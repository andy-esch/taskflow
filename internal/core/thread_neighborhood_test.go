package core

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func neighborhoodProjection() ThreadGraphProjection {
	node := func(id, label string, role ThreadTaskRole) ThreadGraphNode {
		return ThreadGraphNode{
			TaskID: id, Label: label, Status: domain.StatusNextUp, Role: role,
			State: TaskGraphState{TaskID: id, Role: RoleQueued, Gate: GateClear},
		}
	}
	return ThreadGraphProjection{
		View: ThreadView{GraphHealth: GraphDegraded, ProjectionHealth: GraphDegraded},
		Nodes: []ThreadGraphNode{
			node("6g0000000001", "left-gate", ThreadTaskExternalGate),
			node("6g0000000002", "right-parent", ThreadTaskMember),
			node("6g0000000003", "focus-task", ThreadTaskMember),
			node("6g0000000004", "left-child", ThreadTaskMember),
			node("6g0000000005", "right-child", ThreadTaskMember),
			node("6g0000000006", "shared-join", ThreadTaskMember),
			node("6g0000000007", "isolated", ThreadTaskMember),
		},
		Edges: []ThreadGraphEdge{
			{From: "6g0000000001", To: "6g0000000003"},
			{From: "6g0000000002", To: "6g0000000003"},
			{From: "6g0000000003", To: "6g0000000004"},
			{From: "6g0000000003", To: "6g0000000005"},
			{From: "6g0000000004", To: "6g0000000006"},
			{From: "6g0000000005", To: "6g0000000006"},
		},
		Waves: []ThreadGraphWave{
			{Index: 1, TaskIDs: []string{"6g0000000002", "6g0000000007"}},
			{Index: 2, TaskIDs: []string{"6g0000000003"}},
			{Index: 3, TaskIDs: []string{"6g0000000004", "6g0000000005"}},
			{Index: 4, TaskIDs: []string{"6g0000000006"}},
		},
		TopologyComplete: false,
	}
}

func TestSelectThreadGraphNeighborhoodKeepsDirectedInducedEdgesAndBoundaryEvidence(t *testing.T) {
	projection := neighborhoodProjection()
	selected, err := SelectThreadGraphNeighborhood(projection, "focus", 1)
	if err != nil {
		t.Fatal(err)
	}
	wantIDs := []string{"6g0000000001", "6g0000000002", "6g0000000003", "6g0000000004", "6g0000000005"}
	gotIDs := make([]string, len(selected.Nodes))
	for index, node := range selected.Nodes {
		gotIDs[index] = node.TaskID
	}
	if !slices.Equal(gotIDs, wantIDs) {
		t.Fatalf("nodes=%v want=%v", gotIDs, wantIDs)
	}
	wantEdges := projection.Edges[:4]
	if !slices.Equal(selected.Edges, wantEdges) {
		t.Fatalf("edges=%v want=%v", selected.Edges, wantEdges)
	}
	wantBoundary := []ThreadGraphEdge{
		{From: "6g0000000004", To: "6g0000000006"},
		{From: "6g0000000005", To: "6g0000000006"},
	}
	if selected.Scope == nil || selected.Scope.Kind != ThreadGraphScopeNeighborhood ||
		selected.Scope.FocalTaskID != "6g0000000003" || selected.Scope.Depth != 1 ||
		selected.Scope.TotalNodes != 7 || selected.Scope.ShownNodes != 5 || selected.Scope.HiddenNodes != 2 ||
		selected.Scope.TotalEdges != 6 || selected.Scope.ShownEdges != 4 || selected.Scope.HiddenEdges != 2 ||
		!slices.Equal(selected.Scope.BoundaryEdges, wantBoundary) {
		t.Fatalf("scope=%+v", selected.Scope)
	}
	if selected.TopologyComplete || selected.View.GraphHealth != GraphDegraded || selected.View.ProjectionHealth != GraphDegraded {
		t.Fatalf("source diagnosis was changed: %+v", selected)
	}
	wantWaves := []ThreadGraphWave{
		{Index: 1, TaskIDs: []string{"6g0000000002"}},
		{Index: 2, TaskIDs: []string{"6g0000000003"}},
		{Index: 3, TaskIDs: []string{"6g0000000004", "6g0000000005"}},
	}
	if !reflect.DeepEqual(selected.Waves, wantWaves) {
		t.Fatalf("waves=%v want=%v", selected.Waves, wantWaves)
	}
}

