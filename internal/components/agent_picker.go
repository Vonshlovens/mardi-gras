package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/matt-wright86/mardi-gras/internal/agent"
	"github.com/matt-wright86/mardi-gras/internal/ui"
)

// AgentPickerResult is sent when the agent runtime picker closes.
type AgentPickerResult struct {
	Runtime   agent.Runtime
	Cancelled bool
}

// AgentPicker is a compact runtime selection modal. It intentionally lists
// every supported runtime, including ones not currently installed, so users
// can discover the available launch options without consulting a config file.
type AgentPicker struct {
	runtimes []agent.Runtime
	cursor   int
	width    int
	height   int
}

// NewAgentPicker constructs a picker and selects preferred when it is one of
// the known runtimes. The preference usually comes from MG_AGENT_RUNTIME or
// the --agent flag, preserving its value as a convenient default.
func NewAgentPicker(width, height int, preferred agent.Runtime) AgentPicker {
	runtimes := agent.SupportedRuntimes()
	cursor := 0
	for i, runtime := range runtimes {
		if runtime == preferred {
			cursor = i
			break
		}
	}
	return AgentPicker{
		runtimes: runtimes,
		cursor:   cursor,
		width:    width,
		height:   height,
	}
}

// SetSize updates the modal's placement dimensions.
func (p *AgentPicker) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// Update handles navigation and selection for the picker.
func (p AgentPicker) Update(msg tea.Msg) (AgentPicker, tea.Cmd) {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}

	switch km.String() {
	case "esc", "q":
		return p, func() tea.Msg {
			return AgentPickerResult{Cancelled: true}
		}
	case "j", "down", "ctrl+n":
		if p.cursor < len(p.runtimes)-1 {
			p.cursor++
		}
	case "k", "up", "ctrl+p":
		if p.cursor > 0 {
			p.cursor--
		}
	case "enter":
		if p.cursor >= 0 && p.cursor < len(p.runtimes) {
			runtime := p.runtimes[p.cursor]
			return p, func() tea.Msg {
				return AgentPickerResult{Runtime: runtime}
			}
		}
		return p, func() tea.Msg {
			return AgentPickerResult{Cancelled: true}
		}
	}

	return p, nil
}

// View renders the centered agent picker modal.
func (p AgentPicker) View() string {
	contentWidth := p.width - 12
	if contentWidth > 62 {
		contentWidth = 62
	}
	if contentWidth < 42 {
		contentWidth = 42
	}

	title := ui.HelpTitle.Width(contentWidth).Render(ui.FleurDeLis + " CHOOSE AN AGENT")
	subtitle := ui.HelpSubtitle.Width(contentWidth).Render("Start an interactive session with permissive tool access")
	separator := lipgloss.NewStyle().Foreground(ui.DimPurple).Render(strings.Repeat("─", contentWidth))

	rows := make([]string, 0, len(p.runtimes)*2)
	for i, runtime := range p.runtimes {
		cursor := "  "
		nameStyle := ui.HelpDesc
		modeStyle := ui.HelpHint
		availability := ui.HelpKey.Render("ready")
		if !runtime.Installed() {
			availability = lipgloss.NewStyle().Foreground(ui.Muted).Render("not installed")
			nameStyle = lipgloss.NewStyle().Foreground(ui.Muted)
			modeStyle = lipgloss.NewStyle().Foreground(ui.Dim)
		}
		if i == p.cursor {
			cursor = ui.ItemCursor.Render(ui.Cursor + " ")
			if runtime.Installed() {
				nameStyle = lipgloss.NewStyle().Foreground(ui.White).Bold(true)
			}
		}

		name := nameStyle.Render(runtime.RuntimeLabel())
		gap := contentWidth - lipgloss.Width(cursor+name) - lipgloss.Width(availability)
		if gap < 1 {
			gap = 1
		}
		row := cursor + name + strings.Repeat(" ", gap) + availability
		if i == p.cursor {
			row = ui.ItemSelectedBg.Width(contentWidth).Render(row)
		}
		rows = append(rows, row)
		rows = append(rows, modeStyle.Render("    "+runtime.PermissionModeLabel()))
	}

	hint := ui.HelpHint.Width(contentWidth).Render("enter to launch  esc to cancel")
	note := ui.HelpHint.Width(contentWidth).Render("tmux opens a new window and leaves the issue note unsent")
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		separator,
		strings.Join(rows, "\n"),
		"",
		note,
		hint,
	)

	box := ui.HelpOverlayBg.Width(contentWidth + 4).Render(content)
	return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center, box)
}
