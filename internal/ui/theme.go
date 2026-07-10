// Package ui defines the Mardi Gras design system: color palette, role and
// state colors, lipgloss styles, unicode symbols, gradients, and sparkline
// renderers. It contains no business logic.
package ui

import (
	"errors"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
)

// Theme is the colour contract used by Tuxedo. The five built-in values below
// are kept byte-for-byte equivalent to webstonehq/tuxedo's theme.rs palette.
// Mardi Gras maps its richer status and Gas Town vocabulary onto these values.
type Theme struct {
	Name string

	BG        color.Color
	Panel     color.Color
	Border    color.Color
	FG        color.Color
	Dim       color.Color
	Accent    color.Color
	Cursor    color.Color
	Selection color.Color

	StatusBar color.Color
	StatusFG  color.Color
	ModeFG    color.Color
	ModeBG    color.Color

	PriA     color.Color
	PriB     color.Color
	PriC     color.Color
	PriD     color.Color
	PriOther color.Color

	Project  color.Color
	Context  color.Color
	Due      color.Color
	Overdue  color.Color
	Today    color.Color
	Done     color.Color
	Selected color.Color
	Matched  color.Color

	// Terminal avoids true-colour backgrounds and uses the terminal's ANSI
	// palette, matching Tuxedo's Terminal theme.
	Terminal bool
}

func themeColor(hex string) color.Color { return lipgloss.Color(hex) }

