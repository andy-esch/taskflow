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
		{Index: 1, Checked: true, Text: "first is done"},
		{Index: 2, Checked: false, Text: "second is not"},
		{Index: 3, Checked: true, Text: "third is done (capital X)"},
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
func TestSetAcceptanceCriterion_MultiLine(t *testing.T) {
	body := "## Acceptance criteria\n\n- [ ] first criterion spans\n      a continuation line\n- [x] second is done\n"
	cs := ListAcceptanceCriteria(body)
	if len(cs) != 2 || cs[0].Text != "first criterion spans" {
		t.Fatalf("multi-line criterion should count once: %+v", cs)
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
