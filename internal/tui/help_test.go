package tui

import "testing"

// TestModel_HelpScrollClampedToVisibleMax pins that j/k clamp the help overlay
// to its visible bottom (helpMaxScroll), not the total line count. Before the
// fix, helpScroll ran up to len(helpLines()); after a burst of j's at the bottom
// the view sat still while a pile of dead k-presses had to be spent first.
func TestModel_HelpScrollClampedToVisibleMax(t *testing.T) {
	// A short terminal so the help content overflows its box (helpMaxScroll > 0).
	m := loaded(t, 100, 14)
	tm, _ := m.Update(press("?"))
	m = tm.(Model)
	if !m.showHelp {
		t.Fatal("? should open the help overlay")
	}
	maxScroll := m.helpMaxScroll()
	if maxScroll <= 0 || maxScroll >= len(helpLines()) {
		t.Fatalf("test needs an overflowing overlay with a real clamp; helpMaxScroll=%d, lines=%d", maxScroll, len(helpLines()))
	}

	// Press j well past the bottom — it must stop at the visible max.
	for i := 0; i < len(helpLines())+5; i++ {
		tm, _ := m.Update(press("j"))
		m = tm.(Model)
	}
	if m.helpScroll != maxScroll {
		t.Errorf("helpScroll should clamp to %d, got %d (ran past the visible bottom)", maxScroll, m.helpScroll)
	}

	// A single k must scroll up immediately — no backlog of dead presses.
	tm, _ = m.Update(press("k"))
	m = tm.(Model)
	if m.helpScroll != maxScroll-1 {
		t.Errorf("one k should scroll up immediately to %d, got %d", maxScroll-1, m.helpScroll)
	}
}
