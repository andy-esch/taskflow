package tui

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/theme"
)

// Shared column-alignment helpers for the dashboard widgets and the list loaders.
// Both used to hand-roll the same measure-then-pad logic per widget (a dateW/countsW
// pre-measure plus %-*s), the open-coded counterpart of the CLI's writeTable — exactly
// the kind of thing that drifts. These describe the alignment once (audit M1).

// relDateCells renders the aligned relative-date column for a set of items: each cell
// is the item's relative date (theme.RelativeDate of the raw date), left-justified to
// the column's widest cell and dimmed, so whatever follows lines up. When NO item is
// dated the column is width 0 and every cell is "", so the caller drops it (a blank
// cell among dated ones still pads, holding the next column). The dashboard's
// in-progress + epics widgets share this rather than each pre-measuring a dateW.
func relDateCells[T any](items []T, raw func(T) string) []string {
	cells := make([]string, len(items))
	w := 0
	for i, it := range items {
		cells[i] = theme.RelativeDate(raw(it))
		w = max(w, len(cells[i])) // RelativeDate output is ASCII, so len == display width
	}
	if w == 0 {
		return cells // all undated → empty cells; the caller omits the column
	}
	for i := range cells {
		cells[i] = dim(fmt.Sprintf("%-*s", w, cells[i]))
	}
	return cells
}

// countsWidth is the display width of the widest done/total rollup label across items
// (rollupCounts at its natural width), so a column of them can be right-justified to
// one width via rollupCounts(done, total, countsWidth(...)). Shared by the dashboard
// epics widget and the epic/audit list loaders.
func countsWidth[T any](items []T, counts func(T) (done, total int)) int {
	w := 0
	for _, it := range items {
		done, total := counts(it)
		w = max(w, len(rollupCounts(done, total, 0)))
	}
	return w
}
