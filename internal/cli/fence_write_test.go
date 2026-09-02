package cli

import (
	"strings"
	"testing"
)

// A body that ends inside an open fence parses fine and renders fine — the failure
// is that every heading, criterion and finding after the fence becomes invisible to
// the scanners. The write is the last moment the author still has the text, so it
// is refused here rather than discovered later as a finding that cannot be stamped.
func TestBodyWrite_RejectsUnterminatedFence(t *testing.T) {
	root := setupRepo(t)

	_, err := runRootRC(t, "-C", root, "task", "set", "alpha", "--body",
		"# Alpha\n\n```\nnever closed\n")
	if err == nil {
		t.Fatal("a body ending inside an open fence must be refused")
	}
	if got := ExitCode(err); got != 11 {
		t.Errorf("exit = %d, want 11", got)
	}
	for _, want := range []string{"unterminated", "line 3"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error should locate the fence (missing %q): %v", want, err)
		}
	}
}

// The same guard on the append path, which is where a truncated write actually
// produced this in the field.
func TestBodyAppend_RejectsUnterminatedFence(t *testing.T) {
	root := setupRepo(t)

	_, err := runRootRC(t, "-C", root, "task", "append", "alpha", "--body",
		"## Notes\n\n~~~\ntruncated")
	if err == nil {
		t.Fatal("an append ending inside an open fence must be refused")
	}
	if !strings.Contains(err.Error(), "unterminated") {
		t.Errorf("want an unterminated-fence error, got: %v", err)
	}
}

// An append that CLOSES a fence the stored body left open is legitimate: the guard
// is about the resulting document, not the fragment being written.
func TestBodyAppend_JudgesTheResultingBodyNotTheFragment(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "set", "alpha", "--body", "# Alpha\n\nprose\n")

	// A fragment that opens AND closes is fine on its own.
	runRoot(t, "-C", root, "task", "append", "alpha", "--body", "```\ncode\n```\n")

	out := runRoot(t, "-C", root, "task", "show", "alpha")
	if !strings.Contains(out, "code") {
		t.Errorf("the balanced append should have landed:\n%s", out)
	}
}

// The end-to-end shape of H4: an audit whose first finding documents fence syntax
// inside a four-backtick block. Every later finding must remain visible to the
// index, because a finding the index cannot see can never be stamped — `audit
// finding H3 --status fixed` fails permanently — while the progress bar reports the
// survivors as the whole audit.
func TestAuditFindings_SurviveANestedFenceInAnEarlierFinding(t *testing.T) {
	root := setupRepo(t)
	body := "# Audit: x\n\n## Findings\n\n" +
		"#### H1. fence syntax · **Status:** open\n\n" +
		"````\n```mermaid\n````\n\n" +
		"#### H2. second · **Status:** open\n\n" +
		"```\ncode\n```\n\n" +
		"#### H3. third · **Status:** open\n"

	runRoot(t, "-C", root, "audit", "new", "fences", "--date", "2026-09-02", "--body", body)

	out := runRoot(t, "-C", root, "audit", "info", "2026-09-02-fences")
	if !strings.Contains(out, "3 total") {
		t.Errorf("all three findings should be indexed:\n%s", out)
	}
	// The decisive one: a swallowed finding cannot be stamped at all.
	runRoot(t, "-C", root, "audit", "finding", "2026-09-02-fences", "H3", "--status", "fixed", "--note", "done")
	if out := runRoot(t, "-C", root, "audit", "info", "2026-09-02-fences"); !strings.Contains(out, "1 done") {
		t.Errorf("H3 should be stampable:\n%s", out)
	}
}

// Nested fences are valid CommonMark and must survive the write guard — a
// four-backtick block wrapping a three-backtick example is how fence syntax itself
// gets documented.
func TestBodyWrite_AcceptsNestedFences(t *testing.T) {
	root := setupRepo(t)

	runRoot(t, "-C", root, "task", "set", "alpha", "--body",
		"# Alpha\n\n````\n```mermaid\ngraph TD;\n```\n````\n")

	out := runRoot(t, "-C", root, "task", "show", "alpha")
	if !strings.Contains(out, "mermaid") {
		t.Errorf("a nested fence should be written verbatim:\n%s", out)
	}
}
