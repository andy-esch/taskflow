package userconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// spaceHome points the registry at a fresh temp dir.
func spaceHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv(DirEnv, dir)
	return dir
}

// TestSpaces_MissingFileIsNormal: most machines have no registry, and that must read as
// "no spaces" rather than an error — the registry is advisory.
func TestSpaces_MissingFileIsNormal(t *testing.T) {
	spaceHome(t)
	got, err := Spaces()
	if err != nil || len(got) != 0 {
		t.Errorf("a missing registry should be silent and empty, got %v / %v", got, err)
	}
}

// TestAddSpace_DedupsOnPhysicalPath pins identity: the physical directory, never the
// label and never verify_id. Relative, absolute and symlinked spellings are one repo.
func TestAddSpace_DedupsOnPhysicalPath(t *testing.T) {
	spaceHome(t)
	repo := t.TempDir()
	if added, _, err := AddSpace(Space{ID: "a", Path: repo}); err != nil || !added {
		t.Fatalf("first add: %v / %v", added, err)
	}
	// a different spelling of the same directory
	spelling := filepath.Join(repo, "sub", "..")
	if err := os.MkdirAll(filepath.Join(repo, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	added, existing, err := AddSpace(Space{ID: "b", Path: spelling})
	if err != nil {
		t.Fatalf("re-add: %v", err)
	}
	if added {
		t.Error("a different spelling of a registered path must not add a second entry")
	}
	if existing.ID != "a" {
		t.Errorf("should report the existing entry, got %q", existing.ID)
	}
}

// TestAddSpace_SharedVerifyIDIsAllowed is the counterpart, and easy to get wrong: two
// checkouts of ONE repo legitimately share a durable id (it lives in a committed file), so
// dedup must never collapse them — they are separately addressable working trees.
func TestAddSpace_SharedVerifyIDIsAllowed(t *testing.T) {
	spaceHome(t)
	a, b := t.TempDir(), t.TempDir()
	if _, _, err := AddSpace(Space{ID: "main", Path: a, VerifyID: "shared0000aa"}); err != nil {
		t.Fatal(err)
	}
	added, _, err := AddSpace(Space{ID: "wt", Path: b, VerifyID: "shared0000aa"})
	if err != nil || !added {
		t.Fatalf("two checkouts of one repo must both register: added=%v err=%v", added, err)
	}
	if got, _ := Spaces(); len(got) != 2 {
		t.Errorf("want 2 entries, got %d", len(got))
	}
}

// TestAddSpace_LabelCollisionRefused: a clash is visible, never silently suffixed into
// something like `taskflow-2` that nobody chose.
func TestAddSpace_LabelCollisionRefused(t *testing.T) {
	spaceHome(t)
	if _, _, err := AddSpace(Space{ID: "dup", Path: t.TempDir()}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := AddSpace(Space{ID: "dup", Path: t.TempDir()}); err == nil {
		t.Error("a second space claiming the same label must be refused")
	}
}

// TestForgetSpace_OnlyDropsTheEntry: forgetting is a registry edit, not a deletion.
func TestForgetSpace_OnlyDropsTheEntry(t *testing.T) {
	spaceHome(t)
	repo := t.TempDir()
	marker := filepath.Join(repo, "keep.txt")
	if err := os.WriteFile(marker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := AddSpace(Space{ID: "gone", Path: repo}); err != nil {
		t.Fatal(err)
	}
	removed, err := ForgetSpace("gone")
	if err != nil || !removed {
		t.Fatalf("forget: %v / %v", removed, err)
	}
	if got, _ := Spaces(); len(got) != 0 {
		t.Errorf("entry should be gone, got %v", got)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Error("forgetting must NOT touch the repo on disk")
	}
	if again, _ := ForgetSpace("gone"); again {
		t.Error("forgetting an absent id should report no change, not fail")
	}
}

// TestWriteSpaces_SortedAndRoundTrips: the file is tool-owned, so entries are sorted by
// label — there is no human ordering intent to preserve, and a stable order keeps diffs
// from churning when an unrelated space is added.
func TestWriteSpaces_SortedAndRoundTrips(t *testing.T) {
	dir := spaceHome(t)
	for _, id := range []string{"zeta", "alpha", "mid"} {
		if _, _, err := AddSpace(Space{ID: id, Path: t.TempDir()}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := Spaces()
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, s := range got {
		ids = append(ids, s.ID)
	}
	if strings.Join(ids, ",") != "alpha,mid,zeta" {
		t.Errorf("entries should be sorted by label, got %v", ids)
	}
	b, err := os.ReadFile(filepath.Join(dir, SpacesFile))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "schema_version") {
		t.Error("the registry should carry its on-disk schema version")
	}
	if !strings.Contains(string(b), "MANAGED BY THE TOOL") {
		t.Error("the file should say it is tool-owned, since hand edits are not preserved")
	}
}

// TestTildePath_RoundTrip: entries store `~` so the file stays portable and committable to
// a dotfiles repo, while comparison still happens on physical paths.
func TestTildePath_RoundTrip(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	p := filepath.Join(home, "git", "thing")
	abbrev := TildePath(p)
	if !strings.HasPrefix(abbrev, "~/") {
		t.Errorf("a path under home should abbreviate, got %q", abbrev)
	}
	if ExpandTilde(abbrev) != p {
		t.Errorf("round trip failed: %q -> %q", abbrev, ExpandTilde(abbrev))
	}
	outside := filepath.Join(t.TempDir(), "x")
	if TildePath(outside) != outside {
		t.Errorf("a path outside home must be left alone, got %q", TildePath(outside))
	}
	if ExpandTilde("~someone/x") != "~someone/x" {
		t.Error("~user forms are deliberately unsupported and must pass through unchanged")
	}
}
