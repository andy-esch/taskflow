package tui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// row renders one list line with the shared cursor convention: a "› " accent
// marker when selected, two spaces otherwise, the content truncated to the list
// width. Used by every entity delegate so rows look consistent across tabs.
func row(w io.Writer, m list.Model, index int, content string, st *styles) {
	line := truncate(content, max1(m.Width()-2))
	if index == m.Index() {
		fmt.Fprint(w, st.selected.Render("› "+line))
		return
	}
	fmt.Fprint(w, "  "+line)
}

// rollupCounts formats a done/total rollup right-justified to width (0 = natural
// width). Shared by the epic + audit list rows and the dashboard's epics widget so
// the "12/166"-style column lines up the same way on every surface, padded to the
// widest value in its set rather than a fixed guess that a 3-digit total overflows.
// Counts are ASCII, so a byte-width pad (%*s) is also the display width.
func rollupCounts(done, total, width int) string {
	return fmt.Sprintf("%*s", width, theme.Counts(done, total))
}

// --- tasks ---

// taskItem adapts a domain.Task to a bubbles/list item. due is whether its revisit
// (snooze) date has arrived — computed once at load against the service clock (see
// loadTaskList), so the render path stays clock-free and a WithClock injection
// reaches the marker too.
type taskItem struct {
	t   domain.Task
	due bool
}

// FilterValue feeds the `/` fuzzy filter: slug, description, and tags so a tag
// query (e.g. "/go") narrows the list (S2b broadened this from slug+desc).
func (i taskItem) FilterValue() string {
	return i.t.Slug + " " + i.t.Description + " " + strings.Join(i.t.Tags, " ")
}
func (i taskItem) id() string   { return i.t.Slug }
func (i taskItem) path() string { return i.t.Path }

// lifecycleState is the task's current status — the action menu drops the no-op
// transition that lands on it (M10).
func (i taskItem) lifecycleState() string { return string(i.t.Status) }
func (i taskItem) sortFields() sortFields {
	return sortFields{priorityRank: priorityRank(i.t.Priority), updated: i.t.Updated, tier: i.t.Tier, slug: i.t.Slug}
}

// taskDelegate renders one task row: colored status glyph, a ⚠ if misfiled, the
// slug, and a dim relative date — truncated to fit the list width.
type taskDelegate struct{ st *styles }

