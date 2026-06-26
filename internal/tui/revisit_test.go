package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/testutil"
)

// revisitRepo seeds a tree with an active task plus three deferred tasks: one whose
// revisit date is clearly past (due), one clearly future (not due), and one with no
// date (indefinite). Past/future keep the due/not-due split wall-clock-stable.
func revisitRepo(t *testing.T) string {
	t.Helper()
	r := testutil.NewRepo(t)
	r.Epic("01-test.md", "---\nstatus: active\ndescription: a test epic\npriority: high\n---\n# Test epic\n")
	r.Task("in-progress", "alpha.md", "---\nstatus: in-progress\nepic: 01-test\ndescription: the alpha task\n---\n# alpha\n")
	deferred := func(slug, revisit string) {
		fm := fmt.Sprintf("---\nstatus: deferred\nepic: 01-test\ndescription: %s\ntags: [t]\n", slug)
		if revisit != "" {
			fm += fmt.Sprintf("revisit_at: %q\n", revisit)
		}
		fm += "---\n# " + slug + "\n"
		r.Task("deferred", slug+".md", fm)
	}
	deferred("overdue", "2020-01-01") // due
	deferred("later", "2099-01-01")   // not due
	deferred("parked", "")            // indefinite, not due
	return r.Root
}

// jumpToView drives the `:` command bar to switch to the named view and settles
// the load, exercising the real resolveView → applyView path.
func jumpToView(t *testing.T, m Model, word string) Model {
	t.Helper()
	tm, _ := m.Update(press(":"))
	m = tm.(Model)
	for _, r := range word {
		tm, _ = m.Update(press(string(r)))
		m = tm.(Model)
	}
	tm, cmd := m.Update(press("enter"))
	return drain(t, tm.(Model), cmd)
}

// taskSlugs returns the ids of the active tab's loaded task rows, in order.
func taskSlugs(m Model) []string {
	items := m.cur().list.Items()
	out := make([]string, 0, len(items))
	for _, it := range items {
		if ti, ok := it.(taskItem); ok {
			out = append(out, ti.id())
		}
	}
	return out
}

// TestModel_RevisitView pins the TUI mirror of `task list --revisit-due`: the
// `:revisit` view lists ONLY deferred tasks whose snooze date has arrived,
// excluding future/no-date deferred tasks and active tasks.
func TestModel_RevisitView(t *testing.T) {
	m := loadedAt(t, revisitRepo(t), 120, 40)
	m = jumpToView(t, m, "revisit")
	got := taskSlugs(m)
	if len(got) != 1 || got[0] != "overdue" {
		t.Errorf(":revisit should list only the due-deferred task, got %v", got)
	}
}

// TestModel_RevisitMarkerAndSortInDeferred pins that browsing `:deferred` floats
// the due-for-revisit task to the top and flags it with the ↻ marker (no emoji),
// while not-due deferred tasks stay unmarked below.
func TestModel_RevisitMarkerAndSortInDeferred(t *testing.T) {
	m := loadedAt(t, revisitRepo(t), 120, 40)
	m = jumpToView(t, m, "deferred")

	got := taskSlugs(m)
	if len(got) != 3 || got[0] != "overdue" {
		t.Errorf("due-for-revisit task should sort to the top of :deferred, got %v", got)
	}
	v := ansi.Strip(m.View().Content)
	if !strings.Contains(v, "↻") {
		t.Errorf("expected the ↻ revisit marker in :deferred view:\n%s", v)
	}
}
