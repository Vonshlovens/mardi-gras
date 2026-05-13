package app

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/matt-wright86/mardi-gras/internal/data"
)

func setupDeferredKeyModel(t *testing.T) (model Model, filter func(tea.Model, tea.Msg) tea.Msg) {
	t.Helper()

	guard := NewOSCGuard()
	filter = guard.Filter()
	issues := []data.Issue{
		testIssue("open-1", data.StatusOpen),
		testIssue("open-2", data.StatusOpen),
	}
	m := NewWithGuard(issues, data.Source{}, data.DefaultBlockingTypes, guard, false)
	m.startedAt = time.Now().Add(-time.Second)

	readyModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
	return readyModel.(Model), filter
}

func sendFiltered(t *testing.T, m Model, filter func(tea.Model, tea.Msg) tea.Msg, msg tea.Msg) (Model, tea.Cmd, bool) {
	t.Helper()

	filtered := filter(nil, msg)
	if filtered == nil {
		return m, nil, false
	}

	model, cmd := m.Update(filtered)
	return model.(Model), cmd, true
}

func TestDeferredKeyPassesAfterDelay(t *testing.T) {
	m, filter := setupDeferredKeyModel(t)

	var cmd tea.Cmd
	var ok bool
	m, cmd, ok = sendFiltered(t, m, filter, tea.KeyPressMsg{Code: 'q', Text: "q"})
	if !ok {
		t.Fatal("expected q to pass filter")
	}
	if cmd == nil {
		t.Fatal("expected deferred command after staging q")
	}

	msg := cmd()
	deferred, ok := msg.(deferredKeyMsg)
	if !ok {
		t.Fatalf("expected deferredKeyMsg, got %T", msg)
	}

	model, quitCmd := m.Update(deferred)
	m = model.(Model)
	if quitCmd == nil {
		t.Fatal("expected q to produce a quit command after delay")
	}
	quitMsg := quitCmd()
	if _, ok := quitMsg.(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", quitMsg)
	}
	if len(m.pendingKeys) != 0 {
		t.Fatal("expected pending key queue to be cleared after deferred delivery")
	}
}

func TestDeferredKeyDropsSuspiciousPair(t *testing.T) {
	m, filter := setupDeferredKeyModel(t)

	m, firstCmd, ok := sendFiltered(t, m, filter, tea.KeyPressMsg{Code: '1', Text: "1"})
	if !ok {
		t.Fatal("expected 1 to pass filter")
	}
	if firstCmd == nil {
		t.Fatal("expected deferred command after staging 1")
	}

	time.Sleep(20 * time.Millisecond)

	m, secondCmd, ok := sendFiltered(t, m, filter, tea.KeyPressMsg{Code: ';', Text: ";"})
	if !ok {
		t.Fatal("expected ; to reach Update for pair detection")
	}
	if secondCmd != nil {
		t.Fatal("expected suspicious pair to be dropped without routing a command")
	}
	if len(m.pendingKeys) != 0 {
		t.Fatal("expected pending key queue to be cleared after suspicious pair drop")
	}

	model, cmd := m.Update(firstCmd())
	m = model.(Model)
	if cmd != nil {
		t.Fatal("expected stale deferred message to be ignored after pair drop")
	}
	if len(m.pendingKeys) != 0 {
		t.Fatal("expected no pending key after stale deferred message")
	}
}

func TestDeferredKeyDropsAfterFilterSuppression(t *testing.T) {
	m, filter := setupDeferredKeyModel(t)

	m, firstCmd, ok := sendFiltered(t, m, filter, tea.KeyPressMsg{Code: ']', Text: "]"})
	if !ok {
		t.Fatal("expected ] to pass filter and stage")
	}
	if firstCmd == nil {
		t.Fatal("expected deferred command after staging ]")
	}

	if filtered := filter(nil, tea.KeyPressMsg{Code: '1', Text: "1"}); filtered != nil {
		t.Fatal("expected 1 to be suppressed by the shared guard filter")
	}

	model, cmd := m.Update(firstCmd())
	m = model.(Model)
	if cmd != nil {
		t.Fatal("expected deferred ] to be dropped after later filter suppression")
	}
	if len(m.pendingKeys) != 0 {
		t.Fatal("expected pending key queue to be cleared after deferred drop")
	}
}

