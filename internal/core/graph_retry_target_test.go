package core

import (
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// Conflicts before planning establish nothing; an empty result after a prior
// selection must not erase it. Snapshots deliberately reassign every old alias.
type retryEndpointStore struct {
	nopStore
	calls         int
	graphPlans    []TaskGraphMutationPlan
	threadPlans   []ThreadMutationPlan
	creationPlans []ThreadCreationPlan
}

func (s *retryEndpointStore) graph() *TaskGraph {
	dep, pre := "dependent", "prerequisite"
	if s.calls > 2 {
		dep, pre = "renamed-dependent", "renamed-prerequisite"
	}
	tasks := []domain.Task{
		{ID: testutil.TaskID("dependent"), Slug: dep, Status: domain.StatusReadyToStart},
		{ID: testutil.TaskID("prerequisite"), Slug: pre, Status: domain.StatusCompleted},
	}
	if s.calls > 2 {
		tasks = append(tasks, domain.Task{ID: testutil.TaskID("impostor-dependent"), Slug: "dependent", Status: domain.StatusReadyToStart}, domain.Task{ID: testutil.TaskID("impostor-prerequisite"), Slug: "prerequisite", Status: domain.StatusCompleted})
	}
	return NewTaskGraph(tasks, nil)
}

func (s *retryEndpointStore) threads() []domain.Thread {
	name := "thread"
	if s.calls > 2 {
		name = "renamed"
	}
	return []domain.Thread{
		{ID: testutil.TaskID("thread"), Slug: name, Status: domain.ThreadStatusUnstarted},
		{ID: testutil.TaskID("impostor-thread"), Slug: "thread-two", Status: domain.ThreadStatusUnstarted},
	}
}

func (s *retryEndpointStore) MutateTaskGraph(_ time.Time, _ bool, planner TaskGraphPlanner) (TaskGraphMutationResult, error) {
	s.calls++
	if s.calls == 1 || s.calls == 3 {
		return TaskGraphMutationResult{}, domain.ErrConflict
	}
	plan, err := planner(s.graph())
	if err != nil {
		return TaskGraphMutationResult{}, err
	}
	s.graphPlans = append(s.graphPlans, plan)
	result := TaskGraphMutationResult{Plan: plan}
	if s.calls == 2 {
		return result, domain.ErrConflict
	}
	return result, nil
}

func (s *retryEndpointStore) MutateThread(_ time.Time, _ bool, planner ThreadMutationPlanner) (ThreadMutationResult, error) {
	s.calls++
	if s.calls == 1 || s.calls == 3 {
		return ThreadMutationResult{}, domain.ErrConflict
	}
	plan, err := planner(ThreadMutationSnapshot{Graph: s.graph(), Threads: s.threads()})
	if err != nil {
		return ThreadMutationResult{}, err
	}
	s.threadPlans = append(s.threadPlans, plan)
	result := ThreadMutationResult{Plan: plan}
	if s.calls == 2 {
		// A conflict after selection but before validation may have no result plan.
		return ThreadMutationResult{}, domain.ErrConflict
	}
	return result, nil
}

func (s *retryEndpointStore) MutateThreadCreation(_ time.Time, _ bool, planner ThreadCreationPlanner) (ThreadCreationMutationResult, error) {
	s.calls++
	if s.calls == 1 || s.calls == 3 {
		return ThreadCreationMutationResult{}, domain.ErrConflict
	}
	plan, err := planner(ThreadCreationSnapshot{Graph: s.graph(), Threads: s.threads()})
	if err != nil {
		return ThreadCreationMutationResult{}, err
	}
	s.creationPlans = append(s.creationPlans, plan)
	result := ThreadCreationMutationResult{Plan: plan}
	if s.calls == 2 {
		// Creation-ID collision validation currently returns exactly this shape.
		return ThreadCreationMutationResult{}, domain.ErrConflict
	}
	return result, nil
}

func TestGraphRetriesRetainBindingAcrossUnplannedConflicts(t *testing.T) {
	for _, operation := range []string{"dependency", "membership", "start", "creation"} {
		t.Run(operation, func(t *testing.T) {
			st := &retryEndpointStore{}
			svc := MustNewService(st, WithRetry(3, func(int) {}))
			var err error
			switch operation {
			case "dependency":
				_, err = svc.AddTaskDependencies("dependent", []string{"prerequisite"}, false)
			case "membership":
				_, err = svc.AddThreadMembers("thread", []string{"prerequisite"}, false)
			case "start":
				_, err = svc.StartThread("thread", false)
			case "creation":
				_, err = svc.NewThread(NewThreadParams{Title: "new", Description: "creation", Goal: "fixed members", Tasks: []string{"prerequisite"}})
			}
			if err != nil || st.calls != 4 {
				t.Fatalf("retry lifecycle: calls=%d err=%v", st.calls, err)
			}
			for _, plan := range st.graphPlans {
				if len(plan.TaskWrites) != 1 || plan.TaskWrites[0].TaskID != testutil.TaskID("dependent") || len(plan.TaskWrites[0].DependsOn) != 1 || plan.TaskWrites[0].DependsOn[0] != testutil.TaskID("prerequisite") {
					t.Fatalf("dependency binding drift: %+v", plan)
				}
			}
			for _, plan := range st.threadPlans {
				if plan.ThreadID != testutil.TaskID("thread") || (operation == "membership" && (len(plan.TaskIDs) != 1 || plan.TaskIDs[0] != testutil.TaskID("prerequisite"))) {
					t.Fatalf("Thread binding drift: %+v", plan)
				}
			}
			for _, plan := range st.creationPlans {
				if len(plan.Thread.Tasks) != 1 || plan.Thread.Tasks[0] != testutil.TaskID("prerequisite") {
					t.Fatalf("creation binding drift: %+v", plan)
				}
			}
		})
	}
}
