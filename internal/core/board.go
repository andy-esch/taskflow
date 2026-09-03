package core

import "github.com/andy-esch/taskflow/internal/domain"

// BoardColumn is one status's tasks, in the board's (active-pipeline) order.
type BoardColumn struct {
	Status domain.Status
	Tasks  []domain.Task
}

// Board is the active-work view: tasks grouped by their active status — the
// next-up → ready-to-start → in-progress pipeline (domain.Status.IsActive()) — as the
// on-demand replacement for browsing tasks/<status>/. Terminal (completed/
// deprecated) and parked (deferred) tasks are excluded; those live in `task list`.
// It is a core projection like Summary, rendered by the `board` command (and later
// the web read endpoint) — distinct from Summary, which is the aggregation dashboard.
type Board struct {
	Columns  []BoardColumn
	Problems []domain.FileProblem // unreadable files, surfaced not swallowed (mirrors Summary)
	// Blocked holds the ids of listed tasks the repository graph says cannot be started
	// — a hard prerequisite is unmet. The board is the tool's answer to "what should I
	// do next", so answering it with work `task start` will refuse is the one thing it
	// must not do: a human eventually remembers the dependency, while an agent
	// re-derives the board every session and re-makes the same wrong choice.
	//
	// Empty when the graph is not healthy: eligibility is only meaningful over a graph
	// the tool trusts, and GraphHealth below says so rather than implying everything is
	// startable.
	Blocked map[string]bool
	// GraphHealth and GraphDetail carry the repository-wide dependency-graph verdict.
	// Degradation is latched at mutation time — `task complete` refuses on a degraded
	// graph — but was previously invisible until then, so it surfaced mid-operation in
	// an unfamiliar repo. Reporting it on the read surfaces is what lets it be fixed
	// calmly instead.
	GraphHealth GraphHealth
	GraphDetail string
}

// Board composes the active-work view from a single task scan. Every active status
// is a column (an empty status shows an empty column, not a gap), and each column
// keeps the store's task order.
func (s *Service) Board() (Board, error) {
	read, err := loadTaskGraphRecords(s.taskGraphs)
	if err != nil {
		return Board{}, err
	}
	tasks := read.Tasks
	problems := taskGraphFileProblems(read.Problems)
	byStatus := map[domain.Status][]domain.Task{}
	for _, t := range tasks {
		if t.Status.IsActive() {
			byStatus[t.Status] = append(byStatus[t.Status], t)
		}
	}
	// One graph build for the whole board: eligibility is repository-global, so asking
	// per task would re-derive the same snapshot for every row.
	graph := NewTaskGraphRead(read)
	blocked := map[string]bool{}
	if graph.Health() == GraphHealthy {
		for _, t := range tasks {
			if !t.Status.IsActive() || t.ID == "" {
				continue
			}
			// Only pending work can be "blocked": an in-progress task has already
			// started, so reporting its gate would be advice about a decision already
			// taken.
			if state := graph.State(t.ID); isPendingWorkRole(state.Role) && !state.Eligible {
				blocked[t.ID] = true
			}
		}
	}
	active := domain.ActiveStatuses()
	cols := make([]BoardColumn, len(active))
	for i, st := range active {
		cols[i] = BoardColumn{Status: st, Tasks: sortEligibleFirst(byStatus[st], blocked)}
	}
	board := Board{Columns: cols, Problems: problems, Blocked: blocked, GraphHealth: graph.Health()}
	if board.GraphHealth != GraphHealthy {
		board.GraphDetail = taskGraphHealthDetail(graph)
	}
	return board, nil
}

// sortEligibleFirst parks blocked work at the end of its column while keeping the
// store's order within each group. Marking alone still leaves the first row of the
// board unstartable; ordering is what makes the top of the list answerable.
func sortEligibleFirst(tasks []domain.Task, blocked map[string]bool) []domain.Task {
	if len(blocked) == 0 {
		return tasks
	}
	out := make([]domain.Task, 0, len(tasks))
	for _, t := range tasks {
		if !blocked[t.ID] {
			out = append(out, t)
		}
	}
	for _, t := range tasks {
		if blocked[t.ID] {
			out = append(out, t)
		}
	}
	return out
}
