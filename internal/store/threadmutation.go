package store

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

var _ core.ThreadMutationStore = (*FS)(nil)

// MutateThread owns one existing-Thread read/authorize/write transaction over
// the authoritative task graph and complete Thread set.
func (s *FS) MutateThread(now time.Time, dryRun bool, planner core.ThreadMutationPlanner) (result core.ThreadMutationResult, err error) {
	result.DryRun = dryRun
	if planner == nil {
		return result, fmt.Errorf("%w: Thread mutation planner is required", domain.ErrValidation)
	}
	if now.IsZero() {
		return result, fmt.Errorf("%w: Thread mutation time is required", domain.ErrValidation)
	}
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return result, err
	}

	unlock, err := s.checkedWriteLock()
	if err != nil {
		return result, err
	}
	defer func() {
		if releaseErr := unlock(); releaseErr != nil {
			wrapped := fmt.Errorf("release repository Thread mutation guard: %w", releaseErr)
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
	threads, problems, err := s.ListThreads()
	if err != nil {
		return result, fmt.Errorf("load authoritative Threads: %w", err)
	}
	if err := core.ValidateThreadMutationSource(graph, threads, problems); err != nil {
		return result, err
	}
	snapshot := core.ThreadMutationSnapshot{Graph: graph, Threads: clonePlannerThreads(threads)}
	plan, err := callThreadMutationPlanner(s, planner, snapshot)
	if err != nil {
		return result, err
	}
	validated, analysis, err := core.ValidateThreadMutationPlan(snapshot, plan, now)
	if err != nil {
		return result, err
	}
	result.Plan = validated
	result.Before = analysis.Before
	result.After = analysis.After
	result.MemberOutcomes = append([]core.ThreadMemberOutcome(nil), analysis.MemberOutcomes...)
	result.Changed = analysis.Changed

	materialized, err := s.materializeThreadMutation(validated, analysis)
	if err != nil {
		return result, err
	}
	result.Thread = materialized.thread
	if dryRun || !materialized.changed {
		return result, nil
	}

	if testHookBeforeThreadMutationVerify != nil {
		testHookBeforeThreadMutationVerify()
	}
	currentGraph, err := core.LoadTaskGraph(s)
	if err != nil {
		return result, fmt.Errorf("re-read authoritative task graph before Thread write: %w", err)
	}
	currentThreads, currentProblems, err := s.ListThreads()
	if err != nil {
		return result, fmt.Errorf("re-read authoritative Threads before Thread write: %w", err)
	}
	if !graph.SameSourceSnapshot(currentGraph) || !sameThreadSourceSnapshot(threads, problems, currentThreads, currentProblems) {
		return result, fmt.Errorf("repository tasks or Threads changed while authorizing Thread mutation; retry: %w", domain.ErrConflict)
	}

	if testHookBeforeThreadMutationWrite != nil {
		testHookBeforeThreadMutationWrite(materialized.thread.ID)
	}
	if err := verifyUnchanged(s.resolveThread, materialized.thread.ID, materialized.path, materialized.ifVersion, "Thread", "mutation"); err != nil {
		return result, err
	}
	if err := writeFileAtomic(materialized.path, materialized.content, 0o644); err != nil {
		return result, fmt.Errorf("write Thread mutation for %s: %w", materialized.thread.ID, err)
	}
	result.Committed = true
	return result, nil
}

func callThreadMutationPlanner(store *FS, planner core.ThreadMutationPlanner, snapshot core.ThreadMutationSnapshot) (core.ThreadMutationPlan, error) {
	leave, err := store.enterRepositoryPlanner()
	if err != nil {
		return core.ThreadMutationPlan{}, err
	}
	defer leave()
	return planner(snapshot)
}

type materializedThreadMutation struct {
	thread    domain.Thread
	path      string
	ifVersion string
	content   []byte
	changed   bool
}

// materializeThreadMutation is deliberately lock-free and update-only so a
// future compound apply can compose it under one outer repository guard.
func (s *FS) materializeThreadMutation(plan core.ThreadMutationPlan, analysis core.ThreadMutationAnalysis) (materializedThreadMutation, error) {
	before, after := analysis.Before.Thread, analysis.After.Thread
	if before.ID != plan.ThreadID || after.ID != plan.ThreadID || before.Path == "" {
		return materializedThreadMutation{}, fmt.Errorf("%w: Thread mutation analysis does not identify its target document", domain.ErrValidation)
	}
	resolved, err := s.resolveThread(plan.ThreadID)
	if err != nil {
		return materializedThreadMutation{}, err
	}
	if resolved != before.Path {
		return materializedThreadMutation{}, fmt.Errorf("thread %s changed path during mutation snapshot: %w", plan.ThreadID, domain.ErrConflict)
	}
	content, err := os.ReadFile(before.Path)
	if err != nil {
		return materializedThreadMutation{}, fmt.Errorf("read Thread %s for mutation: %w", before.Path, err)
	}
	materialized := materializedThreadMutation{
		thread: before, path: before.Path, ifVersion: hashContent(content), content: content,
	}
	if !analysis.Changed {
		return materialized, nil
	}
	updates := make(map[string]any)
	if !slices.Equal(before.Tasks, after.Tasks) {
		updates["tasks"] = append([]string(nil), after.Tasks...)
	}
	if before.Status != after.Status {
		updates["status"] = string(after.Status)
	}
	if before.Updated != after.Updated {
		updates["updated_at"] = after.Updated
	}
	if before.StartedAt != after.StartedAt {
		if after.StartedAt == "" {
			updates["started_at"] = domain.UnsetField{}
		} else {
			updates["started_at"] = after.StartedAt
		}
	}
	if before.EndedAt != after.EndedAt {
		if after.EndedAt == "" {
			updates["ended_at"] = domain.UnsetField{}
		} else {
			updates["ended_at"] = after.EndedAt
		}
	}
	newContent, err := updateFrontmatter(content, updates)
	if err != nil {
		return materializedThreadMutation{}, err
	}
	parsed, err := parseThread(newContent, before.Path)
	if err != nil {
		return materializedThreadMutation{}, fmt.Errorf("%w: Thread mutation for %s would not reload: %v", domain.ErrValidation, plan.ThreadID, err)
	}
	if err := domain.ValidateThreadDocument(parsed); err != nil {
		return materializedThreadMutation{}, err
	}
	if parsed.Status != after.Status || parsed.Updated != after.Updated || parsed.StartedAt != after.StartedAt || parsed.EndedAt != after.EndedAt || !slices.Equal(parsed.Tasks, after.Tasks) {
		return materializedThreadMutation{}, fmt.Errorf("%w: Thread mutation for %s did not materialize the authorized state", domain.ErrValidation, plan.ThreadID)
	}
	materialized.thread = parsed
	materialized.content = newContent
	materialized.changed = !bytes.Equal(content, newContent)
	return materialized, nil
}

var testHookBeforeThreadMutationVerify func()
var testHookBeforeThreadMutationWrite func(threadID string)
