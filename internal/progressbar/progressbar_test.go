package progressbar

import (
	"regexp"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestRender_Width pins the exact display width across pct/width — both surfaces
// (CLI status table, TUI rows) align against it.
func TestRender_Width(t *testing.T) {
	for _, w := range []int{1, 4, 8, 10, 20} {
		for _, pct := range []int{-5, 0, 1, 33, 50, 99, 100, 150} {
			if got := ansi.StringWidth(Render(pct, w)); got != w {
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
	for _, m := range sgrRe.FindAllString(Render(100, 10), -1) {
		distinct[m] = struct{}{}
	}
	if len(distinct) < 2 {
		t.Errorf("a full bar should be a gradient (≥2 colors), got %d", len(distinct))
	}
}
