package core

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestThreadLegacyDiagnosticsRetainRepositoryScopeAndOwnTheirValues(t *testing.T) {
	member := graphRecord("legacy-member", domain.StatusNextUp)
	peer := graphRecord("legacy-peer", domain.StatusCompleted)
	owner := graphRecord("legacy-outside", domain.StatusReadyToStart)
	owner.LegacyBlockedBy = []string{peer.Slug}
	owner.LegacyDependencyFields = []string{"blocked_by", "blocks"}
	record := func(task domain.Task) VersionedRecord[domain.Task] {
		return VersionedRecord[domain.Task]{Record: LoadedRecord[domain.Task]{Value: task,
			Source: RecordSource{ID: task.ID, Location: "db://tasks/" + task.Slug}}}
	}
	read := TaskGraphRead{GuardedRecords: []VersionedRecord[domain.Task]{record(owner), record(peer), record(member)}}
	graph := NewTaskGraphRead(read)
	thread := threadRecord(domain.ThreadStatusInProgress, member.ID)
	view := ProjectThread(thread, graph)
	if view.GraphHealth != GraphDegraded || view.ProjectionHealth != GraphDegraded || len(view.Frontier) != 0 || len(view.GraphProblems) != 0 || len(view.GraphLegacy) != 2 {
		t.Fatalf("legacy-only cause was lost or dispatch changed: %+v", view)
	}
	if len(view.ExternalGates) != 0 || view.GraphLegacy[0].TaskID != owner.ID || view.GraphLegacy[0].TaskPath != "" || view.GraphLegacy[0].Location != "db://tasks/legacy-outside" || view.GraphLegacy[0].recordRef != 0 {
		t.Fatalf("repository context became a gate or lost portable attribution: %+v", view)
	}
	for _, diagnostic := range view.GraphLegacy {
		if !diagnostic.MigrationReady() || diagnostic.MigrationGuidance() != "run task depend migrate" {
			t.Fatalf("resolved/empty field lost safe guidance: %+v", diagnostic)
		}
	}
	want := graph.LegacyDiagnostics()
	slices.Reverse(read.GuardedRecords)
	if got := ProjectThread(thread, NewTaskGraphRead(read)); !reflect.DeepEqual(got.GraphLegacy, want) {
		t.Fatalf("record order changed diagnostic order: got=%+v want=%+v", got.GraphLegacy, want)
	}
	copy := cloneThreadView(view)
	copy.GraphLegacy[0].References[0].CandidateIDs[0] = "changed"
	copy.GraphLegacy[0].References[0].Value = "changed"
	if !reflect.DeepEqual(view.GraphLegacy, want) || !reflect.DeepEqual(graph.LegacyDiagnostics(), want) {
		t.Fatal("receipt clone aliases shared diagnosis")
	}
	view.GraphLegacy[0].References[0].CandidateIDs[0] = "changed"
	if !reflect.DeepEqual(graph.LegacyDiagnostics(), want) {
		t.Fatal("Thread projection aliases graph diagnosis")
	}
	read.GuardedRecords[2].Record.Source.ID = testutil.TaskID("drifted-source")
	drift := NewTaskGraphRead(read)
	found := 0
	for _, diagnostic := range drift.LegacyDiagnostics() {
		if diagnostic.TaskSlug == owner.Slug {
			found++
			if diagnostic.TaskID != testutil.TaskID("drifted-source") {
				t.Fatalf("diagnosis used declared ID instead of source ID: %+v", diagnostic)
			}
		}
	}
	if found != 2 || drift.Health() != GraphBroken {
		t.Fatalf("source drift erased the legacy occurrences: %+v", drift.LegacyDiagnostics())
	}
}

func TestThreadLegacyDiagnosticsPreserveBrokenSeverityAndGrouping(t *testing.T) {
	member := graphRecord("group-member", domain.StatusNextUp)
	owner := graphRecord("group-owner", domain.StatusNextUp)
	one, two := graphRecord("one", domain.StatusCompleted), graphRecord("two", domain.StatusCompleted)
	one.Slug, two.Slug = "shared", "shared"
	owner.LegacyBlockedBy = []string{"missing", "shared"}
	owner.DependsOn = []string{"invalid-token"}
	graph := NewTaskGraph([]domain.Task{member, owner, one, two}, nil)
	view := ProjectThread(threadRecord(domain.ThreadStatusUnstarted, member.ID), graph)
	if view.GraphHealth != GraphBroken || len(view.Frontier) != 0 || len(view.GraphLegacy) != 1 || len(view.GraphProblems) != 3 {
		t.Fatalf("broken evidence or fail-closed behavior changed: %+v", view)
	}
	diagnostic := view.GraphLegacy[0]
	if diagnostic.MigrationReady() || diagnostic.MigrationGuidance() != "run task depend repair, then task depend migrate" || !strings.Contains(diagnostic.Summary(), `"missing": missing`) || !strings.Contains(diagnostic.Summary(), `"shared": ambiguous`) {
		t.Fatalf("unsafe references lack diagnosis/guidance: %+v", diagnostic)
	}
	for _, problem := range view.GraphProblems {
		wantGrouped := problem.Code == ProblemLegacyMissing || problem.Code == ProblemLegacyAmbiguous
		if diagnostic.ExplainsGraphProblem(problem) != wantGrouped {
			t.Fatalf("wrong diagnostic owner: %+v", problem)
		}
		problem.Location = "db://other-occurrence"
		if diagnostic.ExplainsGraphProblem(problem) {
			t.Fatal("grouping hid a different source occurrence")
		}
	}
}

