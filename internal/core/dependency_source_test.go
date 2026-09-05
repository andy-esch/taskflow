package core

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestTaskGraphSourceViewsSeparateRawCanonicalAndProjectedDependencies(t *testing.T) {
	alpha := graphRecord("source-alpha", domain.StatusCompleted)
	beta := graphRecord("source-beta", domain.StatusCompleted)
	gamma := graphRecord("source-gamma", domain.StatusCompleted)
	owner := graphRecord("source-owner", domain.StatusReadyToStart, beta.ID, beta.ID)
	owner.LegacyBlockedBy = []string{gamma.Slug}
	owner.LegacyDependencies = []string{beta.ID}
	owner.LegacyBlocks = []string{alpha.Slug}
	owner.LegacyDependencyFields = []string{"blocked_by", "dependencies", "blocks"}
	tasks := []domain.Task{owner, gamma, alpha, beta}
	graph := NewTaskGraph(tasks, nil)

	if got := graph.CanonicalDependencies(owner.ID); !slices.Equal(got, []string{beta.ID}) {
		t.Fatalf("canonical dependencies = %v", got)
	}
	wantProjected := []string{beta.ID, gamma.ID}
	slices.Sort(wantProjected)
	if got := graph.Prerequisites(owner.ID); !slices.Equal(got, wantProjected) {
		t.Fatalf("projected prerequisites = %v, want %v", got, wantProjected)
	}

	declarations := mustSourceDeclarations(t, graph)
	assertSourceDeclaration(t, declarations, owner, TaskDependencyDependsOn, beta.ID, 0, DependencyEdge{From: beta.ID, To: owner.ID})
	assertSourceDeclaration(t, declarations, owner, TaskDependencyDependsOn, beta.ID, 1, DependencyEdge{From: beta.ID, To: owner.ID})
	assertSourceDeclaration(t, declarations, owner, TaskDependencyBlockedBy, gamma.Slug, 0, DependencyEdge{From: gamma.ID, To: owner.ID})
	assertSourceDeclaration(t, declarations, owner, TaskDependencyDependencies, beta.ID, 0, DependencyEdge{From: beta.ID, To: owner.ID})
	assertSourceDeclaration(t, declarations, owner, TaskDependencyBlocks, alpha.Slug, 0, DependencyEdge{From: owner.ID, To: alpha.ID})
	reversed := cloneTasks(tasks)
	slices.Reverse(reversed)
	if got := mustSourceDeclarations(t, NewTaskGraph(reversed, nil)); !reflect.DeepEqual(got, declarations) {
		t.Fatalf("adapter record order changed declarations:\nfirst=%+v\nsecond=%+v", declarations, got)
	}

	records := mustSourceRecords(t, graph)
	ownerRecord := sourceRecordFor(t, records, owner.Path)
	ownerRecord.Source.TaskID = "mutated"
	ownerRecord.Fields[0].Values[0] = "mutated"
	if fresh := sourceRecordFor(t, mustSourceRecords(t, graph), owner.Path); fresh.Source.TaskID != owner.ID || fresh.Fields[0].Values[0] != beta.ID {
		t.Fatalf("source query leaked mutable storage: %+v", fresh)
	}
	tasks[0].DependsOn[0] = "input-mutated"
	if fresh := sourceRecordFor(t, mustSourceRecords(t, graph), owner.Path); fresh.Fields[0].Values[0] != beta.ID {
		t.Fatalf("constructor retained input alias: %+v", fresh)
	}
}

