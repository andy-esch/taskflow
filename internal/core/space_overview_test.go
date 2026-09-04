package core

import (
	"errors"
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
	return TaskGraphReadFromFiles(s.tasks, s.problems), nil
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
	if summary.GraphHealth != GraphBroken || len(summary.InProgress) != 0 {
		t.Fatalf("summary did not use explicit graph snapshot: %+v", summary)
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
	if len(overview.Spaces) != 3 || overview.Spaces[0].LoadError != "no healthy entry point" {
		t.Fatalf("all-broken group was not retained: %+v", overview.Spaces)
	}
	if overview.Spaces[1].LoadError == "" || overview.Spaces[2].Summary == nil {
		t.Fatalf("per-group load failure was not isolated: %+v", overview.Spaces)
	}
	if len(overview.InProgress) != 1 || overview.InProgress[0].Task.Slug != "working" {
		t.Fatalf("healthy groups should still contribute work: %+v", overview.InProgress)
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
