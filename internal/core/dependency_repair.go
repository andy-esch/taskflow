package core

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

// TaskGraphRepairReason is stable taskflow-owned diagnosis vocabulary. Repair
// operates on source declarations rather than projected edges so invalid raw
// values, duplicate occurrences, and legacy blocks ownership remain visible.
type TaskGraphRepairReason string

const (
	RepairDuplicate       TaskGraphRepairReason = "duplicate-dependency"
	RepairSelf            TaskGraphRepairReason = "self-dependency"
	RepairInvalidID       TaskGraphRepairReason = "invalid-dependency-id"
	RepairDangling        TaskGraphRepairReason = "missing-dependency"
	RepairCycle           TaskGraphRepairReason = "cycle"
	RepairLegacyMissing   TaskGraphRepairReason = "legacy-reference-missing"
	RepairLegacyAmbiguous TaskGraphRepairReason = "legacy-reference-ambiguous"
	RepairLegacyUnsafe    TaskGraphRepairReason = "legacy-reference-unsafe"
	RepairLegacyEmpty     TaskGraphRepairReason = "legacy-empty-field"
	RepairDirectEdit      TaskGraphRepairReason = "needs-direct-edit"
)

// TaskGraphRepairDefect attributes one repair choice or one residual problem.
// Target is zero only for defects such as unreadable or duplicate task identity
// that dependency repair cannot safely change.
type TaskGraphRepairDefect struct {
	Reason           TaskGraphRepairReason
	Target           TaskGraphSourceEdit
	ProjectedEdge    DependencyEdge
	HasProjectedEdge bool
	Automatic        bool
	Repairable       bool
	Problem          GraphProblem
	CandidateIDs     []string
}

// TaskGraphRepairDiagnosis is a deterministic source-level inventory. Defects
// may include several cycle-breaking alternatives; Problems remains the graph's
// ordinary diagnostic projection for other consumers.
type TaskGraphRepairDiagnosis struct {
	Health   GraphHealth
	Defects  []TaskGraphRepairDefect
	Problems []GraphProblem
}

// TaskGraphRepairRequest is convergent user intent. Auto selects only operations
// whose information-loss policy is explicitly safe; Edits are reauthorized
// against the current guarded snapshot rather than trusted as an old plan.
type TaskGraphRepairRequest struct {
	Auto  bool
	Edits []TaskGraphSourceEdit
}

// TaskGraphRepairOperation is one normalized, currently authorized edit.
type TaskGraphRepairOperation struct {
	Edit      TaskGraphSourceEdit
	Reason    TaskGraphRepairReason
	Automatic bool
}

// TaskGraphRepairPlan is a removal-only plan ordered by source file. The store
// materializes a complete group per file so every durable prefix is syntactically
// valid and has already passed the structural and preservation checks.
type TaskGraphRepairPlan struct {
	Selections []TaskGraphSourceEdit
	Operations []TaskGraphRepairOperation
}

// TaskGraphRepairMutationResult is the guarded store's adapter-neutral result.
// FinalGraph reflects the durable prefix on failure and the prospective graph on
// dry-run; it never escapes through the wire mapper.
type TaskGraphRepairMutationResult struct {
	Plan             TaskGraphRepairPlan
	Analysis         TaskGraphRepairAnalysis
	FinalGraph       *TaskGraph
	Threads          []domain.Thread
	ThreadProblems   []ThreadReadProblem
	PlannedSources   []TaskGraphSourceRef
	AppliedSources   []TaskGraphSourceRef
	RemainingSources []TaskGraphSourceRef
	Changed          bool
	DryRun           bool
	Committed        bool
}

// TaskGraphRepairReceipt is the public semantic recovery result. Workspace
// identity remains adapter-owned and is added only at the wire boundary.
type TaskGraphRepairReceipt struct {
	Changed           bool
	DryRun            bool
	Committed         bool
	InitialHealth     GraphHealth
	FinalHealth       GraphHealth
	Selected          []TaskGraphSourceEdit
	Operations        []TaskGraphRepairOperation
	Removed           []TaskGraphSourceDeclaration
	Addressed         []TaskGraphRepairDefect
	Residual          []TaskGraphRepairDefect
	Problems          []GraphProblem
	PlannedFiles      []string
	AppliedFiles      []string
	RemainingFiles    []string
	Impacts           []TaskGraphStateImpact
	ThreadImpacts     []ThreadProjectionImpact
	IncompleteThreads []ThreadReadProblem
}

// TaskGraphRepairFailure preserves a committed prefix and its recovery receipt.
type TaskGraphRepairFailure struct {
	Cause   error
	Receipt TaskGraphRepairReceipt
}

func (e *TaskGraphRepairFailure) Error() string {
	if e == nil || e.Cause == nil {
		return "task graph repair failed"
	}
	if len(e.Receipt.AppliedFiles) == 0 {
		return e.Cause.Error()
	}
	if len(e.Receipt.RemainingFiles) == 0 {
		return fmt.Sprintf("%v; the selected graph repair committed, but repository cleanup failed; inspect current graph state before retrying", e.Cause)
	}
	return fmt.Sprintf("%v; durable repair prefix applied to %s; retry the same repair intent to converge remaining files %s",
		e.Cause, strings.Join(e.Receipt.AppliedFiles, ", "), strings.Join(e.Receipt.RemainingFiles, ", "))
}

func (e *TaskGraphRepairFailure) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// taskGraphRepairMeasure is intentionally internal. Public consumers receive
// the attributable defects and before/after health, not an implementation tuple
// that would freeze today's proof strategy into the wire contract.
type taskGraphRepairMeasure struct {
	unreadable    int
	identity      int
	self          int
	cyclicEdges   int
	legacyUnsafe  int
	declarations  int
	legacyPresent int
}

// TaskGraphRepairAnalysis is the pure proof result consumed by guarded stores.
// Prospective is intentionally core-internal data, not public wire vocabulary.
type TaskGraphRepairAnalysis struct {
	beforeGraph  *TaskGraph
	Before       TaskGraphRepairDiagnosis
	After        TaskGraphRepairDiagnosis
	Prospective  *TaskGraph
	Removed      []TaskGraphSourceDeclaration
	SourceGroups []TaskGraphRepairSourceGroup
	Impacts      []TaskGraphStateImpact
}

