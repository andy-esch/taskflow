package tui

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
)

type atlasTestAdapter struct {
	entries  []core.SpaceEntryPoint
	trees    map[string]*store.FS
	openErrs map[string]error
	listErr  error
}

func (a *atlasTestAdapter) ListSpaceEntries() ([]core.SpaceEntryPoint, error) {
	if a.listErr != nil {
		return nil, a.listErr
	}
	return append([]core.SpaceEntryPoint(nil), a.entries...), nil
}

func (a *atlasTestAdapter) PrepareSpace(string) (core.SpaceRegistration, error) {
	return core.SpaceRegistration{}, errors.New("not used")
}

func (a *atlasTestAdapter) AddSpace(core.SpaceRegistration, bool) (core.SpaceEntryPoint, bool, error) {
	return core.SpaceEntryPoint{}, false, errors.New("not used")
}

func (a *atlasTestAdapter) ForgetSpace(string, bool) (core.SpaceEntryPoint, bool, error) {
	return core.SpaceEntryPoint{}, false, errors.New("not used")
}

func (a *atlasTestAdapter) OpenPlanningStore(root string) (core.SummaryStore, error) {
	fs, ok := a.trees[root]
	if !ok {
		return nil, errors.New("missing summary tree")
	}
	return fs, nil
}

func (a *atlasTestAdapter) OpenWorkspace(start string) (core.WorkspaceSource, error) {
	if err := a.openErrs[start]; err != nil {
		return core.WorkspaceSource{}, err
	}
	fs, ok := a.trees[start]
	if !ok {
		return core.WorkspaceSource{}, errors.New("missing workspace tree")
	}
	planningRoot, planningID := start, ""
	for _, entry := range a.entries {
		if entry.Checkout == start {
			planningRoot, planningID = entry.Root, entry.PlanningID
			break
		}
	}
	return core.WorkspaceSource{
		Checkout: start, PlanningRoot: planningRoot, PlanningID: planningID,
		Store: fs, Layout: noWatchLayout{},
	}, nil
}

type noWatchLayout struct{}

func (noWatchLayout) WatchPaths() []string { return nil }

func atlasTestModel(t *testing.T) (Model, *atlasTestAdapter, string, string) {
	t.Helper()
	alpha := seedRepo(t)
	beta := seedRepo(t)
	adapter := &atlasTestAdapter{
		entries: []core.SpaceEntryPoint{
			{ID: "alpha", Path: "~/git/alpha", Checkout: alpha, Root: alpha, PlanningID: "planning-alpha", Role: core.SpaceRoleDirect, State: core.SpaceStateOK},
			{ID: "alpha-impl", Path: "~/git/alpha-impl", Checkout: alpha + "-impl", Root: alpha, PlanningID: "planning-alpha", Role: core.SpaceRolePointer, State: core.SpaceStateOK},
			{ID: "beta", Path: "~/git/beta", Checkout: beta, Root: beta, PlanningID: "planning-beta", Role: core.SpaceRoleDirect, State: core.SpaceStateOK},
		},
		trees: map[string]*store.FS{
			alpha: store.NewFS(alpha), beta: store.NewFS(beta),
		},
		openErrs: make(map[string]error),
	}
	// The pointer entry resolves to alpha's planning store while retaining its own
	// selected checkout address.
	adapter.trees[alpha+"-impl"] = store.NewFS(alpha)
	registry := core.NewSpaceRegistryService(adapter)
	m := New(core.NewService(adapter.trees[alpha]),
		WithWorkspaceOpening(core.NewWorkspaceService(adapter)),
		WithAtlas(core.NewSpaceOverviewService(registry, adapter)))
	m.workspace = core.Workspace{
		Checkout: alpha, PlanningRoot: alpha, PlanningID: "planning-alpha",
		Planning: m.svc, Layout: noWatchLayout{},
	}
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	m = drainBatch(t, m, m.Init())
	return m, adapter, alpha, beta
}

func applyAtlasOpen(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected workspace-open command")
	}
	return settleAtlasCmd(t, m, cmd)
}

