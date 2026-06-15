package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TestFS_Move_RejectsUnreloadableWithoutMoving pins parse-before-commit on the
// move path: a task whose frontmatter wouldn't read back after the update must
// fail with the file still in its ORIGINAL directory. (Previously the parse ran
// after the write+remove, so the move succeeded on disk while reporting failure
// — a phantom failure a retrying agent would act on.)
func TestFS_Move_RejectsUnreloadableWithoutMoving(t *testing.T) {
	root := t.TempDir()
	// tier as a quoted string survives the surgical status update but fails the
	// strict typed decode.
	const original = "---\nstatus: ready-to-start\ntier: \"4\"\n---\n# Alpha\n"
	writeTask(t, root, "ready-to-start", "alpha.md", original)

	_, err := NewFS(root).Move("alpha", domain.StatusInProgress, time.Now(), false)
	if err == nil {
		t.Fatal("want an error for a move that wouldn't reload")
	}
	oldPath := filepath.Join(root, "tasks", "ready-to-start", "alpha.md")
	newPath := filepath.Join(root, "tasks", "in-progress", "alpha.md")
	if _, statErr := os.Stat(oldPath); statErr != nil {
		t.Error("a rejected move must leave the file in its original directory")
	}
	if _, statErr := os.Stat(newPath); statErr == nil {
		t.Error("a rejected move must not create the target file")
	}
}

// TestFS_MoveAudit_RejectsMalformedWithoutMoving is the audit sibling: parse
// failure must precede the rename.
func TestFS_MoveAudit_RejectsMalformedWithoutMoving(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "audits", "open")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Unterminated frontmatter: an opening fence with no closing one.
	if err := os.WriteFile(filepath.Join(dir, "a1.md"), []byte("---\narea: store\n# no closing fence\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := NewFS(root).MoveAudit("a1", domain.AuditClosed, false)
	if err == nil {
		t.Fatal("want an error for a malformed audit")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "a1.md")); statErr != nil {
		t.Error("a rejected audit move must leave the file in its original bucket")
	}
}