// ExtendTaskGraphRepairAnalysis composes one freshly validated source-file
// step onto an already durable prefix. The step must have been planned from
// the current repository snapshot; prior repair-owned timestamp and revision
// changes are the only representation differences tolerated at that seam.
// This lets adapters keep truthful prefix receipts without revalidating every
// earlier source group after each write.
func ExtendTaskGraphRepairAnalysis(prefix, step TaskGraphRepairAnalysis) (TaskGraphRepairAnalysis, error) {
	if prefix.beforeGraph == nil || prefix.Prospective == nil || step.beforeGraph == nil || step.Prospective == nil {
		return TaskGraphRepairAnalysis{}, fmt.Errorf("%w: complete repair analyses are required for prefix composition", domain.ErrValidation)
	}
	if len(step.SourceGroups) != 1 {
		return TaskGraphRepairAnalysis{}, fmt.Errorf("%w: repair prefix step must contain exactly one source group", domain.ErrValidation)
	}
	changed := make([]TaskGraphSourceRef, 0, len(prefix.SourceGroups))
	for _, group := range prefix.SourceGroups {
		changed = append(changed, group.Source)
	}
	if !prefix.Prospective.SameRepairSnapshot(step.beforeGraph, changed) {
		return TaskGraphRepairAnalysis{}, fmt.Errorf("%w: repair prefix step was not validated against the current durable snapshot", domain.ErrConflict)
	}
	groups := append([]TaskGraphRepairSourceGroup(nil), prefix.SourceGroups...)
	groups = append(groups, step.SourceGroups[0])
	removed := append([]TaskGraphSourceDeclaration(nil), prefix.Removed...)
	removed = append(removed, step.Removed...)
	return TaskGraphRepairAnalysis{
		beforeGraph:  prefix.beforeGraph,
		Before:       prefix.Before,
		After:        step.After,
		Prospective:  step.Prospective,
		Removed:      removed,
		SourceGroups: groups,
		Impacts:      taskGraphRepairImpacts(prefix.beforeGraph, step.Prospective, groups),
	}, nil
}

// TaskGraphRepairSourceGroup is one atomic task-file replacement in a validated
// durable order. Source identity, rather than a representative task ID, keeps
// duplicate-ID shadows distinguishable and gives the adapter an exact CAS target.
type TaskGraphRepairSourceGroup struct {
	Source TaskGraphSourceRef
	Edits  []TaskGraphSourceEdit
}

// DiagnoseTaskGraphRepair inventories the repairable source declarations and
// the identity/read failures that need direct editing. It never guesses which
// edge should break a cycle or whether legacy human intent should be discarded.
func DiagnoseTaskGraphRepair(graph *TaskGraph) (TaskGraphRepairDiagnosis, error) {
	if graph == nil {
		return TaskGraphRepairDiagnosis{}, fmt.Errorf("%w: authoritative task graph is required", domain.ErrValidation)
	}
	if err := graph.requireCompleteSource(); err != nil {
		return TaskGraphRepairDiagnosis{}, err
	}
	diagnosis := TaskGraphRepairDiagnosis{
		Health: graph.Health(), Problems: graph.Problems(), Defects: make([]TaskGraphRepairDefect, 0),
	}
	records := graph.sourceRecords()
	declarations, _ := graph.SourceDeclarations()

	legacyRefs := legacyReferenceIndex(graph.legacy)
	seenCanonical := make(map[string]bool)
	for _, declaration := range declarations {
		edit := TaskGraphSourceEdit{
			Action: TaskGraphSourceDropDeclaration, Source: declaration.Source,
			Field: declaration.Field, Value: declaration.Value, Occurrence: declaration.Occurrence,
		}
		defect := TaskGraphRepairDefect{
			Target: edit, ProjectedEdge: declaration.ProjectedEdge,
			HasProjectedEdge: declaration.HasProjectedEdge, Repairable: true,
		}
		switch declaration.Field {
		case TaskDependencyDependsOn:
			duplicateKey := sourceFieldValueKey(declaration.Source, declaration.Field, declaration.Value)
			switch {
			case declaration.Value == declaration.Source.TaskID && declaration.Source.TaskID != "":
				defect.Reason = RepairSelf
				defect.Automatic = true
			case declaration.Occurrence > 0:
				if seenCanonical[duplicateKey] {
					continue
				}
				seenCanonical[duplicateKey] = true
				defect.Reason = RepairDuplicate
				defect.Automatic = true
				defect.Target.Action = TaskGraphSourceDedupe
				defect.Target.Occurrence = 0
			case !id.Valid(declaration.Value):
				defect.Reason = RepairInvalidID
			case !graph.hasKnownOrUnreadableTask(declaration.Value):
				defect.Reason = RepairDangling
			case declaration.HasProjectedEdge && graph.edgeInCycle(declaration.ProjectedEdge):
				defect.Reason = RepairCycle
			default:
				continue
			}
		default:
			ref, ok := legacyRefs[sourceFieldOccurrenceKey(declaration.Source, declaration.Field, declaration.Value, declaration.Occurrence)]
			if !ok {
				continue
			}
			defect.CandidateIDs = append([]string(nil), ref.CandidateIDs...)
			switch ref.Resolution {
			case LegacyMissing:
				defect.Reason = RepairLegacyMissing
			case LegacyAmbiguous:
				defect.Reason = RepairLegacyAmbiguous
			case LegacyUnsafe:
				defect.Reason = RepairLegacyUnsafe
			default:
				continue
			}
		}
		diagnosis.Defects = append(diagnosis.Defects, defect)
	}
	for _, record := range records {
		for _, field := range record.Fields {
			if field.Field == TaskDependencyDependsOn || len(field.Values) != 0 {
				continue
			}
			diagnosis.Defects = append(diagnosis.Defects, TaskGraphRepairDefect{
				Reason: RepairLegacyEmpty, Automatic: true, Repairable: true,
				Target: TaskGraphSourceEdit{Action: TaskGraphSourceDropEmptyField, Source: record.Source, Field: field.Field},
			})
		}
	}
	for _, problem := range diagnosis.Problems {
		if repairProblemHasDeclarationCandidate(problem.Code) {
			continue
		}
		diagnosis.Defects = append(diagnosis.Defects, TaskGraphRepairDefect{
			Reason: RepairDirectEdit, Repairable: false, Problem: problem,
		})
	}
	sort.SliceStable(diagnosis.Defects, func(i, j int) bool {
		return repairDefectSortKey(diagnosis.Defects[i]) < repairDefectSortKey(diagnosis.Defects[j])
	})
	return diagnosis, nil
}