func settleAtlasCmd(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		return m
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, child := range batch {
			m = settleAtlasCmd(t, m, child)
		}
		return m
	}
	tm, next := m.Update(msg)
	return settleAtlasCmd(t, tm.(Model), next)
}

func TestAtlasLoadsGroupedCardsAndNavigatesThroughSelectedEntry(t *testing.T) {
	m, _, alpha, _ := atlasTestModel(t)
	if !m.onAtlas || len(m.atlas.spaces) != 2 {
		t.Fatalf("atlas landing = on:%v spaces:%d", m.onAtlas, len(m.atlas.spaces))
	}
	view := m.View().Content
	for _, want := range []string{"atlas", "alpha", "alpha-impl", "beta", "~/git/alpha", "current", "order name"} {
		if !strings.Contains(view, want) {
			t.Fatalf("atlas view missing %q:\n%s", want, view)
		}
	}

	// The grouped alpha card defaults to its direct entry; h/l selects the pointer
	// without duplicating the logical space, and Enter retains that local address.
	tm, _ := m.Update(press("l"))
	m = tm.(Model)
	entry, _ := m.atlas.selectedEntry()
	if entry.ID != "alpha-impl" {
		t.Fatalf("selected entry = %q", entry.ID)
	}
	tm, cmd := m.Update(press("enter"))
	m = applyAtlasOpen(t, tm.(Model), cmd)
	if m.onAtlas || m.workspace.SpaceID != "alpha-impl" || m.workspace.Checkout != alpha+"-impl" {
		t.Fatalf("opened workspace = %+v, onAtlas=%v", m.workspace, m.onAtlas)
	}
	if m.configStart != alpha+"-impl" {
		t.Fatalf("configuration editor still points at old workspace: %q", m.configStart)
	}
	if m.workspace.PlanningID != "planning-alpha" || m.workspace.PlanningRoot != alpha {
		t.Fatalf("pointer lost logical planning identity: %+v", m.workspace)
	}
}

func TestAtlasOrderCyclesWithoutLosingTheSelectedLogicalSpace(t *testing.T) {
	a := atlas{}
	active := core.Summary{InProgress: []domain.Task{{}, {}}}
	a.setOverview(core.SpaceOverview{Spaces: []core.SpaceSummary{
		{ID: "zeta", PlanningID: "planning-zeta", Summary: &active},
		{ID: "alpha", PlanningID: "planning-alpha", Summary: &core.Summary{}},
	}})
	if got := a.spaces[0].summary.ID; got != "alpha" {
		t.Fatalf("default order starts with %q, want alphabetical alpha", got)
	}
	a.cursor = 1 // keep zeta selected while the sort reshapes the cards
	a.cycleOrder()
	if got := a.spaces[0].summary.ID; got != "zeta" || a.cursor != 0 || a.order != atlasOrderActivity {
		t.Fatalf("activity order = first:%q cursor:%d order:%s", got, a.cursor, a.order.label())
	}
	a.cycleOrder()
	if got := a.spaces[0].summary.ID; got != "zeta" || a.order != atlasOrderRegistry {
		t.Fatalf("registry order = first:%q order:%s", got, a.order.label())
	}
	a.reverse = true
	a.applyOrder()
	if got := a.spaces[0].summary.ID; got != "alpha" || a.cursor != 1 {
		t.Fatalf("reversed registry order = first:%q selected cursor:%d", got, a.cursor)
	}
}

func TestAtlasUsesItsOwnHomeScopedStyles(t *testing.T) {
	atlasTheme, ok := design.Lookup("catppuccin")
	if !ok {
		t.Fatal("catppuccin test theme is not registered")
	}
	m := New(nil, WithAtlasTheme(atlasTheme))
	if m.atlasSt.pal.Accent.Hex != atlasTheme.Dark.Accent.Hex {
		t.Fatalf("atlas accent = %q, want global %q", m.atlasSt.pal.Accent.Hex, atlasTheme.Dark.Accent.Hex)
	}
	if m.st == m.atlasSt {
		t.Fatal("atlas and repository screens must not share mutable style state")
	}
}

