package tui

import "testing"

// TestToggleFilterMode: F flips the list filter fuzzy ⇄ substring, session-wide
// across every tab, with the active mode reflected in each list's filter prompt;
// the default is fuzzy.
func TestToggleFilterMode(t *testing.T) {
	m := loaded(t, 120, 40)

	// Default: fuzzy everywhere.
	for _, tab := range m.tabs {
		if tab.filterExact {
			t.Fatalf("default filter mode should be fuzzy, tab %q is exact", tab.name)
		}
		if got := tab.list.FilterInput.Prompt; got != "filter (fuzzy): " {
			t.Errorf("tab %q prompt = %q, want fuzzy", tab.name, got)
		}
	}

	// F → substring, applied to ALL tabs (session-wide), prompt updated.
	tm, _ := m.Update(press("F"))
	m = tm.(Model)
	for _, tab := range m.tabs {
		if !tab.filterExact {
			t.Errorf("after F, tab %q should be exact", tab.name)
		}
		if got := tab.list.FilterInput.Prompt; got != "filter (exact): " {
			t.Errorf("tab %q prompt = %q, want exact", tab.name, got)
		}
	}

	// F again → back to fuzzy (the default).
	tm, _ = m.Update(press("F"))
	m = tm.(Model)
	if m.cur().filterExact {
		t.Error("a second F should return to fuzzy")
	}
}
