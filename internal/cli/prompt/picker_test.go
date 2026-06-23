package prompt

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
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
	fm := drive(twoOpts, tea.KeyPressMsg{Code: tea.KeyEnter})
	if fm.choice != "a" || fm.aborted {
		t.Errorf("enter should select the first item: choice=%q aborted=%v", fm.choice, fm.aborted)
	}
}

func TestPicker_DownThenEnter(t *testing.T) {
	fm := drive(twoOpts, tea.KeyPressMsg{Code: tea.KeyDown}, tea.KeyPressMsg{Code: tea.KeyEnter})
	if fm.choice != "b" {
		t.Errorf("down then enter should select the second item, got %q", fm.choice)
	}
}

func TestPicker_EscAborts(t *testing.T) {
	fm := drive(twoOpts, tea.KeyPressMsg{Code: tea.KeyEsc})
	if !fm.aborted || fm.choice != "" {
		t.Errorf("esc on an unfiltered list should abort: aborted=%v choice=%q", fm.aborted, fm.choice)
	}
}

func TestPicker_CtrlCAborts(t *testing.T) {
	fm := drive(twoOpts, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if !fm.aborted {
		t.Error("ctrl-c should abort")
	}
}

// TestPicker_EscIsFilterAware pins the bug-prone branch: while filtering, esc is
// the list's (cancel the filter), NOT an abort; only esc on an unfiltered list
// aborts.
func TestPicker_EscIsFilterAware(t *testing.T) {
	slash := tea.KeyPressMsg{Code: '/', Text: "/"}
	esc := tea.KeyPressMsg{Code: tea.KeyEsc}

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
func TestListItem_FilterValueIsLabel(t *testing.T) {
	it := listItem{label: "04-multiuser-13 · admin tooling", value: "04-multiuser-13"}
	if it.FilterValue() != it.Title() || it.FilterValue() != "04-multiuser-13 · admin tooling" {
		t.Errorf("FilterValue should be the label so the fuzzy filter matches visible text")
	}
}
