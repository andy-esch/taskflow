package tui

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// threadRepo seeds a planning tree with one Thread over two tasks: a dependency
// and the queued task it gates. The projection therefore has a rollup, a
// frontier, and an eligibility that a task-only edit can flip — which is what
// makes the "task changes must reload Threads too" path observable.
func threadRepo(t *testing.T) string {
	t.Helper()
	r := testutil.NewRepo(t)
	first, second := testutil.TaskID("first"), testutil.TaskID("second")
	r.Task("completed", "first.md", fmt.Sprintf(
		"---\nid: %s\nstatus: completed\ndescription: the dependency\n---\n# first\n", first))
	r.Task("next-up", "second.md", fmt.Sprintf(
		"---\nid: %s\nstatus: next-up\ndescription: the dependent\ndepends_on: [%s]\n---\n# second\n", second, first))
	r.File("threads/6g503c6pfqeb-delivery.md", fmt.Sprintf(
		"---\nschema: 1\nid: 6g503c6pfqeb\nstatus: in-progress\ndescription: the delivery thread\n"+
			"goal: ship it\ncreated: \"2026-08-29\"\ntasks: [%s, %s]\n---\n# Thread: Delivery\n\nbody\n",
		first, second))
	return r.Root
}

func threadModel(t *testing.T) (Model, string) {
	t.Helper()
	root := threadRepo(t)
	return New(core.NewService(store.NewFS(root))), root
}

// TestThreadProjectionsAreCoreValuesNotAReDerivation pins that the adapter
// stores exactly what the core service returned. Anything the TUI recomputed
// locally — ordering, readiness, health — would show up as a difference here.
func TestThreadProjectionsAreCoreValuesNotAReDerivation(t *testing.T) {
	m, _ := threadModel(t)
	wantList, wantProblems, err := m.svc.ListThreadViews()
	if err != nil {
		t.Fatal(err)
	}
	if len(wantList.Threads) != 1 {
		t.Fatalf("setup: expected one Thread in the fixture, got %d", len(wantList.Threads))
	}
	if len(wantList.Threads[0].Frontier) == 0 || len(wantList.Threads[0].Members) != 2 {
		t.Fatalf("setup: expected a 2-member Thread with a frontier, got %+v", wantList.Threads[0].Rollup)
	}

	m = drain(t, m, m.requestThreadList())
	if !m.threads.loaded || m.threads.loadErr != nil {
		t.Fatalf("thread list should have loaded: loaded=%v err=%v", m.threads.loaded, m.threads.loadErr)
	}
	if !reflect.DeepEqual(m.threads.list, wantList) {
		t.Errorf("cached list projection differs from the core service's:\n got %+v\nwant %+v", m.threads.list, wantList)
	}
	if !reflect.DeepEqual(m.threads.problems, wantProblems) {
		t.Errorf("cached read problems = %+v, want %+v", m.threads.problems, wantProblems)
	}

	wantView, wantBody, err := m.svc.ShowThread("delivery")
	if err != nil {
		t.Fatal(err)
	}
	m = drain(t, m, m.requestThreadDetail("delivery"))
	if !m.threads.detailOK || m.threads.detailErr != nil {
		t.Fatalf("thread detail should have loaded: ok=%v err=%v", m.threads.detailOK, m.threads.detailErr)
	}
	if !reflect.DeepEqual(m.threads.detail, wantView) || m.threads.detailBody != wantBody {
		t.Errorf("cached detail projection differs from the core service's:\n got %+v\nwant %+v", m.threads.detail, wantView)
	}
	if m.threads.detailRef != "delivery" {
		t.Errorf("detail ref = %q, want the requested ref", m.threads.detailRef)
	}
	if m.threads.detailLoadedRef != "delivery" {
		t.Errorf("loaded detail ref = %q, want delivery", m.threads.detailLoadedRef)
	}
}

func TestThreadDetailSeparatesRequestedAndLoadedIdentity(t *testing.T) {
	m, _ := threadModel(t)
	m = drain(t, m, m.requestThreadDetail("delivery"))
	wantView, wantBody := m.threads.detail, m.threads.detailBody

	m = drain(t, m, m.requestThreadDetail("does-not-exist"))
	if m.threads.detailRef != "does-not-exist" {
		t.Errorf("requested ref = %q, want does-not-exist", m.threads.detailRef)
	}
	if m.threads.detailLoadedRef != "delivery" {
		t.Errorf("failed selection changed loaded ref to %q", m.threads.detailLoadedRef)
	}
	if m.threads.detailOK {
		t.Error("a failed new selection claims its retained predecessor is current")
	}
	if m.threads.detailErr == nil {
		t.Error("the failed selected Thread should retain its error")
	}
	if !reflect.DeepEqual(m.threads.detail, wantView) || m.threads.detailBody != wantBody {
		t.Error("a failed new selection should keep, but not misidentify, the last coherent payload")
	}
}

