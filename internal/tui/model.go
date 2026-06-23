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

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/listfilter"
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
	svc *core.Service

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
	modals []modal // the ordered overlay registry (help, action, follow); see overlay.go

	showHelp   bool       // the `?` keybinding overlay is open
	helpScroll int        // overlay scroll offset (j/k while open; clamped to helpMaxScroll)
	action     actionMenu // the `a` lifecycle action menu (S4)
	follow     followMenu // the `f` reference picker (S6, epics → their tasks)
	edit       editMenu   // the `e` inline field editor (task set with a GUI)
	navStack   []navLoc   // where each `f` jump came from; ctrl+o pops (S6)
	flash      string     // transient post-action feedback line (cleared on the next key)
	flashErr   bool       // the flash is an error (rendered red)
	movedAway  string     // slug just relocated by a lifecycle action: its absence from the
	// active tab after the post-move reload is the success, not a dangling reference

	watch     *watcher // fsnotify source (nil when unavailable / in tests); see watch.go
	watchOff  bool     // the watcher failed to start: live reload is off (footer note)
	dirtyGen  int      // bumped per fs event; the debounce tick fires a reload only when it matches
	detailGen int      // bumped per detail request; orders concurrent loads for the same id
}

// New constructs the root model over the same *core.Service the CLI uses.
func New(svc *core.Service) Model {
	return Model{
		svc: svc, focus: focusList,
		tabs: newEntityTabs(), active: 0,
		detail: newDetailPane(theme.MarkdownStyleDark), cmd: newCommandBar(),
		modals: defaultModals(),
	}
}

func (m Model) Init() tea.Cmd {
	if m.watch != nil {
		return tea.Batch(m.cur().reload(m.svc, ""), waitForFS(m.watch))
	}
	return m.cur().reload(m.svc, "")
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
		// markReload picks the restore id (a pending jump target, else the cursor);
		// reload stamps it onto the load it fires.
		cmds = append(cmds, t.reload(m.svc, t.markReload()))
	}
	return tea.Batch(cmds...)
}

// cur returns the active entity tab. The tab is a pointer, so reads use a value
// receiver yet callers can still mutate the tab's list in place.
func (m Model) cur() *entityTab { return m.tabs[m.active] }

// Update is the reducer. It delegates to update, then syncs the active tab's chip
// into its list title — so View can stay a pure function of state (the title slot
// is also where bubbles draws the `/` filter prompt, so the chip can't be a
// separate line without losing it).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.update(msg)
	mm := next.(Model)
	t := mm.cur()
	t.list.Title = t.chip()
	return mm, cmd
}

