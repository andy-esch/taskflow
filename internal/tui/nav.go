package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// Cross-link navigation (S6): follow structured references — a task's `epic:`
// field, an epic's task list, or a Thread's bounded task nodes — with `f`, and
// walk back with ctrl+o (vim jumplist style). Only *structured* references for
// now; body [[wikilinks]] and the peek-overlay are deferred (see the task file).

// navLoc is one entry in the follow back-stack: where the user was when they
// followed a reference.
type navLoc struct {
	kind entityKind
	ref  entityRef
}

// followMenu is the reference picker for an entity with several outgoing task
// links (an epic or Thread). Modal like the action menu: the model routes every
// key to it while active and floats it over the body.
type followMenu struct {
	active      bool
	sourceLabel string        // the entity whose references are listed
	tasks       []domain.Task // the rows
	cursor      int
}

func (f *followMenu) open(sourceLabel string, tasks []domain.Task) {
	*f = followMenu{active: true, sourceLabel: sourceLabel, tasks: tasks}
}

func (f *followMenu) selectTask(taskID string) {
	for index, task := range f.tasks {
		if task.CanonicalID() == taskID {
			f.cursor = index
			return
		}
	}
}

func (f *followMenu) close() { f.active = false }

func (f *followMenu) move(d int) {
	if n := len(f.tasks); n > 0 {
		f.cursor = ((f.cursor+d)%n + n) % n
	}
}

func (f followMenu) selected() domain.Task { return f.tasks[f.cursor] }

// view renders the picker as a centered box + hint line for overlay().
func (f followMenu) view(s *styles, maxW, maxH int) string {
	var b strings.Builder
	position := ""
	if len(f.tasks) > 0 {
		position = fmt.Sprintf(" · %d/%d", f.cursor+1, len(f.tasks))
	}
	b.WriteString(s.actionHeading.Render("follow " + truncate(f.sourceLabel, max(maxW-8-ansi.StringWidth(position), 12)) + position))
	b.WriteString("\n\n")
	refs := make([]entityRef, 0, len(f.tasks))
	for _, task := range f.tasks {
		refs = append(refs, entityRef{key: task.CanonicalID(), label: task.Slug})
	}
	hints := duplicateIdentityHints(refs)
	start, end := f.visibleRange(maxH)
	for i := start; i < end; i++ {
		t := f.tasks[i]
		tok := theme.Status(t.Status)
		label := s.fg(tok.Color, tok.Glyph) + " " + truncate(labelWithIdentityHint(t.Slug, hints[t.CanonicalID()]), max(maxW-10, 12))
		if i == f.cursor {
			b.WriteString(s.selected.Render("› ") + label + "\n")
		} else {
			b.WriteString("  " + label + "\n")
		}
	}
	box := s.actionBorder.Render(strings.TrimRight(b.String(), "\n"))
	hint := s.dim("↑↓/jk select · ⏎ follow · esc cancel")
	return clampBox(lipgloss.JoinVertical(lipgloss.Center, box, hint), maxW, maxH)
}

// visibleRange keeps the selected reference on screen in a short terminal. The
// box consumes two border rows, a heading plus spacer, and one hint row; every
// remaining row can hold exactly one task. Clamping is presentation-only and
// never changes the underlying task order or cursor.
func (f followMenu) visibleRange(maxH int) (int, int) {
	count := len(f.tasks)
	if count == 0 || maxH <= 0 {
		return 0, count
	}
	visible := max(1, maxH-5)
	if visible >= count {
		return 0, count
	}
	start := f.cursor - visible/2
	if start < 0 {
		start = 0
	}
	if end := start + visible; end > count {
		start = count - visible
	}
	return start, start + visible
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
		return m.jumpTo(entityTasks, entityRef{key: target.CanonicalID(), label: target.Slug})
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit):
		m.follow.close()
	}
	return nil
}

