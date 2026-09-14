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

// Column is one projectable field shared by the list formats. Name and Extract
// retain the stable table/CSV header and display value. projection is an optional
// canonical wire-contract override for `--json -c`; when present,
// both the canonical JSON name and the legacy presentation name are accepted as
// selectors. selectedName records which spelling an explicit projection used so
// legacy JSON callers keep their key while canonical callers get the wire key.
// This keeps compatibility without letting a canonical header carry a display
// fallback. The first column is the concise command handle (slug / epic id),
// which `-o name` projects; a later column may carry a distinct durable id.
type Column[T any] struct {
	Name         string
	Desc         string
	Extract      func(T) string
	projection   *columnProjection[T]
	selectedName string
}

// columnProjection keeps a canonical selector inseparable from the raw value it
// claims to expose. A nil projection means Name/Extract are already canonical.
// The fields stay package-private so contractColumn is the production construction
// path; validateColumnRegistry still fails closed against malformed declarations.
type columnProjection[T any] struct {
	name    string
	extract func(T) string
}

func column[T any](name, desc string, extract func(T) string) Column[T] {
	return Column[T]{Name: name, Desc: desc, Extract: extract}
}

func contractColumn[T any](name, jsonName, desc string, display, project func(T) string) Column[T] {
	if jsonName == "" || project == nil {
		panic("render: a contract column requires both a canonical selector and raw projector")
	}
	return Column[T]{
		Name: name, Desc: desc, Extract: display,
		projection: &columnProjection[T]{name: jsonName, extract: project},
	}
}

func (c Column[T]) selectorName() string {
	if c.projection != nil {
		return c.projection.name
	}
	return c.Name
}

func (c Column[T]) projectedValue(item T) string {
	if c.projection != nil {
		return c.projection.extract(item)
	}
	return c.Extract(item)
}

func (c Column[T]) projectedName() string {
	if c.selectedName != "" {
		return c.selectedName
	}
	return c.selectorName()
}

// ColumnSpec is the name+description of a column without the (typed) extractor,
// so the cli completion/help layer can offer and describe columns without the
// generic type parameter.
type ColumnSpec struct {
	Name    string
	Desc    string
	Aliases []string
}

type columnSelectorOwner struct {
	column string
	index  int
}

// validateColumnRegistry rejects declaration bugs before a selector map can
// silently let a later column shadow an earlier canonical name or alias.
func validateColumnRegistry[T any](cols []Column[T]) error {
	seen := make(map[string]columnSelectorOwner, len(cols)*2)
	for i, c := range cols {
		if c.Name == "" {
			return fmt.Errorf("invalid column registry: column %d has an empty display name", i)
		}
		if c.Extract == nil {
			return fmt.Errorf("invalid column registry: column %q has no display extractor", c.Name)
		}
		if c.projection != nil {
			if c.projection.name == "" {
				return fmt.Errorf("invalid column registry: column %q has an empty canonical selector", c.Name)
			}
			if c.projection.extract == nil {
				return fmt.Errorf("invalid column registry: canonical selector %q has no raw projector", c.projection.name)
			}
		}

		selectors := []string{c.selectorName()}
		if c.Name != c.selectorName() {
			selectors = append(selectors, c.Name)
		}
		for _, selector := range selectors {
			if prior, ok := seen[selector]; ok {
				return fmt.Errorf(
					"invalid column registry: selector %q for column %q at index %d collides with column %q at index %d",
					selector, c.Name, i, prior.column, prior.index,
				)
			}
			seen[selector] = columnSelectorOwner{column: c.Name, index: i}
		}
	}
	return nil
}

func mustValidateColumnRegistry[T any](cols []Column[T]) {
	if err := validateColumnRegistry(cols); err != nil {
		// Official registries are static program declarations, not user input. A
		// malformed one is a programmer error that must fail during command setup;
		// SelectColumns returns the same validation as an error for caller-supplied
		// or test-constructed slices.
		panic(err)
	}
}

func columnRegistry[T any](cols ...Column[T]) []Column[T] {
	mustValidateColumnRegistry(cols)
	return cols
}

// Specs projects a typed column set to its name/description pairs.
func Specs[T any](cols []Column[T]) []ColumnSpec {
	mustValidateColumnRegistry(cols)
	out := make([]ColumnSpec, len(cols))
	for i, c := range cols {
		out[i] = ColumnSpec{Name: c.selectorName(), Desc: c.Desc}
		if c.Name != c.selectorName() {
			out[i].Aliases = []string{c.Name}
		}
	}
	return out
}

