package core

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// DependencyOperation is the stable use-case vocabulary carried by mutation
// receipts. It deliberately names user intent rather than persistence mechanics.
type DependencyOperation string

const (
	DependencyAdd     DependencyOperation = "add"
	DependencyRemove  DependencyOperation = "remove"
	DependencyMigrate DependencyOperation = "migrate"
)

// DependencyEdgeOutcome reports one canonical edge intent. Outcome is the
// semantic result; DryRun on the containing receipt distinguishes preview from
// durable application.
type DependencyEdgeOutcome struct {
	DependentID    string
	PrerequisiteID string
	Action         DependencyOperation
	Outcome        string // added | removed | skipped
}

// LegacyFieldClear identifies one legacy field occurrence removed by migration.
type LegacyFieldClear struct {
	TaskID string
	Field  string
}

// DependencyMutationReceipt is the adapter-neutral mutation result. Planned and
// Applied refer to task-file replacements; Remaining is populated only on failure
// so an interrupted multi-file migration is explicitly resumable.
type DependencyMutationReceipt struct {
	Operation           DependencyOperation
	Changed             bool
	DryRun              bool
	Edges               []DependencyEdgeOutcome
	Impacts             []TaskGraphStateImpact
	ClearedLegacyFields []LegacyFieldClear
	PlannedTaskIDs      []string
	AppliedTaskIDs      []string
	RemainingTaskIDs    []string
}

// DependencyMutationFailure preserves a typed receipt on error. Error includes
// the durable prefix for human adapters; machine adapters can errors.As and emit
// the complete structured receipt without parsing this text.
type DependencyMutationFailure struct {
	Cause   error
	Receipt DependencyMutationReceipt
}

func (e *DependencyMutationFailure) Error() string {
	if e == nil || e.Cause == nil {
		return "dependency mutation failed"
	}
	if len(e.Receipt.AppliedTaskIDs) == 0 {
		return e.Cause.Error()
	}
	if len(e.Receipt.RemainingTaskIDs) == 0 {
		return fmt.Sprintf("%v; all planned dependency task files were durably applied to %s; verify current graph state before deciding whether to retry",
			e.Cause, strings.Join(e.Receipt.AppliedTaskIDs, ", "))
	}
	return fmt.Sprintf("%v; durable dependency prefix applied to %s; retry the same command to converge remaining tasks %s",
		e.Cause, strings.Join(e.Receipt.AppliedTaskIDs, ", "), strings.Join(e.Receipt.RemainingTaskIDs, ", "))
}

func (e *DependencyMutationFailure) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type dependencyPlanDetails struct {
	edges  []DependencyEdgeOutcome
	clears []LegacyFieldClear
}

// AddTaskDependencies adds hard prerequisites to one dependent task through the
// repository-global graph guard.
func (s *Service) AddTaskDependencies(taskRef string, prerequisiteRefs []string, dryRun bool) (DependencyMutationReceipt, error) {
	return s.mutateTaskDependencies(DependencyAdd, taskRef, prerequisiteRefs, dryRun)
}

// RemoveTaskDependencies removes hard prerequisites from one dependent task
// through the repository-global graph guard.
func (s *Service) RemoveTaskDependencies(taskRef string, prerequisiteRefs []string, dryRun bool) (DependencyMutationReceipt, error) {
	return s.mutateTaskDependencies(DependencyRemove, taskRef, prerequisiteRefs, dryRun)
}

func (s *Service) mutateTaskDependencies(operation DependencyOperation, taskRef string, prerequisiteRefs []string, dryRun bool) (DependencyMutationReceipt, error) {
	if operation != DependencyAdd && operation != DependencyRemove {
		return DependencyMutationReceipt{}, fmt.Errorf("%w: unsupported dependency operation %q", domain.ErrValidation, operation)
	}
	if strings.TrimSpace(taskRef) == "" {
		return DependencyMutationReceipt{}, fmt.Errorf("%w: dependent task is required", domain.ErrValidation)
	}
	if len(prerequisiteRefs) == 0 {
		return DependencyMutationReceipt{}, fmt.Errorf("%w: at least one --on prerequisite is required", domain.ErrValidation)
	}
	return s.runDependencyMutation(operation, dryRun, func(graph *TaskGraph) (TaskGraphMutationPlan, dependencyPlanDetails, error) {
		return planDependencyEdges(graph, operation, taskRef, prerequisiteRefs)
	})
}

// MigrateTaskDependencies converts every safe legacy dependency occurrence in
// the repository. Broken legacy references are rejected by the guarded source
// validator before this planner runs.
func (s *Service) MigrateTaskDependencies(dryRun bool) (DependencyMutationReceipt, error) {
	return s.runDependencyMutation(DependencyMigrate, dryRun, planLegacyDependencyMigration)
}

