package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// cleanTaskRepo seeds a lint-clean in-progress task "clean" (with tags, so the
// active-task invariant lets a single-field SetFields through) and its epic.
func cleanTaskRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mk := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("epics/e1.md", "---\nstatus: active\n---\n# E1\n")
	mk("tasks/in-progress/clean.md",
		"---\nstatus: in-progress\nepic: e1\ntier: 2\npriority: low\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\ndescription: d\n---\n# Clean\n")
	return root
}

func loadedAt(t *testing.T, root string, w, h int) Model {
	t.Helper()
	m := New(core.NewService(store.NewFS(root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()()) // landing dashboard load
	return toTasks(t, tm.(Model))
}

// editCursorTo drives the field picker onto key via j-presses.
func editCursorTo(t *testing.T, m Model, key string) Model {
	t.Helper()
	for i := 0; i < len(m.edit.fields); i++ {
		if m.edit.cur().key == key {
			return m
		}
		tm, _ := m.Update(press("j"))
		m = tm.(Model)
	}
	t.Fatalf("field %q not reachable in the picker", key)
	return m
}

// enumCursorTo drives an open enum widget onto opt via j-presses.
func enumCursorTo(t *testing.T, m Model, opt string) Model {
	t.Helper()
	for i := 0; i <= len(m.edit.cur().options); i++ {
		if m.edit.value() == opt {
			return m
		}
		tm, _ := m.Update(press("j"))
		m = tm.(Model)
	}
	t.Fatalf("enum option %q not reachable", opt)
	return m
}

// TestModel_EditPriorityViaMenu pins the enum happy path: e → pick priority → choose
// high → apply persists via SetFields, flashes success, and reloads.
func TestModel_EditPriorityViaMenu(t *testing.T) {
	m := loadedAt(t, cleanTaskRepo(t), 120, 40)
	if m.selectedID() != "clean" {
		t.Fatalf("setup: want clean selected, got %q", m.selectedID())
	}
	tm, _ := m.Update(press("e"))
	m = tm.(Model)
	if !m.edit.active {
		t.Fatal("e should open the edit field picker on a task")
	}
	m = editCursorTo(t, m, "priority")
	tm, _ = m.Update(press("enter")) // begin editing the enum (starts on current "low")
	m = tm.(Model)
	if !m.edit.editing {
		t.Fatal("enter should begin editing the field")
	}
	m = enumCursorTo(t, m, "high")
	tm, cmd := m.Update(press("enter")) // apply (write in flight; still on the field)
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("apply should return a SetFields command")
	}
	tm, _ = m.Update(cmd()) // run SetFields → editedMsg
	m = tm.(Model)
	if !m.edit.active || m.edit.editing {
		t.Error("a successful apply should return to the picker (open, not editing)")
	}
	if m.flash == "" || m.flashErr {
		t.Errorf("expected a success flash, got %q (err=%v)", m.flash, m.flashErr)
	}
	task, _, err := m.svc.ShowTask("clean")
	if err != nil || task.Priority != "high" {
		t.Errorf("priority should be high after the edit: %q (%v)", task.Priority, err)
	}
}

// TestModel_EditStaysOpenForMultipleFields pins the form flow: applying a field
// returns to the picker (not closing) so several fields can be edited in one
// session; only Esc from the picker closes.
func TestModel_EditStaysOpenForMultipleFields(t *testing.T) {
	m := loadedAt(t, cleanTaskRepo(t), 120, 40)
	tm, _ := m.Update(press("e"))
	m = tm.(Model)

	// First field: priority → high.
	m = editCursorTo(t, m, "priority")
	tm, _ = m.Update(press("enter"))
	m = enumCursorTo(t, tm.(Model), "high")
	tm, cmd := m.Update(press("enter")) // apply
	m = tm.(Model)
	tm, _ = m.Update(cmd()) // land editedMsg → back to the picker, value refreshed
	m = tm.(Model)
	if !m.edit.active || m.edit.editing {
		t.Fatal("after a successful apply, the editor should be back at the picker (still open)")
	}
	for _, f := range m.edit.fields {
		if f.key == "priority" && f.current != "high" {
			t.Errorf("the picker should show the updated priority, got %q", f.current)
		}
	}

	// Second field, without re-opening: tier → 1.
	m = editCursorTo(t, m, "tier")
	tm, _ = m.Update(press("enter"))
	m = enumCursorTo(t, tm.(Model), "1")
	tm, cmd = m.Update(press("enter"))
	m = tm.(Model)
	tm, _ = m.Update(cmd())
	m = tm.(Model)

	// Esc from the picker closes.
	tm, _ = m.Update(press("esc"))
	m = tm.(Model)
	if m.edit.active {
		t.Error("esc from the picker should close the editor")
	}
	if task, _, _ := m.svc.ShowTask("clean"); task.Priority != "high" || task.Tier != 1 {
		t.Errorf("both edits should persist: priority=%q tier=%d", task.Priority, task.Tier)
	}
}

// TestModel_EditDescriptionViaTextInput pins the long-text happy path: a typed
// keystroke reaches the word-wrapped description box and the new value persists via
// SetFields (Enter submits — it never inserts a newline, so description stays one line).
func TestModel_EditDescriptionViaTextInput(t *testing.T) {
	m := loadedAt(t, cleanTaskRepo(t), 120, 40)
	tm, _ := m.Update(press("e"))
	m = editCursorTo(t, tm.(Model), "description")
	tm, _ = m.Update(press("enter")) // begin editing (prefilled "d", cursor at end)
	m = tm.(Model)
	tm, _ = m.Update(press("x")) // type a char — exercises the textarea key path
	m = tm.(Model)
	if !strings.Contains(m.edit.area.Value(), "x") {
		t.Fatalf("the keystroke should reach the description box, got %q", m.edit.area.Value())
	}
	_, cmd := m.Update(press("enter")) // apply
	if cmd == nil {
		t.Fatal("apply should return a SetFields command")
	}
	cmd() // run SetFields
	task, _, err := m.svc.ShowTask("clean")
	if err != nil || task.Description != "dx" {
		t.Errorf("description should be the typed value 'dx', got %q (%v)", task.Description, err)
	}
}

// TestModel_EditRejectedSurfacesError pins the validation contract: clearing tags on
// an active task is rejected by core, the error is shown ON the field (not a bounce
// to the picker or a silent revert), what was typed is kept, nothing is written —
// and the user can fix it in place and re-submit.
func TestModel_EditRejectedSurfacesError(t *testing.T) {
	m := loadedAt(t, cleanTaskRepo(t), 120, 40)
	tm, _ := m.Update(press("e"))
	m = editCursorTo(t, tm.(Model), "tags")
	tm, _ = m.Update(press("enter")) // edit tags (prefilled "a")
	m = tm.(Model)
	tm, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace}) // clear → ""
	m = tm.(Model)
	if m.edit.input.Value() != "" {
		t.Fatalf("tags input should be empty, got %q", m.edit.input.Value())
	}
	tm, cmd := m.Update(press("enter")) // submit empty tags
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("submit should return a command")
	}
	tm, _ = m.Update(cmd()) // run SetFields → rejection
	m = tm.(Model)
	if !m.edit.active || !m.edit.editing {
		t.Error("a rejected edit should keep the field open for a fix, not bounce to the picker")
	}
	if m.edit.err == "" || !strings.Contains(m.edit.err, "tag") {
		t.Errorf("the validation error should be shown on the field, got %q", m.edit.err)
	}
	if task, _, _ := m.svc.ShowTask("clean"); len(task.Tags) != 1 || task.Tags[0] != "a" {
		t.Errorf("a rejected edit must not write; tags=%v", task.Tags)
	}
	// The user fixes it in place: type a valid tag and re-submit.
	tm, _ = m.Update(press("x"))
	m = tm.(Model)
	if m.edit.err != "" {
		t.Error("typing should clear the stale error")
	}
	tm, cmd = m.Update(press("enter"))
	m = tm.(Model)
	tm, _ = m.Update(cmd()) // success
	m = tm.(Model)
	if m.edit.editing {
		t.Error("a successful re-submit should return to the picker")
	}
	if task, _, _ := m.svc.ShowTask("clean"); len(task.Tags) != 1 || task.Tags[0] != "x" {
		t.Errorf("the fixed tag should persist; tags=%v", task.Tags)
	}
}

