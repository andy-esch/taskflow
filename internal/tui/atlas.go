package tui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/theme"
)

// atlas is the cross-space navigation screen. It owns only the registry projection and
// cursor state; opened planning data lives in the active/cached spaceSession.
type atlas struct {
	loaded  bool
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
		return m, nil
	}
	return m, m.activateWorkspace(msg.workspace, msg.watcher, msg.watchErr)
}

func (m Model) handleAtlasKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "j" || msg.String() == "down":
		m.atlas.move(1)
	case msg.String() == "k" || msg.String() == "up":
		m.atlas.move(-1)
	case msg.String() == "h" || msg.String() == "left":
		m.atlas.moveEntry(-1)
	case msg.String() == "l" || msg.String() == "right":
		m.atlas.moveEntry(1)
	case key.Matches(msg, keys.Sort):
		m.atlas.cycleOrder()
	case key.Matches(msg, keys.SortRev):
		m.atlas.reverse = !m.atlas.reverse
		m.atlas.applyOrder()
	case msg.String() == "enter":
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
	if refresh || !m.atlas.loaded {
		m.atlas.loadGen++
		return loadAtlas(m.spaceOverviewSvc, m.atlas.loadGen)
	}
	return nil
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

func (a *atlas) setOverview(overview core.SpaceOverview) {
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
	a.loadErr = nil
	a.openErr = ""
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

	direction := "↑"
	if a.reverse {
		direction = "↓"
	}
	// Header and status are PINNED: only the cards scroll. They were both inside the
	// scrolled slice once, which meant a registry big enough to need scrolling was also
	// big enough to hide the sort order you just changed and the ✘ explaining why ⏎ did
	// nothing. Everything the screen says about itself has to stay on screen.
	header := []string{st.dashHeading.Render(fmt.Sprintf(
		"atlas · %d space(s) · order %s %s", len(a.spaces), a.order.label(), direction,
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

	var body []string
	focusRow := 0
	for i, space := range a.spaces {
		card, cardFocus := renderAtlasCard(st, space, i == a.cursor, current)
		if i == a.cursor {
			focusRow = len(body) + cardFocus
		}
		body = append(body, card...)
		if i < len(a.spaces)-1 {
			body = append(body, "")
		}
	}
	if maxH > 0 {
		// One card row is the floor: a terminal too short for even that is already being
		// hard-clamped by View, and scrollTo cannot take a non-positive window.
		budget := max1(maxH - len(header) - len(status))
		if len(body) > budget {
			body = scrollTo(body, focusRow, budget)
		}
	}

	lines := append(append(header, body...), status...)
	for i := range lines {
		lines[i] = truncate(lines[i], max1(maxW))
	}
	return strings.Join(lines, "\n")
}

func renderAtlasCard(st *styles, space atlasSpace, selected bool, current core.Workspace) ([]string, int) {
	summary := space.summary
	entry, hasEntry := space.selectedEntry()
	marker := "  "
	if selected {
		marker = st.selected.Render("› ")
	}
	title := atlasSpaceName(summary)
	if hasEntry && entry.Label != "" && entry.Label != title {
		title += "  " + st.dim(entry.Label)
	}
	if isCurrentSpace(current, summary) {
		title += "  " + st.fg(theme.ColorGreen, "current")
	}
	lines := []string{marker + st.dashHeading.Render(title)}
	focusRow := 0

	switch {
	case summary.Summary != nil:
		lines = append(lines, "  "+atlasSummaryLine(st, *summary.Summary))
	case summary.LoadError != "":
		lines = append(lines, "  "+st.fg(theme.ColorRed, "! "+summary.LoadError))
	default:
		lines = append(lines, "  "+st.dim("summary unavailable"))
	}
	if hasEntry {
		entryLine := fmt.Sprintf("  entry points · %d/%d selected", space.entry+1, len(summary.Entries))
		if len(summary.Entries) > 1 {
			entryLine += st.dim("  h/l choose")
		}
		lines = append(lines, entryLine)
		for i, candidate := range summary.Entries {
			glyph := st.fg(theme.ColorGreen, "●")
			if !candidate.Healthy() {
				glyph = st.fg(theme.ColorRed, "!")
			}
			cursor := "  "
			if selected && i == space.entry {
				cursor = st.selected.Render("› ")
				focusRow = len(lines)
			}
			directory := candidate.Path
			if directory == "" {
				directory = candidate.Checkout
			}
			line := fmt.Sprintf("  %s%s %s  %s  %s", cursor, glyph, candidate.ID,
				st.dim(string(candidate.Role)), st.dim(directory))
			if isCurrentEntry(current, candidate) {
				line += "  " + st.fg(theme.ColorGreen, "current")
			}
			lines = append(lines, line)
			if !candidate.Healthy() && candidate.Detail != "" {
				lines = append(lines, "      "+st.fg(theme.ColorRed, candidate.Detail))
			}
			if !candidate.Healthy() && candidate.Remedy != "" {
				lines = append(lines, "      "+st.dim("remedy: "+candidate.Remedy))
			}
		}
	}
	return lines, focusRow
}

func atlasSummaryLine(st *styles, summary core.Summary) string {
	parts := []string{fmt.Sprintf("%d in progress", len(summary.InProgress))}
	if len(summary.Epics) > 0 {
		parts = append(parts, fmt.Sprintf("%d epic(s)", len(summary.Epics)))
	}
	if len(summary.OpenAudits) > 0 {
		parts = append(parts, fmt.Sprintf("%d open audit(s)", len(summary.OpenAudits)))
	}
	if n := summary.Findings.Open + summary.Findings.InProgress; n > 0 {
		parts = append(parts, fmt.Sprintf("%d finding(s)", n))
	}
	return strings.Join(parts, st.dim(" · "))
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
