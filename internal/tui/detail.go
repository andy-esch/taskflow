package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// detailPane is the right pane: a scrollable view of one task's metadata + body.
type detailPane struct {
	vp         viewport.Model
	title      string
	width      int
	loading    bool
	hasContent bool
}

func newDetailPane() detailPane { return detailPane{vp: viewport.New(0, 0)} }

func (d *detailPane) SetSize(w, h int) {
	d.width = w
	d.vp.Width = w
	d.vp.Height = h
}

func (d *detailPane) SetContent(t domain.Task, body string) {
	d.title = t.Slug
	d.vp.SetContent(renderDetail(t, body, d.width))
	d.vp.GotoTop()
	d.hasContent = true
	d.loading = false
}

// SetError shows a per-task load error in the pane (keeps the browser alive).
func (d *detailPane) SetError(slug, msg string) {
	d.title = slug
	d.vp.SetContent(fg(theme.ColorRed, "⚠ "+msg))
	d.vp.GotoTop()
	d.hasContent = true
	d.loading = false
}

func (d detailPane) View() string {
	switch {
	case d.loading && !d.hasContent:
		return dim("loading…")
	case !d.hasContent:
		return dim("(select a task)")
	}
	return d.vp.View()
}

var detailLabel = lipgloss.NewStyle().Faint(true)

// renderDetail formats a task's frontmatter fields + markdown body, wrapped to
// width. Body is plain text for now (glamour is a later sprint).
func renderDetail(t domain.Task, body string, width int) string {
	var b strings.Builder
	field := func(label, val string) {
		if val == "" {
			return
		}
		fmt.Fprintf(&b, "%s %s\n", detailLabel.Render(fmt.Sprintf("%-9s", label+":")), val)
	}
	field("status", statusText(t.Status))
	field("epic", t.Epic)
	field("priority", priorityText(t.Priority))
	if t.Tier != 0 {
		field("tier", strconv.Itoa(t.Tier))
	}
	if len(t.Tags) > 0 {
		field("tags", strings.Join(t.Tags, ", "))
	}
	if t.Updated != "" {
		field("updated", fmt.Sprintf("%s (%s)", t.Updated, theme.RelativeDate(t.Updated)))
	}
	if t.Misfiled() {
		field("⚠", fg(theme.ColorYellow, fmt.Sprintf("frontmatter says %q (folder wins)", t.Declared)))
	}
	b.WriteString("\n")
	b.WriteString(body)
	if width > 0 {
		return lipgloss.NewStyle().Width(width).Render(b.String())
	}
	return b.String()
}
