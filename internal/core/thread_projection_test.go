package core

import (
	"slices"
	"sort"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func threadRecord(status domain.ThreadStatus, taskIDs ...string) domain.Thread {
	sort.Strings(taskIDs)
	return domain.Thread{
		ID: testutil.TaskID("thread"), FilenameID: testutil.TaskID("thread"), Slug: "thread",
		Path: "threads/" + testutil.TaskID("thread") + "-thread.md", Status: status,
		Description: "Thread projection", Goal: "Prove projection semantics", Created: "2026-08-29",
		Tasks: append([]string(nil), taskIDs...),
	}
}

func TestProjectThreadRollupExternalGatesAndFrontier(t *testing.T) {
	internalDone := graphRecord("internal-done", domain.StatusCompleted)
	external := graphRecord("external", domain.StatusNextUp)
	candidate := graphRecord("candidate", domain.StatusReadyToStart, internalDone.ID, external.ID)
	parked := graphRecord("parked-member", domain.StatusDeferred)
	withdrawn := graphRecord("withdrawn-member", domain.StatusDeprecated)
	thread := threadRecord(domain.ThreadStatusInProgress, internalDone.ID, candidate.ID, parked.ID, withdrawn.ID)

	view := ProjectThread(thread, NewTaskGraph([]domain.Task{candidate, withdrawn, external, internalDone, parked}, nil))
	if view.GraphHealth != GraphHealthy || view.ProjectionHealth != GraphHealthy {
		t.Fatalf("graph=%s projection=%s problems=%+v", view.GraphHealth, view.ProjectionHealth, view.Problems)
	}
	if view.Rollup != (ThreadRollup{Done: 1, Total: 3, Drained: 1, Deprecated: 1}) {
		t.Fatalf("rollup = %+v", view.Rollup)
	}
	if len(view.ExternalGates) != 1 || view.ExternalGates[0].Task.ID != external.ID || !view.ExternalGates[0].Outstanding {
		t.Fatalf("external gates = %+v", view.ExternalGates)
	}
	if len(view.Frontier) != 0 {
		t.Fatalf("frontier = %+v; externally blocked candidate must not dispatch", view.Frontier)
	}
}

func TestProjectThreadFrontierUsesSharedEligibility(t *testing.T) {
	done := graphRecord("done", domain.StatusCompleted)
	ready := graphRecord("ready", domain.StatusReadyToStart, done.ID)
	queued := graphRecord("queued", domain.StatusNextUp)
	thread := threadRecord(domain.ThreadStatusUnstarted, done.ID, ready.ID, queued.ID)
	view := ProjectThread(thread, NewTaskGraph([]domain.Task{queued, ready, done}, nil))
	if len(view.Frontier) != 2 || !view.Frontier[0].State.Eligible || !view.Frontier[1].State.Eligible {
		t.Fatalf("frontier = %+v", view.Frontier)
	}
	got := []string{view.Frontier[0].Task.ID, view.Frontier[1].Task.ID}
	sort.Strings(got)
	want := []string{queued.ID, ready.ID}
	sort.Strings(want)
	if !slices.Equal(got, want) {
		t.Fatalf("frontier IDs = %v, want %v", got, want)
	}
}

func TestProjectThreadBrokenProjectionSuppressesOtherwiseEligibleFrontier(t *testing.T) {
	queued := graphRecord("queued-with-missing-thread-member", domain.StatusNextUp)
	missing := testutil.TaskID("missing-thread-member")
	view := ProjectThread(
		threadRecord(domain.ThreadStatusUnstarted, queued.ID, missing),
		NewTaskGraph([]domain.Task{queued}, nil),
	)

	if view.GraphHealth != GraphHealthy || view.ProjectionHealth != GraphBroken {
		t.Fatalf("graph=%s projection=%s problems=%+v", view.GraphHealth, view.ProjectionHealth, view.Problems)
	}
	if len(view.Members) != 2 || !view.Members[0].State.Eligible && !view.Members[1].State.Eligible {
		t.Fatalf("members = %+v; expected the present queued member to remain graph-eligible", view.Members)
	}
	if len(view.Frontier) != 0 {
		t.Fatalf("frontier = %+v; broken Thread-local evidence must fail closed", view.Frontier)
	}
}

func TestProjectThreadIdentityDriftAndTaskCollisionFailClosed(t *testing.T) {
	task := graphRecord("colliding-identity", domain.StatusNextUp)
	thread := threadRecord(domain.ThreadStatusUnstarted, task.ID)
	thread.ID = task.ID
	thread.FilenameID = testutil.TaskID("different-thread-filename")

	view := ProjectThread(thread, NewTaskGraph([]domain.Task{task}, nil))
	if view.ProjectionHealth != GraphBroken || len(view.Frontier) != 0 {
		t.Fatalf("identity-defective projection = %+v", view)
	}
	codes := make(map[ThreadProblemCode]bool)
	for _, problem := range view.Problems {
		codes[problem.Code] = true
	}
	if !codes[ThreadProblemIDDrift] || !codes[ThreadProblemTaskIDCollision] {
		t.Fatalf("identity problems = %+v", view.Problems)
	}
}

func TestProjectThreadMissingAndInvalidIDsDoNotBecomeTaskCollisions(t *testing.T) {
	task := graphRecord("unrelated-task", domain.StatusNextUp)
	for _, threadID := range []string{"", "not-a-stable-id"} {
		thread := threadRecord(domain.ThreadStatusUnstarted, task.ID)
		thread.ID, thread.FilenameID = threadID, ""
		view := ProjectThread(thread, NewTaskGraph([]domain.Task{task}, nil))
		if view.ProjectionHealth != GraphBroken || len(view.Frontier) != 0 {
			t.Fatalf("invalid identity projection = %+v", view)
		}
		if len(view.Problems) != 1 || view.Problems[0].Code != ThreadProblemInvalidDocument {
			t.Fatalf("invalid identity problems = %+v", view.Problems)
		}
	}
}

func TestProjectThreadCompletedInconsistencyAndMissingMember(t *testing.T) {
	done := graphRecord("done-member", domain.StatusCompleted)
	missing := testutil.TaskID("missing-member")
	thread := threadRecord(domain.ThreadStatusCompleted, done.ID, missing)
	view := ProjectThread(thread, NewTaskGraph([]domain.Task{done}, nil))
	if view.GraphHealth != GraphHealthy || view.ProjectionHealth != GraphBroken || !view.Inconsistent {
		t.Fatalf("view = graph %s projection %s inconsistent=%v", view.GraphHealth, view.ProjectionHealth, view.Inconsistent)
	}
	if view.Rollup.Total != 2 || view.Rollup.Done != 1 || view.Rollup.Drained != 1 {
		t.Fatalf("missing member must remain in denominator: %+v", view.Rollup)
	}
	if len(view.Problems) != 3 || view.Problems[0].Code != ThreadProblemMissingMember ||
		view.Problems[1].Code != ThreadProblemCompletedUnhealthyEvidence ||
		view.Problems[2].Code != ThreadProblemCompletedUndrained {
		t.Fatalf("problems = %+v", view.Problems)
	}
	if len(view.Frontier) != 0 {
		t.Fatal("broken Thread projection exposed dispatchable work")
	}
}

func TestProjectThreadCompletedRequiresSoundExternalGate(t *testing.T) {
	upstream := graphRecord("upstream-open", domain.StatusNextUp)
	externalDone := graphRecord("external-done", domain.StatusCompleted, upstream.ID)
	memberDone := graphRecord("member-done", domain.StatusCompleted, externalDone.ID)
	thread := threadRecord(domain.ThreadStatusCompleted, memberDone.ID)
	view := ProjectThread(thread, NewTaskGraph([]domain.Task{memberDone, externalDone, upstream}, nil))
	if view.Rollup.Done != 1 || view.Rollup.Drained != 0 || !view.Inconsistent {
		t.Fatalf("view = rollup %+v inconsistent=%v", view.Rollup, view.Inconsistent)
	}
	if len(view.ExternalGates) != 1 || !view.ExternalGates[0].Outstanding {
		t.Fatalf("external gates = %+v", view.ExternalGates)
	}
}

func TestProjectThreadEmptyCompletedIsInconsistent(t *testing.T) {
	view := ProjectThread(threadRecord(domain.ThreadStatusCompleted), NewTaskGraph(nil, nil))
	if !view.Inconsistent || view.Rollup.Total != 0 || len(view.Problems) != 1 || view.Problems[0].Code != ThreadProblemCompletedEmpty {
		t.Fatalf("empty completed view = %+v", view)
	}
}

func TestProjectThreadDeprecatedMemberDoesNotContributeExternalGate(t *testing.T) {
	external := graphRecord("withdrawn-external", domain.StatusReadyToStart)
	withdrawn := graphRecord("withdrawn", domain.StatusDeprecated, external.ID)
	done := graphRecord("live-done", domain.StatusCompleted)
	thread := threadRecord(domain.ThreadStatusCompleted, withdrawn.ID, done.ID)

	view := ProjectThread(thread, NewTaskGraph([]domain.Task{withdrawn, external, done}, nil))
	if view.Inconsistent || len(view.ExternalGates) != 0 {
		t.Fatalf("withdrawn member gated closure: %+v", view)
	}
	if view.Rollup != (ThreadRollup{Done: 1, Total: 1, Drained: 1, Deprecated: 1}) {
		t.Fatalf("rollup = %+v", view.Rollup)
	}
}

func TestProjectThreadCompletedInconsistencyNamesOutstandingGate(t *testing.T) {
	external := graphRecord("external-open", domain.StatusReadyToStart)
	member := graphRecord("member-complete", domain.StatusCompleted, external.ID)
	view := ProjectThread(threadRecord(domain.ThreadStatusCompleted, member.ID), NewTaskGraph([]domain.Task{member, external}, nil))

	want := map[ThreadProblemCode]bool{
		ThreadProblemCompletedUndrained:    false,
		ThreadProblemCompletedExternalGate: false,
	}
	for _, problem := range view.Problems {
		if _, exists := want[problem.Code]; exists {
			want[problem.Code] = true
		}
	}
	for code, found := range want {
		if !found {
			t.Errorf("missing %s in %+v", code, view.Problems)
		}
	}
}
