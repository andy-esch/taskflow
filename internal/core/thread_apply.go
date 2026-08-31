package core

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

const ThreadApplyPlanSchema = 1

type ThreadComposeInput struct {
	Title       string   `json:"title" yaml:"title"`
	Description string   `json:"description" yaml:"description"`
	Goal        string   `json:"goal" yaml:"goal"`
	TargetDate  string   `json:"target_date,omitempty" yaml:"target_date,omitempty"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type ThreadComposeNode struct {
	Key    string `json:"key" yaml:"key"`
	TaskID string `json:"task_id" yaml:"task_id"`
	Member *bool  `json:"member,omitempty" yaml:"member,omitempty"`
}

type ThreadComposeDependency struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}

// ThreadComposeManifest is the human-authored, local-key input. Schema zero is
// accepted as the documented V1 shorthand; materialized plans always carry an
// explicit schema.
type ThreadComposeManifest struct {
	Schema       int                       `json:"schema,omitempty" yaml:"schema,omitempty"`
	Thread       ThreadComposeInput        `json:"thread" yaml:"thread"`
	Nodes        []ThreadComposeNode       `json:"nodes" yaml:"nodes"`
	Dependencies []ThreadComposeDependency `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
}

// ThreadApplyThread contains every byte-significant semantic input needed to
// distinguish an idempotent retry from a same-ID collision.
type ThreadApplyThread struct {
	ID          string              `json:"id" yaml:"id"`
	Slug        string              `json:"slug" yaml:"slug"`
	Status      domain.ThreadStatus `json:"status" yaml:"status"`
	Description string              `json:"description" yaml:"description"`
	Goal        string              `json:"goal" yaml:"goal"`
	TargetDate  string              `json:"target_date,omitempty" yaml:"target_date,omitempty"`
	Created     string              `json:"created" yaml:"created"`
	Tags        []string            `json:"tags,omitempty" yaml:"tags,omitempty"`
	Tasks       []string            `json:"tasks" yaml:"tasks"`
	Body        string              `json:"body" yaml:"body"`
}

func (planned ThreadApplyThread) domainThread() domain.Thread {
	return domain.Thread{
		ID: planned.ID, Slug: planned.Slug, Status: planned.Status,
		Description: planned.Description, Goal: planned.Goal, TargetDate: planned.TargetDate,
		Created: planned.Created, Tags: append([]string(nil), planned.Tags...),
		Tasks: append([]string(nil), planned.Tasks...),
	}
}

type ThreadApplyDependency struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}

