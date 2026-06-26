package tui

import (
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TestSortWorkingView_UnknownStatusLast is the A4 guard the polish-batch acceptance
// requires: a foreign/legacy status (which the loader tolerates) must sort LAST in
// the working view, not float up among in-progress work via a rank-0 map miss.
func TestSortWorkingView_UnknownStatusLast(t *testing.T) {
	tasks := []domain.Task{
		{Slug: "weird", Status: domain.Status("legacy-word")},
		{Slug: "rts", Status: domain.StatusReadyToStart},
		{Slug: "ip", Status: domain.StatusInProgress},
	}
	sortWorkingView(tasks, time.Now()) // no deferred tasks here, so the clock is irrelevant to the rank check
	if tasks[0].Status != domain.StatusInProgress {
		t.Errorf("in-progress should lead the working set, got %q", tasks[0].Status)
	}
	if last := tasks[len(tasks)-1]; last.Slug != "weird" {
		t.Errorf("an unknown status should sort last, got %q (%s)", last.Slug, last.Status)
	}
}

// TestPerEntitySortColumns pins that each entity's `o`-cycle offers only columns it
// actually has — so cycling can't land on a no-op sort that shows a chip while
// nothing reorders (the per-entity-columns fix).
func TestPerEntitySortColumns(t *testing.T) {
	tabs := newEntityTabs()
	has := func(cols []sortKey, k sortKey) bool {
		for _, c := range cols {
			if c == k {
				return true
			}
		}
		return false
	}
	epics := tabs[indexOfKind(tabs, entityEpics)].sortCols
	if has(epics, sortTier) || has(epics, sortUpdated) {
		t.Errorf("epics have no tier/updated — must not offer them: %v", epics)
	}
	audits := tabs[indexOfKind(tabs, entityAudits)].sortCols
	if has(audits, sortTier) || has(audits, sortPriority) {
		t.Errorf("audits have no tier/priority — must not offer them: %v", audits)
	}
	tasks := tabs[indexOfKind(tabs, entityTasks)].sortCols
	for _, k := range []sortKey{sortDefault, sortPriority, sortUpdated, sortTier, sortSlug} {
		if !has(tasks, k) {
			t.Errorf("tasks should offer every column, missing %v", k)
		}
	}
}

// TestCycleSortStaysOnOfferedColumns drives the real `o` cycle on the epics tab
// and asserts it never lands on a column epics don't have.
func TestCycleSortStaysOnOfferedColumns(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, cmd := m.Update(press("]")) // → epics
	m = drain(t, tm.(Model), cmd)
	if m.cur().kind != entityEpics {
		t.Skip("epics tab not active")
	}
	for i := 0; i < 8; i++ { // more than the cycle length, to wrap
		tm, _ := m.Update(press("o"))
		m = tm.(Model)
		if k := m.cur().sortKey; k == sortTier || k == sortUpdated {
			t.Fatalf("o landed on %v, which epics don't offer", k)
		}
	}
}

func TestPadRight_DisplayCells(t *testing.T) {
	if got := padRight("ab", 5); got != "ab   " {
		t.Errorf("padRight(ascii) = %q, want %q", got, "ab   ")
	}
	// A wide rune is 2 display cells, so the result is still 5 cells wide (not
	// over-padded by byte count) — keeping the date column aligned.
	if w := lipgloss.Width(padRight("你", 5)); w != 5 {
		t.Errorf("padRight should reach 5 display cells, got width %d", w)
	}
	if got := padRight("abcdef", 3); got != "abcdef" {
		t.Errorf("padRight must not truncate an overlong value, got %q", got)
	}
}
