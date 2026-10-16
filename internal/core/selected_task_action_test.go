package core

import (
	"errors"
	"maps"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestSelectedTaskFieldsSkipResolutionAndShareRetryAndCoercion(t *testing.T) {
	for _, dry := range []bool{false, true} {
		st := &selectedTaskStore{sourceID: "must-not-resolve"}
		clocks, sleeps := 0, 0
		svc := MustNewService(st, WithClock(func() time.Time {
			clocks++
			return time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
		}), WithRetry(2, func(int) { sleeps++ }))
		updates := map[string]any{"tier": "2", "tags": "go,core", "effort": domain.UnsetField{}}
		original := maps.Clone(updates)
		target := TaskMutationTarget{ID: "urn:task:opaque"}
		_, err := svc.SetSelectedTaskFields(target, updates, false, dry)
		attempts := 3
		if dry {
			attempts = 1
			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("preview retried or swallowed conflict: %v", err)
			}
		} else if err != nil {
			t.Fatal(err)
		}
		if st.reads != 0 || st.typedReads != 0 || clocks != 1 || len(st.targets) != attempts || sleeps != attempts-1 {
			t.Fatalf("selected identity was re-resolved or retry/clock drifted: store=%+v clocks=%d sleeps=%d", st, clocks, sleeps)
		}
		for i, got := range st.targets {
			if got != target || st.updates[i]["tier"] != 2 || !reflect.DeepEqual(st.updates[i]["tags"], []string{"go", "core"}) || st.updates[i]["updated_at"] != "2026-10-15" {
				t.Fatalf("attempt changed target, coerced values or clock: target=%+v updates=%+v", got, st.updates[i])
			}
			if _, ok := st.updates[i]["effort"].(domain.UnsetField); !ok {
				t.Fatal("unset was not preserved")
			}
		}
		if !reflect.DeepEqual(updates, original) {
			t.Fatalf("caller updates were mutated: %+v", updates)
		}
	}
}

func TestSelectedTaskFieldsShareReferenceValidation(t *testing.T) {
	for _, updates := range []map[string]any{
		{}, {"priority": "urgent"}, {"tier": "bad"}, {"status": "completed"},
		{"depends_on": []string{"other"}}, {"unknown": "value"},
		{"updated_at": domain.UnsetField{}}, {"epic": "missing"},
	} {
		for _, dry := range []bool{false, true} {
			st := &selectedTaskStore{}
			svc := MustNewService(st)
			_, refErr := svc.SetFields("human-ref", updates, false, dry)
			_, selectedErr := svc.SetSelectedTaskFields(TaskMutationTarget{ID: "opaque-id"}, updates, false, dry)
			if !errors.Is(refErr, domain.ErrValidation) || !errors.Is(selectedErr, domain.ErrValidation) || refErr.Error() != selectedErr.Error() || len(st.targets) != 0 || st.reads != 0 {
				t.Fatalf("validation diverged: updates=%+v ref=%v selected=%v store=%+v", updates, refErr, selectedErr, st)
			}
		}
	}
}

type selectedLifecycleActionStore struct {
	nopStore
	plans []TaskLifecyclePlan
	times []time.Time
}

func (st *selectedLifecycleActionStore) MutateTaskLifecycle(now time.Time, dry bool, planner TaskLifecyclePlanner) (TaskLifecycleMutationResult, error) {
	// The intended task is absent. A different task's human slug is its ID;
	// reference-based planning would silently select the impostor instead.
	graph := NewTaskGraph([]domain.Task{{ID: testutil.TaskID("impostor"), Slug: testutil.TaskID("selected"), Status: domain.StatusReadyToStart, Description: "impostor", Tags: []string{"go"}}}, nil)
	plan, err := planner(graph)
	if err != nil {
		return TaskLifecycleMutationResult{}, err
	}
	st.plans = append(st.plans, plan)
	st.times = append(st.times, now)
	result := TaskLifecycleMutationResult{Plan: plan, DryRun: dry}
	if len(st.plans) < 3 {
		return result, domain.ErrConflict
	}
	_, _, err = ValidateTaskLifecyclePlan(graph, plan, "")
	return result, err
}

