package core

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

// GraphHealth describes whether one immutable repository task snapshot is safe
// for graph-sensitive decisions. Degraded is intentionally distinct from broken:
// every legacy reference resolves and can be projected for diagnostic reads, but
// those constraints are not canonical edges yet. Both degraded and broken
// snapshots fail closed for ordinary mutations and dispatch-oriented selectors.
type GraphHealth string

const (
	GraphHealthy  GraphHealth = "healthy"
	GraphDegraded GraphHealth = "degraded"
	GraphBroken   GraphHealth = "broken"
)

// GraphProblemCode is taskflow-owned diagnostic vocabulary. A graph library may
// help implement algorithms, but its error types and wording never cross this seam.
type GraphProblemCode string

const (
	ProblemUnreadable          GraphProblemCode = "unreadable-task"
	ProblemMissingTaskID       GraphProblemCode = "missing-task-id"
	ProblemTaskIDDrift         GraphProblemCode = "task-id-drift"
	ProblemDuplicateTaskID     GraphProblemCode = "duplicate-task-id"
	ProblemInvalidStatus       GraphProblemCode = "invalid-status"
	ProblemDuplicateDependency GraphProblemCode = "duplicate-dependency"
	ProblemSelfDependency      GraphProblemCode = "self-dependency"
	ProblemInvalidDependencyID GraphProblemCode = "invalid-dependency-id"
	ProblemMissingDependency   GraphProblemCode = "missing-dependency"
	ProblemCycle               GraphProblemCode = "cycle"
	ProblemLegacyMissing       GraphProblemCode = "legacy-reference-missing"
	ProblemLegacyAmbiguous     GraphProblemCode = "legacy-reference-ambiguous"
)

// GraphProblem is one deterministic, attributable reason a strict snapshot is
// broken. Cycle repeats its first ID at the end; it is empty for non-cycle defects.
type GraphProblem struct {
	Code          GraphProblemCode
	TaskID        string
	RelatedTaskID string
	Field         string
	Path          string
	Message       string
	Cycle         []string
}

type LegacyResolution string

const (
	LegacyResolved  LegacyResolution = "resolved"
	LegacyUnsafe    LegacyResolution = "unsafe"
	LegacyMissing   LegacyResolution = "missing"
	LegacyAmbiguous LegacyResolution = "ambiguous"
)

// DependencyEdge follows graph direction: From is the prerequisite and To is the
// dependent whose task file owns From in depends_on.
type DependencyEdge struct {
	From string
	To   string
}

// LegacyReference records how one legacy slug/ID maps to the canonical namespace.
// A resolved reference also carries the edge the guarded migration will add.
type LegacyReference struct {
	Value        string
	Resolution   LegacyResolution
	CandidateIDs []string
	Edge         DependencyEdge
}

// LegacyDependencyDiagnostic groups one legacy field occurrence. The production
// repository currently has six such occurrences (some contain several references),
// so lint reports six focused issues rather than one line per edge.
type LegacyDependencyDiagnostic struct {
	TaskID     string
	TaskSlug   string
	TaskPath   string
	Field      string
	References []LegacyReference
}

// MigrationReady reports whether the guarded legacy migration can rewrite this
// field occurrence. Present-but-empty fields are safe to remove; every populated
// reference must resolve to a structurally safe canonical edge.
func (d LegacyDependencyDiagnostic) MigrationReady() bool {
	for _, ref := range d.References {
		if ref.Resolution != LegacyResolved {
			return false
		}
	}
	return true
}

type LifecycleRole string

const (
	RoleQueued            LifecycleRole = "queued"
	RoleCandidate         LifecycleRole = "candidate"
	RoleInFlight          LifecycleRole = "in-flight"
	RoleParked            LifecycleRole = "parked"
	RoleNominallyComplete LifecycleRole = "nominally-complete"
	RoleWithdrawn         LifecycleRole = "withdrawn"
	RoleUnknown           LifecycleRole = "unknown"
)

// isPendingWorkRole reports whether a task is waiting to enter in-progress.
// Queued and candidate retain distinct planning meaning, but neither is an
// authorization gate: the repository graph decides whether pending work can
// start.
func isPendingWorkRole(role LifecycleRole) bool {
	return role == RoleQueued || role == RoleCandidate
}

type GateState string

const (
	GateClear   GateState = "clear"
	GateBlocked GateState = "blocked"
	GateBroken  GateState = "broken"
)

// BlockerReason is stable explanatory vocabulary. Invalid-status and cycle are
// deliberately explicit: normal commands prevent them, but hand edits and older
// binaries can still create states that must be diagnosed without euphemism.
type BlockerReason string

const (
	BlockerNotStarted       BlockerReason = "not-started"
	BlockerInFlight         BlockerReason = "in-flight"
	BlockerUnsoundCompleted BlockerReason = "unsound-completed"
	BlockerWithdrawn        BlockerReason = "withdrawn"
	BlockerMissing          BlockerReason = "missing"
	BlockerParked           BlockerReason = "parked"
	BlockerInvalidStatus    BlockerReason = "invalid-status"
	BlockerCycle            BlockerReason = "cycle"
	BlockerUnreadable       BlockerReason = "unreadable"
	BlockerInvalidReference BlockerReason = "invalid-reference"
	BlockerInvalidTask      BlockerReason = "invalid-task"
)

// Blocker explains one unfinished or broken prerequisite reachable from a task.
// Path starts at the queried task and ends at TaskID.
type Blocker struct {
	TaskID string
	Reason BlockerReason
	Path   []string
	Direct bool
}

// DependentImpact is one task reachable downstream from a queried task. Path
// starts at the queried task and ends at TaskID; Direct is true for one edge.
type DependentImpact struct {
	TaskID string
	Path   []string
	Direct bool
}

