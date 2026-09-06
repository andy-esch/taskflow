package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

var _ core.TaskGraphRepairStore = (*FS)(nil)

// MutateTaskGraphRepair is the only filesystem capability permitted to plan
// against GraphBroken. It protects task and Thread evidence under one repository
// guard, validates a removal-only source plan, and reports every durable prefix.
func (s *FS) MutateTaskGraphRepair(now time.Time, dryRun bool, planner core.TaskGraphRepairPlanner) (result core.TaskGraphRepairMutationResult, err error) {
	result.DryRun = dryRun
	if planner == nil {
		return result, fmt.Errorf("%w: graph repair planner is required", domain.ErrValidation)
	}
	if now.IsZero() {
		return result, fmt.Errorf("%w: graph repair time is required", domain.ErrValidation)
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
			wrapped := fmt.Errorf("release repository graph repair guard: %w", releaseErr)
			if err == nil {
				err = wrapped
			} else {
				err = errors.Join(err, wrapped)
			}
			setRepairRemainingSources(&result)
		}
	}()

	threadRead, err := s.ReadThreads()
	if err != nil {
		return result, fmt.Errorf("load Thread impact evidence for repair: %w", err)
	}
	graph, err := core.LoadTaskGraph(s)
	if err != nil {
		return result, fmt.Errorf("load authoritative task graph for repair: %w", err)
	}
	result.Threads = threadRead.Threads
	result.ThreadProblems = threadRead.Problems

	plan, err := callTaskGraphRepairPlanner(s, planner, graph)
	if err != nil {
		return result, err
	}
	result.Plan, result.Analysis, err = core.ValidateTaskGraphRepairPlan(graph, plan)
	if err != nil {
		return result, err
	}
	writes, err := s.materializeTaskGraphRepair(result.Analysis, now)
	if err != nil {
		return result, err
	}
	for _, write := range writes {
		result.PlannedSources = append(result.PlannedSources, write.source)
	}
	result.FinalGraph = result.Analysis.Prospective
	if dryRun {
		result.Changed = len(writes) > 0
		return result, nil
	}
	// From this point until a write lands, receipt state describes the unchanged
	// repository, not the complete prospective plan.
	_, result.Analysis, _ = core.ValidateTaskGraphRepairPlan(graph, core.TaskGraphRepairPlan{})
	result.FinalGraph = graph
	if len(writes) == 0 {
		return result, nil
	}

	if testHookBeforeGraphRepairVerify != nil {
		testHookBeforeGraphRepairVerify()
	}
	currentThreads, err := s.ReadThreads()
	if err != nil {
		return result, fmt.Errorf("re-read Thread evidence before repair: %w", err)
	}
	currentGraph, err := core.LoadTaskGraph(s)
	if err != nil {
		return result, fmt.Errorf("re-read authoritative task graph before repair: %w", err)
	}
	if graphErr := verifyTaskGraphSourceSnapshot(graph, currentGraph); graphErr != nil || verifyThreadSourceSnapshot(threadRead, currentThreads) != nil {
		setRepairRemainingSources(&result)
		return result, fmt.Errorf("repository task or Thread evidence changed while planning repair; retry: %w", domain.ErrConflict)
	}
	expectedGraph := currentGraph

	for index, write := range writes {
		if testHookBeforeGraphRepairWrite != nil {
			testHookBeforeGraphRepairWrite(write.source)
		}
		currentThreads, evidenceErr := s.ReadThreads()
		if evidenceErr != nil || verifyThreadSourceSnapshot(threadRead, currentThreads) != nil {
			setRepairRemainingSources(&result)
			return result, fmt.Errorf("thread evidence changed before graph repair replacement; retry: %w", domain.ErrConflict)
		}
		currentGraph, evidenceErr = core.LoadTaskGraph(s)
		if evidenceErr != nil || verifyTaskGraphSourceSnapshot(expectedGraph, currentGraph) != nil {
			setRepairRemainingSources(&result)
			return result, fmt.Errorf("task evidence changed before graph repair replacement; retry: %w", domain.ErrConflict)
		}
		content, readErr := os.ReadFile(write.path)
		if readErr != nil {
			setRepairRemainingSources(&result)
			return result, fmt.Errorf("re-read graph repair source %s: %w", write.path, domain.ErrConflict)
		}
		if hashContent(content) != write.ifVersion {
			setRepairRemainingSources(&result)
			return result, fmt.Errorf("graph repair source %s changed before replacement: %w", write.path, domain.ErrConflict)
		}
		stepPlan := repairPlanForSource(result.Plan, write.source)
		if testHookGraphRepairStepValidation != nil {
			testHookGraphRepairStepValidation(stepPlan)
		}
		_, stepAnalysis, analysisErr := core.ValidateTaskGraphRepairPlan(currentGraph, stepPlan)
		if analysisErr != nil {
			setRepairRemainingSources(&result)
			return result, fmt.Errorf("validate next graph repair source %s: %w", write.path, analysisErr)
		}
		nextAnalysis, analysisErr := core.ExtendTaskGraphRepairAnalysis(result.Analysis, stepAnalysis)
		if analysisErr != nil {
			setRepairRemainingSources(&result)
			return result, fmt.Errorf("compose durable graph repair prefix: %w", analysisErr)
		}
		if writeErr := writeFileAtomic(write.path, write.content, 0o644); writeErr != nil {
			setRepairRemainingSources(&result)
			return result, fmt.Errorf("write graph repair for %s: %w", write.path, writeErr)
		}
		result.AppliedSources = append(result.AppliedSources, write.source)
		result.Changed = true
		result.Committed = true
		result.Analysis = nextAnalysis
		result.FinalGraph = nextAnalysis.Prospective
		postThreads, evidenceErr := s.ReadThreads()
		if evidenceErr != nil || verifyThreadSourceSnapshot(threadRead, postThreads) != nil {
			setRepairRemainingSources(&result)
			return result, fmt.Errorf("thread evidence changed while graph repair committed a prefix; inspect before retrying: %w", domain.ErrConflict)
		}
		postGraph, evidenceErr := core.LoadTaskGraph(s)
		if evidenceErr != nil || !stepAnalysis.Prospective.SameRepairSnapshot(postGraph, []core.TaskGraphSourceRef{write.source}) {
			setRepairRemainingSources(&result)
			return result, fmt.Errorf("task evidence changed while graph repair committed a prefix; inspect before retrying: %w", domain.ErrConflict)
		}
		expectedGraph = postGraph
		if testHookAfterGraphRepairWrite != nil {
			if hookErr := testHookAfterGraphRepairWrite(write.source, index); hookErr != nil {
				setRepairRemainingSources(&result)
				return result, fmt.Errorf("after graph repair for %s: %w", write.path, hookErr)
			}
		}
	}
	return result, nil
}

