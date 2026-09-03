package tui

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	clirender "github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/graphfmt"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
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
	if !strings.Contains(footer, "v summary") || !strings.Contains(footer, "f tasks") ||
		!strings.Contains(footer, "j/k task") || !strings.Contains(footer, "⏎ open") ||
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
	detail := selectedThreadDetail(t, m)
	if detail.detailViewName() != string(threadDetailSummary) || detail.rawBody() != "# Thread: Delivery\n\nbody\n" {
		t.Fatalf("summary was not restored after resize: view=%q body=%q", detail.detailViewName(), detail.rawBody())
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
