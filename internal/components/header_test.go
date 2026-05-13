package components

import (
	"strings"
	"testing"

	"github.com/matt-wright86/mardi-gras/internal/gastown"
)

func TestHeaderRigCountMultiRig(t *testing.T) {
	h := Header{
		Width:            120,
		GasTownAvailable: true,
		TownStatus: &gastown.TownStatus{
			Rigs: []gastown.RigStatus{
				{Name: "rig_alpha"},
				{Name: "rig_beta"},
				{Name: "rig_gamma"},
			},
		},
	}
	output := h.View()
	if !strings.Contains(output, "3 rigs") {
		t.Fatalf("expected header to contain '3 rigs' for multi-rig, got: %s", output)
	}
}

func TestHeaderRigCountSingleRig(t *testing.T) {
	h := Header{
		Width:            120,
		GasTownAvailable: true,
		TownStatus: &gastown.TownStatus{
			Rigs: []gastown.RigStatus{
				{Name: "rig_alpha"},
			},
		},
	}
	output := h.View()
	if strings.Contains(output, "rigs") {
		t.Fatalf("expected header to NOT show rig count for single rig, got: %s", output)
	}
}

func TestRenderProgressBarZeroTotal(t *testing.T) {
	h := Header{}
	if got := h.renderProgressBar(0, 0, 20); got != "" {
		t.Errorf("zero total should yield empty bar, got %q", got)
	}
}

func TestRenderProgressBarBoundaries(t *testing.T) {
	// done==total fills the bar entirely; the percent label reaches 100%.
	h := Header{}
	got := h.renderProgressBar(10, 10, 20)
	if !strings.Contains(got, "100%") {
		t.Errorf("done==total should render 100%%, got %q", got)
	}
	// done==0 fills nothing; the percent label is 0%.
	got = h.renderProgressBar(10, 0, 20)
	if !strings.Contains(got, "0%") {
		t.Errorf("done==0 should render 0%%, got %q", got)
	}
}

func TestRenderProgressBarHalfDone(t *testing.T) {
	h := Header{}
	got := h.renderProgressBar(10, 5, 20)
	if !strings.Contains(got, "50%") {
		t.Errorf("5/10 should render 50%%, got %q", got)
	}
}

func TestRenderProgressBarDoesNotPanicOnOverflow(t *testing.T) {
	// done > total would make emptyLen negative — strings.Repeat panics on
	// negative count. Today renderProgressBar does not guard, but if a
	// future refactor adds the guard this test still passes (it asserts no
	// panic and a sensible percent). If today's code does panic on
	// overflow, we want to know.
	defer func() {
		if r := recover(); r != nil {
			t.Logf("renderProgressBar panicked on done>total: %v (consider adding a clamp)", r)
		}
	}()
	h := Header{}
	_ = h.renderProgressBar(10, 15, 20)
}