func TestAtlasFailedOpenLeavesCurrentSessionIntact(t *testing.T) {
	m, adapter, alpha, beta := atlasTestModel(t)
	adapter.openErrs[beta] = errors.New("checkout disappeared")
	tm, _ := m.Update(press("j"))
	m = tm.(Model)
	tm, cmd := m.Update(press("enter"))
	m = applyAtlasOpen(t, tm.(Model), cmd)
	if !m.onAtlas || m.workspace.Checkout != alpha || !strings.Contains(m.atlas.openErr, "disappeared") {
		t.Fatalf("failed open changed session: workspace=%+v atlas=%v err=%q", m.workspace, m.onAtlas, m.atlas.openErr)
	}
}

func TestAtlasIdentityDriftLeavesCurrentSessionIntact(t *testing.T) {
	m, adapter, alpha, beta := atlasTestModel(t)
	// The registry projection still identifies beta as planning-beta, but discovery at
	// Enter now observes a replacement checkout. The core workspace boundary must reject
	// the time-of-check/time-of-use drift before the TUI swaps services.
	adapter.entries[2].PlanningID = "replacement-id"
	tm, _ := m.Update(press("j"))
	m = tm.(Model)
	tm, cmd := m.Update(press("enter"))
	m = applyAtlasOpen(t, tm.(Model), cmd)
	if !m.onAtlas || m.workspace.Checkout != alpha ||
		!strings.Contains(m.atlas.openErr, "no longer matches") {
		t.Fatalf("identity drift changed session: workspace=%+v atlas=%v err=%q beta=%q",
			m.workspace, m.onAtlas, m.atlas.openErr, beta)
	}
}

func TestAtlasRoundTripRestoresPerSpaceBrowsingState(t *testing.T) {
	m, _, alpha, _ := atlasTestModel(t)
	// Return to alpha, enter its task tab, and select the second row.
	tm, _ := m.Update(press("a"))
	m = tm.(Model)
	tm, cmd := m.Update(press("]"))
	m = drain(t, tm.(Model), cmd)
	m.cur().list.Select(1)
	wantID := m.selectedID()

	// Enter beta through the atlas.
	tm, _ = m.Update(press("a"))
	m = tm.(Model)
	tm, _ = m.Update(press("j"))
	m = tm.(Model)
	tm, cmd = m.Update(press("enter"))
	m = applyAtlasOpen(t, tm.(Model), cmd)
	if m.workspace.SpaceID != "beta" {
		t.Fatalf("first switch = %+v", m.workspace)
	}

	// Return to alpha. The card cursor is still beta, so j wraps to alpha; its cached
	// task tab and cursor return while the rows refresh against a freshly-opened service.
	tm, _ = m.Update(press("a"))
	m = tm.(Model)
	tm, _ = m.Update(press("j"))
	m = tm.(Model)
	tm, cmd = m.Update(press("enter"))
	m = applyAtlasOpen(t, tm.(Model), cmd)
	if workspaceKey(m.workspace) != workspaceKey(core.Workspace{Checkout: alpha}) || m.onDash ||
		m.cur().kind != entityTasks || m.selectedID() != wantID {
		t.Fatalf("restored session = checkout:%q dash:%v tab:%s selected:%q want:%q",
			m.workspace.Checkout, m.onDash, m.cur().name, m.selectedID(), wantID)
	}
}

