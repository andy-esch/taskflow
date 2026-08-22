package core

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

type fakeSpaceRegistryStore struct {
	entries     []SpaceEntryPoint
	listErr     error
	prepared    SpaceRegistration
	prepareErr  error
	preparePath string
	added       SpaceRegistration
	addEntry    SpaceEntryPoint
	addChanged  bool
	addErr      error
	forgotID    string
	forgetEntry SpaceEntryPoint
	forget      bool
	forgetErr   error
	dryRun      bool
}

func (f *fakeSpaceRegistryStore) ListSpaceEntries() ([]SpaceEntryPoint, error) {
	return f.entries, f.listErr
}

func (f *fakeSpaceRegistryStore) PrepareSpace(path string) (SpaceRegistration, error) {
	f.preparePath = path
	return f.prepared, f.prepareErr
}

func (f *fakeSpaceRegistryStore) AddSpace(registration SpaceRegistration, dryRun bool) (SpaceEntryPoint, bool, error) {
	f.added, f.dryRun = registration, dryRun
	return f.addEntry, f.addChanged, f.addErr
}

func (f *fakeSpaceRegistryStore) ForgetSpace(id string, dryRun bool) (SpaceEntryPoint, bool, error) {
	f.forgotID, f.dryRun = id, dryRun
	return f.forgetEntry, f.forget, f.forgetErr
}

func TestSpaceRegistryCatalog_GroupsIdentityRootAndBrokenEntriesInRegistryOrder(t *testing.T) {
	entries := []SpaceEntryPoint{
		{ID: "impl", PlanningID: "plan-a", Root: "/plan"},
		{ID: "idless", Root: "/legacy"},
		{ID: "planning", PlanningID: "plan-a", Root: "/plan"},
		{ID: "idless-pointer", Root: "/legacy"},
		{ID: "gone-a"},
		{ID: "gone-b"},
	}
	catalog, err := NewSpaceRegistryService(&fakeSpaceRegistryStore{entries: entries}).Catalog()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(catalog.Entries, entries) {
		t.Fatalf("catalog entries changed registry order: %+v", catalog.Entries)
	}
	if len(catalog.Groups) != 4 {
		t.Fatalf("groups = %+v, want identity group, root group, and two isolated failures", catalog.Groups)
	}
	if got := []string{catalog.Groups[0].Entries[0].ID, catalog.Groups[0].Entries[1].ID}; !reflect.DeepEqual(got, []string{"impl", "planning"}) {
		t.Errorf("identity group = %v", got)
	}
	if got := []string{catalog.Groups[1].Entries[0].ID, catalog.Groups[1].Entries[1].ID}; !reflect.DeepEqual(got, []string{"idless", "idless-pointer"}) {
		t.Errorf("legacy root group = %v", got)
	}
}

func TestSpaceRegistryResolve_HealthyUnknownDuplicateAndBroken(t *testing.T) {
	tests := []struct {
		name    string
		entries []SpaceEntryPoint
		id      string
		wantID  string
		class   domain.Class
		text    string
	}{
		{name: "healthy", entries: []SpaceEntryPoint{{ID: "good", State: SpaceStateEmpty, Checkout: "/repo"}}, id: "good", wantID: "good"},
		{name: "empty registry", id: "gone", class: domain.ClassNotFound, text: "none are registered"},
		{name: "known labels sorted", entries: []SpaceEntryPoint{{ID: "z"}, {ID: "a"}}, id: "gone", class: domain.ClassNotFound, text: "known: a, z"},
		{name: "duplicate", entries: []SpaceEntryPoint{{ID: "dup"}, {ID: "dup"}}, id: "dup", class: domain.ClassValidation, text: "appears more than once"},
		{name: "missing", entries: []SpaceEntryPoint{{ID: "gone", State: SpaceStateMissing, Detail: "not found", Remedy: "repair"}}, id: "gone", class: domain.ClassNotFound, text: "not found; repair"},
		{name: "mismatch", entries: []SpaceEntryPoint{{ID: "wrong", State: SpaceStateMismatch, Detail: "does not match"}}, id: "wrong", class: domain.ClassConflict, text: "does not match"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSpaceRegistryService(&fakeSpaceRegistryStore{entries: tt.entries}).Resolve(tt.id)
			if tt.class == domain.ClassUnknown {
				if err != nil || got.ID != tt.wantID {
					t.Fatalf("Resolve = %+v, %v", got, err)
				}
				return
			}
			if domain.Classify(err) != tt.class || !strings.Contains(err.Error(), tt.text) {
				t.Fatalf("Resolve error = %v (class %q), want class %q containing %q", err, domain.Classify(err), tt.class, tt.text)
			}
		})
	}
}

func TestSpaceRegistryAdd_PreparesDefaultsValidatesThenMutates(t *testing.T) {
	store := &fakeSpaceRegistryStore{
		prepared: coreRegistration("/repos/My-Plan"),
		addEntry: SpaceEntryPoint{ID: "my-plan"}, addChanged: true,
	}
	mutation, err := NewSpaceRegistryService(store).Add("nested", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if store.preparePath != "nested" || store.added.ID != "my-plan" || !store.dryRun {
		t.Fatalf("store calls = prepare %q, add %+v, dry=%v", store.preparePath, store.added, store.dryRun)
	}
	if !mutation.Changed || !mutation.DryRun || mutation.Entry.ID != "my-plan" {
		t.Fatalf("mutation = %+v", mutation)
	}

	invalid := &fakeSpaceRegistryStore{prepared: coreRegistration("/repo")}
	_, err = NewSpaceRegistryService(invalid).Add("/repo", "bad label", false)
	if domain.Classify(err) != domain.ClassValidation || invalid.added.ID != "" {
		t.Fatalf("invalid label should stop before AddSpace: added=%+v err=%v", invalid.added, err)
	}
}

func TestSpaceRegistryOperations_PreserveAdapterErrorsAndForgetContract(t *testing.T) {
	want := errors.New("disk unavailable")
	if _, err := NewSpaceRegistryService(&fakeSpaceRegistryStore{listErr: want}).Catalog(); !errors.Is(err, want) {
		t.Fatalf("Catalog error = %v", err)
	}
	if _, err := NewSpaceRegistryService(&fakeSpaceRegistryStore{prepareErr: want}).Add(".", "x", false); !errors.Is(err, want) {
		t.Fatalf("Add error = %v", err)
	}
	if _, err := NewSpaceRegistryService(&fakeSpaceRegistryStore{}).Forget("gone", false); domain.Classify(err) != domain.ClassNotFound {
		t.Fatalf("Forget missing error = %v", err)
	}
}

func coreRegistration(path string) SpaceRegistration {
	return SpaceRegistration{Path: path, VerifyID: "plan-id"}
}
