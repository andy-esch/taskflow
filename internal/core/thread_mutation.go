package core

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// ThreadMutationOperation is the closed set of existing-Thread writes. Membership
// and lifecycle share one guarded capability because both mutate the same owning
// document and must authorize against the same task/Thread snapshot.
type ThreadMutationOperation string

const (
	ThreadMutationAddMembers    ThreadMutationOperation = "add-members"
	ThreadMutationRemoveMembers ThreadMutationOperation = "remove-members"
	ThreadMutationStart         ThreadMutationOperation = "start"
	ThreadMutationComplete      ThreadMutationOperation = "complete"
	ThreadMutationCancel        ThreadMutationOperation = "cancel"
	ThreadMutationReopen        ThreadMutationOperation = "reopen"
)

// ThreadMutationSnapshot is the immutable semantic input to a guarded mutation
// planner. Persistence versions are stripped before the callback receives it.
type ThreadMutationSnapshot struct {
	Graph   *TaskGraph
	Threads []domain.Thread
}

// ResolveThreadID applies the ordinary exact/prefix/substring tiers without
// re-entering the Store from a guarded planner callback.
func (s ThreadMutationSnapshot) ResolveThreadID(ref string) (string, error) {
	if ref == "" || strings.ContainsAny(ref, `/\`) || strings.Contains(ref, "..") {
		return "", fmt.Errorf("%w: Thread name %q must be a plain name (no path separators)", domain.ErrValidation, ref)
	}
	threads := cloneThreads(s.Threads)
	sort.Slice(threads, func(i, j int) bool {
		if threads[i].ID != threads[j].ID {
			return threads[i].ID < threads[j].ID
		}
		return threads[i].Slug < threads[j].Slug
	})
	query := strings.ToLower(ref)
	tiers := []func(string) bool{
		func(key string) bool { return key == ref || strings.ToLower(key) == query },
		func(key string) bool { return strings.HasPrefix(strings.ToLower(key), query) },
		func(key string) bool { return strings.Contains(strings.ToLower(key), query) },
	}
	for _, matches := range tiers {
		hits := make([]domain.Thread, 0)
		for _, thread := range threads {
			if matches(thread.ID) || (thread.Slug != "" && matches(thread.Slug)) {
				hits = append(hits, thread)
			}
		}
		switch len(hits) {
		case 0:
			continue
		case 1:
			return hits[0].ID, nil
		default:
			details := make([]string, len(hits))
			for i, hit := range hits {
				details[i] = fmt.Sprintf("%s (%s)", hit.Slug, hit.ID)
			}
			return "", fmt.Errorf("%q matches %d Threads: %s: %w", ref, len(hits), strings.Join(details, ", "), domain.ErrAmbiguous)
		}
	}
	return "", fmt.Errorf("thread %q: %w", ref, domain.ErrNotFound)
}

type ThreadMutationPlan struct {
	ThreadID  string
	Operation ThreadMutationOperation
	TaskIDs   []string
}

type ThreadMutationPlanner func(ThreadMutationSnapshot) (ThreadMutationPlan, error)

type ThreadMemberOutcome struct {
	TaskID  string
	Action  string
	Outcome string
}

type ThreadMutationAnalysis struct {
	Before         ThreadView
	After          ThreadView
	MemberOutcomes []ThreadMemberOutcome
	Changed        bool
}

type ThreadMutationResult struct {
	Plan           ThreadMutationPlan
	Thread         domain.Thread
	Before         ThreadView
	After          ThreadView
	MemberOutcomes []ThreadMemberOutcome
	Changed        bool
	DryRun         bool
	Committed      bool
}

type ThreadMutationReceipt struct {
	Operation      ThreadMutationOperation
	Thread         domain.Thread
	Before         ThreadView
	After          ThreadView
	MemberOutcomes []ThreadMemberOutcome
	Changed        bool
	DryRun         bool
	Committed      bool
	Remedy         string
}

// ThreadMutationFailure preserves a durable Thread outcome when guard cleanup
// fails. Callers must inspect the receipt and never blindly retry it.
type ThreadMutationFailure struct {
	Cause   error
	Receipt ThreadMutationReceipt
}

func (e *ThreadMutationFailure) Error() string {
	if e == nil {
		return "Thread mutation committed, but repository cleanup failed"
	}
	return fmt.Sprintf("Thread mutation committed, but repository cleanup failed: %v; inspect the current Thread before retrying", e.Cause)
}

func (e *ThreadMutationFailure) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// ThreadMutationPolicyError makes a refused lifecycle/membership intent
// machine-inspectable while preserving ErrValidation classification.
type ThreadMutationPolicyError struct {
	ThreadID  string
	Operation ThreadMutationOperation
	Status    domain.ThreadStatus
	Reason    string
	Remedy    string
}

func (e *ThreadMutationPolicyError) Error() string {
	if e == nil {
		return "Thread mutation is not allowed"
	}
	message := fmt.Sprintf("Thread %s cannot %s from %s: %s", e.ThreadID, e.Operation, e.Status, e.Reason)
	if e.Remedy != "" {
		message += "; " + e.Remedy
	}
	return message
}

func (e *ThreadMutationPolicyError) Unwrap() error { return domain.ErrValidation }

func threadPolicyError(thread domain.Thread, operation ThreadMutationOperation, reason, remedy string) error {
	return &ThreadMutationPolicyError{
		ThreadID: thread.ID, Operation: operation, Status: thread.Status, Reason: reason, Remedy: remedy,
	}
}

// ValidateThreadMutationSource applies the same strict repository-wide evidence
// gate as Thread creation. A mutation cannot silently ignore an unreadable Thread
// or an invalid/missing member while claiming an authoritative receipt.
func ValidateThreadMutationSource(graph *TaskGraph, threads []domain.Thread, unreadable []domain.FileProblem) error {
	return ValidateThreadCreationSource(graph, threads, unreadable)
}

// ValidateThreadMutationPlan normalizes one planner intent and derives the exact
// before/after projection. now is an injected semantic input, not a wall-clock read.
func ValidateThreadMutationPlan(snapshot ThreadMutationSnapshot, plan ThreadMutationPlan, now time.Time) (ThreadMutationPlan, ThreadMutationAnalysis, error) {
	if now.IsZero() {
		return ThreadMutationPlan{}, ThreadMutationAnalysis{}, fmt.Errorf("%w: Thread mutation time is required", domain.ErrValidation)
	}
	if snapshot.Graph == nil {
		return ThreadMutationPlan{}, ThreadMutationAnalysis{}, fmt.Errorf("%w: authoritative task graph is required", domain.ErrValidation)
	}
	thread, exists := threadByID(snapshot.Threads, plan.ThreadID)
	if !exists {
		return ThreadMutationPlan{}, ThreadMutationAnalysis{}, fmt.Errorf("thread %q: %w", plan.ThreadID, domain.ErrNotFound)
	}
	plan.TaskIDs = append([]string(nil), plan.TaskIDs...)
	sort.Strings(plan.TaskIDs)
	for i, taskID := range plan.TaskIDs {
		if i > 0 && taskID == plan.TaskIDs[i-1] {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, fmt.Errorf("%w: Thread mutation repeats task %s", domain.ErrValidation, taskID)
		}
		if _, exists := snapshot.Graph.Task(taskID); !exists {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, fmt.Errorf("%w: Thread member task %s does not exist", domain.ErrValidation, taskID)
		}
	}

	analysis := ThreadMutationAnalysis{Before: ProjectThread(thread, snapshot.Graph)}
	after := cloneThread(thread)
	date := now.Format("2006-01-02")
	switch plan.Operation {
	case ThreadMutationAddMembers, ThreadMutationRemoveMembers:
		if len(plan.TaskIDs) == 0 {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, fmt.Errorf("%w: %s requires at least one task", domain.ErrValidation, plan.Operation)
		}
		if thread.Status != domain.ThreadStatusUnstarted && thread.Status != domain.ThreadStatusInProgress {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, threadPolicyError(thread, plan.Operation,
				"terminal Threads are membership-immutable", "reopen a completed Thread before changing membership; cancelled Threads cannot reopen in V1")
		}
		members := make(map[string]bool, len(thread.Tasks)+len(plan.TaskIDs))
		for _, taskID := range thread.Tasks {
			members[taskID] = true
		}
		for _, taskID := range plan.TaskIDs {
			had := members[taskID]
			action, outcome := "add", "skipped"
			if plan.Operation == ThreadMutationAddMembers {
				if !had {
					members[taskID], outcome = true, "added"
				}
			} else {
				action = "remove"
				if had {
					delete(members, taskID)
					outcome = "removed"
				}
			}
			analysis.MemberOutcomes = append(analysis.MemberOutcomes, ThreadMemberOutcome{TaskID: taskID, Action: action, Outcome: outcome})
		}
		after.Tasks = make([]string, 0, len(members))
		for taskID := range members {
			after.Tasks = append(after.Tasks, taskID)
		}
		sort.Strings(after.Tasks)
		analysis.Changed = !slices.Equal(thread.Tasks, after.Tasks)
	case ThreadMutationStart:
		if len(plan.TaskIDs) != 0 {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, fmt.Errorf("%w: start does not accept member task ids", domain.ErrValidation)
		}
		if thread.Status != domain.ThreadStatusUnstarted && thread.Status != domain.ThreadStatusInProgress {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, threadPolicyError(thread, plan.Operation,
				"only unstarted or already-active Threads can start", "reopen a completed Thread; cancelled Threads cannot reopen in V1")
		}
		if analysis.Before.Rollup.Total == 0 {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, threadPolicyError(thread, plan.Operation,
				"at least one non-deprecated member is required", "add a live member task before starting")
		}
		if thread.Status != domain.ThreadStatusInProgress {
			after.Status = domain.ThreadStatusInProgress
			after.StartedAt = date
			after.EndedAt = ""
			analysis.Changed = true
		}
	case ThreadMutationComplete:
		if len(plan.TaskIDs) != 0 {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, fmt.Errorf("%w: complete does not accept member task ids", domain.ErrValidation)
		}
		if thread.Status != domain.ThreadStatusInProgress && thread.Status != domain.ThreadStatusCompleted {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, threadPolicyError(thread, plan.Operation,
				"only in-progress or already-completed Threads can complete", "start the Thread before completing it")
		}
		if err := validateThreadCompletion(thread, analysis.Before); err != nil {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, err
		}
		if thread.Status != domain.ThreadStatusCompleted {
			after.Status = domain.ThreadStatusCompleted
			after.EndedAt = date
			analysis.Changed = true
		}
	case ThreadMutationCancel:
		if len(plan.TaskIDs) != 0 {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, fmt.Errorf("%w: cancel does not accept member task ids", domain.ErrValidation)
		}
		if thread.Status != domain.ThreadStatusUnstarted && thread.Status != domain.ThreadStatusInProgress && thread.Status != domain.ThreadStatusCancelled {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, threadPolicyError(thread, plan.Operation,
				"completed Threads cannot be reclassified as cancelled", "leave it completed, or reopen it before making further project changes")
		}
		if thread.Status != domain.ThreadStatusCancelled {
			after.Status = domain.ThreadStatusCancelled
			after.EndedAt = date
			analysis.Changed = true
		}
	case ThreadMutationReopen:
		if len(plan.TaskIDs) != 0 {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, fmt.Errorf("%w: reopen does not accept member task ids", domain.ErrValidation)
		}
		if thread.Status != domain.ThreadStatusCompleted && thread.Status != domain.ThreadStatusInProgress {
			return ThreadMutationPlan{}, ThreadMutationAnalysis{}, threadPolicyError(thread, plan.Operation,
				"only completed or already-active Threads can reopen", "cancelled Threads are terminal in V1; start an unstarted Thread instead")
		}
		if thread.Status == domain.ThreadStatusCompleted {
			after.Status = domain.ThreadStatusInProgress
			after.EndedAt = ""
			analysis.Changed = true
		}
	default:
		return ThreadMutationPlan{}, ThreadMutationAnalysis{}, fmt.Errorf("%w: unknown Thread mutation operation %q", domain.ErrValidation, plan.Operation)
	}
	if analysis.Changed {
		after.Updated = date
	}
	if err := domain.ValidateThreadDocument(after); err != nil {
		return ThreadMutationPlan{}, ThreadMutationAnalysis{}, err
	}
	analysis.After = ProjectThread(after, snapshot.Graph)
	return plan, analysis, nil
}

func validateThreadCompletion(thread domain.Thread, view ThreadView) error {
	if view.ProjectionHealth != GraphHealthy {
		return threadPolicyError(thread, ThreadMutationComplete, "Thread projection evidence is not healthy", "repair the reported Thread or graph problems before completing")
	}
	if view.Rollup.Total == 0 {
		return threadPolicyError(thread, ThreadMutationComplete, "at least one non-deprecated member is required", "add and finish a live member task before completing")
	}
	if view.Rollup.Drained != view.Rollup.Total {
		return threadPolicyError(thread, ThreadMutationComplete,
			fmt.Sprintf("only %d of %d live members are soundly completed", view.Rollup.Drained, view.Rollup.Total),
			"finish or remove unfinished work; deferred tasks remain live and block completion")
	}
	// Drained live members currently imply closed external gates and sound member
	// evidence. Keep the explicit checks as defense against a future projection
	// representation that can express those states independently.
	for _, gate := range view.ExternalGates {
		if gate.Outstanding {
			return threadPolicyError(thread, ThreadMutationComplete,
				fmt.Sprintf("external gate %s is not soundly completed", gate.State.TaskID),
				"complete or repair the external prerequisite before completing the Thread")
		}
	}
	for _, member := range view.Members {
		// Deprecated work remains in Thread membership as historical context, but
		// it is withdrawn from the completion denominator and cannot gate closure.
		if member.State.Role == RoleWithdrawn {
			continue
		}
		if member.State.Inconsistent || member.State.Gate == GateBroken {
			return threadPolicyError(thread, ThreadMutationComplete,
				fmt.Sprintf("member %s has inconsistent or broken graph state", member.State.TaskID),
				"inspect the member blockers and restore sound dependency evidence")
		}
	}
	return nil
}

func threadByID(threads []domain.Thread, threadID string) (domain.Thread, bool) {
	for _, thread := range threads {
		if thread.ID == threadID {
			return cloneThread(thread), true
		}
	}
	return domain.Thread{}, false
}

// ThreadProjectionImpact is one Thread whose derived view changes across a task
// lifecycle mutation. Direct says the transitioned task is itself a member;
// ChangedTaskIDs also captures downstream-member and external-gate effects.
type ThreadProjectionImpact struct {
	ThreadID       string
	Slug           string
	Direct         bool
	ChangedTaskIDs []string
	Before         ThreadView
	After          ThreadView
}

// TaskLifecycleThreadImpacts compares every Thread projection using the same
// before/after task graphs as lifecycle authorization. Thread files are never
// mutated by this analysis.
func TaskLifecycleThreadImpacts(threads []domain.Thread, graph *TaskGraph, plan TaskLifecyclePlan) []ThreadProjectionImpact {
	if graph == nil || plan.Create != nil || plan.TaskID == "" {
		return nil
	}
	task, exists := graph.Task(plan.TaskID)
	if !exists || task.Status == plan.To {
		return nil
	}
	afterGraph := taskGraphWithStatus(graph, plan.TaskID, plan.To)
	ordered := cloneThreads(threads)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	impacts := make([]ThreadProjectionImpact, 0)
	for _, thread := range ordered {
		before, after := ProjectThread(thread, graph), ProjectThread(thread, afterGraph)
		if reflect.DeepEqual(before, after) {
			continue
		}
		impacts = append(impacts, ThreadProjectionImpact{
			ThreadID: thread.ID, Slug: thread.Slug, Direct: slices.Contains(thread.Tasks, plan.TaskID),
			ChangedTaskIDs: changedThreadProjectionTaskIDs(before, after), Before: before, After: after,
		})
	}
	return impacts
}

func changedThreadProjectionTaskIDs(before, after ThreadView) []string {
	type projectionTask struct {
		Role        ThreadTaskRole
		Task        domain.Task
		State       TaskGraphState
		Outstanding bool
	}
	items := func(view ThreadView) map[string]projectionTask {
		out := make(map[string]projectionTask, len(view.Members)+len(view.ExternalGates))
		for _, member := range view.Members {
			out[member.State.TaskID] = projectionTask{Role: member.Role, Task: member.Task, State: member.State}
		}
		for _, gate := range view.ExternalGates {
			out[gate.State.TaskID] = projectionTask{Role: gate.Role, Task: gate.Task, State: gate.State, Outstanding: gate.Outstanding}
		}
		return out
	}
	left, right := items(before), items(after)
	ids := make(map[string]bool, len(left)+len(right))
	for taskID := range left {
		ids[taskID] = true
	}
	for taskID := range right {
		ids[taskID] = true
	}
	changed := make([]string, 0)
	for taskID := range ids {
		if !reflect.DeepEqual(left[taskID], right[taskID]) {
			changed = append(changed, taskID)
		}
	}
	sort.Strings(changed)
	return changed
}

func cloneThreadProjectionImpacts(values []ThreadProjectionImpact) []ThreadProjectionImpact {
	out := make([]ThreadProjectionImpact, len(values))
	copy(out, values)
	for i := range out {
		out[i].ChangedTaskIDs = append([]string(nil), values[i].ChangedTaskIDs...)
		out[i].Before = cloneThreadView(values[i].Before)
		out[i].After = cloneThreadView(values[i].After)
	}
	return out
}

func cloneThreadView(view ThreadView) ThreadView {
	view.Thread = cloneThread(view.Thread)
	view.Members = append([]ThreadTaskView(nil), view.Members...)
	view.ExternalGates = append([]ThreadExternalGate(nil), view.ExternalGates...)
	view.Frontier = append([]ThreadTaskView(nil), view.Frontier...)
	view.GraphProblems = append([]GraphProblem(nil), view.GraphProblems...)
	view.Problems = append([]ThreadProblem(nil), view.Problems...)
	return view
}
