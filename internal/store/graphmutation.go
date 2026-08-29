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

var _ core.TaskGraphMutationStore = (*FS)(nil)

// MutateTaskGraph owns the complete cooperating-writer critical section: take
// the repository lock, load the canonical strict snapshot, invoke the pure core
// planner, validate and materialize its semantic writes, apply them through the
// package's lock-free atomic helper, and finally release the lock.
//
// Multi-file plans are deliberately resumable rather than transactionally
// atomic: every file replacement is atomic and the result names the durable
// prefix if a later replacement fails. The planner must therefore be convergent
// when re-run, matching bulk-linking's additive/idempotent contract in ADR-0006.
func (s *FS) MutateTaskGraph(now time.Time, dryRun bool, planner core.TaskGraphPlanner) (result core.TaskGraphMutationResult, err error) {
	result.DryRun = dryRun
	if planner == nil {
		return result, fmt.Errorf("%w: graph mutation planner is required", domain.ErrValidation)
	}
	if now.IsZero() {
		return result, fmt.Errorf("%w: graph mutation time is required", domain.ErrValidation)
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
			wrapped := fmt.Errorf("release repository graph mutation guard: %w", releaseErr)
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
	if err := core.ValidateTaskGraphMutationSource(graph); err != nil {
		return result, err
	}

	plan, err := callTaskGraphPlanner(s, planner, graph)
	if err != nil {
		return result, err
	}
	result.Plan, err = core.ValidateTaskGraphMutationPlan(graph, plan)
	if err != nil {
		return result, err
	}
	writes, err := s.materializeTaskGraphPlan(graph, result.Plan, now)
	if err != nil {
		return result, err
	}
	if dryRun {
		return result, nil
	}

	// Verify the whole materialized prefix before changing the first file. The
	// repository lock excludes cooperating writers; the hashes still catch a raw
	// editor that raced the advisory guard before application begins.
	if testHookBeforeGraphVerify != nil {
		testHookBeforeGraphVerify()
	}
	currentGraph, err := core.LoadTaskGraph(s)
	if err != nil {
		return result, fmt.Errorf("re-read authoritative task graph before write: %w", err)
	}
	if !graph.SameSourceSnapshot(currentGraph) {
		return result, fmt.Errorf("repository task graph changed while planning; retry: %w", domain.ErrConflict)
	}
	for _, write := range writes {
		// Keep the raw-editor CAS window bounded to one atomic replacement. The
		// whole-snapshot check above is still the all-or-nothing preflight; this
		// second per-file check protects later writes after a durable prefix.
		if testHookBeforeGraphWrite != nil {
			testHookBeforeGraphWrite(write.taskID)
		}
		if err := verifyUnchanged(s.resolvePath, write.taskID, write.path, write.ifVersion, "task", "dependency update"); err != nil {
			return result, err
		}
		if err := writeFileAtomic(write.path, write.content, 0o644); err != nil {
			return result, fmt.Errorf("write dependency update for task %s: %w", write.taskID, err)
		}
		result.AppliedTaskIDs = append(result.AppliedTaskIDs, write.taskID)
		if testHookAfterGraphWrite != nil {
			if err := testHookAfterGraphWrite(write.taskID); err != nil {
				return result, fmt.Errorf("after dependency update for task %s: %w", write.taskID, err)
			}
		}
	}
	return result, nil
}

func callTaskGraphPlanner(store *FS, planner core.TaskGraphPlanner, graph *core.TaskGraph) (core.TaskGraphMutationPlan, error) {
	leave, err := store.enterRepositoryPlanner()
	if err != nil {
		return core.TaskGraphMutationPlan{}, err
	}
	defer leave() // also clears the re-entry sentinel when planner panics
	return planner(graph)
}

type materializedTaskGraphWrite struct {
	taskID    string
	path      string
	ifVersion string
	content   []byte
}

func (s *FS) materializeTaskGraphPlan(graph *core.TaskGraph, plan core.TaskGraphMutationPlan, now time.Time) ([]materializedTaskGraphWrite, error) {
	writes := make([]materializedTaskGraphWrite, 0, len(plan.TaskWrites))
	for _, planned := range plan.TaskWrites {
		task, _ := graph.Task(planned.TaskID)
		path, err := s.resolvePath(planned.TaskID)
		if err != nil {
			return nil, err
		}
		if path != task.Path {
			return nil, fmt.Errorf("task %s changed path during graph snapshot: %w", planned.TaskID, domain.ErrConflict)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read task %s for dependency update: %w", path, err)
		}
		updates := map[string]any{}
		if len(planned.DependsOn) == 0 {
			updates["depends_on"] = domain.UnsetField{}
		} else {
			updates["depends_on"] = planned.DependsOn
		}
		if planned.ClearLegacy {
			updates["blocked_by"] = domain.UnsetField{}
			updates["dependencies"] = domain.UnsetField{}
			updates["blocks"] = domain.UnsetField{}
		}
		newContent, err := updateFrontmatter(content, updates)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(content, newContent) {
			continue
		}
		updates["updated_at"] = now.Format("2006-01-02")
		newContent, err = updateFrontmatter(content, updates)
		if err != nil {
			return nil, err
		}
		parsed, err := parseTask(newContent, path)
		if err != nil {
			return nil, fmt.Errorf("%w: dependency update for task %s would not reload: %v", domain.ErrValidation, planned.TaskID, err)
		}
		if !slices.Equal(parsed.DependsOn, planned.DependsOn) {
			return nil, fmt.Errorf("%w: dependency update for task %s did not materialize the planned canonical set", domain.ErrValidation, planned.TaskID)
		}
		if planned.ClearLegacy && (len(parsed.LegacyBlockedBy) > 0 || len(parsed.LegacyDependencies) > 0 || len(parsed.LegacyBlocks) > 0 || len(parsed.LegacyDependencyFields) > 0) {
			return nil, fmt.Errorf("%w: dependency update for task %s did not clear legacy fields", domain.ErrValidation, planned.TaskID)
		}
		writes = append(writes, materializedTaskGraphWrite{
			taskID: planned.TaskID, path: path, ifVersion: hashContent(content), content: newContent,
		})
	}
	return writes, nil
}

// testHookAfterGraphWrite injects a failure after one durable file replacement,
// pinning the prefix-result recovery contract. Nil outside tests.
var testHookAfterGraphWrite func(taskID string) error

// testHookBeforeGraphVerify is the graph-mutation counterpart to the ordinary
// OCC seams: it interleaves a non-cooperating raw edit before the all-file CAS.
var testHookBeforeGraphVerify func()

// testHookBeforeGraphWrite interleaves a raw edit after the whole-snapshot
// preflight but immediately before one target's per-file CAS. Nil outside tests.
var testHookBeforeGraphWrite func(taskID string)