type dependencyPlanner func(*TaskGraph) (TaskGraphMutationPlan, dependencyPlanDetails, error)

func (s *Service) runDependencyMutation(operation DependencyOperation, dryRun bool, planner dependencyPlanner) (DependencyMutationReceipt, error) {
	if s.graphMutations == nil {
		return DependencyMutationReceipt{}, fmt.Errorf("dependency mutations are unavailable from this store")
	}
	now := s.now()
	var receipt DependencyMutationReceipt
	for attempt := 0; ; attempt++ {
		details := dependencyPlanDetails{}
		var sourceGraph *TaskGraph
		result, err := s.graphMutations.MutateTaskGraph(now, dryRun, func(graph *TaskGraph) (TaskGraphMutationPlan, error) {
			sourceGraph = graph
			plan, plannedDetails, planErr := planner(graph)
			details = plannedDetails
			return plan, planErr
		})
		impacts := []TaskGraphStateImpact(nil)
		if operation == DependencyAdd || operation == DependencyRemove {
			impacts = directDependencyImpacts(sourceGraph, result.Plan)
		}
		receipt = dependencyReceipt(operation, result, details, impacts, err)
		// A graph conflict is safe to retry only before any replacement landed.
		// Once a durable prefix exists, surface it; silently replaying would erase
		// the recovery event the caller needs to understand.
		if dryRun || !errors.Is(err, domain.ErrConflict) || len(result.AppliedTaskIDs) > 0 || attempt >= s.maxRetries {
			if err != nil {
				return receipt, &DependencyMutationFailure{Cause: err, Receipt: receipt}
			}
			return receipt, nil
		}
		s.retrySleep(attempt + 1)
	}
}

func dependencyReceipt(operation DependencyOperation, result TaskGraphMutationResult, details dependencyPlanDetails, impacts []TaskGraphStateImpact, mutationErr error) DependencyMutationReceipt {
	planned := make([]string, len(result.Plan.TaskWrites))
	for i, write := range result.Plan.TaskWrites {
		planned[i] = write.TaskID
	}
	receipt := DependencyMutationReceipt{
		Operation: operation, Changed: len(planned) > 0, DryRun: result.DryRun,
		Edges:               append([]DependencyEdgeOutcome(nil), details.edges...),
		Impacts:             cloneTaskGraphStateImpacts(impacts),
		ClearedLegacyFields: append([]LegacyFieldClear(nil), details.clears...),
		PlannedTaskIDs:      planned, AppliedTaskIDs: append([]string(nil), result.AppliedTaskIDs...),
	}
	if mutationErr != nil {
		applied := make(map[string]bool, len(result.AppliedTaskIDs))
		for _, taskID := range result.AppliedTaskIDs {
			applied[taskID] = true
		}
		for _, taskID := range planned {
			if !applied[taskID] {
				receipt.RemainingTaskIDs = append(receipt.RemainingTaskIDs, taskID)
			}
		}
	}
	return receipt
}

func planDependencyEdges(graph *TaskGraph, operation DependencyOperation, taskRef string, prerequisiteRefs []string) (TaskGraphMutationPlan, dependencyPlanDetails, error) {
	dependentID, err := graph.ResolveTaskID(taskRef)
	if err != nil {
		return TaskGraphMutationPlan{}, dependencyPlanDetails{}, err
	}
	dependent, ok := graph.Task(dependentID)
	if !ok {
		return TaskGraphMutationPlan{}, dependencyPlanDetails{}, fmt.Errorf("task %q resolved to unreadable task %s: %w", taskRef, dependentID, domain.ErrValidation)
	}

	prerequisites := make([]string, 0, len(prerequisiteRefs))
	seen := make(map[string]string, len(prerequisiteRefs))
	for _, ref := range prerequisiteRefs {
		prerequisiteID, resolveErr := graph.ResolveTaskID(ref)
		if resolveErr != nil {
			return TaskGraphMutationPlan{}, dependencyPlanDetails{}, resolveErr
		}
		if previous, duplicate := seen[prerequisiteID]; duplicate {
			return TaskGraphMutationPlan{}, dependencyPlanDetails{}, fmt.Errorf("%w: prerequisite references %q and %q both resolve to task %s", domain.ErrValidation, previous, ref, prerequisiteID)
		}
		seen[prerequisiteID] = ref
		prerequisites = append(prerequisites, prerequisiteID)
	}
	sort.Strings(prerequisites)

	dependencies := append([]string(nil), dependent.DependsOn...)
	details := dependencyPlanDetails{edges: make([]DependencyEdgeOutcome, 0, len(prerequisites))}
	changed := false
	for _, prerequisiteID := range prerequisites {
		if prerequisiteID == dependentID {
			return TaskGraphMutationPlan{}, dependencyPlanDetails{}, fmt.Errorf("%w: task %s cannot depend on itself", domain.ErrValidation, dependentID)
		}
		present := slices.Contains(dependencies, prerequisiteID)
		outcome := DependencyEdgeOutcome{DependentID: dependentID, PrerequisiteID: prerequisiteID, Action: operation, Outcome: "skipped"}
		switch operation {
		case DependencyAdd:
			if !present {
				dependencies = append(dependencies, prerequisiteID)
				outcome.Outcome = "added"
				changed = true
			}
		case DependencyRemove:
			if present {
				dependencies = slices.DeleteFunc(dependencies, func(id string) bool { return id == prerequisiteID })
				outcome.Outcome = "removed"
				changed = true
			}
		}
		details.edges = append(details.edges, outcome)
	}
	if !changed {
		return TaskGraphMutationPlan{}, details, nil
	}
	return TaskGraphMutationPlan{TaskWrites: []TaskDependencyWrite{{TaskID: dependentID, DependsOn: dependencies}}}, details, nil
}

