package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

type threadReadFake struct {
	threads  []domain.Thread
	problems []domain.FileProblem
}

func (f *threadReadFake) ListThreads() ([]domain.Thread, []domain.FileProblem, error) {
	return f.threads, f.problems, nil
}

func (f *threadReadFake) GetThread(string) (domain.Thread, string, error) {
	return domain.Thread{}, "", fmt.Errorf("unused")
}

func (f *threadReadFake) ResolveThreadPath(string) (string, error) {
	return "", fmt.Errorf("unused")
}

type threadCreationFake struct {
	snapshot ThreadCreationSnapshot
	result   ThreadCreationMutationResult
	err      error
	calls    int
}

func (f *threadCreationFake) MutateThreadCreation(_ time.Time, dryRun bool, planner ThreadCreationPlanner) (ThreadCreationMutationResult, error) {
	f.calls++
	plan, err := planner(f.snapshot)
	if err != nil {
		return ThreadCreationMutationResult{DryRun: dryRun}, err
	}
	result := f.result
	result.Plan, result.Thread, result.DryRun = plan, plan.Thread, dryRun
	return result, f.err
}

func TestServiceNewThreadResolvesAndSortsMembersInsidePlanner(t *testing.T) {
	a := graphRecord("a-member", domain.StatusReadyToStart)
	b := graphRecord("b-member", domain.StatusReadyToStart)
	fake := &threadCreationFake{snapshot: ThreadCreationSnapshot{Graph: NewTaskGraph([]domain.Task{b, a}, nil)}, result: ThreadCreationMutationResult{Changed: true, Committed: true}}
	svc := NewService(nil, WithThreadCreationMutationStore(fake), WithIDGen(func() string { return "6g3q4rtmv4ak" }), WithClock(func() time.Time {
		return time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	}))
	receipt, err := svc.NewThread(NewThreadParams{
		Title: "Implementation", Description: "Thread implementation", Goal: "Ship Thread reads", Tasks: []string{b.Slug, a.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Thread.Tasks) != 2 || receipt.Thread.Tasks[0] > receipt.Thread.Tasks[1] {
		t.Fatalf("members = %v", receipt.Thread.Tasks)
	}
	if receipt.Thread.Status != domain.ThreadStatusUnstarted || receipt.Thread.Created != "2026-08-29" {
		t.Fatalf("Thread = %+v", receipt.Thread)
	}
}

func TestServiceNewThreadRejectsDuplicateResolvedMember(t *testing.T) {
	task := graphRecord("member", domain.StatusReadyToStart)
	fake := &threadCreationFake{snapshot: ThreadCreationSnapshot{Graph: NewTaskGraph([]domain.Task{task}, nil)}}
	svc := NewService(nil, WithThreadCreationMutationStore(fake), WithIDGen(func() string { return "6g3q4rtmv4ak" }))
	_, err := svc.NewThread(NewThreadParams{
		Title: "Implementation", Description: "Thread implementation", Goal: "Ship", Tasks: []string{task.ID, task.Slug},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("error = %v", err)
	}
}

func TestServiceThreadCommittedFailureIsNotRetried(t *testing.T) {
	fake := &threadCreationFake{
		snapshot: ThreadCreationSnapshot{Graph: NewTaskGraph(nil, nil)},
		result:   ThreadCreationMutationResult{Changed: true, Committed: true},
		err:      errors.New("unlock failed"),
	}
	svc := NewService(nil, WithThreadCreationMutationStore(fake), WithIDGen(func() string { return "6g3q4rtmv4ak" }), WithRetry(4, func(int) {}))
	receipt, err := svc.NewThread(NewThreadParams{Title: "Implementation", Description: "Thread implementation", Goal: "Ship"})
	var committed *ThreadCreationMutationFailure
	if !errors.As(err, &committed) || !receipt.Committed || fake.calls != 1 {
		t.Fatalf("receipt=%+v err=%v calls=%d", receipt, err, fake.calls)
	}
}

func TestServiceLintIncludesThreadIntegrityAndCrossKindIdentity(t *testing.T) {
	task := domain.Task{
		ID: "6g3q4rtmv4ak", FilenameID: "6g3q4rtmv4ak", Slug: "task", Path: "tasks/6g3q4rtmv4ak-task.md",
		Status: domain.StatusCompleted, Epic: "01-e", Created: "2026-08-29",
	}
	threadStore := &threadReadFake{
		threads: []domain.Thread{{
			ID: "6g3q4rtmv4ak", FilenameID: "6g3q4rtmv4ak", Slug: "thread", Path: "threads/6g3q4rtmv4ak-thread.md",
			Status: domain.ThreadStatusUnstarted, Description: "Valid description", Goal: "Ship it",
			Created: "2026-08-29", Tasks: []string{"6g3q4rtmv4az"},
		}},
		problems: []domain.FileProblem{{Path: "threads/bad.md", Message: "bad frontmatter"}},
	}
	svc := NewService(&fakeStore{tasks: []domain.Task{task}}, WithThreadStore(threadStore))
	results, problems, err := svc.Lint()
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 || problems[0].Path != "threads/bad.md" {
		t.Fatalf("problems = %+v", problems)
	}
	got := make(map[string]string)
	for _, result := range results {
		parts := make([]string, 0, len(result.Issues))
		for _, issue := range result.Issues {
			parts = append(parts, issue.Field+": "+issue.Message)
		}
		got[result.Slug] = strings.Join(parts, "; ")
	}
	if !strings.Contains(got["task"], "also used by a Thread") {
		t.Fatalf("task lint = %q", got["task"])
	}
	if !strings.Contains(got["thread"], "also used by a task") || !strings.Contains(got["thread"], "unknown member") {
		t.Fatalf("Thread lint = %q", got["thread"])
	}
}

func TestServiceThreadListHoistsRepositoryGraphDiagnostics(t *testing.T) {
	task := graphRecord("broken-member", domain.StatusReadyToStart, "6g0000000009")
	threadStore := &threadReadFake{threads: []domain.Thread{{
		ID: "6g3q4rtmv4ak", FilenameID: "6g3q4rtmv4ak", Slug: "thread", Path: "threads/6g3q4rtmv4ak-thread.md",
		Status: domain.ThreadStatusUnstarted, Description: "Broken graph list", Goal: "Report once",
		Created: "2026-08-29", Tasks: []string{task.ID},
	}}}
	svc := NewService(&fakeStore{tasks: []domain.Task{task}}, WithThreadStore(threadStore))

	list, problems, err := svc.ListThreadViews()
	if err != nil || len(problems) != 0 {
		t.Fatalf("list err=%v problems=%+v", err, problems)
	}
	if list.GraphHealth != GraphBroken || len(list.GraphProblems) == 0 || len(list.Threads) != 1 {
		t.Fatalf("list = %+v", list)
	}
	if len(list.Threads[0].GraphProblems) != 0 || list.Threads[0].GraphHealth != GraphBroken {
		t.Fatalf("row repeated repository diagnostics: %+v", list.Threads[0])
	}
}
