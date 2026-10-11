package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// A successful bind must be readable on its very next discovery, regardless of
// which spelling addresses the same physical, marker-governed planning tree.
func TestPlanningTarget_BindingRoundTrip(t *testing.T) {
	for _, tree := range []string{".", "planning", "docs/work"} {
		for _, address := range []string{"repo", "data", "symlink"} {
			for _, mode := range []string{"init", "migrate"} {
				t.Run(tree+"/"+address+"/"+mode, func(t *testing.T) {
					repo, impl := t.TempDir(), t.TempDir()
					mkdirs(t, filepath.Join(repo, tree, domain.TasksDir))
					marker := fmt.Sprintf("# keep topology\nid = \"stable-id\"\ntaskflow_root = %q\nunknown = \"keep\"\n[theme]\nname = \"dark\"\n", tree)
					writeConfig(t, repo, marker)
					data := eval(t, filepath.Join(repo, tree))
					target := repo
					if address != "repo" {
						target = data
					}
					if address == "symlink" {
						target = filepath.Join(t.TempDir(), "alias")
						if err := os.Symlink(data, target); err != nil {
							t.Skipf("symlinks unavailable: %v", err)
						}
					}
					before := testutil.SnapshotTree(t, repo)
					if mode == "init" {
						result, err := InitPointer(impl, target, false)
						if err != nil || result.PlanningID != "stable-id" {
							t.Fatalf("init = %+v, %v", result, err)
						}
					} else {
						writeConfig(t, impl, fmt.Sprintf("# keep pointer\nplanning_repo = %q\nunknown = \"keep\"\n", filepath.ToSlash(target)))
						preview, err := Migrate(impl, true)
						if err != nil || len(preview.Steps) != 1 || preview.Steps[0].Value != "stable-id" {
							t.Fatalf("preview = %+v, %v", preview, err)
						}
						if result, err := Migrate(impl, false); err != nil || !reflect.DeepEqual(preview.Steps, result.Steps) {
							t.Fatalf("migrate = %+v, %v", result, err)
						}
						b, err := os.ReadFile(filepath.Join(impl, ConfigFile))
						if err != nil || !strings.Contains(string(b), "# keep pointer") || !strings.Contains(string(b), "unknown = \"keep\"") {
							t.Fatalf("migration lost unknown pointer configuration: %s, %v", b, err)
						}
					}
					cfg, err := Discover(impl)
					if err != nil || cfg.Root != data || cfg.ID != "stable-id" {
						t.Fatalf("bound pointer discovery = %+v, %v", cfg, err)
					}
					// Direct discovery of a data-directory alias has the same governing
					// marker, identity and preferences, not a new markerless space.
					cfg, err = Discover(target)
					if err != nil || cfg.Root != data || cfg.ID != "stable-id" || cfg.Dir != eval(t, repo) || cfg.Theme.Name != "dark" {
						t.Fatalf("direct discovery = %+v, %v", cfg, err)
					}
					if !reflect.DeepEqual(before, testutil.SnapshotTree(t, repo)) {
						t.Fatal("binding changed planning marker topology or content")
					}
				})
			}
		}
	}
}