// followSelected follows the selected item's outgoing reference: a task jumps
// to its epic; an epic opens the picker over its tasks; a Thread opens the same
// picker over member and immediate-external-gate tasks from its loaded graph
// projection. Audits have no structured references (yet).
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
		return m, m.jumpTo(entityEpics, entityRef{key: task.Epic, label: task.Epic})
	case entityEpics:
		ref := m.selectedRef()
		if ref.empty() {
			return m, nil
		}
		// The epic's task list rides in the already-loaded detail content (the
		// pane is stale-guarded, so a matching ID means current data).
		ed, ok := m.detail.content.(epicDetail)
		if !ok || ed.es.Epic.ID != ref.key {
			m.flash, m.flashErr = "references still loading…", true
			return m, nil
		}
		if len(ed.tasks) == 0 {
			m.flash, m.flashErr = fmt.Sprintf("%s has no tasks", ref.label), true
			return m, nil
		}
		m.follow.open(ref.label, ed.tasks)
		return m, nil
	case entityThreads:
		ref := m.selectedRef()
		if ref.empty() {
			return m, nil
		}
		detail, ok := m.detail.content.(threadDetail)
		if !ok || detail.projection.View.Thread.CanonicalID() != ref.key {
			m.flash, m.flashErr = "references still loading…", true
			return m, nil
		}
		tasks := threadFollowTasks(detail.projection)
		if len(tasks) == 0 {
			m.flash, m.flashErr = fmt.Sprintf("%s has no readable task references", ref.label), true
			return m, nil
		}
		m.follow.open(ref.label, tasks)
		m.follow.selectTask(detail.detailSelectionKey())
		return m, nil
	default:
		m.flash, m.flashErr = "no linked entities here", true
		return m, nil
	}
}

// openDetailSelection follows a structured row selected inside a detail
// presentation. The presentation supplies only a canonical target; the shell
// retains ownership of history, tab switching, lifecycle-view widening, and
// asynchronous loading exactly as it does for the `f` picker.
func (m Model) openDetailSelection() (tea.Model, tea.Cmd) {
	kind, ref, ok := m.detail.selectionTarget()
	if !ok {
		return m, nil
	}
	m.pushLoc()
	return m, m.jumpTo(kind, ref)
}

// threadFollowTasks retains the stable node order supplied by the projection
// while recovering the semantic task values carried by its Thread view. Missing
// or unreadable nodes remain visible in topology diagnostics but are not offered
// as navigation targets that cannot resolve on the task tab.
func threadFollowTasks(projection core.ThreadGraphProjection) []domain.Task {
	byID := make(map[string]domain.Task, len(projection.View.Members)+len(projection.View.ExternalGates))
	for _, member := range projection.View.Members {
		if member.Task.Slug != "" {
			byID[member.State.TaskID] = member.Task
		}
	}
	for _, gate := range projection.View.ExternalGates {
		if gate.Task.Slug != "" {
			byID[gate.State.TaskID] = gate.Task
		}
	}
	tasks := make([]domain.Task, 0, len(byID))
	for _, node := range projection.Nodes {
		if task, ok := byID[node.TaskID]; ok {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// navStackMax bounds the back-stack so a long bounce between an epic and its
// tasks can't grow it for the whole session (a vim-jumplist-style cap).
const navStackMax = 50

// pushLoc records the current position on the back-stack (no-op on an empty
// selection — there is nothing to come back to). It skips a push identical to the
// current top (re-following the same place adds no useful history) and caps the
// stack at navStackMax, dropping the oldest entries.
func (m *Model) pushLoc() {
	ref := m.selectedRef()
	if ref.empty() {
		return
	}
	loc := navLoc{kind: m.cur().kind, ref: ref}
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
	return m, m.jumpTo(loc.kind, loc.ref)
}

// jumpTo makes (kind, canonical ref) the active selection: switches the tab, clears any
// applied filter (a jump is explicit navigation — a filter must not hide the
// target), and selects the row. A task hidden by the current status view
// escalates the view to :all and reloads with the cursor restore pending; a
// genuinely missing target flashes instead of crashing.
func (m *Model) jumpTo(kind entityKind, ref entityRef) tea.Cmd {
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
		// An explicit navigation target may be outside an axis-bearing tab's
		// default view (for example, a completed Thread member while Tasks shows
		// only working states). Nothing has been browsed on an unloaded tab yet, so
		// load :all immediately rather than landing on an arbitrary visible row and
		// requiring a second read. Loaded tabs retain their view unless the target
		// is actually absent below.
		if len(tab.viewAxis) > 0 {
			tab.statusView = "all"
		}
		return tab.reload(m.svc, ref)
	}
	if tab.selectByKey(ref.key) {
		return m.refreshDetail()
	}
	if tab.viewAxis != nil && tab.statusView != "all" {
		// A non-default view hides rows the target may live in — the task working set
		// hides archived tasks; the epics default hides retired/deprecated buckets —
		// so widen to :all rather than fail (the chip shows view:all afterwards).
		tab.statusView = "all"
		m.flash, m.flashErr = fmt.Sprintf("showing :all to reach %s", ref.label), false
		return tab.reload(m.svc, ref)
	}
	m.flash, m.flashErr = fmt.Sprintf("%s not found", ref.label), true
	return m.refreshDetail()
}
