package core

import (
	"errors"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

type workspaceStoreFake struct {
	source WorkspaceSource
	err    error
	start  string
}

func (f *workspaceStoreFake) OpenWorkspace(start string) (WorkspaceSource, error) {
	f.start = start
	return f.source, f.err
}

type workspaceLayoutFake struct{}

func (workspaceLayoutFake) WatchPaths() []string { return []string{"/plan/tasks"} }

func TestWorkspaceService_OpenAssemblesRuntimeAndPreservesSelection(t *testing.T) {
	store := &fakeStore{tasks: []domain.Task{graphRecord("aggregate-only-member", domain.StatusReadyToStart)}}
	layout := workspaceLayoutFake{}
	member := graphRecord("workspace-thread-member", domain.StatusReadyToStart)
	thread := domain.Thread{
		ID: "6g3q4rtmv4ak", Slug: "workspace-thread", Status: domain.ThreadStatusUnstarted,
		Description: "Workspace split ports", Goal: "Preserve adapter neutrality", Created: "2026-08-31",
		Tasks: []string{member.ID},
	}
	adapter := &workspaceStoreFake{source: WorkspaceSource{
		Checkout: "/checkout", PlanningRoot: "/plan", PlanningID: "planning-id",
		Store: store, TaskGraphs: &taskGraphReadFake{tasks: []domain.Task{member}},
		Threads: &threadReadFake{threads: []domain.Thread{thread}, thread: thread}, Layout: layout,
	}}
	workspace, err := NewWorkspaceService(adapter).Open(WorkspaceRequest{
		Start: "/checkout/nested", SpaceID: "implementation", ExpectedPlanningID: "planning-id",
	})
	if err != nil {
		t.Fatal(err)
	}
	if adapter.start != "/checkout/nested" || workspace.SpaceID != "implementation" ||
		workspace.Checkout != "/checkout" || workspace.PlanningRoot != "/plan" ||
		workspace.PlanningID != "planning-id" || workspace.Planning == nil || workspace.Layout == nil {
		t.Fatalf("workspace = %+v, start = %q", workspace, adapter.start)
	}
	view, _, err := workspace.Planning.ShowThread(thread.ID)
	if err != nil || view.ProjectionHealth != GraphHealthy || len(view.Members) != 1 ||
		view.Members[0].Task.ID != member.ID || view.Members[0].Task.Slug != member.Slug {
		t.Fatalf("workspace Thread view = %+v, err = %v", view, err)
	}
}

func TestWorkspaceService_OpenRejectsFallbackAndPreservesAdapterError(t *testing.T) {
	service := NewWorkspaceService(&workspaceStoreFake{})
	if _, err := service.Open(WorkspaceRequest{}); domain.Classify(err) != domain.ClassValidation {
		t.Fatalf("empty start error = %v", err)
	}

	want := errors.New("identity mismatch")
	adapter := &workspaceStoreFake{err: want}
	if _, err := NewWorkspaceService(adapter).Open(WorkspaceRequest{Start: "/wrong"}); !errors.Is(err, want) {
		t.Fatalf("adapter error = %v, want identity %v", err, want)
	}
}

func TestWorkspaceService_OpenRejectsIncompleteAdapterResult(t *testing.T) {
	_, err := NewWorkspaceService(&workspaceStoreFake{source: WorkspaceSource{
		PlanningRoot: "/plan",
	}}).Open(WorkspaceRequest{Start: "/checkout"})
	if domain.Classify(err) != domain.ClassValidation {
		t.Fatalf("incomplete source error = %v", err)
	}
}

func TestWorkspaceService_OpenRejectsPlanningIdentityDrift(t *testing.T) {
	store := &fakeStore{}
	layout := workspaceLayoutFake{}
	adapter := &workspaceStoreFake{source: WorkspaceSource{
		Checkout: "/checkout", PlanningRoot: "/other-plan", PlanningID: "replacement-id",
		Store: store, Layout: layout,
	}}
	_, err := NewWorkspaceService(adapter).Open(WorkspaceRequest{
		Start: "/checkout", SpaceID: "planning", ExpectedPlanningID: "original-id",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("identity drift error = %v", err)
	}
}

// A primary adapter uses Checkout as its config start and as the identity of the tree it
// has open; an empty one would silently make `:config` edit the launching directory.
func TestWorkspaceService_OpenRejectsAnEmptyCheckout(t *testing.T) {
	store := &fakeStore{}
	layout := workspaceLayoutFake{}
	_, err := NewWorkspaceService(&workspaceStoreFake{source: WorkspaceSource{
		PlanningRoot: "/plan", Store: store, Layout: layout,
	}}).Open(WorkspaceRequest{Start: "/checkout"})
	if domain.Classify(err) != domain.ClassValidation {
		t.Fatalf("empty checkout error = %v", err)
	}
}

func TestWorkspaceService_OpenRejectsTypedNilCapabilities(t *testing.T) {
	var store *fakeStore
	_, err := NewWorkspaceService(&workspaceStoreFake{source: WorkspaceSource{
		Checkout: "/checkout", PlanningRoot: "/plan", Store: store, Layout: workspaceLayoutFake{},
	}}).Open(WorkspaceRequest{Start: "/checkout"})
	if domain.Classify(err) != domain.ClassValidation {
		t.Fatalf("typed-nil store error = %v", err)
	}
}

// A service that cannot open anything is a different problem from a bad path, and the
// caller has to be told which — so the capability check runs first.
func TestWorkspaceService_OpenReportsAnUnavailableOpenerBeforeTheRequest(t *testing.T) {
	_, err := (*WorkspaceService)(nil).Open(WorkspaceRequest{})
	if err == nil || !strings.Contains(err.Error(), "opener is unavailable") {
		t.Fatalf("nil opener error = %v, want the unavailable-opener cause", err)
	}
}