func (taskDelegate) Height() int                         { return 1 }
func (taskDelegate) Spacing() int                        { return 0 }
func (taskDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (d taskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(taskItem)
	if !ok {
		return
	}
	st := d.st
	tok := theme.Status(it.t.Status)
	// One marker cell: a ↻ when a deferred task's revisit (snooze) date has arrived
	// (it.due, set at load) — the per-row twin of the `:revisit` view.
	marker := " "
	if it.due {
		marker = st.glyph(theme.MarkerRevisit)
	}
	date := theme.RelativeDate(theme.TaskDate(it.t))

	// Reserve: glyph(1) + marker(1) + 3 spaces + date(≤10) within the row budget.
	slugW := m.Width() - 2 - 2 - 3 - 10
	if slugW < 8 {
		slugW = 8
	}
	// Pad by display cells, not bytes (%-*s) — a non-ASCII slug would otherwise
	// shove the date column out of alignment.
	slug := padRight(truncate(it.t.Slug, slugW), slugW)
	row(w, m, index, fmt.Sprintf("%s %s %s  %s", st.fg(tok.Color, tok.Glyph), marker, slug, st.dim(date)), st)
}

// --- epics ---

// epicItem adapts a core.EpicSummary (epic + rollup) to a list item. countsW is
// the done/total column width measured across the whole list at load (see
// loadEpicList), so the delegate can pad to it without re-scanning siblings.
type epicItem struct {
	es      core.EpicSummary
	countsW int
}

func (i epicItem) FilterValue() string {
	return i.es.Epic.ID + " " + i.es.Epic.Description + " " + strings.Join(i.es.Epic.Tags, " ")
}
func (i epicItem) id() string   { return i.es.Epic.ID }
func (i epicItem) path() string { return i.es.Epic.Path }

// lifecycleState is the epic's current status (active/retired/deprecated) — the
// action menu drops the no-op transition that lands on it.
func (i epicItem) lifecycleState() string { return i.es.Epic.Status }
func (i epicItem) sortFields() sortFields {
	// Epics have no tier/updated; priority + id (slug) carry the sort.
	return sortFields{priorityRank: priorityRank(i.es.Epic.Priority), slug: i.es.Epic.ID}
}

// epicGlyph is the leading state glyph for an epic row, shared by the epics tab and
// the dashboard so they read alike. A ⚠ when the status is non-conforming (outside
// active/retired/deprecated) — a fixable data problem that takes priority over
// liveness; otherwise the liveness band glyph (working/fresh/dormant), mirroring the
// audit row's bucket glyph.
func epicGlyph(es core.EpicSummary, st *styles) string {
	if !domain.IsKnownEpicStatus(es.Epic.Status) {
		return st.glyph(theme.MarkerWarn)
	}
	tok := theme.Liveness(string(es.Liveness()))
	return st.fg(tok.Color, tok.Glyph)
}

// epicStatusNote annotates a non-conforming epic row with its offending status, so
// the ⚠ says WHAT to fix (set active/retired/deprecated via the m-menu or `epic
// move`). "" when the status conforms; "—" stands in for an empty status.
func epicStatusNote(es core.EpicSummary, st *styles) string {
	if domain.IsKnownEpicStatus(es.Epic.Status) {
		return ""
	}
	s := es.Epic.Status
	if s == "" {
		s = "—"
	}
	return "  " + st.fg(theme.ColorYellow, "status:"+s)
}

// epicDelegate renders one epic row: a leading glyph (liveness, or ⚠ for a
// non-conforming status), then a rollup bar + colored percent + done/total + the
// epic id and description. A dormant (drained) epic dims its id so a quiet bucket
// recedes even on a mono terminal; a non-conforming one shows its raw status.
type epicDelegate struct{ st *styles }

func (epicDelegate) Height() int                         { return 1 }
func (epicDelegate) Spacing() int                        { return 0 }
func (epicDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (d epicDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(epicItem)
	if !ok {
		return
	}
	st := d.st
	pct := it.es.Percent()
	bar := st.miniBar(pct, 8)
	pctStr := st.fg(theme.Percent(pct), theme.PercentLabelPadded(pct))
	counts := rollupCounts(it.es.Done, it.es.Total, it.countsW)
	id := it.es.Epic.ID
	if !it.es.Live() { // dormant buckets recede: the id dims like the description
		id = st.dim(id)
	}
	idAndDesc := id + epicStatusNote(it.es, st) + "  " + st.dim(it.es.Epic.Description)
	row(w, m, index, fmt.Sprintf("%s %s %s %s  %s",
		epicGlyph(it.es, st), bar, pctStr, counts, idAndDesc), st)
}

// --- audits ---

// auditItem adapts a domain.Audit to a list item. countsW is the resolved/total
// column width measured across the list at load (see loadAuditList).
type auditItem struct {
	a       domain.Audit
	countsW int
}

func (i auditItem) FilterValue() string { return i.a.Slug + " " + i.a.Area }
func (i auditItem) id() string          { return i.a.Slug }
func (i auditItem) path() string        { return i.a.Path }

// lifecycleState is the audit's current bucket — the action menu drops the no-op
// transition that lands on it (e.g. reopen on an already-open audit).
func (i auditItem) lifecycleState() string { return string(i.a.Bucket) }
func (i auditItem) sortFields() sortFields {
	// Audits sort by date (as "updated") + slug; no priority/tier.
	return sortFields{updated: i.a.Date, slug: i.a.Slug}
}

// auditDelegate renders one audit row: a bucket glyph (state), then the same
// rollup bar + colored percent + resolved/total the epic row uses, the slug, and
// a dim area.
type auditDelegate struct{ st *styles }

func (auditDelegate) Height() int                         { return 1 }
func (auditDelegate) Spacing() int                        { return 0 }
func (auditDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (d auditDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(auditItem)
	if !ok {
		return
	}
	st := d.st
	tok := theme.Bucket(it.a.Bucket)
	pct := it.a.Percent()
	bar := st.segBar(it.a.DoneFindings, it.a.ActiveFindings, it.a.DroppedFindings, it.a.Findings, 8)
	pctStr := st.fg(theme.Percent(pct), theme.PercentLabelPadded(pct))
	counts := rollupCounts(it.a.Resolved(), it.a.Findings, it.countsW)
	row(w, m, index, fmt.Sprintf("%s %s %s %s  %s  %s",
		st.fg(tok.Color, tok.Glyph), bar, pctStr, counts, it.a.Slug, st.dim(it.a.Area)), st)
}
