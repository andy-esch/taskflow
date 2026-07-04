//go:build unix

package store

import (
	"fmt"
	"os"
	"syscall"
)

// writeLock takes the process-wide advisory write lock — flock(LOCK_EX) on the repo root
// directory — so the version-CAS's verify→write becomes an ATOMIC compare-and-swap: no
// other cooperating writer can land a rename between a verify and its own rename (the
// lost-update window that the check-then-write, non-atomic on a filesystem, otherwise
// leaves open — widened to milliseconds by the temp-file fsync). Advisory: a
// non-cooperating writer (a raw hand-edit) isn't blocked, but the content hash still
// catches that on the next verify. Repo-wide (not per-file) for simplicity — writes are
// brief and infrequent, so serializing them is imperceptible; per-file locking is a future
// refinement. flock auto-releases if the process dies (no stale lock files). Returns an
// unlock func the caller defers after the write.
func (s *FS) writeLock() (func(), error) {
	f, err := os.Open(s.root)
	if err != nil {
		return nil, fmt.Errorf("open repo root for write lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("acquire write lock: %w", err)
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}
