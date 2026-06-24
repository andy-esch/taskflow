package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	tea "charm.land/bubbletea/v2"
)

func ctrlC() tea.KeyPressMsg { return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl} }

func quitsApp(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

// TestModal_ForceQuitFromEachOverlay pins that ForceQuit (ctrl+c) escapes every
// modal overlay — the guarantee the handleKey preamble centralizes, instead of the
// per-overlay ForceQuit cases the old guard chain duplicated.
func TestModal_ForceQuitFromEachOverlay(t *testing.T) {
	// help (?)
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press("?"))
	m = tm.(Model)
	if !m.showHelp {
		t.Fatal("setup: ? should open help")
	}
	if _, cmd := m.Update(ctrlC()); !quitsApp(cmd) {
		t.Error("ctrl+c must quit from the help overlay")
	}

	// action menu (m)
	m = loaded(t, 120, 40)
	tm, _ = m.Update(press("m"))
	m = tm.(Model)
	if !m.action.active {
		t.Fatal("setup: a should open the action menu")
	}
	if _, cmd := m.Update(ctrlC()); !quitsApp(cmd) {
		t.Error("ctrl+c must quit from the action menu")
	}

	// follow picker (f on an epic)
	m = loaded(t, 120, 40)
	m = followToEpic(t, m)
	tm, _ = m.Update(press("f"))
	m = tm.(Model)
	if !m.follow.active {
		t.Fatalf("setup: f on an epic should open the follow picker (flash=%q)", m.flash)
	}
	if _, cmd := m.Update(ctrlC()); !quitsApp(cmd) {
		t.Error("ctrl+c must quit from the follow picker")
	}

	// command palette (ctrl+p)
	m = loaded(t, 120, 40)
	tm, _ = m.Update(ctrlP())
	m = tm.(Model)
	if !m.palette.active {
		t.Fatal("setup: ctrl+p should open the palette")
	}
	if _, cmd := m.Update(ctrlC()); !quitsApp(cmd) {
		t.Error("ctrl+c must quit from the command palette")
	}
}

// TestModal_CapturesKeysWhileActive pins overlay precedence: an active modal owns
// keys that would otherwise be global hotkeys, so they don't leak through. With the
// action menu open, `]` (next-tab) must not switch tabs.
func TestModal_CapturesKeysWhileActive(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press("m"))
	m = tm.(Model)
	if !m.action.active {
		t.Fatal("setup: a should open the action menu")
	}
	before := m.cur().name
	tm, _ = m.Update(press("]")) // next-tab — must be swallowed by the open menu
	m = tm.(Model)
	if m.cur().name != before {
		t.Errorf("an open modal must capture `]`; tab switched %s → %s", before, m.cur().name)
	}
	if !m.action.active {
		t.Error("`]` should be a no-op inside the menu, not close it")
	}
}

// TestModal_FallsThroughWhenNoneActive pins the other side: with no overlay active,
// a key reaches base routing — `]` switches tabs.
func TestModal_FallsThroughWhenNoneActive(t *testing.T) {
	m := loaded(t, 120, 40)
	if m.showHelp || m.action.active || m.follow.active {
		t.Fatal("setup: no overlay should be active")
	}
	tm, cmd := m.Update(press("]"))
	m = drain(t, tm.(Model), cmd)
	if m.cur().name != "epics" {
		t.Errorf("with no overlay active, `]` should switch to epics, got %q", m.cur().name)
	}
}

// TestModal_BodyViewCompositesActiveOverlay pins that the active modal is floated
// over the body (the bodyView modal loop), so View() shows the menu's content.
func TestModal_BodyViewCompositesActiveOverlay(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press("m"))
	m = tm.(Model)
	v := ansi.Strip(m.View().Content)
	if !strings.Contains(v, "move alpha") {
		t.Errorf("the action menu should be composited over the body:\n%s", v)
	}
	// The base panes stay visible around the floated box.
	if !strings.Contains(v, "beta") {
		t.Errorf("the underlying list should remain visible around the overlay:\n%s", v)
	}
}
