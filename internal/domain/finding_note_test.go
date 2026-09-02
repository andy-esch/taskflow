package domain

import (
	"strings"
	"testing"
)

// strayLabelAudit is the shape `audit new` scaffolds and authors leave behind: a
// `**Resolution:**` label with no paragraph, sitting at the end of the finding.
const strayLabelAudit = "# Audit: x\n\n## Findings\n\n" +
	"#### H1. first · **Status:** open\n\n" +
	"Some prose.\n\n" +
	"**Resolution:**\n\n" +
	"#### H2. second · **Status:** open\n\n" +
	"More prose.\n"

// TestSetFindingNote_FillsAStrayLabelInsteadOfAppending is the regression for H7.
// Appending beside the stray label produced TWO labels, and `audit lint` then
// reported "only the first is read" — meaning the note just written is the one
// being ignored, while the empty placeholder wins. The command reported success.
func TestSetFindingNote_FillsAStrayLabelInsteadOfAppending(t *testing.T) {
	out, err := SetFindingNote(strayLabelAudit, "H1", "fixed by rewriting the scanner")
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(out, FindingNoteLabel); n != 1 {
		t.Errorf("want exactly one resolution label, got %d:\n%s", n, out)
	}
	if !strings.Contains(out, "fixed by rewriting the scanner") {
		t.Errorf("the note should have been written:\n%s", out)
	}

	// The parse must agree: the note is readable and lint is quiet about it.
	var h1 Finding
	for _, f := range ParseFindings(out) {
		if f.Code == "H1" {
			h1 = f
		}
	}
	if h1.Note != "fixed by rewriting the scanner" {
		t.Errorf("the written note should parse back, got %q", h1.Note)
	}
	if h1.NoteLabels != 1 {
		t.Errorf("NoteLabels = %d, want 1", h1.NoteLabels)
	}
	for _, issue := range LintFindings("open", ParseFindings(out)) {
		if strings.Contains(issue.Message, "Resolution") {
			t.Errorf("a filled label should not lint: %s", issue.Message)
		}
	}
}

// The neighbouring finding must be untouched — the fill replaces one label, not
// the region between findings.
func TestSetFindingNote_FillingOneLabelLeavesTheNextFindingAlone(t *testing.T) {
	out, err := SetFindingNote(strayLabelAudit, "H1", "done")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "#### H2. second · **Status:** open") {
		t.Errorf("H2's header was disturbed:\n%s", out)
	}
	if !strings.Contains(out, "More prose.") {
		t.Errorf("H2's body was disturbed:\n%s", out)
	}
}

// A label with prose BELOW it is a different mistake: the paragraph belongs on the
// label's own line. Replacing the label alone would delete that prose, so this case
// keeps the old behaviour and lint names the stray label instead.
func TestSetFindingNote_DoesNotSwallowProseBelowAStrayLabel(t *testing.T) {
	body := "# Audit: x\n\n## Findings\n\n" +
		"#### H1. first · **Status:** open\n\n" +
		"**Resolution:**\n\n" +
		"the paragraph the author meant as the note\n"

	out, err := SetFindingNote(body, "H1", "replacement")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "the paragraph the author meant as the note") {
		t.Errorf("prose below the label must not be deleted:\n%s", out)
	}
}

// Re-noting a filled label still replaces rather than stacking.
func TestSetFindingNote_ReplacingAFilledLabelIsUnchanged(t *testing.T) {
	once, err := SetFindingNote(strayLabelAudit, "H1", "first note")
	if err != nil {
		t.Fatal(err)
	}
	twice, err := SetFindingNote(once, "H1", "second note")
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(twice, FindingNoteLabel); n != 1 {
		t.Errorf("re-noting should replace, got %d labels:\n%s", n, twice)
	}
	if strings.Contains(twice, "first note") {
		t.Errorf("the old note should be gone:\n%s", twice)
	}
}

// Clearing a note must also clear a stray label, so the audit ends up neither
// carrying a placeholder nor a note.
func TestSetFindingNote_EmptyNoteRemovesAStrayLabel(t *testing.T) {
	out, err := SetFindingNote(strayLabelAudit, "H1", "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(out, FindingNoteLabel) != 0 {
		t.Errorf("the stray label should be gone:\n%s", out)
	}
	if !strings.Contains(out, "#### H2. second") {
		t.Errorf("H2 was disturbed:\n%s", out)
	}
}