func planLegacyDependencyMigration(graph *TaskGraph) (TaskGraphMutationPlan, dependencyPlanDetails, error) {
	diagnostics := graph.LegacyDiagnostics()
	if len(diagnostics) == 0 {
		return TaskGraphMutationPlan{}, dependencyPlanDetails{}, nil
	}
	details := dependencyPlanDetails{}
	clearOwners := make(map[string]bool)
	desired := make(map[string][]string)
	edges := make(map[DependencyEdge]DependencyEdgeOutcome)
	clears := make(map[LegacyFieldClear]bool)

	for _, diagnostic := range diagnostics {
		clearOwners[diagnostic.TaskID] = true
		clears[LegacyFieldClear{TaskID: diagnostic.TaskID, Field: diagnostic.Field}] = true
		for _, ref := range diagnostic.References {
			if ref.Resolution != LegacyResolved {
				return TaskGraphMutationPlan{}, dependencyPlanDetails{}, fmt.Errorf("%w: legacy %s reference %q on task %s is %s", domain.ErrValidation, diagnostic.Field, ref.Value, diagnostic.TaskID, ref.Resolution)
			}
			dependent, ok := graph.Task(ref.Edge.To)
			if !ok {
				return TaskGraphMutationPlan{}, dependencyPlanDetails{}, fmt.Errorf("%w: resolved legacy edge targets unreadable task %s", domain.ErrValidation, ref.Edge.To)
			}
			dependencies, initialized := desired[ref.Edge.To]
			if !initialized {
				dependencies = append([]string(nil), dependent.DependsOn...)
			}
			outcome := DependencyEdgeOutcome{DependentID: ref.Edge.To, PrerequisiteID: ref.Edge.From, Action: DependencyAdd, Outcome: "skipped"}
			if !slices.Contains(dependencies, ref.Edge.From) {
				dependencies = append(dependencies, ref.Edge.From)
				outcome.Outcome = "added"
			}
			desired[ref.Edge.To] = dependencies
			if previous, exists := edges[ref.Edge]; !exists || previous.Outcome == "skipped" {
				edges[ref.Edge] = outcome
			}
		}
	}

	for _, outcome := range edges {
		details.edges = append(details.edges, outcome)
	}
	sort.Slice(details.edges, func(i, j int) bool {
		if details.edges[i].DependentID != details.edges[j].DependentID {
			return details.edges[i].DependentID < details.edges[j].DependentID
		}
		return details.edges[i].PrerequisiteID < details.edges[j].PrerequisiteID
	})
	for clear := range clears {
		details.clears = append(details.clears, clear)
	}
	sort.Slice(details.clears, func(i, j int) bool {
		if details.clears[i].TaskID != details.clears[j].TaskID {
			return details.clears[i].TaskID < details.clears[j].TaskID
		}
		return details.clears[i].Field < details.clears[j].Field
	})

	affected := make(map[string]bool)
	for taskID := range clearOwners {
		affected[taskID] = true
	}
	for taskID, dependencies := range desired {
		task, _ := graph.Task(taskID)
		sort.Strings(dependencies)
		desired[taskID] = dependencies
		if !slices.Equal(task.DependsOn, dependencies) {
			affected[taskID] = true
		}
	}

	// Write dependents before prerequisites. This makes a legacy `blocks` edge
	// canonical on its dependent before the prerequisite-owner clears the legacy
	// declaration; duplicate projected edges are harmless, disappearing edges are
	// avoided, and every prefix remains semantically conservative.
	waves, _ := graph.TopologicalWaves()
	waveByTask := make(map[string]int, len(graph.TaskIDs()))
	for waveIndex, wave := range waves {
		for _, taskID := range wave {
			waveByTask[taskID] = waveIndex
		}
	}
	ordered := make([]string, 0, len(affected))
	for taskID := range affected {
		if _, ok := waveByTask[taskID]; !ok {
			return TaskGraphMutationPlan{}, dependencyPlanDetails{}, fmt.Errorf("%w: legacy migration cannot order task %s in the projected graph", domain.ErrValidation, taskID)
		}
		ordered = append(ordered, taskID)
	}
	sort.Slice(ordered, func(i, j int) bool {
		leftWave, rightWave := waveByTask[ordered[i]], waveByTask[ordered[j]]
		if leftWave != rightWave {
			return leftWave > rightWave
		}
		return ordered[i] < ordered[j]
	})

	plan := TaskGraphMutationPlan{TaskWrites: make([]TaskDependencyWrite, 0, len(ordered))}
	for _, taskID := range ordered {
		task, _ := graph.Task(taskID)
		dependencies, ok := desired[taskID]
		if !ok {
			dependencies = append([]string(nil), task.DependsOn...)
		}
		plan.TaskWrites = append(plan.TaskWrites, TaskDependencyWrite{
			TaskID: taskID, DependsOn: dependencies, ClearLegacy: clearOwners[taskID],
		})
	}
	return plan, details, nil
}

