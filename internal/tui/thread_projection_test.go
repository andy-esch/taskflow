package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"unsafe"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	clirender "github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/graphfmt"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/theme"
)

// threadRepo is the same semantic projection shape exercised by core and CLI:
// one completed prerequisite and one queued dependent member. A task-only edit
// can therefore flip the Thread frontier without touching its document.
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

func threadTab(m Model) *entityTab { return m.tabs[indexOfKind(m.tabs, entityThreads)] }

func openThreads(t *testing.T, m Model) Model {
	t.Helper()
	cmd := m.switchTab(indexOfKind(m.tabs, entityThreads))
	return drainNested(t, m, cmd)
}

func selectedThreadView(t *testing.T, m Model) core.ThreadView {
	t.Helper()
	it, ok := m.cur().list.SelectedItem().(threadItem)
	if !ok {
		t.Fatalf("selected Thread row = %T", m.cur().list.SelectedItem())
	}
	return it.view
}

func selectedThreadDetail(t *testing.T, m Model) threadDetail {
	t.Helper()
	detail, ok := m.detail.content.(threadDetail)
	if !ok {
		t.Fatalf("Thread detail = %T", m.detail.content)
	}
	return detail
}

func frontierIDs(view core.ThreadView) []string {
	ids := make([]string, 0, len(view.Frontier))
	for _, member := range view.Frontier {
		ids = append(ids, member.State.TaskID)
	}
	return ids
}

// drainNested resolves a command tree and every follow-up load. Thread list
// success naturally chains into the ordinary registry detail request.
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
	return drainNested(t, tm.(Model), next)
}

func TestThreadsUseOneRegistryListAndDetailOwner(t *testing.T) {
	m, _ := threadModel(t)
	tab := threadTab(m)
	if tab.loaded || tab.loadGen != 0 {
		t.Fatalf("an unvisited Thread tab read eagerly: loaded=%v gen=%d", tab.loaded, tab.loadGen)
	}

	wantList, wantProblems, err := m.svc.ListThreadViews()
	if err != nil {
		t.Fatal(err)
	}
	m = openThreads(t, m)
	tab = threadTab(m)
	if m.cur() != tab || !tab.loaded || tab.loadGen != 1 || tab.loadErr != nil {
		t.Fatalf("Thread route did not settle through registry: active=%q loaded=%v gen=%d err=%v",
			m.cur().name, tab.loaded, tab.loadGen, tab.loadErr)
	}
	gotList := make([]core.ThreadView, 0, len(tab.list.Items()))
	for _, raw := range tab.list.Items() {
		gotList = append(gotList, raw.(threadItem).view)
	}
	if !reflect.DeepEqual(gotList, wantList.Threads) {
		t.Fatalf("registry rows changed core projections:\n got %+v\nwant %+v", gotList, wantList.Threads)
	}
	if tab.threadDiagnostics == nil ||
		!reflect.DeepEqual(tab.threadDiagnostics.graphProblems, wantList.GraphProblems) ||
		!reflect.DeepEqual(tab.threadDiagnostics.readProblems, wantProblems) {
		t.Fatalf("repository diagnostics were not retained on the registry tab: %+v", tab.threadDiagnostics)
	}

	wantProjection, wantBody, err := m.svc.ShowThreadGraphDetail("delivery")
	if err != nil {
		t.Fatal(err)
	}
	detail := selectedThreadDetail(t, m)
	if !reflect.DeepEqual(detail.projection, wantProjection) || detail.body != wantBody || m.detail.loadedKey != wantProjection.View.Thread.CanonicalID() {
		t.Fatalf("registry detail changed core projection/body: key=%q detail=%+v", m.detail.loadedKey, detail)
	}
	if detail.path == "" || m.selectedPath() != detail.path {
		t.Fatal("local Thread path capability did not reach the selected registry detail")
	}

	var route, jump bool
	for _, item := range m.paletteIndex() {
		route = route || item.kind == palCommand && item.word == "threads"
		jump = jump || item.kind == palJump && item.ek == entityThreads && item.ref.key == wantProjection.View.Thread.CanonicalID()
	}
	if !route || !jump {
		t.Fatalf("Thread route/item missing from palette: route=%v jump=%v", route, jump)
	}
}

// splitWorkspaceStore supplies aggregate records, graph reads, and Thread reads
// independently. Omitting ThreadPaths is intentional: portable Thread browsing
// must not inherit the aggregate filesystem path resolver.
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
	threads  []domain.Thread
	problems []core.ThreadReadProblem
	getErr   error
	calls    int
}

func (s *countingThreadStore) ReadThreads() (core.ThreadRead, error) {
	s.calls++
	return core.ThreadRead{Threads: s.threads, Problems: s.problems}, nil
}

func (s *countingThreadStore) GetThread(string) (domain.Thread, string, error) {
	s.calls++
	if s.getErr != nil {
		return domain.Thread{}, "", s.getErr
	}
	return s.threads[0], "split body\n", nil
}

type tuiThreadPathFake struct{ path string }

func (f tuiThreadPathFake) ResolveThreadPath(string) (string, error) { return f.path, nil }

func TestThreadRouteSurvivesSplitPathlessCapabilities(t *testing.T) {
	root := threadRepo(t)
	graphs := &countingGraphSource{tasks: []domain.Task{{
		ID: "6g5rwjqeh6a6", Slug: "split-only", Status: domain.StatusNextUp,
		Description: "only the graph source has this",
	}}}
	threads := &countingThreadStore{
		threads: []domain.Thread{{
			ID: "6g503c6pfqe1", Slug: "split", Status: domain.ThreadStatusInProgress,
			Description: "only the thread store has this", Goal: "prove the split",
			Created: "2026-08-29", Tasks: []string{"6g5rwjqeh6a6"},
		}},
		problems: []core.ThreadReadProblem{{ThreadSlug: "broken", Location: "remote://thread", Message: "unreadable"}},
	}
	opener := core.NewWorkspaceService(&splitWorkspaceStore{
		root: root, store: store.NewFS(root), graphs: graphs, threads: threads,
	})
	workspace, err := opener.Open(core.WorkspaceRequest{Start: root})
	if err != nil {
		t.Fatal(err)
	}

	m := openThreads(t, New(workspace.Planning))
	if graphs.calls == 0 || threads.calls < 2 {
		t.Fatalf("split capabilities were not used: graph=%d threads=%d", graphs.calls, threads.calls)
	}
	if got := selectedThreadView(t, m); got.Thread.Slug != "split" || got.Members[0].Task.ID != "6g5rwjqeh6a6" {
		t.Fatalf("Thread route escaped the injected capabilities: %+v", got)
	}
	if got := threadTab(m).threadDiagnostics; got == nil || len(got.readProblems) != 1 {
		t.Fatalf("portable read diagnostics were lost: %+v", got)
	}
	detail := selectedThreadDetail(t, m)
	if detail.body != "split body\n" || detail.path != "" || detail.pathIssue == "" {
		t.Fatalf("pathless detail = body %q path %q issue %q", detail.body, detail.path, detail.pathIssue)
	}
	tm, cmd := m.yankSelectedPath()
	m = tm.(Model)
	if cmd != nil || !m.flashErr || !strings.Contains(m.flash, "local path unavailable") {
		t.Fatalf("pathless yank did not degrade explicitly: flash=%q cmd=%v", m.flash, cmd != nil)
	}
	tm, cmd = m.openInEditor()
	m = tm.(Model)
	if cmd != nil || !m.flashErr || !strings.Contains(m.flash, "local path unavailable") {
		t.Fatalf("pathless editor did not degrade explicitly: flash=%q cmd=%v", m.flash, cmd != nil)
	}
	tm, cmd = m.Update(press("e"))
	m = tm.(Model)
	if cmd != nil || !m.flashErr || !strings.Contains(m.flash, "editing is unavailable") {
		t.Fatalf("pathless inline edit did not explain the read-only capability: flash=%q cmd=%v", m.flash, cmd != nil)
	}
}

func TestLocalThreadPathSurvivesSemanticDetailFailure(t *testing.T) {
	thread := domain.Thread{
		ID: testutil.TaskID("repair-thread"), Slug: "repair-thread", Status: domain.ThreadStatusUnstarted,
		Description: "repair me", Goal: "retain local navigation", Created: "2026-09-02",
	}
	threads := &countingThreadStore{threads: []domain.Thread{thread}, getErr: domain.ErrValidation}
	svc := core.NewService(nil,
		core.WithThreadStore(threads),
		core.WithTaskGraphSource(&countingGraphSource{}),
		core.WithThreadPathSource(tuiThreadPathFake{path: "/planning/threads/repair-thread.md"}),
	)
	m := openThreads(t, New(svc))
	if m.detail.content != nil || m.detail.loadedKey != thread.CanonicalID() {
		t.Fatalf("semantic error pane lost selection identity: key=%q content=%T", m.detail.loadedKey, m.detail.content)
	}
	if got := m.selectedPath(); got != "/planning/threads/repair-thread.md" {
		t.Fatalf("semantic detail failure lost local repair path: %q", got)
	}
	tm, copyCmd := m.yankSelectedPath()
	m = tm.(Model)
	if copyCmd == nil || m.flashErr {
		t.Fatalf("repair path was not copyable: flash=%q cmd=%v", m.flash, copyCmd != nil)
	}
	tm, editorCmd := m.openInEditor()
	m = tm.(Model)
	if editorCmd == nil || m.flashErr {
		t.Fatalf("repair path was not openable: flash=%q cmd=%v", m.flash, editorCmd != nil)
	}
}

func TestThreadRegistryReloadsOnTaskAndThreadChanges(t *testing.T) {
	m, root := threadModel(t)
	m = openThreads(t, m)
	second := testutil.TaskID("second")
	if got := frontierIDs(selectedThreadView(t, m)); len(got) != 1 || got[0] != second {
		t.Fatalf("setup frontier = %v", got)
	}

	first := filepath.Join(root, domain.TasksDir, testutil.TaskID("first")+"-first.md")
	testutil.Write(t, first, "---\nid: "+testutil.TaskID("first")+
		"\nstatus: next-up\ndescription: the dependency\n---\n# first\n")
	before := threadTab(m).loadGen
	m = drainNested(t, m, m.reloadAll())
	if threadTab(m).loadGen != before+1 {
		t.Fatalf("one watcher reload advanced Thread generation by %d", threadTab(m).loadGen-before)
	}
	if got := frontierIDs(selectedThreadView(t, m)); len(got) == 1 && got[0] == second {
		t.Error("task-only edit did not refresh the Thread list projection")
	}
	if got := frontierIDs(selectedThreadDetail(t, m).projection.View); len(got) == 1 && got[0] == second {
		t.Error("task-only edit did not refresh the selected Thread detail")
	}

	testutil.Write(t, filepath.Join(root, "threads", "6g503c6pfqeb-delivery.md"),
		"---\nschema: 1\nid: 6g503c6pfqeb\nstatus: in-progress\ndescription: the delivery thread\n"+
			"goal: ship it\ncreated: \"2026-08-29\"\ntasks: ["+testutil.TaskID("first")+"]\n---\n# Thread: Delivery\n\nbody\n")
	m = drainNested(t, m, m.reloadAll())
	if got := len(selectedThreadView(t, m).Members); got != 1 {
		t.Fatalf("Thread-document edit left %d list members", got)
	}
	if got := len(selectedThreadDetail(t, m).projection.View.Members); got != 1 {
		t.Fatalf("Thread-document edit left %d detail members", got)
	}
}

func TestThreadSpatialReloadAddsAndRenamesNodesWithoutLosingSelection(t *testing.T) {
	m, root := threadModel(t)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m = openThreads(t, tm.(Model))
	m.setFocus(focusDetail)
	for range 2 {
		tm, _ = m.Update(press("v"))
		m = tm.(Model)
	}
	tm, _ = m.Update(press("l"))
	m = tm.(Model)
	selected := testutil.TaskID("second")
	if got := selectedThreadDetail(t, m).detailSelectionKey(); got != selected {
		t.Fatalf("setup spatial selection=%q want %q", got, selected)
	}

	third := testutil.TaskID("watch-added-third")
	testutil.Write(t, filepath.Join(root, domain.TasksDir, third+"-watch-added-third.md"), fmt.Sprintf(
		"---\nid: %s\nstatus: next-up\ndescription: added while the graph is open\ndepends_on: [%s]\n---\n# third\n", third, selected))
	testutil.Write(t, filepath.Join(root, "threads", "6g503c6pfqeb-delivery.md"), fmt.Sprintf(
		"---\nschema: 1\nid: 6g503c6pfqeb\nstatus: in-progress\ndescription: the delivery thread\n"+
			"goal: ship it\ncreated: \"2026-08-29\"\ntasks: [%s, %s, %s]\n---\n# Thread: Delivery\n\nbody\n",
		testutil.TaskID("first"), selected, third))
	m = drainNested(t, m, m.reloadAll())
	detail := selectedThreadDetail(t, m)
	if detail.detailViewName() != string(threadDetailSpatial) || detail.detailSelectionKey() != selected || !m.zoom {
		t.Fatalf("added node disturbed spatial context: view=%q selected=%q zoom=%v", detail.detailViewName(), detail.detailSelectionKey(), m.zoom)
	}
	if _, ok := spatialPlacement(buildThreadSpatialLayout(detail.projection), third); !ok {
		t.Fatal("reload did not add the new Thread member to the open spatial graph")
	}

	oldPath := filepath.Join(root, domain.TasksDir, selected+"-second.md")
	newPath := filepath.Join(root, domain.TasksDir, selected+"-renamed-second.md")
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
	m = drainNested(t, m, m.reloadAll())
	detail = selectedThreadDetail(t, m)
	if detail.detailSelectionKey() != selected {
		t.Fatalf("rename changed stable spatial selection to %q", detail.detailSelectionKey())
	}
	placement, ok := spatialPlacement(buildThreadSpatialLayout(detail.projection), selected)
	if !ok || placement.node.Label != "renamed-second" {
		t.Fatalf("renamed node did not refresh in place: %+v present=%v", placement.node, ok)
	}
}

