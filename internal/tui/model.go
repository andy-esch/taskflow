// Package tui is the second primary adapter: an interactive Bubble Tea front-end
// over the same core.Service the CLI uses. It never touches the store/fs — all
// reads run as tea.Cmds against the service (see commands.go).
package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

type focus int

const (
	focusList focus = iota
	focusDetail
)

// Model is the root TUI model: a two-pane read-only task browser over the core
// service. List on the left, detail preview on the right.
type Model struct {
	svc  *core.Service
	root string

	width, height int
	twoPane       bool
	listOuterW    int
	detailOuterW  int
	paneOuterH    int

	focus    focus
	list     list.Model
	detail   detailPane
	loading  bool
	err      error
	problems []domain.FileProblem
	restore  string // slug to re-select after a reload
}

// New constructs the root model over the same *core.Service the CLI uses.
func New(svc *core.Service, root string) Model {
	l := list.New(nil, taskDelegate{}, 0, 0)
	l.Title = "Tasks"
	l.Styles.Title = lipgloss.NewStyle().Bold(true)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	// Built-in fuzzy `/` filter over FilterValue (slug + description); the list's
	// title bar doubles as the filter input. Sprint 2 adds the persistent filter
	// chip, `:` status views, and sortable columns.
	return Model{svc: svc, root: root, focus: focusList, list: l, detail: newDetailPane(), loading: true}
}

func (m Model) Init() tea.Cmd { return loadTasks(m.svc) }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.recomputeLayout()
		return m, nil

	case tea.KeyMsg:
		// While the list's filter input is capturing text, the list owns every
		// key — don't let global hotkeys (q/r/…) leak into the query.
		if m.list.SettingFilter() {
			return m.updateList(msg)
		}
		switch {
		case key.Matches(msg, keys.ForceQuit), key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Refresh):
			m.restore = m.selectedSlug()
			m.loading = true
			return m, loadTasks(m.svc)
		case key.Matches(msg, keys.ToggleFocus):
			m.toggleFocus()
			return m, nil
		}
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
		// detail focus — viewport handles j/k/ctrl+d/u; g/G aren't in its keymap.
		switch {
		case key.Matches(msg, keys.Left), key.Matches(msg, keys.Back):
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

	case tasksLoadedMsg:
		m.loading = false
		m.problems = msg.problems
		cmd := m.list.SetItems(msg.items)
		if m.restore != "" {
			m.selectSlug(m.restore)
			m.restore = ""
		}
		m.detail.loading = true
		return m, tea.Batch(cmd, loadBody(m.svc, m.selectedSlug()))

	case taskBodyMsg:
		if msg.slug != m.selectedSlug() {
			return m, nil // stale: selection changed since this load fired
		}
		m.detail.SetContent(msg.task, msg.body)
		return m, nil

	case bodyErrMsg:
		// A per-task load failure (e.g. an ambiguous duplicate slug) shows in the
		// detail pane — it must not blank the whole browser.
		if msg.slug != m.selectedSlug() {
			return m, nil
		}
		m.detail.SetError(msg.slug, msg.err.Error())
		return m, nil

	case reloadMsg:
		m.restore = m.selectedSlug()
		return m, loadTasks(m.svc)

	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	}
	// Forward anything else (notably the list's async FilterMatchesMsg, which
	// applies the `/` filter) to the list.
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// updateList forwards a key to the list, lazily loading the detail body when the
// selection changes.
func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	prev := m.selectedSlug()
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	if s := m.selectedSlug(); s != prev && s != "" {
		m.detail.loading = true
		return m, tea.Batch(cmd, loadBody(m.svc, s))
	}
	return m, cmd
}

func (m *Model) setFocus(f focus) {
	m.focus = f
	m.recomputeLayout()
}

func (m *Model) toggleFocus() {
	if m.focus == focusList {
		m.setFocus(focusDetail)
	} else {
		m.setFocus(focusList)
	}
}

