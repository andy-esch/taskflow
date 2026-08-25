package domain

import (
	"errors"
	"strings"
	"testing"
)

const acBody = `# Title

Some intro prose.

## Acceptance criteria

- [x] first is done
- [ ] second is not
- [X] third is done (capital X)

## Notes

- [ ] this checkbox is NOT acceptance criteria and must not count
`

func TestCountAcceptanceCriteria(t *testing.T) {
	got := CountAcceptanceCriteria(acBody)
	if got.Checked != 2 || got.Total != 3 {
		t.Fatalf("CountAcceptanceCriteria = %+v, want {Checked:2 Total:3}", got)
	}
}

func TestCountAcceptanceCriteria_NoSection(t *testing.T) {
	if got := CountAcceptanceCriteria("# Title\n\njust prose, no criteria\n"); got != (ACCount{}) {
		t.Fatalf("no AC section should be zero tally, got %+v", got)
	}
}

func TestCountAcceptanceCriteria_IgnoresFencedCheckboxes(t *testing.T) {
	body := "## Acceptance criteria\n\n- [x] real one\n\n```\n- [ ] fenced example, not real\n```\n"
	if got := CountAcceptanceCriteria(body); got.Checked != 1 || got.Total != 1 {
		t.Fatalf("fenced checkbox must not count, got %+v", got)
	}
}

func TestSection(t *testing.T) {
	sec, ok := Section(acBody, "acceptance")
	if !ok {
		t.Fatal("expected to find the acceptance section")
	}
	want := "## Acceptance criteria\n\n- [x] first is done\n- [ ] second is not\n- [X] third is done (capital X)"
	if sec != want {
		t.Fatalf("Section(acceptance) =\n%q\nwant\n%q", sec, want)
	}
}

func TestSection_TrailingSectionRunsToEnd(t *testing.T) {
	sec, ok := Section(acBody, "notes")
	if !ok {
		t.Fatal("expected to find the notes section")
	}
	want := "## Notes\n\n- [ ] this checkbox is NOT acceptance criteria and must not count"
	if sec != want {
		t.Fatalf("Section(notes) =\n%q\nwant\n%q", sec, want)
	}
}

func TestSection_NestedDeeperHeadingsStayInside(t *testing.T) {
	body := "## Design\n\ntop.\n\n### Sub\n\nnested.\n\n## After\n\nout.\n"
	sec, ok := Section(body, "design")
	if !ok {
		t.Fatal("expected to find the design section")
	}
	want := "## Design\n\ntop.\n\n### Sub\n\nnested."
	if sec != want {
		t.Fatalf("Section(design) =\n%q\nwant\n%q", sec, want)
	}
}

func TestSection_NotFound(t *testing.T) {
	if _, ok := Section(acBody, "nonexistent"); ok {
		t.Fatal("expected no match for a missing section")
	}
}

// CRLF line endings (Windows / core.autocrlf checkout / a CRLF --body-file) must
// not blind the heading + checkbox scanners.
func TestCountAcceptanceCriteria_CRLF(t *testing.T) {
	body := "# Title\r\n\r\n## Acceptance criteria\r\n\r\n- [x] a\r\n- [ ] b\r\n"
	if got := CountAcceptanceCriteria(body); got.Checked != 1 || got.Total != 2 {
		t.Fatalf("CRLF body tally = %+v, want {Checked:1 Total:2}", got)
	}
}

func TestSection_CRLF(t *testing.T) {
	body := "## Notes\r\n\r\nsome text\r\n"
	sec, ok := Section(body, "notes")
	if !ok || sec != "## Notes\n\nsome text" {
		t.Fatalf("Section on CRLF body = (%q, %v), want normalized LF block", sec, ok)
	}
}

