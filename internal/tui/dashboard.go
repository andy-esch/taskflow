package tui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// The dashboard is the TUI's landing screen — the in-app counterpart of
// `tskflwctl status`. It's NOT an entityTab (no list/filter/sort); it's a
// composite of read-only widgets rendered from one core.Summary, with a cursor
// over the navigable rows. Selecting a row jumps into the relevant tab/view
// (see Model.dashJump), so the dashboard never mutates — it orients and routes.
// v1 widgets: in-progress · due-for-revisit · epic rollups · needs-attention. (A
// cross-cutting "goals" widget waits on the Projects entity.)

// dashListCap bounds each list widget so the dashboard stays a glanceable summary;
// the overflow collapses into a "+N more →" row that jumps to the full tab.
const dashListCap = 6

var dashHeading = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))

// dashTarget is where selecting a row navigates: a specific item (id set) via
// jumpTo, or a whole view (view set) via applyView, on the named entity's tab.
type dashTarget struct {
	kind entityKind
	id   string
	view string
}

// dashRow is one rendered line. A nil target is a heading/info line (not
// selectable); a non-nil target is a navigable row.
type dashRow struct {
	text   string
	target *dashTarget
}

// dashboard holds the rendered rows plus a cursor over the navigable ones.
type dashboard struct {
	loaded  bool
	loadErr error // last Summary load failure; the rows below are the last good load (or none)
	rows    []dashRow
	nav     []int // indices into rows that carry a target (selectable), in order
	cursor  int   // index into nav
}

// setSummary (re)builds the widget rows from a fresh core.Summary, recomputing the
// navigable set and clamping the cursor.
func (d *dashboard) setSummary(s core.Summary) {
	var rows []dashRow
	head := func(t string) { rows = append(rows, dashRow{text: dashHeading.Render(t)}) }
	info := func(t string) { rows = append(rows, dashRow{text: dim("  " + t)}) }
	blank := func() { rows = append(rows, dashRow{}) }
	nav := func(text string, tgt dashTarget) { t := tgt; rows = append(rows, dashRow{text: text, target: &t}) }
	capList := func(n int) (shown, more int) {
		if n > dashListCap {
			return dashListCap, n - dashListCap
		}
		return n, 0
	}

	// In progress — the active work, each with how long since it was last touched
	// (a staleness cue) in an aligned column, the slug last so it absorbs truncation.
	head(fmt.Sprintf("in progress (%d)", len(s.InProgress)))
	if len(s.InProgress) == 0 {
		info("nothing in progress")
	} else {
		shown, more := capList(len(s.InProgress))
		vis := s.InProgress[:shown]
		dates := make([]string, len(vis))
		dateW := 0
		for i, t := range vis {
			dates[i] = theme.RelativeDate(theme.TaskDate(t)) // ASCII, so len == display width
			dateW = max(dateW, len(dates[i]))
		}
		for i, t := range vis {
			tok := theme.Status(t.Status)
			cell := fg(tok.Color, tok.Glyph) + " "
			if dateW > 0 { // blank (undated) cells still pad, so the slug column holds
				cell += dim(fmt.Sprintf("%-*s", dateW, dates[i])) + "  "
			}
			cell += t.Slug
			nav(cell, dashTarget{kind: entityTasks, id: t.Slug})
		}
		if more > 0 {
			nav(dim(fmt.Sprintf("+%d more →", more)), dashTarget{kind: entityTasks, view: "in-progress"})
		}
	}

	// Due for revisit — only when a snooze has come due.
	if s.RevisitDue > 0 {
		blank()
		head("due for revisit")
		nav(fg(theme.ColorYellow, "↻")+fmt.Sprintf(" %d snoozed task(s) now due", s.RevisitDue),
			dashTarget{kind: entityTasks, view: "revisit"})
	}

	// Epics — rollup progress, most-recently-touched first (the dashboard's "what
	// moved lately" lens; the epics tab keeps its own sort). The counts and date are
	// pre-measured and padded to a shared width so the columns line up; the epic id
	// goes LAST, where width-truncation naturally falls and the date stays visible.
	// The +N overflow still jumps to the full tab.
	blank()
	head("epics")
	if len(s.Epics) == 0 {
		info("no epics")
	} else {
		epics := epicsByRecent(s.Epics)
		shown, more := capList(len(epics))
		vis := epics[:shown]
		dates := make([]string, len(vis))
		countsW, dateW := 0, 0
		for i, es := range vis {
			dates[i] = theme.RelativeDate(es.LastUpdated) // ASCII, so len == display width
			countsW = max(countsW, len(rollupCounts(es.Done, es.Total, 0)))
			dateW = max(dateW, len(dates[i]))
		}
		for i, es := range vis {
			pct := es.Percent()
			row := fmt.Sprintf("%s %s  %s", miniBar(pct, 8),
				fg(theme.Percent(pct), fmt.Sprintf("%3d%%", pct)), rollupCounts(es.Done, es.Total, countsW))
			if dateW > 0 { // a blank (undated) cell still pads, so the id column holds
				row += "  " + dim(fmt.Sprintf("%-*s", dateW, dates[i]))
			}
			row += "  " + es.Epic.ID
			nav(row, dashTarget{kind: entityEpics, id: es.Epic.ID})
		}
		if more > 0 {
			nav(dim(fmt.Sprintf("+%d more →", more)), dashTarget{kind: entityEpics})
		}
	}

	// Needs attention — misfiled tasks, the open-audit queue, and unreadable files.
	// Under a non-specific heading a bare count says nothing, so every row names its
	// own category and wears its entity's glyph (the audit ◆ matches the audits tab);
	// "all clear" when there's nothing.
	blank()
	head("needs attention")
	allClear := true
	if s.Misfiled > 0 {
		nav(fg(theme.ColorYellow, "⚠")+fmt.Sprintf(" %d misfiled task(s) (status ≠ folder)", s.Misfiled),
			dashTarget{kind: entityTasks, view: "all"})
		allClear = false
	}
	if n := len(s.OpenAudits); n > 0 {
		tok := theme.Bucket(domain.AuditOpen)
		nav(fg(tok.Color, tok.Glyph)+fmt.Sprintf(" %d open audit(s)", n), dashTarget{kind: entityAudits})
		allClear = false
	}
	if len(s.Problems) > 0 {
		info(fg(theme.ColorRed, "!") + fmt.Sprintf(" %d unreadable file(s) (run lint)", len(s.Problems)))
		allClear = false
	}
	if allClear {
		info(fg(theme.ColorGreen, "✔") + " all clear")
	}

	d.rows = rows
	d.nav = d.nav[:0]
	for i, r := range rows {
		if r.target != nil {
			d.nav = append(d.nav, i)
		}
	}
	if d.cursor >= len(d.nav) {
		d.cursor = 0
	}
	d.loaded = true
}

