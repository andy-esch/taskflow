package core

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

type fakeSpaceOverviewStore struct {
	stores  map[string]PlanningSummarySource
	openErr map[string]error
	opened  []string
}

type splitSummaryStore struct {
	*fakeStore
	read TaskGraphRead
}

func (s *splitSummaryStore) ReadTaskGraph() (TaskGraphRead, error) { return s.read, nil }

type planningSummaryFake struct{ *fakeStore }

func (s *planningSummaryFake) ReadTaskGraph() (TaskGraphRead, error) {
	read := TaskGraphReadFromFiles(s.tasks, s.problems)
	read.GuardedRecords = nil           // use the fake's explicitly assigned portable IDs
	read.Records = loadedTasks(s.tasks) // the fake store assigns source IDs even without frontmatter IDs
	return read, nil
}

func (f *fakeSpaceOverviewStore) OpenPlanningStore(root string) (PlanningSummarySource, error) {
	f.opened = append(f.opened, root)
	if err := f.openErr[root]; err != nil {
		return nil, err
	}
	return f.stores[root], nil
}

func TestSpaceOverview_PrefersHealthyDirectAndCombinesInProgress(t *testing.T) {
	registry := NewSpaceRegistryService(&fakeSpaceRegistryStore{
		entries: flattenSpaceGroups([]SpaceGroup{
			{PlanningID: "plan-a", Entries: []SpaceEntryPoint{
				{ID: "impl-a", PlanningID: "plan-a", Role: SpaceRolePointer, State: SpaceStateOK, Root: "/pointer-a"},
				{ID: "planning-a", PlanningID: "plan-a", Role: SpaceRoleDirect, State: SpaceStateOK, Root: "/direct-a"},
			}},
			{PlanningID: "plan-b", Entries: []SpaceEntryPoint{
				{ID: "broken-direct", PlanningID: "plan-b", Role: SpaceRoleDirect, State: SpaceStateMissing, Root: "/broken-b"},
				{ID: "impl-b", PlanningID: "plan-b", Role: SpaceRolePointer, State: SpaceStateEmpty, Root: "/pointer-b"},
			}},
		}),
	})
	source := &fakeSpaceOverviewStore{
		stores: map[string]PlanningSummarySource{
			"/direct-a":  &planningSummaryFake{fakeStore: &fakeStore{tasks: []domain.Task{{Slug: "a", Status: domain.StatusInProgress}}}},
			"/pointer-b": &planningSummaryFake{fakeStore: &fakeStore{tasks: []domain.Task{{Slug: "b", Status: domain.StatusInProgress}}}},
		},
	}

	overview, err := NewSpaceOverviewService(registry, source).Overview()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"/direct-a", "/pointer-b"}; !reflect.DeepEqual(source.opened, want) {
		t.Fatalf("opened roots = %v, want %v", source.opened, want)
	}
	if len(overview.Spaces) != 2 || overview.Spaces[0].ID != "planning-a" || overview.Spaces[1].ID != "impl-b" {
		t.Fatalf("selected spaces = %+v", overview.Spaces)
	}
	if len(overview.InProgress) != 2 || overview.InProgress[0].SpaceID != "planning-a" || overview.InProgress[1].SpaceID != "impl-b" {
		t.Fatalf("combined working set = %+v", overview.InProgress)
	}
}

func TestSpaceOverviewUsesPlanningStoresGraphSnapshot(t *testing.T) {
	member := graphRecord("space-summary-member", domain.StatusReadyToStart)
	member.LegacyBlockedBy = []string{"gone"}
	root := "/planning"
	registry := NewSpaceRegistryService(&fakeSpaceRegistryStore{entries: []SpaceEntryPoint{{
		ID: "planning", PlanningID: "planning", Role: SpaceRoleDirect, State: SpaceStateOK, Root: root,
	}}})
	source := &fakeSpaceOverviewStore{stores: map[string]PlanningSummarySource{
		root: &splitSummaryStore{
			fakeStore: &fakeStore{tasks: []domain.Task{{Slug: "wrong", Status: domain.StatusInProgress}}},
			read:      TaskGraphRead{Tasks: []domain.Task{member}},
		},
	}}

	overview, err := NewSpaceOverviewService(registry, source).Overview()
	if err != nil {
		t.Fatal(err)
	}
	if len(overview.Spaces) != 1 || overview.Spaces[0].Summary == nil {
		t.Fatalf("overview = %+v", overview)
	}
	summary := overview.Spaces[0].Summary
	if summary.GraphHealth != GraphBroken || len(summary.InProgressRecords) != 0 {
		t.Fatalf("summary did not use explicit graph snapshot: %+v", summary)
	}
}

