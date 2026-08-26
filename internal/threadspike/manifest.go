package threadspike

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

type ThreadInput struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Goal        string   `yaml:"goal"`
	Tags        []string `yaml:"tags,omitempty"`
}

type NewTaskInput struct {
	Title       string        `yaml:"title"`
	Status      domain.Status `yaml:"status,omitempty"`
	Epic        string        `yaml:"epic"`
	Description string        `yaml:"description,omitempty"`
	Effort      string        `yaml:"effort,omitempty"`
	Tier        int           `yaml:"tier,omitempty"`
	Priority    string        `yaml:"priority,omitempty"`
	Autonomy    int           `yaml:"autonomy_level,omitempty"`
	Tags        []string      `yaml:"tags"`
	Body        string        `yaml:"body,omitempty"`
}

type NodeInput struct {
	Key     string        `yaml:"key"`
	TaskID  string        `yaml:"task_id,omitempty"`
	NewTask *NewTaskInput `yaml:"new_task,omitempty"`
	Member  *bool         `yaml:"member,omitempty"`
}

type Manifest struct {
	Thread       ThreadInput `yaml:"thread"`
	Nodes        []NodeInput `yaml:"nodes"`
	Dependencies []Edge      `yaml:"dependencies,omitempty"`
}

// PlannedTask carries every byte-significant input needed to make retry
// convergence distinguish an identical prior create from an ID collision.
type PlannedTask struct {
	ID          string        `yaml:"id"`
	Slug        string        `yaml:"slug"`
	Title       string        `yaml:"title"`
	Status      domain.Status `yaml:"status"`
	Epic        string        `yaml:"epic"`
	Description string        `yaml:"description,omitempty"`
	Effort      string        `yaml:"effort"`
	Tier        int           `yaml:"tier"`
	Priority    string        `yaml:"priority"`
	Autonomy    int           `yaml:"autonomy_level"`
	Tags        []string      `yaml:"tags"`
	Created     string        `yaml:"created"`
	Body        string        `yaml:"body"`
}

func (p PlannedTask) Task(dependencies []string) Task {
	return Task{
		Record: domain.Task{
			ID:          p.ID,
			FilenameID:  p.ID,
			Slug:        p.Slug,
			Status:      p.Status,
			Epic:        p.Epic,
			Description: p.Description,
			Effort:      p.Effort,
			Tier:        p.Tier,
			Priority:    p.Priority,
			Autonomy:    p.Autonomy,
			Tags:        append([]string(nil), p.Tags...),
			Created:     p.Created,
		},
		DependsOn: append([]string(nil), dependencies...),
		Body:      p.Body,
	}
}

type PlannedThread struct {
	ID          string       `yaml:"id"`
	Slug        string       `yaml:"slug"`
	Title       string       `yaml:"title"`
	Status      ThreadStatus `yaml:"status"`
	Description string       `yaml:"description"`
	Goal        string       `yaml:"goal"`
	Created     string       `yaml:"created"`
	Tags        []string     `yaml:"tags,omitempty"`
	Tasks       []string     `yaml:"tasks"`
	Body        string       `yaml:"body"`
}

func (p PlannedThread) Thread() Thread {
	return Thread{
		ID:          p.ID,
		Slug:        p.Slug,
		Status:      p.Status,
		Description: p.Description,
		Goal:        p.Goal,
		Created:     p.Created,
		Tags:        append([]string(nil), p.Tags...),
		Tasks:       append([]string(nil), p.Tasks...),
		Body:        p.Body,
	}
}

type MaterializedPlan struct {
	Schema int            `yaml:"schema"`
	RepoID string         `yaml:"planning_space_id"`
	Thread *PlannedThread `yaml:"thread,omitempty"`
	Tasks  []PlannedTask  `yaml:"tasks,omitempty"`
	Edges  []Edge         `yaml:"dependencies,omitempty"`
}

type DependencyUpdate struct {
	TaskID    string
	DependsOn []string
}

type TaskCreate struct {
	Task      PlannedTask
	DependsOn []string
}

type ApplyDecision struct {
	CreateTasks        []TaskCreate
	UpdateDependencies []DependencyUpdate
	CreateThread       *PlannedThread
	Skipped            []Operation
}

type Operation struct {
	Kind   string `json:"kind" yaml:"kind"`
	ID     string `json:"id" yaml:"id"`
	Action string `json:"action" yaml:"action"`
}

type Receipt struct {
	Complete bool        `json:"complete" yaml:"complete"`
	Entries  []Operation `json:"entries" yaml:"entries"`
}

