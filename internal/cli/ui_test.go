package cli

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/spacestore"
	"github.com/andy-esch/taskflow/internal/userconfig"
	"github.com/andy-esch/taskflow/internal/workspacestore"
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

// Standing in a planning repo says which space you meant, so `ui` opens THAT repo and
// leaves the atlas one keystroke away — however many spaces are registered.
func TestUIStartupOpensTheRepoYouAreStandingIn(t *testing.T) {
	repo := t.TempDir()
	if _, err := config.Init(repo, "", false); err != nil {
		t.Fatal(err)
	}
	app := newUITestApp(t, repo)
	app.Chdir = repo // mirror `ui -C <repo>`, so the config start is deterministic here
	workspace, start, landOnAtlas, err := app.uiStartup()
	if err != nil {
		t.Fatalf("uiStartup() error = %v", err)
	}
	if landOnAtlas {
		t.Error("in-repo launch must not land on the atlas")
	}
	if workspace.PlanningRoot != app.Cfg.Root || workspace.Checkout != app.Cfg.Dir || start != repo {
		t.Fatalf("workspace = %+v, start = %q, want the discovered repo", workspace, start)
	}
}

// Outside a repo there is no such statement, so the atlas is the landing screen and a
// registered space is opened behind it to give the browser a real planning context.
func TestUIStartupOutsideARepoLandsOnTheAtlasOverASeededSpace(t *testing.T) {
	home := t.TempDir()
	t.Setenv(userconfig.DirEnv, home)
	repo := t.TempDir()
	if _, err := config.Init(repo, "", false); err != nil {
		t.Fatal(err)
	}
	app := newUITestApp(t, filepath.Join(t.TempDir(), "nowhere-near-a-repo"))
	if _, err := app.SpaceSvc.Add(repo, "registered", false); err != nil {
		t.Fatal(err)
	}
	workspace, start, landOnAtlas, err := app.uiStartup()
	if err != nil {
		t.Fatalf("uiStartup() error = %v", err)
	}
	if !landOnAtlas {
		t.Error("a launch outside any planning repo must land on the atlas")
	}
	if workspace.SpaceID != "registered" || workspace.Planning == nil || workspace.Layout == nil {
		t.Fatalf("seeded workspace = %+v, want the registered space fully opened", workspace)
	}
	// The config start follows the seeded space, never the directory that launched us.
	if start != workspace.Checkout {
		t.Fatalf("config start = %q, want the seeded checkout %q", start, workspace.Checkout)
	}
}

// Nothing to stand in and nothing registered is a dead end, and has to say so with both
// ways out rather than opening an empty browser.
func TestUIStartupWithoutRepoOrRegistryExplainsBothRemedies(t *testing.T) {
	t.Setenv(userconfig.DirEnv, t.TempDir())
	app := newUITestApp(t, filepath.Join(t.TempDir(), "nowhere-near-a-repo"))
	_, _, _, err := app.uiStartup()
	if err == nil || ExitCode(err) != 10 {
		t.Fatalf("uiStartup() error = %v (exit %d), want a not-found failure", err, ExitCode(err))
	}
	for _, want := range []string{"tskflwctl init", "space add"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err.Error(), want)
		}
	}
}

// An explicit --space/-C is an assertion about WHICH tree to open; a broken one must fail
// loudly rather than being quietly replaced by another registered space.
func TestUIDoesNotSubstituteASpaceForABrokenExplicitSelection(t *testing.T) {
	t.Setenv(userconfig.DirEnv, t.TempDir())
	missing := filepath.Join(t.TempDir(), "not-a-repo")
	_, err := runRootRC(t, "-C", missing, "ui")
	if err == nil || !strings.Contains(err.Error(), "not a taskflow planning repo") {
		t.Fatalf("explicit -C into a non-repo = %v, want the discovery failure", err)
	}
}

// newUITestApp wires the two registry/workspace capabilities `ui` startup uses and then
// mirrors the command's own hook: resolve when there is a repo to resolve, tolerate the
// ordinary discovery miss, and leave Cfg nil when there was none.
func newUITestApp(t *testing.T, start string) *App {
	t.Helper()
	app := &App{
		In: strings.NewReader(""), Out: io.Discard, ErrOut: io.Discard,
		SpaceSvc:     core.NewSpaceRegistryService(spacestore.New()),
		WorkspaceSvc: core.NewWorkspaceService(workspacestore.New()),
	}
	app.setStyle()
	if err := app.resolveFrom(start); err != nil && !errors.Is(err, config.ErrNoConfig) {
		t.Fatalf("unexpected resolve failure: %v", err)
	}
	return app
}
