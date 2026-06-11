package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// taskItem adapts a domain.Task to a bubbles/list item.
type taskItem struct{ t domain.Task }

func (i taskItem) FilterValue() string { return i.t.Slug + " " + i.t.Description }

func taskDate(t domain.Task) string {
	if t.Updated != "" {
		return t.Updated
	}
	return t.Created
}

// taskDelegate renders one task row: colored status glyph, a ⚠ if misfiled, the
// slug, and a dim relative date — truncated to fit the list width.
type taskDelegate struct{}

func (taskDelegate) Height() int                         { return 1 }
func (taskDelegate) Spacing() int                        { return 0 }
func (taskDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (taskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(taskItem)
	if !ok {
		return
	}
	tok := theme.Status(it.t.Status)
	marker := " "
	if it.t.Misfiled() {
		marker = fg(theme.ColorYellow, "⚠")
	}
	date := theme.RelativeDate(taskDate(it.t))

	// Reserve: cursor(2) + glyph(1) + marker(1) + 3 spaces + date(≤10).
	slugW := m.Width() - 2 - 2 - 3 - 10
	if slugW < 8 {
		slugW = 8
	}
	slug := fmt.Sprintf("%-*s", slugW, truncate(it.t.Slug, slugW))

	row := fmt.Sprintf("%s %s %s  %s", fg(tok.Color, tok.Glyph), marker, slug, dim(date))
	if index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render("› "+row))
		return
	}
	fmt.Fprint(w, "  "+row)
}
