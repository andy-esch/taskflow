package tui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/theme"
)

// atlasScreen is which question the atlas is answering. The two want opposite layouts —
// "where can I go" is spatial and comparative, "what am I working on" is temporal and
// sorted — so they are separate views rather than two bands fighting over one viewport.
// Keeping them separate is also what lets the work view exist before the spaces view is
// re-laid out as tiles.
type atlasScreen uint8

const (
	atlasScreenSpaces atlasScreen = iota
	atlasScreenWork
)

func (v atlasScreen) label() string {
	if v == atlasScreenWork {
		return "work"
	}
	return "spaces"
}

func (v atlasScreen) next(delta int) atlasScreen {
	const n = 2
	return atlasScreen((int(v) + delta%n + n) % n)
}

// atlas is the cross-space navigation screen. It owns only the registry projection and
// cursor state; opened planning data lives in the active/cached spaceSession.
type atlas struct {
	// screen is which view is showing; it survives an atlas round trip inside a session
	// but not a fresh launch, so `ui` always opens on spaces.
	screen atlasScreen
	// work is the cross-space in-progress set. It is the projection SpaceOverviewService
	// already builds for `status --all`; the atlas kept only its length until now.
	work       []core.SpaceInProgress
	workCursor int
	workOrder  atlasWorkOrder
	loaded     bool
	// stale marks the projection as out of date because planning data changed underneath
	// it. Kept separate from loaded: the atlas keeps rendering its last good snapshot,
	// it just knows to re-read before showing it again.
	stale   bool
	loadErr error
	spaces  []atlasSpace
	cursor  int
	// startup marks the load that ran because the atlas IS the landing screen. Its only
	// job now is degradation: if the registry cannot be read at all, fall back to the
	// seeded workspace rather than opening onto an empty screen.
	startup bool
	loadGen int
	opening bool
	openGen int
	openErr string
	order   atlasOrder
	reverse bool
}

// atlasWorkOrder is the work view's own ordering axis. The spaces view sorts by identity
// (name, activity, registration); the work list sorts by the shape of the work itself, so
// it needs a separate vocabulary rather than borrowing atlasOrder.
type atlasWorkOrder uint8

const (
	// Started first: "what have I been in the middle of longest" is the question the view
	// exists to answer, and it is the one a flat list answers best.
	atlasWorkByStarted atlasWorkOrder = iota
	atlasWorkBySpace
	atlasWorkByPriority
)

func (o atlasWorkOrder) label() string {
	switch o {
	case atlasWorkBySpace:
		return "space"
	case atlasWorkByPriority:
		return "priority"
	default:
		return "started"
	}
}

// grouped reports whether this order draws per-space headings. Only the space order does:
// grouping by anything else would repeat the heading for one row at a time.
func (o atlasWorkOrder) grouped() bool { return o == atlasWorkBySpace }

type atlasSpace struct {
	summary core.SpaceSummary
	entry   int
	// registry retains source order even while the visible cards use another order,
	// making registration order a reversible user choice rather than a one-way sort.
	registry int
}

type atlasOrder uint8

const (
	atlasOrderName atlasOrder = iota
	atlasOrderActivity
	atlasOrderRegistry
)

func (o atlasOrder) label() string {
	switch o {
	case atlasOrderActivity:
		return "activity"
	case atlasOrderRegistry:
		return "registered"
	default:
		return "name"
	}
}

type atlasLoadedMsg struct {
	gen      int
	overview core.SpaceOverview
	err      error
}

func loadAtlas(svc *core.SpaceOverviewService, gen int) tea.Cmd {
	return func() tea.Msg {
		overview, err := svc.Overview()
		return atlasLoadedMsg{gen: gen, overview: overview, err: err}
	}
}

type workspaceOpenedMsg struct {
	gen       int
	workspace core.Workspace
	watcher   *watcher
	watchErr  error
	err       error
}

func openWorkspace(svc *core.WorkspaceService, request core.WorkspaceRequest, gen int) tea.Cmd {
	return func() tea.Msg {
		workspace, err := svc.Open(request)
		if err != nil {
			return workspaceOpenedMsg{gen: gen, err: err}
		}
		w, watchErr := newWatcher(workspace.Layout.WatchPaths())
		return workspaceOpenedMsg{gen: gen, workspace: workspace, watcher: w, watchErr: watchErr}
	}
}

func (m Model) handleAtlasLoaded(msg atlasLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.atlas.loadGen {
		return m, nil
	}
	startup := m.atlas.startup
	m.atlas.startup = false
	if msg.err != nil {
		m.atlas.loadErr = msg.err
		if startup {
			m.onAtlas = false
			m.flash, m.flashErr = "atlas unavailable: "+msg.err.Error(), true
		}
		return m, nil
	}
	m.atlas.setOverview(msg.overview)
	return m, nil
}