func TestThreadRegistryPreservesSelectionFilterAndSortState(t *testing.T) {
	m, root := threadModel(t)
	testutil.Write(t, filepath.Join(root, "threads", "6g503c6pfqec-second-thread.md"),
		"---\nschema: 1\nid: 6g503c6pfqec\nstatus: unstarted\ndescription: another initiative\n"+
			"goal: test registry behavior\ncreated: \"2026-08-28\"\ntasks: []\n---\n# Second Thread\n")
	m = openThreads(t, m)
	tab := threadTab(m)
	delivery := entityRef{key: "6g503c6pfqeb", label: "delivery"}
	if !tab.selectByKey(delivery.key) {
		t.Fatal("could not select delivery by canonical Thread id")
	}
	tm, cmd := m.afterSelectionChange("", nil)
	m = drainNested(t, tm.(Model), cmd)

	if fv := m.cur().list.SelectedItem().(threadItem).FilterValue(); !strings.Contains(fv, "delivery") || !strings.Contains(fv, "ship it") || !strings.Contains(fv, "in-progress") {
		t.Fatalf("Thread filter vocabulary = %q", fv)
	}
	tab.sortKey = sortSlug
	m = drainNested(t, m, m.applySortToCurrent())
	if m.selectedKey() != delivery.key {
		t.Fatal("sorting lost the canonical Thread selection")
	}
	tab.list.SetFilterText("delivery")
	m = drainNested(t, m, tab.reload(m.svc, tab.markReload()))
	if m.selectedKey() != delivery.key || tab.sortKey != sortSlug || tab.list.FilterValue() != "delivery" {
		t.Fatalf("reload lost Thread state: selected=%q sort=%v filter=%q",
			m.selectedKey(), tab.sortKey, tab.list.FilterValue())
	}
}

func TestThreadRowsDegradeWithoutLosingEssentialState(t *testing.T) {
	done := domain.Task{ID: testutil.TaskID("done"), Slug: "done", Status: domain.StatusCompleted}
	active := domain.Task{ID: testutil.TaskID("active"), Slug: "active", Status: domain.StatusInProgress}
	queued := domain.Task{ID: testutil.TaskID("queued"), Slug: "queued", Status: domain.StatusNextUp}
	taskIDs := []string{active.ID, done.ID, queued.ID}
	sort.Strings(taskIDs)
	thread := domain.Thread{
		ID: testutil.TaskID("initiative"), Slug: "migrate-configuration-subsystem-phase-one", Status: domain.ThreadStatusInProgress,
		Description: "deliver the feature", Goal: "ship", Created: "2026-09-02",
		Tasks: taskIDs,
	}
	view := core.ProjectThread(thread, core.NewTaskGraph([]domain.Task{done, active, queued}, nil))
	it := threadItem{view: view, countsW: 3}
	other := it
	other.view.Thread.ID = testutil.TaskID("initiative-two")
	other.view.Thread.FilenameID = other.view.Thread.ID
	other.view.Thread.Slug = "migrate-configuration-subsystem-phase-two"

	for _, width := range []int{72, 48, 26} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			l := list.New([]list.Item{it, other}, threadDelegate{st: &testStyles}, width, 5)
			render := func(index int, item threadItem) string {
				var out strings.Builder
				threadDelegate{st: &testStyles}.Render(&out, l, index, item)
				return ansi.Strip(out.String())
			}
			plain, otherPlain := render(0, it), render(1, other)
			for _, want := range []string{"▶1", "✓1", "×0"} {
				if !strings.Contains(plain, want) {
					t.Errorf("width %d lost %q: %q", width, want, plain)
				}
			}
			if width >= 64 && (!strings.Contains(plain, "d:") || !strings.Contains(plain, "s:")) {
				t.Errorf("width %d lost nominal/sound progress: %q", width, plain)
			}
			if ansi.StringWidth(plain) > width {
				t.Errorf("row width %d overflowed: %q", width, plain)
			}
			if plain == otherPlain {
				t.Errorf("width %d collapsed distinct long Thread identities: %q", width, plain)
			}
		})
	}
}

func TestThreadRowsProtectDuplicateIdentityHints(t *testing.T) {
	base := core.ThreadView{Thread: domain.Thread{
		ID: "6g503c6pfqe1", FilenameID: "6g503c6pfqe1", Slug: "same-very-long-thread-name",
		Status: domain.ThreadStatusInProgress,
	}, GraphHealth: core.GraphHealthy, ProjectionHealth: core.GraphHealthy}
	first := threadItem{view: base, identityHint: "6g503c6pfqe1"}
	second := first
	second.view.Thread.ID, second.view.Thread.FilenameID = "6g503c6pfqe2", "6g503c6pfqe2"
	second.identityHint = "6g503c6pfqe2"
	l := list.New([]list.Item{first, second}, threadDelegate{st: &testStyles}, 26, 5)
	render := func(index int, item threadItem) string {
		var out strings.Builder
		threadDelegate{st: &testStyles}.Render(&out, l, index, item)
		return ansi.Strip(out.String())
	}
	a, b := render(0, first), render(1, second)
	if a == b || !strings.Contains(a, "e1]") || !strings.Contains(b, "e2]") {
		t.Fatalf("narrow duplicate hints lost their distinguishing tail:\nA %q\nB %q", a, b)
	}
}

func TestThreadRowsUseCellAwareBudgetsForUnicodeAndLargeCounts(t *testing.T) {
	view := core.ThreadView{Thread: domain.Thread{
		ID: "6g503c6pfqez", FilenameID: "6g503c6pfqez",
		Slug: "移行-configuration-subsystem-phase-終端", Status: domain.ThreadStatusCompleted,
	}, GraphHealth: core.GraphDegraded, ProjectionHealth: core.GraphBroken, Inconsistent: true}
	for i := 0; i < 120; i++ {
		view.Members = append(view.Members, core.ThreadTaskView{State: core.TaskGraphState{
			TaskID: fmt.Sprintf("member-%03d", i), Role: core.RoleQueued, Gate: core.GateClear,
		}})
	}
	item := threadItem{view: view}
	for _, width := range []int{64, 42, 26} {
		l := list.New([]list.Item{item}, threadDelegate{st: &testStyles}, width, 5)
		var out strings.Builder
		threadDelegate{st: &testStyles}.Render(&out, l, 0, item)
		plain := ansi.Strip(out.String())
		if got := ansi.StringWidth(plain); got > width {
			t.Errorf("width %d rendered %d terminal cells: %q", width, got, plain)
		}
		if !strings.Contains(plain, "×") {
			t.Errorf("width %d lost not-dispatchable work: %q", width, plain)
		}
	}
}

func TestThreadActivityUsesAuthoritativeFrontierOnUnhealthyGraph(t *testing.T) {
	legacyRoot := domain.Task{ID: testutil.TaskID("legacy-root"), Slug: "legacy-root", Status: domain.StatusCompleted}
	legacyUser := domain.Task{
		ID: testutil.TaskID("legacy-user"), Slug: "legacy-user", Status: domain.StatusNextUp,
		LegacyBlockedBy: []string{legacyRoot.Slug},
	}
	queued := domain.Task{ID: testutil.TaskID("clear-queued"), Slug: "clear-queued", Status: domain.StatusNextUp}
	ready := domain.Task{ID: testutil.TaskID("clear-ready"), Slug: "clear-ready", Status: domain.StatusReadyToStart}
	active := domain.Task{ID: testutil.TaskID("active"), Slug: "active", Status: domain.StatusInProgress}
	graph := core.NewTaskGraph([]domain.Task{legacyRoot, legacyUser, queued, ready, active}, nil)
	thread := domain.Thread{
		ID: testutil.TaskID("unhealthy-thread"), FilenameID: testutil.TaskID("unhealthy-thread"),
		Slug: "unhealthy-thread", Status: domain.ThreadStatusInProgress,
		Description: "global graph evidence prevents dispatch", Goal: "show every pending member",
		Created: "2026-09-03", Tasks: []string{queued.ID, ready.ID, active.ID},
	}
	view := core.ProjectThread(thread, graph)
	activity := activityForThread(view)
	if view.GraphHealth == core.GraphHealthy || len(view.Frontier) != 0 {
		t.Fatalf("fixture did not separate clear gates from dispatchability: health=%s frontier=%d", view.GraphHealth, len(view.Frontier))
	}
	if activity.inFlight != 1 || activity.dispatchable != len(view.Frontier) || activity.notDispatchable != 2 {
		t.Fatalf("activity did not account for authoritative frontier: %+v", activity)
	}
	meta := ansi.Strip(renderThreadMeta(threadDetail{projection: core.ThreadGraphProjection{View: view}}, 100, &testStyles))
	if !strings.Contains(meta, "0 dispatchable · 2 pending not dispatchable") {
		t.Fatalf("detail hid globally unsafe pending work:\n%s", meta)
	}
}

func TestThreadDetailKeepsNominalAndSoundProgressDistinct(t *testing.T) {
	missing := testutil.TaskID("missing-prerequisite")
	done := domain.Task{
		ID: testutil.TaskID("nominal-only"), Slug: "nominal-only", Status: domain.StatusCompleted,
		DependsOn: []string{missing},
	}
	thread := domain.Thread{
		ID: testutil.TaskID("unsound"), FilenameID: testutil.TaskID("unsound"), Slug: "unsound",
		Status: domain.ThreadStatusCompleted, Description: "nominal is not sound", Goal: "show the difference",
		Created: "2026-09-03", Tasks: []string{done.ID},
	}
	view := core.ProjectThread(thread, core.NewTaskGraph([]domain.Task{done}, nil))
	meta := ansi.Strip(renderThreadMeta(threadDetail{projection: core.ThreadGraphProjection{View: view}}, 100, &testStyles))
	if !strings.Contains(meta, "1/1 nominally done · 0/1 soundly drained") {
		t.Fatalf("detail collapsed nominal and sound progress:\n%s", meta)
	}
}

func TestThreadRegistryIsReadOnly(t *testing.T) {
	m, _ := threadModel(t)
	m = openThreads(t, m)
	tab := threadTab(m)
	if len(tab.transitions) != 0 || tab.applyMove != nil {
		t.Fatalf("Thread registry armed mutations: transitions=%v applyMove=%v", tab.transitions, tab.applyMove != nil)
	}
	footer := ansi.Strip(m.footer())
	if strings.Contains(footer, "move") || strings.Contains(footer, " edit") {
		t.Fatalf("Thread footer advertised mutation actions: %q", footer)
	}
	options := make(map[string]bool)
	for _, option := range m.commandOptions() {
		options[option] = true
	}
	for _, transition := range taskTransitions {
		if options[transition.verb] {
			t.Errorf("Thread command completion advertised task lifecycle verb %q", transition.verb)
		}
	}
	tm, _ := m.Update(press("m"))
	m = tm.(Model)
	if m.action.active {
		t.Fatal("Thread action key opened a lifecycle menu")
	}
}

func TestThreadInlineEditExplainsPathLoading(t *testing.T) {
	m, _ := threadModel(t)
	cmd := m.switchTab(indexOfKind(m.tabs, entityThreads))
	if cmd == nil {
		t.Fatal("first Thread visit did not request the list")
	}
	tm, _ := m.Update(cmd()) // apply the list, but deliberately hold its detail command
	m = tm.(Model)
	if !m.detail.loading || m.selectedRef().empty() {
		t.Fatalf("fixture did not stop with a selected Thread path loading: loading=%v selected=%+v", m.detail.loading, m.selectedRef())
	}
	tm, cmd = m.Update(press("e"))
	m = tm.(Model)
	if cmd != nil || !m.flashErr || !strings.Contains(m.flash, "path is still loading") {
		t.Fatalf("inline edit did not explain the loading path: flash=%q cmd=%v", m.flash, cmd != nil)
	}
}

func TestThreadRegistryDropsOutOfOrderListAndDetailMessages(t *testing.T) {
	m, _ := threadModel(t)
	m = openThreads(t, m)
	tab := threadTab(m)
	wantItems := append([]list.Item(nil), tab.list.Items()...)
	wantDiagnostics := tab.threadDiagnostics
	wantKey := m.detail.loadedKey
	wantDetail := selectedThreadDetail(t, m)

	tm, _ := m.Update(listLoadedMsg{
		kind: entityThreads, gen: tab.loadGen - 1, items: nil,
		threadDiagnostics: &threadListDiagnostics{graphHealth: core.GraphBroken},
	})
	m = tm.(Model)
	if !reflect.DeepEqual(tab.list.Items(), wantItems) || tab.threadDiagnostics != wantDiagnostics {
		t.Fatal("stale Thread list message replaced the current registry state")
	}

	stale := wantDetail
	stale.body = "stale body"
	tm, _ = m.Update(detailMsg{kind: entityThreads, id: wantKey, gen: m.detailGen - 1, content: stale})
	m = tm.(Model)
	if got := selectedThreadDetail(t, m).body; got != wantDetail.body {
		t.Fatalf("stale same-Thread detail landed: %q", got)
	}
	tm, _ = m.Update(detailMsg{kind: entityThreads, id: "another-thread", gen: m.detailGen, content: stale})
	m = tm.(Model)
	if got := selectedThreadDetail(t, m).body; got != wantDetail.body {
		t.Fatalf("current-generation wrong-Thread detail landed: %q", got)
	}
}

func TestThreadDetailPresentsCoreProjectionAndBody(t *testing.T) {
	m, _ := threadModel(t)
	m = openThreads(t, m)
	detail := selectedThreadDetail(t, m)
	meta := ansi.Strip(detail.meta(100, &testStyles))
	for _, want := range []string{
		"lifecycle:", "health:", "progress:", "work:", "goal:",
		"Dispatchable frontier", "Members (persisted order)",
	} {
		if !strings.Contains(meta, want) {
			t.Errorf("Thread detail omitted %q:\n%s", want, meta)
		}
	}
	if detail.rawBody() != "# Thread: Delivery\n\nbody\n" {
		t.Fatalf("persisted body = %q", detail.rawBody())
	}
}

