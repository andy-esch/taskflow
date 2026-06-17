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
