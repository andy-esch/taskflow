package core

import (
	"fmt"
	"sort"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

// ValidateTaskGraphMutationSource rejects an authoritative snapshot that cannot
// support a sound write. Degraded legacy snapshots remain eligible for the one
// guarded migration that converges them to the canonical field.
func ValidateTaskGraphMutationSource(graph *TaskGraph) error {
	if graph == nil {
		return fmt.Errorf("%w: authoritative task graph is required", domain.ErrValidation)
	}
	if graph.Health() == GraphBroken {
		return fmt.Errorf("%w: repository task graph is broken: %s",
			domain.ErrValidation, taskGraphHealthDetail(graph))
	}
	return nil
}

// ValidateTaskGraphMutationPlan normalizes task-owned sets and proves that every
// planner-ordered durable prefix, plus the final state, remains a sound graph. It
// is pure so command planners can validate and preview without a filesystem; the
// store owns only locking, materialization, CAS, and atomic replacement.
func ValidateTaskGraphMutationPlan(graph *TaskGraph, plan TaskGraphMutationPlan) (TaskGraphMutationPlan, error) {
	if graph == nil {
		return TaskGraphMutationPlan{}, fmt.Errorf("%w: authoritative task graph is required", domain.ErrValidation)
	}
	normalized := TaskGraphMutationPlan{TaskWrites: make([]TaskDependencyWrite, len(plan.TaskWrites))}
	copy(normalized.TaskWrites, plan.TaskWrites)
	for i := range normalized.TaskWrites {
		normalized.TaskWrites[i].DependsOn = append([]string(nil), normalized.TaskWrites[i].DependsOn...)
	}
	// A plan that only adds canonical edges has a monotone safety property: every
	// physical prefix is an edge-subset of the final graph. If the final graph is
	// healthy, no prefix can introduce a missing reference or cycle that the final
	// graph does not also contain. Keep per-prefix graph reconstruction for removals
	// and legacy-field clearing, where that argument does not hold.
	additiveOnly := taskGraphMutationOnlyAddsEdges(graph, normalized)

	seenTasks := make(map[string]bool, len(normalized.TaskWrites))
	taskIDs := graph.TaskIDs()
	prospective := make(map[string]domain.Task, len(taskIDs))
	for _, taskID := range taskIDs {
		task, _ := graph.Task(taskID)
		prospective[taskID] = task
	}
	for i := range normalized.TaskWrites {
		write := &normalized.TaskWrites[i]
		if !id.Valid(write.TaskID) {
			return TaskGraphMutationPlan{}, fmt.Errorf("%w: planned task id %q is not a stable task id", domain.ErrValidation, write.TaskID)
		}
		if seenTasks[write.TaskID] {
			return TaskGraphMutationPlan{}, fmt.Errorf("%w: graph mutation plans task %s more than once", domain.ErrValidation, write.TaskID)
		}
		seenTasks[write.TaskID] = true
		task, exists := prospective[write.TaskID]
		if !exists {
			return TaskGraphMutationPlan{}, fmt.Errorf("%w: planned task %s does not exist in the authoritative snapshot", domain.ErrValidation, write.TaskID)
		}

		seenDependencies := make(map[string]bool, len(write.DependsOn))
		for _, prerequisite := range write.DependsOn {
			switch {
			case !id.Valid(prerequisite):
				return TaskGraphMutationPlan{}, fmt.Errorf("%w: planned dependency %q for task %s is not a stable task id", domain.ErrValidation, prerequisite, write.TaskID)
			case prerequisite == write.TaskID:
				return TaskGraphMutationPlan{}, fmt.Errorf("%w: task %s cannot depend on itself", domain.ErrValidation, write.TaskID)
			case seenDependencies[prerequisite]:
				return TaskGraphMutationPlan{}, fmt.Errorf("%w: task %s repeats planned dependency %s", domain.ErrValidation, write.TaskID, prerequisite)
			}
			if _, exists := prospective[prerequisite]; !exists {
				return TaskGraphMutationPlan{}, fmt.Errorf("%w: planned dependency %s for task %s does not exist", domain.ErrValidation, prerequisite, write.TaskID)
			}
			seenDependencies[prerequisite] = true
		}
		sort.Strings(write.DependsOn)
		task.DependsOn = append([]string(nil), write.DependsOn...)
		if write.ClearLegacy {
			task.LegacyBlockedBy = nil
			task.LegacyDependencies = nil
			task.LegacyBlocks = nil
			task.LegacyDependencyFields = nil
		}
		prospective[write.TaskID] = task

		if !additiveOnly {
			prefixGraph := taskGraphFromMap(taskIDs, prospective)
			if prefixGraph.Health() == GraphBroken {
				return TaskGraphMutationPlan{}, fmt.Errorf("%w: planned write prefix ending at task %s would leave a broken graph: %s",
					domain.ErrValidation, write.TaskID, taskGraphHealthDetail(prefixGraph))
			}
		}
	}

	finalGraph := taskGraphFromMap(taskIDs, prospective)
	if !finalGraph.MutationReady() {
		return TaskGraphMutationPlan{}, fmt.Errorf("%w: planned dependency state is %s; mutation requires a healthy final graph: %s",
			domain.ErrValidation, finalGraph.Health(), taskGraphHealthDetail(finalGraph))
	}
	return normalized, nil
}

func taskGraphMutationOnlyAddsEdges(graph *TaskGraph, plan TaskGraphMutationPlan) bool {
	for _, write := range plan.TaskWrites {
		if write.ClearLegacy {
			return false
		}
		task, exists := graph.Task(write.TaskID)
		if !exists {
			return false
		}
		for _, existing := range task.DependsOn {
			if !sliceContainsExact(write.DependsOn, existing) {
				return false
			}
		}
	}
	return true
}

func taskGraphFromMap(taskIDs []string, tasksByID map[string]domain.Task) *TaskGraph {
	tasks := make([]domain.Task, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		tasks = append(tasks, tasksByID[taskID])
	}
	return NewTaskGraph(tasks, nil)
}

func taskGraphHealthDetail(graph *TaskGraph) string {
	if problems := graph.Problems(); len(problems) > 0 {
		first := problems[0]
		location := ""
		if first.Path != "" {
			location = fmt.Sprintf(" in %s", first.Path)
			if first.Field != "" {
				location += fmt.Sprintf(" (field %s)", first.Field)
			}
		}
		detail := first.Message + location
		if len(problems) > 1 {
			detail += fmt.Sprintf(" (%d additional problem(s))", len(problems)-1)
		}
		return detail + "; repair the graph-owned frontmatter directly, then run `tskflwctl lint`"
	}
	if legacy := graph.LegacyDiagnostics(); len(legacy) > 0 {
		return fmt.Sprintf("%d legacy dependency field occurrence(s) remain; run `tskflwctl task depend migrate`", len(legacy))
	}
	return "graph health is not mutation-ready"
}
