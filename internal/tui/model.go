// Package tui is the second primary adapter: an interactive Bubble Tea front-end
// over the same core.Service the CLI uses. It never touches the store/fs — every
// read runs as a tea.Cmd against the service (commands.go), so Update and View
// stay I/O-free. Entities (tasks/epics/audits) are declared in a registry
// (entity.go); the lists live-reload via fsnotify (watch.go). See
// docs/ARCHITECTURE.md for the subsystem map.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/theme"
)

type focus int

const (
	focusList focus = iota
	focusDetail
)

// Model is the root TUI model: a multi-entity browser (tasks/epics/audits) over
// the core service. A tab strip + `:` command-jump switch the active entity; each
// entity keeps its own list (and cursor). The right pane shows the selection's
// detail.
type Model struct {
	svc  *core.Service
	root string // planning root; reserved for the S3 fsnotify watch (not read yet)

	width, height int
	twoPane       bool
	listOuterW    int
	detailOuterW  int
	paneOuterH    int

	focus  focus
	tabs   []*entityTab
	active int
	detail detailPane
	cmd    commandBar

	showHelp bool // the `?` keybinding overlay is open

	watch     *watcher // fsnotify source (nil when unavailable / in tests); see watch.go
	watchOff  bool     // the watcher failed to start: live reload is off (footer note)
	dirtyGen  int      // bumped per fs event; the debounce tick fires a reload only when it matches
	detailGen int      // bumped per detail request; orders concurrent loads for the same id
}

// New constructs the root model over the same *core.Service the CLI uses.
func New(svc *core.Service, root string) Model {
	return Model{
		svc: svc, root: root, focus: focusList,
		tabs: newEntityTabs(), active: 0,
		detail: newDetailPane(), cmd: newCommandBar(),
	}
}

func (m Model) Init() tea.Cmd {
	if m.watch != nil {
		return tea.Batch(m.cur().reload(m.svc), waitForFS(m.watch))
	}
	return m.cur().reload(m.svc)
}

// reloadAll re-fires the loader for every loaded tab, each preserving its own
// cursor by id. Unvisited tabs are left alone (they reload fresh on first visit)
// — except the active tab, which always reloads: after a failed *initial* load
// nothing is `loaded`, and `r` must still be able to recover the session.
// This is the `r` / fsnotify path: a change from another process is reflected on
// whichever tab you land on, not just the active one.
func (m *Model) reloadAll() tea.Cmd {
	var cmds []tea.Cmd
	for i, t := range m.tabs {
		if !t.loaded && i != m.active {
			continue
		}
		t.markReload()
		cmds = append(cmds, t.reload(m.svc))
	}
	return tea.Batch(cmds...)
}

// cur returns the active entity tab. The tab is a pointer, so reads use a value
// receiver yet callers can still mutate the tab's list in place.
func (m Model) cur() *entityTab { return m.tabs[m.active] }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.recomputeLayout()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case listLoadedMsg:
		return m.handleListLoaded(msg)

	case detailMsg:
		if !m.isCurrentSelection(msg.kind, msg.id) || msg.gen != m.detailGen {
			return m, nil // stale: tab/selection changed, or a newer load is in flight
		}
		m.detail.SetContent(msg.content)
		return m, nil

	case detailErrMsg:
		// A per-item load failure (e.g. an ambiguous duplicate slug) shows in the
		// detail pane — it must not blank the whole browser.
		if !m.isCurrentSelection(msg.kind, msg.id) || msg.gen != m.detailGen {
			return m, nil
		}
		m.detail.SetError(msg.id, msg.err.Error())
		return m, nil

	case tabMsg:
		return m.handleTabMsg(msg)

	case reloadMsg:
		return m, m.reloadAll()

	case fsEventMsg:
		// A filesystem change: keep listening, and (re)arm the debounce. The reload
		// only fires from a debounce tick whose generation is still current, so an
		// editor's save-storm of events coalesces into one reload.
		m.dirtyGen++
		return m, tea.Batch(waitForFS(m.watch), debounceTick(m.dirtyGen))

	case debounceMsg:
		if msg.gen != m.dirtyGen {
			return m, nil // a newer event re-armed the debounce; this tick is stale
		}
		return m, func() tea.Msg { return reloadMsg{} }

	case errMsg:
		if i := indexOfKind(m.tabs, msg.kind); i >= 0 && msg.gen == m.tabs[i].loadGen {
			m.tabs[i].loadErr = msg.err // stale failures (an older gen) are dropped
		}
		return m, nil
	}
	// Forward anything else (e.g. cursor-blink ticks) to the active list. List
	// messages with a tab identity arrive as tabMsg above and route themselves.
	var cmd tea.Cmd
	m.cur().list, cmd = m.cur().list.Update(msg)
	return m, routeToTab(m.cur().kind, cmd)
}