// PlanTaskGraphRepair reauthorizes convergent intent against graph. An edit that
// is already satisfied is skipped; a present edit must match a diagnosed defect
// exactly, preventing repair from becoming a generic constraint-deletion path.
func PlanTaskGraphRepair(graph *TaskGraph, request TaskGraphRepairRequest) (TaskGraphRepairPlan, error) {
	diagnosis, err := DiagnoseTaskGraphRepair(graph)
	if err != nil {
		return TaskGraphRepairPlan{}, err
	}
	selected := make(map[string]TaskGraphRepairOperation)
	selections := make(map[string]TaskGraphSourceEdit)
	deferredEmptyFields := make([]TaskGraphSourceEdit, 0)
	if request.Auto {
		for _, defect := range diagnosis.Defects {
			if defect.Repairable && defect.Automatic {
				selected[sourceEditKey(defect.Target)] = TaskGraphRepairOperation{
					Edit: defect.Target, Reason: defect.Reason, Automatic: true,
				}
				selections[sourceEditKey(defect.Target)] = defect.Target
			}
		}
	}
	for _, requested := range request.Edits {
		if shapeErr := validateTaskGraphRepairEditShape(requested); shapeErr != nil {
			return TaskGraphRepairPlan{}, shapeErr
		}
		requested, sourcePresent, normalizeErr := normalizeRepairEditSource(graph, requested)
		if normalizeErr != nil {
			return TaskGraphRepairPlan{}, normalizeErr
		}
		if !sourcePresent {
			selections[sourceEditKey(requested)] = requested
			continue
		}
		selections[sourceEditKey(requested)] = requested
		if requested.Action == TaskGraphSourceDropEmptyField {
			deferredEmptyFields = append(deferredEmptyFields, requested)
			continue
		}
		matched, alreadySatisfied, matchErr := authorizeRequestedRepair(graph, diagnosis.Defects, requested)
		if matchErr != nil {
			return TaskGraphRepairPlan{}, matchErr
		}
		if alreadySatisfied {
			continue
		}
		selections[sourceEditKey(matched.Target)] = matched.Target
		selected[sourceEditKey(matched.Target)] = TaskGraphRepairOperation{Edit: matched.Target, Reason: matched.Reason}
	}
	if len(deferredEmptyFields) > 0 || request.Auto {
		preliminary := make([]TaskGraphSourceEdit, 0, len(selected))
		for _, operation := range selected {
			preliminary = append(preliminary, operation.Edit)
		}
		prospective, simulateErr := graph.SimulateSourceEdits(preliminary)
		if simulateErr != nil {
			return TaskGraphRepairPlan{}, simulateErr
		}
		prospectiveDiagnosis, diagnoseErr := DiagnoseTaskGraphRepair(prospective)
		if diagnoseErr != nil {
			return TaskGraphRepairPlan{}, diagnoseErr
		}
		for _, requested := range deferredEmptyFields {
			matched, alreadySatisfied, matchErr := authorizeRequestedRepair(prospective, prospectiveDiagnosis.Defects, requested)
			if matchErr != nil {
				return TaskGraphRepairPlan{}, matchErr
			}
			if !alreadySatisfied {
				selected[sourceEditKey(matched.Target)] = TaskGraphRepairOperation{Edit: matched.Target, Reason: matched.Reason}
			}
		}
		if request.Auto {
			for _, defect := range prospectiveDiagnosis.Defects {
				if defect.Reason != RepairLegacyEmpty || !defect.Automatic {
					continue
				}
				selected[sourceEditKey(defect.Target)] = TaskGraphRepairOperation{Edit: defect.Target, Reason: defect.Reason, Automatic: true}
				selections[sourceEditKey(defect.Target)] = defect.Target
			}
		}
	}
	plan := TaskGraphRepairPlan{
		Selections: make([]TaskGraphSourceEdit, 0, len(selections)),
		Operations: make([]TaskGraphRepairOperation, 0, len(selected)),
	}
	for _, selection := range selections {
		plan.Selections = append(plan.Selections, selection)
	}
	sort.Slice(plan.Selections, func(i, j int) bool { return sourceEditKey(plan.Selections[i]) < sourceEditKey(plan.Selections[j]) })
	for _, operation := range selected {
		plan.Operations = append(plan.Operations, operation)
	}
	sort.Slice(plan.Operations, func(i, j int) bool {
		return sourceEditKey(plan.Operations[i].Edit) < sourceEditKey(plan.Operations[j].Edit)
	})
	return plan, nil
}

func validateTaskGraphRepairEditShape(edit TaskGraphSourceEdit) error {
	if !edit.Field.valid() {
		return fmt.Errorf("%w: unsupported graph-owned field %q", domain.ErrValidation, edit.Field)
	}
	switch edit.Action {
	case TaskGraphSourceDropDeclaration:
		if edit.Occurrence < 0 {
			return fmt.Errorf("%w: declaration occurrence must be non-negative", domain.ErrValidation)
		}
	case TaskGraphSourceDedupe:
		if edit.Field != TaskDependencyDependsOn {
			return fmt.Errorf("%w: dedupe is supported only for canonical depends_on declarations", domain.ErrValidation)
		}
		if edit.Occurrence != 0 {
			return fmt.Errorf("%w: dedupe occurrence must be zero", domain.ErrValidation)
		}
	case TaskGraphSourceDropEmptyField:
		if edit.Field == TaskDependencyDependsOn {
			return fmt.Errorf("%w: empty-field removal is supported only for legacy dependency fields", domain.ErrValidation)
		}
		if edit.Value != "" || edit.Occurrence != 0 {
			return fmt.Errorf("%w: empty-field removal cannot select a value or occurrence", domain.ErrValidation)
		}
	default:
		return fmt.Errorf("%w: unsupported repair action %q", domain.ErrValidation, edit.Action)
	}
	return nil
}

