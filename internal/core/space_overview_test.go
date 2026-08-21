package core

import (
	"errors"
	"reflect"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

type fakeSpaceOverviewStore struct {
	groups  []SpaceGroup
	listErr error
	stores  map[string]SummaryStore
	openErr map[string]error
	opened  []string
}

func (f *fakeSpaceOverviewStore) ListSpaceGroups() ([]SpaceGroup, error) {
	return f.groups, f.listErr
}

func (f *fakeSpaceOverviewStore) OpenPlanningStore(root string) (SummaryStore, error) {
	f.opened = append(f.opened, root)
	if err := f.openErr[root]; err != nil {
		return nil, err
	}
	return f.stores[root], nil
}

func TestSpaceOverview_PrefersHealthyDirectAndCombinesInProgress(t *testing.T) {
	source := &fakeSpaceOverviewStore{
		groups: []SpaceGroup{
			{PlanningID: "plan-a", Entries: []SpaceEntryPoint{
				{ID: "impl-a", Role: SpaceRolePointer, State: SpaceStateOK, Root: "/pointer-a"},
				{ID: "planning-a", Role: SpaceRoleDirect, State: SpaceStateOK, Root: "/direct-a"},
			}},
			{PlanningID: "plan-b", Entries: []SpaceEntryPoint{
				{ID: "broken-direct", Role: SpaceRoleDirect, State: SpaceStateMissing, Root: "/broken-b"},
				{ID: "impl-b", Role: SpaceRolePointer, State: SpaceStateEmpty, Root: "/pointer-b"},
			}},
		},
		stores: map[string]SummaryStore{
			"/direct-a":  &fakeStore{tasks: []domain.Task{{Slug: "a", Status: domain.StatusInProgress}}},
			"/pointer-b": &fakeStore{tasks: []domain.Task{{Slug: "b", Status: domain.StatusInProgress}}},
		},
	}

	overview, err := NewSpaceOverviewService(source).Overview()
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

func TestSpaceOverview_IsolatesBrokenAndUnreadableGroups(t *testing.T) {
	source := &fakeSpaceOverviewStore{
		groups: []SpaceGroup{
			{PlanningID: "gone", Entries: []SpaceEntryPoint{{ID: "gone", State: SpaceStateMissing}}},
			{PlanningID: "raced", Entries: []SpaceEntryPoint{{ID: "raced", State: SpaceStateOK, Root: "/raced"}}},
			{PlanningID: "healthy", Entries: []SpaceEntryPoint{{ID: "healthy", State: SpaceStateOK, Root: "/healthy"}}},
		},
		stores: map[string]SummaryStore{
			"/healthy": &fakeStore{tasks: []domain.Task{{Slug: "working", Status: domain.StatusInProgress}}},
		},
		openErr: map[string]error{"/raced": errors.New("checkout disappeared")},
	}

	overview, err := NewSpaceOverviewService(source).Overview()
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
	_, err := NewSpaceOverviewService(&fakeSpaceOverviewStore{listErr: want}).Overview()
	if !errors.Is(err, want) {
		t.Fatalf("Overview error = %v, want %v", err, want)
	}
}
