package core

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

type NewThreadParams struct {
	Title       string
	Description string
	Goal        string
	TargetDate  string
	Tags        []string
	Tasks       []string
	Body        string
	Template    string
	DryRun      bool
}

// NewThread creates one explicitly unstarted Thread. Task references are
// resolved only inside the guarded planner against the authoritative graph.
func (s *Service) NewThread(params NewThreadParams) (ThreadCreationReceipt, error) {
	if err := templateBodyConflict(params.Body, params.Template); err != nil {
		return ThreadCreationReceipt{}, err
	}
	if strings.TrimSpace(params.Description) == "" {
		return ThreadCreationReceipt{}, fmt.Errorf("%w: Thread description is required", domain.ErrValidation)
	}
	if err := domain.ValidateDescription(params.Description); err != nil {
		return ThreadCreationReceipt{}, err
	}
	if strings.TrimSpace(params.Goal) == "" {
		return ThreadCreationReceipt{}, fmt.Errorf("%w: Thread goal is required", domain.ErrValidation)
	}
	if strings.ContainsAny(params.Goal, "\r\n") {
		return ThreadCreationReceipt{}, fmt.Errorf("%w: Thread goal must be a single line", domain.ErrValidation)
	}
	if params.TargetDate != "" {
		if err := domain.ValidateDate(params.TargetDate); err != nil {
			return ThreadCreationReceipt{}, err
		}
	}
	slug := domain.Slugify(params.Title)
	if slug == "" {
		return ThreadCreationReceipt{}, fmt.Errorf("%w: title produced an empty slug: %q", domain.ErrValidation, params.Title)
	}
	thread := domain.Thread{
		ID: s.newID(), Slug: slug, Status: domain.ThreadStatusUnstarted,
		Description: params.Description, Goal: params.Goal, TargetDate: params.TargetDate,
		Created: s.now().Format("2006-01-02"), Tags: append([]string(nil), params.Tags...),
	}
	body := params.Body
	if body == "" {
		template, err := s.templateBody("thread", params.Template)
		if err != nil {
			return ThreadCreationReceipt{}, err
		}
		body = renderTemplate(template, map[string]string{"title": params.Title, "goal": params.Goal})
	}
	refs := append([]string(nil), params.Tasks...)
	return s.runThreadCreationMutation(params.DryRun, func(snapshot ThreadCreationSnapshot) (ThreadCreationPlan, error) {
		memberIDs := make([]string, 0, len(refs))
		seen := make(map[string]bool, len(refs))
		for _, ref := range refs {
			taskID, err := snapshot.Graph.ResolveTaskID(ref)
			if err != nil {
				return ThreadCreationPlan{}, err
			}
			if seen[taskID] {
				return ThreadCreationPlan{}, fmt.Errorf("%w: task references resolve to duplicate member %s", domain.ErrValidation, taskID)
			}
			seen[taskID] = true
			memberIDs = append(memberIDs, taskID)
		}
		sort.Strings(memberIDs)
		planned := thread
		planned.Tasks = memberIDs
		return ThreadCreationPlan{Thread: planned, Body: body}, nil
	})
}

func (s *Service) runThreadCreationMutation(dryRun bool, planner ThreadCreationPlanner) (ThreadCreationReceipt, error) {
	if s.threadCreations == nil {
		return ThreadCreationReceipt{}, fmt.Errorf("thread creation is unavailable from this store")
	}
	now := s.now()
	result, err := s.threadCreations.MutateThreadCreation(now, dryRun, planner)
	if !dryRun {
		for attempt := 1; attempt <= s.maxRetries && errors.Is(err, domain.ErrConflict) && !result.Committed; attempt++ {
			s.retrySleep(attempt)
			result, err = s.threadCreations.MutateThreadCreation(now, dryRun, planner)
		}
	}
	receipt := ThreadCreationReceipt{
		Thread: result.Thread, Changed: result.Changed, DryRun: result.DryRun, Committed: result.Committed,
	}
	if err != nil && result.Committed {
		return receipt, &ThreadCreationMutationFailure{Cause: err, Receipt: receipt}
	}
	return receipt, err
}

