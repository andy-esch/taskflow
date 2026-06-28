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

// loadTaskList reads the task list for the tab's current status view. The default
// view ("") is the WORKING set — active work plus deferred (snoozed tasks stay in
// view as reminders), hiding only completed/deprecated; "all" includes those
// archived states; any other value is an exact status filter. The view is
// snapshotted here so a later change can't race this load.
func loadTaskList(t *entityTab, svc *core.Service) tea.Cmd {
	view, gen := t.statusView, t.loadGen
	return func() tea.Msg {
		f := core.TaskFilter{}
		switch view {
		case "":
			f.All = true // working view: load all, drop completed/deprecated below
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
		now := svc.Now() // one clock read drives both the sort and the per-row due flag
		switch view {
		case "":
			tasks = dropArchived(tasks) // working view excludes completed/deprecated
			sortWorkingView(tasks, now) // active first, then deferred (due-for-revisit leading)
		case "revisit":
			sortByRevisitDate(tasks) // oldest-overdue first
		case string(domain.StatusDeferred):
			sortRevisitDueFirst(tasks, now) // browsing all deferred: the due ones lead
		}
		items := make([]list.Item, 0, len(tasks))
		for _, t := range tasks {
			items = append(items, taskItem{t: t, due: domain.IsTaskRevisitDue(t, now)})
		}
		return listLoadedMsg{kind: entityTasks, gen: gen, items: items, problems: problems}
	}
}

// loadDashboard reads the at-a-glance Summary for the landing screen (off the
// event loop → dashLoadedMsg) — the same core.Summary the `status` command renders.
func loadDashboard(svc *core.Service) tea.Cmd {
	return func() tea.Msg {
		s, err := svc.Summary()
		if err != nil {
			return dashLoadedMsg{err: err}
		}
		return dashLoadedMsg{summary: s}
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

// loadEpicList reads the epic roster for the tab's current status view. The default
// view ("") is the live working set — only `active` domain buckets, with dormant
// (drained) ones floated to the bottom so liveness reads at a glance; "all" spans
// every status; any other value is an exact stored-status filter (retired/
// deprecated). The view is snapshotted here so a later change can't race this load.
func loadEpicList(t *entityTab, svc *core.Service) tea.Cmd {
	view, gen := t.statusView, t.loadGen
	return func() tea.Msg {
		epics, problems, err := svc.ListEpics()
		if err != nil {
			return errMsg{kind: entityEpics, gen: gen, err: err}
		}
		epics = filterEpicsByView(epics, view)
		sortEpicsForView(epics, view)
		countsW := countsWidth(epics, func(es core.EpicSummary) (int, int) { return es.Done, es.Total })
		items := make([]list.Item, 0, len(epics))
		for _, es := range epics {
			items = append(items, epicItem{es: es, countsW: countsW})
		}
		return listLoadedMsg{kind: entityEpics, gen: gen, items: items, problems: problems}
	}
}

// filterEpicsByView narrows the roster to a status view: "" (default) is the LIVE
// set — every epic that isn't a known terminal (retired/deprecated), so it fails
// open on an unknown/foreign status rather than hiding it; "all" keeps everything;
// any other value is an exact status filter (retired/deprecated). The epic echo of
// loadTaskList's view switch, but on the stored status FIELD (epics live flat).
func filterEpicsByView(epics []core.EpicSummary, view string) []core.EpicSummary {
	if view == "all" {
		return epics
	}
	out := make([]core.EpicSummary, 0, len(epics))
	for _, e := range epics {
		keep := e.Epic.Status == view // exact match for a named terminal view
		if view == "" {               // live: anything not retired/deprecated
			keep = !domain.IsEpicArchived(e.Epic.Status)
		}
		if keep {
			out = append(out, e)
		}
	}
	return out
}

// sortEpicsForView floats live epics (working/fresh) above dormant ones in the
// default view, so a drained bucket recedes without leaving the list — the epics-tab
// echo of the dashboard's live-first lens. Stable, so the underlying store order is
// preserved within each band. Non-default views (an exact status, or "all") keep
// their order untouched.
func sortEpicsForView(epics []core.EpicSummary, view string) {
	if view != "" {
		return
	}
	sort.SliceStable(epics, func(i, j int) bool { return epics[i].Live() && !epics[j].Live() })
}

func loadEpicDetail(svc *core.Service, id string) tea.Cmd {
	return func() tea.Msg {
		es, tasks, body, err := svc.ShowEpic(id)
		if err != nil {
			return detailErrMsg{kind: entityEpics, id: id, err: err}
		}
		return detailMsg{kind: entityEpics, id: id, content: epicDetail{es: es, tasks: tasks, body: body}}
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
		countsW := countsWidth(audits, func(a domain.Audit) (int, int) { return a.Resolved(), a.Findings })
		items := make([]list.Item, 0, len(audits))
		for _, a := range audits {
			items = append(items, auditItem{a: a, countsW: countsW})
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

// dropArchived removes completed/deprecated tasks — the genuinely "done" states —
// from the working view. Deferred is NOT archived (it's "snoozed, come back"), so
// it stays in view as a reminder. Filters in place (the slice is freshly returned
// by ListTasks, so reusing its backing array is safe).
func dropArchived(tasks []domain.Task) []domain.Task {
	out := tasks[:0]
	for _, t := range tasks {
		if t.Status != domain.StatusCompleted && t.Status != domain.StatusDeprecated {
			out = append(out, t)
		}
	}
	return out
}

// sortWorkingView orders the default view: active work first (by working-set rank),
// then the deferred tail — and within deferred, the due-for-revisit ones lead, so a
// fired snooze sits right under your active work instead of buried at the bottom.
// now is the service clock, so "due" matches the marker and the core filter.
func sortWorkingView(tasks []domain.Task, now time.Time) {
	sort.SliceStable(tasks, func(i, j int) bool {
		ri, rj := rankOf(tasks[i].Status), rankOf(tasks[j].Status)
		if ri != rj {
			return ri < rj
		}
		return domain.IsTaskRevisitDue(tasks[i], now) && !domain.IsTaskRevisitDue(tasks[j], now)
	})
}

// sortRevisitDueFirst floats due-for-revisit deferred tasks to the top of the
// `:deferred` view (stable otherwise), so a snooze that came due isn't buried.
func sortRevisitDueFirst(tasks []domain.Task, now time.Time) {
	sort.SliceStable(tasks, func(i, j int) bool {
		return domain.IsTaskRevisitDue(tasks[i], now) && !domain.IsTaskRevisitDue(tasks[j], now)
	})
}

// sortByRevisitDate orders the `:revisit` view oldest-overdue first (every task
// there is already due, so the date drives the order, not the marker).
func sortByRevisitDate(tasks []domain.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].RevisitAt < tasks[j].RevisitAt
	})
}
