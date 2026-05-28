package components

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/matt-wright86/mardi-gras/internal/ui"
)

// ApprovalDialogResult is sent when the approval dialog completes. Decision is a
// codex ReviewDecision value ("approved", "approved_for_session", "denied",
// "abort"). Cancelled is true when the user dismissed the dialog (esc/q); the app
// treats that as a denial.
type ApprovalDialogResult struct {
	Decision  string
	Cancelled bool
}

// approvalDecision is one selectable choice in the dialog.
type approvalDecision struct {
	Label string
	Value string
}

// approvalDecisions is the lean-minimum decision set offered for every exec/patch
// approval. Amendment variants (execpolicy/network) are intentionally omitted —
// their response shapes are unverified upstream.
var approvalDecisions = []approvalDecision{
	{"Approve once", "approved"},
	{"Approve for this session", "approved_for_session"},
	{"Deny", "denied"},
	{"Abort turn", "abort"},
}

// ApprovalDialog prompts the user to approve or deny a codex action (a shell
// command or a patch). It mirrors RecoveryDialog's Update/View shape and is
// decoupled from codexmcp — the app passes plain fields.
type ApprovalDialog struct {
	kind    string // "exec" | "patch"
	message string
	command []string
	cwd     string
	reason  string
	files   []string
	selIdx  int
	width   int
	height  int
}

// NewApprovalDialog builds an approval dialog. For exec approvals pass command +
// cwd; for patch approvals pass files. reason is optional for both.
func NewApprovalDialog(kind, message string, command []string, cwd, reason string, files []string, width, height int) ApprovalDialog {
	return ApprovalDialog{
		kind:    kind,
		message: message,
		command: command,
		cwd:     cwd,
		reason:  reason,
		files:   files,
		width:   width,
		height:  height,
	}
}

// Update handles key events for the approval dialog.
func (ad ApprovalDialog) Update(msg tea.Msg) (ApprovalDialog, tea.Cmd) {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return ad, nil
	}

	switch km.String() {
	case "esc", "q":
		return ad, func() tea.Msg {
			return ApprovalDialogResult{Cancelled: true}
		}

	case "j", "down":
		if ad.selIdx < len(approvalDecisions)-1 {
			ad.selIdx++
		}

	case "k", "up":
		if ad.selIdx > 0 {
			ad.selIdx--
		}

	case "enter":
		selected := approvalDecisions[ad.selIdx]
		return ad, func() tea.Msg {
			return ApprovalDialogResult{Decision: selected.Value}
		}
	}

	return ad, nil
}

// View renders the approval prompt.
func (ad ApprovalDialog) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.BrightGold)
	dimStyle := lipgloss.NewStyle().Foreground(ui.Dim)
	normalStyle := lipgloss.NewStyle().Foreground(ui.Light)
	selectedStyle := lipgloss.NewStyle().Foreground(ui.BrightGreen)

	var lines []string

	// Title
	heading := "CODEX WANTS TO RUN A COMMAND"
	if ad.kind == "patch" {
		heading = "CODEX WANTS TO APPLY A PATCH"
	}
	lines = append(lines, titleStyle.Render(fmt.Sprintf("  %s %s", ui.SymGate, heading)))
	lines = append(lines, "")

	// Body
	switch ad.kind {
	case "patch":
		lines = append(lines, normalStyle.Render(fmt.Sprintf("  %d file(s) changed:", len(ad.files))))
		for _, f := range ad.files {
			lines = append(lines, fmt.Sprintf("    %s", dimStyle.Render(f)))
		}
	default:
		lines = append(lines, normalStyle.Render(fmt.Sprintf("  %s", strings.Join(ad.command, " "))))
		if ad.cwd != "" {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("  cwd: %s", ad.cwd)))
		}
	}
	if ad.reason != "" {
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render(fmt.Sprintf("  reason: %s", ad.reason)))
	}
	lines = append(lines, "")

	// Decisions
	for i, d := range approvalDecisions {
		cursor := "    "
		labelStyle := normalStyle
		if i == ad.selIdx {
			cursor = selectedStyle.Render("  > ")
			labelStyle = selectedStyle
		}
		lines = append(lines, fmt.Sprintf("%s%s", cursor, labelStyle.Render(d.Label)))
	}
	lines = append(lines, "")
	lines = append(lines, dimStyle.Render("  ↑/↓ select   enter confirm   esc deny"))

	return strings.Join(lines, "\n")
}
