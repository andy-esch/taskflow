// Package tui is the second primary adapter: an interactive Bubble Tea front-end
// over the same core.Service the CLI uses. It never touches the store/fs — all
// reads run as tea.Cmds against the service (see commands.go).
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

	err      error
	restore  string // id to re-select after a reload of the active tab
	showHelp bool   // the `?` keybinding overlay is open
}

// New constructs the root model over the same *core.Service the CLI uses.
func New(svc *core.Service, root string) Model {
	return Model{
		svc: svc, root: root, focus: focusList,
		tabs: newEntityTabs(), active: 0,
		detail: newDetailPane(), cmd: newCommandBar(),
	}
}

func (m Model) Init() tea.Cmd { return m.cur().reload(m.svc) }

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
		if !m.isCurrentSelection(msg.kind, msg.id) {
			return m, nil // stale: tab or selection changed since this load fired
		}
		m.detail.SetContent(msg.content)
		return m, nil

	case detailErrMsg:
		// A per-item load failure (e.g. an ambiguous duplicate slug) shows in the
		// detail pane — it must not blank the whole browser.
		if !m.isCurrentSelection(msg.kind, msg.id) {
			return m, nil
		}
		m.detail.SetError(msg.id, msg.err.Error())
		return m, nil

	case reloadMsg:
		m.restore = m.selectedID()
		return m, m.cur().reload(m.svc)

	case errMsg:
		m.err = msg.err
		return m, nil
	}
	// Forward anything else (notably the list's async FilterMatchesMsg, which
	// applies the `/` filter) to the active list.
	var cmd tea.Cmd
	m.cur().list, cmd = m.cur().list.Update(msg)
	return m, cmd
}

// handleListLoaded applies an entity-list load to its tab (by kind, so a load
// that finishes after a tab switch still lands correctly). For the active tab it
// also restores the cursor and kicks off the selected item's detail load.
func (m Model) handleListLoaded(msg listLoadedMsg) (tea.Model, tea.Cmd) {
	i := indexOfKind(m.tabs, msg.kind)
	if i < 0 {
		return m, nil
	}
	// A successful load clears any prior fatal error so a transient failure (e.g.
	// the planning dir briefly unreadable) recovers on the next `r`/reload.
	m.err = nil
	tab := m.tabs[i]
	sortItems(msg.items, tab.sortKey, tab.sortRev) // honor the tab's sort across reloads
	cmd := tab.list.SetItems(msg.items)
	tab.loaded = true
	tab.problems = msg.problems
	if msg.kind != m.cur().kind {
		return m, cmd // a background tab loaded; leave the active view alone
	}
	if m.restore != "" {
		m.selectByID(m.restore)
		m.restore = ""
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

	// 2. The list's filter input owns every key while capturing a query.
	if m.cur().list.SettingFilter() {
		var cmd tea.Cmd
		m.cur().list, cmd = m.cur().list.Update(msg)
		return m, cmd
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
	case key.Matches(msg, keys.ForceQuit), key.Matches(msg, keys.Quit):
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
	prev := m.selectedID()
	var cmd tea.Cmd
	m.cur().list, cmd = m.cur().list.Update(msg)
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
		return m, tea.Batch(cmd, m.cur().loadItem(m.svc, id))
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
	return m.cur().loadItem(m.svc, id)
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

func (m *Model) selectByID(id string) {
	// Range the *visible* items: list.Select indexes the filtered/paginated view
	// (Page*PerPage+cursor), so an unfiltered Items() index would land on the
	// wrong row whenever a `/` filter is active. When unfiltered, VisibleItems ==
	// Items, so this is identical to the naive version.
	for i, it := range m.cur().list.VisibleItems() {
		if ei, ok := it.(entityItem); ok && ei.id() == id {
			m.cur().list.Select(i)
			return
		}
	}
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
	cmd := t.list.SetItems(items)
	m.selectByID(id)
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
	if it, ok := tab.list.SelectedItem().(entityItem); ok {
		m.restore = it.id()
	}
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
	switch {
	case m.width == 0 || m.height == 0:
		// No WindowSizeMsg yet. Rendering panes now would use unset (0) sizes →
		// negative border dimensions → a broken oversized frame. Wait for size.
		return "loading…"
	case m.err != nil:
		return fg(theme.ColorRed, "error: "+m.err.Error())
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

// renderBody is the pane layout: a loading note until the active tab loads, then
// the two-pane (or single-pane drill) view.
func (m Model) renderBody() string {
	if !m.cur().loaded {
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
	}
	if p := m.cur().problems; len(p) > 0 {
		hints = fmt.Sprintf("! %d unreadable · ", len(p)) + hints
	}
	return dim(truncate(hints, m.width))
}
