package appwiring

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

func TestPlanningOpenRetainsInitialCorpusWhenMarkerChangesDuringDiscovery(t *testing.T) {
	t.Setenv(userconfig.DirEnv, t.TempDir())
	for _, pointer := range []bool{false, true} {
		t.Run(fmt.Sprintf("pointer=%t", pointer), func(t *testing.T) {
			entry := newThreadApplyEntry(t, pointer, "")
			var alternate string
			calls := 0
			reads := localSources{
				discover: func(start string) (*config.Config, error) {
					cfg, err := config.Discover(start)
					calls++
					if err == nil && calls == 1 {
						// Return the actual first observation, but make any illicit
						// rediscovery see a different, otherwise valid same-ID corpus.
						alternate = repointThreadApplyEntry(t, entry, pointer)
						testutil.Write(t, filepath.Join(alternate, domain.TasksDir, entry.memberID+"-parity-member.md"),
							"---\nid: "+entry.memberID+"\nstatus: ready-to-start\ndescription: replacement corpus\ntags: [test]\n---\n")
					}
					return cfg, err
				},
				user: userconfig.Load,
			}
			opened, err := testutil.Must(bindingsFor(reads).Compose(func() error { return nil })).OpenPlanning(entry.start)
			if err != nil {
				t.Fatal(err)
			}
			if calls != 1 || opened.Repository.PlanningRoot != entry.cfg.Root || opened.Repository.ID != entry.cfg.ID || opened.Repository.Dir != entry.cfg.Dir {
				t.Fatalf("opening lost the initial observation: repository=%+v calls=%d", opened.Repository, calls)
			}
			member, err := opened.Service.ShowTask(entry.memberID)
			if err != nil || member.Value.Task.Description != "parity fixture" {
				t.Fatalf("service and metadata address different corpora: member=%+v err=%v", member, err)
			}
			var paths []string
			for _, dir := range []string{domain.EpicsDir, domain.TasksDir, domain.AuditsDir, domain.ResearchDir, domain.ThreadsDir} {
				paths = append(paths, filepath.Join(entry.cfg.Root, dir))
			}
			if !slices.Equal(opened.Layout.WatchPaths(), paths) {
				t.Fatalf("watcher lost the initially observed corpus: %v", opened.Layout.WatchPaths())
			}
			before := planningDocuments(t, entry.cfg.Root, alternate)
			plan := composeThreadApplyEntry(t, opened.Service, opened.Repository.ID, entry)
			if calls != 1 || !maps.Equal(before, planningDocuments(t, entry.cfg.Root, alternate)) {
				t.Fatal("compose rediscovered the corpus or changed documents")
			}
			for _, dryRun := range []bool{true, false} {
				beforeCalls := calls
				receipt, err := opened.Service.ApplyThreadPlan(plan, dryRun)
				assertThreadApplyRefused(t, plan, receipt, err, domain.ErrConflict)
				if calls <= beforeCalls {
					t.Fatal("apply did not re-read the changed identity")
				}
				if !strings.Contains(err.Error(), "instead of guarded root") {
					t.Fatalf("expected physical-root refusal, got %v", err)
				}
				if !maps.Equal(before, planningDocuments(t, entry.cfg.Root, alternate)) {
					t.Fatal("apply across the changed marker wrote to a corpus")
				}
			}
		})
	}
}

