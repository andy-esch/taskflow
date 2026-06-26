package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/domain"
)

// cursorTo drives the action menu's cursor onto the given verb via j-presses (so
// the real key path is exercised), failing if the verb isn't offered.
func cursorTo(t *testing.T, m Model, verb string) Model {
	t.Helper()
	idx := -1
	for i, tr := range m.action.options {
		if tr.verb == verb {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("verb %q not offered; options=%v", verb, m.action.options)
	}
	for m.action.cursor != idx {
		tm, _ := m.Update(press("j"))
		m = tm.(Model)
	}
	return m
}

// TestTransitionTablesValidStates guards M10's stringly-typed transition.to: a
// typo'd destination compiles (it's a string now) but would fail only at runtime
// when applyMove casts it. Pin every entity's table to its real state vocabulary.
func TestTransitionTablesValidStates(t *testing.T) {
	for _, tr := range taskTransitions {
		if !domain.Status(tr.to).Valid() {
			t.Errorf("task transition %q -> %q is not a valid status", tr.verb, tr.to)
		}
	}
	for _, tr := range auditTransitions {
		if !domain.AuditBucket(tr.to).Valid() {
			t.Errorf("audit transition %q -> %q is not a valid bucket", tr.verb, tr.to)
		}
	}
	for _, tr := range epicTransitions {
		if err := domain.ValidateEpicStatus(tr.to); err != nil {
			t.Errorf("epic transition %q -> %q is not a valid epic status: %v", tr.verb, tr.to, err)
		}
	}
}

func TestValidTransitions(t *testing.T) {
	got := validTransitions(taskTransitions, string(domain.StatusInProgress))
	if len(got) != len(taskTransitions)-1 {
		t.Errorf("want %d transitions (all but current), got %d", len(taskTransitions)-1, len(got))
	}
	for _, tr := range got {
		if tr.to == string(domain.StatusInProgress) {
			t.Error("the current status must be excluded from the menu")
		}
	}
	if tr, ok := transitionFor(taskTransitions, "complete"); !ok || tr.to != string(domain.StatusCompleted) {
		t.Errorf("complete should map to completed, got %v ok=%v", tr, ok)
	}
	if tr, ok := transitionFor(taskTransitions, "deprecate"); !ok || !tr.destructive {
		t.Errorf("deprecate should be destructive, got %v ok=%v", tr, ok)
	}
	if _, ok := transitionFor(taskTransitions, "bogus"); ok {
		t.Error("bogus is not a lifecycle verb")
	}
}

// TestModel_SuccessfulMoveKeepsSuccessFlash pins H5 (2026-06-22 audit):
// completing a task from the default view moves it out of the working set, and the
// post-move reload's cursor-restore must NOT report the just-moved task as
// "<slug> not found" and clobber the green success flash.
func TestModel_SuccessfulMoveKeepsSuccessFlash(t *testing.T) {
	m := loaded(t, 120, 40)
	if m.selectedID() != "alpha" {
		t.Fatalf("setup: want alpha selected, got %q", m.selectedID())
	}
	tm, _ := m.Update(press("m"))
	m = tm.(Model)
	m = cursorTo(t, m, "complete")
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	// Run the Move → movedMsg: the success flash is set and reloadAll kicked off.
	tm, reload := m.Update(cmd())
	m = tm.(Model)
	if m.flash == "" || m.flashErr {
		t.Fatalf("want a success flash after the move, got %q (err=%v)", m.flash, m.flashErr)
	}
	want := m.flash
	// Drive the reload: alpha is now completed and absent from the working-set view,
	// so the cursor-restore for "alpha" fails — but that absence is the success.
	m = drain(t, m, reload)
	if m.flashErr || m.flash != want {
		t.Errorf("post-move reload clobbered the success flash: got %q (err=%v), want %q", m.flash, m.flashErr, want)
	}
}

// TestModel_ActionMenuMovesTask is the end-to-end path: open the menu on a task,
// pick a non-destructive transition, and the real Service.Move relocates it.
func TestModel_ActionMenuMovesTask(t *testing.T) {
	m := loaded(t, 120, 40)
	if m.selectedID() != "alpha" {
		t.Fatalf("setup: want alpha selected, got %q", m.selectedID())
	}
	tm, _ := m.Update(press("m"))
	m = tm.(Model)
	if !m.action.active {
		t.Fatal("a should open the action menu on a task")
	}
	for _, tr := range m.action.options {
		if tr.verb == "start" {
			t.Error("start must be excluded for an already in-progress task")
		}
	}
	m = cursorTo(t, m, "complete")
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	if m.action.active {
		t.Error("a non-destructive apply should close the menu")
	}
	if cmd == nil {
		t.Fatal("apply should return a Move command")
	}
	// Run the Move, then apply its result.
	tm, _ = m.Update(cmd())
	m = tm.(Model)
	if m.flash == "" || m.flashErr {
		t.Errorf("expected a success flash, got %q (err=%v)", m.flash, m.flashErr)
	}
	// The file actually moved: alpha is now completed on disk.
	task, _, err := m.svc.ShowTask("alpha")
	if err != nil || task.Status != domain.StatusCompleted {
		t.Errorf("alpha should be completed after the action: status=%s err=%v", task.Status, err)
	}
}

// TestModel_ActionMenuConfirmGatesDeprecate pins the destructive guard: deprecate
// opens a y/n confirm instead of applying; n returns to the menu, y fires the move.
func TestModel_ActionMenuConfirmGatesDeprecate(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press("m"))
	m = cursorTo(t, tm.(Model), "deprecate")
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	if !m.action.active || !m.action.confirm {
		t.Fatal("deprecate should open the confirm gate, not apply")
	}
	if cmd != nil {
		t.Error("no move should fire before confirmation")
	}
	// n returns to the menu (still open, no longer confirming).
	tm, _ = m.Update(press("n"))
	m = tm.(Model)
	if !m.action.active || m.action.confirm {
		t.Error("n should return to the menu without applying or closing")
	}
	// Enter → confirm again, y → apply.
	tm, _ = m.Update(press("enter"))
	m = tm.(Model)
	tm, cmd = m.Update(press("y"))
	m = tm.(Model)
	if m.action.active {
		t.Error("y should close the menu")
	}
	if cmd == nil {
		t.Fatal("y should fire the move")
	}
	if msg, ok := cmd().(movedMsg); !ok || msg.to != string(domain.StatusDeprecated) {
		t.Fatalf("expected a deprecate movedMsg, got %T %+v", cmd(), cmd())
	}
}