// A fenced block that contains an info-stringed inner fence (```go inside
// ```markdown) or a shorter inner fence must NOT close the outer block — a naive
// boolean toggle would, leaking the fenced `##`/`- [ ]` into structure scanning.
func TestFenceScanner_NestedFencesDoNotLeak(t *testing.T) {
	body := "## Acceptance criteria\n\n" +
		"```markdown\n" +
		"```go\n" +
		"## Fake nested heading\n" +
		"- [ ] fenced example, not a real criterion\n" +
		"```\n\n" +
		"- [x] the only real criterion\n"
	// The fenced `## Fake nested heading` must not truncate the section...
	sec, ok := Section(body, "acceptance")
	if !ok || !strings.Contains(sec, "the only real criterion") {
		t.Fatalf("nested fence truncated the section:\n%q", sec)
	}
	// ...and the fenced `- [ ]` must not count.
	if got := CountAcceptanceCriteria(body); got.Checked != 1 || got.Total != 1 {
		t.Fatalf("nested-fence tally = %+v, want {Checked:1 Total:1}", got)
	}
}

func TestListAcceptanceCriteria(t *testing.T) {
	got := ListAcceptanceCriteria(acBody)
	want := []Criterion{
		{Index: 1, Checked: true, Text: "first is done", State: CriterionMet},
		{Index: 2, Checked: false, Text: "second is not", State: CriterionUnmet},
		{Index: 3, Checked: true, Text: "third is done (capital X)", State: CriterionMet},
	}
	if len(got) != len(want) {
		t.Fatalf("ListAcceptanceCriteria len = %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("criterion %d = %+v, want %+v", i+1, got[i], want[i])
		}
	}
}

func TestListAcceptanceCriteria_NoSection(t *testing.T) {
	if got := ListAcceptanceCriteria("# Title\n\njust prose\n"); len(got) != 0 {
		t.Fatalf("no AC section should list nothing, got %+v", got)
	}
}

