package ui

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestPriorityColor(t *testing.T) {
	tests := []struct {
		name     string
		priority int
		want     color.Color
	}{
		{"P0", 0, PrioP0},
		{"P1", 1, PrioP1},
		{"P2", 2, PrioP2},
		{"P3", 3, PrioP3},
		{"P4", 4, PrioP4},
		{"negative falls back to Muted", -1, Muted},
		{"out of range falls back to Muted", 5, Muted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PriorityColor(tt.priority)
			if got != tt.want {
				t.Errorf("PriorityColor(%d) = %v, want %v", tt.priority, got, tt.want)
			}
		})
	}
}

func TestIssueTypeColor(t *testing.T) {
	tests := []struct {
		name      string
		issueType string
		want      color.Color
	}{
		{"bug", "bug", ColorBug},
		{"feature", "feature", ColorFeature},
		{"task", "task", ColorTask},
		{"chore", "chore", ColorChore},
		{"epic", "epic", ColorEpic},
		{"spike", "spike", ColorSpike},
		{"story", "story", ColorStory},
		{"milestone", "milestone", ColorMilestone},
		{"empty string falls back to Muted", "", Muted},
		{"unknown falls back to Muted", "unknown", Muted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IssueTypeColor(tt.issueType)
			if got != tt.want {
				t.Errorf("IssueTypeColor(%q) = %v, want %v", tt.issueType, got, tt.want)
			}
		})
	}
}

