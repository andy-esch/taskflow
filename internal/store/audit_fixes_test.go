package store

// Regression tests for the 2026-06-22 code-quality audit fixes (store package):
// H1 exit-code classification, H2 active-task write invariants, M2 file-mode
// preservation, M5 MoveAudit compare-and-swap, L4 fence whitespace tolerance.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// H1: a malformed file on the single-item read path must classify as
// ErrValidation (exit 11), matching the write paths — not fall through to exit 1.
func TestFS_GetTask_MalformedFrontmatterIsValidation(t *testing.T) {
	root := t.TempDir()
	// tier as a quoted string survives the byte split but fails the strict typed
	// decode — a malformed-frontmatter parse error.
	writeTask(t, root, "ready-to-start", "alpha.md",
		"---\nstatus: ready-to-start\ntier: \"4\"\ntags: [seed]\n---\n# Alpha\n")
	if _, _, err := NewFS(root).GetTask("alpha"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("malformed frontmatter on the read path must be ErrValidation (exit 11), got %v", err)
	}
}

// H2: `set` must not be able to write a file the linter rejects — emptying tags
// on an active task is refused, with nothing written.
func TestFS_SetFields_RejectsEmptyTagsOnActiveTask(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "alpha.md",
		"---\nstatus: ready-to-start\ndescription: a\ntags: [seed]\n---\n# Alpha\n")
	_, err := NewFS(root).SetFields("alpha", map[string]any{"tags": []string{}}, false)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("emptying tags on an active task must be rejected, got %v", err)
	}
	if tk, _, _ := NewFS(root).GetTask("alpha"); len(tk.Tags) != 1 || tk.Tags[0] != "seed" {
		t.Errorf("a rejected SetFields must not have written (tags=%v)", tk.Tags)
	}
}

// H2: clearing the description of a next-up/in-progress task is refused (lint
// requires one there).
func TestFS_SetFields_RejectsClearedDescriptionOnActive(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "in-progress", "alpha.md",
		"---\nstatus: in-progress\ndescription: a\ntags: [seed]\n---\n# Alpha\n")
	if _, err := NewFS(root).SetFields("alpha", map[string]any{"description": ""}, false); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("clearing the description of an in-progress task must be rejected, got %v", err)
	}
}

// H2 (gating): archived tasks are NOT held to the active-only field rules, so
// editing an unrelated field on an untagged completed task still succeeds — the
// invariant must mirror Lint (active-only), not over-enforce.
func TestFS_SetFields_AllowsUntaggedCompletedTask(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "completed", "done.md",
		"---\nstatus: completed\ndescription: d\n---\n# Done\n")
	if _, err := NewFS(root).SetFields("done", map[string]any{"priority": "high"}, false); err != nil {
		t.Fatalf("SetFields on an untagged completed task should succeed, got %v", err)
	}
}

// M2: overwriting a user-restricted file must preserve its mode, not silently
// widen it to the caller's 0644; a brand-new file falls back to the passed perm.
func TestWriteFileAtomic_PreservesExistingMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.md")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(path, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(path); err != nil {
		t.Fatal(err)
	} else if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("writeFileAtomic widened mode to %#o, want preserved 0600", got)
	}
	newPath := filepath.Join(dir, "new.md")
	if err := writeFileAtomic(newPath, []byte("x"), 0o640); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(newPath); err != nil {
		t.Fatal(err)
	} else if got := info.Mode().Perm(); got != 0o640 {
		t.Errorf("new file mode = %#o, want fallback 0640", got)
	}
}

// M5: an audit relocated between MoveAudit's resolve and its rename must produce
// ErrConflict (exit 14 retry signal) with no duplicate left behind — the audit
// sibling of the task SetFields/Move compare-and-swap guard.
func TestFS_MoveAudit_ConflictsWhenMovedConcurrently(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "a1.md", "---\narea: store\ndate: \"2026-06-01\"\n---\n# Audit: store\n")
	fs := NewFS(root)

	openPath := filepath.Join(root, "audits", "open", "a1.md")
	closedDir := filepath.Join(root, "audits", "closed")
	testHookBeforeMoveAuditWrite = func() {
		if err := os.MkdirAll(closedDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(openPath, filepath.Join(closedDir, "a1.md")); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { testHookBeforeMoveAuditWrite = nil }()

	_, err := fs.MoveAudit("a1", domain.AuditDeferred, false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want ErrConflict for a concurrently-moved audit, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "audits", "deferred", "a1.md")); statErr == nil {
		t.Error("a conflicted audit move must not write into the target bucket")
	}
}

// M4: closing/deferring an audit that still has open findings must be refused
// (the bucket↔state invariant `audit lint` enforces), with nothing moved.
func TestFS_MoveAudit_RejectsOpenFindings(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "x.md", "---\narea: a\n---\n#### H1. t  · **Status:** open\n")
	_, err := NewFS(root).MoveAudit("x", domain.AuditClosed, false)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("closing an audit with open findings must be rejected, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "audits", "open", "x.md")); statErr != nil {
		t.Error("a rejected move must leave the audit in its original bucket")
	}
	// Deferring (also a non-open bucket) is refused for the same reason.
	if _, err := NewFS(root).MoveAudit("x", domain.AuditDeferred, false); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("deferring an audit with open findings must be rejected, got %v", err)
	}
}

// L5: when a SetFields update fails to reload because the file was ALREADY broken
// the same way (pre-existing duplicate keys), blame the file, not the user's update.
func TestFS_SetFields_PreexistingCorruptionAttributedToFile(t *testing.T) {
	root := t.TempDir()
	// Duplicate top-level key: yaml.Node accepts it, but the typed decode rejects it.
	writeTask(t, root, "ready-to-start", "dup.md",
		"---\nstatus: ready-to-start\ntier: 2\ntier: 3\ntags: [x]\n---\n# Dup\n")
	_, err := NewFS(root).SetFields("dup", map[string]any{"priority": "high"}, false)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "already has malformed frontmatter") {
		t.Errorf("error should blame the pre-existing corruption, not the update: %v", err)
	}
}

// L4: a closing fence with trailing whitespace (a common editor artifact) must
// be recognized, not reported as an unterminated-frontmatter problem.
func TestSplitFrontmatter_ToleratesTrailingWhitespaceFence(t *testing.T) {
	content := []byte("---\nstatus: open\n--- \nbody\n")
	fm, body, err := splitFrontmatterStrict(content)
	if err != nil {
		t.Fatalf("a closing fence with trailing whitespace must be recognized, got %v", err)
	}
	if !strings.Contains(string(fm), "status: open") {
		t.Errorf("frontmatter not captured: %q", fm)
	}
	if string(body) != "body\n" {
		t.Errorf("body = %q, want %q", body, "body\n")
	}
}
