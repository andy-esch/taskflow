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

// writeTaskAt writes a flat task fixture (like writeTask) and returns its id-led
// path, so a lifecycle test can assert the file stays put and only its
// frontmatter changes (no relocation under the flat layout).
func writeTaskAt(t *testing.T, root, status, name, content string) string {
	t.Helper()
	path, out := testutil.TaskFixture(root, status, name, content)
	testutil.Write(t, path, out)
	return path
}

func TestFS_Move(t *testing.T) {
	root := t.TempDir()
	path := writeTaskAt(t, root, "ready-to-start", "alpha.md", "---\nstatus: ready-to-start\nepic: 01-x\n---\n# Alpha\n")

	now := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)
	task, err := NewFS(root).Move("alpha", domain.StatusInProgress, now, false)
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != domain.StatusInProgress {
		t.Errorf("status = %s", task.Status)
	}

	// The move is an in-place frontmatter edit under the flat layout: the file
	// stays at its original id-led path, only status: changes.
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file missing at original path: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "status: in-progress") || !strings.Contains(s, "started_at") || !strings.Contains(s, "epic: 01-x") {
		t.Errorf("moved content wrong:\n%s", s)
	}
}

func TestFS_Move_Idempotent(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "in-progress", "beta.md", "---\nstatus: in-progress\n---\n# B\n")
	task, err := NewFS(root).Move("beta", domain.StatusInProgress, time.Now(), false)
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != domain.StatusInProgress {
		t.Errorf("status = %s", task.Status)
	}
}

// TestFS_Move_RevisitAt pins the revisit_at lifecycle at the store layer: a
// re-defer (deferred->deferred, the idempotent no-op) KEEPS the snooze date, and
// any move OUT of deferred CLEARS it while preserving the historical deferred_at.
// The CLI tests only cover the resume (next) path; this guards the load-bearing
// from==to early-return (a reorder that cleared the date on re-defer would be
// silent data loss otherwise) and the to-independent clear directly.
func TestFS_Move_RevisitAt(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC)
	fs := NewFS(root)
	deferred := func(name string) string {
		return writeTaskAt(t, root, "deferred", name,
			"---\nstatus: deferred\nrevisit_at: \"2026-09-01\"\ndeferred_at: \"2026-06-01\"\n---\n# X\n")
	}
	read := func(path string) string {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return string(b)
	}

	// Re-defer (deferred -> deferred): idempotent no-op, snooze date untouched.
	redeferPath := deferred("redefer.md")
	task, err := fs.Move("redefer", domain.StatusDeferred, now, false)
	if err != nil {
		t.Fatal(err)
	}
	if task.RevisitAt != "2026-09-01" {
		t.Errorf("re-defer should keep revisit_at, got %q", task.RevisitAt)
	}
	if !strings.Contains(read(redeferPath), "revisit_at") {
		t.Error("re-defer must not strip revisit_at from disk")
	}

	// Leaving deferred clears the snooze date for ANY destination (resume via
	// next/ready, or archive via deprecate); deferred_at history survives. The
	// file stays at its original flat path — only frontmatter changes.
	for _, tc := range []struct {
		to  domain.Status
		dir string
	}{
		{domain.StatusReadyToStart, "ready-to-start"},
		{domain.StatusDeprecated, "deprecated"},
	} {
		name := "leave-" + tc.dir + ".md"
		path := deferred(name)
		if _, err := fs.Move(strings.TrimSuffix(name, ".md"), tc.to, now, false); err != nil {
			t.Fatalf("move to %s: %v", tc.to, err)
		}
		got := read(path)
		if strings.Contains(got, "revisit_at") {
			t.Errorf("leaving deferred -> %s should clear revisit_at:\n%s", tc.to, got)
		}
		if !strings.Contains(got, "deferred_at") {
			t.Errorf("leaving deferred -> %s must keep historical deferred_at:\n%s", tc.to, got)
		}
	}
}

