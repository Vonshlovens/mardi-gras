package ui

import "testing"

func deref(t *testing.T, p *string, field string) string {
	t.Helper()
	if p == nil {
		t.Fatalf("%s color is nil; expected a brand hex", field)
	}
	return *p
}

// TestMardiGrasGlamourStyle verifies the brand theme maps mg's palette onto the
// markdown elements it recolors, and that it doesn't disturb the cloned dark
// base (a value copy, so the package var stays untouched).
func TestMardiGrasGlamourStyle(t *testing.T) {
	c := MardiGrasGlamourStyle()

	cases := []struct {
		field string
		got   *string
		want  string
	}{
		{"H1.Color", c.H1.Color, "#FFD700"},
		{"H1.BackgroundColor", c.H1.BackgroundColor, "#7B2D8E"},
		{"Heading.Color", c.Heading.Color, "#9B59B6"},
		{"H6.Color", c.H6.Color, "#9B59B6"},
		{"Emph.Color", c.Emph.Color, "#F5C518"},
		{"Strong.Color", c.Strong.Color, "#FFD700"},
		{"Link.Color", c.Link.Color, "#F5C518"},
		{"LinkText.Color", c.LinkText.Color, "#F5C518"},
		{"Code.Color", c.Code.Color, "#2ECC71"},
		{"Item.Color", c.Item.Color, "#9B59B6"},
		{"Enumeration.Color", c.Enumeration.Color, "#9B59B6"},
	}
	for _, tc := range cases {
		if got := deref(t, tc.got, tc.field); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.field, got, tc.want)
		}
	}

	if c.H1.Bold == nil || !*c.H1.Bold {
		t.Error("H1 should be bold")
	}
}

// TestMardiGrasGlamourStyleDoesNotMutateBase guards against accidentally
// mutating the shared DarkStyleConfig through a copied pointer: two calls must
// return identical, independent colors.
func TestMardiGrasGlamourStyleDoesNotMutateBase(t *testing.T) {
	a := MardiGrasGlamourStyle()
	b := MardiGrasGlamourStyle()
	if a.H1.Color == b.H1.Color {
		t.Error("each call should allocate fresh color pointers, not share them")
	}
	if *a.H1.Color != *b.H1.Color {
		t.Errorf("repeated calls disagree: %q vs %q", *a.H1.Color, *b.H1.Color)
	}
}

func TestMardiGrasGlamourStyleFollowsActiveTheme(t *testing.T) {
	original := CurrentThemeIndex()
	t.Cleanup(func() { SetThemeIndex(original) })

	if _, ok := SetTheme("Dawn"); !ok {
		t.Fatal("expected Dawn theme")
	}
	c := MardiGrasGlamourStyle()
	if got := deref(t, c.H1.Color, "Dawn H1.Color"); got != "#A35D3A" {
		t.Errorf("Dawn H1.Color = %q, want #A35D3A", got)
	}
	if got := deref(t, c.Heading.Color, "Dawn Heading.Color"); got != "#7A4A8A" {
		t.Errorf("Dawn Heading.Color = %q, want #7A4A8A", got)
	}
	if got := deref(t, c.Code.Color, "Dawn Code.Color"); got != "#5A7A3A" {
		t.Errorf("Dawn Code.Color = %q, want #5A7A3A", got)
	}

	SetTheme("Terminal")
	terminal := MardiGrasGlamourStyle()
	if terminal.H1.Color != nil || terminal.H1.BackgroundColor != nil {
		t.Errorf("Terminal glamour style should not inject a fixed palette: %+v", terminal.H1.StylePrimitive)
	}
}
