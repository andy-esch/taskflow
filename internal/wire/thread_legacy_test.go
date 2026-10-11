package wire

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestThreadLegacyJSONPreservesOccurrenceAcrossConsumers(t *testing.T) {
	legacy := []core.LegacyDependencyDiagnostic{{
		TaskID: "6g0000000001", Field: "blocked_by", Location: "db://tasks/outside",
		References: []core.LegacyReference{{Value: "outside prerequisite", Resolution: core.LegacyResolved,
			CandidateIDs: []string{"6g0000000002"}, Edge: core.DependencyEdge{From: "6g0000000002", To: "6g0000000001"}}},
	}, {TaskID: "6g0000000003", Field: "blocks"}}
	view := core.ThreadView{Thread: domain.Thread{ID: "6g0000000000", Slug: "thread", Status: domain.ThreadStatusUnstarted},
		GraphHealth: core.GraphDegraded, ProjectionHealth: core.GraphDegraded, GraphLegacy: legacy}
	update := ToThreadUpdateEnvelope(core.ThreadMutationReceipt{Before: view, After: view}, "", WorkspaceJSON{})
	impacts := toThreadProjectionImpactsJSON([]core.ThreadProjectionImpact{{Before: view, After: view}})
	want := ToLegacyDependenciesJSON(legacy)
	for name, result := range map[string]ThreadViewJSON{
		"show":          ToThreadShowEnvelope(view, "").View,
		"frontier":      ToThreadFrontierEnvelope(view).View,
		"graph":         ToThreadGraphEnvelope(core.ThreadGraphProjection{View: view}).Projection.View,
		"plan":          ToThreadPlanEnvelope(core.ThreadGraphProjection{View: view}).Projection.View,
		"update-before": update.Before, "update-after": update.After,
		"impact-before": impacts[0].Before, "impact-after": impacts[0].After,
	} {
		t.Run(name, func(t *testing.T) {
			if !reflect.DeepEqual(result.GraphLegacy, want) || result.GraphLegacy[0].Path != "" || result.GraphLegacy[0].Location != legacy[0].Location || result.GraphLegacy[1].References == nil || result.Frontier == nil {
				t.Fatalf("lost portable legacy context/defaults: %+v", result)
			}
		})
	}
	list := ToThreadsEnvelope(core.ThreadListView{GraphHealth: core.GraphDegraded, GraphLegacy: legacy, Threads: []core.ThreadView{{Thread: view.Thread}}}, nil)
	if !reflect.DeepEqual(list.GraphLegacy, want) || len(list.Threads[0].GraphLegacy) != 0 || list.Threads[0].GraphLegacy == nil {
		t.Fatalf("list did not retain hoisted context: %+v", list)
	}
	local := ToLegacyDependenciesJSON([]core.LegacyDependencyDiagnostic{{TaskID: "local", Field: "blocks", TaskPath: "tasks/local.md", Location: "tasks/local.md"}})
	if local[0].Path != "tasks/local.md" || local[0].Location != "" {
		t.Fatalf("local path was duplicated as opaque context: %+v", local)
	}

	schemaBytes, err := JSONSchema()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
	if err != nil {
		t.Fatal(err)
	}
	id := doc.(map[string]any)["$id"].(string)
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(id, doc); err != nil {
		t.Fatal(err)
	}
	for name, envelope := range map[string]any{"ThreadFrontierEnvelope": ToThreadFrontierEnvelope(view), "ThreadsEnvelope": list} {
		schema, err := compiler.Compile(id + "#/$defs/" + name)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(envelope)
		if err != nil {
			t.Fatal(err)
		}
		instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(encoded))
		if err != nil {
			t.Fatal(err)
		}
		if err := schema.Validate(instance); err != nil {
			t.Fatalf("populated %s failed emitted schema: %v", name, err)
		}
		context := instance.(map[string]any)
		if name == "ThreadFrontierEnvelope" {
			context = context["view"].(map[string]any)
		}
		diagnostic := context["graph_legacy"].([]any)[0].(map[string]any)
		for _, field := range []string{"location", "references"} {
			original := diagnostic[field]
			diagnostic[field] = 42
			if err := schema.Validate(instance); err == nil {
				t.Fatalf("%s schema accepted invalid legacy %s", name, field)
			}
			diagnostic[field] = original
		}
	}
}
