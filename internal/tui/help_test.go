package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestHelpBoxFixedWidthAcrossScroll pins the fix for the reported resize bug: the
// overlay is a CONSTANT width — every line the same, and the same at any scroll
// offset — instead of growing/shrinking to the widest currently-visible line.
func TestHelpBoxFixedWidthAcrossScroll(t *testing.T) {
	const maxW, maxH = 80, 16 // short enough that content overflows and scroll matters
	lineWidths := func(scroll int) []int {
		var ws []int
		for _, ln := range strings.Split(helpBox(maxW, maxH, scroll, focusList, entityTasks), "\n") {
			ws = append(ws, ansi.StringWidth(ln))
		}
		return ws
	}
	want := helpWidth(maxW) // the fixed outer width
	for _, scroll := range []int{0, 3, 8} {
		for i, w := range lineWidths(scroll) {
			if w != want {
				t.Fatalf("scroll %d line %d width = %d, want a constant %d (box resized)", scroll, i, w, want)
			}
		}
	}
}

// TestHelpWrapsLongDescriptions pins that an over-long description wraps within its
// column (continuation lines indented under the description) instead of widening the
// box. The `m`/`e` rows carry the longest descriptions.
func TestHelpWrapsLongDescriptions(t *testing.T) {
	const maxW = 80
	contentW := helpWidth(maxW) - helpHFrame
	lines := helpLines(focusList, entityTasks, contentW)
	wrapped := false
	for _, ln := range lines {
		if ansi.StringWidth(ln) > contentW {
			t.Errorf("line exceeds the content width %d: %q (%d)", contentW, ln, ansi.StringWidth(ln))
		}
	}
	// The long `m` description must span more than one line (proof it wrapped, not
	// truncated): a continuation line carries text but no key.
	for i, ln := range lines {
		if strings.Contains(ln, "lifecycle") && i+1 < len(lines) {
			if next := strings.TrimSpace(ansi.Strip(lines[i+1])); next != "" && !strings.HasPrefix(next, "e ") {
				wrapped = true
			}
		}
	}
	if !wrapped {
		t.Error("the long 'm' description should wrap to a continuation line")
	}
}

// TestHelpContextDependent pins the focus-aware `?`: the active pane's nav section
// shows, the inactive pane's is hidden, and Global + Notes always show.
func TestHelpContextDependent(t *testing.T) {
	// A wide content width so descriptions don't wrap — this test checks which
	// entries appear and their order, not the wrapping.
	const wide = 120
	list := strings.Join(helpLines(focusList, entityTasks, wide), "\n")
	detail := strings.Join(helpLines(focusDetail, entityTasks, wide), "\n")

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
		// "command palette" = Global; the find note = Notes (kind-independent, unlike
		// the entity-specific views note). Both must always show.
		if !strings.Contains(v, "command palette") || !strings.Contains(v, "matches the rendered text on screen") {
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
// fix, helpScroll ran up to len(helpLines(m.focus, m.cur().kind)); after a burst of j's at the bottom
// the view sat still while a pile of dead k-presses had to be spent first.
func TestModel_HelpScrollClampedToVisibleMax(t *testing.T) {
	// A short terminal so the help content overflows its box (helpMaxScroll > 0).
	m := loaded(t, 100, 14)
	tm, _ := m.Update(press("?"))
	m = tm.(Model)
	if !m.showHelp {
		t.Fatal("? should open the help overlay")
	}
	// Count lines at the SAME content width helpMaxScroll/helpBox use, so the clamp
	// and the line count agree under wrapping.
	cw := helpWidth(m.width-2) - helpHFrame
	maxScroll := m.helpMaxScroll()
	if maxScroll <= 0 || maxScroll >= len(helpLines(m.focus, m.cur().kind, cw)) {
		t.Fatalf("test needs an overflowing overlay with a real clamp; helpMaxScroll=%d, lines=%d", maxScroll, len(helpLines(m.focus, m.cur().kind, cw)))
	}

	// Press j well past the bottom — it must stop at the visible max.
	for i := 0; i < len(helpLines(m.focus, m.cur().kind, cw))+5; i++ {
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
