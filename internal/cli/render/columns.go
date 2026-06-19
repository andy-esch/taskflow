package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// Column is one projectable column of a list table: a machine name (the header
// in `-o table`, and what `-c/--columns` selects), a short description (for
// completion), and an extractor. A per-entity []Column is the single source of
// truth for the default table columns, `-c` validation, `-c` completion, and the
// projection itself — so the four can't drift. The first column is the id (slug
// / epic id), which `-o name` projects.
type Column[T any] struct {
	Name    string
	Desc    string
	Extract func(T) string
}

// ColumnSpec is the name+description of a column without the (typed) extractor,
// so the cli completion/help layer can offer and describe columns without the
// generic type parameter.
type ColumnSpec struct{ Name, Desc string }

// Specs projects a typed column set to its name/description pairs.
func Specs[T any](cols []Column[T]) []ColumnSpec {
	out := make([]ColumnSpec, len(cols))
	for i, c := range cols {
		out[i] = ColumnSpec{Name: c.Name, Desc: c.Desc}
	}
	return out
}

// SelectColumns returns the columns named by `names`, in that order; empty
// `names` returns all (the default table). An unknown name is a validation error
// listing the available columns.
func SelectColumns[T any](all []Column[T], names []string) ([]Column[T], error) {
	if len(names) == 0 {
		return all, nil
	}
	byName := make(map[string]Column[T], len(all))
	for _, c := range all {
		byName[c.Name] = c
	}
	out := make([]Column[T], 0, len(names))
	for _, n := range names {
		c, ok := byName[n]
		if !ok {
			return nil, fmt.Errorf("%w: unknown column %q (available: %s)",
				domain.ErrValidation, n, columnNames(all))
		}
		out = append(out, c)
	}
	return out, nil
}

// columnNames joins the column names for help/error text.
func columnNames[T any](cols []Column[T]) string {
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.Name
	}
	return strings.Join(names, ", ")
}

// WriteTablePlain writes the stable tab-separated table for the given columns: a
// header row of column names, then one tab-separated row per item. No ANSI, no
// truncation — the documented `-o table` contract.
func WriteTablePlain[T any](w io.Writer, cols []Column[T], items []T) {
	header := make([]string, len(cols))
	for i, c := range cols {
		header[i] = c.Name
	}
	rows := make([][]string, 0, len(items))
	for _, it := range items {
		row := make([]string, len(cols))
		for i, c := range cols {
			row[i] = c.Extract(it)
		}
		rows = append(rows, row)
	}
	writePlain(w, header, rows)
}

// TaskColumns is the projectable column set for `task list` (slug first — the id
// projected by `-o name`).
func TaskColumns() []Column[domain.Task] {
	return []Column[domain.Task]{
		{"slug", "task identifier", func(t domain.Task) string { return t.Slug }},
		{"status", "lifecycle status", func(t domain.Task) string { return string(t.Status) }},
		{"tier", "priority tier 1-5", func(t domain.Task) string { return fmt.Sprintf("%d", t.Tier) }},
		{"priority", "high|medium|low", func(t domain.Task) string { return t.Priority }},
		{"epic", "parent epic id", func(t domain.Task) string { return t.Epic }},
		{"updated", "last-updated date", func(t domain.Task) string {
			if t.Updated != "" {
				return t.Updated
			}
			return t.Created
		}},
		{"description", "one-line summary", func(t domain.Task) string { return t.Description }},
	}
}

// EpicColumns is the projectable column set for `epic list` (id first; done/total
// as plain numbers, not the human "2/3 (66%)" cell).
func EpicColumns() []Column[core.EpicSummary] {
	return []Column[core.EpicSummary]{
		{"id", "epic identifier", func(e core.EpicSummary) string { return e.Epic.ID }},
		{"status", "epic status", func(e core.EpicSummary) string { return e.Epic.Status }},
		{"priority", "high|medium|low", func(e core.EpicSummary) string { return e.Epic.Priority }},
		{"done", "completed task count", func(e core.EpicSummary) string { return fmt.Sprintf("%d", e.Done) }},
		{"total", "total task count", func(e core.EpicSummary) string { return fmt.Sprintf("%d", e.Total) }},
		{"description", "one-line summary", func(e core.EpicSummary) string { return e.Epic.Description }},
	}
}

// AuditColumns is the projectable column set for `audit list` (slug first).
func AuditColumns() []Column[domain.Audit] {
	return []Column[domain.Audit]{
		{"slug", "audit identifier", func(a domain.Audit) string { return a.Slug }},
		{"bucket", "open|closed|deferred", func(a domain.Audit) string { return string(a.Bucket) }},
		{"area", "area under audit", func(a domain.Audit) string { return a.Area }},
		{"date", "audit date", func(a domain.Audit) string { return a.Date }},
		{"findings", "total findings", func(a domain.Audit) string { return fmt.Sprintf("%d", a.Findings) }},
		{"open", "open findings", func(a domain.Audit) string { return fmt.Sprintf("%d", a.OpenFindings) }},
	}
}
