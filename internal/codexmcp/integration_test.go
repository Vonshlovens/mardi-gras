//go:build integration

package codexmcp_test

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"
	"time"

	"github.com/matt-wright86/mardi-gras/internal/codexmcp"
)

// TestIntegrationRealCodex exercises a real `codex mcp-server` end-to-end:
// spawn the subprocess, perform the initialize handshake, start a session
// with a trivial prompt, observe at least one agent_message event, and
// confirm Done() fires with a non-empty content + threadId.
//
// Gated by build tag `integration` because it requires `codex` on PATH and
// makes a real model call (cost + latency).
//
// Run with:
//
//	go test -tags=integration ./internal/codexmcp -run TestIntegrationRealCodex -v -timeout 120s
func TestIntegrationRealCodex(t *testing.T) {
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex not on PATH")
	}

	transport, err := codexmcp.SpawnSubprocess(codexmcp.WithDir("/tmp"))
	if err != nil {
		t.Fatalf("spawn: %v", err)
	}
	defer transport.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := codexmcp.Dial(ctx, transport, codexmcp.WithClientVersion("integration-test"))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()

	sess, err := client.StartSession(ctx, codexmcp.SessionOptions{
		Prompt:         "Say the single word 'pong' and stop.",
		Cwd:            "/tmp",
		Sandbox:        "read-only",
		ApprovalPolicy: "never",
	})
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}

	sawAgentMessage := false
	timeout := time.NewTimer(110 * time.Second)
	defer timeout.Stop()

loop:
	for {
		select {
		case ev, ok := <-sess.Events():
			if !ok {
				break loop
			}
			t.Logf("event: %s thread=%s", ev.EventType(), ev.Meta.ThreadID)
			if ev.EventType() == "agent_message" {
				sawAgentMessage = true
				var am codexmcp.AgentMessageEvent
				if err := json.Unmarshal(ev.Msg, &am); err != nil {
					t.Fatalf("decode agent_message: %v", err)
				}
				t.Logf("agent: %q", am.Message)
			}
		case res := <-sess.Done():
			t.Logf("done: threadId=%s err=%v content=%q", res.ThreadID, res.Err, res.Content)
			if res.Err != nil {
				t.Fatalf("session ended with error: %v", res.Err)
			}
			if res.ThreadID == "" {
				t.Fatal("expected threadId")
			}
			if res.Content == "" {
				t.Fatal("expected content")
			}
			break loop
		case <-timeout.C:
			t.Fatal("integration test timeout")
		}
	}

	if !sawAgentMessage {
		t.Fatal("did not observe any agent_message event")
	}
}
