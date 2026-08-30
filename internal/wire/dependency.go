package wire

import (
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// GraphTaskJSON is the compact task identity used by graph-query envelopes.
// TaskID remains present even when the underlying task is unreadable.
type GraphTaskJSON struct {
	TaskID string `json:"task_id" jsonschema:"description=canonical stable task identifier"`
	Slug   string `json:"slug,omitempty" jsonschema:"description=human task slug when the task is readable"`
	Status string `json:"status,omitempty" jsonschema:"description=persisted lifecycle status when readable"`
	Epic   string `json:"epic,omitempty" jsonschema:"description=epic identifier when readable"`
}

func toGraphTaskJSON(taskID string, task domain.Task) GraphTaskJSON {
	return GraphTaskJSON{TaskID: taskID, Slug: task.Slug, Status: string(task.Status), Epic: task.Epic}
}

// TaskGraphStateJSON is the derived lifecycle/gate projection for one task.
// Eligible is true for queued (next-up) or candidate (ready-to-start) work only
// when the authoritative repository graph is healthy and the gate is clear.
type TaskGraphStateJSON struct {
	Role             string `json:"role"`
	Gate             string `json:"gate"`
	SoundlyCompleted bool   `json:"soundly_completed"`
	Eligible         bool   `json:"eligible" jsonschema:"description=true only for next-up or ready-to-start work with a clear gate in a healthy graph"`
	Drained          bool   `json:"drained"`
	Inconsistent     bool   `json:"inconsistent"`
}

func toTaskGraphStateJSON(state core.TaskGraphState) TaskGraphStateJSON {
	return TaskGraphStateJSON{
		Role: string(state.Role), Gate: string(state.Gate), SoundlyCompleted: state.SoundlyCompleted,
		Eligible: state.Eligible, Drained: state.Drained, Inconsistent: state.Inconsistent,
	}
}

// GraphProblemJSON is taskflow-owned repository-graph diagnostic data.
type GraphProblemJSON struct {
	Code          string   `json:"code"`
	TaskID        string   `json:"task_id,omitempty"`
	RelatedTaskID string   `json:"related_task_id,omitempty"`
	Field         string   `json:"field,omitempty"`
	Path          string   `json:"path,omitempty"`
	Message       string   `json:"message"`
	Cycle         []string `json:"cycle,omitempty"`
}

func toGraphProblemsJSON(problems []core.GraphProblem) []GraphProblemJSON {
	out := make([]GraphProblemJSON, 0, len(problems))
	for _, problem := range problems {
		out = append(out, GraphProblemJSON{
			Code: string(problem.Code), TaskID: problem.TaskID, RelatedTaskID: problem.RelatedTaskID,
			Field: problem.Field, Path: problem.Path, Message: problem.Message,
			Cycle: append([]string(nil), problem.Cycle...),
		})
	}
	return out
}

// LegacyReferenceJSON records one legacy dependency resolution.
type LegacyReferenceJSON struct {
	Value          string   `json:"value"`
	Resolution     string   `json:"resolution"`
	CandidateIDs   []string `json:"candidate_ids"`
	PrerequisiteID string   `json:"prerequisite_id,omitempty"`
	DependentID    string   `json:"dependent_id,omitempty"`
}

// LegacyDependencyJSON is one legacy field occurrence reported by a graph query.
type LegacyDependencyJSON struct {
	TaskID     string                `json:"task_id"`
	TaskSlug   string                `json:"task_slug,omitempty"`
	Path       string                `json:"path,omitempty"`
	Field      string                `json:"field"`
	References []LegacyReferenceJSON `json:"references"`
}

func toLegacyDependenciesJSON(diagnostics []core.LegacyDependencyDiagnostic) []LegacyDependencyJSON {
	out := make([]LegacyDependencyJSON, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		item := LegacyDependencyJSON{
			TaskID: diagnostic.TaskID, TaskSlug: diagnostic.TaskSlug, Path: diagnostic.TaskPath,
			Field: diagnostic.Field, References: make([]LegacyReferenceJSON, 0, len(diagnostic.References)),
		}
		for _, ref := range diagnostic.References {
			item.References = append(item.References, LegacyReferenceJSON{
				Value: ref.Value, Resolution: string(ref.Resolution), CandidateIDs: append([]string{}, ref.CandidateIDs...),
				PrerequisiteID: ref.Edge.From, DependentID: ref.Edge.To,
			})
		}
		out = append(out, item)
	}
	return out
}

// TaskBlockerJSON is one direct/transitive blocker with explanatory path.
type TaskBlockerJSON struct {
	Task   GraphTaskJSON      `json:"task"`
	State  TaskGraphStateJSON `json:"state"`
	Reason string             `json:"reason"`
	Path   []string           `json:"path"`
	Direct bool               `json:"direct"`
}

// TaskBlockersEnvelope is `task blockers --json`.
type TaskBlockersEnvelope struct {
	SchemaVersion string                 `json:"schema_version"`
	Task          GraphTaskJSON          `json:"task"`
	State         TaskGraphStateJSON     `json:"state"`
	Projection    string                 `json:"projection"`
	Health        string                 `json:"health"`
	Blockers      []TaskBlockerJSON      `json:"blockers"`
	Problems      []GraphProblemJSON     `json:"problems"`
	Legacy        []LegacyDependencyJSON `json:"legacy_dependencies"`
}

// ToTaskBlockersEnvelope maps the core diagnostic query into its wire contract.
func ToTaskBlockersEnvelope(result core.TaskBlockersResult) TaskBlockersEnvelope {
	envelope := TaskBlockersEnvelope{
		SchemaVersion: SchemaVersion, Task: toGraphTaskJSON(result.TaskID, result.Task),
		State: toTaskGraphStateJSON(result.State), Projection: result.Projection, Health: string(result.Health),
		Blockers: make([]TaskBlockerJSON, 0, len(result.Blockers)),
		Problems: toGraphProblemsJSON(result.Problems), Legacy: toLegacyDependenciesJSON(result.Legacy),
	}
	for _, blocker := range result.Blockers {
		envelope.Blockers = append(envelope.Blockers, TaskBlockerJSON{
			Task: toGraphTaskJSON(blocker.Blocker.TaskID, blocker.Task), State: toTaskGraphStateJSON(blocker.State),
			Reason: string(blocker.Blocker.Reason), Path: append([]string(nil), blocker.Blocker.Path...), Direct: blocker.Blocker.Direct,
		})
	}
	return envelope
}

// TaskUnblockJSON is one transitive downstream dependent and its current state.
type TaskUnblockJSON struct {
	Task   GraphTaskJSON      `json:"task"`
	State  TaskGraphStateJSON `json:"state"`
	Path   []string           `json:"path"`
	Direct bool               `json:"direct"`
}

// TaskUnblocksEnvelope is `task unblocks --json`.
type TaskUnblocksEnvelope struct {
	SchemaVersion string                 `json:"schema_version"`
	Task          GraphTaskJSON          `json:"task"`
	State         TaskGraphStateJSON     `json:"state"`
	Health        string                 `json:"health"`
	Unblocks      []TaskUnblockJSON      `json:"unblocks"`
	Problems      []GraphProblemJSON     `json:"problems"`
	Legacy        []LegacyDependencyJSON `json:"legacy_dependencies"`
}

// ToTaskUnblocksEnvelope maps the downstream-impact query into its wire contract.
func ToTaskUnblocksEnvelope(result core.TaskUnblocksResult) TaskUnblocksEnvelope {
	envelope := TaskUnblocksEnvelope{
		SchemaVersion: SchemaVersion, Task: toGraphTaskJSON(result.TaskID, result.Task),
		State: toTaskGraphStateJSON(result.State), Health: string(result.Health), Unblocks: make([]TaskUnblockJSON, 0, len(result.Unblocks)),
		Problems: toGraphProblemsJSON(result.Problems), Legacy: toLegacyDependenciesJSON(result.Legacy),
	}
	for _, dependent := range result.Unblocks {
		envelope.Unblocks = append(envelope.Unblocks, TaskUnblockJSON{
			Task: toGraphTaskJSON(dependent.Impact.TaskID, dependent.Task), State: toTaskGraphStateJSON(dependent.State),
			Path: append([]string(nil), dependent.Impact.Path...), Direct: dependent.Impact.Direct,
		})
	}
	return envelope
}

// DependencyEdgeOutcomeJSON is one canonical edge result in a mutation receipt.
type DependencyEdgeOutcomeJSON struct {
	DependentID    string `json:"dependent_id"`
	PrerequisiteID string `json:"prerequisite_id"`
	Action         string `json:"action"`
	Outcome        string `json:"outcome"`
}

// LegacyFieldClearJSON identifies one legacy field occurrence cleared by migration.
type LegacyFieldClearJSON struct {
	TaskID string `json:"task_id"`
	Field  string `json:"field"`
}

// TaskGraphStateImpactJSON is one before/after derived-state consequence of a
// graph or lifecycle mutation.
type TaskGraphStateImpactJSON struct {
	TaskID string             `json:"task_id"`
	Before TaskGraphStateJSON `json:"before"`
	After  TaskGraphStateJSON `json:"after"`
	Path   []string           `json:"path,omitempty"`
	Direct bool               `json:"direct"`
}

func toTaskGraphStateImpactsJSON(impacts []core.TaskGraphStateImpact) []TaskGraphStateImpactJSON {
	out := make([]TaskGraphStateImpactJSON, 0, len(impacts))
	for _, impact := range impacts {
		out = append(out, TaskGraphStateImpactJSON{
			TaskID: impact.TaskID, Before: toTaskGraphStateJSON(impact.Before), After: toTaskGraphStateJSON(impact.After),
			Path: append([]string(nil), impact.Path...), Direct: impact.Direct,
		})
	}
	return out
}

// TaskLifecycleBlockerJSON is one blocker deliberately retained by a forced
// transition receipt.
type TaskLifecycleBlockerJSON struct {
	TaskID string   `json:"task_id"`
	Reason string   `json:"reason"`
	Path   []string `json:"path"`
	Direct bool     `json:"direct"`
}

// TaskLifecycleJSON is the task-specific detail nested into the shared move
// envelope. Audit and epic transitions omit it.
type TaskLifecycleJSON struct {
	TaskID              string                       `json:"task_id"`
	From                string                       `json:"from"`
	Changed             bool                         `json:"changed"`
	Committed           bool                         `json:"committed" jsonschema:"description=true only after the lifecycle write became durable"`
	Forced              bool                         `json:"forced"`
	Override            string                       `json:"override,omitempty"`
	Before              TaskGraphStateJSON           `json:"before"`
	After               TaskGraphStateJSON           `json:"after"`
	OutstandingBlockers []TaskLifecycleBlockerJSON   `json:"outstanding_blockers"`
	ImpactCount         int                          `json:"impact_count"`
	Impacts             []TaskGraphStateImpactJSON   `json:"impacts"`
	ThreadImpactCount   int                          `json:"thread_impact_count"`
	ThreadImpacts       []ThreadProjectionImpactJSON `json:"thread_impacts"`
	Remedy              string                       `json:"remedy,omitempty"`
}

// ThreadProjectionImpactJSON is one Thread view changed by a task lifecycle
// transition; direct distinguishes membership from downstream/boundary effects.
type ThreadProjectionImpactJSON struct {
	ThreadID       string         `json:"thread_id"`
	Slug           string         `json:"slug"`
	Direct         bool           `json:"direct"`
	ChangedTaskIDs []string       `json:"changed_task_ids"`
	Before         ThreadViewJSON `json:"before"`
	After          ThreadViewJSON `json:"after"`
}

func toThreadProjectionImpactsJSON(impacts []core.ThreadProjectionImpact) []ThreadProjectionImpactJSON {
	out := make([]ThreadProjectionImpactJSON, 0, len(impacts))
	for _, impact := range impacts {
		out = append(out, ThreadProjectionImpactJSON{
			ThreadID: impact.ThreadID, Slug: impact.Slug, Direct: impact.Direct,
			ChangedTaskIDs: append([]string{}, impact.ChangedTaskIDs...),
			Before:         ToThreadViewJSON(impact.Before), After: ToThreadViewJSON(impact.After),
		})
	}
	return out
}

func ToTaskLifecycleJSON(receipt core.TaskLifecycleReceipt) TaskLifecycleJSON {
	payload := TaskLifecycleJSON{
		TaskID: receipt.Task.ID, From: string(receipt.From), Changed: receipt.Changed,
		Committed: receipt.Committed,
		Forced:    receipt.Forced, Override: string(receipt.Override),
		Before: toTaskGraphStateJSON(receipt.Before), After: toTaskGraphStateJSON(receipt.After),
		OutstandingBlockers: make([]TaskLifecycleBlockerJSON, 0, len(receipt.OutstandingBlockers)),
		ImpactCount:         len(receipt.Impacts),
		Impacts:             toTaskGraphStateImpactsJSON(receipt.Impacts),
		ThreadImpactCount:   len(receipt.ThreadImpacts),
		ThreadImpacts:       toThreadProjectionImpactsJSON(receipt.ThreadImpacts),
		Remedy:              receipt.Remedy,
	}
	for _, blocker := range receipt.OutstandingBlockers {
		payload.OutstandingBlockers = append(payload.OutstandingBlockers, TaskLifecycleBlockerJSON{
			TaskID: blocker.TaskID, Reason: string(blocker.Reason),
			Path: append([]string(nil), blocker.Path...), Direct: blocker.Direct,
		})
	}
	return payload
}

// TaskEligibilityFailureJSON preserves the typed default-start refusal in each
// failed batch row. Error text remains alongside it for human compatibility.
type TaskEligibilityFailureJSON struct {
	TaskID            string                     `json:"task_id"`
	State             TaskGraphStateJSON         `json:"state"`
	Blockers          []TaskLifecycleBlockerJSON `json:"blockers"`
	RequestedOverride string                     `json:"requested_override,omitempty"`
	OverrideAllowed   bool                       `json:"override_allowed" jsonschema:"description=true only when the refusal is a blocked dependency gate on next-up or ready-to-start work; force never bypasses a broken gate or another lifecycle role"`
	Remedy            string                     `json:"remedy"`
}

func ToTaskEligibilityFailureJSON(failure *core.TaskEligibilityError) TaskEligibilityFailureJSON {
	payload := TaskEligibilityFailureJSON{
		TaskID: failure.TaskID, State: toTaskGraphStateJSON(failure.State),
		Blockers:          make([]TaskLifecycleBlockerJSON, 0, len(failure.Blockers)),
		RequestedOverride: string(failure.RequestedOverride),
		OverrideAllowed:   failure.DependencyGateOverrideAllowed(),
		Remedy:            failure.Remedy(),
	}
	for _, blocker := range failure.Blockers {
		payload.Blockers = append(payload.Blockers, TaskLifecycleBlockerJSON{
			TaskID: blocker.TaskID, Reason: string(blocker.Reason),
			Path: append([]string(nil), blocker.Path...), Direct: blocker.Direct,
		})
	}
	return payload
}

// DependencyMutationJSON is the reusable dependency-mutation receipt payload.
type DependencyMutationJSON struct {
	Operation           string                      `json:"operation"`
	Changed             bool                        `json:"changed"`
	DryRun              bool                        `json:"dry_run"`
	Edges               []DependencyEdgeOutcomeJSON `json:"edges"`
	Impacts             []TaskGraphStateImpactJSON  `json:"impacts"`
	ClearedLegacyFields []LegacyFieldClearJSON      `json:"cleared_legacy_fields"`
	PlannedTaskIDs      []string                    `json:"planned_task_ids"`
	AppliedTaskIDs      []string                    `json:"applied_task_ids"`
	RemainingTaskIDs    []string                    `json:"remaining_task_ids"`
	Workspace           WorkspaceJSON               `json:"workspace"`
}

// ToDependencyMutationJSON maps a core receipt and adapter-owned workspace identity.
func ToDependencyMutationJSON(receipt core.DependencyMutationReceipt, workspace WorkspaceJSON) DependencyMutationJSON {
	payload := DependencyMutationJSON{
		Operation: string(receipt.Operation), Changed: receipt.Changed, DryRun: receipt.DryRun,
		Edges:               make([]DependencyEdgeOutcomeJSON, 0, len(receipt.Edges)),
		Impacts:             toTaskGraphStateImpactsJSON(receipt.Impacts),
		ClearedLegacyFields: make([]LegacyFieldClearJSON, 0, len(receipt.ClearedLegacyFields)),
		PlannedTaskIDs:      append([]string{}, receipt.PlannedTaskIDs...),
		AppliedTaskIDs:      append([]string{}, receipt.AppliedTaskIDs...),
		RemainingTaskIDs:    append([]string{}, receipt.RemainingTaskIDs...), Workspace: workspace,
	}
	for _, edge := range receipt.Edges {
		payload.Edges = append(payload.Edges, DependencyEdgeOutcomeJSON{
			DependentID: edge.DependentID, PrerequisiteID: edge.PrerequisiteID,
			Action: string(edge.Action), Outcome: edge.Outcome,
		})
	}
	for _, clear := range receipt.ClearedLegacyFields {
		payload.ClearedLegacyFields = append(payload.ClearedLegacyFields, LegacyFieldClearJSON{TaskID: clear.TaskID, Field: clear.Field})
	}
	return payload
}

// DependencyMutationEnvelope is the successful `task depend` machine receipt.
type DependencyMutationEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	DependencyMutationJSON
}

// ToDependencyMutationEnvelope wraps a dependency mutation receipt.
func ToDependencyMutationEnvelope(receipt core.DependencyMutationReceipt, workspace WorkspaceJSON) DependencyMutationEnvelope {
	return DependencyMutationEnvelope{
		SchemaVersion: SchemaVersion, DependencyMutationJSON: ToDependencyMutationJSON(receipt, workspace),
	}
}