// BuiltInThemes mirrors Tuxedo's canonical built-in theme order. Do not
// change their values without also changing the upstream source of truth.
var BuiltInThemes = []Theme{
	{
		Name: "Muted Slate",
		BG:   themeColor("#1A1D23"), Panel: themeColor("#1F232B"), Border: themeColor("#2A2F38"),
		FG: themeColor("#C8CCD4"), Dim: themeColor("#6B7280"), Accent: themeColor("#8AA9C9"),
		Cursor: themeColor("#3A4150"), Selection: themeColor("#2F3947"),
		StatusBar: themeColor("#252A33"), StatusFG: themeColor("#A8B0BC"),
		ModeFG: themeColor("#1A1D23"), ModeBG: themeColor("#8AA9C9"),
		PriA: themeColor("#E07A7A"), PriB: themeColor("#D4B06A"), PriC: themeColor("#7AA67A"),
		PriD: themeColor("#7A9EC9"), PriOther: themeColor("#9A8FC4"),
		Project: themeColor("#7FB3A8"), Context: themeColor("#C89A6E"), Due: themeColor("#D4B06A"),
		Overdue: themeColor("#E07A7A"), Today: themeColor("#E07A7A"), Done: themeColor("#5A6270"),
		Selected: themeColor("#2F3947"), Matched: themeColor("#D4B06A"),
	},
	{
		Name: "Dawn",
		BG:   themeColor("#FAF6F0"), Panel: themeColor("#F3EDE2"), Border: themeColor("#E0D6C4"),
		FG: themeColor("#3D3528"), Dim: themeColor("#8A7E6A"), Accent: themeColor("#A35D3A"),
		Cursor: themeColor("#E8DEC8"), Selection: themeColor("#EDE0C8"),
		StatusBar: themeColor("#EDE2CC"), StatusFG: themeColor("#5A4F3D"),
		ModeFG: themeColor("#FAF6F0"), ModeBG: themeColor("#A35D3A"),
		PriA: themeColor("#B8483A"), PriB: themeColor("#A3722A"), PriC: themeColor("#5A7A3A"),
		PriD: themeColor("#3A6A8A"), PriOther: themeColor("#7A4A8A"),
		Project: themeColor("#3A7A6A"), Context: themeColor("#A35D3A"), Due: themeColor("#A3722A"),
		Overdue: themeColor("#B8483A"), Today: themeColor("#B8483A"), Done: themeColor("#A89A82"),
		Selected: themeColor("#EDE0C8"), Matched: themeColor("#A3722A"),
	},
	{
		Name: "Nord",
		BG:   themeColor("#2E3440"), Panel: themeColor("#3B4252"), Border: themeColor("#434C5E"),
		FG: themeColor("#D8DEE9"), Dim: themeColor("#6C7686"), Accent: themeColor("#88C0D0"),
		Cursor: themeColor("#434C5E"), Selection: themeColor("#434C5E"),
		StatusBar: themeColor("#3B4252"), StatusFG: themeColor("#D8DEE9"),
		ModeFG: themeColor("#2E3440"), ModeBG: themeColor("#88C0D0"),
		PriA: themeColor("#BF616A"), PriB: themeColor("#EBCB8B"), PriC: themeColor("#A3BE8C"),
		PriD: themeColor("#81A1C1"), PriOther: themeColor("#B48EAD"),
		Project: themeColor("#A3BE8C"), Context: themeColor("#D08770"), Due: themeColor("#EBCB8B"),
		Overdue: themeColor("#BF616A"), Today: themeColor("#BF616A"), Done: themeColor("#4C566A"),
		Selected: themeColor("#434C5E"), Matched: themeColor("#EBCB8B"),
	},
	{
		Name: "Matrix",
		BG:   themeColor("#0A120A"), Panel: themeColor("#0F1A0F"), Border: themeColor("#1A2A1A"),
		FG: themeColor("#7FCC7F"), Dim: themeColor("#3F6A3F"), Accent: themeColor("#9FFF9F"),
		Cursor: themeColor("#1A2E1A"), Selection: themeColor("#1F3A1F"),
		StatusBar: themeColor("#0F1A0F"), StatusFG: themeColor("#7FCC7F"),
		ModeFG: themeColor("#0A120A"), ModeBG: themeColor("#9FFF9F"),
		PriA: themeColor("#FF8C8C"), PriB: themeColor("#FFD66E"), PriC: themeColor("#9FFF9F"),
		PriD: themeColor("#7FD0FF"), PriOther: themeColor("#CF9FFF"),
		Project: themeColor("#9FFF9F"), Context: themeColor("#FFB56E"), Due: themeColor("#FFD66E"),
		Overdue: themeColor("#FF8C8C"), Today: themeColor("#FF8C8C"), Done: themeColor("#3F6A3F"),
		Selected: themeColor("#1F3A1F"), Matched: themeColor("#FFD66E"),
	},
	{
		Name: "Terminal",
		BG:   lipgloss.NoColor{}, Panel: lipgloss.NoColor{}, Border: lipgloss.Color("8"),
		FG: themeColor("#DCDCDC"), Dim: lipgloss.Color("8"), Accent: lipgloss.Color("6"),
		Cursor: lipgloss.Color("8"), Selection: lipgloss.Color("8"),
		StatusBar: lipgloss.NoColor{}, StatusFG: lipgloss.NoColor{},
		ModeFG: lipgloss.Color("0"), ModeBG: lipgloss.Color("6"),
		PriA: lipgloss.Color("1"), PriB: lipgloss.Color("3"), PriC: lipgloss.Color("2"),
		PriD: lipgloss.Color("4"), PriOther: lipgloss.Color("5"),
		Project: lipgloss.Color("2"), Context: lipgloss.Color("3"), Due: lipgloss.Color("3"),
		Overdue: lipgloss.Color("1"), Today: lipgloss.Color("1"), Done: lipgloss.Color("8"),
		Selected: lipgloss.Color("8"), Matched: lipgloss.Color("3"), Terminal: true,
	},
}

// MardiGrasTheme is the original palette retained as an opt-in sixth theme.
// The Tuxedo palettes remain the first five entries and their values are not
// altered by this compatibility option.
var MardiGrasTheme = Theme{
	Name: "Mardi Gras",
	BG:   themeColor("#1A1A1A"), Panel: themeColor("#121521"), Border: themeColor("#4A1259"),
	FG: themeColor("#FAFAFA"), Dim: themeColor("#888888"), Accent: themeColor("#FFD700"),
	Cursor: themeColor("#4A1259"), Selection: themeColor("#4A1259"),
	StatusBar: themeColor("#4A1259"), StatusFG: themeColor("#CCCCCC"),
	ModeFG: themeColor("#1A1A1A"), ModeBG: themeColor("#FFD700"),
	PriA: themeColor("#FF3333"), PriB: themeColor("#FF8C00"), PriC: themeColor("#FFD700"),
	PriD: themeColor("#2ECC71"), PriOther: themeColor("#888888"),
	Project: themeColor("#2ECC71"), Context: themeColor("#9B59B6"), Due: themeColor("#FF8C00"),
	Overdue: themeColor("#E74C3C"), Today: themeColor("#E74C3C"), Done: themeColor("#888888"),
	Selected: themeColor("#4A1259"), Matched: themeColor("#FFD700"),
}

