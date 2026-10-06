package workspacestore

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestFS_OpenWorkspaceResolvesDirectAndPointerEntries(t *testing.T) {
	planning := t.TempDir()
	initialized, err := config.Init(planning, "planning", false)
	if err != nil {
		t.Fatal(err)
	}
	planningConfig, err := config.Discover(planning)
	if err != nil {
		t.Fatal(err)
	}
	threadID := testutil.TaskID("workspace-thread")
	threadPath := filepath.Join(planningConfig.Root, domain.ThreadsDir, threadID+"-workspace-thread.md")
	testutil.Write(t, threadPath, "malformed but path-resolvable\n")
	pointer := t.TempDir()
	if _, err := config.InitPointer(pointer, planning, false); err != nil {
		t.Fatal(err)
	}
	pointerConfig, err := config.Discover(pointer)
	if err != nil {
		t.Fatal(err)
	}

	service := core.NewWorkspaceService(testutil.Must(New(core.ReadOnlyMutations())))
	direct, err := service.Open(core.WorkspaceRequest{Start: planning, SpaceID: "planning"})
	if err != nil {
		t.Fatal(err)
	}
	fromPointer, err := service.Open(core.WorkspaceRequest{Start: pointer, SpaceID: "implementation"})
	if err != nil {
		t.Fatal(err)
	}
	if direct.PlanningRoot != planningConfig.Root || direct.PlanningID != initialized.PlanningID || direct.Checkout != planningConfig.Dir {
		t.Fatalf("direct workspace = %+v", direct)
	}
	if fromPointer.PlanningRoot != direct.PlanningRoot || fromPointer.PlanningID != direct.PlanningID ||
		fromPointer.Checkout != pointerConfig.Dir || fromPointer.SpaceID != "implementation" {
		t.Fatalf("pointer workspace = %+v, direct = %+v", fromPointer, direct)
	}
	if len(fromPointer.Layout.WatchPaths()) != 5 {
		t.Fatalf("watch paths = %v", fromPointer.Layout.WatchPaths())
	}
	for name, workspace := range map[string]core.Workspace{"direct": direct, "pointer": fromPointer} {
		got, err := workspace.Planning.ThreadPath("workspace-thread")
		if err != nil || got != threadPath {
			t.Fatalf("%s workspace Thread path = %q, %v; want %q", name, got, err, threadPath)
		}
	}
}

func TestFS_OpenWorkspacePreservesDiscoveryFailureCause(t *testing.T) {
	planning := t.TempDir()
	initialized, err := config.Init(planning, "", false)
	if err != nil {
		t.Fatal(err)
	}
	pointer := t.TempDir()
	if _, err := config.InitPointer(pointer, planning, false); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(planning, config.ConfigFile)
	if err := os.WriteFile(configPath, []byte("id = \"different\"\ntaskflow_root = \".\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if initialized.PlanningID == "different" {
		t.Fatal("test requires a changed planning id")
	}

	_, err = core.NewWorkspaceService(testutil.Must(New(core.ReadOnlyMutations()))).Open(core.WorkspaceRequest{Start: pointer})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("identity mismatch error = %v", err)
	}
}

func TestFS_MultiWorkspaceSourceSetsCannotBeCrossWired(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	for _, root := range []string{first, second} {
		if _, err := config.Init(root, "", false); err != nil {
			t.Fatal(err)
		}
	}
	opener := testutil.Must(New(core.ReadOnlyMutations()))
	a, err := opener.OpenWorkspace(first)
	if err != nil {
		t.Fatal(err)
	}
	b, err := opener.OpenWorkspace(second)
	if err != nil {
		t.Fatal(err)
	}
	aID := a.Store.(core.SourceSetProvider).SourceSetID()
	bID := b.Store.(core.SourceSetProvider).SourceSetID()
	if a.TaskPaths == nil || a.EpicPaths == nil || a.AuditPaths == nil || a.ResearchPaths == nil {
		t.Fatalf("filesystem workspace omitted entity-local path capabilities: %+v", a)
	}
	if aID.IsZero() || bID.IsZero() || aID == bID {
		t.Fatal("independent workspaces must expose distinct source sets")
	}
	if svc, err := core.NewService(a.Store, core.WithThreadStore(b.Threads)); svc != nil ||
		!errors.Is(err, core.ErrIncompatibleCapabilities) {
		t.Fatalf("cross-wired workspace = %v, %v", svc, err)
	}
	service := core.NewWorkspaceService(opener)
	if _, err := service.Open(core.WorkspaceRequest{Start: first}); err != nil {
		t.Fatalf("first independent workspace: %v", err)
	}
	if _, err := service.Open(core.WorkspaceRequest{Start: second}); err != nil {
		t.Fatalf("second independent workspace: %v", err)
	}
}

