package tui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

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
	t            domain.Task
	due          bool
	identityHint string
}

// FilterValue feeds the `/` fuzzy filter: slug, description, and tags so a tag
// query (e.g. "/go") narrows the list (S2b broadened this from slug+desc).
func (i taskItem) FilterValue() string {
	return i.t.Slug + " " + i.t.CanonicalID() + " " + i.t.Description + " " + strings.Join(i.t.Tags, " ")
}
func (i taskItem) ref() entityRef { return entityRef{key: i.t.CanonicalID(), label: i.t.Slug} }
func (i taskItem) displayLabel() string {
	return labelWithIdentityHint(i.t.Slug, i.identityHint)
}
func (i taskItem) hasIdentityHint() bool { return i.identityHint != "" }
func (i taskItem) path() string          { return i.t.Path }

// lifecycleState is the task's current status — the action menu drops the no-op
// transition that lands on it (M10).
func (i taskItem) lifecycleState() string { return string(i.t.Status) }
func (i taskItem) sortFields() sortFields {
	return sortFields{priorityRank: priorityRank(i.t.Priority), updated: i.t.Updated, tier: i.t.Tier, slug: i.t.Slug}
}

// taskDelegate renders one task row: colored status glyph, a ↻ marker when a
// deferred task's revisit date has arrived, the slug, and a dim relative date —
// truncated to fit the list width.
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
	slug := padRight(truncate(it.displayLabel(), slugW), slugW)
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
func (i epicItem) ref() entityRef {
	return entityRef{key: i.es.Epic.ID, label: i.es.Epic.ID}
}
func (i epicItem) displayLabel() string  { return i.es.Epic.ID }
func (i epicItem) hasIdentityHint() bool { return false }
func (i epicItem) path() string          { return i.es.Epic.Path }

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

// --- threads ---

// threadItem retains the complete shared core projection. Rendering may count
// supplied role/gate values for compact presentation, but it never traverses
// dependencies or recreates health, eligibility, or frontier rules.
type threadItem struct {
	view         core.ThreadView
	countsW      int
	identityHint string
}

func (i threadItem) FilterValue() string {
	t := i.view.Thread
	return strings.Join([]string{
		t.Slug, t.CanonicalID(), string(t.Status), t.Description, t.Goal,
		strings.Join(t.Tags, " "), string(i.view.GraphHealth), string(i.view.ProjectionHealth),
	}, " ")
}
func (i threadItem) ref() entityRef {
	return entityRef{key: i.view.Thread.CanonicalID(), label: i.view.Thread.Slug}
}
func (i threadItem) displayLabel() string {
	return labelWithIdentityHint(i.view.Thread.Slug, i.identityHint)
}
func (i threadItem) hasIdentityHint() bool { return i.identityHint != "" }

// Thread paths are resolved only through the optional ThreadPathSource during
// detail loading. The portable semantic record is never treated as that port.
func (i threadItem) path() string { return "" }
func (i threadItem) sortFields() sortFields {
	updated := i.view.Thread.Updated
	if updated == "" {
		updated = i.view.Thread.Created
	}
	return sortFields{updated: updated, slug: i.view.Thread.Slug}
}

// threadActivity partitions work using the core projection's authoritative
// frontier. A clear local gate does not imply dispatchability when repository
// graph evidence is unhealthy, so pending work outside view.Frontier belongs in
// notDispatchable instead of disappearing from the row.
type threadActivity struct {
	inFlight        int
	dispatchable    int
	notDispatchable int
}

func activityForThread(view core.ThreadView) threadActivity {
	a := threadActivity{dispatchable: len(view.Frontier)}
	pending := 0
	for _, member := range view.Members {
		switch member.State.Role {
		case core.RoleInFlight:
			a.inFlight++
		case core.RoleQueued, core.RoleCandidate:
			pending++
		}
	}
	a.notDispatchable = max(0, pending-a.dispatchable)
	return a
}

