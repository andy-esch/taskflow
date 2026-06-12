package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// row renders one list line with the shared cursor convention: a "› " accent
// marker when selected, two spaces otherwise, the content truncated to the list
// width. Used by every entity delegate so rows look consistent across tabs.
func row(w io.Writer, m list.Model, index int, content string) {
	line := truncate(content, max1(m.Width()-2))
	if index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render("› "+line))
		return
	}
	fmt.Fprint(w, "  "+line)
}

// --- tasks ---

// taskItem adapts a domain.Task to a bubbles/list item.
type taskItem struct{ t domain.Task }

// FilterValue feeds the `/` fuzzy filter: slug, description, and tags so a tag
// query (e.g. "/go") narrows the list (S2b broadened this from slug+desc).
func (i taskItem) FilterValue() string {
	return i.t.Slug + " " + i.t.Description + " " + strings.Join(i.t.Tags, " ")
}
func (i taskItem) id() string { return i.t.Slug }
func (i taskItem) sortFields() sortFields {
	return sortFields{priorityRank: priorityRank(i.t.Priority), updated: i.t.Updated, tier: i.t.Tier, slug: i.t.Slug}
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
	date := theme.RelativeDate(theme.TaskDate(it.t))

	// Reserve: glyph(1) + marker(1) + 3 spaces + date(≤10) within the row budget.
	slugW := m.Width() - 2 - 2 - 3 - 10
	if slugW < 8 {
		slugW = 8
	}
	slug := fmt.Sprintf("%-*s", slugW, truncate(it.t.Slug, slugW))
	row(w, m, index, fmt.Sprintf("%s %s %s  %s", fg(tok.Color, tok.Glyph), marker, slug, dim(date)))
}

// --- epics ---

// epicItem adapts a core.EpicSummary (epic + rollup) to a list item.
type epicItem struct{ es core.EpicSummary }

func (i epicItem) FilterValue() string {
	return i.es.Epic.ID + " " + i.es.Epic.Description + " " + strings.Join(i.es.Epic.Tags, " ")
}
func (i epicItem) id() string { return i.es.Epic.ID }
func (i epicItem) sortFields() sortFields {
	// Epics have no tier/updated; priority + id (slug) carry the sort.
	return sortFields{priorityRank: priorityRank(i.es.Epic.Priority), slug: i.es.Epic.ID}
}

// epicDelegate renders one epic row: a rollup bar + colored percent + done/total
// + the epic id and description.
type epicDelegate struct{}

func (epicDelegate) Height() int                         { return 1 }
func (epicDelegate) Spacing() int                        { return 0 }
func (epicDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (epicDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(epicItem)
	if !ok {
		return
	}
	pct := it.es.Percent()
	bar := miniBar(pct, 8)
	pctStr := fg(theme.Percent(pct), fmt.Sprintf("%3d%%", pct))
	counts := fmt.Sprintf("%5s", fmt.Sprintf("%d/%d", it.es.Done, it.es.Total))
	idAndDesc := it.es.Epic.ID + "  " + dim(it.es.Epic.Description)
	row(w, m, index, fmt.Sprintf("%s %s %s  %s", bar, pctStr, counts, idAndDesc))
}

// --- audits ---

// auditItem adapts a domain.Audit to a list item.
type auditItem struct{ a domain.Audit }

func (i auditItem) FilterValue() string { return i.a.Slug + " " + i.a.Area }
func (i auditItem) id() string          { return i.a.Slug }
func (i auditItem) sortFields() sortFields {
	// Audits sort by date (as "updated") + slug; no priority/tier.
	return sortFields{updated: i.a.Date, slug: i.a.Slug}
}

// auditDelegate renders one audit row: a bucket-colored marker, the slug, the
// open/total finding count, and a dim area.
type auditDelegate struct{}

func (auditDelegate) Height() int                         { return 1 }
func (auditDelegate) Spacing() int                        { return 0 }
func (auditDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (auditDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(auditItem)
	if !ok {
		return
	}
	marker := fg(theme.Bucket(it.a.Bucket), "■")
	findings := fmt.Sprintf("%d/%d open", it.a.OpenFindings, it.a.Findings)
	row(w, m, index, fmt.Sprintf("%s %s  %s  %s", marker, it.a.Slug, findings, dim(it.a.Area)))
}
