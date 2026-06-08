package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func writeAudit(t *testing.T, root, bucket, name, content string) {
	t.Helper()
	dir := filepath.Join(root, "audits", bucket)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
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

func TestFS_MoveAudit(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "x.md", "---\narea: a\n---\n#### H1. t  · **Status:** open\n")

	a, err := NewFS(root).MoveAudit("x", domain.AuditClosed)
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
}

func TestFS_GetAudit_NotFound(t *testing.T) {
	if _, _, err := NewFS(t.TempDir()).GetAudit("nope"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}