func TestTaskGraphSourceDeclarationsOnlyNameEdgesInTheSemanticProjection(t *testing.T) {
	target := graphRecord("projected-edge-target", domain.StatusCompleted)
	danglingID := testutil.TaskID("projected-edge-dangling")
	unreadableID := testutil.TaskID("projected-edge-unreadable")
	owner := graphRecord("projected-edge-owner", domain.StatusReadyToStart,
		target.ID, danglingID, unreadableID, "invalid-token")
	owner.LegacyBlockedBy = []string{"missing-legacy-target"}
	owner.LegacyDependencyFields = []string{"blocked_by"}

	representative := graphRecord("projected-edge-duplicate-a", domain.StatusCompleted, target.ID)
	shadow := graphRecord("projected-edge-duplicate-b", domain.StatusCompleted, target.ID)
	shadow.ID = representative.ID
	shadow.FilenameID = representative.ID
	representative.Path = "tasks/a-representative.md"
	shadow.Path = "tasks/b-shadow.md"

	graph := NewTaskGraphRead(TaskGraphRead{
		Tasks: []domain.Task{shadow, owner, target, representative},
		Problems: []TaskGraphLoadProblem{{
			TaskID: unreadableID, TaskSlug: "projected-edge-unreadable",
			Message: "remote record is unreadable", SourceVersion: "unreadable-revision",
		}},
	})
	declarations := mustSourceDeclarations(t, graph)
	assertSourceDeclaration(t, declarations, owner, TaskDependencyDependsOn, target.ID, 0,
		DependencyEdge{From: target.ID, To: owner.ID})
	assertSourceDeclaration(t, declarations, representative, TaskDependencyDependsOn, target.ID, 0,
		DependencyEdge{From: target.ID, To: representative.ID})
	assertSourceDeclarationWithoutEdge(t, declarations, owner, TaskDependencyDependsOn, danglingID)
	assertSourceDeclarationWithoutEdge(t, declarations, owner, TaskDependencyDependsOn, unreadableID)
	assertSourceDeclarationWithoutEdge(t, declarations, owner, TaskDependencyDependsOn, "invalid-token")
	assertSourceDeclarationWithoutEdge(t, declarations, owner, TaskDependencyBlockedBy, "missing-legacy-target")
	assertSourceDeclarationWithoutEdge(t, declarations, shadow, TaskDependencyDependsOn, target.ID)
	for _, declaration := range declarations {
		if declaration.HasProjectedEdge && !slices.Contains(graph.outgoing[declaration.ProjectedEdge.From], declaration.ProjectedEdge.To) {
			t.Fatalf("source declaration claims edge absent from semantic graph: %+v outgoing=%+v", declaration, graph.outgoing)
		}
	}
}