func TestThreadLegacyDiagnosticsDistinguishDuplicateSourceOccurrences(t *testing.T) {
	owner := graphRecord("duplicate-legacy-owner", domain.StatusNextUp)
	owner.LegacyBlockedBy = []string{"missing"}
	record := func(location string) VersionedRecord[domain.Task] {
		return VersionedRecord[domain.Task]{Record: LoadedRecord[domain.Task]{Value: owner,
			Source: RecordSource{ID: owner.ID, Location: location}}}
	}
	read := TaskGraphRead{GuardedRecords: []VersionedRecord[domain.Task]{record("db://second"), record("db://first")}}
	graph := NewTaskGraphRead(read)
	want := graph.LegacyDiagnostics()
	if graph.Health() != GraphBroken || len(want) != 2 || want[0].Location == want[1].Location {
		t.Fatalf("duplicate sources lost distinct legacy occurrences: %+v", want)
	}
	missing := 0
	for _, problem := range graph.Problems() {
		owners := 0
		for _, diagnostic := range want {
			if diagnostic.ExplainsGraphProblem(problem) {
				owners++
			}
		}
		if problem.Code == ProblemLegacyMissing {
			missing++
			if owners != 1 {
				t.Fatalf("missing reference has %d occurrence owners: %+v", owners, problem)
			}
		} else if owners != 0 {
			t.Fatalf("legacy grouping hid structural duplicate evidence: %+v", problem)
		}
	}
	if missing != 2 {
		t.Fatalf("expected both missing-reference occurrences, got %d", missing)
	}
	slices.Reverse(read.GuardedRecords)
	if got := NewTaskGraphRead(read).LegacyDiagnostics(); !reflect.DeepEqual(got, want) {
		t.Fatalf("source order changed duplicate diagnostics: got=%+v want=%+v", got, want)
	}
}

type threadLegacyGraphSource struct {
	testSourceSetProvider
	read  TaskGraphRead
	calls int
}

func (s *threadLegacyGraphSource) ReadTaskGraph() (TaskGraphRead, error) {
	s.calls++
	return s.read, nil
}

func TestThreadListLegacyDiagnosticsHoistAndInvalidateContinuation(t *testing.T) {
	member := graphRecord("listed-legacy-member", domain.StatusNextUp)
	owner := graphRecord("listed-legacy-outside", domain.StatusCompleted)
	owner.LegacyBlockedBy = []string{member.ID}
	graphs := &threadLegacyGraphSource{read: TaskGraphRead{GuardedRecords: []VersionedRecord[domain.Task]{
		{Record: LoadedRecord[domain.Task]{Value: member, Source: RecordSource{ID: member.ID}}},
		{Record: LoadedRecord[domain.Task]{Value: owner, Source: RecordSource{ID: owner.ID, Location: "db://first"}}},
	}}}
	first := threadRecord(domain.ThreadStatusUnstarted, member.ID)
	second := first
	second.ID, second.Slug = testutil.TaskID("second-thread"), "second-thread"
	threads := &threadReadFake{threads: []domain.Thread{second, first}, thread: first}
	svc := MustNewService(nil, WithThreadStore(threads), WithTaskGraphSource(graphs))
	list, page, _, err := svc.QueryThreadViews("", ListQuery{Limit: 1})
	if err != nil || page == nil || page.NextCursor == "" || graphs.calls != 1 || len(list.GraphLegacy) != 1 || len(list.Threads) != 1 || len(list.Threads[0].GraphLegacy) != 0 {
		t.Fatalf("diagnostics not hoisted from one snapshot: list=%+v page=%+v reads=%d err=%v", list, page, graphs.calls, err)
	}
	q := ListQuery{Limit: 1, Cursor: page.NextCursor}
	if _, _, _, err := svc.QueryThreadViews("", q); err != nil {
		t.Fatalf("unchanged context refused continuation: %v", err)
	}
	graphs.read.GuardedRecords[1].Record.Source.Location = "db://second"
	if _, _, _, err := svc.QueryThreadViews("", q); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("changed outside diagnosis permitted stale continuation: %v", err)
	}
	empty, _, _, err := svc.QueryThreadViews("", ListQuery{Search: "no-match"})
	if err != nil || len(empty.Threads) != 0 || len(empty.GraphLegacy) != 1 {
		t.Fatalf("empty selected list lost repository cause: %+v %v", empty, err)
	}
	show, _, err := svc.ShowThread(first.ID)
	if err != nil || len(show.GraphLegacy) != 1 || show.GraphLegacy[0].Location != "db://second" {
		t.Fatalf("show lost the same shared diagnosis: %+v %v", show, err)
	}
}

func TestThreadGraphRepairImpactsCarryLegacyBeforeAndAfter(t *testing.T) {
	member := graphRecord("impact-legacy-member", domain.StatusNextUp)
	owner := graphRecord("impact-legacy-outside", domain.StatusCompleted)
	owner.LegacyDependencyFields = []string{"blocks"}
	before := NewTaskGraph([]domain.Task{member, owner}, nil)
	owner.LegacyDependencyFields = nil
	after := NewTaskGraph([]domain.Task{member, owner}, nil)
	thread := threadRecord(domain.ThreadStatusUnstarted, member.ID)
	impacts := TaskGraphThreadImpacts([]LoadedRecord[domain.Thread]{{Value: thread, Source: RecordSource{ID: thread.ID}}}, before, after, []string{owner.ID})
	if len(impacts) != 1 || impacts[0].Direct || len(impacts[0].Before.GraphLegacy) != 1 || len(impacts[0].After.GraphLegacy) != 0 || len(impacts[0].Before.Frontier) != 0 || len(impacts[0].After.Frontier) != 1 {
		t.Fatalf("repair impacts lost repository diagnosis or eligibility: %+v", impacts)
	}
}
