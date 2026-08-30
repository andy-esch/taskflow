package core

import (
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

var threadMutationNow = time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)

func mutationSnapshot(thread domain.Thread, tasks ...domain.Task) ThreadMutationSnapshot {
	return ThreadMutationSnapshot{Graph: NewTaskGraph(tasks, nil), Threads: []domain.Thread{thread}}
}

func TestThreadMutationSnapshotResolvesThreadInsideGuardedSnapshot(t *testing.T) {
	alpha := threadRecord(domain.ThreadStatusUnstarted)
	alpha.ID, alpha.Slug = testutil.TaskID("thread-alpha"), "alpha-delivery"
	beta := threadRecord(domain.ThreadStatusUnstarted)
	beta.ID, beta.Slug = testutil.TaskID("thread-beta"), "beta-delivery"
	snapshot := ThreadMutationSnapshot{Threads: []domain.Thread{beta, alpha}}

	for _, ref := range []string{alpha.ID, "alpha", "pha-del"} {
		if got, err := snapshot.ResolveThreadID(ref); err != nil || got != alpha.ID {
			t.Fatalf("ResolveThreadID(%q) = %q, %v", ref, got, err)
		}
	}
	if _, err := snapshot.ResolveThreadID("delivery"); !errors.Is(err, domain.ErrAmbiguous) {
		t.Fatalf("ambiguous resolution error = %v", err)
	}
}

func TestValidateThreadMembershipMutationIsAtomicSortedAndIdempotent(t *testing.T) {
	a := graphRecord("member-a", domain.StatusNextUp)
	b := graphRecord("member-b", domain.StatusReadyToStart)
	c := graphRecord("member-c", domain.StatusCompleted)
	thread := threadRecord(domain.ThreadStatusUnstarted, a.ID)
	snapshot := mutationSnapshot(thread, a, b, c)

	plan, analysis, err := ValidateThreadMutationPlan(snapshot, ThreadMutationPlan{
		ThreadID: thread.ID, Operation: ThreadMutationAddMembers, TaskIDs: []string{c.ID, a.ID, b.ID},
	}, threadMutationNow)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan.TaskIDs, []string{a.ID, b.ID, c.ID}) {
		t.Fatalf("normalized task IDs = %v", plan.TaskIDs)
	}
	if !reflect.DeepEqual(analysis.After.Thread.Tasks, []string{a.ID, b.ID, c.ID}) || !analysis.Changed {
		t.Fatalf("after = %+v", analysis.After.Thread)
	}
	wantOutcomes := []ThreadMemberOutcome{
		{TaskID: a.ID, Action: "add", Outcome: "skipped"},
		{TaskID: b.ID, Action: "add", Outcome: "added"},
		{TaskID: c.ID, Action: "add", Outcome: "added"},
	}
	if !reflect.DeepEqual(analysis.MemberOutcomes, wantOutcomes) {
		t.Fatalf("outcomes = %+v, want %+v", analysis.MemberOutcomes, wantOutcomes)
	}

	_, removal, err := ValidateThreadMutationPlan(snapshot, ThreadMutationPlan{
		ThreadID: thread.ID, Operation: ThreadMutationRemoveMembers, TaskIDs: []string{b.ID, a.ID},
	}, threadMutationNow)
	if err != nil {
		t.Fatal(err)
	}
	if len(removal.After.Thread.Tasks) != 0 || !removal.Changed || removal.MemberOutcomes[1].Outcome != "skipped" {
		t.Fatalf("removal = %+v", removal)
	}

	if _, _, err := ValidateThreadMutationPlan(snapshot, ThreadMutationPlan{
		ThreadID: thread.ID, Operation: ThreadMutationAddMembers, TaskIDs: []string{b.ID, b.ID},
	}, threadMutationNow); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("duplicate request error = %v", err)
	}
	if _, _, err := ValidateThreadMutationPlan(snapshot, ThreadMutationPlan{
		ThreadID: thread.ID, Operation: ThreadMutationAddMembers, TaskIDs: []string{testutil.TaskID("missing")},
	}, threadMutationNow); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("missing request error = %v", err)
	}
}

func TestValidateThreadMembershipRejectsTerminalThreadsEvenForNoOps(t *testing.T) {
	task := graphRecord("terminal-member", domain.StatusCompleted)
	for _, status := range []domain.ThreadStatus{domain.ThreadStatusCompleted, domain.ThreadStatusCancelled} {
		thread := threadRecord(status, task.ID)
		_, _, err := ValidateThreadMutationPlan(mutationSnapshot(thread, task), ThreadMutationPlan{
			ThreadID: thread.ID, Operation: ThreadMutationAddMembers, TaskIDs: []string{task.ID},
		}, threadMutationNow)
		var policy *ThreadMutationPolicyError
		if !errors.As(err, &policy) || policy.Remedy == "" {
			t.Fatalf("status %s error = %v", status, err)
		}
	}
}