// TaskBlockerDetail enriches one graph blocker with any readable task metadata
// and its current derived state. Missing/unreadable blockers keep zero Task data
// while retaining their stable ID, reason, and path.
type TaskBlockerDetail struct {
	Blocker Blocker
	Task    domain.Task
	State   TaskGraphState
}

type TaskBlockersResult struct {
	TaskID     string
	Task       domain.Task
	State      TaskGraphState
	Projection string // frontier | causal
	Health     GraphHealth
	Problems   []GraphProblem
	Legacy     []LegacyDependencyDiagnostic
	Blockers   []TaskBlockerDetail
}

// TaskBlockers returns either the action frontier (default) or the full causal
// closure. It is diagnostic and therefore reports degraded/broken snapshots
// rather than failing the read.
func (s *Service) TaskBlockers(ref string, causal bool) (TaskBlockersResult, error) {
	graph, taskID, task, err := s.resolveTaskGraphQuery(ref)
	if err != nil {
		return TaskBlockersResult{}, err
	}
	projection := "frontier"
	blockers := graph.BlockingFrontier(taskID)
	if causal {
		projection = "causal"
		blockers = graph.CausalBlockers(taskID)
	}
	result := TaskBlockersResult{
		TaskID: taskID, Task: task, State: graph.State(taskID), Projection: projection, Health: graph.Health(),
		Problems: graph.Problems(), Legacy: graph.LegacyDiagnostics(),
		Blockers: make([]TaskBlockerDetail, 0, len(blockers)),
	}
	for _, blocker := range blockers {
		blockerTask, _ := graph.Task(blocker.TaskID)
		result.Blockers = append(result.Blockers, TaskBlockerDetail{
			Blocker: blocker, Task: blockerTask, State: graph.State(blocker.TaskID),
		})
	}
	return result, nil
}

type TaskDependentDetail struct {
	Impact DependentImpact
	Task   domain.Task
	State  TaskGraphState
}

type TaskUnblocksResult struct {
	TaskID   string
	Task     domain.Task
	State    TaskGraphState
	Health   GraphHealth
	Problems []GraphProblem
	Legacy   []LegacyDependencyDiagnostic
	Unblocks []TaskDependentDetail
}

// TaskUnblocks reports transitive downstream impact. It does not simulate a
// completion or claim every dependent would become eligible.
func (s *Service) TaskUnblocks(ref string) (TaskUnblocksResult, error) {
	graph, taskID, task, err := s.resolveTaskGraphQuery(ref)
	if err != nil {
		return TaskUnblocksResult{}, err
	}
	result := TaskUnblocksResult{
		TaskID: taskID, Task: task, State: graph.State(taskID), Health: graph.Health(), Problems: graph.Problems(),
		Legacy: graph.LegacyDiagnostics(),
	}
	for _, impact := range graph.DownstreamImpact(taskID) {
		dependent, _ := graph.Task(impact.TaskID)
		result.Unblocks = append(result.Unblocks, TaskDependentDetail{
			Impact: impact, Task: dependent, State: graph.State(impact.TaskID),
		})
	}
	return result, nil
}

func (s *Service) resolveTaskGraphQuery(ref string) (*TaskGraph, string, domain.Task, error) {
	graph, err := LoadTaskGraph(s.store)
	if err != nil {
		return nil, "", domain.Task{}, err
	}
	taskID, err := graph.ResolveTaskID(ref)
	if err != nil {
		return nil, "", domain.Task{}, err
	}
	task, _ := graph.Task(taskID)
	return graph, taskID, task, nil
}
