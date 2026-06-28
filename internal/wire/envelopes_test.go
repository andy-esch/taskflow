package wire

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// emit encodes the envelope value a constructor returns, so each case proves the
// constructor's output (the same value render's *JSON funcs encode, and the value a
// web handler would wrap) validates against the schema.
func emit(w io.Writer, v any) error { return EncodeJSON(w, v) }

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
	epic := domain.Epic{ID: "e1", Status: "active", Description: "d"}
	epicSum := core.EpicSummary{Epic: epic, Total: 2, Done: 1}

	// Every envelope, validated against its own $defs entry — the whole --json
	// contract, not a sample. The embedded-struct envelopes (schema/schema_kind,
	// epic rollup) are here precisely because reflection of embedded fields is the
	// likeliest place the schema and the real output drift.
	cases := []struct {
		def  string
		emit func(io.Writer) error
	}{
		{"TasksEnvelope", func(w io.Writer) error { return emit(w, ToTasksEnvelope([]domain.Task{task}, nil)) }},
		{"TaskShowEnvelope", func(w io.Writer) error { return emit(w, ToTaskShowEnvelope(task, "# body")) }},
		{"TaskMutationEnvelope", func(w io.Writer) error { return emit(w, ToTaskMutationEnvelope(task, "# new body", true)) }},
		{"EpicMutationEnvelope", func(w io.Writer) error { return emit(w, ToEpicMutationEnvelope(epic, true)) }},
		{"CreatedEnvelope", func(w io.Writer) error {
			return emit(w, ToCreatedEnvelope("task", "alpha", "ready-to-start", "tasks/ready-to-start/alpha.md", false))
		}},
		{"MovesEnvelope", func(w io.Writer) error {
			return emit(w, ToMovesEnvelope([]MoveResult{{Slug: "alpha", To: "in-progress"}}, false))
		}},
		{"SummaryEnvelope", func(w io.Writer) error {
			return emit(w, ToSummaryEnvelope(core.Summary{
				Counts:     []core.StatusCount{{Status: domain.StatusInProgress, Count: 1}},
				InProgress: []domain.Task{task},
				Epics:      []core.EpicSummary{epicSum},
			}))
		}},
		{"VersionEnvelope", func(w io.Writer) error { return emit(w, ToVersionEnvelope("v0.6.0")) }},
		{"EpicsEnvelope", func(w io.Writer) error { return emit(w, ToEpicsEnvelope([]core.EpicSummary{epicSum}, nil)) }},
		{"EpicShowEnvelope", func(w io.Writer) error {
			return emit(w, ToEpicShowEnvelope(epic, []domain.Task{task}, "# body"))
		}},
		{"AuditsEnvelope", func(w io.Writer) error {
			return emit(w, ToAuditsEnvelope([]domain.Audit{{Slug: "x", Bucket: domain.AuditOpen, Findings: 1, OpenFindings: 1}}, nil))
		}},
		{"AuditShowEnvelope", func(w io.Writer) error {
			return emit(w, ToAuditShowEnvelope(domain.Audit{Slug: "x", Bucket: domain.AuditOpen, Findings: 2, OpenFindings: 1}, "# body"))
		}},
		{"FindingsEnvelope", func(w io.Writer) error {
			return emit(w, ToFindingsEnvelope([]core.AuditFinding{{
				Finding: domain.Finding{Code: "S1", Title: "tighten the gateway", Status: "open", Effort: "S", Urgency: "soon"},
				Audit:   "2026-01-01-area", Bucket: "open",
			}}, nil))
		}},
		{"LintEnvelope", func(w io.Writer) error {
			return emit(w, ToLintEnvelope([]core.LintResult{{Slug: "alpha", Issues: []domain.Issue{{Field: "epic", Message: "missing"}}}}, nil))
		}},
		{"FixEnvelope", func(w io.Writer) error {
			return emit(w, ToFixEnvelope(nil, nil, nil, false)) // the nil-slice path: must emit [] and validate
		}},
		{"InitEnvelope", func(w io.Writer) error {
			return emit(w, NormalizeInitEnvelope(InitEnvelope{Mode: "scaffold", Root: "/root", Created: []string{"tasks"}}))
		}},
		{"DoctorEnvelope", func(w io.Writer) error {
			return emit(w, ToDoctorEnvelope("/root", []DoctorProblem{{Repo: "../impl", Message: "one-sided link"}}))
		}},
		{"SchemaEnvelope", func(w io.Writer) error {
			return emit(w, ToSchemaEnvelope(SchemaContract{
				Statuses:     []SchemaStatus{{Value: "in-progress", Active: true}},
				EpicStatuses: []string{"active"},
				AuditBuckets: []string{"open"},
				TaskFields:   []SchemaField{{Name: "tier", Type: "int"}},
				EpicFields:   []string{"status", "description"},
				ExitCodes:    []SchemaExitCode{{Code: 10, Name: "not-found"}},
				Kinds:        []string{"task"},
			}))
		}},
		{"SchemaKindEnvelope", func(w io.Writer) error {
			return emit(w, ToSchemaKindEnvelope(KindSchema{
				Kind:         "task",
				Sections:     []string{"Objective"},
				BodyTemplate: "## Objective\n",
				Fields:       []domain.FieldDoc{{Name: "tier", Type: "int", Required: true, Description: "d", Example: "3"}},
				Conventions:  []string{"c"},
				Templates:    []TemplateInfo{{Kind: "task", Name: "default", Description: "d"}},
			}))
		}},
		{"TemplatesEnvelope", func(w io.Writer) error {
			return emit(w, ToTemplatesEnvelope([]TemplateInfo{{Kind: "task", Name: "default", Description: "d"}}))
		}},
		{"TemplateShowEnvelope", func(w io.Writer) error {
			return emit(w, ToTemplateShowEnvelope(TemplateInfo{Kind: "task", Name: "default", Description: "d"}, "# body"))
		}},
		{"ErrorEnvelope", func(w io.Writer) error {
			// Built by cli.WriteError (not a constructor here) — marshal the named type
			// directly to prove its schema matches.
			return emit(w, ErrorEnvelope{SchemaVersion: SchemaVersion, Error: ErrorItem{Code: "not-found", Message: "task not found"}})
		}},
	}
	// Registry-derived coverage guard (replaces a brittle literal count): every
	// envelope type the jsonEnvelopes registry pulls into the schema must have a
	// case here, so a newly-added envelope can't be silently left unvalidated. The
	// $defs key is the Go type name, which is also each case's `def`. ErrorEnvelope
	// is built by cli.WriteError (not a constructor here) but is still a registered
	// envelope with a case, so it's covered too.
	covered := make(map[string]bool, len(cases))
	for _, tc := range cases {
		covered[tc.def] = true
	}
	rt := reflect.TypeOf(Envelopes())
	for i := range rt.NumField() {
		def := rt.Field(i).Type.Name()
		if !covered[def] {
			t.Errorf("envelope %q is in the jsonEnvelopes registry but has no validation case", def)
		}
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
