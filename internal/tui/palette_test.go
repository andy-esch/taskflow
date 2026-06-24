package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func ctrlP() tea.KeyPressMsg { return tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl} }

// TestPalette_OpenClose: ctrl+p opens the palette; esc closes it.
func TestPalette_OpenClose(t *testing.T) {
	m := loaded(t, 80, 24)
	tm, _ := m.Update(ctrlP())
	m = tm.(Model)
	if !m.palette.active {
		t.Fatal("ctrl+p should open the palette")
	}
	tm, _ = m.Update(press("esc"))
	m = tm.(Model)
	if m.palette.active {
		t.Error("esc should close the palette")
	}
}

// TestPalette_IndexHasEntitiesAndCommands: the candidate set spans loaded entities
// (jump) and the `:` command words (dispatch).
func TestPalette_IndexHasEntitiesAndCommands(t *testing.T) {
	m := loaded(t, 80, 24)
	idx := m.paletteIndex()
	jumps, cmds, sawEpicsCmd := 0, 0, false
	for _, it := range idx {
		switch it.kind {
		case palJump:
			jumps++
		case palCommand:
			cmds++
			if it.word == "epics" {
				sawEpicsCmd = true
			}
		}
	}
	if jumps == 0 {
		t.Error("index should contain entity jumps for the loaded tasks")
	}
	if cmds == 0 || !sawEpicsCmd {
		t.Errorf("index should contain command words including 'epics' (jumps=%d cmds=%d)", jumps, cmds)
	}
}

// TestPalette_JumpSelectsEntity: running a jump candidate selects that entity on
// its tab.
func TestPalette_JumpSelectsEntity(t *testing.T) {
	m := loaded(t, 80, 24)
	items := m.tabs[0].list.Items() // the tasks tab
	if len(items) < 2 {
		t.Skip("need ≥2 seeded tasks to prove the jump moves the selection")
	}
	target := items[len(items)-1].(entityItem).id() // not the default-selected first row
	if m.selectedID() == target {
		t.Fatalf("setup: target %q is already selected", target)
	}
	m.runPaletteItem(paletteItem{kind: palJump, ek: entityTasks, id: target})
	if got := m.selectedID(); got != target {
		t.Errorf("jump should select %q, selected %q", target, got)
	}
}

// TestPalette_CommandSwitchesTab: running a command word dispatches it (here, a
// tab switch), exactly like the `:` bar.
func TestPalette_CommandSwitchesTab(t *testing.T) {
	m := loaded(t, 80, 24)
	if m.cur().kind != entityTasks {
		t.Fatalf("setup: expected to start on the tasks tab, got %v", m.cur().kind)
	}
	m.runPaletteCommand("epics")
	if m.cur().kind != entityEpics {
		t.Errorf("running :epics should switch to the epics tab, on %v", m.cur().kind)
	}
}

// TestPalette_FilterNarrows: typing into the palette fuzzy-narrows the candidate
// list to (and keeps) the matching entity.
func TestPalette_FilterNarrows(t *testing.T) {
	m := loaded(t, 80, 24)
	tm, _ := m.Update(ctrlP())
	m = tm.(Model)
	total := len(m.palette.list.Items())

	target := m.tabs[0].list.Items()[0].(entityItem).id()
	for _, r := range target {
		tm, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = tm.(Model)
	}
	got := m.palette.list.Items()
	if len(got) == 0 || len(got) >= total {
		t.Fatalf("typing %q should narrow the list (was %d, now %d)", target, total, len(got))
	}
	found := false
	for _, it := range got {
		if pi, ok := it.(paletteItem); ok && pi.id == target {
			found = true
		}
	}
	if !found {
		t.Errorf("the typed target %q should remain in the filtered candidates", target)
	}
}