func TestTaskGraphLegacyShadowDeclarationsDoNotEnterSemanticProjection(t *testing.T) {
	tests := []struct {
		field     TaskDependencyField
		configure func(*domain.Task, domain.Task)
		assert    func(*testing.T, *TaskGraph, domain.Task, domain.Task)
	}{
		{
			field: TaskDependencyBlockedBy,
			configure: func(shadow *domain.Task, peer domain.Task) {
				shadow.LegacyBlockedBy = []string{peer.ID}
			},
			assert: assertLegacyShadowDoesNotGateOwner,
		},
		{
			field: TaskDependencyDependencies,
			configure: func(shadow *domain.Task, peer domain.Task) {
				shadow.LegacyDependencies = []string{peer.ID}
			},
			assert: assertLegacyShadowDoesNotGateOwner,
		},
		{
			field: TaskDependencyBlocks,
			configure: func(shadow *domain.Task, peer domain.Task) {
				shadow.LegacyBlocks = []string{peer.ID}
			},
			assert: func(t *testing.T, graph *TaskGraph, representative, peer domain.Task) {
				t.Helper()
				if got := graph.Prerequisites(peer.ID); slices.Contains(got, representative.ID) {
					t.Fatalf("shadow blocks declaration entered peer prerequisites: %v", got)
				}
				if got := graph.State(peer.ID); got.Gate != GateClear {
					t.Fatalf("shadow blocks declaration changed peer gate: %+v", got)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(string(test.field), func(t *testing.T) {
			peer := graphRecord("legacy-shadow-peer", domain.StatusReadyToStart)
			representative := graphRecord("legacy-shadow-owner-a", domain.StatusCompleted)
			shadow := graphRecord("legacy-shadow-owner-b", domain.StatusCompleted)
			shadow.ID = representative.ID
			shadow.FilenameID = representative.ID
			representative.Path = "tasks/a-legacy-representative.md"
			shadow.Path = "tasks/b-legacy-shadow.md"
			test.configure(&shadow, peer)
			shadow.LegacyDependencyFields = []string{string(test.field)}

			graph := NewTaskGraph([]domain.Task{shadow, peer, representative}, nil)
			test.assert(t, graph, representative, peer)
			foundDiagnostic := false
			for _, diagnostic := range graph.LegacyDiagnostics() {
				if diagnostic.TaskPath == shadow.Path && diagnostic.Field == string(test.field) {
					foundDiagnostic = true
				}
			}
			if !foundDiagnostic {
				t.Fatalf("shadow legacy source evidence disappeared: %+v", graph.LegacyDiagnostics())
			}
			assertSourceDeclarationWithoutEdge(t, mustSourceDeclarations(t, graph), shadow, test.field, peer.ID)
		})
	}
}

func assertLegacyShadowDoesNotGateOwner(t *testing.T, graph *TaskGraph, representative, peer domain.Task) {
	t.Helper()
	if got := graph.Prerequisites(representative.ID); slices.Contains(got, peer.ID) {
		t.Fatalf("shadow prerequisite declaration entered representative prerequisites: %v", got)
	}
}

func TestTaskGraphSourceSimulationPreservesRawInvalidAndDanglingIntent(t *testing.T) {
	prerequisite := graphRecord("source-valid-prerequisite", domain.StatusCompleted)
	owner := graphRecord("source-invalid-owner", domain.StatusReadyToStart)
	missingID := testutil.TaskID("source-missing-prerequisite")
	owner.DependsOn = []string{"human-authored-slug", missingID, prerequisite.ID}
	graph := NewTaskGraph([]domain.Task{owner, prerequisite}, nil)
	source := sourceRefForTask(owner)

	if got := graph.CanonicalDependencies(owner.ID); !slices.Equal(got, sortedUnique([]string{missingID, prerequisite.ID})) {
		t.Fatalf("canonical stable-ID set = %v", got)
	}
	if got := graph.Prerequisites(owner.ID); !slices.Equal(got, sortedUnique(owner.DependsOn)) {
		t.Fatalf("fail-closed behavior prerequisites = %v, want raw canonical values %v", got, sortedUnique(owner.DependsOn))
	}
	assertRawValues(t, graph, owner.Path, TaskDependencyDependsOn, owner.DependsOn)

	simulated, err := graph.SimulateSourceEdits([]TaskGraphSourceEdit{
		{Action: TaskGraphSourceDropDeclaration, Source: source, Field: TaskDependencyDependsOn, Value: "human-authored-slug"},
		{Action: TaskGraphSourceDropDeclaration, Source: source, Field: TaskDependencyDependsOn, Value: missingID},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertRawValues(t, simulated, owner.Path, TaskDependencyDependsOn, []string{prerequisite.ID})
	assertRawValues(t, graph, owner.Path, TaskDependencyDependsOn, owner.DependsOn)
	if simulated.Health() != GraphHealthy {
		t.Fatalf("simulated health = %s problems=%+v", simulated.Health(), simulated.Problems())
	}
}

func TestTaskGraphSourceSimulationPreservesUntouchedInvalidLiteralsVerbatim(t *testing.T) {
	target := graphRecord("source-literal-target", domain.StatusCompleted)
	owner := graphRecord("source-literal-owner", domain.StatusReadyToStart,
		target.ID, "human-authored-slug", "  spaced literal  ", "")
	owner.LegacyBlockedBy = []string{target.Slug, " legacy literal "}
	owner.LegacyDependencyFields = []string{"blocked_by"}
	graph := NewTaskGraph([]domain.Task{owner, target}, nil)

	simulated, err := graph.SimulateSourceEdits([]TaskGraphSourceEdit{
		{Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(owner), Field: TaskDependencyDependsOn, Value: ""},
		{Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(owner), Field: TaskDependencyBlockedBy, Value: target.Slug},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertRawValues(t, simulated, owner.Path, TaskDependencyDependsOn,
		[]string{target.ID, "human-authored-slug", "  spaced literal  "})
	assertRawValues(t, simulated, owner.Path, TaskDependencyBlockedBy, []string{" legacy literal "})
}

func TestTaskGraphSourceDedupeIsRetryStableAndExactDropNamesAnOccurrence(t *testing.T) {
	alpha := graphRecord("source-dedupe-alpha", domain.StatusCompleted)
	beta := graphRecord("source-dedupe-beta", domain.StatusCompleted)
	owner := graphRecord("source-dedupe-owner", domain.StatusReadyToStart, alpha.ID, alpha.ID, alpha.ID, beta.ID)
	graph := NewTaskGraph([]domain.Task{owner, beta, alpha}, nil)
	source := sourceRefForTask(owner)

	dropped, err := graph.SimulateSourceEdits([]TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: source, Field: TaskDependencyDependsOn,
		Value: alpha.ID, Occurrence: 1,
	}})
	if err != nil {
		t.Fatal(err)
	}
	assertRawValues(t, dropped, owner.Path, TaskDependencyDependsOn, []string{alpha.ID, alpha.ID, beta.ID})

	dedupe := TaskGraphSourceEdit{Action: TaskGraphSourceDedupe, Source: source, Field: TaskDependencyDependsOn, Value: alpha.ID}
	deduped, err := graph.SimulateSourceEdits([]TaskGraphSourceEdit{dedupe})
	if err != nil {
		t.Fatal(err)
	}
	assertRawValues(t, deduped, owner.Path, TaskDependencyDependsOn, []string{alpha.ID, beta.ID})
	retried, err := deduped.SimulateSourceEdits([]TaskGraphSourceEdit{dedupe})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mustSourceRecords(t, retried), mustSourceRecords(t, deduped)) {
		t.Fatalf("dedupe retry changed source:\nfirst=%+v\nretry=%+v", mustSourceRecords(t, deduped), mustSourceRecords(t, retried))
	}
	combined := []TaskGraphSourceEdit{
		{Action: TaskGraphSourceDropDeclaration, Source: source, Field: TaskDependencyDependsOn, Value: alpha.ID, Occurrence: 1},
		dedupe,
	}
	forward, err := graph.SimulateSourceEdits(combined)
	if err != nil {
		t.Fatal(err)
	}
	slices.Reverse(combined)
	reverse, err := graph.SimulateSourceEdits(combined)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mustSourceRecords(t, forward), mustSourceRecords(t, reverse)) {
		t.Fatalf("edit order changed source:\nforward=%+v\nreverse=%+v", mustSourceRecords(t, forward), mustSourceRecords(t, reverse))
	}
}

func TestTaskGraphSourceSimulationRemovesOnlyNamedLegacyState(t *testing.T) {
	alpha := graphRecord("source-legacy-alpha", domain.StatusCompleted)
	beta := graphRecord("source-legacy-beta", domain.StatusCompleted)
	gamma := graphRecord("source-legacy-gamma", domain.StatusCompleted)
	delta := graphRecord("source-legacy-delta", domain.StatusCompleted)
	owner := graphRecord("source-legacy-owner", domain.StatusReadyToStart)
	owner.LegacyBlockedBy = []string{alpha.ID, beta.ID}
	owner.LegacyDependencies = []string{gamma.ID}
	owner.LegacyBlocks = []string{delta.ID}
	owner.LegacyDependencyFields = []string{"blocked_by", "dependencies", "blocks"}
	graph := NewTaskGraph([]domain.Task{owner, delta, gamma, beta, alpha}, nil)

	simulated, err := graph.SimulateSourceEdits([]TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(owner),
		Field: TaskDependencyBlockedBy, Value: beta.ID,
	}})
	if err != nil {
		t.Fatal(err)
	}
	assertRawValues(t, simulated, owner.Path, TaskDependencyBlockedBy, []string{alpha.ID})
	assertRawValues(t, simulated, owner.Path, TaskDependencyDependencies, []string{gamma.ID})
	assertRawValues(t, simulated, owner.Path, TaskDependencyBlocks, []string{delta.ID})

	empty := graphRecord("source-empty-legacy", domain.StatusReadyToStart)
	empty.LegacyDependencyFields = []string{"blocked_by", "dependencies", "blocks"}
	emptyGraph := NewTaskGraph([]domain.Task{empty}, nil)
	cleaned, err := emptyGraph.SimulateSourceEdits([]TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropEmptyField, Source: sourceRefForTask(empty), Field: TaskDependencyBlockedBy,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if fields := sourceRecordFor(t, mustSourceRecords(t, cleaned), empty.Path).Fields; len(fields) != 2 ||
		fields[0].Field != TaskDependencyDependencies || fields[1].Field != TaskDependencyBlocks {
		t.Fatalf("empty legacy cleanup changed unrelated field presence: %+v", fields)
	} else if fields[0].Values == nil || fields[1].Values == nil {
		t.Fatalf("present-but-empty legacy fields must expose non-nil values: %+v", fields)
	}
	if _, err := graph.SimulateSourceEdits([]TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropEmptyField, Source: sourceRefForTask(owner), Field: TaskDependencyBlockedBy,
	}}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("non-empty field removal error = %v, want validation", err)
	}
	if _, err := emptyGraph.SimulateSourceEdits([]TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropEmptyField, Source: sourceRefForTask(empty), Field: TaskDependencyDependsOn,
	}}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("canonical empty-field removal error = %v, want validation", err)
	}
}