func TestFS_OpenWorkspaceDoesNotFallbackFromMissingOrMalformedEntry(t *testing.T) {
	service := core.NewWorkspaceService(testutil.Must(New(core.ReadOnlyMutations())))
	missing := t.TempDir()
	if workspace, err := service.Open(core.WorkspaceRequest{Start: missing}); err == nil ||
		workspace.Planning != nil || !strings.Contains(err.Error(), "not a taskflow planning repo") {
		t.Fatalf("missing entry = %+v, %v", workspace, err)
	}

	malformed := t.TempDir()
	if err := os.WriteFile(filepath.Join(malformed, config.ConfigFile), []byte("[[broken\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if workspace, err := service.Open(core.WorkspaceRequest{Start: malformed}); err == nil ||
		workspace.Planning != nil || !strings.Contains(err.Error(), "parse") {
		t.Fatalf("malformed entry = %+v, %v", workspace, err)
	}
}

func TestWorkspaceOpeningCarriesPolicyAcrossDirectAndPointerEntryPoints(t *testing.T) {
	for _, pointer := range []bool{false, true} {
		for _, mode := range []string{"read-only", "guarded", "unrestricted"} {
			t.Run(map[bool]string{true: "pointer", false: "direct"}[pointer]+"/"+mode, func(t *testing.T) {
				repo := t.TempDir()
				if _, err := config.Init(repo, "", false); err != nil {
					t.Fatal(err)
				}
				cfg, err := config.Discover(repo)
				if err != nil {
					t.Fatal(err)
				}
				testutil.Write(t, filepath.Join(cfg.Root, domain.TasksDir, testutil.TaskID("probe")+"-probe.md"),
					"---\nid: "+testutil.TaskID("probe")+"\nstatus: ready-to-start\ndescription: probe\ntags: [test]\n---\n")
				start := repo
				if pointer {
					start = t.TempDir()
					if _, err := config.InitPointer(start, repo, false); err != nil {
						t.Fatal(err)
					}
				}
				denied := errors.New("late workspace denied")
				calls := 0
				decision := denied
				policy := core.ReadOnlyMutations()
				switch mode {
				case "unrestricted":
					policy = core.UnrestrictedMutations()
				case "guarded":
					policy = core.GuardedMutations(func() error { calls++; return decision })
				}
				opener := testutil.Must(New(policy))
				workspace, err := core.NewWorkspaceService(opener).Open(core.WorkspaceRequest{Start: start})
				if err != nil || calls != 0 {
					t.Fatalf("opening authorized mutation: workspace=%v calls=%d err=%v", workspace, calls, err)
				}
				for _, dryRun := range []bool{true, false} {
					_, err := workspace.Planning.SetFields("probe", map[string]any{"priority": "low"}, false, dryRun)
					want := core.ErrReadOnlyPersistence
					switch mode {
					case "guarded":
						want = denied
					case "unrestricted":
						want = nil
					}
					if !errors.Is(err, want) {
						t.Fatalf("late workspace mutation = %v; want %v", err, want)
					}
				}
				if mode == "guarded" {
					if calls != 2 {
						t.Fatalf("guard was not invoked for each refusal: %d", calls)
					}
					decision = nil
					if _, err := workspace.Planning.SetFields("probe", map[string]any{"priority": "low"}, false, false); err != nil || calls <= 2 {
						t.Fatalf("authorization cached after opening: calls=%d err=%v", calls, err)
					}
					// A successful operation may cross multiple guarded persistence
					// boundaries. Assert a fresh decision, not their implementation count.
					previous := calls
					decision = denied
					if _, err := workspace.Planning.SetFields("probe", map[string]any{"priority": "high"}, false, true); !errors.Is(err, denied) || calls != previous+1 {
						t.Fatalf("allowed decision was cached: calls=%d previous=%d err=%v", calls, previous, err)
					}
				}
			})
		}
	}
}

func TestZeroWorkspaceAdapterRefusesBeforeDiscovery(t *testing.T) {
	if opened, err := (&FS{}).OpenWorkspace(t.TempDir()); !errors.Is(err, core.ErrInvalidMutationPolicy) || opened.Store != nil {
		t.Fatalf("zero workspace opened/discovered planning: %+v %v", opened, err)
	}
}
