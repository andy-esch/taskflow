package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestTaskStart_ChangesStatusInPlace(t *testing.T) {
	root := setupRepo(t)
	path := filepath.Join(root, "tasks", testutil.TaskID("alpha")+"-alpha.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("alpha fixture missing: %v", err)
	}
	out := runRoot(t, "-C", root, "task", "start", "alpha")
	if !strings.Contains(out, "alpha -> in-progress") {
		t.Errorf("unexpected output: %q", out)
	}
	// Flat layout: the file path never changes; status is an in-place frontmatter edit.
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("alpha no longer at original path: %v", err)
	}
	if !strings.Contains(string(b), "status: in-progress") {
		t.Errorf("alpha frontmatter status not updated:\n%s", b)
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
