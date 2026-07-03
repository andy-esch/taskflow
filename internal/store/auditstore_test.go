package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func writeAudit(t *testing.T, root, bucket, name, content string) {
	t.Helper()
	testutil.Write(t, filepath.Join(root, domain.AuditsDir, bucket, name), content)
}

// TestFS_ListAudits_MissingFrontmatterIsLoud: a fence-less audit file is surfaced
// as a loud FileProblem naming the valid shape, not silently read as an empty
// audit (empty area/date) that slips through.
func TestFS_ListAudits_MissingFrontmatterIsLoud(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "notes.md", "# Some audit notes\n\nno frontmatter\n")

	audits, problems, err := NewFS(root).ListAudits()
	if err != nil {
		t.Fatal(err)
	}
	if len(audits) != 0 {
		t.Errorf("a fence-less audit must not parse as an audit, got %+v", audits)
	}
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "missing frontmatter") || !strings.Contains(problems[0].Message, "schema audit") {
		t.Errorf("want one loud, shape-naming problem, got %+v", problems)
	}
}

func TestFS_ListAudits_FindingCounts(t *testing.T) {
	root := t.TempDir()
	body := "# Audit\n\n#### H1. thing  · **Status:** open\n\nblah\n\n#### M2. other  · **Status:** fixed 2026-01-01\n"
	writeAudit(t, root, "open", "a.md", "---\narea: dispatcher\ndate: 2026-06-01\n---\n"+body)

	audits, problems, err := NewFS(root).ListAudits()
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Fatalf("unexpected problems: %+v", problems)
	}
	if len(audits) != 1 {
		t.Fatalf("want 1 audit, got %d", len(audits))
	}
	a := audits[0]
	if a.Slug != "a" || a.Bucket != domain.AuditOpen || a.Area != "dispatcher" {
		t.Errorf("metadata wrong: %+v", a)
	}
	if a.Findings != 2 || a.OpenFindings != 1 {
		t.Errorf("findings=%d open=%d, want 2/1", a.Findings, a.OpenFindings)
	}
}

func TestFS_FindingCounts_IgnoresFencesAndOpenIsh(t *testing.T) {
	root := t.TempDir()
	body := "# Audit\n\n" +
		"#### H1. real  · **Status:** open\n\n" + // real open finding
		"#### M2. doc-ish  · **Status:** open-ish\n\n" + // NOT open (open-ish)
		"```\n#### S9. example in a code fence  · **Status:** open\n```\n\n" + // fenced: not counted
		"#### L3. done  · **Status:** fixed\n"
	writeAudit(t, root, "open", "b.md", "---\narea: x\n---\n"+body)

	audits, _, err := NewFS(root).ListAudits()
	if err != nil {
		t.Fatal(err)
	}
	a := audits[0]
	// 3 real findings (H1, M2, L3 — the fenced S9 is excluded); 1 open (H1 only).
	if a.Findings != 3 {
		t.Errorf("findings = %d, want 3 (fenced example excluded)", a.Findings)
	}
	if a.OpenFindings != 1 {
		t.Errorf("open = %d, want 1 (open-ish and fenced excluded)", a.OpenFindings)
	}
}

// TestFS_ListAuditsWithFindings pins the single-scan port (H2): it returns the
// same per-audit tally ListAudits does AND the findings parsed from that same body
// read, in document order — so Summary needs no GetAuditByPath re-read.
func TestFS_ListAuditsWithFindings(t *testing.T) {
	root := t.TempDir()
	body := "# Audit\n\n#### H1. open thing  · **Status:** open\n\n#### M2. fixed thing  · **Status:** fixed 2026-01-01\n"
	writeAudit(t, root, "open", "a.md", "---\narea: dispatcher\ndate: 2026-06-01\n---\n"+body)

	got, problems, err := NewFS(root).ListAuditsWithFindings()
	if err != nil || len(problems) != 0 {
		t.Fatalf("ListAuditsWithFindings: %v / %+v", err, problems)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 audit, got %d", len(got))
	}
	a := got[0]
	// Tally is identical to ListAudits'.
	if a.Audit.Slug != "a" || a.Audit.Findings != 2 || a.Audit.OpenFindings != 1 {
		t.Errorf("audit tally wrong: %+v", a.Audit)
	}
	// Findings come back parsed, in document order (H1 then M2).
	if len(a.Findings) != 2 || a.Findings[0].Code != "H1" || a.Findings[1].Code != "M2" {
		t.Errorf("findings wrong: %+v", a.Findings)
	}
}

