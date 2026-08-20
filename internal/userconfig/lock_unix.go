//go:build unix

package userconfig

import (
	"fmt"
	"os"
	"syscall"
)

// registryWriteLock serializes the registry's read→validate→write transaction across
// cooperating CLI processes. Lock the stable config DIRECTORY rather than spaces.toml:
// the file may not exist yet, and every atomic write replaces its inode. flock releases
// automatically if the process dies, so there is no stale lock file to recover.
func registryWriteLock(dir string) (func(), error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("open config directory for registry lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("acquire registry lock: %w", err)
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}