func TestPlanningTarget_AliasCannotBorrowUnrelatedIdentity(t *testing.T) {
	for _, kind := range []string{"different-root", "git-directory", "git-file", "pointer-ancestor", "shadow-marker", "missing-id", "changed-id", "malformed-marker", "escaping-root"} {
		t.Run(kind, func(t *testing.T) {
			repo, impl := t.TempDir(), t.TempDir()
			data := filepath.Join(repo, "planning")
			mkdirs(t, filepath.Join(data, domain.TasksDir))
			marker := "id = \"expected\"\ntaskflow_root = \"planning\"\n"
			wantErr := domain.ErrConflict
			switch kind {
			case "different-root":
				mkdirs(t, filepath.Join(repo, "other", domain.TasksDir))
				marker = "id = \"expected\"\ntaskflow_root = \"other\"\n"
			case "git-directory":
				mkdirs(t, filepath.Join(data, ".git"))
			case "git-file":
				if err := os.WriteFile(filepath.Join(data, ".git"), []byte("gitdir: missing\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			case "pointer-ancestor":
				// A self-referencing pointer must not recurse during identity lookup.
				marker = "id = \"expected\"\nplanning_repo = \"planning\"\n"
			case "shadow-marker":
				writeConfig(t, data, "taskflow_root = \".\"\n")
			case "missing-id":
				marker = "taskflow_root = \"planning\"\n"
			case "changed-id":
				marker = "id = \"replacement\"\ntaskflow_root = \"planning\"\n"
			case "malformed-marker":
				marker = "id = [invalid\n"
				wantErr = domain.ErrValidation
			case "escaping-root":
				outside := t.TempDir()
				mkdirs(t, filepath.Join(outside, domain.TasksDir))
				if err := os.Symlink(outside, filepath.Join(repo, "escape")); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
				data = filepath.Join(repo, "escape")
				marker = "id = \"expected\"\ntaskflow_root = \"escape\"\n"
			}
			writeConfig(t, repo, marker)
			writeConfig(t, impl, fmt.Sprintf("planning_repo = %q\nplanning_repo_id = \"expected\"\n", filepath.ToSlash(data)))
			before, beforeImpl := testutil.SnapshotTree(t, repo), testutil.SnapshotTree(t, impl)
			for _, attempt := range []func() error{
				func() error { _, err := Discover(impl); return err },
				func() error { _, err := Migrate(impl, false); return err },
				func() error { _, err := InitPointer(impl, data, false); return err },
			} {
				if err := attempt(); !errors.Is(err, wantErr) {
					t.Fatalf("bound alias error = %v, want %v", err, wantErr)
				}
			}
			if !reflect.DeepEqual(before, testutil.SnapshotTree(t, repo)) || !reflect.DeepEqual(beforeImpl, testutil.SnapshotTree(t, impl)) {
				t.Fatal("refusal mutated planning configuration or pointer binding")
			}
		})
	}
}

func TestPlanningTarget_WorktreeAliasUsesCanonicalCheckout(t *testing.T) {
	parent := t.TempDir()
	home, elsewhere := filepath.Join(parent, "home"), filepath.Join(parent, "elsewhere")
	plan, impl := filepath.Join(home, "plan"), filepath.Join(home, "impl")
	result, err := Init(plan, "docs/work", false)
	if err != nil {
		t.Fatal(err)
	}
	mkdirs(t, impl, elsewhere)
	git(t, impl, "init", "-q", ".")
	const pointer = "../plan/docs/work"
	writeConfig(t, impl, "planning_repo = \""+pointer+"\"\n")
	if _, err := Migrate(impl, false); err != nil {
		t.Fatal(err)
	}
	git(t, impl, "add", ConfigFile)
	git(t, impl, "commit", "-q", "-m", "bound pointer")
	wt := filepath.Join(elsewhere, "impl-wt")
	git(t, impl, "worktree", "add", "-q", wt, "-b", "alias")
	if _, err := Init(filepath.Join(elsewhere, "plan"), "docs/work", false); err != nil {
		t.Fatal(err)
	}
	for _, start := range []string{impl, wt} {
		cfg, err := Discover(start)
		if err != nil || cfg.ID != result.PlanningID || cfg.Root != eval(t, filepath.Join(plan, "docs/work")) {
			t.Fatalf("discovery from %s = %+v, %v", start, cfg, err)
		}
		if _, err := InitPointer(start, pointer, true); err != nil {
			t.Fatalf("idempotent init from %s: %v", start, err)
		}
		if migration, err := Migrate(start, true); err != nil || len(migration.Steps) != 0 {
			t.Fatalf("migration from %s = %+v, %v", start, migration, err)
		}
	}
}

func TestPlanningTarget_LegacyBareTreeDoesNotInheritParentIdentity(t *testing.T) {
	repo, impl := t.TempDir(), t.TempDir()
	data := filepath.Join(repo, "independent")
	mkdirs(t, filepath.Join(data, domain.TasksDir), filepath.Join(repo, "other", domain.TasksDir))
	writeConfig(t, repo, "id = \"unrelated\"\ntaskflow_root = \"other\"\n")
	if result, err := InitPointer(impl, data, false); err != nil || result.PlanningID != "" {
		t.Fatalf("legacy init = %+v, %v", result, err)
	}
	for _, start := range []string{impl, data} {
		cfg, err := Discover(start)
		if err != nil || cfg.ID != "" || cfg.Root != eval(t, data) {
			t.Fatalf("markerless discovery = %+v, %v", cfg, err)
		}
	}
	before := testutil.SnapshotTree(t, impl)
	if _, err := Migrate(impl, false); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("legacy migration must require a deliberate target identity: %v", err)
	}
	if !reflect.DeepEqual(before, testutil.SnapshotTree(t, impl)) {
		t.Fatal("failed migration changed pointer")
	}
}

func TestPlanningTarget_DirectAliasMigratesGoverningMarker(t *testing.T) {
	repo := t.TempDir()
	data := filepath.Join(repo, "docs", "work")
	mkdirs(t, filepath.Join(data, domain.TasksDir))
	writeConfig(t, repo, "# preserve\ntaskflow_root = \"docs/work\"\nunknown = \"keep\"\n")
	result, err := Migrate(data, false)
	if err != nil || result.ConfigPath != filepath.Join(eval(t, repo), ConfigFile) || len(result.Steps) != 1 {
		t.Fatalf("migration through data alias = %+v, %v", result, err)
	}
	for _, start := range []string{repo, data} {
		cfg, err := Discover(start)
		if err != nil || cfg.ID == "" || cfg.ID != result.Steps[0].Value {
			t.Fatalf("migrated alias identity = %+v, %v", cfg, err)
		}
	}
	if _, err := os.Stat(filepath.Join(data, ConfigFile)); !os.IsNotExist(err) {
		t.Fatalf("migration created a second marker: %v", err)
	}
}
