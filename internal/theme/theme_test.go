package theme

import (
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// theme is the single source of truth both the CLI (→ANSI) and TUI (→lipgloss)
// render from, so its decision table is pinned here — a silent glyph/color/edge
// change would otherwise shift both surfaces with no test catching it.

func TestStatus(t *testing.T) {
	cases := []struct {
		status domain.Status
		glyph  string
		color  Color
	}{
		{domain.StatusInProgress, "●", ColorYellow},
		{domain.StatusNextUp, "●", ColorBlue},
		{domain.StatusReadyToStart, "○", ColorCyan},
		{domain.StatusCompleted, "✔", ColorGreen},
		{domain.StatusDeprecated, "✘", ColorRed},
		{domain.StatusDeferred, "◌", ColorGray},
		{domain.Status("something-foreign"), "•", ColorGray}, // default arm
	}
	for _, c := range cases {
		got := Status(c.status)
		if got.Glyph != c.glyph || got.Color != c.color {
			t.Errorf("Status(%q) = {%q,%d}, want {%q,%d}", c.status, got.Glyph, got.Color, c.glyph, c.color)
		}
	}
}

func TestBucket(t *testing.T) {
	cases := []struct {
		bucket domain.AuditBucket
		glyph  string
		color  Color
	}{
		{domain.AuditOpen, "◆", ColorYellow},
		{domain.AuditClosed, "✔", ColorGreen},
		{domain.AuditDeferred, "◌", ColorGray},
		{domain.AuditBucket("weird"), "■", ColorNone}, // default arm
	}
	for _, c := range cases {
		got := Bucket(c.bucket)
		if got.Glyph != c.glyph || got.Color != c.color {
			t.Errorf("Bucket(%q) = {%q,%d}, want {%q,%d}", c.bucket, got.Glyph, got.Color, c.glyph, c.color)
		}
	}
}

func TestLiveness(t *testing.T) {
	cases := []struct {
		band  string
		glyph string
		color Color
	}{
		{"working", "●", ColorYellow},
		{"fresh", "✦", ColorBlue},
		{"dormant", "○", ColorGray},
		{"", "•", ColorGray},      // default arm (unknown)
		{"bogus", "•", ColorGray}, // default arm
	}
	for _, c := range cases {
		got := Liveness(c.band)
		if got.Glyph != c.glyph || got.Color != c.color {
			t.Errorf("Liveness(%q) = {%q,%d}, want {%q,%d}", c.band, got.Glyph, got.Color, c.glyph, c.color)
		}
	}
}

func TestFindingStatus(t *testing.T) {
	cases := []struct {
		status string
		glyph  string
		color  Color
	}{
		{"open", "○", ColorYellow},
		{"in-progress", "●", ColorYellow},
		{"fixed", "✔", ColorGreen},
		{"landed", "✔", ColorGreen},
		{"deferred", "◌", ColorGray},
		{"superseded", "◌", ColorGray},
		{"wontfix", "✘", ColorRed},
		{"FIXED", "✔", ColorGreen}, // case-insensitive
		{"", "•", ColorGray},       // default arm (missing status)
		{"bogus", "•", ColorGray},
	}
	for _, c := range cases {
		got := FindingStatus(c.status)
		if got.Glyph != c.glyph || got.Color != c.color {
			t.Errorf("FindingStatus(%q) = {%q,%d}, want {%q,%d}", c.status, got.Glyph, got.Color, c.glyph, c.color)
		}
	}
}

func TestPriority(t *testing.T) {
	cases := []struct {
		priority string
		color    Color
	}{
		{"high", ColorRed},
		{"medium", ColorYellow},
		{"low", ColorGray},
		{"", ColorNone},
		{"bogus", ColorNone},
	}
	for _, c := range cases {
		if got := Priority(c.priority); got != c.color {
			t.Errorf("Priority(%q) = %d, want %d", c.priority, got, c.color)
		}
	}
}

func TestPercent(t *testing.T) {
	// Pin the boundaries: <34 gray, 34..99 yellow, 100 green.
	cases := []struct {
		pct   int
		color Color
	}{
		{0, ColorGray}, {33, ColorGray},
		{34, ColorYellow}, {99, ColorYellow},
		{100, ColorGreen},
	}
	for _, c := range cases {
		if got := Percent(c.pct); got != c.color {
			t.Errorf("Percent(%d) = %d, want %d", c.pct, got, c.color)
		}
	}
}

func TestTaskDate(t *testing.T) {
	if got := TaskDate(domain.Task{Updated: "2026-06-10", Created: "2026-06-01"}); got != "2026-06-10" {
		t.Errorf("TaskDate prefers Updated, got %q", got)
	}
	if got := TaskDate(domain.Task{Created: "2026-06-01"}); got != "2026-06-01" {
		t.Errorf("TaskDate falls back to Created, got %q", got)
	}
}

// TestMarkers pins the cross-surface marker glyphs so a change is deliberate, and in
// particular that ✓ (ready-to-close) and ✔ (all-clear / done) stay DISTINCT glyphs —
// the reconciliation the legend review called for.
func TestMarkers(t *testing.T) {
	if MarkerReadyToClose.Glyph != "✓" || MarkerAllClear.Glyph != "✔" {
		t.Errorf("ready-to-close=%q all-clear=%q, want ✓ (U+2713) / ✔ (U+2714)", MarkerReadyToClose.Glyph, MarkerAllClear.Glyph)
	}
	if MarkerReadyToClose.Glyph == MarkerAllClear.Glyph {
		t.Error("✓ ready-to-close and ✔ all-clear must stay distinct glyphs")
	}
	if MarkerWarn.Glyph != "⚠" || MarkerRevisit.Glyph != "↻" || MarkerUnreadable.Glyph != "!" {
		t.Errorf("marker glyphs drifted: warn=%q revisit=%q unreadable=%q", MarkerWarn.Glyph, MarkerRevisit.Glyph, MarkerUnreadable.Glyph)
	}
}