// TestModel_EditTagsCoercedToList pins the one field with non-trivial coercion: a
// comma-list typed in the editor lands as a YAML list (the same coercion `task set`
// applies), so the GUI and agent faces agree.
func TestModel_EditTagsCoercedToList(t *testing.T) {
	m := loadedAt(t, cleanTaskRepo(t), 120, 40)
	tm, _ := m.Update(press("e"))
	m = editCursorTo(t, tm.(Model), "tags")
	tm, _ = m.Update(press("enter")) // edit tags
	m = tm.(Model)
	m.edit.input.SetValue("net, ui") // a comma-list, like `task set tags=net,ui`
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("apply should return a SetFields command")
	}
	tm, _ = m.Update(cmd())
	m = tm.(Model)
	task, _, err := m.svc.ShowTask("clean")
	if err != nil {
		t.Fatal(err)
	}
	if len(task.Tags) != 2 || task.Tags[0] != "net" || task.Tags[1] != "ui" {
		t.Errorf("tags should coerce to a 2-item list, got %v", task.Tags)
	}
}

// TestModel_EditCancelNoWrite pins that Esc is a no-op: the editor closes and the
// file is untouched.
func TestModel_EditCancelNoWrite(t *testing.T) {
	root := cleanTaskRepo(t)
	path := filepath.Join(root, "tasks", "in-progress", "clean.md")
	before, _ := os.ReadFile(path)
	m := loadedAt(t, root, 120, 40)
	tm, _ := m.Update(press("e"))
	m = tm.(Model)
	tm, _ = m.Update(press("esc"))
	m = tm.(Model)
	if m.edit.active {
		t.Error("esc should close the editor")
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Error("cancel must not write")
	}
}

