package components

import (
	"strings"
	"testing"
	"time"
)

func TestToastActiveBeforeExpiry(t *testing.T) {
	toast, _ := ShowToast("hello", ToastInfo, 5*time.Second)
	if !toast.Active() {
		t.Fatal("toast should be active before expiry")
	}
}

func TestToastInactiveAfterExpiry(t *testing.T) {
	toast := Toast{
		Message:   "expired",
		Level:     ToastInfo,
		ExpiresAt: time.Now().Add(-1 * time.Second),
	}
	if toast.Active() {
		t.Fatal("toast should be inactive after expiry")
	}
}

func TestToastInactiveWhenEmpty(t *testing.T) {
	toast := Toast{
		Message:   "",
		ExpiresAt: time.Now().Add(5 * time.Second),
	}
	if toast.Active() {
		t.Fatal("toast with empty message should not be active")
	}
}

func TestShowToastSetsExpiry(t *testing.T) {
	before := time.Now()
	toast, _ := ShowToast("test", ToastSuccess, 3*time.Second)
	after := time.Now()

	expectedMin := before.Add(3 * time.Second)
	expectedMax := after.Add(3 * time.Second)

	if toast.ExpiresAt.Before(expectedMin) || toast.ExpiresAt.After(expectedMax) {
		t.Fatalf("ExpiresAt %v not in expected range [%v, %v]", toast.ExpiresAt, expectedMin, expectedMax)
	}
}

func TestToastViewEmptyWhenInactive(t *testing.T) {
	toast := Toast{
		Message:   "",
		ExpiresAt: time.Now().Add(-1 * time.Second),
	}
	if got := toast.View(80); got != "" {
		t.Fatalf("inactive toast View() = %q, want empty string", got)
	}
}

func TestToastViewIncludesMessageForEachLevel(t *testing.T) {
	tests := []struct {
		name  string
		level ToastLevel
	}{
		{"info", ToastInfo},
		{"success", ToastSuccess},
		{"warn", ToastWarn},
		{"error", ToastError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			toast := Toast{Message: "hello " + tc.name, Level: tc.level}
			got := toast.View(80)
			if !strings.Contains(got, "hello "+tc.name) {
				t.Errorf("View() = %q, want to contain message", got)
			}
			if !strings.Contains(got, "\x1b[") {
				t.Errorf("View() missing ANSI styling for level %v: %q", tc.level, got)
			}
		})
	}
}

func TestToastViewDistinctStylePerLevel(t *testing.T) {
	// Each level applies a distinct background; the rendered string must
	// differ between Info/Success/Warn/Error for the same message+width.
	msg := "alert"
	info := Toast{Message: msg, Level: ToastInfo}.View(40)
	success := Toast{Message: msg, Level: ToastSuccess}.View(40)
	warn := Toast{Message: msg, Level: ToastWarn}.View(40)
	errToast := Toast{Message: msg, Level: ToastError}.View(40)

	pairs := [][2]string{{info, success}, {info, warn}, {info, errToast}, {success, errToast}, {warn, errToast}}
	for i, p := range pairs {
		if p[0] == p[1] {
			t.Errorf("pair %d: levels rendered identically (style switch broken)", i)
		}
	}
}

func TestToastViewZeroWidthDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Toast.View(0) panicked: %v", r)
		}
	}()
	_ = Toast{Message: "x", Level: ToastInfo}.View(0)
}