// SelectColumns returns the columns named by `names`, in that order; empty
// `names` returns all (the default table). An unknown name is a validation error
// listing the available columns.
func SelectColumns[T any](all []Column[T], names []string) ([]Column[T], error) {
	if err := validateColumnRegistry(all); err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return all, nil
	}
	byName := make(map[string]Column[T], len(all)*2)
	for _, c := range all {
		byName[c.selectorName()] = c
		if c.Name != c.selectorName() {
			byName[c.Name] = c // compatibility alias for the old table/CSV name
		}
	}
	out := make([]Column[T], 0, len(names))
	selected := make(map[string]string, len(names))
	for _, n := range names {
		c, ok := byName[n]
		if !ok {
			return nil, fmt.Errorf("%w: unknown column %q (available: %s)",
				domain.ErrValidation, n, columnNames(all))
		}
		canonical := c.selectorName()
		if prior, exists := selected[canonical]; exists {
			return nil, fmt.Errorf("%w: duplicate column %q (already selected as %q)",
				domain.ErrValidation, n, prior)
		}
		selected[canonical] = n
		c.selectedName = n
		// Preserve legacy/default headers, but echo an explicitly requested
		// canonical selector in table/CSV with the canonical value semantics.
		// Otherwise an `updated_at` header could carry the display-only created
		// fallback. An explicitly selected legacy alias retains its old header and
		// display value, and projected JSON retains the requested key while using
		// the raw projected value.
		if n == canonical && c.projection != nil {
			c.Name = canonical
			c.Extract = c.projection.extract
		}
		out = append(out, c)
	}
	return out, nil
}

