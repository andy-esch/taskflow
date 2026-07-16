package domain

import (
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
