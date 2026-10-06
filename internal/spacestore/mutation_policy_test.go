package spacestore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

func TestRegistryPoliciesDenyBeforeEffectsIncludingPreview(t *testing.T) {
	denied := errors.New("registry refused")
	for _, tc := range []struct {
		name   string
		policy core.MutationPolicy
		want   error
	}{
		{"guarded", core.GuardedMutations(func() error { return denied }), denied},
		{"read-only", core.ReadOnlyMutations(), core.ErrReadOnlyPersistence},
		{"zero adapter", core.MutationPolicy{}, core.ErrInvalidMutationPolicy},
	} {
		for _, dryRun := range []bool{true, false} {
			t.Run(tc.name+map[bool]string{true: "/preview", false: "/commit"}[dryRun], func(t *testing.T) {
				base := t.TempDir()
				t.Setenv(userconfig.DirEnv, filepath.Join(base, "missing-home"))
				adapter := &FS{}
				if tc.name != "zero adapter" {
					adapter = testutil.Must(New(tc.policy))
				}
				if _, _, err := adapter.AddSpace(core.SpaceRegistration{ID: "probe", Checkout: "/missing"}, dryRun); !errors.Is(err, tc.want) {
					t.Fatalf("add error=%v; want %v", err, tc.want)
				}
				if _, _, err := adapter.ForgetSpace("probe", dryRun); !errors.Is(err, tc.want) {
					t.Fatalf("forget error=%v; want %v", err, tc.want)
				}
				if entries, err := os.ReadDir(base); err != nil || len(entries) != 0 {
					t.Fatalf("refusal touched storage: entries=%v err=%v", entries, err)
				}
			})
		}
	}
}

func TestSummaryPlanningStoresCannotInheritRegistryMutationPrivilege(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Task("ready-to-start", "probe.md", "---\ndescription: probe\ntags: [test]\n---\n")
	for _, policy := range []core.MutationPolicy{core.UnrestrictedMutations(), core.ReadOnlyMutations(), core.GuardedMutations(func() error {
		t.Fatal("summary opener invoked registry authorizer")
		return nil
	})} {
		adapter := testutil.Must(New(policy))
		summary, err := adapter.OpenPlanningStore(r.Root)
		if err != nil {
			t.Fatal(err)
		}
		fs, ok := summary.(*store.FS)
		if !ok {
			t.Fatalf("unexpected summary adapter %T", summary)
		}
		if tasks, _, err := fs.ListTasks(); err != nil || len(tasks) != 1 {
			t.Fatalf("read-only summary tasks=%v err=%v", tasks, err)
		}
		for _, dryRun := range []bool{true, false} {
			if _, err := fs.SetFields("probe", map[string]any{"priority": "low"}, dryRun); !errors.Is(err, core.ErrReadOnlyPersistence) {
				t.Fatalf("widened summary capability authorized mutation: %v", err)
			}
		}
		if task, _, err := fs.GetTask("probe"); err != nil || task.Status != domain.StatusReadyToStart || task.Priority == "low" {
			t.Fatalf("summary mutation changed task: task=%v err=%v", task, err)
		}
	}
	if source, err := (&FS{}).OpenPlanningStore(r.Root); source != nil || !errors.Is(err, core.ErrInvalidMutationPolicy) {
		t.Fatalf("zero registry minted summary store: %v %v", source, err)
	}
}
