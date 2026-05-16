package codexmcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// Transport is the read/write side of a JSON-RPC connection. The default
// transport spawns `codex mcp-server` as a subprocess (see SubprocessTransport);
// tests use an in-memory pipe pair.
type Transport interface {
	// Reader returns a stream of newline-delimited JSON objects from the server.
	Reader() io.Reader
	// Writer accepts newline-delimited JSON objects to send to the server.
	Writer() io.Writer
	// Close shuts the transport down. After Close, Reader will return io.EOF
	// after draining any in-flight bytes, and Writer will return an error.
	Close() error
}

// Client speaks JSON-RPC to a Codex MCP server over the supplied Transport.
// One Client manages one subprocess (or pipe pair). The zero value is not
// usable — construct with Dial.
type Client struct {
	t     Transport
	enc   *json.Encoder
	dec   *json.Decoder
	w     io.Writer
	wMu   sync.Mutex
	rdErr atomic.Value // error

	nextID  atomic.Int64
	pending sync.Map // map[int]chan response

	eventsCh     chan CodexEvent
	eventsClosed atomic.Bool

	closeOnce sync.Once
	closeErr  error
	done      chan struct{}
}

// ClientOption customizes Dial behavior.
type ClientOption func(*clientOptions)

type clientOptions struct {
	clientVersion string
	eventBuffer   int
}

// WithClientVersion overrides the version reported to the server in initialize.
// Defaults to "dev".
func WithClientVersion(v string) ClientOption {
	return func(o *clientOptions) { o.clientVersion = v }
}

// WithEventBuffer sets the buffered channel size for event delivery. Defaults
// to 64. A larger buffer reduces backpressure on the reader goroutine when the
// consumer (e.g. BubbleTea) is slow to drain.
func WithEventBuffer(n int) ClientOption {
	return func(o *clientOptions) {
		if n > 0 {
			o.eventBuffer = n
		}
	}
}

// Dial constructs a Client around the given Transport and performs the MCP
// initialize handshake. It does not start a Codex session — call StartSession
// for that.
//
// On any handshake failure the transport is closed before returning.
func Dial(ctx context.Context, t Transport, opts ...ClientOption) (*Client, error) {
	o := clientOptions{
		clientVersion: clientVersionFallback,
		eventBuffer:   64,
	}
	for _, opt := range opts {
		opt(&o)
	}

	c := &Client{
		t:        t,
		dec:      json.NewDecoder(bufio.NewReader(t.Reader())),
		w:        t.Writer(),
		eventsCh: make(chan CodexEvent, o.eventBuffer),
		done:     make(chan struct{}),
	}
	c.enc = json.NewEncoder(c.w)

	go c.readLoop()

	if err := c.initialize(ctx, o.clientVersion); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := c.notify(methodNotifyInitialized, nil); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("notify initialized: %w", err)
	}
	return c, nil
}

// Events returns a receive channel of `codex/event` notifications. The channel
// is closed when the client is shut down. Consumers must drain promptly or
// risk the reader goroutine blocking; the buffer is sized via WithEventBuffer.
func (c *Client) Events() <-chan CodexEvent {
	return c.eventsCh
}

// Done returns a channel that is closed when the underlying transport hits
// EOF or an unrecoverable read error. Reading from it is non-blocking only
// after shutdown.
func (c *Client) Done() <-chan struct{} {
	return c.done
}

// ReadError returns the error that terminated the read loop, if any.
func (c *Client) ReadError() error {
	if v := c.rdErr.Load(); v != nil {
		if e, ok := v.(error); ok {
			return e
		}
	}
	return nil
}

// Close shuts the client and transport down. Safe to call multiple times.
// In-flight Call goroutines will receive context.Canceled on the next read.
func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		c.closeErr = c.t.Close()
		// Wake any in-flight callers.
		c.pending.Range(func(_, v any) bool {
			if ch, ok := v.(chan response); ok {
				close(ch)
			}
			return true
		})
	})
	return c.closeErr
}

// Call issues a JSON-RPC request and waits for the matching response. It
// blocks until ctx is canceled, the server responds, or the transport closes.
func (c *Client) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	id := int(c.nextID.Add(1))
	ch := make(chan response, 1)
	c.pending.Store(id, ch)
	defer c.pending.Delete(id)

	raw, err := marshalParams(params)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}
	req := request{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Method:  method,
		Params:  raw,
	}
	if err := c.writeJSON(req); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		if err := c.ReadError(); err != nil {
			return nil, fmt.Errorf("transport closed: %w", err)
		}
		return nil, errors.New("transport closed")
	case resp, ok := <-ch:
		if !ok {
			return nil, errors.New("client closed")
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	}
}

// notify sends a JSON-RPC notification (no id, no response expected).
func (c *Client) notify(method string, params any) error {
	raw, err := marshalParams(params)
	if err != nil {
		return err
	}
	n := notification{
		JSONRPC: jsonRPCVersion,
		Method:  method,
		Params:  raw,
	}
	return c.writeJSON(n)
}

func (c *Client) initialize(ctx context.Context, version string) error {
	params := initializeParams{
		ProtocolVersion: mcpProtocolVersion,
		Capabilities:    map[string]any{},
		ClientInfo: clientInfo{
			Name:    clientName,
			Version: version,
		},
	}
	_, err := c.Call(ctx, methodInitialize, params)
	if err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	return nil
}

func (c *Client) writeJSON(v any) error {
	c.wMu.Lock()
	defer c.wMu.Unlock()
	return c.enc.Encode(v)
}

func marshalParams(params any) (json.RawMessage, error) {
	if params == nil {
		return nil, nil
	}
	if raw, ok := params.(json.RawMessage); ok {
		return raw, nil
	}
	b, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// readLoop dispatches inbound JSON-RPC messages. Responses (id != null) are
// routed to the matching pending channel; notifications named codex/event are
// decoded and forwarded on eventsCh. Unknown notifications are dropped.
func (c *Client) readLoop() {
	defer func() {
		if !c.eventsClosed.Swap(true) {
			close(c.eventsCh)
		}
		close(c.done)
	}()
	for {
		var msg response
		if err := c.dec.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			c.rdErr.Store(err)
			return
		}
		switch {
		case msg.ID != nil:
			if v, ok := c.pending.LoadAndDelete(*msg.ID); ok {
				if ch, ok := v.(chan response); ok {
					ch <- msg
				}
			}
		case msg.Method == methodCodexEvent:
			var ev CodexEvent
			if err := json.Unmarshal(msg.Params, &ev); err == nil {
				c.eventsCh <- ev
			}
		default:
			// Unknown notification — drop.
		}
	}
}
