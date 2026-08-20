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
	// a relative spelling of the same directory
	if err := os.MkdirAll(filepath.Join(repo, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	spelling, err := filepath.Rel(cwd, filepath.Join(repo, "sub", ".."))
	if err != nil {
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
	link := filepath.Join(t.TempDir(), "repo-link")
	if err := os.Symlink(repo, link); err != nil {
		t.Fatal(err)
	}
	added, existing, err = AddSpace(Space{ID: "via-link", Path: link})
	if err != nil || added || existing.ID != "a" {
		t.Fatalf("a symlink spelling must dedup to a: added=%v existing=%q err=%v", added, existing.ID, err)
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

// TestSpaceWrites_PreserveInsertionOrderAndRoundTrip pins the ordering decision:
// registrations stay in insertion order. Re-sorting a hand-edited array of tables would
// move comments with surprising entries and violate the surgical-write contract.
func TestSpaceWrites_PreserveInsertionOrderAndRoundTrip(t *testing.T) {
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
	if strings.Join(ids, ",") != "zeta,alpha,mid" {
		t.Errorf("entries should retain insertion order, got %v", ids)
	}
	b, err := os.ReadFile(filepath.Join(dir, SpacesFile))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "schema_version") {
		t.Error("the registry should carry its on-disk schema version")
	}
	if !strings.Contains(string(b), "comments and unknown keys survive") {
		t.Error("the generated header should document the surgical editing contract")
	}
}

// TestSpaceWrites_AreSurgical is the central persistence contract. Adding appends to the
// original bytes; forgetting removes only the selected table. Top-level metadata, comments,
// key order, unknown entry keys, and unrelated tables all survive.
func TestSpaceWrites_AreSurgical(t *testing.T) {
	dir := spaceHome(t)
	path := filepath.Join(dir, SpacesFile)
	original := `# hand-written registry header
schema_version = 1 # keep this inline comment
owner = "andy"      # unknown top-level key

[[space]] # first
path = "/tmp/alpha" # deliberately before id
id = "alpha"
mystery = "keep me"

[[space]]
# beta's comment
id = "beta"
path = "/tmp/beta"
custom_beta = 42 # unknown entry key

[notes]
text = "unrelated table survives"
`
	if err := os.WriteFile(path, []byte(original), 0o640); err != nil {
		t.Fatal(err)
	}
	if added, _, err := AddSpace(Space{ID: "gamma", Path: t.TempDir(), Added: "2026-08-20"}); err != nil || !added {
		t.Fatalf("add gamma: added=%v err=%v", added, err)
	}
	afterAdd, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(afterAdd), original) {
		t.Errorf("add must preserve every existing byte and append one table:\n%s", afterAdd)
	}
	if fi, err := os.Stat(path); err != nil || fi.Mode().Perm() != 0o640 {
		t.Errorf("atomic edit should preserve file mode 0640, got %v / %v", fi, err)
	}

	if removed, err := ForgetSpace("alpha"); err != nil || !removed {
		t.Fatalf("forget alpha: removed=%v err=%v", removed, err)
	}
	afterForget, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(afterForget)
	for _, want := range []string{
		"# hand-written registry header",
		"schema_version = 1 # keep this inline comment",
		`owner = "andy"      # unknown top-level key`,
		"# beta's comment",
		`id = "beta"`,
		"custom_beta = 42 # unknown entry key",
		`text = "unrelated table survives"`,
		`id = "gamma"`,
	} {
		if !strings.Contains(text, want) {
			t.Errorf("forget dropped preserved content %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, `id = "alpha"`) || strings.Contains(text, `mystery = "keep me"`) {
		t.Errorf("the forgotten entry should be the only removed block:\n%s", text)
	}
	got, err := Spaces()
	if err != nil || len(got) != 2 || got[0].ID != "beta" || got[1].ID != "gamma" {
		t.Errorf("edited registry must still decode in retained order, got %v / %v", got, err)
	}
}

// TestSpaceWrites_FollowRegistrySymlink protects the dotfiles use case. Atomic rename
// targets the symlink's destination; it must not replace spaces.toml with a regular file.
func TestSpaceWrites_FollowRegistrySymlink(t *testing.T) {
	dir := spaceHome(t)
	target := filepath.Join(t.TempDir(), "committed-spaces.toml")
	if err := os.WriteFile(target, []byte(initialSpacesText), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, SpacesFile)
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, _, err := AddSpace(Space{ID: "linked", Path: t.TempDir()}); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Lstat(link); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("registry symlink was replaced: %v / %v", fi, err)
	}
	b, err := os.ReadFile(target)
	if err != nil || !strings.Contains(string(b), `id = "linked"`) {
		t.Errorf("symlink target was not edited: %v\n%s", err, b)
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
