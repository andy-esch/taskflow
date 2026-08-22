package tui

import (
	"errors"
	"fmt"
	"path/filepath"
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
	// The atlas-landing shape: `ui` outside any planning repo, with alpha seeded behind
	// the atlas as the space `esc` falls back into.
	m := New(core.NewService(adapter.trees[alpha]),
		WithWorkspaceOpening(core.NewWorkspaceService(adapter)),
		WithAtlas(core.NewSpaceOverviewService(registry, adapter)),
		WithAtlasLanding())
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
	// No "current" here on purpose: nothing is chosen yet on the landing atlas — see
	// TestAtlasMarksNoCardCurrentUntilASpaceIsChosen. It is asserted below, once a space
	// has actually been entered.
	for _, want := range []string{"atlas", "alpha", "alpha-impl", "beta", "~/git/alpha", "order name"} {
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

	// Back on the atlas, the entry point just entered is the one marked current.
	tm, _ = m.Update(press("a"))
	m = tm.(Model)
	if !strings.Contains(m.View().Content, "current") {
		t.Fatalf("atlas did not mark the entered entry point current:\n%s", m.View().Content)
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
	if m.atlasTheme.Name != atlasTheme.Name {
		t.Fatalf("atlas theme = %q, want %q", m.atlasTheme.Name, atlasTheme.Name)
	}
	// Run owns the light/dark choice for both palettes; the option must not pre-bake one.
	*m.atlasSt = newStyles(m.atlasTheme.For(true))
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

// In-repo startup opens THAT repo however many spaces are registered: standing in a
// planning repo says which space you meant, so the atlas must not interpose itself.
func TestAtlasInRepoStartupOpensTheRepoAndKeepsTheAtlasOneKeyAway(t *testing.T) {
	_, adapter, alpha, _ := atlasTestModel(t)
	registry := core.NewSpaceRegistryService(adapter)
	m := New(core.NewService(adapter.trees[alpha]),
		WithWorkspaceOpening(core.NewWorkspaceService(adapter)),
		WithAtlas(core.NewSpaceOverviewService(registry, adapter))) // no WithAtlasLanding
	m.workspace = core.Workspace{Checkout: alpha, PlanningRoot: alpha, Planning: m.svc, Layout: noWatchLayout{}}
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = tm.(Model)
	m = drainBatch(t, m, m.Init())
	if m.onAtlas || !m.dash.loaded {
		t.Fatalf("in-repo startup should land on overview: atlas=%v dash=%v", m.onAtlas, m.dash.loaded)
	}
	if len(m.atlas.spaces) != 2 {
		t.Fatalf("the atlas projection should still load in the background: spaces=%d", len(m.atlas.spaces))
	}
	tm, _ = m.Update(press("a"))
	if m = tm.(Model); !m.onAtlas {
		t.Fatal("explicit atlas should remain one keystroke away")
	}
}

// Outside a repo the atlas IS the landing screen, single registered space or not — it is
// the only surface that can say which space the seeded workspace even is.
func TestAtlasLandingHoldsEvenWithOneRegisteredSpace(t *testing.T) {
	_, adapter, alpha, _ := atlasTestModel(t)
	adapter.entries = adapter.entries[:1]
	registry := core.NewSpaceRegistryService(adapter)
	m := New(core.NewService(adapter.trees[alpha]),
		WithWorkspaceOpening(core.NewWorkspaceService(adapter)),
		WithAtlas(core.NewSpaceOverviewService(registry, adapter)),
		WithAtlasLanding())
	m.workspace = core.Workspace{Checkout: alpha, PlanningRoot: alpha, Planning: m.svc, Layout: noWatchLayout{}}
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = tm.(Model)
	m = drainBatch(t, m, m.Init())
	if !m.onAtlas || len(m.atlas.spaces) != 1 {
		t.Fatalf("no-repo startup should land on the atlas: atlas=%v spaces=%d", m.onAtlas, len(m.atlas.spaces))
	}
	tm, _ = m.Update(press("esc"))
	if m = tm.(Model); m.onAtlas {
		t.Fatal("esc should fall back into the seeded workspace")
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
		WithAtlas(core.NewSpaceOverviewService(registry, adapter)),
		WithAtlasLanding())
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

// The header names the sort order and the status row explains why ⏎ did nothing; both
// have to survive a registry too tall for the terminal, whichever card is focused.
func TestAtlasPinsHeaderAndStatusAroundTheScrollingCards(t *testing.T) {
	m, _, _, _ := atlasTestModel(t)
	spaces := make([]atlasSpace, 6)
	for i := range spaces {
		id := fmt.Sprintf("space-%d", i)
		spaces[i] = atlasSpace{summary: core.SpaceSummary{
			ID: id, PlanningID: id, Summary: &core.Summary{},
			Entries: []core.SpaceEntryPoint{{ID: id, Checkout: "/" + id, State: core.SpaceStateOK}},
		}}
	}
	m.atlas.spaces = spaces
	m.atlas.openErr = "beta is unreachable"
	for _, cursor := range []int{0, 3, 5} {
		m.atlas.cursor = cursor
		view := m.atlas.view(m.st, m.workspace, 100, 12)
		if rows := strings.Split(view, "\n"); len(rows) > 12 {
			t.Fatalf("cursor %d: atlas overran its %d-row budget with %d rows", cursor, 12, len(rows))
		}
		if !strings.Contains(view, "atlas · 6 space(s)") {
			t.Errorf("cursor %d: header scrolled off:\n%s", cursor, view)
		}
		if !strings.Contains(view, "beta is unreachable") {
			t.Errorf("cursor %d: open error scrolled off:\n%s", cursor, view)
		}
		if !strings.Contains(view, fmt.Sprintf("space-%d", cursor)) {
			t.Errorf("cursor %d: focused card scrolled off:\n%s", cursor, view)
		}
	}
}

// Durable identity, not path spelling: the registry keeps a symlinked/relatively-spelled
// path while an opened workspace carries the resolved one, so a string compare of the two
// would drop the badge from the space the user is standing in.
func TestAtlasCurrentBadgeFollowsIdentityNotPathSpelling(t *testing.T) {
	entry := core.SpaceEntryPoint{
		ID: "alpha", Checkout: "/logical/alpha", PlanningID: "planning-alpha", State: core.SpaceStateOK,
	}
	summary := core.SpaceSummary{ID: "alpha", PlanningID: "planning-alpha", Entries: []core.SpaceEntryPoint{entry}}
	current := core.Workspace{
		SpaceID: "alpha", Checkout: "/physical/alpha", PlanningRoot: "/physical/alpha",
		PlanningID: "planning-alpha",
	}
	if !isCurrentEntry(current, entry) || !isCurrentSpace(current, summary) {
		t.Fatal("an entry opened by the atlas must stay marked current under a different path spelling")
	}
	// The launch workspace carries no SpaceID; identity still resolves it by planning id.
	launch := core.Workspace{Checkout: "/physical/alpha", PlanningID: "planning-alpha"}
	if !isCurrentSpace(launch, summary) {
		t.Fatal("the launch workspace should still mark its own logical space")
	}
	other := core.Workspace{SpaceID: "beta", Checkout: "/logical/alpha", PlanningID: "planning-beta"}
	if isCurrentEntry(other, entry) || isCurrentSpace(other, summary) {
		t.Fatal("a different space must not be marked current merely by sharing a path string")
	}
}

// Entering the atlas is not navigation away: the space stays open, so the detail focus
// and zoom the atlas had to clobber come back with it.
func TestAtlasRoundTripRestoresDetailFocusAndZoom(t *testing.T) {
	m, _, _, _ := atlasTestModel(t)
	tm, _ := m.Update(press("a")) // atlas → the seeded space
	m = tm.(Model)
	m.onDash = false
	m.focus = focusDetail
	m.zoom = true
	tm, _ = m.Update(press("a"))
	if m = tm.(Model); !m.onAtlas || m.focus != focusList || m.zoom {
		t.Fatalf("the atlas needs list focus and no zoom: atlas=%v focus=%v zoom=%v", m.onAtlas, m.focus, m.zoom)
	}
	tm, _ = m.Update(press("a"))
	if m = tm.(Model); m.onAtlas || m.focus != focusDetail || !m.zoom {
		t.Fatalf("round trip lost browsing state: atlas=%v focus=%v zoom=%v", m.onAtlas, m.focus, m.zoom)
	}
}

// Switching spaces is navigation away, so the incoming space's own cached state wins over
// the snapshot taken from the space that was left.
func TestAtlasSwitchDiscardsTheResumeSnapshotOfTheSpaceLeft(t *testing.T) {
	m, _, _, beta := atlasTestModel(t)
	tm, _ := m.Update(press("a"))
	m = tm.(Model)
	m.onDash = false
	m.focus = focusDetail
	m.zoom = true
	tm, _ = m.Update(press("a")) // into the atlas, snapshot taken
	m = tm.(Model)
	tm, _ = m.Update(press("j")) // select beta
	m = tm.(Model)
	tm, cmd := m.Update(press("enter"))
	m = applyAtlasOpen(t, tm.(Model), cmd)
	if m.workspace.Checkout != beta {
		t.Fatalf("expected to enter beta, got %q", m.workspace.Checkout)
	}
	if m.focus != focusList || m.zoom {
		t.Fatalf("a first visit must start clean, not on the previous space's state: focus=%v zoom=%v", m.focus, m.zoom)
	}
}

// Launched outside a repo the browser seeds a space to keep its surfaces live; the chip
// must not claim the user is IN it until they say so. Any route off the atlas is that
// choice, after which the chip names the space for the rest of the session.
func TestAtlasChipNamesTheAtlasUntilASpaceIsChosen(t *testing.T) {
	m, _, alpha, _ := atlasTestModel(t) // the atlas-landing shape
	if got := m.spaceName(); got != atlasName {
		t.Fatalf("chip = %q on the landing atlas, want %q", got, atlasName)
	}
	if strip := m.tabStrip(); !strings.Contains(strip, "["+atlasName+"]") {
		t.Fatalf("tab strip = %q, want the atlas chip", strip)
	}
	tm, _ := m.Update(press("esc")) // fall back into the seeded space
	m = tm.(Model)
	if got := m.spaceName(); got != filepath.Base(alpha) {
		t.Fatalf("chip = %q after choosing, want the seeded space %q", got, filepath.Base(alpha))
	}
	tm, _ = m.Update(press("a")) // back to the atlas: the space is chosen now, so it stays named
	if m = tm.(Model); m.spaceName() != filepath.Base(alpha) {
		t.Fatalf("chip = %q on a later atlas visit, want the chosen space", m.spaceName())
	}
}

// In-repo startup means the space was chosen by standing in it, so visiting the atlas
// never blanks the chip.
func TestAtlasChipKeepsTheRepoNameWhenLaunchedInsideOne(t *testing.T) {
	_, adapter, alpha, _ := atlasTestModel(t)
	registry := core.NewSpaceRegistryService(adapter)
	m := New(core.NewService(adapter.trees[alpha]),
		WithWorkspaceOpening(core.NewWorkspaceService(adapter)),
		WithAtlas(core.NewSpaceOverviewService(registry, adapter))) // no WithAtlasLanding
	m.workspace = core.Workspace{Checkout: alpha, PlanningRoot: alpha, Planning: m.svc, Layout: noWatchLayout{}}
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = drainBatch(t, tm.(Model), m.Init())
	tm, _ = m.Update(press("a"))
	if m = tm.(Model); !m.onAtlas || m.spaceName() != filepath.Base(alpha) {
		t.Fatalf("chip = %q while visiting the atlas from a repo, want %q", m.spaceName(), filepath.Base(alpha))
	}
}

// The chip reading `atlas` and a card reading `current` are contradictory claims about
// the same moment. Until a space is chosen the seeded workspace is scaffolding, so no
// card is marked; once chosen, the badge tracks it normally.
func TestAtlasMarksNoCardCurrentUntilASpaceIsChosen(t *testing.T) {
	m, _, alpha, _ := atlasTestModel(t) // the atlas-landing shape, alpha seeded behind it
	if m.spaceName() != atlasName {
		t.Fatalf("setup: expected the unchosen landing atlas, chip = %q", m.spaceName())
	}
	if got := m.currentForAtlas(); got.Checkout != "" || got.PlanningID != "" {
		t.Fatalf("unchosen atlas marked %+v as current; want the zero workspace", got)
	}
	if strings.Contains(m.renderBody(), "current") {
		t.Errorf("landing atlas rendered a `current` badge while the chip reads %q:\n%s",
			atlasName, m.renderBody())
	}
	tm, _ := m.Update(press("esc")) // fall into the seeded space — now it IS chosen
	m = tm.(Model)
	tm, _ = m.Update(press("a"))
	m = tm.(Model)
	if m.currentForAtlas().Checkout != alpha {
		t.Fatalf("after choosing, current = %+v, want %q", m.currentForAtlas(), alpha)
	}
	if !strings.Contains(m.renderBody(), "current") {
		t.Errorf("atlas dropped the `current` badge for a chosen space:\n%s", m.renderBody())
	}
}
