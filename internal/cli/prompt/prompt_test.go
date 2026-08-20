package prompt

import "testing"

func TestGate(t *testing.T) {
	if !NewGate(true).On() {
		t.Error("NewGate(true) should be On")
	}
	if NewGate(false).On() {
		t.Error("NewGate(false) should be off")
	}
}

func TestFake_ScriptsAndRecords(t *testing.T) {
	f := &Fake{SelectAnswers: []string{"e1"}, TextAnswers: []string{"hello"}}

	if got, err := f.SelectOne("pick epic", []Option{{Label: "E1", Value: "e1"}}); err != nil || got != "e1" {
		t.Fatalf("SelectOne = %q, %v; want e1", got, err)
	}
	if got, err := f.Text("title", ""); err != nil || got != "hello" {
		t.Fatalf("Text = %q, %v; want hello", got, err)
	}
	if len(f.Asked) != 2 || f.Asked[0] != "pick epic" || f.Asked[1] != "title" {
		t.Errorf("Asked = %v; want [pick epic title]", f.Asked)
	}
	// Exhausted queue errors rather than silently returning "" (a test-author bug).
	if _, err := f.SelectOne("again", nil); err == nil {
		t.Error("exhausted SelectOne should error")
	}
}

// TestSelectLayout pins the two defaults that made a two-option menu render as a filter
// line plus a single visible choice.
//
// huh's Height counts the TITLE as well as the rows, so the old Height(len(options)) left
// room for one fewer row than there were options — hiding the last one. And filtering was
// on unconditionally, putting a free-text input above a list you could simply arrow to.
func TestSelectLayout(t *testing.T) {
	cases := []struct {
		n          int
		wantH      int
		wantFilter bool
		why        string
	}{
		{2, 0, false, "init's here-vs-elsewhere: no cap, no filter — every option visible"},
		{3, 0, false, "the placement menu: same"},
		{7, 0, false, "still scannable at a glance"},
		{8, 0, true, "long enough that filtering beats scanning"},
		{10, 0, true, "at the cap: still no scrolling needed"},
		{11, selectMenuMax + 1, true, "past the cap: scroll, and reserve a line for the title"},
		{50, selectMenuMax + 1, true, "an epic picker stays a compact window"},
	}
	for _, c := range cases {
		h, f := selectLayout(c.n)
		if h != c.wantH || f != c.wantFilter {
			t.Errorf("selectLayout(%d) = (%d, %v), want (%d, %v) — %s",
				c.n, h, f, c.wantH, c.wantFilter, c.why)
		}
	}
}

// TestSelectLayout_HeightNeverHidesAnOption is the invariant the old code broke: whenever a
// height IS set, it must leave room for the full window plus the title, never fewer rows
// than the menu promises.
func TestSelectLayout_HeightNeverHidesAnOption(t *testing.T) {
	for n := 1; n <= 60; n++ {
		h, _ := selectLayout(n)
		if h == 0 {
			continue // huh sizes to the options itself
		}
		if rows := h - 1; rows < selectMenuMax {
			t.Errorf("selectLayout(%d) height %d leaves %d rows, fewer than the %d-row window",
				n, h, rows, selectMenuMax)
		}
	}
}

// TestAbortKeyMap_BindsEsc pins the cancel affordance. huh binds Quit to ctrl+c alone, so
// backing out of a prompt required an interrupt — which reads as "something broke" for what
// is just changing your mind. ctrl+c is kept alongside so muscle memory still works.
func TestAbortKeyMap_BindsEsc(t *testing.T) {
	keys := abortKeyMap().Quit.Keys()
	has := func(k string) bool {
		for _, got := range keys {
			if got == k {
				return true
			}
		}
		return false
	}
	if !has("esc") {
		t.Errorf("esc must cancel a prompt, got %v", keys)
	}
	if !has("ctrl+c") {
		t.Errorf("ctrl+c must keep working, got %v", keys)
	}
}
