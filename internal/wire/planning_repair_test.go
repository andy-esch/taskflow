package wire

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func TestPlanningRepairEnvelopeSeparatesPartialRelocationFromFixedRows(t *testing.T) {
	result := core.PlanningRepairResult{
		Fixes: []domain.FixResult{{Path: "earlier.md", Changes: []string{"tags normalized"}}},
		Relocation: &core.FrontmatterRepairRelocation{
			SourcePath: "/planning/bad.md", DestinationPath: "/planning/corrected.md", DestinationWritten: true,
		},
	}
	envelope := ToPlanningRepairEnvelope(result, false, WorkspaceJSON{})
	if len(envelope.Fixed) != 1 || envelope.Fixed[0].Path != "earlier.md" || envelope.Relocation == nil || envelope.Relocation.SourcePath != result.Relocation.SourcePath || envelope.Relocation.DestinationPath != result.Relocation.DestinationPath || !envelope.Relocation.DestinationWritten || envelope.Relocation.SourceRemoved {
		t.Fatalf("incorrect partial receipt: %+v", envelope)
	}
	result.Relocation.DestinationPath = "mutated source evidence"
	if envelope.Relocation.DestinationPath != "/planning/corrected.md" {
		t.Fatal("wire retained mutable core outcome")
	}
	ordinary, err := json.Marshal(ToPlanningRepairEnvelope(core.PlanningRepairResult{}, true, WorkspaceJSON{}))
	if err != nil || strings.Contains(string(ordinary), "relocation") || !strings.Contains(string(ordinary), `"fixed":[]`) {
		t.Fatalf("ordinary preview shape drift: %s err=%v", ordinary, err)
	}
}

// The DTO copies historical action flags; it does not infer current existence
// or authorize producer states. False flags must remain explicit, not omitted.
func TestPlanningRepairRelocationSchemaRequiresExplicitEvidence(t *testing.T) {
	schemaBytes, err := JSONSchema()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
	if err != nil {
		t.Fatal(err)
	}
	schemaID := doc.(map[string]any)["$id"].(string)
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(schemaID, doc); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(schemaID + "#/$defs/FrontmatterRepairRelocationJSON")
	if err != nil {
		t.Fatal(err)
	}
	for _, written := range []bool{false, true} {
		for _, removed := range []bool{false, true} {
			t.Run(fmt.Sprintf("written=%t/removed=%t", written, removed), func(t *testing.T) {
				envelope := ToPlanningRepairEnvelope(core.PlanningRepairResult{
					Relocation: &core.FrontmatterRepairRelocation{
						SourcePath: "/planning/old.md", DestinationPath: "/planning/new.md",
						DestinationWritten: written, SourceRemoved: removed,
					},
				}, false, WorkspaceJSON{})
				data, err := json.Marshal(envelope.Relocation)
				if err != nil {
					t.Fatal(err)
				}
				var fields map[string]any
				if err := json.Unmarshal(data, &fields); err != nil {
					t.Fatal(err)
				}
				if len(fields) != 4 || fields["source_path"] != "/planning/old.md" || fields["destination_path"] != "/planning/new.md" || fields["destination_written"] != written || fields["source_removed"] != removed {
					t.Fatalf("mapper lost or invented evidence: %s", data)
				}
				if err := schema.Validate(fields); err != nil {
					t.Fatalf("explicit flags violate schema: %v", err)
				}
				for _, name := range []string{"source_path", "destination_path", "destination_written", "source_removed"} {
					missing := maps.Clone(fields)
					delete(missing, name)
					if err := schema.Validate(missing); err == nil {
						t.Errorf("schema accepted missing %s", name)
					}
				}
				for _, name := range []string{"destination_written", "source_removed"} {
					wrongType := maps.Clone(fields)
					wrongType[name] = "false"
					if err := schema.Validate(wrongType); err == nil {
						t.Errorf("schema accepted string-valued %s", name)
					}
				}
			})
		}
	}
}
