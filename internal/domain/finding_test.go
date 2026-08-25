package domain

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

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
	if iss := LintFindings("open", []Finding{{Code: "S1", Status: "IN-PROGRESS"}, {Code: "H1", Status: "tracked", StatusDecoration: "by 6fbwhsw024nr"}}); len(iss) != 0 {
		t.Errorf("legal statuses (case-insensitive) should pass, got %v", iss)
	}
}

func TestParseFindings_EmptyStatusBeforeBoldLabel(t *testing.T) {
	// `**Status:** **Effort:** S` — an empty status immediately before another bold
	// label must parse as "" (so lint flags it), not grab "**Effort:**" as garbage.
	fs := ParseFindings("#### S1. t\n**Status:** **Effort:** S\n")
	if len(fs) != 1 || fs[0].Status != "" {
		t.Errorf("empty status before a bold label should be \"\", got %q", fs[0].Status)
	}
}

// TestTallyFindings pins the status→disposition mapping the segmented bar bands
// by, case-insensitively, with unrecognized/missing statuses counting toward none.
func TestTallyFindings(t *testing.T) {
	fs := []Finding{
		{Status: "open"}, {Status: "open"},
		{Status: "in-progress"},
		// `tracked` counts as done for the AUDIT: the finding was transferred to a task and
		// is no longer this document's business — the distinction it exists to draw.
		{Status: "fixed"}, {Status: "tracked"}, {Status: "FIXED"}, // case-insensitive → done
		{Status: "deferred"}, {Status: "superseded"}, {Status: "wontfix"},
		{Status: "bogus"}, {Status: ""}, // unrecognized / missing → counted toward none
	}
	got := TallyFindings(fs)
	want := FindingTally{Open: 2, Active: 1, Done: 3, Dropped: 3}
	if got != want {
		t.Errorf("TallyFindings = %+v, want %+v", got, want)
	}
}

// The validated write path finding H1 asked for. Surgical by construction: the span located
// at parse time is what gets replaced, so prose elsewhere containing the same word — and
// every other finding — is untouched without any care being taken.
func TestSetFindingStatus(t *testing.T) {
	body := "## Findings\n\n#### H1. A thing  · **Status:** open\n\n" +
		"**File:** a.go | **Component:** x\n\nProse that mentions open elsewhere.\n\n" +
		"#### M1. Another\n\n**Status:** open\n\nmore\n"

	out, err := SetFindingStatus(body, "H1", "fixed 2026-08-24 (PR #12)")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "**Status:** fixed 2026-08-24 (PR #12)") {
		t.Errorf("status not stamped:\n%s", out)
	}
	if !strings.Contains(out, "Prose that mentions open elsewhere.") {
		t.Error("prose containing the status word was rewritten")
	}
	if strings.Count(out, "**Status:** open") != 1 {
		t.Error("the other finding's status was touched")
	}

	// Re-stamping REPLACES the decorated value rather than appending beside it.
	out, err = SetFindingStatus(out, "H1", "deferred (see ADR-0003)")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "2026-08-24") {
		t.Errorf("the previous decoration survived a re-stamp:\n%s", out)
	}
	// Decoration keeps its case: it carries links and document names.
	if !strings.Contains(out, "(see ADR-0003)") {
		t.Errorf("decoration was case-flattened:\n%s", out)
	}
	// …while the token is normalised, so it parses back as vocabulary.
	if got := ParseFindings(out)[0].Status; got != "deferred" {
		t.Errorf("status parses back as %q, want deferred", got)
	}

	if _, err := SetFindingStatus(body, "H1", "bogus"); !errors.Is(err, ErrValidation) {
		t.Errorf("an unknown status was accepted: %v", err)
	}
	if _, err := SetFindingStatus(body, "Z9", "fixed"); !errors.Is(err, ErrNotFound) {
		t.Errorf("a missing code should be ErrNotFound, got %v", err)
	}
}

