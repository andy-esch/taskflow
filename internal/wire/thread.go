package wire

import (
	"sort"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// ThreadJSON is the persisted metadata and membership of one Thread. It never
// contains dependency edges or cached graph projections.
type ThreadJSON struct {
	ID          string   `json:"id"`
	Slug        string   `json:"slug"`
	Status      string   `json:"status" jsonschema:"description=unstarted|in-progress|completed|cancelled"`
	Description string   `json:"description"`
	Goal        string   `json:"goal"`
	TargetDate  string   `json:"target_date,omitempty"`
	Created     string   `json:"created"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
	StartedAt   string   `json:"started_at,omitempty"`
	EndedAt     string   `json:"ended_at,omitempty"`
	Tags        []string `json:"tags"`
	Tasks       []string `json:"tasks"`
}

func ToThreadJSON(thread domain.Thread) ThreadJSON {
	return ThreadJSON{
		ID: thread.ID, Slug: thread.Slug, Status: string(thread.Status), Description: thread.Description,
		Goal: thread.Goal, TargetDate: thread.TargetDate, Created: thread.Created, UpdatedAt: thread.Updated,
		StartedAt: thread.StartedAt, EndedAt: thread.EndedAt,
		Tags: append([]string{}, thread.Tags...), Tasks: nonNilSortedStrings(thread.Tasks),
	}
}

func nonNilSortedStrings(values []string) []string {
	out := append([]string{}, values...)
	sort.Strings(out)
	return out
}

// ThreadTaskJSON is one task in a Thread projection, with an explicit member or
// external-gate role and the shared repository graph state.
type ThreadTaskJSON struct {
	Role  string             `json:"role" jsonschema:"description=member|external-gate"`
	Task  GraphTaskJSON      `json:"task"`
	State TaskGraphStateJSON `json:"state"`
}

// ThreadExternalGateJSON marks whether a direct outside prerequisite is still
// preventing sound closure.
type ThreadExternalGateJSON struct {
	ThreadTaskJSON
	Outstanding bool `json:"outstanding"`
}

type ThreadRollupJSON struct {
	Done       int `json:"done"`
	Total      int `json:"total"`
	Drained    int `json:"drained"`
	Deprecated int `json:"deprecated"`
}

type ThreadProblemJSON struct {
	Code     string `json:"code" jsonschema:"description=invalid-thread-document|thread-id-drift|duplicate-thread-id|thread-task-id-collision|missing-thread-member|completed-thread-empty|completed-thread-undrained|completed-thread-external-gate|completed-thread-unhealthy-evidence"`
	ThreadID string `json:"thread_id"`
	TaskID   string `json:"task_id,omitempty"`
	Path     string `json:"path,omitempty"`
	Message  string `json:"message"`
}

// ThreadViewJSON is the shared runtime projection consumed by list, show, and
// frontier adapters.
type ThreadViewJSON struct {
	Thread           ThreadJSON               `json:"thread"`
	Rollup           ThreadRollupJSON         `json:"rollup"`
	GraphHealth      string                   `json:"graph_health" jsonschema:"description=repository task-DAG verdict: healthy|degraded|broken"`
	ProjectionHealth string                   `json:"projection_health" jsonschema:"description=Thread-local evidence verdict after combining graph and document integrity: healthy|degraded|broken"`
	Inconsistent     bool                     `json:"inconsistent" jsonschema:"description=true when a completed Thread is not soundly closed"`
	Members          []ThreadTaskJSON         `json:"members"`
	ExternalGates    []ThreadExternalGateJSON `json:"external_gates" jsonschema:"description=direct prerequisites outside the Thread membership boundary; deeper upstream context is available through causal blocker queries"`
	Frontier         []ThreadTaskJSON         `json:"frontier" jsonschema:"description=next-up and ready-to-start Thread members with clear dependency gates; empty unless both graph_health and projection_health are healthy"`
	Problems         []ThreadProblemJSON      `json:"problems"`
	GraphProblems    []GraphProblemJSON       `json:"graph_problems"`
}

func ToThreadViewJSON(view core.ThreadView) ThreadViewJSON {
	payload := ThreadViewJSON{
		Thread: ToThreadJSON(view.Thread),
		Rollup: ThreadRollupJSON{
			Done: view.Rollup.Done, Total: view.Rollup.Total, Drained: view.Rollup.Drained, Deprecated: view.Rollup.Deprecated,
		},
		GraphHealth: string(view.GraphHealth), ProjectionHealth: string(view.ProjectionHealth), Inconsistent: view.Inconsistent,
		Members:       make([]ThreadTaskJSON, 0, len(view.Members)),
		ExternalGates: make([]ThreadExternalGateJSON, 0, len(view.ExternalGates)),
		Frontier:      make([]ThreadTaskJSON, 0, len(view.Frontier)),
		Problems:      make([]ThreadProblemJSON, 0, len(view.Problems)),
		GraphProblems: toGraphProblemsJSON(view.GraphProblems),
	}
	for _, member := range view.Members {
		payload.Members = append(payload.Members, toThreadTaskJSON(member))
	}
	for _, gate := range view.ExternalGates {
		payload.ExternalGates = append(payload.ExternalGates, ThreadExternalGateJSON{
			ThreadTaskJSON: toThreadTaskJSON(gate.ThreadTaskView), Outstanding: gate.Outstanding,
		})
	}
	for _, member := range view.Frontier {
		payload.Frontier = append(payload.Frontier, toThreadTaskJSON(member))
	}
	for _, problem := range view.Problems {
		payload.Problems = append(payload.Problems, ThreadProblemJSON{
			Code: string(problem.Code), ThreadID: problem.ThreadID, TaskID: problem.TaskID,
			Path: problem.Path, Message: problem.Message,
		})
	}
	return payload
}

func toThreadTaskJSON(item core.ThreadTaskView) ThreadTaskJSON {
	return ThreadTaskJSON{
		Role: string(item.Role), Task: toGraphTaskJSON(item.State.TaskID, item.Task), State: toTaskGraphStateJSON(item.State),
	}
}

// ThreadsEnvelope is `thread list --json`.
type ThreadsEnvelope struct {
	SchemaVersion string                  `json:"schema_version"`
	GraphHealth   string                  `json:"graph_health" jsonschema:"description=repository task-DAG verdict: healthy|degraded|broken"`
	GraphProblems []GraphProblemJSON      `json:"graph_problems"`
	Threads       []ThreadViewJSON        `json:"threads"`
	Unreadable    []ThreadReadProblemJSON `json:"unreadable"`
}

// ThreadReadProblemJSON is an adapter-neutral failed-record diagnostic. Identity
// is explicit when recoverable; location is optional repair context and must not
// be parsed by consumers to reconstruct identity.
type ThreadReadProblemJSON struct {
	ThreadID   string `json:"thread_id,omitempty"`
	ThreadSlug string `json:"thread_slug,omitempty"`
	Location   string `json:"location,omitempty"`
	Message    string `json:"message"`
}

func ToThreadsEnvelope(list core.ThreadListView, problems []core.ThreadReadProblem) ThreadsEnvelope {
	payload := ThreadsEnvelope{
		SchemaVersion: SchemaVersion, GraphHealth: string(list.GraphHealth), GraphProblems: toGraphProblemsJSON(list.GraphProblems),
		Threads: make([]ThreadViewJSON, 0, len(list.Threads)), Unreadable: make([]ThreadReadProblemJSON, 0, len(problems)),
	}
	for _, problem := range problems {
		payload.Unreadable = append(payload.Unreadable, ThreadReadProblemJSON{
			ThreadID: problem.ThreadID, ThreadSlug: problem.ThreadSlug,
			Location: problem.Location, Message: problem.Message,
		})
	}
	for _, view := range list.Threads {
		payload.Threads = append(payload.Threads, ToThreadViewJSON(view))
	}
	return payload
}

// ThreadShowEnvelope is `thread show --json`.
type ThreadShowEnvelope struct {
	SchemaVersion string         `json:"schema_version"`
	View          ThreadViewJSON `json:"view"`
	Body          string         `json:"body"`
}

func ToThreadShowEnvelope(view core.ThreadView, body string) ThreadShowEnvelope {
	return ThreadShowEnvelope{SchemaVersion: SchemaVersion, View: ToThreadViewJSON(view), Body: body}
}

// ThreadFrontierEnvelope is `thread frontier --json`. It retains the complete
// diagnosis so an empty frontier never masquerades as permission or completion.
type ThreadFrontierEnvelope struct {
	SchemaVersion string         `json:"schema_version"`
	View          ThreadViewJSON `json:"view"`
}

func ToThreadFrontierEnvelope(view core.ThreadView) ThreadFrontierEnvelope {
	return ThreadFrontierEnvelope{SchemaVersion: SchemaVersion, View: ToThreadViewJSON(view)}
}

// ThreadGraphNodeJSON is one renderer-neutral vertex. Labels and descriptions
// remain raw transport data; Mermaid and DOT escaping belongs to graphfmt.
type ThreadGraphNodeJSON struct {
	TaskID      string             `json:"task_id"`
	Label       string             `json:"label"`
	Description string             `json:"description"`
	Status      string             `json:"status"`
	Role        string             `json:"role" jsonschema:"description=member|external-gate"`
	State       TaskGraphStateJSON `json:"state"`
}

// ThreadGraphEdgeJSON follows prerequisite-to-dependent direction.
type ThreadGraphEdgeJSON struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ThreadGraphWaveJSON is a one-based explanatory generation of member task IDs.
type ThreadGraphWaveJSON struct {
	Index   int      `json:"index"`
	TaskIDs []string `json:"task_ids"`
}

// ThreadGraphProjectionJSON is the shared machine graph/plan projection. It is
// intentionally renderer- and framework-neutral.
type ThreadGraphProjectionJSON struct {
	View             ThreadViewJSON        `json:"view"`
	Nodes            []ThreadGraphNodeJSON `json:"nodes"`
	Edges            []ThreadGraphEdgeJSON `json:"edges"`
	Waves            []ThreadGraphWaveJSON `json:"waves"`
	TopologyComplete bool                  `json:"topology_complete" jsonschema:"description=true only when the member topology and qualifying Thread projection are healthy and complete"`
}

func ToThreadGraphProjectionJSON(projection core.ThreadGraphProjection) ThreadGraphProjectionJSON {
	payload := ThreadGraphProjectionJSON{
		View:             ToThreadViewJSON(projection.View),
		Nodes:            make([]ThreadGraphNodeJSON, 0, len(projection.Nodes)),
		Edges:            make([]ThreadGraphEdgeJSON, 0, len(projection.Edges)),
		Waves:            make([]ThreadGraphWaveJSON, 0, len(projection.Waves)),
		TopologyComplete: projection.TopologyComplete,
	}
	for _, node := range projection.Nodes {
		payload.Nodes = append(payload.Nodes, ThreadGraphNodeJSON{
			TaskID: node.TaskID, Label: node.Label, Description: node.Description,
			Status: string(node.Status), Role: string(node.Role), State: toTaskGraphStateJSON(node.State),
		})
	}
	for _, edge := range projection.Edges {
		payload.Edges = append(payload.Edges, ThreadGraphEdgeJSON{From: edge.From, To: edge.To})
	}
	for _, wave := range projection.Waves {
		payload.Waves = append(payload.Waves, ThreadGraphWaveJSON{Index: wave.Index, TaskIDs: append([]string{}, wave.TaskIDs...)})
	}
	return payload
}

// ThreadGraphEnvelope is `thread graph --json`. Renderer choice is deliberately
// absent: machine callers receive the reusable semantic projection.
type ThreadGraphEnvelope struct {
	SchemaVersion string                    `json:"schema_version"`
	Projection    ThreadGraphProjectionJSON `json:"projection"`
}

func ToThreadGraphEnvelope(projection core.ThreadGraphProjection) ThreadGraphEnvelope {
	return ThreadGraphEnvelope{SchemaVersion: SchemaVersion, Projection: ToThreadGraphProjectionJSON(projection)}
}

// ThreadPlanEnvelope is `thread plan --json`. It shares the exact neutral
// projection with graph export while retaining a named schema entry per command.
type ThreadPlanEnvelope struct {
	SchemaVersion string                    `json:"schema_version"`
	Projection    ThreadGraphProjectionJSON `json:"projection"`
}

func ToThreadPlanEnvelope(projection core.ThreadGraphProjection) ThreadPlanEnvelope {
	return ThreadPlanEnvelope{SchemaVersion: SchemaVersion, Projection: ToThreadGraphProjectionJSON(projection)}
}

// ThreadMutationJSON is the reusable committed-outcome receipt payload. It is
// standalone inside the success envelope and nested in a post-commit error.
type ThreadMutationJSON struct {
	Thread    ThreadJSON    `json:"thread"`
	Changed   bool          `json:"changed"`
	DryRun    bool          `json:"dry_run"`
	Committed bool          `json:"committed" jsonschema:"description=true only after the Thread file became durable"`
	Path      string        `json:"path"`
	Workspace WorkspaceJSON `json:"workspace"`
}

func ToThreadMutationJSON(receipt core.ThreadCreationReceipt, path string, workspace WorkspaceJSON) ThreadMutationJSON {
	return ThreadMutationJSON{
		Thread: ToThreadJSON(receipt.Thread), Changed: receipt.Changed,
		DryRun: receipt.DryRun, Committed: receipt.Committed, Path: path, Workspace: workspace,
	}
}

// ThreadMutationEnvelope is the successful `thread new --json` receipt.
type ThreadMutationEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	ThreadMutationJSON
}

func ToThreadMutationEnvelope(receipt core.ThreadCreationReceipt, path string, workspace WorkspaceJSON) ThreadMutationEnvelope {
	return ThreadMutationEnvelope{
		SchemaVersion: SchemaVersion, ThreadMutationJSON: ToThreadMutationJSON(receipt, path, workspace),
	}
}

// ThreadMemberOutcomeJSON records one resolved member intent, including an
// idempotent skip that deliberately caused no set change.
type ThreadMemberOutcomeJSON struct {
	TaskID  string `json:"task_id"`
	Action  string `json:"action" jsonschema:"description=add|remove"`
	Outcome string `json:"outcome" jsonschema:"description=added|removed|skipped"`
}

// ThreadUpdateJSON is the existing-Thread membership/lifecycle mutation receipt.
// Before and after retain the complete projection so graph consequences remain
// inspectable without a follow-up read.
type ThreadUpdateJSON struct {
	Operation      string                    `json:"operation" jsonschema:"description=add-members|remove-members|start|complete|cancel|reopen"`
	ThreadID       string                    `json:"thread_id"`
	Changed        bool                      `json:"changed"`
	DryRun         bool                      `json:"dry_run"`
	Committed      bool                      `json:"committed" jsonschema:"description=true only after the Thread file became durable"`
	MemberOutcomes []ThreadMemberOutcomeJSON `json:"member_outcomes"`
	Before         ThreadViewJSON            `json:"before"`
	After          ThreadViewJSON            `json:"after"`
	Remedy         string                    `json:"remedy,omitempty"`
	Path           string                    `json:"path"`
	Workspace      WorkspaceJSON             `json:"workspace"`
}

func ToThreadUpdateJSON(receipt core.ThreadMutationReceipt, path string, workspace WorkspaceJSON) ThreadUpdateJSON {
	payload := ThreadUpdateJSON{
		Operation: string(receipt.Operation), ThreadID: receipt.Thread.ID,
		Changed: receipt.Changed, DryRun: receipt.DryRun, Committed: receipt.Committed,
		MemberOutcomes: make([]ThreadMemberOutcomeJSON, 0, len(receipt.MemberOutcomes)),
		Before:         ToThreadViewJSON(receipt.Before), After: ToThreadViewJSON(receipt.After),
		Remedy: receipt.Remedy, Path: path, Workspace: workspace,
	}
	for _, outcome := range receipt.MemberOutcomes {
		payload.MemberOutcomes = append(payload.MemberOutcomes, ThreadMemberOutcomeJSON{
			TaskID: outcome.TaskID, Action: outcome.Action, Outcome: outcome.Outcome,
		})
	}
	return payload
}

// ThreadUpdateEnvelope is a successful existing-Thread mutation receipt.
type ThreadUpdateEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	ThreadUpdateJSON
}

func ToThreadUpdateEnvelope(receipt core.ThreadMutationReceipt, path string, workspace WorkspaceJSON) ThreadUpdateEnvelope {
	return ThreadUpdateEnvelope{SchemaVersion: SchemaVersion, ThreadUpdateJSON: ToThreadUpdateJSON(receipt, path, workspace)}
}

type ThreadApplyOperationJSON struct {
	Kind           string `json:"kind" jsonschema:"description=dependency|thread"`
	Action         string `json:"action" jsonschema:"description=add|create"`
	State          string `json:"state" jsonschema:"description=pending|applied|skipped"`
	ThreadID       string `json:"thread_id,omitempty"`
	DependentID    string `json:"dependent_id,omitempty"`
	PrerequisiteID string `json:"prerequisite_id,omitempty"`
}

func toThreadApplyOperations(operations []core.ThreadApplyOperation) []ThreadApplyOperationJSON {
	out := make([]ThreadApplyOperationJSON, 0, len(operations))
	for _, operation := range operations {
		out = append(out, ThreadApplyOperationJSON{
			Kind: operation.Kind, Action: operation.Action, State: string(operation.State),
			ThreadID: operation.ThreadID, DependentID: operation.DependentID,
			PrerequisiteID: operation.PrerequisiteID,
		})
	}
	return out
}

// ThreadApplyDependencyJSON is one stable-ID prerequisite-to-dependent edge in
// a generated apply plan.
type ThreadApplyDependencyJSON struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ThreadApplyThreadJSON is the exact new-Thread intent embedded in a generated
// apply plan. Body is retained so an identical retry can be distinguished from
// a same-ID document collision.
type ThreadApplyThreadJSON struct {
	ID          string   `json:"id"`
	Slug        string   `json:"slug"`
	Status      string   `json:"status" jsonschema:"description=unstarted for V1 generated plans"`
	Description string   `json:"description"`
	Goal        string   `json:"goal"`
	TargetDate  string   `json:"target_date,omitempty"`
	Created     string   `json:"created"`
	Tags        []string `json:"tags,omitempty"`
	Tasks       []string `json:"tasks"`
	Body        string   `json:"body"`
}

// ThreadApplyPlanJSON is the machine projection of the durable schema-1 retry
// token. Its repository ID authorizes apply against one planning space; all
// dependency and membership references are exact stable IDs.
type ThreadApplyPlanJSON struct {
	Schema         int                         `json:"schema"`
	PlanningRepoID string                      `json:"planning_repo_id"`
	ComposedAt     string                      `json:"composed_at"`
	Thread         ThreadApplyThreadJSON       `json:"thread"`
	Dependencies   []ThreadApplyDependencyJSON `json:"dependencies,omitempty"`
}

func ToThreadApplyPlanJSON(plan core.ThreadApplyPlan) ThreadApplyPlanJSON {
	payload := ThreadApplyPlanJSON{
		Schema: plan.Schema, PlanningRepoID: plan.PlanningRepoID, ComposedAt: plan.ComposedAt,
		Thread: ThreadApplyThreadJSON{
			ID: plan.Thread.ID, Slug: plan.Thread.Slug, Status: string(plan.Thread.Status),
			Description: plan.Thread.Description, Goal: plan.Thread.Goal,
			TargetDate: plan.Thread.TargetDate, Created: plan.Thread.Created,
			Tags:  append([]string(nil), plan.Thread.Tags...),
			Tasks: append([]string{}, plan.Thread.Tasks...), Body: plan.Thread.Body,
		},
		Dependencies: make([]ThreadApplyDependencyJSON, 0, len(plan.Dependencies)),
	}
	for _, dependency := range plan.Dependencies {
		payload.Dependencies = append(payload.Dependencies, ThreadApplyDependencyJSON{
			From: dependency.From, To: dependency.To,
		})
	}
	return payload
}

// ThreadApplyComposeEnvelope is `thread compose --json`. Plan is the exact
// durable document written to PlanPath (or previewed when DryRun is true).
type ThreadApplyComposeEnvelope struct {
	SchemaVersion string              `json:"schema_version"`
	DryRun        bool                `json:"dry_run"`
	Written       bool                `json:"written"`
	PlanPath      string              `json:"plan_path"`
	Plan          ThreadApplyPlanJSON `json:"plan"`
	Workspace     WorkspaceJSON       `json:"workspace"`
}

func ToThreadApplyComposeEnvelope(plan core.ThreadApplyPlan, planPath string, dryRun bool, workspace WorkspaceJSON) ThreadApplyComposeEnvelope {
	return ThreadApplyComposeEnvelope{
		SchemaVersion: SchemaVersion, DryRun: dryRun, Written: !dryRun,
		PlanPath: planPath, Plan: ToThreadApplyPlanJSON(plan), Workspace: workspace,
	}
}

// ThreadApplyJSON is both the successful apply payload and the recovery detail
// nested in an error envelope after a durable prefix.
type ThreadApplyJSON struct {
	ThreadID   string                     `json:"thread_id"`
	ThreadSlug string                     `json:"thread_slug"`
	Changed    bool                       `json:"changed"`
	DryRun     bool                       `json:"dry_run"`
	Complete   bool                       `json:"complete"`
	Committed  bool                       `json:"committed"`
	Operations []ThreadApplyOperationJSON `json:"operations"`
	PlanPath   string                     `json:"plan_path"`
	Workspace  WorkspaceJSON              `json:"workspace"`
}

func ToThreadApplyJSON(receipt core.ThreadApplyReceipt, planPath string, workspace WorkspaceJSON) ThreadApplyJSON {
	return ThreadApplyJSON{
		ThreadID: receipt.Plan.Thread.ID, ThreadSlug: receipt.Plan.Thread.Slug,
		Changed: receipt.Changed, DryRun: receipt.DryRun, Complete: receipt.Complete,
		Committed: receipt.Committed, Operations: toThreadApplyOperations(receipt.Operations),
		PlanPath: planPath, Workspace: workspace,
	}
}

type ThreadApplyEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	ThreadApplyJSON
}

func ToThreadApplyEnvelope(receipt core.ThreadApplyReceipt, planPath string, workspace WorkspaceJSON) ThreadApplyEnvelope {
	return ThreadApplyEnvelope{
		SchemaVersion:   SchemaVersion,
		ThreadApplyJSON: ToThreadApplyJSON(receipt, planPath, workspace),
	}
}

// ThreadMutationFailureJSON is a pre-commit membership/lifecycle policy refusal.
type ThreadMutationFailureJSON struct {
	ThreadID  string `json:"thread_id"`
	Operation string `json:"operation"`
	Status    string `json:"status"`
	Reason    string `json:"reason"`
	Remedy    string `json:"remedy"`
}

func ToThreadMutationFailureJSON(failure *core.ThreadMutationPolicyError) ThreadMutationFailureJSON {
	return ThreadMutationFailureJSON{
		ThreadID: failure.ThreadID, Operation: string(failure.Operation), Status: string(failure.Status),
		Reason: failure.Reason, Remedy: failure.Remedy,
	}
}
