// Package tui is the second primary adapter: an interactive Bubble Tea front-end
// over the same core.Service the CLI uses. It never touches the store/fs — every
// read runs as a tea.Cmd against the service (commands.go), so Update and View
// stay I/O-free. Entities (tasks/epics/audits) are declared in a registry
// (entity.go); the lists live-reload via fsnotify (watch.go). See
// docs/ARCHITECTURE.md for the subsystem map.
package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/editor"
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
	zoom          bool // full-screen the detail pane (z): hide the list, give detail the full width
	listOuterW    int
	detailOuterW  int
	paneOuterH    int

	focus   focus
	tabs    []*entityTab
	active  int
	onDash  bool      // showing the landing dashboard (a non-list screen left of the tabs)
	dash    dashboard // the landing screen's widgets (see dashboard.go)
	detail  detailPane
	cmd     commandBar
	palette palette // the ctrl+p command palette (fuzzy launcher); see palette.go
	modals  []modal // the ordered overlay registry (see overlay.go / defaultModals)

	showHelp   bool       // the `?` keybinding overlay is open
	helpScroll int        // overlay scroll offset (j/k while open; clamped to helpMaxScroll)
	action     actionMenu // the `m` lifecycle action menu (S4)
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
		onDash: true, // the dashboard is the landing view; `]`/:tasks drops into work
		detail: newDetailPane(theme.MarkdownStyleDark), cmd: newCommandBar(),
		palette: newPalette(),
		modals:  defaultModals(),
	}
}

func (m Model) Init() tea.Cmd {
	// Load only the landing surface — the dashboard's summary or the active tab.
	// Don't fire the tab loader when the dashboard is the landing: its result is
	// discarded but the call still churns the tab's load generation / restore state.
	load := loadDashboard(m.svc)
	if !m.onDash {
		load = m.cur().reload(m.svc, "")
	}
	if m.watch != nil {
		return tea.Batch(load, waitForFS(m.watch))
	}
	return load
}

