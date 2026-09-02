package domain

import (
	"strings"
	"testing"
)

// nestedFenceAudit is the shape that swallowed two findings in
// 2026-09-01-cli-ergonomics-from-an-agent-session: a four-backtick fence wrapping a
// three-backtick example. Valid CommonMark, and the rendered document looks right —
// which is why the failure went unnoticed at every surface an author would check.
// The inner example is an OPENING fence shown on its own — the shape an author
// writes to document fence syntax itself. That makes the count of ``` runs odd, so
// a pair-them-up regex pairs the outer closer with the NEXT finding's fence and
// masks every heading in between.
const nestedFenceAudit = "# Audit: x\n\n## Findings\n\n" +
	"#### H1. first · **Status:** open\n\n" +
	"An example wrapped in a longer fence:\n\n" +
	"````\n" +
	"```mermaid\n" +
	"````\n\n" +
	"#### H2. second · **Status:** open\n\n" +
	"```\n" +
	"code\n" +
	"```\n\n" +
	"#### H3. third · **Status:** open\n"

// TestParseFindings_NestedFenceDoesNotSwallowLaterFindings is the regression for
// H4. A swallowed finding can never be stamped — `audit finding H3 --status fixed`
// fails permanently — and the progress bar reports every remaining finding settled,
// so an audit silently under-reports its own scope.
func TestParseFindings_NestedFenceDoesNotSwallowLaterFindings(t *testing.T) {
	got := ParseFindings(nestedFenceAudit)

	if len(got) != 3 {
		codes := make([]string, len(got))
		for i, f := range got {
			codes[i] = f.Code
		}
		t.Fatalf("parsed %d findings %v, want 3 (H1, H2, H3)", len(got), codes)
	}
	for i, want := range []string{"H1", "H2", "H3"} {
		if got[i].Code != want {
			t.Errorf("finding %d = %q, want %q", i, got[i].Code, want)
		}
	}
}

// The point of masking is that example syntax inside a fence is NOT structure.
// A `#### H9.` inside the nested block must stay invisible.
func TestParseFindings_FencedExampleIsNotAFinding(t *testing.T) {
	body := "# Audit: x\n\n## Findings\n\n" +
		"#### H1. real · **Status:** open\n\n" +
		"````\n" +
		"```\n" +
		"#### H9. not a finding · **Status:** open\n" +
		"```\n" +
		"````\n\n" +
		"#### H2. also real · **Status:** open\n"

	got := ParseFindings(body)

	for _, f := range got {
		if f.Code == "H9" {
			t.Errorf("a fenced example was parsed as a real finding: %+v", f)
		}
	}
	if len(got) != 2 {
		t.Errorf("parsed %d findings, want 2", len(got))
	}
}

// Tilde fences are equally valid CommonMark and appear in bodies that need to show
// backtick examples. The old regex only knew backticks.
func TestParseFindings_TildeFenceMasksLikeBackticks(t *testing.T) {
	body := "# Audit: x\n\n## Findings\n\n" +
		"#### H1. real · **Status:** open\n\n" +
		"~~~\n" +
		"#### H9. not a finding · **Status:** open\n" +
		"~~~\n\n" +
		"#### H2. also real · **Status:** open\n"

	got := ParseFindings(body)

	for _, f := range got {
		if f.Code == "H9" {
			t.Errorf("a tilde-fenced example was parsed as a real finding: %+v", f)
		}
	}
	if len(got) != 2 {
		t.Errorf("parsed %d findings, want 2", len(got))
	}
}

// blankFences must preserve every byte offset, because finding spans are computed
// against the masked text and then used to write into the ORIGINAL body. A length
// change here lands a status write somewhere else entirely.
func TestBlankFences_PreservesLengthAndLineStructure(t *testing.T) {
	for _, body := range []string{
		nestedFenceAudit,
		"a\n```\nb\n```\nc\n",
		"a\n~~~\nb\n~~~\nc\n",
		"unterminated:\n```\nb\nc\n",
		"",
		"no fences here\n",
	} {
		masked := blankFences(body)
		if len(masked) != len(body) {
			t.Errorf("length changed %d -> %d for %q", len(body), len(masked), body)
		}
		if strings.Count(masked, "\n") != strings.Count(body, "\n") {
			t.Errorf("line count changed for %q", body)
		}
	}
}
