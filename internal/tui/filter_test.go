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

// TestToggleFilterMode_SwapsMatcher proves the toggle actually swaps the live
// matcher, not just a cosmetic flag: "tat" is a fuzzy subsequence of the seed
// descriptions ("the alpha/beta task") but a substring of neither, so the same
// query yields different visible items per mode.
func TestToggleFilterMode_SwapsMatcher(t *testing.T) {
	m := loaded(t, 120, 40)

	// Fuzzy (default): "tat" matches both as a scattered subsequence.
	m.cur().list.SetFilterText("tat")
	fuzzyN := len(m.cur().list.VisibleItems())
	if fuzzyN == 0 {
		t.Fatalf("fuzzy should match 'tat' as a subsequence, got %d visible", fuzzyN)
	}

	// F → substring: the toggle re-ranks live, and "tat" is a substring of neither,
	// so the visible set must shrink (proving Substring is actually installed).
	tm, _ := m.toggleFilterMode()
	m = tm.(Model)
	if !m.cur().filterExact {
		t.Fatal("toggle should switch to exact")
	}
	if exactN := len(m.cur().list.VisibleItems()); exactN >= fuzzyN {
		t.Errorf("substring 'tat' should match fewer than fuzzy (%d) — got %d; matcher not swapped?", fuzzyN, exactN)
	}
}
