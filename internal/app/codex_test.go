package app

import (
	"encoding/json"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/matt-wright86/mardi-gras/internal/codexmcp"
	"github.com/matt-wright86/mardi-gras/internal/views"
)

func TestKeyMTogglesCodexOverlay(t *testing.T) {
	got := setupModel(t)
	if got.showCodex {
		t.Fatal("expected showCodex=false initially")
	}

	// First press: should open the overlay.
	model, _ := got.Update(tea.KeyPressMsg{Code: 'M', Text: "M"})
	got = model.(Model)
	if !got.showCodex {
		t.Fatal("expected showCodex=true after M")
	}

	// Second press: should close the overlay.
	model, _ = got.Update(tea.KeyPressMsg{Code: 'M', Text: "M"})
	got = model.(Model)
	if got.showCodex {
		t.Fatal("expected showCodex=false after second M")
	}
}

func TestKeyMOpeningClearsOtherOverlays(t *testing.T) {
	got := setupModel(t)
	got.showProblems = true
	got.showGasTown = true
	got.showDoctor = true

	model, _ := got.Update(tea.KeyPressMsg{Code: 'M', Text: "M"})
	got = model.(Model)
	if !got.showCodex {
		t.Fatal("showCodex should be true")
	}
	if got.showProblems || got.showGasTown || got.showDoctor {
		t.Fatal("other overlays should be cleared")
	}
}

func TestCodexEventMsgAppendsToTranscript(t *testing.T) {
	got := setupModel(t)
	issueID := got.parade.SelectedIssue.ID

	got.codexSessions[issueID] = &codexSession{
		state: &views.CodexTranscriptState{
			IssueID: issueID,
			Status:  "running",
			StartAt: time.Now(),
		},
	}

	raw, _ := json.Marshal(map[string]string{
		"type":    "agent_message",
		"message": "hi from codex",
	})
	ev := codexmcp.CodexEvent{Msg: raw}
	model, _ := got.Update(codexEventMsg{issueID: issueID, ev: ev})
	got = model.(Model)

	sess := got.codexSessions[issueID]
	if len(sess.state.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(sess.state.Entries))
	}
	if sess.state.Entries[0].Title != "hi from codex" {
		t.Fatalf("title = %q", sess.state.Entries[0].Title)
	}
}

func TestCodexDoneMsgMarksSessionTerminal(t *testing.T) {
	got := setupModel(t)
	issueID := got.parade.SelectedIssue.ID

	got.codexSessions[issueID] = &codexSession{
		state: &views.CodexTranscriptState{
			IssueID: issueID,
			Status:  "running",
			StartAt: time.Now(),
		},
	}

	model, _ := got.Update(codexDoneMsg{
		issueID: issueID,
		result:  codexmcp.SessionResult{ThreadID: "tid", Content: "all done"},
	})
	got = model.(Model)
	sess := got.codexSessions[issueID]
	if sess.state.Status != "done" {
		t.Fatalf("status = %q, want done", sess.state.Status)
	}
	if sess.state.EndAt.IsZero() {
		t.Fatal("EndAt should be set")
	}
}
