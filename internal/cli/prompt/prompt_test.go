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
