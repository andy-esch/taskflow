package tui

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// The list rows (taskDelegate / epicDelegate / auditDelegate) are the most-seen UI
// in the tool, yet only the detail panes were tested. These render a one-item list
// and assert the row's COLUMNS — present and in left-to-right order — so a dropped
// glyph, a missing bar, or a reordered column fails CI. They reference the theme +
// render helpers (theme.Status/Bucket, miniBar/segBar, theme.Counts) rather than
// hardcoding glyphs, so they pin row LAYOUT without duplicating the glyph→colour
// decision table (that's theme's own tests' job).

// renderDelegateRow renders one delegate row to an ANSI-stripped string. It renders
// at index 1 while the fresh list's selected index is 0, so the shared "› " cursor
// prefix is absent and the assertions see just the content columns.
func renderDelegateRow(t *testing.T, d list.ItemDelegate, it list.Item, width int) string {
	t.Helper()
	m := list.New([]list.Item{it}, d, width, 10)
	var buf bytes.Buffer
	d.Render(&buf, m, 1, it)
	return ansi.Strip(buf.String())
}

// assertColumns checks every part is present AND appears in the given left-to-right
// order (first-occurrence indexing) — so a dropped column OR a reordered one fails.
// Parts must be distinct enough not to collide.
func assertColumns(t *testing.T, row string, parts ...string) {
	t.Helper()
	prev, prevPart := -1, ""
	for _, p := range parts {
		i := strings.Index(row, p)
		if i < 0 {
			t.Fatalf("row is missing column %q\nrow: %q", p, row)
		}
		if i < prev {
			t.Errorf("column %q precedes %q (wrong order)\nrow: %q", p, prevPart, row)
		}
		prev, prevPart = i, p
	}
}

// A task row: status glyph, slug, relative date — in that order.
func TestTaskDelegateRow(t *testing.T) {
	task := domain.Task{Slug: "alpha", Status: domain.StatusInProgress, Updated: "2020-01-02"}
	row := renderDelegateRow(t, taskDelegate{st: &testStyles}, taskItem{t: task}, 80)

	wantDate := theme.RelativeDate(theme.TaskDate(task)) // computed the same way the row does
	if wantDate == "" {
		t.Fatal("setup: expected a non-empty relative date")
	}
	assertColumns(t, row, theme.Status(domain.StatusInProgress).Glyph, "alpha", wantDate)
}

// An epic row: liveness glyph, rollup bar, colored percent, done/total, id, dim
// description — in that order.
func TestEpicDelegateRow(t *testing.T) {
	es := core.EpicSummary{
		Epic:  domain.Epic{ID: "17-pm-go-cli", Status: "active", Description: "planning tool"},
		Total: 4, Done: 1,
	}
	row := renderDelegateRow(t, epicDelegate{st: &testStyles}, epicItem{es: es}, 80)

	wantGlyph := ansi.Strip(epicGlyph(es, &testStyles))                  // working-band glyph, via production logic
	wantBar := ansi.Strip(testStyles.miniBar(es.Percent(), 8))           // the 8-cell rollup bar
	wantPct := strings.TrimSpace(theme.PercentLabelPadded(es.Percent())) // "25%"
	wantCounts := theme.Counts(es.Done, es.Total)                        // "1/4"
	assertColumns(t, row, wantGlyph, wantBar, wantPct, wantCounts, "17-pm-go-cli", "planning tool")
}

// An audit row: bucket glyph, segmented bar, colored percent, resolved/total, slug,
// dim area — in that order.
func TestAuditDelegateRow(t *testing.T) {
	a := domain.Audit{
		Slug: "2026-06-20-api", Area: "gateway", Bucket: domain.AuditOpen,
		Findings: 4, DoneFindings: 1, ActiveFindings: 1, OpenFindings: 2,
	}
	row := renderDelegateRow(t, auditDelegate{st: &testStyles}, auditItem{a: a}, 80)

	wantBar := ansi.Strip(testStyles.segBar(a.DoneFindings, a.ActiveFindings, a.DroppedFindings, a.Findings, 8))
	wantPct := strings.TrimSpace(theme.PercentLabelPadded(a.Percent())) // "25%"
	wantCounts := theme.Counts(a.Resolved(), a.Findings)                // "1/4"
	assertColumns(t, row, theme.Bucket(domain.AuditOpen).Glyph, wantBar, wantPct, wantCounts, "2026-06-20-api", "gateway")
}
