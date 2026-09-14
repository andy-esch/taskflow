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

func TestSelectColumns_CanonicalNamesAndLegacyAliases(t *testing.T) {
	cols := TaskColumns()
	canonical, err := SelectColumns(cols, []string{"updated_at"})
	if err != nil {
		t.Fatalf("canonical updated_at selector: %v", err)
	}
	legacy, err := SelectColumns(cols, []string{"updated"})
	if err != nil {
		t.Fatalf("legacy updated selector: %v", err)
	}
	if canonical[0].selectorName() != "updated_at" || legacy[0].selectorName() != "updated_at" {
		t.Fatalf("canonical/legacy selectors did not resolve to updated_at: canonical=%q legacy=%q",
			canonical[0].selectorName(), legacy[0].selectorName())
	}
	if canonical[0].Name != "updated_at" || legacy[0].Name != "updated" {
		t.Errorf("canonical selection should echo its header while legacy keeps compatibility: canonical=%q legacy=%q",
			canonical[0].Name, legacy[0].Name)
	}
	neverEdited := domain.Task{Created: "2026-01-01"}
	if got := canonical[0].Extract(neverEdited); got != "" {
		t.Errorf("canonical table/CSV selector should expose raw updated_at, got %q", got)
	}
	if got := legacy[0].Extract(neverEdited); got != "2026-01-01" {
		t.Errorf("legacy table/CSV selector should retain the created fallback, got %q", got)
	}
	var table bytes.Buffer
	WriteTablePlain(&table, canonical, []domain.Task{neverEdited})
	if got := table.String(); got != "updated_at\n\n" {
		t.Errorf("canonical table should pair its header with the raw empty value, got %q", got)
	}
	table.Reset()
	WriteTablePlain(&table, legacy, []domain.Task{neverEdited})
	if got := table.String(); got != "updated\n2026-01-01\n" {
		t.Errorf("legacy table should retain its header and fallback, got %q", got)
	}
	var csv bytes.Buffer
	if err := WriteCSV(&csv, canonical, []domain.Task{neverEdited}); err != nil {
		t.Fatal(err)
	}
	if got := csv.String(); got != "updated_at\n\n" {
		t.Errorf("canonical CSV should pair its header with the raw empty value, got %q", got)
	}
	csv.Reset()
	if err := WriteCSV(&csv, legacy, []domain.Task{neverEdited}); err != nil {
		t.Fatal(err)
	}
	if got := csv.String(); got != "updated\n2026-01-01\n" {
		t.Errorf("legacy CSV should retain its header and fallback, got %q", got)
	}
	if canonical[0].projectedName() != "updated_at" || legacy[0].projectedName() != "updated" {
		t.Errorf("projected JSON should echo the selected spelling: canonical=%q legacy=%q",
			canonical[0].projectedName(), legacy[0].projectedName())
	}
	spec := Specs(cols)[5]
	if spec.Name != "updated_at" || len(spec.Aliases) != 1 || spec.Aliases[0] != "updated" {
		t.Errorf("completion/help should advertise canonical updated_at and retain its alias metadata: %#v", spec)
	}
	if _, err := SelectColumns(cols, []string{"updated", "updated_at"}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("alias plus canonical name should be rejected as a duplicate, got %v", err)
	}

	audit, err := SelectColumns(AuditColumns(), []string{"open"})
	if err != nil || audit[0].selectorName() != "open_findings" {
		t.Fatalf("legacy audit selector should resolve to open_findings: column=%q err=%v",
			audit[0].selectorName(), err)
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

func TestProjectedListJSON_UsesCanonicalWireKeysAndRawValues(t *testing.T) {
	tasks := []domain.Task{
		{Slug: "never-edited", Created: "2026-01-01"},
		{Slug: "edited", Created: "2026-01-01", Updated: "2026-02-03"},
	}
	// Select through the legacy presentation name to prove compatibility input
	// retains its established key without leaking the created-date fallback into
	// projected JSON.
	selected, err := SelectColumns(TaskColumns(), []string{"slug", "updated"})
	if err != nil {
		t.Fatal(err)
	}
	var projected bytes.Buffer
	if err := ProjectedListJSON(&projected, "tasks", selected, tasks, nil); err != nil {
		t.Fatal(err)
	}
	var narrow struct {
		Tasks []map[string]string `json:"tasks"`
	}
	if err := json.Unmarshal(projected.Bytes(), &narrow); err != nil {
		t.Fatalf("projected task JSON: %v\n%s", err, projected.String())
	}
	if len(narrow.Tasks) != 2 || narrow.Tasks[0]["updated"] != "" || narrow.Tasks[1]["updated"] != "2026-02-03" {
		t.Fatalf("legacy projection should preserve its key with raw values: %#v", narrow.Tasks)
	}
	if _, leaked := narrow.Tasks[0]["updated_at"]; leaked {
		t.Fatalf("legacy selector unexpectedly changed its output key: %#v", narrow.Tasks[0])
	}

	var full bytes.Buffer
	if err := TasksJSON(&full, tasks, nil); err != nil {
		t.Fatal(err)
	}
	var authoritative struct {
		Tasks []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal(full.Bytes(), &authoritative); err != nil {
		t.Fatal(err)
	}
	if _, invented := authoritative.Tasks[0]["updated_at"]; invented {
		t.Fatalf("full envelope should omit an absent updated_at: %#v", authoritative.Tasks[0])
	}
	if got := authoritative.Tasks[1]["updated_at"]; got != narrow.Tasks[1]["updated"] {
		t.Fatalf("projected updated value %v disagrees with full envelope %v", narrow.Tasks[1]["updated"], got)
	}

	researchSelected, err := SelectColumns(ResearchColumns(), []string{"updated_at"})
	if err != nil {
		t.Fatal(err)
	}
	projected.Reset()
	if err := ProjectedListJSON(&projected, "research", researchSelected,
		[]domain.Research{{Created: "2026-01-01"}}, nil); err != nil {
		t.Fatal(err)
	}
	if got := projected.String(); !strings.Contains(got, `"updated_at":""`) || strings.Contains(got, `"updated":`) {
		t.Fatalf("research projection should use raw canonical updated_at:\n%s", got)
	}

	auditSelected, err := SelectColumns(AuditColumns(), []string{"open_findings"})
	if err != nil {
		t.Fatal(err)
	}
	projected.Reset()
	if err := ProjectedListJSON(&projected, "audits", auditSelected,
		[]domain.Audit{{OpenFindings: 3}}, nil); err != nil {
		t.Fatal(err)
	}
	if got := projected.String(); !strings.Contains(got, `"open_findings":"3"`) || strings.Contains(got, `"open":`) {
		t.Fatalf("audit projection should use the canonical open_findings key:\n%s", got)
	}

	auditSelected, err = SelectColumns(AuditColumns(), []string{"open"})
	if err != nil {
		t.Fatal(err)
	}
	projected.Reset()
	if err := ProjectedListJSON(&projected, "audits", auditSelected,
		[]domain.Audit{{OpenFindings: 3}}, nil); err != nil {
		t.Fatal(err)
	}
	if got := projected.String(); !strings.Contains(got, `"open":"3"`) || strings.Contains(got, `"open_findings":`) {
		t.Fatalf("legacy audit projection should preserve its open key:\n%s", got)
	}
}

// TestColumnRegistries_FirstColumnIsQuietHandle pins the invariant renderList
// relies on for `-o name`/`-q`: the first column is the concise human handle,
// which is deliberately distinct from the durable task/audit id.
func TestColumnRegistries_FirstColumnIsQuietHandle(t *testing.T) {
	if got := TaskColumns()[0].Name; got != "slug" {
		t.Errorf("TaskColumns first column must be the quiet slug handle, got %q", got)
	}
	if got := EpicColumns()[0].Name; got != "id" {
		t.Errorf("EpicColumns first column must be the id, got %q", got)
	}
	if got := AuditColumns()[0].Name; got != "slug" {
		t.Errorf("AuditColumns first column must be the quiet slug handle, got %q", got)
	}
}

func TestWriteTablePlain_TaskExtractors(t *testing.T) {
	var b bytes.Buffer
	WriteTablePlain(&b, TaskColumns(), []domain.Task{{
		ID: "6ga000000001", Slug: "alpha", Status: domain.StatusInProgress, Tier: 2, Priority: "high",
		Epic: "20-cli", Updated: "2026-06-19", Description: "do the thing",
		RevisitAt: "2026-09-01",
	}})
	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	// revisit_at remains after description so the pre-existing columns keep their
	// positions; the stable id is the intentional new trailing column.
	if lines[0] != "slug\tstatus\ttier\tpriority\tepic\tupdated\tdescription\trevisit_at\tid" {
		t.Errorf("task header: %q", lines[0])
	}
	if lines[1] != "alpha\tin-progress\t2\thigh\t20-cli\t2026-06-19\tdo the thing\t2026-09-01\t6ga000000001" {
		t.Errorf("task row: %q", lines[1])
	}
}

// TestTaskRevisitAt_FlowsThroughCSVAndJSON pins that a SET revisit_at reaches the
// CSV cell and the full --json payload — the extractor (columns.go) and the DTO
// json tag (dto.go) are otherwise only ever exercised empty, so a broken
// extractor or json tag would slip past the other render tests.
func TestTaskRevisitAt_FlowsThroughCSVAndJSON(t *testing.T) {
	task := domain.Task{Slug: "snz", Status: domain.StatusDeferred, Tags: []string{"x"}, RevisitAt: "2026-09-01"}

	var cb bytes.Buffer
	if err := WriteCSV(&cb, TaskColumns(), []domain.Task{task}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(cb.String()), "\n")
	fields := strings.Split(lines[1], ",")
	if got := fields[len(fields)-2]; got != "2026-09-01" {
		t.Errorf("csv revisit_at cell = %q, want 2026-09-01:\n%s", got, cb.String())
	}

	var jb bytes.Buffer
	if err := TasksJSON(&jb, []domain.Task{task}, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(jb.String(), `"revisit_at":"2026-09-01"`) {
		t.Errorf("full --json should carry revisit_at:\n%s", jb.String())
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
		Epic: domain.Epic{ID: "20-cli", Status: "active", Priority: "medium", Description: "ux"},
		Done: 2, Total: 5,
	}})
	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	// percent + deprecated are appended LAST (after description) so the pre-existing
	// default columns kept their positions; done/total/percent/deprecated are plain
	// numbers (deprecated is 0 here — none set on the fixture).
	if lines[0] != "id\tstatus\tpriority\tdone\ttotal\tdescription\tpercent\tdeprecated" {
		t.Errorf("epic header: %q", lines[0])
	}
	if lines[1] != "20-cli\tactive\tmedium\t2\t5\tux\t40\t0" {
		t.Errorf("epic row: %q", lines[1])
	}
}

func TestWriteTablePlain_AuditExtractors(t *testing.T) {
	var b bytes.Buffer
	WriteTablePlain(&b, AuditColumns(), []domain.Audit{{
		ID: "6ga000000002", Slug: "2026-06-19-x", Bucket: domain.AuditOpen, Area: "cli",
		Date: "2026-06-19", Findings: 4, OpenFindings: 1,
	}})
	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	if lines[0] != "slug\tbucket\tarea\tdate\tfindings\topen\tid" {
		t.Errorf("audit header: %q", lines[0])
	}
	if lines[1] != "2026-06-19-x\topen\tcli\t2026-06-19\t4\t1\t6ga000000002" {
		t.Errorf("audit row: %q", lines[1])
	}
}

// TestWriteTablePlain_EmptyIsHeaderOnly pins the porcelain contract: a zero-row
// table still emits the header, so a consumer always gets a stable schema and
// detects "no rows" by line count.
func TestWriteTablePlain_EmptyIsHeaderOnly(t *testing.T) {
	var b bytes.Buffer
	WriteTablePlain(&b, TaskColumns(), nil)
	if got := strings.TrimSpace(b.String()); got != "slug\tstatus\ttier\tpriority\tepic\tupdated\tdescription\trevisit_at\tid" {
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
	if lines[0] != "slug,status,tier,priority,epic,updated,description,revisit_at,id" {
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
	if got := strings.TrimSpace(b.String()); got != "slug,bucket,area,date,findings,open,id" {
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
	cols := []Column[string]{column("v", "value", func(s string) string { return s })}
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
