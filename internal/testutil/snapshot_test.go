package testutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotTreeCapturesBytesEmptyDirectoriesAndSymlinkTargets(t *testing.T) {
	root := t.TempDir()
	Write(t, filepath.Join(root, "file"), "original")
	if err := os.Mkdir(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink("file", link); err != nil {
		t.Skipf("symlinks unsupported here: %v", err)
	}
	before := SnapshotTree(t, root)
	if len(before) != 4 || before["file"].Content != "original" || !before["empty"].Mode.IsDir() || before["link"].Content != "file" || before["link"].Mode&fs.ModeSymlink == 0 {
		t.Fatalf("incomplete fixture snapshot: %v", before)
	}
	Write(t, filepath.Join(root, "file"), "changed")
	if err := os.Chmod(filepath.Join(root, "file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("empty", link); err != nil {
		t.Fatal(err)
	}
	after := SnapshotTree(t, root)
	if after["file"].Content != "changed" || after["file"].Mode.Perm() != 0o600 || after["link"].Content != "empty" {
		t.Fatalf("snapshot missed changed bytes, permissions, or target: %v", after)
	}
}
