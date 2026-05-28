package codexmcp

import (
	"encoding/json"
	"testing"
)

func TestParseElicitApprovalExec(t *testing.T) {
	params := json.RawMessage(`{
		"message": "Allow Codex to run a command?",
		"codex_elicitation": "exec-approval",
		"codex_command": ["rm", "-rf", "/tmp/foo"],
		"codex_cwd": "/work/proj",
		"codex_reason": "cleanup"
	}`)
	a, ok := ParseElicitApproval(params)
	if !ok {
		t.Fatal("ok = false, want true")
	}
	if a.Kind != "exec" {
		t.Fatalf("Kind = %q, want exec", a.Kind)
	}
	if len(a.Command) != 3 || a.Command[0] != "rm" {
		t.Fatalf("Command = %v", a.Command)
	}
	if a.Cwd != "/work/proj" {
		t.Fatalf("Cwd = %q", a.Cwd)
	}
	if a.Reason != "cleanup" {
		t.Fatalf("Reason = %q", a.Reason)
	}
	if a.Message == "" {
		t.Fatal("Message empty")
	}
}

func TestParseElicitApprovalPatch(t *testing.T) {
	params := json.RawMessage(`{
		"message": "Allow Codex to apply a patch?",
		"codex_elicitation": "patch-approval",
		"codex_changes": {
			"a.go": {"type": "add", "content": "x"},
			"b.go": {"type": "update", "unified_diff": "..."}
		},
		"codex_reason": "implement feature"
	}`)
	a, ok := ParseElicitApproval(params)
	if !ok {
		t.Fatal("ok = false, want true")
	}
	if a.Kind != "patch" {
		t.Fatalf("Kind = %q, want patch", a.Kind)
	}
	if len(a.Changes) != 2 {
		t.Fatalf("Changes len = %d, want 2", len(a.Changes))
	}
	if _, ok := a.Changes["a.go"]; !ok {
		t.Fatalf("missing a.go in changes: %v", a.Changes)
	}
}

func TestParseElicitApprovalUnknownDiscriminator(t *testing.T) {
	params := json.RawMessage(`{"message":"?","codex_elicitation":"network-approval"}`)
	a, ok := ParseElicitApproval(params)
	if ok {
		t.Fatalf("ok = true, want false for unknown discriminator (got kind %q)", a.Kind)
	}
}

func TestParseElicitApprovalEmpty(t *testing.T) {
	if _, ok := ParseElicitApproval(nil); ok {
		t.Fatal("ok = true for nil params, want false")
	}
	if _, ok := ParseElicitApproval(json.RawMessage(`{}`)); ok {
		t.Fatal("ok = true for empty object (no discriminator), want false")
	}
}

func TestParseIntID(t *testing.T) {
	if id, ok := parseIntID(json.RawMessage(`42`)); !ok || id != 42 {
		t.Fatalf("parseIntID(42) = %d,%v", id, ok)
	}
	if _, ok := parseIntID(json.RawMessage(`"abc"`)); ok {
		t.Fatal("parseIntID(string) returned ok=true, want false")
	}
	if _, ok := parseIntID(nil); ok {
		t.Fatal("parseIntID(nil) returned ok=true, want false")
	}
}
