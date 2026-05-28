package components

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestApprovalDialogDefaultApprove(t *testing.T) {
	ad := NewApprovalDialog("exec", "Run a command?", []string{"ls", "-la"}, "/work", "", nil, 80, 24)
	_, cmd := ad.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd from enter")
	}
	res, ok := cmd().(ApprovalDialogResult)
	if !ok {
		t.Fatalf("expected ApprovalDialogResult, got %T", cmd())
	}
	if res.Cancelled {
		t.Fatal("unexpected Cancelled")
	}
	if res.Decision != "approved" {
		t.Fatalf("Decision = %q, want approved (default selection)", res.Decision)
	}
}

func TestApprovalDialogNavigateToDeny(t *testing.T) {
	ad := NewApprovalDialog("exec", "Run?", []string{"rm", "-rf", "/"}, "/work", "danger", nil, 80, 24)
	// Decisions: approved, approved_for_session, denied, abort. Two downs → denied.
	ad, _ = ad.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	ad, _ = ad.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	_, cmd := ad.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	res := cmd().(ApprovalDialogResult)
	if res.Decision != "denied" {
		t.Fatalf("Decision = %q, want denied", res.Decision)
	}
}

func TestApprovalDialogClampsSelection(t *testing.T) {
	ad := NewApprovalDialog("exec", "Run?", []string{"ls"}, "/work", "", nil, 80, 24)
	// Up at top stays at top.
	ad, _ = ad.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	// Many downs clamp at the last entry (abort).
	for range 10 {
		ad, _ = ad.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	_, cmd := ad.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	res := cmd().(ApprovalDialogResult)
	if res.Decision != "abort" {
		t.Fatalf("Decision = %q, want abort (clamped to last)", res.Decision)
	}
}

func TestApprovalDialogEscCancels(t *testing.T) {
	ad := NewApprovalDialog("exec", "Run?", []string{"ls"}, "/work", "", nil, 80, 24)
	_, cmd := ad.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	res := cmd().(ApprovalDialogResult)
	if !res.Cancelled {
		t.Fatal("expected Cancelled=true from esc")
	}
}

func TestApprovalDialogViewExec(t *testing.T) {
	ad := NewApprovalDialog("exec", "Allow?", []string{"git", "push"}, "/repo", "deploy", nil, 80, 24)
	v := ad.View()
	if !strings.Contains(v, "RUN A COMMAND") {
		t.Fatalf("exec view missing heading:\n%s", v)
	}
	if !strings.Contains(v, "git push") {
		t.Fatalf("exec view missing command:\n%s", v)
	}
	if !strings.Contains(v, "/repo") {
		t.Fatalf("exec view missing cwd:\n%s", v)
	}
	if !strings.Contains(v, "deploy") {
		t.Fatalf("exec view missing reason:\n%s", v)
	}
}

func TestApprovalDialogViewPatch(t *testing.T) {
	ad := NewApprovalDialog("patch", "Allow?", nil, "", "implement", []string{"a.go", "b.go"}, 80, 24)
	v := ad.View()
	if !strings.Contains(v, "APPLY A PATCH") {
		t.Fatalf("patch view missing heading:\n%s", v)
	}
	if !strings.Contains(v, "a.go") || !strings.Contains(v, "b.go") {
		t.Fatalf("patch view missing file list:\n%s", v)
	}
	if !strings.Contains(v, "2 file(s)") {
		t.Fatalf("patch view missing file count:\n%s", v)
	}
}