func TestSpaceOverviewWorkRetainsPortableTaskSource(t *testing.T) {
	const sourceID = "portable-task-source"
	task := domain.Task{ID: "declared-id", Slug: "work", Status: domain.StatusInProgress}
	root := "/planning"
	registry := NewSpaceRegistryService(&fakeSpaceRegistryStore{entries: []SpaceEntryPoint{{
		ID: "planning", PlanningID: "planning", Role: SpaceRoleDirect, State: SpaceStateOK, Root: root,
	}}})
	source := &fakeSpaceOverviewStore{stores: map[string]PlanningSummarySource{
		root: &splitSummaryStore{fakeStore: &fakeStore{}, read: TaskGraphRead{Records: []LoadedRecord[domain.Task]{
			{Value: task, Source: RecordSource{ID: sourceID}},
		}}},
	}}
	overview, err := NewSpaceOverviewService(registry, source).Overview()
	if err != nil || len(overview.InProgress) != 1 {
		t.Fatalf("overview=%+v err=%v", overview, err)
	}
	if got := overview.InProgress[0].Source.ID; got != sourceID {
		t.Fatalf("cross-space work source = %q, want %q", got, sourceID)
	}
	if got := overview.Spaces[0].Summary.InProgressRecords[0].Source.ID; got != sourceID {
		t.Fatalf("summary work source = %q, want %q", got, sourceID)
	}
}

func TestSpaceOverviewPreservesPathlessTaskLoadProblemIdentity(t *testing.T) {
	root := "/planning"
	registry := NewSpaceRegistryService(&fakeSpaceRegistryStore{entries: []SpaceEntryPoint{{
		ID: "planning", PlanningID: "planning", Role: SpaceRoleDirect, State: SpaceStateOK, Root: root,
	}}})
	source := &fakeSpaceOverviewStore{stores: map[string]PlanningSummarySource{
		root: &splitSummaryStore{fakeStore: &fakeStore{}, read: TaskGraphRead{Problems: []TaskGraphLoadProblem{{
			TaskID: "6gpathless01", TaskSlug: "broken-task",
			Location: "db://planning/tasks/1", Message: "decode failed",
		}}}},
	}}

	overview, err := NewSpaceOverviewService(registry, source).Overview()
	if err != nil {
		t.Fatal(err)
	}
	if len(overview.Spaces) != 1 || overview.Spaces[0].Summary == nil ||
		len(overview.Spaces[0].Summary.Problems) != 1 {
		t.Fatalf("overview = %+v", overview)
	}
	problem := overview.Spaces[0].Summary.Problems[0]
	if problem.EntityKind != EntityTask || problem.EntityID != "6gpathless01" ||
		problem.EntitySlug != "broken-task" || problem.Location != "db://planning/tasks/1" ||
		problem.LocalPath != "" {
		t.Fatalf("cross-space diagnostic = %+v", problem)
	}
}

func TestSpaceOverview_IsolatesBrokenAndUnreadableGroups(t *testing.T) {
	registry := NewSpaceRegistryService(&fakeSpaceRegistryStore{entries: []SpaceEntryPoint{
		{ID: "gone", PlanningID: "gone", State: SpaceStateMissing},
		{ID: "raced", PlanningID: "raced", State: SpaceStateOK, Root: "/raced"},
		{ID: "healthy", PlanningID: "healthy", State: SpaceStateOK, Root: "/healthy"},
	}})
	source := &fakeSpaceOverviewStore{
		stores: map[string]PlanningSummarySource{
			"/healthy": &planningSummaryFake{fakeStore: &fakeStore{tasks: []domain.Task{{Slug: "working", Status: domain.StatusInProgress}}}},
		},
		openErr: map[string]error{"/raced": errors.New("checkout disappeared")},
	}

	overview, err := NewSpaceOverviewService(registry, source).Overview()
	if err != nil {
		t.Fatal(err)
	}
	if len(overview.Spaces) != 3 || overview.Spaces[0].Failure == nil ||
		overview.Spaces[0].Failure.Message != "no healthy entry point" {
		t.Fatalf("all-broken group was not retained: %+v", overview.Spaces)
	}
	if overview.Spaces[1].Failure == nil || overview.Spaces[2].Summary == nil {
		t.Fatalf("per-group load failure was not isolated: %+v", overview.Spaces)
	}
	if len(overview.InProgress) != 1 || overview.InProgress[0].Task.Slug != "working" {
		t.Fatalf("healthy groups should still contribute work: %+v", overview.InProgress)
	}
}