// reloadAll re-fires the loader for every loaded tab, each preserving its own
// cursor by id. Unvisited tabs are left alone (they reload fresh on first visit)
// — except the active tab, which always reloads: after a failed *initial* load
// nothing is `loaded`, and `r` must still be able to recover the session.
// This is the `r` / fsnotify path: a change from another process is reflected on
// whichever surface — a tab or the dashboard — you've landed on, not just the
// active tab.
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
	// The dashboard reads the same files, so keep the landing screen live too —
	// otherwise an fsnotify/`r` reload refreshes the (hidden) tabs while the summary
	// you're actually looking at silently goes stale. Loaded eagerly in Init.
	if m.dash.loaded {
		cmds = append(cmds, loadDashboard(m.svc))
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

	case dashLoadedMsg:
		// Store the error durably (like a tab's loadErr) rather than flashing it: a
		// flash clears on the next key, which would leave a never-loaded dashboard
		// stuck on "loading…" with no clue why. renderBody/footer surface it instead —
		// a failed *refresh* keeps the last good rows, a failed first load shows the
		// error pane.
		if msg.err != nil {
			m.dash.loadErr = msg.err
			return m, nil
		}
		m.dash.loadErr = nil
		m.dash.setSummary(msg.summary)
		return m, nil

	case movedMsg:
		// A transition succeeded: flash it and reload so the relocated task shows in
		// its new status (folder-authoritative), each tab's cursor preserved by id.
		m.flash = fmt.Sprintf("moved %s → %s", msg.slug, msg.to)
		if msg.revisit != "" {
			m.flash += fmt.Sprintf(" (revisit %s)", msg.revisit)
		}
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

	case editorClosedMsg:
		// The external $EDITOR (`E`) exited. A launch failure flashes; otherwise the
		// editor may have changed the file, so reload — each tab's cursor preserved by
		// id. status==dir, so a frontmatter `status:` edit can't move the file; it'll
		// show as misfiled on reload, exactly as the CLI's `task edit` flags it.
		if msg.err != nil {
			m.flash, m.flashErr = "editor: "+msg.err.Error(), true
			return m, nil
		}
		return m, m.reloadAll()

	case actionErrMsg:
		// A mutation failed. Route on the domain error class (shared with the CLI's
		// exit-code mapping, audit H4) rather than painting every failure red:
		//   - Conflict: the compare-and-swap saw the file change on disk under us
		//     (another process moved/edited it). Retrying our stale write is wrong —
		//     reload so the user acts on the current state. A stale inline editor is
		//     closed first, since its prefilled values may no longer match the file.
		//   - Validation (while editing): keep the field open with what was typed and
		//     show the bare reason inline, so the fix happens in place. domain.Reason
		//     strips the "validation failed:" prefix in one tested helper (no inline
		//     string-trim coupling the TUI to the wrap format).
		//   - Everything else: a red flash.
		switch domain.Classify(msg.err) {
		case domain.ClassConflict:
			if m.edit.active {
				m.edit.close()
			}
			m.flash, m.flashErr = "changed on disk — reloading", true
			return m, m.reloadAll()
		case domain.ClassValidation:
			if m.edit.active {
				m.edit.err = domain.Reason(msg.err) // bare reason, inline by the field
				return m, nil
			}
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
	// Guarded executably by TestModel_UntaggedMsgRoutesToActiveTabOnly (audit M8).
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
	if m.palette.active {
		m.palette.reindex(m.paletteIndex()) // a tab finished loading while the palette is open
	}
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

	// Modal overlays take precedence in registry order (see defaultModals): the
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

	// 2c. The dashboard (landing screen) owns its own keys — cursor + jump-to-entity
	// plus a handful of globals. It's not a list, so the list-scoped hotkeys below
	// (m/e/s/o/F/f/…) don't apply; routing here keeps them from acting on the hidden
	// active tab's selection.
	if m.onDash {
		return m.handleDashKey(msg)
	}

	// 3. Global hotkeys.
	switch {
	case key.Matches(msg, keys.Quit):
		// q is a *context* quit: full-screen detail and the single-pane drill are
		// both layers, so q pops back (to the split / the list) rather than exiting.
		// In two-pane, detail focus isn't a layer — q quits from either pane.
		if m.zoom {
			m.toggleZoom()
			return m, nil
		}
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
		// declares transitions (tasks: statuses; audits: buckets; epics: statuses —
		// epicTransitions/moveEpic, which rewrite the status: field in place).
		if cur := m.cur(); len(cur.transitions) > 0 {
			if id, state, ok := m.selectedLifecycle(); ok {
				m.action.open(id, cur.transitions, state)
			}
		}
		return m, nil
	case key.Matches(msg, keys.Edit):
		// Inline field edit — tasks (SetFields) and epics (SetEpicFields), each with
		// its own typed field set: tasks get description/priority/tags/effort/tier,
		// epics description/priority/tags (no effort/tier; status moves in the `m`
		// menu). Audits have no field-level write — they edit the whole file via the
		// entity-agnostic `E` ($EDITOR) — so on an audit selection we point at `E`
		// rather than dying as a silent no-op.
		if t, ok := m.selectedTask(); ok {
			m.edit.open(t)
		} else if ep, ok := m.selectedEpic(); ok {
			m.edit.openEpic(ep)
		} else if m.selectedPath() != "" {
			m.flash, m.flashErr = "no inline edit here — press E to edit in $EDITOR", true
		}
		return m, nil
	case key.Matches(msg, keys.OpenEditor):
		return m.openInEditor()
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
	case key.Matches(msg, keys.Palette):
		return m, m.openPalette()
	case key.Matches(msg, keys.NextTab):
		if m.active == len(m.tabs)-1 {
			return m, m.enterDash() // past the last tab wraps to the dashboard
		}
		return m, m.switchTab(m.active + 1)
	case key.Matches(msg, keys.PrevTab):
		if m.active == 0 {
			return m, m.enterDash() // before the first tab wraps to the dashboard
		}
		return m, m.switchTab(m.active - 1)
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
	case key.Matches(msg, keys.Zoom):
		// Full-screen the detail pane (toggle). Entity-tab only — the dashboard
		// routes its keys in handleDashKey above and never reaches here.
		m.toggleZoom()
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
		// First Esc/h clears an active find; then it leaves full-screen (back to the
		// split) or, in the split, returns focus to the list.
		if m.detail.findActive() {
			m.detail.clearFind()
			return m, nil
		}
		if m.zoom {
			m.toggleZoom() // exit full-screen back to the split (focus → list)
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

// handleActionKey drives the lifecycle action menu while it's open: vim-select a
// transition, Enter applies it (a destructive one gates on y/n), Esc cancels. It
// mutates the model copy directly (the modal loop passes &m) and returns the cmd;
// ForceQuit is handled by handleKey's preamble, ahead of the modal loop.
func (m *Model) handleActionKey(msg tea.KeyPressMsg) tea.Cmd {
	if m.action.revisit {
		switch msg.String() {
		case "enter":
			date, err := domain.ParseRevisitDate(m.action.dateInput.Value(), m.svc.Now())
			if err != nil {
				m.action.dateErr = err.Error() // keep what was typed, show the error
				return nil
			}
			slug := m.action.slug
			m.action.close()
			return deferTaskCmd(m.svc, slug, date)
		case "esc":
			if len(m.action.options) > 0 {
				m.action.revisit = false // came from the menu → return to it
				m.action.dateInput.Blur()
			} else {
				m.action.close() // cold `:defer`/palette entry → nothing to return to
			}
			return nil
		}
		m.action.dateErr = "" // any keystroke clears the stale parse error
		var cmd tea.Cmd
		m.action.dateInput, cmd = m.action.dateInput.Update(msg)
		return cmd
	}
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
		// A task defer opens the revisit-date prompt instead of applying at once
		// (the audit "defer" bucket has no revisit date, so it falls through).
		if m.cur().kind == entityTasks && tr.to == string(domain.StatusDeferred) {
			return m.action.beginRevisit(m.action.slug)
		}
		slug := m.action.slug
		m.action.close()
		return m.cur().applyMove(m.svc, slug, tr)
	case "esc", "h", "a", "q":
		m.action.close()
	}
	return nil
}

// beginTransition routes a chosen lifecycle transition for slug: a destructive
// move opens the y/n confirm; the tasks-tab defer opens the revisit-date prompt
// (mirroring the CLI snooze); everything else applies immediately. Shared by the
// `:`-command and palette entry points so all three paths agree.
func (m *Model) beginTransition(id string, tr transition) tea.Cmd {
	if tr.destructive {
		m.action.openConfirm(id, tr)
		return nil
	}
	if m.cur().kind == entityTasks && tr.to == string(domain.StatusDeferred) {
		return m.action.beginRevisit(id)
	}
	return m.cur().applyMove(m.svc, id, tr)
}

// selectedTask returns the selected row as a task — ok only on the tasks tab. Used
// by followSelected (the task→epic reference jump), which is task-specific.
func (m Model) selectedTask() (domain.Task, bool) {
	if it, ok := m.cur().list.SelectedItem().(taskItem); ok {
		return it.t, true
	}
	return domain.Task{}, false
}

// selectedEpic returns the selected row as an epic — ok only on the epics tab.
// Mirrors selectedTask; the `e` handler uses it to open the inline editor on an
// epic (description/priority/tags via SetEpicFields).
func (m Model) selectedEpic() (domain.Epic, bool) {
	if it, ok := m.cur().list.SelectedItem().(epicItem); ok {
		return it.es.Epic, true
	}
	return domain.Epic{}, false
}

// selectedLifecycle returns the selected row's id and current lifecycle state (a
// task's status, an audit's bucket, or an epic's status) for the action menu, or
// ok=false on an empty list. Every entity now implements lifecycleState() (epics
// via moveEpic, which rewrites the status: field in place rather than moving the
// file), so it asks the row via the lifecycleItem interface, never switching on
// concrete item types.
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
	if i == m.active && !m.onDash {
		return nil // already here (but leaving the dashboard to the active tab still re-renders)
	}
	m.exitDashboard(i)
	if !m.cur().loaded {
		return m.cur().reload(m.svc, "")
	}
	return m.refreshDetail()
}

// exitDashboard clears the landing-screen state and makes tab i active — the
// shared half of every dashboard→tab transition (switchTab / leaveDashTo /
// applyView / jumpTo all need it). Callers still own how tab i (re)loads: a
// first-visit load, a view filter, or a cursor restore.
func (m *Model) exitDashboard(i int) {
	m.onDash = false
	m.active = i
	m.focus = focusList
	m.detail.clear()
	m.unzoom() // a tab switch drops full-screen — the new tab opens on its list
}

// unzoom leaves full-screen detail and restores the layout, if zoomed. A no-op
// otherwise. Used where the item/tab context changes out from under the zoom
// (switching tabs, entering the dashboard) so it never strands a full-screen pane
// over a just-cleared selection.
func (m *Model) unzoom() {
	if m.zoom {
		m.zoom = false
		m.recomputeLayout()
	}
}

// handleDashKey drives the landing dashboard: j/k move the cursor over the
// navigable rows, ⏎/l jumps to the selected item or view, [ / ] move between the
// dashboard and the tabs, and the usual globals (:, ctrl+p, ?, r, q) work. The
// list-scoped hotkeys (m/e/s/o/F/f/…) don't apply — the dashboard isn't a list.
func (m Model) handleDashKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "j" || msg.String() == "down":
		m.dash.move(1)
	case msg.String() == "k" || msg.String() == "up":
		m.dash.move(-1)
	case msg.String() == "enter" || msg.String() == "l":
		if tgt, ok := m.dash.selectedTarget(); ok {
			return m, m.dashJump(tgt)
		}
	case key.Matches(msg, keys.NextTab):
		return m, m.leaveDashTo(0) // ] → first tab
	case key.Matches(msg, keys.PrevTab):
		return m, m.leaveDashTo(len(m.tabs) - 1) // [ → last tab
	case key.Matches(msg, keys.Command):
		return m, m.cmd.focus()
	case key.Matches(msg, keys.Palette):
		return m, m.openPalette()
	case key.Matches(msg, keys.Help):
		m.showHelp = true
	case key.Matches(msg, keys.Refresh):
		return m, loadDashboard(m.svc)
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

// enterDash switches to the landing dashboard and refreshes its summary.
func (m *Model) enterDash() tea.Cmd {
	m.onDash = true
	m.focus = focusList
	m.detail.clear()
	m.unzoom() // the dashboard has no detail pane to full-screen
	return loadDashboard(m.svc)
}

// leaveDashTo drops from the dashboard onto tab i (loading it if needed). Unlike
// switchTab it doesn't early-return on i == m.active — coming off the dashboard the
// body changes even when the same tab was "active" underneath.
func (m *Model) leaveDashTo(i int) tea.Cmd {
	m.exitDashboard(i)
	if !m.cur().loaded {
		return m.cur().reload(m.svc, "")
	}
	return m.refreshDetail()
}

// dashJump leaves the dashboard for the selected row's target: a specific item
// (jumpTo) or a whole view (applyView) on its entity's tab. Pure routing — jumpTo
// and applyView each own the full dashboard→tab transition (see exitDashboard),
// so this holds no half-set state of its own.
func (m *Model) dashJump(tgt dashTarget) tea.Cmd {
	if tgt.id != "" {
		return m.jumpTo(tgt.kind, tgt.id)
	}
	i := indexOfKind(m.tabs, tgt.kind)
	if i < 0 {
		return nil
	}
	return m.applyView(i, tgt.view)
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
	if m.zoom { // only the detail pane is visible; tab returns to the split
		m.toggleZoom()
		return
	}
	if m.focus == focusList {
		m.setFocus(focusDetail)
	} else {
		m.setFocus(focusList)
	}
}

// toggleZoom flips the detail pane between its split share and full-screen. Zoomed,
// the list is hidden and the detail takes the whole width (recomputeLayout reads
// m.zoom); focus follows the visible pane — detail on the way in, list on the way
// back to the split. Entity-tab only: the dashboard handles its own keys and never
// reaches here.
func (m *Model) toggleZoom() {
	m.zoom = !m.zoom
	if m.zoom {
		m.setFocus(focusDetail)
	} else {
		m.setFocus(focusList)
	}
	m.recomputeLayout()
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

// openInEditor suspends the TUI and opens the current selection's file in the
// user's $EDITOR: tea.ExecProcess releases the terminal for the editor, then
// restores it. It edits the file directly rather than through core — the same
// path as editing it in another terminal, which the fsnotify watcher already
// handles — and reloads on return (via editorClosedMsg) so the change shows at
// once, instant even when the watcher is off (the debounce coalesces the
// duplicate fs event). It works on any entity (tasks/epics/audits) and from
// either pane, since it acts on the selected row's path.
func (m Model) openInEditor() (tea.Model, tea.Cmd) {
	path := m.selectedPath()
	if path == "" {
		m.flash, m.flashErr = "nothing to edit", true
		return m, nil
	}
	cmd := editor.Command(editor.Resolve(), path)
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorClosedMsg{err: err}
	})
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
	m.exitDashboard(i)
	tab := m.tabs[i]
	restoreID := tab.markReload() // preserve the cursor across the view change
	// Reset the active filter when switching views, matching jumpTo: otherwise a
	// stale `/foo` silently carries into the new status view (the chip still reads
	// filter:foo and the view can look unexpectedly empty).
	tab.list.ResetFilter()
	tab.statusView = view
	return tab.reload(m.svc, restoreID)
}