func TestThreadDetailCyclesToTopologyAndPreservesItAcrossReload(t *testing.T) {
	m, _ := threadModel(t)
	m = openThreads(t, m)
	m.setFocus(focusDetail)
	if selectedThreadDetail(t, m).detailViewName() != string(threadDetailSummary) {
		t.Fatal("Thread detail did not start in the summary view")
	}

	tm, cmd := m.Update(press("v"))
	m = tm.(Model)
	if cmd != nil || selectedThreadDetail(t, m).detailViewName() != string(threadDetailTopology) {
		t.Fatalf("v did not switch to topology: cmd=%v view=%q", cmd != nil, selectedThreadDetail(t, m).detailViewName())
	}
	if got, want := selectedThreadDetail(t, m).detailSelectionKey(), testutil.TaskID("first"); got != want {
		t.Fatalf("topology cursor=%q want first visual task %q", got, want)
	}
	plain := ansi.Strip(m.detail.styled)
	for _, want := range []string{
		"view:     topology", "Wave 1", "Wave 2", "[prerequisite] ─▶ [dependent]", "needs [", "› ",
		testutil.TaskID("first"), testutil.TaskID("second"),
	} {
		if !strings.Contains(plain, want) {
			t.Errorf("topology view omitted %q:\n%s", want, plain)
		}
	}
	footer := ansi.Strip(m.detailFooterBody())
	if !strings.Contains(footer, "v spatial") || !strings.Contains(footer, "f tasks") ||
		!strings.Contains(footer, "j/k task") || !strings.Contains(footer, "⏎ open") || !strings.Contains(footer, "y copy task") ||
		strings.Contains(footer, "raw/pretty") || strings.Contains(footer, "j/k scroll") {
		t.Fatalf("topology footer did not describe the active controls: %q", footer)
	}

	tm, cmd = m.Update(press("j"))
	m = tm.(Model)
	if cmd != nil {
		t.Fatal("topology cursor movement unexpectedly returned a command")
	}
	if got, want := selectedThreadDetail(t, m).detailSelectionKey(), testutil.TaskID("second"); got != want {
		t.Fatalf("j selected %q want %q", got, want)
	}
	if line, ok := selectedThreadDetail(t, m).detailSelectionLine(ansi.Strip(m.detail.styled)); !ok ||
		!strings.Contains(strings.Split(ansi.Strip(m.detail.styled), "\n")[line], testutil.TaskID("second")) {
		t.Fatalf("visible cursor did not move to the second task:\n%s", ansi.Strip(m.detail.styled))
	}

	// A watcher refresh may change graph evidence without changing the selected
	// Thread. The pane retains the reader's chosen representation and stable task
	// selection rather than snapping back to summary/the first row.
	m = drainNested(t, m, m.reloadAll())
	if selectedThreadDetail(t, m).detailViewName() != string(threadDetailTopology) {
		t.Fatal("same-Thread reload discarded the topology view")
	}
	if got, want := selectedThreadDetail(t, m).detailSelectionKey(), testutil.TaskID("second"); got != want {
		t.Fatalf("same-Thread reload changed topology cursor from %q to %q", want, got)
	}

	// Resize while focused exercises the same content through the minimum-width
	// single-pane path; cycling back still restores the persisted markdown body.
	tm, _ = m.Update(tea.WindowSizeMsg{Width: 20, Height: 8})
	m = tm.(Model)
	_ = m.View()
	tm, _ = m.Update(press("v"))
	m = tm.(Model)
	if selectedThreadDetail(t, m).detailViewName() != string(threadDetailSpatial) || !m.zoom {
		t.Fatalf("topology did not cycle into the immersive spatial view: view=%q zoom=%v",
			selectedThreadDetail(t, m).detailViewName(), m.zoom)
	}
	tm, _ = m.Update(press("v"))
	m = tm.(Model)
	detail := selectedThreadDetail(t, m)
	if detail.detailViewName() != string(threadDetailSummary) || detail.rawBody() != "# Thread: Delivery\n\nbody\n" {
		t.Fatalf("summary was not restored after resize: view=%q body=%q", detail.detailViewName(), detail.rawBody())
	}
}

func TestThreadSpatialGraphUsesDirectionalStableIdentityNavigation(t *testing.T) {
	m, _ := threadModel(t)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m = openThreads(t, tm.(Model))
	m.setFocus(focusDetail)

	// summary → topology → spatial. The presentation, rather than the root
	// model knowing about Threads, requests the full content region.
	tm, _ = m.Update(press("v"))
	m = tm.(Model)
	tm, _ = m.Update(press("v"))
	m = tm.(Model)
	if got := selectedThreadDetail(t, m).detailViewName(); got != string(threadDetailSpatial) || !m.zoom || m.focus != focusDetail {
		t.Fatalf("spatial entry = view %q zoom=%v focus=%v", got, m.zoom, m.focus)
	}
	plain := ansi.Strip(m.detail.styled)
	for _, want := range []string{
		"spatial graph", "prerequisite ─▶ dependent", "status", "roles", "╔ gate",
		"focus", "about", "the dependency", "┌", "▶", "┐",
	} {
		if !strings.Contains(plain, want) {
			t.Errorf("spatial graph omitted %q:\n%s", want, plain)
		}
	}
	footer := ansi.Strip(m.footer())
	if !strings.Contains(footer, "hjkl node") || !strings.Contains(footer, "⏎ open") ||
		!strings.Contains(footer, "y copy task") || !strings.Contains(footer, "esc waves") || strings.Contains(footer, "j/k scroll") {
		t.Fatalf("spatial footer did not describe its controls: %q", footer)
	}

	first, second := testutil.TaskID("first"), testutil.TaskID("second")
	if got := selectedThreadDetail(t, m).detailSelectionKey(); got != first {
		t.Fatalf("initial spatial selection=%q want %q", got, first)
	}
	tm, _ = m.Update(press("l"))
	m = tm.(Model)
	if got := selectedThreadDetail(t, m).detailSelectionKey(); got != second {
		t.Fatalf("l selected %q want connected dependent %q", got, second)
	}
	m = drainNested(t, m, m.reloadAll())
	if selectedThreadDetail(t, m).detailViewName() != string(threadDetailSpatial) ||
		selectedThreadDetail(t, m).detailSelectionKey() != second || !m.zoom {
		t.Fatalf("same-Thread reload discarded spatial context: view=%q selected=%q zoom=%v",
			selectedThreadDetail(t, m).detailViewName(), selectedThreadDetail(t, m).detailSelectionKey(), m.zoom)
	}
	tm, _ = m.Update(press("h"))
	m = tm.(Model)
	if got := selectedThreadDetail(t, m).detailSelectionKey(); got != first {
		t.Fatalf("h selected %q want connected prerequisite %q", got, first)
	}
	tm, _ = m.Update(press("l"))
	m = tm.(Model)

	// Enter and ctrl+o keep the canonical node identity plus the presentation
	// context, rather than returning to an arbitrary Thread summary.
	tm, cmd := m.Update(press("enter"))
	m = drainNested(t, tm.(Model), cmd)
	if m.cur().kind != entityTasks || m.selectedKey() != second {
		t.Fatalf("spatial enter did not open task %q: kind=%v selected=%q", second, m.cur().kind, m.selectedKey())
	}
	tm, cmd = m.Update(press("ctrl+o"))
	m = drainNested(t, tm.(Model), cmd)
	if m.cur().kind != entityThreads || selectedThreadDetail(t, m).detailViewName() != string(threadDetailSpatial) ||
		selectedThreadDetail(t, m).detailSelectionKey() != second || !m.zoom {
		t.Fatalf("ctrl+o did not restore spatial context: kind=%v view=%q selected=%q zoom=%v",
			m.cur().kind, selectedThreadDetail(t, m).detailViewName(), selectedThreadDetail(t, m).detailSelectionKey(), m.zoom)
	}

	// Esc is the explicit complexity step-down, not an application quit or a
	// jump to the unrelated list pane.
	tm, cmd = m.Update(press("esc"))
	m = tm.(Model)
	if cmd != nil || selectedThreadDetail(t, m).detailViewName() != string(threadDetailTopology) || m.zoom || m.focus != focusDetail {
		t.Fatalf("Esc did not return to waves: cmd=%v view=%q zoom=%v focus=%v",
			cmd != nil, selectedThreadDetail(t, m).detailViewName(), m.zoom, m.focus)
	}
}

func TestThreadSpatialGraphPreservesManualZoomAndOwnsNoZoomKey(t *testing.T) {
	m, _ := threadModel(t)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m = openThreads(t, tm.(Model))
	m.setFocus(focusDetail)

	// A user-entered zoom predates the spatial presentation and remains theirs.
	tm, _ = m.Update(press("z"))
	m = tm.(Model)
	if !m.zoom || m.immersiveZoom {
		t.Fatalf("manual zoom ownership = zoom:%v immersive:%v", m.zoom, m.immersiveZoom)
	}
	for range 2 {
		tm, _ = m.Update(press("v"))
		m = tm.(Model)
	}
	if got := selectedThreadDetail(t, m).detailViewName(); got != string(threadDetailSpatial) ||
		!m.zoom || m.immersiveZoom {
		t.Fatalf("spatial entry consumed manual zoom: view=%q zoom=%v immersive=%v", got, m.zoom, m.immersiveZoom)
	}

	// Spatial navigation owns z, so it cannot accidentally rewrite shell zoom
	// ownership while the presentation is immersive.
	tm, _ = m.Update(press("z"))
	m = tm.(Model)
	if !m.zoom || m.immersiveZoom {
		t.Fatalf("z changed spatial zoom ownership: zoom=%v immersive=%v", m.zoom, m.immersiveZoom)
	}
	tm, _ = m.Update(press("esc"))
	m = tm.(Model)
	if got := selectedThreadDetail(t, m).detailViewName(); got != string(threadDetailTopology) || !m.zoom || m.immersiveZoom {
		t.Fatalf("spatial retreat consumed manual zoom: view=%q zoom=%v immersive=%v", got, m.zoom, m.immersiveZoom)
	}
}

func TestThreadStructuredDetailYankCopiesHighlightedTask(t *testing.T) {
	m, _ := threadModel(t)
	m = openThreads(t, m)
	threadSlug := m.selectedLabel()

	// List focus still means the Thread row itself.
	tm, cmd := m.Update(press("y"))
	m = tm.(Model)
	if cmd == nil || m.flash != "copied slug: "+threadSlug {
		t.Fatalf("Thread-list yank = %q cmd=%v, want parent Thread", m.flash, cmd != nil)
	}

	m.setFocus(focusDetail)
	tm, _ = m.Update(press("v")) // summary → topology
	m = tm.(Model)
	first := testutil.TaskID("first")
	if got := selectedThreadDetail(t, m).detailSelectionKey(); got != first {
		t.Fatalf("topology selection=%q want %q", got, first)
	}
	tm, cmd = m.Update(press("y"))
	m = tm.(Model)
	if cmd == nil || m.flash != "copied slug: first" {
		t.Fatalf("topology yank = %q cmd=%v, want highlighted task", m.flash, cmd != nil)
	}

	tm, _ = m.Update(press("v")) // topology → spatial
	m = tm.(Model)
	tm, _ = m.Update(press("l"))
	m = tm.(Model)
	if got, want := selectedThreadDetail(t, m).detailSelectionKey(), testutil.TaskID("second"); got != want {
		t.Fatalf("spatial selection=%q want %q", got, want)
	}
	tm, cmd = m.Update(press("y"))
	m = tm.(Model)
	if cmd == nil || m.flash != "copied slug: second" {
		t.Fatalf("spatial yank = %q cmd=%v, want highlighted task", m.flash, cmd != nil)
	}

	// An unreadable supplied node cannot be opened as a task, but its stable ID
	// remains more useful than silently copying the parent Thread.
	projection := hostileThreadGraphProjection()
	missing := ""
	for _, node := range projection.Nodes {
		if node.State.Role == core.RoleUnknown {
			missing = node.TaskID
			break
		}
	}
	detail := threadDetail{projection: projection, view: threadDetailSpatial, selection: missing}
	if text, label, ok := detail.detailSelectionYankRef(); !ok || text != missing || label != "id" {
		t.Fatalf("unreadable-node yank = (%q, %q, %v), want stable id", text, label, ok)
	}
}

func TestThreadSpatialLayoutPlacesExternalGateBetweenMemberWaves(t *testing.T) {
	before := domain.Task{ID: testutil.TaskID("before-gate"), Slug: "before-gate", Status: domain.StatusCompleted}
	gate := domain.Task{
		ID: testutil.TaskID("middle-gate"), Slug: "middle-gate", Status: domain.StatusCompleted,
		DependsOn: []string{before.ID},
	}
	after := domain.Task{
		ID: testutil.TaskID("after-gate"), Slug: "after-gate", Status: domain.StatusNextUp,
		DependsOn: []string{gate.ID},
	}
	thread := domain.Thread{
		ID: testutil.TaskID("gate-layout-thread"), FilenameID: testutil.TaskID("gate-layout-thread"),
		Slug: "gate-layout", Status: domain.ThreadStatusInProgress, Created: "2026-09-08",
		Tasks: []string{before.ID, after.ID},
	}
	projection := core.ProjectThreadGraph(thread, core.NewTaskGraph([]domain.Task{after, gate, before}, nil))
	layout := buildThreadSpatialLayout(projection)
	beforeNode, beforeOK := spatialPlacement(layout, before.ID)
	gateNode, gateOK := spatialPlacement(layout, gate.ID)
	afterNode, afterOK := spatialPlacement(layout, after.ID)
	if !beforeOK || !gateOK || !afterOK {
		t.Fatalf("layout omitted supplied nodes: before=%v gate=%v after=%v", beforeOK, gateOK, afterOK)
	}
	if beforeNode.column >= gateNode.column || gateNode.column >= afterNode.column {
		t.Fatalf("external gate did not preserve left-to-right dependency order: before=%d gate=%d after=%d",
			beforeNode.column, gateNode.column, afterNode.column)
	}
	if gateNode.node.Role != core.ThreadTaskExternalGate || !strings.Contains(strings.Join(layout.columnLabels, " "), "external") {
		t.Fatalf("interposed gate lost its bounded role/label: node=%+v labels=%v", gateNode, layout.columnLabels)
	}
}

