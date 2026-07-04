package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestHashContent_StrongValidator(t *testing.T) {
	a := []byte("---\nstatus: next-up\n---\n# x\n")
	// Identical bytes in a distinct slice must hash identically (determinism).
	if hashContent(a) != hashContent(append([]byte(nil), a...)) {
		t.Fatal("hashContent must be deterministic for identical bytes")
	}
	if hashContent(a) == hashContent([]byte("---\nstatus: next-up\n---\n# y\n")) {
		t.Error("hashContent must differ when a byte differs")
	}
	// A single trailing space flips it — a STRONG (byte-exact) validator, by design:
	// a lost-update guard must not treat a cosmetically-different file as unchanged.
	if hashContent(a) == hashContent(append(append([]byte{}, a...), ' ')) {
		t.Error("hashContent must be byte-exact (trailing whitespace must change it)")
	}
}

func TestVerifyUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "t.md")
	content := []byte("---\nstatus: next-up\n---\n# x\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	ver := hashContent(content)
	here := func(string) (string, error) { return path, nil }

	// Unchanged source + matching version → passes.
	if err := verifyUnchanged(here, "t", path, ver, "task", "update"); err != nil {
		t.Errorf("an unchanged file must pass: %v", err)
	}
	// Empty ifVersion → the content check is skipped (unconditional write).
	if err := verifyUnchanged(here, "t", path, "", "task", "update"); err != nil {
		t.Errorf("an empty ifVersion must skip the content check: %v", err)
	}

	// A concurrent in-place edit → ErrConflict (the NEW coverage version-CAS adds).
	if err := os.WriteFile(path, []byte("---\nstatus: in-progress\n---\n# x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := verifyUnchanged(here, "t", path, ver, "task", "update"); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("a concurrent content edit must be ErrConflict, got %v", err)
	}

	// A concurrent relocation (resolver now points elsewhere) → ErrConflict, before any
	// content read — the old path-CAS coverage, preserved.
	moved := func(string) (string, error) { return filepath.Join(dir, "sub", "t.md"), nil }
	if err := verifyUnchanged(moved, "t", path, "", "task", "update"); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("a concurrent relocation must be ErrConflict, got %v", err)
	}

	// A vanished file (resolver errors) → ErrConflict.
	gone := func(string) (string, error) { return "", domain.ErrNotFound }
	if err := verifyUnchanged(gone, "t", path, ver, "task", "update"); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("a vanished file must be ErrConflict, got %v", err)
	}
}

// The conflict message reproduces the existing per-site wording exactly, so step 3 can
// route the current re-resolve blocks through verifyUnchanged without changing the error
// text callers/tests match on.
func TestVerifyUnchanged_MessageShape(t *testing.T) {
	moved := func(string) (string, error) { return "/elsewhere/t.md", nil }
	err := verifyUnchanged(moved, "retr", "/here/t.md", "", "task", "move")
	if err == nil || !strings.Contains(err.Error(), `task "retr" changed on disk during move; retry:`) {
		t.Errorf("conflict message shape drifted: %v", err)
	}
}