// columnNames joins the column names for help/error text.
func columnNames[T any](cols []Column[T]) string {
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.selectorName()
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

// projectedField is one key/value of a projected `--json -c` row. Values remain
// strings, but a column may supply a canonical wire key/value distinct from its
// human table/CSV presentation.
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
// This is a string-valued VIEW, NOT the canonical typed envelope: rows omit
// unselected fields, so projected output does NOT validate against
// `schema --json-schema` (which describes the full envelopes). Columns that
// represent wire fields use their canonical keys and source values when selected
// canonically. Explicit compatibility aliases retain their requested output key
// but still use the raw projected value. Presentation-only fallbacks are reserved
// for default or explicitly legacy table/CSV selections. Only bare `--json` is
// the schema-validated contract.
func ProjectedListJSON[T any](w io.Writer, listKey string, cols []Column[T], items []T, problems []domain.FileProblem) error {
	rows := make([]projectedRow, 0, len(items))
	for _, it := range items {
		row := make(projectedRow, len(cols))
		for i, c := range cols {
			row[i] = projectedField{key: c.projectedName(), value: c.projectedValue(it)}
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
	b = append(b, '\n') // match wire.EncodeJSON's single trailing newline
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

// TaskColumns is the projectable column set for `task list`. Slug stays first as
// the human-facing handle projected by `-o name`; the durable id is selectable
// explicitly and appended last so established columns retain their positions.
func TaskColumns() []Column[domain.Task] {
	return columnRegistry(
		column("slug", "task slug", func(t domain.Task) string { return t.Slug }),
		column("status", "lifecycle status", func(t domain.Task) string { return string(t.Status) }),
		column("tier", "priority tier 1-5", func(t domain.Task) string { return fmt.Sprintf("%d", t.Tier) }),
		column("priority", "high|medium|low", func(t domain.Task) string { return t.Priority }),
		column("epic", "parent epic id", func(t domain.Task) string { return t.Epic }),
		contractColumn("updated", "updated_at", "last-updated date", func(t domain.Task) string {
			if t.Updated != "" {
				return t.Updated
			}
			return t.Created
		}, func(t domain.Task) string { return t.Updated }),
		column("description", "one-line summary", func(t domain.Task) string { return t.Description }),
		// revisit_at remains after the original columns so their positions stay
		// stable (description is column 7); id is the new trailing addition. Both
		// are `-c`-selectable in any order the caller asks.
		column("revisit_at", "snooze-until date (deferred tasks)", func(t domain.Task) string { return t.RevisitAt }),
		column("id", "stable task identifier", func(t domain.Task) string { return t.ID }),
	)
}

// EpicColumns is the projectable column set for `epic list` (id first; done/total
// as plain numbers, not the human "2/3 (66%)" cell).
func EpicColumns() []Column[core.EpicSummary] {
	return columnRegistry(
		column("id", "epic identifier", func(e core.EpicSummary) string { return e.Epic.ID }),
		column("status", "epic status", func(e core.EpicSummary) string { return e.Epic.Status }),
		column("priority", "high|medium|low", func(e core.EpicSummary) string { return e.Epic.Priority }),
		column("done", "completed task count", func(e core.EpicSummary) string { return fmt.Sprintf("%d", e.Done) }),
		column("total", "total task count", func(e core.EpicSummary) string { return fmt.Sprintf("%d", e.Total) }),
		column("description", "one-line summary", func(e core.EpicSummary) string { return e.Epic.Description }),
		// percent/deprecated are appended LAST so adding them didn't shift the
		// pre-existing default `epic list -o table`/`csv` columns (description stays
		// column 6); both are still `-c`-selectable in any position the caller asks.
		column("percent", "rollup % complete", func(e core.EpicSummary) string { return fmt.Sprintf("%d", e.Percent()) }),
		column("deprecated", "withdrawn (excluded) task count", func(e core.EpicSummary) string { return fmt.Sprintf("%d", e.Deprecated) }),
	)
}

// FindingColumns is the projectable column set for `audit findings`. The first
// column is the addressable id `audit:code` (what `-o name` projects) — unique
// across audits, unlike a bare finding code which repeats.
func FindingColumns() []Column[core.AuditFinding] {
	return columnRegistry(
		column("ref", "addressable id: audit:code", func(f core.AuditFinding) string { return f.Audit + ":" + f.Code }),
		column("code", "finding code (H1/M2/…)", func(f core.AuditFinding) string { return f.Code }),
		column("audit", "audit slug", func(f core.AuditFinding) string { return f.Audit }),
		column("status", "finding status", func(f core.AuditFinding) string { return f.Status }),
		column("effort", "XS|S|M|L", func(f core.AuditFinding) string { return f.Effort }),
		column("urgency", "acute|soon|eventually", func(f core.AuditFinding) string { return f.Urgency }),
		column("component", "component", func(f core.AuditFinding) string { return f.Component }),
		column("file", "file:line", func(f core.AuditFinding) string { return f.File }),
		column("title", "finding title", func(f core.AuditFinding) string { return f.Title }),
	)
}

// ResearchColumns are the projectable columns for `research list -o table/csv` and
// `--json -c`. No status/bucket column exists because research has no lifecycle.
func ResearchColumns() []Column[domain.Research] {
	return columnRegistry(
		column("slug", "research identifier", func(r domain.Research) string { return r.Slug }),
		column("created", "date the research was done", func(r domain.Research) string { return r.Created }),
		column("description", "one-line summary", func(r domain.Research) string { return r.Description }),
		column("tags", "topical tags", func(r domain.Research) string { return strings.Join(r.Tags, ",") }),
		// Falls back to created when never edited, matching the task `updated` column: a doc
		// written once has been "last touched" on its created date, and an empty cell here
		// would sort last under an updated sort for no good reason.
		contractColumn("updated", "updated_at", "last-updated date", func(r domain.Research) string {
			if r.Updated != "" {
				return r.Updated
			}
			return r.Created
		}, func(r domain.Research) string { return r.Updated }),
		column("id", "stable identifier", func(r domain.Research) string { return r.ID }),
	)
}

// AuditColumns is the projectable column set for `audit list`. Slug stays first
// as the human-facing handle projected by `-o name`; the durable id is appended
// last for explicit selection without shifting established columns.
func AuditColumns() []Column[domain.Audit] {
	return columnRegistry(
		column("slug", "audit slug", func(a domain.Audit) string { return a.Slug }),
		column("bucket", "open|closed|deferred", func(a domain.Audit) string { return string(a.Bucket) }),
		column("area", "area under audit", func(a domain.Audit) string { return a.Area }),
		column("date", "audit date", func(a domain.Audit) string { return a.Date }),
		column("findings", "total findings", func(a domain.Audit) string { return fmt.Sprintf("%d", a.Findings) }),
		contractColumn("open", "open_findings", "open findings",
			func(a domain.Audit) string { return fmt.Sprintf("%d", a.OpenFindings) },
			func(a domain.Audit) string { return fmt.Sprintf("%d", a.OpenFindings) }),
		column("id", "stable audit identifier", func(a domain.Audit) string { return a.ID }),
	)
}
