package tui

import (
	"os"
	"path/filepath"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
)

// TestEntityTab_MarkReloadCarriesPendingTarget pins the core of the M6 fix: a
// background reload firing while a jump's load is in flight must carry the jump
// target forward, not capture the (stale) current cursor.
func TestEntityTab_MarkReloadCarriesPendingTarget(t *testing.T) {
	m := loaded(t, 120, 40) // cursor on alpha
	tab := m.cur()
	if got := tab.markReload(); got != m.selectedID() {
		t.Errorf("with nothing pending, markReload should capture the cursor, got %q want %q", got, m.selectedID())
	}
	// Simulate a jump in flight: reload() records the target as the pending restore.
	tab.restore = "jump-target"
	if got := tab.markReload(); got != "jump-target" {
		t.Errorf("markReload must carry the pending jump target forward, got %q", got)
	}
}

// TestModel_StaleReloadDoesNotStealRestore pins the per-message half of M6: the
// cursor restore rides on each load's message (gen-stamped), so a stale load
// finishing late can't apply a restore meant for a different generation.
func TestModel_StaleReloadDoesNotStealRestore(t *testing.T) {
	m := loaded(t, 120, 40) // cursor on alpha
	gen := m.cur().loadGen
	items := []list.Item{
		taskItem{t: domain.Task{Slug: "alpha", Status: domain.StatusInProgress}},
		taskItem{t: domain.Task{Slug: "beta", Status: domain.StatusReadyToStart}},
	}
	// A current-gen load carrying restore=beta selects beta.
	tm, _ := m.Update(listLoadedMsg{kind: entityTasks, gen: gen, items: items, restore: "beta"})
	m = tm.(Model)
	if m.selectedID() != "beta" {
		t.Fatalf("a current load's per-message restore should select beta, got %q", m.selectedID())
	}
	// A STALE load (older gen) carrying restore=alpha is dropped wholesale — its
	// restore must not steal the cursor back, nor fire a spurious not-found flash.
	tm, _ = m.Update(listLoadedMsg{kind: entityTasks, gen: gen - 1, items: items, restore: "alpha"})
	m = tm.(Model)
	if m.selectedID() != "beta" {
		t.Errorf("a stale-gen load must not apply its restore; cursor moved to %q", m.selectedID())
	}
	if m.flashErr {
		t.Errorf("a stale load must not flash, got %q", m.flash)
	}
}

// TestModel_ReloadDuringJumpKeepsTarget is the end-to-end M6 race: a jump to a
// view-hidden task escalates to :all and fires a reload (in flight); before it
// lands, a background reload fires. The jump target must win — not the cursor the
// background reload would have captured — with no spurious "not found".
func TestModel_ReloadDuringJumpKeepsTarget(t *testing.T) {
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
	write("tasks/in-progress/active-one.md", "---\nstatus: in-progress\nepic: 01-x\ndescription: a\n---\n# a\n")
	write("tasks/completed/done-one.md", "---\nstatus: completed\nepic: 01-x\ndescription: d\n---\n# d\n")

	m := New(core.NewService(store.NewFS(root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = toTasks(t, tm.(Model)) // default landing is the dashboard; drop onto tasks
	if n := len(m.cur().list.Items()); n != 1 {
		t.Fatalf("setup: working set should hide the completed task, got %d items", n)
	}

	// Jump to the hidden completed task: escalates to :all and fires the reload (its
	// load is left in flight — we deliberately don't drain the returned cmd).
	_ = m.jumpTo(entityTasks, "done-one")
	if m.cur().restore != "done-one" || m.cur().statusView != "all" {
		t.Fatalf("jump should escalate to :all and pend restore=done-one, got view=%q restore=%q",
			m.cur().statusView, m.cur().restore)
	}
	g1 := m.cur().loadGen

	// A background reload fires before the jump's load lands. markReload must carry
	// the jump target forward, so the new (higher-gen) load also aims at done-one.
	tm, reloadCmd := m.Update(reloadMsg{})
	m = tm.(Model)
	if m.cur().restore != "done-one" {
		t.Fatalf("a reload mid-jump must carry the jump target forward, got %q", m.cur().restore)
	}
	if m.cur().loadGen == g1 {
		t.Fatal("the background reload should have bumped the load generation")
	}

	// Land the reload; the jump target must win, with no false not-found flash.
	m = pump(t, m, reloadCmd, 8)
	if m.selectedID() != "done-one" {
		t.Errorf("the jump target should survive the concurrent reload, got %q", m.selectedID())
	}
	if m.flashErr {
		t.Errorf("no spurious not-found flash expected, got %q", m.flash)
	}
}
