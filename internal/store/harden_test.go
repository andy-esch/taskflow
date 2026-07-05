package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// TestFS_Move_RejectsUnreloadableWithoutMoving pins parse-before-commit on the
// move path: a task whose frontmatter wouldn't read back after the update must
// fail with nothing on disk changed. Under the flat layout a move is an in-place
// frontmatter edit — the file path never changes — so the guarantee is that the
// file is left byte-for-byte untouched (its status: not flipped).
func TestFS_Move_RejectsUnreloadableWithoutMoving(t *testing.T) {
	root := t.TempDir()
	// tier as a quoted string survives the surgical status update but fails the
	// strict typed decode.
	const original = "---\nstatus: ready-to-start\ntier: \"4\"\n---\n# Alpha\n"
	path, out := testutil.TaskFixture(root, "ready-to-start", "alpha.md", original)
	testutil.Write(t, path, out)

	_, err := NewFS(root).Move("alpha", domain.StatusInProgress, time.Now(), false)
	if err == nil {
		t.Fatal("want an error for a move that wouldn't reload")
	}
	b, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal("a rejected move must leave the file in place")
	}
	if string(b) != out {
		t.Errorf("a rejected move must leave the file untouched (status not flipped):\n%s", b)
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

// TestFS_SetFields_ConflictsWhenEditedConcurrently pins the version-CAS: a task
// whose bytes are edited in place between SetFields' read and its write must
// produce ErrConflict with nothing written — the flat layout has no relocation,
// so content drift is the lost-update hazard the compare-and-swap must catch.
func TestFS_SetFields_ConflictsWhenEditedConcurrently(t *testing.T) {
	root := t.TempDir()
	path, out := testutil.TaskFixture(root, "ready-to-start", "alpha.md",
		"---\nstatus: ready-to-start\ntags: [seed]\n---\n# Alpha\n")
	testutil.Write(t, path, out)
	fs := NewFS(root)

	const concurrent = "---\nstatus: ready-to-start\ntags: [seed]\ndescription: raced\n---\n# Alpha\n"
	testHookBeforeSetFieldsWrite = func() {
		// Interleave a concurrent in-place edit while SetFields is between read and write.
		if err := os.WriteFile(path, []byte(concurrent), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { testHookBeforeSetFieldsWrite = nil }()

	_, err := fs.SetFields("alpha", map[string]any{"priority": "high"}, false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want ErrConflict for a concurrently-edited task, got %v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != concurrent {
		t.Error("the concurrent edit must NOT be clobbered by the stale write")
	}
}

// TestFS_Move_ConflictsWhenEditedConcurrently is the Move sibling of the CAS
// guard above: a task whose bytes change in place between Move's resolve and its
// write must produce ErrConflict with the concurrent edit intact, rather than a
// stale write that silently clobbers the racing writer's change.
func TestFS_Move_ConflictsWhenEditedConcurrently(t *testing.T) {
	root := t.TempDir()
	path, out := testutil.TaskFixture(root, "ready-to-start", "alpha.md",
		"---\nstatus: ready-to-start\n---\n# Alpha\n")
	testutil.Write(t, path, out)
	fs := NewFS(root)

	const concurrent = "---\nstatus: ready-to-start\ndescription: raced\n---\n# Alpha\n"
	testHookBeforeMoveWrite = func() {
		// A concurrent in-place edit lands between this Move's resolve and write.
		if err := os.WriteFile(path, []byte(concurrent), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { testHookBeforeMoveWrite = nil }()

	_, err := fs.Move("alpha", domain.StatusInProgress, time.Now(), false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want ErrConflict for a concurrently-edited task, got %v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != concurrent {
		t.Error("a conflicted move must not clobber the concurrent edit")
	}
}

// TestFS_SetFields_CRLFRoundTrip pins the line-ending promise: a CRLF file
// surgically edited must come back with consistent CRLF endings (no mixed
// LF-frontmatter/CRLF-body file) and correct values.
func TestFS_SetFields_CRLFRoundTrip(t *testing.T) {
	root := t.TempDir()
	crlf := strings.ReplaceAll("---\nstatus: ready-to-start\ndescription: old\ntags: [seed]\n---\n# Alpha\nbody\n", "\n", "\r\n")
	path, out := testutil.TaskFixture(root, "ready-to-start", "alpha.md", crlf)
	testutil.Write(t, path, out)

	task, err := NewFS(root).SetFields("alpha", map[string]any{"description": "new desc"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if task.Description != "new desc" {
		t.Errorf("description not updated: %+v", task)
	}
	b, err := os.ReadFile(path)
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
	path, out := testutil.TaskFixture(root, "ready-to-start", "alpha.md", broken)
	testutil.Write(t, path, out)
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
	b, _ := os.ReadFile(path)
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

	// A symlink in the tasks dir, named like a flat id-led task, pointing at a
	// file outside the tree.
	outside := filepath.Join(t.TempDir(), "secret.md")
	if err := os.WriteFile(outside, []byte("---\nstatus: ready-to-start\n---\n# Secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "tasks", testutil.TaskID("link")+"-link.md")
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
