package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

// TaskLifecycleOverride names the one policy gate a caller deliberately bypasses.
// The CLI may spell both contextual choices --force, but core and persistence never
// carry an ambiguous boolean.
type TaskLifecycleOverride string

const (
	TaskLifecycleOverrideNone               TaskLifecycleOverride = ""
	TaskLifecycleOverrideDependencyGate     TaskLifecycleOverride = "dependency-gate"
	TaskLifecycleOverrideAcceptanceCriteria TaskLifecycleOverride = "acceptance-criteria"
)

// TaskGraphStateImpact is the shared before/after projection used by dependency
// and lifecycle receipts. Path is present for transitive lifecycle impact and
// empty for a directly mutated task.
type TaskGraphStateImpact struct {
	TaskID string
	Before TaskGraphState
	After  TaskGraphState
	Path   []string
	Direct bool
}

// TaskLifecycleCreation is the semantic seed document carried by the one guarded
// create-and-start operation. The seed uses ready-to-start as a canonical internal
// shape; it is never persisted in that state, and readiness is not an authorization
// prerequisite for starting an existing task.
type TaskLifecycleCreation struct {
	Task domain.Task
	Body string
}

// TaskLifecyclePlan is returned by a pure planner while the store owns the
// authoritative graph snapshot and repository guard. Existing transitions name
// TaskID. Create-and-start instead supplies Create and leaves TaskID empty.
type TaskLifecyclePlan struct {
	TaskID    string
	To        domain.Status
	Override  TaskLifecycleOverride
	RevisitAt string
	Create    *TaskLifecycleCreation
}

// TaskLifecycleAnalysis is the pure authorization and impact result for a
// validated plan. OutstandingBlockers is retained when a dependency override is
// actually required so a forced receipt remains explanatory.
type TaskLifecycleAnalysis struct {
	From                domain.Status
	Before              TaskGraphState
	After               TaskGraphState
	Impacts             []TaskGraphStateImpact
	OutstandingBlockers []Blocker
	OverrideApplied     bool
}

// TaskLifecycleMutationResult is the store-owned guarded operation result.
// Changed means the semantic task document differs; dry runs report the same
// result without a durable replacement.
type TaskLifecycleMutationResult struct {
	Plan                TaskLifecyclePlan
	Task                domain.Task
	From                domain.Status
	Before              TaskGraphState
	After               TaskGraphState
	Impacts             []TaskGraphStateImpact
	OutstandingBlockers []Blocker
	OverrideApplied     bool
	Changed             bool
	DryRun              bool
	// Committed distinguishes a durable write followed by a repository-guard
	// cleanup failure from an ordinary pre-commit failure. Callers must never
	// blindly retry a committed result.
	Committed bool
}

// TaskLifecycleReceipt is the adapter-neutral public result returned by Service.
type TaskLifecycleReceipt struct {
	Task                domain.Task
	From                domain.Status
	To                  domain.Status
	Changed             bool
	DryRun              bool
	Committed           bool
	Override            TaskLifecycleOverride
	Forced              bool
	Before              TaskGraphState
	After               TaskGraphState
	OutstandingBlockers []Blocker
	Impacts             []TaskGraphStateImpact
	Remedy              string
}

// TaskEligibilityError keeps an ineligible start machine-inspectable while its
// Error text remains a deterministic, actionable CLI diagnosis.
type TaskEligibilityError struct {
	TaskID            string
	State             TaskGraphState
	Blockers          []Blocker
	RequestedOverride TaskLifecycleOverride
}

func (e *TaskEligibilityError) Error() string {
	if e == nil {
		return "task is not eligible to start"
	}
	base := fmt.Sprintf("task %s is not eligible to start: role=%s gate=%s", e.TaskID, e.State.Role, e.State.Gate)
	if !isPendingWorkRole(e.State.Role) {
		return base + "; move it to next-up or ready-to-start before starting"
	}
	if len(e.Blockers) == 0 {
		return base
	}
	parts := make([]string, 0, len(e.Blockers))
	for _, blocker := range e.Blockers {
		parts = append(parts, fmt.Sprintf("%s (%s via %s)", blocker.TaskID, blocker.Reason, strings.Join(blocker.Path, " -> ")))
	}
	base += "; outstanding blockers: " + strings.Join(parts, ", ")
	if e.DependencyGateOverrideAllowed() {
		return base + "; pass --force to bypass only the blocked dependency gate"
	}
	return base + "; repair the broken dependency evidence before starting"
}