type ApplyOptions struct {
	DryRun     bool
	AfterWrite func(Operation) error
}

// Compose compiles human-local keys into one durable, stable-ID apply plan.
// It does not mutate the repository.
func Compose(snapshot Snapshot, manifest Manifest, newID func() string, now time.Time) (MaterializedPlan, error) {
	if err := snapshotHealthy(snapshot); err != nil {
		return MaterializedPlan{}, err
	}
	if strings.TrimSpace(manifest.Thread.Title) == "" || strings.TrimSpace(manifest.Thread.Description) == "" || strings.TrimSpace(manifest.Thread.Goal) == "" {
		return MaterializedPlan{}, fmt.Errorf("%w: Thread title, description, and goal are required", domain.ErrValidation)
	}
	if err := domain.ValidateDescription(manifest.Thread.Description); err != nil {
		return MaterializedPlan{}, err
	}
	if len(manifest.Nodes) == 0 {
		return MaterializedPlan{}, fmt.Errorf("%w: manifest must contain at least one node", domain.ErrValidation)
	}
	usedIDs := map[string]bool{}
	for taskID := range snapshot.Tasks {
		usedIDs[taskID] = true
	}
	for threadID := range snapshot.Threads {
		usedIDs[threadID] = true
	}
	mint := func(kind string) (string, error) {
		for range 16 {
			candidate := newID()
			if !id.Valid(candidate) {
				return "", fmt.Errorf("%w: generated %s id %q is invalid", domain.ErrValidation, kind, candidate)
			}
			if !usedIDs[candidate] {
				usedIDs[candidate] = true
				return candidate, nil
			}
		}
		return "", fmt.Errorf("%w: could not mint a unique %s id", domain.ErrConflict, kind)
	}

	threadID, err := mint("Thread")
	if err != nil {
		return MaterializedPlan{}, err
	}
	threadSlug := domain.Slugify(manifest.Thread.Title)
	if threadSlug == "" {
		return MaterializedPlan{}, fmt.Errorf("%w: Thread title produced an empty slug", domain.ErrValidation)
	}
	created := now.Format(time.DateOnly)
	plan := MaterializedPlan{Schema: PlanSchema, RepoID: snapshot.RepoID}
	plan.Thread = &PlannedThread{
		ID:          threadID,
		Slug:        threadSlug,
		Title:       manifest.Thread.Title,
		Status:      ThreadUnstarted,
		Description: manifest.Thread.Description,
		Goal:        manifest.Thread.Goal,
		Created:     created,
		Tags:        sortedUnique(manifest.Thread.Tags),
		Body:        fmt.Sprintf("# Thread: %s\n\n## Context\n\n%s\n", manifest.Thread.Title, manifest.Thread.Goal),
	}

	keys := map[string]string{}
	memberIDs := []string{}
	for _, node := range manifest.Nodes {
		key := strings.TrimSpace(node.Key)
		if key == "" {
			return MaterializedPlan{}, fmt.Errorf("%w: every manifest node needs a local key", domain.ErrValidation)
		}
		if _, exists := keys[key]; exists {
			return MaterializedPlan{}, fmt.Errorf("%w: duplicate manifest node key %q", domain.ErrValidation, key)
		}
		if (node.TaskID == "") == (node.NewTask == nil) {
			return MaterializedPlan{}, fmt.Errorf("%w: node %q requires exactly one of task_id or new_task", domain.ErrValidation, key)
		}
		member := node.Member == nil || *node.Member
		var taskID string
		if node.TaskID != "" {
			taskID = node.TaskID
			if _, exists := snapshot.Tasks[taskID]; !exists {
				return MaterializedPlan{}, fmt.Errorf("%w: node %q references missing task %s", domain.ErrValidation, key, taskID)
			}
		} else {
			if !member {
				return MaterializedPlan{}, fmt.Errorf("%w: new task node %q cannot be an external-only gate", domain.ErrValidation, key)
			}
			planned, err := planTask(*node.NewTask, snapshot, mint, created)
			if err != nil {
				return MaterializedPlan{}, fmt.Errorf("node %q: %w", key, err)
			}
			taskID = planned.ID
			plan.Tasks = append(plan.Tasks, planned)
		}
		keys[key] = taskID
		if member {
			memberIDs = append(memberIDs, taskID)
		}
	}
	plan.Thread.Tasks = sortedUnique(memberIDs)
	for _, edge := range manifest.Dependencies {
		from, ok := keys[edge.From]
		if !ok {
			return MaterializedPlan{}, fmt.Errorf("%w: dependency references unknown local key %q", domain.ErrValidation, edge.From)
		}
		to, ok := keys[edge.To]
		if !ok {
			return MaterializedPlan{}, fmt.Errorf("%w: dependency references unknown local key %q", domain.ErrValidation, edge.To)
		}
		plan.Edges = append(plan.Edges, Edge{From: from, To: to})
	}
	sort.Slice(plan.Tasks, func(i, j int) bool { return plan.Tasks[i].ID < plan.Tasks[j].ID })
	sort.Slice(plan.Edges, func(i, j int) bool {
		if plan.Edges[i].From != plan.Edges[j].From {
			return plan.Edges[i].From < plan.Edges[j].From
		}
		return plan.Edges[i].To < plan.Edges[j].To
	})
	if _, err := PrepareApply(snapshot, plan); err != nil {
		return MaterializedPlan{}, err
	}
	return plan, nil
}