var themes = append(append([]Theme{}, BuiltInThemes...), MardiGrasTheme)

// Keep the existing Mardi Gras presentation on first run. Selecting a Tuxedo
// palette is persisted by the picker and restored during mg startup.
var activeThemeIndex = len(themes) - 1

// Core palette names are retained for the existing view code. They are updated
// together by SetThemeIndex, followed by style and gradient regeneration.
var (
	Background color.Color = MardiGrasTheme.BG
	Panel      color.Color = MardiGrasTheme.Panel
	Border     color.Color = MardiGrasTheme.Border

	Purple color.Color = themeColor("#7B2D8E")
	Gold   color.Color = themeColor("#F5C518")
	Green  color.Color = themeColor("#1D8348")

	BrightPurple color.Color = themeColor("#9B59B6")
	BrightGold   color.Color = themeColor("#FFD700")
	BrightGreen  color.Color = themeColor("#2ECC71")

	DimPurple color.Color = themeColor("#4A1259")
	DimGold   color.Color = themeColor("#8B7D00")
	DimGreen  color.Color = themeColor("#145A32")

	White   color.Color = themeColor("#FAFAFA")
	Light   color.Color = themeColor("#CCCCCC")
	Muted   color.Color = themeColor("#888888")
	Dim     color.Color = themeColor("#555555")
	Dark    color.Color = themeColor("#333333")
	Darkest color.Color = themeColor("#1A1A1A")

	StatusRolling color.Color = BrightGreen
	StatusLinedUp color.Color = BrightGold
	StatusStalled color.Color = themeColor("#E74C3C")
	StatusPassed  color.Color = Muted
	StatusAgent   color.Color = BrightPurple
	StatusConvoy  color.Color = BrightGold
	StatusMail    color.Color = BrightGreen

	PrioP0 color.Color = themeColor("#FF3333")
	PrioP1 color.Color = themeColor("#FF8C00")
	PrioP2 color.Color = BrightGold
	PrioP3 color.Color = BrightGreen
	PrioP4 color.Color = Muted

	ColorBug       color.Color = themeColor("#E74C3C")
	ColorFeature   color.Color = BrightPurple
	ColorTask      color.Color = BrightGold
	ColorChore     color.Color = Muted
	ColorEpic      color.Color = themeColor("#3498DB")
	ColorSpike     color.Color = themeColor("#F39C12")
	ColorStory     color.Color = themeColor("#A569BD")
	ColorMilestone color.Color = themeColor("#00BCD4")

	Silver color.Color = themeColor("#AAAAAA")

	RoleMayor    color.Color = BrightGold
	RoleDeacon   color.Color = themeColor("#3498DB")
	RolePolecat  color.Color = BrightGreen
	RoleCrew     color.Color = BrightPurple
	RoleWitness  color.Color = themeColor("#E67E22")
	RoleRefinery color.Color = themeColor("#1ABC9C")
	RoleDog      color.Color = themeColor("#8E44AD")
	RoleDefault  color.Color = Silver

	StateWorking    color.Color = BrightGreen
	StateIdle       color.Color = Silver
	StateBackoff    color.Color = themeColor("#E74C3C")
	StateStuck      color.Color = themeColor("#FF8C00")
	StateSpawn      color.Color = themeColor("#3498DB")
	StateGate       color.Color = BrightGold
	StateFixNeeded  color.Color = themeColor("#E056A0")
	StatePropelled  color.Color = themeColor("#00CED1")
	StatePatrolling color.Color = themeColor("#5DADE2")
)

// Themes returns a copy so callers cannot mutate the registered palette.
func Themes() []Theme {
	result := make([]Theme, len(themes))
	copy(result, themes)
	return result
}