func (e *TaskEligibilityError) Unwrap() error { return domain.ErrValidation }

// Remedy is kept separate from Error so machine adapters can expose recovery
// without asking callers to parse prose.
func (e *TaskEligibilityError) Remedy() string {
	if e == nil {
		return "inspect the task's lifecycle and dependency state before retrying"
	}
	if !isPendingWorkRole(e.State.Role) {
		return "move the task to next-up or ready-to-start before starting"
	}
	if e.DependencyGateOverrideAllowed() {
		return "resolve the outstanding blockers, or pass --force to bypass only the dependency gate"
	}
	if e.State.Gate == GateBroken {
		return "repair the broken dependency evidence, then inspect the task before retrying"
	}
	return "inspect the task with `tskflwctl task blockers <task>` before retrying"
}

// DependencyGateOverrideAllowed reports whether the refusal is solely a blocked
// dependency gate that --force may bypass. A broken gate can still occur in a
// healthy repository when a prerequisite is withdrawn; force cannot make that
// dependency evidence sound.
func (e *TaskEligibilityError) DependencyGateOverrideAllowed() bool {
	return e != nil && isPendingWorkRole(e.State.Role) && e.State.Gate == GateBlocked
}

// TaskLifecycleMutationFailure reports a post-commit failure without erasing
// the durable outcome. Receipt.Committed is always true. Keeping the cause
// wrapped preserves error classification while adapters can render recovery
// guidance and, critically, retry loops can stop.
type TaskLifecycleMutationFailure struct {
	Cause   error
	Receipt TaskLifecycleReceipt
}

func (e *TaskLifecycleMutationFailure) Error() string {
	if e == nil {
		return "task lifecycle transition committed, but repository cleanup failed"
	}
	return fmt.Sprintf("task lifecycle transition committed, but repository cleanup failed: %v; inspect current task state before retrying", e.Cause)
}

func (e *TaskLifecycleMutationFailure) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// ValidateTaskLifecycleSource applies the fail-closed mutation rule. Unlike the
// dedicated legacy migration, lifecycle writes require a fully healthy canonical
// graph; degraded or broken input cannot authorize dispatch or impact receipts.
func ValidateTaskLifecycleSource(graph *TaskGraph) error {
	if graph == nil {
		return fmt.Errorf("%w: authoritative task graph is required", domain.ErrValidation)
	}
	if !graph.MutationReady() {
		return fmt.Errorf("%w: repository task graph is %s: %s",
			domain.ErrValidation, graph.Health(), taskGraphHealthDetail(graph))
	}
	return nil
}

// ValidateTaskLifecyclePlan is the pure lifecycle authorization boundary. The
// store supplies body from the exact source snapshot for completion gating, then
// owns timestamp/frontmatter materialization and the atomic replacement.
func ValidateTaskLifecyclePlan(graph *TaskGraph, plan TaskLifecyclePlan, body string) (TaskLifecyclePlan, TaskLifecycleAnalysis, error) {
	if err := ValidateTaskLifecycleSource(graph); err != nil {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, err
	}
	if !plan.To.Valid() {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%q: %w", plan.To, domain.ErrValidation)
	}
	if !validTaskLifecycleOverride(plan.Override) {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: unknown task lifecycle override %q", domain.ErrValidation, plan.Override)
	}
	if plan.RevisitAt != "" {
		if plan.To != domain.StatusDeferred {
			return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: revisit_at is valid only for a transition to deferred", domain.ErrValidation)
		}
		if err := domain.ValidateDate(plan.RevisitAt); err != nil {
			return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, err
		}
	}

	if plan.Create != nil {
		return validateCreateAndStartPlan(graph, plan)
	}
	return validateExistingLifecyclePlan(graph, plan, body)
}

func validTaskLifecycleOverride(override TaskLifecycleOverride) bool {
	switch override {
	case TaskLifecycleOverrideNone, TaskLifecycleOverrideDependencyGate, TaskLifecycleOverrideAcceptanceCriteria:
		return true
	default:
		return false
	}
}