func TestThreadSpatialLayoutPlacesCascadedExternalGatesAtOneStableFixedPoint(t *testing.T) {
	memberA, memberB, memberC, memberD := "member-a", "member-b", "member-c", "member-d"
	// Canonical presentation order visits the upstream gate first. Its
	// downstream gate moves later in pass one, forcing the upstream placement
	// to settle on a subsequent fixed-point pass.
	upstreamGate, downstreamGate := "a-upstream-gate", "z-downstream-gate"
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: memberA, Role: core.ThreadTaskMember},
			{TaskID: memberB, Role: core.ThreadTaskMember},
			{TaskID: memberC, Role: core.ThreadTaskMember},
			{TaskID: memberD, Role: core.ThreadTaskMember},
			{TaskID: upstreamGate, Role: core.ThreadTaskExternalGate},
			{TaskID: downstreamGate, Role: core.ThreadTaskExternalGate},
		},
		Edges: []core.ThreadGraphEdge{
			{From: memberA, To: memberB},
			{From: memberB, To: memberC},
			{From: memberC, To: memberD},
			{From: upstreamGate, To: downstreamGate},
			{From: downstreamGate, To: memberD},
		},
		Waves: []core.ThreadGraphWave{
			{Index: 1, TaskIDs: []string{memberA}},
			{Index: 2, TaskIDs: []string{memberB}},
			{Index: 3, TaskIDs: []string{memberC}},
			{Index: 4, TaskIDs: []string{memberD}},
		},
	}
	layout := buildThreadSpatialLayout(projection)
	upstream := layout.byID[upstreamGate]
	downstream := layout.byID[downstreamGate]
	member := layout.byID[memberD]
	if upstream.column >= downstream.column || downstream.column >= member.column {
		t.Fatalf("cascaded gates did not settle inside their dependency intervals: upstream=%d downstream=%d member=%d",
			upstream.column, downstream.column, member.column)
	}
	if downstream.column != member.column-1 || upstream.column != downstream.column-1 {
		t.Fatalf("cascaded gates did not settle beside their dependents: upstream=%d downstream=%d member=%d",
			upstream.column, downstream.column, member.column)
	}

	permuted := projection
	permuted.Nodes = []core.ThreadGraphNode{
		projection.Nodes[5], projection.Nodes[4], projection.Nodes[3], projection.Nodes[2], projection.Nodes[1], projection.Nodes[0],
	}
	permuted.Edges = []core.ThreadGraphEdge{
		projection.Edges[4], projection.Edges[3], projection.Edges[2], projection.Edges[1], projection.Edges[0],
	}
	if got := buildThreadSpatialLayout(permuted); !reflect.DeepEqual(got, layout) {
		t.Fatalf("cascaded gate placement changed under equivalent evidence permutation:\n got: %#v\nwant: %#v", got, layout)
	}
}

func TestThreadSpatialColumnLabelsDistinguishLayoutLayersFromMemberWaves(t *testing.T) {
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: "a", Role: core.ThreadTaskMember},
			{TaskID: "b", Role: core.ThreadTaskMember},
			{TaskID: "gate", Role: core.ThreadTaskExternalGate},
		},
		Edges: []core.ThreadGraphEdge{{From: "a", To: "gate"}, {From: "gate", To: "b"}},
		Waves: []core.ThreadGraphWave{
			{Index: 1, TaskIDs: []string{"a"}}, {Index: 2, TaskIDs: []string{"b"}},
		},
	}
	labels := buildThreadSpatialLayout(projection).columnLabels
	if len(labels) != 3 {
		t.Fatalf("column labels=%v want three layout layers", labels)
	}
	for index, label := range labels {
		if !strings.HasPrefix(label, fmt.Sprintf("layer %d", index+1)) {
			t.Errorf("column %d label %q misrepresents a layout layer as a wave", index+1, label)
		}
	}
	if !strings.Contains(labels[0], "wave 1") || !strings.Contains(labels[1], "external") ||
		!strings.Contains(labels[2], "wave 2") {
		t.Fatalf("layer labels lost actual wave/role evidence: %v", labels)
	}
}

func TestThreadSpatialColumnLabelsDoNotInventContiguousWaveRanges(t *testing.T) {
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: "a", Role: core.ThreadTaskMember},
			{TaskID: "b", Role: core.ThreadTaskMember},
			{TaskID: "c", Role: core.ThreadTaskMember},
		},
		Waves: []core.ThreadGraphWave{
			{Index: 1, TaskIDs: []string{"a"}},
			{Index: 3, TaskIDs: []string{"b"}},
			{Index: 5, TaskIDs: []string{"c"}},
		},
	}
	byNode := map[string]core.ThreadGraphNode{
		"a": projection.Nodes[0], "b": projection.Nodes[1], "c": projection.Nodes[2],
	}
	labels := labelThreadSpatialColumns([][]string{{"a", "b", "c"}}, projection, byNode)
	if len(labels) != 1 || !strings.Contains(labels[0], "waves 1+3+5") || strings.Contains(labels[0], "1–5") {
		t.Fatalf("non-contiguous wave label=%v want the exact supplied wave set", labels)
	}
}

func TestThreadSpatialLayoutPullsSourceGateBesideItsFirstDependent(t *testing.T) {
	a := testutil.TaskID("early-member")
	b := testutil.TaskID("middle-member")
	c := testutil.TaskID("late-member")
	gate := testutil.TaskID("late-external-gate")
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: a, Role: core.ThreadTaskMember},
			{TaskID: gate, Role: core.ThreadTaskExternalGate},
			{TaskID: b, Role: core.ThreadTaskMember},
			{TaskID: c, Role: core.ThreadTaskMember},
		},
		Edges: []core.ThreadGraphEdge{
			{From: a, To: b},
			{From: b, To: c},
			{From: gate, To: c},
		},
	}
	layout := buildThreadSpatialLayout(projection)
	if got, want := layout.byID[gate].column, layout.byID[c].column-1; got != want {
		t.Fatalf("source gate column=%d want immediately before dependent column %d", got, want)
	}
	if layout.byID[gate].column <= layout.byID[a].column {
		t.Fatalf("late gate remained beside unrelated graph roots: gate=%d early=%d", layout.byID[gate].column, layout.byID[a].column)
	}
}

func TestThreadSpatialHorizontalNavigationPrefersGraphEdgesAcrossVisibleColumns(t *testing.T) {
	a := testutil.TaskID("skip-source")
	d := testutil.TaskID("layer-source")
	b := testutil.TaskID("nearest-prerequisite")
	c := testutil.TaskID("selected-dependent")
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: a, Label: "skip source", Role: core.ThreadTaskMember},
			{TaskID: d, Label: "layer source", Role: core.ThreadTaskMember},
			{TaskID: b, Label: "nearest prerequisite", Role: core.ThreadTaskMember},
			{TaskID: c, Label: "selected dependent", Role: core.ThreadTaskMember},
		},
		Edges: []core.ThreadGraphEdge{
			{From: a, To: c}, // direct, but skips the nearest visible column
			{From: d, To: b},
			{From: b, To: c}, // direct and in the nearest visible column
		},
	}
	layout := buildThreadSpatialLayout(projection)
	if layout.byID[a].column != 0 || layout.byID[b].column != 1 || layout.byID[c].column != 2 {
		t.Fatalf("fixture did not produce three presentation columns: a=%d b=%d c=%d", layout.byID[a].column, layout.byID[b].column, layout.byID[c].column)
	}
	if got := threadSpatialMove(projection, c, -1, 0); got != b {
		t.Fatalf("h did not choose the nearest directly connected column: got %q want %q", got, b)
	}
	if got := threadSpatialMove(projection, b, -1, 0); got != d {
		t.Fatalf("second h did not follow the prerequisite edge: got %q want %q", got, d)
	}
	if got := threadSpatialMove(projection, a, 1, 0); got != c {
		t.Fatalf("l stopped on unrelated task in the next visible column: got %q want directly connected %q", got, c)
	}
}

func TestThreadSpatialCrossingPreservesProductionRoutesAndArrowheads(t *testing.T) {
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: "a", Role: core.ThreadTaskMember}, {TaskID: "b", Role: core.ThreadTaskMember},
			{TaskID: "c", Role: core.ThreadTaskMember}, {TaskID: "d", Role: core.ThreadTaskMember},
		},
		Edges: []core.ThreadGraphEdge{{From: "a", To: "d"}, {From: "b", To: "c"}},
	}
	layout := buildThreadSpatialLayout(projection)
	canvas := renderThreadSpatialCanvasWindow(projection, layout, "a", 0, 0, layout.width, layout.height)
	crossings := 0
	for row := range canvas.cells {
		for column := range canvas.cells[row] {
			cell := canvas.cells[row][column]
			if cell.conflict || cell.overlap {
				t.Fatalf("production route collision at (%d,%d) was not a straight crossing: %+v", column, row, cell)
			}
			if cell.crossing {
				crossings++
			}
		}
	}
	if crossings == 0 {
		t.Fatal("crossed-edge fixture did not produce a pass-through crossing")
	}
	for _, route := range layout.routes {
		cell := canvas.cells[route.arrow.y][route.arrow.x]
		if cell.text != string(route.arrowRune) || cell.color != theme.ColorYellow {
			t.Errorf("production route %s -> %s endpoint=%+v want yellow %q",
				route.edge.From, route.edge.To, cell, route.arrowRune)
		}
		for index := 1; index < len(route.segments); index++ {
			corner := route.segments[index].from
			cell := canvas.cells[corner.y][corner.x]
			if cell.crossing || cell.conflict || cell.overlap {
				t.Errorf("route %s -> %s lost its corner at %+v: %+v",
					route.edge.From, route.edge.To, corner, cell)
			}
		}
	}
}

func TestThreadSpatialCanvasDistinguishesCrossingsSharedStubsAndOverlaps(t *testing.T) {
	crossing := newThreadSpatialCanvas(8, 8)
	crossing.routeHorizontal(1, 5, 3, threadSpatialRouteStyle{id: 1, from: "a", to: "d"})
	crossing.routeVertical(3, 1, 5, threadSpatialRouteStyle{id: 2, from: "b", to: "d"})
	if cell := crossing.cells[3][3]; !cell.crossing || cell.shared || cell.connector != 0 {
		t.Fatalf("perpendicular routes with a common endpoint became a junction: %+v", cell)
	}

	shared := newThreadSpatialCanvas(8, 3)
	shared.routeHorizontal(1, 6, 1, threadSpatialRouteStyle{id: 1, from: "a", to: "c"})
	shared.routeHorizontal(1, 5, 1, threadSpatialRouteStyle{id: 2, from: "a", to: "d"})
	if cell := shared.cells[1][3]; !cell.shared || threadSpatialSharedConnectorGlyph(cell.connector) != "═" {
		t.Fatalf("common-source stub lost the shared-route grammar: %+v", cell)
	}
	merged := newThreadSpatialCanvas(9, 8)
	turnStyle := threadSpatialRouteStyle{id: 1, from: "a", to: "d"}
	merged.routeVertical(3, 3, 6, turnStyle)
	merged.routeHorizontal(3, 7, 3, turnStyle)
	merged.routeHorizontal(1, 7, 3, threadSpatialRouteStyle{id: 2, from: "b", to: "d"})
	if cell := merged.cells[3][3]; !cell.shared || cell.crossing || cell.connector !=
		(threadSpatialLeft|threadSpatialRight|threadSpatialDown) {
		t.Fatalf("common-target turn did not preserve every real bundle arm: %+v", cell)
	}

	overlap := newThreadSpatialCanvas(8, 3)
	overlap.routeHorizontal(1, 6, 1, threadSpatialRouteStyle{id: 1, from: "a", to: "b"})
	overlap.routeHorizontal(3, 7, 1, threadSpatialRouteStyle{id: 2, from: "c", to: "d"})
	if cell := overlap.cells[1][4]; !cell.overlap || cell.shared || cell.connector != 0 {
		t.Fatalf("unrelated collinear routes were presented as a shared bundle: %+v", cell)
	}
}