func TestModel_CommandVerbMovesTask(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press(":"))
	m = tm.(Model)
	for _, r := range "complete" {
		tm, _ = m.Update(press(string(r)))
		m = tm.(Model)
	}
	_, cmd := m.Update(press("enter"))
	if cmd == nil {
		t.Fatal(":complete should fire a move")
	}
	if _, ok := cmd().(movedMsg); !ok {
		t.Fatalf(":complete should yield movedMsg, got %T", cmd())
	}

	// :deprecate gates on confirm even when typed explicitly.
	m2 := loaded(t, 120, 40)
	tm, _ = m2.Update(press(":"))
	m2 = tm.(Model)
	for _, r := range "deprecate" {
		tm, _ = m2.Update(press(string(r)))
		m2 = tm.(Model)
	}
	tm, _ = m2.Update(press("enter"))
	m2 = tm.(Model)
	if !m2.action.active || !m2.action.confirm {
		t.Error(":deprecate should open the confirm gate")
	}
	if !m2.action.confirmOnly() {
		t.Error("a :verb confirm has a single option (no menu to fall back to)")
	}
}

// TestModel_ActionMenuOpensOnEpic pins that the action menu is now registry-driven
// across entities: it opens on an epic (which declares status transitions), not
// just tasks. (Historically the `m` menu was task-only; audits then epics gained
// their own transition tables.) The full move flow lives in epic_action_test.go.
func TestModel_ActionMenuOpensOnEpic(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, cmd := m.Update(press("]")) // → epics
	m = drain(t, tm.(Model), cmd)
	if m.cur().name != "epics" {
		t.Fatalf("setup: expected epics, got %q", m.cur().name)
	}
	tm, _ = m.Update(press("m"))
	m = tm.(Model)
	if !m.action.active {
		t.Error("the action menu should open on an epic (epics now declare transitions)")
	}
}

