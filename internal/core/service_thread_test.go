package core

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

type threadReadFake struct {
	threads  []domain.Thread
	problems []domain.FileProblem
	thread   domain.Thread
	body     string
	getErr   error
	onList   func()
	onGet    func()
}

func (f *threadReadFake) ListThreads() ([]domain.Thread, []domain.FileProblem, error) {
	if f.onList != nil {
		f.onList()
	}
	return f.threads, f.problems, nil
}

func (f *threadReadFake) GetThread(string) (domain.Thread, string, error) {
	if f.onGet != nil {
		f.onGet()
	}
	return f.thread, f.body, f.getErr
}

func (f *threadReadFake) ResolveThreadPath(string) (string, error) {
	return "", fmt.Errorf("unused")
}

type taskGraphReadFake struct {
	tasks    []domain.Task
	problems []domain.FileProblem
	err      error
	onList   func()
	calls    int
}

func (f *taskGraphReadFake) ReadTaskGraph() (TaskGraphRead, error) {
	f.calls++
	if f.onList != nil {
		f.onList()
	}
	return TaskGraphReadFromFiles(f.tasks, f.problems), f.err
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

func TestServiceThreadReadsComposeIndependentGraphAndThreadPorts(t *testing.T) {
	gate := graphRecord("split-gate", domain.StatusCompleted)
	member := graphRecord("split-member", domain.StatusReadyToStart, gate.ID)
	thread := domain.Thread{
		ID: "6g3q4rtmv4ak", FilenameID: "6g3q4rtmv4ak", Slug: "split-thread",
		Status: domain.ThreadStatusUnstarted, Description: "Split read ports", Goal: "Keep core reusable",
		Created: "2026-08-31", Tasks: []string{member.ID},
	}
	graphs := &taskGraphReadFake{tasks: []domain.Task{member, gate}}
	threads := &threadReadFake{threads: []domain.Thread{thread}, thread: thread, body: "# Split Thread\n"}
	svc := NewService(nil, WithTaskGraphSource(graphs), WithThreadStore(threads))

	view, body, err := svc.ShowThread(thread.ID)
	if err != nil || body != "# Split Thread\n" || len(view.Members) != 1 || len(view.ExternalGates) != 1 {
		t.Fatalf("show view=%+v body=%q err=%v", view, body, err)
	}
	list, problems, err := svc.ListThreadViews()
	if err != nil || len(problems) != 0 || len(list.Threads) != 1 || list.Threads[0].Thread.ID != thread.ID {
		t.Fatalf("list=%+v problems=%+v err=%v", list, problems, err)
	}

	blockers, err := svc.TaskBlockers(member.ID, false)
	if err != nil || len(blockers.Blockers) != 0 || blockers.State.Gate != GateClear {
		t.Fatalf("blockers=%+v err=%v", blockers, err)
	}
	unblocks, err := svc.TaskUnblocks(gate.ID)
	if err != nil || len(unblocks.Unblocks) != 1 || unblocks.Unblocks[0].Task.ID != member.ID {
		t.Fatalf("unblocks=%+v err=%v", unblocks, err)
	}
}

func TestServiceThreadReadsFailExplicitlyWithoutTaskGraphSource(t *testing.T) {
	thread := domain.Thread{
		ID: "6g3q4rtmv4ak", Slug: "missing-graph", Status: domain.ThreadStatusUnstarted,
		Description: "Missing graph source", Goal: "Fail explicitly", Created: "2026-08-31",
	}
	svc := NewService(nil, WithThreadStore(&threadReadFake{thread: thread}))

	if _, _, err := svc.ShowThread(thread.ID); err == nil || !strings.Contains(err.Error(), "task graph reads are unavailable") {
		t.Fatalf("show error = %v", err)
	}
	if _, _, err := svc.ListThreadViews(); err == nil || !strings.Contains(err.Error(), "task graph reads are unavailable") {
		t.Fatalf("list error = %v", err)
	}
	if _, err := svc.TaskBlockers(thread.ID, false); err == nil || !strings.Contains(err.Error(), "task graph reads are unavailable") {
		t.Fatalf("blockers error = %v", err)
	}
}

func TestServiceGraphQueriesDoNotRequireThreadSupport(t *testing.T) {
	task := graphRecord("graph-only", domain.StatusReadyToStart)
	graphs := &taskGraphReadFake{tasks: []domain.Task{task}}
	svc := NewService(nil, WithTaskGraphSource(graphs))

	result, err := svc.TaskBlockers(task.ID, false)
	if err != nil || result.Task.ID != task.ID || result.State.Gate != GateClear {
		t.Fatalf("blockers = %+v, err = %v", result, err)
	}
	if _, _, err := svc.ShowThread("any-thread"); err == nil || !strings.Contains(err.Error(), "thread reads are unavailable") {
		t.Fatalf("show error = %v", err)
	}
	tasks, problems, err := svc.ListTasks(TaskFilter{Unblocked: true})
	if err != nil || len(problems) != 0 || len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Fatalf("task list = %+v, problems = %+v, err = %v", tasks, problems, err)
	}
	board, err := svc.Board()
	if err != nil || len(board.Columns) != len(domain.ActiveStatuses()) {
		t.Fatalf("board = %+v, err = %v", board, err)
	}
	if _, _, err := svc.ListTasks(TaskFilter{Epic: "30"}); err == nil || !strings.Contains(err.Error(), "epic reads are unavailable") {
		t.Fatalf("epic-filtered list error = %v", err)
	}
	if graphs.calls != 3 {
		t.Fatalf("graph source calls = %d, want one per blockers/list/board operation", graphs.calls)
	}
}

