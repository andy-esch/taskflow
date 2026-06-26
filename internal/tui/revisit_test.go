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
	// Archived states — hidden from the default working view.
	r.Task("completed", "done.md", "---\nstatus: completed\nepic: 01-test\ndescription: done\n---\n# done\n")
	r.Task("deprecated", "dead.md", "---\nstatus: deprecated\nepic: 01-test\ndescription: dead\n---\n# dead\n")
	return r.Root
}

func hasSlug(slugs []string, want string) bool {
	for _, s := range slugs {
		if s == want {
			return true
		}
	}
	return false
}

// openHelp opens the `?` overlay on a copy of m and returns its rendered (ANSI-
// stripped) content — for asserting page-specific help.
func openHelp(t *testing.T, m Model) string {
	t.Helper()
	tm, _ := m.Update(press("?"))
	return ansi.Strip(tm.(Model).View().Content)
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

// TestModel_WorkingViewShowsDeferredHidesArchived pins the default ("working")
// view: active work plus deferred (snoozed reminders) show, completed/deprecated
// are hidden, and due-for-revisit deferred floats to just under the active work.
func TestModel_WorkingViewShowsDeferredHidesArchived(t *testing.T) {
	m := loadedAt(t, revisitRepo(t), 120, 40) // opens on the default working view
	got := taskSlugs(m)

	for _, want := range []string{"alpha", "overdue", "later", "parked"} {
		if !hasSlug(got, want) {
			t.Errorf("working view should include %q, got %v", want, got)
		}
	}
	for _, hidden := range []string{"done", "dead"} {
		if hasSlug(got, hidden) {
			t.Errorf("working view should hide archived %q, got %v", hidden, got)
		}
	}
	if len(got) < 2 || got[0] != "alpha" || got[1] != "overdue" {
		t.Errorf("active work should lead, then due-for-revisit deferred; got %v", got)
	}
}

// TestModel_HelpNotesEntityScoped pins that the `?` Notes are page-specific: the
// tasks tab shows task views, the audits tab shows bucket views, and neither
// advertises the other's.
func TestModel_HelpNotesEntityScoped(t *testing.T) {
	m := loaded(t, 120, 40) // tasks tab
	help := openHelp(t, m)
	if !strings.Contains(help, ":revisit") {
		t.Errorf("tasks help should list task views (:revisit):\n%s", help)
	}
	if strings.Contains(help, "switch bucket") {
		t.Errorf("tasks help must not show audit bucket views:\n%s", help)
	}

	m = auditsTab(t, m)
	help = openHelp(t, m)
	if !strings.Contains(help, "switch bucket") {
		t.Errorf("audits help should show bucket views:\n%s", help)
	}
	if strings.Contains(help, ":revisit") {
		t.Errorf("audits help must not show task views (:revisit):\n%s", help)
	}
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
