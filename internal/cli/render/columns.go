package render

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
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

// WriteCSV writes the same projected columns as WriteTablePlain, but as RFC 4180
// CSV (encoding/csv quotes any cell containing a comma, quote, or newline) — the
// `-o csv` format, for spreadsheets. Like the table, an empty result still emits
// the header row. encoding/csv defaults to LF line endings, matching the rest of
// our output.
func WriteCSV[T any](w io.Writer, cols []Column[T], items []T) error {
	cw := csv.NewWriter(w)
	header := make([]string, len(cols))
	for i, c := range cols {
		header[i] = c.Name
	}
	if err := cw.Write(header); err != nil {
		return err
	}
	row := make([]string, len(cols))
	for _, it := range items {
		for i, c := range cols {
			row[i] = csvInjectionSafe(c.Extract(it))
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// projectedField is one key/value of a projected `--json -c` row. The value is
// the column extractor's string — the projection is a column VIEW, so it mirrors
// the table/csv cells (numbers and lists render as their string form), keeping
// the column registry the single source of truth for table, csv, and json alike.
type projectedField struct{ key, value string }

// projectedRow marshals as a JSON object whose keys stay in column (i.e. `-c`)
// order. A plain map can't: encoding/json sorts map keys, which would silently
// drop the requested ordering.
type projectedRow []projectedField

func (r projectedRow) MarshalJSON() ([]byte, error) {
	fields := make([]orderedField, len(r))
	for i, f := range r {
		fields[i] = orderedField{f.key, f.value}
	}
	return marshalOrderedObject(fields)
}

// orderedField is one key/value of an order-preserving JSON object.
type orderedField struct {
	key   string
	value any
}

// marshalOrderedObject renders a compact JSON object whose keys appear in the
// given order — what a struct gives for free but a map (sorted keys) does not.
func marshalOrderedObject(fields []orderedField) ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, f := range fields {
		if i > 0 {
			b.WriteByte(',')
		}
		k, err := json.Marshal(f.key)
		if err != nil {
			return nil, err
		}
		v, err := json.Marshal(f.value)
		if err != nil {
			return nil, err
		}
		b.Write(k)
		b.WriteByte(':')
		b.Write(v)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

// ProjectedListJSON writes a `--json -c …` projection: the standard versioned
// envelope (schema_version + the entity's listKey + an `unreadable` array only
// when non-empty, mirroring the full envelope's omitempty) but with each row
// narrowed to the selected columns as column-named string fields. listKey is the
// entity's envelope key ("tasks", "epics", …), so a projected list lands under
// the same key as its full envelope.
//
// This is a column VIEW (like -o table/csv), NOT the canonical typed envelope:
// every value is a string and rows omit unselected fields, so projected output
// does NOT validate against `schema --json-schema` (which describes the full
// envelopes). Only bare `--json` is the schema-validated contract.
func ProjectedListJSON[T any](w io.Writer, listKey string, cols []Column[T], items []T, problems []domain.FileProblem) error {
	rows := make([]projectedRow, 0, len(items))
	for _, it := range items {
		row := make(projectedRow, len(cols))
		for i, c := range cols {
			row[i] = projectedField{key: c.Name, value: c.Extract(it)}
		}
		rows = append(rows, row)
	}
	// schema_version first, then the entity list, then unreadable (only when
	// non-empty, mirroring the full envelope's omitempty) — a fixed, contract-
	// stable order a map's sorted keys wouldn't give.
	fields := []orderedField{
		{"schema_version", SchemaVersion},
		{listKey, rows},
	}
	if len(problems) > 0 {
		fields = append(fields, orderedField{"unreadable", problems})
	}
	b, err := marshalOrderedObject(fields)
	if err != nil {
		return err
	}
	b = append(b, '\n') // match encodeJSON's single trailing newline
	_, err = w.Write(b)
	return err
}

// csvInjectionSafe neutralizes spreadsheet formula injection: a cell whose first
// byte a spreadsheet treats as a formula (= + - @) or as a control prefix (tab,
// CR) is prefixed with a single quote so Excel/Sheets render it as literal text.
// Free-text cells (e.g. a finding title pasted from external review) are the risk;
// header names are fixed and safe, so only data cells are guarded.
func csvInjectionSafe(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + s
	}
	return s
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
		// percent is appended LAST so adding it didn't shift the pre-existing
		// default `epic list -o table`/`csv` columns (description stays column 6);
		// it's still `-c`-selectable in any position the caller asks for.
		{"percent", "rollup % complete", func(e core.EpicSummary) string { return fmt.Sprintf("%d", e.Percent()) }},
	}
}

// FindingColumns is the projectable column set for `audit findings`. The first
// column is the addressable id `audit:code` (what `-o name` projects) — unique
// across audits, unlike a bare finding code which repeats.
func FindingColumns() []Column[core.AuditFinding] {
	return []Column[core.AuditFinding]{
		{"ref", "addressable id: audit:code", func(f core.AuditFinding) string { return f.Audit + ":" + f.Code }},
		{"code", "finding code (H1/M2/…)", func(f core.AuditFinding) string { return f.Code }},
		{"audit", "audit slug", func(f core.AuditFinding) string { return f.Audit }},
		{"status", "finding status", func(f core.AuditFinding) string { return f.Status }},
		{"effort", "XS|S|M|L", func(f core.AuditFinding) string { return f.Effort }},
		{"urgency", "acute|soon|eventually", func(f core.AuditFinding) string { return f.Urgency }},
		{"component", "component", func(f core.AuditFinding) string { return f.Component }},
		{"file", "file:line", func(f core.AuditFinding) string { return f.File }},
		{"title", "finding title", func(f core.AuditFinding) string { return f.Title }},
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