// routeToTab wraps a list-internal Cmd so its message comes back tagged with the
// tab that owns it. The list's filter machinery is asynchronous (SetItems and
// filter keystrokes return a Cmd whose FilterMatchesMsg arrives later) — without
// the tag, a background tab's matches would be applied to the active tab.
func routeToTab(kind entityKind, cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		if msg == nil {
			return nil
		}
		if batch, ok := msg.(tea.BatchMsg); ok {
			wrapped := make([]tea.Cmd, len(batch))
			for i, c := range batch {
				wrapped[i] = routeToTab(kind, c)
			}
			return tea.BatchMsg(wrapped)
		}
		return tabMsg{kind: kind, msg: msg}
	}
}

// handleTabMsg applies a tab-tagged list message to its owning tab. If the tab
// has a pending cursor restore (a reload landed while a filter was applied, so
// the refilter was still in flight), it's retried once the matches arrive. For
// the active tab, a selection moved by the filter also refreshes the detail pane.
func (m Model) handleTabMsg(msg tabMsg) (tea.Model, tea.Cmd) {
	i := indexOfKind(m.tabs, msg.kind)
	if i < 0 || msg.msg == nil {
		return m, nil
	}
	tab := m.tabs[i]
	prev := ""
	if i == m.active {
		prev = m.selectedID()
	}
	var cmd tea.Cmd
	tab.list, cmd = tab.list.Update(msg.msg)
	if tab.restore != "" && tab.selectByID(tab.restore) {
		tab.restore = ""
	}
	cmd = routeToTab(msg.kind, cmd)
	if i != m.active {
		return m, cmd
	}
	return m.afterSelectionChange(prev, cmd)
}