// epicsByRecent orders epics most-recently-updated first for the dashboard's
// recency lens. Stable, so equal dates keep core's order; undated epics ("" date)
// sink to the bottom.
func epicsByRecent(src []core.EpicSummary) []core.EpicSummary {
	out := append([]core.EpicSummary(nil), src...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].LastUpdated > out[j].LastUpdated })
	return out
}

// move steps the cursor over the navigable rows (wrapping).
func (d *dashboard) move(delta int) {
	if n := len(d.nav); n > 0 {
		d.cursor = ((d.cursor+delta)%n + n) % n
	}
}

// selectedTarget returns the cursor row's navigation target.
func (d dashboard) selectedTarget() (dashTarget, bool) {
	if len(d.nav) == 0 {
		return dashTarget{}, false
	}
	if t := d.rows[d.nav[d.cursor]].target; t != nil {
		return *t, true
	}
	return dashTarget{}, false
}

// view renders the widgets into the body, the cursor row accented. v1 doesn't
// scroll lists (each widget is capped); but the composed widgets can still overrun
// a short terminal, so when they do the output is a window that keeps the cursor
// row on screen — a selectable row must never be navigable-but-invisible.
func (d dashboard) view(maxW, maxH int) string {
	if !d.loaded {
		return dim("loading…")
	}
	cursorRow := -1
	if len(d.nav) > 0 {
		cursorRow = d.nav[d.cursor]
	}
	lines := make([]string, len(d.rows))
	for i, r := range d.rows {
		switch {
		case r.target == nil:
			lines[i] = r.text // heading / info, rendered with its own style + indent
		case i == cursorRow:
			lines[i] = truncate(selectedStyle.Render("› ")+r.text, max1(maxW))
		default:
			lines[i] = truncate("  "+r.text, max1(maxW))
		}
	}
	if maxH > 0 && len(lines) > maxH {
		lines = scrollTo(lines, cursorRow, maxH)
	}
	return strings.Join(lines, "\n")
}

// scrollTo returns the maxH-line window over lines that keeps the focused row
// visible — centered when it can be, clamped at the ends. A negative focus (no
// cursor) shows the top. Only called when len(lines) > maxH.
func scrollTo(lines []string, focus, maxH int) []string {
	start := 0
	if focus >= 0 {
		start = focus - maxH/2
	}
	if start > len(lines)-maxH {
		start = len(lines) - maxH
	}
	if start < 0 {
		start = 0
	}
	return lines[start : start+maxH]
}
