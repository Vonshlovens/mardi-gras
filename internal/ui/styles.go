package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Pre-built styles are regenerated whenever a theme is selected. Keeping this
// central prevents a palette switch from leaving old Mardi Gras accents in a
// footer, form, toast, or overlay.
var (
	HeaderStyle  lipgloss.Style
	HeaderCounts lipgloss.Style

	BeadStylePurple lipgloss.Style
	BeadStyleGold   lipgloss.Style
	BeadStyleGreen  lipgloss.Style

	SectionRolling lipgloss.Style
	SectionLinedUp lipgloss.Style
	SectionStalled lipgloss.Style
	SectionPassed  lipgloss.Style

	StatusRollingStr string
	StatusLinedUpStr string
	StatusStalledStr string
	StatusPassedStr  string

	ItemNormal     lipgloss.Style
	ItemSelected   lipgloss.Style
	ItemCursor     lipgloss.Style
	ItemSelectedBg lipgloss.Style

	DetailBorder  lipgloss.Style
	DetailTitle   lipgloss.Style
	DetailLabel   lipgloss.Style
	DetailValue   lipgloss.Style
	DetailSection lipgloss.Style

	BadgePriority lipgloss.Style
	BadgeP0       string
	BadgeP1       string
	BadgeP2       string
	BadgeP3       string
	BadgeP4       string
	BadgeType     lipgloss.Style

	FooterStyle lipgloss.Style
	FooterKey   lipgloss.Style
	FooterDesc  lipgloss.Style

	DepBlocked     lipgloss.Style
	DepBlocks      lipgloss.Style
	DepMissing     lipgloss.Style
	DepResolved    lipgloss.Style
	DepNonBlocking lipgloss.Style
	OverdueBadge   lipgloss.Style
	DueSoonBadge   lipgloss.Style
	DeferredStyle  lipgloss.Style
	DepRelated     lipgloss.Style
	DepDuplicates  lipgloss.Style
	DepSupersedes  lipgloss.Style
	AgentBadge     lipgloss.Style
	ConvoyBadge    lipgloss.Style
	GasTownTag     lipgloss.Style

	GasTownBorder        lipgloss.Style
	GasTownTitle         lipgloss.Style
	GasTownLabel         lipgloss.Style
	GasTownValue         lipgloss.Style
	GasTownAgentSelected lipgloss.Style
	GasTownHint          lipgloss.Style
	FooterSource         lipgloss.Style

	MolStepDone    lipgloss.Style
	MolStepActive  lipgloss.Style
	MolStepReady   lipgloss.Style
	MolStepBlocked lipgloss.Style
	MolTierLabel   lipgloss.Style
	MolDAGFlow     lipgloss.Style
	MolCritical    lipgloss.Style

	MetaFieldName    lipgloss.Style
	MetaFieldNameDim lipgloss.Style
	MetaFieldType    lipgloss.Style
	MetaFieldValue   lipgloss.Style
	MetaRequired     lipgloss.Style

	InputPrompt lipgloss.Style
	InputText   lipgloss.Style
	InputCursor lipgloss.Style

	HelpOverlayBg lipgloss.Style
	HelpTitle     lipgloss.Style
	HelpSubtitle  lipgloss.Style
	HelpSection   lipgloss.Style
	HelpKey       lipgloss.Style
	HelpDesc      lipgloss.Style
	HelpHint      lipgloss.Style

	ToastInfo    lipgloss.Style
	ToastSuccess lipgloss.Style
	ToastWarn    lipgloss.Style
	ToastError   lipgloss.Style

	matchStyle lipgloss.Style
)

func init() {
	rebuildStyles()
}

