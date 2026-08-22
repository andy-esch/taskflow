package spacestore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
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
