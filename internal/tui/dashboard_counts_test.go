package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/testutil"
)

// countsRepo has tasks across both sides of the split, and deliberately leaves
// ready-to-start and deprecated empty so the "omit empty buckets" rule is exercised
// rather than assumed.
func countsRepo(t *testing.T) string {
	t.Helper()
	r := testutil.NewRepo(t)
	r.Epic("01-e.md", "---\nstatus: active\ndescription: e\n---\n# e\n")
	task := func(name, status string) {
		r.Task(status, name,
			"---\nstatus: "+status+"\nepic: 01-e\ndescription: d\nupdated_at: 2026-09-10\n---\n# t\n")
	}
	task("a.md", "next-up")
	task("b.md", "next-up")
	task("c.md", "in-progress")
	task("d.md", "completed")
	task("e.md", "deferred")
	return r.Root
}

// The Overview leads with the shape of the whole task set, the way `status` does.
// Before this it opened on "in progress" and showed no next-up count anywhere.
func TestModel_DashboardShowsTaskCounts(t *testing.T) {
	v := ansi.Strip(loadedDashAt(t, countsRepo(t), 120, 40).View().Content)
	for _, want := range []string{"tasks", "active", "archived", "2 ● next-up", "1 ● in-progress", "1 ✔ completed", "1 ◌ deferred"} {
		if !strings.Contains(v, want) {
			t.Errorf("Overview should show %q:\n%s", want, v)
		}
	}
}

// An empty bucket is omitted, not rendered as a zero — Counts carries every status,
// so without the filter the line would read "0 ready-to-start".
func TestModel_DashboardCountsOmitEmptyBuckets(t *testing.T) {
	v := ansi.Strip(loadedDashAt(t, countsRepo(t), 120, 40).View().Content)
	for _, unwanted := range []string{"0 ready-to-start", "0 deprecated", "ready-to-start"} {
		if strings.Contains(v, unwanted) {
			t.Errorf("no task is ready-to-start, so %q should not appear:\n%s", unwanted, v)
		}
	}
}

// The counts orient; they are not destinations. A selectable row that opens nothing
// would be a dead end for the cursor, so they must carry no target.
func TestModel_DashboardCountRowsAreNotNavigable(t *testing.T) {
	m := loadedDashAt(t, countsRepo(t), 120, 40)
	for _, i := range m.dash.nav {
		if text := ansi.Strip(m.dash.rows[i].text); strings.Contains(text, "next-up") && strings.Contains(text, "·") {
			t.Errorf("the counts line is selectable but opens nothing: %q", text)
		}
	}
	if len(m.dash.nav) == 0 {
		t.Fatal("the dashboard should still have navigable rows (in-progress, epics)")
	}
}

// A repo with no tasks at all renders no counts block rather than a bare heading
// over two empty lines.
func TestModel_DashboardCountsAbsentWhenNoTasks(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Epic("01-e.md", "---\nstatus: active\ndescription: e\n---\n# e\n")
	if v := ansi.Strip(loadedDashAt(t, r.Root, 120, 40).View().Content); strings.Contains(v, "archived") {
		t.Errorf("an empty repo should show no counts block:\n%s", v)
	}
}
