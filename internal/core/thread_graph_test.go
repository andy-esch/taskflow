package core

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestProjectThreadGraphBoundsAndOrdersNeutralProjection(t *testing.T) {
	external := graphRecord("external-gate", domain.StatusCompleted)
	first := graphRecord("first-member", domain.StatusCompleted, external.ID)
	second := graphRecord("second-member", domain.StatusReadyToStart, first.ID)
	disconnected := graphRecord("disconnected-member", domain.StatusNextUp)
	thread := domain.Thread{
		ID: "6g3q4rtmv4ak", Slug: "delivery", Status: domain.ThreadStatusInProgress,
		Description: "Ship the graph", Goal: "Render it", Created: "2026-08-31",
		Tasks: []string{second.ID, disconnected.ID, first.ID},
	}
	slices.Sort(thread.Tasks)

	projection := ProjectThreadGraph(thread, NewTaskGraph([]domain.Task{second, external, disconnected, first}, nil))
	reordered := ProjectThreadGraph(thread, NewTaskGraph([]domain.Task{first, disconnected, external, second}, nil))
	if !reflect.DeepEqual(projection, reordered) {
		t.Fatalf("repository scan order changed projection:\nfirst=%+v\nsecond=%+v", projection, reordered)
	}
	if projection.View.ProjectionHealth != GraphHealthy || !projection.TopologyComplete {
		t.Fatalf("health=%s complete=%v", projection.View.ProjectionHealth, projection.TopologyComplete)
	}
	wantNodeIDs := []string{disconnected.ID, external.ID, first.ID, second.ID}
	slices.Sort(wantNodeIDs)
	gotNodeIDs := make([]string, len(projection.Nodes))
	roles := make(map[string]ThreadTaskRole)
	for i, node := range projection.Nodes {
		gotNodeIDs[i], roles[node.TaskID] = node.TaskID, node.Role
	}
	if !slices.Equal(gotNodeIDs, wantNodeIDs) {
		t.Fatalf("nodes=%v want=%v", gotNodeIDs, wantNodeIDs)
	}
	if roles[external.ID] != ThreadTaskExternalGate || roles[first.ID] != ThreadTaskMember {
		t.Fatalf("roles=%v", roles)
	}
	wantEdges := []ThreadGraphEdge{{From: external.ID, To: first.ID}, {From: first.ID, To: second.ID}}
	slices.SortFunc(wantEdges, func(a, b ThreadGraphEdge) int {
		if a.From != b.From {
			return compareString(a.From, b.From)
		}
		return compareString(a.To, b.To)
	})
	if !slices.Equal(projection.Edges, wantEdges) {
		t.Fatalf("edges=%v want=%v", projection.Edges, wantEdges)
	}
	if len(projection.Waves) != 2 || projection.Waves[0].Index != 1 || projection.Waves[1].Index != 2 {
		t.Fatalf("waves=%v", projection.Waves)
	}
	if !slices.Contains(projection.Waves[0].TaskIDs, first.ID) ||
		!slices.Contains(projection.Waves[0].TaskIDs, disconnected.ID) ||
		!slices.Equal(projection.Waves[1].TaskIDs, []string{second.ID}) {
		t.Fatalf("member-only waves=%v", projection.Waves)
	}
}

func TestProjectThreadGraphPreservesMemberOrderThroughIncludedGate(t *testing.T) {
	beforeGate := fixedGraphRecord("6g0000000001", "before-gate", domain.StatusNextUp)
	gate := fixedGraphRecord("6g0000000002", "external-gate", domain.StatusNextUp, beforeGate.ID)
	afterGate := fixedGraphRecord("6g0000000003", "after-gate", domain.StatusNextUp, gate.ID)
	thread := domain.Thread{
		ID: "6g0000000004", Slug: "gate-path", Status: domain.ThreadStatusUnstarted,
		Description: "Keep the bounded graph honest", Goal: "Preserve member ordering", Created: "2026-08-31",
		Tasks: []string{beforeGate.ID, afterGate.ID},
	}
	slices.Sort(thread.Tasks)

	projection := ProjectThreadGraph(thread, NewTaskGraph([]domain.Task{afterGate, gate, beforeGate}, nil))
	wantEdges := []ThreadGraphEdge{
		{From: beforeGate.ID, To: gate.ID},
		{From: gate.ID, To: afterGate.ID},
	}
	if !slices.Equal(projection.Edges, wantEdges) {
		t.Fatalf("edges=%v want=%v", projection.Edges, wantEdges)
	}
	if len(projection.Waves) != 2 ||
		!slices.Equal(projection.Waves[0].TaskIDs, []string{beforeGate.ID}) ||
		!slices.Equal(projection.Waves[1].TaskIDs, []string{afterGate.ID}) {
		t.Fatalf("member waves did not preserve gate-traversing order: %v", projection.Waves)
	}
	for _, wave := range projection.Waves {
		if slices.Contains(wave.TaskIDs, gate.ID) {
			t.Fatalf("external gate became Thread-owned work: %v", projection.Waves)
		}
	}
	if !projection.TopologyComplete {
		t.Fatalf("healthy bounded topology reported partial: %+v", projection)
	}
}

