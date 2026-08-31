package core

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

type threadApplyFakeOutcome struct {
	result ThreadApplyMutationResult
	err    error
}

type threadApplyFake struct {
	snapshot ThreadApplySnapshot
	outcomes []threadApplyFakeOutcome
	calls    int
}

func (f *threadApplyFake) MutateThreadApply(_ time.Time, dryRun bool, planner ThreadApplyPlanner) (ThreadApplyMutationResult, error) {
	f.calls++
	plan, err := planner(f.snapshot)
	if err != nil {
		return ThreadApplyMutationResult{DryRun: dryRun}, err
	}
	index := f.calls - 1
	if index >= len(f.outcomes) {
		index = len(f.outcomes) - 1
	}
	outcome := f.outcomes[index]
	outcome.result.Plan = plan
	outcome.result.DryRun = dryRun
	return outcome.result, outcome.err
}

func TestServiceComposeThreadApplyRendersDefaultTemplate(t *testing.T) {
	task := graphRecord("compose-member", domain.StatusNextUp)
	threadID := testutil.TaskID("composed-thread")
	svc := NewService(nil,
		WithTaskGraphSource(&taskGraphReadFake{tasks: []domain.Task{task}}),
		WithThreadStore(&threadReadFake{}),
		WithIDGen(func() string { return threadID }),
		WithClock(func() time.Time { return threadApplyNow }),
	)
	plan, err := svc.ComposeThreadApply("planning", ThreadComposeManifest{
		Thread: ThreadComposeInput{
			Title: "Composed delivery", Description: "Compile one plan", Goal: "Exercise the production template",
		},
		Nodes: []ThreadComposeNode{{Key: "member", TaskID: task.ID}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Thread.ID != threadID || !strings.Contains(plan.Thread.Body, "# Thread: Composed delivery") || !strings.Contains(plan.Thread.Body, "Exercise the production template") {
		t.Fatalf("plan = %+v", plan)
	}
}

func TestServiceComposeThreadApplyReportsEachMissingReadCapability(t *testing.T) {
	tests := []struct {
		name string
		svc  *Service
		want string
	}{
		{name: "task graph", svc: NewService(nil, WithThreadStore(&threadReadFake{})), want: "task graph reads are unavailable"},
		{name: "Threads", svc: NewService(nil, WithTaskGraphSource(&taskGraphReadFake{})), want: "thread reads are unavailable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.svc.ComposeThreadApply("planning", ThreadComposeManifest{}); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("compose error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestServiceComposeThreadApplyReadsThreadsBeforeTasks(t *testing.T) {
	calls := make([]string, 0, 2)
	svc := NewService(nil,
		WithTaskGraphSource(&taskGraphReadFake{onList: func() { calls = append(calls, "tasks") }}),
		WithThreadStore(&threadReadFake{onList: func() { calls = append(calls, "threads") }}),
	)
	_, _ = svc.ComposeThreadApply("planning", ThreadComposeManifest{})
	if !slices.Equal(calls, []string{"threads", "tasks"}) {
		t.Fatalf("calls = %v", calls)
	}
}

func TestServiceThreadApplyRetriesOnlyPreCommitConflict(t *testing.T) {
	task := graphRecord("apply-member", domain.StatusNextUp)
	plan := ThreadApplyPlan{
		Schema: ThreadApplyPlanSchema, PlanningRepoID: "planning", ComposedAt: "2026-08-30",
		Thread: ThreadApplyThread{
			ID: testutil.TaskID("apply-service-thread"), Slug: "apply-service", Status: domain.ThreadStatusUnstarted,
			Description: "Apply service", Goal: "Pin retry policy", Created: "2026-08-30",
			Tasks: []string{task.ID}, Body: "# Apply\n",
		},
	}
	fake := &threadApplyFake{
		snapshot: ThreadApplySnapshot{PlanningRepoID: "planning", Graph: NewTaskGraph([]domain.Task{task}, nil)},
		outcomes: []threadApplyFakeOutcome{
			{err: domain.ErrConflict},
			{result: ThreadApplyMutationResult{
				Changed: true, Complete: true, Committed: true,
				Operations: []ThreadApplyOperation{{Kind: "thread", Action: "create", State: ThreadApplyApplied, ThreadID: plan.Thread.ID}},
			}},
		},
	}
	svc := NewService(nil, WithThreadApplyMutationStore(fake), WithRetry(3, func(int) {}))
	receipt, err := svc.ApplyThreadPlan(plan, false)
	if err != nil || fake.calls != 2 || !receipt.Complete || !receipt.Committed {
		t.Fatalf("receipt=%+v err=%v calls=%d", receipt, err, fake.calls)
	}
}

func TestServiceThreadApplyDoesNotRetryDurableOrCompleteFailure(t *testing.T) {
	task := graphRecord("apply-failure-member", domain.StatusNextUp)
	plan := ThreadApplyPlan{
		Schema: ThreadApplyPlanSchema, PlanningRepoID: "planning", ComposedAt: "2026-08-30",
		Thread: ThreadApplyThread{
			ID: testutil.TaskID("apply-failure-thread"), Slug: "apply-failure", Status: domain.ThreadStatusUnstarted,
			Description: "Apply failure", Goal: "Preserve recovery", Created: "2026-08-30",
			Tasks: []string{task.ID}, Body: "# Apply\n",
		},
	}
	for _, tc := range []struct {
		name      string
		committed bool
		complete  bool
	}{
		{name: "durable prefix", committed: true},
		{name: "final cleanup", committed: true, complete: true},
		{name: "complete no-op cleanup", complete: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &threadApplyFake{
				snapshot: ThreadApplySnapshot{PlanningRepoID: "planning", Graph: NewTaskGraph([]domain.Task{task}, nil)},
				outcomes: []threadApplyFakeOutcome{{
					result: ThreadApplyMutationResult{Changed: true, Committed: tc.committed, Complete: tc.complete},
					err:    domain.ErrConflict,
				}},
			}
			svc := NewService(nil, WithThreadApplyMutationStore(fake), WithRetry(3, func(int) {}))
			receipt, err := svc.ApplyThreadPlan(plan, false)
			var failure *ThreadApplyFailure
			if !errors.As(err, &failure) || fake.calls != 1 || receipt.Committed != tc.committed || receipt.Complete != tc.complete {
				t.Fatalf("receipt=%+v err=%v calls=%d", receipt, err, fake.calls)
			}
		})
	}
}

func TestServiceThreadApplyRetainsPlanIdentityOnPreplanFailure(t *testing.T) {
	task := graphRecord("apply-identity-member", domain.StatusNextUp)
	plan := ThreadApplyPlan{
		Schema: ThreadApplyPlanSchema, PlanningRepoID: "wrong-repository", ComposedAt: "2026-08-30",
		Thread: ThreadApplyThread{
			ID: testutil.TaskID("apply-identity-thread"), Slug: "apply-identity", Status: domain.ThreadStatusUnstarted,
			Description: "Apply identity", Goal: "Retain recovery identity", Created: "2026-08-30",
			Tasks: []string{task.ID}, Body: "# Apply\n",
		},
	}
	fake := &threadApplyFake{
		snapshot: ThreadApplySnapshot{PlanningRepoID: "planning", Graph: NewTaskGraph([]domain.Task{task}, nil)},
		outcomes: []threadApplyFakeOutcome{{}},
	}
	svc := NewService(nil, WithThreadApplyMutationStore(fake), WithRetry(0, func(int) {}))
	receipt, err := svc.ApplyThreadPlan(plan, false)
	var failure *ThreadApplyFailure
	if !errors.As(err, &failure) || receipt.Plan.Thread.ID != plan.Thread.ID || failure.Receipt.Plan.Thread.ID != plan.Thread.ID {
		t.Fatalf("receipt=%+v err=%v failure=%+v", receipt, err, failure)
	}
}