func TestPlanningOpenersThreadApplyParity(t *testing.T) {
	for _, opener := range threadApplyOpeners() {
		t.Run(opener.name, func(t *testing.T) {
			entry := newThreadApplyEntry(t, opener.pointer, "planning")
			service, planningID := openThreadApplyEntry(t, entry.start, opener.workspace, func() error { return nil })
			if planningID != entry.cfg.ID {
				t.Fatalf("opened planning id = %q, want %q", planningID, entry.cfg.ID)
			}
			before := planningDocuments(t, entry.cfg.Root)
			plan := composeThreadApplyEntry(t, service, planningID, entry)
			if !maps.Equal(before, planningDocuments(t, entry.cfg.Root)) {
				t.Fatal("compose changed planning documents")
			}
			receipt, err := service.ApplyThreadPlan(plan, true)
			if err != nil || !receipt.DryRun || !receipt.Changed || receipt.Committed || receipt.Complete || len(receipt.Operations) != 2 {
				t.Fatalf("dry-run receipt=%+v err=%v", receipt, err)
			}
			if !maps.Equal(before, planningDocuments(t, entry.cfg.Root)) {
				t.Fatal("dry-run changed planning documents")
			}
			receipt, err = service.ApplyThreadPlan(plan, false)
			if err != nil || receipt.DryRun || !receipt.Committed || !receipt.Complete {
				t.Fatalf("apply receipt=%+v err=%v", receipt, err)
			}
			member, err := service.ShowTask(entry.memberID)
			if err != nil || !slices.Equal(member.Value.Task.DependsOn, []string{entry.gateID}) {
				t.Fatalf("dependency did not persist: %+v err=%v", member, err)
			}
			view, _, err := service.ShowThread(plan.Thread.ID)
			if err != nil || !slices.Equal(view.Thread.Tasks, []string{entry.memberID}) {
				t.Fatalf("Thread did not persist: %+v err=%v", view, err)
			}
		})
	}
}

func TestPlanningOpenersRejectChangedIdentityBeforeThreadApply(t *testing.T) {
	for _, opener := range threadApplyOpeners() {
		for _, change := range []string{"repoint-root", "replace-id", "remove-marker", "malformed-marker"} {
			for _, dryRun := range []bool{true, false} {
				t.Run(fmt.Sprintf("%s/%s/dry=%t", opener.name, change, dryRun), func(t *testing.T) {
					planningSubdir := "planning"
					if opener.pointer && change == "repoint-root" {
						// Following the resolved target instead of the pointer marker
						// would still pass an initial apply against this flat corpus.
						planningSubdir = ""
					}
					entry := newThreadApplyEntry(t, opener.pointer, planningSubdir)
					service, planningID := openThreadApplyEntry(t, entry.start, opener.workspace, func() error { return nil })
					plan := composeThreadApplyEntry(t, service, planningID, entry)
					if _, err := service.ApplyThreadPlan(plan, true); err != nil {
						t.Fatalf("fixture must be valid before changing identity: %v", err)
					}
					wantClass, wantMessage := error(domain.ErrConflict), ""
					roots := []string{entry.cfg.Root}
					switch change {
					case "repoint-root":
						roots = append(roots, repointThreadApplyEntry(t, entry, opener.pointer))
						wantMessage = "instead of guarded root"
					case "replace-id":
						testutil.Write(t, filepath.Join(entry.planning, config.ConfigFile),
							fmt.Sprintf("id = %q\ntaskflow_root = \"planning\"\n", testutil.TaskID("replaced-planning")))
						wantMessage = "apply plan belongs to planning repository"
						if opener.pointer {
							wantMessage = "this pointer expects"
						}
					case "remove-marker":
						if err := os.Remove(filepath.Join(entry.planning, config.ConfigFile)); err != nil {
							t.Fatal(err)
						}
						wantClass, wantMessage = domain.ErrValidation, "no durable id"
						if opener.pointer {
							wantClass, wantMessage = domain.ErrConflict, "carries no id"
						}
					case "malformed-marker":
						testutil.Write(t, filepath.Join(entry.cfg.Dir, config.ConfigFile), "[[broken\n")
						wantClass, wantMessage = domain.ErrValidation, "parse "+config.ConfigFile
					}
					before := planningDocuments(t, roots...)
					receipt, err := service.ApplyThreadPlan(plan, dryRun)
					assertThreadApplyRefused(t, plan, receipt, err, wantClass)
					if !strings.Contains(err.Error(), wantMessage) {
						t.Fatalf("wrong refusal, want %q: %v", wantMessage, err)
					}
					if !maps.Equal(before, planningDocuments(t, roots...)) {
						t.Fatal("identity refusal changed planning documents")
					}
				})
			}
		}
	}
}