// AddThreadMembers atomically resolves and adds every requested task to one
// mutable Thread. Existing members are retained as explicit idempotent outcomes.
func (s *Service) AddThreadMembers(threadRef string, taskRefs []string, dryRun bool) (ThreadMutationReceipt, error) {
	return s.mutateThreadMembers(threadRef, taskRefs, ThreadMutationAddMembers, dryRun)
}

// RemoveThreadMembers atomically resolves and removes every requested task from
// one mutable Thread. Absent members are explicit idempotent outcomes.
func (s *Service) RemoveThreadMembers(threadRef string, taskRefs []string, dryRun bool) (ThreadMutationReceipt, error) {
	return s.mutateThreadMembers(threadRef, taskRefs, ThreadMutationRemoveMembers, dryRun)
}

func (s *Service) mutateThreadMembers(threadRef string, taskRefs []string, operation ThreadMutationOperation, dryRun bool) (ThreadMutationReceipt, error) {
	refs := append([]string(nil), taskRefs...)
	if len(refs) == 0 {
		return ThreadMutationReceipt{}, fmt.Errorf("%w: %s requires at least one task", domain.ErrValidation, operation)
	}
	return s.runThreadMutation(dryRun, func(snapshot ThreadMutationSnapshot) (ThreadMutationPlan, error) {
		threadID, err := snapshot.ResolveThreadID(threadRef)
		if err != nil {
			return ThreadMutationPlan{}, err
		}
		taskIDs := make([]string, 0, len(refs))
		seen := make(map[string]bool, len(refs))
		for _, ref := range refs {
			taskID, err := snapshot.Graph.ResolveTaskID(ref)
			if err != nil {
				return ThreadMutationPlan{}, err
			}
			if seen[taskID] {
				return ThreadMutationPlan{}, fmt.Errorf("%w: task references resolve to duplicate member intent %s", domain.ErrValidation, taskID)
			}
			seen[taskID] = true
			taskIDs = append(taskIDs, taskID)
		}
		return ThreadMutationPlan{ThreadID: threadID, Operation: operation, TaskIDs: taskIDs}, nil
	})
}

func (s *Service) StartThread(ref string, dryRun bool) (ThreadMutationReceipt, error) {
	return s.moveThread(ref, ThreadMutationStart, dryRun)
}

func (s *Service) CompleteThread(ref string, dryRun bool) (ThreadMutationReceipt, error) {
	return s.moveThread(ref, ThreadMutationComplete, dryRun)
}

func (s *Service) CancelThread(ref string, dryRun bool) (ThreadMutationReceipt, error) {
	return s.moveThread(ref, ThreadMutationCancel, dryRun)
}

func (s *Service) ReopenThread(ref string, dryRun bool) (ThreadMutationReceipt, error) {
	return s.moveThread(ref, ThreadMutationReopen, dryRun)
}

func (s *Service) moveThread(ref string, operation ThreadMutationOperation, dryRun bool) (ThreadMutationReceipt, error) {
	return s.runThreadMutation(dryRun, func(snapshot ThreadMutationSnapshot) (ThreadMutationPlan, error) {
		threadID, err := snapshot.ResolveThreadID(ref)
		if err != nil {
			return ThreadMutationPlan{}, err
		}
		return ThreadMutationPlan{ThreadID: threadID, Operation: operation}, nil
	})
}

