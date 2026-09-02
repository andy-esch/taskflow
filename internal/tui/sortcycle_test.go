package tui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// sortRepo seeds tasks whose LOADER order deliberately contradicts every sortable column,
// which is what makes a stuck sort visible. The shared seedRepo can't show it: its
// alpha/beta agree on loader order AND slug order, so a cycle that wrongly kept the slug
// sort still looked correct.
//
//	loader (working set): zulu (in-progress) → alpha (ready-to-start)
//	slug:                 alpha → zulu
//	priority:             alpha (high) → zulu (low)
func sortRepo(t *testing.T) string {
	t.Helper()
	r := testutil.NewRepo(t)
	task := func(status, slug, priority string) {
		r.Task(status, slug+".md", fmt.Sprintf(
			"---\nstatus: %s\nepic: 01-test\ndescription: the %s task\npriority: %s\ntier: 2\neffort: 1h\ncreated: 2026-01-01\ntags: [x]\n---\n# %s\n",
			status, slug, priority, slug))
	}
	task("in-progress", "zulu", "low")
	task("ready-to-start", "alpha", "high")
	r.Epic("01-test.md", "---\nstatus: active\ndescription: a test epic\npriority: high\n---\n# Test epic\n")
	return r.Root
}

func sortModel(t *testing.T) Model {
	t.Helper()
	m := New(core.NewService(store.NewFS(sortRepo(t))))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	return toTasks(t, tm.(Model))
}

// order returns the visible row ids top-to-bottom.
func order(m Model) []string {
	items := m.cur().list.Items()
	out := make([]string, 0, len(items))
	for _, it := range items {
		if ei, ok := it.(entityItem); ok {
			out = append(out, ei.ref().label)
		}
	}
	return out
}

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Cycling `o` all the way round must land back on the LOADER's order. It didn't:
// applySortToCurrent sorted the list's live slice and sortItems reorders IN PLACE, so
// sorts compounded and sortDefault — which by definition doesn't reorder — had nothing
// left to restore. Whichever column ran last silently stuck.
func TestSortCycle_ReturnsToLoaderOrder(t *testing.T) {
	m := sortModel(t)
	loader := order(m)
	if !eq(loader, []string{"zulu", "alpha"}) {
		t.Fatalf("loader order should be the working set (zulu first), got %v", loader)
	}

	n := len(m.cur().sortCols)
	for i := 0; i < n; i++ {
		tm, cmd := m.Update(press("o"))
		m = drain(t, tm.(Model), cmd)
	}
	if m.cur().sortKey != sortDefault {
		t.Fatalf("after %d presses sortKey should be sortDefault, got %v", n, m.cur().sortKey)
	}
	if got := order(m); !eq(got, loader) {
		t.Errorf("sortDefault should restore the loader's order %v, got %v", loader, got)
	}
}

// Sorts must not COMPOUND either: each one is computed from the loader's order, so
// sorting by slug and then by priority gives a pure priority order rather than
// "priority with the slug sort underneath".
func TestSortCycle_SortsDoNotCompound(t *testing.T) {
	m := sortModel(t)

	// Reach slug, then keep cycling to priority, checking each lands on a pure order.
	seen := map[sortKey][]string{}
	for i := 0; i < len(m.cur().sortCols)*2; i++ {
		tm, cmd := m.Update(press("o"))
		m = drain(t, tm.(Model), cmd)
		k := m.cur().sortKey
		got := order(m)
		if prev, ok := seen[k]; ok && !eq(prev, got) {
			t.Errorf("sort %v is not idempotent across cycles: first %v, later %v", k, prev, got)
		}
		seen[k] = got
	}
	// And the pure orders are what we expect for this fixture.
	if got := seen[sortSlug]; got != nil && !eq(got, []string{"alpha", "zulu"}) {
		t.Errorf("slug sort = %v, want [alpha zulu]", got)
	}
	if got := seen[sortPriority]; got != nil && !eq(got, []string{"alpha", "zulu"}) {
		t.Errorf("priority sort = %v, want [alpha zulu] (high first)", got)
	}
}

// A reload must refresh the sort BASE, not just re-apply the current sort — otherwise the
// recorded loader order goes stale and sortDefault restores a list that no longer matches
// disk.
func TestSortCycle_ReloadRefreshesTheSortBase(t *testing.T) {
	root := sortRepo(t)
	m := New(core.NewService(store.NewFS(root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = toTasks(t, tm.(Model))

	tm, cmd := m.Update(press("o")) // move off default
	m = drain(t, tm.(Model), cmd)
	if m.cur().sortKey == sortDefault {
		t.Fatal("expected to have moved off sortDefault")
	}

	// A third task lands from outside the TUI.
	path, out := testutil.TaskFixture(root, "in-progress", "mid.md",
		"---\nstatus: in-progress\nepic: 01-test\ndescription: the mid task\npriority: medium\ntier: 2\neffort: 1h\ncreated: 2026-01-01\ntags: [x]\n---\n# mid\n")
	testutil.Write(t, path, out)

	// reloadAll returns a BATCH (one reload per loaded tab plus the dashboard), so this
	// needs drainBatch rather than drain.
	tm, cmd = m.Update(reloadMsg{})
	m = drainBatch(t, tm.(Model), cmd)
	if got := len(m.cur().list.Items()); got != 3 {
		t.Fatalf("reload should see 3 tasks, got %d", got)
	}
	for m.cur().sortKey != sortDefault {
		tm, cmd = m.Update(press("o"))
		m = drain(t, tm.(Model), cmd)
	}
	if got := len(order(m)); got != 3 {
		t.Errorf("sortDefault restored a stale base: %d rows, want 3", got)
	}
}
