package components

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/matt-wright86/mardi-gras/internal/ui"
)

func TestNewThemePickerStartsOnCurrentTheme(t *testing.T) {
	picker := NewThemePicker(90, 30, 2)
	if picker.cursor != 2 || picker.original != 2 {
		t.Fatalf("picker = %+v, want cursor and original at 2", picker)
	}
}

func TestThemePickerNavigationPreviewsAndWraps(t *testing.T) {
	picker := NewThemePicker(90, 30, 0)
	picker, cmd := picker.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if picker.cursor != len(ui.Themes())-1 {
		t.Fatalf("up from first cursor = %d, want wrapped last index", picker.cursor)
	}
	if cmd == nil {
		t.Fatal("expected preview command after navigation")
	}
	result, ok := cmd().(ThemePickerResult)
	if !ok || !result.Preview || result.Index != picker.cursor {
		t.Fatalf("navigation result = %#v, want preview for %d", cmd(), picker.cursor)
	}

	picker, cmd = picker.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if picker.cursor != 0 {
		t.Fatalf("down from last cursor = %d, want wrapped first index", picker.cursor)
	}
	result, ok = cmd().(ThemePickerResult)
	if !ok || !result.Preview || result.Index != 0 {
		t.Fatalf("wrap preview result = %#v", cmd())
	}
}

func TestThemePickerEnterAndEscape(t *testing.T) {
	picker := NewThemePicker(90, 30, 1)
	picker, _ = picker.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, cmd := picker.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	result, ok := cmd().(ThemePickerResult)
	if !ok || !result.Accepted || result.Index != 2 {
		t.Fatalf("enter result = %#v, want accepted index 2", cmd())
	}

	_, cmd = picker.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	result, ok = cmd().(ThemePickerResult)
	if !ok || !result.Cancelled || result.Index != 1 {
		t.Fatalf("escape result = %#v, want cancelled original index 1", cmd())
	}
}

func TestThemePickerViewListsEveryTheme(t *testing.T) {
	picker := NewThemePicker(90, 30, 0)
	view := picker.View()
	for _, theme := range ui.Themes() {
		if !strings.Contains(view, theme.Name) {
			t.Errorf("picker view missing theme %q:\n%s", theme.Name, view)
		}
	}
	for _, want := range []string{"THEMES", "j/k navigate", "enter select", "esc cancel"} {
		if !strings.Contains(view, want) {
			t.Errorf("picker view missing %q:\n%s", want, view)
		}
	}
}