func TestAgentStateColor(t *testing.T) {
	tests := []struct {
		name  string
		state string
		want  color.Color
	}{
		{"working", "working", StateWorking},
		{"spawning", "spawning", StateSpawn},
		{"idle", "idle", StateIdle},
		{"backoff", "backoff", StateBackoff},
		{"degraded maps to backoff", "degraded", StateBackoff},
		{"stuck", "stuck", StateStuck},
		{"awaiting-gate", "awaiting-gate", StateGate},
		{"fix_needed", "fix_needed", StateFixNeeded},
		{"paused", "paused", Dim},
		{"muted", "muted", Dim},
		{"propelled", "propelled", StatePropelled},
		{"unknown falls back to idle", "unknown", StateIdle},
		{"empty falls back to idle", "", StateIdle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AgentStateColor(tt.state)
			if got != tt.want {
				t.Errorf("AgentStateColor(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestAgentStateColorDistinctCategories(t *testing.T) {
	// Each state category should map to a distinct color
	colors := map[string]color.Color{
		"working":       AgentStateColor("working"),
		"spawning":      AgentStateColor("spawning"),
		"idle":          AgentStateColor("idle"),
		"backoff":       AgentStateColor("backoff"),
		"stuck":         AgentStateColor("stuck"),
		"awaiting-gate": AgentStateColor("awaiting-gate"),
		"fix_needed":    AgentStateColor("fix_needed"),
		"paused":        AgentStateColor("paused"),
		"propelled":     AgentStateColor("propelled"),
	}

	// Verify distinct colors across categories (some share intentionally)
	pairs := [][2]string{
		{"working", "idle"},
		{"working", "backoff"},
		{"working", "stuck"},
		{"idle", "backoff"},
		{"idle", "stuck"},
		{"stuck", "backoff"},
		{"fix_needed", "stuck"},
		{"fix_needed", "backoff"},
		{"fix_needed", "working"},
		{"spawning", "idle"},
		{"propelled", "working"},
		{"propelled", "idle"},
		{"propelled", "spawning"},
	}
	for _, pair := range pairs {
		if colors[pair[0]] == colors[pair[1]] {
			t.Errorf("AgentStateColor(%q) == AgentStateColor(%q), should be distinct", pair[0], pair[1])
		}
	}
}

func TestApplyMardiGrasGradientEmpty(t *testing.T) {
	result := ApplyMardiGrasGradient("")
	if result != "" {
		t.Errorf("ApplyMardiGrasGradient(\"\") = %q, want \"\"", result)
	}
}

func TestApplyMardiGrasGradientNonEmpty(t *testing.T) {
	input := "hello"
	result := ApplyMardiGrasGradient(input)

	if result == "" {
		t.Fatal("ApplyMardiGrasGradient(\"hello\") returned empty string")
	}

	for _, r := range input {
		if !strings.Contains(result, string(r)) {
			t.Errorf("result missing character %q from input", string(r))
		}
	}
}

func TestApplyPartialGradientZeroLength(t *testing.T) {
	result := ApplyPartialMardiGrasGradient("hello", 0)
	if result != "" {
		t.Errorf("ApplyPartialMardiGrasGradient(\"hello\", 0) = %q, want \"\"", result)
	}
}

func TestApplyPartialGradientNonEmpty(t *testing.T) {
	result := ApplyPartialMardiGrasGradient("hello", 10)
	if result == "" {
		t.Fatal("ApplyPartialMardiGrasGradient(\"hello\", 10) returned empty string")
	}
}

func TestRoleColor(t *testing.T) {
	tests := []struct {
		name string
		role string
		want color.Color
	}{
		{"mayor", "mayor", RoleMayor},
		{"coordinator alias of mayor", "coordinator", RoleMayor},
		{"deacon", "deacon", RoleDeacon},
		{"health-check alias of deacon", "health-check", RoleDeacon},
		{"polecat", "polecat", RolePolecat},
		{"crew", "crew", RoleCrew},
		{"witness", "witness", RoleWitness},
		{"refinery", "refinery", RoleRefinery},
		{"dog", "dog", RoleDog},
		{"unknown falls back to default", "frobnicator", RoleDefault},
		{"empty falls back to default", "", RoleDefault},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RoleColor(tc.role)
			if got != tc.want {
				t.Errorf("RoleColor(%q) = %v, want %v", tc.role, got, tc.want)
			}
		})
	}
}

func TestAgentStateColorPatrolling(t *testing.T) {
	// Patrolling state was added 2026-05-13; verify it maps to the sky-blue
	// StatePatrolling color and not the default Idle fallback.
	if got := AgentStateColor("patrolling"); got != StatePatrolling {
		t.Errorf("AgentStateColor(\"patrolling\") = %v, want StatePatrolling", got)
	}
	if AgentStateColor("patrolling") == AgentStateColor("idle") {
		t.Error("patrolling and idle must map to distinct colors")
	}
}

func testThemeHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}

func themeSignature(theme Theme) string {
	colors := []color.Color{
		theme.BG, theme.Panel, theme.Border, theme.FG, theme.Dim, theme.Accent,
		theme.Cursor, theme.Selection, theme.StatusBar, theme.StatusFG, theme.ModeFG,
		theme.ModeBG, theme.PriA, theme.PriB, theme.PriC, theme.PriD, theme.PriOther,
		theme.Project, theme.Context, theme.Due, theme.Overdue, theme.Today, theme.Done,
		theme.Selected, theme.Matched,
	}
	parts := []string{theme.Name}
	for _, c := range colors {
		parts = append(parts, testThemeHex(c))
	}
	return strings.Join(parts, "|")
}

func TestTuxedoBuiltInThemesMatchCanonicalPalette(t *testing.T) {
	if len(BuiltInThemes) != 5 {
		t.Fatalf("Tuxedo built-ins = %d, want 5", len(BuiltInThemes))
	}

	// Field order is Theme's complete colour contract, excluding Terminal's
	// Reset/ANSI entries which are verified below. These snapshots come from
	// webstonehq/tuxedo src/theme.rs at upstream main.
	want := []string{
		"Muted Slate|#1A1D23|#1F232B|#2A2F38|#C8CCD4|#6B7280|#8AA9C9|#3A4150|#2F3947|#252A33|#A8B0BC|#1A1D23|#8AA9C9|#E07A7A|#D4B06A|#7AA67A|#7A9EC9|#9A8FC4|#7FB3A8|#C89A6E|#D4B06A|#E07A7A|#E07A7A|#5A6270|#2F3947|#D4B06A",
		"Dawn|#FAF6F0|#F3EDE2|#E0D6C4|#3D3528|#8A7E6A|#A35D3A|#E8DEC8|#EDE0C8|#EDE2CC|#5A4F3D|#FAF6F0|#A35D3A|#B8483A|#A3722A|#5A7A3A|#3A6A8A|#7A4A8A|#3A7A6A|#A35D3A|#A3722A|#B8483A|#B8483A|#A89A82|#EDE0C8|#A3722A",
		"Nord|#2E3440|#3B4252|#434C5E|#D8DEE9|#6C7686|#88C0D0|#434C5E|#434C5E|#3B4252|#D8DEE9|#2E3440|#88C0D0|#BF616A|#EBCB8B|#A3BE8C|#81A1C1|#B48EAD|#A3BE8C|#D08770|#EBCB8B|#BF616A|#BF616A|#4C566A|#434C5E|#EBCB8B",
		"Matrix|#0A120A|#0F1A0F|#1A2A1A|#7FCC7F|#3F6A3F|#9FFF9F|#1A2E1A|#1F3A1F|#0F1A0F|#7FCC7F|#0A120A|#9FFF9F|#FF8C8C|#FFD66E|#9FFF9F|#7FD0FF|#CF9FFF|#9FFF9F|#FFB56E|#FFD66E|#FF8C8C|#FF8C8C|#3F6A3F|#1F3A1F|#FFD66E",
	}

	for i, expected := range want {
		if got := themeSignature(BuiltInThemes[i]); got != expected {
			t.Errorf("theme %d differs from Tuxedo\n got: %s\nwant: %s", i, got, expected)
		}
	}

	terminal := BuiltInThemes[4]
	if !terminal.Terminal {
		t.Fatal("Terminal theme should preserve terminal palette mode")
	}
	for _, c := range []color.Color{terminal.BG, terminal.Panel, terminal.StatusBar, terminal.StatusFG} {
		if _, ok := c.(lipgloss.NoColor); !ok {
			t.Errorf("Terminal reset color = %T, want lipgloss.NoColor", c)
		}
	}
	if terminal.Accent != lipgloss.Color("6") || terminal.PriA != lipgloss.Color("1") || terminal.PriD != lipgloss.Color("4") {
		t.Errorf("Terminal ANSI palette changed: accent=%v priA=%v priD=%v", terminal.Accent, terminal.PriA, terminal.PriD)
	}
}

func TestSetThemeRegeneratesStylesAndGradients(t *testing.T) {
	original := CurrentThemeIndex()
	t.Cleanup(func() { SetThemeIndex(original) })

	if _, ok := SetTheme("Dawn"); !ok {
		t.Fatal("expected Dawn to be selectable")
	}
	if got := testThemeHex(Background); got != "#FAF6F0" {
		t.Errorf("Background = %s, want Dawn background", got)
	}
	if got := testThemeHex(HelpOverlayBg.GetBackground()); got != "#F3EDE2" {
		t.Errorf("help overlay background = %s, want Dawn panel", got)
	}
	if got := testThemeHex(GradientHeat.At(0).GetForeground()); got != "#5A7A3A" {
		t.Errorf("heat gradient start = %s, want Dawn green", got)
	}
	if got := testThemeHex(DetailBorder.GetBorderLeftForeground()); got != "#E0D6C4" {
		t.Errorf("detail border = %s, want Dawn border", got)
	}
}

func TestThemePreferenceRoundTrip(t *testing.T) {
	original := CurrentThemeIndex()
	t.Cleanup(func() { SetThemeIndex(original) })
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if _, ok := SetTheme("Nord"); !ok {
		t.Fatal("expected Nord to be selectable")
	}
	if err := SaveThemePreference(); err != nil {
		t.Fatalf("SaveThemePreference() error: %v", err)
	}
	path, err := ThemeConfigPath()
	if err != nil {
		t.Fatalf("ThemeConfigPath() error: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved preference: %v", err)
	}
	if !strings.Contains(string(body), "theme = Nord") {
		t.Fatalf("saved config missing selected theme: %q", body)
	}

	SetThemeIndex(len(themes) - 1)
	if err := LoadThemePreference(); err != nil {
		t.Fatalf("LoadThemePreference() error: %v", err)
	}
	if got := CurrentTheme().Name; got != "Nord" {
		t.Errorf("restored theme = %q, want Nord", got)
	}
}
