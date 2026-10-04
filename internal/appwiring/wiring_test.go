package appwiring

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

func TestLocalCompositionSelectsOneObservedDirectOrPointerCorpus(t *testing.T) {
	planning, pointer, decoy := t.TempDir(), t.TempDir(), t.TempDir()
	if _, err := config.Init(planning, "planning", false); err != nil {
		t.Fatal(err)
	}
	if _, err := config.InitPointer(pointer, planning, false); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Init(decoy, "", false); err != nil {
		t.Fatal(err)
	}
	intendedID, decoyID := testutil.TaskID("intended"), testutil.TaskID("decoy")
	testutil.Write(t, filepath.Join(planning, "planning", domain.TasksDir, intendedID+"-intended.md"),
		"---\nid: "+intendedID+"\nstatus: ready-to-start\ndescription: intended\ntags: [test]\n---\n")
	testutil.Write(t, filepath.Join(decoy, domain.TasksDir, decoyID+"-decoy.md"),
		"---\nid: "+decoyID+"\nstatus: ready-to-start\ndescription: decoy\ntags: [test]\n---\n")
	t.Chdir(decoy) // a wrong-root store must observe distinguishable real records
	services := LocalBindings().Compose(func() error { return nil })
	for _, entry := range []string{planning, pointer} {
		opened, err := services.OpenPlanning(entry)
		cfg, cfgErr := config.Discover(entry)
		if err != nil || cfgErr != nil || opened.Service == nil || opened.Layout == nil || opened.Repository.Dir != cfg.Dir || opened.Repository.PlanningRoot != cfg.Root || opened.Repository.ID != cfg.ID {
			t.Fatalf("entry=%q opened=%+v err=%v config=%+v err=%v", entry, opened, err, cfg, cfgErr)
		}
		expectedPaths := []string{}
		for _, dir := range []string{domain.EpicsDir, domain.TasksDir, domain.AuditsDir, domain.ResearchDir, domain.ThreadsDir} {
			expectedPaths = append(expectedPaths, filepath.Join(cfg.Root, dir))
		}
		if !slices.Equal(opened.Layout.WatchPaths(), expectedPaths) {
			t.Fatalf("watch layout not preserved: %v", opened.Layout.WatchPaths())
		}
		records, problems, err := opened.Service.ListTasks(core.TaskFilter{})
		if err != nil || len(problems) != 0 || len(records) != 1 || records[0].Value.ID != intendedID {
			t.Fatalf("service opened another corpus: records=%+v problems=%+v err=%v", records, problems, err)
		}
	}
}

func TestLocalBindingReadsAreDeferredUntilTheirHooks(t *testing.T) {
	t.Setenv(userconfig.DirEnv, t.TempDir())
	repo := t.TempDir()
	if _, err := config.Init(repo, "", false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo) // an accidental early local open would succeed here
	var starts []string
	userReads := 0
	bindings := bindingsFor(localSources{
		discover: func(start string) (*config.Config, error) {
			starts = append(starts, start)
			return config.Discover(start)
		},
		user: func() (*userconfig.Config, error) {
			userReads++
			return userconfig.Load()
		},
	})
	services := bindings.Compose(func() error { return nil })
	if len(starts) != 0 || userReads != 0 {
		t.Fatalf("real adapter composition read startup data: starts=%v userReads=%d", starts, userReads)
	}
	if _, err := services.OpenPlanning(repo); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(starts, []string{repo}) || userReads != 0 {
		t.Fatalf("opening performed extra discovery/presentation reads: starts=%v userReads=%d", starts, userReads)
	}
	if _, err := bindings.ReadRepository(repo); err != nil {
		t.Fatal(err)
	}
	if _, err := bindings.ReadUser(); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(starts, []string{repo, repo}) || userReads != 1 {
		t.Fatalf("explicit readers performed extra reads: starts=%v userReads=%d", starts, userReads)
	}
}

func TestLocalCompositionPropagatesAuthorizationToEveryPersistenceFamily(t *testing.T) {
	t.Setenv(userconfig.DirEnv, t.TempDir())
	repo := t.TempDir()
	if _, err := config.Init(repo, "", false); err != nil {
		t.Fatal(err)
	}
	taskID := testutil.TaskID("authorization")
	path := filepath.Join(repo, domain.TasksDir, taskID+"-authorization.md")
	before := "---\nid: " + taskID + "\nstatus: ready-to-start\npriority: medium\ndescription: authorization\ntags: [test]\n---\n"
	testutil.Write(t, path, before)
	sentinel := errors.New("composition authorizer refused")
	calls := 0
	services := LocalBindings().Compose(func() error { calls++; return sentinel })
	opened, err := services.OpenPlanning(repo)
	if err != nil {
		t.Fatal(err)
	}
	workspace, err := services.Workspaces.Open(core.WorkspaceRequest{Start: repo})
	if err != nil {
		t.Fatal(err)
	}
	for _, dryRun := range []bool{true, false} {
		operations := []func() error{
			func() error {
				_, err := opened.Service.SetFields("authorization", map[string]any{"priority": "low"}, false, dryRun)
				return err
			},
			func() error { _, err := opened.Service.RepairPlanning(dryRun); return err },
			func() error {
				_, err := services.Configuration.SetPreference(repo, core.PreferenceChange{Scope: core.ConfigScopeUser, Field: core.PreferencePagerCommand, Value: "less"}, dryRun)
				return err
			},
			func() error { _, err := services.Spaces.Add(repo, "probe", dryRun); return err },
			func() error {
				_, err := workspace.Planning.SetFields("authorization", map[string]any{"priority": "low"}, false, dryRun)
				return err
			},
		}
		for index, operation := range operations {
			previous := calls
			if err := operation(); !errors.Is(err, sentinel) || calls != previous+1 {
				t.Fatalf("dryRun=%t operation=%d err=%v calls=%d previous=%d", dryRun, index, err, calls, previous)
			}
		}
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != before {
		t.Fatal("denied composition changed planning data")
	}
	home, err := userconfig.Dir()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 0 {
		t.Fatalf("denied composition changed home storage: entries=%v err=%v", entries, err)
	}
}

func TestOpenedPlanningRechecksIdentityBeforeThreadApply(t *testing.T) {
	repo := t.TempDir()
	if _, err := config.Init(repo, "", false); err != nil {
		t.Fatal(err)
	}
	taskID := testutil.TaskID("member")
	testutil.Write(t, filepath.Join(repo, domain.TasksDir, taskID+"-member.md"), "---\nid: "+taskID+"\nstatus: ready-to-start\n---\n")
	services := LocalBindings().Compose(func() error { return nil })
	opened, err := services.OpenPlanning(repo)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := opened.Service.ComposeThreadApply(opened.Repository.ID, core.ThreadComposeManifest{
		Thread: core.ThreadComposeInput{Title: "Probe", Description: "Probe", Goal: "Probe", Tags: []string{"test"}},
		Nodes:  []core.ThreadComposeNode{{Key: "member", TaskID: taskID}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := opened.Service.ApplyThreadPlan(plan, true); err != nil {
		t.Fatalf("initial dry-run could not re-read identity: %v", err)
	}
	testutil.Write(t, filepath.Join(repo, config.ConfigFile), "id = \"replaced\"\ntaskflow_root = \".\"\n")
	if receipt, err := opened.Service.ApplyThreadPlan(plan, true); !errors.Is(err, domain.ErrConflict) || receipt.Committed {
		t.Fatalf("replaced planning identity was accepted: receipt=%+v err=%v", receipt, err)
	}
}
