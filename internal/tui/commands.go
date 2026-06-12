package tui

import (
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// loadTasks reads the active task list via the service (off the event loop) and
// sorts it into working-set order. Never call the service from Update/View.
func loadTasks(svc *core.Service) tea.Cmd {
	return func() tea.Msg {
		tasks, problems, err := svc.ListTasks(core.TaskFilter{})
		if err != nil {
			return errMsg{err}
		}
		sortWorkingSet(tasks)
		items := make([]list.Item, 0, len(tasks))
		for _, t := range tasks {
			items = append(items, taskItem{t})
		}
		return tasksLoadedMsg{items: items, problems: problems}
	}
}

// loadBody reads one task's body for the detail pane (lazy, on selection). A
// failure (e.g. an ambiguous duplicate slug) becomes a bodyErrMsg shown in the
// detail pane, not a fatal errMsg that would blank the browser.
func loadBody(svc *core.Service, slug string) tea.Cmd {
	return func() tea.Msg {
		t, body, err := svc.ShowTask(slug)
		if err != nil {
			return bodyErrMsg{slug: slug, err: err}
		}
		return taskBodyMsg{slug: slug, task: t, body: body}
	}
}

// statusRank orders statuses for a "what am I doing" view: active work first,
// archived last. (Default scan is active-only; archived views come in sprint 2.)
var statusRank = map[domain.Status]int{
	domain.StatusInProgress:   0,
	domain.StatusNextUp:       1,
	domain.StatusReadyToStart: 2,
	domain.StatusCompleted:    3,
	domain.StatusDeprecated:   4,
	domain.StatusDeferred:     5,
}

func sortWorkingSet(tasks []domain.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		return statusRank[tasks[i].Status] < statusRank[tasks[j].Status]
	})
}