// ValidateTaskGraphRepairPlan proves source containment, authorization, and
// componentwise monotone progress independently. Unlike ordinary mutation
// validation it deliberately accepts GraphBroken before and after.
func ValidateTaskGraphRepairPlan(graph *TaskGraph, plan TaskGraphRepairPlan) (TaskGraphRepairPlan, TaskGraphRepairAnalysis, error) {
	before, err := DiagnoseTaskGraphRepair(graph)
	if err != nil {
		return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, err
	}
	// Re-plan the normalized operations as explicit intent. This prevents callers
	// implementing the port directly from smuggling an unauthorized source edit.
	requested := make([]TaskGraphSourceEdit, len(plan.Operations))
	for i, operation := range plan.Operations {
		requested[i] = operation.Edit
	}
	authorized, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Edits: requested})
	if err != nil {
		return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, err
	}
	if len(authorized.Operations) != len(plan.Operations) {
		return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, fmt.Errorf("%w: repair plan contains an already-satisfied or duplicate operation", domain.ErrValidation)
	}
	requestedSelections := append([]TaskGraphSourceEdit(nil), plan.Selections...)
	if len(requestedSelections) == 0 {
		for _, operation := range authorized.Operations {
			requestedSelections = append(requestedSelections, operation.Edit)
		}
	}
	selectionPlan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Edits: requestedSelections})
	if err != nil {
		return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, fmt.Errorf("%w: invalid selected repair intent: %v", domain.ErrValidation, err)
	}
	if !sameRepairOperationSet(authorized.Operations, selectionPlan.Operations) {
		return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, fmt.Errorf("%w: selected repair intent does not exactly account for the active operations", domain.ErrValidation)
	}
	authorizedReasons := make(map[string]TaskGraphRepairReason, len(authorized.Operations))
	for _, operation := range authorized.Operations {
		authorizedReasons[sourceEditKey(operation.Edit)] = operation.Reason
	}
	claimedAutomatic := make(map[string]bool, len(plan.Operations))
	for _, operation := range plan.Operations {
		key := sourceEditKey(operation.Edit)
		if operation.Automatic && !automaticRepairReason(authorizedReasons[key]) {
			return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, fmt.Errorf("%w: repair operation %s is not safe for automatic selection", domain.ErrValidation, repairEditDisplay(operation.Edit))
		}
		claimedAutomatic[key] = operation.Automatic
	}
	for index := range authorized.Operations {
		authorized.Operations[index].Automatic = claimedAutomatic[sourceEditKey(authorized.Operations[index].Edit)]
	}
	authorized.Selections = selectionPlan.Selections
	plan = authorized
	if len(plan.Operations) == 0 {
		return plan, TaskGraphRepairAnalysis{beforeGraph: graph, Before: before, After: before, Prospective: graph}, nil
	}
	if err := rejectOverlappingRepairOperations(plan.Operations); err != nil {
		return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, err
	}
	groups := groupRepairOperations(plan.Operations)
	current := graph
	previousMeasure, err := measureTaskGraphRepair(current)
	if err != nil {
		return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, err
	}
	for _, group := range groups {
		next, simulateErr := current.SimulateSourceEdits(group.Edits)
		if simulateErr != nil {
			return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, simulateErr
		}
		nextMeasure, measureErr := measureTaskGraphRepair(next)
		if measureErr != nil {
			return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, measureErr
		}
		if !repairMeasureComponentwiseNonIncreasing(previousMeasure, nextMeasure) || previousMeasure == nextMeasure {
			return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, fmt.Errorf("%w: repair of %s does not strictly improve its durable source-file prefix", domain.ErrValidation, displayRepairSource(group.Source))
		}
		current, previousMeasure = next, nextMeasure
	}
	removed, err := selectedRepairDeclarations(graph, plan.Operations)
	if err != nil {
		return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, err
	}
	if err := validateRepairPreservation(graph, current, plan.Operations, removed); err != nil {
		return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, err
	}
	after, err := DiagnoseTaskGraphRepair(current)
	if err != nil {
		return TaskGraphRepairPlan{}, TaskGraphRepairAnalysis{}, err
	}
	analysis := TaskGraphRepairAnalysis{
		beforeGraph: graph, Before: before, After: after, Prospective: current, Removed: removed,
		SourceGroups: groups, Impacts: taskGraphRepairImpacts(graph, current, groups),
	}
	return plan, analysis, nil
}

func rejectOverlappingRepairOperations(operations []TaskGraphRepairOperation) error {
	type declarationGroup struct {
		source TaskGraphSourceRef
		field  TaskDependencyField
		value  string
	}
	actions := make(map[declarationGroup]map[TaskGraphSourceEditAction]bool)
	for _, operation := range operations {
		if operation.Edit.Action == TaskGraphSourceDropEmptyField {
			continue
		}
		key := declarationGroup{source: operation.Edit.Source, field: operation.Edit.Field, value: operation.Edit.Value}
		if actions[key] == nil {
			actions[key] = make(map[TaskGraphSourceEditAction]bool)
		}
		actions[key][operation.Edit.Action] = true
		if actions[key][TaskGraphSourceDedupe] && actions[key][TaskGraphSourceDropDeclaration] {
			return fmt.Errorf("%w: dedupe and exact-drop repair intents overlap for %s:%s=%q; apply one intent, then re-diagnose", domain.ErrValidation, displayRepairSource(key.source), key.field, key.value)
		}
	}
	return nil
}

func automaticRepairReason(reason TaskGraphRepairReason) bool {
	return reason == RepairDuplicate || reason == RepairSelf || reason == RepairLegacyEmpty
}

func sameRepairOperationSet(left, right []TaskGraphRepairOperation) bool {
	if len(left) != len(right) {
		return false
	}
	leftKeys := make([]string, len(left))
	rightKeys := make([]string, len(right))
	for index := range left {
		leftKeys[index] = string(left[index].Reason) + "\x00" + sourceEditKey(left[index].Edit)
	}
	for index := range right {
		rightKeys[index] = string(right[index].Reason) + "\x00" + sourceEditKey(right[index].Edit)
	}
	sort.Strings(leftKeys)
	sort.Strings(rightKeys)
	return reflect.DeepEqual(leftKeys, rightKeys)
}

func repairEditDisplay(edit TaskGraphSourceEdit) string {
	return fmt.Sprintf("%s:%s=%q#%d", displayRepairSource(edit.Source), edit.Field, edit.Value, edit.Occurrence)
}

// InspectTaskGraphRepair performs the bare, read-only diagnosis. Mutation
// selectors use RepairTaskGraph so their authorization happens under the store's
// repository guard instead of racing this preview.
func (s *Service) InspectTaskGraphRepair() (TaskGraphRepairReceipt, error) {
	var threadProblems []ThreadReadProblem
	if s.threads != nil {
		read, readErr := s.threads.ReadThreads()
		if readErr != nil {
			return TaskGraphRepairReceipt{}, readErr
		}
		threadProblems = read.Problems
	}
	graph, err := LoadTaskGraph(s.taskGraphs)
	if err != nil {
		return TaskGraphRepairReceipt{}, err
	}
	diagnosis, err := DiagnoseTaskGraphRepair(graph)
	if err != nil {
		return TaskGraphRepairReceipt{}, err
	}
	receipt := TaskGraphRepairReceipt{
		InitialHealth: diagnosis.Health, FinalHealth: diagnosis.Health,
		Residual: cloneRepairDefects(diagnosis.Defects), Problems: graph.Problems(),
		IncompleteThreads: cloneThreadReadProblemsPublic(threadProblems),
	}
	return receipt, nil
}

// RepairTaskGraph applies source-declaration repair intent through the only
// capability permitted to accept GraphBroken. Pre-write conflicts are retried;
// a durable prefix is always returned immediately for operator inspection.
func (s *Service) RepairTaskGraph(request TaskGraphRepairRequest, dryRun bool) (TaskGraphRepairReceipt, error) {
	if s.graphRepairs == nil {
		return TaskGraphRepairReceipt{}, fmt.Errorf("task graph repair is unavailable from this store")
	}
	now := s.now()
	var receipt TaskGraphRepairReceipt
	for attempt := 0; ; attempt++ {
		result, err := s.graphRepairs.MutateTaskGraphRepair(now, dryRun, func(graph *TaskGraph) (TaskGraphRepairPlan, error) {
			return PlanTaskGraphRepair(graph, request)
		})
		receipt = taskGraphRepairReceipt(result)
		if dryRun || !errors.Is(err, domain.ErrConflict) || len(result.AppliedSources) > 0 || attempt >= s.maxRetries {
			if err != nil {
				return receipt, &TaskGraphRepairFailure{Cause: err, Receipt: receipt}
			}
			return receipt, nil
		}
		s.retrySleep(attempt + 1)
	}
}

