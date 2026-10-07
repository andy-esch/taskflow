package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

const driftedAudit = "---\nid: 6fjangd7kva1\nbucket: open\narea: drift\ndate: \"2026-01-01\"\n---\n" +
	"# Audit: drift\n\n## Findings\n\n" +
	"### 1. Ordinary numbered section\n\nLegitimate structure.\n\n" +
	"#### BTA-01 — A finding the parser cannot see\n\n**Status:** open\n\nBody.\n"

// `lint --fix` repairs a drifted finding header in place, so the author is not sent
// back to re-edit the document by hand. The ordinary numbered section beside it must
// survive untouched — that is the whole reason the recognizer is letter-led.
func TestLintFix_CanonicalizesNearMissFindingHeader(t *testing.T) {
	root := setupRepo(t)
	p, content := testutil.AuditFixture(root, "open", "2026-01-01-drift.md", driftedAudit)
	testutil.Write(t, p, content)

	// setupRepo's own tasks are not lint-clean, so --fix exits non-zero for the
	// leftovers it cannot repair; the header repair below is what this pins.
	out, _ := runRootRC(t, "-C", root, "lint", "--fix")
	if !strings.Contains(out, "BTA1.") {
		t.Errorf("fix output should name the replacement, got:\n%s", out)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, "#### BTA1. A finding the parser cannot see") {
		t.Errorf("header was not canonicalized:\n%s", got)
	}
	if !strings.Contains(got, "### 1. Ordinary numbered section") {
		t.Errorf("an ordinary numbered heading was rewritten:\n%s", got)
	}
	if !strings.Contains(got, "**Status:** open") {
		t.Errorf("the finding's status line was disturbed:\n%s", got)
	}
}

func TestLintFix_StatusExamplesNeverManufactureFindings(t *testing.T) {
	for _, example := range []string{
		"Inline example: `· **Status:** fixed ` is documentation.",
		"> · **Status:** fixed",
		"The example is · **Status:** fixed",
		"    **Status:** fixed",
	} {
		t.Run(example, func(t *testing.T) {
			root := setupRepo(t)
			path, content := testutil.AuditFixture(root, "open", "probe.md",
				"---\nid: "+testutil.TaskID("probe")+"\narea: probe\ndate: 2026-10-06\n---\n## Findings\n\n#### H1 Example syntax\n\n"+example+"\n")
			testutil.Write(t, path, content)
			for range 2 {
				out, _ := runRootRC(t, "-C", root, "lint", "--fix")
				after, err := os.ReadFile(path)
				if err != nil || string(after) != content || strings.Contains(out, "missing **Status:**") {
					t.Fatalf("repair changed example into a finding: err=%v\n%s\n%s", err, after, out)
				}
				var shown wire.AuditShowEnvelope
				out = runRoot(t, "-C", root, "audit", "show", "probe", "--frontmatter-only", "--json")
				if err := json.Unmarshal([]byte(out), &shown); err != nil {
					t.Fatal(err)
				}
				if shown.Audit.Findings != 0 || shown.Audit.DoneFindings != 0 || shown.Audit.ReadyToClose || shown.Audit.UnparsedFindings != 1 {
					t.Fatalf("example manufactured audit progress: %+v", shown.Audit)
				}
			}
			// Narrative examples must remain writable, including a newly introduced
			// code-shaped header; only real metadata grants the write guard authority.
			runRoot(t, "-C", root, "audit", "append", "probe", "--body", "#### H2 Another example\n\n"+example)
			after, err := os.ReadFile(path)
			if err != nil || !strings.Contains(string(after), "#### H2 Another example") || len(domain.ParseFindings(string(after))) != 0 {
				t.Fatalf("narrative append failed or invented findings: err=%v\n%s", err, after)
			}
		})
	}
}

// A dry run reports the repair without writing it.
func TestLintFix_DryRunDoesNotWriteHeaderRepair(t *testing.T) {
	root := setupRepo(t)
	p, content := testutil.AuditFixture(root, "open", "2026-01-01-drift.md", driftedAudit)
	testutil.Write(t, p, content)
	before, _ := os.ReadFile(p)

	out := runRoot(t, "-C", root, "--dry-run", "lint", "--fix")
	if !strings.Contains(out, "BTA1.") {
		t.Errorf("dry run should still preview the repair, got:\n%s", out)
	}
	after, _ := os.ReadFile(p)
	if string(after) != string(before) {
		t.Errorf("a dry run must not write:\n%s", after)
	}
}

// Repair is idempotent: a second --fix over the repaired tree changes nothing, and
// an already-canonical audit is never rewritten.
func TestLintFix_HeaderRepairIsIdempotent(t *testing.T) {
	root := setupRepo(t)
	p, content := testutil.AuditFixture(root, "open", "2026-01-01-drift.md", driftedAudit)
	testutil.Write(t, p, content)

	_, _ = runRootRC(t, "-C", root, "lint", "--fix")
	first, _ := os.ReadFile(p)
	_, _ = runRootRC(t, "-C", root, "lint", "--fix")
	second, _ := os.ReadFile(p)
	if string(first) != string(second) {
		t.Errorf("a second --fix rewrote the file:\n%s", second)
	}
}

func TestLintFix_PreservesOrdinaryAndAmbiguousHeadingsWithoutPhantomFindings(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"ordinary", "## Findings\n\n### S3 Storage Architecture\n\n### V2 Migration Guide\n\n### MP3 Audio Support\n\n### Top3 Recommendations\n", 0},
		{"ambiguous", "## Findings\n\n#### H1 Missing metadata\n", 0},
		{"duplicate", "#### H1. Real · **Status:** fixed\n\n#### h-1: Same code · **Status:** open\n", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := setupRepo(t)
			path, content := testutil.AuditFixture(root, "open", "probe.md",
				"---\nid: "+testutil.TaskID("probe")+"\narea: probe\ndate: 2026-10-06\n---\n"+tc.body)
			testutil.Write(t, path, content)
			out, _ := runRootRC(t, "-C", root, "lint", "--fix")
			after, err := os.ReadFile(path)
			if err != nil || string(after) != content {
				t.Fatalf("non-repairable audit was rewritten: err=%v\n%s", err, after)
			}
			if got := len(domain.ParseFindings(string(after))); got != tc.findings {
				t.Fatalf("repair manufactured findings: got %d, want %d", got, tc.findings)
			}
			if strings.Contains(out, "missing **Status:**") {
				t.Fatalf("repair manufactured a missing-status error:\n%s", out)
			}
		})
	}
}
