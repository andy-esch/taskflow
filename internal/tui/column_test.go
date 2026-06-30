package tui

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestRelDateCells(t *testing.T) {
	id := func(s string) string { return s }
	// All undated → width 0, every cell empty so the caller drops the column.
	for _, c := range relDateCells([]string{"", ""}, id, testStyles) {
		if c != "" {
			t.Errorf("all-undated column should be empty cells, got %q", c)
		}
	}
	// Mixed: every cell padded to one width (a blank cell still pads, holding alignment).
	cells := relDateCells([]string{"2026-06-01", ""}, id, testStyles)
	w0, w1 := ansi.StringWidth(cells[0]), ansi.StringWidth(cells[1])
	if w0 == 0 || w0 != w1 {
		t.Errorf("a dated column's cells should share one width, got %d and %d", w0, w1)
	}
}

func TestCountsWidth(t *testing.T) {
	got := countsWidth([]int{0, 1}, func(i int) (int, int) {
		if i == 0 {
			return 7, 100 // "7/100" — the widest label
		}
		return 1, 2
	})
	if want := len(rollupCounts(7, 100, 0)); got != want {
		t.Errorf("countsWidth = %d, want %d (width of the widest done/total label)", got, want)
	}
}