func (m Model) selectedSlug() string {
	if it, ok := m.list.SelectedItem().(taskItem); ok {
		return it.t.Slug
	}
	return ""
}

func (m *Model) selectSlug(slug string) {
	for i, it := range m.list.Items() {
		if ti, ok := it.(taskItem); ok && ti.t.Slug == slug {
			m.list.Select(i)
			return
		}
	}
}

// recomputeLayout sizes the panes from the terminal size + responsive mode.
// Borders are subtracted before sizing children (the #1 lipgloss bug). The list
// renders its own title bar; the detail pane gets a manual title line.
func (m *Model) recomputeLayout() {
	const (
		footerH = 1
		titleH  = 1 // the detail pane's manual title line
	)
	bodyH := m.height - footerH
	if bodyH < 4 {
		bodyH = 4
	}
	m.paneOuterH = bodyH
	listH := max1(bodyH - paneVFrame)
	detailH := max1(bodyH - paneVFrame - titleH)
	m.twoPane = m.width >= 90
	if m.twoPane {
		listOuterW := m.width * 2 / 5
		if listOuterW < 28 {
			listOuterW = 28
		}
		m.listOuterW = listOuterW
		m.detailOuterW = m.width - listOuterW
		m.list.SetSize(max1(listOuterW-paneHFrame), listH)
		m.detail.SetSize(max1(m.detailOuterW-paneHFrame), detailH)
	} else {
		m.listOuterW, m.detailOuterW = m.width, m.width
		full := max1(m.width - paneHFrame)
		m.list.SetSize(full, listH)
		m.detail.SetSize(full, detailH)
	}
}

func (m Model) View() string {
	switch {
	case m.width == 0 || m.height == 0:
		// No WindowSizeMsg yet. Rendering panes now would use unset (0) sizes →
		// negative border dimensions → a broken oversized frame that corrupts the
		// renderer's height tracking (the clipped-top-border bug). Wait for size.
		return "loading…"
	case m.err != nil:
		return fg(theme.ColorRed, "error: "+m.err.Error())
	case m.loading:
		return "loading…"
	case len(m.list.Items()) == 0:
		return m.emptyView()
	}

	listPane := m.pane(focusList, m.list.View(), m.listOuterW)

	var view string
	switch {
	case m.twoPane:
		view = lipgloss.JoinHorizontal(lipgloss.Top, listPane, m.detailPaneView())
	case m.focus == focusDetail:
		view = m.detailPaneView()
	default:
		view = listPane
	}
	full := lipgloss.JoinVertical(lipgloss.Left, view, m.footer())
	// Last-line-of-defense clamp: a single missed truncation degrades gracefully
	// instead of overflowing/corrupting the screen.
	return lipgloss.NewStyle().MaxWidth(m.width).MaxHeight(m.height).Render(full)
}

// detailPaneView composes the detail pane: a title line + the scrollable body.
func (m Model) detailPaneView() string {
	titleStyle := dimStyle
	if m.focus == focusDetail {
		titleStyle = selectedStyle
	}
	content := lipgloss.JoinVertical(lipgloss.Left, titleStyle.Render(m.detailTitle()), m.detail.View())
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

func (m Model) footer() string {
	hints := "j/k move · l/⏎ detail · / filter · tab focus · r refresh · q quit"
	if m.focus == focusDetail {
		hints = "j/k scroll · g/G top/bottom · h/esc back · q quit"
	}
	if len(m.problems) > 0 {
		hints = fmt.Sprintf("! %d unreadable · ", len(m.problems)) + hints
	}
	// Truncate to the terminal width — otherwise JoinVertical pads every pane row
	// out to the footer's width and the whole frame overflows the terminal.
	return dim(truncate(hints, m.width))
}

func (m Model) emptyView() string {
	msg := "No active tasks.\n\nCreate one:  tskflwctl task new \"Title\" --epic <id>"
	if len(m.problems) > 0 {
		msg += fmt.Sprintf("\n\n! %d unreadable file(s) — run `tskflwctl lint`", len(m.problems))
	}
	return msg
}
