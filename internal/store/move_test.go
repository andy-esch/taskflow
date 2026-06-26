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

func TestFS_Move(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "alpha.md", "---\nstatus: ready-to-start\nepic: 01-x\n---\n# Alpha\n")

	now := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)
	task, err := NewFS(root).Move("alpha", domain.StatusInProgress, now, false)
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != domain.StatusInProgress {
		t.Errorf("status = %s", task.Status)
	}

	if _, err := os.Stat(filepath.Join(root, "tasks", "ready-to-start", "alpha.md")); !os.IsNotExist(err) {
		t.Error("old file should be gone")
	}
	newPath := filepath.Join(root, "tasks", "in-progress", "alpha.md")
	b, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("new file missing: %v", err)
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
	deferred := func(name string) {
		writeTask(t, root, "deferred", name,
			"---\nstatus: deferred\nrevisit_at: \"2026-09-01\"\ndeferred_at: \"2026-06-01\"\n---\n# X\n")
	}
	read := func(status, name string) string {
		b, err := os.ReadFile(filepath.Join(root, "tasks", status, name))
		if err != nil {
			t.Fatalf("read %s/%s: %v", status, name, err)
		}
		return string(b)
	}

	// Re-defer (deferred -> deferred): idempotent no-op, snooze date untouched.
	deferred("redefer.md")
	task, err := fs.Move("redefer", domain.StatusDeferred, now, false)
	if err != nil {
		t.Fatal(err)
	}
	if task.RevisitAt != "2026-09-01" {
		t.Errorf("re-defer should keep revisit_at, got %q", task.RevisitAt)
	}
	if !strings.Contains(read("deferred", "redefer.md"), "revisit_at") {
		t.Error("re-defer must not strip revisit_at from disk")
	}

	// Leaving deferred clears the snooze date for ANY destination (resume via
	// next/ready, or archive via deprecate); deferred_at history survives.
	for _, tc := range []struct {
		to  domain.Status
		dir string
	}{
		{domain.StatusReadyToStart, "ready-to-start"},
		{domain.StatusDeprecated, "deprecated"},
	} {
		name := "leave-" + tc.dir + ".md"
		deferred(name)
		if _, err := fs.Move(strings.TrimSuffix(name, ".md"), tc.to, now, false); err != nil {
			t.Fatalf("move to %s: %v", tc.to, err)
		}
		got := read(tc.dir, name)
		if strings.Contains(got, "revisit_at") {
			t.Errorf("leaving deferred -> %s should clear revisit_at:\n%s", tc.to, got)
		}
		if !strings.Contains(got, "deferred_at") {
			t.Errorf("leaving deferred -> %s must keep historical deferred_at:\n%s", tc.to, got)
		}
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
	writeTask(t, root, "ready-to-start", "dup.md", "---\nstatus: ready-to-start\n---\n")
	writeTask(t, root, "in-progress", "dup.md", "---\nstatus: in-progress\n---\n")
	_, err := NewFS(root).Move("dup", domain.StatusCompleted, time.Now(), false)
	if !errors.Is(err, domain.ErrAmbiguous) {
		t.Errorf("want ErrAmbiguous, got %v", err)
	}
	// The message should name where the duplicates live, so the user can clean up.
	for _, want := range []string{"ready-to-start", "in-progress"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("ambiguous error should name %q: %v", want, err)
		}
	}
}
