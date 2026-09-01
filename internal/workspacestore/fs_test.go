package workspacestore

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestFS_OpenWorkspaceResolvesDirectAndPointerEntries(t *testing.T) {
	planning := t.TempDir()
	initialized, err := config.Init(planning, "planning", false)
	if err != nil {
		t.Fatal(err)
	}
	planningConfig, err := config.Discover(planning)
	if err != nil {
		t.Fatal(err)
	}
	threadID := testutil.TaskID("workspace-thread")
	threadPath := filepath.Join(planningConfig.Root, domain.ThreadsDir, threadID+"-workspace-thread.md")
	testutil.Write(t, threadPath, "malformed but path-resolvable\n")
	pointer := t.TempDir()
	if _, err := config.InitPointer(pointer, planning, false); err != nil {
		t.Fatal(err)
	}
	pointerConfig, err := config.Discover(pointer)
	if err != nil {
		t.Fatal(err)
	}

	service := core.NewWorkspaceService(New())
	direct, err := service.Open(core.WorkspaceRequest{Start: planning, SpaceID: "planning"})
	if err != nil {
		t.Fatal(err)
	}
	fromPointer, err := service.Open(core.WorkspaceRequest{Start: pointer, SpaceID: "implementation"})
	if err != nil {
		t.Fatal(err)
	}
	if direct.PlanningRoot != planningConfig.Root || direct.PlanningID != initialized.PlanningID || direct.Checkout != planningConfig.Dir {
		t.Fatalf("direct workspace = %+v", direct)
	}
	if fromPointer.PlanningRoot != direct.PlanningRoot || fromPointer.PlanningID != direct.PlanningID ||
		fromPointer.Checkout != pointerConfig.Dir || fromPointer.SpaceID != "implementation" {
		t.Fatalf("pointer workspace = %+v, direct = %+v", fromPointer, direct)
	}
	if len(fromPointer.Layout.WatchPaths()) != 5 {
		t.Fatalf("watch paths = %v", fromPointer.Layout.WatchPaths())
	}
	for name, workspace := range map[string]core.Workspace{"direct": direct, "pointer": fromPointer} {
		got, err := workspace.Planning.ThreadPath("workspace-thread")
		if err != nil || got != threadPath {
			t.Fatalf("%s workspace Thread path = %q, %v; want %q", name, got, err, threadPath)
		}
	}
}

func TestFS_OpenWorkspacePreservesDiscoveryFailureCause(t *testing.T) {
	planning := t.TempDir()
	initialized, err := config.Init(planning, "", false)
	if err != nil {
		t.Fatal(err)
	}
	pointer := t.TempDir()
	if _, err := config.InitPointer(pointer, planning, false); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(planning, config.ConfigFile)
	if err := os.WriteFile(configPath, []byte("id = \"different\"\ntaskflow_root = \".\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if initialized.PlanningID == "different" {
		t.Fatal("test requires a changed planning id")
	}

	_, err = core.NewWorkspaceService(New()).Open(core.WorkspaceRequest{Start: pointer})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("identity mismatch error = %v", err)
	}
}

func TestFS_OpenWorkspaceDoesNotFallbackFromMissingOrMalformedEntry(t *testing.T) {
	service := core.NewWorkspaceService(New())
	missing := t.TempDir()
	if workspace, err := service.Open(core.WorkspaceRequest{Start: missing}); err == nil ||
		workspace.Planning != nil || !strings.Contains(err.Error(), "not a taskflow planning repo") {
		t.Fatalf("missing entry = %+v, %v", workspace, err)
	}

	malformed := t.TempDir()
	if err := os.WriteFile(filepath.Join(malformed, config.ConfigFile), []byte("[[broken\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if workspace, err := service.Open(core.WorkspaceRequest{Start: malformed}); err == nil ||
		workspace.Planning != nil || !strings.Contains(err.Error(), "parse") {
		t.Fatalf("malformed entry = %+v, %v", workspace, err)
	}
}
