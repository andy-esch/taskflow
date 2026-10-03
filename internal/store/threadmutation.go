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
	threadRead, err := s.ReadThreads()
	if err != nil {
		return result, fmt.Errorf("load authoritative Threads: %w", err)
	}
	if err := core.ValidateThreadMutationSource(graph, threadRead); err != nil {
		return result, err
	}
	snapshot := core.ThreadMutationSnapshot{Graph: graph, Threads: clonePlannerThreads(threadRead.SemanticThreads())}
	plan, err := callThreadMutationPlanner(s, planner, snapshot)
	if err != nil {
		return result, err
	}
	validated, analysis, err := core.ValidateThreadMutationPlan(snapshot, plan, now)
	if err != nil {
		return result, err
	}
	source, err := threadMutationSource(threadRead, validated.ThreadID)
	if err != nil {
		return result, err
	}
	result.Plan = validated
	result.Before = core.ProjectLoadedThread(core.LoadedRecord[domain.Thread]{Value: analysis.Before.Thread, Source: source.Record.Source}, graph)
	result.After = core.ProjectLoadedThread(core.LoadedRecord[domain.Thread]{Value: analysis.After.Thread, Source: source.Record.Source}, graph)
	result.MemberOutcomes = append([]core.ThreadMemberOutcome(nil), analysis.MemberOutcomes...)
	result.Changed = analysis.Changed

	materialized, err := s.materializeThreadMutation(source, validated, analysis)
	if err != nil {
		return result, err
	}
	result.Thread = materialized.thread
	result.LocalPath = materialized.path
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
	currentThreadRead, err := s.ReadThreads()
	if err != nil {
		return result, fmt.Errorf("re-read authoritative Threads before Thread write: %w", err)
	}
	if graphErr := verifyTaskGraphSourceSnapshot(graph, currentGraph); graphErr != nil || verifyThreadSourceSnapshot(threadRead, currentThreadRead) != nil {
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
	// Callback-owned slices must not rewrite the snapshot used to authorize its plan.
	snapshot.Threads = clonePlannerThreads(snapshot.Threads)
	return planner(snapshot)
}

type materializedThreadMutation struct {
	thread    domain.Thread
	path      string
	ifVersion string
	content   []byte
	changed   bool
}

func threadMutationSource(read core.ThreadRead, threadID string) (core.VersionedRecord[domain.Thread], error) {
	var source core.VersionedRecord[domain.Thread]
	for _, record := range read.Records {
		if record.Record.Source.ID != threadID {
			continue
		}
		if source.Record.Source.ID != "" {
			return core.VersionedRecord[domain.Thread]{}, fmt.Errorf("%w: Thread mutation target %s has duplicate source records", domain.ErrValidation, threadID)
		}
		source = record
	}
	if source.Record.Source.ID == "" {
		return core.VersionedRecord[domain.Thread]{}, fmt.Errorf("%w: Thread mutation target %s has no source record", domain.ErrValidation, threadID)
	}
	return source, nil
}

// materializeThreadMutation is deliberately lock-free and update-only so a
// future compound apply can compose it under one outer repository guard. The
// source record supplies its local target and original byte revision explicitly.
func (s *FS) materializeThreadMutation(source core.VersionedRecord[domain.Thread], plan core.ThreadMutationPlan, analysis core.ThreadMutationAnalysis) (materializedThreadMutation, error) {
	before, after := analysis.Before.Thread, analysis.After.Thread
	if source.Record.Source.ID != plan.ThreadID || before.ID != plan.ThreadID || after.ID != plan.ThreadID || source.Record.Value.ID != plan.ThreadID || source.LocalPath == "" || source.SourceVersion == "" {
		return materializedThreadMutation{}, fmt.Errorf("%w: Thread mutation analysis does not identify its target document", domain.ErrValidation)
	}
	resolved, err := s.resolveThread(plan.ThreadID)
	if err != nil {
		return materializedThreadMutation{}, err
	}
	if resolved != source.LocalPath {
		return materializedThreadMutation{}, fmt.Errorf("thread %s changed path during mutation snapshot: %w", plan.ThreadID, domain.ErrConflict)
	}
	content, err := os.ReadFile(source.LocalPath)
	if err != nil {
		return materializedThreadMutation{}, fmt.Errorf("read Thread %s for mutation: %w", source.LocalPath, err)
	}
	if hashContent(content) != source.SourceVersion {
		return materializedThreadMutation{}, fmt.Errorf("thread %s changed content during mutation snapshot: %w", plan.ThreadID, domain.ErrConflict)
	}
	materialized := materializedThreadMutation{
		thread: before, path: source.LocalPath, ifVersion: source.SourceVersion, content: content,
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
	parsed, err := parseThread(newContent, source.LocalPath)
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