func TestSpaceOverviewRetainsStructuredContentionAndRetriesOnlyFailedGroups(t *testing.T) {
	entries := []SpaceEntryPoint{
		{ID: "alpha", PlanningID: "plan-alpha", Role: SpaceRoleDirect, State: SpaceStateOK, Root: "/alpha"},
		{ID: "beta", PlanningID: "plan-beta", Role: SpaceRoleDirect, State: SpaceStateOK, Root: "/beta"},
	}
	registry := NewSpaceRegistryService(&fakeSpaceRegistryStore{entries: entries})
	alpha := &planningSummaryFake{fakeStore: &fakeStore{tasks: []domain.Task{{Slug: "alpha-old", Status: domain.StatusInProgress}}}}
	beta := &planningSummaryFake{fakeStore: &fakeStore{tasks: []domain.Task{{Slug: "beta-old", Status: domain.StatusInProgress}}}}
	source := &fakeSpaceOverviewStore{
		stores:  map[string]PlanningSummarySource{"/alpha": alpha, "/beta": beta},
		openErr: map[string]error{},
	}
	service := NewSpaceOverviewService(registry, source)

	previous, err := service.Overview()
	if err != nil {
		t.Fatal(err)
	}
	alpha.tasks = []domain.Task{{Slug: "alpha-new", Status: domain.StatusInProgress}}
	source.openErr["/beta"] = fmt.Errorf("planner active: %w", domain.ErrConflict)
	current, err := service.Overview()
	if err != nil {
		t.Fatal(err)
	}
	if current.Spaces[1].Failure == nil || current.Spaces[1].Failure.Class != domain.ClassConflict ||
		current.Spaces[1].Summary != nil {
		t.Fatalf("contended summary = %+v", current.Spaces[1])
	}

	merged := RetainContendedSpaceSummaries(previous, current)
	if merged.Spaces[0].Summary == nil || merged.Spaces[0].Summary.InProgressRecords[0].Value.Slug != "alpha-new" {
		t.Fatalf("successful space did not advance: %+v", merged.Spaces[0])
	}
	if !merged.Spaces[1].Stale || merged.Spaces[1].Summary == nil ||
		merged.Spaces[1].Summary.InProgressRecords[0].Value.Slug != "beta-old" {
		t.Fatalf("contended space did not retain its coherent summary: %+v", merged.Spaces[1])
	}
	if len(merged.InProgress) != 2 || !merged.InProgress[1].Stale {
		t.Fatalf("combined work did not retain stale provenance: %+v", merged.InProgress)
	}

	delete(source.openErr, "/beta")
	beta.tasks = []domain.Task{{Slug: "beta-new", Status: domain.StatusInProgress}}
	source.opened = nil
	refresh := service.RetryContended(merged)
	if want := []string{"/beta"}; !reflect.DeepEqual(source.opened, want) {
		t.Fatalf("retry opened %v, want only %v", source.opened, want)
	}
	reconciled := ApplySpaceOverviewRefresh(merged, refresh)
	if reconciled.Spaces[1].Stale || reconciled.Spaces[1].Failure != nil ||
		reconciled.Spaces[1].Summary == nil || reconciled.Spaces[1].Summary.InProgressRecords[0].Value.Slug != "beta-new" {
		t.Fatalf("successful retry did not replace stale data: %+v", reconciled.Spaces[1])
	}
}

