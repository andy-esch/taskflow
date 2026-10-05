package wire

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestImpactWireCopiesCoreDecisionsAndDependencyRemedy(t *testing.T) {
	const ownerGuidance = "inspect the owner's durable receipt before choosing the next operation"
	// Cover every gate/inconsistency pair, not just transitions from clear.
	// Otherwise a mapper using only After.Gate can agree accidentally with core.
	states := []core.TaskGraphState{
		{Gate: core.GateClear},
		{Gate: core.GateClear, Inconsistent: true},
		{Gate: core.GateBlocked},
		{Gate: core.GateBlocked, Inconsistent: true},
		{Gate: core.GateBroken},
		{Gate: core.GateBroken, Inconsistent: true},
	}
	for _, before := range states {
		for _, after := range states {
			t.Run(fmt.Sprintf("task/%s-%t_to_%s-%t", before.Gate, before.Inconsistent, after.Gate, after.Inconsistent), func(t *testing.T) {
				impact := core.TaskGraphStateImpact{TaskID: "6g0000000001", Before: before, After: after}
				// Role-only changes must not become warnings either.
				impact.Before.Role, impact.After.Role = core.RoleQueued, core.RoleCandidate
				dependency := ToDependencyMutationJSON(core.DependencyMutationReceipt{
					Impacts: []core.TaskGraphStateImpact{impact}, Remedy: ownerGuidance,
				}, WorkspaceJSON{})
				lifecycle := ToTaskLifecycleJSON(core.TaskLifecycleReceipt{
					Impacts: []core.TaskGraphStateImpact{impact}, Remedy: ownerGuidance,
				})
				if dependency.Remedy != ownerGuidance || lifecycle.Remedy != ownerGuidance ||
					len(dependency.Impacts) != 1 || len(lifecycle.Impacts) != 1 ||
					dependency.Impacts[0].NewlyUnsafe != impact.NewlyUnsafe() ||
					lifecycle.Impacts[0].NewlyUnsafe != impact.NewlyUnsafe() {
					t.Fatalf("wire re-derived/lost the core decision: dependency=%+v lifecycle=%+v", dependency, lifecycle)
				}
				data, err := json.Marshal(dependency)
				if err != nil || !strings.Contains(string(data), `"newly_unsafe":`) || !strings.Contains(string(data), `"remedy":`) {
					t.Fatalf("public decision/guidance disappeared: %s, %v", data, err)
				}
			})
		}
	}
	for _, before := range []bool{false, true} {
		for _, after := range []bool{false, true} {
			t.Run(fmt.Sprintf("thread/%t_to_%t", before, after), func(t *testing.T) {
				impact := core.ThreadProjectionImpact{ThreadID: "6g0000000002",
					Before: core.ThreadView{Inconsistent: before}, After: core.ThreadView{Inconsistent: after},
				}
				lifecycle := ToTaskLifecycleJSON(core.TaskLifecycleReceipt{
					ThreadImpacts: []core.ThreadProjectionImpact{impact}, Remedy: ownerGuidance,
				})
				if lifecycle.Remedy != ownerGuidance || len(lifecycle.ThreadImpacts) != 1 ||
					lifecycle.ThreadImpacts[0].NewlyInconsistent != impact.NewlyInconsistent() {
					t.Fatalf("wire re-derived/lost the core Thread decision: %+v", lifecycle)
				}
				data, err := json.Marshal(lifecycle)
				if err != nil || !strings.Contains(string(data), `"newly_inconsistent":`) || !strings.Contains(string(data), `"remedy":`) {
					t.Fatalf("public Thread decision/guidance disappeared: %s, %v", data, err)
				}
			})
		}
	}
	data, err := json.Marshal(ToDependencyMutationJSON(core.DependencyMutationReceipt{}, WorkspaceJSON{}))
	if err != nil || strings.Contains(string(data), `"remedy"`) {
		t.Fatalf("empty recovery intent must stay optional: %s, %v", data, err)
	}
}