func TestTaskGraphSourceSimulationSeparatesDeclarationAndEmptyLegacyFieldRemoval(t *testing.T) {
	prerequisite := graphRecord("source-last-legacy-prerequisite", domain.StatusCompleted)
	owner := graphRecord("source-last-legacy-owner", domain.StatusReadyToStart)
	owner.LegacyBlockedBy = []string{prerequisite.ID}
	owner.LegacyDependencyFields = []string{"blocked_by"}
	graph := NewTaskGraph([]domain.Task{owner, prerequisite}, nil)
	drop := TaskGraphSourceEdit{
		Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(owner),
		Field: TaskDependencyBlockedBy, Value: prerequisite.ID,
	}
	dropEmpty := TaskGraphSourceEdit{
		Action: TaskGraphSourceDropEmptyField, Source: sourceRefForTask(owner),
		Field: TaskDependencyBlockedBy,
	}

	valuesRemoved, err := graph.SimulateSourceEdits([]TaskGraphSourceEdit{drop})
	if err != nil {
		t.Fatal(err)
	}
	field := sourceRecordFor(t, mustSourceRecords(t, valuesRemoved), owner.Path).Fields[0]
	if field.Field != TaskDependencyBlockedBy || field.Values == nil || len(field.Values) != 0 {
		t.Fatalf("declaration drop also removed or obscured legacy field presence: %+v", field)
	}
	sequential, err := valuesRemoved.SimulateSourceEdits([]TaskGraphSourceEdit{dropEmpty})
	if err != nil {
		t.Fatal(err)
	}
	batched, err := graph.SimulateSourceEdits([]TaskGraphSourceEdit{drop, dropEmpty})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mustSourceRecords(t, batched), mustSourceRecords(t, sequential)) {
		t.Fatalf("batched and sequential empty-field cleanup diverged:\nbatched=%+v\nsequential=%+v",
			mustSourceRecords(t, batched), mustSourceRecords(t, sequential))
	}
}

