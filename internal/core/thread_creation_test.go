package core

import (
	"errors"
	"strings"
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

func TestValidateThreadCreationSourceOrdersPortableDiagnosticsAndNamesRemoteThreads(t *testing.T) {
	graph := NewTaskGraph(nil, nil)
	unreadable := []ThreadReadProblem{
		{ThreadID: "6g0000000002", ThreadSlug: "later", Message: "later failure"},
		{ThreadID: "6g0000000001", ThreadSlug: "first", Message: "first failure"},
	}
	if err := ValidateThreadCreationSource(graph, nil, unreadable); err == nil ||
		!strings.Contains(err.Error(), "first (6g0000000001)") || strings.Contains(err.Error(), "later failure") {
		t.Fatalf("ordered unreadable error = %v", err)
	}

	duplicateID := testutil.TaskID("portable-duplicate")
	threads := []domain.Thread{
		{ID: duplicateID, Slug: "z-remote", Status: domain.ThreadStatusUnstarted, Description: "Remote duplicate", Goal: "Fail closed", Created: "2026-09-01"},
		{ID: duplicateID, Slug: "a-remote", Status: domain.ThreadStatusUnstarted, Description: "Remote duplicate", Goal: "Fail closed", Created: "2026-09-01"},
	}
	if err := ValidateThreadCreationSource(graph, threads, nil); err == nil ||
		!strings.Contains(err.Error(), "a-remote ("+duplicateID+")") ||
		!strings.Contains(err.Error(), "z-remote ("+duplicateID+")") {
		t.Fatalf("portable duplicate error = %v", err)
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
