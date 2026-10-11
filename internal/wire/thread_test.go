package wire

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestThreadViewPlanningContextPreservesRawValuesAndCanonicalIdentity(t *testing.T) {
	items := []core.ThreadTaskView{
		{Role: core.ThreadTaskMember, Task: domain.Task{ID: "declared-drift", Slug: "small", Status: domain.StatusNextUp, Priority: "low", Tier: 5, Effort: "XS"}, State: core.TaskGraphState{TaskID: "6g0000000001"}},
		{Role: core.ThreadTaskMember, Task: domain.Task{ID: "6g0000000002", Slug: "unknown", Status: domain.StatusReadyToStart, Priority: "high", Tier: 1, Effort: "Unknown"}, State: core.TaskGraphState{TaskID: "6g0000000002"}},
		{Role: core.ThreadTaskMember, Task: domain.Task{ID: "6g0000000003", Slug: "unset", Status: domain.StatusNextUp}, State: core.TaskGraphState{TaskID: "6g0000000003"}},
	}
	gate := core.ThreadExternalGate{ThreadTaskView: core.ThreadTaskView{
		Role:  core.ThreadTaskExternalGate,
		Task:  domain.Task{ID: "6g0000000004", Priority: "medium", Tier: 3, Effort: "  raw\nestimate\x1b[31m"},
		State: core.TaskGraphState{TaskID: "6g0000000004"},
	}, Outstanding: true}
	view := core.ThreadView{Members: items, Frontier: items, ExternalGates: []core.ThreadExternalGate{gate}}
	update := ToThreadUpdateEnvelope(core.ThreadMutationReceipt{Before: view, After: view}, "", WorkspaceJSON{})
	impacts := toThreadProjectionImpactsJSON([]core.ThreadProjectionImpact{{Before: view, After: view}})
	for name, payload := range map[string]ThreadViewJSON{
		"list":                 ToThreadsEnvelope(core.ThreadListView{Threads: []core.ThreadView{view}}, nil).Threads[0],
		"show":                 ToThreadShowEnvelope(view, "body").View,
		"frontier":             ToThreadFrontierEnvelope(view).View,
		"graph":                ToThreadGraphEnvelope(core.ThreadGraphProjection{View: view}).Projection.View,
		"plan":                 ToThreadPlanEnvelope(core.ThreadGraphProjection{View: view}).Projection.View,
		"thread-update-before": update.Before,
		"thread-update-after":  update.After,
		"task-impact-before":   impacts[0].Before,
		"task-impact-after":    impacts[0].After,
	} {
		t.Run(name, func(t *testing.T) {
			for _, members := range [][]ThreadTaskJSON{payload.Members, payload.Frontier} {
				if len(members) != len(items) {
					t.Fatalf("lost tasks: %+v", members)
				}
				for i, item := range members {
					task := items[i].Task
					want := ThreadTaskPlanningJSON{Priority: task.Priority, Tier: task.Tier, Effort: task.Effort}
					if item.Planning != want || item.Task.TaskID != items[i].State.TaskID || item.Task.Status != string(task.Status) {
						t.Fatalf("task[%d] lost metadata, identity, or order: %+v want planning %+v", i, item, want)
					}
				}
			}
			if len(payload.ExternalGates) != 1 || !payload.ExternalGates[0].Outstanding || payload.ExternalGates[0].Role != "external-gate" ||
				payload.ExternalGates[0].Planning != (ThreadTaskPlanningJSON{Priority: "medium", Tier: 3, Effort: gate.Task.Effort}) {
				t.Fatalf("lost raw external prerequisite context: %+v", payload.ExternalGates)
			}
		})
	}
}

func TestThreadPlanningJSONSchemaAndUnsetRepresentation(t *testing.T) {
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
	schema, err := compiler.Compile(id + "#/$defs/ThreadFrontierEnvelope")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []ThreadTaskPlanningJSON{
		{Priority: "high", Tier: 1, Effort: "Unknown"},
		{Priority: "low", Tier: 5, Effort: "XS"},
		{},
	} {
		t.Run(want.Effort+want.Priority, func(t *testing.T) {
			item := core.ThreadTaskView{Role: core.ThreadTaskMember,
				Task:  domain.Task{ID: "6g0000000001", Status: domain.StatusNextUp, Priority: want.Priority, Tier: want.Tier, Effort: want.Effort},
				State: core.TaskGraphState{TaskID: "6g0000000001", Role: core.RoleQueued, Gate: core.GateClear, Eligible: true},
			}
			encoded, err := json.Marshal(ToThreadFrontierEnvelope(core.ThreadView{
				Thread:      domain.Thread{ID: "6g0000000000", Slug: "thread", Status: domain.ThreadStatusUnstarted},
				GraphHealth: core.GraphHealthy, ProjectionHealth: core.GraphHealthy,
				Members: []core.ThreadTaskView{item}, Frontier: []core.ThreadTaskView{item},
			}))
			if err != nil {
				t.Fatal(err)
			}
			var instance map[string]any
			if err := json.Unmarshal(encoded, &instance); err != nil {
				t.Fatal(err)
			}
			if err := schema.Validate(instance); err != nil {
				t.Fatalf("Thread envelope violates published schema: %v", err)
			}
			frontier := instance["view"].(map[string]any)["frontier"].([]any)
			planning := frontier[0].(map[string]any)["planning"].(map[string]any)
			if len(planning) != 3 || planning["priority"] != want.Priority || planning["tier"] != float64(want.Tier) || planning["effort"] != want.Effort {
				t.Fatalf("wire defaulted or omitted metadata: %+v want %+v", planning, want)
			}
			planning["tier"] = "1"
			if err := schema.Validate(instance); err == nil {
				t.Fatal("schema accepted a string-valued planning tier")
			}
			planning["tier"] = float64(want.Tier)
			delete(planning, "effort")
			if err := schema.Validate(instance); err == nil {
				t.Fatal("schema accepted omitted planning effort")
			}
		})
	}
}