func TestTaskGraphSourceSimulationRetainsDuplicateAndUnreadableRecords(t *testing.T) {
	duplicateA := graphRecord("source-duplicate-a", domain.StatusCompleted)
	duplicateB := graphRecord("source-duplicate-b", domain.StatusCompleted)
	duplicateB.ID = duplicateA.ID
	duplicateB.FilenameID = duplicateA.ID
	unreadableID := testutil.TaskID("source-unreadable")
	owner := graphRecord("source-mixed-owner", domain.StatusReadyToStart, unreadableID, "invalid-human-token")
	duplicateA.DependsOn = []string{owner.ID}
	duplicateB.DependsOn = []string{owner.ID}
	owner.SourceVersion = "owner-revision"
	duplicateA.SourceVersion = "duplicate-a-revision"
	duplicateB.SourceVersion = "duplicate-b-revision"
	readProblem := TaskGraphLoadProblem{
		TaskID: unreadableID, TaskSlug: "source-unreadable", Path: "opaque://source-unreadable",
		Message: "remote decode failed", SourceVersion: "unreadable-revision",
	}
	graph := NewTaskGraphRead(TaskGraphRead{Tasks: []domain.Task{duplicateB, owner, duplicateA}, Problems: []TaskGraphLoadProblem{readProblem}})

	simulated, err := graph.SimulateSourceEdits([]TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(owner),
		Field: TaskDependencyDependsOn, Value: "invalid-human-token",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(simulated.sourceTasks) != 3 || len(simulated.loadProblems) != 1 || simulated.loadProblems[0] != readProblem {
		t.Fatalf("simulation lost source evidence: tasks=%+v problems=%+v", simulated.sourceTasks, simulated.loadProblems)
	}
	codes := graphProblemCodes(simulated.Problems())
	if !slices.Contains(codes, ProblemDuplicateTaskID) || !slices.Contains(codes, ProblemUnreadable) {
		t.Fatalf("simulation discarded broken source classes: %+v", simulated.Problems())
	}
	for _, graphProblem := range simulated.Problems() {
		if graphProblem.Code == ProblemMissingDependency && graphProblem.RelatedTaskID == unreadableID {
			t.Fatalf("unreadable prerequisite was misclassified as missing: %+v", graphProblem)
		}
	}
	assertRawValues(t, simulated, owner.Path, TaskDependencyDependsOn, []string{unreadableID})

	if _, err := graph.SimulateSourceEdits([]TaskGraphSourceEdit{{
		Action: TaskGraphSourceDedupe, Source: TaskGraphSourceRef{TaskID: duplicateA.ID},
		Field: TaskDependencyDependsOn, Value: owner.ID,
	}}); !errors.Is(err, domain.ErrAmbiguous) {
		t.Fatalf("duplicate source identity error = %v, want ambiguous", err)
	}

	targeted, err := graph.SimulateSourceEdits([]TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(duplicateB),
		Field: TaskDependencyDependsOn, Value: owner.ID,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(mustSourceRecords(t, targeted)) != len(mustSourceRecords(t, graph)) {
		t.Fatalf("path-qualified edit lost a duplicate-ID record: %+v", mustSourceRecords(t, targeted))
	}
	assertRawValues(t, targeted, duplicateA.Path, TaskDependencyDependsOn, duplicateA.DependsOn)
	assertRawValues(t, targeted, duplicateB.Path, TaskDependencyDependsOn, nil)
}

func TestTaskGraphSourceSimulationRejectsMalformedEditIntent(t *testing.T) {
	task := graphRecord("source-edit-validation", domain.StatusReadyToStart)
	graph := NewTaskGraph([]domain.Task{task}, nil)
	source := sourceRefForTask(task)
	tests := []struct {
		name string
		edit TaskGraphSourceEdit
		want error
	}{
		{name: "missing source identity", edit: TaskGraphSourceEdit{Action: TaskGraphSourceDedupe, Field: TaskDependencyDependsOn}, want: domain.ErrValidation},
		{name: "unknown field", edit: TaskGraphSourceEdit{Action: TaskGraphSourceDedupe, Source: source, Field: "invented"}, want: domain.ErrValidation},
		{name: "unknown action", edit: TaskGraphSourceEdit{Action: "replace", Source: source, Field: TaskDependencyDependsOn}, want: domain.ErrValidation},
		{name: "negative occurrence", edit: TaskGraphSourceEdit{Action: TaskGraphSourceDropDeclaration, Source: source, Field: TaskDependencyDependsOn, Occurrence: -1}, want: domain.ErrValidation},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := graph.SimulateSourceEdits([]TaskGraphSourceEdit{test.edit}); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestTaskGraphSourceSnapshotCASIncludesEveryDuplicateIDRecord(t *testing.T) {
	first := graphRecord("source-cas-first", domain.StatusCompleted)
	second := graphRecord("source-cas-second", domain.StatusCompleted)
	second.ID = first.ID
	second.FilenameID = first.ID
	first.SourceVersion = "first-revision"
	second.SourceVersion = "second-revision"
	baseline := NewTaskGraphRead(TaskGraphRead{Tasks: []domain.Task{first, second}})
	reordered := NewTaskGraphRead(TaskGraphRead{Tasks: []domain.Task{second, first}})
	if !baseline.SameSourceSnapshot(reordered) {
		t.Fatal("reordered duplicate-ID source records changed the snapshot")
	}

	changed := second
	changed.SourceVersion = "changed-shadow-revision"
	if baseline.SameSourceSnapshot(NewTaskGraphRead(TaskGraphRead{Tasks: []domain.Task{first, changed}})) {
		t.Fatal("duplicate-ID shadow edit was absent from whole-snapshot CAS")
	}
	if baseline.SameSourceSnapshot(NewTaskGraphRead(TaskGraphRead{Tasks: []domain.Task{first}})) {
		t.Fatal("duplicate-ID shadow removal was absent from whole-snapshot CAS")
	}
}

func TestTaskGraphSourceQueriesRejectRepresentativeOnlyDerivedGraphs(t *testing.T) {
	first := graphRecord("source-derived-first", domain.StatusCompleted)
	shadow := graphRecord("source-derived-shadow", domain.StatusCompleted)
	shadow.ID = first.ID
	shadow.FilenameID = first.ID
	unreadableID := testutil.TaskID("source-derived-unreadable")
	full := NewTaskGraphRead(TaskGraphRead{
		Tasks: []domain.Task{first, shadow},
		Problems: []TaskGraphLoadProblem{{
			TaskID: unreadableID, Message: "unreadable", SourceVersion: "unreadable-revision",
		}},
	})
	derived := taskGraphAfterDependencyPlan(full, TaskGraphMutationPlan{})
	if _, err := derived.SourceRecords(); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("derived source records error = %v, want validation", err)
	}
	if _, err := derived.SourceDeclarations(); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("derived source declarations error = %v, want validation", err)
	}
	if _, err := derived.SimulateSourceEdits(nil); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("derived source simulation error = %v, want validation", err)
	}
	if derived.SameSourceSnapshot(derived) {
		t.Fatal("representative-only graph compared as an authoritative source snapshot")
	}
}

func TestTaskGraphSourceDeclarationsProjectEveryLegacyCycleDirection(t *testing.T) {
	tests := []struct {
		name      string
		field     TaskDependencyField
		configure func(owner, peer *domain.Task)
		source    func(owner, peer domain.Task) domain.Task
	}{
		{
			name: "blocked_by", field: TaskDependencyBlockedBy,
			configure: func(_ *domain.Task, peer *domain.Task) {
				peer.LegacyBlockedBy = []string{"cycle-owner"}
				peer.LegacyDependencyFields = []string{"blocked_by"}
			}, source: func(_, peer domain.Task) domain.Task { return peer },
		},
		{
			name: "dependencies", field: TaskDependencyDependencies,
			configure: func(_ *domain.Task, peer *domain.Task) {
				peer.LegacyDependencies = []string{"cycle-owner"}
				peer.LegacyDependencyFields = []string{"dependencies"}
			}, source: func(_, peer domain.Task) domain.Task { return peer },
		},
		{
			name: "blocks", field: TaskDependencyBlocks,
			configure: func(owner, _ *domain.Task) {
				owner.LegacyBlocks = []string{"cycle-peer"}
				owner.LegacyDependencyFields = []string{"blocks"}
			}, source: func(owner, _ domain.Task) domain.Task { return owner },
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := graphRecord("cycle-owner", domain.StatusReadyToStart)
			peer := graphRecord("cycle-peer", domain.StatusReadyToStart)
			owner.DependsOn = []string{peer.ID}
			test.configure(&owner, &peer)
			graph := NewTaskGraph([]domain.Task{peer, owner}, nil)
			if graph.Health() != GraphBroken || !slices.Contains(graphProblemCodes(graph.Problems()), ProblemCycle) {
				t.Fatalf("legacy cycle health=%s problems=%+v", graph.Health(), graph.Problems())
			}
			source := test.source(owner, peer)
			assertSourceDeclaration(t, mustSourceDeclarations(t, graph), source, test.field,
				map[TaskDependencyField]string{
					TaskDependencyBlockedBy: "cycle-owner", TaskDependencyDependencies: "cycle-owner", TaskDependencyBlocks: "cycle-peer",
				}[test.field], 0, DependencyEdge{From: owner.ID, To: peer.ID})
			foundUnsafe := false
			for _, diagnostic := range graph.LegacyDiagnostics() {
				if diagnostic.Field == string(test.field) && len(diagnostic.References) == 1 && diagnostic.References[0].Resolution == LegacyUnsafe {
					foundUnsafe = true
				}
			}
			if !foundUnsafe {
				t.Fatalf("legacy field %s was not attributed unsafe: %+v", test.field, graph.LegacyDiagnostics())
			}
		})
	}
}

