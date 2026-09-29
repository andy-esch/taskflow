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

// detailDirectionMenu resolves a semantic directional fan-in/fan-out without
// guessing among equally direct graph neighbors. It is shell-owned and reusable:
// the detail supplies canonical tasks while the menu owns only transient cursor
// and overlay presentation state.
type detailDirectionMenu struct {
	active          bool
	label           string
	tasks           []core.LoadedRecord[domain.Task]
	cursor          int
	loadedKey       string
	originSelection string
	dx              int
	dy              int
}

func (m *detailDirectionMenu) open(
	label string,
	tasks []core.LoadedRecord[domain.Task],
	loadedKey, originSelection string,
	dx, dy int,
) {
	*m = detailDirectionMenu{
		active: true, label: label, tasks: append([]core.LoadedRecord[domain.Task](nil), tasks...),
		loadedKey: loadedKey, originSelection: originSelection, dx: dx, dy: dy,
	}
}

func (m *detailDirectionMenu) close() { *m = detailDirectionMenu{} }

func (m *detailDirectionMenu) move(delta int) {
	if count := len(m.tasks); count > 0 {
		m.cursor = ((m.cursor+delta)%count + count) % count
	}
}

func (m detailDirectionMenu) selected() (core.LoadedRecord[domain.Task], bool) {
	if m.cursor < 0 || m.cursor >= len(m.tasks) {
		return core.LoadedRecord[domain.Task]{}, false
	}
	return m.tasks[m.cursor], true
}

func (m detailDirectionMenu) view(s *styles, maxW, maxH int) string {
	position := fmt.Sprintf(" · %d/%d", m.cursor+1, len(m.tasks))
	var b strings.Builder
	b.WriteString(s.actionHeading.Render("choose " + truncate(m.label, max(maxW-8-ansi.StringWidth(position), 12)) + position))
	b.WriteString("\n\n")
	refs := make([]entityRef, 0, len(m.tasks))
	for _, record := range m.tasks {
		refs = append(refs, entityRef{key: record.Source.ID, label: record.Value.Slug})
	}
	hints := duplicateIdentityHints(refs)
	start, end := visibleTaskPickerRange(len(m.tasks), m.cursor, maxH)
	for index := start; index < end; index++ {
		record := m.tasks[index]
		task := record.Value
		token := theme.Status(task.Status)
		label := s.fg(token.Color, token.Glyph) + " " +
			truncate(labelWithIdentityHint(task.Slug, hints[record.Source.ID]), max(maxW-10, 12))
		if index == m.cursor {
			b.WriteString(s.selected.Render("› ") + label + "\n")
		} else {
			b.WriteString("  " + label + "\n")
		}
	}
	box := s.actionBorder.Render(strings.TrimRight(b.String(), "\n"))
	hint := s.dim("↑↓/jk select · ⏎ move · esc cancel")
	return clampBox(lipgloss.JoinVertical(lipgloss.Center, box, hint), maxW, maxH)
}

func (m *Model) handleDetailDirectionKey(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case msg.String() == "j" || msg.String() == "down":
		m.direction.move(1)
	case msg.String() == "k" || msg.String() == "up":
		m.direction.move(-1)
	case msg.String() == "enter":
		menu := m.direction
		task, ok := m.direction.selected()
		m.direction.close()
		if !ok || !m.detail.directionTargetStillValid(menu, task.Source.ID) ||
			!m.detail.selectDetailTask(task.Source.ID) {
			m.flash, m.flashErr = "directional target is no longer available", true
		}
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit):
		m.direction.close()
	}
	return nil
}