func threadWorkText(activity threadActivity, compact bool) string {
	if compact {
		compactCount := func(value int) string {
			if value > 99 {
				return "99+"
			}
			return fmt.Sprint(value)
		}
		return "▶" + compactCount(activity.inFlight) +
			"✓" + compactCount(activity.dispatchable) +
			"×" + compactCount(activity.notDispatchable)
	}
	return fmt.Sprintf("▶%d ✓%d ×%d", activity.inFlight, activity.dispatchable, activity.notDispatchable)
}

// truncateMiddle retains both ends of a long plain-text identity. Thread slugs
// commonly share an initiative prefix and differ at the tail, so ordinary
// right-truncation can make distinct rows identical. ANSI's grapheme-aware
// helpers keep this safe for wide and combining Unicode too.
func truncateMiddle(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(value) <= width {
		return value
	}
	if width == 1 {
		return "…"
	}
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1
	left := ansi.Truncate(value, leftWidth, "")
	right := ansi.TruncateLeft(value, ansi.StringWidth(value)-rightWidth, "")
	return left + "…" + right
}

// threadIdentityLabel gives a duplicate's canonical hint first claim on the
// remaining row budget. If even the hint is unusually long, middle elision keeps
// its distinguishing tail instead of clipping it away.
func threadIdentityLabel(item threadItem, width int) string {
	if width <= 0 {
		return ""
	}
	if item.identityHint == "" {
		return truncateMiddle(item.view.Thread.Slug, width)
	}
	hint := "[" + item.identityHint + "]"
	if ansi.StringWidth(hint) >= width {
		return truncateMiddle(hint, width)
	}
	remaining := width - ansi.StringWidth(hint) - 1
	if remaining <= 0 {
		return hint
	}
	return hint + " " + truncateMiddle(item.view.Thread.Slug, remaining)
}

func threadRowTail(item threadItem, width int, st *styles) string {
	if width <= 0 {
		return ""
	}
	label := threadIdentityLabel(item, width)
	remaining := width - ansi.StringWidth(label) - 2
	if remaining <= 0 || item.view.Thread.Description == "" {
		return label
	}
	return label + "  " + st.dim(truncate(item.view.Thread.Description, remaining))
}

func threadHealthText(view core.ThreadView, s *styles, compact bool) string {
	mark := func(health core.GraphHealth) string {
		switch health {
		case core.GraphHealthy:
			return s.fg(theme.ColorGreen, "✓")
		case core.GraphDegraded:
			return s.fg(theme.ColorYellow, "~")
		default:
			return s.fg(theme.ColorRed, "!")
		}
	}
	if compact {
		if view.GraphHealth == view.ProjectionHealth {
			return mark(view.ProjectionHealth)
		}
		// At the smallest widths, retain the graph/projection ordering while
		// dropping only the redundant letters and separator.
		return mark(view.GraphHealth) + mark(view.ProjectionHealth)
	}
	return "g" + mark(view.GraphHealth) + "/v" + mark(view.ProjectionHealth)
}

type threadDelegate struct{ st *styles }

