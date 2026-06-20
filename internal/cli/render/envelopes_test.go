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

	cases := []struct {
		def  string
		emit func(io.Writer) error
	}{
		{"TasksEnvelope", func(w io.Writer) error {
			return TasksJSON(w, []domain.Task{{Slug: "alpha", Status: domain.StatusReadyToStart, Tier: 2, Tags: []string{"x"}}}, nil)
		}},
		{"TaskShowEnvelope", func(w io.Writer) error {
			return TaskShowJSON(w, domain.Task{Slug: "alpha", Status: domain.StatusInProgress}, "# body")
		}},
		{"CreatedEnvelope", func(w io.Writer) error {
			return CreatedJSON(w, "task", "alpha", "ready-to-start", "tasks/ready-to-start/alpha.md", false)
		}},
		{"MovesEnvelope", func(w io.Writer) error {
			return MovesJSON(w, []MoveResult{{Slug: "alpha", To: "in-progress"}}, false)
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
