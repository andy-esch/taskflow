package core

import (
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

type countingTaskStore struct {
	fakeStore
	listCalls int
}

type neutralTaskGraphSource struct {
	read  TaskGraphRead
	calls int
}

func (s *neutralTaskGraphSource) ReadTaskGraph() (TaskGraphRead, error) {
	s.calls++
	return s.read, nil
}

func (s *countingTaskStore) ListTasks() ([]domain.Task, []domain.FileProblem, error) {
	s.listCalls++
	return s.fakeStore.ListTasks()
}

func (s *countingTaskStore) ReadTasks() (TaskRead, error) {
	s.listCalls++
	return s.fakeStore.ReadTasks()
}

func TestBoard_ActivePipelineOnlyInOrder(t *testing.T) {
	svc := NewService(&fakeStore{tasks: []domain.Task{
		{Slug: "a", Status: domain.StatusInProgress},
		{Slug: "b", Status: domain.StatusNextUp},
		{Slug: "c", Status: domain.StatusInProgress},
		{Slug: "d", Status: domain.StatusReadyToStart},
		{Slug: "done", Status: domain.StatusCompleted},  // terminal → excluded
		{Slug: "dead", Status: domain.StatusDeprecated}, // terminal → excluded
		{Slug: "parked", Status: domain.StatusDeferred}, // parked → excluded
	}})
	b, err := svc.Board()
	if err != nil {
		t.Fatal(err)
	}

	want := []domain.Status{domain.StatusNextUp, domain.StatusReadyToStart, domain.StatusInProgress}
	if len(b.Columns) != len(want) {
		t.Fatalf("got %d columns, want %d", len(b.Columns), len(want))
	}
	counts := map[domain.Status]int{}
	for i, c := range b.Columns {
		if c.Status != want[i] {
			t.Errorf("column %d = %q, want %q", i, c.Status, want[i])
		}
		counts[c.Status] = len(c.Tasks)
		for _, tk := range c.Tasks {
			if !tk.Status.IsActive() {
				t.Errorf("non-active task %q leaked onto the board (status %q)", tk.Slug, tk.Status)
			}
		}
	}
	if counts[domain.StatusInProgress] != 2 || counts[domain.StatusNextUp] != 1 || counts[domain.StatusReadyToStart] != 1 {
		t.Errorf("wrong per-column counts: %v", counts)
	}
}

func TestBoard_EmptyColumnsStillPresent(t *testing.T) {
	svc := NewService(&fakeStore{tasks: []domain.Task{
		{Slug: "a", Status: domain.StatusInProgress},
	}})
	b, err := svc.Board()
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Columns) != 3 {
		t.Fatalf("got %d columns, want 3 (empty statuses still show a column)", len(b.Columns))
	}
	for _, c := range b.Columns {
		if c.Status != domain.StatusInProgress && len(c.Tasks) != 0 {
			t.Errorf("column %q should be empty, has %d", c.Status, len(c.Tasks))
		}
	}
}

func TestBoard_CompleteStoreFallbackScansTasksOnce(t *testing.T) {
	store := &countingTaskStore{fakeStore: fakeStore{tasks: []domain.Task{{
		ID: "6g3q4rtmv4ak", Slug: "single-scan", Status: domain.StatusNextUp,
	}}}}
	board, err := NewService(store).Board()
	if err != nil || len(board.Columns) != len(domain.ActiveStatuses()) {
		t.Fatalf("board = %+v, err = %v", board, err)
	}
	if store.listCalls != 1 {
		t.Fatalf("task snapshot reads = %d, want 1", store.listCalls)
	}
}

func TestBoard_PreservesPathlessTaskLoadProblemIdentity(t *testing.T) {
	source := &neutralTaskGraphSource{read: TaskGraphRead{Problems: []TaskGraphLoadProblem{{
		TaskID: "6gpathless01", TaskSlug: "known-task",
		Location: "records/task/contradictory-name",
		Path:     "/planning/tasks/6gpathless01-known-task.md",
		Message:  "invalid frontmatter",
	}}}}

	board, err := NewService(&fakeStore{}, WithTaskGraphSource(source)).Board()
	if err != nil {
		t.Fatal(err)
	}
	if source.calls != 1 {
		t.Fatalf("task graph reads = %d, want exactly 1", source.calls)
	}
	if len(board.Problems) != 1 {
		t.Fatalf("problems = %+v", board.Problems)
	}
	problem := board.Problems[0]
	if problem.EntityKind != EntityTask || problem.EntityID != "6gpathless01" ||
		problem.EntitySlug != "known-task" || problem.Location != "records/task/contradictory-name" ||
		problem.LocalPath != "/planning/tasks/6gpathless01-known-task.md" ||
		problem.Message != "invalid frontmatter" {
		t.Fatalf("portable task diagnostic = %+v", problem)
	}
}