func TestProjectThreadGraphSortsEdgesByPrerequisiteThenDependent(t *testing.T) {
	firstPrerequisite := fixedGraphRecord("6g0000000004", "first-prerequisite", domain.StatusNextUp)
	firstDependent := fixedGraphRecord("6g0000000009", "first-dependent", domain.StatusNextUp, firstPrerequisite.ID)
	secondPrerequisite := fixedGraphRecord("6g0000000008", "second-prerequisite", domain.StatusNextUp)
	secondDependent := fixedGraphRecord("6g0000000002", "second-dependent", domain.StatusNextUp, secondPrerequisite.ID)
	thread := domain.Thread{
		ID: "6g0000000001", Slug: "edge-order", Status: domain.ThreadStatusUnstarted,
		Description: "Pin the edge contract", Goal: "Remain deterministic", Created: "2026-08-31",
		Tasks: []string{firstPrerequisite.ID, firstDependent.ID, secondPrerequisite.ID, secondDependent.ID},
	}
	slices.Sort(thread.Tasks)

	projection := ProjectThreadGraph(thread, NewTaskGraph([]domain.Task{
		firstDependent, secondDependent, firstPrerequisite, secondPrerequisite,
	}, nil))
	want := []ThreadGraphEdge{
		{From: firstPrerequisite.ID, To: firstDependent.ID},
		{From: secondPrerequisite.ID, To: secondDependent.ID},
	}
	if !slices.Equal(projection.Edges, want) {
		t.Fatalf("edges=%v want=%v", projection.Edges, want)
	}
}

func TestProjectThreadGraphQualifiesUsefulTopologyWithHealth(t *testing.T) {
	member := graphRecord("member", domain.StatusNextUp)
	brokenElsewhere := graphRecord("broken-elsewhere", domain.StatusNextUp, "6g0000000009")
	thread := domain.Thread{
		ID: "6g3q4rtmv4ak", Slug: "delivery", Status: domain.ThreadStatusUnstarted,
		Description: "Diagnose the graph", Goal: "Never overclaim", Created: "2026-08-31",
		Tasks: []string{member.ID},
	}

	projection := ProjectThreadGraph(thread, NewTaskGraph([]domain.Task{member, brokenElsewhere}, nil))
	if projection.View.GraphHealth != GraphBroken || projection.TopologyComplete {
		t.Fatalf("health=%s complete=%v", projection.View.GraphHealth, projection.TopologyComplete)
	}
	if len(projection.Waves) != 1 || !slices.Equal(projection.Waves[0].TaskIDs, []string{member.ID}) {
		t.Fatalf("useful partial topology lost: %v", projection.Waves)
	}
}

func TestProjectThreadGraphNilGraphIsDiagnosticAndEmpty(t *testing.T) {
	thread := domain.Thread{ID: "6g3q4rtmv4ak", Slug: "delivery"}
	projection := ProjectThreadGraph(thread, nil)
	if projection.TopologyComplete || len(projection.Nodes) != 0 || len(projection.View.Problems) == 0 {
		t.Fatalf("projection=%+v", projection)
	}
}

func TestProjectThreadGraphLeavesBrokenMembersUnranked(t *testing.T) {
	missingID := "6g0000000009"
	thread := domain.Thread{
		ID: "6g3q4rtmv4ak", Slug: "broken", Status: domain.ThreadStatusUnstarted,
		Description: "Keep broken work visible", Goal: "Do not rank it", Created: "2026-08-31",
		Tasks: []string{missingID},
	}
	projection := ProjectThreadGraph(thread, NewTaskGraph(nil, nil))
	if projection.TopologyComplete || len(projection.Nodes) != 1 || len(projection.Waves) != 0 ||
		projection.Nodes[0].State.Gate != GateBroken {
		t.Fatalf("projection=%+v", projection)
	}
}