func TestThreadSpatialRouteGrammarKeepsFocusSharedRoutesAndCollisionsDistinct(t *testing.T) {
	ordinary := newThreadSpatialCanvas(8, 3)
	ordinary.routeHorizontal(1, 6, 1, threadSpatialRouteStyle{id: 1, from: "a", to: "b"})
	if got := threadSpatialConnectorGlyph(ordinary.cells[1][3].connector); got != "─" {
		t.Fatalf("ordinary route glyph=%q want light stroke", got)
	}

	focused := newThreadSpatialCanvas(8, 3)
	focused.routeHorizontal(1, 6, 1, threadSpatialRouteStyle{id: 1, from: "a", to: "b", selected: true})
	if cell := focused.cells[1][3]; !cell.accent || threadSpatialFocusConnectorGlyph(cell.connector) != "━" {
		t.Fatalf("focused route lacks redundant accent and heavy-stroke grammar: %+v", cell)
	}
	if got := ansi.Strip(focused.renderLine(1, &testStyles)); !strings.Contains(got, "━") {
		t.Fatalf("focused route did not retain its non-color channel after rendering: %q", got)
	}
	focused.routeVertical(6, 1, 2, threadSpatialRouteStyle{id: 1, from: "a", to: "b", selected: true})
	if got := threadSpatialFocusConnectorGlyph(focused.cells[1][6].connector); got != "┓" {
		t.Fatalf("focused route corner=%q want one continuous heavy elbow", got)
	}

	shared := newThreadSpatialCanvas(8, 3)
	shared.routeHorizontal(1, 6, 1, threadSpatialRouteStyle{id: 1, from: "a", to: "b", selected: true})
	shared.routeHorizontal(1, 5, 1, threadSpatialRouteStyle{id: 2, from: "a", to: "c"})
	if cell := shared.cells[1][3]; !cell.shared || !cell.accent || threadSpatialSharedConnectorGlyph(cell.connector) != "═" {
		t.Fatalf("focused shared bundle did not preserve both meanings: %+v", cell)
	}

	crossing := newThreadSpatialCanvas(8, 8)
	crossing.routeHorizontal(1, 5, 3, threadSpatialRouteStyle{id: 1, from: "a", to: "b", selected: true})
	crossing.routeVertical(3, 1, 5, threadSpatialRouteStyle{id: 2, from: "c", to: "d"})
	if cell := crossing.cells[3][3]; !cell.crossing || cell.accent || cell.color != theme.ColorGray {
		t.Fatalf("unrelated crossing inherited focused-route treatment: %+v", cell)
	}
	if got := ansi.Strip(crossing.renderLine(3, &testStyles)); !strings.Contains(got, "╳") {
		t.Fatalf("neutral crossing grammar missing from rendered row: %q", got)
	}
}

func TestThreadSpatialFanCountRelocatesInsteadOfErasingCongestion(t *testing.T) {
	canvas := newThreadSpatialCanvas(10, 6)
	canvas.routeHorizontal(1, 8, 3, threadSpatialRouteStyle{id: 1, from: "a", to: "b"})
	canvas.routeVertical(3, 1, 5, threadSpatialRouteStyle{id: 2, from: "c", to: "d"})
	if !canvas.cells[3][3].crossing {
		t.Fatal("fixture did not put a crossing on the preferred count cell")
	}
	point, ok := canvas.putRouteCountAlong(threadSpatialPoint{x: 3, y: 3}, 1, 3, 5, false)
	if !ok || point != (threadSpatialPoint{x: 4, y: 3}) {
		t.Fatalf("fan count placement=(%+v,%v) want first free approach cell", point, ok)
	}
	if !canvas.cells[3][3].crossing || canvas.cells[3][3].text != "" {
		t.Fatalf("relocated count erased crossing grammar: %+v", canvas.cells[3][3])
	}
	if cell := canvas.cells[3][4]; !cell.routeCount || cell.text != "5" || cell.color != theme.ColorYellow {
		t.Fatalf("relocated fan count lost multiplicity grammar: %+v", cell)
	}
}

func TestThreadSpatialFanCountFallsBackToNodeBorderWhenStubIsFull(t *testing.T) {
	placement := threadSpatialNode{
		node: core.ThreadGraphNode{TaskID: "a", Label: "source", Role: core.ThreadTaskMember},
		x:    3, y: 2,
	}
	layout := threadSpatialLayout{byID: map[string]threadSpatialNode{"a": placement}}
	canvas := newThreadSpatialCanvas(50, 10)
	drawThreadSpatialNode(canvas, placement, false)
	canvas.cells[4][27] = threadSpatialCell{crossing: true, color: theme.ColorGray}
	routes := []threadSpatialRoute{
		{id: 1, edge: core.ThreadGraphEdge{From: "a", To: "b"}, segments: []threadSpatialRouteSegment{{
			from: threadSpatialPoint{x: 25, y: 4}, to: threadSpatialPoint{x: 28, y: 4},
		}}, arrow: threadSpatialPoint{x: 40, y: 4}, arrowRune: '▶'},
		{id: 2, edge: core.ThreadGraphEdge{From: "a", To: "c"}, segments: []threadSpatialRouteSegment{{
			from: threadSpatialPoint{x: 25, y: 4}, to: threadSpatialPoint{x: 28, y: 4},
		}}, arrow: threadSpatialPoint{x: 44, y: 4}, arrowRune: '▶'},
	}
	drawThreadSpatialRouteCounts(canvas, layout, routes, "")
	fallback := threadSpatialCountFallback(layout, "a", "", 's', '▶')
	if cell := canvas.cells[fallback.y][fallback.x]; !cell.routeCount || cell.text != "2" || cell.color != theme.ColorYellow {
		t.Fatalf("fully congested fan-out silently lost its node-border fallback: %+v", cell)
	}
	if !canvas.cells[4][27].crossing {
		t.Fatalf("node-border fallback erased congested stub evidence: %+v", canvas.cells[4][27])
	}
}

func TestThreadSpatialRoutingConflictIsNeutralAndReportedOutOfBand(t *testing.T) {
	canvas := newThreadSpatialCanvas(9, 8)
	turn := threadSpatialRouteStyle{id: 1, from: "a", to: "b", selected: true}
	canvas.routeVertical(3, 3, 6, turn)
	canvas.routeHorizontal(3, 7, 3, turn)
	canvas.routeHorizontal(1, 7, 3, threadSpatialRouteStyle{id: 2, from: "c", to: "d"})
	cell := canvas.cells[3][3]
	if !cell.conflict || cell.accent || cell.color != theme.ColorGray {
		t.Fatalf("renderer conflict leaked into focus or task-health presentation: %+v", cell)
	}
	if got := threadSpatialRouteConflictSummary(canvas); got != "1 routing conflict" {
		t.Fatalf("routing conflict summary=%q want explicit out-of-band diagnostic", got)
	}
}

func TestThreadSpatialInlineLegendStaysCompactAndDefersRareGrammarToHelp(t *testing.T) {
	legend := ansi.Strip(threadSpatialRoleLegend(&testStyles))
	for _, grammar := range []string{"┌ member", "╔ gate", "━ focus route", "▶ direction", "2 fan"} {
		if !strings.Contains(legend, grammar) {
			t.Errorf("route legend omitted %q: %q", grammar, legend)
		}
	}
	if width := ansi.StringWidth(legend); width > 72 {
		t.Errorf("inline route legend width=%d want <=72: %q", width, legend)
	}
	for _, rare := range []string{"╳", "≋", "conflict"} {
		if strings.Contains(legend, rare) {
			t.Errorf("inline legend should leave uncommon %q grammar to contextual help: %q", rare, legend)
		}
	}
}

func TestThreadSpatialRouteCountsPreserveSideAndEndpointSemantics(t *testing.T) {
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: "a", Role: core.ThreadTaskMember},
			{TaskID: "b", Role: core.ThreadTaskMember},
			{TaskID: "c", Role: core.ThreadTaskMember},
			{TaskID: "d", Role: core.ThreadTaskMember},
		},
		Edges: []core.ThreadGraphEdge{
			{From: "a", To: "c"}, {From: "b", To: "c"},
			{From: "c", To: "c"}, {From: "c", To: "d"}, {From: "d", To: "c"},
		},
	}
	layout := buildThreadSpatialLayout(projection)
	placement := layout.byID["c"]
	row := placement.y + threadSpatialNodeSlotHeight/2
	canvas := renderThreadSpatialCanvas(projection, layout, "c")
	for _, check := range []struct {
		x    int
		text string
		kind string
	}{
		{x: placement.x - 2, text: "2", kind: "left-entry fan-in count"},
		{x: placement.x - 1, text: "▶", kind: "left-entry direction"},
		{x: placement.x + threadSpatialNodeWidth, text: "◀", kind: "right-entry direction"},
		{x: placement.x + threadSpatialNodeWidth + 1, text: "2", kind: "right-entry fan-in count"},
		{x: placement.x + threadSpatialNodeWidth + 2, text: "2", kind: "source fan-out count"},
	} {
		if got := canvas.cells[row][check.x].text; got != check.text {
			t.Errorf("%s at x=%d is %q want %q", check.kind, check.x, got, check.text)
		}
	}
	for _, x := range []int{placement.x - 2, placement.x + threadSpatialNodeWidth + 1, placement.x + threadSpatialNodeWidth + 2} {
		if !canvas.cells[row][x].routeCount {
			t.Errorf("cell x=%d contains an incidental numeral rather than a route count: %+v", x, canvas.cells[row][x])
		}
	}
}

func TestThreadSpatialDenseRoutesDistinguishCrossingsBundlesAndCounts(t *testing.T) {
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: "a", Role: core.ThreadTaskMember},
			{TaskID: "b", Role: core.ThreadTaskMember},
			{TaskID: "c", Role: core.ThreadTaskMember},
			{TaskID: "d", Role: core.ThreadTaskMember},
			{TaskID: "e", Role: core.ThreadTaskMember},
			{TaskID: "f", Role: core.ThreadTaskMember},
		},
		Edges: []core.ThreadGraphEdge{
			{From: "a", To: "c"}, {From: "b", To: "d"},
			{From: "c", To: "f"}, {From: "d", To: "e"},
			{From: "b", To: "e"},
		},
	}
	layout := buildThreadSpatialLayout(projection)
	canvas := renderThreadSpatialCanvas(projection, layout, "b")
	crossings, shared, routeCounts, trackedRoutes := 0, 0, 0, 0
	for _, route := range layout.routes {
		for _, segment := range route.segments {
			if segment.corridor {
				trackedRoutes++
			}
		}
	}
	for row := range canvas.cells {
		for column := range canvas.cells[row] {
			cell := canvas.cells[row][column]
			if cell.crossing {
				crossings++
				if cell.connector != 0 {
					t.Errorf("crossing at (%d,%d) became a junction: %+v", column, row, cell)
				}
				if cell.accent || cell.color != theme.ColorGray {
					t.Errorf("crossing at (%d,%d) inherited focus or status color: %+v", column, row, cell)
				}
				if got := ansi.Strip(canvas.renderLine(row, &testStyles)); !strings.Contains(got, "╳") {
					t.Errorf("crossing row omitted the non-junction glyph: %q", got)
				}
			}
			if cell.shared {
				shared++
			}
			if cell.routeCount {
				routeCounts++
			}
		}
	}
	if crossings == 0 {
		t.Fatal("fixture did not expose an unrelated perpendicular route crossing")
	}
	if shared == 0 {
		t.Fatal("fan-in/fan-out did not expose explicit shared-route geometry")
	}
	if trackedRoutes == 0 {
		t.Fatal("fixture did not exercise a skipped-layer route")
	}
	if routeCounts == 0 {
		t.Fatalf("dense fan-in/fan-out omitted every endpoint multiplicity marker: count markers=%d", routeCounts)
	}
	for _, route := range layout.routes {
		if route.edge.From != "b" && route.edge.To != "b" {
			continue
		}
		for _, segment := range route.segments {
			for _, point := range threadSpatialTestSegmentPoints(segment) {
				cell := canvas.cells[point.y][point.x]
				collision := cell.crossing || cell.overlap || cell.conflict
				if !cell.accent && !collision && !strings.ContainsAny(cell.text, "▶◀2+") {
					t.Errorf("selected incident route %s -> %s lost emphasis at %+v: %+v",
						route.edge.From, route.edge.To, point, cell)
				}
			}
		}
	}
	for row := range canvas.cells {
		plain := ansi.Strip(canvas.renderLine(row, &testStyles))
		if strings.ContainsAny(plain, "┄◇") {
			t.Errorf("route row retained fragmented track/waypoint grammar: %q", plain)
		}
	}
}

func TestThreadSpatialDenseLayeredFixtureIsBoundedAndDeterministic(t *testing.T) {
	const layerCount, rowCount = 5, 4
	projection := core.ThreadGraphProjection{
		View:             core.ThreadView{GraphHealth: core.GraphHealthy, ProjectionHealth: core.GraphHealthy},
		TopologyComplete: true,
	}
	ids := make([][]string, layerCount)
	for layer := range layerCount {
		wave := core.ThreadGraphWave{Index: layer + 1}
		for row := range rowCount {
			taskID := fmt.Sprintf("l%d-r%d", layer, row)
			ids[layer] = append(ids[layer], taskID)
			wave.TaskIDs = append(wave.TaskIDs, taskID)
			projection.Nodes = append(projection.Nodes, core.ThreadGraphNode{
				TaskID: taskID, Label: taskID, Status: domain.StatusNextUp, Role: core.ThreadTaskMember,
				State: core.TaskGraphState{Role: core.RoleQueued, Gate: core.GateClear},
			})
		}
		projection.Waves = append(projection.Waves, wave)
	}
	for layer := 0; layer < layerCount-1; layer++ {
		for _, from := range ids[layer] {
			for _, to := range ids[layer+1] {
				projection.Edges = append(projection.Edges, core.ThreadGraphEdge{From: from, To: to})
			}
		}
	}
	for layer := 0; layer < layerCount-2; layer++ {
		for row := range rowCount {
			projection.Edges = append(projection.Edges, core.ThreadGraphEdge{
				From: ids[layer][row], To: ids[layer+2][row],
			})
		}
	}

	layout := buildThreadSpatialLayout(projection)
	if issue := threadSpatialCapacityIssue(layout, len(projection.Edges)); issue != "" {
		t.Fatalf("representative dense fixture exceeded its bounded canvas: %s", issue)
	}
	if len(layout.routes) != len(projection.Edges) {
		t.Fatalf("routed edges=%d want every supplied edge=%d", len(layout.routes), len(projection.Edges))
	}
	seenRoutes := make(map[int32]bool, len(layout.routes))
	for _, route := range layout.routes {
		if seenRoutes[route.id] {
			t.Fatalf("route identity repeated: %d", route.id)
		}
		seenRoutes[route.id] = true
		for _, segment := range route.segments {
			for _, point := range threadSpatialTestSegmentPoints(segment) {
				for _, node := range layout.nodes {
					inside := point.x >= node.x && point.x < node.x+threadSpatialNodeWidth &&
						point.y >= node.y && point.y < node.y+threadSpatialNodeSlotHeight
					if inside {
						t.Fatalf("dense route %s -> %s crossed node %s at %+v",
							route.edge.From, route.edge.To, node.node.TaskID, point)
					}
				}
			}
		}
	}
	selected := ids[2][1]
	canvas := renderThreadSpatialCanvas(projection, layout, selected)
	for row := range canvas.cells {
		for column := range canvas.cells[row] {
			cell := canvas.cells[row][column]
			if cell.conflict || cell.overlap {
				point := threadSpatialPoint{x: column, y: row}
				touches := make([]string, 0)
				for _, route := range layout.routes {
					for _, segment := range route.segments {
						for _, candidate := range threadSpatialTestSegmentPoints(segment) {
							if candidate == point {
								touches = append(touches, fmt.Sprintf("%d:%s->%s:%+v", route.id, route.edge.From, route.edge.To, segment))
								break
							}
						}
					}
				}
				t.Fatalf("dense production route collision at (%d,%d) is ambiguous: %+v routes=%v", column, row, cell, touches)
			}
		}
	}
	first := renderThreadSpatial(projection, "", selected, 100, 28, &testStyles)
	for iteration := 0; iteration < 10; iteration++ {
		if got := renderThreadSpatial(projection, "", selected, 100, 28, &testStyles); got != first {
			t.Fatalf("dense render changed on iteration %d", iteration)
		}
	}
}

