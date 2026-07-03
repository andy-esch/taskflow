package store

import (
	"os"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Frontmatter bucket wins over the folder; the folder is captured as FolderBucket, and
// a missing/foreign bucket falls back to the folder (not misfiled).
func TestParseAudit_FrontmatterBucketIsAuthoritative(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "2026-01-02-drift.md",
		"---\nid: 6fjjt6s9ttz0\nbucket: closed\narea: x\ndate: 2026-01-02\n---\n#### H1. t  · **Status:** fixed\n")
	writeAudit(t, root, "open", "2026-01-02-legacy.md",
		"---\nbucket: archived\narea: y\ndate: 2026-01-02\n---\n# x\n") // foreign bucket word
	writeAudit(t, root, "open", "2026-01-02-nobucket.md",
		"---\narea: z\ndate: 2026-01-02\n---\n# x\n") // no bucket frontmatter

	audits, _, err := NewFS(root).ListAudits()
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]domain.Audit{}
	for _, a := range audits {
		by[a.Slug] = a
	}

	if d := by["2026-01-02-drift"]; d.Bucket != domain.AuditClosed || d.FolderBucket != domain.AuditOpen || !d.Misfiled() {
		t.Errorf("drift: bucket=%q folder=%q misfiled=%v (want closed/open/true)", d.Bucket, d.FolderBucket, d.Misfiled())
	}
	if l := by["2026-01-02-legacy"]; l.Bucket != domain.AuditOpen || l.Misfiled() {
		t.Errorf("foreign bucket should fall back to the folder: bucket=%q misfiled=%v", l.Bucket, l.Misfiled())
	}
	if n := by["2026-01-02-nobucket"]; n.Bucket != domain.AuditOpen || n.Misfiled() {
		t.Errorf("no bucket frontmatter should fall back to the folder: bucket=%q misfiled=%v", n.Bucket, n.Misfiled())
	}
}

// lint --fix backfills bucket: <dir> into an audit that lacks it (pre-Phase-A state),
// idempotently.
func TestFixFrontmatter_BackfillsMissingAuditBucket(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "closed", "2026-01-02-nb.md",
		"---\nid: 6fjjt6s9ttz1\narea: x\ndate: 2026-01-02\n---\n# x\n")

	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(root + "/audits/closed/2026-01-02-nb.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "bucket: closed") {
		t.Errorf("bucket: closed not backfilled:\n%s", b)
	}
	res2, err := NewFS(root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range res2 {
		if strings.Contains(r.Path, "nb") {
			t.Errorf("bucket backfill should be idempotent, got a repeat: %+v", r)
		}
	}
}

// lint --fix relocates a misfiled audit (frontmatter bucket ≠ folder) to match, without
// rewriting the bucket.
func TestFixFrontmatter_MovesMisfiledAudit(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "2026-01-02-mv.md",
		"---\nid: 6fjjt6s9ttz2\nbucket: deferred\narea: x\ndate: 2026-01-02\n---\n#### H1. t  · **Status:** deferred\n")

	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root + "/audits/open/2026-01-02-mv.md"); !os.IsNotExist(err) {
		t.Error("misfiled audit should have moved out of open/")
	}
	b, err := os.ReadFile(root + "/audits/deferred/2026-01-02-mv.md")
	if err != nil {
		t.Fatalf("misfiled audit should be relocated to deferred/: %v", err)
	}
	if !strings.Contains(string(b), "bucket: deferred") {
		t.Errorf("the relocation must not rewrite the frontmatter bucket:\n%s", b)
	}
}

// MoveAudit rewrites the authoritative bucket frontmatter, not just the file location.
func TestMoveAudit_WritesBucketFrontmatter(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "2026-01-02-c.md",
		"---\nid: 6fjjt6s9ttz3\nbucket: open\narea: x\ndate: 2026-01-02\n---\n#### H1. t  · **Status:** fixed\n")

	a, err := NewFS(root).MoveAudit("2026-01-02-c", domain.AuditClosed, false)
	if err != nil {
		t.Fatal(err)
	}
	if a.Bucket != domain.AuditClosed {
		t.Errorf("moved audit bucket = %q, want closed", a.Bucket)
	}
	b, err := os.ReadFile(root + "/audits/closed/2026-01-02-c.md")
	if err != nil {
		t.Fatalf("file should be relocated to closed/: %v", err)
	}
	if !strings.Contains(string(b), "bucket: closed") {
		t.Errorf("MoveAudit must rewrite the bucket frontmatter:\n%s", b)
	}
}

// lint --fix must NOT relocate a misfiled audit into a non-open bucket while it still has
// open findings — that would create the bucket↔state violation MoveAudit refuses (and the
// re-lint could never repair). It stays put, flagged.
func TestFixFrontmatter_MisfiledAuditWithOpenFindings_NotRelocated(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "2026-01-02-of.md",
		"---\nid: 6fjjt6s9ttz4\nbucket: closed\narea: x\ndate: 2026-01-02\n---\n#### H1. t  · **Status:** open\n")

	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root + "/audits/open/2026-01-02-of.md"); err != nil {
		t.Errorf("audit with open findings must not be relocated into closed/: %v", err)
	}
	if _, err := os.Stat(root + "/audits/closed/2026-01-02-of.md"); !os.IsNotExist(err) {
		t.Error("no closed/ copy should be created (bucket↔state gate)")
	}
}

// A foreign/legacy bucket word is NOT clobbered by the backfill — only a truly absent
// bucket is filled (a bad value is the re-lint's job to surface, not the repair's to eat).
func TestFixFrontmatter_ForeignBucketNotClobbered(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "2026-01-02-fb.md",
		"---\nid: 6fjjt6s9ttz5\nbucket: archived\narea: x\ndate: 2026-01-02\n---\n# x\n")

	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(root + "/audits/open/2026-01-02-fb.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "bucket: archived") {
		t.Errorf("a foreign bucket word must not be overwritten by backfill:\n%s", b)
	}
}
