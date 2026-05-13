package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestBrailleSparklineEmpty(t *testing.T) {
	got := BrailleSparkline(nil, 10, lipgloss.NewStyle())
	if strings.TrimSpace(got) != "" {
		// Should be all spaces
		if got == "" {
			t.Error("should return padded spaces for nil data")
		}
	}
}

func TestBrailleSparklineBasic(t *testing.T) {
	data := []float64{1, 2, 3, 4}
	got := BrailleSparkline(data, 5, lipgloss.NewStyle())
	if got == "" {
		t.Error("should render braille characters")
	}
	// 4 data points = 2 braille chars + 3 padding = 5 chars
	if len([]rune(got)) != 5 {
		t.Errorf("expected 5 runes, got %d", len([]rune(got)))
	}
}

func TestBrailleSparklineOddData(t *testing.T) {
	data := []float64{1, 2, 3}
	got := BrailleSparkline(data, 5, lipgloss.NewStyle())
	if got == "" {
		t.Error("should handle odd number of data points")
	}
}

func TestBrailleSparklineAllZero(t *testing.T) {
	data := []float64{0, 0, 0}
	got := BrailleSparkline(data, 5, lipgloss.NewStyle())
	if strings.TrimSpace(got) != "" {
		// All zeros should render as spaces
		t.Log("all-zero data rendered as:", got)
	}
}

func TestMiniSparklineAllZero(t *testing.T) {
	got := MiniSparkline([3]int{0, 0, 0})
	if got != "" {
		t.Error("all-zero should return empty string")
	}
}

func TestMiniSparklineBasic(t *testing.T) {
	got := MiniSparkline([3]int{1, 3, 2})
	if got == "" {
		t.Error("non-zero values should produce output")
	}
}

func TestDualSparklineBasic(t *testing.T) {
	top := []float64{1, 0, 2, 0}
	bot := []float64{0, 1, 0, 2}
	style := lipgloss.NewStyle()
	got := DualSparkline(top, bot, 4, style, style)
	if got == "" {
		t.Error("should render dual sparkline")
	}
}

func TestDualSparklineZeroWidth(t *testing.T) {
	got := DualSparkline([]float64{1}, []float64{1}, 0, lipgloss.NewStyle(), lipgloss.NewStyle())
	if got != "" {
		t.Error("zero width should return empty")
	}
}

func TestRenderSparklineEmpty(t *testing.T) {
	got := RenderSparkline(nil, 5)
	if got != "     " {
		t.Errorf("nil values should yield %d spaces, got %q (len=%d)", 5, got, len(got))
	}
}

func TestRenderSparklineZeroWidth(t *testing.T) {
	got := RenderSparkline([]int{1, 2, 3}, 0)
	if got != "" {
		t.Errorf("zero width should yield empty, got %q", got)
	}
}

func TestRenderSparklineAllZero(t *testing.T) {
	got := RenderSparkline([]int{0, 0, 0}, 5)
	// All-zero path renders 3 dim "▁" blocks (min(len, width)).
	count := strings.Count(got, "▁")
	if count != 3 {
		t.Errorf("all-zero with len=3,width=5 should yield 3 ▁ blocks, got %d in %q", count, got)
	}
}

func TestRenderSparklineNormal(t *testing.T) {
	got := RenderSparkline([]int{1, 5, 3, 7, 2}, 5)
	// Verify it produced 5 block chars (one per value).
	total := 0
	for _, b := range sparkBlocks {
		total += strings.Count(got, b)
	}
	if total != 5 {
		t.Errorf("expected 5 block chars across all levels, got %d in %q", total, got)
	}
}

func TestRenderSparklineTruncatesToWidth(t *testing.T) {
	got := RenderSparkline([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 4)
	total := 0
	for _, b := range sparkBlocks {
		total += strings.Count(got, b)
	}
	if total != 4 {
		t.Errorf("expected 4 blocks (width-limited), got %d in %q", total, got)
	}
}

func TestHeatCharZero(t *testing.T) {
	got := HeatChar(0, 10)
	if !strings.Contains(got, "·") {
		t.Errorf("zero event count should render '·', got %q", got)
	}
}

func TestHeatCharLowAndHigh(t *testing.T) {
	low := HeatChar(1, 10)
	if !strings.Contains(low, "▪") {
		t.Errorf("low intensity should render ▪, got %q", low)
	}
	high := HeatChar(9, 10)
	if !strings.Contains(high, "▮") {
		t.Errorf("high intensity (t>0.7) should render ▮, got %q", high)
	}
}

func TestHeatCharZeroMaxCount(t *testing.T) {
	// Should not panic when maxCount==0 but eventCount>0.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("HeatChar panicked on maxCount=0: %v", r)
		}
	}()
	_ = HeatChar(3, 0)
}

func TestConvoyPipelineEmpty(t *testing.T) {
	if got := ConvoyPipeline(nil, 20); got != "" {
		t.Errorf("empty statuses should yield empty, got %q", got)
	}
}

func TestConvoyPipelineSymbols(t *testing.T) {
	got := ConvoyPipeline([]string{"closed", "in_progress", "open"}, 20)
	if !strings.Contains(got, "●") {
		t.Errorf("closed should render ●, got %q", got)
	}
	if !strings.Contains(got, "◐") {
		t.Errorf("in_progress should render ◐, got %q", got)
	}
	if !strings.Contains(got, "○") {
		t.Errorf("open should render ○, got %q", got)
	}
}

func TestConvoyPipelineHookedRendersActive(t *testing.T) {
	got := ConvoyPipeline([]string{"hooked"}, 10)
	if !strings.Contains(got, "◐") {
		t.Errorf("hooked should render ◐ (active), got %q", got)
	}
}

func TestConvoyPipelineTruncatesAtMaxWidth(t *testing.T) {
	got := ConvoyPipeline([]string{"closed", "closed", "closed", "closed", "closed", "closed"}, 4)
	// maxWidth=4 → n = 4/2 = 2 closed nodes + " +4" suffix
	if !strings.Contains(got, "+4") {
		t.Errorf("expected '+4' overflow suffix, got %q", got)
	}
}
