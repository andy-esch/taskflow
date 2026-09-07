package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// An audit finding defect must fail the TOP-LEVEL `lint`, not just `audit lint`.
// Before audits joined the roster, `lint` was the command the repo actually runs
// and it reported green while a finding carried an unusable status.
func TestLintFoldsAuditFindingIssues(t *testing.T) {
	root := setupRepo(t)
	p, content := testutil.AuditFixture(root, "open", "2026-01-01-x.md",
		"---\narea: x\ndate: 2026-01-01\n---\n#### S1. t · **Status:** opne\n")
	testutil.Write(t, p, content)

	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "lint"})
	if err := cmd.Execute(); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("top-level lint must flag an audit finding defect (exit 11), got %v", err)
	}
	if got := out.String(); !strings.Contains(got, "2026-01-01-x") {
		t.Errorf("lint output should name the offending audit, got:\n%s", got)
	}
}

// The fold carries the audit's FRONTMATTER checks too, not just finding grammar:
// a foreign bucket is an audit-shaped defect the top-level gate must now catch.
func TestLintFoldsAuditFrontmatterIssues(t *testing.T) {
	root := setupRepo(t)
	p, content := testutil.AuditFixture(root, "open", "2026-01-01-y.md",
		"---\nid: "+testutil.TaskID("2026-01-01-y")+"\nbucket: nonsense\narea: y\ndate: 2026-01-01\n---\n#### S1. t · **Status:** open\n")
	testutil.Write(t, p, content)

	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "lint"})
	if err := cmd.Execute(); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("top-level lint must flag an audit frontmatter defect, got %v", err)
	}
	if got := out.String(); !strings.Contains(got, "bucket") {
		t.Errorf("lint output should name the bucket defect, got:\n%s", got)
	}
}

// The audit results reach the shared --json envelope, not just human output.
func TestLintJSONIncludesAuditIssues(t *testing.T) {
	root := setupRepo(t)
	p, content := testutil.AuditFixture(root, "open", "2026-01-01-z.md",
		"---\nid: "+testutil.TaskID("2026-01-01-z")+"\nbucket: open\narea: z\ndate: 2026-01-01\n---\n#### S1. t · **Status:** opne\n")
	testutil.Write(t, p, content)

	out, err := runRootRC(t, "-C", root, "--json", "lint")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("dirty audit should make lint --json exit 11, got %v", err)
	}
	var env struct {
		Issues []struct {
			Slug   string                            `json:"slug"`
			Issues []struct{ Field, Message string } `json:"issues"`
		} `json:"issues"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("lint --json is not valid JSON: %v\n%s", err, out)
	}
	for _, r := range env.Issues {
		if r.Slug == "2026-01-01-z" {
			return
		}
	}
	t.Errorf("lint --json envelope should carry the audit's issues, got: %s", out)
}

// A repo whose audits are clean still passes the top-level gate — the fold must
// not turn ordinary audits into noise. Mirrors TestLint_Clean's fixture, which is
// lint-clean by construction, plus one well-formed audit.
func TestLintPassesWithCleanAudit(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("epics/01-e1.md", "---\nstatus: active\npriority: high\ndescription: the epic\n---\n# E1\n")
	goodPath, goodOut := testutil.TaskFixture(root, "ready-to-start", "good.md",
		"---\nid: "+testutil.TaskID("good")+"\nstatus: ready-to-start\nepic: 01-e1\ntier: 2\npriority: high\neffort: 2h\ncreated: 2026-01-01\ntags: [a]\n---\n# Good\n")
	testutil.Write(t, goodPath, goodOut)

	auditPath, auditOut := testutil.AuditFixture(root, "open", "2026-01-01-x.md",
		"---\nid: "+testutil.TaskID("2026-01-01-x")+"\nbucket: open\narea: x\ndate: 2026-01-01\n---\n#### S1. t · **Status:** open\n")
	testutil.Write(t, auditPath, auditOut)

	out := runRoot(t, "-C", root, "lint")
	if !strings.Contains(out, "pass lint") {
		t.Errorf("a clean audit must not fail the top-level lint, got: %q", out)
	}
}

func TestLintReportsEveryAuditSharingAStableID(t *testing.T) {
	root := t.TempDir()
	audits := filepath.Join(root, domain.AuditsDir)
	for _, dir := range []string{filepath.Join(root, domain.TasksDir), audits} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	const shared = "6g7s4k845fsb"
	for _, slug := range []string{"2026-09-07-alpha", "2026-09-07-beta"} {
		content := "---\nid: " + shared + "\nbucket: open\narea: " + slug + "\ndate: 2026-09-07\n---\n# Audit\n"
		if err := os.WriteFile(filepath.Join(audits, shared+"-"+slug+".md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	out, err := runRootRC(t, "-C", root, "lint", "--color=never")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("duplicate audit ids must fail ordinary lint, got %v\n%s", err, out)
	}
	if !strings.Contains(out, "duplicate stable id") || !strings.Contains(out, shared) {
		t.Fatalf("lint did not explain the duplicate audit id:\n%s", out)
	}
	for _, slug := range []string{"2026-09-07-alpha", "2026-09-07-beta"} {
		if !strings.Contains(out, slug) {
			t.Errorf("lint did not identify %s:\n%s", slug, out)
		}
	}
}

func TestLintReportsMalformedAuditInDuplicateIdentity(t *testing.T) {
	root := t.TempDir()
	audits := filepath.Join(root, domain.AuditsDir)
	for _, dir := range []string{filepath.Join(root, domain.TasksDir), audits} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	const shared = "6g7s4k845fsb"
	for _, slug := range []string{"2026-09-07-alpha", "2026-09-07-beta"} {
		content := "---\nid: " + shared + "\nbucket: open\narea: " + slug + "\ndate: 2026-09-07\n---\n# Audit\n"
		if err := os.WriteFile(filepath.Join(audits, shared+"-"+slug+".md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	malformedSlug := "2026-09-07-gamma"
	malformedPath := filepath.Join(audits, shared+"-"+malformedSlug+".md")
	if err := os.WriteFile(malformedPath, []byte("---\nid: [unterminated\n---\n# Audit\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runRootRC(t, "-C", root, "lint", "--color=never")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("mixed readable/unreadable duplicate ids must fail lint, got %v\n%s", err, out)
	}
	for _, want := range []string{
		"shared by 3 docs", "all are unresolvable", malformedSlug,
		shared + "-2026-09-07-alpha.md", shared + "-2026-09-07-beta.md", shared + "-" + malformedSlug + ".md",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("lint output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "both are unresolvable") {
		t.Errorf("three-way collision used two-way wording:\n%s", out)
	}
}