// TestModel_EditEpicPriorityViaMenu mirrors TestModel_EditPriorityViaMenu for an
// epic: e on the epics tab opens the SAME inline form (description/priority/tags —
// no effort/tier), and applying priority=high routes through SetEpicFields, flashes
// success, and reloads. This is the epic side of the inline `e` parity.
func TestModel_EditEpicPriorityViaMenu(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Epic("01-e.md", "---\nstatus: active\ndescription: a goal\npriority: low\ntags: [x]\n---\n# Epic\n")
	m := New(core.NewService(store.NewFS(r.Root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = cmdJump(t, tm.(Model), "epics")
	if m.cur().kind != entityEpics || m.selectedID() != "01-e" {
		t.Fatalf("setup: want the epic selected on the epics tab, got tab=%q id=%q", m.cur().name, m.selectedID())
	}

	tm, _ = m.Update(press("e"))
	m = tm.(Model)
	if !m.edit.active {
		t.Fatal("e should open the inline editor on an epic")
	}
	// The epic form must not offer effort/tier (those are task-only) or status.
	for _, f := range m.edit.fields {
		if f.key == "effort" || f.key == "tier" || f.key == "status" {
			t.Errorf("epic editor must not offer %q", f.key)
		}
	}

	m = editCursorTo(t, m, "priority")
	tm, _ = m.Update(press("enter")) // begin editing the enum (starts on current "low")
	m = enumCursorTo(t, tm.(Model), "high")
	tm, cmd := m.Update(press("enter")) // apply (write in flight; still on the field)
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("apply should return a SetEpicFields command")
	}
	tm, _ = m.Update(cmd()) // run SetEpicFields → editedMsg
	m = tm.(Model)
	if !m.edit.active || m.edit.editing {
		t.Error("a successful apply should return to the picker (open, not editing)")
	}
	if m.flash == "" || m.flashErr {
		t.Errorf("expected a success flash, got %q (err=%v)", m.flash, m.flashErr)
	}
	epic, _, _, err := m.svc.ShowEpic("01-e")
	if err != nil || epic.Priority != "high" {
		t.Errorf("priority should be high after the edit: %q (%v)", epic.Priority, err)
	}
}

// TestModel_EditNotOnAudits pins that the inline editor does NOT open on an audit
// (the entity with no field-level write — it edits the whole file via E). Tasks and
// epics edit fields in place; audits flash the E hint (see
// TestModel_EditKeyOnAudit_FlashesEditorHint).
func TestModel_EditNotOnAudits(t *testing.T) {
	m := auditsTab(t, loaded(t, 120, 40))
	tm, _ := m.Update(press("e"))
	m = tm.(Model)
	if m.edit.active {
		t.Error("e must not open the inline editor on an audit")
	}
}

// TestModel_EditFormFitsTerminal keeps the layout invariant with the editor open —
// including the tall description box: the composited view must not change height or
// overflow width at any size.
func TestModel_EditFormFitsTerminal(t *testing.T) {
	for _, d := range []struct{ w, h int }{
		{120, 40}, {100, 24}, {80, 20}, {60, 16},
	} {
		m := loadedAt(t, cleanTaskRepo(t), d.w, d.h)
		tm, _ := m.Update(press("e"))
		m = editCursorTo(t, tm.(Model), "description")
		tm, _ = m.Update(press("enter")) // open the tallest state (the wrapped box)
		m = tm.(Model)
		lines := strings.Split(m.View().Content, "\n")
		if len(lines) != d.h {
			t.Errorf("%dx%d with the editor: %d lines, want %d", d.w, d.h, len(lines), d.h)
		}
		for i, ln := range lines {
			if w := ansi.StringWidth(ln); w > d.w {
				t.Errorf("%dx%d with the editor: line %d is %d wide > %d", d.w, d.h, i, w, d.w)
			}
		}
	}
}

// TestModel_EditMenuComposites pins that the picker floats over the body showing the
// editable fields.
func TestModel_EditMenuComposites(t *testing.T) {
	m := loadedAt(t, cleanTaskRepo(t), 120, 40)
	tm, _ := m.Update(press("e"))
	m = tm.(Model)
	v := ansi.Strip(m.View().Content)
	for _, want := range []string{"edit clean", "priority", "description", "tags", "tier"} {
		if !strings.Contains(v, want) {
			t.Errorf("edit picker should show %q:\n%s", want, v)
		}
	}
}

// TestEditableFields_RevisitOnlyWhenDeferred pins that the snooze date is editable
// in place ONLY for deferred tasks (the invariant: revisit_at ⟺ deferred).
func TestEditableFields_RevisitOnlyWhenDeferred(t *testing.T) {
	has := func(tk domain.Task) bool {
		for _, f := range editableFields(tk) {
			if f.key == "revisit_at" {
				return true
			}
		}
		return false
	}
	if !has(domain.Task{Status: domain.StatusDeferred}) {
		t.Error("a deferred task's edit form should include revisit_at")
	}
	if has(domain.Task{Status: domain.StatusInProgress}) {
		t.Error("a non-deferred task's edit form must NOT include revisit_at")
	}
}

// selectOverdue moves the cursor to the deferred 'overdue' task (row after the
// active alpha) in a revisitRepo working view.
func selectOverdue(t *testing.T, m Model) Model {
	t.Helper()
	tm, _ := m.Update(press("j"))
	m = tm.(Model)
	if m.selectedID() != "overdue" {
		t.Fatalf("setup: want overdue selected, got %q", m.selectedID())
	}
	return m
}

// clearInput backspaces the focused text input empty (prefill is ≤10 chars).
func clearInput(t *testing.T, m Model) Model {
	t.Helper()
	for i := 0; i < 16; i++ {
		tm, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
		m = tm.(Model)
	}
	return m
}

// TestModel_EditRevisitDate pins re-editing a deferred task's snooze in place:
// `e` → revisit → type a new absolute date → it persists.
func TestModel_EditRevisitDate(t *testing.T) {
	m := loadedAt(t, revisitRepo(t), 120, 40)
	m = selectOverdue(t, m)
	tm, _ := m.Update(press("e"))
	m = editCursorTo(t, tm.(Model), "revisit_at")
	tm, _ = m.Update(press("enter")) // begin editing (prefilled 2020-01-01)
	m = clearInput(t, tm.(Model))
	for _, r := range "2030-05-05" {
		tm, _ = m.Update(press(string(r)))
		m = tm.(Model)
	}
	tm, cmd := m.Update(press("enter")) // submit
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("submit should return a SetFields command")
	}
	m.Update(cmd()) // run the write
	if task, _, err := m.svc.ShowTask("overdue"); err != nil || task.RevisitAt != "2030-05-05" {
		t.Errorf("revisit_at should update to 2030-05-05, got %q (%v)", task.RevisitAt, err)
	}
}

// TestModel_EditRevisitDateBlankClears pins that blanking the snooze field clears
// it (unset), parking the task indefinitely again.
func TestModel_EditRevisitDateBlankClears(t *testing.T) {
	m := loadedAt(t, revisitRepo(t), 120, 40)
	m = selectOverdue(t, m)
	tm, _ := m.Update(press("e"))
	m = editCursorTo(t, tm.(Model), "revisit_at")
	tm, _ = m.Update(press("enter"))
	m = clearInput(t, tm.(Model))
	tm, cmd := m.Update(press("enter")) // submit blank → clear
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("a blank revisit submit should return an unset command")
	}
	m.Update(cmd())
	if task, _, _ := m.svc.ShowTask("overdue"); task.RevisitAt != "" {
		t.Errorf("blanking revisit should clear it, got %q", task.RevisitAt)
	}
}

