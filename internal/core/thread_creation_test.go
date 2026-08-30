package core

import (
	"errors"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestValidateThreadCreationSourceRejectsCrossKindAndMissingMember(t *testing.T) {
	task := graphRecord("member", domain.StatusReadyToStart)
	graph := NewTaskGraph([]domain.Task{task}, nil)
	existing := threadRecord(domain.ThreadStatusUnstarted, task.ID)
	existing.ID, existing.FilenameID = task.ID, task.ID
	if err := ValidateThreadCreationSource(graph, []domain.Thread{existing}, nil); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("cross-kind error = %v", err)
	}

	existing.ID, existing.FilenameID = testutil.TaskID("existing-thread"), testutil.TaskID("existing-thread")
	existing.Tasks = []string{testutil.TaskID("missing")}
	if err := ValidateThreadCreationSource(graph, []domain.Thread{existing}, nil); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("missing-member error = %v", err)
	}
}

func TestValidateThreadCreationPlanAllowsEmptyAndRejectsLifecycleOrCollision(t *testing.T) {
	task := graphRecord("member", domain.StatusReadyToStart)
	graph := NewTaskGraph([]domain.Task{task}, nil)
	thread := threadRecord(domain.ThreadStatusUnstarted)
	thread.Slug = "new-thread"
	snapshot := ThreadCreationSnapshot{Graph: graph}
	if _, err := ValidateThreadCreationPlan(snapshot, ThreadCreationPlan{Thread: thread}); err != nil {
		t.Fatalf("empty unstarted Thread: %v", err)
	}

	thread.Status = domain.ThreadStatusInProgress
	if _, err := ValidateThreadCreationPlan(snapshot, ThreadCreationPlan{Thread: thread}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("lifecycle error = %v", err)
	}
	thread.Status = domain.ThreadStatusUnstarted
	thread.ID, thread.FilenameID = task.ID, task.ID
	if _, err := ValidateThreadCreationPlan(snapshot, ThreadCreationPlan{Thread: thread}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("collision error = %v", err)
	}
}
