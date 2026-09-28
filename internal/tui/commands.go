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
		if view != "" && view != "all" && view != "revisit" {
			if _, err := domain.ParseStatus(view); err != nil {
				return errMsg{kind: entityTasks, gen: gen, err: err}
			}
		}
		// One complete read lets the registry reject an ID duplicated in a hidden
		// status before this view makes its other occurrence actionable.
		records, problems, err := svc.ListTasks(core.TaskFilter{All: true})
		if err != nil {
			return errMsg{kind: entityTasks, gen: gen, err: err}
		}
		fullRefs := make([]entityRef, 0, len(records))
		for _, record := range records {
			fullRefs = append(fullRefs, entityRef{key: record.Source.ID, label: record.Value.Slug})
		}
		identityErr := validateEntityRefs(fullRefs)
		now := svc.Now() // one clock read drives both the sort and the per-row due flag
		switch view {
		case "":
			records = dropArchivedRecords(records) // working view excludes completed/deprecated
			sortWorkingRecords(records, now)       // active first, then deferred (due-for-revisit leading)
		case "revisit":
			records = filterTaskRecords(records, func(task domain.Task) bool { return domain.IsTaskRevisitDue(task, now) })
			sortRecordsByRevisitDate(records) // oldest-overdue first
		case string(domain.StatusDeferred):
			records = filterTaskRecords(records, func(task domain.Task) bool { return task.Status == domain.StatusDeferred })
			sortRecordsRevisitDueFirst(records, now) // browsing all deferred: the due ones lead
		case "all":
		default:
			records = filterTaskRecords(records, func(task domain.Task) bool { return string(task.Status) == view })
		}
		items := make([]list.Item, 0, len(records))
		refs := make([]entityRef, 0, len(records))
		for _, record := range records {
			refs = append(refs, entityRef{key: record.Source.ID, label: record.Value.Slug})
		}
		hints := duplicateIdentityHints(refs)
		for _, record := range records {
			t := record.Value
			items = append(items, taskItem{t: t, sourceID: record.Source.ID, due: domain.IsTaskRevisitDue(t, now), identityHint: hints[record.Source.ID]})
		}
		return listLoadedMsg{kind: entityTasks, gen: gen, items: items, problems: problems, identityErr: identityErr}
	}
}

func filterTaskRecords(records []core.LoadedRecord[domain.Task], keep func(domain.Task) bool) []core.LoadedRecord[domain.Task] {
	out := records[:0]
	for _, record := range records {
		if keep(record.Value) {
			out = append(out, record)
		}
	}
	return out
}

// loadDashboard reads the at-a-glance Summary for the landing screen (off the
// event loop → dashLoadedMsg) — the same core.Summary the `status` command renders.
func loadDashboard(svc *core.Service, gen int) tea.Cmd {
	return func() tea.Msg {
		s, err := svc.Summary()
		if err != nil {
			return dashLoadedMsg{gen: gen, err: err}
		}
		return dashLoadedMsg{gen: gen, summary: s}
	}
}

func loadTaskDetail(svc *core.Service, id string) tea.Cmd {
	return func() tea.Msg {
		record, err := svc.ShowTask(id)
		if err != nil {
			return detailErrMsg{kind: entityTasks, id: id, err: err}
		}
		return detailMsg{kind: entityTasks, id: id, sourceID: record.Source.ID, content: taskDetail{t: record.Value.Task, body: record.Value.Body}}
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
		fullRefs := make([]entityRef, 0, len(epics))
		for _, epic := range epics {
			fullRefs = append(fullRefs, entityRef{key: epic.Source.ID, label: epic.Epic.ID})
		}
		identityErr := validateEntityRefs(fullRefs)
		epics = filterEpicsByView(epics, view)
		sortEpicsForView(epics, view)
		countsW := countsWidth(epics, func(es core.EpicSummary) (int, int) { return es.Done, es.Total })
		items := make([]list.Item, 0, len(epics))
		refs := make([]entityRef, 0, len(epics))
		for _, es := range epics {
			refs = append(refs, entityRef{key: es.Source.ID, label: es.Epic.ID})
		}
		hints := duplicateIdentityHints(refs)
		for _, es := range epics {
			items = append(items, epicItem{es: es, countsW: countsW, identityHint: hints[es.Source.ID]})
		}
		return listLoadedMsg{kind: entityEpics, gen: gen, items: items, problems: problems, identityErr: identityErr}
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
		detail, err := svc.ShowEpic(id)
		if err != nil {
			return detailErrMsg{kind: entityEpics, id: id, err: err}
		}
		return detailMsg{kind: entityEpics, id: id, sourceID: detail.Summary.Source.ID,
			content: epicDetail{es: detail.Summary, tasks: detail.Tasks, body: detail.Body}}
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
		if view != "" && view != "all" {
			if _, err := domain.ParseAuditBucket(view); err != nil {
				return errMsg{kind: entityAudits, gen: gen, err: err}
			}
		}
		records, problems, err := svc.ListAudits("", true)
		if err != nil {
			return errMsg{kind: entityAudits, gen: gen, err: err}
		}
		fullRefs := make([]entityRef, 0, len(records))
		for _, record := range records {
			fullRefs = append(fullRefs, entityRef{key: record.Source.ID, label: record.Value.Slug})
		}
		identityErr := validateEntityRefs(fullRefs)
		if view != "all" {
			bucket := view
			if bucket == "" {
				bucket = string(domain.AuditOpen)
			}
			records = filterAuditRecords(records, bucket)
		}
		countsW := countsWidth(records, func(record core.LoadedRecord[domain.Audit]) (int, int) {
			return record.Value.Resolved(), record.Value.Findings
		})
		items := make([]list.Item, 0, len(records))
		refs := make([]entityRef, 0, len(records))
		for _, record := range records {
			refs = append(refs, entityRef{key: record.Source.ID, label: record.Value.Slug})
		}
		hints := duplicateIdentityHints(refs)
		for _, record := range records {
			items = append(items, auditItem{a: record.Value, sourceID: record.Source.ID,
				countsW: countsW, identityHint: hints[record.Source.ID]})
		}
		return listLoadedMsg{kind: entityAudits, gen: gen, items: items, problems: problems, identityErr: identityErr}
	}
}