func TestProjectThreadGraphSharedTaskRoleIsLocalToEachThread(t *testing.T) {
	shared := graphRecord("shared", domain.StatusCompleted)
	dependent := graphRecord("dependent", domain.StatusNextUp, shared.ID)
	graph := NewTaskGraph([]domain.Task{dependent, shared}, nil)
	memberThread := domain.Thread{
		ID: "6g0000000001", Slug: "member", Status: domain.ThreadStatusUnstarted,
		Description: "Own shared membership", Goal: "Project a member", Created: "2026-08-31", Tasks: []string{shared.ID},
	}
	gateThread := domain.Thread{
		ID: "6g0000000002", Slug: "gate", Status: domain.ThreadStatusUnstarted,
		Description: "Use shared boundary", Goal: "Project a gate", Created: "2026-08-31", Tasks: []string{dependent.ID},
	}
	memberProjection := ProjectThreadGraph(memberThread, graph)
	gateProjection := ProjectThreadGraph(gateThread, graph)
	if len(memberProjection.Nodes) != 1 || memberProjection.Nodes[0].Role != ThreadTaskMember {
		t.Fatalf("member projection=%+v", memberProjection.Nodes)
	}
	roles := make(map[string]ThreadTaskRole)
	for _, node := range gateProjection.Nodes {
		roles[node.TaskID] = node.Role
	}
	if roles[shared.ID] != ThreadTaskExternalGate || roles[dependent.ID] != ThreadTaskMember {
		t.Fatalf("gate projection roles=%v", roles)
	}
}

func TestProjectThreadGraphLargeDeepWideTopology(t *testing.T) {
	const width, depth = 128, 128
	tasks := make([]domain.Task, 0, width+depth)
	memberIDs := make([]string, 0, width+depth)
	waveOneIDs := make([]string, 0, width)
	for index := range width {
		task := graphRecord(fmt.Sprintf("wide-%03d", index), domain.StatusNextUp)
		tasks = append(tasks, task)
		memberIDs = append(memberIDs, task.ID)
		waveOneIDs = append(waveOneIDs, task.ID)
	}
	previous := append([]string(nil), waveOneIDs...)
	for index := range depth {
		task := graphRecord(fmt.Sprintf("deep-%03d", index), domain.StatusNextUp, previous...)
		tasks = append(tasks, task)
		memberIDs = append(memberIDs, task.ID)
		previous = []string{task.ID}
	}
	slices.Sort(memberIDs)
	thread := domain.Thread{
		ID: "6g3q4rtmv4ak", Slug: "large", Status: domain.ThreadStatusUnstarted,
		Description: "Stress a large projection", Goal: "Stay deterministic", Created: "2026-08-31", Tasks: memberIDs,
	}
	projection := ProjectThreadGraph(thread, NewTaskGraph(tasks, nil))
	if !projection.TopologyComplete || len(projection.Nodes) != width+depth || len(projection.Waves) != depth+1 {
		t.Fatalf("nodes=%d waves=%d complete=%v", len(projection.Nodes), len(projection.Waves), projection.TopologyComplete)
	}
	if len(projection.Waves[0].TaskIDs) != width || len(projection.Waves[len(projection.Waves)-1].TaskIDs) != 1 {
		t.Fatalf("wave widths first=%d last=%d", len(projection.Waves[0].TaskIDs), len(projection.Waves[len(projection.Waves)-1].TaskIDs))
	}
}

func compareString(left, right string) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func fixedGraphRecord(id, slug string, status domain.Status, dependencies ...string) domain.Task {
	return domain.Task{
		ID: id, FilenameID: id, Slug: slug, Path: "tasks/" + id + "-" + slug + ".md",
		Status: status, Description: slug, Tags: []string{"graph"},
		DependsOn: append([]string(nil), dependencies...),
	}
}

func TestServiceShowThreadGraphUsesIndependentPortsInOrder(t *testing.T) {
	task := graphRecord("member", domain.StatusReadyToStart)
	thread := domain.Thread{
		ID: "6g3q4rtmv4ak", Slug: "ordered", Status: domain.ThreadStatusUnstarted,
		Description: "Ordered graph reads", Goal: "Keep adapters portable", Created: "2026-08-31",
		Tasks: []string{task.ID},
	}
	calls := make([]string, 0, 2)
	graphs := &taskGraphReadFake{tasks: []domain.Task{task}, onList: func() { calls = append(calls, "tasks") }}
	threads := &threadReadFake{thread: thread, onGet: func() { calls = append(calls, "threads") }}
	projection, err := NewService(nil, WithTaskGraphSource(graphs), WithThreadStore(threads)).ShowThreadGraph(thread.ID)
	if err != nil || !slices.Equal(calls, []string{"threads", "tasks"}) || len(projection.Nodes) != 1 {
		t.Fatalf("calls=%v projection=%+v err=%v", calls, projection, err)
	}
}