func TestActivateWorkspaceReplacesAndClosesTheOnlyActiveWatcher(t *testing.T) {
	m, adapter, _, beta := atlasTestModel(t)
	oldDir, nextDir := t.TempDir(), t.TempDir()
	oldWatcher, err := newWatcher([]string{oldDir})
	if err != nil {
		t.Fatal(err)
	}
	nextWatcher, err := newWatcher([]string{nextDir})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = oldWatcher.close() })
	t.Cleanup(func() { _ = nextWatcher.close() })
	m.watch = oldWatcher
	next := core.Workspace{
		SpaceID: "beta", Checkout: beta, PlanningRoot: beta, PlanningID: "planning-beta",
		Planning: core.NewService(adapter.trees[beta]), Layout: noWatchLayout{},
	}

	cmd := m.activateWorkspace(next, nextWatcher, nil)
	if m.watch != nextWatcher || m.workspace.Checkout != beta {
		t.Fatalf("active watcher/workspace were not replaced: watcher=%p workspace=%+v", m.watch, m.workspace)
	}
	result := cmd()
	batch, ok := result.(tea.BatchMsg)
	if !ok || len(batch) < 2 {
		t.Fatalf("workspace activation command = %T, want close/load/watch batch", result)
	}
	if msg := batch[0](); msg != nil {
		t.Fatalf("watcher close command returned %T", msg)
	}
	if err := oldWatcher.fsw.Add(oldDir); err == nil {
		t.Fatal("previous workspace watcher still accepts paths after activation")
	}
}

func TestAtlasDropsStaleWorkspaceResultsAndOldSessionMessages(t *testing.T) {
	m, _, alpha, _ := atlasTestModel(t)
	// Fire beta, then choose alpha before beta lands. The older open generation must
	// not win even though both commands belong to the same source workspace session.
	tm, _ := m.Update(press("j"))
	m = tm.(Model)
	tm, betaCmd := m.Update(press("enter"))
	m = tm.(Model)
	m.atlas.move(1) // beta → alpha while the first open command is still running
	tm, alphaCmd := m.Update(press("enter"))
	m = applyAtlasOpen(t, tm.(Model), alphaCmd)
	if m.workspace.Checkout != alpha {
		t.Fatalf("newest open did not win: %+v", m.workspace)
	}
	tm, _ = m.Update(betaCmd())
	m = tm.(Model)
	if m.workspace.Checkout != alpha {
		t.Fatalf("stale open replaced workspace: %+v", m.workspace)
	}

	oldGen := m.sessionGen - 1
	tm, _ = m.Update(sessionMsg{gen: oldGen, msg: movedMsg{slug: "foreign", to: "completed"}})
	m = tm.(Model)
	if m.flash != "" {
		t.Fatalf("old session mutation landed in current workspace: %q", m.flash)
	}
}

func TestAtlasSingleSpaceStartupFallsBackButRemainsExplicitlyReachable(t *testing.T) {
	_, adapter, alpha, _ := atlasTestModel(t)
	adapter.entries = adapter.entries[:1]
	registry := core.NewSpaceRegistryService(adapter)
	m := New(core.NewService(adapter.trees[alpha]),
		WithWorkspaceOpening(core.NewWorkspaceService(adapter)),
		WithAtlas(core.NewSpaceOverviewService(registry, adapter)))
	m.workspace = core.Workspace{Checkout: alpha, PlanningRoot: alpha, Planning: m.svc, Layout: noWatchLayout{}}
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = tm.(Model)
	m = drainBatch(t, m, m.Init())
	if m.onAtlas || !m.dash.loaded {
		t.Fatalf("one-space startup should land on overview: atlas=%v dash=%v", m.onAtlas, m.dash.loaded)
	}
	tm, _ = m.Update(press("a"))
	m = tm.(Model)
	if !m.onAtlas || len(m.atlas.spaces) != 1 {
		t.Fatalf("explicit atlas should remain available: atlas=%v spaces=%d", m.onAtlas, len(m.atlas.spaces))
	}
}

func TestAtlasBrokenEntryIsVisibleButCannotReplaceWorkspace(t *testing.T) {
	m, adapter, alpha, _ := atlasTestModel(t)
	adapter.entries[1].State = core.SpaceStateMissing
	adapter.entries[1].Detail = "checkout is missing"
	// Refresh the projection, then select the broken pointer in alpha's grouped card.
	tm, cmd := m.Update(press("r"))
	m = settleAtlasCmd(t, tm.(Model), cmd)
	tm, _ = m.Update(press("l"))
	m = tm.(Model)
	tm, cmd = m.Update(press("enter"))
	m = tm.(Model)
	if cmd != nil || m.workspace.Checkout != alpha || !strings.Contains(m.atlas.openErr, "missing") {
		t.Fatalf("broken entry opened: cmd=%v workspace=%+v err=%q", cmd != nil, m.workspace, m.atlas.openErr)
	}
}

