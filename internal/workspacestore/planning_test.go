package workspacestore

import (
	"errors"
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
)

func TestNewPlanningStoreUsesOneObservedCorpusAndDefersIdentityReads(t *testing.T) {
	cfg, manifest := planningStoreFixture(t)
	observed := *cfg
	var starts []string
	fs, err := NewPlanningStore(&observed, func(start string) (*config.Config, error) {
		starts = append(starts, start)
		return config.Discover(start)
	}, func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	service, err := core.NewService(fs)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := service.ComposeThreadApply(cfg.ID, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(starts) != 0 {
		t.Fatalf("construction/read-only compose rediscovered the corpus: %v", starts)
	}
	if fs.SourceSetID().IsZero() {
		t.Fatal("shared opening lost its checked source-set witness")
	}
	// Mutating a caller's config value must not redirect the retained start/root.
	observed.Dir, observed.Root = t.TempDir(), t.TempDir()
	if receipt, err := service.ApplyThreadPlan(plan, true); err != nil || receipt.Committed || !receipt.DryRun {
		t.Fatalf("dry-run receipt=%+v err=%v", receipt, err)
	}
	if !slices.Equal(starts, []string{cfg.Dir}) {
		t.Fatalf("identity read did not use the pinned marker: %v", starts)
	}
	if fs.WatchPaths()[0] != filepath.Join(cfg.Root, domain.EpicsDir) {
		t.Fatal("shared opening changed the observed watcher corpus")
	}
}

func TestNewPlanningStoreRejectsMissingConstructionInputs(t *testing.T) {
	for _, tc := range []struct {
		name     string
		cfg      *config.Config
		discover func(string) (*config.Config, error)
	}{
		{name: "nil config", discover: config.Discover},
		{name: "empty root", cfg: &config.Config{}, discover: config.Discover},
		{name: "nil discovery", cfg: &config.Config{Root: t.TempDir()}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, err := NewPlanningStore(tc.cfg, tc.discover, nil)
			if fs != nil || !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("incomplete construction = %v, %v", fs, err)
			}
		})
	}
}

func TestNewPlanningStoreRechecksLegacyRootWithoutInventingIdentity(t *testing.T) {
	root := t.TempDir()
	if _, err := config.Init(root, "", false); err != nil {
		t.Fatal(err)
	}
	taskID := testutil.TaskID("legacy-reader-member")
	testutil.Write(t, filepath.Join(root, domain.TasksDir, taskID+"-legacy-reader-member.md"),
		"---\nid: "+taskID+"\nstatus: ready-to-start\ndescription: legacy fixture\ntags: [test]\n---\n")
	cfg, err := config.Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := NewPlanningStore(cfg, config.Discover, func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	plan, err := core.MustNewService(initial).ComposeThreadApply(cfg.ID, core.ThreadComposeManifest{
		Thread: core.ThreadComposeInput{Title: "Legacy", Description: "Legacy fixture", Goal: "No invented identity", Tags: []string{"test"}},
		Nodes:  []core.ThreadComposeNode{{Key: "member", TaskID: taskID}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, config.ConfigFile)); err != nil {
		t.Fatal(err)
	}
	cfg, err = config.Discover(root)
	if err != nil || cfg.Dir != "" || cfg.ID != "" {
		t.Fatalf("legacy fixture after marker loss = %+v, %v", cfg, err)
	}
	var starts []string
	fs, err := NewPlanningStore(cfg, func(start string) (*config.Config, error) {
		starts = append(starts, start)
		return config.Discover(start)
	}, func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	service := core.MustNewService(fs)
	if records, problems, err := service.ListTasks(core.TaskFilter{}); err != nil || len(records) != 1 || len(problems) != 0 || len(starts) != 0 {
		t.Fatalf("legacy reads=%v problems=%v err=%v starts=%v", records, problems, err, starts)
	}
	// The durable plan from before marker loss must not adopt a cached/invented ID.
	receipt, err := service.ApplyThreadPlan(plan, true)
	if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "no durable id") || receipt.Committed || !slices.Equal(starts, []string{cfg.Root}) {
		t.Fatalf("legacy apply receipt=%+v err=%v starts=%v", receipt, err, starts)
	}
}

func TestNewPlanningStoreIdentityReaderFailuresNeverUseCachedIdentity(t *testing.T) {
	sentinel := errors.New("identity discovery failed")
	for _, tc := range []struct {
		name    string
		fresh   *config.Config
		cause   error
		want    error
		message string
		partial bool
	}{
		{name: "nil result", want: domain.ErrValidation, message: "returned no root"},
		{name: "empty root", fresh: &config.Config{ID: "cached-id"}, want: domain.ErrValidation, message: "returned no root"},
		{name: "reader error", cause: sentinel, want: sentinel, message: "identity discovery failed"},
		{name: "partial result with reader error", cause: sentinel, want: sentinel, message: "identity discovery failed", partial: true},
	} {
		for _, dryRun := range []bool{true, false} {
			name := tc.name + "/commit"
			if dryRun {
				name = tc.name + "/dry-run"
			}
			t.Run(name, func(t *testing.T) {
				cfg, manifest := planningStoreFixture(t)
				broken := false
				fs, err := NewPlanningStore(cfg, func(start string) (*config.Config, error) {
					if broken {
						if tc.partial {
							return cfg, tc.cause // valid cached-looking data must not mask the error
						}
						return tc.fresh, tc.cause
					}
					return config.Discover(start)
				}, func() error { return nil })
				if err != nil {
					t.Fatal(err)
				}
				service := core.MustNewService(fs)
				plan, err := service.ComposeThreadApply(cfg.ID, manifest)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := service.ApplyThreadPlan(plan, true); err != nil {
					t.Fatalf("fixture must be valid before re-reader failure: %v", err)
				}
				before := planningStoreDocuments(t, cfg.Root)
				broken = true
				receipt, err := service.ApplyThreadPlan(plan, dryRun)
				if !errors.Is(err, tc.want) || !strings.Contains(err.Error(), tc.message) || receipt.Committed || receipt.Plan.Thread.ID != plan.Thread.ID {
					t.Fatalf("failed identity read receipt=%+v err=%v", receipt, err)
				}
				if !maps.Equal(before, planningStoreDocuments(t, cfg.Root)) {
					t.Fatal("failed identity re-read changed planning documents")
				}
			})
		}
	}
}

func planningStoreFixture(t *testing.T) (*config.Config, core.ThreadComposeManifest) {
	t.Helper()
	root := t.TempDir()
	if _, err := config.Init(root, "planning", false); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	taskID := testutil.TaskID("identity-reader-member")
	testutil.Write(t, filepath.Join(cfg.Root, domain.TasksDir, taskID+"-identity-reader-member.md"),
		"---\nid: "+taskID+"\nstatus: ready-to-start\ndescription: identity reader fixture\ntags: [test]\n---\n")
	return cfg, core.ThreadComposeManifest{
		Thread: core.ThreadComposeInput{Title: "Identity reader", Description: "Identity reader fixture", Goal: "Guarded identity", Tags: []string{"test"}},
		Nodes:  []core.ThreadComposeNode{{Key: "member", TaskID: taskID}},
	}
}

func planningStoreDocuments(t *testing.T, root string) map[string]string {
	t.Helper()
	files := make(map[string]string)
	for _, dir := range []string{domain.TasksDir, domain.ThreadsDir} {
		paths, err := filepath.Glob(filepath.Join(root, dir, "*"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			files[path] = string(content)
		}
	}
	return files
}