type TaskGraphState struct {
	TaskID           string
	Role             LifecycleRole
	Gate             GateState
	SoundlyCompleted bool
	Eligible         bool
	Drained          bool
	Inconsistent     bool
}

// GateExplanation keeps authorization and diagnosis coupled without making an
// empty prerequisite projection look like permission. State.Eligible is the
// authorization result; LocalProblems and Frontier explain a refusal.
type GateExplanation struct {
	Health        GraphHealth
	State         TaskGraphState
	LocalProblems []GraphProblem
	Frontier      []Blocker
}

// dagInput and dagAnalysis keep the owned structural algorithm independent from
// task lifecycle and diagnostic policy.
type dagInput struct {
	Nodes []string
	Edges []DependencyEdge
}

type dagAnalysis struct {
	CyclicComponents     [][]string
	RepresentativeCycles [][]string
	TopologicalWaves     [][]string
	TopologicalComplete  bool
}

// analyzeDAG is deliberately taskflow-owned. The bounded implementation bake-off
// did not justify retaining a public adapter seam or a third-party dependency.
func analyzeDAG(input dagInput) dagAnalysis {
	nodes := sortedUnique(input.Nodes)
	known := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		known[node] = true
	}
	outgoing := make(map[string][]string, len(nodes))
	indegree := make(map[string]int, len(nodes))
	seenEdges := make(map[DependencyEdge]bool, len(input.Edges))
	for _, edge := range input.Edges {
		if !known[edge.From] || !known[edge.To] || seenEdges[edge] {
			continue
		}
		seenEdges[edge] = true
		outgoing[edge.From] = append(outgoing[edge.From], edge.To)
		indegree[edge.To]++
	}
	for node := range outgoing {
		outgoing[node] = sortedUnique(outgoing[node])
	}

	components, cycles := stronglyConnectedCycles(nodes, outgoing)
	if len(components) > 0 {
		return dagAnalysis{CyclicComponents: components, RepresentativeCycles: cycles}
	}
	remaining := len(nodes)
	current := make([]string, 0)
	for _, node := range nodes {
		if indegree[node] == 0 {
			current = append(current, node)
		}
	}
	waves := make([][]string, 0)
	for len(current) > 0 {
		waves = append(waves, append([]string(nil), current...))
		next := make([]string, 0)
		for _, node := range current {
			remaining--
			for _, dependent := range outgoing[node] {
				indegree[dependent]--
				if indegree[dependent] == 0 {
					next = append(next, dependent)
				}
			}
		}
		sort.Strings(next)
		current = next
	}
	if remaining != 0 {
		components, cycles = stronglyConnectedCycles(nodes, outgoing)
		return dagAnalysis{CyclicComponents: components, RepresentativeCycles: cycles}
	}
	return dagAnalysis{TopologicalWaves: waves, TopologicalComplete: true}
}

type soundResult struct {
	sound  bool
	broken bool
}

type taskReferenceCandidate struct {
	id   string
	slug string
}

// TaskGraph is an immutable projection over one repository scan. Its internal
// query caches are synchronized; callers always receive copies of slices/maps.
type TaskGraph struct {
	tasks               map[string]domain.Task
	ids                 []string
	dependencies        map[string][]string
	outgoing            map[string][]string
	problems            []GraphProblem
	legacy              []LegacyDependencyDiagnostic
	health              GraphHealth
	hardBroken          map[string]bool
	unreadableIDs       map[string]bool
	referenceCandidates []taskReferenceCandidate
	cycleMembers        map[string]bool
	sound               map[string]soundResult
	states              map[string]TaskGraphState
	waves               [][]string
	wavesComplete       bool

	mu            sync.Mutex
	causalCache   map[string][]Blocker
	frontierCache map[string][]Blocker
	impactCache   map[string][]DependentImpact
	soundVisits   map[string]int
}

// NewTaskGraph builds the production strict snapshot with the owned analyzer.
func NewTaskGraph(tasks []domain.Task, unreadable []domain.FileProblem) *TaskGraph {
	return NewTaskGraphRead(TaskGraphReadFromFiles(tasks, unreadable))
}

// NewTaskGraphRead builds the strict snapshot from the neutral adapter read
// contract used by Service graph consumers.
func NewTaskGraphRead(read TaskGraphRead) *TaskGraph {
	return newTaskGraph(read.Tasks, read.Problems)
}

