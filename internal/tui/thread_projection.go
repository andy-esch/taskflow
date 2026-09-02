package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
)

// threadProjectionState is the adapter's typed cache of the core Thread read
// models. It adds only asynchronous UI bookkeeping; lifecycle, readiness,
// health, gates, and diagnostics remain owned by core.ThreadListView/ThreadView,
// so the next slice can render these values without recreating their semantics.
//
// A zero value means "no Thread surface has asked yet": like an unvisited tab,
// an unrequested projection is never read, and the generations below only ever
// count requests that some consumer actually made.
type threadProjectionState struct {
	list     core.ThreadListView
	problems []core.ThreadReadProblem
	loaded   bool
	loadErr  error // last list failure; the values above are the last good load (or none)
	listGen  int   // bumped per list request; older results and retries are dropped

	detail          core.ThreadView
	detailBody      string
	detailRef       string // latest requested ref
	detailLoadedRef string // ref owning detail/detailBody; independent of selection
	detailOK        bool   // the latest requested ref has a coherent loaded projection
	detailErr       error  // last detail failure for detailRef
	detailGen       int    // bumped per detail request; older results and retries are dropped
}

type threadListLoadedMsg struct {
	gen      int
	list     core.ThreadListView
	problems []core.ThreadReadProblem
	err      error
}

func (msg threadListLoadedMsg) readError() error { return msg.err }

type threadDetailLoadedMsg struct {
	gen  int
	ref  string
	view core.ThreadView
	body string
	err  error
}

func (msg threadDetailLoadedMsg) readError() error { return msg.err }

// loadThreadList reads the shared list projection off the event loop. It returns
// core's values untouched: no local traversal, ordering, or readiness rules.
func loadThreadList(svc *core.Service, gen int) tea.Cmd {
	return func() tea.Msg {
		list, problems, err := svc.ListThreadViews()
		return threadListLoadedMsg{gen: gen, list: list, problems: problems, err: err}
	}
}

// loadThreadDetail reads one Thread's projection and body through the same core
// service method the CLI's `thread show` uses.
func loadThreadDetail(svc *core.Service, ref string, gen int) tea.Cmd {
	return func() tea.Msg {
		view, body, err := svc.ShowThread(ref)
		return threadDetailLoadedMsg{gen: gen, ref: ref, view: view, body: body, err: err}
	}
}

// requestThreadList is the entry point a Thread surface calls to (re)read the
// list. Bumping first is what makes an older result — or a retry scheduled for
// it — stale rather than something that can overwrite this request.
func (m *Model) requestThreadList() tea.Cmd {
	m.threads.listGen++
	request := readRequest{surface: readThreadList, gen: m.threads.listGen}
	return withReadConflictRetry(request, loadThreadList(m.svc, request.gen))
}

// requestThreadDetail loads one Thread's projection, replacing whichever ref was
// selected before. The cache keeps the last coherent detail until the new one
// lands; deciding what to show meanwhile belongs to the visible slice.
func (m *Model) requestThreadDetail(ref string) tea.Cmd {
	m.threads.detailGen++
	m.threads.detailRef = ref
	m.threads.detailOK = ref != "" && ref == m.threads.detailLoadedRef
	m.threads.detailErr = nil
	request := readRequest{surface: readThreadDetail, id: ref, gen: m.threads.detailGen}
	return withReadConflictRetry(request, loadThreadDetail(m.svc, ref, request.gen))
}

// reloadThreadProjections is the fs/`r` path: it re-reads exactly the Thread
// projections some surface has already asked for. Threads project over both
// Thread documents AND the repository-global task graph, so a task-only edit
// reaches them through this same reload — eligibility, frontier, and gates can
// change with no Thread file touched at all.
func (m *Model) reloadThreadProjections() tea.Cmd {
	var cmds []tea.Cmd
	if m.threads.listGen > 0 {
		cmds = append(cmds, m.requestThreadList())
	}
	if m.threads.detailGen > 0 && m.threads.detailRef != "" {
		cmds = append(cmds, m.requestThreadDetail(m.threads.detailRef))
	}
	return tea.Batch(cmds...)
}