// handleListLoaded applies an entity-list load to its tab (by kind, so a load
// that finishes after a tab switch still lands correctly). Every tab restores its
// own cursor by id (so an all-tabs reload preserves each); only the active tab
// also kicks off the selected item's detail load.
func (m Model) handleListLoaded(msg listLoadedMsg) (tea.Model, tea.Cmd) {
	i := indexOfKind(m.tabs, msg.kind)
	if i < 0 {
		return m, nil
	}
	tab := m.tabs[i]
	if msg.gen != tab.loadGen {
		return m, nil // an older load finishing late must not clobber the newer one
	}
	// A successful load clears the tab's error so a transient failure (e.g. the
	// planning dir briefly unreadable) recovers on the next `r`/reload.
	tab.loadErr = nil
	sortItems(msg.items, tab.sortKey, tab.sortRev) // honor the tab's sort across reloads
	// SetItems' refilter is async — route its FilterMatchesMsg back to THIS tab,
	// and keep the cursor restore pending until the matches land (selectByID sees
	// nothing while filteredItems is nil).
	cmd := routeToTab(msg.kind, tab.list.SetItems(msg.items))
	tab.loaded = true
	tab.problems = msg.problems
	if tab.restore != "" && tab.selectByID(tab.restore) {
		tab.restore = ""
	}
	if msg.kind != m.cur().kind {
		return m, cmd // a background tab loaded; leave the active view alone
	}
	return m, tea.Batch(cmd, m.refreshDetail())
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 0. The help overlay is modal: any key dismisses it (ctrl+c still quits).
	if m.showHelp {
		if key.Matches(msg, keys.ForceQuit) {
			return m, tea.Quit
		}
		m.showHelp = false
		return m, nil
	}

	// 1. The command bar captures every key while open (so `:tasks` typing never
	// leaks into global hotkeys).
	if m.cmd.active {
		switch {
		case key.Matches(msg, keys.ForceQuit):
			return m, tea.Quit
		case key.Matches(msg, keys.Back):
			m.cmd.blur()
			return m, nil
		case msg.Type == tea.KeyEnter:
			return m.dispatchCommand()
		case msg.Type == tea.KeyTab:
			m.cmd.complete(m.commandOptions())
			return m, nil
		}
		return m, m.cmd.update(msg)
	}

	// 2. The list's filter input owns every key while capturing a query. It runs
	// through updateList (not a bare forward) so the live-filter cursor moves
	// keep the detail pane in sync while typing.
	if m.cur().list.SettingFilter() {
		return m.updateList(msg)
	}

	// 2b. The detail pane's find input owns keys while a query is being typed
	// (ctrl+c still force-quits).
	if m.focus == focusDetail && m.detail.finding() {
		if key.Matches(msg, keys.ForceQuit) {
			return m, tea.Quit
		}
		return m, m.detail.updateFind(msg)
	}

	// 3. Global hotkeys.
	switch {
	case key.Matches(msg, keys.ForceQuit):
		return m, tea.Quit
	case key.Matches(msg, keys.Quit):
		// q is a *context* quit: in single-pane drill the detail pane is a layer,
		// so q pops back to the list (like Esc/h) instead of exiting the app.
		// In two-pane, detail focus isn't a layer — q quits from either pane.
		if m.focus == focusDetail && !m.twoPane {
			m.setFocus(focusList)
			return m, nil
		}
		return m, tea.Quit
	case key.Matches(msg, keys.Help):
		m.showHelp = true
		return m, nil
	case key.Matches(msg, keys.Command):
		return m, m.cmd.focus()
	case key.Matches(msg, keys.NextTab):
		return m, m.switchTab((m.active + 1) % len(m.tabs))
	case key.Matches(msg, keys.PrevTab):
		return m, m.switchTab((m.active - 1 + len(m.tabs)) % len(m.tabs))
	case key.Matches(msg, keys.Sort):
		return m, m.cycleSort(1)
	case key.Matches(msg, keys.SortRev):
		m.cur().sortRev = !m.cur().sortRev
		return m, m.applySortToCurrent()
	case key.Matches(msg, keys.StatusView):
		return m, m.cycleStatusView(1)
	case key.Matches(msg, keys.StatusRev):
		return m, m.cycleStatusView(-1)
	case key.Matches(msg, keys.Refresh):
		return m, func() tea.Msg { return reloadMsg{} }
	case key.Matches(msg, keys.ToggleFocus):
		m.toggleFocus()
		return m, nil
	}

	// 4. Focus-routed keys (list vs detail).
	if m.focus == focusList {
		switch {
		case key.Matches(msg, keys.Right):
			m.setFocus(focusDetail)
			return m, nil
		case key.Matches(msg, keys.Left):
			return m, nil // already leftmost
		}
		return m.updateList(msg)
	}
	switch {
	case key.Matches(msg, keys.Find):
		return m, m.detail.startFind()
	case key.Matches(msg, keys.FindNext):
		m.detail.findNext(1)
		return m, nil
	case key.Matches(msg, keys.FindPrev):
		m.detail.findNext(-1)
		return m, nil
	case key.Matches(msg, keys.Left), key.Matches(msg, keys.Back):
		// First Esc/h clears an active find; a second leaves the detail pane.
		if m.detail.findActive() {
			m.detail.clearFind()
			return m, nil
		}
		m.setFocus(focusList)
		return m, nil
	case key.Matches(msg, keys.Top):
		m.detail.vp.GotoTop()
		return m, nil
	case key.Matches(msg, keys.Bottom):
		m.detail.vp.GotoBottom()
		return m, nil
	}
	var cmd tea.Cmd
	m.detail.vp, cmd = m.detail.vp.Update(msg)
	return m, cmd
}

// updateList forwards a key to the active list, lazily loading the detail body
// when the selection changes.
func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	t := m.cur()
	prev := m.selectedID()
	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return m.afterSelectionChange(prev, routeToTab(t.kind, cmd))
}

// afterSelectionChange is the shared tail of every path that may move the active
// list's cursor (keys, filter keystrokes, async filter matches, reloads): if the
// selection changed, the detail pane follows it.
func (m Model) afterSelectionChange(prev string, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch id := m.selectedID(); id {
	case prev:
		return m, cmd
	case "":
		// A `/` filter narrowed the list to zero matches: drop the now-stale detail
		// instead of leaving the last item showing.
		m.detail.showEmpty()
		return m, cmd
	default:
		m.detail.loading = true
		return m, tea.Batch(cmd, m.loadDetail(id))
	}
}

