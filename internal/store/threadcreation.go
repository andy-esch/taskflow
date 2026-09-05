package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

var _ core.ThreadCreationMutationStore = (*FS)(nil)

func (s *FS) MutateThreadCreation(now time.Time, dryRun bool, planner core.ThreadCreationPlanner) (result core.ThreadCreationMutationResult, err error) {
	result.DryRun = dryRun
	if planner == nil {
		return result, fmt.Errorf("%w: Thread creation planner is required", domain.ErrValidation)
	}
	if now.IsZero() {
		return result, fmt.Errorf("%w: Thread creation time is required", domain.ErrValidation)
	}
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return result, err
	}
	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return result, fmt.Errorf("mkdir planning root %s: %w", s.root, err)
	}
	unlock, err := s.checkedWriteLock()
	if err != nil {
		return result, err
	}
	defer func() {
		if releaseErr := unlock(); releaseErr != nil {
			wrapped := fmt.Errorf("release repository Thread creation guard: %w", releaseErr)
			if err == nil {
				err = wrapped
			} else {
				err = errors.Join(err, wrapped)
			}
		}
	}()

	graph, err := core.LoadTaskGraph(s)
	if err != nil {
		return result, fmt.Errorf("load authoritative task graph: %w", err)
	}
	threadRead, err := s.ReadThreads()
	if err != nil {
		return result, fmt.Errorf("load authoritative Threads: %w", err)
	}
	if err := core.ValidateThreadCreationSource(graph, threadRead.Threads, threadRead.Problems); err != nil {
		return result, err
	}
	snapshot := core.ThreadCreationSnapshot{Graph: graph, Threads: clonePlannerThreads(threadRead.Threads)}
	plan, err := callThreadCreationPlanner(s, planner, snapshot)
	if err != nil {
		return result, err
	}
	result.Plan, err = core.ValidateThreadCreationPlan(snapshot, plan)
	if err != nil {
		return result, err
	}
	materialized, err := s.materializeThreadCreation(result.Plan)
	if err != nil {
		return result, err
	}
	result.Thread = materialized.thread
	result.Changed = true
	if dryRun {
		return result, nil
	}

	if testHookBeforeThreadCreationVerify != nil {
		testHookBeforeThreadCreationVerify()
	}
	currentGraph, err := core.LoadTaskGraph(s)
	if err != nil {
		return result, fmt.Errorf("re-read authoritative task graph before Thread create: %w", err)
	}
	currentThreadRead, err := s.ReadThreads()
	if err != nil {
		return result, fmt.Errorf("re-read authoritative Threads before create: %w", err)
	}
	if graphErr := verifyTaskGraphSourceSnapshot(graph, currentGraph); graphErr != nil || verifyThreadSourceSnapshot(threadRead, currentThreadRead) != nil {
		return result, fmt.Errorf("repository tasks or Threads changed while planning; retry: %w", domain.ErrConflict)
	}
	if err := s.writeNewFileUnlocked(s.threadsDir, materialized.path, materialized.content, "Thread", materialized.thread.ID+"-"+materialized.thread.Slug); err != nil {
		return result, err
	}
	result.Committed = true
	return result, nil
}

func callThreadCreationPlanner(store *FS, planner core.ThreadCreationPlanner, snapshot core.ThreadCreationSnapshot) (core.ThreadCreationPlan, error) {
	leave, err := store.enterRepositoryPlanner()
	if err != nil {
		return core.ThreadCreationPlan{}, err
	}
	defer leave()
	return planner(snapshot)
}

type materializedThreadCreation struct {
	thread  domain.Thread
	path    string
	content []byte
}

// materializeThreadCreation is intentionally lock-free and creation-only. A
// compound bulk capability may reuse it for its final new-Thread write while one
// outer repository guard is held. Mutations of an existing Thread must instead
// use surgical frontmatter updates so lifecycle timestamps, unknown fields,
// comments, and key order are preserved.
func (s *FS) materializeThreadCreation(plan core.ThreadCreationPlan) (materializedThreadCreation, error) {
	thread := plan.Thread
	if err := validEntityID(thread.ID); err != nil {
		return materializedThreadCreation{}, err
	}
	stem := thread.ID + "-" + thread.Slug
	path := filepath.Join(s.threadsDir, stem+".md")
	if _, err := os.Stat(path); err == nil {
		return materializedThreadCreation{}, fmt.Errorf("thread %q already exists: %w", stem, domain.ErrConflict)
	} else if !os.IsNotExist(err) {
		return materializedThreadCreation{}, fmt.Errorf("stat Thread %s: %w", path, err)
	}
	content, err := buildFile(threadCreationFields(thread), plan.Body)
	if err != nil {
		return materializedThreadCreation{}, err
	}
	parsed, err := parseThread(content, path)
	if err != nil {
		return materializedThreadCreation{}, fmt.Errorf("%w: Thread creation would not reload: %v", domain.ErrValidation, err)
	}
	if err := domain.ValidateThreadDocument(parsed); err != nil {
		return materializedThreadCreation{}, err
	}
	return materializedThreadCreation{thread: parsed, path: path, content: content}, nil
}

// threadCreationFields emits exactly the state legal on a new unstarted Thread.
// ValidateThreadCreationPlan rejects update/lifecycle timestamps before this
// builder is reached; this is not a general Thread serializer.
func threadCreationFields(thread domain.Thread) []fmField {
	tasks := append([]string(nil), thread.Tasks...)
	sort.Strings(tasks)
	fields := []fmField{
		{"schema", domain.FileSchemaVersion},
		{"id", thread.ID},
		{"status", string(thread.Status)},
		{"description", thread.Description},
		{"goal", thread.Goal},
	}
	if thread.TargetDate != "" {
		fields = append(fields, fmField{"target_date", thread.TargetDate})
	}
	fields = append(fields, fmField{"created", thread.Created})
	if len(thread.Tags) > 0 {
		fields = append(fields, fmField{"tags", thread.Tags})
	}
	fields = append(fields, fmField{"tasks", tasks})
	return fields
}

func clonePlannerThreads(threads []domain.Thread) []domain.Thread {
	out := make([]domain.Thread, len(threads))
	for i, thread := range threads {
		out[i] = thread
		out[i].SourceVersion = ""
		out[i].Tags = append([]string(nil), thread.Tags...)
		out[i].Tasks = append([]string(nil), thread.Tasks...)
	}
	return out
}

var testHookBeforeThreadCreationVerify func()
