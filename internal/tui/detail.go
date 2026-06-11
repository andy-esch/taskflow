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

// detailContent is an entity-agnostic right-pane payload: a title for the pane
// header and a width-aware renderer (wrapping happens in the model, not the load
// Cmd, so it re-wraps on resize). Tasks, epics, and audits each implement it.
type detailContent interface {
	Title() string
	Render(width int) string
}

// detailPane is the right pane: a scrollable view of the selected item's detail.
type detailPane struct {
	vp         viewport.Model
	title      string
	width      int
	content    detailContent // current payload (re-rendered on resize); nil for errors
	errMsg     string
	loading    bool
	hasContent bool
}

func newDetailPane() detailPane { return detailPane{vp: viewport.New(0, 0)} }

func (d *detailPane) SetSize(w, h int) {
	d.width = w
	d.vp.Width = w
	d.vp.Height = h
	// Re-wrap the current payload to the new width (keeps the body from clipping
	// when the pane grows/shrinks).
	switch {
	case d.content != nil:
		d.vp.SetContent(d.content.Render(w))
	case d.errMsg != "":
		d.vp.SetContent(fg(theme.ColorRed, "⚠ "+d.errMsg))
	}
}

func (d *detailPane) SetContent(c detailContent) {
	d.content = c
	d.errMsg = ""
	d.title = c.Title()
	d.vp.SetContent(c.Render(d.width))
	d.vp.GotoTop()
	d.hasContent = true
	d.loading = false
}

// SetError shows a per-item load error in the pane (keeps the browser alive).
func (d *detailPane) SetError(title, msg string) {
	d.content = nil
	d.errMsg = msg
	d.title = title
	d.vp.SetContent(fg(theme.ColorRed, "⚠ "+msg))
	d.vp.GotoTop()
	d.hasContent = true
	d.loading = false
}

// clear resets the pane to its loading state — used when switching tabs so the
// previous entity's detail doesn't linger while the new selection loads.
func (d *detailPane) clear() {
	d.content = nil
	d.errMsg = ""
	d.title = ""
	d.hasContent = false
	d.loading = true
}

func (d detailPane) View() string {
	switch {
	case d.loading && !d.hasContent:
		return dim("loading…")
	case !d.hasContent:
		return dim("(nothing selected)")
	}
	return d.vp.View()
}

func detailField(b *strings.Builder, label, val string) {
	if val == "" {
		return
	}
	fmt.Fprintf(b, "%s %s\n", dimStyle.Render(fmt.Sprintf("%-9s", label+":")), val)
}

func wrap(s string, width int) string {
	if width > 0 {
		return lipgloss.NewStyle().Width(width).Render(s)
	}
	return s
}

// --- task detail ---

type taskDetail struct {
	t    domain.Task
	body string
}

func (d taskDetail) Title() string       { return d.t.Slug }
func (d taskDetail) Render(w int) string { return renderTaskDetail(d.t, d.body, w) }

// renderTaskDetail formats a task's frontmatter fields + markdown body, wrapped
// to width. Body is plain text for now (glamour is a later sprint).
func renderTaskDetail(t domain.Task, body string, width int) string {
	var b strings.Builder
	detailField(&b, "status", statusText(t.Status))
	detailField(&b, "epic", t.Epic)
	detailField(&b, "priority", priorityText(t.Priority))
	if t.Tier != 0 {
		detailField(&b, "tier", strconv.Itoa(t.Tier))
	}
	if len(t.Tags) > 0 {
		detailField(&b, "tags", strings.Join(t.Tags, ", "))
	}
	if t.Updated != "" {
		detailField(&b, "updated", fmt.Sprintf("%s (%s)", t.Updated, theme.RelativeDate(t.Updated)))
	}
	if t.Misfiled() {
		detailField(&b, "⚠", fg(theme.ColorYellow, fmt.Sprintf("frontmatter says %q (folder wins)", t.Declared)))
	}
	b.WriteString("\n")
	b.WriteString(body)
	return wrap(b.String(), width)
}

// --- epic detail ---

type epicDetail struct {
	e     domain.Epic
	tasks []domain.Task
	body  string
}

func (d epicDetail) Title() string       { return d.e.ID }
func (d epicDetail) Render(w int) string { return renderEpicDetail(d.e, d.tasks, d.body, w) }

func renderEpicDetail(e domain.Epic, tasks []domain.Task, body string, width int) string {
	var b strings.Builder
	detailField(&b, "epic", e.ID)
	detailField(&b, "status", e.Status)
	detailField(&b, "priority", priorityText(e.Priority))
	if len(e.Tags) > 0 {
		detailField(&b, "tags", strings.Join(e.Tags, ", "))
	}
	done := 0
	for _, t := range tasks {
		if t.Status == domain.StatusCompleted {
			done++
		}
	}
	pct := 0
	if len(tasks) > 0 {
		pct = done * 100 / len(tasks)
	}
	detailField(&b, "progress", fmt.Sprintf("%s %s  %d/%d",
		miniBar(pct, 12), fg(theme.Percent(pct), fmt.Sprintf("%d%%", pct)), done, len(tasks)))
	if len(tasks) > 0 {
		b.WriteString("\n")
		for _, t := range tasks {
			tok := theme.Status(t.Status)
			fmt.Fprintf(&b, "  %s %s\n", fg(tok.Color, tok.Glyph), t.Slug)
		}
	}
	b.WriteString("\n")
	b.WriteString(body)
	return wrap(b.String(), width)
}

// --- audit detail ---

type auditDetail struct {
	a    domain.Audit
	body string
}

func (d auditDetail) Title() string       { return d.a.Slug }
func (d auditDetail) Render(w int) string { return renderAuditDetail(d.a, d.body, w) }

func renderAuditDetail(a domain.Audit, body string, width int) string {
	var b strings.Builder
	detailField(&b, "audit", a.Slug)
	detailField(&b, "bucket", fg(theme.Bucket(a.Bucket), string(a.Bucket)))
	detailField(&b, "area", a.Area)
	detailField(&b, "date", a.Date)
	detailField(&b, "findings", fmt.Sprintf("%d open / %d total", a.OpenFindings, a.Findings))
	b.WriteString("\n")
	b.WriteString(body)
	return wrap(b.String(), width)
}