func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.recomputeLayout()
		return m, nil

	case tea.KeyPressMsg:
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

	case movedMsg:
		// A transition succeeded: flash it and reload so the relocated task shows in
		// its new status (folder-authoritative), each tab's cursor preserved by id.
		m.flash = fmt.Sprintf("moved %s → %s", msg.slug, msg.to)
		m.flashErr = false
		// The moved task leaves the active list (folder-authoritative); its
		// disappearance on the reload below is expected, so don't let the post-reload
		// restore mistake it for a dangling reference and overwrite this success.
		m.movedAway = msg.slug
		return m, m.reloadAll()

	case editedMsg:
		// A field edit succeeded: flash it and reload so the new value shows. The
		// task keeps its dir (SetFields isn't a move), so this is a plain refresh —
		// each tab's cursor preserved by id, no movedAway dance. If the editor is
		// still open (the user can keep editing), refresh the field it just set so
		// it isn't stale.
		m.flash = fmt.Sprintf("set %s on %s", msg.field, msg.slug)
		m.flashErr = false
		if m.edit.active {
			m.edit.applied(msg.field, msg.value) // back to the picker, value refreshed
		}
		return m, m.reloadAll()

	case actionErrMsg:
		// An inline-edit write failed: keep the field open with what was typed and
		// show the validation error there, so the fix happens in place rather than
		// the edit silently reverting. Other mutations (the action menu) flash.
		if m.edit.active {
			// Trim the "validation failed:" sentinel prefix — inline by the field, the
			// bare reason ("at least one tag is required") reads cleaner.
			m.edit.err = strings.TrimPrefix(msg.err.Error(), domain.ErrValidation.Error()+": ")
			return m, nil
		}
		m.flash = msg.err.Error()
		m.flashErr = true
		return m, nil

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
	// Forward anything else (e.g. cursor-blink ticks) to the active list ONLY.
	// INVARIANT: any async message that must reach a *background* tab's list has to
	// carry a tab identity (arrive as tabMsg above, which routes itself) — an
	// untagged list-affecting message would be misdelivered or dropped here. This
	// holds today because background tabs never focus their FilterInput (so generate
	// no blink/spinner ticks); a future background component with its own ticks must
	// be tab-tagged, or this fall-through changed to broadcast (per-tab routeToTab).
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
	// Retry a pending restore now its async refilter may have populated VisibleItems
	// — but only for the gen that set it, so a newer reload's target isn't applied
	// against this (now-superseded) filter pass.
	if tab.restore != "" && tab.restoreGen == tab.loadGen && tab.selectByID(tab.restore) {
		tab.restore, tab.restoreGen = "", 0
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
	// Resolve the cursor restore carried by THIS load (gen-matched above). Reading
	// the per-message id — not a mutable tab slot two triggers share — means a
	// dropped stale load can't apply a restore meant for another (M6). This load
	// supersedes any pending restore; clear it, then re-pend only if still unfound
	// under an async filter.
	tab.restore, tab.restoreGen = "", 0
	if msg.restore != "" && !tab.selectByID(msg.restore) {
		switch {
		case tab.list.FilterState() == list.Unfiltered:
			// The id is genuinely absent from a fully-visible list — a dangling
			// reference jump (`f` to an epic that doesn't exist) or a selection
			// deleted externally. Say so once. EXCEPT the task we just relocated: its
			// absence here is the success (flashed green already), not a not-found.
			if msg.kind == m.cur().kind && msg.restore != m.movedAway {
				m.flash, m.flashErr = msg.restore+" not found", true
			}
		default:
			// Filtered: SetItems' refilter is async (VisibleItems is empty now), so
			// keep the target pending — keyed to this gen — for handleTabMsg to retry
			// when the FilterMatchesMsg lands.
			tab.restore, tab.restoreGen = msg.restore, msg.gen
		}
	}
	if msg.kind != m.cur().kind {
		return m, cmd // a background tab loaded; leave the active view alone
	}
	m.movedAway = "" // consumed: the active tab's post-move reload has landed
	return m, tea.Batch(cmd, m.refreshDetail())
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Any key dismisses the post-action flash (it's a one-shot confirmation).
	m.flash = ""

	// ForceQuit is the one key no layer may swallow: handle it once here, ahead of
	// the modal loop and every input capture, so it isn't re-implemented in each.
	if key.Matches(msg, keys.ForceQuit) {
		return m, tea.Quit
	}

	// Modal overlays (help, action, follow) take precedence in registry order: the
	// first active one owns the key. The markers are stateless and mutate the
	// by-value model copy through &m (the "mutate the copy, return it" idiom), so a
	// new overlay is one entry in defaultModals — no new guard block here. Input
	// captures (the command bar, the list filter, the detail-find below) stay as
	// special early returns: they're text inputs, not floating boxes.
	for _, o := range m.modals {
		if o.active(&m) {
			if handled, cmd := o.handleKey(&m, msg); handled {
				return m, cmd
			}
		}
	}

	// 1. The command bar captures every key while open (so `:tasks` typing never
	// leaks into global hotkeys).
	if m.cmd.active {
		switch {
		case key.Matches(msg, keys.Back):
			m.cmd.blur()
			return m, nil
		case msg.String() == "enter":
			return m.dispatchCommand()
		case msg.String() == "tab":
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

	// 2b. The detail pane's find input owns keys while a query is being typed.
	if m.focus == focusDetail && m.detail.finding() {
		return m, m.detail.updateFind(msg)
	}

	// 3. Global hotkeys.
	switch {
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
	case key.Matches(msg, keys.Action):
		// Lifecycle actions are registry-driven: open the menu for any entity that
		// declares transitions (tasks: statuses; audits: buckets; epics: none).
		if cur := m.cur(); len(cur.transitions) > 0 {
			if id, state, ok := m.selectedLifecycle(); ok {
				m.action.open(id, cur.transitions, state)
			}
		}
		return m, nil
	case key.Matches(msg, keys.Edit):
		// Inline field edit via SetFields — task-only (status stays in the `a`
		// menu); a no-op on epics/audits, which have no SetFields path in core.
		if t, ok := m.selectedTask(); ok {
			m.edit.open(t)
		}
		return m, nil
	case key.Matches(msg, keys.RawToggle):
		m.detail.toggleMode() // raw ⇄ pretty markdown (cached, no recompile)
		return m, nil
	case key.Matches(msg, keys.Follow):
		// f shadows the (undocumented) f-paging alias in list/viewport — d/u and
		// ctrl+d/u remain the documented paging keys.
		return m.followSelected()
	case key.Matches(msg, keys.JumpBack):
		return m.navBack()
	case key.Matches(msg, keys.Yank):
		return m.yank(m.selectedID(), "slug")
	case key.Matches(msg, keys.YankPath):
		return m.yank(m.selectedPath(), "path")
	case key.Matches(msg, keys.Command):
		return m, m.cmd.focus()
	case key.Matches(msg, keys.NextTab):
		return m, m.switchTab((m.active + 1) % len(m.tabs))
	case key.Matches(msg, keys.PrevTab):
		return m, m.switchTab((m.active - 1 + len(m.tabs)) % len(m.tabs))
	// Sort/status-view/filter-mode reshape the LIST, so they're list-scoped: gate
	// them on list focus. Otherwise pressing e.g. `s` while reading the detail pane
	// snaps focus back to the list, clears the detail, and triggers a reload —
	// silently wiping the body being read (the detail-focus footer advertises none
	// of these keys). In detail focus they fall through to the find handler / no-op.
	case m.focus == focusList && key.Matches(msg, keys.Sort):
		return m, m.cycleSort(1)
	case m.focus == focusList && key.Matches(msg, keys.SortRev):
		m.cur().sortRev = !m.cur().sortRev
		return m, m.applySortToCurrent()
	case m.focus == focusList && key.Matches(msg, keys.StatusView):
		return m, m.cycleView(1)
	case m.focus == focusList && key.Matches(msg, keys.StatusRev):
		return m, m.cycleView(-1)
	case m.focus == focusList && key.Matches(msg, keys.FilterMode):
		return m.toggleFilterMode()
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
	if view, i, ok := m.resolveView(word); ok {
		return m, m.applyView(i, view)
	}
	if tr, ok := transitionFor(m.cur().transitions, word); ok {
		id, _, ok := m.selectedLifecycle()
		if !ok {
			cmd := m.cmd.focus()
			m.cmd.err = "select a row first"
			return m, cmd
		}
		if tr.destructive {
			m.action.openConfirm(id, tr) // gate even an explicit :deprecate
			return m, nil
		}
		return m, m.cur().applyMove(m.svc, id, tr)
	}
	cmd := m.cmd.focus()
	m.cmd.err = "unknown: " + word
	return m, cmd
}

// handleActionKey drives the lifecycle action menu while it's open: vim-select a
// transition, Enter applies it (a destructive one gates on y/n), Esc cancels. It
// mutates the model copy directly (the modal loop passes &m) and returns the cmd;
// ForceQuit is handled by handleKey's preamble, ahead of the modal loop.
func (m *Model) handleActionKey(msg tea.KeyPressMsg) tea.Cmd {
	if m.action.confirm {
		switch msg.String() {
		case "y", "Y":
			tr, slug := m.action.selected(), m.action.slug
			m.action.close()
			return m.cur().applyMove(m.svc, slug, tr)
		case "n", "N", "esc":
			if m.action.confirmOnly() {
				m.action.close() // a bare `:deprecate` confirm has no menu to return to
			} else {
				m.action.confirm = false // back to the menu
			}
		}
		return nil
	}
	switch msg.String() {
	case "j", "down":
		m.action.move(1)
	case "k", "up":
		m.action.move(-1)
	case "enter", "l":
		tr := m.action.selected()
		if tr.destructive {
			m.action.confirm = true
			return nil
		}
		slug := m.action.slug
		m.action.close()
		return m.cur().applyMove(m.svc, slug, tr)
	case "esc", "h", "a", "q":
		m.action.close()
	}
	return nil
}

// selectedTask returns the selected row as a task — ok only on the tasks tab. Used
// by followSelected (the task→epic reference jump), which is task-specific.
func (m Model) selectedTask() (domain.Task, bool) {
	if it, ok := m.cur().list.SelectedItem().(taskItem); ok {
		return it.t, true
	}
	return domain.Task{}, false
}

// selectedLifecycle returns the selected row's id and current lifecycle state (a
// task's status or an audit's bucket) for the action menu, or ok=false on an
// entity without a lifecycle (epics) or an empty list. It asks the row via the
// lifecycleItem interface, so the reducer needn't switch on concrete item types.
func (m Model) selectedLifecycle() (id, state string, ok bool) {
	if li, ok := m.cur().list.SelectedItem().(lifecycleItem); ok {
		return li.id(), li.lifecycleState(), true
	}
	return "", "", false
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
		return m.cur().reload(m.svc, "")
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

// selectedPath is the file path of the active tab's selection (empty if none) —
// the clipboard yank target for Y.
func (m Model) selectedPath() string {
	if it, ok := m.cur().list.SelectedItem().(entityItem); ok {
		return it.path()
	}
	return ""
}

// yank copies text to the system clipboard (native OS utility when available,
// else OSC 52 — see copyToClipboard) and flashes a confirmation; an empty target
// (no selection) flashes an error instead. label names what was copied.
func (m Model) yank(text, label string) (tea.Model, tea.Cmd) {
	if text == "" {
		m.flash, m.flashErr = "nothing to copy", true
		return m, nil
	}
	m.flash, m.flashErr = fmt.Sprintf("copied %s: %s", label, text), false
	return m, copyToClipboard(text)
}

func (m Model) entityNames() []string {
	names := make([]string, len(m.tabs))
	for i, t := range m.tabs {
		names[i] = t.name
	}
	return names
}

// commandOptions is the `:` Tab-completion set: every tab's names + aliases +
// view-axis words (so any tab is reachable by name), plus the ACTIVE tab's
// lifecycle verbs. Verbs are context-scoped to match dispatchCommand, which
// resolves them against m.cur().transitions — so completion never offers a verb
// that would be "unknown" on the tab in view. View words are deduped because tasks
// and audits share "deferred"/"all".
func (m Model) commandOptions() []string {
	opts := make([]string, 0, len(m.tabs)*2+len(m.cur().transitions)+12)
	seen := make(map[string]bool)
	add := func(w string) {
		if !seen[w] {
			seen[w] = true
			opts = append(opts, w)
		}
	}
	for _, t := range m.tabs {
		add(t.name)
		for _, a := range t.aliases {
			add(a)
		}
		for _, w := range t.viewWords() {
			add(w)
		}
	}
	for _, tr := range m.cur().transitions {
		add(tr.verb)
	}
	return opts
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

// cycleSort advances the active tab's sort column (wrapping over the columns that
// entity actually offers) and re-sorts.
func (m *Model) cycleSort(dir int) tea.Cmd {
	cols := m.cur().sortCols
	cur := 0
	for i, k := range cols {
		if k == m.cur().sortKey {
			cur = i
			break
		}
	}
	n := len(cols)
	m.cur().sortKey = cols[((cur+dir)%n+n)%n]
	return m.applySortToCurrent()
}

// --- status / bucket views ---

// cycleView steps the active tab's view axis (no-op on entities without one, e.g.
// epics). The cycle order lives in the tab's viewAxis (statusview.go).
func (m *Model) cycleView(dir int) tea.Cmd {
	t := m.cur()
	if len(t.viewAxis) == 0 {
		return nil
	}
	return m.applyView(m.active, viewStep(t.viewAxis, t.statusView, dir))
}

// toggleFilterMode flips the list filter between fuzzy (the default) and substring
// — session-wide across every tab, so the choice is consistent no matter which
// tab you filter in — and re-runs the visible filter so results update live. The
// substring matcher is the shared listfilter.Substring the CLI picker uses, so the
// two faces can't drift. Default stays fuzzy (the TUI's exploratory default).
func (m Model) toggleFilterMode() (tea.Model, tea.Cmd) {
	exact := !m.cur().filterExact
	var f list.FilterFunc = list.DefaultFilter
	if exact {
		f = listfilter.Substring
	}
	prev := m.selectedID()
	for _, t := range m.tabs {
		t.filterExact = exact
		t.list.Filter = f
		t.list.FilterInput.Prompt = filterPrompt(exact)
		// Re-rank EVERY tab that has a filter applied (not just the visible one), so
		// a background tab's rankings can't go stale under the new mode's chip.
		if t.list.FilterState() != list.Unfiltered {
			t.list.SetFilterText(t.list.FilterValue())
		}
	}
	// The current tab's re-rank can move or empty the selection — keep the detail
	// pane in sync, like every other selection-moving path.
	return m.afterSelectionChange(prev, nil)
}

// resolveView maps a `:` word to a view value and the tab it applies to. The
// active tab is tried first, so words shared across axes (":all"/":deferred", on
// both tasks and audits) act in context; otherwise the first tab in the registry
// that defines the word wins (tasks before audits — back-compat).
func (m Model) resolveView(word string) (view string, tab int, ok bool) {
	if v, ok := m.cur().viewFor(word); ok {
		return v, m.active, true
	}
	for i, t := range m.tabs {
		if v, ok := t.viewFor(word); ok {
			return v, i, true
		}
	}
	return "", -1, false
}

// applyView switches to tab i, sets its view filter, and reloads — preserving the
// cursor by id when the item survives into the new view.
func (m *Model) applyView(i int, view string) tea.Cmd {
	m.active = i
	m.focus = focusList
	tab := m.tabs[i]
	restoreID := tab.markReload() // preserve the cursor across the view change
	// Reset the active filter when switching views, matching jumpTo: otherwise a
	// stale `/foo` silently carries into the new status view (the chip still reads
	// filter:foo and the view can look unexpectedly empty).
	tab.list.ResetFilter()
	tab.statusView = view
	m.detail.clear()
	return tab.reload(m.svc, restoreID)
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

func (m Model) View() tea.View {
	// Alt-screen is declarative in v2 (a View field, not a program option) — set it
	// on every view so the browser owns the full screen and restores cleanly on exit.
	if m.width == 0 || m.height == 0 {
		// No WindowSizeMsg yet. Rendering panes now would use unset (0) sizes →
		// negative border dimensions → a broken oversized frame. Wait for size.
		v := tea.NewView("loading…")
		v.AltScreen = true
		return v
	}
	// Hard-clamp the body to its budget so the tab strip and footer (the chrome)
	// are ALWAYS rendered — a child that overflows its box loses its own bottom
	// edge, never the load-bearing navigation/command line. Belt-and-suspenders
	// with the per-list pagination reserve above.
	body := lipgloss.NewStyle().MaxHeight(m.paneOuterH).Render(m.bodyView())
	full := lipgloss.JoinVertical(lipgloss.Left, m.tabStrip(), body, m.footer())
	v := tea.NewView(lipgloss.NewStyle().MaxWidth(m.width).MaxHeight(m.height).Render(full))
	v.AltScreen = true
	return v
}

// bodyView renders the pane area, floating the topmost active modal (help/action/
// follow) over it — the underlying panes stay visible around the box. The modal
// registry is looped in precedence order, so the first active one wins (matching
// the old switch order) and a new overlay needs no case added here.
func (m Model) bodyView() string {
	base := m.renderBody()
	for _, o := range m.modals {
		if o.active(&m) {
			// Normalize the body to exact dimensions, then composite the box on top
			// so it floats over the items rather than blanking them.
			canvas := lipgloss.Place(m.width, m.paneOuterH, lipgloss.Left, lipgloss.Top, base)
			return overlay(canvas, o.view(&m, m.width-2, m.paneOuterH-2), m.width, m.paneOuterH)
		}
	}
	return base
}

// helpMaxScroll is the largest in-bounds scroll offset for the `?` overlay,
// mirroring helpBox's window math: the box gets paneOuterH-2 (see bodyView) and
// spends 2 of that on its border rows. The j/k handler clamps to this so
// helpScroll can't run past the visible bottom (leaving k presses doing nothing).
func (m Model) helpMaxScroll() int {
	innerH := m.paneOuterH - 2 - 2 // box height (paneOuterH-2) minus top+bottom border
	if innerH <= 0 {
		return 0
	}
	return max(len(helpLines())-innerH, 0)
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
// or audits tab (epics fall back to the list's own "No items."). The audits hint
// names the empty bucket and points at the s/S cycle, since archived buckets are
// reachable in-TUI now rather than only via `audit list --all`.
func (m Model) listPaneContent() string {
	t := m.cur()
	// The chip is synced into t.list.Title by Update (not here) so View is pure.
	if t.kind == entityTasks && t.statusView == "" && len(t.list.Items()) == 0 {
		return "No active tasks.\n\nCreate one:\n  tskflwctl task new \"Title\" --epic <id>"
	}
	if t.kind == entityAudits && len(t.list.Items()) == 0 {
		switch t.statusView {
		case "":
			return "No open audits.\n\nOther buckets: s/S or :closed / :deferred / :all"
		case "all":
			return "No audits in any bucket."
		default:
			return fmt.Sprintf("No %s audits.\n\nOther buckets: s/S or :open / :closed / :deferred / :all", t.statusView)
		}
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
	t := m.detail.title
	if t == "" {
		return "Detail"
	}
	// Flag raw mode in the title (pretty is the default, left unlabeled).
	if m.detail.hasContent && !m.detail.pretty {
		t += " · raw"
	}
	return t
}

// pane wraps content in a focus-colored border. lipgloss v2 sizes by border-box
// (Width/Height are the OUTER size; the border is subtracted internally to get the
// content area), so we pass the outer dims directly — content lands in
// outerW-paneHFrame × paneOuterH-paneVFrame, the inner budget children are sized
// to. Clamped to ≥1 so a tiny terminal never produces a negative-sized frame.
func (m Model) pane(f focus, content string, outerW int) string {
	border := paneInactive
	if m.focus == f {
		border = paneActive
	}
	return border.Width(max1(outerW)).Height(max1(m.paneOuterH)).Render(content)
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
	// A post-action result takes over the footer until the next key.
	if m.flash != "" {
		if m.flashErr {
			return truncate(fg(theme.ColorRed, "✘ "+m.flash), m.width)
		}
		return truncate(fg(theme.ColorGreen, "✔ "+m.flash), m.width)
	}
	// The detail find input/status takes over the footer while searching a body.
	if m.focus == focusDetail && (m.detail.finding() || m.detail.findActive()) {
		return truncate(m.detail.findStatus(), m.width)
	}
	hints := ": cmd · / filter · a act · e edit · s view · [ ] tabs · l/⏎ detail · ? help · q quit"
	if m.focus == focusDetail {
		hints = ": cmd · / find · n/N match · R raw/pretty · j/k scroll · g/G top/bottom · h/esc back · q quit"
		if !m.twoPane {
			// Single-pane drill: q pops back to the list (context quit), so the
			// hint must not promise it exits the app.
			hints = ": cmd · / find · n/N match · R raw/pretty · j/k scroll · g/G top/bottom · h/esc/q back"
		}
	}
	if n := len(m.navStack); n > 0 {
		// Breadcrumb for the follow stack: where ctrl+o leads, and how deep.
		hints = fmt.Sprintf("↩ ctrl+o %s (%d) · ", m.navStack[n-1].id, n) + hints
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