// splitWorkspaceStore hands the TUI a workspace whose aggregate store, task
// graph source, and Thread store are three distinct values. The fakes return
// data the on-disk tree does NOT contain, so a Thread read that reached back
// through the aggregate store would be visible as the fixture's own values.
type splitWorkspaceStore struct {
	root    string
	store   core.Store
	graphs  *countingGraphSource
	threads *countingThreadStore
}

func (s *splitWorkspaceStore) OpenWorkspace(start string) (core.WorkspaceSource, error) {
	return core.WorkspaceSource{
		Checkout: start, PlanningRoot: s.root, PlanningID: "planning-split",
		Store: s.store, TaskGraphs: s.graphs, Threads: s.threads, Layout: noWatchLayout{},
	}, nil
}

type countingGraphSource struct {
	tasks []domain.Task
	calls int
}

func (s *countingGraphSource) ReadTaskGraph() (core.TaskGraphRead, error) {
	s.calls++
	return core.TaskGraphRead{Tasks: s.tasks}, nil
}

type countingThreadStore struct {
	threads []domain.Thread
	calls   int
}

func (s *countingThreadStore) ReadThreads() (core.ThreadRead, error) {
	s.calls++
	return core.ThreadRead{Threads: s.threads}, nil
}

func (s *countingThreadStore) GetThread(string) (domain.Thread, string, error) {
	s.calls++
	return s.threads[0], "split body\n", nil
}

// TestThreadReadsSurviveSplitWorkspaceCapabilities proves the reviewed port
// boundary is real from the TUI's side: a workspace may supply the aggregate
// store, the task graph source, and the Thread store as three independent
// capabilities, and Thread reads must come from the injected pair.
func TestThreadReadsSurviveSplitWorkspaceCapabilities(t *testing.T) {
	root := threadRepo(t)
	graphs := &countingGraphSource{tasks: []domain.Task{
		{ID: "6g5rwjqeh6a6", Slug: "split-only", Status: domain.StatusNextUp, Description: "only the graph source has this"},
	}}
	threads := &countingThreadStore{threads: []domain.Thread{{
		ID: "6g503c6pfqe1", Slug: "split", Status: domain.ThreadStatusInProgress,
		Description: "only the thread store has this", Goal: "prove the split",
		Created: "2026-08-29", Tasks: []string{"6g5rwjqeh6a6"},
	}}}
	opener := core.NewWorkspaceService(&splitWorkspaceStore{
		root: root, store: store.NewFS(root), graphs: graphs, threads: threads,
	})
	workspace, err := opener.Open(core.WorkspaceRequest{Start: root})
	if err != nil {
		t.Fatal(err)
	}

	m := New(workspace.Planning)
	m = drain(t, m, m.requestThreadList())
	if m.threads.loadErr != nil {
		t.Fatalf("split-capability thread list failed: %v", m.threads.loadErr)
	}
	if graphs.calls == 0 || threads.calls == 0 {
		t.Fatalf("both injected capabilities should have been read: graphs=%d threads=%d", graphs.calls, threads.calls)
	}
	if len(m.threads.list.Threads) != 1 || m.threads.list.Threads[0].Thread.Slug != "split" {
		t.Fatalf("thread list did not come from the injected Thread store: %+v", m.threads.list.Threads)
	}
	view := m.threads.list.Threads[0]
	if len(view.Members) != 1 || view.Members[0].Task.ID != "6g5rwjqeh6a6" {
		t.Fatalf("membership was not joined to the injected task graph source: %+v", view.Members)
	}

	m = drain(t, m, m.requestThreadDetail("split"))
	if m.threads.detailErr != nil || m.threads.detailBody != "split body\n" {
		t.Fatalf("split-capability thread detail = %q err=%v", m.threads.detailBody, m.threads.detailErr)
	}
}

func frontierIDs(view core.ThreadView) []string {
	ids := make([]string, 0, len(view.Frontier))
	for _, member := range view.Frontier {
		ids = append(ids, member.State.TaskID)
	}
	return ids
}

// drainNested resolves a command tree, applying every message a batch (or a
// batch of batches — reloadAll fans out per surface) eventually produces.
func drainNested(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		return m
	}
	msg := cmd()
	if msg == nil {
		return m
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, child := range batch {
			m = drainNested(t, m, child)
		}
		return m
	}
	tm, next := m.Update(msg)
	// Follow the reducer's own follow-up loads (a list result asks for the
	// selection's detail), so a drained model is a settled one.
	return drainNested(t, tm.(Model), next)
}