func planTask(input NewTaskInput, snapshot Snapshot, mint func(string) (string, error), created string) (PlannedTask, error) {
	if strings.TrimSpace(input.Title) == "" {
		return PlannedTask{}, fmt.Errorf("%w: task title is required", domain.ErrValidation)
	}
	if !snapshot.Epics[input.Epic] {
		return PlannedTask{}, fmt.Errorf("%w: unknown epic %q", domain.ErrValidation, input.Epic)
	}
	if input.Status == "" {
		input.Status = domain.StatusReadyToStart
	}
	if input.Status != domain.StatusReadyToStart && input.Status != domain.StatusNextUp {
		return PlannedTask{}, fmt.Errorf("%w: bulk-created task status must be ready-to-start or next-up", domain.ErrValidation)
	}
	if input.Effort == "" {
		input.Effort = "Unknown"
	}
	if input.Priority == "" {
		input.Priority = "medium"
	}
	if input.Tier == 0 {
		input.Tier = 3
	}
	if input.Autonomy == 0 {
		input.Autonomy = 3
	}
	if err := domain.ValidateDescription(input.Description); err != nil {
		return PlannedTask{}, err
	}
	if err := domain.ValidatePriority(input.Priority); err != nil {
		return PlannedTask{}, err
	}
	if err := domain.ValidateTier(input.Tier); err != nil {
		return PlannedTask{}, err
	}
	if err := domain.ValidateAutonomy(input.Autonomy); err != nil {
		return PlannedTask{}, err
	}
	taskID, err := mint("task")
	if err != nil {
		return PlannedTask{}, err
	}
	slug := domain.Slugify(input.Title)
	if slug == "" {
		return PlannedTask{}, fmt.Errorf("%w: task title produced an empty slug", domain.ErrValidation)
	}
	record := domain.Task{
		ID: taskID, Slug: slug, Status: input.Status, Epic: input.Epic,
		Description: input.Description, Effort: input.Effort, Tier: input.Tier,
		Priority: input.Priority, Autonomy: input.Autonomy, Tags: input.Tags, Created: created,
	}
	if err := domain.ActiveTaskFieldErr(record); err != nil {
		return PlannedTask{}, err
	}
	body := input.Body
	if body == "" {
		body = fmt.Sprintf("# %s\n\n## Objective\n\nPrototype task created by the Thread spike.\n", input.Title)
	}
	return PlannedTask{
		ID: taskID, Slug: slug, Title: input.Title, Status: input.Status, Epic: input.Epic,
		Description: input.Description, Effort: input.Effort, Tier: input.Tier,
		Priority: input.Priority, Autonomy: input.Autonomy, Tags: append([]string(nil), input.Tags...),
		Created: created, Body: body,
	}, nil
}

