//go:build unix

package store

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// Isolate process-wide umask; 0600 fixtures cannot detect a caller accidentally
// choosing ordinary exclusive creation instead of exact-mode relocation.
func TestRenameTask_PreservesModeDespiteUmask(t *testing.T) {
	const helper = "TSKFLWCTL_TEST_RENAME_UMASK"
	if os.Getenv(helper) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestRenameTask_PreservesModeDespiteUmask$")
		cmd.Env = append(os.Environ(), helper+"=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("umask subprocess: %v\n%s", err, output)
		}
		return
	}
	previous := syscall.Umask(0o022)
	defer syscall.Umask(previous)
	root, source, _ := renameRepo(t)
	if err := os.Chmod(source, 0o664); err != nil {
		t.Fatal(err)
	}
	result, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).RenameTask("old", "New title", false)
	if err != nil || !result.Complete || !result.SourceRemoved {
		t.Fatalf("rename result=%+v err=%v", result, err)
	}
	destination := filepath.Join(root, domain.TasksDir, "6fjangd7kva1-new-title.md")
	if info, err := os.Stat(destination); err != nil || info.Mode().Perm() != 0o664 {
		t.Fatalf("umask changed renamed mode: info=%v err=%v", info, err)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("successful rename retained source: %v", err)
	}
}
