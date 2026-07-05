package store

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

const editSeed = "---\nstatus: ready-to-start\ndescription: original\ntier: 2\n---\n# Title\n\nbody\n"

func editRepo(t *testing.T) (*FS, string) {
	t.Helper()
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "edit-me.md", editSeed)
	path, _ := testutil.TaskFixture(root, "ready-to-start", "edit-me.md", editSeed)
	return NewFS(root), path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// A valid edit is parsed, written atomically, and reported as changed.
func TestEditTask_ValidEdit_Writes(t *testing.T) {
	fs, path := editRepo(t)
	want := strings.Replace(editSeed, "body", "edited body", 1)
	task, changed, err := fs.EditTask("edit-me", bodyNow, func(cur string, _ error) (string, error) {
		if cur != editSeed {
			t.Errorf("editor got unexpected current content:\n%q", cur)
		}
		return want, nil
	})
	if err != nil {
		t.Fatalf("EditTask: %v", err)
	}
	if !changed {
		t.Error("expected changed=true")
	}
	if task.Slug != "edit-me" {
		t.Errorf("reloaded task slug = %q", task.Slug)
	}
	got := readFile(t, path)
	if !strings.Contains(got, "edited body") {
		t.Errorf("the edit's body change should land:\n%s", got)
	}
	// An accepted edit stamps updated_at to now (2026-06-20) — uniform with set/append.
	if task.Updated != "2026-06-20" || !strings.Contains(got, `updated_at: "2026-06-20"`) {
		t.Errorf("an edit should stamp updated_at to now; task.Updated=%q\n%s", task.Updated, got)
	}
}

// No net change → no write, changed=false (the file is untouched).
func TestEditTask_NoChange_DoesNotWrite(t *testing.T) {
	fs, path := editRepo(t)
	info, _ := os.Stat(path)
	_, changed, err := fs.EditTask("edit-me", bodyNow, func(cur string, _ error) (string, error) {
		return cur, nil // saved as-is
	})
	if err != nil {
		t.Fatalf("EditTask: %v", err)
	}
	if changed {
		t.Error("expected changed=false for an unchanged save")
	}
	if got := readFile(t, path); got != editSeed {
		t.Errorf("file should be untouched, got %q", got)
	}
	if after, _ := os.Stat(path); !after.ModTime().Equal(info.ModTime()) {
		t.Error("unchanged save should not rewrite the file")
	}
}

// A broken edit reopens the editor on the broken content (parse-before-accept):
// the disk never holds the invalid bytes, and a follow-up fix lands.
func TestEditTask_InvalidThenFixed_Reopens(t *testing.T) {
	fs, path := editRepo(t)
	broken := "---\nstatus: ready-to-start\ntier: not-an-int\n---\n# Title\n"
	fixed := strings.Replace(editSeed, "body", "recovered", 1)
	calls := 0
	task, changed, err := fs.EditTask("edit-me", bodyNow, func(cur string, prevErr error) (string, error) {
		calls++
		switch calls {
		case 1:
			if prevErr != nil {
				t.Errorf("first call should have no prevErr, got %v", prevErr)
			}
			return broken, nil
		default:
			if prevErr == nil {
				t.Error("reopen should carry the parse error")
			}
			if cur != broken {
				t.Errorf("reopen should show the broken content, got %q", cur)
			}
			return fixed, nil
		}
	})
	if err != nil {
		t.Fatalf("EditTask: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 editor calls (broken→fixed), got %d", calls)
	}
	if !changed || task.Slug != "edit-me" {
		t.Errorf("expected a changed write, got changed=%v slug=%q", changed, task.Slug)
	}
	got := readFile(t, path)
	if !strings.Contains(got, "recovered") {
		t.Errorf("file should hold the fixed content, got %q", got)
	}
	if task.Updated != "2026-06-20" || !strings.Contains(got, `updated_at: "2026-06-20"`) {
		t.Errorf("the fixed edit should stamp updated_at to now; task.Updated=%q\n%s", task.Updated, got)
	}
}