func TestSelectedTaskLifecycleDoesNotResolveBeforeOrDuringRetry(t *testing.T) {
	for _, verb := range []string{"move", "defer"} {
		for _, dry := range []bool{false, true} {
			st := &selectedLifecycleActionStore{}
			clocks, sleeps := 0, 0
			svc := MustNewService(st, WithClock(func() time.Time {
				clocks++
				return time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
			}), WithRetry(2, func(int) { sleeps++ }))
			target := TaskMutationTarget{ID: testutil.TaskID("selected")}
			var receipt TaskLifecycleReceipt
			var err error
			if verb == "move" {
				receipt, err = svc.MoveSelectedTask(target, domain.StatusInProgress, dry, TaskLifecycleOverrideDependencyGate)
			} else {
				receipt, err = svc.DeferSelectedTask(target, "2099-01-01", dry)
			}
			want, attempts := domain.ErrNotFound, 3
			if dry {
				want, attempts = domain.ErrConflict, 1
			}
			if !errors.Is(err, want) || receipt.Committed || len(st.plans) != attempts || clocks != 1 || sleeps != attempts-1 {
				t.Fatalf("selected lifecycle drifted: verb=%s dry=%v err=%v receipt=%+v plans=%+v clocks=%d sleeps=%d", verb, dry, err, receipt, st.plans, clocks, sleeps)
			}
			for i, plan := range st.plans {
				if plan.TaskID != target.ID || plan.Create != nil || st.times[i] != st.times[0] {
					t.Fatalf("retry changed intent or time: plan=%+v time=%v", plan, st.times[i])
				}
				if verb == "move" && (plan.To != domain.StatusInProgress || plan.Override != TaskLifecycleOverrideDependencyGate) {
					t.Fatalf("move override changed: %+v", plan)
				}
				if verb == "defer" && (plan.To != domain.StatusDeferred || plan.RevisitAt != "2099-01-01") {
					t.Fatalf("defer date changed: %+v", plan)
				}
			}
		}
	}
}

func TestSelectedTaskActionsRefuseEmptyIdentityBeforeAdapterCalls(t *testing.T) {
	for _, id := range []string{"", "  "} {
		for _, dry := range []bool{false, true} {
			st := &selectedTaskStore{}
			lifecycle := &selectedLifecycleActionStore{}
			for _, mutate := range []func() error{
				func() error {
					_, err := MustNewService(st).SetSelectedTaskFields(TaskMutationTarget{ID: id}, map[string]any{"priority": "high"}, false, dry)
					return err
				},
				func() error {
					_, err := MustNewService(lifecycle).MoveSelectedTask(TaskMutationTarget{ID: id}, domain.StatusInProgress, dry, TaskLifecycleOverrideNone)
					return err
				},
				func() error {
					_, err := MustNewService(lifecycle).DeferSelectedTask(TaskMutationTarget{ID: id}, "", dry)
					return err
				},
			} {
				if err := mutate(); !errors.Is(err, domain.ErrValidation) {
					t.Fatalf("missing identity should fail closed: %v", err)
				}
			}
			if st.reads != 0 || len(st.targets) != 0 || len(lifecycle.plans) != 0 {
				t.Fatal("empty identity reached an adapter")
			}
		}
	}
	lifecycle := &selectedLifecycleActionStore{}
	_, err := MustNewService(lifecycle).DeferSelectedTask(TaskMutationTarget{ID: testutil.TaskID("selected")}, "tomorrow", false)
	if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "YYYY-MM-DD") || len(lifecycle.plans) != 0 {
		t.Fatalf("invalid date reached the adapter: err=%v plans=%v", err, lifecycle.plans)
	}
}
