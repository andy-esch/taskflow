package tui

import (
	"strings"
	"testing"
)

// TestHelpContextDependent pins the focus-aware `?`: the active pane's nav section
// shows, the inactive pane's is hidden, and Global + Notes always show.
func TestHelpContextDependent(t *testing.T) {
	list := strings.Join(helpLines(focusList), "\n")
	detail := strings.Join(helpLines(focusDetail), "\n")

	if !strings.Contains(list, "open detail") { // a List-only entry
		t.Error("list-focus help should include List keys")
	}
	if strings.Contains(list, "raw ⇄ pretty markdown") { // a Detail-only entry
		t.Error("list-focus help must hide Detail keys")
	}
	if !strings.Contains(detail, "raw ⇄ pretty markdown") {
		t.Error("detail-focus help should include Detail keys")
	}
	if strings.Contains(detail, "open detail") {
		t.Error("detail-focus help must hide List keys")
	}
	for _, v := range []string{list, detail} {
		if !strings.Contains(v, "command palette") || !strings.Contains(v, "switch bucket") {
			t.Error("help should always include Global and Notes sections")
		}
	}
	// Ordering: the active pane's keys come first; Global (command palette) is last.
	if strings.Index(list, "open detail") > strings.Index(list, "command palette") {
		t.Error("list-focus help should list the List section before Global")
	}
	if strings.Index(detail, "raw ⇄ pretty markdown") > strings.Index(detail, "command palette") {
		t.Error("detail-focus help should list the Detail section before Global")
	}
}

// TestModel_HelpScrollClampedToVisibleMax pins that j/k clamp the help overlay
// to its visible bottom (helpMaxScroll), not the total line count. Before the
// fix, helpScroll ran up to len(helpLines(m.focus)); after a burst of j's at the bottom
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
	if maxScroll <= 0 || maxScroll >= len(helpLines(m.focus)) {
		t.Fatalf("test needs an overflowing overlay with a real clamp; helpMaxScroll=%d, lines=%d", maxScroll, len(helpLines(m.focus)))
	}

	// Press j well past the bottom — it must stop at the visible max.
	for i := 0; i < len(helpLines(m.focus))+5; i++ {
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