func newTaskGraph(tasks []domain.Task, unreadable []TaskGraphLoadProblem) *TaskGraph {
	g := &TaskGraph{
		tasks:         make(map[string]domain.Task, len(tasks)),
		dependencies:  make(map[string][]string, len(tasks)),
		outgoing:      make(map[string][]string, len(tasks)),
		hardBroken:    make(map[string]bool),
		unreadableIDs: make(map[string]bool),
		cycleMembers:  make(map[string]bool),
		sound:         make(map[string]soundResult, len(tasks)),
		states:        make(map[string]TaskGraphState, len(tasks)),
		causalCache:   make(map[string][]Blocker),
		frontierCache: make(map[string][]Blocker),
		impactCache:   make(map[string][]DependentImpact),
		soundVisits:   make(map[string]int, len(tasks)),
	}
	for _, problem := range unreadable {
		taskID, taskSlug := problem.TaskID, problem.TaskSlug
		if id.Valid(taskID) {
			g.unreadableIDs[taskID] = true
			g.referenceCandidates = append(g.referenceCandidates, taskReferenceCandidate{id: taskID, slug: taskSlug})
			g.hardBroken[taskID] = true
		}
		message := "unreadable task record: " + problem.Message
		if problem.Path != "" {
			message = "unreadable task file: " + problem.Message
		}
		g.problems = append(g.problems, GraphProblem{
			Code: ProblemUnreadable, TaskID: taskID, Path: problem.Path,
			Message: message,
		})
	}

	ordered := append([]domain.Task(nil), tasks...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := canonicalTaskID(ordered[i]), canonicalTaskID(ordered[j])
		if left != right {
			return left < right
		}
		if ordered[i].Path != ordered[j].Path {
			return ordered[i].Path < ordered[j].Path
		}
		return ordered[i].Slug < ordered[j].Slug
	})
	idCounts := make(map[string]int, len(ordered))
	idPaths := make(map[string][]string, len(ordered))
	for _, task := range ordered {
		if taskID := canonicalTaskID(task); taskID != "" {
			idCounts[taskID]++
			idPaths[taskID] = append(idPaths[taskID], displayPath(task.Path))
		}
	}
	for _, task := range ordered {
		taskID := canonicalTaskID(task)
		if taskID != "" {
			g.referenceCandidates = append(g.referenceCandidates, taskReferenceCandidate{id: taskID, slug: task.Slug})
		}
		if strings.TrimSpace(task.ID) == "" {
			g.addProblem(GraphProblem{Code: ProblemMissingTaskID, TaskID: taskID, Field: "id", Path: task.Path,
				Message: "missing stable task id in frontmatter"})
			g.hardBroken[taskID] = true
		}
		if taskID == "" {
			continue
		}
		if task.ID != "" && task.FilenameID != "" && task.ID != task.FilenameID {
			g.addProblem(GraphProblem{Code: ProblemTaskIDDrift, TaskID: taskID, RelatedTaskID: task.ID, Field: "id", Path: task.Path,
				Message: fmt.Sprintf("frontmatter id %q disagrees with filename id %q", task.ID, task.FilenameID)})
			g.hardBroken[taskID] = true
		}
		if idCounts[taskID] > 1 {
			g.addProblem(GraphProblem{Code: ProblemDuplicateTaskID, TaskID: taskID, Field: "id", Path: task.Path,
				Message: fmt.Sprintf("duplicate stable task id %q across %s; no source is uniquely authoritative", taskID, strings.Join(idPaths[taskID], ", "))})
			g.hardBroken[taskID] = true
		}
		if _, exists := g.tasks[taskID]; !exists {
			g.tasks[taskID] = cloneTask(task)
			g.ids = append(g.ids, taskID)
		}
		if !task.Status.Valid() {
			g.addProblem(GraphProblem{Code: ProblemInvalidStatus, TaskID: taskID, Field: "status", Path: task.Path,
				Message: fmt.Sprintf("task %s has missing or invalid status %q", taskID, task.Status)})
			g.hardBroken[taskID] = true
		}
	}
	sort.Strings(g.ids)

	canonicalEdges := make([]DependencyEdge, 0)
	for _, task := range ordered {
		taskID := canonicalTaskID(task)
		if taskID == "" {
			continue
		}
		representative, isRepresentative := g.tasks[taskID]
		isRepresentative = isRepresentative && representative.Path == task.Path && representative.Slug == task.Slug
		dependencies := append([]string(nil), task.DependsOn...)
		sort.Strings(dependencies)
		if isRepresentative {
			g.dependencies[taskID] = sortedUnique(dependencies)
		}
		seen := make(map[string]bool, len(dependencies))
		for _, prerequisite := range dependencies {
			if seen[prerequisite] {
				g.addProblem(GraphProblem{Code: ProblemDuplicateDependency, TaskID: taskID, RelatedTaskID: prerequisite,
					Field: "depends_on", Path: task.Path,
					Message: fmt.Sprintf("task %s repeats dependency %s", taskID, prerequisite)})
				g.hardBroken[taskID] = true
				continue
			}
			seen[prerequisite] = true
			switch {
			case prerequisite == taskID:
				g.addProblem(GraphProblem{Code: ProblemSelfDependency, TaskID: taskID, RelatedTaskID: prerequisite,
					Field: "depends_on", Path: task.Path,
					Message: fmt.Sprintf("task %s cannot depend on itself", taskID)})
				g.hardBroken[taskID] = true
				// Retain the representative self-edge for exact SCC membership.
				if isRepresentative {
					canonicalEdges = append(canonicalEdges, DependencyEdge{From: prerequisite, To: taskID})
					g.outgoing[prerequisite] = append(g.outgoing[prerequisite], taskID)
				}
			case !id.Valid(prerequisite):
				g.addProblem(GraphProblem{Code: ProblemInvalidDependencyID, TaskID: taskID, RelatedTaskID: prerequisite,
					Field: "depends_on", Path: task.Path,
					Message: fmt.Sprintf("task %s depends_on value %q is not a stable task id", taskID, prerequisite)})
				g.hardBroken[taskID] = true
			case !taskExists(g.tasks, prerequisite) && !g.unreadableIDs[prerequisite]:
				g.addProblem(GraphProblem{Code: ProblemMissingDependency, TaskID: taskID, RelatedTaskID: prerequisite,
					Field: "depends_on", Path: task.Path,
					Message: fmt.Sprintf("task %s depends on missing task %s", taskID, prerequisite)})
				g.hardBroken[taskID] = true
			default:
				if isRepresentative && taskExists(g.tasks, prerequisite) {
					canonicalEdges = append(canonicalEdges, DependencyEdge{From: prerequisite, To: taskID})
					g.outgoing[prerequisite] = append(g.outgoing[prerequisite], taskID)
				}
			}
		}
	}
	legacyDiagnostics, legacyEdges := g.resolveLegacyDiagnostics(ordered)
	g.legacy = legacyDiagnostics
	// Resolved legacy edges are semantically real constraints. Project them into
	// explanatory reads and derived state as well as structural analysis so a
	// degraded snapshot never reports a false all-clear before migration. The
	// persisted/canonical source remains Task.DependsOn; unioning here is read-only.
	for _, edge := range legacyEdges {
		if taskExists(g.tasks, edge.From) && taskExists(g.tasks, edge.To) {
			g.dependencies[edge.To] = append(g.dependencies[edge.To], edge.From)
			g.outgoing[edge.From] = append(g.outgoing[edge.From], edge.To)
		}
	}
	for taskID := range g.dependencies {
		g.dependencies[taskID] = sortedUnique(g.dependencies[taskID])
	}
	for taskID := range g.outgoing {
		g.outgoing[taskID] = sortedUnique(g.outgoing[taskID])
	}
	projectedEdges := append(append([]DependencyEdge(nil), canonicalEdges...), legacyEdges...)
	structure := analyzeDAG(dagInput{Nodes: append([]string(nil), g.ids...), Edges: projectedEdges})
	g.waves = cloneWaves(structure.TopologicalWaves)
	g.wavesComplete = structure.TopologicalComplete
	componentByTask := make(map[string]int)
	for componentIndex, component := range structure.CyclicComponents {
		for _, taskID := range component {
			g.cycleMembers[taskID] = true
			componentByTask[taskID] = componentIndex
		}
		cycle := structure.RepresentativeCycles[componentIndex]
		for _, taskID := range component {
			if len(component) == 1 && g.hasProblem(ProblemSelfDependency, taskID) {
				continue
			}
			path := ""
			if task, ok := g.tasks[taskID]; ok {
				path = task.Path
			}
			g.addProblem(GraphProblem{Code: ProblemCycle, TaskID: taskID, Field: "depends_on", Path: path,
				Message: "dependency cycle: " + strings.Join(cycle, " -> "), Cycle: append([]string(nil), cycle...)})
		}
	}
	g.markUnsafeLegacy(componentByTask)
	sortGraphProblems(g.problems)
	sort.Slice(g.legacy, func(i, j int) bool {
		if g.legacy[i].TaskID != g.legacy[j].TaskID {
			return g.legacy[i].TaskID < g.legacy[j].TaskID
		}
		if g.legacy[i].TaskPath != g.legacy[j].TaskPath {
			return g.legacy[i].TaskPath < g.legacy[j].TaskPath
		}
		return g.legacy[i].Field < g.legacy[j].Field
	})
	switch {
	case len(g.problems) > 0:
		g.health = GraphBroken
	case len(g.legacy) > 0:
		g.health = GraphDegraded
	default:
		g.health = GraphHealthy
	}

	for _, taskID := range g.ids {
		g.computeSound(taskID, make(map[string]bool))
	}
	for _, taskID := range g.ids {
		g.states[taskID] = g.deriveState(taskID)
	}
	return g
}