func TestToThreadJSONPreservesTagOrderAndCanonicalizesMembership(t *testing.T) {
	payload := ToThreadJSON(domain.Thread{
		Tags: []string{"z", "a", "m"}, Tasks: []string{"6g0000000002", "6g0000000001"},
	})
	if !slices.Equal(payload.Tags, []string{"z", "a", "m"}) {
		t.Fatalf("tags = %v", payload.Tags)
	}
	if !slices.Equal(payload.Tasks, []string{"6g0000000001", "6g0000000002"}) {
		t.Fatalf("tasks = %v", payload.Tasks)
	}
}

func TestThreadViewEnvelopesPreserveDeclaredIDWhenSourceDiffers(t *testing.T) {
	view := core.ThreadView{
		Thread: domain.Thread{ID: "declared-thread-id", Slug: "drifted-thread"},
		Source: core.RecordSource{ID: "canonical-source-id", Location: "remote://misleading-thread"},
	}
	for name, payload := range map[string]ThreadViewJSON{
		"list":     ToThreadsEnvelope(core.ThreadListView{Threads: []core.ThreadView{view}}, nil).Threads[0],
		"show":     ToThreadShowEnvelope(view, "body").View,
		"frontier": ToThreadFrontierEnvelope(view).View,
	} {
		if payload.Thread.ID != view.Thread.ID || payload.Thread.ID == view.Source.ID {
			t.Errorf("%s thread.id = %q, want declared ID %q", name, payload.Thread.ID, view.Thread.ID)
		}
	}
	if projection := ToThreadGraphProjectionJSON(core.ThreadGraphProjection{View: view}); projection.View.Thread.ID != view.Thread.ID {
		t.Errorf("graph projection thread.id = %q, want declared ID %q", projection.View.Thread.ID, view.Thread.ID)
	}
}

func TestToThreadsEnvelopeRetainsPathlessIdentityWithoutParsingLocation(t *testing.T) {
	const sourceVersion = "opaque-thread-source-revision"
	payload := ToThreadsEnvelope(core.ThreadListView{}, []core.ThreadReadProblem{
		{ThreadID: "6g0000000002", ThreadSlug: "pathless", Message: "remote decode failed", SourceVersion: sourceVersion},
		{ThreadID: "6g0000000001", ThreadSlug: "explicit", Location: "opaque://6g9999999999-wrong", Message: "bad record"},
	})
	if payload.Unreadable == nil || len(payload.Unreadable) != 2 {
		t.Fatalf("unreadable = %+v", payload.Unreadable)
	}
	if payload.Unreadable[0].ThreadID != "6g0000000002" || payload.Unreadable[0].Location != "" {
		t.Fatalf("pathless problem = %+v", payload.Unreadable[0])
	}
	if payload.Unreadable[1].ThreadID != "6g0000000001" || payload.Unreadable[1].Location != "opaque://6g9999999999-wrong" {
		t.Fatalf("located problem = %+v", payload.Unreadable[1])
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), sourceVersion) || strings.Contains(string(encoded), "source_version") {
		t.Fatalf("opaque Thread source revision leaked through wire JSON: %s", encoded)
	}
}

func TestToThreadGraphProjectionJSONAddsTitleWithoutReplacingLabel(t *testing.T) {
	payload := ToThreadGraphProjectionJSON(core.ThreadGraphProjection{Nodes: []core.ThreadGraphNode{{
		TaskID: "6g0000000001", Label: "canonical-slug", Title: "Canonical TUI Title",
	}}})
	if len(payload.Nodes) != 1 || payload.Nodes[0].Label != "canonical-slug" ||
		payload.Nodes[0].Title != "Canonical TUI Title" {
		t.Fatalf("node = %+v", payload.Nodes)
	}
}

func TestToThreadGraphProjectionJSONKeepsOptionalNeighborhoodScopeExact(t *testing.T) {
	full := ToThreadGraphProjectionJSON(core.ThreadGraphProjection{})
	if full.Scope != nil {
		t.Fatalf("full graph unexpectedly has scope: %+v", full.Scope)
	}

	bounded := ToThreadGraphProjectionJSON(core.ThreadGraphProjection{Scope: &core.ThreadGraphScope{
		Kind: core.ThreadGraphScopeNeighborhood, FocalTaskID: "6g0000000001", Depth: 2,
		TotalNodes: 8, ShownNodes: 4, HiddenNodes: 4, TotalEdges: 10, ShownEdges: 3, HiddenEdges: 7,
		BoundaryEdges: []core.ThreadGraphEdge{{From: "6g0000000004", To: "6g0000000005"}},
	}})
	if bounded.Scope == nil || bounded.Scope.Kind != "neighborhood" || bounded.Scope.FocalTaskID != "6g0000000001" ||
		bounded.Scope.Depth != 2 || bounded.Scope.TotalNodes != 8 || bounded.Scope.ShownNodes != 4 ||
		bounded.Scope.HiddenNodes != 4 || bounded.Scope.TotalEdges != 10 || bounded.Scope.ShownEdges != 3 ||
		bounded.Scope.HiddenEdges != 7 || !slices.Equal(bounded.Scope.BoundaryEdges, []ThreadGraphEdgeJSON{{From: "6g0000000004", To: "6g0000000005"}}) {
		t.Fatalf("scope=%+v", bounded.Scope)
	}
	encoded, err := json.Marshal(full)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"scope"`) {
		t.Fatalf("full graph encoded optional scope: %s", encoded)
	}
}