func taskGraphRepairReceipt(result TaskGraphRepairMutationResult) TaskGraphRepairReceipt {
	finalGraph := result.FinalGraph
	if finalGraph == nil {
		finalGraph = result.Analysis.Prospective
	}
	finalDiagnosis := result.Analysis.After
	removed := result.Analysis.Removed
	impacts := result.Analysis.Impacts
	threadImpacts := []ThreadProjectionImpact(nil)
	if finalGraph != nil && result.Analysis.Prospective != nil && finalGraph != result.Analysis.Prospective {
		if diagnosis, err := DiagnoseTaskGraphRepair(finalGraph); err == nil {
			finalDiagnosis = diagnosis
		}
		if result.Analysis.Prospective != nil {
			// Partial-prefix analysis is populated by the store when possible. This
			// fallback intentionally leaves full-plan impact out rather than lying.
			impacts = nil
		}
	}
	if finalGraph != nil && result.Analysis.Prospective != nil {
		directSources := result.AppliedSources
		if result.DryRun {
			directSources = result.PlannedSources
		}
		direct := make([]string, 0, len(directSources))
		for _, source := range directSources {
			direct = append(direct, source.TaskID)
		}
		threadImpacts = TaskGraphThreadImpacts(result.Threads, repairBeforeGraph(result), finalGraph, direct)
	}
	receipt := TaskGraphRepairReceipt{
		Changed: result.Changed, DryRun: result.DryRun, Committed: result.Committed,
		InitialHealth: result.Analysis.Before.Health, FinalHealth: finalDiagnosis.Health,
		Selected:   append([]TaskGraphSourceEdit(nil), result.Plan.Selections...),
		Operations: append([]TaskGraphRepairOperation(nil), result.Plan.Operations...),
		Removed:    append([]TaskGraphSourceDeclaration(nil), removed...),
		Addressed:  repairDefectDifference(result.Analysis.Before.Defects, finalDiagnosis.Defects),
		Residual:   cloneRepairDefects(finalDiagnosis.Defects), Problems: append([]GraphProblem(nil), finalDiagnosis.Problems...),
		Impacts: cloneTaskGraphStateImpacts(impacts), ThreadImpacts: cloneThreadProjectionImpacts(threadImpacts),
		IncompleteThreads: cloneThreadReadProblemsPublic(result.ThreadProblems),
	}
	for _, source := range result.PlannedSources {
		receipt.PlannedFiles = append(receipt.PlannedFiles, displayRepairSource(source))
	}
	for _, source := range result.AppliedSources {
		receipt.AppliedFiles = append(receipt.AppliedFiles, displayRepairSource(source))
	}
	for _, source := range result.RemainingSources {
		receipt.RemainingFiles = append(receipt.RemainingFiles, displayRepairSource(source))
	}
	return receipt
}

func repairBeforeGraph(result TaskGraphRepairMutationResult) *TaskGraph {
	// SourceGroups can be reversed from the fully prospective graph only by
	// inventing declarations, which repair forbids. The store supplies prefix
	// impacts directly; Thread comparison uses the original graph captured in the
	// analysis through a private field added below.
	return result.Analysis.beforeGraph
}

func cloneRepairDefects(values []TaskGraphRepairDefect) []TaskGraphRepairDefect {
	out := append([]TaskGraphRepairDefect(nil), values...)
	for i := range out {
		out[i].CandidateIDs = append([]string(nil), values[i].CandidateIDs...)
		out[i].Problem.Cycle = append([]string(nil), values[i].Problem.Cycle...)
	}
	return out
}

func cloneThreadReadProblemsPublic(values []ThreadReadProblem) []ThreadReadProblem {
	out := append([]ThreadReadProblem(nil), values...)
	for i := range out {
		out[i].SourceVersion = ""
	}
	return out
}

func repairDefectDifference(before, after []TaskGraphRepairDefect) []TaskGraphRepairDefect {
	remaining := make(map[string]int, len(after))
	for _, defect := range after {
		remaining[repairDefectIdentity(defect)]++
	}
	addressed := make([]TaskGraphRepairDefect, 0)
	for _, defect := range before {
		key := repairDefectIdentity(defect)
		if remaining[key] > 0 {
			remaining[key]--
			continue
		}
		addressed = append(addressed, defect)
	}
	return cloneRepairDefects(addressed)
}

func repairDefectIdentity(defect TaskGraphRepairDefect) string {
	if defect.Repairable {
		return string(defect.Reason) + "\x00" + sourceEditKey(defect.Target)
	}
	problem := defect.Problem
	return strings.Join([]string{string(defect.Reason), string(problem.Code), problem.TaskID, problem.RelatedTaskID, problem.Field, problem.Path, problem.Message}, "\x00")
}

func authorizeRequestedRepair(graph *TaskGraph, defects []TaskGraphRepairDefect, requested TaskGraphSourceEdit) (TaskGraphRepairDefect, bool, error) {
	if requested.Action != TaskGraphSourceDropDeclaration && requested.Action != TaskGraphSourceDedupe && requested.Action != TaskGraphSourceDropEmptyField {
		return TaskGraphRepairDefect{}, false, fmt.Errorf("%w: unsupported repair action %q", domain.ErrValidation, requested.Action)
	}
	for _, defect := range defects {
		if sameRepairEditIntent(defect.Target, requested) {
			return defect, false, nil
		}
	}
	// Convergent retries may name an exact source/value that disappeared in an
	// earlier durable prefix. A still-present non-defect is never treated as done.
	present, err := repairEditTargetPresent(graph, requested)
	if err != nil {
		return TaskGraphRepairDefect{}, false, err
	}
	if !present {
		return TaskGraphRepairDefect{}, true, nil
	}
	return TaskGraphRepairDefect{}, false, fmt.Errorf("%w: selected declaration is not a repairable defect; use ordinary dependency removal after graph repair", domain.ErrValidation)
}

func repairEditTargetPresent(graph *TaskGraph, requested TaskGraphSourceEdit) (bool, error) {
	records, err := graph.SourceRecords()
	if err != nil {
		return false, err
	}
	for _, record := range records {
		if !sameSourceRef(record.Source, requested.Source) {
			continue
		}
		for _, field := range record.Fields {
			if field.Field != requested.Field {
				continue
			}
			switch requested.Action {
			case TaskGraphSourceDropEmptyField:
				return true, nil
			case TaskGraphSourceDedupe:
				count := 0
				for _, value := range field.Values {
					if value == requested.Value {
						count++
					}
				}
				return count > 1, nil
			case TaskGraphSourceDropDeclaration:
				occurrence := 0
				for _, value := range field.Values {
					if value != requested.Value {
						continue
					}
					if occurrence == requested.Occurrence {
						return true, nil
					}
					occurrence++
				}
				return false, nil
			}
		}
		return false, nil
	}
	return false, nil
}