func TestSetAcceptanceCriterion(t *testing.T) {
	// Check the currently-unchecked #2.
	out, err := SetAcceptanceCriterion(acBody, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	got := ListAcceptanceCriteria(out)
	if !got[1].Checked {
		t.Errorf("criterion 2 should be checked after the flip:\n%s", out)
	}
	// Only that one line changed: #1 and #3 stay checked, everything else identical.
	if !got[0].Checked || !got[2].Checked || got[1].Text != "second is not" {
		t.Errorf("flip must not disturb other criteria or text:\n%s", out)
	}
	// The non-AC checkbox under ## Notes must be untouched.
	if !strings.Contains(out, "- [ ] this checkbox is NOT acceptance criteria") {
		t.Errorf("flip must not touch checkboxes outside the AC section:\n%s", out)
	}
}

func TestSetAcceptanceCriterion_Uncheck(t *testing.T) {
	out, err := SetAcceptanceCriterion(acBody, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if ListAcceptanceCriteria(out)[0].Checked {
		t.Errorf("criterion 1 should be unchecked:\n%s", out)
	}
}

func TestSetAcceptanceCriterion_Idempotent(t *testing.T) {
	out, err := SetAcceptanceCriterion(acBody, 1, true) // #1 already checked
	if err != nil {
		t.Fatal(err)
	}
	if out != acBody {
		t.Errorf("flipping to the current state must return the body unchanged")
	}
}

func TestSetAcceptanceCriterion_OutOfRange(t *testing.T) {
	if _, err := SetAcceptanceCriterion(acBody, 9, true); !errors.Is(err, ErrValidation) {
		t.Errorf("out-of-range index should be ErrValidation, got %v", err)
	}
	if _, err := SetAcceptanceCriterion(acBody, 0, true); !errors.Is(err, ErrValidation) {
		t.Errorf("index 0 should be ErrValidation, got %v", err)
	}
}

func TestSetAcceptanceCriterion_NoSection(t *testing.T) {
	if _, err := SetAcceptanceCriterion("# Title\n\nno criteria\n", 1, true); !errors.Is(err, ErrValidation) {
		t.Errorf("no AC section should be ErrValidation, got %v", err)
	}
}

// A multi-line criterion (a checkbox with an indented continuation line — the shape
// real tasks use) is ONE criterion: the continuation isn't a separate checkbox, and a
// flip touches only the checkbox line, leaving the continuation intact.
//
// Its Text is the WHOLE criterion. This previously asserted the truncated first line,
// which is how `task ac --list`, `task show`, and the JSON all came to show
// "…rather than introducing a" and silently drop the rest of the sentence.
func TestSetAcceptanceCriterion_MultiLine(t *testing.T) {
	body := "## Acceptance criteria\n\n- [ ] first criterion spans\n      a continuation line\n- [x] second is done\n"
	cs := ListAcceptanceCriteria(body)
	if len(cs) != 2 || cs[0].Text != "first criterion spans a continuation line" {
		t.Fatalf("multi-line criterion should count once and read whole: %+v", cs)
	}
	out, err := SetAcceptanceCriterion(body, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "- [x] first criterion spans") || !strings.Contains(out, "      a continuation line") {
		t.Errorf("flip must tick the checkbox line and preserve the continuation:\n%s", out)
	}
}

// check-then-uncheck restores the body byte-for-byte — the surgical guarantee.
func TestSetAcceptanceCriterion_RoundTripByteIdentical(t *testing.T) {
	checked, err := SetAcceptanceCriterion(acBody, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	back, err := SetAcceptanceCriterion(checked, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if back != acBody {
		t.Errorf("check then uncheck must restore the body exactly:\n got %q\nwant %q", back, acBody)
	}
}

// List and flip both skip a checkbox inside a fenced block: the index numbers only
// the real criteria, and a flip targets the real checkbox, never the fenced example.
func TestAcceptanceCriteria_FenceAware(t *testing.T) {
	body := "## Acceptance criteria\n\n- [x] real one\n\n```\n- [ ] fenced example, not real\n```\n\n- [ ] real two\n"
	cs := ListAcceptanceCriteria(body)
	if len(cs) != 2 || cs[1].Text != "real two" {
		t.Fatalf("fenced checkbox must not be listed: %+v", cs)
	}
	out, err := SetAcceptanceCriterion(body, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "- [x] real two") || strings.Contains(out, "- [x] fenced example") {
		t.Errorf("flip index 2 must target the real checkbox, not the fenced one:\n%s", out)
	}
}

func TestLintAcceptanceCriteria_Clean(t *testing.T) {
	if iss := LintAcceptanceCriteria(acBody); len(iss) != 0 {
		t.Errorf("well-formed acceptance criteria should not lint, got %+v", iss)
	}
}

func TestLintAcceptanceCriteria_Malformed(t *testing.T) {
	body := "## Acceptance criteria\n\n- [x] ok\n- [] empty\n- [ x] spaced\n- [  ] two spaces\n"
	iss := LintAcceptanceCriteria(body)
	if len(iss) != 3 {
		t.Fatalf("expected 3 malformed-checkbox warnings, got %d: %+v", len(iss), iss)
	}
	for _, i := range iss {
		if i.Field != "acceptance" || !strings.Contains(i.Message, "malformed") {
			t.Errorf("unexpected issue: %+v", i)
		}
	}
}

// The malformed heuristic is deliberately narrow: citations, partial markers, and
// links must NOT be flagged (they'd break lint-clean on legit content).
func TestLintAcceptanceCriteria_NoFalsePositives(t *testing.T) {
	body := "## Acceptance criteria\n\n- [x] ok\n- [1] a citation\n- [-] a partial marker\n- [see docs](http://x) a link\n"
	if iss := LintAcceptanceCriteria(body); len(iss) != 0 {
		t.Errorf("citations/markers/links must not be flagged, got %+v", iss)
	}
}

func TestLintAcceptanceCriteria_MultipleSections(t *testing.T) {
	body := "## Acceptance criteria\n\n- [ ] a\n\n## Acceptance criteria\n\n- [ ] b\n"
	iss := LintAcceptanceCriteria(body)
	found := false
	for _, i := range iss {
		if strings.Contains(i.Message, "acceptance-criteria sections") {
			found = true
		}
	}
	if !found {
		t.Errorf("duplicate acceptance sections should be flagged: %+v", iss)
	}
}

// A later heading that merely MENTIONS "acceptance" (e.g. a Progress note) must not
// be miscounted as a second acceptance section — the guard must not false-positive on
// its own kind of prose. (Regression: this bit the tool's own self-hosted task.)
func TestLintAcceptanceCriteria_MentionHeadingNotCounted(t *testing.T) {
	body := "## Acceptance criteria\n\n- [ ] a\n\n## Progress — notes on acceptance criteria\n\nsome text\n"
	if iss := LintAcceptanceCriteria(body); len(iss) != 0 {
		t.Errorf("a heading merely mentioning 'acceptance' must not count as a second section, got %+v", iss)
	}
}

func TestLintAcceptanceCriteria_FencedNotFlagged(t *testing.T) {
	body := "## Acceptance criteria\n\n- [x] real\n\n```\n- [] fenced botched\n```\n"
	if iss := LintAcceptanceCriteria(body); len(iss) != 0 {
		t.Errorf("a malformed checkbox inside a fence must not be flagged, got %+v", iss)
	}
}

// A botched checkbox OUTSIDE any acceptance section isn't the tally's business.
func TestLintAcceptanceCriteria_OnlyInSection(t *testing.T) {
	if iss := LintAcceptanceCriteria("# Title\n\n- [] not in an AC section\n"); len(iss) != 0 {
		t.Errorf("malformed checkbox outside the AC section must not be flagged, got %+v", iss)
	}
}

func TestSection_IgnoresFencedHeadings(t *testing.T) {
	body := "## Real\n\ntext.\n\n```\n## Fake heading in a fence\n```\n\nmore text under Real.\n"
	sec, ok := Section(body, "real")
	if !ok {
		t.Fatal("expected to find the real section")
	}
	// The fenced `## Fake` must not close the section — everything is one block.
	want := "## Real\n\ntext.\n\n```\n## Fake heading in a fence\n```\n\nmore text under Real."
	if sec != want {
		t.Fatalf("Section(real) =\n%q\nwant\n%q", sec, want)
	}
}

// A tally of "1/4" cannot distinguish three criteria still to do from one to do and two
// deliberately not happening. Explained is what makes the difference legible — and it is
// what a future `task complete` reconciliation would have to reason about.
func TestCountAcceptanceCriteriaSeparatesExplainedFromOutstanding(t *testing.T) {
	body := "# T\n\n## Acceptance criteria\n\n" +
		"- [x] done\n" +
		"- [ ] still to do\n" +
		"- [ ] parked · **deferred:** waiting on the ADR\n" +
		"- [ ] dropped · **n/a:** scope removed\n"
	got := CountAcceptanceCriteria(body)
	if got.Total != 4 || got.Checked != 1 || got.Explained != 2 {
		t.Errorf("tally = %+v; want 4 total, 1 checked, 2 explained", got)
	}
	// A body written before the vocabulary existed tallies exactly as it always did.
	legacy := CountAcceptanceCriteria("# T\n\n## Acceptance criteria\n\n- [x] a\n- [ ] b\n")
	if legacy.Total != 2 || legacy.Checked != 1 || legacy.Explained != 0 {
		t.Errorf("legacy tally = %+v; want 2 total, 1 checked, 0 explained", legacy)
	}
}

// The write path is what the vocabulary needed to ship WITH. A state reachable only by
// hand-editing is one nobody can be held to — the habit that let the finding vocabulary
// drift from its own documentation.
func TestSetCriterionState(t *testing.T) {
	body := "# T\n\n## Acceptance criteria\n\n- [ ] alpha\n- [x] beta\n  - [ ] nested\n"

	out, err := SetCriterionState(body, 1, CriterionDeferred, "waiting on the ADR")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "- [ ] alpha · **deferred:** waiting on the ADR") {
		t.Errorf("deferred not written as expected:\n%s", out)
	}

	// Re-setting must REPLACE the suffix, not stack another one.
	out, err = SetCriterionState(out, 1, CriterionWontFix, "changed approach")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(out, "**deferred:**") != 0 || strings.Count(out, "**wontfix:**") != 1 {
		t.Errorf("suffixes stacked instead of replacing:\n%s", out)
	}

	// Met clears the suffix and checks the box: the two halves cannot disagree.
	out, err = SetCriterionState(out, 1, CriterionMet, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "- [x] alpha\n") || strings.Contains(out, "**wontfix:**") {
		t.Errorf("met did not clear the suffix:\n%s", out)
	}

	// A nested criterion keeps its indentation and bullet.
	out, err = SetCriterionState(body, 3, CriterionNA, "scope removed")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "  - [ ] nested · **n/a:** scope removed") {
		t.Errorf("nested indentation was not preserved:\n%s", out)
	}

	// A reason is required for the non-binary states and rejected for the binary ones.
	if _, err := SetCriterionState(body, 1, CriterionDeferred, ""); err == nil {
		t.Error("deferred without a reason was accepted")
	}
	if _, err := SetCriterionState(body, 1, CriterionMet, "why"); err == nil {
		t.Error("met with a reason was accepted")
	}
	if _, err := SetCriterionState(body, 9, CriterionMet, ""); err == nil {
		t.Error("an out-of-range index was accepted")
	}
}

// Audit 2026-08-24-planning-state-vocabulary M3 and L1. The disposition suffix is the LAST
// state marker on the line, not the first: a criterion may legitimately carry its own bold
// labels, and matching leftmost silently dropped the trailing deferral. And a reason must be
// one line — a newline inside a checkbox list manufactures phantom criteria.
func TestCriterionSuffixIsTrailingAndSingleLine(t *testing.T) {
	body := "# T\n\n## Acceptance criteria\n\n" +
		"- [ ] Tests · **Coverage:** 100% · **deferred:** waiting on harness\n" +
		"- [ ] No state · **Note:** just prose\n"
	cs := ListAcceptanceCriteria(body)
	if cs[0].State != CriterionDeferred || cs[0].Reason != "waiting on harness" {
		t.Errorf("trailing suffix missed: state=%q reason=%q", cs[0].State, cs[0].Reason)
	}
	if cs[0].Text != "Tests · **Coverage:** 100%" {
		t.Errorf("the criterion's own bold label was eaten: %q", cs[0].Text)
	}
	if cs[1].State != CriterionUnmet || cs[1].Text != "No state · **Note:** just prose" {
		t.Errorf("a non-state label was treated as a disposition: %+v", cs[1])
	}
	if got := CountAcceptanceCriteria(body); got.Explained != 1 {
		t.Errorf("tally lost the explained criterion: %+v", got)
	}
	// Re-setting replaces only the trailing disposition, leaving the earlier label alone.
	out, err := SetCriterionState(body, 1, CriterionWontFix, "changed approach")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "- [ ] Tests · **Coverage:** 100% · **wontfix:** changed approach") {
		t.Errorf("re-set mangled the line:\n%s", out)
	}
	// A multi-line reason would break out of the list item.
	if _, err := SetCriterionState(body, 1, CriterionDeferred, "why\n- [ ] injected"); err == nil {
		t.Error("a newline in a reason was accepted")
	}
}

// A wrapped criterion's state suffix belongs at the END of the criterion, not the end of
// its marker line. Appending it to the first line splits the sentence and leaves its tail
// dangling beneath the reason:
//
//   - [ ] Criterion states reuse the finding glyph vocabulary rather than introducing a · **deferred:** not yet
//     parallel one.
//
// which is what this repo's own planning file looked like until the scanner learned that
// criteria wrap.
func TestSetCriterionState_WrappedCriterion(t *testing.T) {
	const body = "## Acceptance criteria\n\n" +
		"- [ ] Criterion states reuse the finding glyph vocabulary rather than introducing a\n" +
		"  parallel one.\n" +
		"- [ ] A short one.\n"
	const whole = "Criterion states reuse the finding glyph vocabulary rather than introducing a parallel one."

	deferred, err := SetCriterionState(body, 1, CriterionDeferred, "not yet")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(deferred, "  parallel one. · **deferred:** not yet\n") {
		t.Errorf("suffix must land at the end of the criterion, indentation intact:\n%s", deferred)
	}
	c := ListAcceptanceCriteria(deferred)[0]
	if c.Text != whole || c.State != CriterionDeferred || c.Reason != "not yet" {
		t.Errorf("round-trip: text=%q state=%q reason=%q", c.Text, c.State, c.Reason)
	}

	// Re-setting REPLACES the suffix wherever it sits, and must not strip the
	// continuation's indent — losing it would make the line stop being a continuation, so
	// the next write would leave its suffix stranded and `met` would not clear it.
	wont, err := SetCriterionState(deferred, 1, CriterionWontFix, "changed my mind")
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(wont, "· **"); n != 1 {
		t.Errorf("want exactly one suffix after re-setting, got %d:\n%s", n, wont)
	}
	if !strings.Contains(wont, "  parallel one. · **wontfix:** changed my mind\n") {
		t.Errorf("re-set lost the continuation's indentation:\n%s", wont)
	}

	// met clears the suffix from wherever it is and ticks the box — the whole criterion
	// returns to its original text.
	met, err := SetCriterionState(wont, 1, CriterionMet, "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(met, "· **") {
		t.Errorf("met must drop the suffix:\n%s", met)
	}
	if met != strings.Replace(body, "- [ ] Criterion", "- [x] Criterion", 1) {
		t.Errorf("met should restore the criterion's text exactly:\n%s", met)
	}
}

// The continuation rule stops where a markdown reader stops: a blank line, a new list
// item, a heading, or a fence. Without those bounds one criterion would swallow the rest
// of the section.
func TestListAcceptanceCriteria_ContinuationBounds(t *testing.T) {
	body := "## Acceptance criteria\n\n" +
		"- [ ] wraps\n  onto here\n\n" + // blank line ends it
		"  orphaned prose that is not part of it\n" +
		"- [ ] next one\n" +
		"  - a nested bullet is a sub-list, not more sentence\n" +
		"- [ ] third\n" +
		"## Next section\n" +
		"  not a criterion either\n"
	cs := ListAcceptanceCriteria(body)
	want := []string{"wraps onto here", "next one", "third"}
	if len(cs) != len(want) {
		t.Fatalf("got %d criteria, want %d: %+v", len(cs), len(want), cs)
	}
	for i, w := range want {
		if cs[i].Text != w {
			t.Errorf("criterion %d = %q, want %q", i+1, cs[i].Text, w)
		}
	}
}

// The roll-up's source. A bare "3/8" reads as five things still to do; the tally is what
// lets a surface say that three of those five were DECIDED rather than forgotten.
func TestTallyCriteria(t *testing.T) {
	body := "## Acceptance criteria\n\n" +
		"- [x] done one\n" +
		"- [x] done two\n" +
		"- [ ] still open\n" +
		"- [ ] parked · **deferred:** waiting on the ADR\n" +
		"- [ ] abandoned · **wontfix:** superseded\n" +
		"- [ ] moved · **tracked:** carried by 6g3ag8py12y9\n" +
		"- [ ] moot · **n/a:** the tile grid was dropped\n"
	got := TallyCriteria(body)
	want := []CriterionCount{
		{CriterionMet, 2}, {CriterionUnmet, 1}, {CriterionDeferred, 1},
		{CriterionWontFix, 1}, {CriterionTracked, 1}, {CriterionNA, 1},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("position %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	// States with no members are omitted rather than rendered as zeroes.
	if n := len(TallyCriteria("## Acceptance criteria\n\n- [x] only one\n")); n != 1 {
		t.Errorf("a single-state body should tally one entry, got %d", n)
	}
	// No section, no tally — which is most tasks.
	if n := len(TallyCriteria("# Title\n\nprose\n")); n != 0 {
		t.Errorf("a body with no criteria should tally nothing, got %d", n)
	}
}