func TestAtlasRegistryFailurePreservesOrdinaryTUIAndExplicitDiagnosis(t *testing.T) {
	_, adapter, alpha, _ := atlasTestModel(t)
	adapter.listErr = errors.New("registry malformed")
	registry := core.NewSpaceRegistryService(adapter)
	m := New(core.NewService(adapter.trees[alpha]),
		WithWorkspaceOpening(core.NewWorkspaceService(adapter)),
		WithAtlas(core.NewSpaceOverviewService(registry, adapter)))
	m.workspace = core.Workspace{Checkout: alpha, PlanningRoot: alpha, Planning: m.svc, Layout: noWatchLayout{}}
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = tm.(Model)
	m = drainBatch(t, m, m.Init())
	if m.onAtlas || !m.dash.loaded || !strings.Contains(m.flash, "registry malformed") {
		t.Fatalf("startup fallback = atlas:%v dash:%v flash:%q", m.onAtlas, m.dash.loaded, m.flash)
	}
	tm, cmd := m.Update(press("a"))
	m = settleAtlasCmd(t, tm.(Model), cmd)
	if !m.onAtlas || m.atlas.loadErr == nil || !strings.Contains(m.View().Content, "atlas unavailable") {
		t.Fatalf("explicit diagnosis = atlas:%v err:%v\n%s", m.onAtlas, m.atlas.loadErr, m.View().Content)
	}
}

func TestAtlasHelpAndPaletteExposeNavigationWithoutHiddenLifecycleActions(t *testing.T) {
	m, _, _, _ := atlasTestModel(t)
	commands := strings.Join(m.paletteCommands(), " ")
	words := strings.Fields(commands)
	if !slices.Contains(words, atlasName) || slices.Contains(words, "complete") {
		t.Fatalf("atlas palette commands = %q", commands)
	}
	tm, _ := m.Update(press("?"))
	m = tm.(Model)
	help := m.View().Content
	for _, want := range []string{"choose a space", "choose an entry", "enter", "return"} {
		if !strings.Contains(help, want) {
			t.Fatalf("atlas help missing %q:\n%s", want, help)
		}
	}
}

func TestAtlasSelectedEntryStaysVisibleOnShortTerminal(t *testing.T) {
	m, _, _, _ := atlasTestModel(t)
	entries := make([]core.SpaceEntryPoint, 8)
	for i := range entries {
		entries[i] = core.SpaceEntryPoint{
			ID: "entry-" + string(rune('a'+i)), Checkout: "/entry", Role: core.SpaceRolePointer, State: core.SpaceStateOK,
		}
	}
	m.atlas.spaces = []atlasSpace{{summary: core.SpaceSummary{ID: "crowded", Entries: entries}, entry: 7}}
	m.atlas.cursor = 0
	view := m.atlas.view(m.st, m.workspace, 60, 6)
	if !strings.Contains(view, "entry-h") {
		t.Fatalf("selected entry was navigable but clipped:\n%s", view)
	}
}

func TestAtlasEnabledModelDoesNotHideEditorControlCommand(t *testing.T) {
	m, _, _, _ := atlasTestModel(t)
	tm, _ := m.Update(press("a")) // atlas → current overview
	m = tm.(Model)
	tm, cmd := m.Update(press("]"))
	m = drain(t, tm.(Model), cmd)
	tm, cmd = m.Update(press("E"))
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("E should return Bubble Tea's ExecProcess command")
	}
	msg := cmd() // returns the control value; it does not launch the editor itself
	if _, hidden := msg.(sessionMsg); hidden || !strings.Contains(fmt.Sprintf("%T", msg), "execMsg") {
		t.Fatalf("ExecProcess control message was hidden by session scoping: %T", msg)
	}
	clipboard := scopeSession(m.sessionGen, tea.SetClipboard("value"))()
	if _, hidden := clipboard.(sessionMsg); hidden {
		t.Fatalf("OSC52 control message was hidden by session scoping: %T", clipboard)
	}
}
