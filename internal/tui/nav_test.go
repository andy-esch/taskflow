package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/store"
)

// followToEpic presses f on the tasks tab and settles the epics load.
func followToEpic(t *testing.T, m Model) Model {
	t.Helper()
	tm, cmd := m.Update(press("f"))
	m = tm.(Model)
	if m.cur().name != "epics" {
		t.Fatalf("f on a task should jump to epics, got %q", m.cur().name)
	}
	return pump(t, m, cmd, 8)
}

func TestModel_FollowTaskToEpicAndBack(t *testing.T) {
	m := loaded(t, 120, 40) // alpha selected; seed tasks carry epic: 01-test
	m = followToEpic(t, m)
	if m.selectedID() != "01-test" {
		t.Fatalf("should land on the task's epic, got %q", m.selectedID())
	}
	if len(m.navStack) != 1 || m.navStack[0] != (navLoc{entityTasks, "alpha"}) {
		t.Fatalf("the origin should be on the back-stack, got %v", m.navStack)
	}
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "ctrl+o alpha") {
		t.Errorf("the footer should show the back breadcrumb:\n%s", v)
	}
	// ctrl+o returns to the task with the cursor restored, stack emptied.
	tm, cmd := m.Update(press("ctrl+o"))
	m = pump(t, tm.(Model), cmd, 8)
	if m.cur().name != "tasks" || m.selectedID() != "alpha" {
		t.Errorf("back should restore tasks/alpha, got %s/%s", m.cur().name, m.selectedID())
	}
	if len(m.navStack) != 0 {
		t.Errorf("the stack should be empty after back, got %v", m.navStack)
	}
	// An empty stack flashes instead of crashing.
	tm, _ = m.Update(press("ctrl+o"))
	m = tm.(Model)
	if !m.flashErr || !strings.Contains(m.flash, "nothing") {
		t.Errorf("back on an empty stack should flash, got %q", m.flash)
	}
}

func TestModel_FollowEpicToTaskViaMenuMultiHop(t *testing.T) {
	m := loaded(t, 120, 40)
	m = followToEpic(t, m) // hop 1: tasks/alpha → epics/01-test (detail pumped in)

	// f on the epic opens the reference picker over its tasks.
	tm, _ := m.Update(press("f"))
	m = tm.(Model)
	if !m.follow.active {
		t.Fatalf("f on an epic should open the follow picker (flash=%q)", m.flash)
	}
	if len(m.follow.tasks) != 2 {
		t.Fatalf("the picker should list the epic's 2 tasks, got %d", len(m.follow.tasks))
	}
	// The picker is modal: q must not quit, j moves the cursor.
	tm, cmd := m.Update(press("j"))
	m = tm.(Model)
	if cmd != nil {
		if _, quits := cmd().(tea.QuitMsg); quits {
			t.Fatal("picker keys must not quit the app")
		}
	}
	want := m.follow.selected().Slug
	tm, cmd = m.Update(press("enter")) // hop 2: epics/01-test → tasks/<want>
	m = pump(t, tm.(Model), cmd, 8)
	if m.cur().name != "tasks" || m.selectedID() != want {
		t.Fatalf("enter should jump to tasks/%s, got %s/%s", want, m.cur().name, m.selectedID())
	}
	if len(m.navStack) != 2 {
		t.Fatalf("two hops should stack two origins, got %v", m.navStack)
	}
	// Unwind both hops.
	tm, cmd = m.Update(press("ctrl+o"))
	m = pump(t, tm.(Model), cmd, 8)
	if m.cur().name != "epics" || m.selectedID() != "01-test" {
		t.Errorf("first back should restore epics/01-test, got %s/%s", m.cur().name, m.selectedID())
	}
	tm, cmd = m.Update(press("ctrl+o"))
	m = pump(t, tm.(Model), cmd, 8)
	if m.cur().name != "tasks" || m.selectedID() != "alpha" {
		t.Errorf("second back should restore tasks/alpha, got %s/%s", m.cur().name, m.selectedID())
	}
	// Esc closes the picker without following.
	m = followToEpic(t, m) // back on epics with detail loaded
	tm, _ = m.Update(press("f"))
	m = tm.(Model)
	tm, _ = m.Update(press("esc"))
	m = tm.(Model)
	if m.follow.active {
		t.Error("esc should close the picker")
	}
}

func TestModel_FollowGracefulDeadEnds(t *testing.T) {
	// A task with no epic reference.
	root := t.TempDir()
	dir := filepath.Join(root, "tasks", "in-progress")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "loner.md"),
		[]byte("---\nstatus: active\ndescription: x\n---\n# loner\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := New(core.NewService(store.NewFS(root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = tm.(Model)
	tm, _ = m.Update(press("f"))
	m = tm.(Model)
	if !m.flashErr || !strings.Contains(m.flash, "no epic") {
		t.Errorf("following without an epic should flash, got %q", m.flash)
	}
	if m.cur().name != "tasks" || len(m.navStack) != 0 {
		t.Error("a dead-end follow must not move or push the stack")
	}

	// Audits have no structured references.
	a := loaded(t, 120, 40)
	tm, _ = a.Update(press(":"))
	a = tm.(Model)
	for _, r := range "audits" {
		tm, _ = a.Update(press(string(r)))
		a = tm.(Model)
	}
	var cmd tea.Cmd
	tm, cmd = a.Update(press("enter"))
	a = pump(t, tm.(Model), cmd, 8)
	tm, _ = a.Update(press("f"))
	a = tm.(Model)
	if !a.flashErr || !strings.Contains(a.flash, "no linked") {
		t.Errorf("f on audits should flash, got %q", a.flash)
	}
}

