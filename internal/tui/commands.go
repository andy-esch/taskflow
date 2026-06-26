package tui

import (
	"sort"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

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
	view, gen := t.statusView, t.loadGen
	return func() tea.Msg {
		f := core.TaskFilter{}
		switch view {
		case "":
			// active-only default
		case "all":
			f.All = true
		case "revisit":
			f.RevisitDue = true // synthetic view: deferred tasks whose snooze date has arrived
		default:
			f.Status = view
		}
		tasks, problems, err := svc.ListTasks(f)
		if err != nil {
			return errMsg{kind: entityTasks, gen: gen, err: err}
		}
		switch view {
		case "":
			sortWorkingSet(tasks) // working-set order only for the default view
		case "revisit":
			sortByRevisitDate(tasks) // oldest-overdue first
		case string(domain.StatusDeferred):
			sortRevisitDueFirst(tasks) // browsing all deferred: the due ones lead
		}
		items := make([]list.Item, 0, len(tasks))
		for _, t := range tasks {
			items = append(items, taskItem{t})
		}
		return listLoadedMsg{kind: entityTasks, gen: gen, items: items, problems: problems}
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

func loadEpicList(t *entityTab, svc *core.Service) tea.Cmd {
	gen := t.loadGen
	return func() tea.Msg {
		epics, problems, err := svc.ListEpics()
		if err != nil {
			return errMsg{kind: entityEpics, gen: gen, err: err}
		}
		items := make([]list.Item, 0, len(epics))
		for _, es := range epics {
			items = append(items, epicItem{es})
		}
		return listLoadedMsg{kind: entityEpics, gen: gen, items: items, problems: problems}
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

// loadAuditList reads the audits for the tab's current bucket view. The default
// view ("") is the open bucket (the working set, matching the CLI default); "all"
// spans every bucket; any other value is an exact bucket (closed/deferred). The
// view is snapshotted here so a later change can't race this load.
func loadAuditList(t *entityTab, svc *core.Service) tea.Cmd {
	view, gen := t.statusView, t.loadGen
	return func() tea.Msg {
		bucket, all := view, false
		switch view {
		case "":
			// open-only default: bucket "" + all=false
		case "all":
			bucket, all = "", true
		}
		audits, problems, err := svc.ListAudits(bucket, all)
		if err != nil {
			return errMsg{kind: entityAudits, gen: gen, err: err}
		}
		items := make([]list.Item, 0, len(audits))
		for _, a := range audits {
			items = append(items, auditItem{a})
		}
		return listLoadedMsg{kind: entityAudits, gen: gen, items: items, problems: problems}
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

// rankOf returns a status's working-set rank. An unrecognized status (a foreign or
// legacy word the loader tolerates) sorts LAST — a bare map index would give it
// rank 0 and float it up among in-progress work.
func rankOf(s domain.Status) int {
	if r, ok := statusRank[s]; ok {
		return r
	}
	return len(statusRank)
}

func sortWorkingSet(tasks []domain.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		return rankOf(tasks[i].Status) < rankOf(tasks[j].Status)
	})
}

// revisitDue reports whether a task is parked in deferred AND its revisit date has
// arrived — the same predicate `task list --revisit-due` uses, evaluated against
// the wall clock at load time (refreshed on every reload, like the relative dates).
func revisitDue(t domain.Task) bool {
	return t.Status == domain.StatusDeferred && domain.IsRevisitDue(t.RevisitAt, time.Now())
}

// sortRevisitDueFirst floats due-for-revisit deferred tasks to the top of the
// `:deferred` view (stable otherwise), so a snooze that came due isn't buried.
func sortRevisitDueFirst(tasks []domain.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		return revisitDue(tasks[i]) && !revisitDue(tasks[j])
	})
}

// sortByRevisitDate orders the `:revisit` view oldest-overdue first (every task
// there is already due, so the date drives the order, not the marker).
func sortByRevisitDate(tasks []domain.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].RevisitAt < tasks[j].RevisitAt
	})
}