// Audit 2026-08-24-planning-state-vocabulary H1. Fence stripping used to DELETE the fenced
// text before computing spans, so every offset after a fence was short by its length and a
// status write landed somewhere else — on the scaffold `audit new` emits, `## Candidate
// tasks` became `## Candidafixedasks` while the finding's status went unchanged, and the
// command still reported success. Blanking preserves offsets, so spans stay valid indices
// into the body they were computed from.
func TestSetFindingStatusIsOffsetSafeAcrossFences(t *testing.T) {
	body := "## Findings\n\n" +
		"<!-- example, un-fence it: -->\n\n" +
		"```\n#### H9. <title>  · **Status:** open\n\n**Recommendation:** <minimum fix>\n```\n\n" +
		"## Candidate tasks\n\n" +
		"#### H1. A real finding after the fence  · **Status:** open\n\nBody.\n"

	out, err := SetFindingStatus(body, "H1", "fixed")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "#### H1. A real finding after the fence  · **Status:** fixed") {
		t.Errorf("the finding's own status was not rewritten:\n%s", out)
	}
	if !strings.Contains(out, "## Candidate tasks") {
		t.Errorf("an unrelated heading was corrupted:\n%s", out)
	}
	if !strings.Contains(out, "**Recommendation:** <minimum fix>") {
		t.Errorf("fenced example text was corrupted:\n%s", out)
	}
	// The fenced example must not be parsed as a real finding at all.
	codes := make([]string, 0, 2)
	for _, f := range ParseFindings(body) {
		codes = append(codes, f.Code)
	}
	if len(codes) != 1 || codes[0] != "H1" {
		t.Errorf("parsed %v; the fenced example should be invisible", codes)
	}
	// Every span must be a valid index into the ORIGINAL body, not a stripped copy.
	for _, f := range ParseFindings(body) {
		if f.StatusSpan.End > len(body) || body[f.StatusSpan.Start:f.StatusSpan.End] != f.Status {
			t.Errorf("span for %s does not locate its own status in the body", f.Code)
		}
	}
}

// Audit L1: a newline in a status would break the finding header apart.
func TestSetFindingStatusRejectsNewlines(t *testing.T) {
	body := "#### H1. T  · **Status:** open\n\nprose\n"
	if _, err := SetFindingStatus(body, "H1", "fixed\n#### H2. injected  · **Status:** open"); err == nil {
		t.Error("a newline in a status was accepted")
	}
}

// M2 of the 2026-08-17 finding-status-surface audit: a leading ✅/⏳/⛔ — written to keep a
// finding visually in sync with the candidate list below it — was captured AS the status,
// which made it the single highest-frequency cause of red audits in the corpus. The glyph
// is decoration; the word after it is the status. The span still covers the glyph, so
// re-stamping replaces it rather than leaving `✅ deferred` asserting two things at once.
func TestParseFindings_LeadingDecorationIsNotTheStatus(t *testing.T) {
	for _, tc := range []struct {
		name, body, want, wantSpan string
	}{
		{"emoji before token", "#### H1. T  · **Status:** ✅ fixed 2026-06-04\n", "fixed", "✅ fixed 2026-06-04"},
		{"undecorated is unchanged", "#### H1. T  · **Status:** fixed 2026-06-04\n", "fixed", "fixed 2026-06-04"},
		{"decoration on a status line", "#### H1. T\n\n**Status:** ⛔ wontfix (too costly) | **Effort:** S\n", "wontfix", "⛔ wontfix (too costly)"},
		// Decoration with no word is not a status — but the span still locates it, so
		// `audit finding --status` can repair the line instead of appending beside it.
		{"decoration only", "#### H1. T  · **Status:** ✅\n", "", "✅"},
		// The pre-existing empty-status case must not regress: `*` stays excluded, so a
		// following bold label is not swallowed as decoration.
		{"empty before a bold label", "#### H1. T  · **Status:** **Effort:** S\n", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := ParseFindings(tc.body)[0]
			if f.Status != tc.want {
				t.Errorf("Status = %q, want %q", f.Status, tc.want)
			}
			got := ""
			if !f.StatusSpan.Empty() {
				got = tc.body[f.StatusSpan.Start:f.StatusSpan.End]
			}
			if got != tc.wantSpan {
				t.Errorf("span text = %q, want %q", got, tc.wantSpan)
			}
		})
	}
}

