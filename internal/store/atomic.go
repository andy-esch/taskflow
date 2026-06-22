package store

import (
	"fmt"
	"os"
	"path/filepath"
)

// stageTemp writes data to a fsync'd, chmod'd temp file in dir and returns its
// path for the caller to move into place (rename or link). On any error the
// temp is cleaned up.
func stageTemp(dir string, data []byte, perm os.FileMode) (string, error) {
	tmp, err := os.CreateTemp(dir, ".tskflwctl-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temp: %w", err)
	}
	name := tmp.Name()
	cleanup := func(wrap string, e error) (string, error) {
		_ = os.Remove(name)
		return "", fmt.Errorf(wrap+": %w", e)
	}
	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return cleanup("write temp", err)
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return cleanup("sync temp", err)
	}
	if err = tmp.Close(); err != nil {
		return cleanup("close temp", err)
	}
	if err = os.Chmod(name, perm); err != nil {
		return cleanup("chmod temp", err)
	}
	return name, nil
}

// writeFileAtomic writes data via a temp file in the same directory, fsync, and
// rename — so a crash or Ctrl-C mid-write can't leave a truncated file. Rename
// overwrites an existing file in place.
//
// When overwriting, the destination's existing mode is preserved rather than
// reset to perm: a user (or synced/encrypted setup) that chmod'd a task to 0600
// must not have it silently widened to 0644 on the next edit. perm is the
// fallback for a file that doesn't yet exist.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	}
	tmp, err := stageTemp(filepath.Dir(path), data, perm)
	if err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename temp into place: %w", err)
	}
	syncDir(filepath.Dir(path))
	return nil
}

// syncDir best-effort fsyncs a directory so a just-completed rename/create in
// it survives a power loss. Errors are ignored: some filesystems (network and
// FUSE mounts in particular) reject directory fsync, and degraded durability
// there beats failing every write.
func syncDir(dir string) {
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}
}

// createFileAtomic writes a *new* file with O_EXCL: it fails with an os.IsExist
// error if path is already taken (no stat/write race), using the portable
// O_CREATE|O_EXCL flag rather than a hard link — hard links are restricted on
// many container/network/VM mounts. A crash mid-write can leave a partial *new*
// file (cleaned up on a write error here); it never corrupts an existing one.
func createFileAtomic(path string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err // os.IsExist(err) on collision
	}
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return fmt.Errorf("sync %s: %w", path, err)
	}
	if err = f.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	syncDir(filepath.Dir(path))
	return nil
}