func validateCreateAndStartPlan(graph *TaskGraph, plan TaskLifecyclePlan) (TaskLifecyclePlan, TaskLifecycleAnalysis, error) {
	if plan.TaskID != "" {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: create-and-start cannot also name an existing task", domain.ErrValidation)
	}
	if plan.To != domain.StatusInProgress {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: guarded task creation supports only create-and-start", domain.ErrValidation)
	}
	if plan.Override != TaskLifecycleOverrideNone {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: create-and-start has no dependency gate to override", domain.ErrValidation)
	}
	if plan.RevisitAt != "" {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: create-and-start cannot set revisit_at", domain.ErrValidation)
	}

	creation := *plan.Create
	creation.Task = cloneTask(creation.Task)
	plan.Create = &creation
	task := creation.Task
	if task.ID == "" || !id.Valid(task.ID) {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: create-and-start requires a valid stable task id", domain.ErrValidation)
	}
	if task.Slug == "" {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: create-and-start requires a task slug", domain.ErrValidation)
	}
	if task.Status != domain.StatusReadyToStart {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: create-and-start seed document must use ready-to-start, got %s", domain.ErrValidation, task.Status)
	}
	if len(task.DependsOn) > 0 || len(task.LegacyBlockedBy) > 0 || len(task.LegacyDependencies) > 0 || len(task.LegacyBlocks) > 0 {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: create-and-start cannot set graph-owned dependency fields", domain.ErrValidation)
	}
	if _, exists := graph.Task(task.ID); exists {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("task %q already exists: %w", task.ID, domain.ErrConflict)
	}
	started := task
	started.Status = domain.StatusInProgress
	if err := domain.ActiveTaskFieldErr(started); err != nil {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, err
	}

	beforeGraph := taskGraphWithTask(graph, task)
	if err := ValidateTaskLifecycleSource(beforeGraph); err != nil {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, err
	}
	before := beforeGraph.State(task.ID)
	if !before.Eligible {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, eligibilityError(beforeGraph, task.ID, plan.Override)
	}
	afterGraph := taskGraphWithStatus(beforeGraph, task.ID, domain.StatusInProgress)
	return plan, TaskLifecycleAnalysis{From: domain.StatusReadyToStart, Before: before, After: afterGraph.State(task.ID)}, nil
}

func validateExistingLifecyclePlan(graph *TaskGraph, plan TaskLifecyclePlan, body string) (TaskLifecyclePlan, TaskLifecycleAnalysis, error) {
	if plan.TaskID == "" || !id.Valid(plan.TaskID) {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: lifecycle plan requires a valid stable task id", domain.ErrValidation)
	}
	task, exists := graph.Task(plan.TaskID)
	if !exists {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("task %q: %w", plan.TaskID, domain.ErrNotFound)
	}
	before := graph.State(plan.TaskID)
	switch plan.Override {
	case TaskLifecycleOverrideDependencyGate:
		if plan.To != domain.StatusInProgress {
			return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: dependency-gate override applies only when entering in-progress", domain.ErrValidation)
		}
	case TaskLifecycleOverrideAcceptanceCriteria:
		if plan.To != domain.StatusCompleted {
			return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf("%w: acceptance-criteria override applies only when completing a task", domain.ErrValidation)
		}
	}
	if task.Status == plan.To && plan.RevisitAt == "" {
		return plan, TaskLifecycleAnalysis{From: task.Status, Before: before, After: before}, nil
	}

	analysis := TaskLifecycleAnalysis{From: task.Status, Before: before}

	if task.Status != plan.To && plan.To == domain.StatusInProgress {
		if !isPendingWorkRole(before.Role) {
			return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, eligibilityError(graph, plan.TaskID, plan.Override)
		}
		if before.Gate != GateClear {
			analysis.OutstandingBlockers = graph.BlockingFrontier(plan.TaskID)
			if before.Gate != GateBlocked || plan.Override != TaskLifecycleOverrideDependencyGate {
				return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, eligibilityError(graph, plan.TaskID, plan.Override)
			}
			analysis.OverrideApplied = true
		}
	}

	if plan.To == domain.StatusCompleted {
		if unmet := domain.UnexplainedCriteria(body); len(unmet) > 0 {
			if plan.Override != TaskLifecycleOverrideAcceptanceCriteria {
				return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, fmt.Errorf(
					"%w: task %q has %d acceptance criterion/criteria still unmet with no reason (%s); tick them, give each a state (`task ac --defer|--wontfix|--tracked|--na`), or pass --force",
					domain.ErrValidation, task.Slug, len(unmet), criterionIndexesCore(unmet))
			}
			analysis.OverrideApplied = true
		}
	}

	afterGraph := taskGraphWithStatus(graph, plan.TaskID, plan.To)
	afterTask, _ := afterGraph.Task(plan.TaskID)
	if err := domain.ActiveTaskFieldErr(afterTask); err != nil {
		return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, err
	}
	analysis.After = afterGraph.State(plan.TaskID)
	for _, downstream := range graph.DownstreamImpact(plan.TaskID) {
		beforeState, afterState := graph.State(downstream.TaskID), afterGraph.State(downstream.TaskID)
		if beforeState == afterState {
			continue
		}
		analysis.Impacts = append(analysis.Impacts, TaskGraphStateImpact{
			TaskID: downstream.TaskID, Before: beforeState, After: afterState,
			Path: append([]string(nil), downstream.Path...), Direct: downstream.Direct,
		})
	}
	return plan, analysis, nil
}

