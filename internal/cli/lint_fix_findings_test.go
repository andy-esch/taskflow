package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
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
