package threadspike

import (
	"fmt"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

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

type GateState string

const (
	GateClear   GateState = "clear"
	GateBlocked GateState = "blocked"
	GateBroken  GateState = "broken"
)

// Edge follows the graph/rendering direction: From is the prerequisite and To
// is the dependent whose task file stores From in depends_on.
type Edge struct {
	From string `yaml:"from" json:"from"`
	To   string `yaml:"to" json:"to"`
}

type TaskView struct {
	ID           string        `json:"id"`
	Slug         string        `json:"slug,omitempty"`
	Status       domain.Status `json:"status,omitempty"`
	Role         LifecycleRole `json:"role"`
	Gate         GateState     `json:"gate"`
	External     bool          `json:"external"`
	Eligible     bool          `json:"eligible"`
	Drained      bool          `json:"drained"`
	Inconsistent bool          `json:"inconsistent"`
	DependsOn    []string      `json:"depends_on,omitempty"`
}

type ThreadView struct {
	Thread        Thread     `json:"thread"`
	Members       []TaskView `json:"members"`
	ExternalGates []TaskView `json:"external_gates"`
	Frontier      []string   `json:"frontier"`
	Done          int        `json:"done"`
	Total         int        `json:"total"`
	Withdrawn     int        `json:"withdrawn"`
	Broken        bool       `json:"broken"`
	Inconsistent  bool       `json:"inconsistent"`
}

// Graph owns no persistence. It is a deterministic projection over one task
// snapshot and deliberately uses plain stable IDs at its boundary.
type Graph struct {
	tasks    map[string]Task
	outgoing map[string][]string
}

func NewGraph(tasks map[string]Task) *Graph {
	g := &Graph{tasks: make(map[string]Task, len(tasks)), outgoing: map[string][]string{}}
	for taskID, task := range tasks {
		copyTask := task
		copyTask.DependsOn = append([]string(nil), task.DependsOn...)
		g.tasks[taskID] = copyTask
		for _, prerequisite := range task.DependsOn {
			g.outgoing[prerequisite] = append(g.outgoing[prerequisite], taskID)
		}
	}
	for prerequisite, dependents := range g.outgoing {
		g.outgoing[prerequisite] = sortedUnique(dependents)
	}
	return g
}

func (g *Graph) Tasks() map[string]Task {
	out := make(map[string]Task, len(g.tasks))
	for taskID, task := range g.tasks {
		copyTask := task
		copyTask.DependsOn = append([]string(nil), task.DependsOn...)
		out[taskID] = copyTask
	}
	return out
}

// Validate checks the repository-global hard DAG contract. Errors are stable:
// task and dependency iteration is ID-sorted, and a cycle names its exact path.
func (g *Graph) Validate() error {
	ids := sortedTaskIDs(g.tasks)
	for _, taskID := range ids {
		seen := map[string]bool{}
		dependencies := append([]string(nil), g.tasks[taskID].DependsOn...)
		sort.Strings(dependencies)
		for _, prerequisite := range dependencies {
			switch {
			case prerequisite == taskID:
				return fmt.Errorf("%w: task %s cannot depend on itself", domain.ErrValidation, taskID)
			case seen[prerequisite]:
				return fmt.Errorf("%w: task %s repeats dependency %s", domain.ErrValidation, taskID, prerequisite)
			}
			seen[prerequisite] = true
			if _, ok := g.tasks[prerequisite]; !ok {
				return fmt.Errorf("%w: task %s depends on missing task %s", domain.ErrValidation, taskID, prerequisite)
			}
		}
	}
	if cycle := g.CyclePath(); len(cycle) > 0 {
		return fmt.Errorf("%w: dependency cycle: %s", domain.ErrValidation, strings.Join(cycle, " -> "))
	}
	return nil
}

// CyclePath returns one deterministic cycle in graph direction
// prerequisite→dependent, repeating the first vertex at the end.
func (g *Graph) CyclePath() []string {
	state := map[string]uint8{}
	positions := map[string]int{}
	var stack []string
	var visit func(string) []string
	visit = func(taskID string) []string {
		state[taskID] = 1
		positions[taskID] = len(stack)
		stack = append(stack, taskID)
		for _, dependent := range g.outgoing[taskID] {
			if _, exists := g.tasks[dependent]; !exists {
				continue
			}
			switch state[dependent] {
			case 0:
				if cycle := visit(dependent); len(cycle) > 0 {
					return cycle
				}
			case 1:
				start := positions[dependent]
				cycle := append([]string(nil), stack[start:]...)
				return append(cycle, dependent)
			}
		}
		stack = stack[:len(stack)-1]
		delete(positions, taskID)
		state[taskID] = 2
		return nil
	}
	for _, taskID := range sortedTaskIDs(g.tasks) {
		if state[taskID] == 0 {
			if cycle := visit(taskID); len(cycle) > 0 {
				return cycle
			}
		}
	}
	return nil
}

func (g *Graph) WithEdges(edges []Edge) (*Graph, error) {
	tasks := g.Tasks()
	for _, edge := range edges {
		if edge.From == edge.To {
			return nil, fmt.Errorf("%w: task %s cannot depend on itself", domain.ErrValidation, edge.To)
		}
		if _, ok := tasks[edge.From]; !ok {
			return nil, fmt.Errorf("%w: prerequisite task %s does not exist", domain.ErrValidation, edge.From)
		}
		dependent, ok := tasks[edge.To]
		if !ok {
			return nil, fmt.Errorf("%w: dependent task %s does not exist", domain.ErrValidation, edge.To)
		}
		if contains(dependent.DependsOn, edge.From) {
			return nil, fmt.Errorf("%w: dependency %s -> %s already exists", domain.ErrValidation, edge.From, edge.To)
		}
		dependent.DependsOn = append(dependent.DependsOn, edge.From)
		sort.Strings(dependent.DependsOn)
		tasks[edge.To] = dependent
	}
	candidate := NewGraph(tasks)
	if err := candidate.Validate(); err != nil {
		return nil, err
	}
	return candidate, nil
}

func (g *Graph) Gate(taskID string) GateState {
	task, ok := g.tasks[taskID]
	if !ok {
		return GateBroken
	}
	blocked := false
	for _, prerequisite := range task.DependsOn {
		sound, broken := g.soundlyCompleted(prerequisite, map[string]bool{})
		if broken {
			return GateBroken
		}
		if !sound {
			blocked = true
		}
	}
	if blocked {
		return GateBlocked
	}
	return GateClear
}

func (g *Graph) soundlyCompleted(taskID string, visiting map[string]bool) (sound, broken bool) {
	task, ok := g.tasks[taskID]
	if !ok || visiting[taskID] {
		return false, true
	}
	if task.Record.Status == domain.StatusDeprecated || !task.Record.Status.Valid() {
		return false, true
	}
	visiting[taskID] = true
	defer delete(visiting, taskID)
	allPrerequisitesSound := true
	for _, prerequisite := range task.DependsOn {
		prerequisiteSound, prerequisiteBroken := g.soundlyCompleted(prerequisite, visiting)
		if prerequisiteBroken {
			return false, true
		}
		if !prerequisiteSound {
			allPrerequisitesSound = false
		}
	}
	return task.Record.Status == domain.StatusCompleted && allPrerequisitesSound, false
}

func roleFor(status domain.Status) LifecycleRole {
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

func (g *Graph) ViewTask(taskID string, external bool) TaskView {
	task, ok := g.tasks[taskID]
	if !ok {
		return TaskView{ID: taskID, Role: RoleUnknown, Gate: GateBroken, External: external}
	}
	gate := g.Gate(taskID)
	role := roleFor(task.Record.Status)
	return TaskView{
		ID:           taskID,
		Slug:         task.Record.Slug,
		Status:       task.Record.Status,
		Role:         role,
		Gate:         gate,
		External:     external,
		Eligible:     role == RoleCandidate && gate == GateClear,
		Drained:      role == RoleNominallyComplete && gate == GateClear,
		Inconsistent: (role == RoleInFlight || role == RoleNominallyComplete) && gate != GateClear,
		DependsOn:    append([]string(nil), task.DependsOn...),
	}
}

func (g *Graph) ViewThread(thread Thread) ThreadView {
	view := ThreadView{Thread: thread}
	members := sortedUnique(thread.Tasks)
	memberSet := make(map[string]bool, len(members))
	for _, taskID := range members {
		memberSet[taskID] = true
	}
	externalSet := map[string]bool{}
	for _, taskID := range members {
		taskView := g.ViewTask(taskID, false)
		view.Members = append(view.Members, taskView)
		if taskView.Role == RoleWithdrawn {
			view.Withdrawn++
		} else {
			view.Total++
			if taskView.Status == domain.StatusCompleted {
				view.Done++
			}
		}
		if taskView.Eligible {
			view.Frontier = append(view.Frontier, taskID)
		}
		if taskView.Gate == GateBroken || taskView.Role == RoleUnknown {
			view.Broken = true
		}
		if taskView.Inconsistent {
			view.Inconsistent = true
		}
		if task, ok := g.tasks[taskID]; ok {
			for _, prerequisite := range task.DependsOn {
				if !memberSet[prerequisite] {
					externalSet[prerequisite] = true
				}
			}
		}
	}
	for _, taskID := range sortedKeys(externalSet) {
		external := g.ViewTask(taskID, true)
		view.ExternalGates = append(view.ExternalGates, external)
		if external.Gate == GateBroken || external.Role == RoleUnknown {
			view.Broken = true
		}
	}
	if thread.Status == ThreadCompleted {
		for _, member := range view.Members {
			if member.Role != RoleWithdrawn && !member.Drained {
				view.Inconsistent = true
			}
		}
	}
	return view
}

// Plan returns deterministic member-only topological waves. External gates do
// not become Thread-owned work and therefore do not enter the waves.
func (g *Graph) Plan(memberIDs []string) ([][]string, error) {
	members := sortedUnique(memberIDs)
	memberSet := make(map[string]bool, len(members))
	indegree := make(map[string]int, len(members))
	for _, taskID := range members {
		if _, ok := g.tasks[taskID]; !ok {
			return nil, fmt.Errorf("%w: Thread member %s does not exist", domain.ErrValidation, taskID)
		}
		memberSet[taskID] = true
	}
	for _, taskID := range members {
		for _, prerequisite := range g.tasks[taskID].DependsOn {
			if memberSet[prerequisite] {
				indegree[taskID]++
			}
		}
	}
	remaining := len(members)
	var waves [][]string
	for remaining > 0 {
		var wave []string
		for _, taskID := range members {
			if memberSet[taskID] && indegree[taskID] == 0 {
				wave = append(wave, taskID)
			}
		}
		if len(wave) == 0 {
			return nil, fmt.Errorf("%w: Thread projection contains a dependency cycle", domain.ErrValidation)
		}
		waves = append(waves, wave)
		for _, taskID := range wave {
			memberSet[taskID] = false
			remaining--
			for _, dependent := range g.outgoing[taskID] {
				if memberSet[dependent] {
					indegree[dependent]--
				}
			}
		}
	}
	return waves, nil
}

func (g *Graph) Blockers(taskID string) []string {
	seen := map[string]bool{}
	var walk func(string)
	walk = func(current string) {
		task, ok := g.tasks[current]
		if !ok {
			return
		}
		for _, prerequisite := range task.DependsOn {
			sound, broken := g.soundlyCompleted(prerequisite, map[string]bool{})
			if (broken || !sound) && !seen[prerequisite] {
				seen[prerequisite] = true
				walk(prerequisite)
			}
		}
	}
	walk(taskID)
	return sortedKeys(seen)
}

func (g *Graph) Unblocks(taskID string) []string {
	seen := map[string]bool{}
	var walk func(string)
	walk = func(current string) {
		for _, dependent := range g.outgoing[current] {
			if !seen[dependent] {
				seen[dependent] = true
				walk(dependent)
			}
		}
	}
	walk(taskID)
	return sortedKeys(seen)
}

func sortedTaskIDs(tasks map[string]Task) []string {
	ids := make([]string, 0, len(tasks))
	for taskID := range tasks {
		ids = append(ids, taskID)
	}
	sort.Strings(ids)
	return ids
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key, present := range values {
		if present {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}