// dispatchCommand resolves the typed `:` word to an entity tab or a task status
// view and applies it; an unknown word reopens the bar with an inline error.
func (m Model) dispatchCommand() (tea.Model, tea.Cmd) {
	word := m.cmd.value()
	m.cmd.blur()
	if word == "" {
		return m, nil
	}
	for i, t := range m.tabs {
		if t.matches(word) {
			return m, m.switchTab(i)
		}
	}
	if view, ok := statusViewFor(word); ok {
		return m, m.applyStatusView(view)
	}
	cmd := m.cmd.focus()
	m.cmd.err = "unknown: " + word
	return m, cmd
}

// switchTab makes tab i active: it resets focus + the detail pane, loads the tab
// on first visit, and (re)loads the selected item's detail. Per-tab cursors are
// preserved because each tab owns its list.
func (m *Model) switchTab(i int) tea.Cmd {
	if i == m.active {
		return nil
	}
	m.active = i
	m.focus = focusList
	m.detail.clear()
	if !m.cur().loaded {
		return m.cur().reload(m.svc)
	}
	return m.refreshDetail()
}

// refreshDetail (re)loads the detail for the current selection, or settles the
// pane into its empty state when the active tab has no items — so an empty tab
// (e.g. a repo with no audits) never sits on a perpetual "loading…".
func (m *Model) refreshDetail() tea.Cmd {
	id := m.selectedID()
	if id == "" {
		m.detail.loading = false
		return nil
	}
	m.detail.loading = true
	return m.loadDetail(id)
}

// loadDetail fires the active tab's item loader, stamping the response with a
// fresh request generation. (kind, id) alone can't order two loads for the SAME
// id — e.g. `r` plus an fs-debounce reload — and Cmds run concurrently, so the
// older read could land last and win without the stamp.
func (m *Model) loadDetail(id string) tea.Cmd {
	m.detailGen++
	gen := m.detailGen
	load := m.cur().loadItem(m.svc, id)
	return func() tea.Msg {
		switch msg := load().(type) {
		case detailMsg:
			msg.gen = gen
			return msg
		case detailErrMsg:
			msg.gen = gen
			return msg
		default:
			return msg
		}
	}
}

// isCurrentSelection reports whether (kind, id) still matches the active tab's
// selection — the stale guard for async detail loads.
func (m Model) isCurrentSelection(kind entityKind, id string) bool {
	return kind == m.cur().kind && id == m.selectedID()
}

func (m *Model) setFocus(f focus) { m.focus = f }

func (m *Model) toggleFocus() {
	if m.focus == focusList {
		m.setFocus(focusDetail)
	} else {
		m.setFocus(focusList)
	}
}

func (m Model) selectedID() string {
	if it, ok := m.cur().list.SelectedItem().(entityItem); ok {
		return it.id()
	}
	return ""
}

func (m Model) entityNames() []string {
	names := make([]string, len(m.tabs))
	for i, t := range m.tabs {
		names[i] = t.name
	}
	return names
}

// commandOptions is the full `:` Tab-completion set: entity names + their
// aliases + the task status-view words. (Aliases were missing in S2a.)
func (m Model) commandOptions() []string {
	words := statusViewWords()
	opts := make([]string, 0, len(m.tabs)*2+len(words))
	for _, t := range m.tabs {
		opts = append(opts, t.name)
		opts = append(opts, t.aliases...)
	}
	return append(opts, words...)
}

// --- interactive sort ---

// applySortToCurrent reorders the active list under its current sort state,
// preserving the cursor by id (SetItems re-applies any active `/` filter).
func (m *Model) applySortToCurrent() tea.Cmd {
	t := m.cur()
	id := m.selectedID()
	items := t.list.Items()
	sortItems(items, t.sortKey, t.sortRev)
	cmd := routeToTab(t.kind, t.list.SetItems(items))
	t.selectByID(id)
	return cmd
}

// cycleSort advances the active tab's sort column (wrapping) and re-sorts.
func (m *Model) cycleSort(dir int) tea.Cmd {
	cur := 0
	for i, k := range sortCols {
		if k == m.cur().sortKey {
			cur = i
			break
		}
	}
	n := len(sortCols)
	m.cur().sortKey = sortCols[((cur+dir)%n+n)%n]
	return m.applySortToCurrent()
}