func TestIsLikelyDeferredFragmentPair(t *testing.T) {
	tests := []struct {
		name  string
		first tea.KeyPressMsg
		next  tea.KeyPressMsg
		want  bool
	}{
		{"[ + digit is fragment", tea.KeyPressMsg{Code: '[', Text: "["}, tea.KeyPressMsg{Code: '1', Text: "1"}, true},
		{"[ + shift CSI tail is fragment", tea.KeyPressMsg{Code: '[', Text: "["}, tea.KeyPressMsg{Code: 'A', Text: "A", Mod: tea.ModShift}, true},
		{"[ + non-digit non-CSI is not fragment", tea.KeyPressMsg{Code: '[', Text: "["}, tea.KeyPressMsg{Code: 'q', Text: "q"}, false},
		{"] + digit is fragment", tea.KeyPressMsg{Code: ']', Text: "]"}, tea.KeyPressMsg{Code: '5', Text: "5"}, true},
		{"] + letter is not fragment", tea.KeyPressMsg{Code: ']', Text: "]"}, tea.KeyPressMsg{Code: 'a', Text: "a"}, false},
		{"digit + ; is fragment", tea.KeyPressMsg{Code: '7', Text: "7"}, tea.KeyPressMsg{Code: ';', Text: ";"}, true},
		{"digit + non-semicolon is not", tea.KeyPressMsg{Code: '7', Text: "7"}, tea.KeyPressMsg{Code: '0', Text: "0"}, false},
		{"; + r is fragment", tea.KeyPressMsg{Code: ';', Text: ";"}, tea.KeyPressMsg{Code: 'r', Text: "r"}, true},
		{"; + g is fragment", tea.KeyPressMsg{Code: ';', Text: ";"}, tea.KeyPressMsg{Code: 'g', Text: "g"}, true},
		{"; + other is not", tea.KeyPressMsg{Code: ';', Text: ";"}, tea.KeyPressMsg{Code: 'x', Text: "x"}, false},
		{"default case (letter + letter)", tea.KeyPressMsg{Code: 'q', Text: "q"}, tea.KeyPressMsg{Code: 'q', Text: "q"}, false},
		{"non-printable first rejects", tea.KeyPressMsg{Code: 0x01}, tea.KeyPressMsg{Code: '1', Text: "1"}, false},
		{"non-printable second rejects", tea.KeyPressMsg{Code: '[', Text: "["}, tea.KeyPressMsg{Code: 0x01}, false},
		{"ModCtrl on first rejects fragment", tea.KeyPressMsg{Code: '[', Text: "[", Mod: tea.ModCtrl}, tea.KeyPressMsg{Code: '1', Text: "1"}, false},
		{"ModCtrl on second rejects fragment", tea.KeyPressMsg{Code: '[', Text: "["}, tea.KeyPressMsg{Code: '1', Text: "1", Mod: tea.ModCtrl}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isLikelyDeferredFragmentPair(tc.first, tc.next)
			if got != tc.want {
				t.Errorf("isLikelyDeferredFragmentPair(%v, %v) = %v, want %v", tc.first, tc.next, got, tc.want)
			}
		})
	}
}

func TestDeferredQuickActionDigitWaitsForTimer(t *testing.T) {
	m, filter := setupDeferredKeyModel(t)

	m, firstCmd, ok := sendFiltered(t, m, filter, tea.KeyPressMsg{Code: '3', Text: "3"})
	if !ok {
		t.Fatal("expected 3 to pass filter and stage")
	}
	if firstCmd == nil {
		t.Fatal("expected deferred command after staging 3")
	}
	if len(m.pendingKeys) != 1 {
		t.Fatalf("expected 1 pending key after staging 3, got %d", len(m.pendingKeys))
	}

	time.Sleep(20 * time.Millisecond)

	m, secondCmd, ok := sendFiltered(t, m, filter, tea.KeyPressMsg{Code: '2', Text: "2"})
	if !ok {
		t.Fatal("expected 2 to pass filter and stage")
	}
	if secondCmd == nil {
		t.Fatal("expected second deferred command after staging 2 behind 3")
	}
	if len(m.pendingKeys) != 2 {
		t.Fatalf("expected 2 pending keys after staging 3 and 2, got %d", len(m.pendingKeys))
	}

	if filtered := filter(nil, uv.UnknownEvent("\x1b]11;rgb:1f1f/2323/3535")); filtered != nil {
		t.Fatal("expected UnknownEvent to be dropped by shared guard")
	}

	model, cmd := m.Update(firstCmd())
	m = model.(Model)
	if cmd != nil {
		t.Fatal("expected deferred 3 to be dropped after later suspicious input")
	}
	if len(m.pendingKeys) != 1 {
		t.Fatalf("expected only the second pending key to remain, got %d", len(m.pendingKeys))
	}

	model, cmd = m.Update(secondCmd())
	m = model.(Model)
	if cmd != nil {
		t.Fatal("expected deferred 2 to be dropped after later suspicious input")
	}
	if len(m.pendingKeys) != 0 {
		t.Fatal("expected pending key queue to be cleared after both deferred drops")
	}
}
