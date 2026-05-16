package codexmcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeServer is a long-lived JSON-RPC peer for tests. A single goroutine reads
// inbound requests and forwards them to the Incoming channel; tests respond
// via Respond / SendEvent which encode onto the writer side.
type fakeServer struct {
	dec *json.Decoder
	enc *json.Encoder

	Incoming chan request

	wMu     sync.Mutex
	stopped bool

	stopOnce sync.Once
	stop     chan struct{}
}

func newFakeServer(t *testing.T) (*Client, *fakeServer) {
	t.Helper()
	transport, sr, sw := newPipePair()
	fs := &fakeServer{
		dec:      json.NewDecoder(bufio.NewReader(sr)),
		enc:      json.NewEncoder(sw),
		Incoming: make(chan request, 16),
		stop:     make(chan struct{}),
	}

	go fs.readLoop(sr, sw)

	// Pre-arm: the handshake fires the initialize request and the
	// notifications/initialized message. Auto-respond to initialize.
	handshakeDone := make(chan struct{})
	go func() {
		defer close(handshakeDone)
		for req := range fs.Incoming {
			if req.Method == methodInitialize {
				fs.Respond(req.ID, json.RawMessage(`{
					"protocolVersion":"2025-03-26",
					"capabilities":{"tools":{"listChanged":true}},
					"serverInfo":{"name":"fake","version":"0.1"}
				}`))
				return
			}
			// Drain anything else — shouldn't happen pre-init.
		}
	}()

	c, err := Dial(context.Background(), transport)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	select {
	case <-handshakeDone:
	case <-time.After(2 * time.Second):
		t.Fatal("handshake timeout")
	}
	// notifications/initialized is a fire-and-forget; ensure it arrived but
	// don't require a response.
	select {
	case n := <-fs.Incoming:
		if n.Method != methodNotifyInitialized {
			t.Fatalf("expected notifications/initialized, got %q", n.Method)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("notifications/initialized not received")
	}

	t.Cleanup(func() {
		_ = c.Close()
		fs.Stop()
	})
	return c, fs
}

func (f *fakeServer) readLoop(sr, sw interface{ Close() error }) {
	defer func() {
		_ = sr.Close()
		_ = sw.Close()
		close(f.Incoming)
	}()
	for {
		var req request
		if err := f.dec.Decode(&req); err != nil {
			return
		}
		select {
		case <-f.stop:
			return
		case f.Incoming <- req:
		}
	}
}

func (f *fakeServer) Stop() {
	f.stopOnce.Do(func() {
		f.wMu.Lock()
		f.stopped = true
		f.wMu.Unlock()
		close(f.stop)
	})
}

func (f *fakeServer) Respond(id int, result json.RawMessage) {
	f.wMu.Lock()
	defer f.wMu.Unlock()
	if f.stopped {
		return
	}
	_ = f.enc.Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	})
}

func (f *fakeServer) RespondError(id, code int, msg string) {
	f.wMu.Lock()
	defer f.wMu.Unlock()
	if f.stopped {
		return
	}
	_ = f.enc.Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error":   map[string]any{"code": code, "message": msg},
	})
}

func (f *fakeServer) SendEvent(reqID int, threadID, msg string) {
	f.wMu.Lock()
	defer f.wMu.Unlock()
	if f.stopped {
		return
	}
	_ = f.enc.Encode(map[string]any{
		"jsonrpc": "2.0",
		"method":  methodCodexEvent,
		"params": map[string]any{
			"_meta": map[string]any{"requestId": reqID, "threadId": threadID},
			"id":    "",
			"msg":   json.RawMessage(msg),
		},
	})
}

func (f *fakeServer) SendRaw(v any) {
	f.wMu.Lock()
	defer f.wMu.Unlock()
	if f.stopped {
		return
	}
	_ = f.enc.Encode(v)
}

func (f *fakeServer) Expect(t *testing.T, method string, timeout time.Duration) request {
	t.Helper()
	select {
	case req := <-f.Incoming:
		if req.Method != method {
			t.Fatalf("expected method %q, got %q", method, req.Method)
		}
		return req
	case <-time.After(timeout):
		t.Fatalf("timed out waiting for %s", method)
		return request{}
	}
}

func TestDialPerformsInitialize(t *testing.T) {
	c, _ := newFakeServer(t)
	if c == nil {
		t.Fatal("expected client")
	}
}

func TestCallSuccess(t *testing.T) {
	c, fs := newFakeServer(t)
	go func() {
		req := fs.Expect(t, "ping", 2*time.Second)
		fs.Respond(req.ID, json.RawMessage(`{"ok":true}`))
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	res, err := c.Call(ctx, "ping", nil)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if !strings.Contains(string(res), "\"ok\":true") {
		t.Fatalf("unexpected result %s", res)
	}
}

func TestCallError(t *testing.T) {
	c, fs := newFakeServer(t)
	go func() {
		req := fs.Expect(t, "bogus", 2*time.Second)
		fs.RespondError(req.ID, -32601, "method not found")
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := c.Call(ctx, "bogus", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "method not found") {
		t.Fatalf("error message lost: %v", err)
	}
}

func TestCallContextCancel(t *testing.T) {
	c, fs := newFakeServer(t)
	// Drain the request so writes don't backpressure.
	go func() { _ = fs.Expect(t, "hang", 2*time.Second) }()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	_, err := c.Call(ctx, "hang", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected Canceled, got %v", err)
	}
}

func TestEventsRoutedToClientChannel(t *testing.T) {
	c, fs := newFakeServer(t)
	go fs.SendEvent(1, "thread-A", `{"type":"agent_message","message":"hello"}`)
	select {
	case ev := <-c.Events():
		if ev.Meta.RequestID != 1 || ev.Meta.ThreadID != "thread-A" {
			t.Fatalf("bad meta: %+v", ev.Meta)
		}
		if ev.EventType() != "agent_message" {
			t.Fatalf("bad event type: %q", ev.EventType())
		}
		var am AgentMessageEvent
		if err := json.Unmarshal(ev.Msg, &am); err != nil {
			t.Fatalf("decode agent_message: %v", err)
		}
		if am.Message != "hello" {
			t.Fatalf("bad message: %q", am.Message)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestUnknownNotificationIsDropped(t *testing.T) {
	c, fs := newFakeServer(t)
	fs.SendRaw(map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/cancelled",
		"params":  map[string]any{"requestId": 99},
	})
	fs.SendEvent(2, "t", `{"type":"task_started"}`)
	select {
	case ev := <-c.Events():
		if ev.EventType() != "task_started" {
			t.Fatalf("got %q — was the unknown notification leaked?", ev.EventType())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestCloseIsIdempotentAndUnblocksCallers(t *testing.T) {
	c, _ := newFakeServer(t)
	go func() {
		time.Sleep(20 * time.Millisecond)
		_ = c.Close()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := c.Call(ctx, "hang", nil)
	if err == nil {
		t.Fatal("expected error after Close")
	}
	if err := c.Close(); err != nil {
		t.Fatalf("double close: %v", err)
	}
}