func TestSpaceOverviewRepeatedContentionIsBoundedAndDurableFailureReplacesStaleData(t *testing.T) {
	entry := SpaceEntryPoint{ID: "planning", PlanningID: "plan", Role: SpaceRoleDirect, State: SpaceStateOK, Root: "/planning"}
	registry := NewSpaceRegistryService(&fakeSpaceRegistryStore{entries: []SpaceEntryPoint{entry}})
	store := &planningSummaryFake{fakeStore: &fakeStore{tasks: []domain.Task{{Slug: "working", Status: domain.StatusInProgress}}}}
	source := &fakeSpaceOverviewStore{
		stores: map[string]PlanningSummarySource{"/planning": store}, openErr: map[string]error{},
	}
	service := NewSpaceOverviewService(registry, source)
	previous, err := service.Overview()
	if err != nil {
		t.Fatal(err)
	}

	source.openErr["/planning"] = fmt.Errorf("still busy: %w", domain.ErrConflict)
	contended, err := service.Overview()
	if err != nil {
		t.Fatal(err)
	}
	merged := RetainContendedSpaceSummaries(previous, contended)
	source.opened = nil
	repeated := service.RetryContended(merged)
	if len(repeated.Spaces) != 1 || len(source.opened) != 1 {
		t.Fatalf("bounded retry = %+v, opened=%v", repeated, source.opened)
	}
	merged = ApplySpaceOverviewRefresh(merged, repeated)
	if !merged.Spaces[0].Stale || merged.Spaces[0].Summary == nil || !merged.Spaces[0].Contended() {
		t.Fatalf("repeated contention lost retained evidence: %+v", merged.Spaces[0])
	}

	source.openErr["/planning"] = errors.New("durable read failure")
	durable, err := service.Overview()
	if err != nil {
		t.Fatal(err)
	}
	merged = RetainContendedSpaceSummaries(merged, durable)
	if merged.Spaces[0].Stale || merged.Spaces[0].Summary != nil || merged.Spaces[0].Failure == nil ||
		merged.Spaces[0].Failure.Class != domain.ClassUnknown {
		t.Fatalf("durable failure incorrectly retained stale summary: %+v", merged.Spaces[0])
	}
}

func TestSpaceOverviewConsecutiveFirstLoadContentionDoesNotInventSummary(t *testing.T) {
	contended := SpaceSummary{
		ID: "planning", PlanningID: "plan",
		Failure: &SpaceLoadFailure{Class: domain.ClassConflict, Message: "still busy"},
	}

	merged := RetainContendedSpaceSummaries(
		SpaceOverview{Spaces: []SpaceSummary{contended}},
		SpaceOverview{Spaces: []SpaceSummary{contended}},
	)
	if len(merged.Spaces) != 1 || merged.Spaces[0].Summary != nil || merged.Spaces[0].Stale ||
		!merged.Spaces[0].Contended() {
		t.Fatalf("consecutive first-load contention invented retained data: %+v", merged.Spaces)
	}
}

func TestSpaceOverviewRetainedSummaryOwnsMutableSnapshotData(t *testing.T) {
	previousSummary := Summary{
		Counts: []StatusCount{{Status: domain.StatusInProgress, Count: 1}},
		InProgressRecords: []LoadedRecord[domain.Task]{{
			Value:  domain.Task{Slug: "working", Tags: []string{"original"}, DependsOn: []string{"6g0000000001"}},
			Source: RecordSource{ID: "task-original"},
		}},
		TaskSourceIDs:  []string{"task-original"},
		Epics:          []EpicSummary{{Epic: domain.Epic{ID: "01-domain", Tags: []string{"original"}}}},
		OpenAudits:     []LoadedRecord[domain.Audit]{{Value: domain.Audit{Slug: "review"}, Source: RecordSource{ID: "audit-original"}}},
		AuditSourceIDs: []string{"audit-original"},
		Findings: FindingsRollup{
			ByUrgency: []CountBy{{Key: "soon", Count: 1}},
			Acute:     []AuditFinding{{Finding: domain.Finding{Code: "H1", Title: "original"}}},
		},
		Problems: []LoadProblem{{
			EntityKind: EntityTask, Location: "tasks/broken.md",
			LocalPath: "tasks/broken.md", Message: "original",
		}},
	}
	previous := SpaceOverview{Spaces: []SpaceSummary{{
		ID: "planning", PlanningID: "plan", Summary: &previousSummary,
	}}}
	current := SpaceOverview{Spaces: []SpaceSummary{{
		ID: "planning", PlanningID: "plan",
		Failure: &SpaceLoadFailure{Class: domain.ClassConflict, Message: "busy"},
	}}}

	retained := RetainContendedSpaceSummaries(previous, current).Spaces[0].Summary
	previousSummary.Counts[0].Count = 9
	previousSummary.InProgressRecords[0].Value.Tags[0] = "changed"
	previousSummary.InProgressRecords[0].Value.DependsOn[0] = "changed"
	previousSummary.InProgressRecords[0].Source.ID = "changed"
	previousSummary.TaskSourceIDs[0] = "changed"
	previousSummary.Epics[0].Epic.Tags[0] = "changed"
	previousSummary.OpenAudits[0].Value.Slug = "changed"
	previousSummary.OpenAudits[0].Source.ID = "changed"
	previousSummary.AuditSourceIDs[0] = "changed"
	previousSummary.Findings.ByUrgency[0].Key = "changed"
	previousSummary.Findings.Acute[0].Title = "changed"
	previousSummary.Problems[0].Message = "changed"

	if retained.Counts[0].Count != 1 || retained.InProgressRecords[0].Value.Tags[0] != "original" ||
		retained.InProgressRecords[0].Value.DependsOn[0] != "6g0000000001" ||
		retained.InProgressRecords[0].Source.ID != "task-original" ||
		retained.TaskSourceIDs[0] != "task-original" ||
		retained.Epics[0].Epic.Tags[0] != "original" || retained.OpenAudits[0].Value.Slug != "review" ||
		retained.OpenAudits[0].Source.ID != "audit-original" ||
		retained.AuditSourceIDs[0] != "audit-original" ||
		retained.Findings.ByUrgency[0].Key != "soon" || retained.Findings.Acute[0].Title != "original" ||
		retained.Problems[0].Message != "original" {
		t.Fatalf("retained summary aliases previous mutable data: %+v", retained)
	}
}