func eligibilityError(graph *TaskGraph, taskID string, override TaskLifecycleOverride) error {
	return &TaskEligibilityError{
		TaskID: taskID, State: graph.State(taskID), Blockers: graph.BlockingFrontier(taskID), RequestedOverride: override,
	}
}

func criterionIndexesCore(criteria []domain.Criterion) string {
	parts := make([]string, len(criteria))
	for i, criterion := range criteria {
		parts[i] = fmt.Sprintf("#%d", criterion.Index)
	}
	return strings.Join(parts, ", ")
}

func taskGraphWithStatus(graph *TaskGraph, taskID string, status domain.Status) *TaskGraph {
	task, _ := graph.Task(taskID)
	task.Status = status
	return taskGraphWithTask(graph, task)
}

func taskGraphWithTask(graph *TaskGraph, replacement domain.Task) *TaskGraph {
	ids := graph.TaskIDs()
	tasks := make(map[string]domain.Task, len(ids)+1)
	found := false
	for _, taskID := range ids {
		task, _ := graph.Task(taskID)
		if taskID == replacement.ID {
			task = replacement
			found = true
		}
		tasks[taskID] = task
	}
	if !found {
		ids = append(ids, replacement.ID)
		tasks[replacement.ID] = replacement
		sort.Strings(ids)
	}
	return taskGraphFromMap(ids, tasks)
}

func taskGraphAfterDependencyPlan(graph *TaskGraph, plan TaskGraphMutationPlan) *TaskGraph {
	ids := graph.TaskIDs()
	tasks := make(map[string]domain.Task, len(ids))
	for _, taskID := range ids {
		task, _ := graph.Task(taskID)
		tasks[taskID] = task
	}
	for _, write := range plan.TaskWrites {
		task, ok := tasks[write.TaskID]
		if !ok {
			continue
		}
		task.DependsOn = append([]string(nil), write.DependsOn...)
		if write.ClearLegacy {
			task.LegacyBlockedBy = nil
			task.LegacyDependencies = nil
			task.LegacyBlocks = nil
			task.LegacyDependencyFields = nil
		}
		tasks[write.TaskID] = task
	}
	return taskGraphFromMap(ids, tasks)
}

func directDependencyImpacts(before *TaskGraph, plan TaskGraphMutationPlan) []TaskGraphStateImpact {
	if before == nil || len(plan.TaskWrites) == 0 {
		return nil
	}
	after := taskGraphAfterDependencyPlan(before, plan)
	ids := make([]string, 0, len(plan.TaskWrites))
	seen := make(map[string]bool, len(plan.TaskWrites))
	for _, write := range plan.TaskWrites {
		if !seen[write.TaskID] {
			seen[write.TaskID] = true
			ids = append(ids, write.TaskID)
		}
	}
	sort.Strings(ids)
	impacts := make([]TaskGraphStateImpact, 0, len(ids))
	for _, taskID := range ids {
		impacts = append(impacts, TaskGraphStateImpact{
			TaskID: taskID, Before: before.State(taskID), After: after.State(taskID), Direct: true,
		})
	}
	return impacts
}

func cloneTaskGraphStateImpacts(values []TaskGraphStateImpact) []TaskGraphStateImpact {
	out := make([]TaskGraphStateImpact, len(values))
	copy(out, values)
	for i := range out {
		out[i].Path = append([]string(nil), values[i].Path...)
	}
	return out
}

func cloneLifecycleBlockers(values []Blocker) []Blocker {
	out := make([]Blocker, len(values))
	copy(out, values)
	for i := range out {
		out[i].Path = append([]string(nil), values[i].Path...)
	}
	return out
}