// CurrentTheme returns the active palette.
func CurrentTheme() Theme { return themes[activeThemeIndex] }

// CurrentThemeIndex returns the active palette's index in Themes.
func CurrentThemeIndex() int { return activeThemeIndex }

// SetThemeIndex applies a palette, wrapping arbitrary indexes into range.
func SetThemeIndex(index int) Theme {
	if len(themes) == 0 {
		return Theme{}
	}
	index %= len(themes)
	if index < 0 {
		index += len(themes)
	}
	activeThemeIndex = index
	applyTheme(themes[index])
	return themes[index]
}

// SetTheme applies a palette by its display name.
func SetTheme(name string) (Theme, bool) {
	for i, theme := range themes {
		if theme.Name == name {
			return SetThemeIndex(i), true
		}
	}
	return CurrentTheme(), false
}

func applyTheme(theme Theme) {
	Background = theme.BG
	Panel = theme.Panel
	Border = theme.Border

	if theme.Name == MardiGrasTheme.Name {
		applyMardiGrasPalette()
	} else {
		applyTuxedoPalette(theme)
	}

	rebuildStyles()
	rebuildGradients()
}

func applyTuxedoPalette(theme Theme) {
	Purple = theme.Accent
	Gold = theme.Matched
	Green = theme.Project
	BrightPurple = theme.PriOther
	BrightGold = theme.Accent
	BrightGreen = theme.PriC
	DimPurple = theme.Border
	DimGold = theme.Due
	DimGreen = theme.Done
	White = theme.FG
	Light = theme.StatusFG
	Muted = theme.Dim
	Dim = theme.Dim
	Dark = theme.Panel
	Darkest = theme.BG

	StatusRolling = theme.PriC
	StatusLinedUp = theme.Accent
	StatusStalled = theme.Overdue
	StatusPassed = theme.Done
	StatusAgent = theme.Accent
	StatusConvoy = theme.Project
	StatusMail = theme.PriC

	PrioP0 = theme.PriA
	PrioP1 = theme.PriB
	PrioP2 = theme.PriC
	PrioP3 = theme.PriD
	PrioP4 = theme.PriOther

	ColorBug = theme.Overdue
	ColorFeature = theme.Accent
	ColorTask = theme.PriB
	ColorChore = theme.Dim
	ColorEpic = theme.PriD
	ColorSpike = theme.Context
	ColorStory = theme.PriOther
	ColorMilestone = theme.Project

	Silver = theme.FG
	RoleMayor = theme.Accent
	RoleDeacon = theme.PriD
	RolePolecat = theme.PriC
	RoleCrew = theme.PriOther
	RoleWitness = theme.Context
	RoleRefinery = theme.Project
	RoleDog = theme.PriOther
	RoleDefault = theme.FG

	StateWorking = theme.PriC
	StateIdle = theme.FG
	StateBackoff = theme.Overdue
	StateStuck = theme.Due
	StateSpawn = theme.PriD
	StateGate = theme.Accent
	StateFixNeeded = theme.PriOther
	StatePropelled = theme.Context
	StatePatrolling = theme.Project
}