func taskExists(tasks map[string]domain.Task, taskID string) bool {
	_, ok := tasks[taskID]
	return ok
}

func canonicalTaskID(task domain.Task) string {
	if task.FilenameID != "" {
		return task.FilenameID
	}
	return task.ID
}

func taskIdentityFromPath(path string) (string, string) {
	base := filepath.Base(path)
	stem := strings.TrimSuffix(base, ".md")
	if len(stem) < id.Length+2 || stem[id.Length] != '-' {
		return "", ""
	}
	candidate := stem[:id.Length]
	if !id.Valid(candidate) {
		return "", ""
	}
	return candidate, stem[id.Length+1:]
}

func cloneTask(task domain.Task) domain.Task {
	task.Tags = append([]string(nil), task.Tags...)
	task.DependsOn = append([]string(nil), task.DependsOn...)
	task.LegacyBlockedBy = append([]string(nil), task.LegacyBlockedBy...)
	task.LegacyDependencies = append([]string(nil), task.LegacyDependencies...)
	task.LegacyBlocks = append([]string(nil), task.LegacyBlocks...)
	task.LegacyDependencyFields = append([]string(nil), task.LegacyDependencyFields...)
	return task
}

func displayPath(path string) string {
	if path == "" {
		return "<unknown path>"
	}
	return filepath.ToSlash(path)
}

func (g *TaskGraph) addProblem(problem GraphProblem) {
	g.problems = append(g.problems, problem)
}

func (g *TaskGraph) hasProblem(code GraphProblemCode, taskID string) bool {
	for _, problem := range g.problems {
		if problem.Code == code && problem.TaskID == taskID {
			return true
		}
	}
	return false
}

