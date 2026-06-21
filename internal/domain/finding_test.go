package domain

import "testing"

func TestParseFindings(t *testing.T) {
	body := "# Audit\n\n" +
		"#### H1. retry waste  · **Status:** open\n\n" +
		"**File:** dispatcher/x.go:42 | **Component:** dispatcher / retry\n" +
		"**Effort:** S · **Urgency:** soon\n\n" +
		"why it matters\n\n" +
		"#### M2. other thing  · **Status:** fixed 2026-01-01 (PR #9)\n\n" +
		"```\n#### S9. fenced example  · **Status:** open\n```\n\n" +
		"#### L3. later\n\n**Status:** open-ish\n"

	fs := ParseFindings(body)
	if len(fs) != 3 {
		t.Fatalf("want 3 findings (the fenced S9 excluded), got %d: %+v", len(fs), fs)
	}

	h1 := fs[0]
	if h1.Code != "H1" || h1.Title != "retry waste" || h1.Status != "open" {
		t.Errorf("H1 header/status wrong: %+v", h1)
	}
	if h1.File != "dispatcher/x.go:42" || h1.Component != "dispatcher / retry" ||
		h1.Effort != "S" || h1.Urgency != "soon" {
		t.Errorf("H1 metadata wrong: %+v", h1)
	}
	// Status keeps only the first token, dropping the date/PR tail.
	if fs[1].Code != "M2" || fs[1].Status != "fixed" {
		t.Errorf("M2 status = %q, want fixed", fs[1].Status)
	}
	// Status on its own line; "open-ish" must NOT read as "open".
	if fs[2].Code != "L3" || fs[2].Title != "later" || fs[2].Status != "open-ish" {
		t.Errorf("L3 parsed wrong: %+v", fs[2])
	}
	if got := CountOpenFindings(fs); got != 1 {
		t.Errorf("open count = %d, want 1 (only H1; open-ish and fenced excluded)", got)
	}
}

func TestParseFindings_Empty(t *testing.T) {
	if fs := ParseFindings("# Audit\n\nno findings yet\n"); len(fs) != 0 {
		t.Errorf("want no findings, got %+v", fs)
	}
}

// TestParseFindings_LiteralStatusInTitle pins the fix the 2026-06-17 self-review
// surfaced: a finding whose TITLE contains a literal `**Status:**` must not have
// that mistaken for its status, and the title must survive intact. The marker is
// authoritative only at line start or after the header's `· ` separator.
func TestParseFindings_LiteralStatusInTitle(t *testing.T) {
	body := "#### X1. parser takes the first `**Status:**` token  · **Status:** open\n\nbody\n"
	fs := ParseFindings(body)
	if len(fs) != 1 {
		t.Fatalf("want 1 finding, got %d: %+v", len(fs), fs)
	}
	if fs[0].Status != "open" {
		t.Errorf("status = %q, want open (the literal **Status:** in the title must not win)", fs[0].Status)
	}
	if fs[0].Title != "parser takes the first `**Status:**` token" {
		t.Errorf("title = %q, want it kept intact (incl. the literal marker)", fs[0].Title)
	}
}

func TestLintFindings(t *testing.T) {
	// Clean: open bucket, legal statuses → no issues.
	if iss := LintFindings("open", []Finding{{Code: "S1", Status: "open"}, {Code: "H1", Status: "fixed"}}); len(iss) != 0 {
		t.Errorf("clean findings should lint clean, got %v", iss)
	}
	// Typo'd status → one issue on the finding code.
	if iss := LintFindings("open", []Finding{{Code: "S1", Status: "opne"}}); len(iss) != 1 || iss[0].Field != "S1" {
		t.Errorf("typo status should be one issue on S1, got %v", iss)
	}
	// Missing status → flagged.
	if iss := LintFindings("open", []Finding{{Code: "M2", Status: ""}}); len(iss) != 1 || iss[0].Field != "M2" {
		t.Errorf("missing status should be flagged on M2, got %v", iss)
	}
	// bucket↔state: a closed audit with a still-open finding → bucket issue.
	if iss := LintFindings("closed", []Finding{{Code: "S1", Status: "open"}}); len(iss) != 1 || iss[0].Field != "bucket" {
		t.Errorf("closed audit with an open finding should flag bucket, got %v", iss)
	}
	// Vocabulary is case-insensitive and covers the full set.
	if iss := LintFindings("open", []Finding{{Code: "S1", Status: "IN-PROGRESS"}, {Code: "H1", Status: "landed"}}); len(iss) != 0 {
		t.Errorf("legal statuses (case-insensitive) should pass, got %v", iss)
	}
}