// TestFS_Defer pins the audit-M4 atomic defer: a SINGLE Defer call sets status
// deferred AND stamps revisit_at + deferred_at in one write (no
// Move-then-SetFields window), and a re-defer rewrites revisit_at in place.
func TestFS_Defer(t *testing.T) {
	root := t.TempDir()
	path := writeTaskAt(t, root, "ready-to-start", "alpha.md", "---\nstatus: ready-to-start\nepic: 01-x\n---\n# Alpha\n")
	now := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)
	fs := NewFS(root)
	read := func() string {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return string(b)
	}

	task, err := fs.Defer("alpha", "2026-09-01", now, false)
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != domain.StatusDeferred || task.RevisitAt != "2026-09-01" {
		t.Errorf("deferred task = (status %q, revisit %q), want (deferred, 2026-09-01)", task.Status, task.RevisitAt)
	}
	// The file stays at its original flat path and carries BOTH the snooze date
	// and the deferred_at stamp — written together, so the two can't land separately.
	got := read()
	if !strings.Contains(got, "status: deferred") || !strings.Contains(got, "2026-09-01") || !strings.Contains(got, "deferred_at") {
		t.Errorf("deferred file should carry status + revisit_at + deferred_at in one write:\n%s", got)
	}

	// Re-defer with a NEW date: in-place rewrite, revisit_at updated.
	later := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	redeferred, err := fs.Defer("alpha", "2026-12-25", later, false)
	if err != nil {
		t.Fatalf("re-defer: %v", err)
	}
	if redeferred.RevisitAt != "2026-12-25" {
		t.Errorf("re-defer should update revisit_at, got %q", redeferred.RevisitAt)
	}
	if got := read(); !strings.Contains(got, "2026-12-25") || strings.Contains(got, "2026-09-01") {
		t.Errorf("re-defer should replace the snooze date on disk:\n%s", got)
	}
}

// TestFS_Defer_BareNoDate pins that a defer with no date is a plain status change
// to deferred — deferred_at stamped, and no revisit_at written.
func TestFS_Defer_BareNoDate(t *testing.T) {
	root := t.TempDir()
	path := writeTaskAt(t, root, "ready-to-start", "beta.md", "---\nstatus: ready-to-start\n---\n# Beta\n")
	now := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)

	task, err := NewFS(root).Defer("beta", "", now, false)
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != domain.StatusDeferred || task.RevisitAt != "" {
		t.Errorf("bare defer = (status %q, revisit %q), want (deferred, \"\")", task.Status, task.RevisitAt)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	got := string(b)
	if strings.Contains(got, "revisit_at") {
		t.Errorf("bare defer must not write revisit_at:\n%s", got)
	}
	if !strings.Contains(got, "deferred_at") {
		t.Errorf("bare defer should stamp deferred_at:\n%s", got)
	}
}

func TestFS_Move_NotFound(t *testing.T) {
	_, err := NewFS(t.TempDir()).Move("nope", domain.StatusCompleted, time.Now(), false)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestFS_Resolve_Ambiguous(t *testing.T) {
	root := t.TempDir()
	// Two distinct id-led files share the slug "dup" (allowed under the flat
	// layout — files are unique by id). Resolving by the shared slug is ambiguous.
	idA := testutil.TaskID("dup-a")
	idB := testutil.TaskID("dup-b")
	testutil.Write(t, filepath.Join(root, "tasks", idA+"-dup.md"), "---\nstatus: ready-to-start\n---\n")
	testutil.Write(t, filepath.Join(root, "tasks", idB+"-dup.md"), "---\nstatus: in-progress\n---\n")
	_, err := NewFS(root).Move("dup", domain.StatusCompleted, time.Now(), false)
	if !errors.Is(err, domain.ErrAmbiguous) {
		t.Errorf("want ErrAmbiguous, got %v", err)
	}
	// The message should name each candidate's id, so the user can retype an
	// unambiguous name.
	for _, want := range []string{idA, idB} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("ambiguous error should name %q: %v", want, err)
		}
	}
}
