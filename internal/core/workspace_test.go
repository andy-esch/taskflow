package core

import (
	"errors"
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

type workspaceCapabilitiesFake struct {
	Store
}

func (workspaceCapabilitiesFake) WatchPaths() []string { return []string{"/plan/tasks"} }

func TestWorkspaceService_OpenAssemblesRuntimeAndPreservesSelection(t *testing.T) {
	capabilities := workspaceCapabilitiesFake{}
	adapter := &workspaceStoreFake{source: WorkspaceSource{
		Checkout: "/checkout", PlanningRoot: "/plan", PlanningID: "planning-id",
		Store: capabilities, Layout: capabilities,
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
	capabilities := workspaceCapabilitiesFake{}
	adapter := &workspaceStoreFake{source: WorkspaceSource{
		Checkout: "/checkout", PlanningRoot: "/other-plan", PlanningID: "replacement-id",
		Store: capabilities, Layout: capabilities,
	}}
	_, err := NewWorkspaceService(adapter).Open(WorkspaceRequest{
		Start: "/checkout", SpaceID: "planning", ExpectedPlanningID: "original-id",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("identity drift error = %v", err)
	}
}