func TestPlanningOpenersPreserveThreadApplyAuthorization(t *testing.T) {
	for _, opener := range threadApplyOpeners() {
		for _, dryRun := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/dry=%t", opener.name, dryRun), func(t *testing.T) {
				entry := newThreadApplyEntry(t, opener.pointer, "planning")
				denied := errors.New("read-only invocation refuses mutation")
				allowed, calls := false, 0
				service, planningID := openThreadApplyEntry(t, entry.start, opener.workspace, func() error {
					calls++
					if !allowed {
						return denied
					}
					return nil
				})
				before := planningDocuments(t, entry.cfg.Root)
				plan := composeThreadApplyEntry(t, service, planningID, entry)
				if calls != 0 {
					t.Fatal("opening/compose must not require mutation authorization")
				}
				receipt, err := service.ApplyThreadPlan(plan, dryRun)
				assertThreadApplyRefused(t, plan, receipt, err, denied)
				if calls != 1 || !maps.Equal(before, planningDocuments(t, entry.cfg.Root)) {
					t.Fatalf("denied apply bypassed authorization or changed documents: calls=%d", calls)
				}
				allowed = true // the opened store must retain the live invocation policy
				receipt, err = service.ApplyThreadPlan(plan, dryRun)
				if err != nil || calls != 2 || receipt.DryRun != dryRun || receipt.Committed == dryRun {
					t.Fatalf("authorized receipt=%+v err=%v calls=%d", receipt, err, calls)
				}
				unchanged := maps.Equal(before, planningDocuments(t, entry.cfg.Root))
				if unchanged != dryRun {
					t.Fatalf("dryRun=%t documents unchanged=%t", dryRun, unchanged)
				}
			})
		}
	}
}

func threadApplyOpeners() []struct {
	name               string
	pointer, workspace bool
} {
	return []struct {
		name               string
		pointer, workspace bool
	}{
		{name: "direct/ordinary"},
		{name: "direct/workspace", workspace: true},
		{name: "pointer/ordinary", pointer: true},
		{name: "pointer/workspace", pointer: true, workspace: true},
	}
}

func assertThreadApplyRefused(t *testing.T, plan core.ThreadApplyPlan, receipt core.ThreadApplyReceipt, err, want error) {
	t.Helper()
	var failure *core.ThreadApplyFailure
	if !errors.Is(err, want) || !errors.As(err, &failure) || receipt.Committed || receipt.Complete || receipt.Plan.Thread.ID != plan.Thread.ID || failure.Receipt.Plan.Thread.ID != plan.Thread.ID {
		t.Fatalf("refusal lost its error class or uncommitted recovery token: receipt=%+v err=%v, want %v", receipt, err, want)
	}
}