func TestFS_MoveAudit(t *testing.T) {
	root := t.TempDir()
	// No open findings, so the bucket↔state invariant permits closing.
	writeAudit(t, root, "open", "x.md", "---\narea: a\n---\n#### H1. t  · **Status:** fixed\n")

	a, err := NewFS(root).MoveAudit("x", domain.AuditClosed, false)
	if err != nil {
		t.Fatal(err)
	}
	if a.Bucket != domain.AuditClosed {
		t.Errorf("bucket = %s", a.Bucket)
	}
	if _, err := os.Stat(filepath.Join(root, "audits", "open", "x.md")); !os.IsNotExist(err) {
		t.Error("old file should be gone")
	}
	if _, err := os.Stat(filepath.Join(root, "audits", "closed", "x.md")); err != nil {
		t.Errorf("closed file missing: %v", err)
	}
	// A bucket move is a pure relocation — it touches neither frontmatter nor body, so
	// it must NOT stamp updated_at (unlike edit/append). The directory carries the
	// state change; the content date stays put.
	moved, _ := os.ReadFile(filepath.Join(root, "audits", "closed", "x.md"))
	if a.Updated != "" || strings.Contains(string(moved), "updated_at") {
		t.Errorf("a bucket move must not stamp updated_at:\n%s", moved)
	}
}

func TestFS_GetAudit_NotFound(t *testing.T) {
	if _, _, err := NewFS(t.TempDir()).GetAudit("nope"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// TestFS_GetAuditByPath pins M16: a read-by-path returns the same audit+body as
// GetAudit and derives the bucket from the parent directory (not the frontmatter).
func TestFS_GetAuditByPath(t *testing.T) {
	root := t.TempDir()
	body := "# Audit\n\n#### H1. t  · **Status:** open\n"
	writeAudit(t, root, "deferred", "2026-06-01-x.md", "---\narea: dispatcher\ndate: 2026-06-01\n---\n"+body)
	fs := NewFS(root)

	// Discover the path the way the sweeps do: ListAudits populates .Path.
	audits, _, err := fs.ListAudits()
	if err != nil || len(audits) != 1 {
		t.Fatalf("ListAudits: %v (n=%d)", err, len(audits))
	}
	wantPath := audits[0].Path

	a, gotBody, err := fs.GetAuditByPath(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	if a.Slug != "2026-06-01-x" || a.Bucket != domain.AuditDeferred || a.Area != "dispatcher" {
		t.Errorf("metadata wrong (bucket must come from the parent dir): %+v", a)
	}
	if a.Findings != 1 || a.OpenFindings != 1 {
		t.Errorf("findings=%d open=%d, want 1/1", a.Findings, a.OpenFindings)
	}
	// Body matches the GetAudit (slug-resolved) read of the same file.
	if _, slugBody, err := fs.GetAudit("2026-06-01-x"); err != nil || gotBody != slugBody {
		t.Errorf("by-path body diverges from by-slug: %q vs %q (%v)", gotBody, slugBody, err)
	}
}

// TestFS_GetAuditByPath_RejectsNonBucketDir pins that a path outside
// audits/<bucket>/ is rejected (ErrValidation), not silently mis-bucketed.
func TestFS_GetAuditByPath_RejectsNonBucketDir(t *testing.T) {
	root := t.TempDir()
	stray := filepath.Join(root, "audits", "bogus", "x.md")
	testutil.Write(t, stray, "---\narea: a\n---\n# x\n")
	if _, _, err := NewFS(root).GetAuditByPath(stray); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("want ErrValidation for a non-bucket parent dir, got %v", err)
	}
}