func assertSourceDeclaration(t *testing.T, declarations []TaskGraphSourceDeclaration, source domain.Task, field TaskDependencyField, value string, occurrence int, edge DependencyEdge) {
	t.Helper()
	for _, declaration := range declarations {
		if declaration.Source == sourceRefForTask(source) && declaration.Field == field &&
			declaration.Value == value && declaration.Occurrence == occurrence {
			if !declaration.HasProjectedEdge || declaration.ProjectedEdge != edge {
				t.Fatalf("declaration %+v projected edge = %+v/%t, want %+v", declaration, declaration.ProjectedEdge, declaration.HasProjectedEdge, edge)
			}
			return
		}
	}
	t.Fatalf("missing declaration source=%+v field=%s value=%q occurrence=%d in %+v", sourceRefForTask(source), field, value, occurrence, declarations)
}

func assertSourceDeclarationWithoutEdge(t *testing.T, declarations []TaskGraphSourceDeclaration, source domain.Task, field TaskDependencyField, value string) {
	t.Helper()
	for _, declaration := range declarations {
		if declaration.Source == sourceRefForTask(source) && declaration.Field == field && declaration.Value == value {
			if declaration.HasProjectedEdge || declaration.ProjectedEdge != (DependencyEdge{}) {
				t.Fatalf("declaration %+v unexpectedly claims a projected edge", declaration)
			}
			return
		}
	}
	t.Fatalf("missing declaration source=%+v field=%s value=%q in %+v", sourceRefForTask(source), field, value, declarations)
}