func applyMardiGrasPalette() {
	Purple = themeColor("#7B2D8E")
	Gold = themeColor("#F5C518")
	Green = themeColor("#1D8348")
	BrightPurple = themeColor("#9B59B6")
	BrightGold = themeColor("#FFD700")
	BrightGreen = themeColor("#2ECC71")
	DimPurple = themeColor("#4A1259")
	DimGold = themeColor("#8B7D00")
	DimGreen = themeColor("#145A32")
	White = themeColor("#FAFAFA")
	Light = themeColor("#CCCCCC")
	Muted = themeColor("#888888")
	Dim = themeColor("#555555")
	Dark = themeColor("#333333")
	Darkest = themeColor("#1A1A1A")

	StatusRolling = BrightGreen
	StatusLinedUp = BrightGold
	StatusStalled = themeColor("#E74C3C")
	StatusPassed = Muted
	StatusAgent = BrightPurple
	StatusConvoy = BrightGold
	StatusMail = BrightGreen

	PrioP0 = themeColor("#FF3333")
	PrioP1 = themeColor("#FF8C00")
	PrioP2 = BrightGold
	PrioP3 = BrightGreen
	PrioP4 = Muted

	ColorBug = themeColor("#E74C3C")
	ColorFeature = BrightPurple
	ColorTask = BrightGold
	ColorChore = Muted
	ColorEpic = themeColor("#3498DB")
	ColorSpike = themeColor("#F39C12")
	ColorStory = themeColor("#A569BD")
	ColorMilestone = themeColor("#00BCD4")

	Silver = themeColor("#AAAAAA")
	RoleMayor = BrightGold
	RoleDeacon = themeColor("#3498DB")
	RolePolecat = BrightGreen
	RoleCrew = BrightPurple
	RoleWitness = themeColor("#E67E22")
	RoleRefinery = themeColor("#1ABC9C")
	RoleDog = themeColor("#8E44AD")
	RoleDefault = Silver

	StateWorking = BrightGreen
	StateIdle = Silver
	StateBackoff = themeColor("#E74C3C")
	StateStuck = themeColor("#FF8C00")
	StateSpawn = themeColor("#3498DB")
	StateGate = BrightGold
	StateFixNeeded = themeColor("#E056A0")
	StatePropelled = themeColor("#00CED1")
	StatePatrolling = themeColor("#5DADE2")
}

// ThemeConfigPath returns the XDG-style location used to persist a selection.
func ThemeConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mardi-gras", "config.toml"), nil
}

// LoadThemePreference restores the saved theme. Missing config is a normal
// first-run condition. Unknown or malformed themes are ignored safely.
func LoadThemePreference() error {
	path, err := ThemeConfigPath()
	if err != nil {
		return err
	}
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(body), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "theme" {
			continue
		}
		name := strings.Trim(strings.TrimSpace(value), "\"")
		if _, found := SetTheme(name); found {
			return nil
		}
		return nil
	}
	return nil
}

// SaveThemePreference persists the active choice in the same simple theme =
// name format used by Tuxedo's config file.
func SaveThemePreference() error {
	path, err := ThemeConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body := "# mardi-gras config\n" + "theme = " + CurrentTheme().Name + "\n"
	return os.WriteFile(path, []byte(body), 0o644)
}

// PriorityColor returns the theme color for a priority level.
func PriorityColor(p int) color.Color {
	switch p {
	case 0:
		return PrioP0
	case 1:
		return PrioP1
	case 2:
		return PrioP2
	case 3:
		return PrioP3
	case 4:
		return PrioP4
	default:
		return Muted
	}
}

// RoleColor returns the theme color for a Gas Town agent role.
func RoleColor(role string) color.Color {
	switch role {
	case "mayor", "coordinator":
		return RoleMayor
	case "deacon", "health-check":
		return RoleDeacon
	case "polecat":
		return RolePolecat
	case "crew":
		return RoleCrew
	case "witness":
		return RoleWitness
	case "refinery":
		return RoleRefinery
	case "dog":
		return RoleDog
	default:
		return RoleDefault
	}
}

// AgentStateColor returns the theme color for a Gas Town agent state.
func AgentStateColor(state string) color.Color {
	switch state {
	case "working":
		return StateWorking
	case "spawning":
		return StateSpawn
	case "backoff", "degraded":
		return StateBackoff
	case "stuck":
		return StateStuck
	case "awaiting-gate":
		return StateGate
	case "fix_needed":
		return StateFixNeeded
	case "propelled":
		return StatePropelled
	case "patrolling":
		return StatePatrolling
	case "paused", "muted":
		return Dim
	default:
		return StateIdle
	}
}

// IssueTypeColor returns the theme color for an issue type.
func IssueTypeColor(t string) color.Color {
	switch t {
	case "bug":
		return ColorBug
	case "feature":
		return ColorFeature
	case "task":
		return ColorTask
	case "chore":
		return ColorChore
	case "epic":
		return ColorEpic
	case "spike":
		return ColorSpike
	case "story":
		return ColorStory
	case "milestone":
		return ColorMilestone
	default:
		return Muted
	}
}