func TestThreadSpatialRoutesStayOutsideNodeBoxes(t *testing.T) {
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: "a", Role: core.ThreadTaskMember},
			{TaskID: "b", Role: core.ThreadTaskMember},
			{TaskID: "c", Role: core.ThreadTaskMember},
			{TaskID: "d", Role: core.ThreadTaskMember},
		},
		Edges: []core.ThreadGraphEdge{
			{From: "a", To: "c"}, {From: "b", To: "d"}, {From: "a", To: "d"},
		},
	}
	layout := buildThreadSpatialLayout(projection)
	for _, route := range layout.routes {
		for _, segment := range route.segments {
			for _, point := range threadSpatialTestSegmentPoints(segment) {
				for _, node := range layout.nodes {
					inside := point.x >= node.x && point.x < node.x+threadSpatialNodeWidth &&
						point.y >= node.y && point.y < node.y+threadSpatialNodeSlotHeight
					if inside {
						t.Fatalf("route %s -> %s crossed node %s at %+v", route.edge.From, route.edge.To, node.node.TaskID, point)
					}
				}
			}
		}
	}
}

func TestThreadSpatialClippedIncidentRoutesNameTheirOffscreenEndpoint(t *testing.T) {
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: "a", Role: core.ThreadTaskMember},
			{TaskID: "b", Role: core.ThreadTaskMember},
			{TaskID: "c", Role: core.ThreadTaskMember},
			{TaskID: "d", Role: core.ThreadTaskMember},
		},
		Edges: []core.ThreadGraphEdge{
			{From: "a", To: "b"}, {From: "b", To: "c"}, {From: "c", To: "d"}, {From: "a", To: "d"},
		},
	}
	for _, test := range []struct {
		selected string
		want     string
	}{
		{selected: "a", want: "…▶[M4]"},
		{selected: "d", want: "[M1]…"},
	} {
		rendered := ansi.Strip(renderThreadSpatial(projection, "", test.selected, 60, 20, &testStyles))
		graph := strings.Join(strings.Split(rendered, "\n")[:15], "\n")
		if !strings.Contains(graph, test.want) {
			t.Errorf("selected %s omitted clipped route endpoint %q:\n%s", test.selected, test.want, graph)
		}
	}
}

func TestThreadSpatialBoundaryLabelsBundleAliasesWithoutOverwritingNodes(t *testing.T) {
	aliases := make([]string, 20)
	for index := range aliases {
		aliases[index] = fmt.Sprintf("M%d", index+1)
	}
	if got, want := threadSpatialBoundaryLabel('t', '▶', aliases), "…▶[M1,M2,M3,+17]"; got != want {
		t.Fatalf("large clipped fan-in label=%q want compact diagnostic %q", got, want)
	}

	layout := threadSpatialLayout{
		byID: map[string]threadSpatialNode{
			"a": {node: core.ThreadGraphNode{TaskID: "a"}, alias: "M1", x: 3, y: 2},
			"b": {node: core.ThreadGraphNode{TaskID: "b"}, alias: "M2", x: 100, y: 2},
			"c": {node: core.ThreadGraphNode{TaskID: "c"}, alias: "M3", x: 100, y: 8},
			"d": {node: core.ThreadGraphNode{TaskID: "d"}, alias: "M4", x: 100, y: 14},
		},
	}
	for index, target := range []string{"b", "c", "d"} {
		layout.routes = append(layout.routes, threadSpatialRoute{
			id: int32(index + 1), edge: core.ThreadGraphEdge{From: "a", To: target},
			segments: []threadSpatialRouteSegment{{
				from: threadSpatialPoint{x: 25, y: 4}, to: threadSpatialPoint{x: 100, y: 4},
			}},
			arrow: threadSpatialPoint{x: 100, y: 4}, arrowRune: '▶',
		})
	}
	canvas := newThreadSpatialCanvas(110, 20)
	drawThreadSpatialNode(canvas, layout.byID["a"], true)
	before := append([]threadSpatialCell(nil), canvas.cells[4][3:25]...)
	annotateThreadSpatialRouteBoundaries(canvas, layout, "a", 0, 0, 40, 10)
	if got := ansi.Strip(canvas.renderLine(4, &testStyles)); !strings.Contains(got, "…▶[M2,M3,M4]") {
		t.Fatalf("clipped fan-out aliases were not bundled deterministically: %q", got)
	}
	for column := 25; column < 40; column++ {
		if cell := canvas.cells[4][column]; cell.text != "" && !cell.accent {
			t.Fatalf("selected-route boundary alias did not use the focus channel at column %d: %+v", column, cell)
		}
	}
	if got := canvas.cells[4][3:25]; !reflect.DeepEqual(got, before) {
		t.Fatalf("boundary annotation overwrote the selected node: got=%+v want=%+v", got, before)
	}

	// A centered top/bottom annotation beside a node must move into free space
	// instead of replacing the node's border at column 24.
	putThreadSpatialBoundaryLabel(canvas, 't', 26, 4, "[M9]…", 0, 40)
	if got := canvas.cells[4][24]; !reflect.DeepEqual(got, before[21]) {
		t.Fatalf("collision-aware boundary label replaced node border: got=%+v want=%+v", got, before[21])
	}
}

func TestThreadSpatialViewportOmitsRoutesBetweenTwoOffscreenNodes(t *testing.T) {
	layout := threadSpatialLayout{
		byID: map[string]threadSpatialNode{
			"left":  {node: core.ThreadGraphNode{TaskID: "left"}, x: 3, y: 2},
			"focus": {node: core.ThreadGraphNode{TaskID: "focus"}, x: 63, y: 2},
			"right": {node: core.ThreadGraphNode{TaskID: "right"}, x: 123, y: 2},
		},
		routes: []threadSpatialRoute{
			{id: 1, edge: core.ThreadGraphEdge{From: "left", To: "right"}},
			{id: 2, edge: core.ThreadGraphEdge{From: "focus", To: "right"}},
		},
	}
	routes := threadSpatialRoutesForWindow(layout, "focus", 50, 0, 40, 12)
	if len(routes) != 1 || routes[0].id != 2 {
		t.Fatalf("viewport routes=%+v want only the selected incident route", routes)
	}
}

func threadSpatialTestSegmentPoints(segment threadSpatialRouteSegment) []threadSpatialPoint {
	points := make([]threadSpatialPoint, 0)
	switch {
	case segment.from.y == segment.to.y:
		left, right := min(segment.from.x, segment.to.x), max(segment.from.x, segment.to.x)
		for x := left; x <= right; x++ {
			points = append(points, threadSpatialPoint{x: x, y: segment.from.y})
		}
	case segment.from.x == segment.to.x:
		top, bottom := min(segment.from.y, segment.to.y), max(segment.from.y, segment.to.y)
		for y := top; y <= bottom; y++ {
			points = append(points, threadSpatialPoint{x: segment.from.x, y: y})
		}
	}
	return points
}

func TestThreadSpatialCyclicResidueUsesDistinctDeterministicLoops(t *testing.T) {
	base := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: "a", Role: core.ThreadTaskMember},
			{TaskID: "b", Role: core.ThreadTaskMember},
			{TaskID: "c", Role: core.ThreadTaskMember},
		},
		Edges: []core.ThreadGraphEdge{
			{From: "a", To: "a"}, {From: "a", To: "b"}, {From: "b", To: "c"}, {From: "c", To: "a"},
		},
	}
	permuted := base
	permuted.Nodes = []core.ThreadGraphNode{base.Nodes[2], base.Nodes[1], base.Nodes[0]}
	permuted.Edges = []core.ThreadGraphEdge{base.Edges[3], base.Edges[2], base.Edges[1], base.Edges[0]}

	layout := buildThreadSpatialLayout(base)
	if len(layout.columns) != 1 || !strings.Contains(layout.columnLabels[0], "unranked") {
		t.Fatalf("cyclic residue was not kept in one explicit partial layer: columns=%v labels=%v", layout.columns, layout.columnLabels)
	}
	if len(layout.routes) != 4 {
		t.Fatalf("cyclic residue routes=%d want 4 including the supplied self-edge", len(layout.routes))
	}
	lanes := make(map[int]bool)
	for _, route := range layout.routes {
		if route.arrowRune != '◀' {
			t.Errorf("same-column cyclic edge %s -> %s arrow=%q want reverse-entry marker", route.edge.From, route.edge.To, route.arrowRune)
		}
		lane := 0
		for _, segment := range route.segments {
			if segment.from.x == segment.to.x {
				lane = max(lane, segment.from.x)
			}
		}
		lanes[lane] = true
	}
	if len(lanes) != len(layout.routes) {
		t.Fatalf("cyclic edges shared indistinguishable loop lanes: lanes=%v routes=%d", lanes, len(layout.routes))
	}
	if got := buildThreadSpatialLayout(permuted); !reflect.DeepEqual(got, layout) {
		t.Fatalf("equivalent cyclic evidence changed route layout:\n got: %#v\nwant: %#v", got, layout)
	}
}

func TestThreadSpatialReverseRouteUsesNodeFreeTrack(t *testing.T) {
	const routeID int32 = 1
	placements := map[string]threadSpatialNode{
		"target": {node: core.ThreadGraphNode{TaskID: "target"}, column: 0, row: 0, x: 3, y: 2},
		"source": {node: core.ThreadGraphNode{TaskID: "source"}, column: 2, row: 0, x: 63, y: 2},
	}
	seed := threadSpatialRouteSeed{
		id: routeID, edge: core.ThreadGraphEdge{From: "source", To: "target"},
		fromColumn: 2, toColumn: 0, fromRow: 0, toRow: 0, needsTrack: true,
	}
	routes := materializeThreadSpatialRoutes(
		[]threadSpatialRouteSeed{seed}, placements,
		map[int32]threadSpatialRouteLanes{routeID: {source: 88, target: 27}},
		map[int32]int{routeID: 7},
	)
	if len(routes) != 1 || routes[0].arrowRune != '◀' || routes[0].arrow != (threadSpatialPoint{x: 25, y: 4}) {
		t.Fatalf("reverse route did not enter its actual target from the right: %+v", routes)
	}
	corridors := 0
	for _, segment := range routes[0].segments {
		if segment.corridor {
			corridors++
		}
		for _, point := range threadSpatialTestSegmentPoints(segment) {
			for taskID, node := range placements {
				inside := point.x >= node.x && point.x < node.x+threadSpatialNodeWidth &&
					point.y >= node.y && point.y < node.y+threadSpatialNodeSlotHeight
				if inside {
					t.Fatalf("reverse route crossed %s at %+v", taskID, point)
				}
			}
		}
	}
	if corridors != 1 {
		t.Fatalf("reverse route lost its attributed horizontal track: %+v", routes[0])
	}
}

func TestThreadSpatialLongEdgeUsesNodeFreeTrack(t *testing.T) {
	a := testutil.TaskID("long-edge-source")
	d := testutil.TaskID("other-source")
	b := testutil.TaskID("intermediate-node")
	c := testutil.TaskID("long-edge-target")
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: a, Label: "long source", Role: core.ThreadTaskMember},
			{TaskID: d, Label: "other source", Role: core.ThreadTaskMember},
			{TaskID: b, Label: "intermediate", Role: core.ThreadTaskMember},
			{TaskID: c, Label: "long target", Role: core.ThreadTaskMember},
		},
		Edges: []core.ThreadGraphEdge{
			{From: a, To: c},
			{From: d, To: b},
			{From: b, To: c},
		},
	}
	layout := buildThreadSpatialLayout(projection)
	canvas := renderThreadSpatialCanvas(projection, layout, a)
	from, middle, to := layout.byID[a], layout.byID[b], layout.byID[c]
	if to.column-from.column <= 1 || middle.column != from.column+1 {
		t.Fatalf("fixture did not create a skipped visible column: from=%d middle=%d to=%d", from.column, middle.column, to.column)
	}
	route, ok := threadSpatialTestRoute(layout, a, c)
	if !ok {
		t.Fatalf("layout omitted route %s -> %s", a, c)
	}
	corridor, ok := threadSpatialTestCorridor(route)
	if !ok {
		t.Fatalf("skipped-layer route has no corridor: %+v", route)
	}
	probeX := (corridor.from.x + corridor.to.x) / 2
	if !canvas.cells[corridor.from.y][probeX].accent {
		t.Fatal("selected long edge did not use the node-free inter-row track")
	}
	middleY := middle.y + threadSpatialNodeSlotHeight/2
	if canvas.cells[middleY][middle.x-2].accent {
		t.Fatal("selected long edge still borrowed the intermediate node's incoming route")
	}
}

