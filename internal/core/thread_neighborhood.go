package core

import (
	"errors"
	"fmt"
	"sort"

	"github.com/andy-esch/taskflow/internal/domain"
)

// ThreadGraphScope describes an explicit bounded excerpt of a supplied Thread
// graph. BoundaryEdges retain the exact directed evidence crossing the excerpt;
// they are diagnostics and are not part of the induced Edges collection.
type ThreadGraphScope struct {
	Kind          ThreadGraphScopeKind
	FocalTaskID   string
	Depth         int
	TotalNodes    int
	ShownNodes    int
	HiddenNodes   int
	TotalEdges    int
	ShownEdges    int
	HiddenEdges   int
	BoundaryEdges []ThreadGraphEdge
}

type ThreadGraphScopeKind string

const ThreadGraphScopeNeighborhood ThreadGraphScopeKind = "neighborhood"

// SelectThreadGraphNeighborhood returns the subgraph induced by every supplied
// node within depth undirected hops of focusRef. Dependency direction is retained
// in Edges and BoundaryEdges. The function is pure: it never reads a repository,
// applies terminal geometry, or chooses a focal node for the caller.
func SelectThreadGraphNeighborhood(projection ThreadGraphProjection, focusRef string, depth int) (ThreadGraphProjection, error) {
	if depth != 1 && depth != 2 {
		return ThreadGraphProjection{}, fmt.Errorf("%w: Thread graph neighborhood depth must be 1 or 2, got %d", domain.ErrValidation, depth)
	}
	if projection.Scope != nil {
		return ThreadGraphProjection{}, fmt.Errorf("%w: cannot select a Thread graph neighborhood from an already bounded projection", domain.ErrValidation)
	}

	nodesByID := make(map[string]ThreadGraphNode, len(projection.Nodes))
	candidates := make([]taskReferenceCandidate, 0, len(projection.Nodes))
	for index, node := range projection.Nodes {
		if node.TaskID == "" {
			return ThreadGraphProjection{}, fmt.Errorf("%w: Thread graph node %d has an empty task ID", domain.ErrValidation, index)
		}
		if _, duplicate := nodesByID[node.TaskID]; duplicate {
			return ThreadGraphProjection{}, fmt.Errorf("%w: Thread graph repeats task ID %s", domain.ErrValidation, node.TaskID)
		}
		if node.Role != ThreadTaskMember && node.Role != ThreadTaskExternalGate {
			return ThreadGraphProjection{}, fmt.Errorf("%w: Thread graph task %s has unknown role %q", domain.ErrValidation, node.TaskID, node.Role)
		}
		nodesByID[node.TaskID] = node
		candidates = append(candidates, taskReferenceCandidate{id: node.TaskID, slug: node.Label})
	}

	adjacent := make(map[string][]string, len(projection.Nodes))
	seenEdges := make(map[ThreadGraphEdge]bool, len(projection.Edges))
	for _, edge := range projection.Edges {
		if _, ok := nodesByID[edge.From]; !ok {
			return ThreadGraphProjection{}, fmt.Errorf("%w: Thread graph edge %s -> %s has an unknown endpoint", domain.ErrValidation, edge.From, edge.To)
		}
		if _, ok := nodesByID[edge.To]; !ok {
			return ThreadGraphProjection{}, fmt.Errorf("%w: Thread graph edge %s -> %s has an unknown endpoint", domain.ErrValidation, edge.From, edge.To)
		}
		if seenEdges[edge] {
			return ThreadGraphProjection{}, fmt.Errorf("%w: Thread graph repeats edge %s -> %s", domain.ErrValidation, edge.From, edge.To)
		}
		seenEdges[edge] = true
		adjacent[edge.From] = append(adjacent[edge.From], edge.To)
		if edge.From != edge.To {
			adjacent[edge.To] = append(adjacent[edge.To], edge.From)
		}
	}
	for taskID := range adjacent {
		sort.Strings(adjacent[taskID])
	}
	waveTaskIDs := make(map[string]bool)
	waveIndexes := make(map[int]bool)
	for _, wave := range projection.Waves {
		if wave.Index < 1 || waveIndexes[wave.Index] {
			return ThreadGraphProjection{}, fmt.Errorf("%w: Thread graph has invalid or repeated wave index %d", domain.ErrValidation, wave.Index)
		}
		waveIndexes[wave.Index] = true
		for _, taskID := range wave.TaskIDs {
			node, exists := nodesByID[taskID]
			if !exists || node.Role != ThreadTaskMember {
				return ThreadGraphProjection{}, fmt.Errorf("%w: Thread graph wave %d names unknown or non-member task %s", domain.ErrValidation, wave.Index, taskID)
			}
			if waveTaskIDs[taskID] {
				return ThreadGraphProjection{}, fmt.Errorf("%w: Thread graph waves repeat task %s", domain.ErrValidation, taskID)
			}
			waveTaskIDs[taskID] = true
		}
	}

	focalTaskID, err := resolveTaskReference(focusRef, candidates)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ThreadGraphProjection{}, fmt.Errorf("task %q is not a node in the supplied Thread graph: %w", focusRef, domain.ErrNotFound)
		}
		return ThreadGraphProjection{}, err
	}
	focal := nodesByID[focalTaskID]
	if focal.State.Role == RoleUnknown {
		return ThreadGraphProjection{}, fmt.Errorf("%w: focal task %s is unreadable or has invalid lifecycle state", domain.ErrValidation, focalTaskID)
	}

	distance := map[string]int{focalTaskID: 0}
	queue := []string{focalTaskID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if distance[current] == depth {
			continue
		}
		for _, neighbor := range adjacent[current] {
			if _, visited := distance[neighbor]; visited {
				continue
			}
			distance[neighbor] = distance[current] + 1
			queue = append(queue, neighbor)
		}
	}

	selected := ThreadGraphProjection{
		View: cloneThreadView(projection.View), TopologyComplete: projection.TopologyComplete,
		Nodes: make([]ThreadGraphNode, 0, len(distance)),
		Edges: make([]ThreadGraphEdge, 0, len(projection.Edges)),
		Waves: make([]ThreadGraphWave, 0, len(projection.Waves)),
	}
	for taskID := range distance {
		selected.Nodes = append(selected.Nodes, nodesByID[taskID])
	}
	sort.Slice(selected.Nodes, func(i, j int) bool { return selected.Nodes[i].TaskID < selected.Nodes[j].TaskID })

	boundary := make([]ThreadGraphEdge, 0)
	for _, edge := range projection.Edges {
		_, fromShown := distance[edge.From]
		_, toShown := distance[edge.To]
		switch {
		case fromShown && toShown:
			selected.Edges = append(selected.Edges, edge)
		case fromShown != toShown:
			boundary = append(boundary, edge)
		}
	}
	sortThreadGraphEdges(selected.Edges)
	sortThreadGraphEdges(boundary)

	for _, wave := range projection.Waves {
		filtered := make([]string, 0, len(wave.TaskIDs))
		for _, taskID := range wave.TaskIDs {
			if _, shown := distance[taskID]; shown {
				filtered = append(filtered, taskID)
			}
		}
		if len(filtered) == 0 {
			continue
		}
		sort.Strings(filtered)
		selected.Waves = append(selected.Waves, ThreadGraphWave{Index: wave.Index, TaskIDs: filtered})
	}
	selected.Scope = &ThreadGraphScope{
		Kind: ThreadGraphScopeNeighborhood, FocalTaskID: focalTaskID, Depth: depth,
		TotalNodes: len(projection.Nodes), ShownNodes: len(selected.Nodes), HiddenNodes: len(projection.Nodes) - len(selected.Nodes),
		TotalEdges: len(projection.Edges), ShownEdges: len(selected.Edges), HiddenEdges: len(projection.Edges) - len(selected.Edges),
		BoundaryEdges: boundary,
	}
	return selected, nil
}

func sortThreadGraphEdges(edges []ThreadGraphEdge) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})
}