func sourceRecordFor(t *testing.T, records []TaskGraphSourceRecord, location string) TaskGraphSourceRecord {
	t.Helper()
	for _, record := range records {
		if record.Source.Location == location {
			return record
		}
	}
	t.Fatalf("missing source record at %q in %+v", location, records)
	return TaskGraphSourceRecord{}
}

func assertRawValues(t *testing.T, graph *TaskGraph, location string, field TaskDependencyField, want []string) {
	t.Helper()
	record := sourceRecordFor(t, mustSourceRecords(t, graph), location)
	for _, item := range record.Fields {
		if item.Field == field {
			if !slices.Equal(item.Values, want) {
				t.Fatalf("%s values = %v, want %v", field, item.Values, want)
			}
			return
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing %s values, want %v in %+v", field, want, record.Fields)
	}
}

func graphProblemCodes(problems []GraphProblem) []GraphProblemCode {
	codes := make([]GraphProblemCode, len(problems))
	for index, problem := range problems {
		codes[index] = problem.Code
	}
	return codes
}

func mustSourceRecords(t *testing.T, graph *TaskGraph) []TaskGraphSourceRecord {
	t.Helper()
	records, err := graph.SourceRecords()
	if err != nil {
		t.Fatal(err)
	}
	return records
}

func mustSourceDeclarations(t *testing.T, graph *TaskGraph) []TaskGraphSourceDeclaration {
	t.Helper()
	declarations, err := graph.SourceDeclarations()
	if err != nil {
		t.Fatal(err)
	}
	return declarations
}
