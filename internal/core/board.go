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
}

// Board composes the active-work view from a single task scan. Every active status
// is a column (an empty status shows an empty column, not a gap), and each column
// keeps the store's task order.
func (s *Service) Board() (Board, error) {
	tasks, problems, err := s.store.ListTasks()
	if err != nil {
		return Board{}, err
	}
	byStatus := map[domain.Status][]domain.Task{}
	for _, t := range tasks {
		if t.Status.IsActive() {
			byStatus[t.Status] = append(byStatus[t.Status], t)
		}
	}
	active := domain.ActiveStatuses()
	cols := make([]BoardColumn, len(active))
	for i, st := range active {
		cols[i] = BoardColumn{Status: st, Tasks: byStatus[st]}
	}
	return Board{Columns: cols, Problems: problems}, nil
}