func TestModel_ActionErrorFlashes(t *testing.T) {
	m := loaded(t, 120, 40)
	cmd := m.cur().applyMove(m.svc, "ghost-slug", transition{verb: "complete", to: string(domain.StatusCompleted)})
	msg := cmd()
	if _, ok := msg.(actionErrMsg); !ok {
		t.Fatalf("a failed move should yield actionErrMsg, got %T", msg)
	}
	tm, _ := m.Update(msg)
	m = tm.(Model)
	if m.flash == "" || !m.flashErr {
		t.Errorf("a move error should flash red, got %q (err=%v)", m.flash, m.flashErr)
	}
	if !strings.Contains(ansi.Strip(m.View().Content), "✘") {
		t.Error("the error flash should show in the footer")
	}
}

// typeRunes feeds each rune of s as a printable keypress (the text-input path).
func typeRunes(t *testing.T, m Model, s string) Model {
	t.Helper()
	for _, r := range s {
		tm, _ := m.Update(press(string(r)))
		m = tm.(Model)
	}
	return m
}

// TestModel_DeferPromptsForRevisitDate is the TUI parity for `task defer --until`:
// selecting defer opens a revisit-date prompt (not an immediate move), and the
// typed date is recorded on disk and confirmed in the flash.
func TestModel_DeferPromptsForRevisitDate(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press("m"))
	m = cursorTo(t, tm.(Model), "defer")
	tm, _ = m.Update(press("enter"))
	m = tm.(Model)
	if !m.action.revisit {
		t.Fatal("selecting defer should open the revisit-date prompt, not apply at once")
	}
	if task, _, _ := m.svc.ShowTask("alpha"); task.Status == domain.StatusDeferred {
		t.Fatal("defer must not move the task before a date is entered")
	}

	m = typeRunes(t, m, "2026-09-01")
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	if m.action.active {
		t.Error("applying the date should close the menu")
	}
	if cmd == nil {
		t.Fatal("entering the date should fire a defer command")
	}
	tm, _ = m.Update(cmd()) // run DeferTask → movedMsg
	m = tm.(Model)

	task, _, err := m.svc.ShowTask("alpha")
	if err != nil || task.Status != domain.StatusDeferred {
		t.Fatalf("alpha should be deferred: status=%s err=%v", task.Status, err)
	}
	if task.RevisitAt != "2026-09-01" {
		t.Errorf("defer should record the revisit date, got %q", task.RevisitAt)
	}
	if !strings.Contains(m.flash, "2026-09-01") {
		t.Errorf("flash should confirm the revisit date, got %q", m.flash)
	}
}

// TestModel_DeferBlankParksIndefinitely pins that an empty revisit prompt defers
// with no revisit_at (the snooze stays opt-in).
func TestModel_DeferBlankParksIndefinitely(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press("m"))
	m = cursorTo(t, tm.(Model), "defer")
	tm, _ = m.Update(press("enter")) // open the prompt
	m = tm.(Model)
	tm, cmd := m.Update(press("enter")) // blank → apply
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("a blank date should still defer (park indefinitely)")
	}
	tm, _ = m.Update(cmd())
	m = tm.(Model)
	task, _, err := m.svc.ShowTask("alpha")
	if err != nil || task.Status != domain.StatusDeferred {
		t.Fatalf("alpha should be deferred: status=%s err=%v", task.Status, err)
	}
	if task.RevisitAt != "" {
		t.Errorf("a blank revisit prompt must not set revisit_at, got %q", task.RevisitAt)
	}
}