func (g *TaskGraph) resolveLegacyDiagnostics(records []domain.Task) ([]LegacyDependencyDiagnostic, []DependencyEdge) {
	bySlug := make(map[string][]string, len(g.tasks))
	for _, taskID := range g.ids {
		bySlug[g.tasks[taskID].Slug] = append(bySlug[g.tasks[taskID].Slug], taskID)
	}
	for slug := range bySlug {
		bySlug[slug] = sortedUnique(bySlug[slug])
	}

	type legacyField struct {
		name   string
		values func(domain.Task) []string
		blocks bool
	}
	fields := []legacyField{
		{name: "blocked_by", values: func(t domain.Task) []string { return t.LegacyBlockedBy }},
		{name: "dependencies", values: func(t domain.Task) []string { return t.LegacyDependencies }},
		{name: "blocks", values: func(t domain.Task) []string { return t.LegacyBlocks }, blocks: true},
	}

	var diagnostics []LegacyDependencyDiagnostic
	var edges []DependencyEdge
	seenEdges := make(map[DependencyEdge]bool)
	for _, task := range records {
		taskID := canonicalTaskID(task)
		if taskID == "" {
			continue
		}
		for _, field := range fields {
			values := sortedUnique(field.values(task))
			if len(values) == 0 && !slices.Contains(task.LegacyDependencyFields, field.name) {
				continue
			}
			diagnostic := LegacyDependencyDiagnostic{
				TaskID: taskID, TaskSlug: task.Slug, TaskPath: task.Path, Field: field.name,
			}
			for _, value := range values {
				ref := LegacyReference{Value: value}
				var candidates []string
				if taskExists(g.tasks, value) {
					candidates = []string{value}
				} else {
					candidates = append([]string(nil), bySlug[value]...)
				}
				ref.CandidateIDs = candidates
				switch len(candidates) {
				case 0:
					ref.Resolution = LegacyMissing
					g.addProblem(GraphProblem{Code: ProblemLegacyMissing, TaskID: taskID, Field: field.name, Path: task.Path,
						Message: fmt.Sprintf("legacy %s reference %q on task %s has no exact task ID or slug match", field.name, value, taskID)})
					g.hardBroken[taskID] = true
				case 1:
					ref.Resolution = LegacyResolved
					if field.blocks {
						ref.Edge = DependencyEdge{From: taskID, To: candidates[0]}
					} else {
						ref.Edge = DependencyEdge{From: candidates[0], To: taskID}
					}
					if ref.Edge.From == ref.Edge.To {
						ref.Resolution = LegacyUnsafe
						g.addProblem(GraphProblem{Code: ProblemSelfDependency, TaskID: taskID, RelatedTaskID: taskID,
							Field: field.name, Path: task.Path,
							Message: fmt.Sprintf("legacy %s reference %q makes task %s depend on itself", field.name, value, taskID)})
						g.hardBroken[taskID] = true
					}
					if !seenEdges[ref.Edge] {
						seenEdges[ref.Edge] = true
						edges = append(edges, ref.Edge)
					}
				default:
					ref.Resolution = LegacyAmbiguous
					g.addProblem(GraphProblem{Code: ProblemLegacyAmbiguous, TaskID: taskID, Field: field.name, Path: task.Path,
						Message: fmt.Sprintf("legacy %s reference %q on task %s is ambiguous across task IDs %s", field.name, value, taskID, strings.Join(candidates, ", "))})
					g.hardBroken[taskID] = true
				}
				diagnostic.References = append(diagnostic.References, ref)
			}
			diagnostics = append(diagnostics, diagnostic)
		}
	}
	return diagnostics, edges
}

func (g *TaskGraph) markUnsafeLegacy(componentByTask map[string]int) {
	for i := range g.legacy {
		for j := range g.legacy[i].References {
			ref := &g.legacy[i].References[j]
			if ref.Resolution != LegacyResolved {
				continue
			}
			fromComponent, fromCyclic := componentByTask[ref.Edge.From]
			toComponent, toCyclic := componentByTask[ref.Edge.To]
			if fromCyclic && toCyclic && fromComponent == toComponent {
				ref.Resolution = LegacyUnsafe
				g.hardBroken[g.legacy[i].TaskID] = true
			}
		}
	}
}

func (g *TaskGraph) computeSound(taskID string, visiting map[string]bool) soundResult {
	if result, ok := g.sound[taskID]; ok {
		return result
	}
	g.soundVisits[taskID]++
	task, ok := g.tasks[taskID]
	if !ok {
		result := soundResult{broken: true}
		g.sound[taskID] = result
		return result
	}
	if visiting[taskID] || g.cycleMembers[taskID] || g.hardBroken[taskID] || !task.Status.Valid() || task.Status == domain.StatusDeprecated {
		result := soundResult{broken: true}
		g.sound[taskID] = result
		return result
	}
	visiting[taskID] = true
	defer delete(visiting, taskID)
	allSound := true
	for _, prerequisite := range g.dependencies[taskID] {
		result := g.computeSound(prerequisite, visiting)
		if result.broken {
			result = soundResult{broken: true}
			g.sound[taskID] = result
			return result
		}
		if !result.sound {
			allSound = false
		}
	}
	result := soundResult{sound: task.Status == domain.StatusCompleted && allSound}
	g.sound[taskID] = result
	return result
}

func roleForStatus(status domain.Status) LifecycleRole {
	switch status {
	case domain.StatusNextUp:
		return RoleQueued
	case domain.StatusReadyToStart:
		return RoleCandidate
	case domain.StatusInProgress:
		return RoleInFlight
	case domain.StatusDeferred:
		return RoleParked
	case domain.StatusCompleted:
		return RoleNominallyComplete
	case domain.StatusDeprecated:
		return RoleWithdrawn
	default:
		return RoleUnknown
	}
}

func (g *TaskGraph) gate(taskID string) GateState {
	task, ok := g.tasks[taskID]
	if !ok || !task.Status.Valid() || g.hardBroken[taskID] {
		return GateBroken
	}
	blocked := false
	for _, prerequisite := range g.dependencies[taskID] {
		result := g.computeSound(prerequisite, make(map[string]bool))
		if result.broken {
			return GateBroken
		}
		if !result.sound {
			blocked = true
		}
	}
	if blocked {
		return GateBlocked
	}
	return GateClear
}

func (g *TaskGraph) deriveState(taskID string) TaskGraphState {
	task, ok := g.tasks[taskID]
	if !ok {
		return TaskGraphState{TaskID: taskID, Role: RoleUnknown, Gate: GateBroken}
	}
	role := roleForStatus(task.Status)
	gate := g.gate(taskID)
	sound := g.computeSound(taskID, make(map[string]bool)).sound
	return TaskGraphState{
		TaskID: taskID, Role: role, Gate: gate, SoundlyCompleted: sound,
		Eligible:     g.health == GraphHealthy && isPendingWorkRole(role) && gate == GateClear,
		Drained:      role == RoleNominallyComplete && sound,
		Inconsistent: (role == RoleInFlight || role == RoleNominallyComplete) && gate != GateClear,
	}
}

func (g *TaskGraph) Health() GraphHealth { return g.health }

// MutationReady is deliberately stricter than "no cycle": legacy constraints must
// be migrated and every task must be readable before an ordinary graph write.
func (g *TaskGraph) MutationReady() bool { return g.health == GraphHealthy }

func (g *TaskGraph) Problems() []GraphProblem {
	out := make([]GraphProblem, len(g.problems))
	copy(out, g.problems)
	for i := range out {
		out[i].Cycle = append([]string(nil), out[i].Cycle...)
	}
	return out
}

