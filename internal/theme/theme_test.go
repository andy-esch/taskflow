package theme

import (
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// theme is the single source of truth both the CLI (→ANSI) and TUI (→lipgloss)
// render from, so its decision table is pinned here — a silent glyph/color/edge
// change would otherwise shift both surfaces with no test catching it.

// TestMarkdownStyleFor pins the body theme shared by `show` and the TUI: dracula
// on a dark terminal, light on a light one.
func TestMarkdownStyleFor(t *testing.T) {
	if got := MarkdownStyleFor(true); got != "dracula" {
		t.Errorf("dark background should be dracula, got %q", got)
	}
	if got := MarkdownStyleFor(false); got != "light" {
		t.Errorf("light background should be light, got %q", got)
	}
}

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
		color  Color
	}{
		{domain.AuditOpen, ColorYellow},
		{domain.AuditClosed, ColorGreen},
		{domain.AuditDeferred, ColorGray},
		{domain.AuditBucket("weird"), ColorNone},
	}
	for _, c := range cases {
		if got := Bucket(c.bucket); got != c.color {
			t.Errorf("Bucket(%q) = %d, want %d", c.bucket, got, c.color)
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

// TestBarFill pins the shared fill arithmetic (one source for CLI + TUI bars).
func TestBarFill(t *testing.T) {
	for _, tc := range []struct{ pct, width, want int }{
		{0, 10, 0}, {50, 10, 5}, {100, 10, 10}, {33, 12, 3},
		{150, 10, 10}, {-5, 10, 0}, // clamped
	} {
		if got := BarFill(tc.pct, tc.width); got != tc.want {
			t.Errorf("BarFill(%d, %d) = %d, want %d", tc.pct, tc.width, got, tc.want)
		}
	}
}