func filterAuditRecords(records []core.LoadedRecord[domain.Audit], bucket string) []core.LoadedRecord[domain.Audit] {
	out := records[:0]
	for _, record := range records {
		if string(record.Value.Bucket) == bucket {
			out = append(out, record)
		}
	}
	return out
}

func loadAuditDetail(svc *core.Service, id string) tea.Cmd {
	return func() tea.Msg {
		record, err := svc.ShowAudit(id)
		if err != nil {
			return detailErrMsg{kind: entityAudits, id: id, err: err}
		}
		return detailMsg{kind: entityAudits, id: id, sourceID: record.Source.ID, content: auditDetail{a: record.Value.Audit, body: record.Value.Body}}
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

// Keep the source envelope attached while filtering and sorting rows. Copying
// Source.ID into domain.FilenameID would make identity depend on a local-field
// compatibility shim and lose the adapter's explicit source contract.
func dropArchivedRecords(records []core.LoadedRecord[domain.Task]) []core.LoadedRecord[domain.Task] {
	out := records[:0]
	for _, record := range records {
		if record.Value.Status != domain.StatusCompleted && record.Value.Status != domain.StatusDeprecated {
			out = append(out, record)
		}
	}
	return out
}

func workingTaskLess(left, right domain.Task, now time.Time) bool {
	leftRank, rightRank := rankOf(left.Status), rankOf(right.Status)
	if leftRank != rightRank {
		return leftRank < rightRank
	}
	return domain.IsTaskRevisitDue(left, now) && !domain.IsTaskRevisitDue(right, now)
}

// sortWorkingView orders the default view: active work first (by working-set rank),
// then the deferred tail — and within deferred, the due-for-revisit ones lead, so a
// fired snooze sits right under your active work instead of buried at the bottom.
// now is the service clock, so "due" matches the marker and the core filter.
func sortWorkingView(tasks []domain.Task, now time.Time) {
	sort.SliceStable(tasks, func(i, j int) bool { return workingTaskLess(tasks[i], tasks[j], now) })
}

func sortWorkingRecords(records []core.LoadedRecord[domain.Task], now time.Time) {
	sort.SliceStable(records, func(i, j int) bool { return workingTaskLess(records[i].Value, records[j].Value, now) })
}

// sortRecordsRevisitDueFirst leads the `:deferred` view with due tasks.
func sortRecordsRevisitDueFirst(records []core.LoadedRecord[domain.Task], now time.Time) {
	sort.SliceStable(records, func(i, j int) bool {
		return domain.IsTaskRevisitDue(records[i].Value, now) && !domain.IsTaskRevisitDue(records[j].Value, now)
	})
}

// sortRecordsByRevisitDate orders the `:revisit` view oldest-overdue first.
func sortRecordsByRevisitDate(records []core.LoadedRecord[domain.Task]) {
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].Value.RevisitAt < records[j].Value.RevisitAt
	})
}

// loadResearchList loads the whole corpus. Unlike tasks/audits there is no view axis to
// snapshot — research has no status, so there is no working-set-vs-all distinction and
// `s`/`S` are inert on this tab. Order is the service's: newest first, slug-tiebroken.
func loadResearchList(t *entityTab, svc *core.Service) tea.Cmd {
	gen := t.loadGen
	return func() tea.Msg {
		records, problems, err := svc.ListResearch("")
		if err != nil {
			return errMsg{kind: entityResearch, gen: gen, err: err}
		}
		items := make([]list.Item, 0, len(records))
		refs := make([]entityRef, 0, len(records))
		for _, record := range records {
			refs = append(refs, entityRef{key: record.Source.ID, label: record.Value.Slug})
		}
		hints := duplicateIdentityHints(refs)
		for _, record := range records {
			items = append(items, researchItem{r: record.Value, sourceID: record.Source.ID, identityHint: hints[record.Source.ID]})
		}
		return listLoadedMsg{kind: entityResearch, gen: gen, items: items, problems: problems}
	}
}

func loadResearchDetail(svc *core.Service, id string) tea.Cmd {
	return func() tea.Msg {
		record, err := svc.ShowResearch(id)
		if err != nil {
			return detailErrMsg{kind: entityResearch, id: id, err: err}
		}
		return detailMsg{kind: entityResearch, id: id, sourceID: record.Source.ID, content: researchDetail{r: record.Value.Research, body: record.Value.Body}}
	}
}