func TestSelectThreadGraphNeighborhoodDepthTwoAndExternalFocus(t *testing.T) {
	projection := neighborhoodProjection()
	selected, err := SelectThreadGraphNeighborhood(projection, "left-gate", 2)
	if err != nil {
		t.Fatal(err)
	}
	wantIDs := []string{
		"6g0000000001", "6g0000000002", "6g0000000003", "6g0000000004", "6g0000000005",
	}
	gotIDs := make([]string, len(selected.Nodes))
	for index, node := range selected.Nodes {
		gotIDs[index] = node.TaskID
	}
	if !slices.Equal(gotIDs, wantIDs) || selected.Scope.HiddenNodes != 2 || len(selected.Scope.BoundaryEdges) != 2 {
		t.Fatalf("nodes=%v scope=%+v", gotIDs, selected.Scope)
	}
	wantWaves := []ThreadGraphWave{
		{Index: 1, TaskIDs: []string{"6g0000000002"}},
		{Index: 2, TaskIDs: []string{"6g0000000003"}},
		{Index: 3, TaskIDs: []string{"6g0000000004", "6g0000000005"}},
	}
	if !reflect.DeepEqual(selected.Waves, wantWaves) {
		t.Fatalf("external focus leaked into waves or changed indexes: got=%v want=%v", selected.Waves, wantWaves)
	}

	allConnected, err := SelectThreadGraphNeighborhood(projection, "focus-task", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(allConnected.Nodes) != 6 || allConnected.Scope.HiddenNodes != 1 ||
		len(allConnected.Edges) != len(projection.Edges) || len(allConnected.Scope.BoundaryEdges) != 0 {
		t.Fatalf("depth-two=%+v", allConnected)
	}
}

func TestSelectThreadGraphNeighborhoodHandlesIsolatedAndSharedNeighbors(t *testing.T) {
	projection := neighborhoodProjection()
	isolated, err := SelectThreadGraphNeighborhood(projection, "isolated", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(isolated.Nodes) != 1 || len(isolated.Edges) != 0 || len(isolated.Scope.BoundaryEdges) != 0 || isolated.Scope.HiddenNodes != 6 {
		t.Fatalf("isolated=%+v", isolated)
	}

	join, err := SelectThreadGraphNeighborhood(projection, "shared-join", 1)
	if err != nil {
		t.Fatal(err)
	}
	want := []ThreadGraphEdge{
		{From: "6g0000000004", To: "6g0000000006"},
		{From: "6g0000000005", To: "6g0000000006"},
	}
	wantWaves := []ThreadGraphWave{
		{Index: 3, TaskIDs: []string{"6g0000000004", "6g0000000005"}},
		{Index: 4, TaskIDs: []string{"6g0000000006"}},
	}
	if !slices.Equal(join.Edges, want) || join.Scope.TotalEdges != 6 || join.Scope.ShownEdges != 2 ||
		join.Scope.HiddenEdges != 4 || len(join.Scope.BoundaryEdges) != 2 || !reflect.DeepEqual(join.Waves, wantWaves) {
		t.Fatalf("shared join edges=%v scope=%+v", join.Edges, join.Scope)
	}
}

func TestSelectThreadGraphNeighborhoodIsDeterministicAcrossInputOrder(t *testing.T) {
	projection := neighborhoodProjection()
	reordered := neighborhoodProjection()
	slices.Reverse(reordered.Nodes)
	slices.Reverse(reordered.Edges)
	for index := range reordered.Waves {
		slices.Reverse(reordered.Waves[index].TaskIDs)
	}
	first, err := SelectThreadGraphNeighborhood(projection, "6g0000000003", 1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := SelectThreadGraphNeighborhood(reordered, "FOCUS-TASK", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("input order changed neighborhood:\nfirst=%+v\nsecond=%+v", first, second)
	}
}

func TestSelectThreadGraphNeighborhoodRejectsBadScopeAndFocus(t *testing.T) {
	base := neighborhoodProjection()
	tests := []struct {
		name       string
		projection ThreadGraphProjection
		focus      string
		depth      int
		want       error
		message    string
	}{
		{name: "depth", projection: base, focus: "focus", depth: 3, want: domain.ErrValidation, message: "depth"},
		{name: "unsafe ref", projection: base, focus: "../focus", depth: 1, want: domain.ErrValidation, message: "plain name"},
		{name: "outside", projection: base, focus: "not-in-thread", depth: 1, want: domain.ErrNotFound, message: "not a node in the supplied Thread graph"},
		{name: "already bounded", projection: func() ThreadGraphProjection {
			value := base
			value.Scope = &ThreadGraphScope{Kind: ThreadGraphScopeNeighborhood}
			return value
		}(), focus: "focus", depth: 1, want: domain.ErrValidation, message: "already bounded"},
		{name: "unreadable focal", projection: func() ThreadGraphProjection {
			value := base
			value.Nodes = append([]ThreadGraphNode(nil), base.Nodes...)
			value.Nodes[2].State.Role = RoleUnknown
			return value
		}(), focus: "focus", depth: 1, want: domain.ErrValidation, message: "unreadable"},
		{name: "dangling edge", projection: func() ThreadGraphProjection {
			value := base
			value.Edges = append([]ThreadGraphEdge(nil), base.Edges...)
			value.Edges[0].From = "missing"
			return value
		}(), focus: "focus", depth: 1, want: domain.ErrValidation, message: "unknown endpoint"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := SelectThreadGraphNeighborhood(test.projection, test.focus, test.depth)
			if !errors.Is(err, test.want) || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("error=%v want %v containing %q", err, test.want, test.message)
			}
		})
	}

	ambiguous := base
	ambiguous.Nodes = append([]ThreadGraphNode(nil), base.Nodes...)
	ambiguous.Nodes[0].Label = "focus-left"
	ambiguous.Nodes[1].Label = "focus-right"
	if _, err := SelectThreadGraphNeighborhood(ambiguous, "focus", 1); !errors.Is(err, domain.ErrAmbiguous) ||
		!strings.Contains(err.Error(), ambiguous.Nodes[0].TaskID) || !strings.Contains(err.Error(), ambiguous.Nodes[1].TaskID) {
		t.Fatalf("ambiguous error=%v", err)
	}
}
