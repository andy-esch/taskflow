package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSweepStaleTemps pins L18's orphan sweep: only the tool's own .tmp prefix,
// only past the age threshold — a fresh temp (could be a live write) and any user
// file are left alone.
func TestSweepStaleTemps(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, ".tskflwctl-old.tmp")
	fresh := filepath.Join(dir, ".tskflwctl-new.tmp")
	user := filepath.Join(dir, "keep.md")
	for _, p := range []string{stale, fresh, user} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now()
	old := now.Add(-2 * staleTempAge)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatal(err)
	}

	removed := sweepStaleTemps(dir, now)
	if len(removed) != 1 || filepath.Base(removed[0]) != ".tskflwctl-old.tmp" {
		t.Fatalf("sweep should remove only the stale temp, got %v", removed)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("the stale temp should be gone")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("a fresh temp must be kept — it could be a live write")
	}
	if _, err := os.Stat(user); err != nil {
		t.Error("a user file must never be swept")
	}
}

// TestWriteFileAtomic_FailureLeavesNoOrphan pins L18's crash/error-rollback claim
// on the failure path: a write into a read-only dir fails and leaves neither an
// orphan temp nor a partial target. (chmod is a no-op as root, so skip there.)
func TestWriteFileAtomic_FailureLeavesNoOrphan(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses 0555 dir perms")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) }) // so t.TempDir cleanup can rm it

	if err := writeFileAtomic(filepath.Join(dir, "x.md"), []byte("data"), 0o644); err == nil {
		t.Fatal("writeFileAtomic into a read-only dir should fail")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		t.Errorf("a failed write must leave nothing behind, found %q", e.Name())
	}
}

// TestCreateFileAtomic_FailureLeavesNoOrphan pins the same for the exclusive-create
// helper: a failed create leaves no partial file.
func TestCreateFileAtomic_FailureLeavesNoOrphan(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses 0555 dir perms")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	if err := createFileAtomic(filepath.Join(dir, "y.md"), []byte("data"), 0o644); err == nil {
		t.Fatal("createFileAtomic into a read-only dir should fail")
	}
	if entries, err := os.ReadDir(dir); err != nil || len(entries) != 0 {
		t.Errorf("a failed create must leave no file, got %d entries (%v)", len(entries), err)
	}
}
