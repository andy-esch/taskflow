package render

import (
	"bytes"
	"io"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// TestJSONSchema_ValidatesRealOutput is the round-trip proof: the emitted schema
// actually validates real --json output across a representative spread of
// envelopes (list, show, mutation, nested item, lint, and the nil-slice fix path).
func TestJSONSchema_ValidatesRealOutput(t *testing.T) {
	schemaBytes, err := JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema: %v", err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
	if err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	id := doc.(map[string]any)["$id"].(string)
	c := jsonschema.NewCompiler()
	if err := c.AddResource(id, doc); err != nil {
		t.Fatalf("add resource: %v", err)
	}

	task := domain.Task{Slug: "alpha", Status: domain.StatusInProgress, Tier: 2, Tags: []string{"x"}}
	epic := domain.Epic{ID: "e1", Status: "in-progress", Description: "d"}
	epicSum := core.EpicSummary{Epic: epic, Total: 2, Done: 1}

	// Every envelope, validated against its own $defs entry — the whole --json
	// contract, not a sample. The embedded-struct envelopes (schema/schema_kind,
	// epic rollup) are here precisely because reflection of embedded fields is the
	// likeliest place the schema and the real output drift.
	cases := []struct {
		def  string
		emit func(io.Writer) error
	}{
		{"TasksEnvelope", func(w io.Writer) error { return TasksJSON(w, []domain.Task{task}, nil) }},
		{"TaskShowEnvelope", func(w io.Writer) error { return TaskShowJSON(w, task, "# body") }},
		{"CreatedEnvelope", func(w io.Writer) error {
			return CreatedJSON(w, "task", "alpha", "ready-to-start", "tasks/ready-to-start/alpha.md", false)
		}},
		{"MovesEnvelope", func(w io.Writer) error {
			return MovesJSON(w, []MoveResult{{Slug: "alpha", To: "in-progress"}}, false)
		}},
		{"SummaryEnvelope", func(w io.Writer) error {
			return SummaryJSON(w, core.Summary{
				Counts:     []core.StatusCount{{Status: domain.StatusInProgress, Count: 1}},
				InProgress: []domain.Task{task},
				Epics:      []core.EpicSummary{epicSum},
			})
		}},
		{"VersionEnvelope", func(w io.Writer) error { return VersionJSON(w, "v0.6.0") }},
		{"EpicsEnvelope", func(w io.Writer) error { return EpicsJSON(w, []core.EpicSummary{epicSum}, nil) }},
		{"EpicShowEnvelope", func(w io.Writer) error {
			return EpicShowJSON(w, epic, []domain.Task{task}, "# body")
		}},
		{"AuditsEnvelope", func(w io.Writer) error {
			return AuditsJSON(w, []domain.Audit{{Slug: "x", Bucket: domain.AuditOpen, Findings: 1, OpenFindings: 1}}, nil)
		}},
		{"AuditShowEnvelope", func(w io.Writer) error {
			return AuditShowJSON(w, domain.Audit{Slug: "x", Bucket: domain.AuditOpen, Findings: 2, OpenFindings: 1}, "# body")
		}},
		{"LintEnvelope", func(w io.Writer) error {
			return LintJSON(w, []core.LintResult{{Slug: "alpha", Issues: []domain.Issue{{Field: "epic", Message: "missing"}}}}, nil)
		}},
		{"FixEnvelope", func(w io.Writer) error {
			return FixJSON(w, nil, nil, false) // the nil-slice path: must emit [] and validate
		}},
		{"InitEnvelope", func(w io.Writer) error { return InitJSON(w, "/root", []string{"tasks"}, false) }},
		{"SchemaEnvelope", func(w io.Writer) error {
			return SchemaJSON(w, SchemaContract{
				Statuses:     []SchemaStatus{{Value: "in-progress", Active: true}},
				EpicStatuses: []string{"in-progress"},
				AuditBuckets: []string{"open"},
				TaskFields:   []SchemaField{{Name: "tier", Type: "int"}},
				ExitCodes:    []SchemaExitCode{{Code: 10, Name: "not-found"}},
				Kinds:        []string{"task"},
			})
		}},
		{"SchemaKindEnvelope", func(w io.Writer) error {
			return SchemaKindJSON(w, KindSchema{
				Kind:         "task",
				Sections:     []string{"Objective"},
				BodyTemplate: "## Objective\n",
				Fields:       []domain.FieldDoc{{Name: "tier", Type: "int", Required: true, Description: "d", Example: "3"}},
				Conventions:  []string{"c"},
			})
		}},
		{"ErrorEnvelope", func(w io.Writer) error {
			// Not emitted by a render func (cli.WriteError builds it) — marshal the
			// named type directly to prove its schema matches.
			return encodeJSON(w, ErrorEnvelope{SchemaVersion: SchemaVersion, Error: ErrorItem{Code: "not-found", Message: "task not found"}})
		}},
	}
	if len(cases) != 16 {
		t.Fatalf("expected all 16 envelopes covered, got %d", len(cases))
	}
	for _, tc := range cases {
		sch, err := c.Compile(id + "#/$defs/" + tc.def)
		if err != nil {
			t.Errorf("compile %s: %v", tc.def, err)
			continue
		}
		var buf bytes.Buffer
		if err := tc.emit(&buf); err != nil {
			t.Errorf("emit %s: %v", tc.def, err)
			continue
		}
		inst, err := jsonschema.UnmarshalJSON(&buf)
		if err != nil {
			t.Errorf("unmarshal %s output: %v", tc.def, err)
			continue
		}
		if err := sch.Validate(inst); err != nil {
			t.Errorf("%s output does NOT validate against its own schema:\n%v", tc.def, err)
		}
	}
}
