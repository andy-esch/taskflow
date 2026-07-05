package store

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// auditEditSeed carries an explicit `bucket:` so AuditFixture writes it verbatim
// (bucket is authoritative in frontmatter under the flat layout, ADR-0003 §4) —
// the editor-content assertions below compare against these exact bytes.
const auditEditSeed = "---\nbucket: open\narea: store\ndate: \"2026-06-20\"\n---\n# Audit\n\n#### H1. thing  · **Status:** open\n\nbody\n"

// auditEditNow is deliberately LATER than the seed's immutable date (2026-06-20),
// so a stamped updated_at is visibly distinct from the slug's date.
var auditEditNow = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

func auditEditRepo(t *testing.T) (*FS, string) {
	t.Helper()
	root := t.TempDir()
	path, out := testutil.AuditFixture(root, "open", "2026-06-20-store.md", auditEditSeed)
	testutil.Write(t, path, out)
	return NewFS(root), path
}

// EditAudit mirrors EditTask: a valid edit parses, writes atomically, reports changed,
// and stamps updated_at — while the audit's `date` (its immutable slug) stays put.
func TestEditAudit_ValidEdit_Writes(t *testing.T) {
	fs, path := auditEditRepo(t)
	want := strings.Replace(auditEditSeed, "body", "edited body", 1)
	a, changed, err := fs.EditAudit("2026-06-20-store", auditEditNow, func(cur string, _ error) (string, error) {
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
	got := readFile(t, path)
	if !strings.Contains(got, "edited body") {
		t.Errorf("the edit's body change should land:\n%s", got)
	}
	if a.Updated != "2026-07-01" || !strings.Contains(got, `updated_at: "2026-07-01"`) {
		t.Errorf("an audit edit should stamp updated_at to now; a.Updated=%q\n%s", a.Updated, got)
	}
	if a.Date != "2026-06-20" {
		t.Errorf("the audit date is immutable (the slug); got %q", a.Date)
	}
}

// Saving a broken edit unchanged (the user gave up) is ErrValidation, and the
// original audit survives intact — parse-before-accept never lands invalid bytes.
func TestEditAudit_GiveUpOnBroken_ErrValidation(t *testing.T) {
	fs, path := auditEditRepo(t)
	broken := "---\narea: store\n  bad: : indent\n---\n# x\n"
	_, changed, err := fs.EditAudit("2026-06-20-store", bodyNow, func(string, error) (string, error) {
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
	_, _, err := fs.EditAudit("nope", bodyNow, func(string, error) (string, error) { ran = true; return "", nil })
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if ran {
		t.Error("editor must not run for an unresolvable slug")
	}
}

// AppendAuditBody appends a section, preserves the rest of the frontmatter, AND
// stamps updated_at (uniform with the task body path) — while the audit's `date`
// (its immutable slug) stays put.
func TestAppendAuditBody_StampsUpdatedAt(t *testing.T) {
	fs, path := auditEditRepo(t)
	a, body, err := fs.AppendAuditBody("2026-06-20-store", "#### M2. new  · **Status:** open", auditEditNow, false)
	if err != nil {
		t.Fatalf("AppendAuditBody: %v", err)
	}
	got := readFile(t, path)
	if a.Updated != "2026-07-01" || !strings.Contains(got, `updated_at: "2026-07-01"`) {
		t.Errorf("audit append should stamp updated_at to now; a.Updated=%q\n%s", a.Updated, got)
	}
	if a.Area != "store" || a.Date != "2026-06-20" {
		t.Errorf("frontmatter not preserved (date is immutable): area=%q date=%q", a.Area, a.Date)
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
	_, body, err := fs.AppendAuditBody("2026-06-20-store", "PREVIEW", bodyNow, true)
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
	if _, _, err := fs.AppendAuditBody("nope", "x", bodyNow, false); !errors.Is(err, domain.ErrNotFound) {
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
	a, changed, err := fs.EditAudit("2026-06-20-store", auditEditNow, func(cur string, prevErr error) (string, error) {
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
	got := readFile(t, path)
	if a.Slug != "2026-06-20-store" || !strings.Contains(got, "recovered") {
		t.Error("the fixed content should land")
	}
	if a.Updated != "2026-07-01" || !strings.Contains(got, `updated_at: "2026-07-01"`) {
		t.Errorf("the fixed edit should stamp updated_at; a.Updated=%q\n%s", a.Updated, got)
	}
}

// No net change → no write, changed=false, the file untouched.
func TestEditAudit_NoChange_DoesNotWrite(t *testing.T) {
	fs, path := auditEditRepo(t)
	info, _ := os.Stat(path)
	_, changed, err := fs.EditAudit("2026-06-20-store", bodyNow, func(cur string, _ error) (string, error) { return cur, nil })
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

// EditAudit compare-and-swaps: a concurrent in-place edit during the editor window
// (audit close/reopen/defer now rewrite the frontmatter in place — the path never
// changes under the flat layout) is a version conflict, not a silent clobber of the
// winning edit. Mirrors the task edit CAS test.
func TestEditAudit_EditedDuringEdit_Conflict(t *testing.T) {
	fs, path := auditEditRepo(t)
	_, changed, err := fs.EditAudit("2026-06-20-store", bodyNow, func(cur string, _ error) (string, error) {
		// Simulate a concurrent `audit close` mid-edit: an in-place frontmatter rewrite.
		testutil.Write(t, path, strings.Replace(cur, "bucket: open", "bucket: closed", 1))
		return strings.Replace(cur, "body", "edited", 1), nil
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent in-place edit should be ErrConflict, got %v", err)
	}
	if changed {
		t.Error("a conflicted edit must not report a change")
	}
	if got := readFile(t, path); !strings.Contains(got, "bucket: closed") || strings.Contains(got, "edited body") {
		t.Errorf("the losing edit must not clobber the concurrent in-place edit:\n%s", got)
	}
}

// AppendAuditBody compare-and-swaps in the write window too: a concurrent in-place edit
// there is a conflict, never a clobber. Mirrors TestAppendAuditBody_ConflictsOnConcurrentContentEdit.
func TestAppendAuditBody_EditedDuringWrite_Conflict(t *testing.T) {
	fs, path := auditEditRepo(t)
	testHookBeforeBodyWrite = func() {
		testutil.Write(t, path, strings.Replace(auditEditSeed, "bucket: open", "bucket: closed", 1))
	}
	defer func() { testHookBeforeBodyWrite = nil }()
	if _, _, err := fs.AppendAuditBody("2026-06-20-store", "x", bodyNow, false); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a concurrent in-place edit in the write window should be ErrConflict, got %v", err)
	}
	if got := readFile(t, path); !strings.Contains(got, "bucket: closed") {
		t.Errorf("the losing append must not clobber the concurrent in-place edit:\n%s", got)
	}
}

// Appending to an audit whose frontmatter won't parse errors rather than mangling it.
func TestAppendAuditBody_BrokenFrontmatter_Errors(t *testing.T) {
	root := t.TempDir()
	path, out := testutil.AuditFixture(root, "open", "2026-06-20-broken.md", "---\narea: store\nno closing fence\n")
	testutil.Write(t, path, out)
	if _, _, err := NewFS(root).AppendAuditBody("2026-06-20-broken", "y", bodyNow, false); err == nil {
		t.Fatal("appending to a file with unterminated frontmatter should error")
	}
}