func TestBoard_BareProjectionUsesExplicitSourceIdentity(t *testing.T) {
	source := &neutralTaskGraphSource{read: TaskGraphRead{Records: []LoadedRecord[domain.Task]{{
		Value:  domain.Task{ID: "stale-declaration", FilenameID: "stale-file-id", Slug: "task", Status: domain.StatusInProgress},
		Source: RecordSource{ID: "6g0000000001", Location: "db://tasks/misleading"},
	}}}}
	board, err := NewService(&fakeStore{}, WithTaskGraphSource(source)).Board()
	if err != nil {
		t.Fatal(err)
	}
	if source.calls != 1 {
		t.Fatalf("graph read calls = %d", source.calls)
	}
	got := board.Columns[len(board.Columns)-1].Tasks[0].FilenameID
	if got != "6g0000000001" {
		t.Fatalf("board identity = %q, want source ID", got)
	}
}

func TestEmptyExplicitTaskSourceIDNeverBecomesEligible(t *testing.T) {
	read := TaskGraphRead{Records: []LoadedRecord[domain.Task]{{
		Value:  domain.Task{ID: "6g0000000001", FilenameID: "stale-id", Slug: "orphan", Status: domain.StatusReadyToStart},
		Source: RecordSource{Location: "db://tasks/misleading-id"},
	}}}
	graph := NewTaskGraphRead(read)
	if graph.Health() != GraphBroken || len(graph.TaskIDs()) != 0 {
		t.Fatalf("empty source ID promoted to graph node: health=%s ids=%v", graph.Health(), graph.TaskIDs())
	}
	if state := graph.State("6g0000000001"); state.Eligible {
		t.Fatalf("declared ID became eligible: %+v", state)
	}
	source := &neutralTaskGraphSource{read: read}
	svc := NewService(&fakeStore{}, WithTaskGraphSource(source))
	board, err := svc.Board()
	if err != nil {
		t.Fatal(err)
	}
	for _, column := range board.Columns {
		if len(column.Tasks) != 0 {
			t.Fatalf("board published source-less record: %+v", board.Columns)
		}
	}
	if board.GraphHealth != GraphBroken || len(board.Problems) != 1 ||
		board.Problems[0].EntityID != "" || board.Problems[0].EntitySlug != "orphan" ||
		board.Problems[0].Location != "db://tasks/misleading-id" || board.Problems[0].LocalPath != "" {
		t.Fatalf("source-less board diagnostic = %+v", board)
	}
	tasks, problems, err := svc.ListTasks(TaskFilter{})
	if err != nil || len(tasks) != 0 || len(problems) != 1 || problems[0].EntityID != "" {
		t.Fatalf("source-less list = tasks:%+v problems:%+v err:%v", tasks, problems, err)
	}
}

func TestBoard_CanonicalizesPortableLoadProblems(t *testing.T) {
	source := &neutralTaskGraphSource{read: TaskGraphRead{Problems: []TaskGraphLoadProblem{
		{TaskID: "6g0000000002", TaskSlug: "second", Location: "db://tasks/2", Message: "z"},
		{TaskID: "6g0000000001", TaskSlug: "first", Location: "db://tasks/1", Message: "a"},
	}}}

	board, err := NewService(&fakeStore{}, WithTaskGraphSource(source)).Board()
	if err != nil {
		t.Fatal(err)
	}
	if len(board.Problems) != 2 || board.Problems[0].EntityID != "6g0000000001" ||
		board.Problems[1].EntityID != "6g0000000002" {
		t.Fatalf("portable task diagnostics are not canonical: %+v", board.Problems)
	}
}

func TestBoard_DoesNotInferIdentityFromOpaqueLocation(t *testing.T) {
	source := &neutralTaskGraphSource{read: TaskGraphRead{Problems: []TaskGraphLoadProblem{{
		Location: "db://tasks/6g0000000009-wrong.md", Message: "invalid frontmatter",
	}}}}

	board, err := NewService(&fakeStore{}, WithTaskGraphSource(source)).Board()
	if err != nil {
		t.Fatal(err)
	}
	if len(board.Problems) != 1 || board.Problems[0].EntityID != "" || board.Problems[0].EntitySlug != "" {
		t.Fatalf("opaque location manufactured identity: %+v", board.Problems)
	}
}
