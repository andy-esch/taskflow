//go:build unix

package store

import (
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// Umask is process-wide. Re-execute just this test so even parallel tests cannot
// observe its change. 0664 under 022 discriminates exact-mode relocation from an
// ordinary exclusive create (0644); 0600 alone cannot prove that contract.
func TestFixIDRelocationPreservesModeDespiteUmask(t *testing.T) {
	const helper = "TSKFLWCTL_TEST_FIX_RELOCATION_UMASK"
	if os.Getenv(helper) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestFixIDRelocationPreservesModeDespiteUmask$")
		cmd.Env = append(os.Environ(), helper+"=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("umask subprocess: %v\n%s", err, output)
		}
		return
	}
	previous := syscall.Umask(0o022)
	defer syscall.Umask(previous)
	root, source, destination, _ := invalidIDRepairFixture(t)
	if err := os.Chmod(source, 0o664); err != nil {
		t.Fatal(err)
	}
	results, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).FixFrontmatter(false)
	if err != nil || len(results) != 1 || results[0].Skipped {
		t.Fatalf("relocation results=%+v err=%v", results, err)
	}
	if info, err := os.Stat(destination); err != nil || info.Mode().Perm() != 0o664 {
		t.Fatalf("umask changed relocated mode: info=%v err=%v", info, err)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("successful relocation retained source: %v", err)
	}
}