// ThreadApplyPlan is the durable retry token. It is additive intent, not a
// frozen snapshot of unrelated repository state.
type ThreadApplyPlan struct {
	Schema         int                     `json:"schema" yaml:"schema"`
	PlanningRepoID string                  `json:"planning_repo_id" yaml:"planning_repo_id"`
	ComposedAt     string                  `json:"composed_at" yaml:"composed_at"`
	Thread         ThreadApplyThread       `json:"thread" yaml:"thread"`
	Dependencies   []ThreadApplyDependency `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
}

type ThreadApplySnapshot struct {
	PlanningRepoID string
	Graph          *TaskGraph
	Threads        []domain.Thread
	ThreadBodies   map[string]string
}

type ThreadApplyOperationState string

const (
	ThreadApplyPending ThreadApplyOperationState = "pending"
	ThreadApplyApplied ThreadApplyOperationState = "applied"
	ThreadApplySkipped ThreadApplyOperationState = "skipped"
)

type ThreadApplyOperation struct {
	Kind           string                    `json:"kind"`
	Action         string                    `json:"action"`
	State          ThreadApplyOperationState `json:"state"`
	ThreadID       string                    `json:"thread_id,omitempty"`
	DependentID    string                    `json:"dependent_id,omitempty"`
	PrerequisiteID string                    `json:"prerequisite_id,omitempty"`
}

type ThreadApplyDecision struct {
	Plan       ThreadApplyPlan
	GraphPlan  TaskGraphMutationPlan
	ThreadPlan *ThreadCreationPlan
	Operations []ThreadApplyOperation
}

// ThreadApplyPlanner is control-inverted so the durable plan is selected while
// the store owns the canonical repository guard. The store revalidates the
// returned plan against the same snapshot before materializing any writes.
type ThreadApplyPlanner func(ThreadApplySnapshot) (ThreadApplyPlan, error)

// ThreadApplyMutationResult is the store-owned recovery record. Operations are
// always in deterministic durable order: dependency additions first and the
// Thread create last. Pending entries are the exact remaining intent after an
// interrupted prefix.
type ThreadApplyMutationResult struct {
	Plan       ThreadApplyPlan
	Operations []ThreadApplyOperation
	Changed    bool
	DryRun     bool
	Complete   bool
	Committed  bool
}

type ThreadApplyReceipt struct {
	Plan       ThreadApplyPlan        `json:"plan"`
	Operations []ThreadApplyOperation `json:"operations"`
	Changed    bool                   `json:"changed"`
	DryRun     bool                   `json:"dry_run"`
	Complete   bool                   `json:"complete"`
	Committed  bool                   `json:"committed"`
}

// ThreadApplyFailure retains the durable prefix when apply cannot finish. A
// caller may retry the same materialized plan when pending operations remain;
// Complete means all semantic writes landed even if guard cleanup later failed.
type ThreadApplyFailure struct {
	Cause   error
	Receipt ThreadApplyReceipt
}

func (e *ThreadApplyFailure) Error() string {
	if e == nil || e.Cause == nil {
		return "Thread apply failed"
	}
	if e.Receipt.Complete {
		return fmt.Sprintf("%v; the Thread apply is complete and must not be retried blindly", e.Cause)
	}
	if !e.Receipt.Committed {
		return e.Cause.Error()
	}
	if pending := pendingThreadApplyOperations(e.Receipt.Operations); pending > 0 {
		return fmt.Sprintf("%v; a durable Thread-apply prefix landed and %d operation(s) remain pending; retry the same materialized plan to converge",
			e.Cause, pending)
	}
	return fmt.Sprintf("%v; a durable Thread-apply prefix landed but final convergence could not be verified; inspect the repository and retry the same materialized plan after resolving the reported error", e.Cause)
}

func (e *ThreadApplyFailure) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func pendingThreadApplyOperations(operations []ThreadApplyOperation) int {
	count := 0
	for _, operation := range operations {
		if operation.State == ThreadApplyPending {
			count++
		}
	}
	return count
}

// ComposeThreadApplyPlan resolves local keys and mints the one Thread ID without
// mutating repository state. body is rendered by the Service from the ordinary
// Thread template before the pure compiler runs.
func ComposeThreadApplyPlan(snapshot ThreadApplySnapshot, manifest ThreadComposeManifest, body string, newID func() string, now time.Time) (ThreadApplyPlan, error) {
	if err := ValidateThreadCreationSource(snapshot.Graph, snapshot.Threads, nil); err != nil {
		return ThreadApplyPlan{}, err
	}
	if snapshot.PlanningRepoID == "" {
		return ThreadApplyPlan{}, fmt.Errorf("%w: planning repository has no durable id; run `tskflwctl config migrate` before composing an apply plan", domain.ErrValidation)
	}
	if manifest.Schema != 0 && manifest.Schema != ThreadApplyPlanSchema {
		return ThreadApplyPlan{}, fmt.Errorf("%w: unsupported Thread authoring manifest schema %d", domain.ErrValidation, manifest.Schema)
	}
	if now.IsZero() {
		return ThreadApplyPlan{}, fmt.Errorf("%w: compose time is required", domain.ErrValidation)
	}
	if newID == nil {
		return ThreadApplyPlan{}, fmt.Errorf("%w: Thread id generator is required", domain.ErrValidation)
	}
	title := strings.TrimSpace(manifest.Thread.Title)
	if title == "" {
		return ThreadApplyPlan{}, fmt.Errorf("%w: Thread title is required", domain.ErrValidation)
	}
	slug := domain.Slugify(title)
	if slug == "" {
		return ThreadApplyPlan{}, fmt.Errorf("%w: Thread title produced an empty slug", domain.ErrValidation)
	}
	if len(manifest.Nodes) == 0 {
		return ThreadApplyPlan{}, fmt.Errorf("%w: Thread manifest requires at least one node", domain.ErrValidation)
	}

	usedIDs := make(map[string]bool)
	for _, taskID := range snapshot.Graph.TaskIDs() {
		usedIDs[taskID] = true
	}
	for _, thread := range snapshot.Threads {
		usedIDs[thread.ID] = true
	}
	threadID := ""
	for range 16 {
		candidate := newID()
		if !id.Valid(candidate) {
			return ThreadApplyPlan{}, fmt.Errorf("%w: generated Thread id %q is invalid", domain.ErrValidation, candidate)
		}
		if !usedIDs[candidate] {
			threadID = candidate
			break
		}
	}
	if threadID == "" {
		return ThreadApplyPlan{}, fmt.Errorf("%w: could not mint a unique Thread id", domain.ErrConflict)
	}

	keys := make(map[string]string, len(manifest.Nodes))
	declaredTasks := make(map[string]string, len(manifest.Nodes))
	members := make([]string, 0, len(manifest.Nodes))
	external := make([]string, 0)
	for _, node := range manifest.Nodes {
		key := strings.TrimSpace(node.Key)
		switch {
		case key == "":
			return ThreadApplyPlan{}, fmt.Errorf("%w: every manifest node requires a local key", domain.ErrValidation)
		case keys[key] != "":
			return ThreadApplyPlan{}, fmt.Errorf("%w: duplicate manifest node key %q", domain.ErrValidation, key)
		case !id.Valid(node.TaskID):
			return ThreadApplyPlan{}, fmt.Errorf("%w: node %q task_id %q is not an exact stable task id", domain.ErrValidation, key, node.TaskID)
		}
		if _, exists := snapshot.Graph.Task(node.TaskID); !exists {
			return ThreadApplyPlan{}, fmt.Errorf("%w: node %q references missing task %s", domain.ErrValidation, key, node.TaskID)
		}
		if prior := declaredTasks[node.TaskID]; prior != "" {
			return ThreadApplyPlan{}, fmt.Errorf("%w: nodes %q and %q both declare task %s", domain.ErrValidation, prior, key, node.TaskID)
		}
		keys[key] = node.TaskID
		declaredTasks[node.TaskID] = key
		member := node.Member == nil || *node.Member
		if member {
			members = append(members, node.TaskID)
		} else {
			external = append(external, node.TaskID)
		}
	}
	if len(members) == 0 {
		return ThreadApplyPlan{}, fmt.Errorf("%w: Thread manifest requires at least one member node", domain.ErrValidation)
	}
	sort.Strings(members)

	edges := make([]ThreadApplyDependency, 0, len(manifest.Dependencies))
	seenEdges := make(map[DependencyEdge]bool, len(manifest.Dependencies))
	for _, dependency := range manifest.Dependencies {
		from, fromExists := keys[strings.TrimSpace(dependency.From)]
		to, toExists := keys[strings.TrimSpace(dependency.To)]
		if !fromExists {
			return ThreadApplyPlan{}, fmt.Errorf("%w: dependency references unknown local key %q", domain.ErrValidation, dependency.From)
		}
		if !toExists {
			return ThreadApplyPlan{}, fmt.Errorf("%w: dependency references unknown local key %q", domain.ErrValidation, dependency.To)
		}
		edge := DependencyEdge{From: from, To: to}
		if edge.From == edge.To {
			return ThreadApplyPlan{}, fmt.Errorf("%w: task %s cannot depend on itself", domain.ErrValidation, edge.From)
		}
		if seenEdges[edge] {
			return ThreadApplyPlan{}, fmt.Errorf("%w: duplicate manifest dependency %s -> %s", domain.ErrValidation, dependency.From, dependency.To)
		}
		seenEdges[edge] = true
		edges = append(edges, ThreadApplyDependency{From: from, To: to})
	}
	sortThreadApplyDependencies(edges)

	plan := ThreadApplyPlan{
		Schema: ThreadApplyPlanSchema, PlanningRepoID: snapshot.PlanningRepoID,
		ComposedAt: now.Format(time.DateOnly), Dependencies: edges,
		Thread: ThreadApplyThread{
			ID: threadID, Slug: slug, Status: domain.ThreadStatusUnstarted,
			Description: manifest.Thread.Description, Goal: manifest.Thread.Goal,
			TargetDate: manifest.Thread.TargetDate, Created: now.Format(time.DateOnly),
			Tags: sortedUnique(manifest.Thread.Tags), Tasks: members, Body: body,
		},
	}
	decision, err := PrepareThreadApply(snapshot, plan)
	if err != nil {
		return ThreadApplyPlan{}, err
	}
	finalGraph := graphAfterTaskWrites(snapshot.Graph, decision.GraphPlan)
	for _, taskID := range external {
		if !taskReachesAny(finalGraph, taskID, members) {
			return ThreadApplyPlan{}, fmt.Errorf("%w: non-member node %q (%s) is not an upstream gate of any Thread member", domain.ErrValidation, declaredTasks[taskID], taskID)
		}
	}
	return decision.Plan, nil
}

// PrepareThreadApply revalidates a durable plan against current repository
// state and returns only the physical writes still needed.
func PrepareThreadApply(snapshot ThreadApplySnapshot, plan ThreadApplyPlan) (ThreadApplyDecision, error) {
	if err := ValidateThreadCreationSource(snapshot.Graph, snapshot.Threads, nil); err != nil {
		return ThreadApplyDecision{}, err
	}
	if plan.Schema != ThreadApplyPlanSchema {
		return ThreadApplyDecision{}, fmt.Errorf("%w: unsupported Thread apply-plan schema %d", domain.ErrValidation, plan.Schema)
	}
	if plan.PlanningRepoID == "" {
		return ThreadApplyDecision{}, fmt.Errorf("%w: apply plan has no planning_repo_id", domain.ErrValidation)
	}
	if snapshot.PlanningRepoID == "" {
		return ThreadApplyDecision{}, fmt.Errorf("%w: current planning repository has no durable id; run `tskflwctl config migrate`", domain.ErrValidation)
	}
	if plan.PlanningRepoID != snapshot.PlanningRepoID {
		return ThreadApplyDecision{}, fmt.Errorf("%w: apply plan belongs to planning repository %q, current repository is %q", domain.ErrConflict, plan.PlanningRepoID, snapshot.PlanningRepoID)
	}
	if err := domain.ValidateDate(plan.ComposedAt); err != nil {
		return ThreadApplyDecision{}, fmt.Errorf("apply plan composed_at: %w", err)
	}

	normalized := cloneThreadApplyPlan(plan)
	sort.Strings(normalized.Thread.Tasks)
	normalized.Thread.Tags = sortedUnique(normalized.Thread.Tags)
	sortThreadApplyDependencies(normalized.Dependencies)
	thread := normalized.Thread.domainThread()
	if thread.Status != domain.ThreadStatusUnstarted {
		return ThreadApplyDecision{}, fmt.Errorf("%w: planned Thread must be unstarted, got %q", domain.ErrValidation, thread.Status)
	}
	if thread.Slug == "" || domain.Slugify(thread.Slug) != thread.Slug {
		return ThreadApplyDecision{}, fmt.Errorf("%w: planned Thread slug %q is not a canonical filename slug", domain.ErrValidation, thread.Slug)
	}
	if thread.Created != normalized.ComposedAt {
		return ThreadApplyDecision{}, fmt.Errorf("%w: planned Thread created date %q must equal composed_at %q", domain.ErrValidation, thread.Created, normalized.ComposedAt)
	}
	if err := domain.ValidateThreadDocument(thread); err != nil {
		return ThreadApplyDecision{}, err
	}
	if len(thread.Tasks) == 0 {
		return ThreadApplyDecision{}, fmt.Errorf("%w: planned Thread requires at least one member task", domain.ErrValidation)
	}
	if _, collision := snapshot.Graph.Task(thread.ID); collision {
		return ThreadApplyDecision{}, fmt.Errorf("planned Thread id %s is already used by a task: %w", thread.ID, domain.ErrConflict)
	}

	desired := make(map[string][]string)
	changedOwners := make(map[string]bool)
	operations := make([]ThreadApplyOperation, 0, len(normalized.Dependencies)+1)
	seenEdges := make(map[DependencyEdge]bool, len(normalized.Dependencies))
	for _, dependency := range normalized.Dependencies {
		edge := DependencyEdge(dependency)
		switch {
		case !id.Valid(edge.From), !id.Valid(edge.To):
			return ThreadApplyDecision{}, fmt.Errorf("%w: planned dependency %q -> %q requires exact stable task ids", domain.ErrValidation, edge.From, edge.To)
		case edge.From == edge.To:
			return ThreadApplyDecision{}, fmt.Errorf("%w: task %s cannot depend on itself", domain.ErrValidation, edge.From)
		case seenEdges[edge]:
			return ThreadApplyDecision{}, fmt.Errorf("%w: duplicate planned dependency %s -> %s", domain.ErrValidation, edge.From, edge.To)
		}
		seenEdges[edge] = true
		if _, exists := snapshot.Graph.Task(edge.From); !exists {
			return ThreadApplyDecision{}, fmt.Errorf("%w: planned prerequisite %s does not exist", domain.ErrValidation, edge.From)
		}
		dependent, exists := snapshot.Graph.Task(edge.To)
		if !exists {
			return ThreadApplyDecision{}, fmt.Errorf("%w: planned dependent %s does not exist", domain.ErrValidation, edge.To)
		}
		dependencies, initialized := desired[edge.To]
		if !initialized {
			dependencies = append([]string(nil), dependent.DependsOn...)
		}
		state := ThreadApplySkipped
		if !sliceContainsExact(dependencies, edge.From) {
			dependencies = append(dependencies, edge.From)
			state = ThreadApplyPending
			changedOwners[edge.To] = true
		}
		desired[edge.To] = sortedUnique(dependencies)
		operations = append(operations, ThreadApplyOperation{
			Kind: "dependency", Action: "add", State: state,
			DependentID: edge.To, PrerequisiteID: edge.From,
		})
	}

	owners := make([]string, 0, len(desired))
	for taskID := range desired {
		if changedOwners[taskID] {
			owners = append(owners, taskID)
		}
	}
	sort.Strings(owners)
	graphPlan := TaskGraphMutationPlan{}
	for _, taskID := range owners {
		task, _ := snapshot.Graph.Task(taskID)
		if reflect.DeepEqual(task.DependsOn, desired[taskID]) {
			continue
		}
		graphPlan.TaskWrites = append(graphPlan.TaskWrites, TaskDependencyWrite{TaskID: taskID, DependsOn: desired[taskID]})
	}
	validatedGraphPlan, err := ValidateTaskGraphMutationPlan(snapshot.Graph, graphPlan)
	if err != nil {
		return ThreadApplyDecision{}, err
	}

	var threadPlan *ThreadCreationPlan
	threadState := ThreadApplyPending
	if existing, exists := threadByID(snapshot.Threads, thread.ID); exists {
		body, hasBody := snapshot.ThreadBodies[thread.ID]
		if !hasBody {
			return ThreadApplyDecision{}, fmt.Errorf("%w: authoritative body for existing planned Thread %s is unavailable", domain.ErrValidation, thread.ID)
		}
		if difference := plannedThreadDifference(existing, body, normalized.Thread); difference != "" {
			if plannedThreadDefinitionMatches(existing, body, normalized.Thread) {
				return ThreadApplyDecision{}, fmt.Errorf("planned Thread %s already exists and has advanced since this plan was applied (%s); the plan will not overwrite it: %w",
					thread.ID, difference, domain.ErrConflict)
			}
			return ThreadApplyDecision{}, fmt.Errorf("planned Thread id %s already exists with different %s: %w", thread.ID, difference, domain.ErrConflict)
		}
		threadState = ThreadApplySkipped
	} else {
		validated, validateErr := ValidateThreadCreationPlan(
			ThreadCreationSnapshot{Graph: snapshot.Graph, Threads: snapshot.Threads},
			ThreadCreationPlan{Thread: thread, Body: normalized.Thread.Body},
		)
		if validateErr != nil {
			return ThreadApplyDecision{}, validateErr
		}
		threadPlan = &validated
	}
	operations = append(operations, ThreadApplyOperation{
		Kind: "thread", Action: "create", State: threadState, ThreadID: thread.ID,
	})
	return ThreadApplyDecision{
		Plan: normalized, GraphPlan: validatedGraphPlan, ThreadPlan: threadPlan,
		Operations: operations,
	}, nil
}

func cloneThreadApplyPlan(plan ThreadApplyPlan) ThreadApplyPlan {
	plan.Thread.Tags = append([]string(nil), plan.Thread.Tags...)
	plan.Thread.Tasks = append([]string(nil), plan.Thread.Tasks...)
	plan.Dependencies = append([]ThreadApplyDependency(nil), plan.Dependencies...)
	return plan
}

func sortThreadApplyDependencies(edges []ThreadApplyDependency) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].To != edges[j].To {
			return edges[i].To < edges[j].To
		}
		return edges[i].From < edges[j].From
	})
}

func sliceContainsExact(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func graphAfterTaskWrites(graph *TaskGraph, plan TaskGraphMutationPlan) *TaskGraph {
	ids := graph.TaskIDs()
	tasks := make(map[string]domain.Task, len(ids))
	for _, taskID := range ids {
		task, _ := graph.Task(taskID)
		tasks[taskID] = task
	}
	for _, write := range plan.TaskWrites {
		task := tasks[write.TaskID]
		task.DependsOn = append([]string(nil), write.DependsOn...)
		tasks[write.TaskID] = task
	}
	return taskGraphFromMap(ids, tasks)
}

func taskReachesAny(graph *TaskGraph, start string, targets []string) bool {
	wanted := make(map[string]bool, len(targets))
	for _, target := range targets {
		wanted[target] = true
	}
	for _, dependent := range graph.Downstream(start) {
		if wanted[dependent] {
			return true
		}
	}
	return false
}

func plannedThreadDefinitionMatches(existing domain.Thread, body string, planned ThreadApplyThread) bool {
	want := planned.domainThread()
	return existing.ID == want.ID && existing.Slug == want.Slug &&
		existing.Description == want.Description && existing.Goal == want.Goal && existing.TargetDate == want.TargetDate &&
		existing.Created == want.Created &&
		reflect.DeepEqual(existing.Tags, want.Tags) && reflect.DeepEqual(existing.Tasks, want.Tasks) && body == planned.Body
}

func plannedThreadDifference(existing domain.Thread, body string, planned ThreadApplyThread) string {
	want := planned.domainThread()
	differences := []struct {
		name      string
		different bool
	}{
		{"id", existing.ID != want.ID},
		{"slug", existing.Slug != want.Slug},
		{"status", existing.Status != want.Status},
		{"description", existing.Description != want.Description},
		{"goal", existing.Goal != want.Goal},
		{"target_date", existing.TargetDate != want.TargetDate},
		{"created", existing.Created != want.Created},
		{"updated_at", existing.Updated != ""},
		{"started_at", existing.StartedAt != ""},
		{"ended_at", existing.EndedAt != ""},
		{"tags", !reflect.DeepEqual(existing.Tags, want.Tags)},
		{"tasks", !reflect.DeepEqual(existing.Tasks, want.Tasks)},
		{"body", body != planned.Body},
	}
	for _, difference := range differences {
		if difference.different {
			return difference.name
		}
	}
	return ""
}