// The decoration a re-stamp must not strand: stamping a decorated finding replaces the
// glyph along with the word, because the tool cannot know which glyph the new status wants.
func TestSetFindingStatus_ReplacesLeadingDecoration(t *testing.T) {
	body := "#### H1. T  · **Status:** ✅ fixed 2026-06-04\n\nprose\n"
	got, err := SetFindingStatus(body, "H1", "open")
	if err != nil {
		t.Fatalf("SetFindingStatus: %v", err)
	}
	if want := "#### H1. T  · **Status:** open\n\nprose\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// The resolution note: the paragraph saying HOW a finding was resolved, beside the one word
// saying THAT it was. It is tool-written for the same reason the status is — a hand-typed
// block is how this file's grammar drifts.
func TestSetFindingNote(t *testing.T) {
	const body = "## Findings\n\n#### H1. One  · **Status:** open\n\nprose here.\n\n" +
		"#### L1. Two  · **Status:** open\n\nmore prose.\n\n## Candidate tasks\n\n- ⏳ thing\n"

	t.Run("appends as the section's last block", func(t *testing.T) {
		got, err := SetFindingNote(body, "H1", "Widened the regex.")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got, "prose here.\n\n**Resolution:** Widened the regex.\n\n#### L1.") {
			t.Errorf("note not placed at the end of H1's section:\n%s", got)
		}
	})

	t.Run("re-noting replaces rather than stacking", func(t *testing.T) {
		once, _ := SetFindingNote(body, "H1", "First.")
		twice, _ := SetFindingNote(once, "H1", "Second.")
		if n := strings.Count(twice, FindingNoteLabel); n != 1 {
			t.Errorf("want exactly 1 label after re-noting, got %d:\n%s", n, twice)
		}
		if !strings.Contains(twice, "**Resolution:** Second.") {
			t.Errorf("replacement text missing:\n%s", twice)
		}
	})

	t.Run("empty note removes the block and restores the layout", func(t *testing.T) {
		noted, _ := SetFindingNote(body, "H1", "Temporary.")
		cleared, err := SetFindingNote(noted, "H1", "")
		if err != nil {
			t.Fatal(err)
		}
		if cleared != body {
			t.Errorf("clearing did not round-trip to the original:\n%q", cleared)
		}
	})

	// The bound that matters most: a finding's section stops at the next HEADING, not just
	// the next finding. Without it the last finding annexes `## Candidate tasks` and its
	// note lands under the wrong heading, at the bottom of the file.
	t.Run("the last finding does not annex the next section", func(t *testing.T) {
		got, err := SetFindingNote(body, "L1", "Bounded.")
		if err != nil {
			t.Fatal(err)
		}
		note, cand := strings.Index(got, FindingNoteLabel), strings.Index(got, "## Candidate tasks")
		if note < 0 || note > cand {
			t.Errorf("note landed after the candidate-tasks heading:\n%s", got)
		}
	})

	t.Run("a newline is refused", func(t *testing.T) {
		if _, err := SetFindingNote(body, "H1", "bad\n#### H9. injected"); !errors.Is(err, ErrValidation) {
			t.Errorf("want ErrValidation for an embedded newline, got %v", err)
		}
	})

	t.Run("an unknown code is not found", func(t *testing.T) {
		if _, err := SetFindingNote(body, "Z9", "x"); !errors.Is(err, ErrNotFound) {
			t.Errorf("want ErrNotFound, got %v", err)
		}
	})

	// Hard wrapping and the parsed value must round-trip: findingNote reads the paragraph
	// back through strings.Fields, so what goes in is what comes out regardless of wrapping.
	t.Run("wraps by runes and round-trips", func(t *testing.T) {
		long := "The atlas — which renders ✅ and → glyphs throughout its prose — is wrapped " +
			"here to prove that multi-byte runes are measured as one column each rather than " +
			"as their UTF-8 length, which would pull these lines visibly short of the margin."
		got, err := SetFindingNote(body, "H1", long)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(got, "\n") {
			if !strings.Contains(line, "wrapped") && !strings.HasPrefix(line, FindingNoteLabel) &&
				!strings.Contains(line, "runes") && !strings.Contains(line, "margin") &&
				!strings.Contains(line, "UTF-8") {
				continue
			}
			if n := utf8.RuneCountInString(line); n > proseWrapWidth {
				t.Errorf("line is %d runes, over the %d margin: %q", n, proseWrapWidth, line)
			}
		}
		if back := ParseFindings(got)[0].Note; back != long {
			t.Errorf("note did not round-trip:\n got %q\nwant %q", back, long)
		}
	})

	// Fenced example syntax is not a real note, for the same reason it is not a real finding.
	t.Run("a note inside a fence is not the finding's note", func(t *testing.T) {
		fenced := "#### H1. One  · **Status:** open\n\n```\n**Resolution:** an example\n```\n"
		if got := ParseFindings(fenced)[0].Note; got != "" {
			t.Errorf("fenced example parsed as a note: %q", got)
		}
	})
}

// `audit finding` refuses to WRITE a destination-less `tracked`, but the corpus is markdown
// and a hand edit routes around every writer. A handoff with nowhere to follow is precisely
// the improvisation the word was introduced to replace, so lint says so.
func TestLintFindings_TrackedNeedsADestination(t *testing.T) {
	body := "#### H1. One  · **Status:** tracked\n\n#### H2. Two  · **Status:** tracked by 6fbwhsw024nr\n"
	iss := LintFindings("open", ParseFindings(body))
	if len(iss) != 1 || iss[0].Field != "H1" || !strings.Contains(iss[0].Message, "needs a destination") {
		t.Fatalf("want one destination issue on H1, got %v", iss)
	}
}