func normalizeRepairEditSource(graph *TaskGraph, edit TaskGraphSourceEdit) (TaskGraphSourceEdit, bool, error) {
	if edit.Source.TaskID == "" && edit.Source.TaskSlug == "" && edit.Source.Location == "" {
		return TaskGraphSourceEdit{}, false, fmt.Errorf("%w: repair source identity is required", domain.ErrValidation)
	}
	records, err := graph.SourceRecords()
	if err != nil {
		return TaskGraphSourceEdit{}, false, err
	}
	matches := make([]TaskGraphSourceRef, 0, 1)
	for _, record := range records {
		if (edit.Source.TaskID != "" && record.Source.TaskID != edit.Source.TaskID) ||
			(edit.Source.TaskSlug != "" && record.Source.TaskSlug != edit.Source.TaskSlug) ||
			(edit.Source.Location != "" && record.Source.Location != edit.Source.Location) {
			continue
		}
		matches = append(matches, record.Source)
	}
	switch len(matches) {
	case 0:
		return TaskGraphSourceEdit{}, false, fmt.Errorf("%w: repair source %+v does not identify a readable task record", domain.ErrNotFound, edit.Source)
	case 1:
		edit.Source = matches[0]
		return edit, true, nil
	default:
		return TaskGraphSourceEdit{}, false, fmt.Errorf("%w: repair source %+v matches %d task records; use its exact location", domain.ErrAmbiguous, edit.Source, len(matches))
	}
}

func measureTaskGraphRepair(graph *TaskGraph) (taskGraphRepairMeasure, error) {
	diagnosis, err := DiagnoseTaskGraphRepair(graph)
	if err != nil {
		return taskGraphRepairMeasure{}, err
	}
	measure := taskGraphRepairMeasure{}
	for _, problem := range graph.problems {
		switch problem.Code {
		case ProblemUnreadable:
			measure.unreadable++
		case ProblemMissingTaskID, ProblemTaskIDDrift, ProblemDuplicateTaskID, ProblemInvalidStatus:
			measure.identity++
		}
	}
	declarations, _ := graph.SourceDeclarations()
	seenEdges := make(map[DependencyEdge]bool)
	for _, declaration := range declarations {
		if declaration.HasProjectedEdge && graph.edgeInCycle(declaration.ProjectedEdge) && !seenEdges[declaration.ProjectedEdge] {
			seenEdges[declaration.ProjectedEdge] = true
			measure.cyclicEdges++
		}
		if declaration.Field != TaskDependencyDependsOn {
			measure.legacyPresent++
		}
	}
	for _, record := range graph.sourceRecords() {
		for _, field := range record.Fields {
			if field.Field != TaskDependencyDependsOn && len(field.Values) == 0 {
				measure.legacyPresent++
			}
		}
	}
	for _, defect := range diagnosis.Defects {
		switch defect.Reason {
		case RepairSelf:
			measure.self++
		case RepairLegacyUnsafe:
			measure.legacyUnsafe++
		case RepairDuplicate, RepairInvalidID, RepairDangling, RepairLegacyMissing, RepairLegacyAmbiguous:
			measure.declarations++
		}
	}
	return measure, nil
}

func repairMeasureComponentwiseNonIncreasing(before, after taskGraphRepairMeasure) bool {
	return after.unreadable <= before.unreadable && after.identity <= before.identity &&
		after.self <= before.self && after.cyclicEdges <= before.cyclicEdges && after.legacyUnsafe <= before.legacyUnsafe &&
		after.declarations <= before.declarations && after.legacyPresent <= before.legacyPresent
}

func groupRepairOperations(operations []TaskGraphRepairOperation) []TaskGraphRepairSourceGroup {
	groups := make([]TaskGraphRepairSourceGroup, 0)
	for _, operation := range operations {
		if len(groups) == 0 || !sameSourceRef(groups[len(groups)-1].Source, operation.Edit.Source) {
			groups = append(groups, TaskGraphRepairSourceGroup{Source: operation.Edit.Source})
		}
		groups[len(groups)-1].Edits = append(groups[len(groups)-1].Edits, operation.Edit)
	}
	return groups
}

func taskGraphRepairImpacts(before, after *TaskGraph, groups []TaskGraphRepairSourceGroup) []TaskGraphStateImpact {
	direct := make(map[string]bool)
	for _, group := range groups {
		direct[group.Source.TaskID] = true
	}
	ids := sortedUnique(append(before.TaskIDs(), after.TaskIDs()...))
	impacts := make([]TaskGraphStateImpact, 0)
	for _, taskID := range ids {
		left, right := before.State(taskID), after.State(taskID)
		if left == right {
			continue
		}
		impacts = append(impacts, TaskGraphStateImpact{TaskID: taskID, Before: left, After: right, Direct: direct[taskID]})
	}
	return impacts
}

func selectedRepairDeclarations(before *TaskGraph, operations []TaskGraphRepairOperation) ([]TaskGraphSourceDeclaration, error) {
	declarations, err := before.SourceDeclarations()
	if err != nil {
		return nil, err
	}
	removed := make([]TaskGraphSourceDeclaration, 0)
	for _, operation := range operations {
		matches := make([]TaskGraphSourceDeclaration, 0)
		for _, declaration := range declarations {
			if sameSourceRef(declaration.Source, operation.Edit.Source) && declaration.Field == operation.Edit.Field && declaration.Value == operation.Edit.Value {
				matches = append(matches, declaration)
			}
		}
		switch operation.Edit.Action {
		case TaskGraphSourceDropDeclaration:
			for _, declaration := range matches {
				if declaration.Occurrence == operation.Edit.Occurrence {
					removed = append(removed, declaration)
					break
				}
			}
		case TaskGraphSourceDedupe:
			removed = append(removed, matches[1:]...)
		}
	}
	sort.SliceStable(removed, func(i, j int) bool {
		return sourceDeclarationSortKey(removed[i]) < sourceDeclarationSortKey(removed[j])
	})
	return removed, nil
}

