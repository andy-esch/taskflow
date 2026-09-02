package tui

import (
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
)

type tuiWorkspaceStore struct{}

func (tuiWorkspaceStore) OpenWorkspace(string) (core.WorkspaceSource, error) {
	return core.WorkspaceSource{}, nil
}

func TestWithWorkspaceOpeningInjectsCapabilityWithoutMakingItRequired(t *testing.T) {
	service := core.NewWorkspaceService(tuiWorkspaceStore{})
	configured := New(nil, WithWorkspaceOpening(service))
	if configured.workspaceSvc != service {
		t.Fatal("workspace opening capability was not retained")
	}
	if !configured.sessionScope {
		t.Fatal("workspace switching capability must enable asynchronous session scoping")
	}
	if plain := New(nil); plain.workspaceSvc != nil {
		t.Fatal("single-space embedded model should not require workspace opening")
	}
}
