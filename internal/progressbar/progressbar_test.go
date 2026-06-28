package progressbar

import (
	"regexp"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/design"
)

// tp is the default palette the bar renders against in tests.
var tp = design.Default().Dark

// TestRender_Width pins the exact display width across pct/width — both surfaces
// (CLI status table, TUI rows) align against it.
func TestRender_Width(t *testing.T) {
	for _, w := range []int{1, 4, 8, 10, 20} {
		for _, pct := range []int{-5, 0, 1, 33, 50, 99, 100, 150} {
			if got := ansi.StringWidth(Render(pct, w, tp)); got != w {
				t.Errorf("Render(%d, %d) width = %d, want %d", pct, w, got, w)
			}
		}
	}
}

var sgrRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// TestRender_Gradient: a full bar paints multiple distinct colors (the neon
// gradient), not a single solid fill.
func TestRender_Gradient(t *testing.T) {
	distinct := map[string]struct{}{}
	for _, m := range sgrRe.FindAllString(Render(100, 10, tp), -1) {
		distinct[m] = struct{}{}
	}
	if len(distinct) < 2 {
		t.Errorf("a full bar should be a gradient (≥2 colors), got %d", len(distinct))
	}
}

// TestRenderSegments_BandsAndWidth pins the stacked bar: distinct glyphs per band
// (so it survives a mono terminal), exact width, and an all-empty bar for no findings.
func TestRenderSegments_BandsAndWidth(t *testing.T) {
	// 5 done, 1 active, 2 dropped, 2 open over width 10 → an exact, fully-banded split.
	out := ansi.Strip(RenderSegments(Segments{Done: 5, Active: 1, Dropped: 2, Total: 10}, 10, tp))
	if out != "█████▓▒▒░░" {
		t.Errorf("segmented bands = %q, want █████▓▒▒░░", out)
	}
	// width holds across sizes (apportionment sums to exactly width).
	for _, w := range []int{1, 4, 8, 12, 20} {
		if got := ansi.StringWidth(RenderSegments(Segments{Done: 3, Active: 1, Dropped: 2, Total: 9}, w, tp)); got != w {
			t.Errorf("RenderSegments width(%d) = %d, want %d", w, got, w)
		}
	}
	// no findings → all empty track, no done/active/dropped bands.
	if got := ansi.Strip(RenderSegments(Segments{Total: 0}, 6, tp)); got != "░░░░░░" {
		t.Errorf("empty audit bar = %q, want ░░░░░░", got)
	}
}