// headingCol returns the column of text in the first view line that contains it.
func headingCol(view, text string) int {
	for _, line := range strings.Split(view, "\n") {
		if i := strings.Index(line, text); i >= 0 {
			return i
		}
	}
	return -1
}

// TestModel_EditBoxStaysPutOnError pins the layout fix: a validation error renders
// inside the box, so the box doesn't shift horizontally when the error appears
// (same editing state + input before/after; only the error toggles).
func TestModel_EditBoxStaysPutOnError(t *testing.T) {
	m := loadedAt(t, revisitRepo(t), 120, 40)
	m = selectOverdue(t, m)
	tm, _ := m.Update(press("e"))
	m = editCursorTo(t, tm.(Model), "revisit_at")
	tm, _ = m.Update(press("enter")) // begin editing
	m = clearInput(t, tm.(Model))
	for _, r := range "bad" { // an invalid date, not yet submitted (no error)
		tm, _ = m.Update(press(string(r)))
		m = tm.(Model)
	}
	before := headingCol(ansi.Strip(m.View().Content), "edit overdue")

	tm, _ = m.Update(press("enter")) // submit invalid → validation error, field stays open
	m = tm.(Model)
	if m.edit.err == "" {
		t.Fatal("expected a validation error for an invalid date")
	}
	after := headingCol(ansi.Strip(m.View().Content), "edit overdue")

	if before < 0 || before != after {
		t.Errorf("edit box shifted when the error appeared: heading col %d → %d", before, after)
	}
}