func (s *Service) runThreadMutation(dryRun bool, planner ThreadMutationPlanner) (ThreadMutationReceipt, error) {
	if s.threadMutations == nil {
		return ThreadMutationReceipt{}, fmt.Errorf("thread mutations are unavailable from this store")
	}
	now := s.now()
	result, err := s.threadMutations.MutateThread(now, dryRun, planner)
	if !dryRun {
		for attempt := 1; attempt <= s.maxRetries && errors.Is(err, domain.ErrConflict) && !result.Committed; attempt++ {
			s.retrySleep(attempt)
			result, err = s.threadMutations.MutateThread(now, dryRun, planner)
		}
	}
	receipt := threadMutationReceipt(result)
	if err != nil && result.Committed {
		return receipt, &ThreadMutationFailure{Cause: err, Receipt: receipt}
	}
	return receipt, err
}

func threadMutationReceipt(result ThreadMutationResult) ThreadMutationReceipt {
	receipt := ThreadMutationReceipt{
		Operation: result.Plan.Operation, Thread: cloneThread(result.Thread),
		Before: cloneThreadView(result.Before), After: cloneThreadView(result.After),
		MemberOutcomes: append([]ThreadMemberOutcome(nil), result.MemberOutcomes...),
		Changed:        result.Changed, DryRun: result.DryRun, Committed: result.Committed,
	}
	switch {
	case result.Plan.Operation == ThreadMutationCancel && result.Changed:
		receipt.Remedy = "member tasks were not changed; inspect them independently or through their other Threads"
	case !result.Changed:
		receipt.Remedy = "the requested Thread state was already satisfied; no file write was needed"
	}
	return receipt
}

// ListThreadViews reads every Thread and the task graph once, then applies the
// same projection used by show/frontier. Repository-global graph diagnostics are
// hoisted to the list level rather than repeated on every Thread row.
func (s *Service) ListThreadViews() (ThreadListView, []domain.FileProblem, error) {
	if s.threads == nil {
		return ThreadListView{}, nil, fmt.Errorf("thread reads are unavailable from this store")
	}
	threads, problems, err := s.threads.ListThreads()
	if err != nil {
		return ThreadListView{}, nil, err
	}
	graph, err := LoadTaskGraph(s.taskGraphs)
	if err != nil {
		return ThreadListView{}, nil, err
	}
	sort.Slice(threads, func(i, j int) bool {
		if threads[i].ID != threads[j].ID {
			return threads[i].ID < threads[j].ID
		}
		return threads[i].Path < threads[j].Path
	})
	list := ThreadListView{
		Threads: make([]ThreadView, len(threads)), GraphHealth: graph.Health(), GraphProblems: graph.Problems(),
	}
	for i, thread := range threads {
		list.Threads[i] = ProjectThread(thread, graph)
		list.Threads[i].GraphProblems = nil
	}
	return list, problems, nil
}

func (s *Service) ShowThread(ref string) (ThreadView, string, error) {
	if s.threads == nil {
		return ThreadView{}, "", fmt.Errorf("thread reads are unavailable from this store")
	}
	thread, body, err := s.threads.GetThread(ref)
	if err != nil {
		return ThreadView{}, "", err
	}
	graph, err := LoadTaskGraph(s.taskGraphs)
	if err != nil {
		return ThreadView{}, "", err
	}
	return ProjectThread(thread, graph), body, nil
}

// ShowThreadGraph reads the Thread first and the task graph second, preserving
// the same paired-source ordering contract as ShowThread while returning the
// adapter-neutral graph projection used by every presentation surface.
func (s *Service) ShowThreadGraph(ref string) (ThreadGraphProjection, error) {
	if s.threads == nil {
		return ThreadGraphProjection{}, fmt.Errorf("thread reads are unavailable from this store")
	}
	thread, _, err := s.threads.GetThread(ref)
	if err != nil {
		return ThreadGraphProjection{}, err
	}
	graph, err := LoadTaskGraph(s.taskGraphs)
	if err != nil {
		return ThreadGraphProjection{}, err
	}
	return ProjectThreadGraph(thread, graph), nil
}

func (s *Service) ThreadPath(ref string) (string, error) {
	if s.threads == nil {
		return "", fmt.Errorf("thread reads are unavailable from this store")
	}
	return s.threads.ResolveThreadPath(ref)
}
