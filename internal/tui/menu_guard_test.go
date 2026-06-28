package tui

import "testing"

// TestActionMenu_EmptyDoesNotOpen guards the latent index panic (audit L3): a
// transition table that filters to nothing must leave the menu inactive, so
// selected() is never reached on an empty slice.
func TestActionMenu_EmptyDoesNotOpen(t *testing.T) {
	var a actionMenu
	a.open("slug", []transition{{"stay", "here", false}}, "here") // only verb lands on cur → filtered out
	if a.active {
		t.Error("an action menu with no valid transitions must not open")
	}
	var b actionMenu
	b.open("slug", nil, "here")
	if b.active {
		t.Error("a nil transition table must not open the menu")
	}
}

// TestEditMenu_EmptyFieldsDoesNotOpen guards the latent index panic (audit L3): a
// form with no editable fields must stay inactive, so cur() never indexes nil.
func TestEditMenu_EmptyFieldsDoesNotOpen(t *testing.T) {
	if m := newEditMenu("slug", nil, nil); m.active {
		t.Error("a form with no editable fields must not open")
	}
}
