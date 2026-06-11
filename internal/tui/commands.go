package tui

import (
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// Each entity has a list loader (off the event loop → listLoadedMsg) and an item
// loader (lazy detail → detailMsg / detailErrMsg). Never call the service from
// Update/View. The registry in entity.go wires these to their tabs.

// --- tasks ---

// loadTaskList reads the task list for the tab's current status view. The
// default view ("") is the active working set (sorted in-progress→next-up→…);
// "all" includes archived; any other value is an exact status filter. The view
// is snapshotted here so a later change can't race this load.
func loadTaskList(t *entityTab, svc *core.Service) tea.Cmd {
	view := t.statusView
	return func() tea.Msg {
		f := core.TaskFilter{}
		switch view {
		case "":
			// active-only default
		case "all":
			f.All = true
		default:
			f.Status = view
		}
		tasks, problems, err := svc.ListTasks(f)
		if err != nil {
			return errMsg{err}
		}
		if view == "" {
			sortWorkingSet(tasks) // working-set order only for the default view
		}
		items := make([]list.Item, 0, len(tasks))
		for _, t := range tasks {
			items = append(items, taskItem{t})
		}
		return listLoadedMsg{kind: entityTasks, items: items, problems: problems}
	}
}

func loadTaskDetail(svc *core.Service, id string) tea.Cmd {
	return func() tea.Msg {
		t, body, err := svc.ShowTask(id)
		if err != nil {
			return detailErrMsg{kind: entityTasks, id: id, err: err}
		}
		return detailMsg{kind: entityTasks, id: id, content: taskDetail{t: t, body: body}}
	}
}

// --- epics ---

func loadEpicList(_ *entityTab, svc *core.Service) tea.Cmd {
	return func() tea.Msg {
		epics, problems, err := svc.ListEpics()
		if err != nil {
			return errMsg{err}
		}
		items := make([]list.Item, 0, len(epics))
		for _, es := range epics {
			items = append(items, epicItem{es})
		}
		return listLoadedMsg{kind: entityEpics, items: items, problems: problems}
	}
}

func loadEpicDetail(svc *core.Service, id string) tea.Cmd {
	return func() tea.Msg {
		e, tasks, body, err := svc.ShowEpic(id)
		if err != nil {
			return detailErrMsg{kind: entityEpics, id: id, err: err}
		}
		return detailMsg{kind: entityEpics, id: id, content: epicDetail{e: e, tasks: tasks, body: body}}
	}
}

// --- audits ---

// loadAuditList reads open audits (the working set). The audits tab mirrors the
// CLI's default of open-only; interactive sort still applies.
func loadAuditList(_ *entityTab, svc *core.Service) tea.Cmd {
	return func() tea.Msg {
		audits, problems, err := svc.ListAudits("", false)
		if err != nil {
			return errMsg{err}
		}
		items := make([]list.Item, 0, len(audits))
		for _, a := range audits {
			items = append(items, auditItem{a})
		}
		return listLoadedMsg{kind: entityAudits, items: items, problems: problems}
	}
}

func loadAuditDetail(svc *core.Service, id string) tea.Cmd {
	return func() tea.Msg {
		a, body, err := svc.ShowAudit(id)
		if err != nil {
			return detailErrMsg{kind: entityAudits, id: id, err: err}
		}
		return detailMsg{kind: entityAudits, id: id, content: auditDetail{a: a, body: body}}
	}
}

// statusRank orders statuses for a "what am I doing" view: active work first,
// archived last. (Default scan is active-only; archived views come in S2b.)
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
