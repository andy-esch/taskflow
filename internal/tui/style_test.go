package tui

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestMiniBar_Width pins the TUI rollup bar's exact display width across pct/width
// (the colored bar emits per-cell ANSI; a width regression would misalign the
// epic-list rows and detail line, and the plain-glyph CLI test can't catch it).
func TestMiniBar_Width(t *testing.T) {
	for _, w := range []int{1, 4, 8, 12, 20} {
		for _, pct := range []int{-5, 0, 1, 33, 50, 99, 100, 150} {
			if got := ansi.StringWidth(testStyles.miniBar(pct, w)); got != w {
				t.Errorf("miniBar(%d, %d) display width = %d, want %d", pct, w, got, w)
			}
		}
	}
}