// --- status views (tasks) ---

// cycleStatusView steps the tasks tab's status view (no-op on other entities,
// which have no status axis). The cycle order lives in statusViews (statusview.go).
func (m *Model) cycleStatusView(dir int) tea.Cmd {
	if m.cur().kind != entityTasks {
		return nil
	}
	return m.applyStatusView(statusViewStep(m.cur().statusView, dir))
}

// applyStatusView switches to the tasks tab, sets its status view, and reloads —
// preserving the cursor by id when the task survives into the new view.
func (m *Model) applyStatusView(view string) tea.Cmd {
	i := indexOfKind(m.tabs, entityTasks)
	m.active = i
	m.focus = focusList
	tab := m.tabs[i]
	tab.markReload()
	tab.statusView = view
	m.detail.clear()
	return tab.reload(m.svc)
}

// recomputeLayout sizes the tab strip, panes, and footer from the terminal size +
// responsive mode. Borders are subtracted before sizing children (the #1 lipgloss
// bug). Every tab's list is sized so a switch needs no relayout.
func (m *Model) recomputeLayout() {
	const (
		footerH = 1
		tabH    = 1 // the tab strip line
		titleH  = 1 // the detail pane's manual title line
	)
	bodyH := m.height - footerH - tabH
	if bodyH < 4 {
		bodyH = 4
	}
	m.paneOuterH = bodyH
	// bubbles/list renders its pagination footer ONE line *beyond* its SetHeight
	// (the `••` dots), so a paginated list would overflow its pane and shove the
	// footer/command bar off-screen. Reserve that line here so title+items+dots
	// fit the pane's inner height exactly.
	listH := max1(bodyH - paneVFrame - 1)
	detailH := max1(bodyH - paneVFrame - titleH)
	m.twoPane = m.width >= 90

	var listInnerW int
	if m.twoPane {
		listOuterW := m.width * 2 / 5
		if listOuterW < 28 {
			listOuterW = 28
		}
		m.listOuterW = listOuterW
		m.detailOuterW = m.width - listOuterW
		listInnerW = max1(listOuterW - paneHFrame)
		m.detail.SetSize(max1(m.detailOuterW-paneHFrame), detailH)
	} else {
		m.listOuterW, m.detailOuterW = m.width, m.width
		listInnerW = max1(m.width - paneHFrame)
		m.detail.SetSize(listInnerW, detailH)
	}
	for _, t := range m.tabs {
		t.list.SetSize(listInnerW, listH)
	}
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		// No WindowSizeMsg yet. Rendering panes now would use unset (0) sizes →
		// negative border dimensions → a broken oversized frame. Wait for size.
		return "loading…"
	}
	// Hard-clamp the body to its budget so the tab strip and footer (the chrome)
	// are ALWAYS rendered — a child that overflows its box loses its own bottom
	// edge, never the load-bearing navigation/command line. Belt-and-suspenders
	// with the per-list pagination reserve above.
	body := lipgloss.NewStyle().MaxHeight(m.paneOuterH).Render(m.bodyView())
	full := lipgloss.JoinVertical(lipgloss.Left, m.tabStrip(), body, m.footer())
	return lipgloss.NewStyle().MaxWidth(m.width).MaxHeight(m.height).Render(full)
}

// bodyView renders the pane area, with the `?` help panel floated over it when
// open (the underlying panes stay visible around the modal).
func (m Model) bodyView() string {
	base := m.renderBody()
	if !m.showHelp {
		return base
	}
	// Normalize the body to exact body dimensions, then composite the help box on
	// top so it floats over the items rather than blanking them.
	canvas := lipgloss.Place(m.width, m.paneOuterH, lipgloss.Left, lipgloss.Top, base)
	return overlay(canvas, helpBox(m.width-2, m.paneOuterH-2), m.width, m.paneOuterH)
}