func validateRepairPreservation(before, after *TaskGraph, operations []TaskGraphRepairOperation, removed []TaskGraphSourceDeclaration) error {
	allowed := make(map[string]int)
	beforeDeclarations, _ := before.SourceDeclarations()
	for _, operation := range operations {
		switch operation.Edit.Action {
		case TaskGraphSourceDropDeclaration:
			for _, declaration := range beforeDeclarations {
				if sameRepairEditIntent(operation.Edit, sourceEditForDeclaration(declaration)) {
					allowed[sourceDeclarationSemanticKey(declaration)]++
					break
				}
			}
		case TaskGraphSourceDedupe:
			matches := make([]TaskGraphSourceDeclaration, 0)
			for _, declaration := range beforeDeclarations {
				if sameSourceRef(declaration.Source, operation.Edit.Source) && declaration.Field == operation.Edit.Field && declaration.Value == operation.Edit.Value {
					matches = append(matches, declaration)
				}
			}
			for _, declaration := range matches[1:] {
				allowed[sourceDeclarationSemanticKey(declaration)]++
			}
		}
	}
	for _, declaration := range removed {
		key := sourceDeclarationSemanticKey(declaration)
		if allowed[key] == 0 {
			return fmt.Errorf("%w: repair would remove unrelated declaration %s", domain.ErrValidation, displaySourceDeclaration(declaration))
		}
		allowed[key]--
	}
	for key, remaining := range allowed {
		if remaining != 0 {
			return fmt.Errorf("%w: selected repair did not remove its authorized declaration %s", domain.ErrValidation, key)
		}
	}
	if err := validateRepairSourceRecordDifference(before, after, operations); err != nil {
		return err
	}
	if !reflect.DeepEqual(before.loadProblems, after.loadProblems) {
		return fmt.Errorf("%w: repair changed unreadable-source evidence", domain.ErrValidation)
	}
	return nil
}

func validateRepairSourceRecordDifference(before, after *TaskGraph, operations []TaskGraphRepairOperation) error {
	beforeRecords, err := before.SourceRecords()
	if err != nil {
		return err
	}
	afterRecords, err := after.SourceRecords()
	if err != nil {
		return err
	}
	beforeBySource := groupRepairSourceRecords(beforeRecords)
	afterBySource := groupRepairSourceRecords(afterRecords)
	if len(beforeRecords) != len(afterRecords) || len(beforeBySource) != len(afterBySource) {
		return fmt.Errorf("%w: repair changed readable task-record identity", domain.ErrValidation)
	}
	operationsBySource := make(map[TaskGraphSourceRef][]TaskGraphRepairOperation)
	for _, operation := range operations {
		operationsBySource[operation.Edit.Source] = append(operationsBySource[operation.Edit.Source], operation)
	}
	for source, beforeGroup := range beforeBySource {
		afterGroup, exists := afterBySource[source]
		if !exists || len(beforeGroup) != len(afterGroup) {
			return fmt.Errorf("%w: repair changed readable source identity %s", domain.ErrValidation, displayRepairSource(source))
		}
		sourceOperations := operationsBySource[source]
		if len(sourceOperations) == 0 {
			if !sameRepairSourceRecordMultiset(beforeGroup, afterGroup) {
				return fmt.Errorf("%w: repair changed unselected source %s", domain.ErrValidation, displayRepairSource(source))
			}
			continue
		}
		if len(beforeGroup) != 1 {
			return fmt.Errorf("%w: selected repair source %s is not a unique physical record", domain.ErrValidation, displayRepairSource(source))
		}
		if err := validateRepairRecordDifference(beforeGroup[0], afterGroup[0], sourceOperations); err != nil {
			return err
		}
	}
	return nil
}

func groupRepairSourceRecords(records []TaskGraphSourceRecord) map[TaskGraphSourceRef][]TaskGraphSourceRecord {
	grouped := make(map[TaskGraphSourceRef][]TaskGraphSourceRecord)
	for _, record := range records {
		grouped[record.Source] = append(grouped[record.Source], record)
	}
	return grouped
}

func sameRepairSourceRecordMultiset(left, right []TaskGraphSourceRecord) bool {
	if len(left) != len(right) {
		return false
	}
	leftKeys := make([]string, len(left))
	rightKeys := make([]string, len(right))
	for index := range left {
		leftKeys[index] = repairSourceFieldsKey(left[index].Fields)
	}
	for index := range right {
		rightKeys[index] = repairSourceFieldsKey(right[index].Fields)
	}
	sort.Strings(leftKeys)
	sort.Strings(rightKeys)
	return reflect.DeepEqual(leftKeys, rightKeys)
}

func repairSourceFieldsKey(fields []TaskGraphSourceField) string {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		parts = append(parts, string(field.Field)+"\x00"+strings.Join(field.Values, "\x00"))
	}
	return strings.Join(parts, "\x01")
}

func validateRepairRecordDifference(before, after TaskGraphSourceRecord, operations []TaskGraphRepairOperation) error {
	beforeFields := repairSourceFieldsByName(before.Fields)
	afterFields := repairSourceFieldsByName(after.Fields)
	operationsByField := make(map[TaskDependencyField][]TaskGraphRepairOperation)
	for _, operation := range operations {
		operationsByField[operation.Edit.Field] = append(operationsByField[operation.Edit.Field], operation)
	}
	for _, field := range []TaskDependencyField{
		TaskDependencyDependsOn, TaskDependencyBlockedBy, TaskDependencyDependencies, TaskDependencyBlocks,
	} {
		beforeValues, beforePresent := beforeFields[field]
		afterValues, afterPresent := afterFields[field]
		fieldOperations := operationsByField[field]
		if len(fieldOperations) == 0 {
			if beforePresent != afterPresent || !reflect.DeepEqual(beforeValues, afterValues) {
				return fmt.Errorf("%w: repair of %s changed unselected field %s", domain.ErrValidation, displayRepairSource(before.Source), field)
			}
			continue
		}
		expectedCounts := repairValueCounts(beforeValues)
		dropEmpty := false
		for _, operation := range fieldOperations {
			switch operation.Edit.Action {
			case TaskGraphSourceDropDeclaration:
				if expectedCounts[operation.Edit.Value] == 0 {
					return fmt.Errorf("%w: repair of %s did not start with selected %s declaration", domain.ErrValidation, displayRepairSource(before.Source), field)
				}
				expectedCounts[operation.Edit.Value]--
				if expectedCounts[operation.Edit.Value] == 0 {
					delete(expectedCounts, operation.Edit.Value)
				}
			case TaskGraphSourceDedupe:
				expectedCounts[operation.Edit.Value] = 1
			case TaskGraphSourceDropEmptyField:
				dropEmpty = true
			}
		}
		if !reflect.DeepEqual(expectedCounts, repairValueCounts(afterValues)) || !repairValuesAreSubsequence(beforeValues, afterValues) {
			return fmt.Errorf("%w: repair of %s removed, added, or reordered an unauthorized %s declaration", domain.ErrValidation, displayRepairSource(before.Source), field)
		}
		expectedPresent := beforePresent
		if dropEmpty {
			expectedPresent = false
		} else if field == TaskDependencyDependsOn && len(afterValues) == 0 {
			expectedPresent = false
		}
		if afterPresent != expectedPresent {
			return fmt.Errorf("%w: repair of %s changed field presence for %s without authorization", domain.ErrValidation, displayRepairSource(before.Source), field)
		}
	}
	return nil
}