// TestThreadProjectionsReloadOnTaskAndThreadChanges pins that both durable
// prefixes reach the Thread cache through the ordinary reload path: a Thread
// document edit AND a task-only edit that changes nothing but eligibility.
func TestThreadProjectionsReloadOnTaskAndThreadChanges(t *testing.T) {
	m, root := threadModel(t)
	m = drain(t, m, m.requestThreadList())
	m = drain(t, m, m.requestThreadDetail("delivery"))
	second := testutil.TaskID("second")
	if got := frontierIDs(m.threads.list.Threads[0]); len(got) != 1 || got[0] != second {
		t.Fatalf("setup: the queued member should be dispatchable, frontier = %v", got)
	}

	// A task-only edit: no Thread file is touched, but the queued member's
	// dependency becomes unmet, so it leaves the frontier.
	first := filepath.Join(root, domain.TasksDir, testutil.TaskID("first")+"-first.md")
	testutil.Write(t, first, "---\nid: "+testutil.TaskID("first")+
		"\nstatus: next-up\ndescription: the dependency\n---\n# first\n")

	m = drainNested(t, m, m.reloadAll())
	if got := frontierIDs(m.threads.list.Threads[0]); len(got) == 1 && got[0] == second {
		t.Error("a task-only edit must reach the Thread projection: the blocked member is still dispatchable")
	}
	if got := frontierIDs(m.threads.detail); len(got) == 1 && got[0] == second {
		t.Error("the open Thread detail must reload too: the blocked member is still dispatchable")
	}

	// A Thread document edit: membership shrinks with no task file touched.
	testutil.Write(t, filepath.Join(root, "threads", "6g503c6pfqeb-delivery.md"),
		"---\nschema: 1\nid: 6g503c6pfqeb\nstatus: in-progress\ndescription: the delivery thread\n"+
			"goal: ship it\ncreated: \"2026-08-29\"\ntasks: ["+testutil.TaskID("first")+"]\n---\n# Thread: Delivery\n\nbody\n")

	m = drainNested(t, m, m.reloadAll())
	if got := len(m.threads.list.Threads[0].Members); got != 1 {
		t.Errorf("a Thread document edit must reach the projection: members = %d, want 1", got)
	}
	if got := len(m.threads.detail.Members); got != 1 {
		t.Errorf("the open Thread detail must reflect the document edit: members = %d, want 1", got)
	}
}

// TestUnrequestedThreadProjectionsAreNeverRead pins the landing-load discipline:
// like an unvisited tab, a Thread surface nobody has opened is not read on
// startup and is not re-read by every filesystem event.
func TestUnrequestedThreadProjectionsAreNeverRead(t *testing.T) {
	m, _ := threadModel(t)
	if cmd := m.reloadThreadProjections(); cmd != nil {
		t.Fatal("an unrequested Thread projection must not be reloaded")
	}
	if m.threads.listGen != 0 || m.threads.detailGen != 0 {
		t.Fatalf("generations moved without a request: list=%d detail=%d", m.threads.listGen, m.threads.detailGen)
	}
	// Only the list has been requested: the reload must not invent a detail read.
	m.requestThreadList()
	if cmd := m.reloadThreadProjections(); cmd == nil {
		t.Fatal("a requested Thread list should reload")
	}
	if m.threads.listGen != 2 || m.threads.detailGen != 0 {
		t.Errorf("reload fired a detail read for a Thread nobody has selected: list=%d detail=%d",
			m.threads.listGen, m.threads.detailGen)
	}
}

// TestStaleThreadResultsAreDropped pins the generation stamps: a slow read that
// lands after a newer request must not overwrite the newer result, and must not
// resurrect a failure the newer read already cleared.
func TestStaleThreadResultsAreDropped(t *testing.T) {
	m, _ := threadModel(t)
	m = drain(t, m, m.requestThreadList())
	fresh := m.threads.list

	stale := threadListLoadedMsg{gen: m.threads.listGen - 1, list: core.ThreadListView{}, err: domain.ErrNotFound}
	tm, _ := m.Update(stale)
	m = tm.(Model)
	if m.threads.loadErr != nil || !reflect.DeepEqual(m.threads.list, fresh) {
		t.Errorf("an older list result overwrote a newer one: err=%v", m.threads.loadErr)
	}

	m = drain(t, m, m.requestThreadDetail("delivery"))
	body := m.threads.detailBody
	tm, _ = m.Update(threadDetailLoadedMsg{gen: m.threads.detailGen - 1, ref: "delivery", body: "stale"})
	m = tm.(Model)
	if m.threads.detailBody != body {
		t.Errorf("an older detail result overwrote a newer one: body = %q", m.threads.detailBody)
	}
	// A result for a Thread that is no longer selected is stale even at the
	// current generation — the ref moved on.
	tm, _ = m.Update(threadDetailLoadedMsg{gen: m.threads.detailGen, ref: "other", body: "wrong thread"})
	m = tm.(Model)
	if m.threads.detailBody != body {
		t.Errorf("a result for another Thread landed in the cache: body = %q", m.threads.detailBody)
	}
}