// TestModel_FollowEscalatesToAllView pins the hidden-target path: following an
// epic reference to a COMPLETED task (not in the working set) widens the tasks
// tab to :all instead of failing.
func TestModel_FollowEscalatesToAllView(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("epics/01-x.md", "---\nstatus: active\ndescription: e\n---\n# E\n")
	write("tasks/in-progress/active-one.md", "---\nstatus: active\nepic: 01-x\ndescription: a\n---\n# a\n")
	write("tasks/completed/done-one.md", "---\nstatus: completed\nepic: 01-x\ndescription: d\n---\n# d\n")

	m := New(core.NewService(store.NewFS(root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = tm.(Model)
	if n := len(m.cur().list.Items()); n != 1 {
		t.Fatalf("working set should hide the completed task, got %d items", n)
	}
	m = followToEpic(t, m)
	tm, _ = m.Update(press("f"))
	m = tm.(Model)
	if !m.follow.active {
		t.Fatalf("picker should open (flash=%q)", m.flash)
	}
	// Select done-one (store scan order puts completed after in-progress).
	for m.follow.selected().Slug != "done-one" {
		tm, _ = m.Update(press("j"))
		m = tm.(Model)
	}
	tm, cmd := m.Update(press("enter"))
	m = pump(t, tm.(Model), cmd, 8)
	if m.cur().statusView != "all" {
		t.Errorf("a hidden target should escalate the view to :all, got %q", m.cur().statusView)
	}
	if m.selectedID() != "done-one" {
		t.Errorf("the completed task should be selected, got %q", m.selectedID())
	}
}

// TestModel_JumpClearsAppliedFilter pins that a follow/back jump is explicit
// navigation: an applied `/` filter on the target tab is cleared so it can't
// hide the destination.
func TestModel_JumpClearsAppliedFilter(t *testing.T) {
	m := loaded(t, 120, 40)
	// Filter the tasks tab down to beta, then follow alpha's epic? No — the
	// filter moves the selection; follow from the filtered selection (beta),
	// then come BACK into the filtered tab: the filter must be gone.
	m.cur().list.SetFilterText("beta")
	if m.selectedID() != "beta" {
		t.Fatalf("setup: filter should select beta, got %q", m.selectedID())
	}
	m = followToEpic(t, m)
	tm, cmd := m.Update(press("ctrl+o"))
	m = pump(t, tm.(Model), cmd, 8)
	if m.cur().name != "tasks" || m.selectedID() != "beta" {
		t.Fatalf("back should restore tasks/beta, got %s/%s", m.cur().name, m.selectedID())
	}
	if m.cur().list.FilterState() != list.Unfiltered {
		t.Errorf("a jump must clear the target tab's filter, state=%v", m.cur().list.FilterState())
	}
	if n := len(m.cur().list.VisibleItems()); n != 2 {
		t.Errorf("all rows should be visible after the jump, got %d", n)
	}
}

// TestModel_FollowDanglingEpicRef pins the real-world missing-target case: a
// task whose epic: field names an epic that doesn't exist (historical B1 data).
// The jump lands on the epics tab with a clear flash — and ctrl+o still works.
func TestModel_FollowDanglingEpicRef(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "tasks", "in-progress")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "orphan.md"),
		[]byte("---\nstatus: active\nepic: 99-ghost\ndescription: x\n---\n# orphan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := New(core.NewService(store.NewFS(root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = tm.(Model)
	tm, cmd := m.Update(press("f"))
	m = pump(t, tm.(Model), cmd, 8)
	if m.cur().name != "epics" {
		t.Fatalf("the jump should still switch tabs, got %q", m.cur().name)
	}
	if !m.flashErr || !strings.Contains(m.flash, "99-ghost") {
		t.Errorf("a dangling target should flash its id, got %q (err=%v)", m.flash, m.flashErr)
	}
	tm, cmd = m.Update(press("ctrl+o"))
	m = pump(t, tm.(Model), cmd, 8)
	if m.cur().name != "tasks" || m.selectedID() != "orphan" {
		t.Errorf("ctrl+o should still return home, got %s/%s", m.cur().name, m.selectedID())
	}
}

// TestModel_FollowPickerFitsTerminal extends the layout invariant to the `f`
// picker (mirrors TestModel_ActionMenuFitsTerminal): the overlay must never
// change the view's height or overflow its width.
func TestModel_FollowPickerFitsTerminal(t *testing.T) {
	for _, d := range []struct{ w, h int }{
		{120, 40}, {80, 20}, {40, 12}, {24, 8},
	} {
		m := loaded(t, d.w, d.h)
		m = followToEpic(t, m)
		tm, _ := m.Update(press("f"))
		m = tm.(Model)
		if !m.follow.active {
			t.Fatalf("%dx%d: picker should open", d.w, d.h)
		}
		lines := strings.Split(m.View().Content, "\n")
		if len(lines) != d.h {
			t.Errorf("%dx%d with picker: %d lines, want %d", d.w, d.h, len(lines), d.h)
		}
		for i, ln := range lines {
			if w := ansi.StringWidth(ln); w > d.w {
				t.Errorf("%dx%d with picker: line %d is %d wide > %d: %q", d.w, d.h, i, w, d.w, ansi.Strip(ln))
			}
		}
	}
}

// TestPushLoc_SkipsConsecutiveDuplicate pins L9 (2026-06-22 audit): re-pushing the
// current top adds no history, so the back-stack doesn't grow on repeated follows
// of the same place.
func TestPushLoc_SkipsConsecutiveDuplicate(t *testing.T) {
	m := loaded(t, 120, 40)
	if m.selectedID() == "" {
		t.Fatal("setup: expected a selection to push")
	}
	m.pushLoc()
	m.pushLoc()
	if len(m.navStack) != 1 {
		t.Errorf("consecutive identical pushes should dedup to 1, got %d", len(m.navStack))
	}
}