// Shared impact DTOs also appear under repair and error envelopes. Populated
// converter fixtures prove their schema contract, not domain authorization.
func TestMutationImpactSchemaBranches(t *testing.T) {
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
	for _, flag := range []bool{false, true} {
		for _, remedy := range []string{"", "inspect the durable receipt before resuming"} {
			impact := core.TaskGraphStateImpact{TaskID: "6g0000000001",
				Before: core.TaskGraphState{Role: core.RoleCandidate, Gate: core.GateClear},
				After:  core.TaskGraphState{Role: core.RoleCandidate, Gate: core.GateClear},
			}
			if flag {
				impact.After.Gate = core.GateBlocked
			}
			beforeThread := core.ProjectThread(domain.Thread{ID: "6g0000000003", Slug: "schema-thread",
				Status: domain.ThreadStatusUnstarted, Tasks: []string{},
			}, core.NewTaskGraph(nil, nil))
			afterThread := beforeThread
			afterThread.Inconsistent = flag
			dependency := core.DependencyMutationReceipt{Operation: core.DependencyAdd,
				Impacts: []core.TaskGraphStateImpact{impact}, Remedy: remedy,
				PlannedTaskIDs: []string{impact.TaskID}, AppliedTaskIDs: []string{impact.TaskID},
			}
			lifecycle := core.TaskLifecycleReceipt{
				Task:    domain.Task{ID: impact.TaskID, Slug: "schema-task", Status: domain.StatusReadyToStart},
				Impacts: dependency.Impacts, Committed: true, Remedy: remedy,
				ThreadImpacts: []core.ThreadProjectionImpact{{ThreadID: beforeThread.Thread.ID,
					Before: beforeThread, After: afterThread, ChangedTaskIDs: []string{impact.TaskID},
				}},
			}
			repair := core.TaskGraphRepairReceipt{InitialHealth: core.GraphHealthy, FinalHealth: core.GraphHealthy,
				Impacts: dependency.Impacts, ThreadImpacts: lifecycle.ThreadImpacts,
			}
			depJSON := ToDependencyMutationJSON(dependency, WorkspaceJSON{})
			lifeJSON := ToTaskLifecycleRecoveryJSON(lifecycle, WorkspaceJSON{})
			repairJSON := ToTaskGraphRepairJSON(repair, WorkspaceJSON{})
			for _, tc := range []struct {
				name, definition     string
				value                any
				threads, remedyField bool
			}{
				{"dependency", "DependencyMutationEnvelope", ToDependencyMutationEnvelope(dependency, WorkspaceJSON{}), false, true},
				{"moves", "MovesEnvelope", ToMovesEnvelope([]MoveResult{ToTaskMoveResult(lifecycle)}, false, WorkspaceJSON{}), true, true},
				{"repair", "TaskGraphRepairEnvelope", ToTaskGraphRepairEnvelope(repair, WorkspaceJSON{}), true, false},
				{"error.dependency", "ErrorEnvelope", ErrorEnvelope{SchemaVersion: SchemaVersion,
					Error: ErrorItem{Code: "conflict", Message: "durable prefix", DependencyMutation: &depJSON}}, false, true},
				{"error.lifecycle", "ErrorEnvelope", ErrorEnvelope{SchemaVersion: SchemaVersion,
					Error: ErrorItem{Code: "conflict", Message: "committed", TaskLifecycle: &lifeJSON}}, true, true},
				{"error.repair", "ErrorEnvelope", ErrorEnvelope{SchemaVersion: SchemaVersion,
					Error: ErrorItem{Code: "conflict", Message: "partial", GraphRepair: &repairJSON}}, true, false},
			} {
				if !tc.remedyField && remedy != "" {
					continue // Repair has no remedy field; do not duplicate its cases.
				}
				t.Run(fmt.Sprintf("%s/flag=%t/remedy=%t", tc.name, flag, remedy != ""), func(t *testing.T) {
					schema, err := compiler.Compile(schemaID + "#/$defs/" + tc.definition)
					if err != nil {
						t.Fatal(err)
					}
					data, err := json.Marshal(tc.value)
					if err != nil {
						t.Fatal(err)
					}
					instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
					if err != nil {
						t.Fatal(err)
					}
					if err := schema.Validate(instance); err != nil {
						t.Fatalf("populated %s: %v\n%s", tc.definition, err, data)
					}
					if !bytes.Contains(data, fmt.Appendf(nil, `"newly_unsafe":%t`, flag)) ||
						(tc.threads && !bytes.Contains(data, fmt.Appendf(nil, `"newly_inconsistent":%t`, flag))) ||
						bytes.Contains(data, []byte(`"remedy":`)) != (tc.remedyField && remedy != "") {
						t.Fatalf("lost flags or incorrect optional guidance:\n%s", data)
					}
					keys := []string{"newly_unsafe"}
					if tc.threads {
						keys = append(keys, "newly_inconsistent")
					}
					for _, key := range keys {
						// Delete the property structurally, not by text replacement that
						// might create invalid JSON and falsely appear to pin the schema.
						bad, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
						if err != nil {
							t.Fatal(err)
						}
						if removed := removeImpactProperty(bad, key); removed == 0 {
							t.Fatalf("%s did not exercise %s", tc.name, key)
						}
						if err := schema.Validate(bad); err == nil || !strings.Contains(err.Error(), key) {
							t.Fatalf("missing required %s not rejected by its schema: %v", key, err)
						}
					}
				})
			}
		}
	}
}

func removeImpactProperty(value any, key string) int {
	removed := 0
	switch value := value.(type) {
	case map[string]any:
		if _, exists := value[key]; exists {
			delete(value, key)
			removed++
		}
		for _, child := range value {
			removed += removeImpactProperty(child, key)
		}
	case []any:
		for _, child := range value {
			removed += removeImpactProperty(child, key)
		}
	}
	return removed
}
