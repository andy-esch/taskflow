package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/theme"
)

// The dashboard is the TUI's landing screen — the in-app counterpart of
// `tskflwctl status`. It's NOT an entityTab (no list/filter/sort); it's a
// composite of read-only widgets rendered from one core.Summary, with a cursor
// over the navigable rows. Selecting a row jumps into the relevant tab/view
// (see Model.dashJump), so the dashboard never mutates — it orients and routes.
// v1 widgets: in-progress · due-for-revisit · epic rollups · health. (A
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
	loaded bool
	rows   []dashRow
	nav    []int // indices into rows that carry a target (selectable), in order
	cursor int   // index into nav
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

	// In progress — the active work.
	head(fmt.Sprintf("in progress (%d)", len(s.InProgress)))
	if len(s.InProgress) == 0 {
		info("nothing in progress")
	} else {
		shown, more := capList(len(s.InProgress))
		for _, t := range s.InProgress[:shown] {
			tok := theme.Status(t.Status)
			nav(fg(tok.Color, tok.Glyph)+" "+t.Slug, dashTarget{kind: entityTasks, id: t.Slug})
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

	// Epics — rollup progress.
	blank()
	head("epics")
	if len(s.Epics) == 0 {
		info("no epics")
	} else {
		shown, more := capList(len(s.Epics))
		for _, es := range s.Epics[:shown] {
			pct := es.Percent()
			nav(fmt.Sprintf("%s %s  %d/%d  %s",
				miniBar(pct, 8), fg(theme.Percent(pct), fmt.Sprintf("%3d%%", pct)), es.Done, es.Total, es.Epic.ID),
				dashTarget{kind: entityEpics, id: es.Epic.ID})
		}
		if more > 0 {
			nav(dim(fmt.Sprintf("+%d more →", more)), dashTarget{kind: entityEpics})
		}
	}

	// Health — what needs attention.
	blank()
	head("health")
	healthy := true
	if s.Misfiled > 0 {
		nav(fg(theme.ColorYellow, "⚠")+fmt.Sprintf(" %d misfiled (status ≠ folder)", s.Misfiled),
			dashTarget{kind: entityTasks, view: "all"})
		healthy = false
	}
	if len(s.OpenAudits) > 0 {
		nav(fmt.Sprintf("%d open audit(s)", len(s.OpenAudits)), dashTarget{kind: entityAudits})
		healthy = false
	}
	if len(s.Problems) > 0 {
		info(fg(theme.ColorRed, "!") + fmt.Sprintf(" %d unreadable file(s) (run lint)", len(s.Problems)))
		healthy = false
	}
	if healthy {
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

// view renders the widgets into the body, the cursor row accented. v1 keeps the
// content bounded (capped lists) rather than scrolling; the pane clamps the rest.
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
		lines = lines[:maxH]
	}
	return strings.Join(lines, "\n")
}
