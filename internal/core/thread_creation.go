package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// ThreadCreationSnapshot is the immutable semantic input to a guarded creation
// planner. Source versions are stripped from Threads; persistence tokens never
// cross the store-owned critical-section boundary.
type ThreadCreationSnapshot struct {
	Graph   *TaskGraph
	Threads []domain.Thread
}

type ThreadCreationPlan struct {
	Thread domain.Thread
	Body   string
}

type ThreadCreationPlanner func(ThreadCreationSnapshot) (ThreadCreationPlan, error)

type ThreadCreationMutationResult struct {
	Plan      ThreadCreationPlan
	Thread    domain.Thread
	Changed   bool
	DryRun    bool
	Committed bool
}

type ThreadCreationReceipt struct {
	Thread    domain.Thread
	Changed   bool
	DryRun    bool
	Committed bool
}

type ThreadCreationMutationFailure struct {
	Cause   error
	Receipt ThreadCreationReceipt
}

func (e *ThreadCreationMutationFailure) Error() string {
	if e == nil {
		return "thread creation committed, but repository cleanup failed"
	}
	return fmt.Sprintf("thread creation committed, but repository cleanup failed: %v; inspect the current Thread before retrying", e.Cause)
}

func (e *ThreadCreationMutationFailure) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// ValidateThreadCreationSource proves the guarded snapshot is authoritative
// enough to introduce a new cross-linked document.
func ValidateThreadCreationSource(graph *TaskGraph, threads []domain.Thread, unreadable []ThreadReadProblem) error {
	if err := ValidateTaskLifecycleSource(graph); err != nil {
		return err
	}
	if len(unreadable) > 0 {
		orderedProblems := append([]ThreadReadProblem(nil), unreadable...)
		sort.Slice(orderedProblems, func(i, j int) bool { return threadReadProblemLess(orderedProblems[i], orderedProblems[j]) })
		problem := orderedProblems[0]
		return fmt.Errorf("%w: current Thread record %s is unreadable: %s",
			domain.ErrValidation, threadReadProblemName(problem), problem.Message)
	}
	seen := make(map[string]string, len(threads))
	ordered := cloneThreads(threads)
	sort.Slice(ordered, func(i, j int) bool { return threadLess(ordered[i], ordered[j]) })
	for _, thread := range ordered {
		name := threadDiagnosticName(thread)
		if err := domain.ValidateThreadDocument(thread); err != nil {
			return fmt.Errorf("%w: existing Thread %s is invalid: %v", domain.ErrValidation, name, err)
		}
		if thread.FilenameID != "" && thread.FilenameID != thread.ID {
			return fmt.Errorf("%w: existing Thread %s has id drift: frontmatter=%q filename=%q", domain.ErrValidation, name, thread.ID, thread.FilenameID)
		}
		if prior, exists := seen[thread.ID]; exists {
			return fmt.Errorf("%w: duplicate Thread id %s across %s and %s", domain.ErrValidation, thread.ID, prior, name)
		}
		seen[thread.ID] = name
		if _, collision := graph.Task(thread.ID); collision {
			return fmt.Errorf("%w: stable id %s is used by both a task and a Thread", domain.ErrValidation, thread.ID)
		}
		for _, taskID := range thread.Tasks {
			if _, exists := graph.Task(taskID); !exists {
				return fmt.Errorf("%w: existing Thread %s references missing task %s", domain.ErrValidation, thread.ID, taskID)
			}
		}
	}
	return nil
}

// ValidateThreadCreationPlan normalizes the planner's membership set and
// enforces the deliberately narrow unstarted-create contract.
func ValidateThreadCreationPlan(snapshot ThreadCreationSnapshot, plan ThreadCreationPlan) (ThreadCreationPlan, error) {
	plan.Thread = cloneThread(plan.Thread)
	thread := plan.Thread
	if thread.Status != domain.ThreadStatusUnstarted {
		return ThreadCreationPlan{}, fmt.Errorf("%w: Thread creation always starts unstarted, got %q", domain.ErrValidation, thread.Status)
	}
	if thread.StartedAt != "" || thread.EndedAt != "" || thread.Updated != "" {
		return ThreadCreationPlan{}, fmt.Errorf("%w: Thread creation cannot set lifecycle or update timestamps", domain.ErrValidation)
	}
	if strings.TrimSpace(thread.Slug) == "" {
		return ThreadCreationPlan{}, fmt.Errorf("%w: Thread creation requires a non-empty slug", domain.ErrValidation)
	}
	if err := domain.ValidateThreadDocument(thread); err != nil {
		return ThreadCreationPlan{}, err
	}
	if snapshot.Graph == nil {
		return ThreadCreationPlan{}, fmt.Errorf("%w: authoritative task graph is required", domain.ErrValidation)
	}
	if _, exists := snapshot.Graph.Task(thread.ID); exists {
		return ThreadCreationPlan{}, fmt.Errorf("stable id %s is already used by a task: %w", thread.ID, domain.ErrConflict)
	}
	for _, existing := range snapshot.Threads {
		if existing.ID == thread.ID {
			return ThreadCreationPlan{}, fmt.Errorf("thread id %s is already used by %s: %w", thread.ID, existing.Path, domain.ErrConflict)
		}
	}
	for _, taskID := range thread.Tasks {
		if _, exists := snapshot.Graph.Task(taskID); !exists {
			return ThreadCreationPlan{}, fmt.Errorf("%w: Thread member task %s does not exist", domain.ErrValidation, taskID)
		}
	}
	return plan, nil
}

func cloneThreads(threads []domain.Thread) []domain.Thread {
	out := make([]domain.Thread, len(threads))
	for i, thread := range threads {
		out[i] = cloneThread(thread)
		out[i].SourceVersion = ""
	}
	return out
}