func (threadDelegate) Height() int                         { return 1 }
func (threadDelegate) Spacing() int                        { return 0 }
func (threadDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (d threadDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(threadItem)
	if !ok {
		return
	}
	st, view := d.st, it.view
	status := theme.ThreadStatus(view.Thread.Status)
	activity := activityForThread(view)
	health := threadHealthText(view, st, m.Width() < 34)
	if view.Inconsistent {
		health += " " + st.fg(theme.ColorYellow, theme.MarkerWarn.Glyph)
	}

	contentWidth := max1(m.Width() - 2) // row reserves the two-cell cursor prefix
	var prefix string
	switch {
	case m.Width() >= 64:
		progress := fmt.Sprintf("d:%s s:%s",
			rollupCounts(view.Rollup.Done, view.Rollup.Total, it.countsW),
			rollupCounts(view.Rollup.Drained, view.Rollup.Total, it.countsW))
		prefix = fmt.Sprintf("%s %-11s %s %s  %s",
			st.fg(status.Color, status.Glyph), view.Thread.Status,
			padRight(health, 7), progress, threadWorkText(activity, false))
	case m.Width() >= 42:
		prefix = fmt.Sprintf("%s %s %s",
			st.fg(status.Color, status.Glyph), padRight(health, 7), threadWorkText(activity, false))
	default:
		prefix = fmt.Sprintf("%s %s %s",
			st.fg(status.Color, status.Glyph), health, threadWorkText(activity, true))
	}
	line := truncate(prefix, contentWidth)
	if remaining := contentWidth - ansi.StringWidth(prefix) - 2; remaining > 0 {
		line = prefix + "  " + threadRowTail(it, remaining, st)
	}
	row(w, m, index, line, st)
}

// --- audits ---

// auditItem adapts a domain.Audit to a list item. countsW is the resolved/total
// column width measured across the list at load (see loadAuditList).
type auditItem struct {
	a            domain.Audit
	countsW      int
	identityHint string
}

func (i auditItem) FilterValue() string { return i.a.Slug + " " + i.a.CanonicalID() + " " + i.a.Area }
func (i auditItem) ref() entityRef      { return entityRef{key: i.a.CanonicalID(), label: i.a.Slug} }
func (i auditItem) displayLabel() string {
	return labelWithIdentityHint(i.a.Slug, i.identityHint)
}
func (i auditItem) hasIdentityHint() bool { return i.identityHint != "" }
func (i auditItem) path() string          { return i.a.Path }

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
	pctStr := st.fg(theme.Percent(pct), theme.AuditPercentLabelPadded(pct))
	counts := rollupCounts(it.a.Resolved(), it.a.Findings, it.countsW)
	line := fmt.Sprintf("%s %s %s %s  %s  %s",
		st.fg(tok.Color, tok.Glyph), bar, pctStr, counts, it.displayLabel(), st.dim(it.a.Area))
	if it.a.ReadyToClose() {
		line += "  " + st.fg(theme.ColorGreen, "✔ ready to close")
	}
	row(w, m, index, line, st)
}

// researchItem is one row of the research tab. Research has no lifecycle, so unlike
// task/audit items it exposes no lifecycleState — the `m` action menu is inert here,
// which the model already handles by checking len(transitions) before opening it.
type researchItem struct {
	r            domain.Research
	identityHint string
}

// FilterValue spans slug, description, and tags: the corpus is browsed by topic, so `/`
// has to reach a doc by what it's ABOUT, not just what the file is called.
func (i researchItem) FilterValue() string {
	return i.r.Slug + " " + i.r.CanonicalID() + " " + i.r.Description + " " + strings.Join(i.r.Tags, " ")
}
func (i researchItem) ref() entityRef { return entityRef{key: i.r.CanonicalID(), label: i.r.Slug} }
func (i researchItem) displayLabel() string {
	return labelWithIdentityHint(i.r.Slug, i.identityHint)
}
func (i researchItem) hasIdentityHint() bool { return i.identityHint != "" }
func (i researchItem) path() string          { return i.r.Path }

func (i researchItem) sortFields() sortFields {
	// Research sorts by created (the loader's default order), updated, and slug — no
	// priority/tier, and no status axis. `updated` falls back to created for a doc never
	// edited, matching the CLI's `updated` column, so a fresh doc never sinks to the
	// bottom of the updated sort.
	updated := i.r.Updated
	if updated == "" {
		updated = i.r.Created
	}
	return sortFields{updated: updated, slug: i.r.Slug}
}

// researchDelegate renders one research row: the created date as the organizing column
// (research has no status glyph or progress bar — nothing to show), then the slug and a
// dim description. Deliberately the plainest row in the TUI; the entity is a snapshot
// list, not a work board.
type researchDelegate struct{ st *styles }

func (researchDelegate) Height() int                         { return 1 }
func (researchDelegate) Spacing() int                        { return 0 }
func (researchDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (d researchDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(researchItem)
	if !ok {
		return
	}
	st := d.st
	desc := it.r.Description
	if desc == "" {
		desc = "—" // a doc with no summary: show the gap rather than a ragged blank
	}
	line := fmt.Sprintf("%s  %s  %s", st.dim(it.r.Created), it.displayLabel(), st.dim(desc))
	row(w, m, index, line, st)
}
