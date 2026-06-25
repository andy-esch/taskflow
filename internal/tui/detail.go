package tui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// detailContent is an entity-agnostic right-pane payload, split so the pane can
// render the markdown body two ways (raw / glamour) while the styled field block
// above it is rendered once. meta is the non-markdown header (frontmatter fields,
// and for epics the task list); rawBody is the markdown body. Wrapping happens in
// the pane (not the load Cmd) so it re-wraps on resize.
type detailContent interface {
	Title() string
	Path() string // the entity's on-disk file path (for the clickable detail title)
	meta(width int) string
	rawBody() string
}

// detailPane is the right pane: a scrollable view of the selected item's detail.
// The body renders two ways — glamour (pretty) and raw — both cached so the `R`
// toggle just swaps which is shown (no glamour recompile). pretty persists across
// selections and tabs (the pane is a single Model field).
type detailPane struct {
	vp           viewport.Model
	title        string
	width        int
	content      detailContent // current payload (re-rendered on resize); nil for errors
	pretty       bool          // glamour body (true) vs raw markdown (false)
	rawStyled    string        // meta + raw body (cached)
	prettyStyled string        // meta + glamour body (cached)
	styled       string        // the active composition, kept so find can re-highlight
	errMsg       string
	loading      bool
	hasContent   bool
	find         finder

	glamStyle string // glamour standard-style for the terminal background (set at startup)

	glam           *glamour.TermRenderer // cached renderer, rebuilt only when width/style changes
	glamW          int                   // the width glam was built for
	glamStyleBuilt string                // the style glam was built for
}

// prettyBody renders md with the pane's cached glamour renderer, rebuilding it
// only when the width changed — compiling a renderer is the expensive part, so a
// plain selection change (same width) reuses it. Falls back to plain wrapped text
// on any error.
func (d *detailPane) prettyBody(md string) string {
	if strings.TrimSpace(md) == "" {
		return ""
	}
	if d.glam == nil || d.glamW != d.width || d.glamStyleBuilt != d.glamStyle {
		r, err := newGlamourRenderer(d.width, d.glamStyle)
		if err != nil {
			return wrap(md, d.width)
		}
		d.glam, d.glamW, d.glamStyleBuilt = r, d.width, d.glamStyle
	}
	out, ok := renderMarkdown(d.glam, md)
	if !ok {
		return wrap(md, d.width)
	}
	return out
}

func newDetailPane(glamStyle string) detailPane {
	return detailPane{vp: viewport.New(), find: newFinder(), pretty: true, glamStyle: glamStyle}
}

// render rebuilds both body compositions at the current width and points styled at
// the active mode. Called on content/size change (Update) — NEVER in View, since
// glamourBody compiles a renderer.
func (d *detailPane) render() {
	if d.content == nil {
		d.rawStyled, d.prettyStyled, d.styled = "", "", ""
		return
	}
	meta := d.content.meta(d.width)
	body := d.content.rawBody()
	d.rawStyled = joinDetail(meta, wrap(body, d.width))
	d.prettyStyled = joinDetail(meta, d.prettyBody(body))
	d.styled = d.activeStyled()
}

func (d detailPane) activeStyled() string {
	if d.pretty {
		return d.prettyStyled
	}
	return d.rawStyled
}

// toggleMode flips raw ⇄ pretty. Both are pre-rendered, so this only swaps which is
// shown and re-applies any find — no glamour recompile.
func (d *detailPane) toggleMode() {
	d.pretty = !d.pretty
	if d.content == nil {
		return
	}
	d.styled = d.activeStyled()
	d.refreshFind()
}

func joinDetail(meta, body string) string {
	switch {
	case meta == "":
		return body
	case body == "":
		return meta
	default:
		return meta + "\n\n" + body
	}
}

func (d *detailPane) SetSize(w, h int) {
	widthChanged := w != d.width
	d.width = w
	d.vp.SetWidth(w)
	d.vp.SetHeight(h)
	switch {
	case d.content != nil:
		// Body wrap (and glamour) depend on width — re-render only when it changed
		// (a height-only resize must not re-run glamour).
		if widthChanged || d.styled == "" {
			d.render()
		}
		d.refreshFind()
	case d.errMsg != "":
		d.vp.SetContent(fg(theme.ColorRed, "⚠ "+d.errMsg))
	}
}

