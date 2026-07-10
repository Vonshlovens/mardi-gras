package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/matt-wright86/mardi-gras/internal/ui"
)

// ThemePickerResult reports a live preview, accepted selection, or cancellation
// to the root model. The root owns applying the theme so every view can refresh
// its cached styles at the same time.
type ThemePickerResult struct {
	Index     int
	Preview   bool
	Accepted  bool
	Cancelled bool
}

// ThemePicker mirrors Tuxedo's keyboard-first picker: j/k previews, Enter
// keeps the preview, and Esc restores the palette active when it opened.
type ThemePicker struct {
	cursor   int
	original int
	width    int
	height   int
}

// NewThemePicker creates a picker with the current theme highlighted.
func NewThemePicker(width, height, current int) ThemePicker {
	count := len(ui.Themes())
	if count > 0 {
		current %= count
		if current < 0 {
			current += count
		}
	}
	return ThemePicker{cursor: current, original: current, width: width, height: height}
}

// SetSize updates the dimensions used to center the picker.
func (p *ThemePicker) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// Update handles Tuxedo-compatible picker navigation.
func (p ThemePicker) Update(msg tea.Msg) (ThemePicker, tea.Cmd) {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}

	count := len(ui.Themes())
	if count == 0 {
		return p, nil
	}

	switch km.String() {
	case "esc", "q":
		return p, func() tea.Msg {
			return ThemePickerResult{Index: p.original, Cancelled: true}
		}
	case "j", "down", "ctrl+n":
		p.cursor = (p.cursor + 1) % count
		return p, func() tea.Msg {
			return ThemePickerResult{Index: p.cursor, Preview: true}
		}
	case "k", "up", "ctrl+p":
		p.cursor = (p.cursor + count - 1) % count
		return p, func() tea.Msg {
			return ThemePickerResult{Index: p.cursor, Preview: true}
		}
	case "enter":
		return p, func() tea.Msg {
			return ThemePickerResult{Index: p.cursor, Accepted: true}
		}
	}

	return p, nil
}

// View renders the selected palette over the current live preview.
func (p ThemePicker) View() string {
	contentWidth := p.width - 12
	if contentWidth > 62 {
		contentWidth = 62
	}
	if contentWidth < 42 {
		contentWidth = 42
	}

	title := ui.HelpTitle.Width(contentWidth).Render(ui.FleurDeLis + " THEMES")
	subtitle := ui.HelpSubtitle.Width(contentWidth).Render("Tuxedo palettes · preview before choosing")
	separator := lipgloss.NewStyle().Foreground(ui.Border).Render(strings.Repeat("─", contentWidth))

	themes := ui.Themes()
	rows := make([]string, 0, len(themes))
	for i, theme := range themes {
		cursor := "  "
		nameStyle := ui.HelpDesc
		if i == p.cursor {
			cursor = ui.ItemCursor.Render("▶ ")
			nameStyle = lipgloss.NewStyle().Foreground(ui.White).Bold(true)
		}
		row := cursor + nameStyle.Render(theme.Name)
		if i == p.cursor {
			row = ui.ItemSelectedBg.Width(contentWidth).Render(row)
		}
		rows = append(rows, row)
	}

	hint := ui.HelpHint.Width(contentWidth).Render("j/k navigate · enter select · esc cancel")
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		separator,
		strings.Join(rows, "\n"),
		"",
		hint,
	)

	box := ui.HelpOverlayBg.Width(contentWidth + 4).Render(content)
	return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center, box)
}
