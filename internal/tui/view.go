// view.go: the Model's render + layout half — View/renderBody/footer/panes/tabStrip.
package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/theme"
)

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
	listH := max1(bodyH - m.st.paneVFrame - 1)
	detailH := max1(bodyH - m.st.paneVFrame - titleH)
	// Zoom forces the single-pane path (detail full width); otherwise a wide terminal
	// shows the list+detail split.
	m.twoPane = m.width >= 90 && !m.zoom

	var listInnerW int
	if m.twoPane {
		listOuterW := m.width * 2 / 5
		if listOuterW < 28 {
			listOuterW = 28
		}
		m.listOuterW = listOuterW
		m.detailOuterW = m.width - listOuterW
		listInnerW = max1(listOuterW - m.st.paneHFrame)
		m.detail.SetSize(max1(m.detailOuterW-m.st.paneHFrame), detailH)
	} else {
		m.listOuterW, m.detailOuterW = m.width, m.width
		listInnerW = max1(m.width - m.st.paneHFrame)
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
		v.WindowTitle = m.windowTitle()
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
	v.WindowTitle = m.windowTitle()
	return v
}

// windowTitle is the terminal window/tab title (tea.View.WindowTitle): the app
// name plus the current selection (or the active tab when nothing's selected), so
// the entity you're on shows up in the terminal's tab bar.
func (m Model) windowTitle() string {
	if m.onDash {
		return "tskflwctl · dashboard"
	}
	if id := m.selectedID(); id != "" {
		return "tskflwctl · " + id
	}
	return "tskflwctl · " + m.cur().name
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
// spends 2 of that on its border rows. The line COUNT must match helpBox's, so it
// wraps at the same content width (helpWidth(width-2): bodyView hands the modal
// width-2). The j/k handler clamps to this so helpScroll can't run past the bottom.
func (m Model) helpMaxScroll() int {
	innerH := m.paneOuterH - 2 - 2 // box height (paneOuterH-2) minus top+bottom border
	if innerH <= 0 {
		return 0
	}
	contentW := helpWidth(m.width-2) - helpHFrame
	return max(len(helpLines(m.focus, m.helpEntityKind(), contentW, *m.st))-innerH, 0)
}

// helpEntityKind is the kind whose context notes the `?` panel shows — the active
// tab's, or the dashboard sentinel when on the landing screen.
func (m Model) helpEntityKind() entityKind {
	if m.onDash {
		return entityDashboard
	}
	return m.cur().kind
}

// renderBody is the pane layout: a loading note (or this tab's load error) until
// the active tab loads, then the two-pane (or single-pane drill) view. Errors
// are per tab — one failing loader must not blank tabs that loaded fine — and a
// tab that HAS loaded keeps its (stale) rows on a failed reload, with the
// failure flagged in the footer instead.
func (m Model) renderBody() string {
	if m.onDash {
		// Mirror the tab pattern: a failed first load shows the error pane; once it
		// has loaded, a failed refresh keeps the stale rows (flagged in the footer).
		switch {
		case m.dash.loadErr != nil && !m.dash.loaded:
			return m.pane(focusList, m.st.fg(theme.ColorRed, "error: "+m.dash.loadErr.Error()), m.width)
		case !m.dash.loaded:
			return m.pane(focusList, m.st.dim("loading…"), m.width)
		}
		return m.pane(focusList, m.dash.view(*m.st, m.width-m.st.paneHFrame, m.paneOuterH-m.st.paneVFrame), m.width)
	}
	switch t := m.cur(); {
	case t.loadErr != nil && !t.loaded:
		return m.pane(focusList, m.st.fg(theme.ColorRed, "error: "+t.loadErr.Error()), m.width)
	case !t.loaded:
		return m.pane(focusList, m.st.dim("loading…"), m.width)
	}
	if m.zoom { // full-screen: the detail pane takes the whole width (no list)
		return m.detailPaneView()
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
	titleStyle := m.st.dimStyle
	if m.focus == focusDetail {
		titleStyle = m.st.selected
	}
	// Truncate the title to the pane's inner width — an un-truncated long slug
	// would wrap to a second row, growing the pane past its budget and clipping
	// its bottom border (the truncate discipline every Join input must follow).
	title := titleStyle.Render(truncate(m.detailTitle(), max1(m.detail.width)))
	// Make the title click-to-open the entity's file (OSC 8) — only when content is
	// loaded, so it points at a real path. The escape adds no display width, so the
	// truncate/pane sizing above is unaffected.
	if p := m.detail.path(); p != "" {
		title = osc8(title, "file://"+p)
	}
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
	border := m.st.paneInactive
	if m.focus == f {
		border = m.st.paneActive
	}
	return border.Width(max1(outerW)).Height(max1(m.paneOuterH)).Render(content)
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

// tabStrip renders the dashboard plus the entity tabs (active accented),
// collapsing to a single `[name ▾]` chip under ~60 cols. The dashboard sits left
// of the entity tabs as the landing surface.
func (m Model) tabStrip() string {
	if m.width < 60 {
		name := "dashboard"
		if !m.onDash {
			name = m.cur().name
		}
		return truncate(m.st.activeTab.Render("["+name+" ▾]"), m.width)
	}
	parts := make([]string, 0, len(m.tabs)+1)
	if m.onDash {
		parts = append(parts, m.st.activeTab.Render("dashboard"))
	} else {
		parts = append(parts, m.st.dim("dashboard"))
	}
	for i, t := range m.tabs {
		if !m.onDash && i == m.active {
			parts = append(parts, m.st.activeTab.Render(t.name))
		} else {
			parts = append(parts, m.st.dim(t.name))
		}
	}
	return truncate(strings.Join(parts, m.st.dim("  ·  ")), m.width)
}

// keyHint renders a footer token "<key> <label>", the key sourced from its binding so
// a rebind in keys.go reaches the footer too (the other place the key vocabulary is
// shown). keyCombo does the same for a two-key pair ("[ ] tabs", "n/N match").
func keyHint(b key.Binding, label string) string { return b.Help().Key + " " + label }
func keyCombo(a, b key.Binding, sep, label string) string {
	return a.Help().Key + sep + b.Help().Key + " " + label
}

// detailFooterBody is the footer fragment shared by every detail-pane footer (focused,
// full-screen, single-pane drill), so the three variants compose it rather than each
// repeating it.
func detailFooterBody() string {
	return strings.Join([]string{
		keyHint(keys.Find, "find"),
		keyCombo(keys.FindNext, keys.FindPrev, "/", "match"),
		keyHint(keys.RawToggle, "raw/pretty"),
		"j/k scroll", // viewport keys — no keyMap binding
		keyCombo(keys.Top, keys.Bottom, "/", "top/bottom"),
	}, " · ")
}

func (m Model) footer() string {
	if m.cmd.active {
		// Surface the matching commands inline so `:` is self-documenting: the full
		// vocabulary on an empty prompt, narrowing to the prefix as you type.
		if hint := m.commandHint(); hint != "" {
			return truncate(m.cmd.view(*m.st)+m.st.dim("  "+hint), m.width)
		}
		return truncate(m.cmd.view(*m.st), m.width)
	}
	// A post-action result takes over the footer until the next key.
	if m.flash != "" {
		if m.flashErr {
			return truncate(m.st.fg(theme.ColorRed, "✘ "+m.flash), m.width)
		}
		return truncate(m.st.fg(theme.ColorGreen, "✔ "+m.flash), m.width)
	}
	// The detail find input/status takes over the footer while searching a body.
	if m.focus == focusDetail && (m.detail.finding() || m.detail.findActive()) {
		return truncate(m.detail.findStatus(), m.width)
	}
	// The dashboard isn't a list, so it gets its own hint line (no m/e/s/…), dimmed
	// to match the tab footers below. A failed refresh is flagged here (the rows
	// shown are the last good load), mirroring the tab "reload failed" note.
	if m.onDash {
		line := strings.Join([]string{
			"↑↓ move", "⏎ open",
			keyCombo(keys.PrevTab, keys.NextTab, " ", "tabs"),
			keyHint(keys.Command, "cmd"),
			keyHint(keys.Refresh, "refresh"),
			keyHint(keys.Help, "help"),
			keyHint(keys.Quit, "quit"),
		}, " · ")
		if m.dash.loaded && m.dash.loadErr != nil {
			line = "⚠ refresh failed · " + line
		}
		return m.st.dim(truncate(line, m.width))
	}
	hints := strings.Join([]string{
		keyHint(keys.Command, "cmd"),
		"/ filter", // the list's own filter (no keyMap binding)
		keyHint(keys.Action, "move"),
		keyHint(keys.Edit, "edit"),
		keyHint(keys.OpenEditor, "editor"),
		keyHint(keys.StatusView, "view"),
		keyCombo(keys.PrevTab, keys.NextTab, " ", "tabs"),
		keyHint(keys.Right, "detail"),
		keyHint(keys.Zoom, "full"),
		keyHint(keys.Help, "help"),
		keyHint(keys.Quit, "quit"),
	}, " · ")
	if m.focus == focusDetail {
		hints = strings.Join([]string{
			keyHint(keys.Command, "cmd"), detailFooterBody(), keyHint(keys.Zoom, "full"),
			keyCombo(keys.Left, keys.Back, "/", "back"),
		}, " · ")
		switch {
		case m.zoom:
			// Full-screen: the list is hidden, so name the way out and drop the keys
			// (m/e/s/tabs) that only make sense beside the list.
			hints = strings.Join([]string{
				"full-screen", detailFooterBody(), keys.Zoom.Help().Key + "/esc exit",
			}, " · ")
		case !m.twoPane:
			// Single-pane drill: q pops back to the list (context quit), so the
			// hint must not promise it exits the app.
			hints = strings.Join([]string{
				keyHint(keys.Command, "cmd"), detailFooterBody(),
				keys.Left.Help().Key + "/" + keys.Back.Help().Key + "/" + keys.Quit.Help().Key + " back",
			}, " · ")
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
	return m.st.dim(truncate(hints, m.width))
}