// renderBody is the pane layout: a loading note (or this tab's load error) until
// the active tab loads, then the two-pane (or single-pane drill) view. Errors
// are per tab — one failing loader must not blank tabs that loaded fine — and a
// tab that HAS loaded keeps its (stale) rows on a failed reload, with the
// failure flagged in the footer instead.
func (m Model) renderBody() string {
	switch t := m.cur(); {
	case t.loadErr != nil && !t.loaded:
		return m.pane(focusList, fg(theme.ColorRed, "error: "+t.loadErr.Error()), m.width)
	case !t.loaded:
		return m.pane(focusList, dim("loading…"), m.width)
	}
	listPane := m.pane(focusList, m.listPaneContent(), m.listOuterW)
	switch {
	case m.twoPane:
		return lipgloss.JoinHorizontal(lipgloss.Top, listPane, m.detailPaneView())
	case m.focus == focusDetail:
		return m.detailPaneView()
	default:
		return listPane
	}
}

// listPaneContent is the active list, or a helpful empty hint on an empty tasks
// tab (other entities fall back to the list's own "No items."). The chip (status
// view / sort / applied filter) is written into the list's title slot here — a
// pure function of state, idempotent per frame — so it shows above the rows (and
// collapses to nothing in the clean default).
func (m Model) listPaneContent() string {
	t := m.cur()
	t.list.Title = t.chip()
	if t.kind == entityTasks && t.statusView == "" && len(t.list.Items()) == 0 {
		return "No active tasks.\n\nCreate one:\n  tskflwctl task new \"Title\" --epic <id>"
	}
	return t.list.View()
}

// detailPaneView composes the detail pane: a title line + the scrollable body.
func (m Model) detailPaneView() string {
	titleStyle := dimStyle
	if m.focus == focusDetail {
		titleStyle = selectedStyle
	}
	// Truncate the title to the pane's inner width — an un-truncated long slug
	// would wrap to a second row, growing the pane past its budget and clipping
	// its bottom border (the truncate discipline every Join input must follow).
	title := titleStyle.Render(truncate(m.detailTitle(), max1(m.detail.width)))
	content := lipgloss.JoinVertical(lipgloss.Left, title, m.detail.View())
	return m.pane(focusDetail, content, m.detailOuterW)
}

func (m Model) detailTitle() string {
	if m.detail.title == "" {
		return "Detail"
	}
	return m.detail.title
}

// pane wraps content in a focus-colored border. Inner dimensions are clamped to
// ≥1 so a tiny terminal never produces a negative-sized (broken) frame.
func (m Model) pane(f focus, content string, outerW int) string {
	border := paneInactive
	if m.focus == f {
		border = paneActive
	}
	return border.Width(max1(outerW - paneHFrame)).Height(max1(m.paneOuterH - paneVFrame)).Render(content)
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

// tabStrip renders the entity tabs (active accented), collapsing to a single
// `[entity ▾]` chip under ~60 cols.
func (m Model) tabStrip() string {
	if m.width < 60 {
		return truncate(activeTab.Render("["+m.cur().name+" ▾]"), m.width)
	}
	parts := make([]string, len(m.tabs))
	for i, t := range m.tabs {
		if i == m.active {
			parts[i] = activeTab.Render(t.name)
		} else {
			parts[i] = dim(t.name)
		}
	}
	return truncate(strings.Join(parts, dim("  ·  ")), m.width)
}

func (m Model) footer() string {
	if m.cmd.active {
		return truncate(m.cmd.view(), m.width)
	}
	// The detail find input/status takes over the footer while searching a body.
	if m.focus == focusDetail && (m.detail.finding() || m.detail.findActive()) {
		return truncate(m.detail.findStatus(), m.width)
	}
	hints := ": cmd · / filter · o sort · s view · [ ] tabs · l/⏎ detail · ? help · q quit"
	if m.focus == focusDetail {
		hints = ": cmd · / find · n/N match · j/k scroll · g/G top/bottom · h/esc back · q quit"
		if !m.twoPane {
			// Single-pane drill: q pops back to the list (context quit), so the
			// hint must not promise it exits the app.
			hints = ": cmd · / find · n/N match · j/k scroll · g/G top/bottom · h/esc/q back"
		}
	}
	if p := m.cur().problems; len(p) > 0 {
		hints = fmt.Sprintf("! %d unreadable · ", len(p)) + hints
	}
	if m.cur().loaded && m.cur().loadErr != nil {
		hints = "⚠ reload failed · " + hints // the rows shown are the last good load
	}
	if m.watchOff {
		hints = "live-reload off · " + hints
	}
	return dim(truncate(hints, m.width))
}
