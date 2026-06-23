package render

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func TestSelectColumns(t *testing.T) {
	cols := TaskColumns()

	// Empty selection returns the full default set, in registry order.
	if got, err := SelectColumns(cols, nil); err != nil || len(got) != len(cols) {
		t.Fatalf("empty selection should be all %d columns (err=%v): got %d", len(cols), err, len(got))
	}

	// A selection projects to exactly those columns, in the requested order.
	got, err := SelectColumns(cols, []string{"status", "slug"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0].Name != "status" || got[1].Name != "slug" {
		t.Errorf("projection should preserve requested order: got %v", names(got))
	}

	// An unknown column is a validation error that names the offender + the menu.
	_, err = SelectColumns(cols, []string{"slug", "nope"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("unknown column should wrap ErrValidation, got %v", err)
	}
	if err == nil || !strings.Contains(err.Error(), "nope") || !strings.Contains(err.Error(), "slug") {
		t.Errorf("error should name the bad column and the available set: %v", err)
	}
}

// TestProjectedListJSON pins the `--json -c` contract: a schema_version-first
// envelope under the entity key, rows narrowed to the selected columns in -c
// order, values as the column extractors' strings, and `unreadable` omitted
// when there are no problems (mirroring the full envelope's omitempty).
func TestProjectedListJSON(t *testing.T) {
	tasks := []domain.Task{
		{Slug: "alpha", Status: domain.StatusInProgress, Tier: 2, Description: "first"},
		{Slug: "beta", Status: domain.StatusReadyToStart, Tier: 5, Description: "second"},
	}
	sel, err := SelectColumns(TaskColumns(), []string{"slug", "tier"})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := ProjectedListJSON(&buf, "tasks", sel, tasks, nil); err != nil {
		t.Fatalf("ProjectedListJSON: %v", err)
	}
	out := buf.String()

	// Compact (no indentation) with a single trailing newline.
	if strings.Contains(out, "\n  ") || strings.Count(out, "\n") != 1 {
		t.Errorf("projected JSON should be compact with one trailing newline:\n%q", out)
	}
	// schema_version comes first, before the entity key — fixed contract order.
	if sv, tk := strings.Index(out, "schema_version"), strings.Index(out, "\"tasks\""); sv < 0 || sv > tk {
		t.Errorf("schema_version must precede the entity key:\n%s", out)
	}
	// `tier` (an int column) renders as its string form — a column VIEW, like table/csv.
	if !strings.Contains(out, `"tier":"2"`) {
		t.Errorf("numeric column should render as a string in the projection:\n%s", out)
	}
	// No `unreadable` key when there are no problems.
	if strings.Contains(out, "unreadable") {
		t.Errorf("clean projection must omit unreadable:\n%s", out)
	}

	var got struct {
		SchemaVersion string           `json:"schema_version"`
		Tasks         []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if got.SchemaVersion != SchemaVersion {
		t.Errorf("schema_version = %q, want %q", got.SchemaVersion, SchemaVersion)
	}
	for _, row := range got.Tasks {
		if len(row) != 2 || row["slug"] == nil || row["tier"] == nil {
			t.Errorf("each row must carry exactly the selected slug+tier: %v", row)
		}
	}

	// With problems, `unreadable` appears (last).
	buf.Reset()
	probs := []domain.FileProblem{{Path: "tasks/x.md", Message: "bad frontmatter"}}
	if err := ProjectedListJSON(&buf, "tasks", sel, tasks, probs); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "unreadable") {
		t.Errorf("projection with problems must include unreadable:\n%s", buf.String())
	}
}

// TestColumnRegistries_FirstColumnIsID pins the invariant renderList relies on
// for `-o name`/`-q`: the first column of every registry is the id.
func TestColumnRegistries_FirstColumnIsID(t *testing.T) {
	if got := TaskColumns()[0].Name; got != "slug" {
		t.Errorf("TaskColumns first column must be the id (slug), got %q", got)
	}
	if got := EpicColumns()[0].Name; got != "id" {
		t.Errorf("EpicColumns first column must be the id, got %q", got)
	}
	if got := AuditColumns()[0].Name; got != "slug" {
		t.Errorf("AuditColumns first column must be the id (slug), got %q", got)
	}
}

func TestWriteTablePlain_TaskExtractors(t *testing.T) {
	var b bytes.Buffer
	WriteTablePlain(&b, TaskColumns(), []domain.Task{{
		Slug: "alpha", Status: domain.StatusInProgress, Tier: 2, Priority: "high",
		Epic: "20-cli", Updated: "2026-06-19", Description: "do the thing",
	}})
	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	if lines[0] != "slug\tstatus\ttier\tpriority\tepic\tupdated\tdescription" {
		t.Errorf("task header: %q", lines[0])
	}
	if lines[1] != "alpha\tin-progress\t2\thigh\t20-cli\t2026-06-19\tdo the thing" {
		t.Errorf("task row: %q", lines[1])
	}
}

// TestWriteTablePlain_UpdatedFallsBackToCreated covers the one non-trivial
// extractor: updated falls back to created when unset.
func TestWriteTablePlain_UpdatedFallsBackToCreated(t *testing.T) {
	var b bytes.Buffer
	WriteTablePlain(&b, TaskColumns(), []domain.Task{{Slug: "a", Created: "2026-01-01"}})
	if !strings.Contains(b.String(), "2026-01-01") {
		t.Errorf("updated should fall back to created:\n%s", b.String())
	}
}

func TestWriteTablePlain_EpicExtractors(t *testing.T) {
	var b bytes.Buffer
	WriteTablePlain(&b, EpicColumns(), []core.EpicSummary{{
		Epic: domain.Epic{ID: "20-cli", Status: "planning", Priority: "medium", Description: "ux"},
		Done: 2, Total: 5,
	}})
	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	// percent + deprecated are appended LAST (after description) so the pre-existing
	// default columns kept their positions; done/total/percent/deprecated are plain
	// numbers (deprecated is 0 here — none set on the fixture).
	if lines[0] != "id\tstatus\tpriority\tdone\ttotal\tdescription\tpercent\tdeprecated" {
		t.Errorf("epic header: %q", lines[0])
	}
	if lines[1] != "20-cli\tplanning\tmedium\t2\t5\tux\t40\t0" {
		t.Errorf("epic row: %q", lines[1])
	}
}

func TestWriteTablePlain_AuditExtractors(t *testing.T) {
	var b bytes.Buffer
	WriteTablePlain(&b, AuditColumns(), []domain.Audit{{
		Slug: "2026-06-19-x", Bucket: domain.AuditOpen, Area: "cli",
		Date: "2026-06-19", Findings: 4, OpenFindings: 1,
	}})
	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	if lines[0] != "slug\tbucket\tarea\tdate\tfindings\topen" {
		t.Errorf("audit header: %q", lines[0])
	}
	if lines[1] != "2026-06-19-x\topen\tcli\t2026-06-19\t4\t1" {
		t.Errorf("audit row: %q", lines[1])
	}
}

// TestWriteTablePlain_EmptyIsHeaderOnly pins the porcelain contract: a zero-row
// table still emits the header, so a consumer always gets a stable schema and
// detects "no rows" by line count.
func TestWriteTablePlain_EmptyIsHeaderOnly(t *testing.T) {
	var b bytes.Buffer
	WriteTablePlain(&b, TaskColumns(), nil)
	if got := strings.TrimSpace(b.String()); got != "slug\tstatus\ttier\tpriority\tepic\tupdated\tdescription" {
		t.Errorf("empty table should be header-only, got %q", got)
	}
}

func TestStyle_Link(t *testing.T) {
	on := NewStyle(true)
	got := on.Link("planning/x.md", "file:///abs/x.md")
	if !strings.Contains(got, "\x1b]8;;file:///abs/x.md\x1b\\") || !strings.Contains(got, "planning/x.md") {
		t.Errorf("enabled Link should embed an OSC 8 sequence + the text: %q", got)
	}
	// Off (pipe / --color=never): plain text, byte-stable, no escape sequences.
	if off := NewStyle(false).Link("planning/x.md", "file:///abs/x.md"); off != "planning/x.md" {
		t.Errorf("disabled Link should return plain text, got %q", off)
	}
}

func TestWriteCSV(t *testing.T) {
	var b bytes.Buffer
	if err := WriteCSV(&b, TaskColumns(), []domain.Task{
		{Slug: "a", Status: domain.StatusReadyToStart, Description: "has, a comma"},
	}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	if lines[0] != "slug,status,tier,priority,epic,updated,description" {
		t.Errorf("csv header: %q", lines[0])
	}
	// A cell containing a comma must be RFC 4180 quoted (this is exactly what
	// the tab-separated table can't express and why csv earns its place).
	if !strings.Contains(lines[1], `"has, a comma"`) {
		t.Errorf("comma cell should be quoted: %q", lines[1])
	}
}

func TestWriteCSV_EmptyIsHeaderOnly(t *testing.T) {
	var b bytes.Buffer
	if err := WriteCSV(&b, AuditColumns(), nil); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(b.String()); got != "slug,bucket,area,date,findings,open" {
		t.Errorf("empty csv should be header-only, got %q", got)
	}
}

func names[T any](cols []Column[T]) []string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = c.Name
	}
	return out
}

// TestWriteCSV_NeutralizesFormulaInjection pins L15 (2026-06-22 audit): cells whose
// first char a spreadsheet treats as a formula (= + - @) are prefixed with a quote
// so a shared CSV can't execute a pasted formula; safe cells are untouched.
func TestWriteCSV_NeutralizesFormulaInjection(t *testing.T) {
	cols := []Column[string]{{"v", "value", func(s string) string { return s }}}
	var buf bytes.Buffer
	if err := WriteCSV(&buf, cols, []string{"=SUM(A1)", "safe", "-1+2", "@cmd", "+x"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"'=SUM(A1)", "'-1+2", "'@cmd", "'+x"} {
		if !strings.Contains(out, want) {
			t.Errorf("CSV did not neutralize a formula-injection cell %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "\nsafe\n") {
		t.Errorf("a safe cell must be written unchanged:\n%s", out)
	}
}