// Only the FIRST **Resolution:** block is read. A hand edit that adds a second would have
// its text silently ignored, which is the failure mode worth naming out loud.
func TestLintFindings_DuplicateResolutionBlocks(t *testing.T) {
	body := "#### H1. One  · **Status:** fixed\n\n**Resolution:** first.\n\n**Resolution:** second.\n"
	iss := LintFindings("open", ParseFindings(body))
	if len(iss) != 1 || iss[0].Field != "H1" || !strings.Contains(iss[0].Message, "more than one") {
		t.Fatalf("want one duplicate-note issue on H1, got %v", iss)
	}
}

// The date on `fixed 2026-08-24` and the destination on `tracked by <id>` are information
// the wire used to drop, because Status carries only the vocabulary word.
func TestParseFindings_StatusDecoration(t *testing.T) {
	body := "#### H1. A  · **Status:** fixed 2026-08-24 (PR #12)\n" +
		"#### H2. B  · **Status:** ✅ tracked by 6fbwhsw024nr\n" +
		"#### H3. C  · **Status:** open\n"
	want := []string{"2026-08-24 (PR #12)", "by 6fbwhsw024nr", ""}
	for i, f := range ParseFindings(body) {
		if f.StatusDecoration != want[i] {
			t.Errorf("%s decoration = %q, want %q", f.Code, f.StatusDecoration, want[i])
		}
	}
}

// H1 of 2026-08-24-finding-note-and-vocabulary-selfreview: a label with a blank line after
// it has no text on its line, so it is not a note. Reporting a span anyway is how the
// paragraph beneath it got orphaned — the write replaced the label alone and stranded the
// prose it was introducing, while reporting success.
func TestFindingNote_LabelWithNoTextIsNotANote(t *testing.T) {
	body := "#### H1. A  · **Status:** fixed\n\n**Resolution:**\n\nThe text lives here.\n"
	f := ParseFindings(body)[0]
	if f.Note != "" || !f.NoteSpan.Empty() {
		t.Fatalf("want no note and no span, got note=%q span=%v", f.Note, f.NoteSpan)
	}
	// The paragraph must survive a write untouched, wherever the new block lands.
	got, err := SetFindingNote(body, "H1", "replacement")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "The text lives here.") {
		t.Errorf("the orphaned paragraph was destroyed:\n%s", got)
	}
	// …and lint names the stray label rather than leaving it to be discovered later.
	iss := LintFindings("open", ParseFindings(got))
	if len(iss) != 1 || !strings.Contains(iss[0].Message, "more than one") {
		t.Errorf("want the stray label flagged, got %v", iss)
	}
	if iss := LintFindings("open", ParseFindings(body)); len(iss) != 1 ||
		!strings.Contains(iss[0].Message, "empty") {
		t.Errorf("want an empty-label issue on the original, got %v", iss)
	}
}

// M4: a note that quotes the label — likely, in audits about this tool's own grammar — must
// not have the wrap put that text at a line start, where it reads as a second label and
// lint counts it as a duplicate of the note it belongs to.
func TestSetFindingNote_WrapNeverStartsALineWithTheLabel(t *testing.T) {
	for pad := 30; pad < 90; pad++ {
		text := strings.Repeat("x", pad) + " **Resolution:** trailing words here"
		got, err := SetFindingNote("#### H1. A  · **Status:** fixed\n", "H1", text)
		if err != nil {
			t.Fatal(err)
		}
		f := ParseFindings(got)[0]
		if f.NoteLabels != 1 {
			t.Fatalf("pad=%d produced %d labels:\n%s", pad, f.NoteLabels, got)
		}
		if f.Note != text {
			t.Fatalf("pad=%d did not round-trip:\n got %q\nwant %q", pad, f.Note, text)
		}
	}
}

// The bar groups seven statuses into four bands, and the whole point of the grouping is
// that it is TOTAL: every legal status belongs to exactly one band, so the bands always
// sum to the finding count and no status can quietly fall into the empty track. A status
// added without a home in TallyFindings would slip past every other test — the counts
// would simply be short, and the bar would under-fill, which reads as work remaining.
func TestTallyFindings_EveryStatusLandsInABand(t *testing.T) {
	for _, s := range FindingStatuses() {
		tally := TallyFindings([]Finding{{Code: "H1", Status: s}})
		if n := tally.Open + tally.Active + tally.Done + tally.Dropped; n != 1 {
			t.Errorf("status %q lands in %d bands, want exactly 1 — add it to TallyFindings", s, n)
		}
	}
	// …and the whole vocabulary at once fills the bar completely.
	fs := make([]Finding, 0, len(FindingStatuses()))
	for _, s := range FindingStatuses() {
		fs = append(fs, Finding{Code: "H1", Status: s})
	}
	tally := TallyFindings(fs)
	if n := tally.Open + tally.Active + tally.Done + tally.Dropped; n != len(fs) {
		t.Errorf("bands sum to %d of %d findings — the bar would under-fill", n, len(fs))
	}
}
