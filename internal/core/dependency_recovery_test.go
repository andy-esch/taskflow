package core

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestTaskGraphStateImpactNewlyUnsafe(t *testing.T) {
	for _, tc := range []struct {
		name                                        string
		beforeGate, afterGate                       GateState
		beforeInconsistent, afterInconsistent, want bool
	}{
		{name: "unchanged clear", beforeGate: GateClear, afterGate: GateClear},
		{name: "already blocked", beforeGate: GateBlocked, afterGate: GateBlocked},
		{name: "already broken", beforeGate: GateBroken, afterGate: GateBroken},
		{name: "newly blocked", beforeGate: GateClear, afterGate: GateBlocked, want: true},
		{name: "newly broken", beforeGate: GateClear, afterGate: GateBroken, want: true},
		{name: "blocked to broken", beforeGate: GateBlocked, afterGate: GateBroken, want: true},
		{name: "broken to blocked is still a changed non-clear gate", beforeGate: GateBroken, afterGate: GateBlocked, want: true},
		{name: "cleared", beforeGate: GateBlocked, afterGate: GateClear},
		{name: "newly inconsistent", beforeGate: GateClear, afterGate: GateClear, afterInconsistent: true, want: true},
		{name: "already inconsistent", beforeGate: GateClear, afterGate: GateClear, beforeInconsistent: true, afterInconsistent: true},
		{name: "consistency restored", beforeGate: GateClear, afterGate: GateClear, beforeInconsistent: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			impact := TaskGraphStateImpact{
				Before: TaskGraphState{Role: RoleQueued, Gate: tc.beforeGate, Inconsistent: tc.beforeInconsistent},
				After:  TaskGraphState{Role: RoleCandidate, Gate: tc.afterGate, Inconsistent: tc.afterInconsistent},
			}
			if got := impact.NewlyUnsafe(); got != tc.want {
				t.Fatalf("NewlyUnsafe() = %t, want %t: %+v", got, tc.want, impact)
			}
			if got := taskImpactsNeedRepair([]TaskGraphStateImpact{impact}); got != tc.want {
				t.Fatalf("aggregate decision = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestThreadProjectionImpactNewlyInconsistent(t *testing.T) {
	for _, before := range []bool{false, true} {
		for _, after := range []bool{false, true} {
			impact := ThreadProjectionImpact{Before: ThreadView{Inconsistent: before}, After: ThreadView{Inconsistent: after}}
			want := !before && after
			if impact.NewlyInconsistent() != want || threadImpactsNeedRepair([]ThreadProjectionImpact{impact}) != want {
				t.Fatalf("Thread impact warning differs: %+v", impact)
			}
		}
	}
}

func TestServiceDependencyAndLifecycleShareRecoveryIntent(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		for _, prerequisiteStatus := range []domain.Status{domain.StatusCompleted, domain.StatusNextUp} {
			t.Run(fmt.Sprintf("%s/preview=%t", prerequisiteStatus, dryRun), func(t *testing.T) {
				prerequisite := graphRecord("prerequisite", prerequisiteStatus)
				dependent := graphRecord("dependent", domain.StatusReadyToStart)
				store := &graphOperationStore{fakeStore: fakeStore{tasks: []domain.Task{dependent, prerequisite}}}
				receipt, err := MustNewService(store).AddTaskDependencies(dependent.ID, []string{prerequisite.ID}, dryRun)
				if err != nil || len(receipt.Impacts) != 1 {
					t.Fatalf("dependency receipt = %+v, %v", receipt, err)
				}
				unsafe := prerequisiteStatus != domain.StatusCompleted
				if receipt.Impacts[0].NewlyUnsafe() != unsafe || (receipt.Remedy != "") != unsafe {
					t.Fatalf("impact/recovery mismatch: %+v", receipt)
				}
				lifecycle := taskLifecycleReceipt(TaskLifecycleMutationResult{
					Impacts: receipt.Impacts, DryRun: dryRun, Committed: !dryRun,
				})
				if lifecycle.Remedy != receipt.Remedy {
					t.Fatalf("lifecycle remedy %q differs from dependency %q", lifecycle.Remedy, receipt.Remedy)
				}
				if unsafe && strings.Contains(receipt.Remedy, "preview only") != dryRun {
					t.Fatalf("preview recovery claims durable state: %+v", receipt)
				}
				if dryRun && len(store.tasks[0].DependsOn) != 0 {
					t.Fatal("preview persisted a dependency")
				}
			})
		}
	}
}

func TestServiceDependencySafeAlreadyUnsafeAndClearedNeedNoNewRecovery(t *testing.T) {
	for _, dependentStatus := range []domain.Status{domain.StatusReadyToStart, domain.StatusInProgress, domain.StatusCompleted} {
		prerequisite := graphRecord("pending", domain.StatusNextUp)
		other := graphRecord("other-pending", domain.StatusNextUp)
		dependent := graphRecord("already-blocked", dependentStatus, prerequisite.ID)
		store := &graphOperationStore{fakeStore: fakeStore{tasks: []domain.Task{dependent, prerequisite, other}}}
		svc := MustNewService(store)
		receipt, err := svc.AddTaskDependencies(dependent.ID, []string{other.ID}, false)
		if err != nil || len(receipt.Impacts) != 1 || receipt.Impacts[0].NewlyUnsafe() || receipt.Remedy != "" {
			t.Fatalf("already unsafe should not get newly-unsafe advice: %+v, %v", receipt, err)
		}
		noop, err := svc.AddTaskDependencies(dependent.ID, []string{other.ID}, false)
		if err != nil || noop.Changed || len(noop.Impacts) != 0 || noop.Remedy != "" {
			t.Fatalf("no-op recovery = %+v, %v", noop, err)
		}
		cleared, err := svc.RemoveTaskDependencies(dependent.ID, []string{prerequisite.ID, other.ID}, false)
		if err != nil || len(cleared.Impacts) != 1 || cleared.Impacts[0].NewlyUnsafe() || cleared.Remedy != "" || cleared.Impacts[0].After.Gate != GateClear {
			t.Fatalf("cleared recovery = %+v, %v", cleared, err)
		}
	}
}

func TestServiceDependencyRefusalAndCommittedFailureRecovery(t *testing.T) {
	for _, after := range []int{0, 1} {
		prerequisite := graphRecord("pending", domain.StatusNextUp)
		dependent := graphRecord("dependent", domain.StatusReadyToStart)
		store := &graphOperationStore{
			fakeStore: fakeStore{tasks: []domain.Task{dependent, prerequisite}},
			failures:  []graphMutationFailure{{err: fmt.Errorf("guard failure: %w", domain.ErrConflict), after: after}},
		}
		retries := 0
		if after > 0 {
			retries = 4 // A committed conflict must remain visible, not become a no-op retry.
		}
		receipt, err := MustNewService(store, WithRetry(retries, func(int) {})).AddTaskDependencies(dependent.ID, []string{prerequisite.ID}, false)
		var failure *DependencyMutationFailure
		if !errors.As(err, &failure) || !errors.Is(err, domain.ErrConflict) || store.calls != 1 || !reflect.DeepEqual(receipt, failure.Receipt) {
			t.Fatalf("typed failure/receipt lost: %+v, %v", receipt, err)
		}
		if len(receipt.AppliedTaskIDs) != after || len(receipt.RemainingTaskIDs) != 1-after || len(receipt.Impacts) != 1 || !receipt.Impacts[0].NewlyUnsafe() {
			t.Fatalf("prospective impact/durability evidence lost: %+v", receipt)
		}
		want := "no dependency task files were applied"
		if after > 0 {
			want = "writes already landed"
		}
		if !strings.Contains(receipt.Remedy, want) || !strings.Contains(err.Error(), receipt.Remedy) || strings.Contains(receipt.Remedy, "restore sound prerequisites") {
			t.Fatalf("failure confused prospective/committed state: %+v, %v", receipt, err)
		}
	}
}

func TestServiceDependencyPartialRecoveryRequiresInspectionAndNeverAutoRetries(t *testing.T) {
	owner := graphRecord("legacy-owner", domain.StatusCompleted)
	dependent := graphRecord("legacy-dependent", domain.StatusReadyToStart)
	owner.LegacyBlocks = []string{dependent.Slug}
	dependent.LegacyBlockedBy = []string{owner.Slug}
	store := &graphOperationStore{
		fakeStore: fakeStore{tasks: []domain.Task{owner, dependent}},
		failures:  []graphMutationFailure{{err: domain.ErrConflict, after: 1}},
	}
	receipt, err := MustNewService(store, WithRetry(4, func(int) {})).MigrateTaskDependencies(false)
	if !errors.Is(err, domain.ErrConflict) || store.calls != 1 || len(receipt.AppliedTaskIDs) != 1 || len(receipt.RemainingTaskIDs) != 1 {
		t.Fatalf("partial durability/retry = %+v, calls=%d, %v", receipt, store.calls, err)
	}
	if !strings.Contains(receipt.Remedy, "before resuming") || !strings.Contains(err.Error(), receipt.Remedy) || strings.Contains(err.Error(), "retry the same command to converge") {
		t.Fatalf("partial recovery needs reconciliation, not blind retry: %+v, %v", receipt, err)
	}
}

func TestDependencyPreviewRefusalDoesNotClaimAppliedChanges(t *testing.T) {
	task := graphRecord("self", domain.StatusReadyToStart)
	store := &graphOperationStore{fakeStore: fakeStore{tasks: []domain.Task{task}}}
	receipt, err := MustNewService(store).AddTaskDependencies(task.ID, []string{task.ID}, true)
	if !errors.Is(err, domain.ErrValidation) || !receipt.DryRun || len(receipt.AppliedTaskIDs) != 0 || len(receipt.Impacts) != 0 || len(store.tasks[0].DependsOn) != 0 || !strings.Contains(receipt.Remedy, "no dependency task files were applied") {
		t.Fatalf("preview refusal claims a write: %+v, %v", receipt, err)
	}
}

func TestLifecycleThreadRecoveryQualifiesPreviewAndRetainsOverride(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		receipt := taskLifecycleReceipt(TaskLifecycleMutationResult{
			DryRun: dryRun, OverrideApplied: true,
			Plan:          TaskLifecyclePlan{Override: TaskLifecycleOverrideDependencyGate},
			ThreadImpacts: []ThreadProjectionImpact{{After: ThreadView{Inconsistent: true}}},
		})
		if !strings.Contains(receipt.Remedy, "override did not alter dependency edges") ||
			!strings.Contains(receipt.Remedy, "restore sound member or external-gate evidence") ||
			strings.Contains(receipt.Remedy, "preview only") != dryRun {
			t.Fatalf("Thread/override recovery lost context: %+v", receipt)
		}
	}
}