func (g *TaskGraph) LegacyDiagnostics() []LegacyDependencyDiagnostic {
	out := make([]LegacyDependencyDiagnostic, len(g.legacy))
	for i, diagnostic := range g.legacy {
		out[i] = diagnostic
		out[i].References = make([]LegacyReference, len(diagnostic.References))
		copy(out[i].References, diagnostic.References)
		for j := range out[i].References {
			out[i].References[j].CandidateIDs = append([]string(nil), diagnostic.References[j].CandidateIDs...)
		}
	}
	return out
}

func (g *TaskGraph) Task(taskID string) (domain.Task, bool) {
	task, ok := g.tasks[taskID]
	task = cloneTask(task)
	task.SourceVersion = "" // persistence token is not planner/query data
	return task, ok
}

// ResolveTaskID applies the ordinary task-reference tiers to this immutable
// snapshot and returns the canonical stable ID. Keeping resolution on TaskGraph
// lets guarded planners resolve user input without Store re-entry or a pre-lock
// TOCTOU choice. Exact unreadable IDs remain addressable for diagnostic queries.
func (g *TaskGraph) ResolveTaskID(ref string) (string, error) {
	if ref == "" || strings.ContainsAny(ref, `/\`) || strings.Contains(ref, "..") {
		return "", fmt.Errorf("%w: task name %q must be a plain name (no path separators)", domain.ErrValidation, ref)
	}
	candidates := append([]taskReferenceCandidate(nil), g.referenceCandidates...)
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].id != candidates[j].id {
			return candidates[i].id < candidates[j].id
		}
		return candidates[i].slug < candidates[j].slug
	})

	query := strings.ToLower(ref)
	exact := func(key string) bool { return key == ref || strings.ToLower(key) == query }
	exactIDHits := make([]taskReferenceCandidate, 0)
	for _, item := range candidates {
		if exact(item.id) {
			exactIDHits = append(exactIDHits, item)
		}
	}
	if len(exactIDHits) == 1 {
		return exactIDHits[0].id, nil
	}
	if len(exactIDHits) > 1 {
		details := make([]string, len(exactIDHits))
		for i, hit := range exactIDHits {
			if hit.slug == "" {
				details[i] = hit.id
			} else {
				details[i] = fmt.Sprintf("%s (%s)", hit.slug, hit.id)
			}
		}
		return "", fmt.Errorf("%q matches %d tasks by id: %s: %w",
			ref, len(exactIDHits), strings.Join(details, ", "), domain.ErrAmbiguous)
	}
	tiers := []func(string) bool{
		exact,
		func(key string) bool { return strings.HasPrefix(strings.ToLower(key), query) },
		func(key string) bool { return strings.Contains(strings.ToLower(key), query) },
	}
	for tier, matches := range tiers {
		hits := make([]taskReferenceCandidate, 0)
		for _, item := range candidates {
			matched := item.slug != "" && matches(item.slug)
			if tier > 0 {
				matched = matches(item.id) || matched
			}
			if matched {
				hits = append(hits, item)
			}
		}
		switch len(hits) {
		case 0:
			continue
		case 1:
			return hits[0].id, nil
		default:
			details := make([]string, len(hits))
			for i, hit := range hits {
				if hit.slug == "" {
					details[i] = hit.id
				} else {
					details[i] = fmt.Sprintf("%s (%s)", hit.slug, hit.id)
				}
			}
			return "", fmt.Errorf("%q matches %d tasks: %s: %w", ref, len(hits), strings.Join(details, ", "), domain.ErrAmbiguous)
		}
	}
	return "", fmt.Errorf("task %q: %w", ref, domain.ErrNotFound)
}

// SameSourceSnapshot reports whether two graphs came from the same exact task
// files. It exposes no version token: the store can use it as a whole-repository
// CAS after planning, while callbacks receive only Task() projections with the
// tokens cleared. Health and paths catch new unreadable/renamed entities; byte
// hashes catch every in-place edit, including non-graph fields.
func (g *TaskGraph) SameSourceSnapshot(other *TaskGraph) bool {
	if other == nil || g.health != other.health || !slices.Equal(g.ids, other.ids) {
		return false
	}
	for _, taskID := range g.ids {
		left, right := g.tasks[taskID], other.tasks[taskID]
		if left.Path != right.Path || left.SourceVersion == "" || left.SourceVersion != right.SourceVersion {
			return false
		}
	}
	return slices.EqualFunc(g.problems, other.problems, sameGraphProblem) &&
		slices.EqualFunc(g.legacy, other.legacy, sameLegacyDiagnostic)
}

func sameGraphProblem(left, right GraphProblem) bool {
	return left.Code == right.Code && left.TaskID == right.TaskID &&
		left.RelatedTaskID == right.RelatedTaskID && left.Field == right.Field &&
		left.Path == right.Path && left.Message == right.Message &&
		slices.Equal(left.Cycle, right.Cycle)
}

func sameLegacyDiagnostic(left, right LegacyDependencyDiagnostic) bool {
	return left.TaskID == right.TaskID && left.TaskSlug == right.TaskSlug &&
		left.TaskPath == right.TaskPath && left.Field == right.Field &&
		slices.EqualFunc(left.References, right.References, func(a, b LegacyReference) bool {
			return a.Value == b.Value && a.Resolution == b.Resolution &&
				a.Edge == b.Edge && slices.Equal(a.CandidateIDs, b.CandidateIDs)
		})
}

func (g *TaskGraph) TaskIDs() []string { return append([]string(nil), g.ids...) }

// Prerequisites returns the task's direct prerequisite IDs from this immutable
// projection. The result includes safely resolved legacy constraints on a
// degraded snapshot, matching State and blocker queries rather than exposing
// only the canonical field bytes from Task().
func (g *TaskGraph) Prerequisites(taskID string) []string {
	return append([]string(nil), g.dependencies[taskID]...)
}

func (g *TaskGraph) State(taskID string) TaskGraphState {
	if state, ok := g.states[taskID]; ok {
		return state
	}
	return TaskGraphState{TaskID: taskID, Role: RoleUnknown, Gate: GateBroken}
}

func (g *TaskGraph) SoundlyCompleted(taskID string) (sound, broken bool) {
	result, ok := g.sound[taskID]
	if !ok {
		return false, true
	}
	return result.sound, result.broken
}

// CausalBlockers returns every reachable unsound prerequisite. It is a forensic
// projection, not an authorization predicate; use ExplainGate().State.Eligible
// for that decision. Results carry deterministic shortest paths.
func (g *TaskGraph) CausalBlockers(taskID string) []Blocker {
	g.mu.Lock()
	defer g.mu.Unlock()
	if cached, ok := g.causalCache[taskID]; ok {
		return cloneBlockers(cached)
	}
	result := g.projectBlockers(taskID, false)
	g.causalCache[taskID] = result
	return cloneBlockers(result)
}

// BlockingFrontier returns the deepest actionable constraints while stopping at
// terminal damage. It is the bounded projection used by lint and user guidance.
func (g *TaskGraph) BlockingFrontier(taskID string) []Blocker {
	g.mu.Lock()
	defer g.mu.Unlock()
	if cached, ok := g.frontierCache[taskID]; ok {
		return cloneBlockers(cached)
	}
	result := g.projectBlockers(taskID, true)
	g.frontierCache[taskID] = result
	return cloneBlockers(result)
}

func (g *TaskGraph) projectBlockers(taskID string, frontier bool) []Blocker {
	if !taskExists(g.tasks, taskID) {
		reason := g.blockerReason(taskID)
		return []Blocker{{TaskID: taskID, Reason: reason, Path: []string{taskID}}}
	}
	queue := []string{taskID}
	visited := map[string]bool{taskID: true}
	parent := make(map[string]string)
	emitted := make(map[string]bool)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, prerequisite := range g.unsoundPrerequisites(current) {
			if visited[prerequisite] {
				continue
			}
			visited[prerequisite] = true
			parent[prerequisite] = current
			reason := g.blockerReason(prerequisite)
			children := g.unsoundPrerequisites(prerequisite)
			terminal := isTerminalBlocker(reason) || !taskExists(g.tasks, prerequisite)
			if !frontier || terminal || len(children) == 0 {
				emitted[prerequisite] = true
			}
			if taskExists(g.tasks, prerequisite) && (!frontier || (!terminal && len(children) > 0)) {
				queue = append(queue, prerequisite)
			}
		}
	}
	ids := make([]string, 0, len(emitted))
	for taskID := range emitted {
		ids = append(ids, taskID)
	}
	sort.Strings(ids)
	result := make([]Blocker, 0, len(ids))
	for _, blockerID := range ids {
		path := blockerPath(taskID, blockerID, parent)
		result = append(result, Blocker{
			TaskID: blockerID, Reason: g.blockerReason(blockerID),
			Path: path, Direct: len(path) == 2,
		})
	}
	return result
}

func (g *TaskGraph) unsoundPrerequisites(taskID string) []string {
	result := make([]string, 0, len(g.dependencies[taskID]))
	for _, prerequisite := range g.dependencies[taskID] {
		sound, ok := g.sound[prerequisite]
		if !ok || sound.broken || !sound.sound {
			result = append(result, prerequisite)
		}
	}
	return result
}

func blockerPath(root, taskID string, parent map[string]string) []string {
	reversed := []string{taskID}
	for current := taskID; current != root; {
		previous, ok := parent[current]
		if !ok {
			break
		}
		reversed = append(reversed, previous)
		current = previous
	}
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	return reversed
}

func isTerminalBlocker(reason BlockerReason) bool {
	switch reason {
	case BlockerWithdrawn, BlockerMissing, BlockerInvalidStatus, BlockerCycle,
		BlockerUnreadable, BlockerInvalidReference, BlockerInvalidTask:
		return true
	default:
		return false
	}
}

func (g *TaskGraph) blockerReason(taskID string) BlockerReason {
	task, ok := g.tasks[taskID]
	if !ok {
		if g.unreadableIDs[taskID] {
			return BlockerUnreadable
		}
		if !id.Valid(taskID) {
			return BlockerInvalidReference
		}
		return BlockerMissing
	}
	if !task.Status.Valid() {
		return BlockerInvalidStatus
	}
	if g.cycleMembers[taskID] {
		return BlockerCycle
	}
	if g.hardBroken[taskID] {
		return BlockerInvalidTask
	}
	if isPendingWorkRole(roleForStatus(task.Status)) {
		return BlockerNotStarted
	}
	switch task.Status {
	case domain.StatusInProgress:
		return BlockerInFlight
	case domain.StatusCompleted:
		return BlockerUnsoundCompleted
	case domain.StatusDeprecated:
		return BlockerWithdrawn
	case domain.StatusDeferred:
		return BlockerParked
	default:
		return BlockerInvalidStatus
	}
}

func (g *TaskGraph) LocalProblems(taskID string) []GraphProblem {
	problems := make([]GraphProblem, 0)
	for _, problem := range g.problems {
		if problem.TaskID == taskID {
			problem.Cycle = append([]string(nil), problem.Cycle...)
			problems = append(problems, problem)
		}
	}
	return problems
}

func (g *TaskGraph) ExplainGate(taskID string) GateExplanation {
	return GateExplanation{
		Health: g.health, State: g.State(taskID),
		LocalProblems: g.LocalProblems(taskID), Frontier: g.BlockingFrontier(taskID),
	}
}

// Downstream returns all transitive dependents in stable ID order, memoized per task.
func (g *TaskGraph) Downstream(taskID string) []string {
	impacts := g.DownstreamImpact(taskID)
	result := make([]string, len(impacts))
	for i, impact := range impacts {
		result[i] = impact.TaskID
	}
	return result
}

// DownstreamImpact returns every transitive dependent with one deterministic
// shortest path. Stable outgoing adjacency plus BFS fixes tie-breaking; results
// are emitted in stable task-ID order.
func (g *TaskGraph) DownstreamImpact(taskID string) []DependentImpact {
	g.mu.Lock()
	defer g.mu.Unlock()
	if cached, ok := g.impactCache[taskID]; ok {
		return cloneDependentImpacts(cached)
	}
	seen := map[string]bool{taskID: true}
	parent := make(map[string]string)
	queue := []string{taskID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, dependent := range g.outgoing[current] {
			if seen[dependent] {
				continue
			}
			seen[dependent] = true
			parent[dependent] = current
			queue = append(queue, dependent)
		}
	}
	ids := make([]string, 0, len(seen))
	for dependent := range seen {
		if dependent == taskID {
			continue
		}
		ids = append(ids, dependent)
	}
	sort.Strings(ids)
	result := make([]DependentImpact, 0, len(ids))
	for _, dependent := range ids {
		path := blockerPath(taskID, dependent, parent)
		result = append(result, DependentImpact{TaskID: dependent, Path: path, Direct: len(path) == 2})
	}
	g.impactCache[taskID] = result
	return cloneDependentImpacts(result)
}

func (g *TaskGraph) TopologicalWaves() ([][]string, bool) {
	return cloneWaves(g.waves), g.wavesComplete && g.health == GraphHealthy
}

func cloneBlockers(values []Blocker) []Blocker {
	out := make([]Blocker, len(values))
	copy(out, values)
	for i := range out {
		out[i].Path = append([]string(nil), out[i].Path...)
	}
	return out
}

func cloneDependentImpacts(values []DependentImpact) []DependentImpact {
	out := make([]DependentImpact, len(values))
	copy(out, values)
	for i := range out {
		out[i].Path = append([]string(nil), out[i].Path...)
	}
	return out
}

func cloneWaves(values [][]string) [][]string {
	out := make([][]string, len(values))
	for i := range values {
		out[i] = append([]string(nil), values[i]...)
	}
	return out
}

func sortedUnique(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := append([]string(nil), values...)
	sort.Strings(out)
	n := 0
	for _, value := range out {
		if n == 0 || value != out[n-1] {
			out[n] = value
			n++
		}
	}
	return out[:n]
}

// stronglyConnectedCycles uses Tarjan's algorithm to make cycle membership
// exact. It returns one sorted member list and one deterministic representative
// edge-following cycle for every cyclic SCC; it deliberately does not enumerate
// every simple cycle, which can be exponential.
func stronglyConnectedCycles(nodes []string, outgoing map[string][]string) ([][]string, [][]string) {
	index := 0
	indices := make(map[string]int, len(nodes))
	lowlink := make(map[string]int, len(nodes))
	onStack := make(map[string]bool, len(nodes))
	stack := make([]string, 0, len(nodes))
	components := make([][]string, 0)
	var connect func(string)
	connect = func(node string) {
		indices[node] = index
		lowlink[node] = index
		index++
		stack = append(stack, node)
		onStack[node] = true
		for _, dependent := range outgoing[node] {
			dependentIndex, visited := indices[dependent]
			if !visited {
				connect(dependent)
				if lowlink[dependent] < lowlink[node] {
					lowlink[node] = lowlink[dependent]
				}
			} else if onStack[dependent] && dependentIndex < lowlink[node] {
				lowlink[node] = dependentIndex
			}
		}
		if lowlink[node] != indices[node] {
			return
		}
		component := make([]string, 0)
		for {
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[last] = false
			component = append(component, last)
			if last == node {
				break
			}
		}
		sort.Strings(component)
		cyclic := len(component) > 1
		if len(component) == 1 {
			for _, dependent := range outgoing[component[0]] {
				if dependent == component[0] {
					cyclic = true
					break
				}
			}
		}
		if cyclic {
			components = append(components, component)
		}
	}
	for _, node := range nodes {
		if _, visited := indices[node]; !visited {
			connect(node)
		}
	}
	sort.Slice(components, func(i, j int) bool {
		return strings.Join(components[i], "\x00") < strings.Join(components[j], "\x00")
	})
	cycles := make([][]string, 0, len(components))
	for _, component := range components {
		cycles = append(cycles, representativeCycle(component, outgoing))
	}
	return components, cycles
}

func representativeCycle(component []string, outgoing map[string][]string) []string {
	start := component[0]
	if len(component) == 1 {
		return []string{start, start}
	}
	inComponent := make(map[string]bool, len(component))
	for _, taskID := range component {
		inComponent[taskID] = true
	}
	path := []string{start}
	visited := map[string]bool{start: true}
	var find func(string) bool
	find = func(current string) bool {
		for _, next := range outgoing[current] {
			if !inComponent[next] {
				continue
			}
			if next == start && len(path) > 1 {
				path = append(path, start)
				return true
			}
			if visited[next] {
				continue
			}
			visited[next] = true
			path = append(path, next)
			if find(next) {
				return true
			}
			path = path[:len(path)-1]
			delete(visited, next)
		}
		return false
	}
	if find(start) {
		return append([]string(nil), path...)
	}
	panic("cyclic strongly connected component has no representative cycle")
}

func sortGraphProblems(problems []GraphProblem) {
	sort.SliceStable(problems, func(i, j int) bool {
		left, right := problems[i], problems[j]
		lk := strings.Join([]string{left.TaskID, string(left.Code), left.Field, left.RelatedTaskID, left.Path, left.Message}, "\x00")
		rk := strings.Join([]string{right.TaskID, string(right.Code), right.Field, right.RelatedTaskID, right.Path, right.Message}, "\x00")
		return lk < rk
	})
}
