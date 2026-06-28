package store

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

const auditEditSeed = "---\narea: store\ndate: \"2026-06-20\"\n---\n# Audit\n\n#### H1. thing  · **Status:** open\n\nbody\n"

func auditEditRepo(t *testing.T) (*FS, string) {
	t.Helper()
	root := t.TempDir()
	writeAudit(t, root, "open", "2026-06-20-store.md", auditEditSeed)
	return NewFS(root), filepath.Join(root, domain.AuditsDir, "open", "2026-06-20-store.md")
}

// EditAudit mirrors EditTask: a valid edit parses, writes atomically, reports changed.
func TestEditAudit_ValidEdit_Writes(t *testing.T) {
	fs, path := auditEditRepo(t)
	want := strings.Replace(auditEditSeed, "body", "edited body", 1)
	a, changed, err := fs.EditAudit("2026-06-20-store", func(cur string, _ error) (string, error) {
		if cur != auditEditSeed {
			t.Errorf("editor got unexpected content:\n%q", cur)
		}
		return want, nil
	})
	if err != nil || !changed {
		t.Fatalf("EditAudit: changed=%v err=%v", changed, err)
	}
	if a.Slug != "2026-06-20-store" || a.Bucket != domain.AuditOpen {
		t.Errorf("reloaded audit wrong: %+v", a)
	}
	if got := readFile(t, path); got != want {
		t.Errorf("file not written:\n got %q\nwant %q", got, want)
	}
}

// Saving a broken edit unchanged (the user gave up) is ErrValidation, and the
// original audit survives intact — parse-before-accept never lands invalid bytes.
func TestEditAudit_GiveUpOnBroken_ErrValidation(t *testing.T) {
	fs, path := auditEditRepo(t)
	broken := "---\narea: store\n  bad: : indent\n---\n# x\n"
	_, changed, err := fs.EditAudit("2026-06-20-store", func(string, error) (string, error) {
		return broken, nil // same broken bytes every reopen → user gave up
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("a give-up should be ErrValidation, got %v", err)
	}
	if changed || readFile(t, path) != auditEditSeed {
		t.Error("the original audit must survive a give-up")
	}
}

// Unknown slug resolves to ErrNotFound before any editor runs.
func TestEditAudit_UnknownSlug_NotFound(t *testing.T) {
	fs, _ := auditEditRepo(t)
	ran := false
	_, _, err := fs.EditAudit("nope", func(string, error) (string, error) { ran = true; return "", nil })
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if ran {
		t.Error("editor must not run for an unresolvable slug")
	}
}

// AppendAuditBody appends a section AND preserves the frontmatter — crucially WITHOUT
// stamping updated_at (the task body path stamps it; audits have no such field, so
// their date stays the immutable slug).
func TestAppendAuditBody_AppendsNoUpdatedAtStamp(t *testing.T) {
	fs, path := auditEditRepo(t)
	a, body, err := fs.AppendAuditBody("2026-06-20-store", "#### M2. new  · **Status:** open", false)
	if err != nil {
		t.Fatalf("AppendAuditBody: %v", err)
	}
	got := readFile(t, path)
	if strings.Contains(got, "updated_at") {
		t.Errorf("audit append must NOT stamp updated_at:\n%s", got)
	}
	if a.Area != "store" || a.Date != "2026-06-20" {
		t.Errorf("frontmatter not preserved: area=%q date=%q", a.Area, a.Date)
	}
	if !strings.Contains(got, "#### M2. new") || !strings.Contains(body, "#### M2. new") {
		t.Errorf("appended section missing:\n%s", got)
	}
	if !strings.Contains(got, "#### H1. thing") {
		t.Error("append clobbered the existing body")
	}
}

// Dry-run runs every check + returns the would-be body, but writes nothing.
func TestAppendAuditBody_DryRun_NoWrite(t *testing.T) {
	fs, path := auditEditRepo(t)
	before := readFile(t, path)
	_, body, err := fs.AppendAuditBody("2026-06-20-store", "PREVIEW", true)
	if err != nil {
		t.Fatalf("dry-run AppendAuditBody: %v", err)
	}
	if !strings.Contains(body, "PREVIEW") {
		t.Errorf("dry-run should return the would-be body, got %q", body)
	}
	if readFile(t, path) != before {
		t.Error("dry-run must not write")
	}
}

// Unknown slug → ErrNotFound, no write.
func TestAppendAuditBody_UnknownSlug_NotFound(t *testing.T) {
	fs, _ := auditEditRepo(t)
	if _, _, err := fs.AppendAuditBody("nope", "x", false); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
