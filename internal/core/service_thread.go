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
	graph, err := LoadTaskGraph(s.store)
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
	graph, err := LoadTaskGraph(s.store)
	if err != nil {
		return ThreadView{}, "", err
	}
	return ProjectThread(thread, graph), body, nil
}

func (s *Service) ThreadPath(ref string) (string, error) {
	if s.threads == nil {
		return "", fmt.Errorf("thread reads are unavailable from this store")
	}
	return s.threads.ResolveThreadPath(ref)
}
