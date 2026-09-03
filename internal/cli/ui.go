package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/tui"
)

func newUICmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Launch the interactive TUI (Bubble Tea)",
		Long: "Browse and update planning entities without restarting the full-screen TUI.\n" +
			"Inside a planning repo it opens that repo; run it anywhere else and it opens the\n" +
			"atlas of registered spaces instead. Press `a` or run `:atlas` to switch between the\n" +
			"two; `o` changes atlas ordering and `O` reverses it. Open the shared\n" +
			"Config/About editor from Overview, with `:config`, or from the command palette; writes\n" +
			"use the same typed application service as `tskflwctl config edit`. Threads have a\n" +
			"read-only list/detail route with lifecycle, graph health, progress, frontier, gates, and diagnostics.",
		Example:     "  tskflwctl ui",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "mutating"},
		// The atlas is a home-registry surface, so `ui` is the one browsing command that
		// must start where there is no planning repo at all. Only the AMBIENT miss is
		// forgiven: --space, -C, and TSKFLW_SPACE are explicit assertions about which tree
		// to open, and substituting a different one for a broken assertion is exactly the
		// wrong-repo hazard the registry rules exist to prevent.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if isCompletionCommand(cmd) || app.wantsSpace() || app.Chdir != "" {
				return app.repoPreRun(cmd, args)
			}
			app.setStyle()
			if err := app.resolve(); err != nil {
				if !errors.Is(err, config.ErrNoConfig) {
					return err // a config that EXISTS but is broken is still fatal
				}
				app.warnPresentation(cmd)
				return nil
			}
			app.warnLinks()
			app.warnPresentation(cmd)
			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			if app.DryRun {
				return fmt.Errorf("%w: `ui` has no --dry-run preview (it is interactive and includes mutations)", domain.ErrValidation)
			}
			if !app.Gate.On() || !isTerminal(app.Out) {
				return fmt.Errorf("%w: `ui` needs an interactive terminal", domain.ErrValidation)
			}
			workspace, start, landOnAtlas, err := app.uiStartup()
			if err != nil {
				return err
			}
			opts := []tui.Option{
				tui.WithConfiguration(app.ConfigSvc, start, app.configurationOverrides()),
				tui.WithWorkspaceOpening(app.WorkspaceSvc),
				tui.WithAtlas(app.SpaceOverviewSvc),
				tui.WithAtlasTheme(app.atlasTheme()),
			}
			if landOnAtlas {
				opts = append(opts, tui.WithAtlasLanding())
			}
			return tui.Run(workspace, app.Th, opts...)
		},
	}
	return cmd
}

// uiStartup resolves which workspace the browser opens and which screen it lands on.
//
// Standing in a planning repo (or naming one with --space / -C) is an unambiguous
// statement of intent, so that repo's Overview is the landing screen and the atlas stays
// one keystroke away. Outside one there is no such statement: the atlas is the landing
// screen, and a registered space is opened behind it purely so the browser has a real
// planning context — every tab, the dashboard, and `:config` need one, and seeding it is
// far cheaper than teaching all of them a null-workspace mode. The seeded space is also
// where `esc` from the atlas lands, which is why it follows the same direct-over-pointer
// preference the atlas itself uses rather than an arbitrary registry row.
func (a *App) uiStartup() (core.Workspace, string, bool, error) {
	if a.Cfg != nil {
		start, err := a.startDir()
		if err != nil {
			return core.Workspace{}, "", false, err
		}
		return a.runtimeWorkspace(), start, false, nil
	}
	entry, ok, err := a.SpaceSvc.PreferredEntry()
	if err != nil {
		return core.Workspace{}, "", false, err
	}
	if !ok {
		return core.Workspace{}, "", false, fmt.Errorf(
			"%w: no planning repo here and no healthy registered space to open — run `tskflwctl init`, or `tskflwctl space add <path>`",
			domain.ErrNotFound)
	}
	workspace, err := a.WorkspaceSvc.Open(core.WorkspaceRequest{
		Start: entry.Checkout, SpaceID: entry.ID, ExpectedPlanningID: entry.PlanningID,
	})
	if err != nil {
		return core.Workspace{}, "", false, err
	}
	return workspace, workspace.Checkout, true, nil
}

// atlasTheme resolves the process-global visual preference (flag/environment/home)
// while deliberately excluding the checkout-local repository tier. The atlas belongs
// to the user's home registry; whichever repo launched `ui` must not brand the shell.
func (a *App) atlasTheme() design.Theme {
	userName := ""
	if a.User != nil {
		userName = a.User.Theme.Name
	}
	theme, _ := design.Lookup(themeName(a.Theme, os.Getenv("TSKFLW_THEME"), "", userName))
	return theme
}

func (a *App) runtimeWorkspace() core.Workspace {
	checkout := a.Cfg.Dir
	if checkout == "" {
		checkout = a.Cfg.Root
	}
	return core.Workspace{
		SpaceID: a.selectedSpace, Checkout: checkout, PlanningRoot: a.Cfg.Root,
		PlanningID: a.Cfg.ID, Planning: a.Svc, Layout: a.Layout,
	}
}
