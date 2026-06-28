package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// Cross-link navigation (S6): follow structured references — a task's `epic:`
// field, an epic's task list — with `f`, and walk back with ctrl+o (vim
// jumplist style). Only *structured* references for now; body [[wikilinks]]
// and the peek-overlay are deferred (see the task file).

// navLoc is one entry in the follow back-stack: where the user was when they
// followed a reference.
type navLoc struct {
	kind entityKind
	id   string
}

// followMenu is the reference picker for an entity with several outgoing links
// (an epic's tasks). Modal like the action menu: the model routes every key to
// it while active and floats it over the body.
type followMenu struct {
	active bool
	epicID string        // the epic whose references are listed
	tasks  []domain.Task // the rows
	cursor int
}

func (f *followMenu) open(epicID string, tasks []domain.Task) {
	*f = followMenu{active: true, epicID: epicID, tasks: tasks}
}

func (f *followMenu) close() { f.active = false }

func (f *followMenu) move(d int) {
	if n := len(f.tasks); n > 0 {
		f.cursor = ((f.cursor+d)%n + n) % n
	}
}

func (f followMenu) selected() domain.Task { return f.tasks[f.cursor] }

// view renders the picker as a centered box + hint line for overlay().
func (f followMenu) view(maxW, maxH int) string {
	var b strings.Builder
	b.WriteString(actionHeading.Render("follow " + truncate(f.epicID, max(maxW-8, 12))))
	b.WriteString("\n\n")
	for i, t := range f.tasks {
		tok := theme.Status(t.Status)
		label := fg(tok.Color, tok.Glyph) + " " + truncate(t.Slug, max(maxW-10, 12))
		if i == f.cursor {
			b.WriteString(selectedStyle.Render("› ") + label + "\n")
		} else {
			b.WriteString("  " + label + "\n")
		}
	}
	box := actionBorder.Render(strings.TrimRight(b.String(), "\n"))
	hint := dim("↑↓/jk select · ⏎ follow · esc cancel")
	return clampBox(lipgloss.JoinVertical(lipgloss.Center, box, hint), maxW, maxH)
}

// handleFollowKey drives the picker while it's open. It mutates the model copy
// directly (the modal loop passes &m) and returns the cmd; ForceQuit is handled by
// handleKey's preamble, ahead of the modal loop.
func (m *Model) handleFollowKey(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case msg.String() == "j" || msg.String() == "down":
		m.follow.move(1)
	case msg.String() == "k" || msg.String() == "up":
		m.follow.move(-1)
	case msg.String() == "enter":
		target := m.follow.selected()
		m.follow.close()
		m.pushLoc()
		return m.jumpTo(entityTasks, target.Slug)
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit):
		m.follow.close()
	}
	return nil
}

// followSelected follows the selected item's outgoing reference: a task jumps
// to its epic; an epic opens the picker over its tasks. Audits have no
// structured references (yet).
func (m Model) followSelected() (tea.Model, tea.Cmd) {
	switch t := m.cur(); t.kind {
	case entityTasks:
		task, ok := m.selectedTask()
		if !ok {
			return m, nil
		}
		if task.Epic == "" {
			m.flash, m.flashErr = fmt.Sprintf("%s has no epic reference", task.Slug), true
			return m, nil
		}
		m.pushLoc()
		return m, m.jumpTo(entityEpics, task.Epic)
	case entityEpics:
		id := m.selectedID()
		if id == "" {
			return m, nil
		}
		// The epic's task list rides in the already-loaded detail content (the
		// pane is stale-guarded, so a matching ID means current data).
		ed, ok := m.detail.content.(epicDetail)
		if !ok || ed.es.Epic.ID != id {
			m.flash, m.flashErr = "references still loading…", true
			return m, nil
		}
		if len(ed.tasks) == 0 {
			m.flash, m.flashErr = fmt.Sprintf("%s has no tasks", id), true
			return m, nil
		}
		m.follow.open(id, ed.tasks)
		return m, nil
	default:
		m.flash, m.flashErr = "no linked entities here", true
		return m, nil
	}
}

// navStackMax bounds the back-stack so a long bounce between an epic and its
// tasks can't grow it for the whole session (a vim-jumplist-style cap).
const navStackMax = 50

// pushLoc records the current position on the back-stack (no-op on an empty
// selection — there is nothing to come back to). It skips a push identical to the
// current top (re-following the same place adds no useful history) and caps the
// stack at navStackMax, dropping the oldest entries.
func (m *Model) pushLoc() {
	id := m.selectedID()
	if id == "" {
		return
	}
	loc := navLoc{kind: m.cur().kind, id: id}
	if n := len(m.navStack); n > 0 && m.navStack[n-1] == loc {
		return
	}
	m.navStack = append(m.navStack, loc)
	if over := len(m.navStack) - navStackMax; over > 0 {
		m.navStack = m.navStack[over:]
	}
}

// navBack pops the back-stack and returns to where the last follow happened.
func (m Model) navBack() (tea.Model, tea.Cmd) {
	n := len(m.navStack)
	if n == 0 {
		m.flash, m.flashErr = "nothing to go back to", true
		return m, nil
	}
	loc := m.navStack[n-1]
	m.navStack = m.navStack[:n-1]
	return m, m.jumpTo(loc.kind, loc.id)
}

// jumpTo makes (kind, id) the active selection: switches the tab, clears any
// applied filter (a jump is explicit navigation — a filter must not hide the
// target), and selects the row. A task hidden by the current status view
// escalates the view to :all and reloads with the cursor restore pending; a
// genuinely missing target flashes instead of crashing.
func (m *Model) jumpTo(kind entityKind, id string) tea.Cmd {
	i := indexOfKind(m.tabs, kind)
	if i < 0 {
		return nil
	}
	if i != m.active || m.onDash {
		m.exitDashboard(i) // a jump always lands on an entity tab
	}
	tab := m.tabs[i]
	tab.list.ResetFilter()
	if !tab.loaded {
		return tab.reload(m.svc, id)
	}
	if tab.selectByID(id) {
		return m.refreshDetail()
	}
	if tab.viewAxis != nil && tab.statusView != "all" {
		// A non-default view hides rows the target may live in — the task working set
		// hides archived tasks; the epics default hides retired/deprecated buckets —
		// so widen to :all rather than fail (the chip shows view:all afterwards).
		tab.statusView = "all"
		m.flash, m.flashErr = fmt.Sprintf("showing :all to reach %s", id), false
		return tab.reload(m.svc, id)
	}
	m.flash, m.flashErr = fmt.Sprintf("%s not found", id), true
	return m.refreshDetail()
}