func callTaskGraphRepairPlanner(store *FS, planner core.TaskGraphRepairPlanner, graph *core.TaskGraph) (core.TaskGraphRepairPlan, error) {
	leave, err := store.enterRepositoryPlanner()
	if err != nil {
		return core.TaskGraphRepairPlan{}, err
	}
	defer leave()
	return planner(graph)
}

type materializedTaskGraphRepair struct {
	source    core.TaskGraphSourceRef
	path      string
	ifVersion string
	content   []byte
}

func (s *FS) materializeTaskGraphRepair(analysis core.TaskGraphRepairAnalysis, now time.Time) ([]materializedTaskGraphRepair, error) {
	writes := make([]materializedTaskGraphRepair, 0, len(analysis.SourceGroups))
	for _, group := range analysis.SourceGroups {
		path := group.Source.Location
		if path == "" {
			return nil, fmt.Errorf("%w: filesystem graph repair requires source location for task %s", domain.ErrValidation, group.Source.TaskID)
		}
		relative, err := filepath.Rel(s.tasksDir, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
			return nil, fmt.Errorf("%w: graph repair source %q is outside the task directory", domain.ErrValidation, path)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read graph repair source %s: %w", path, err)
		}
		updated, changed, err := updateDependencySourceEdits(content, group.Edits, now.Format("2006-01-02"))
		if err != nil {
			return nil, fmt.Errorf("materialize graph repair for %s: %w", path, err)
		}
		if !changed {
			return nil, fmt.Errorf("%w: validated graph repair for %s produced no source change", domain.ErrValidation, path)
		}
		parsed, err := parseTask(updated, path)
		if err != nil {
			return nil, fmt.Errorf("%w: graph repair for %s would not reload: %v", domain.ErrValidation, path, err)
		}
		actualRecords, err := core.NewTaskGraph([]domain.Task{parsed}, nil).SourceRecords()
		if err != nil || len(actualRecords) != 1 {
			return nil, fmt.Errorf("%w: graph repair for %s has no reloadable source projection", domain.ErrValidation, path)
		}
		expected, ok := repairSourceRecord(analysis.Prospective, group.Source)
		if !ok || !reflect.DeepEqual(expected.Fields, actualRecords[0].Fields) {
			return nil, fmt.Errorf("%w: graph repair for %s did not materialize only the authorized declarations", domain.ErrValidation, path)
		}
		writes = append(writes, materializedTaskGraphRepair{
			source: group.Source, path: path, ifVersion: hashContent(content), content: updated,
		})
	}
	return writes, nil
}

func repairSourceRecord(graph *core.TaskGraph, source core.TaskGraphSourceRef) (core.TaskGraphSourceRecord, bool) {
	records, err := graph.SourceRecords()
	if err != nil {
		return core.TaskGraphSourceRecord{}, false
	}
	for _, record := range records {
		if record.Source == source {
			return record, true
		}
	}
	return core.TaskGraphSourceRecord{}, false
}

func repairPlanForSource(plan core.TaskGraphRepairPlan, source core.TaskGraphSourceRef) core.TaskGraphRepairPlan {
	step := core.TaskGraphRepairPlan{}
	for _, selection := range plan.Selections {
		if selection.Source == source {
			step.Selections = append(step.Selections, selection)
		}
	}
	for _, operation := range plan.Operations {
		if operation.Edit.Source == source {
			step.Operations = append(step.Operations, operation)
		}
	}
	return step
}

func setRepairRemainingSources(result *core.TaskGraphRepairMutationResult) {
	applied := make(map[core.TaskGraphSourceRef]bool, len(result.AppliedSources))
	for _, source := range result.AppliedSources {
		applied[source] = true
	}
	result.RemainingSources = result.RemainingSources[:0]
	for _, source := range result.PlannedSources {
		if !applied[source] {
			result.RemainingSources = append(result.RemainingSources, source)
		}
	}
}

// Concurrency and injected-prefix seams. Nil outside tests.
var testHookBeforeGraphRepairVerify func()
var testHookBeforeGraphRepairWrite func(core.TaskGraphSourceRef)
var testHookAfterGraphRepairWrite func(core.TaskGraphSourceRef, int) error
var testHookGraphRepairStepValidation func(core.TaskGraphRepairPlan)