// Re-saving the same broken content (giving up) returns ErrValidation and never
// writes — the original file survives intact.
func TestEditTask_GiveUpOnBroken_ErrValidation(t *testing.T) {
	fs, path := editRepo(t)
	broken := "---\nstatus: ready-to-start\ntier: not-an-int\n---\n# Title\n"
	_, changed, err := fs.EditTask("edit-me", bodyNow, func(_ string, _ error) (string, error) {
		return broken, nil // same broken bytes every reopen → user gave up
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
	if changed {
		t.Error("a give-up must not report a change")
	}
	if got := readFile(t, path); got != editSeed {
		t.Errorf("original file must survive a give-up, got %q", got)
	}
}

// An editor failure (callback error) propagates and writes nothing.
func TestEditTask_EditorError_Propagates(t *testing.T) {
	fs, path := editRepo(t)
	sentinel := errors.New("editor exploded")
	_, _, err := fs.EditTask("edit-me", bodyNow, func(string, error) (string, error) {
		return "", sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected the editor error, got %v", err)
	}
	if got := readFile(t, path); got != editSeed {
		t.Errorf("file must be untouched after an editor error, got %q", got)
	}
}

// A concurrent in-place edit during the (long) editor window is a compare-and-swap
// conflict: the stale editor's save is rejected, and the concurrent write survives.
func TestEditTask_ConcurrentEditDuringEdit_Conflict(t *testing.T) {
	fs, path := editRepo(t)
	concurrent := strings.Replace(editSeed, "body", "concurrent", 1)
	_, changed, err := fs.EditTask("edit-me", bodyNow, func(cur string, _ error) (string, error) {
		// Simulate a concurrent write landing on the same file mid-edit.
		if err := os.WriteFile(path, []byte(concurrent), 0o644); err != nil {
			t.Fatal(err)
		}
		return strings.Replace(cur, "body", "edited", 1), nil
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict when the file changes mid-edit, got %v", err)
	}
	if changed {
		t.Error("a conflicted edit must not report a change")
	}
	// The concurrent write survives; the stale editor's save was not applied.
	if got := readFile(t, path); !strings.Contains(got, "concurrent") || strings.Contains(got, "edited") {
		t.Errorf("concurrent write should survive the aborted edit, got %q", got)
	}
}

// Opening an already-broken file and saving it unchanged surfaces the parse error
// (ErrValidation) instead of a success report with an empty task.
func TestEditTask_BrokenFileUnchanged_ErrValidation(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "broken.md", "---\nstatus: ready-to-start\ntier: not-an-int\n---\n# x\n")
	fs := NewFS(root)
	_, changed, err := fs.EditTask("broken", bodyNow, func(cur string, _ error) (string, error) {
		return cur, nil // opened to inspect, saved unchanged
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("saving an already-broken file unchanged should surface ErrValidation, got %v", err)
	}
	if changed {
		t.Error("no write should be reported for an unchanged broken file")
	}
}

// Editing the frontmatter status away from the seeded status is an ordinary,
// authoritative in-place edit under the flat layout: the write lands, and Status
// reflects the edit (there is no folder to drift from, so no misfile).
func TestEditTask_StatusEdit_Writes(t *testing.T) {
	fs, path := editRepo(t)
	edited := strings.Replace(editSeed, "status: ready-to-start", "status: completed", 1)
	task, changed, err := fs.EditTask("edit-me", bodyNow, func(string, error) (string, error) {
		return edited, nil
	})
	if err != nil || !changed {
		t.Fatalf("a status edit is an ordinary write: changed=%v err=%v", changed, err)
	}
	if string(task.Status) != "completed" {
		t.Errorf("frontmatter is authoritative, want status completed, got %q", task.Status)
	}
	if got := readFile(t, path); !strings.Contains(got, "status: completed") {
		t.Errorf("the status edit should land in the file at its flat path, got %q", got)
	}
}

// An unknown slug resolves to ErrNotFound before any editor runs.
func TestEditTask_UnknownSlug_NotFound(t *testing.T) {
	fs, _ := editRepo(t)
	ran := false
	_, _, err := fs.EditTask("nope", bodyNow, func(string, error) (string, error) {
		ran = true
		return "", nil
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if ran {
		t.Error("editor must not run for an unresolvable slug")
	}
}
