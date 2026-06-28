package tui

import (
	"fmt"
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
	line := func(t string) { rows = append(rows, dashRow{text: "  " + t}) } // info, but keeps its own colors
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
		nav(glyph(theme.MarkerRevisit)+fmt.Sprintf(" %d snoozed task(s) now due", s.RevisitDue),
			dashTarget{kind: entityTasks, view: "revisit"})
	}

	// Epics — rollup progress in core.Summary's order (most-recently-touched first;
	// the same order CLI `status` renders, so the two dashboards agree — audit M2).
	// The counts and date are pre-measured and padded to a shared width so the columns
	// line up; the epic id goes LAST, where width-truncation naturally falls and the
	// date stays visible. The +N overflow still jumps to the full tab.
	blank()
	head("epics")
	if len(s.Epics) == 0 {
		info("no epics")
	} else {
		epics := s.Epics
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
			// live-first, dormant dimmed; a ⚠ leads instead when the status is
			// non-conforming (the same glyph the epics tab shows — see epicGlyph).
			row := fmt.Sprintf("%s %s %s  %s", epicGlyph(es), miniBar(pct, 8),
				fg(theme.Percent(pct), theme.PercentLabelPadded(pct)), rollupCounts(es.Done, es.Total, countsW))
			if dateW > 0 { // a blank (undated) cell still pads, so the id column holds
				row += "  " + dim(fmt.Sprintf("%-*s", dateW, dates[i]))
			}
			id := es.Epic.ID
			if !es.Live() { // dormant buckets recede on the dashboard too
				id = dim(id)
			}
			row += "  " + id + epicStatusNote(es)
			nav(row, dashTarget{kind: entityEpics, id: es.Epic.ID})
		}
		if more > 0 {
			nav(dim(fmt.Sprintf("+%d more →", more)), dashTarget{kind: entityEpics})
		}
	}

	// Audit findings — the cross-audit actionable inbox (open/in-progress findings),
	// triaged by urgency and by subsystem, with the rare acute ones called out. Each
	// acute row jumps to its parent audit; the breakdown lines are read-only.
	if fr := s.Findings; fr.Open+fr.InProgress > 0 {
		blank()
		head(fmt.Sprintf("audit findings (%d open · %d in progress)", fr.Open, fr.InProgress))
		if len(fr.ByUrgency) > 0 {
			line("by urgency:  " + urgencyLine(fr.ByUrgency))
		}
		if len(fr.ByComponent) > 0 {
			line("by area:     " + componentLine(fr.ByComponent, 5))
		}
		for _, f := range fr.Acute {
			label := strings.TrimSpace(f.Code + " " + f.Title)
			nav(fg(theme.ColorRed, "⚠")+" "+label, dashTarget{kind: entityAudits, id: f.Audit})
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
		nav(glyph(theme.MarkerWarn)+fmt.Sprintf(" %d misfiled task(s) (status ≠ folder)", s.Misfiled),
			dashTarget{kind: entityTasks, view: "all"})
		allClear = false
	}
	if n := len(s.OpenAudits); n > 0 {
		nav(glyph(theme.Bucket(domain.AuditOpen))+fmt.Sprintf(" %d open audit(s)", n), dashTarget{kind: entityAudits})
		allClear = false
	}
	if s.ReadyToClose > 0 {
		nav(glyph(theme.MarkerReadyToClose)+fmt.Sprintf(" %d audit(s) ready to close (all findings resolved)", s.ReadyToClose),
			dashTarget{kind: entityAudits})
		allClear = false
	}
	if s.BadEpicStatus > 0 {
		nav(glyph(theme.MarkerWarn)+fmt.Sprintf(" %d epic(s) with unrecognized status (set active/retired/deprecated)", s.BadEpicStatus),
			dashTarget{kind: entityEpics})
		allClear = false
	}
	if len(s.Problems) > 0 {
		info(glyph(theme.MarkerUnreadable) + fmt.Sprintf(" %d unreadable file(s) (run lint)", len(s.Problems)))
		allClear = false
	}
	if allClear {
		info(glyph(theme.MarkerAllClear) + " all clear")
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

// urgencyLine renders a finding-urgency breakdown ("⚠ 1 acute · 12 soon · 23
// eventually"), coloring acute (red) and soon (yellow) so the sharp end stands out.
// Shares the iterate/join STRUCTURE with the CLI's countByLine via theme.Breakdown;
// only this surface's coloring differs (audit M10).
func urgencyLine(cs []core.CountBy) string {
	return theme.Breakdown(cs, dim(" · "), 0, func(c core.CountBy) string {
		seg := fmt.Sprintf("%d %s", c.Count, c.Key)
		switch c.Key {
		case "acute":
			seg = fg(theme.ColorRed, "⚠ "+seg)
		case "soon":
			seg = fg(theme.ColorYellow, seg)
		}
		return seg
	}, nil)
}

// componentLine renders the top-topN components by finding count ("stravapipe 14 ·
// dispatcher 9 · …"), with a dim "+N more" tail when there are more.
func componentLine(cs []core.CountBy, topN int) string {
	return theme.Breakdown(cs, dim(" · "), topN,
		func(c core.CountBy) string { return fmt.Sprintf("%s %d", c.Key, c.Count) },
		func(remaining int) string { return dim(fmt.Sprintf("+%d more", remaining)) })
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
			lines[i] = truncate(r.text, max1(maxW)) // heading / info / breakdown line — width-safe
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
