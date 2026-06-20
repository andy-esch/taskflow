package prompt

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// drive feeds messages through the model's Update (no tea program / TTY needed)
// and returns the final picker state — so the picker's key handling is testable.
func drive(opts []Option, msgs ...tea.Msg) pickerModel {
	var m tea.Model = newPicker("Pick", opts)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24}) // size the list first
	for _, msg := range msgs {
		m, _ = m.Update(msg)
	}
	return m.(pickerModel)
}

var twoOpts = []Option{{Label: "Alpha", Value: "a"}, {Label: "Beta", Value: "b"}}

func TestPicker_EnterSelectsHighlighted(t *testing.T) {
	fm := drive(twoOpts, tea.KeyMsg{Type: tea.KeyEnter})
	if fm.choice != "a" || fm.aborted {
		t.Errorf("enter should select the first item: choice=%q aborted=%v", fm.choice, fm.aborted)
	}
}

func TestPicker_DownThenEnter(t *testing.T) {
	fm := drive(twoOpts, tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter})
	if fm.choice != "b" {
		t.Errorf("down then enter should select the second item, got %q", fm.choice)
	}
}

func TestPicker_EscAborts(t *testing.T) {
	fm := drive(twoOpts, tea.KeyMsg{Type: tea.KeyEsc})
	if !fm.aborted || fm.choice != "" {
		t.Errorf("esc on an unfiltered list should abort: aborted=%v choice=%q", fm.aborted, fm.choice)
	}
}

func TestPicker_CtrlCAborts(t *testing.T) {
	fm := drive(twoOpts, tea.KeyMsg{Type: tea.KeyCtrlC})
	if !fm.aborted {
		t.Error("ctrl-c should abort")
	}
}

// TestPicker_EscIsFilterAware pins the bug-prone branch: while filtering, esc is
// the list's (cancel the filter), NOT an abort; only esc on an unfiltered list
// aborts.
func TestPicker_EscIsFilterAware(t *testing.T) {
	slash := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	esc := tea.KeyMsg{Type: tea.KeyEsc}

	if fm := drive(twoOpts, slash, esc); fm.aborted {
		t.Error("esc while filtering should cancel the filter, not abort")
	}
	if fm := drive(twoOpts, slash, esc, esc); !fm.aborted {
		t.Error("a second esc (now unfiltered) should abort")
	}
}

// TestPickerErr pins the SIGINT→130 mapping: a signal interrupt becomes ErrAborted
// (so it exits 130 with a quiet "aborted"); other errors pass through visibly.
func TestPickerErr(t *testing.T) {
	if pickerErr(nil) != nil {
		t.Error("nil should pass through")
	}
	if !errors.Is(pickerErr(tea.ErrInterrupted), ErrAborted) {
		t.Error("a SIGINT interrupt should map to ErrAborted")
	}
	boom := errors.New("boom")
	if !errors.Is(pickerErr(boom), boom) {
		t.Error("other errors should pass through (stay visible)")
	}
}

// TestSubstringFilter pins the fix: substring matching, NOT fuzzy. The slack slug
// fuzzy-matches "multiuser" (m-u-l-t-i-u-s-e-r scattered) but must not match here.
func TestSubstringFilter(t *testing.T) {
	targets := []string{
		"04-multiuser-10-frontend-backfill",
		"04-multiuser-13-internal-admin",
		"move-slack-webhook-url-secret-from-github-to-infisical", // fuzzy-matches, substring must not
	}
	if got := substringFilter("multiuser", targets); len(got) != 2 ||
		got[0].Index != 0 || got[1].Index != 1 {
		t.Errorf("substring 'multiuser' should match exactly the 2 multiuser items in order, got %+v", got)
	}
	if len(substringFilter("MULTI", targets)) != 2 {
		t.Error("filter should be case-insensitive")
	}
	if len(substringFilter("", targets)) != 3 {
		t.Error("empty term should match all (in order)")
	}
	if len(substringFilter("nope", targets)) != 0 {
		t.Error("no substring match should yield nothing")
	}
}

// TestSubstringFilter_Unicode covers the rune-offset path and the lengthening
// ToLower case that previously panicked.
func TestSubstringFilter_Unicode(t *testing.T) {
	got := substringFilter("multi", []string{"café-multiuser", "plain"})
	if len(got) != 1 || got[0].Index != 0 {
		t.Fatalf("should match only the café label, got %+v", got)
	}
	if len(got[0].MatchedIndexes) == 0 || got[0].MatchedIndexes[0] != 5 {
		t.Errorf("match should start at rune 5 (after 'café-'), got %v", got[0].MatchedIndexes)
	}
	// Ⱥ→ⱥ grows in ToLower; this must not panic (it did before the fix).
	_ = substringFilter("x", []string{"ȺȺȺx"})
}

func TestListItem_FilterValueIsLabel(t *testing.T) {
	it := listItem{label: "04-multiuser-13 · admin tooling", value: "04-multiuser-13"}
	if it.FilterValue() != it.Title() || it.FilterValue() != "04-multiuser-13 · admin tooling" {
		t.Errorf("FilterValue should be the label so the fuzzy filter matches visible text")
	}
}