func TestValidateThreadLifecycleContractsAndNoOps(t *testing.T) {
	done := graphRecord("lifecycle-done", domain.StatusCompleted)
	thread := threadRecord(domain.ThreadStatusUnstarted, done.ID)
	snapshot := mutationSnapshot(thread, done)

	_, started, err := ValidateThreadMutationPlan(snapshot, ThreadMutationPlan{ThreadID: thread.ID, Operation: ThreadMutationStart}, threadMutationNow)
	if err != nil {
		t.Fatal(err)
	}
	if started.After.Thread.Status != domain.ThreadStatusInProgress || started.After.Thread.StartedAt != "2026-08-30" || !started.Changed {
		t.Fatalf("started = %+v", started.After.Thread)
	}

	inProgress := started.After.Thread
	_, completed, err := ValidateThreadMutationPlan(mutationSnapshot(inProgress, done), ThreadMutationPlan{ThreadID: thread.ID, Operation: ThreadMutationComplete}, threadMutationNow)
	if err != nil {
		t.Fatal(err)
	}
	if completed.After.Thread.Status != domain.ThreadStatusCompleted || completed.After.Thread.EndedAt != "2026-08-30" {
		t.Fatalf("completed = %+v", completed.After.Thread)
	}

	_, same, err := ValidateThreadMutationPlan(
		mutationSnapshot(completed.After.Thread, done),
		ThreadMutationPlan{ThreadID: thread.ID, Operation: ThreadMutationComplete}, threadMutationNow,
	)
	if err != nil || same.Changed {
		t.Fatalf("completed no-op = %+v, %v", same, err)
	}

	_, reopened, err := ValidateThreadMutationPlan(
		mutationSnapshot(completed.After.Thread, done),
		ThreadMutationPlan{ThreadID: thread.ID, Operation: ThreadMutationReopen}, threadMutationNow,
	)
	if err != nil || reopened.After.Thread.Status != domain.ThreadStatusInProgress || reopened.After.Thread.EndedAt != "" {
		t.Fatalf("reopened = %+v, %v", reopened.After.Thread, err)
	}

	_, cancelled, err := ValidateThreadMutationPlan(snapshot, ThreadMutationPlan{ThreadID: thread.ID, Operation: ThreadMutationCancel}, threadMutationNow)
	if err != nil || cancelled.After.Thread.Status != domain.ThreadStatusCancelled || cancelled.After.Thread.StartedAt != "" {
		t.Fatalf("cancelled unstarted = %+v, %v", cancelled.After.Thread, err)
	}
	if _, _, err := ValidateThreadMutationPlan(
		mutationSnapshot(cancelled.After.Thread, done),
		ThreadMutationPlan{ThreadID: thread.ID, Operation: ThreadMutationReopen}, threadMutationNow,
	); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("reopen cancelled error = %v", err)
	}
	if _, _, err := ValidateThreadMutationPlan(
		mutationSnapshot(completed.After.Thread, done),
		ThreadMutationPlan{ThreadID: thread.ID, Operation: ThreadMutationCancel}, threadMutationNow,
	); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("cancel completed error = %v", err)
	}
}

func TestValidateThreadCompleteRequiresLiveSoundlyDrainedMembers(t *testing.T) {
	deferred := graphRecord("deferred-member", domain.StatusDeferred)
	deprecated := graphRecord("deprecated-member", domain.StatusDeprecated)
	for _, tc := range []struct {
		name   string
		tasks  []domain.Task
		member []string
	}{
		{name: "empty"},
		{name: "all deprecated", tasks: []domain.Task{deprecated}, member: []string{deprecated.ID}},
		{name: "deferred is live", tasks: []domain.Task{deferred}, member: []string{deferred.ID}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			thread := threadRecord(domain.ThreadStatusInProgress, tc.member...)
			_, _, err := ValidateThreadMutationPlan(mutationSnapshot(thread, tc.tasks...), ThreadMutationPlan{
				ThreadID: thread.ID, Operation: ThreadMutationComplete,
			}, threadMutationNow)
			var policy *ThreadMutationPolicyError
			if !errors.As(err, &policy) || policy.Remedy == "" {
				t.Fatalf("completion error = %v", err)
			}
		})
	}
}

func TestValidateThreadCompleteIgnoresBrokenEvidenceFromDeprecatedMembers(t *testing.T) {
	live := graphRecord("live-completed-member", domain.StatusCompleted)
	deprecatedPrerequisite := graphRecord("deprecated-prerequisite", domain.StatusDeprecated)
	deprecatedMember := graphRecord("deprecated-broken-member", domain.StatusDeprecated, deprecatedPrerequisite.ID)
	thread := threadRecord(domain.ThreadStatusInProgress, live.ID, deprecatedMember.ID)

	_, analysis, err := ValidateThreadMutationPlan(
		mutationSnapshot(thread, live, deprecatedMember, deprecatedPrerequisite),
		ThreadMutationPlan{ThreadID: thread.ID, Operation: ThreadMutationComplete},
		threadMutationNow,
	)
	if err != nil {
		t.Fatalf("deprecated member blocked completion: %v", err)
	}
	if analysis.Before.Rollup.Total != 1 || analysis.Before.Rollup.Drained != 1 || analysis.Before.Rollup.Deprecated != 1 {
		t.Fatalf("before rollup = %+v", analysis.Before.Rollup)
	}
	if analysis.After.Thread.Status != domain.ThreadStatusCompleted {
		t.Fatalf("completed Thread = %+v", analysis.After.Thread)
	}
}

