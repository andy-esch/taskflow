package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// staleDateCells colours per-item ages through the shared theme.Staleness thresholds;
// relDateCells (the rollup spelling) stays neutral. Asserted as a DIFFERENCE between
// a fresh and an old cell rather than against a hardcoded escape, so the test survives
// a palette change but still fails if the colouring is dropped.
func TestStaleDateCellsColourByAge(t *testing.T) {
	fresh, old := time.Now().Format(time.DateOnly), "2020-01-01"
	cells := staleDateCells([]string{fresh, old}, func(s string) string { return s }, &testStyles)
	// Compare the STYLING, not the strings: the two cells carry different text either
	// way, so a naive inequality would pass even with the colouring removed.
	escapes := func(cell string) string { return strings.ReplaceAll(cell, ansi.Strip(cell), "") }
	if escapes(cells[0]) == escapes(cells[1]) {
		t.Errorf("a fresh and a years-old age should not share a colour: %q vs %q", cells[0], cells[1])
	}
	if ansi.StringWidth(cells[0]) != ansi.StringWidth(cells[1]) {
		t.Errorf("colouring must not disturb the padded column width: %d vs %d",
			ansi.StringWidth(cells[0]), ansi.StringWidth(cells[1]))
	}
	plain := relDateCells([]string{fresh, old}, func(s string) string { return s }, &testStyles)
	if plain[1] == cells[1] {
		t.Error("relDateCells is the rollup spelling and should not carry the staleness colour")
	}
}

// An undated item is neutral, never alarming, and still pads so the next column holds.
func TestStaleDateCellsUndatedIsNeutral(t *testing.T) {
	cells := staleDateCells([]string{"2020-01-01", ""}, func(s string) string { return s }, &testStyles)
	if ansi.StringWidth(cells[0]) != ansi.StringWidth(cells[1]) {
		t.Errorf("a blank cell must still pad to the column width: %d vs %d",
			ansi.StringWidth(cells[0]), ansi.StringWidth(cells[1]))
	}
	if strings.TrimSpace(ansi.Strip(cells[1])) != "" {
		t.Errorf("an undated cell should render empty, got %q", ansi.Strip(cells[1]))
	}
}