// TestModel_DeferBadDateShowsError pins that a malformed date keeps the prompt open
// with an inline error and never moves the task.
func TestModel_DeferBadDateShowsError(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press("m"))
	m = cursorTo(t, tm.(Model), "defer")
	tm, _ = m.Update(press("enter"))
	m = typeRunes(t, tm.(Model), "soon")
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	if cmd != nil {
		t.Error("a bad date must not fire a defer")
	}
	if !m.action.revisit || m.action.dateErr == "" {
		t.Errorf("a bad date should keep the prompt open with an error, got revisit=%v err=%q", m.action.revisit, m.action.dateErr)
	}
	if task, _, _ := m.svc.ShowTask("alpha"); task.Status == domain.StatusDeferred {
		t.Error("a bad date must not move the task")
	}
}

// TestModel_DeferEscReturnsToMenu pins that Esc from the date prompt returns to the
// action menu (it was opened from the menu), rather than cancelling outright.
func TestModel_DeferEscReturnsToMenu(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press("m"))
	m = cursorTo(t, tm.(Model), "defer")
	tm, _ = m.Update(press("enter"))
	m = tm.(Model)
	if !m.action.revisit {
		t.Fatal("defer should open the revisit prompt")
	}
	tm, _ = m.Update(press("esc"))
	m = tm.(Model)
	if !m.action.active || m.action.revisit {
		t.Errorf("esc from the date prompt should return to the menu, got active=%v revisit=%v", m.action.active, m.action.revisit)
	}
}

// TestModel_CommandDeferPromptsForDate pins that the `:defer` verb also opens the
// revisit prompt (cold, no menu open); Esc then closes it since there's no menu to
// return to.
func TestModel_CommandDeferPromptsForDate(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press(":"))
	m = typeRunes(t, tm.(Model), "defer")
	tm, _ = m.Update(press("enter")) // submit the command
	m = tm.(Model)
	if !m.action.revisit {
		t.Fatal(":defer should open the revisit-date prompt")
	}
	tm, _ = m.Update(press("esc"))
	m = tm.(Model)
	if m.action.active {
		t.Error("esc from a cold :defer prompt should close it (no menu to return to)")
	}
}

// TestModel_AuditDeferSkipsRevisitPrompt pins that the revisit prompt is task-only:
// an audit "defer" (a bucket move with no revisit date) still applies immediately.
func TestModel_AuditDeferSkipsRevisitPrompt(t *testing.T) {
	m := loaded(t, 120, 40)
	m = auditsTab(t, m)
	tm, _ := m.Update(press("m"))
	m = cursorTo(t, tm.(Model), "defer")
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	if m.action.revisit {
		t.Error("audit defer must not open a revisit prompt")
	}
	if cmd == nil {
		t.Fatal("audit defer should apply immediately")
	}
	if msg, ok := cmd().(movedMsg); !ok || msg.to != string(domain.AuditDeferred) {
		t.Fatalf("expected an audit defer movedMsg, got %T %+v", cmd(), cmd())
	}
}

// TestModel_ActionMenuFitsTerminal keeps the layout invariant with the menu open:
// the overlay must not change the view height or overflow the width.
func TestModel_ActionMenuFitsTerminal(t *testing.T) {
	for _, d := range []struct{ w, h int }{
		{120, 40}, {100, 24}, {80, 20}, {40, 12}, {24, 8},
	} {
		m := loaded(t, d.w, d.h)
		tm, _ := m.Update(press("m"))
		m = tm.(Model)
		lines := strings.Split(m.View().Content, "\n")
		if len(lines) != d.h {
			t.Errorf("%dx%d with action menu: %d lines, want %d", d.w, d.h, len(lines), d.h)
		}
		for i, ln := range lines {
			if w := ansi.StringWidth(ln); w > d.w {
				t.Errorf("%dx%d with action menu: line %d is %d wide > %d", d.w, d.h, i, w, d.w)
			}
		}
	}
}
