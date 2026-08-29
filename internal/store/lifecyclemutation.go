package store

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

var _ core.TaskLifecycleMutationStore = (*FS)(nil)

// MutateTaskLifecycle owns one complete lifecycle authorization transaction:
// repository guard, authoritative graph snapshot, pure plan, body-aware policy,
// materialization, whole-snapshot CAS, immediate target CAS, and atomic write.
func (s *FS) MutateTaskLifecycle(now time.Time, dryRun bool, planner core.TaskLifecyclePlanner) (result core.TaskLifecycleMutationResult, err error) {
	result.DryRun = dryRun
	if planner == nil {
		return result, fmt.Errorf("%w: task lifecycle planner is required", domain.ErrValidation)
	}
	if now.IsZero() {
		return result, fmt.Errorf("%w: task lifecycle time is required", domain.ErrValidation)
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
			wrapped := fmt.Errorf("release repository task lifecycle guard: %w", releaseErr)
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
	if err := core.ValidateTaskLifecycleSource(graph); err != nil {
		return result, err
	}

	plan, err := callTaskLifecyclePlanner(s, planner, graph)
	if err != nil {
		return result, err
	}
	result.Plan = plan

	body, err := s.readTaskLifecycleBody(graph, plan)
	if err != nil {
		return result, err
	}
	validated, analysis, err := core.ValidateTaskLifecyclePlan(graph, plan, body)
	if err != nil {
		return result, err
	}
	result.Plan = validated
	materialized, _, err := s.prepareTaskLifecycleMaterialization(graph, validated, now)
	if err != nil {
		return result, err
	}
	result.Task = materialized.task
	result.From = analysis.From
	result.Before = analysis.Before
	result.After = analysis.After
	result.Impacts = cloneStoreTaskImpacts(analysis.Impacts)
	result.OutstandingBlockers = cloneStoreBlockers(analysis.OutstandingBlockers)
	result.OverrideApplied = analysis.OverrideApplied
	result.Changed = materialized.changed
	if dryRun || !materialized.changed {
		return result, nil
	}

	if testHookBeforeLifecycleVerify != nil {
		testHookBeforeLifecycleVerify()
	}
	currentGraph, err := core.LoadTaskGraph(s)
	if err != nil {
		return result, fmt.Errorf("re-read authoritative task graph before lifecycle write: %w", err)
	}
	if !graph.SameSourceSnapshot(currentGraph) {
		return result, fmt.Errorf("repository task graph changed while authorizing lifecycle transition; retry: %w", domain.ErrConflict)
	}

	if materialized.create {
		if err := s.writeNewFileUnlocked(s.tasksDir, materialized.path, materialized.content, "task", materialized.task.ID+"-"+materialized.task.Slug); err != nil {
			return result, err
		}
		result.Committed = true
		return result, nil
	}

	if testHookBeforeLifecycleWrite != nil {
		testHookBeforeLifecycleWrite(materialized.task.ID)
	}
	if err := verifyUnchanged(s.resolvePath, materialized.task.ID, materialized.path, materialized.ifVersion, "task", "lifecycle transition"); err != nil {
		return result, err
	}
	if err := writeFileAtomic(materialized.path, materialized.content, 0o644); err != nil {
		return result, fmt.Errorf("write lifecycle transition for task %s: %w", materialized.task.ID, err)
	}
	result.Committed = true
	return result, nil
}

// readTaskLifecycleBody obtains only the policy input needed for validation.
// Materialization deliberately follows validation so an untrusted planner cannot
// trigger YAML surgery before its semantic plan is authorized.
func (s *FS) readTaskLifecycleBody(graph *core.TaskGraph, plan core.TaskLifecyclePlan) (string, error) {
	if plan.Create != nil {
		return plan.Create.Body, nil
	}
	task, ok := graph.Task(plan.TaskID)
	if !ok {
		return "", nil // the pure validator owns the attributable not-found error
	}
	path, err := s.resolvePath(plan.TaskID)
	if err != nil {
		return "", err
	}
	if path != task.Path {
		return "", fmt.Errorf("task %s changed path during lifecycle snapshot: %w", plan.TaskID, domain.ErrConflict)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read task %s for lifecycle authorization: %w", path, err)
	}
	_, body := splitFrontmatter(content)
	return string(body), nil
}

func callTaskLifecyclePlanner(store *FS, planner core.TaskLifecyclePlanner, graph *core.TaskGraph) (core.TaskLifecyclePlan, error) {
	leave, err := store.enterRepositoryPlanner()
	if err != nil {
		return core.TaskLifecyclePlan{}, err
	}
	defer leave()
	return planner(graph)
}

type materializedTaskLifecycle struct {
	task      domain.Task
	path      string
	ifVersion string
	content   []byte
	changed   bool
	create    bool
}

func (s *FS) prepareTaskLifecycleMaterialization(graph *core.TaskGraph, plan core.TaskLifecyclePlan, now time.Time) (materializedTaskLifecycle, string, error) {
	if plan.Create != nil {
		return s.materializeCreateAndStart(plan, now)
	}
	task, ok := graph.Task(plan.TaskID)
	if !ok {
		return materializedTaskLifecycle{}, "", fmt.Errorf("task %q: %w", plan.TaskID, domain.ErrNotFound)
	}
	path := task.Path
	resolved, err := s.resolvePath(plan.TaskID)
	if err != nil {
		return materializedTaskLifecycle{}, "", err
	}
	if resolved != path {
		return materializedTaskLifecycle{}, "", fmt.Errorf("task %s changed path during lifecycle snapshot: %w", plan.TaskID, domain.ErrConflict)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return materializedTaskLifecycle{}, "", fmt.Errorf("read task %s for lifecycle transition: %w", path, err)
	}
	_, body := splitFrontmatter(content)
	updates := lifecycleFrontmatterUpdates(task.Status, plan.To, plan.RevisitAt, now)
	if len(updates) == 0 {
		return materializedTaskLifecycle{
			task: task, path: path, ifVersion: hashContent(content), content: content,
		}, string(body), nil
	}
	newContent, err := updateFrontmatter(content, updates)
	if err != nil {
		return materializedTaskLifecycle{}, "", err
	}
	parsed, err := parseTask(newContent, path)
	if err != nil {
		return materializedTaskLifecycle{}, "", fmt.Errorf("%w: lifecycle transition for task %s would not reload: %v", domain.ErrValidation, plan.TaskID, err)
	}
	return materializedTaskLifecycle{
		task: parsed, path: path, ifVersion: hashContent(content), content: newContent,
		changed: !bytes.Equal(content, newContent),
	}, string(body), nil
}

func (s *FS) materializeCreateAndStart(plan core.TaskLifecyclePlan, now time.Time) (materializedTaskLifecycle, string, error) {
	creation := plan.Create
	if creation == nil {
		return materializedTaskLifecycle{}, "", fmt.Errorf("%w: create-and-start document is required", domain.ErrValidation)
	}
	task := creation.Task
	if err := validEntityID(task.ID); err != nil {
		return materializedTaskLifecycle{}, "", err
	}
	if task.Slug == "" {
		return materializedTaskLifecycle{}, "", fmt.Errorf("%w: empty task slug", domain.ErrValidation)
	}
	task.Status = domain.StatusInProgress
	task.StartedAt = now.Format("2006-01-02")
	stem := task.ID + "-" + task.Slug
	path := filepath.Join(s.tasksDir, stem+".md")
	content, err := buildFile(taskFields(task), creation.Body)
	if err != nil {
		return materializedTaskLifecycle{}, "", err
	}
	if _, err := os.Stat(path); err == nil {
		return materializedTaskLifecycle{}, "", fmt.Errorf("task %q already exists: %w", stem, domain.ErrConflict)
	} else if !os.IsNotExist(err) {
		return materializedTaskLifecycle{}, "", fmt.Errorf("stat task %s: %w", path, err)
	}
	parsed, err := parseTask(content, path)
	if err != nil {
		return materializedTaskLifecycle{}, "", fmt.Errorf("%w: create-and-start task would not reload: %v", domain.ErrValidation, err)
	}
	return materializedTaskLifecycle{
		task: parsed, path: path, content: content, changed: true, create: true,
	}, creation.Body, nil
}

func lifecycleFrontmatterUpdates(from, to domain.Status, revisitAt string, now time.Time) map[string]any {
	date := now.Format("2006-01-02")
	updates := map[string]any{}
	if from != to {
		updates["status"] = string(to)
		updates["updated_at"] = date
		switch to {
		case domain.StatusInProgress:
			updates["started_at"] = date
		case domain.StatusCompleted:
			updates["completed_at"] = date
		case domain.StatusDeprecated:
			updates["deprecated_at"] = date
		case domain.StatusDeferred:
			updates["deferred_at"] = date
		}
		if from == domain.StatusDeferred {
			updates["revisit_at"] = domain.UnsetField{}
		}
	}
	if revisitAt != "" {
		if from == to {
			updates["updated_at"] = date
		}
		updates["revisit_at"] = revisitAt
	}
	return updates
}

func cloneStoreTaskImpacts(values []core.TaskGraphStateImpact) []core.TaskGraphStateImpact {
	out := make([]core.TaskGraphStateImpact, len(values))
	copy(out, values)
	for i := range out {
		out[i].Path = append([]string(nil), values[i].Path...)
	}
	return out
}

func cloneStoreBlockers(values []core.Blocker) []core.Blocker {
	out := make([]core.Blocker, len(values))
	copy(out, values)
	for i := range out {
		out[i].Path = append([]string(nil), values[i].Path...)
	}
	return out
}

// Test-only interleaving seams for non-cooperating raw editors and concurrent
// lifecycle/dependency writers.
var testHookBeforeLifecycleVerify func()
var testHookBeforeLifecycleWrite func(taskID string)