func TestThreadSpatialLongEdgeKeepsDeepestRowTrackInsideCanvas(t *testing.T) {
	a, b, c, d := "a", "b", "c", "d"
	projection := core.ThreadGraphProjection{
		Nodes: []core.ThreadGraphNode{
			{TaskID: a, Label: "upper source", Role: core.ThreadTaskMember},
			{TaskID: b, Label: "deep source", Role: core.ThreadTaskMember},
			{TaskID: c, Label: "middle", Role: core.ThreadTaskMember},
			{TaskID: d, Label: "target", Role: core.ThreadTaskMember},
		},
		Edges: []core.ThreadGraphEdge{
			{From: a, To: c},
			{From: c, To: d},
			{From: b, To: d},
		},
	}
	layout := buildThreadSpatialLayout(projection)
	from, to := layout.byID[b], layout.byID[d]
	if from.row == 0 || to.column-from.column <= 1 {
		t.Fatalf("fixture did not put a skipped-layer endpoint on the deepest row: from=%+v to=%+v", from, to)
	}
	route, ok := threadSpatialTestRoute(layout, b, d)
	if !ok {
		t.Fatalf("layout omitted route %s -> %s", b, d)
	}
	corridor, ok := threadSpatialTestCorridor(route)
	if !ok {
		t.Fatalf("deepest-row skipped-layer route has no corridor: %+v", route)
	}
	trackY := corridor.from.y
	if trackY >= layout.height {
		t.Fatalf("long-edge track row %d falls outside layout height %d", trackY, layout.height)
	}
	canvas := renderThreadSpatialCanvas(projection, layout, b)
	probeX := (corridor.from.x + corridor.to.x) / 2
	if got := threadSpatialConnectorGlyph(canvas.cells[trackY][probeX].connector); got != "─" {
		t.Fatalf("deepest-row long edge was clipped at its horizontal track: glyph=%q", got)
	}
	if !canvas.cells[trackY][probeX].accent {
		t.Fatal("deepest-row long edge lost selected-route emphasis")
	}
}

func threadSpatialTestRoute(layout threadSpatialLayout, from, to string) (threadSpatialRoute, bool) {
	for _, route := range layout.routes {
		if route.edge.From == from && route.edge.To == to {
			return route, true
		}
	}
	return threadSpatialRoute{}, false
}

func threadSpatialTestCorridor(route threadSpatialRoute) (threadSpatialRouteSegment, bool) {
	for _, segment := range route.segments {
		if segment.corridor {
			return segment, true
		}
	}
	return threadSpatialRouteSegment{}, false
}

func TestThreadSpatialPresentationIsInvariantUnderEquivalentProjectionPermutations(t *testing.T) {
	base := core.ThreadGraphProjection{
		View: core.ThreadView{GraphHealth: core.GraphHealthy, ProjectionHealth: core.GraphHealthy},
		Nodes: []core.ThreadGraphNode{
			{TaskID: "a", Label: "alpha", Status: domain.StatusCompleted, Role: core.ThreadTaskMember},
			{TaskID: "b", Label: "beta", Status: domain.StatusNextUp, Role: core.ThreadTaskMember},
			{TaskID: "c", Label: "charlie", Status: domain.StatusReadyToStart, Role: core.ThreadTaskMember},
			{TaskID: "gate", Label: "external", Status: domain.StatusCompleted, Role: core.ThreadTaskExternalGate},
		},
		Edges: []core.ThreadGraphEdge{
			{From: "a", To: "c"}, {From: "b", To: "c"}, {From: "gate", To: "b"},
		},
		Waves: []core.ThreadGraphWave{
			{Index: 1, TaskIDs: []string{"a", "b"}}, {Index: 2, TaskIDs: []string{"c"}},
		},
	}
	permuted := base
	permuted.Nodes = []core.ThreadGraphNode{base.Nodes[3], base.Nodes[2], base.Nodes[1], base.Nodes[0]}
	permuted.Edges = []core.ThreadGraphEdge{base.Edges[2], base.Edges[1], base.Edges[0]}
	permuted.Waves = []core.ThreadGraphWave{
		{Index: 2, TaskIDs: []string{"c"}}, {Index: 1, TaskIDs: []string{"b", "a"}},
	}

	if got, want := buildThreadSpatialLayout(permuted), buildThreadSpatialLayout(base); !reflect.DeepEqual(got, want) {
		t.Fatalf("equivalent projection permutation changed spatial layout:\n got: %#v\nwant: %#v", got, want)
	}
	if got, want := threadGraphAliases(permuted), threadGraphAliases(base); !reflect.DeepEqual(got, want) {
		t.Fatalf("equivalent projection permutation changed aliases: got %v want %v", got, want)
	}
	baseLayout := buildThreadSpatialLayout(base)
	gotNeeds, gotUnlocks := threadSpatialConnections(permuted, baseLayout, "c")
	wantNeeds, wantUnlocks := threadSpatialConnections(base, baseLayout, "c")
	if gotNeeds != wantNeeds || gotUnlocks != wantUnlocks {
		t.Fatalf("equivalent edge permutation changed inspector connections: got (%q, %q) want (%q, %q)",
			gotNeeds, gotUnlocks, wantNeeds, wantUnlocks)
	}
	if got, want := renderThreadSpatial(permuted, "", "c", 120, 30, &testStyles),
		renderThreadSpatial(base, "", "c", 120, 30, &testStyles); got != want {
		t.Fatalf("equivalent projection permutation changed rendered output:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestThreadSpatialSelectedNodeExpandsInsideStableSlot(t *testing.T) {
	projection := core.ThreadGraphProjection{Nodes: []core.ThreadGraphNode{{
		TaskID: "task-id", Label: "selected task", Status: domain.StatusNextUp,
		Role: core.ThreadTaskMember, State: core.TaskGraphState{Role: core.RoleQueued, Gate: core.GateClear},
	}}}
	layout := buildThreadSpatialLayout(projection)
	placement := layout.byID["task-id"]
	canvas := renderThreadSpatialCanvas(projection, layout, "task-id")
	for row := placement.y; row < placement.y+threadSpatialNodeSlotHeight; row++ {
		cell := canvas.cells[row][placement.x]
		if cell.accent || cell.color != theme.Status(domain.StatusNextUp).Color {
			t.Fatalf("expanded selected card row %d changed semantic box color: %+v", row-placement.y, cell)
		}
	}
	if got := canvas.cells[placement.y+1][placement.x+2].color; got != theme.Status(domain.StatusNextUp).Color {
		t.Fatalf("selected status glyph color=%v want semantic next-up color", got)
	}
	if got := canvas.cells[placement.y+2][placement.x-3].text; got != "" {
		t.Fatalf("selected card retained a redundant connector-adjacent focus marker: %q", got)
	}
}

func TestThreadSpatialGraphIsBoundedDeterministicAndExplicitWhenNarrow(t *testing.T) {
	projection := hostileThreadGraphProjection()
	selected := ""
	for _, node := range projection.Nodes {
		if node.State.Role == core.RoleUnknown {
			selected = node.TaskID
			break
		}
	}
	if selected == "" {
		t.Fatal("hostile projection has no unreadable node for the partial-topology viewport")
	}
	for _, size := range []struct{ width, height int }{{120, 28}, {72, 14}, {54, 10}} {
		first := renderThreadSpatial(projection, "remote path unavailable", selected, size.width, size.height, &testStyles)
		second := renderThreadSpatial(projection, "remote path unavailable", selected, size.width, size.height, &testStyles)
		if first != second {
			t.Fatalf("%dx%d spatial render was nondeterministic", size.width, size.height)
		}
		plain := ansi.Strip(first)
		if lines := strings.Count(plain, "\n") + 1; lines > size.height {
			t.Errorf("%dx%d rendered %d lines", size.width, size.height, lines)
		}
		for _, line := range strings.Split(plain, "\n") {
			if got := ansi.StringWidth(line); got > size.width {
				t.Errorf("%dx%d line width=%d: %q", size.width, size.height, got, line)
			}
		}
		if size.width < threadSpatialMinWidth || size.height < threadSpatialMinHeight {
			if !strings.Contains(plain, "needs at least") || !strings.Contains(plain, "Esc returns") {
				t.Errorf("narrow fallback was not explanatory:\n%s", plain)
			}
		} else {
			for _, want := range []string{"spatial graph", "partial", "unranked", "focus", "about"} {
				if !strings.Contains(plain, want) {
					t.Errorf("%dx%d graph omitted %q:\n%s", size.width, size.height, want, plain)
				}
			}
		}
	}
}

func TestThreadSpatialGraphFailsOpenToWavesBeyondPrototypeCapacity(t *testing.T) {
	projection := core.ThreadGraphProjection{Nodes: make([]core.ThreadGraphNode, threadSpatialMaxNodes+1)}
	for index := range projection.Nodes {
		projection.Nodes[index] = core.ThreadGraphNode{
			TaskID: fmt.Sprintf("task-%04d", index), Label: fmt.Sprintf("task-%04d", index),
			Status: domain.StatusNextUp, Role: core.ThreadTaskMember,
			State: core.TaskGraphState{Role: core.RoleQueued, Gate: core.GateClear},
		}
	}
	first := renderThreadSpatial(projection, "", projection.Nodes[0].TaskID, 100, 20, &testStyles)
	second := renderThreadSpatial(projection, "", projection.Nodes[0].TaskID, 100, 20, &testStyles)
	if first != second {
		t.Fatal("capacity fallback was nondeterministic")
	}
	plain := ansi.Strip(first)
	for _, want := range []string{"bounded prototype fallback", "513 nodes", "no partial graph", "complete wave reader", "focus [M1]"} {
		if !strings.Contains(plain, want) {
			t.Errorf("capacity fallback omitted %q:\n%s", want, plain)
		}
	}
	for _, line := range strings.Split(plain, "\n") {
		if got := ansi.StringWidth(line); got > 100 {
			t.Fatalf("capacity fallback line width=%d: %q", got, line)
		}
	}
}

func TestThreadSpatialCanvasCapacityMatchesCellRepresentation(t *testing.T) {
	const maxCellBytes = uintptr(48)
	if got := unsafe.Sizeof(threadSpatialCell{}); got > maxCellBytes {
		t.Fatalf("thread spatial cell grew to %d bytes; reconsider the %d-cell canvas limit", got, threadSpatialMaxCanvasCells)
	}
	if bytes := uintptr(threadSpatialMaxCanvasCells) * unsafe.Sizeof(threadSpatialCell{}); bytes > 24*1024*1024 {
		t.Fatalf("canvas guard permits %d bytes of cell storage; want no more than 24 MiB", bytes)
	}
}

func TestThreadSpatialCapacityGuardsEdgesAndCanvasIndependently(t *testing.T) {
	t.Run("edges", func(t *testing.T) {
		projection := core.ThreadGraphProjection{Nodes: []core.ThreadGraphNode{
			{TaskID: "source", Role: core.ThreadTaskMember},
			{TaskID: "target", Role: core.ThreadTaskMember},
		}}
		for range threadSpatialMaxEdges + 1 {
			projection.Edges = append(projection.Edges, core.ThreadGraphEdge{From: "source", To: "target"})
		}
		layout := buildThreadSpatialLayout(projection)
		if len(layout.nodes) > threadSpatialMaxNodes || layout.width*layout.height > threadSpatialMaxCanvasCells {
			t.Fatalf("edge fixture tripped a different guard: nodes=%d canvas=%dx%d", len(layout.nodes), layout.width, layout.height)
		}
		if issue := threadSpatialCapacityIssue(layout, len(projection.Edges)); !strings.Contains(issue, "edges exceeds") {
			t.Fatalf("edge guard issue=%q", issue)
		}
	})

	t.Run("canvas cells", func(t *testing.T) {
		projection := core.ThreadGraphProjection{Nodes: make([]core.ThreadGraphNode, 505)}
		for index := range projection.Nodes {
			projection.Nodes[index] = core.ThreadGraphNode{TaskID: fmt.Sprintf("task-%03d", index), Role: core.ThreadTaskMember}
		}
		// A 255-node chain creates width while the other 250 sources create
		// height, staying below both the node and edge guards.
		for index := 1; index < 255; index++ {
			projection.Edges = append(projection.Edges, core.ThreadGraphEdge{
				From: projection.Nodes[index-1].TaskID, To: projection.Nodes[index].TaskID,
			})
		}
		layout := buildThreadSpatialLayout(projection)
		if len(layout.nodes) > threadSpatialMaxNodes || len(projection.Edges) > threadSpatialMaxEdges {
			t.Fatalf("canvas fixture tripped a different guard: nodes=%d edges=%d", len(layout.nodes), len(projection.Edges))
		}
		if issue := threadSpatialCapacityIssue(layout, len(projection.Edges)); !strings.Contains(issue, "canvas limit") {
			t.Fatalf("canvas guard issue=%q for %dx%d", issue, layout.width, layout.height)
		}
	})
}

func TestThreadSpatialNarrowFallbackAllocatesNoCanvasBeforeCapacityChecks(t *testing.T) {
	projection := core.ThreadGraphProjection{Nodes: make([]core.ThreadGraphNode, threadSpatialMaxNodes+1)}
	for index := range projection.Nodes {
		projection.Nodes[index] = core.ThreadGraphNode{TaskID: fmt.Sprintf("task-%04d", index), Role: core.ThreadTaskMember}
	}
	plain := ansi.Strip(renderThreadSpatial(projection, "", projection.Nodes[0].TaskID,
		threadSpatialMinWidth-1, threadSpatialMinHeight, &testStyles))
	if !strings.Contains(plain, "needs at least") || strings.Contains(plain, "capacity guard") {
		t.Fatalf("narrow no-canvas path did not remain independent of capacity fallback:\n%s", plain)
	}
}

func TestThreadTopologyCursorOpensSelectedTaskByStableIdentity(t *testing.T) {
	m, _ := threadModel(t)
	m = openThreads(t, m)
	m.setFocus(focusDetail)
	tm, _ := m.Update(press("v"))
	m = tm.(Model)
	target := selectedThreadDetail(t, m).detailSelectionKey()

	tm, cmd := m.Update(press("enter"))
	m = drainNested(t, tm.(Model), cmd)
	if m.cur().kind != entityTasks || m.selectedKey() != target {
		t.Fatalf("topology enter did not land on task %s: kind=%v selected=%q", target, m.cur().kind, m.selectedKey())
	}
	if m.cur().statusView != "all" {
		t.Fatalf("completed topology target did not widen unloaded Tasks view: %q", m.cur().statusView)
	}
	if len(m.navStack) != 1 || m.navStack[0].kind != entityThreads || m.navStack[0].ref.key != "6g503c6pfqeb" {
		t.Fatalf("topology origin was not retained on the back stack: %+v", m.navStack)
	}

	tm, cmd = m.Update(press("ctrl+o"))
	m = drainNested(t, tm.(Model), cmd)
	if m.cur().kind != entityThreads || m.selectedKey() != "6g503c6pfqeb" {
		t.Fatalf("back did not restore the Thread: kind=%v selected=%q", m.cur().kind, m.selectedKey())
	}
}

func TestThreadTopologyRendersHostileDeepWideDisconnectedAndPartialEvidence(t *testing.T) {
	projection := hostileThreadGraphProjection()
	if projection.TopologyComplete {
		t.Fatal("fixture should be partial because it includes a missing member")
	}

	for _, width := range []int{120, 54, 24, 12} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			first := renderThreadTopology(projection, "remote\npath\x1b", "", width, &testStyles)
			second := renderThreadTopology(projection, "remote\npath\x1b", "", width, &testStyles)
			if first != second {
				t.Fatal("same projection rendered nondeterministically")
			}
			plain := ansi.Strip(first)
			for _, line := range strings.Split(plain, "\n") {
				if got := ansi.StringWidth(line); got > width {
					t.Errorf("line width %d exceeds %d: %q", got, width, line)
				}
			}
			for _, want := range []string{"partial", "External", "Wave", "Unranked"} {
				if !strings.Contains(plain, want) {
					t.Errorf("width %d omitted %q:\n%s", width, want, plain)
				}
			}
			if width >= 24 && !strings.Contains(plain, "external-gate") {
				t.Errorf("width %d obscured the external role:\n%s", width, plain)
			}
			if width >= 54 && !strings.Contains(plain, "needs [") {
				t.Errorf("width %d obscured the compact prerequisite relation:\n%s", width, plain)
			}
		})
	}

	if got := terminalText("hostile\nlabel\t\x1b[31m"); got != "hostile↵label⇥�[31m" {
		t.Fatalf("terminal text was not neutralized: %q", got)
	}
}