// TestFS_SetFields_ConflictsWhenMovedConcurrently pins the compare-and-swap: a
// task relocated between SetFields' read and its write must produce ErrConflict
// with nothing written — NOT a write to the stale path, which would resurrect
// the slug in its old status directory (permanent ErrAmbiguous).
func TestFS_SetFields_ConflictsWhenMovedConcurrently(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "alpha.md",
		"---\nstatus: ready-to-start\n---\n# Alpha\n")
	fs := NewFS(root)

	oldPath := filepath.Join(root, "tasks", "ready-to-start", "alpha.md")
	newDir := filepath.Join(root, "tasks", "in-progress")
	testHookBeforeSetFieldsWrite = func() {
		// Interleave a concurrent move while SetFields is between read and write.
		if err := os.MkdirAll(newDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(oldPath, filepath.Join(newDir, "alpha.md")); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { testHookBeforeSetFieldsWrite = nil }()

	_, err := fs.SetFields("alpha", map[string]any{"priority": "high"}, false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want ErrConflict for a concurrently-moved task, got %v", err)
	}
	if _, statErr := os.Stat(oldPath); statErr == nil {
		t.Error("the slug must NOT be resurrected in its old status directory")
	}
}

// TestFS_Move_ConflictsWhenMovedConcurrently is the Move sibling of the CAS guard
// above: a task relocated between Move's resolve and its write must produce
// ErrConflict with no duplicate left behind, rather than writing the new file and
// orphaning a copy in two status dirs.
func TestFS_Move_ConflictsWhenMovedConcurrently(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "alpha.md",
		"---\nstatus: ready-to-start\n---\n# Alpha\n")
	fs := NewFS(root)

	oldPath := filepath.Join(root, "tasks", "ready-to-start", "alpha.md")
	doneDir := filepath.Join(root, "tasks", "completed")
	testHookBeforeMoveWrite = func() {
		// A concurrent move to `completed` lands between this Move's resolve and write.
		if err := os.MkdirAll(doneDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(oldPath, filepath.Join(doneDir, "alpha.md")); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { testHookBeforeMoveWrite = nil }()

	_, err := fs.Move("alpha", domain.StatusInProgress, time.Now(), false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want ErrConflict for a concurrently-moved task, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "tasks", "in-progress", "alpha.md")); statErr == nil {
		t.Error("a conflicted move must not write the file into the target dir")
	}
}

// TestFS_SetFields_CRLFRoundTrip pins the line-ending promise: a CRLF file
// surgically edited must come back with consistent CRLF endings (no mixed
// LF-frontmatter/CRLF-body file) and correct values.
func TestFS_SetFields_CRLFRoundTrip(t *testing.T) {
	root := t.TempDir()
	crlf := strings.ReplaceAll("---\nstatus: ready-to-start\ndescription: old\n---\n# Alpha\nbody\n", "\n", "\r\n")
	writeTask(t, root, "ready-to-start", "alpha.md", crlf)

	task, err := NewFS(root).SetFields("alpha", map[string]any{"description": "new desc"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if task.Description != "new desc" {
		t.Errorf("description not updated: %+v", task)
	}
	b, err := os.ReadFile(filepath.Join(root, "tasks", "ready-to-start", "alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	if lone := strings.Count(string(b), "\n") - strings.Count(string(b), "\r\n"); lone != 0 {
		t.Errorf("CRLF file came back with %d bare-LF line endings (mixed endings):\n%q", lone, b)
	}
	// And it still reads back correctly.
	got, _, err := NewFS(root).GetTask("alpha")
	if err != nil || got.Description != "new desc" {
		t.Errorf("CRLF file should round-trip: %v %+v", err, got)
	}
}

// TestFS_UnterminatedFrontmatterIsAProblemNotAnEmptyTask pins that an opening
// fence with no closing one surfaces as a FileProblem (and SetFields refuses)
// instead of parsing as an empty task that a later edit double-fences.
func TestFS_UnterminatedFrontmatterIsAProblemNotAnEmptyTask(t *testing.T) {
	root := t.TempDir()
	const broken = "---\nstatus: ready-to-start\ndescription: x\n# no closing fence\n"
	writeTask(t, root, "ready-to-start", "alpha.md", broken)
	fs := NewFS(root)

	tasks, problems, err := fs.ListTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 || len(problems) != 1 {
		t.Fatalf("want 0 tasks + 1 problem, got %d tasks %d problems", len(tasks), len(problems))
	}
	if !strings.Contains(problems[0].Message, "unterminated") {
		t.Errorf("problem should name the unterminated fence, got %q", problems[0].Message)
	}

	_, err = fs.SetFields("alpha", map[string]any{"priority": "high"}, false)
	if err == nil {
		t.Fatal("SetFields must refuse an unterminated-frontmatter file")
	}
	b, _ := os.ReadFile(filepath.Join(root, "tasks", "ready-to-start", "alpha.md"))
	if string(b) != broken {
		t.Errorf("the broken file must be left untouched (not double-fenced):\n%s", b)
	}
}

// TestFixFrontmatterText_KeepsCRLF pins that the text-level fixer re-emits a
// CRLF file in CRLF (it previously rejoined with bare LF fences).
func TestFixFrontmatterText_KeepsCRLF(t *testing.T) {
	crlf := strings.ReplaceAll("---\nstatus: x\ntags: a, b\n---\nbody\n", "\n", "\r\n")
	fixed, changes := fixFrontmatterText([]byte(crlf))
	if len(changes) == 0 {
		t.Fatal("the tags line should have been normalized")
	}
	if lone := strings.Count(string(fixed), "\n") - strings.Count(string(fixed), "\r\n"); lone != 0 {
		t.Errorf("fixer produced %d bare-LF endings in a CRLF file:\n%q", lone, fixed)
	}
	if !strings.Contains(string(fixed), "tags: [a, b]") {
		t.Errorf("tags not normalized:\n%s", fixed)
	}
}

// TestFS_ListTasks_SkipsSymlinkedMarkdown pins that store scans accept only
// regular .md files: a symlink named like a task must not be followed (it could
// point out of the planning tree), so it's silently skipped, not parsed.
func TestFS_ListTasks_SkipsSymlinkedMarkdown(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "real.md",
		"---\nstatus: ready-to-start\ndescription: real\n---\n# Real\n")

	// A symlink in the status dir pointing at a file outside the tree.
	outside := filepath.Join(t.TempDir(), "secret.md")
	if err := os.WriteFile(outside, []byte("---\nstatus: ready-to-start\n---\n# Secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "tasks", "ready-to-start", "link.md")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unsupported here: %v", err)
	}

	tasks, problems, err := NewFS(root).ListTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Errorf("a skipped symlink must not surface as a problem, got %v", problems)
	}
	if len(tasks) != 1 || tasks[0].Slug != "real" {
		t.Errorf("only the regular file should be listed, got %d tasks: %+v", len(tasks), tasks)
	}
}
