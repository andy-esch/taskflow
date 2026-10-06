package testutil

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TreeEntry captures fixture bytes (or symlink target) and mode, including empty directories.
type TreeEntry struct {
	Mode    fs.FileMode
	Content string
}

// SnapshotTree compares entries, bytes, link targets, and modes in a throwaway fixture tree.
// Unsupported special files fail the test rather than silently weakening the oracle.
// It is not a transaction/revision token and makes no concurrent snapshot guarantee.
func SnapshotTree(t testing.TB, root string) map[string]TreeEntry {
	t.Helper()
	entries := make(map[string]TreeEntry)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		snapshot := TreeEntry{Mode: info.Mode()}
		switch {
		case info.Mode().IsRegular():
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			snapshot.Content = string(content)
		case info.Mode()&fs.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			snapshot.Content = target
		case info.IsDir():
		default:
			return fmt.Errorf("unsupported fixture entry %s (%s)", path, info.Mode())
		}
		entries[rel] = snapshot
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot fixture: %v", err)
	}
	return entries
}