func TestSpaceOverviewMarksWorkWithHiddenDuplicateIdentityUnavailable(t *testing.T) {
	record := LoadedRecord[domain.Task]{
		Value:  domain.Task{ID: "declared", Slug: "working", Status: domain.StatusInProgress},
		Source: RecordSource{ID: "shared-source"},
	}
	summary := Summary{InProgressRecords: []LoadedRecord[domain.Task]{record},
		TaskSourceIDs: []string{"shared-source", "shared-source"}}
	space := SpaceSummary{ID: "space", PlanningID: "planning", Summary: &summary}
	overview := spaceOverviewFromSummaries([]SpaceSummary{space})
	if len(overview.InProgress) != 1 || !overview.InProgress[0].IdentityUnavailable {
		t.Fatalf("hidden completed duplicate stayed addressable: %+v", overview.InProgress)
	}
	summary.TaskSourceIDs = []string{"shared-source"}
	overview = spaceOverviewFromSummaries([]SpaceSummary{space})
	if len(overview.InProgress) != 1 || overview.InProgress[0].IdentityUnavailable {
		t.Fatalf("unique source identity became unavailable: %+v", overview.InProgress)
	}
}

func TestSpaceSummaryReconciliationKeyPrefersDurableIdentity(t *testing.T) {
	for name, tc := range map[string]struct {
		space SpaceSummary
		want  string
	}{
		"planning identity": {SpaceSummary{ID: "local", PlanningID: "durable"}, "planning:durable"},
		"legacy entry":      {SpaceSummary{ID: "local", Entries: []SpaceEntryPoint{{ID: "entry"}}}, "entry:entry"},
		"summary label":     {SpaceSummary{ID: "local"}, "summary:local"},
	} {
		t.Run(name, func(t *testing.T) {
			if got := tc.space.ReconciliationKey(); got != tc.want {
				t.Fatalf("ReconciliationKey() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSpaceState_HealthyIsDerivedFromResolvedState(t *testing.T) {
	for _, state := range []SpaceState{SpaceStateOK, SpaceStateEmpty} {
		if !state.Healthy() {
			t.Errorf("%q should be healthy", state)
		}
	}
	for _, state := range []SpaceState{"", SpaceStateMissing, SpaceStateNotARepo, SpaceStateUnreadable, SpaceStateMismatch} {
		if state.Healthy() {
			t.Errorf("%q should be broken", state)
		}
	}
}

func TestSpaceOverview_RegistryFailureIsFatal(t *testing.T) {
	want := errors.New("bad registry")
	registry := NewSpaceRegistryService(&fakeSpaceRegistryStore{listErr: want})
	_, err := NewSpaceOverviewService(registry, &fakeSpaceOverviewStore{}).Overview()
	if !errors.Is(err, want) {
		t.Fatalf("Overview error = %v, want %v", err, want)
	}
}

func flattenSpaceGroups(groups []SpaceGroup) []SpaceEntryPoint {
	var entries []SpaceEntryPoint
	for _, group := range groups {
		entries = append(entries, group.Entries...)
	}
	return entries
}
