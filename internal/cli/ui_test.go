package cli

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

type uiLayoutFake struct{}

func (uiLayoutFake) WatchPaths() []string { return []string{"/planning/tasks"} }

func TestUIRefusesNonInteractiveAndDryRunInvocation(t *testing.T) {
	repo := t.TempDir()
	if _, err := config.Init(repo, "", false); err != nil {
		t.Fatal(err)
	}
	_, err := runRootRC(t, "-C", repo, "ui")
	if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "interactive terminal") {
		t.Fatalf("non-interactive ui should fail clearly with validation: %v", err)
	}
	_, err = runRootRC(t, "-C", repo, "ui", "--dry-run")
	if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "dry-run") {
		t.Fatalf("ui dry-run should be rejected explicitly: %v", err)
	}
}

func TestAtlasThemeUsesHomeScopeWithoutRepositoryOverride(t *testing.T) {
	t.Setenv("TSKFLW_THEME", "")
	app := &App{
		Cfg:  &config.Config{Theme: config.ThemeConfig{Name: "neon"}},
		User: &userconfig.Config{Theme: userconfig.ThemeConfig{Name: "catppuccin"}},
	}
	if got := app.atlasTheme().Name; got != "catppuccin" {
		t.Fatalf("atlas theme = %q, want home-scoped catppuccin", got)
	}
}

func TestRuntimeWorkspacePreservesResolvedStartupContext(t *testing.T) {
	svc := core.NewService(nil)
	layout := uiLayoutFake{}
	app := &App{
		Cfg: &config.Config{Dir: "/checkout", Root: "/planning", ID: "planning-id"},
		Svc: svc, Layout: layout, selectedSpace: "registered-entry",
	}
	workspace := app.runtimeWorkspace()
	if workspace.SpaceID != "registered-entry" || workspace.Checkout != "/checkout" ||
		workspace.PlanningRoot != "/planning" || workspace.PlanningID != "planning-id" ||
		workspace.Planning != svc || workspace.Layout != layout {
		t.Fatalf("runtime workspace = %+v", workspace)
	}
}
