package tui

import (
	"errors"
	"testing"
)

// TestOpenInEditor_NoSelection: with nothing selected (an unloaded/empty list),
// `E` flashes an error rather than launching an editor on an empty path.
func TestOpenInEditor_NoSelection(t *testing.T) {
	m := newModel(t) // unloaded → empty list → no selection
	tm, cmd := m.openInEditor()
	mm := tm.(Model)
	if cmd != nil {
		t.Errorf("no selection should not launch an editor; got a cmd")
	}
	if !mm.flashErr || mm.flash != "nothing to edit" {
		t.Errorf("flash = %q (err=%v), want %q (err)", mm.flash, mm.flashErr, "nothing to edit")
	}
}

// TestOpenInEditor_Key: pressing `E` on a real selection returns the (non-nil)
// suspend-and-run command and leaves no error flash. The cmd is NOT invoked — it
// would launch $EDITOR — so this asserts the wiring, not the spawn.
func TestOpenInEditor_Key(t *testing.T) {
	m := loaded(t, 120, 40)
	if m.selectedPath() == "" {
		t.Fatal("setup: expected a selected task with a path")
	}
	tm, cmd := m.Update(press("E"))
	mm := tm.(Model)
	if cmd == nil {
		t.Fatal("E on a selection should return an ExecProcess cmd")
	}
	if mm.flash != "" {
		t.Errorf("a successful launch should not flash; got %q", mm.flash)
	}
}

// TestEditorClosed_Success_Reloads: returning from the editor (nil err) reloads
// so any change shows, and clears any prior flash.
func TestEditorClosed_Success_Reloads(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, cmd := m.Update(editorClosedMsg{})
	mm := tm.(Model)
	if cmd == nil {
		t.Error("a clean editor close should fire a reload")
	}
	if mm.flashErr {
		t.Errorf("a clean editor close should not flash an error; got %q", mm.flash)
	}
}

// TestEditorClosed_Error_Flashes: a launch failure surfaces as an error flash and
// does not reload.
func TestEditorClosed_Error_Flashes(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, cmd := m.Update(editorClosedMsg{err: errors.New("boom")})
	mm := tm.(Model)
	if cmd != nil {
		t.Error("a failed editor launch should not reload")
	}
	if !mm.flashErr || mm.flash == "" {
		t.Errorf("a failed editor launch should flash an error; got %q (err=%v)", mm.flash, mm.flashErr)
	}
}
