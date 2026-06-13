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