// PrepareApply revalidates a materialized plan against the current snapshot and
// produces only the writes still needed. That makes a retry convergent after any
// prefix of the write sequence has landed.
func PrepareApply(snapshot Snapshot, plan MaterializedPlan) (ApplyDecision, error) {
	if err := snapshotHealthy(snapshot); err != nil {
		return ApplyDecision{}, err
	}
	if plan.Schema != PlanSchema {
		return ApplyDecision{}, fmt.Errorf("%w: unsupported Thread apply-plan schema %d", domain.ErrValidation, plan.Schema)
	}
	if plan.RepoID == "" || plan.RepoID != snapshot.RepoID {
		return ApplyDecision{}, fmt.Errorf("%w: apply plan belongs to planning space %q, current space is %q", domain.ErrConflict, plan.RepoID, snapshot.RepoID)
	}
	tasks := NewGraph(snapshot.Tasks).Tasks()
	planned := map[string]PlannedTask{}
	for _, task := range plan.Tasks {
		if err := validatePlannedTask(snapshot, task); err != nil {
			return ApplyDecision{}, err
		}
		if _, duplicate := planned[task.ID]; duplicate {
			return ApplyDecision{}, fmt.Errorf("%w: duplicate planned task %s", domain.ErrValidation, task.ID)
		}
		if _, crossKindCollision := snapshot.Threads[task.ID]; crossKindCollision {
			return ApplyDecision{}, fmt.Errorf("%w: planned task id %s is already used by a Thread", domain.ErrConflict, task.ID)
		}
		planned[task.ID] = task
		if _, exists := tasks[task.ID]; !exists {
			tasks[task.ID] = task.Task(nil)
		}
	}
	seenEdges := map[string]bool{}
	for _, edge := range plan.Edges {
		key := edge.From + "\x00" + edge.To
		if seenEdges[key] {
			return ApplyDecision{}, fmt.Errorf("%w: duplicate planned dependency %s -> %s", domain.ErrValidation, edge.From, edge.To)
		}
		seenEdges[key] = true
		if _, exists := tasks[edge.From]; !exists {
			return ApplyDecision{}, fmt.Errorf("%w: planned prerequisite %s does not exist", domain.ErrValidation, edge.From)
		}
		dependent, exists := tasks[edge.To]
		if !exists {
			return ApplyDecision{}, fmt.Errorf("%w: planned dependent %s does not exist", domain.ErrValidation, edge.To)
		}
		if !contains(dependent.DependsOn, edge.From) {
			dependent.DependsOn = append(dependent.DependsOn, edge.From)
			dependent.DependsOn = sortedUnique(dependent.DependsOn)
			tasks[edge.To] = dependent
		}
	}
	candidate := NewGraph(tasks)
	if err := candidate.Validate(); err != nil {
		return ApplyDecision{}, err
	}

	decision := ApplyDecision{}
	var missingPlanned []string
	for _, taskID := range sortedPlannedTaskIDs(planned) {
		plannedTask := planned[taskID]
		desired := tasks[taskID]
		if existing, exists := snapshot.Tasks[taskID]; exists {
			if !equivalentTask(existing, plannedTask.Task(desired.DependsOn)) {
				return ApplyDecision{}, fmt.Errorf("%w: planned task id %s already exists with different content", domain.ErrConflict, taskID)
			}
			decision.Skipped = append(decision.Skipped, Operation{Kind: "task", ID: taskID, Action: "already-applied"})
			continue
		}
		missingPlanned = append(missingPlanned, taskID)
	}
	// A materialized plan is resumable only if every persisted prefix remains a
	// valid graph. Create missing tasks in topological waves so a newly written
	// task never points at another planned task that has not landed yet.
	waves, err := candidate.Plan(missingPlanned)
	if err != nil {
		return ApplyDecision{}, err
	}
	for _, wave := range waves {
		for _, taskID := range wave {
			plannedTask := planned[taskID]
			desired := tasks[taskID]
			decision.CreateTasks = append(decision.CreateTasks, TaskCreate{
				Task:      plannedTask,
				DependsOn: append([]string(nil), desired.DependsOn...),
			})
		}
	}

	for _, taskID := range sortedTaskIDs(tasks) {
		if _, isNew := planned[taskID]; isNew {
			continue
		}
		existing, exists := snapshot.Tasks[taskID]
		if !exists {
			continue
		}
		desired := tasks[taskID]
		if !reflect.DeepEqual(sortedUnique(existing.DependsOn), sortedUnique(desired.DependsOn)) {
			decision.UpdateDependencies = append(decision.UpdateDependencies, DependencyUpdate{TaskID: taskID, DependsOn: sortedUnique(desired.DependsOn)})
		}
	}

	if plan.Thread != nil {
		if err := validatePlannedThread(*plan.Thread); err != nil {
			return ApplyDecision{}, err
		}
		if _, crossKindCollision := tasks[plan.Thread.ID]; crossKindCollision {
			return ApplyDecision{}, fmt.Errorf("%w: planned Thread id %s is already used by a task", domain.ErrConflict, plan.Thread.ID)
		}
		for _, memberID := range plan.Thread.Tasks {
			if _, exists := tasks[memberID]; !exists {
				return ApplyDecision{}, fmt.Errorf("%w: planned Thread member %s does not exist", domain.ErrValidation, memberID)
			}
		}
		if existing, exists := snapshot.Threads[plan.Thread.ID]; exists {
			if !equivalentThread(existing, plan.Thread.Thread()) {
				return ApplyDecision{}, fmt.Errorf("%w: planned Thread id %s already exists with different content", domain.ErrConflict, plan.Thread.ID)
			}
			decision.Skipped = append(decision.Skipped, Operation{Kind: "thread", ID: plan.Thread.ID, Action: "already-applied"})
		} else {
			copyThread := *plan.Thread
			decision.CreateThread = &copyThread
		}
	}
	return decision, nil
}