func rebuildStyles() {
	theme := CurrentTheme()

	HeaderStyle = lipgloss.NewStyle().Bold(true).Background(theme.StatusBar).Padding(0, 1)
	HeaderCounts = lipgloss.NewStyle().Foreground(theme.StatusFG)

	BeadStylePurple = lipgloss.NewStyle().Foreground(Purple)
	BeadStyleGold = lipgloss.NewStyle().Foreground(Gold)
	BeadStyleGreen = lipgloss.NewStyle().Foreground(Green)

	SectionRolling = lipgloss.NewStyle().Bold(true).Foreground(StatusRolling)
	SectionLinedUp = lipgloss.NewStyle().Bold(true).Foreground(StatusLinedUp)
	SectionStalled = lipgloss.NewStyle().Bold(true).Foreground(StatusStalled)
	SectionPassed = lipgloss.NewStyle().Bold(true).Foreground(StatusPassed)

	StatusRollingStr = lipgloss.NewStyle().Foreground(StatusRolling).Render(SymRolling)
	StatusLinedUpStr = lipgloss.NewStyle().Foreground(StatusLinedUp).Render(SymLinedUp)
	StatusStalledStr = lipgloss.NewStyle().Foreground(StatusStalled).Render(SymStalled)
	StatusPassedStr = lipgloss.NewStyle().Foreground(StatusPassed).Render(SymPassed)

	ItemNormal = lipgloss.NewStyle().PaddingLeft(3)
	ItemSelected = lipgloss.NewStyle().PaddingLeft(1).Bold(true).Foreground(White)
	ItemCursor = lipgloss.NewStyle().Foreground(BrightGold).Bold(true)
	ItemSelectedBg = lipgloss.NewStyle().Background(theme.Cursor)

	DetailBorder = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Border).
		PaddingLeft(1)
	DetailTitle = lipgloss.NewStyle().Bold(true).Foreground(White)
	DetailLabel = lipgloss.NewStyle().Foreground(Muted).Width(12)
	DetailValue = lipgloss.NewStyle().Foreground(Light)
	DetailSection = lipgloss.NewStyle().Bold(true).Foreground(BrightGold).MarginTop(1)

	BadgePriority = lipgloss.NewStyle().Bold(true)
	BadgeP0 = BadgePriority.Foreground(PrioP0).Render("P0")
	BadgeP1 = BadgePriority.Foreground(PrioP1).Render("P1")
	BadgeP2 = BadgePriority.Foreground(PrioP2).Render("P2")
	BadgeP3 = BadgePriority.Foreground(PrioP3).Render("P3")
	BadgeP4 = BadgePriority.Foreground(PrioP4).Render("P4")
	BadgeType = lipgloss.NewStyle().Italic(true)

	FooterStyle = lipgloss.NewStyle().Foreground(theme.StatusFG).Background(theme.StatusBar).Padding(0, 1)
	FooterKey = lipgloss.NewStyle().Bold(true).Foreground(BrightGold)
	FooterDesc = lipgloss.NewStyle().Foreground(theme.StatusFG)

	DepBlocked = lipgloss.NewStyle().Foreground(StatusStalled)
	DepBlocks = lipgloss.NewStyle().Foreground(StatusLinedUp)
	DepMissing = lipgloss.NewStyle().Foreground(StatusStalled).Bold(true)
	DepResolved = lipgloss.NewStyle().Foreground(StatusPassed)
	DepNonBlocking = lipgloss.NewStyle().Foreground(Muted)
	OverdueBadge = lipgloss.NewStyle().Foreground(StatusStalled).Bold(true)
	DueSoonBadge = lipgloss.NewStyle().Foreground(PrioP1)
	DeferredStyle = lipgloss.NewStyle().Foreground(Dim)
	DepRelated = lipgloss.NewStyle().Foreground(BrightPurple)
	DepDuplicates = lipgloss.NewStyle().Foreground(Muted).Italic(true)
	DepSupersedes = lipgloss.NewStyle().Foreground(BrightGold)
	AgentBadge = lipgloss.NewStyle().Foreground(StatusAgent).Bold(true)
	ConvoyBadge = lipgloss.NewStyle().Foreground(StatusConvoy).Bold(true)
	GasTownTag = lipgloss.NewStyle().Foreground(BrightPurple).Italic(true)

	GasTownBorder = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(BrightGold).
		PaddingLeft(1)
	GasTownTitle = lipgloss.NewStyle().Bold(true).Foreground(BrightGold).MarginTop(1)
	GasTownLabel = lipgloss.NewStyle().Foreground(Muted)
	GasTownValue = lipgloss.NewStyle().Foreground(Light)
	GasTownAgentSelected = lipgloss.NewStyle().Background(theme.Cursor)
	GasTownHint = lipgloss.NewStyle().Foreground(Dim).MarginTop(1)
	FooterSource = lipgloss.NewStyle().Foreground(Muted)

	MolStepDone = lipgloss.NewStyle().Foreground(BrightGreen)
	MolStepActive = lipgloss.NewStyle().Foreground(BrightGold).Bold(true)
	MolStepReady = lipgloss.NewStyle().Foreground(Light)
	MolStepBlocked = lipgloss.NewStyle().Foreground(StatusStalled)
	MolTierLabel = lipgloss.NewStyle().Foreground(Dim).Italic(true)
	MolDAGFlow = lipgloss.NewStyle().Foreground(Dim)
	MolCritical = lipgloss.NewStyle().Foreground(BrightGold).Bold(true)

	MetaFieldName = lipgloss.NewStyle().Foreground(Light)
	MetaFieldNameDim = lipgloss.NewStyle().Foreground(Muted)
	MetaFieldType = lipgloss.NewStyle().Foreground(Muted)
	MetaFieldValue = lipgloss.NewStyle().Foreground(BrightGreen)
	MetaRequired = lipgloss.NewStyle().Foreground(StatusStalled)

	InputPrompt = lipgloss.NewStyle().Foreground(BrightGold).Bold(true).PaddingLeft(1)
	InputText = lipgloss.NewStyle().Foreground(White)
	InputCursor = lipgloss.NewStyle().Foreground(Purple)

	HelpOverlayBg = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Background(theme.Panel).
		Padding(1, 2)
	HelpTitle = lipgloss.NewStyle().Bold(true).Foreground(BrightGold).Align(lipgloss.Center)
	HelpSubtitle = lipgloss.NewStyle().Foreground(Light).Align(lipgloss.Center)
	HelpSection = lipgloss.NewStyle().Bold(true).Foreground(BrightGreen).Underline(true)
	HelpKey = lipgloss.NewStyle().Bold(true).Foreground(Gold)
	HelpDesc = lipgloss.NewStyle().Foreground(Light)
	HelpHint = lipgloss.NewStyle().Foreground(Muted).Align(lipgloss.Center)

	ToastInfo = lipgloss.NewStyle().Foreground(theme.StatusFG).Background(theme.StatusBar).Padding(0, 1)
	ToastSuccess = lipgloss.NewStyle().Foreground(theme.ModeFG).Background(BrightGreen).Bold(true).Padding(0, 1)
	ToastWarn = lipgloss.NewStyle().Foreground(theme.ModeFG).Background(BrightGold).Bold(true).Padding(0, 1)
	ToastError = lipgloss.NewStyle().Foreground(White).Background(StatusStalled).Bold(true).Padding(0, 1)

	matchStyle = lipgloss.NewStyle().Foreground(BrightGold).Bold(true).Underline(true)
}

