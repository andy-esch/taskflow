package userconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	if added, _, err := AddSpace(Space{ID: "a", Path: repo}, false); err != nil || !added {
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
	added, existing, err := AddSpace(Space{ID: "b", Path: spelling}, false)
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
	added, existing, err = AddSpace(Space{ID: "via-link", Path: link}, false)
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
	if _, _, err := AddSpace(Space{ID: "main", Path: a, VerifyID: "shared0000aa"}, false); err != nil {
		t.Fatal(err)
	}
	added, _, err := AddSpace(Space{ID: "wt", Path: b, VerifyID: "shared0000aa"}, false)
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
	if _, _, err := AddSpace(Space{ID: "dup", Path: t.TempDir()}, false); err != nil {
		t.Fatal(err)
	}
	if _, _, err := AddSpace(Space{ID: "dup", Path: t.TempDir()}, false); !errors.Is(err, ErrSpaceIDConflict) {
		t.Errorf("a second space claiming the same label must return ErrSpaceIDConflict, got %v", err)
	}
}

// TestAddSpace_DryRunUsesRealValidation regresses the review finding: dry-run must plan
// against the registry snapshot, not report changed=true before path dedup / id-collision
// checks. A new preview includes the date it would write and leaves no file behind.
func TestAddSpace_DryRunUsesRealValidation(t *testing.T) {
	dir := spaceHome(t)
	repo := t.TempDir()
	added, preview, err := AddSpace(Space{ID: "preview", Path: repo}, true)
	if err != nil || !added || preview.Added == "" {
		t.Fatalf("new dry-run plan: added=%v preview=%+v err=%v", added, preview, err)
	}
	if _, err := os.Stat(filepath.Join(dir, SpacesFile)); !os.IsNotExist(err) {
		t.Errorf("dry-run created a registry, stat err=%v", err)
	}
	if _, _, err := AddSpace(Space{ID: "real", Path: repo}, false); err != nil {
		t.Fatal(err)
	}
	if added, existing, err := AddSpace(Space{ID: "alias", Path: repo}, true); err != nil || added || existing.ID != "real" {
		t.Errorf("same-path dry-run must be the real no-op: added=%v existing=%+v err=%v", added, existing, err)
	}
	if _, _, err := AddSpace(Space{ID: "real", Path: t.TempDir()}, true); !errors.Is(err, ErrSpaceIDConflict) {
		t.Errorf("dry-run must detect the real label collision, got %v", err)
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
	if _, _, err := AddSpace(Space{ID: "gone", Path: repo}, false); err != nil {
		t.Fatal(err)
	}
	removed, _, err := ForgetSpace("gone", false)
	if err != nil || !removed {
		t.Fatalf("forget: %v / %v", removed, err)
	}
	if got, _ := Spaces(); len(got) != 0 {
		t.Errorf("entry should be gone, got %v", got)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Error("forgetting must NOT touch the repo on disk")
	}
	if again, _, _ := ForgetSpace("gone", false); again {
		t.Error("forgetting an absent id should report no change, not fail")
	}
}

// TestSpaceWrites_PreserveInsertionOrderAndRoundTrip pins the ordering decision:
// registrations stay in insertion order. Re-sorting a hand-edited array of tables would
// move comments with surprising entries and violate the surgical-write contract.
func TestSpaceWrites_PreserveInsertionOrderAndRoundTrip(t *testing.T) {
	dir := spaceHome(t)
	for _, id := range []string{"zeta", "alpha", "mid"} {
		if _, _, err := AddSpace(Space{ID: id, Path: t.TempDir()}, false); err != nil {
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
	if added, _, err := AddSpace(Space{ID: "gamma", Path: t.TempDir(), Added: "2026-08-20"}, false); err != nil || !added {
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

	if removed, _, err := ForgetSpace("alpha", false); err != nil || !removed {
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

// TestForgetSpace_PreservesQuotedFollowingTable is the adversarial regression for the
// review finding. '#' inside a quoted TOML table key is data, not a comment; missing that
// boundary made forget delete the unrelated table through EOF. Its comment prelude is
// preserved too because ownership of inter-table comments is ambiguous.
func TestForgetSpace_PreservesQuotedFollowingTable(t *testing.T) {
	dir := spaceHome(t)
	path := filepath.Join(dir, SpacesFile)
	original := `schema_version = 1

[[space]]
id = "alpha"
path = "/tmp/alpha"

# documentation for the archive table
["notes#archive"] # quoted hash is part of the key
text = "must survive"
`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	removed, _, err := ForgetSpace("alpha", false)
	if err != nil || !removed {
		t.Fatalf("forget alpha: removed=%v err=%v", removed, err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	for _, want := range []string{
		"# documentation for the archive table",
		`["notes#archive"] # quoted hash is part of the key`,
		`text = "must survive"`,
	} {
		if !strings.Contains(text, want) {
			t.Errorf("forget dropped unrelated TOML %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, `id = "alpha"`) {
		t.Errorf("forgotten entry remains:\n%s", text)
	}
	if got, err := Spaces(); err != nil || len(got) != 0 {
		t.Errorf("result must remain valid registry TOML: %v / %v", got, err)
	}
}

// TestConcurrentAddSpace_NoLostUpdates exercises the real cross-process lock primitive
// with independent file descriptors in concurrent goroutines. Every successful add must
// survive; atomic rename alone would let last-writer-wins snapshots silently drop entries.
func TestConcurrentAddSpace_NoLostUpdates(t *testing.T) {
	spaceHome(t)
	const count = 12
	paths := make([]string, count)
	for i := range paths {
		paths[i] = t.TempDir()
	}
	start := make(chan struct{})
	errs := make(chan error, count)
	var wg sync.WaitGroup
	for i := range count {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			added, _, err := AddSpace(Space{ID: fmt.Sprintf("space-%02d", i), Path: paths[i]}, false)
			if err != nil {
				errs <- err
				return
			}
			if !added {
				errs <- fmt.Errorf("space-%02d unexpectedly reported a no-op", i)
			}
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	spaces, err := Spaces()
	if err != nil {
		t.Fatal(err)
	}
	if len(spaces) != count {
		t.Fatalf("lost concurrent registry updates: got %d entries, want %d: %v", len(spaces), count, spaces)
	}
	seen := make(map[string]bool, count)
	for _, space := range spaces {
		seen[space.ID] = true
	}
	for i := range count {
		if id := fmt.Sprintf("space-%02d", i); !seen[id] {
			t.Errorf("concurrent add lost %s", id)
		}
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
	if _, _, err := AddSpace(Space{ID: "linked", Path: t.TempDir()}, false); err != nil {
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

// A dangling registry symlink must not read as "no spaces registered". The two are the
// same ENOENT but opposite situations: one means nothing was ever registered, the other
// means everything registered is temporarily out of reach — and answering the second with
// the first invites re-registering spaces that still exist, after which the next
// `space add` writes a fresh registry over the dangling link.
func TestSpaces_BrokenRegistrySymlinkIsReportedNotReadAsEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv(DirEnv, home)
	target := filepath.Join(t.TempDir(), "unmounted-dotfiles", SpacesFile)
	if err := os.Symlink(target, filepath.Join(home, SpacesFile)); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	spaces, err := Spaces()
	if err == nil {
		t.Fatalf("Spaces() = %v, nil — a dangling registry symlink must be reported", spaces)
	}
	if !strings.Contains(err.Error(), "broken symlink") || !strings.Contains(err.Error(), target) {
		t.Fatalf("error = %q, want it to name the broken link and its target", err.Error())
	}
	if spaces != nil {
		t.Fatalf("Spaces() returned %v alongside the error; the registry is unknown, not empty", spaces)
	}
}

// A registry that genuinely does not exist is still the ordinary, silent case: a machine
// that has never run `space add` must behave exactly as it did before the file existed.
func TestSpaces_AbsentRegistryStaysSilentlyEmpty(t *testing.T) {
	t.Setenv(DirEnv, t.TempDir())
	spaces, err := Spaces()
	if err != nil || len(spaces) != 0 {
		t.Fatalf("Spaces() = %v, %v; want the silent empty registry", spaces, err)
	}
}

// A symlink whose target IS present is an ordinary readable registry — the dotfiles setup
// working as intended, which the probe must not flag.
func TestSpaces_HealthyRegistrySymlinkReadsThrough(t *testing.T) {
	home := t.TempDir()
	t.Setenv(DirEnv, home)
	dotfiles := t.TempDir()
	real := filepath.Join(dotfiles, SpacesFile)
	if err := os.WriteFile(real, []byte("schema_version = 1\n\n[[space]]\n  id = \"linked\"\n  path = \"~/git/linked\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, filepath.Join(home, SpacesFile)); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	spaces, err := Spaces()
	if err != nil || len(spaces) != 1 || spaces[0].ID != "linked" {
		t.Fatalf("Spaces() = %v, %v; want the linked registry read through", spaces, err)
	}
}
