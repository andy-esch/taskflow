package spacestore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/spacehealth"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

func TestFSRegistryAdapter_PreparesListsAndMutatesThroughCoreValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv(userconfig.DirEnv, home)
	repo := initializedRegistryRepo(t)
	nested := filepath.Join(repo, "tasks", "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	adapter := New()
	prepared, err := adapter.PrepareSpace(nested)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Discover(repo)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Checkout != cfg.Dir || prepared.VerifyID != cfg.ID {
		t.Fatalf("prepared = %+v, config = %+v", prepared, cfg)
	}

	service := core.NewSpaceRegistryService(adapter)
	added, err := service.Add(nested, "primary", false)
	if err != nil || !added.Changed {
		t.Fatalf("Add = %+v, %v", added, err)
	}
	catalog, err := service.Catalog()
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Entries) != 1 {
		t.Fatalf("catalog = %+v", catalog)
	}
	entry := catalog.Entries[0]
	if entry.ID != "primary" || entry.Path != userconfig.TildePath(prepared.Checkout) || entry.Checkout != cfg.Dir ||
		entry.VerifyID != cfg.ID || entry.PlanningID != cfg.ID || entry.Role != core.SpaceRoleDirect ||
		entry.State != core.SpaceStateEmpty {
		t.Fatalf("entry = %+v", entry)
	}

	unchanged, err := service.Add(repo, "alias", true)
	if err != nil || unchanged.Changed || unchanged.Entry.ID != "primary" || !unchanged.DryRun {
		t.Fatalf("same-path preview = %+v, %v", unchanged, err)
	}
	forgotten, err := service.Forget("primary", true)
	if err != nil || !forgotten.Changed || !forgotten.DryRun {
		t.Fatalf("forget preview = %+v, %v", forgotten, err)
	}
	if entries, err := adapter.ListSpaceEntries(); err != nil || len(entries) != 1 {
		t.Fatalf("preview changed registry: entries=%+v err=%v", entries, err)
	}
	if _, err := service.Forget("primary", false); err != nil {
		t.Fatal(err)
	}
	if entries, err := adapter.ListSpaceEntries(); err != nil || len(entries) != 0 {
		t.Fatalf("forget left registry entry: entries=%+v err=%v", entries, err)
	}
}

func TestFSRegistryAdapter_PreservesPointerEntryAndClassifiesRegistryErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv(userconfig.DirEnv, home)
	planning := initializedRegistryRepo(t)
	pointer := t.TempDir()
	if _, err := config.InitPointer(pointer, planning, false); err != nil {
		t.Fatal(err)
	}
	adapter := New()
	service := core.NewSpaceRegistryService(adapter)
	mutation, err := service.Add(pointer, "implementation", false)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := config.Discover(pointer)
	if err != nil {
		t.Fatal(err)
	}
	if mutation.Entry.Role != core.SpaceRolePointer || mutation.Entry.Checkout != resolved.Dir || mutation.Entry.Root != resolved.Root {
		t.Fatalf("pointer entry = %+v", mutation.Entry)
	}

	other := initializedRegistryRepo(t)
	if _, err := service.Add(other, "implementation", true); domain.Classify(err) != domain.ClassConflict {
		t.Fatalf("label collision error = %v", err)
	}

	registryPath := filepath.Join(home, userconfig.SpacesFile)
	if err := os.WriteFile(registryPath, []byte("[[space]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Catalog(); domain.Classify(err) != domain.ClassValidation {
		t.Fatalf("malformed registry error = %v", err)
	}

	operational := errors.New("disk unavailable")
	if got := classifyRegistryError(operational); got != operational {
		t.Fatalf("operational error lost identity: %v", got)
	}
}

func initializedRegistryRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if _, err := config.Init(dir, "", false); err != nil {
		t.Fatalf("init repo: %v", err)
	}
	return dir
}

// A registry row records how a person spelled the path; a runtime workspace records what
// config.Discover resolved. Checkout is the value consumers OPEN and COMPARE by, so it has
// to be the resolved one — otherwise a symlinked registry entry never matches the tree the
// user is standing in, and the atlas drops its "current" badge.
func TestFSRegistryAdapter_CheckoutIsTheResolvedPathNotTheRegistrySpelling(t *testing.T) {
	home := t.TempDir()
	t.Setenv(userconfig.DirEnv, home)
	repo := initializedRegistryRepo(t)
	link := filepath.Join(t.TempDir(), "linked-repo")
	if err := os.Symlink(repo, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	// `space add` normalizes through PrepareSpace, so the case that matters is a row whose
	// stored path is NOT the resolved one: a hand-edited spaces.toml, or a path that became
	// a symlink after it was registered.
	entry := toCoreEntry(spacehealth.DiagnoseSpace(userconfig.Space{ID: "linked", Path: link}))
	cfg, err := config.Discover(link)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Checkout != cfg.Dir {
		t.Fatalf("entry checkout = %q, want the resolved %q", entry.Checkout, cfg.Dir)
	}
	if entry.Path != link {
		t.Fatalf("Path = %q, want the registered spelling %q retained for output", entry.Path, link)
	}
	if entry.Checkout == entry.Path {
		t.Fatal("this case is only meaningful when the two spellings differ")
	}
}

// An entry too broken to discover has nothing resolved to fall back on, so it keeps the
// only thing known about it — the registry spelling, tilde-expanded.
func TestFSRegistryAdapter_UndiscoverableEntryKeepsItsRegisteredPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv(userconfig.DirEnv, home)
	gone := filepath.Join(t.TempDir(), "vanished")
	entry := toCoreEntry(spacehealth.DiagnoseSpace(userconfig.Space{ID: "gone", Path: gone}))
	if entry.Checkout != gone || entry.Healthy() {
		t.Fatalf("missing entry = %+v, want the registered path retained", entry)
	}
}
