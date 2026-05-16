// Package codexmcp speaks the Model Context Protocol (MCP) over stdio to
// `codex mcp-server`. It is a thin client that exposes the JSON-RPC request /
// response surface plus the `codex/event` notification stream that Codex emits
// while a session is running.
//
// The package is intentionally focused on Codex; it does not aim to be a
// general MCP client. Only the subset of protocol surface mg needs is modeled.
package codexmcp

import (
	"encoding/json"
)

// JSON-RPC envelope types. Codex's MCP server uses JSON-RPC 2.0 framing
// (newline-delimited objects on stdout/stdin).

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// MCP method names used by this client.
const (
	methodInitialize        = "initialize"
	methodNotifyInitialized = "notifications/initialized"
	methodToolsCall         = "tools/call"
	methodCodexEvent        = "codex/event"
	mcpProtocolVersion      = "2025-03-26"
	codexToolName           = "codex"
	codexReplyToolName      = "codex-reply"
	clientName              = "mardi-gras"
	clientVersionFallback   = "dev"
	jsonRPCVersion          = "2.0"
)

// initializeParams is sent in the MCP `initialize` request.
type initializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      clientInfo     `json:"clientInfo"`
}

type clientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// toolsCallParams is the params object for `tools/call`.
type toolsCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// CodexEvent is the unmarshaled `codex/event` notification payload.
// The raw shape is:
//
//	{
//	  "_meta": {"requestId": 2, "threadId": "..."},
//	  "id":   "2",                  // turn id or "" for session-level events
//	  "msg":  {"type": "agent_message", ...event-specific fields...}
//	}
//
// We keep `Msg` as RawMessage and surface typed accessors for the variants mg
// cares about; unknown event types degrade to a generic Type + Raw payload.
type CodexEvent struct {
	Meta EventMeta       `json:"_meta"`
	ID   string          `json:"id"`
	Msg  json.RawMessage `json:"msg"`
}

// EventMeta carries the request/thread correlation for a CodexEvent. The
// requestId matches the JSON-RPC id of the tools/call request that triggered
// the session; threadId is stable across replies to the same session.
type EventMeta struct {
	RequestID int    `json:"requestId"`
	ThreadID  string `json:"threadId"`
}

// EventType returns the `msg.type` discriminator. Returns "" if Msg is empty
// or malformed.
func (e CodexEvent) EventType() string {
	if len(e.Msg) == 0 {
		return ""
	}
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(e.Msg, &probe); err != nil {
		return ""
	}
	return probe.Type
}

// Typed event payloads. We don't try to cover every variant codex emits —
// only the ones mg renders or routes on. Anything unknown is still accessible
// via CodexEvent.Msg as raw JSON.

// AgentMessageEvent is `msg.type == "agent_message"` — the assistant's text
// output for a turn.
type AgentMessageEvent struct {
	Message string `json:"message"`
}

// UserMessageEvent is `msg.type == "user_message"` — the user-visible prompt
// sent into the model.
type UserMessageEvent struct {
	Message string `json:"message"`
}

// AgentReasoningEvent is `msg.type == "agent_reasoning"` — a reasoning summary.
type AgentReasoningEvent struct {
	Text string `json:"text"`
}

// ExecCommandBeginEvent is `msg.type == "exec_command_begin"`.
type ExecCommandBeginEvent struct {
	CallID  string   `json:"call_id"`
	Command []string `json:"command"`
	Cwd     string   `json:"cwd"`
}

// ExecCommandEndEvent is `msg.type == "exec_command_end"`.
type ExecCommandEndEvent struct {
	CallID   string `json:"call_id"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// MCPToolCallBeginEvent is `msg.type == "mcp_tool_call_begin"`.
type MCPToolCallBeginEvent struct {
	CallID     string `json:"call_id"`
	Invocation struct {
		Server string          `json:"server"`
		Tool   string          `json:"tool"`
		Args   json.RawMessage `json:"arguments,omitempty"`
	} `json:"invocation"`
}

// MCPToolCallEndEvent is `msg.type == "mcp_tool_call_end"`.
type MCPToolCallEndEvent struct {
	CallID  string `json:"call_id"`
	IsError bool   `json:"is_error"`
}

// TaskStartedEvent is `msg.type == "task_started"`.
type TaskStartedEvent struct {
	TurnID             string `json:"turn_id"`
	StartedAt          int64  `json:"started_at"`
	ModelContextWindow int    `json:"model_context_window"`
	CollaborationMode  string `json:"collaboration_mode_kind"`
}

// TaskCompleteEvent is `msg.type == "task_complete"`.
type TaskCompleteEvent struct {
	LastAgentMessage string `json:"last_agent_message"`
}

// ErrorEvent is `msg.type == "error"`.
type ErrorEvent struct {
	Message string `json:"message"`
}

// SessionConfiguredEvent is `msg.type == "session_configured"`.
type SessionConfiguredEvent struct {
	SessionID      string `json:"session_id"`
	ThreadID       string `json:"thread_id"`
	Model          string `json:"model"`
	ApprovalPolicy string `json:"approval_policy"`
	Cwd            string `json:"cwd"`
	RolloutPath    string `json:"rollout_path"`
}