func TestValidateThreadCompletionDefensivelyRejectsContradictoryLiveEvidence(t *testing.T) {
	thread := threadRecord(domain.ThreadStatusInProgress)
	base := ThreadView{
		ProjectionHealth: GraphHealthy,
		Rollup:           ThreadRollup{Total: 1, Drained: 1},
	}
	tests := []struct {
		name string
		view ThreadView
	}{
		{
			name: "outstanding external gate",
			view: func() ThreadView {
				view := base
				view.ExternalGates = []ThreadExternalGate{{Outstanding: true}}
				return view
			}(),
		},
		{
			name: "broken live member",
			view: func() ThreadView {
				view := base
				view.Members = []ThreadTaskView{{State: TaskGraphState{Role: RoleNominallyComplete, Gate: GateBroken}}}
				return view
			}(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var policy *ThreadMutationPolicyError
			if err := validateThreadCompletion(thread, tc.view); !errors.As(err, &policy) {
				t.Fatalf("completion error = %v", err)
			}
		})
	}
}

func TestValidateThreadLifecycleNoOpsRevalidateDestinationPolicy(t *testing.T) {
	deprecated := graphRecord("noop-deprecated-member", domain.StatusDeprecated)
	for _, tc := range []struct {
		name   string
		thread domain.Thread
		tasks  []domain.Task
	}{
		{name: "empty", thread: threadRecord(domain.ThreadStatusInProgress)},
		{name: "all withdrawn", thread: threadRecord(domain.ThreadStatusInProgress, deprecated.ID), tasks: []domain.Task{deprecated}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := ValidateThreadMutationPlan(mutationSnapshot(tc.thread, tc.tasks...), ThreadMutationPlan{
				ThreadID: tc.thread.ID, Operation: ThreadMutationStart,
			}, threadMutationNow); !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("same-state start accepted invalid membership: %v", err)
			}
		})
	}

	unfinished := graphRecord("noop-unfinished-member", domain.StatusNextUp)
	completed := threadRecord(domain.ThreadStatusCompleted, unfinished.ID)
	if _, _, err := ValidateThreadMutationPlan(mutationSnapshot(completed, unfinished), ThreadMutationPlan{
		ThreadID: completed.ID, Operation: ThreadMutationComplete,
	}, threadMutationNow); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("same-state complete accepted drifted projection: %v", err)
	}
}

func TestTaskLifecycleThreadImpactsIncludeDirectDownstreamAndExternalGateEffects(t *testing.T) {
	upstream := graphRecord("impact-upstream", domain.StatusCompleted)
	downstream := graphRecord("impact-downstream", domain.StatusCompleted, upstream.ID)
	unrelated := graphRecord("impact-unrelated", domain.StatusCompleted)
	graph := NewTaskGraph([]domain.Task{upstream, downstream, unrelated}, nil)

	direct := threadRecord(domain.ThreadStatusInProgress, upstream.ID)
	direct.ID, direct.Slug = testutil.TaskID("direct-thread"), "direct"
	shared := threadRecord(domain.ThreadStatusUnstarted, upstream.ID)
	shared.ID, shared.Slug = testutil.TaskID("shared-thread"), "shared"
	dependent := threadRecord(domain.ThreadStatusCompleted, downstream.ID)
	dependent.ID, dependent.Slug = testutil.TaskID("dependent-thread"), "dependent"
	other := threadRecord(domain.ThreadStatusCompleted, unrelated.ID)
	other.ID, other.Slug = testutil.TaskID("other-thread"), "other"

	impacts := TaskLifecycleThreadImpacts([]domain.Thread{other, dependent, shared, direct}, graph, TaskLifecyclePlan{
		TaskID: upstream.ID, To: domain.StatusInProgress,
	})
	if len(impacts) != 3 {
		t.Fatalf("impacts = %+v", impacts)
	}
	byID := map[string]ThreadProjectionImpact{}
	for _, impact := range impacts {
		byID[impact.ThreadID] = impact
	}
	if !byID[direct.ID].Direct || !byID[shared.ID].Direct || byID[dependent.ID].Direct || !byID[dependent.ID].After.Inconsistent {
		t.Fatalf("impact attribution = %+v", impacts)
	}
	wantChanged := []string{upstream.ID, downstream.ID}
	sort.Strings(wantChanged)
	if !reflect.DeepEqual(byID[dependent.ID].ChangedTaskIDs, wantChanged) {
		t.Fatalf("changed tasks = %v", byID[dependent.ID].ChangedTaskIDs)
	}
}