func repointThreadApplyEntry(t *testing.T, entry threadApplyEntry, pointer bool) string {
	t.Helper()
	alternate := filepath.Join(entry.planning, "alternate")
	for _, dir := range []string{domain.TasksDir, domain.ThreadsDir, domain.EpicsDir, domain.AuditsDir, domain.ResearchDir} {
		if err := os.MkdirAll(filepath.Join(alternate, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for path, content := range planningDocuments(t, entry.cfg.Root) {
		relative, err := filepath.Rel(entry.cfg.Root, path)
		if err != nil {
			t.Fatal(err)
		}
		testutil.Write(t, filepath.Join(alternate, relative), content)
	}
	// Keep the durable ID and contents unchanged, so only physical-root revalidation
	// can refuse this. Pointer-ID validation must not mask a missing root guard.
	if pointer {
		testutil.Write(t, filepath.Join(alternate, config.ConfigFile), fmt.Sprintf("id = %q\ntaskflow_root = \".\"\n", entry.cfg.ID))
		testutil.Write(t, filepath.Join(entry.cfg.Dir, config.ConfigFile),
			fmt.Sprintf("planning_repo = %q\nplanning_repo_id = %q\n", alternate, entry.cfg.ID))
	} else {
		testutil.Write(t, filepath.Join(entry.cfg.Dir, config.ConfigFile), fmt.Sprintf("id = %q\ntaskflow_root = \"alternate\"\n", entry.cfg.ID))
	}
	fresh, err := config.Discover(entry.start)
	if err != nil || fresh.ID != entry.cfg.ID || fresh.Root == entry.cfg.Root {
		t.Fatalf("invalid root-only replacement fixture: cfg=%+v err=%v", fresh, err)
	}
	return alternate
}

type threadApplyEntry struct {
	start    string
	planning string
	cfg      *config.Config
	gateID   string
	memberID string
}

func newThreadApplyEntry(t *testing.T, pointer bool, planningSubdir string) threadApplyEntry {
	t.Helper()
	planning := t.TempDir()
	if _, err := config.Init(planning, planningSubdir, false); err != nil {
		t.Fatal(err)
	}
	entry := planning
	if pointer {
		entry = t.TempDir()
		if _, err := config.InitPointer(entry, planning, false); err != nil {
			t.Fatal(err)
		}
	}
	// Exercise discovery from below the marker, not only a checkout-root shortcut.
	start := filepath.Join(entry, "work", "nested")
	if err := os.MkdirAll(start, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Discover(start)
	if err != nil {
		t.Fatal(err)
	}
	gateID, memberID := testutil.TaskID("parity-gate"), testutil.TaskID("parity-member")
	for slug, status := range map[string]string{"parity-gate": "completed", "parity-member": "ready-to-start"} {
		testutil.Write(t, filepath.Join(cfg.Root, domain.TasksDir, testutil.TaskID(slug)+"-"+slug+".md"),
			"---\nid: "+testutil.TaskID(slug)+"\nstatus: "+status+"\ndescription: parity fixture\ntags: [test]\n---\n")
	}
	return threadApplyEntry{start: start, planning: planning, cfg: cfg, gateID: gateID, memberID: memberID}
}

func openThreadApplyEntry(t *testing.T, start string, workspace bool, authorize func() error) (*core.Service, string) {
	t.Helper()
	services := testutil.Must(LocalBindings().Compose(authorize))
	if workspace {
		opened, err := services.Workspaces.Open(core.WorkspaceRequest{Start: start})
		if err != nil {
			t.Fatal(err)
		}
		return opened.Planning, opened.PlanningID
	}
	opened, err := services.OpenPlanning(start)
	if err != nil {
		t.Fatal(err)
	}
	return opened.Service, opened.Repository.ID
}

func composeThreadApplyEntry(t *testing.T, service *core.Service, planningID string, entry threadApplyEntry) core.ThreadApplyPlan {
	t.Helper()
	nonmember := false
	plan, err := service.ComposeThreadApply(planningID, core.ThreadComposeManifest{
		Thread: core.ThreadComposeInput{Title: "Parity", Description: "Check local opening parity", Goal: "One guarded corpus", Tags: []string{"test"}},
		Nodes: []core.ThreadComposeNode{
			{Key: "gate", TaskID: entry.gateID, Member: &nonmember},
			{Key: "member", TaskID: entry.memberID},
		},
		Dependencies: []core.ThreadComposeDependency{{From: "gate", To: "member"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func planningDocuments(t *testing.T, roots ...string) map[string]string {
	t.Helper()
	files := make(map[string]string)
	for _, root := range roots {
		for _, dir := range []string{domain.TasksDir, domain.ThreadsDir, domain.EpicsDir, domain.AuditsDir, domain.ResearchDir} {
			if err := filepath.WalkDir(filepath.Join(root, dir), func(path string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !entry.IsDir() {
					content, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					files[path] = string(content)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		}
	}
	return files
}