func validatePlannedTask(snapshot Snapshot, task PlannedTask) error {
	if !id.Valid(task.ID) {
		return fmt.Errorf("%w: invalid planned task id %q", domain.ErrValidation, task.ID)
	}
	if strings.TrimSpace(task.Title) == "" || task.Slug == "" || domain.Slugify(task.Title) != task.Slug {
		return fmt.Errorf("%w: planned task %s has inconsistent title/slug identity", domain.ErrValidation, task.ID)
	}
	if task.Status != domain.StatusReadyToStart && task.Status != domain.StatusNextUp {
		return fmt.Errorf("%w: planned task %s status must be ready-to-start or next-up", domain.ErrValidation, task.ID)
	}
	if !snapshot.Epics[task.Epic] {
		return fmt.Errorf("%w: planned task %s references unknown epic %q", domain.ErrValidation, task.ID, task.Epic)
	}
	if err := domain.ValidateDescription(task.Description); err != nil {
		return err
	}
	if err := domain.ValidatePriority(task.Priority); err != nil {
		return err
	}
	if err := domain.ValidateTier(task.Tier); err != nil {
		return err
	}
	if err := domain.ValidateAutonomy(task.Autonomy); err != nil {
		return err
	}
	if err := domain.ValidateDate(task.Created); err != nil {
		return err
	}
	if err := domain.ActiveTaskFieldErr(task.Task(nil).Record); err != nil {
		return err
	}
	return nil
}

func validatePlannedThread(thread PlannedThread) error {
	if !id.Valid(thread.ID) {
		return fmt.Errorf("%w: invalid planned Thread id %q", domain.ErrValidation, thread.ID)
	}
	if strings.TrimSpace(thread.Title) == "" || thread.Slug == "" || domain.Slugify(thread.Title) != thread.Slug {
		return fmt.Errorf("%w: planned Thread %s has inconsistent title/slug identity", domain.ErrValidation, thread.ID)
	}
	if thread.Status != ThreadUnstarted {
		return fmt.Errorf("%w: planned Thread %s must start unstarted", domain.ErrValidation, thread.ID)
	}
	if strings.TrimSpace(thread.Description) == "" || strings.TrimSpace(thread.Goal) == "" {
		return fmt.Errorf("%w: planned Thread %s requires description and goal", domain.ErrValidation, thread.ID)
	}
	if err := domain.ValidateDescription(thread.Description); err != nil {
		return err
	}
	if err := domain.ValidateDate(thread.Created); err != nil {
		return err
	}
	if len(thread.Tasks) != len(sortedUnique(thread.Tasks)) {
		return fmt.Errorf("%w: planned Thread %s repeats a member task", domain.ErrValidation, thread.ID)
	}
	return nil
}

func snapshotHealthy(snapshot Snapshot) error {
	if len(snapshot.Problems) > 0 {
		return fmt.Errorf("%w: planning graph cannot be read soundly: %s: %s", domain.ErrValidation, snapshot.Problems[0].Path, snapshot.Problems[0].Message)
	}
	if err := NewGraph(snapshot.Tasks).Validate(); err != nil {
		return err
	}
	return nil
}

func equivalentTask(existing, desired Task) bool {
	a, b := existing.Record, desired.Record
	a.Path, a.FilenameID, a.StatusFellBack = "", "", false
	b.Path, b.FilenameID, b.StatusFellBack = "", "", false
	return reflect.DeepEqual(a, b) &&
		reflect.DeepEqual(sortedUnique(existing.DependsOn), sortedUnique(desired.DependsOn)) &&
		existing.Body == desired.Body
}

func equivalentThread(existing, desired Thread) bool {
	existing.Path, desired.Path = "", ""
	existing.Tasks, desired.Tasks = sortedUnique(existing.Tasks), sortedUnique(desired.Tasks)
	existing.Tags, desired.Tags = sortedUnique(existing.Tags), sortedUnique(desired.Tags)
	return reflect.DeepEqual(existing, desired)
}

func sortedPlannedTaskIDs(tasks map[string]PlannedTask) []string {
	ids := make([]string, 0, len(tasks))
	for taskID := range tasks {
		ids = append(ids, taskID)
	}
	sort.Strings(ids)
	return ids
}
