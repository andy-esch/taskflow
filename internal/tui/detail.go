package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

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
	styled     string        // last rendered (styled) content, kept so find can re-highlight
	errMsg     string
	loading    bool
	hasContent bool
	find       finder
}

func newDetailPane() detailPane { return detailPane{vp: viewport.New(0, 0), find: newFinder()} }

func (d *detailPane) SetSize(w, h int) {
	d.width = w
	d.vp.Width = w
	d.vp.Height = h
	// Re-wrap the current payload to the new width (keeps the body from clipping
	// when the pane grows/shrinks), then re-apply any active find highlight.
	switch {
	case d.content != nil:
		d.styled = d.content.Render(w)
		d.refreshFind()
	case d.errMsg != "":
		d.vp.SetContent(fg(theme.ColorRed, "⚠ "+d.errMsg))
	}
}

func (d *detailPane) SetContent(c detailContent) {
	d.content = c
	d.errMsg = ""
	d.title = c.Title()
	d.styled = c.Render(d.width)
	d.refreshFind() // recompute matches for the new content (find persists across items)
	d.vp.GotoTop()
	d.hasContent = true
	d.loading = false
}

// SetError shows a per-item load error in the pane (keeps the browser alive).
func (d *detailPane) SetError(title, msg string) {
	d.content = nil
	d.styled = ""
	d.resetFind()
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
	d.styled = ""
	d.resetFind()
	d.errMsg = ""
	d.title = ""
	d.hasContent = false
	d.loading = true
}

// showEmpty settles the pane on "(nothing selected)" — used when a `/` filter
// narrows the list to zero matches, so the previously selected item's detail
// doesn't linger.
func (d *detailPane) showEmpty() {
	d.content = nil
	d.styled = ""
	d.resetFind()
	d.errMsg = ""
	d.title = ""
	d.hasContent = false
	d.loading = false
}

// --- find-in-body (vim-like `/` + n/N) ---

// finding reports whether the find query is being typed (so the model routes
// keys to the find input); active reports whether a query is applied.
func (d detailPane) finding() bool    { return d.find.typing }
func (d detailPane) findActive() bool { return d.find.active() }

// startFind opens the find input over the current content.
func (d *detailPane) startFind() tea.Cmd {
	d.find.typing = true
	d.find.input.SetValue("")
	return d.find.input.Focus()
}

// updateFind feeds a key to the find input: Esc cancels, Enter applies the query
// (computing matches + jumping to the first), anything else edits the query.
func (d *detailPane) updateFind(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, keys.Back):
		d.find.typing = false
		d.find.input.Blur()
		return nil
	case msg.Type == tea.KeyEnter:
		d.find.typing = false
		d.find.input.Blur()
		d.applyQuery(d.find.input.Value())
		return nil
	}
	var cmd tea.Cmd
	d.find.input, cmd = d.find.input.Update(msg)
	return cmd
}

// applyQuery sets the active query, recomputes matches, and scrolls to the first.
func (d *detailPane) applyQuery(q string) {
	d.find.query = strings.TrimSpace(q)
	d.find.cur = 0
	d.refreshFind()
	d.scrollToCurrent()
}

// findNext moves the focused match by dir (wrapping) and scrolls to it.
func (d *detailPane) findNext(dir int) {
	if len(d.find.lines) == 0 {
		return
	}
	n := len(d.find.lines)
	d.find.cur = ((d.find.cur+dir)%n + n) % n
	d.refreshFind()
	d.scrollToCurrent()
}

// clearFind drops the active search and removes the highlight.
func (d *detailPane) clearFind() {
	d.resetFind()
	d.refreshFind()
}

func (d *detailPane) resetFind() {
	d.find.typing = false
	d.find.query = ""
	d.find.lines = nil
	d.find.cur = 0
	d.find.input.Blur()
}

// refreshFind recomputes match line indices for the current query and re-renders
// the viewport content with matches highlighted (current match stronger). Matched
// lines are rebuilt from stripped text so a highlight never splits an existing
// escape; unmatched lines keep their original colors.
func (d *detailPane) refreshFind() {
	if d.find.query == "" {
		d.vp.SetContent(d.styled)
		return
	}
	styled := strings.Split(d.styled, "\n")
	plain := strings.Split(ansi.Strip(d.styled), "\n")
	q := strings.ToLower(d.find.query)
	d.find.lines = d.find.lines[:0]
	for i, pl := range plain {
		if strings.Contains(strings.ToLower(pl), q) {
			d.find.lines = append(d.find.lines, i)
		}
	}
	if d.find.cur >= len(d.find.lines) {
		d.find.cur = 0
	}
	for k, li := range d.find.lines {
		style := findMatch
		if k == d.find.cur {
			style = findCurrent
		}
		styled[li] = highlightOccurrences(plain[li], d.find.query, style)
	}
	d.vp.SetContent(strings.Join(styled, "\n"))
}

// scrollToCurrent brings the focused match into view (a couple of lines of lead).
func (d *detailPane) scrollToCurrent() {
	if len(d.find.lines) == 0 {
		return
	}
	if target := d.find.lines[d.find.cur] - 2; target > 0 {
		d.vp.SetYOffset(target)
	} else {
		d.vp.GotoTop()
	}
}

// findStatus is the footer line shown while finding: the live input, or the
// applied query with a match position and the nav hint.
func (d detailPane) findStatus() string {
	if d.find.typing {
		return d.find.input.View()
	}
	pos := 0
	if len(d.find.lines) > 0 {
		pos = d.find.cur + 1
	}
	return dim(fmt.Sprintf("/%s  [%d/%d]  n/N next/prev · esc clear", d.find.query, pos, len(d.find.lines)))
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
