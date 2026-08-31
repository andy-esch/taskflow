package core

import (
	"sort"

	"github.com/andy-esch/taskflow/internal/domain"
)

// ThreadGraphNode is one raw, adapter-neutral task vertex in a Thread graph.
// Label and Description are deliberately unescaped; output adapters own the
// syntax and safety rules of their target format.
type ThreadGraphNode struct {
	TaskID      string
	Label       string
	Description string
	Status      domain.Status
	Role        ThreadTaskRole
	State       TaskGraphState
}

// ThreadGraphEdge follows the repository dependency direction: prerequisite to
// dependent. Both endpoints always occur in ThreadGraphProjection.Nodes.
type ThreadGraphEdge struct {
	From string
	To   string
}

// ThreadGraphWave is one explanatory generation of member tasks. Index is
// one-based for presentation; TaskIDs are stable-ID ordered. External gates are
// nodes and edges, but never masquerade as Thread-owned work in a wave.
type ThreadGraphWave struct {
	Index   int
	TaskIDs []string
}

// ThreadGraphProjection is the reusable runtime contract for graph-aware CLI,
// TUI, served, and library adapters. It owns no renderer, filesystem, framework,
// or third-party graph types. View retains the complete health and diagnostic
// evidence used to qualify the explanatory topology.
type ThreadGraphProjection struct {
	View             ThreadView
	Nodes            []ThreadGraphNode
	Edges            []ThreadGraphEdge
	Waves            []ThreadGraphWave
	TopologyComplete bool
}

// ProjectThreadGraph projects one Thread over one immutable repository task
// graph. Its boundary is deliberately the same as ProjectThread: members plus
// immediate external gates. Deeper causal context remains a blocker-query concern.
func ProjectThreadGraph(thread domain.Thread, graph *TaskGraph) ThreadGraphProjection {
	view := ProjectThread(thread, graph)
	projection := ThreadGraphProjection{View: view}
	if graph == nil {
		return projection
	}

	projection.Nodes = make([]ThreadGraphNode, 0, len(view.Members)+len(view.ExternalGates))
	for _, member := range view.Members {
		projection.Nodes = append(projection.Nodes, threadGraphNode(member))
	}
	for _, gate := range view.ExternalGates {
		projection.Nodes = append(projection.Nodes, threadGraphNode(gate.ThreadTaskView))
	}
	sort.Slice(projection.Nodes, func(i, j int) bool {
		return projection.Nodes[i].TaskID < projection.Nodes[j].TaskID
	})

	included := make(map[string]bool, len(projection.Nodes))
	memberIDs := make([]string, 0, len(view.Members))
	rankableMembers := make(map[string]bool, len(view.Members))
	rankableMemberIDs := make([]string, 0, len(view.Members))
	for _, node := range projection.Nodes {
		included[node.TaskID] = true
		if node.Role == ThreadTaskMember {
			memberIDs = append(memberIDs, node.TaskID)
			if node.State.Role != RoleUnknown && node.State.Gate != GateBroken {
				rankableMembers[node.TaskID] = true
				rankableMemberIDs = append(rankableMemberIDs, node.TaskID)
			}
		}
	}
	// Keep every repository edge induced by the bounded node set. External gates
	// are nodes rather than decorative annotations: an included member or gate may
	// itself constrain another included gate.
	for _, node := range projection.Nodes {
		for _, prerequisiteID := range graph.Prerequisites(node.TaskID) {
			if included[prerequisiteID] {
				projection.Edges = append(projection.Edges, ThreadGraphEdge{From: prerequisiteID, To: node.TaskID})
			}
		}
	}
	sort.Slice(projection.Edges, func(i, j int) bool {
		if projection.Edges[i].From != projection.Edges[j].From {
			return projection.Edges[i].From < projection.Edges[j].From
		}
		return projection.Edges[i].To < projection.Edges[j].To
	})

	memberEdges := contractThreadGraphMembers(projection.Nodes, projection.Edges, rankableMembers)
	topology := analyzeDAG(dagInput{Nodes: rankableMemberIDs, Edges: memberEdges})
	projection.Waves = make([]ThreadGraphWave, 0, len(topology.TopologicalWaves))
	for index, taskIDs := range topology.TopologicalWaves {
		projection.Waves = append(projection.Waves, ThreadGraphWave{
			Index: index + 1, TaskIDs: append([]string(nil), taskIDs...),
		})
	}
	projection.TopologyComplete = topology.TopologicalComplete && len(rankableMemberIDs) == len(memberIDs) && view.ProjectionHealth == GraphHealthy
	return projection
}

// contractThreadGraphMembers preserves ordering paths through included external
// gates without placing those gates in member-only waves. Traversal stops at the
// next member because that member's own contracted edges carry transitive order
// forward; the resulting graph is sufficient for topological generations without
// manufacturing Thread membership for a gate.
func contractThreadGraphMembers(nodes []ThreadGraphNode, edges []ThreadGraphEdge, members map[string]bool) []DependencyEdge {
	outgoing := make(map[string][]string, len(nodes))
	for _, edge := range edges {
		outgoing[edge.From] = append(outgoing[edge.From], edge.To)
	}
	for taskID := range outgoing {
		sort.Strings(outgoing[taskID])
	}

	contracted := make([]DependencyEdge, 0, len(edges))
	seenEdges := make(map[DependencyEdge]bool)
	for _, source := range nodes {
		if !members[source.TaskID] {
			continue
		}
		seenNodes := map[string]bool{source.TaskID: true}
		queue := append([]string(nil), outgoing[source.TaskID]...)
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			if seenNodes[current] {
				continue
			}
			seenNodes[current] = true
			if members[current] {
				edge := DependencyEdge{From: source.TaskID, To: current}
				if !seenEdges[edge] {
					seenEdges[edge] = true
					contracted = append(contracted, edge)
				}
				continue
			}
			queue = append(queue, outgoing[current]...)
		}
	}
	sort.Slice(contracted, func(i, j int) bool {
		if contracted[i].From != contracted[j].From {
			return contracted[i].From < contracted[j].From
		}
		return contracted[i].To < contracted[j].To
	})
	return contracted
}

func threadGraphNode(item ThreadTaskView) ThreadGraphNode {
	label := item.Task.Slug
	if label == "" {
		label = item.State.TaskID
	}
	return ThreadGraphNode{
		TaskID: item.State.TaskID, Label: label, Description: item.Task.Description,
		Status: item.Task.Status, Role: item.Role, State: item.State,
	}
}