func (m Model) handleWorkspaceOpened(msg workspaceOpenedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.atlas.openGen {
		return m, closeWatcher(msg.watcher)
	}
	m.atlas.opening = false
	if msg.err != nil {
		m.atlas.openErr = msg.err.Error()
		m.pendingJump = pendingJump{} // the tree it was meant for never opened
		return m, nil
	}
	return m, m.activateWorkspace(msg.workspace, msg.watcher, msg.watchErr)
}

func (m Model) handleAtlasKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	work := m.atlas.screen == atlasScreenWork
	switch {
	// `[`/`]` walk the whole page strip, so the atlas is a member of that ring rather than
	// a modal side-trip: `]` continues to the overview, `[` wraps back to the last tab.
	// Cycling the atlas's OWN views is `v`, a different axis and so a different key.
	case key.Matches(msg, keys.NextTab):
		return m, m.enterDash()
	case key.Matches(msg, keys.PrevTab):
		return m, m.switchTab(len(m.tabs) - 1)
	case key.Matches(msg, keys.View):
		m.atlas.screen = m.atlas.screen.next(1)
		m.atlas.openErr = ""
	case msg.String() == "j" || msg.String() == "down":
		if work {
			m.atlas.moveWork(1)
		} else {
			m.atlas.move(1)
		}
	case msg.String() == "k" || msg.String() == "up":
		if work {
			m.atlas.moveWork(-1)
		} else {
			m.atlas.move(-1)
		}
	// h/l choose among a space's entry points, which only the spaces view has.
	case !work && (msg.String() == "h" || msg.String() == "left"):
		m.atlas.moveEntry(-1)
	case !work && (msg.String() == "l" || msg.String() == "right"):
		m.atlas.moveEntry(1)
	case key.Matches(msg, keys.Sort):
		if work {
			m.atlas.cycleWorkOrder(1)
		} else {
			m.atlas.cycleOrder()
		}
	case key.Matches(msg, keys.SortRev):
		m.atlas.reverse = !m.atlas.reverse
		if work {
			m.atlas.applyWorkOrder()
		} else {
			m.atlas.applyOrder()
		}
	case msg.String() == "enter":
		if work {
			return m, m.openAtlasWork()
		}
		return m, m.openAtlasSelection()
	case key.Matches(msg, keys.Atlas), key.Matches(msg, keys.Back):
		m.exitAtlas()
	case key.Matches(msg, keys.Command):
		return m, m.cmd.focus()
	case key.Matches(msg, keys.Palette):
		return m, m.openPalette()
	case key.Matches(msg, keys.Help):
		m.showHelp = true
	case key.Matches(msg, keys.Refresh):
		return m, m.enterAtlas(true)
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

// atlasResume is the browsing state entering the atlas has to clobber: it is a
// full-width registry, so it needs list focus and an un-zoomed body. Returning from it is
// not navigation to a new surface — the space is still open, right where it was — so the
// snapshot is put back. A switch discards it; the incoming space restores its own.
type atlasResume struct {
	focus focus
	zoom  bool
	set   bool
}

func (m *Model) enterAtlas(refresh bool) tea.Cmd {
	if m.spaceOverviewSvc == nil {
		m.flash, m.flashErr = "atlas is unavailable in this session", true
		return nil
	}
	if !m.onAtlas {
		m.atlasResume = atlasResume{focus: m.focus, zoom: m.zoom, set: true}
	}
	m.onAtlas = true
	m.focus = focusList
	m.unzoom()
	if refresh || !m.atlas.loaded || m.atlas.stale {
		m.atlas.loadGen++
		return loadAtlas(m.spaceOverviewSvc, m.atlas.loadGen)
	}
	return nil
}

// markAtlasStale records that planning data changed under the active space, so the
// cross-space projection no longer reflects the tree. It is deliberately NOT refreshed
// eagerly: Overview() re-reads EVERY registered planning tree, which is far too heavy to
// run on each fsnotify debounce. The next entry re-reads instead — unless the atlas is
// already on screen, where a visibly wrong list is worse than one extra read.
//
// This covers changes made inside the TUI as well as outside it: the watcher sees the
// tool's own writes, so a `task start` here reaches the same path an edit in another
// terminal does.
func (m *Model) markAtlasStale() tea.Cmd {
	if m.spaceOverviewSvc == nil || !m.atlas.loaded {
		return nil
	}
	m.atlas.stale = true
	if !m.onAtlas {
		return nil // the next enterAtlas will re-read
	}
	m.atlas.loadGen++
	return loadAtlas(m.spaceOverviewSvc, m.atlas.loadGen)
}

// exitAtlas returns to the space that was already open, restoring what enterAtlas took.
func (m *Model) exitAtlas() {
	m.onAtlas = false
	if !m.atlasResume.set {
		return
	}
	m.focus = m.atlasResume.focus
	if m.atlasResume.zoom != m.zoom {
		m.zoom = m.atlasResume.zoom
		m.recomputeLayout()
	}
	m.atlasResume = atlasResume{}
}

func (m *Model) openAtlasSelection() tea.Cmd {
	// Any new open supersedes a landing intent armed by a previous one. Without this, an
	// abandoned work-view selection would fire against whichever space is opened next —
	// and because slugs are only unique within a tree, it can silently select a DIFFERENT
	// task that happens to share the slug rather than failing visibly.
	m.pendingJump = pendingJump{}
	if m.workspaceSvc == nil {
		m.atlas.openErr = "workspace opening is unavailable"
		return nil
	}
	entry, ok := m.atlas.selectedEntry()
	if !ok {
		m.atlas.openErr = "this space has no registered entry point"
		return nil
	}
	if !entry.Healthy() {
		m.atlas.openErr = entry.Detail
		if m.atlas.openErr == "" {
			m.atlas.openErr = "the selected entry point is not healthy"
		}
		return nil
	}
	m.atlas.openGen++
	m.atlas.opening = true
	m.atlas.openErr = ""
	return openWorkspace(m.workspaceSvc, core.WorkspaceRequest{
		Start: entry.Checkout, SpaceID: entry.ID, ExpectedPlanningID: entry.PlanningID,
	}, m.atlas.openGen)
}

// openAtlasWork enters the space owning the selected task, and lands ON that task rather
// than on the space's dashboard — the row was chosen because of the work it names, so
// arriving at a dashboard would make the user re-find it. The space is resolved by durable
// planning identity, never by the registry label, so a relabelled entry still matches.
func (m *Model) openAtlasWork() tea.Cmd {
	row, ok := m.atlas.selectedWork()
	if !ok {
		m.atlas.openErr = "nothing in progress to open"
		return nil
	}
	space, ok := m.atlas.spaceFor(row)
	if !ok {
		m.atlas.openErr = "the space for " + row.Task.Slug + " is no longer registered"
		return nil
	}
	m.atlas.cursor = space
	cmd := m.openAtlasSelection()
	if cmd == nil {
		return nil // openAtlasSelection already explained why (unhealthy, or unavailable)
	}
	m.pendingJump = pendingJump{kind: entityTasks, id: row.Task.Slug, set: true}
	return cmd
}

// spaceFor locates the card owning a work row by planning identity, falling back to the
// registry label for id-less legacy trees.
func (a atlas) spaceFor(row core.SpaceInProgress) (int, bool) {
	for i, space := range a.spaces {
		if row.PlanningID != "" && space.summary.PlanningID == row.PlanningID {
			return i, true
		}
		if row.PlanningID == "" && space.summary.ID == row.SpaceID {
			return i, true
		}
	}
	return 0, false
}

func (a *atlas) setOverview(overview core.SpaceOverview) {
	a.setWork(overview.InProgress)
	selectedSpace, selectedEntry := "", ""
	if len(a.spaces) > 0 && a.cursor >= 0 && a.cursor < len(a.spaces) {
		selected := a.spaces[a.cursor]
		selectedSpace = atlasSpaceKey(selected.summary)
		if entry, ok := selected.selectedEntry(); ok {
			selectedEntry = entry.ID
		}
	}

	a.spaces = make([]atlasSpace, len(overview.Spaces))
	a.cursor = 0
	for i, summary := range overview.Spaces {
		space := atlasSpace{summary: summary, registry: i}
		if summary.Selected != nil {
			for j, entry := range summary.Entries {
				if entry.ID == summary.Selected.ID {
					space.entry = j
					break
				}
			}
		}
		if atlasSpaceKey(summary) == selectedSpace {
			for j, entry := range summary.Entries {
				if entry.ID == selectedEntry {
					space.entry = j
					break
				}
			}
		}
		a.spaces[i] = space
	}
	a.sortSpaces()
	a.restoreCursor(selectedSpace)
	a.loaded = true
	a.stale = false
	a.loadErr = nil
	a.openErr = ""
}

// setWork refreshes the working set, keeping the cursor on the same task across reloads
// where it survived. Rows are ordered most-recently-touched first: the question this view
// answers is "what am I in the middle of", and recency is the closest proxy the planning
// data has. Space and slug break ties so the order is byte-stable between runs.
func (a *atlas) setWork(rows []core.SpaceInProgress) {
	selected := ""
	if row, ok := a.selectedWork(); ok {
		selected = workKey(row)
	}
	a.work = append([]core.SpaceInProgress(nil), rows...)
	a.sortWork()
	a.workCursor = 0
	for i, row := range a.work {
		if workKey(row) == selected {
			a.workCursor = i
			break
		}
	}
}

// sortWork applies the active order. Every comparison falls through to space then slug, so
// the result is byte-stable between runs whatever the axis.
func (a *atlas) sortWork() {
	sort.SliceStable(a.work, func(i, j int) bool {
		x, y := a.work[i], a.work[j]
		cmp := 0
		switch a.workOrder {
		case atlasWorkBySpace:
			cmp = strings.Compare(strings.ToLower(x.SpaceID), strings.ToLower(y.SpaceID))
		case atlasWorkByPriority:
			cmp = priorityRank(x.Task.Priority) - priorityRank(y.Task.Priority)
		default:
			// Longest-underway first: the oldest start date sorts to the top, which is
			// where a stuck task should be.
			if dx, dy := theme.StartedDate(x.Task), theme.StartedDate(y.Task); dx != dy {
				cmp = strings.Compare(dx, dy)
			}
		}
		if cmp == 0 {
			if cmp = strings.Compare(x.SpaceID, y.SpaceID); cmp == 0 {
				cmp = strings.Compare(x.Task.Slug, y.Task.Slug)
			}
		}
		if a.reverse {
			return cmp > 0
		}
		return cmp < 0
	})
}

// cycleWorkOrder advances the work axis, keeping the cursor on the same task.
func (a *atlas) cycleWorkOrder(delta int) {
	if delta >= 0 {
		a.workOrder = atlasWorkOrder((int(a.workOrder) + 1) % 3)
	}
	a.applyWorkOrder()
}

func (a *atlas) applyWorkOrder() {
	selected := ""
	if row, ok := a.selectedWork(); ok {
		selected = workKey(row)
	}
	a.sortWork()
	a.workCursor = 0
	for i, row := range a.work {
		if workKey(row) == selected {
			a.workCursor = i
			break
		}
	}
	a.openErr = ""
}

func workKey(row core.SpaceInProgress) string {
	return row.PlanningID + "\x00" + row.Task.Slug
}

func (a atlas) selectedWork() (core.SpaceInProgress, bool) {
	if a.workCursor < 0 || a.workCursor >= len(a.work) {
		return core.SpaceInProgress{}, false
	}
	return a.work[a.workCursor], true
}

func (a *atlas) moveWork(delta int) {
	if n := len(a.work); n > 0 {
		a.workCursor = ((a.workCursor+delta)%n + n) % n
		a.openErr = ""
	}
}

func (a *atlas) cycleOrder() {
	a.order = (a.order + 1) % 3
	a.applyOrder()
}

func (a *atlas) applyOrder() {
	selected := ""
	if len(a.spaces) > 0 && a.cursor >= 0 && a.cursor < len(a.spaces) {
		selected = atlasSpaceKey(a.spaces[a.cursor].summary)
	}
	a.sortSpaces()
	a.restoreCursor(selected)
	a.openErr = ""
}

func (a *atlas) sortSpaces() {
	sort.SliceStable(a.spaces, func(i, j int) bool {
		cmp := compareAtlasSpaces(a.spaces[i], a.spaces[j], a.order)
		if a.reverse {
			return cmp > 0
		}
		return cmp < 0
	})
}

func (a *atlas) restoreCursor(key string) {
	a.cursor = 0
	if key == "" {
		return
	}
	for i, space := range a.spaces {
		if atlasSpaceKey(space.summary) == key {
			a.cursor = i
			return
		}
	}
}

func compareAtlasSpaces(a, b atlasSpace, order atlasOrder) int {
	switch order {
	case atlasOrderActivity:
		// Activity's useful default is most-active first; reverse exposes the inverse.
		if aCount, bCount := atlasActivity(a.summary), atlasActivity(b.summary); aCount != bCount {
			return bCount - aCount
		}
	case atlasOrderRegistry:
		if a.registry != b.registry {
			return a.registry - b.registry
		}
	}
	return strings.Compare(strings.ToLower(atlasSpaceName(a.summary)), strings.ToLower(atlasSpaceName(b.summary)))
}

func atlasActivity(summary core.SpaceSummary) int {
	if summary.Summary == nil {
		return 0
	}
	return len(summary.Summary.InProgress)
}

func atlasSpaceName(summary core.SpaceSummary) string {
	if summary.ID != "" {
		return summary.ID
	}
	return summary.PlanningID
}

func atlasSpaceKey(summary core.SpaceSummary) string {
	if summary.PlanningID != "" {
		return "planning:" + summary.PlanningID
	}
	if len(summary.Entries) > 0 {
		return "entry:" + summary.Entries[0].ID
	}
	return "summary:" + summary.ID
}

func (a *atlas) move(delta int) {
	if n := len(a.spaces); n > 0 {
		a.cursor = ((a.cursor+delta)%n + n) % n
		a.openErr = ""
	}
}

func (a *atlas) moveEntry(delta int) {
	if len(a.spaces) == 0 {
		return
	}
	space := &a.spaces[a.cursor]
	if n := len(space.summary.Entries); n > 0 {
		space.entry = ((space.entry+delta)%n + n) % n
		a.openErr = ""
	}
}

func (a atlas) selectedEntry() (core.SpaceEntryPoint, bool) {
	if len(a.spaces) == 0 || a.cursor < 0 || a.cursor >= len(a.spaces) {
		return core.SpaceEntryPoint{}, false
	}
	return a.spaces[a.cursor].selectedEntry()
}

func (s atlasSpace) selectedEntry() (core.SpaceEntryPoint, bool) {
	if len(s.summary.Entries) == 0 || s.entry < 0 || s.entry >= len(s.summary.Entries) {
		return core.SpaceEntryPoint{}, false
	}
	return s.summary.Entries[s.entry], true
}

func (a atlas) view(st *styles, current core.Workspace, maxW, maxH int) string {
	if a.loadErr != nil && !a.loaded {
		return st.fg(theme.ColorRed, "atlas unavailable: "+a.loadErr.Error())
	}
	if !a.loaded {
		return st.dim("loading atlas…")
	}
	if len(a.spaces) == 0 {
		return strings.Join([]string{
			st.dashHeading.Render("atlas"),
			"",
			st.dim("No spaces registered."),
			"Run `tskflwctl space add <path>` to add one.",
		}, "\n")
	}

	if a.screen == atlasScreenWork {
		return a.workView(st, maxW, maxH)
	}
	direction := "↑"
	if a.reverse {
		direction = "↓"
	}
	// Four bands: a pinned header, a scrolling table, the focused space's entry points, and
	// a pinned status row. Only the table scrolls — everything the screen says about itself,
	// and everything about where ⏎ would take you, stays on screen.
	header := []string{st.dashHeading.Render(fmt.Sprintf(
		"atlas · spaces · %s · order %s %s",
		countLabel{len(a.spaces), "space", "spaces"}, a.order.label(), direction,
	)), ""}
	var status []string
	switch {
	case a.opening:
		status = []string{"", st.dim("opening selected space…")}
	case a.openErr != "":
		status = []string{"", st.fg(theme.ColorRed, "✘ "+a.openErr)}
	case a.loadErr != nil:
		status = []string{"", st.fg(theme.ColorYellow, "⚠ refresh failed: "+a.loadErr.Error())}
	}
	band := a.entryBand(st, current, maxW)
	body := a.spaceRows(st, current, maxW)
	if maxH > 0 {
		if budget := max1(maxH - len(header) - len(band) - len(status)); len(body) > budget {
			body = scrollTo(body, a.cursor, budget)
		}
	}
	lines := append(append(append(header, body...), band...), status...)
	return strings.Join(truncateAll(lines, maxW), "\n")
}

// spaceRows renders one aligned row per space. Columns are measured across the whole set —
// through the same relDateCells/countsWidth helpers the dashboard's epic widget uses — so
// the bar column is scannable and two spaces can be compared without reading either.
func (a atlas) spaceRows(st *styles, current core.Workspace, maxW int) []string {
	stats := make([]atlasStats, len(a.spaces))
	nameW, activeW := 0, 0
	for i, space := range a.spaces {
		stats[i] = statsFor(space.summary)
		nameW = maxInt(nameW, ansi.StringWidth(atlasSpaceName(space.summary)))
		activeW = maxInt(activeW, ansi.StringWidth(fmt.Sprintf("%d", stats[i].inProgress)))
	}
	nameW = minInt(nameW, 28)
	countsW := countsWidth(stats, func(s atlasStats) (int, int) { return s.done, s.total })
	dates := relDateCells(stats, func(s atlasStats) string { return s.lastTouched }, st)
	attentionW, countsW2 := 0, 0
	counts := make([]string, len(stats))
	for i, s := range stats {
		counts[i] = atlasCountsLine(st, s)
		countsW2 = maxInt(countsW2, ansi.StringWidth(counts[i]))
		if s.attention > 0 {
			attentionW = maxInt(attentionW, ansi.StringWidth(fmt.Sprintf("⚠%d", s.attention)))
		}
	}

	// Same fitting rule as the work view: drop columns whole rather than shearing every row
	// at the right edge. With real space names this table lands at 92 columns, so an 80-wide
	// terminal has to give something up — the breakdown first, then recency, both of which
	// are context beside the bar and the counts.
	showCounts, showDate := countsW2 > 0, false
	for _, d := range dates {
		showDate = showDate || d != ""
	}
	fixed := func() int {
		w := 2 + 2 + nameW + 2 + 10 + 1 + 4 + 2 + countsW + 2 + 1 + activeW
		if showCounts {
			w += 2 + countsW2
		}
		if attentionW > 0 {
			w += 2 + attentionW
		}
		if showDate {
			w += 2 + ansi.StringWidth(ansi.Strip(dates[0]))
		}
		return w
	}
	for avail := max1(maxW); fixed() > avail; {
		if showCounts {
			showCounts = false
			continue
		}
		if showDate {
			showDate = false
			continue
		}
		break
	}

	rows := make([]string, 0, len(a.spaces))
	for i, space := range a.spaces {
		marker := "  "
		if i == a.cursor {
			marker = st.selected.Render("› ")
		}
		name := padRight(truncate(atlasSpaceName(space.summary), nameW), nameW)
		if isCurrentSpace(current, space.summary) {
			name = st.fg(theme.ColorGreen, name)
		}
		row := marker + a.spaceGlyph(st, space) + " " + name
		if !stats[i].loaded {
			// A space whose summary could not be read keeps its row and says why, rather
			// than showing a misleading empty bar.
			detail := space.summary.LoadError
			if detail == "" {
				detail = "summary unavailable"
			}
			rows = append(rows, row+"  "+st.fg(theme.ColorRed, detail))
			continue
		}
		pct := stats[i].percent()
		row += fmt.Sprintf("  %s %s  %s", st.miniBar(pct, 10),
			st.fg(theme.Percent(pct), theme.PercentLabelPadded(pct)),
			rollupCounts(stats[i].done, stats[i].total, countsW))
		row += "  " + st.fg(theme.ColorCyan, fmt.Sprintf("▸%*d", activeW, stats[i].inProgress))
		if showCounts {
			row += "  " + padRight(counts[i], countsW2)
		}
		if attentionW > 0 {
			cell := strings.Repeat(" ", attentionW)
			if stats[i].attention > 0 {
				cell = st.fg(theme.ColorYellow, padRight(fmt.Sprintf("⚠%d", stats[i].attention), attentionW))
			}
			row += "  " + cell
		}
		if showDate && dates[i] != "" {
			row += "  " + dates[i]
		}
		rows = append(rows, row)
	}
	return rows
}

// spaceGlyph reports at a glance whether a space can be entered at all.
func (a atlas) spaceGlyph(st *styles, space atlasSpace) string {
	if entry, ok := space.selectedEntry(); ok && entry.Healthy() {
		return st.fg(theme.ColorGreen, "●")
	}
	return st.fg(theme.ColorRed, "!")
}

// entryBand is the pinned detail strip under the table: the focused space's registered
// entry points, with the one ⏎ would open marked. It lives outside the scrolled body on
// purpose — this is the answer to "where would Enter take me", and a table tall enough to
// scroll is exactly when that answer must not scroll away.
func (a atlas) entryBand(st *styles, current core.Workspace, maxW int) []string {
	if a.cursor < 0 || a.cursor >= len(a.spaces) {
		return nil
	}
	space := a.spaces[a.cursor]
	entries := space.summary.Entries
	if len(entries) == 0 {
		return []string{"", st.dim("no registered entry point for this space")}
	}
	title := fmt.Sprintf(" %s · %d entry point", atlasSpaceName(space.summary), len(entries))
	if len(entries) != 1 {
		title += "s"
	}
	if len(entries) > 1 {
		title += st.dim("  h/l choose")
	}
	rule := st.dim("─" + title + " " + strings.Repeat("─", max1(maxW-ansi.StringWidth(title)-3)))
	band := []string{"", rule}
	for i, candidate := range entries {
		cursor := "  "
		if i == space.entry {
			cursor = st.selected.Render("› ")
		}
		glyph := st.fg(theme.ColorGreen, "●")
		if !candidate.Healthy() {
			glyph = st.fg(theme.ColorRed, "!")
		}
		directory := candidate.Path
		if directory == "" {
			directory = candidate.Checkout
		}
		line := fmt.Sprintf("%s%s %s  %s  %s", cursor, glyph, candidate.ID,
			st.dim(string(candidate.Role)), st.dim(directory))
		if isCurrentEntry(current, candidate) {
			line += "  " + st.fg(theme.ColorGreen, "current")
		}
		band = append(band, line)
		if !candidate.Healthy() && candidate.Detail != "" {
			band = append(band, "    "+st.fg(theme.ColorRed, candidate.Detail))
		}
		if !candidate.Healthy() && candidate.Remedy != "" {
			band = append(band, "    "+st.dim("remedy: "+candidate.Remedy))
		}
	}
	return band
}

// atlasStats is the per-space aggregate a table row renders. Every number here is already
// computed inside core.Summary for the dashboard; the atlas kept only a few of them, as
// prose. Deriving them once per row keeps the render loop free of arithmetic.
type atlasStats struct {
	loaded      bool
	done, total int
	inProgress  int
	epics       int
	openAudits  int
	findings    int
	// attention folds the things that want a PERSON — acute findings, audits whose findings
	// are all resolved, snoozed tasks come due, files that cannot be read — as opposed to
	// the ordinary open counts beside it. One number because four columns of mostly-zeroes
	// would be worse than one that is usually absent; the space's own overview is where you
	// dig in.
	attention   int
	lastTouched string
}

func statsFor(summary core.SpaceSummary) atlasStats {
	s := summary.Summary
	if s == nil {
		return atlasStats{}
	}
	stats := atlasStats{
		loaded: true, inProgress: len(s.InProgress),
		epics: len(s.Epics), openAudits: len(s.OpenAudits),
		findings:  s.Findings.Open + s.Findings.InProgress,
		attention: len(s.Findings.Acute) + s.ReadyToClose + s.RevisitDue + len(s.Problems),
	}
	for _, e := range s.Epics {
		stats.done += e.Done
		stats.total += e.Total
		if e.LastUpdated > stats.lastTouched {
			stats.lastTouched = e.LastUpdated
		}
	}
	return stats
}

func (a atlasStats) percent() int {
	if a.total <= 0 {
		return 0
	}
	return a.done * 100 / a.total
}

// countLabel is one segment of the "4 epics · 1 audit" breakdown. Zero-count segments are
// dropped rather than rendered, so a quiet space reads quietly.
type countLabel struct {
	n         int
	one, many string
}

func (c countLabel) String() string {
	unit := c.many
	if c.n == 1 {
		unit = c.one
	}
	return fmt.Sprintf("%d %s", c.n, unit)
}

func atlasCountsLine(st *styles, stats atlasStats) string {
	var kept []countLabel
	for _, c := range []countLabel{
		{stats.epics, "epic", "epics"},
		{stats.openAudits, "audit", "audits"},
		{stats.findings, "finding", "findings"},
	} {
		if c.n > 0 {
			kept = append(kept, c)
		}
	}
	return theme.Breakdown(kept, st.dim(" · "), 0,
		func(c countLabel) string { return st.dim(c.String()) }, nil)
}

// workView answers "what am I in the middle of, anywhere?" — the question the design
// sketch calls the atlas's actual payload, and the one `status --all` can answer on the
// CLI but the TUI could not. One row per in-progress task across every healthy space.
func (a atlas) workView(st *styles, maxW, maxH int) string {
	direction := "↑"
	if a.reverse {
		direction = "↓"
	}
	header := []string{st.dashHeading.Render(fmt.Sprintf(
		"atlas · work · %d in progress across %s · order %s %s",
		len(a.work), countLabel{a.workSpaceCount(), "space", "spaces"},
		a.workOrder.label(), direction,
	)), ""}
	var status []string
	switch {
	case a.opening:
		status = []string{"", st.dim("opening selected space…")}
	case a.openErr != "":
		status = []string{"", st.fg(theme.ColorRed, "✘ "+a.openErr)}
	case a.loadErr != nil:
		status = []string{"", st.fg(theme.ColorYellow, "⚠ refresh failed: "+a.loadErr.Error())}
	}
	if len(a.work) == 0 {
		body := []string{st.dim("Nothing in progress across registered spaces."), "",
			"Start something with `tskflwctl task start <slug>`, or press " +
				keys.View.Help().Key + " for spaces."}
		return strings.Join(truncateAll(append(append(header, body...), status...), maxW), "\n")
	}

	body, focus := a.workRows(st, maxW)
	if maxH > 0 {
		if budget := max1(maxH - len(header) - len(status)); len(body) > budget {
			body = scrollTo(body, focus, budget)
		}
	}
	return strings.Join(truncateAll(append(append(header, body...), status...), maxW), "\n")
}

// workRows renders the working set and reports which line the cursor is on. Grouping is a
// property of the ORDER: only the space order draws headings, because grouping by started
// date or priority would emit a heading per row.
//
// Columns are FITTED to the terminal, not truncated at its right edge. Right-edge
// truncation looks fine on short space names and then quietly amputates the useful half of
// every row on real ones: a 92-column terminal showing `[desirelines-planning]` plus a
// 76-character slug has already spent 69 columns before the first data column, so age —
// the whole point of this view — never renders at all. Columns are therefore dropped
// WHOLE, least load-bearing first, so what survives is readable rather than sheared.
func (a atlas) workRows(st *styles, maxW int) ([]string, int) {
	grouped := a.workOrder.grouped()
	spaceW, slugW, epicW, prioW, ageW := 0, 0, 0, 0, 0
	ages := make([]string, len(a.work))
	for i, row := range a.work {
		ages[i] = theme.RelativeDate(theme.StartedDate(row.Task))
		spaceW = maxInt(spaceW, ansi.StringWidth(row.SpaceID)+2) // the [brackets]
		slugW = maxInt(slugW, ansi.StringWidth(row.Task.Slug))
		epicW = maxInt(epicW, ansi.StringWidth(row.Task.Epic))
		prioW = maxInt(prioW, ansi.StringWidth(row.Task.Priority))
		ageW = maxInt(ageW, ansi.StringWidth(ages[i]))
	}

	// Sacrifice order, least load-bearing first. Slug and age always survive: without the
	// slug there is no row, and without the age this view is just a list.
	const (
		gap       = 2
		cursorW   = 2
		slugFloor = 18
		descFloor = 16
	)
	// A very long slug is identity, not information: letting it take the whole row buys one
	// unabbreviated name at the cost of the epic and priority columns. Cap it at a share of
	// the terminal first, fit the rest, then give any leftover back to it.
	naturalSlugW := slugW
	slugW = minInt(slugW, maxInt(slugFloor, maxW*2/5))
	showSpace, showEpic, showPrio := !grouped, epicW > 0, prioW > 0
	fixed := func() int {
		w := cursorW + slugW + gap + ageW
		if showSpace {
			w += spaceW + 1
		}
		if showEpic {
			w += gap + epicW
		}
		if showPrio {
			w += gap + prioW
		}
		return w
	}
	avail := max1(maxW)
	for fixed() > avail {
		switch {
		case showEpic:
			showEpic = false
		case showPrio:
			showPrio = false
		case slugW > slugFloor:
			slugW = maxInt(slugFloor, slugW-(fixed()-avail))
		case showSpace:
			showSpace = false // last resort: grouping or the space column, but not neither
		default:
			slugW = max1(slugW - (fixed() - avail))
		}
		if !showEpic && !showPrio && !showSpace && slugW <= slugFloor {
			break // nothing left to give; truncate handles the remainder
		}
	}
	// Description is the flex column: it takes whatever is left, and is dropped rather than
	// rendered as a two-character stub. Anything still spare after that goes back to the
	// slug, up to its natural width, so a wide terminal shows the whole name.
	descW := avail - fixed() - gap
	showDesc := descW >= descFloor
	if !showDesc {
		descW = 0
		if spare := avail - fixed(); spare > 0 {
			slugW = minInt(naturalSlugW, slugW+spare)
		}
	} else if spare := descW - descFloor; spare > 0 && naturalSlugW > slugW {
		grow := minInt(spare, naturalSlugW-slugW)
		slugW += grow
		descW -= grow
	}

	body := make([]string, 0, len(a.work)+2*a.workSpaceCount())
	focus, lastSpace := 0, ""
	for i, row := range a.work {
		if grouped && row.SpaceID != lastSpace {
			lastSpace = row.SpaceID
			if len(body) > 0 {
				body = append(body, "")
			}
			body = append(body, st.dashHeading.Render(row.SpaceID)+"  "+
				st.dim(countLabel{a.countIn(row.SpaceID), "task", "tasks"}.String()))
		}
		cursor := "  "
		if i == a.workCursor {
			cursor = st.selected.Render("› ")
			focus = len(body)
		}
		line := cursor
		if showSpace {
			// The heading carries the space when grouped; repeating it per row would be
			// noise rather than structure.
			line += st.dim(padRight(truncate("["+row.SpaceID+"]", spaceW), spaceW)) + " "
		}
		line += padRight(truncate(row.Task.Slug, slugW), slugW)
		if showEpic {
			line += "  " + st.dim(padRight(truncate(row.Task.Epic, epicW), epicW))
		}
		if showPrio {
			line += "  " + padRight(st.priorityText(row.Task.Priority), prioW)
		}
		// Age is the point of this view: how long has this been underway, coloured by how
		// much that should worry you.
		line += "  " + st.fg(theme.Staleness(theme.DaysSince(theme.StartedDate(row.Task))),
			padRight(ages[i], ageW))
		if showDesc {
			line += "  " + st.dim(truncate(row.Task.Description, descW))
		}
		body = append(body, line)
	}
	return body, focus
}

// countIn is how many rows of the working set belong to one space, for the group heading.
func (a atlas) countIn(spaceID string) int {
	n := 0
	for _, row := range a.work {
		if row.SpaceID == spaceID {
			n++
		}
	}
	return n
}

// workSpaceCount is how many distinct spaces the working set spans — "5 in progress" reads
// very differently depending on whether that is one repo or five.
func (a atlas) workSpaceCount() int {
	seen := make(map[string]struct{}, len(a.work))
	for _, row := range a.work {
		seen[workSpaceKey(row)] = struct{}{}
	}
	return len(seen)
}

func workSpaceKey(row core.SpaceInProgress) string {
	if row.PlanningID != "" {
		return row.PlanningID
	}
	return row.SpaceID
}

func truncateAll(lines []string, maxW int) []string {
	for i := range lines {
		lines[i] = truncate(lines[i], max1(maxW))
	}
	return lines
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isCurrentEntry reports whether one registered entry point IS the open workspace.
// Durable identity is preferred over path spelling: SpaceID is set whenever the atlas
// opened this workspace, and it survives symlinks, ~-expansion, and case-insensitive
// volumes that a string compare of two paths does not. The path compare remains the
// answer for the launch workspace, which carries no SpaceID unless --space named one;
// both sides of it now come from config.Discover, so both are symlink-evaluated.
func isCurrentEntry(current core.Workspace, entry core.SpaceEntryPoint) bool {
	if current.SpaceID != "" && entry.ID != "" {
		return current.SpaceID == entry.ID
	}
	return filepathKey(entry.Checkout) == workspaceKey(current)
}

// isCurrentSpace reports whether the open workspace belongs to this logical space — a
// weaker question than isCurrentEntry, and the right one for the card title: you are "in"
// desirelines whichever of its three registered checkouts you entered through.
func isCurrentSpace(current core.Workspace, summary core.SpaceSummary) bool {
	if current.PlanningID != "" && summary.PlanningID != "" {
		return current.PlanningID == summary.PlanningID
	}
	for _, entry := range summary.Entries {
		if isCurrentEntry(current, entry) {
			return true
		}
	}
	return false
}

// filepathKey mirrors workspaceKey for a raw checkout without manufacturing a partial
// core.Workspace at each render site.
func filepathKey(checkout string) string {
	return workspaceKey(core.Workspace{Checkout: checkout})
}
