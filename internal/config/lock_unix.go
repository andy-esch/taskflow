//go:build unix

package config

import (
	"fmt"
	"os"
	"syscall"
)

// writeLock serializes the repo-config read→edit→rename transaction across
// cooperating taskflow processes. The stable config directory is locked because
// atomic writes replace the config file's inode. flock releases on process exit.
func writeLock(dir string) (func(), error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("open config directory for write lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("acquire config write lock: %w", err)
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}
