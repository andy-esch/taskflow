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
	Status      string   `json:"status" jsonschema:"description=unstarted|in-progress|completed|abandoned"`
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
	Code     string `json:"code" jsonschema:"description=invalid-thread-document|missing-thread-member|completed-thread-empty|completed-thread-undrained|completed-thread-external-gate|completed-thread-unhealthy-evidence"`
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
	ExternalGates    []ThreadExternalGateJSON `json:"external_gates"`
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
	SchemaVersion string               `json:"schema_version"`
	GraphHealth   string               `json:"graph_health" jsonschema:"description=repository task-DAG verdict: healthy|degraded|broken"`
	GraphProblems []GraphProblemJSON   `json:"graph_problems"`
	Threads       []ThreadViewJSON     `json:"threads"`
	Unreadable    []domain.FileProblem `json:"unreadable"`
}

func ToThreadsEnvelope(list core.ThreadListView, problems []domain.FileProblem) ThreadsEnvelope {
	payload := ThreadsEnvelope{
		SchemaVersion: SchemaVersion, GraphHealth: string(list.GraphHealth), GraphProblems: toGraphProblemsJSON(list.GraphProblems),
		Threads: make([]ThreadViewJSON, 0, len(list.Threads)), Unreadable: problems,
	}
	if payload.Unreadable == nil {
		payload.Unreadable = []domain.FileProblem{}
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
