package gastown

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEventParsing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	content := `{"ts":"2026-02-23T01:00:51Z","source":"gt","type":"session_start","actor":"mayor","payload":{"role":"mayor"},"visibility":"feed"}
{"ts":"2026-02-23T01:02:37Z","source":"gt","type":"sling","actor":"mayor","payload":{"target":"mardi_gras/quartz","bead":"bd-c8q"},"visibility":"feed"}
{"ts":"2026-02-23T01:05:00Z","source":"gt","type":"nudge","actor":"mayor","payload":{"target":"mardi_gras/quartz","reason":"Run gt prime"},"visibility":"feed"}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	events, err := LoadRecentEvents(path, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	// Newest first
	if events[0].Type != "nudge" {
		t.Fatalf("expected newest event type 'nudge', got %q", events[0].Type)
	}
	if events[2].Type != "session_start" {
		t.Fatalf("expected oldest event type 'session_start', got %q", events[2].Type)
	}

	// Payload extraction
	target := EventPayloadString(events[0], "target")
	if target != "mardi_gras/quartz" {
		t.Fatalf("expected target 'mardi_gras/quartz', got %q", target)
	}
}

func TestEventParsingLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	content := `{"ts":"1","source":"gt","type":"a","actor":"x","visibility":"feed"}
{"ts":"2","source":"gt","type":"b","actor":"x","visibility":"feed"}
{"ts":"3","source":"gt","type":"c","actor":"x","visibility":"feed"}
{"ts":"4","source":"gt","type":"d","actor":"x","visibility":"feed"}
{"ts":"5","source":"gt","type":"e","actor":"x","visibility":"feed"}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	events, err := LoadRecentEvents(path, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	// Newest first: e, d, c
	if events[0].Type != "e" {
		t.Fatalf("expected 'e', got %q", events[0].Type)
	}
	if events[2].Type != "c" {
		t.Fatalf("expected 'c', got %q", events[2].Type)
	}
}

func TestEventParsingEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	events, err := LoadRecentEvents(path, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if events != nil {
		t.Fatalf("expected nil events for empty file, got %d", len(events))
	}
}

func TestEventParsingMissingFile(t *testing.T) {
	events, err := LoadRecentEvents("/nonexistent/path/events.jsonl", 20)
	if err != nil {
		t.Fatalf("missing file should return nil error, got %v", err)
	}
	if events != nil {
		t.Fatalf("expected nil events for missing file, got %d", len(events))
	}
}

func TestLoadRecentEventsTracksAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	initial := `{"ts":"1","source":"gt","type":"a","actor":"x","visibility":"feed"}
{"ts":"2","source":"gt","type":"b","actor":"x","visibility":"feed"}
`
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	events, err := LoadRecentEvents(path, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 || events[0].Type != "b" || events[1].Type != "a" {
		t.Fatalf("unexpected initial events: %+v", events)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"ts":"3","source":"gt","type":"c","actor":"x","visibility":"feed"}
`); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	events, err = LoadRecentEvents(path, 2)
	if err != nil {
		t.Fatalf("unexpected error after append: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events after append, got %d", len(events))
	}
	if events[0].Type != "c" || events[1].Type != "b" {
		t.Fatalf("expected newest events [c b], got [%s %s]", events[0].Type, events[1].Type)
	}
}

func TestEventsPath(t *testing.T) {
	// With GT_HOME set
	t.Setenv("GT_HOME", "/tmp/mygt")
	path := EventsPath()
	if path != "/tmp/mygt/.events.jsonl" {
		t.Fatalf("expected '/tmp/mygt/.events.jsonl', got %q", path)
	}

	// Without GT_HOME — should use ~/gt/
	t.Setenv("GT_HOME", "")
	path = EventsPath()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "gt", ".events.jsonl")
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}

func TestEventPayloadStringMissing(t *testing.T) {
	ev := Event{Payload: nil}
	if s := EventPayloadString(ev, "foo"); s != "" {
		t.Fatalf("expected empty string, got %q", s)
	}
}

func TestAgentEventCount(t *testing.T) {
	events := []Event{
		{Actor: "mayor", Type: "sling"},
		{Actor: "mayor", Type: "nudge"},
		{Actor: "polecat-1", Type: "session_start"},
		{Actor: "polecat-1", Type: "session_death"},
		{Actor: "polecat-1", Type: "handoff"},
	}
	got := AgentEventCount(events)
	if got["mayor"] != 2 {
		t.Errorf("mayor count = %d, want 2", got["mayor"])
	}
	if got["polecat-1"] != 3 {
		t.Errorf("polecat-1 count = %d, want 3", got["polecat-1"])
	}
}

func TestAgentEventCountEmpty(t *testing.T) {
	got := AgentEventCount(nil)
	if len(got) != 0 {
		t.Errorf("empty events should yield empty map, got %v", got)
	}
}

func TestAgentActivityHistogramEmpty(t *testing.T) {
	if got := AgentActivityHistogram(nil, 5, time.Hour); got != nil {
		t.Errorf("nil events should yield nil histogram, got %v", got)
	}
}

func TestAgentActivityHistogramZeroBuckets(t *testing.T) {
	events := []Event{{Actor: "x", Timestamp: time.Now().Format(time.RFC3339)}}
	if got := AgentActivityHistogram(events, 0, time.Hour); got != nil {
		t.Errorf("zero buckets should yield nil, got %v", got)
	}
}

func TestAgentActivityHistogramBucketsRecent(t *testing.T) {
	now := time.Now()
	window := 60 * time.Minute
	buckets := 6 // 10 minutes per bucket
	events := []Event{
		// 5 min ago — newest bucket (5)
		{Actor: "mayor", Timestamp: now.Add(-5 * time.Minute).Format(time.RFC3339)},
		// 55 min ago — oldest bucket (0)
		{Actor: "mayor", Timestamp: now.Add(-55 * time.Minute).Format(time.RFC3339)},
		// out-of-window: 2 hours ago — must be dropped
		{Actor: "mayor", Timestamp: now.Add(-2 * time.Hour).Format(time.RFC3339)},
		// invalid timestamp — must be dropped
		{Actor: "mayor", Timestamp: "not-a-timestamp"},
	}

	got := AgentActivityHistogram(events, buckets, window)
	mayor, ok := got["mayor"]
	if !ok {
		t.Fatalf("expected mayor histogram, got keys %v", got)
	}
	if len(mayor) != buckets {
		t.Fatalf("len(mayor) = %d, want %d", len(mayor), buckets)
	}
	if mayor[buckets-1] != 1 {
		t.Errorf("newest bucket = %d, want 1", mayor[buckets-1])
	}
	if mayor[0] != 1 {
		t.Errorf("oldest bucket = %d, want 1", mayor[0])
	}
	total := 0
	for _, n := range mayor {
		total += n
	}
	if total != 2 {
		t.Errorf("total counted = %d, want 2 (in-window valid events)", total)
	}
}

func TestAgentActivityHistogramMultipleAgents(t *testing.T) {
	now := time.Now()
	events := []Event{
		{Actor: "mayor", Timestamp: now.Add(-1 * time.Minute).Format(time.RFC3339)},
		{Actor: "polecat-1", Timestamp: now.Add(-1 * time.Minute).Format(time.RFC3339)},
		{Actor: "polecat-1", Timestamp: now.Add(-2 * time.Minute).Format(time.RFC3339)},
	}
	got := AgentActivityHistogram(events, 5, time.Hour)
	if len(got) != 2 {
		t.Fatalf("expected 2 agents, got %d (%v)", len(got), got)
	}
	if got["mayor"] == nil || got["polecat-1"] == nil {
		t.Errorf("expected both mayor and polecat-1 keys, got %v", got)
	}
}
