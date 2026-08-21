//go:build unix

package userconfig

import (
	"fmt"
	"os"
	"syscall"
)

// userConfigWriteLock serializes home-config read→validate→write transactions across
// cooperating CLI processes. Lock the stable config DIRECTORY rather than config.toml
// or spaces.toml: either file may not exist yet, and atomic writes replace their inodes.
// flock releases automatically if the process dies, so there is no stale lock file.
func userConfigWriteLock(dir string) (func(), error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("open user config directory for write lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("acquire user config write lock: %w", err)
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}