func TestThreadTopologyKeepsCLIAndGraphExportProjectionEvidenceAligned(t *testing.T) {
	projection := hostileThreadGraphProjection()
	plainTUI := ansi.Strip(renderThreadTopology(projection, "", "", 160, &testStyles))
	var cli strings.Builder
	if err := clirender.ThreadPlanHuman(&cli, clirender.NewStyle(false), projection); err != nil {
		t.Fatal(err)
	}
	dot, err := graphfmt.DOT(projection)
	if err != nil {
		t.Fatal(err)
	}

	for _, wave := range projection.Waves {
		heading := fmt.Sprintf("Wave %d", wave.Index)
		if !strings.Contains(cli.String(), heading) || !strings.Contains(plainTUI, heading) {
			t.Errorf("projection wave %d diverged across CLI/TUI", wave.Index)
		}
	}
	for _, node := range projection.Nodes {
		if !strings.Contains(cli.String(), node.TaskID) || !strings.Contains(plainTUI, node.TaskID) || !strings.Contains(dot, node.TaskID) {
			t.Errorf("projection node %s diverged across plan/TUI/DOT", node.TaskID)
		}
	}
	if got := strings.Count(dot, " -> "); got != len(projection.Edges) {
		t.Fatalf("DOT edge count=%d want %d", got, len(projection.Edges))
	}
	aliases := threadGraphAliases(projection)
	incoming := threadGraphIncoming(projection.Edges, aliases)
	references := 0
	for _, prerequisites := range incoming {
		references += len(prerequisites)
		if relation := "needs [" + strings.Join(prerequisites, "], [") + "]"; !strings.Contains(plainTUI, relation) {
			t.Errorf("TUI omitted compact relation %q", relation)
		}
	}
	if references != len(projection.Edges) {
		t.Fatalf("TUI prerequisite reference count=%d want %d", references, len(projection.Edges))
	}
}

func TestThreadFollowPickerNavigatesByStableTaskIdentityAndBack(t *testing.T) {
	m, _ := threadModel(t)
	m = openThreads(t, m)
	m.setFocus(focusDetail)
	tm, _ := m.Update(press("v"))
	m = tm.(Model)
	tm, _ = m.Update(press("j"))
	m = tm.(Model)
	wantTarget := selectedThreadDetail(t, m).detailSelectionKey()

	tm, cmd := m.Update(press("f"))
	m = tm.(Model)
	if cmd != nil || !m.follow.active || len(m.follow.tasks) != 2 {
		t.Fatalf("Thread follow picker did not open over both tasks: active=%v tasks=%d cmd=%v flash=%q",
			m.follow.active, len(m.follow.tasks), cmd != nil, m.flash)
	}
	target := m.follow.selected()
	if target.CanonicalID() != wantTarget {
		t.Fatalf("Thread follow picker selected %q want topology cursor %q", target.CanonicalID(), wantTarget)
	}
	tm, cmd = m.Update(press("enter"))
	m = drainNested(t, tm.(Model), cmd)
	if m.cur().kind != entityTasks || m.selectedKey() != target.CanonicalID() {
		t.Fatalf("Thread follow did not land on task %s: kind=%v selected=%q", target.CanonicalID(), m.cur().kind, m.selectedKey())
	}
	if len(m.navStack) != 1 || m.navStack[0].kind != entityThreads || m.navStack[0].ref.key != "6g503c6pfqeb" {
		t.Fatalf("Thread origin was not retained on the back stack: %+v", m.navStack)
	}

	tm, cmd = m.Update(press("ctrl+o"))
	m = drainNested(t, tm.(Model), cmd)
	if m.cur().kind != entityThreads || m.selectedKey() != "6g503c6pfqeb" {
		t.Fatalf("back did not restore the Thread: kind=%v selected=%q", m.cur().kind, m.selectedKey())
	}
}

func TestThreadFollowTargetsIncludeExternalGatesAndSkipMissingMembers(t *testing.T) {
	projection := hostileThreadGraphProjection()
	tasks := threadFollowTasks(projection)
	if len(tasks) != len(projection.Nodes)-1 {
		t.Fatalf("follow targets=%d want every readable node (%d)", len(tasks), len(projection.Nodes)-1)
	}
	wantIDs := make([]string, 0, len(tasks))
	for _, node := range projection.Nodes {
		if node.Label != node.TaskID { // the missing member falls back to its id label
			wantIDs = append(wantIDs, node.TaskID)
		}
	}
	gotIDs := make([]string, 0, len(tasks))
	for _, task := range tasks {
		gotIDs = append(gotIDs, task.CanonicalID())
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("follow targets lost projection node order:\n got %v\nwant %v", gotIDs, wantIDs)
	}

	picker := followMenu{}
	picker.open("large-thread", tasks)
	picker.move(-1) // wrap to the final task, well below a short terminal's first window
	view := ansi.Strip(picker.view(&testStyles, 60, 10))
	if !strings.Contains(view, tasks[len(tasks)-1].Slug) || !strings.Contains(view, fmt.Sprintf("%d/%d", len(tasks), len(tasks))) {
		t.Fatalf("short picker clipped its selected final task:\n%s", view)
	}
	if lines := strings.Count(view, "\n") + 1; lines > 10 {
		t.Fatalf("short picker rendered %d lines into height 10:\n%s", lines, view)
	}
}

func hostileThreadGraphProjection() core.ThreadGraphProjection {
	gate := domain.Task{ID: testutil.TaskID("external-gate"), Slug: "external-gate", Status: domain.StatusCompleted}
	root := domain.Task{
		ID: testutil.TaskID("root"), Slug: "hostile\n\x1b[31m-根", Status: domain.StatusCompleted,
		DependsOn: []string{gate.ID},
	}
	tasks := []domain.Task{gate, root}
	members := []string{root.ID}
	previous := root.ID
	for i := 0; i < 8; i++ {
		task := domain.Task{
			ID: testutil.TaskID(fmt.Sprintf("deep-%d", i)), Slug: fmt.Sprintf("deep-%d", i),
			Status: domain.StatusNextUp, DependsOn: []string{previous},
		}
		tasks = append(tasks, task)
		members = append(members, task.ID)
		previous = task.ID
	}
	wides := make([]string, 0, 8)
	for i := 0; i < 8; i++ {
		task := domain.Task{
			ID: testutil.TaskID(fmt.Sprintf("wide-%d", i)), Slug: fmt.Sprintf("wide-%d-移行", i),
			Status: domain.StatusReadyToStart, DependsOn: []string{root.ID},
		}
		tasks = append(tasks, task)
		members = append(members, task.ID)
		wides = append(wides, task.ID)
	}
	join := domain.Task{
		ID: testutil.TaskID("join"), Slug: "fan-in", Status: domain.StatusNextUp,
		DependsOn: wides,
	}
	disconnected := domain.Task{ID: testutil.TaskID("disconnected"), Slug: "disconnected", Status: domain.StatusReadyToStart}
	tasks = append(tasks, join, disconnected)
	members = append(members, join.ID, disconnected.ID, testutil.TaskID("missing-member"))
	sort.Strings(members)
	thread := domain.Thread{
		ID: testutil.TaskID("topology-thread"), FilenameID: testutil.TaskID("topology-thread"),
		Slug: "topology-stress", Status: domain.ThreadStatusInProgress,
		Description: "stress terminal topology", Goal: "render supplied graph evidence",
		Created: "2026-09-03", Tasks: members,
	}
	return core.ProjectThreadGraph(thread, core.NewTaskGraph(tasks, nil))
}

func TestThreadProjectionStatesRemainVisuallyDistinct(t *testing.T) {
	task := func(name string, status domain.Status, dependsOn ...string) domain.Task {
		sort.Strings(dependsOn)
		return domain.Task{
			ID: testutil.TaskID(name), Slug: name, Status: status,
			Description: name, DependsOn: dependsOn,
		}
	}
	thread := func(name string, status domain.ThreadStatus, taskIDs ...string) domain.Thread {
		sort.Strings(taskIDs)
		id := testutil.TaskID("thread-" + name)
		return domain.Thread{
			ID: id, FilenameID: id, Slug: name, Status: status,
			Description: name, Goal: "exercise " + name, Created: "2026-09-02",
			Tasks: taskIDs,
		}
	}

	done := task("done", domain.StatusCompleted)
	shared := task("shared-member", domain.StatusReadyToStart)
	external := task("outside-gate", domain.StatusNextUp)
	gated := task("gated-member", domain.StatusReadyToStart, external.ID)
	inFlight := task("in-flight", domain.StatusInProgress)
	ready := task("ready", domain.StatusNextUp)
	graph := core.NewTaskGraph([]domain.Task{done, shared, external, gated, inFlight, ready}, nil)
	missing := testutil.TaskID("missing")

	views := map[string]core.ThreadView{
		"completed-unsound": core.ProjectThread(thread("completed-unsound", domain.ThreadStatusCompleted, missing), graph),
		"cancelled":         core.ProjectThread(thread("cancelled", domain.ThreadStatusCancelled, done.ID), graph),
		"empty":             core.ProjectThread(thread("empty", domain.ThreadStatusUnstarted), graph),
		"shared-a":          core.ProjectThread(thread("shared-a", domain.ThreadStatusInProgress, shared.ID), graph),
		"shared-b":          core.ProjectThread(thread("shared-b", domain.ThreadStatusInProgress, shared.ID), graph),
		"externally-gated":  core.ProjectThread(thread("externally-gated", domain.ThreadStatusInProgress, gated.ID), graph),
		"healthy-active":    core.ProjectThread(thread("healthy-active", domain.ThreadStatusInProgress, inFlight.ID, ready.ID), graph),
	}

	rendered := make(map[string]string, len(views))
	for name, view := range views {
		it := threadItem{view: view, countsW: 3}
		l := list.New([]list.Item{it}, threadDelegate{st: &testStyles}, 96, 5)
		var out strings.Builder
		threadDelegate{st: &testStyles}.Render(&out, l, 0, it)
		rendered[name] = ansi.Strip(out.String()) + "\n" + ansi.Strip(renderThreadMeta(threadDetail{projection: core.ThreadGraphProjection{View: view}}, 120, &testStyles))
	}
	for name, want := range map[string][]string{
		"completed-unsound": {"completed", "inconsistent", "Diagnostics", "missing-thread-member"},
		"cancelled":         {"cancelled", "1/1 nominally done"},
		"empty":             {"unstarted", "0/0 nominally done"},
		"externally-gated":  {"Immediate external gates (not members)", "outside-gate", "outstanding"},
		"healthy-active":    {"▶1", "✓1", "In flight", "Dispatchable frontier"},
	} {
		for _, fragment := range want {
			if !strings.Contains(rendered[name], fragment) {
				t.Errorf("%s lost %q:\n%s", name, fragment, rendered[name])
			}
		}
	}
	for _, name := range []string{"shared-a", "shared-b"} {
		if !strings.Contains(rendered[name], shared.ID) || !strings.Contains(rendered[name], name) {
			t.Errorf("%s does not retain its own identity plus the shared member:\n%s", name, rendered[name])
		}
	}

	seen := make(map[string]string)
	for name, output := range rendered {
		if prior, duplicate := seen[output]; duplicate {
			t.Errorf("%s and %s collapsed to indistinguishable output", prior, name)
		}
		seen[output] = name
	}
}