func repairSourceFieldsByName(fields []TaskGraphSourceField) map[TaskDependencyField][]string {
	indexed := make(map[TaskDependencyField][]string, len(fields))
	for _, field := range fields {
		indexed[field.Field] = append([]string(nil), field.Values...)
	}
	return indexed
}

func repairValueCounts(values []string) map[string]int {
	counts := make(map[string]int)
	for _, value := range values {
		counts[value]++
	}
	return counts
}

func repairValuesAreSubsequence(before, after []string) bool {
	next := 0
	for _, value := range before {
		if next < len(after) && value == after[next] {
			next++
		}
	}
	return next == len(after)
}

func legacyReferenceIndex(diagnostics []LegacyDependencyDiagnostic) map[string]LegacyReference {
	index := make(map[string]LegacyReference)
	for _, diagnostic := range diagnostics {
		occurrences := make(map[string]int)
		for _, ref := range diagnostic.References {
			source := TaskGraphSourceRef{TaskID: diagnostic.TaskID, TaskSlug: diagnostic.TaskSlug, Location: diagnostic.TaskPath}
			occurrence := occurrences[ref.Value]
			occurrences[ref.Value]++
			index[sourceFieldOccurrenceKey(source, TaskDependencyField(diagnostic.Field), ref.Value, occurrence)] = ref
		}
	}
	return index
}

func (g *TaskGraph) hasKnownOrUnreadableTask(taskID string) bool {
	return taskExists(g.tasks, taskID) || g.unreadableIDs[taskID]
}

func (g *TaskGraph) edgeInCycle(edge DependencyEdge) bool {
	from, fromOK := g.cycleComponent[edge.From]
	to, toOK := g.cycleComponent[edge.To]
	return fromOK && toOK && from == to
}

// SameRepairSnapshot reports whether actual is the exact authoritative source
// state represented by a prospective removal-only prefix. Revisions and the
// repair-owned updated_at stamp may differ only for sources the prefix changed;
// every other readable record and every unreadable revision must remain exact.
func (g *TaskGraph) SameRepairSnapshot(actual *TaskGraph, changed []TaskGraphSourceRef) bool {
	if g == nil || actual == nil || g.requireCompleteSource() != nil || actual.requireCompleteSource() != nil {
		return false
	}
	if len(g.sourceTasks) != len(actual.sourceTasks) || len(g.loadProblems) != len(actual.loadProblems) {
		return false
	}
	expectedTasks, actualTasks := cloneTasks(g.sourceTasks), cloneTasks(actual.sourceTasks)
	sort.SliceStable(expectedTasks, func(i, j int) bool { return taskSourceSortKey(expectedTasks[i]) < taskSourceSortKey(expectedTasks[j]) })
	sort.SliceStable(actualTasks, func(i, j int) bool { return taskSourceSortKey(actualTasks[i]) < taskSourceSortKey(actualTasks[j]) })
	changedSet := make(map[TaskGraphSourceRef]bool, len(changed))
	for _, source := range changed {
		changedSet[source] = true
	}
	for i := range expectedTasks {
		if sourceRefForTask(expectedTasks[i]) != sourceRefForTask(actualTasks[i]) {
			return false
		}
		if changedSet[sourceRefForTask(expectedTasks[i])] {
			expectedTasks[i].SourceVersion, actualTasks[i].SourceVersion = "", ""
			expectedTasks[i].Updated, actualTasks[i].Updated = "", ""
		}
		if !reflect.DeepEqual(expectedTasks[i], actualTasks[i]) {
			return false
		}
	}
	for i := range g.loadProblems {
		if !sameTaskGraphLoadProblem(g.loadProblems[i], actual.loadProblems[i]) {
			return false
		}
	}
	return true
}

func repairProblemHasDeclarationCandidate(code GraphProblemCode) bool {
	switch code {
	case ProblemDuplicateDependency, ProblemSelfDependency, ProblemInvalidDependencyID,
		ProblemMissingDependency, ProblemCycle, ProblemLegacyMissing, ProblemLegacyAmbiguous:
		return true
	default:
		return false
	}
}

func sourceEditForDeclaration(declaration TaskGraphSourceDeclaration) TaskGraphSourceEdit {
	return TaskGraphSourceEdit{Action: TaskGraphSourceDropDeclaration, Source: declaration.Source, Field: declaration.Field, Value: declaration.Value, Occurrence: declaration.Occurrence}
}

func sameRepairEditIntent(left, right TaskGraphSourceEdit) bool {
	if left.Action != right.Action || !sameSourceRef(left.Source, right.Source) || left.Field != right.Field || left.Value != right.Value {
		return false
	}
	return left.Action == TaskGraphSourceDedupe || left.Action == TaskGraphSourceDropEmptyField || left.Occurrence == right.Occurrence
}

func sameSourceRef(left, right TaskGraphSourceRef) bool {
	return left.TaskID == right.TaskID && left.TaskSlug == right.TaskSlug && left.Location == right.Location
}

func sourceFieldValueKey(source TaskGraphSourceRef, field TaskDependencyField, value string) string {
	return strings.Join([]string{source.TaskID, source.TaskSlug, source.Location, string(field), value}, "\x00")
}

func sourceFieldOccurrenceKey(source TaskGraphSourceRef, field TaskDependencyField, value string, occurrence int) string {
	return fmt.Sprintf("%s\x00%d", sourceFieldValueKey(source, field, value), occurrence)
}

func sourceEditKey(edit TaskGraphSourceEdit) string {
	return fmt.Sprintf("%s\x00%s\x00%d", sourceFieldValueKey(edit.Source, edit.Field, edit.Value), edit.Action, edit.Occurrence)
}

func sourceDeclarationSemanticKey(declaration TaskGraphSourceDeclaration) string {
	return sourceFieldOccurrenceKey(declaration.Source, declaration.Field, declaration.Value, declaration.Occurrence)
}

func repairDefectSortKey(defect TaskGraphRepairDefect) string {
	return fmt.Sprintf("%s\x00%s\x00%s", sourceEditKey(defect.Target), defect.Reason, defect.Problem.Message)
}

func displayRepairSource(source TaskGraphSourceRef) string {
	if source.Location != "" {
		return source.Location
	}
	if source.TaskID != "" {
		return source.TaskID
	}
	return source.TaskSlug
}

func displaySourceDeclaration(declaration TaskGraphSourceDeclaration) string {
	return fmt.Sprintf("%s:%s=%q#%d", displayRepairSource(declaration.Source), declaration.Field, declaration.Value, declaration.Occurrence)
}