func (d *detailPane) SetContent(c detailContent) {
	// A live-reload refresh of the item already on screen keeps the scroll
	// position (the viewport clamps it to the new content); only a *different*
	// item snaps back to the top. Otherwise every external write under the
	// watched tree would yank the body you're reading back to line one.
	sameItem := d.hasContent && d.content != nil && d.title == c.Title()
	offset := d.vp.YOffset()
	d.content = c
	d.errMsg = ""
	d.title = c.Title()
	d.render()
	d.refreshFind() // recompute matches for the new content (find persists across items)
	if sameItem {
		d.vp.SetYOffset(offset)
	} else {
		d.vp.GotoTop()
	}
	d.hasContent = true
	d.loading = false
}

// path returns the loaded entity's file path, or "" when there's no content (error
// or empty pane) — so the title is linkified only when it points somewhere real.
func (d detailPane) path() string {
	if d.content == nil {
		return ""
	}
	return d.content.Path()
}

// SetError shows a per-item load error in the pane (keeps the browser alive).
func (d *detailPane) SetError(title, msg string) {
	d.content = nil
	d.rawStyled, d.prettyStyled, d.styled = "", "", ""
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
	d.rawStyled, d.prettyStyled, d.styled = "", "", ""
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
	d.rawStyled, d.prettyStyled, d.styled = "", "", ""
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
func (d *detailPane) updateFind(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, keys.Back):
		d.find.typing = false
		d.find.input.Blur()
		return nil
	case msg.String() == "enter":
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

// findNext moves the focused occurrence by dir (wrapping) and scrolls to it.
func (d *detailPane) findNext(dir int) {
	if len(d.find.matches) == 0 {
		return
	}
	n := len(d.find.matches)
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
	d.find.matches = nil
	d.find.cur = 0
	d.find.input.Blur()
}

// refreshFind recomputes the occurrences of the current query and re-renders the
// viewport with each highlighted (the focused one brighter). Each affected line is
// rebuilt over its *styled* text, so a match's highlight preserves the rest of the
// line's field colors and never splits an existing escape (see highlightLine).
func (d *detailPane) refreshFind() {
	if d.find.query == "" {
		d.vp.SetContent(d.styled)
		return
	}
	styled := strings.Split(d.styled, "\n")
	plain := strings.Split(ansi.Strip(d.styled), "\n")
	d.find.matches = d.find.matches[:0]
	for li, pl := range plain {
		for _, r := range foldMatches(pl, d.find.query) {
			d.find.matches = append(d.find.matches, matchPos{line: li, b0: r[0], b1: r[1]})
		}
	}
	if d.find.cur >= len(d.find.matches) {
		d.find.cur = 0
	}
	curLine, curB0 := -1, -1
	if len(d.find.matches) > 0 {
		curLine = d.find.matches[d.find.cur].line
		curB0 = d.find.matches[d.find.cur].b0
	}
	// Group occurrences by line (matches are already ascending by line then b0).
	occByLine := map[int][][2]int{}
	for _, m := range d.find.matches {
		occByLine[m.line] = append(occByLine[m.line], [2]int{m.b0, m.b1})
	}
	for li, occ := range occByLine {
		cb := -1
		if li == curLine {
			cb = curB0
		}
		styled[li] = highlightLine(styled[li], plain[li], occ, cb)
	}
	d.vp.SetContent(strings.Join(styled, "\n"))
}

// scrollToCurrent brings the focused occurrence into view (a couple of lines of lead).
func (d *detailPane) scrollToCurrent() {
	if len(d.find.matches) == 0 {
		return
	}
	if target := d.find.matches[d.find.cur].line - 2; target > 0 {
		d.vp.SetYOffset(target)
	} else {
		d.vp.GotoTop()
	}
}

// findStatus is the footer line shown while finding: the live input, or the
// applied query with the occurrence position and the nav hint.
func (d detailPane) findStatus() string {
	if d.find.typing {
		return d.find.input.View()
	}
	pos := 0
	if len(d.find.matches) > 0 {
		pos = d.find.cur + 1
	}
	status := dim(fmt.Sprintf("/%s  [%d/%d]  n/N next/prev · esc clear", d.find.query, pos, len(d.find.matches)))
	// Find runs per rendered line, so a query straddling a glamour wrap/reflow point
	// reads as 0 matches in pretty mode even though it's there. When the raw render
	// WOULD match, point at R so a real hit isn't mistaken for "not present" (L16).
	if d.pretty && d.find.query != "" && len(d.find.matches) == 0 && rawHasMatch(d.rawStyled, d.find.query) {
		status += dim(" · R: raw (match spans a wrap)")
	}
	return status
}

// rawHasMatch reports whether query occurs in the raw-rendered text (what R shows),
// using the same fold-aware matcher as the live search — so the L16 hint fires only
// when switching to raw would actually reveal the match.
func rawHasMatch(rawStyled, query string) bool {
	for _, line := range strings.Split(ansi.Strip(rawStyled), "\n") {
		if len(foldMatches(line, query)) > 0 {
			return true
		}
	}
	return false
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

func (d taskDetail) Title() string     { return d.t.Slug }
func (d taskDetail) Path() string      { return d.t.Path }
func (d taskDetail) rawBody() string   { return d.body }
func (d taskDetail) meta(w int) string { return renderTaskMeta(d.t, w) }

// renderTaskMeta formats a task's frontmatter field block (no body), wrapped to
// width. The body is rendered separately by the pane (raw or glamour).
func renderTaskMeta(t domain.Task, width int) string {
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
	return wrap(strings.TrimRight(b.String(), "\n"), width)
}

// --- epic detail ---

type epicDetail struct {
	e     domain.Epic
	tasks []domain.Task
	body  string
}

func (d epicDetail) Title() string     { return d.e.ID }
func (d epicDetail) Path() string      { return d.e.Path }
func (d epicDetail) rawBody() string   { return d.body }
func (d epicDetail) meta(w int) string { return renderEpicMeta(d.e, d.tasks, w) }

func renderEpicMeta(e domain.Epic, tasks []domain.Task, width int) string {
	var b strings.Builder
	detailField(&b, "epic", e.ID)
	detailField(&b, "status", e.Status)
	detailField(&b, "priority", priorityText(e.Priority))
	if len(e.Tags) > 0 {
		detailField(&b, "tags", strings.Join(e.Tags, ", "))
	}
	// Shared rollup (deprecated leaves the denominator, counted separately) so this
	// matches epic list / status / epic show.
	done, total, deprecated := core.TaskRollup(tasks)
	pct := 0
	if total > 0 {
		pct = done * 100 / total
	}
	progress := fmt.Sprintf("%s %s  %d/%d",
		miniBar(pct, 12), fg(theme.Percent(pct), fmt.Sprintf("%d%%", pct)), done, total)
	if deprecated > 0 {
		progress += fmt.Sprintf("  (%d deprecated)", deprecated)
	}
	detailField(&b, "progress", progress)
	if len(tasks) > 0 {
		b.WriteString("\n")
		for _, t := range tasks {
			tok := theme.Status(t.Status)
			fmt.Fprintf(&b, "  %s %s\n", fg(tok.Color, tok.Glyph), t.Slug)
		}
	}
	return wrap(strings.TrimRight(b.String(), "\n"), width)
}

// --- audit detail ---

type auditDetail struct {
	a    domain.Audit
	body string
}

func (d auditDetail) Title() string     { return d.a.Slug }
func (d auditDetail) Path() string      { return d.a.Path }
func (d auditDetail) rawBody() string   { return d.body }
func (d auditDetail) meta(w int) string { return renderAuditMeta(d.a, d.body, w) }

func renderAuditMeta(a domain.Audit, body string, width int) string {
	var b strings.Builder
	tok := theme.Bucket(a.Bucket)
	pct := a.Percent()
	progress := fmt.Sprintf("%s %s  %d/%d",
		miniBar(pct, 12), fg(theme.Percent(pct), fmt.Sprintf("%d%%", pct)), a.Resolved(), a.Findings)
	if a.OpenFindings > 0 {
		progress += fmt.Sprintf("  (%d open)", a.OpenFindings)
	}
	detailField(&b, "audit", a.Slug)
	detailField(&b, "bucket", fg(tok.Color, tok.Glyph+" "+string(a.Bucket)))
	detailField(&b, "area", a.Area)
	detailField(&b, "date", a.Date)
	detailField(&b, "findings", progress)
	// A glyph-coded finding index — status glyph + code + title, one scannable line
	// each — mirroring the epic detail's task list. The body below renders the same
	// findings as full prose; this is the at-a-glance, status-colored map of them
	// (which the prose, with its **Status:** buried inline, doesn't give you).
	if findings := domain.ParseFindings(body); len(findings) > 0 {
		b.WriteString("\n")
		for _, f := range findings {
			ftok := theme.FindingStatus(f.Status)
			line := fmt.Sprintf("  %s %s", fg(ftok.Color, ftok.Glyph), f.Code)
			if f.Title != "" {
				line += "  " + dim(truncate(f.Title, max1(width-len(f.Code)-6)))
			}
			b.WriteString(line + "\n")
		}
	}
	return wrap(strings.TrimRight(b.String(), "\n"), width)
}
