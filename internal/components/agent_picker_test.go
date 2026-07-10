package components

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/matt-wright86/mardi-gras/internal/agent"
)

func TestNewAgentPickerListsAllSupportedRuntimes(t *testing.T) {
	picker := NewAgentPicker(90, 30, agent.RuntimeCursor)

	if len(picker.runtimes) != 4 {
		t.Fatalf("picker has %d runtimes, want 4: %v", len(picker.runtimes), picker.runtimes)
	}
	if picker.runtimes[0] != agent.RuntimeCodex ||
		picker.runtimes[1] != agent.RuntimeClaude ||
		picker.runtimes[2] != agent.RuntimeCursor ||
		picker.runtimes[3] != agent.RuntimeCopilot {
		t.Fatalf("unexpected picker runtimes: %v", picker.runtimes)
	}
	if picker.cursor != 2 {
		t.Fatalf("cursor = %d, want cursor on Cursor CLI", picker.cursor)
	}
}

func TestAgentPickerNavigationAndSelection(t *testing.T) {
	picker := NewAgentPicker(90, 30, agent.RuntimeCodex)
	picker, _ = picker.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	if picker.cursor != 1 {
		t.Fatalf("cursor = %d, want 1 after down", picker.cursor)
	}
	_, cmd := picker.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter to return a picker result command")
	}
	result, ok := cmd().(AgentPickerResult)
	if !ok {
		t.Fatalf("expected AgentPickerResult, got %T", cmd())
	}
	if result.Cancelled || result.Runtime != agent.RuntimeClaude {
		t.Fatalf("result = %+v, want Claude selection", result)
	}
}

func TestAgentPickerEscapeCancels(t *testing.T) {
	picker := NewAgentPicker(90, 30, agent.RuntimeCodex)
	_, cmd := picker.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected escape to return a picker result command")
	}
	result, ok := cmd().(AgentPickerResult)
	if !ok || !result.Cancelled {
		t.Fatalf("expected cancelled AgentPickerResult, got %#v", cmd())
	}
}

func TestAgentPickerViewShowsAllChoices(t *testing.T) {
	picker := NewAgentPicker(90, 30, agent.RuntimeCodex)
	view := picker.View()
	for _, want := range []string{"CHOOSE AN AGENT", "Codex", "Claude Code", "Cursor CLI", "GitHub Copilot", "issue note unsent"} {
		if !strings.Contains(view, want) {
			t.Errorf("picker view missing %q:\n%s", want, view)
		}
	}
}