// RoleBadge returns a styled badge for a Gas Town role.
func RoleBadge(role string) string {
	return lipgloss.NewStyle().Foreground(RoleColor(role)).Bold(true).Render(role)
}

// StateBadge returns a styled badge for an agent state.
func StateBadge(state string) string {
	sym := SymIdle
	switch state {
	case "working":
		sym = SymWorking
	case "spawning":
		sym = SymSpawning
	case "backoff", "degraded":
		sym = SymBackoff
	case "stuck":
		sym = SymStuck
	case "awaiting-gate":
		sym = SymGate
	case "fix_needed":
		sym = SymFixNeeded
	case "patrolling":
		sym = SymPatrolling
	case "paused", "muted":
		sym = SymPaused
	}
	return lipgloss.NewStyle().Foreground(AgentStateColor(state)).Render(sym + " " + state)
}

// SectionDivider renders a btop-style section divider: ── ⚜ TITLE ──────────
// When focused, the fleur-de-lis and cursor glow bright gold.
func SectionDivider(title string, width int, focused bool) string {
	cursorPrefix := ""
	extraWidth := 0
	if focused {
		cursorPrefix = lipgloss.NewStyle().Bold(true).Foreground(BrightGold).Render(Cursor) + " "
		extraWidth = 2
	}

	usedWidth := extraWidth + 5 + len([]rune(title)) + 1
	trailWidth := max(width-usedWidth, 3)
	trail := strings.Repeat(BoxHorizontal, trailWidth)

	ruleStyle := lipgloss.NewStyle().Foreground(Dim)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(BrightGold)
	fleurColor := DimGold
	if focused {
		fleurColor = BrightGold
	}
	fleurStyle := lipgloss.NewStyle().Foreground(fleurColor)

	return "\n" + cursorPrefix +
		ruleStyle.Render(BoxHorizontal+BoxHorizontal+" ") +
		fleurStyle.Render(FleurDeLis) + " " +
		titleStyle.Render(title) + " " +
		ruleStyle.Render(trail)
}

// HighlightMatches renders a string with matched character positions highlighted.
func HighlightMatches(text string, indices []int, maxLen int) string {
	runes := []rune(text)
	if maxLen > 0 && len(runes) > maxLen {
		runes = runes[:maxLen]
	}

	matchSet := make(map[int]bool, len(indices))
	for _, idx := range indices {
		matchSet[idx] = true
	}

	var b strings.Builder
	for i, r := range runes {
		if matchSet[i] {
			b.WriteString(matchStyle.Render(string(r)))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