func TestServiceGraphReadsRejectTypedNilCapabilities(t *testing.T) {
	var store *fakeStore
	svc := NewService(store)
	if _, err := svc.TaskBlockers("any-task", false); err == nil || !strings.Contains(err.Error(), "task graph reads are unavailable") {
		t.Fatalf("typed-nil aggregate error = %v", err)
	}

	var graphs *taskGraphReadFake
	svc = NewService(nil, WithTaskGraphSource(graphs))
	if _, err := svc.Board(); err == nil || !strings.Contains(err.Error(), "task graph reads are unavailable") {
		t.Fatalf("typed-nil graph option error = %v", err)
	}

	var threads *threadReadFake
	svc = NewService(nil,
		WithTaskGraphSource(&taskGraphReadFake{}),
		WithThreadStore(threads),
	)
	if _, _, err := svc.ShowThread("any-thread"); err == nil || !strings.Contains(err.Error(), "thread reads are unavailable") {
		t.Fatalf("typed-nil Thread option error = %v", err)
	}
}

func TestServiceThreadViewsReadThreadsBeforeTasks(t *testing.T) {
	task := graphRecord("ordered-member", domain.StatusReadyToStart)
	thread := domain.Thread{
		ID: "6g3q4rtmv4ak", Slug: "ordered-thread", Status: domain.ThreadStatusUnstarted,
		Description: "Ordered reads", Goal: "Pin the compatibility contract", Created: "2026-08-31",
		Tasks: []string{task.ID},
	}

	for _, operation := range []string{"show", "list"} {
		t.Run(operation, func(t *testing.T) {
			calls := make([]string, 0, 2)
			graphs := &taskGraphReadFake{tasks: []domain.Task{task}, onList: func() { calls = append(calls, "tasks") }}
			threads := &threadReadFake{
				threads: []domain.Thread{thread}, thread: thread,
				onList: func() { calls = append(calls, "threads") },
				onGet:  func() { calls = append(calls, "threads") },
			}
			svc := NewService(nil, WithTaskGraphSource(graphs), WithThreadStore(threads))
			var err error
			if operation == "show" {
				_, _, err = svc.ShowThread(thread.ID)
			} else {
				_, _, err = svc.ListThreadViews()
			}
			if err != nil || !slices.Equal(calls, []string{"threads", "tasks"}) {
				t.Fatalf("calls = %v, err = %v", calls, err)
			}
		})
	}
}
