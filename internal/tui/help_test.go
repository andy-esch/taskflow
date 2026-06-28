package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestHelpBoxFixedWidthAcrossScroll pins the fixed-width invariant across the cases a
// single-width test missed: every line is EXACTLY contentW — at every scroll offset,
// every kind, and every terminal width including narrow ones where the column can't
// fit and the truncate backstop must keep rows from widening the box. Checked both at
// the helpLines level (the real invariant) and the rendered-box level (uniform border).
func TestHelpBoxFixedWidthAcrossScroll(t *testing.T) {
	for _, kind := range []entityKind{entityTasks, entityEpics, entityAudits, entityDashboard} {
		for _, maxW := range []int{20, 30, 47, 62, 120} { // 20/30 = narrow (backstop regime)
			contentW := helpWidth(maxW) - helpHFrame
			// Invariant 1 (the one the resize/clip bug violated): no composed line may
			// exceed contentW — every line is forced to exactly contentW.
			for _, ln := range helpLines(focusList, kind, contentW) {
				if w := ansi.StringWidth(ln); w != contentW {
					t.Errorf("kind=%d maxW=%d: line width %d, want exactly contentW %d: %q", kind, maxW, w, contentW, ansi.Strip(ln))
				}
			}
			// Invariant 2: the rendered box is one constant width at every scroll.
			want := helpWidth(maxW)
			for _, scroll := range []int{0, 3, 9} {
				for i, ln := range strings.Split(helpBox(maxW, 16, scroll, focusList, kind), "\n") {
					if w := ansi.StringWidth(ln); w != want {
						t.Errorf("kind=%d maxW=%d scroll=%d: box line %d width %d, want %d", kind, maxW, scroll, i, w, want)
					}
				}
			}
		}
	}
}

// TestHelpWrapsLongDescriptions proves long descriptions WRAP within their column
// (rather than truncate or widen the box): a narrow content width yields strictly more
// lines than a wide one, since several descriptions reflow to continuation lines. No
// dependency on the exact wording of any entry, so a reworded description won't break it.
func TestHelpWrapsLongDescriptions(t *testing.T) {
	wide := len(helpLines(focusList, entityTasks, 200))  // descriptions all fit on one line
	narrow := len(helpLines(focusList, entityTasks, 40)) // several reflow to 2+ lines
	if narrow <= wide {
		t.Errorf("narrow help (%d lines) should wrap to MORE lines than wide (%d) — descriptions not wrapping", narrow, wide)
	}
}

// TestSymbolsLegendIsPageSpecific pins that the glyph legend is context-specific like
// the keys: each tab's Symbols section names its own vocabulary (statuses / liveness /
// buckets+findings), not another tab's.
func TestSymbolsLegendIsPageSpecific(t *testing.T) {
	text := func(kind entityKind) string {
		sec, _ := symbolsFor(kind)
		var b strings.Builder
		for _, e := range sec.entries {
			b.WriteString(e.desc + "\n")
		}
		return b.String()
	}
	tasks, epics, audits := text(entityTasks), text(entityEpics), text(entityAudits)
	if !strings.Contains(tasks, "misfiled") || strings.Contains(tasks, "liveness") {
		t.Error("tasks legend should describe task markers, not epic liveness")
	}
	if !strings.Contains(epics, "dormant") || strings.Contains(epics, "finding:") {
		t.Error("epics legend should describe liveness, not audit findings")
	}
	if !strings.Contains(audits, "finding:") || !strings.Contains(audits, "bucket") {
		t.Error("audits legend should describe buckets + finding statuses")
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
