package store

import (
	"errors"
	"os"
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

// A broken edit reopens the editor on the broken content (parse-before-accept); a
// follow-up fix lands. Mirrors TestEditTask_InvalidThenFixed_Reopens.
func TestEditAudit_InvalidThenFixed_Reopens(t *testing.T) {
	fs, path := auditEditRepo(t)
	broken := "---\narea: store\n  bad: : indent\n---\n# x\n"
	fixed := strings.Replace(auditEditSeed, "body", "recovered", 1)
	calls := 0
	a, changed, err := fs.EditAudit("2026-06-20-store", func(cur string, prevErr error) (string, error) {
		calls++
		if calls == 1 {
			return broken, nil
		}
		if prevErr == nil {
			t.Error("reopen should carry the parse error")
		}
		if cur != broken {
			t.Errorf("reopen should show the broken content, got %q", cur)
		}
		return fixed, nil
	})
	if err != nil || !changed || calls != 2 {
		t.Fatalf("expected a broken→fixed reopen (2 calls), got changed=%v calls=%d err=%v", changed, calls, err)
	}
	if a.Slug != "2026-06-20-store" || readFile(t, path) != fixed {
		t.Error("the fixed content should land")
	}
}

// No net change → no write, changed=false, the file untouched.
func TestEditAudit_NoChange_DoesNotWrite(t *testing.T) {
	fs, path := auditEditRepo(t)
	info, _ := os.Stat(path)
	_, changed, err := fs.EditAudit("2026-06-20-store", func(cur string, _ error) (string, error) { return cur, nil })
	if err != nil || changed {
		t.Fatalf("unchanged save: changed=%v err=%v", changed, err)
	}
	if readFile(t, path) != auditEditSeed {
		t.Error("file should be untouched")
	}
	if after, _ := os.Stat(path); !after.ModTime().Equal(info.ModTime()) {
		t.Error("an unchanged save should not rewrite the file")
	}
}

// EditAudit's signature feature vs EditEpic: a concurrent bucket move (audit
// close/reopen/defer) during the editor window is a compare-and-swap conflict, not a
// silent resurrection of the slug at its old bucket. Mirrors the task relocation test.
func TestEditAudit_RelocatedDuringEdit_Conflict(t *testing.T) {
	fs, path := auditEditRepo(t)
	moved := strings.Replace(path, "open", "closed", 1)
	_, changed, err := fs.EditAudit("2026-06-20-store", func(cur string, _ error) (string, error) {
		_ = os.MkdirAll(filepath.Dir(moved), 0o755)
		_ = os.Rename(path, moved) // simulate a concurrent `audit close` mid-edit
		return strings.Replace(cur, "body", "edited", 1), nil
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a relocation mid-edit should be ErrConflict, got %v", err)
	}
	if changed {
		t.Error("a conflicted edit must not report a change")
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Error("the old path must not be resurrected")
	}
}

// AppendAuditBody compare-and-swaps in the write window too: a bucket move there is a
// conflict, never a resurrection. Mirrors TestEditBody_RelocatedDuringWrite_Conflict.
func TestAppendAuditBody_RelocatedDuringWrite_Conflict(t *testing.T) {
	fs, path := auditEditRepo(t)
	moved := strings.Replace(path, "open", "closed", 1)
	testHookBeforeBodyWrite = func() {
		_ = os.MkdirAll(filepath.Dir(moved), 0o755)
		_ = os.Rename(path, moved)
	}
	defer func() { testHookBeforeBodyWrite = nil }()
	if _, _, err := fs.AppendAuditBody("2026-06-20-store", "x", false); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a relocation in the write window should be ErrConflict, got %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Error("the old path must not be resurrected")
	}
}

// Appending to an audit whose frontmatter won't parse errors rather than mangling it.
func TestAppendAuditBody_BrokenFrontmatter_Errors(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "2026-06-20-broken.md", "---\narea: store\nno closing fence\n")
	if _, _, err := NewFS(root).AppendAuditBody("2026-06-20-broken", "y", false); err == nil {
		t.Fatal("appending to a file with unterminated frontmatter should error")
	}
}
