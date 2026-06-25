package tui

import "testing"

// epicsTab switches to the epics tab and settles its load.
func epicsTab(t *testing.T, m Model) Model {
	t.Helper()
	for i := 0; i < len(m.tabs) && m.cur().name != "epics"; i++ {
		tm, cmd := m.Update(press("]"))
		m = drain(t, tm.(Model), cmd)
	}
	if m.cur().name != "epics" {
		t.Fatalf("could not reach the epics tab, on %q", m.cur().name)
	}
	return m
}

// TestModel_ActionMenuMovesEpic pins the epic half of the registry-driven `m`
// menu: pressing `m` on an active epic offers retire + deprecate (active, the
// no-op, is dropped), and applying one rewrites the epic's status field on disk
// (no file moves) and live-reloads.
func TestModel_ActionMenuMovesEpic(t *testing.T) {
	m := loaded(t, 120, 40)
	m = epicsTab(t, m)
	if m.selectedID() != "01-test" {
		t.Fatalf("setup: want the seeded epic selected, got %q", m.selectedID())
	}
	tm, _ := m.Update(press("m"))
	m = tm.(Model)
	if !m.action.active {
		t.Fatal("m should open the action menu on an epic")
	}
	// Active epic: activate (→active) is the no-op row and dropped; retire + deprecate remain.
	offered := map[string]bool{}
	for _, tr := range m.action.options {
		offered[tr.verb] = true
	}
	if !offered["retire"] || !offered["deprecate"] {
		t.Errorf("epic menu should offer retire + deprecate, got %v", m.action.options)
	}
	if offered["activate"] {
		t.Errorf("activate must be dropped for an already-active epic, got %v", m.action.options)
	}
	m = cursorTo(t, m, "retire")
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	if m.action.active {
		t.Error("a non-destructive apply should close the menu")
	}
	if cmd == nil {
		t.Fatal("apply should return a MoveEpic command")
	}
	tm, _ = m.Update(cmd())
	m = tm.(Model)
	if m.flash == "" || m.flashErr {
		t.Errorf("expected a success flash, got %q (err=%v)", m.flash, m.flashErr)
	}
	// The status field was rewritten in place — the file did not move.
	e, _, _, err := m.svc.ShowEpic("01-test")
	if err != nil || e.Status != "retired" {
		t.Errorf("epic should be retired after the action: status=%q err=%v", e.Status, err)
	}
}

// TestModel_CommandVerbMovesEpic pins the `:`-verb half: a verb is resolved
// against the active tab's transition table, so `:retire` moves an epic.
func TestModel_CommandVerbMovesEpic(t *testing.T) {
	m := loaded(t, 120, 40)
	m = epicsTab(t, m)
	tm, _ := m.Update(press(":"))
	m = tm.(Model)
	for _, r := range "retire" {
		tm, _ = m.Update(press(string(r)))
		m = tm.(Model)
	}
	_, cmd := m.Update(press("enter"))
	if cmd == nil {
		t.Fatal(":retire should fire a MoveEpic")
	}
	msg, ok := cmd().(movedMsg)
	if !ok || msg.to != "retired" {
		t.Fatalf(":retire should yield a movedMsg → retired, got %T %+v", cmd(), cmd())
	}
}
