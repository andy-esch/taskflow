package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestTaskStart_MovesFile(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "start", "alpha")
	if !strings.Contains(out, "alpha -> in-progress") {
		t.Errorf("unexpected output: %q", out)
	}
	if _, err := os.Stat(filepath.Join(root, "tasks", "in-progress", "alpha.md")); err != nil {
		t.Errorf("alpha not in in-progress: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "tasks", "ready-to-start", "alpha.md")); !os.IsNotExist(err) {
		t.Error("alpha still in ready-to-start")
	}
}

func TestTaskShow(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "show", "alpha")
	if !strings.Contains(out, "slug:") || !strings.Contains(out, "# Alpha") {
		t.Errorf("unexpected show output:\n%s", out)
	}
}

func TestTaskStart_NotFound_ExitCode(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "start", "ghost"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing task")
	}
	if ExitCode(err) != 10 {
		t.Errorf("want exit code 10 (not-found), got %d", ExitCode(err))
	}
}

func TestExitCode(t *testing.T) {
	cases := map[error]int{
		nil:                                     0,
		fmt.Errorf("x: %w", domain.ErrNotFound): 10,
		fmt.Errorf("x: %w", domain.ErrValidation): 11,
		fmt.Errorf("x: %w", domain.ErrAmbiguous):  13,
		fmt.Errorf("plain"):                       1,
	}
	for err, want := range cases {
		if got := ExitCode(err); got != want {
			t.Errorf("ExitCode(%v) = %d, want %d", err, got, want)
		}
	}
}
