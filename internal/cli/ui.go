package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/tui"
)

func newUICmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Launch the interactive TUI (Bubble Tea)",
		Long: "Navigate registered spaces from the atlas, then browse and update planning entities\n" +
			"without restarting the full-screen TUI. Press `a` or run `:atlas` to return; `o` changes\n" +
			"atlas ordering and `O` reverses it. Open the shared\n" +
			"Config/About editor from Overview, with `:config`, or from the command palette; writes\n" +
			"use the same typed application service as `tskflwctl config edit`.",
		Example:     "  tskflwctl ui",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "mutating"},
		RunE: func(_ *cobra.Command, _ []string) error {
			if app.DryRun {
				return fmt.Errorf("%w: `ui` has no --dry-run preview (it is interactive and includes mutations)", domain.ErrValidation)
			}
			if !app.Gate.On() || !isTerminal(app.Out) {
				return fmt.Errorf("%w: `ui` needs an interactive terminal", domain.ErrValidation)
			}
			start, err := app.startDir()
			if err != nil {
				return err
			}
			return tui.Run(app.runtimeWorkspace(), app.Th,
				tui.WithConfiguration(app.ConfigSvc, start, app.configurationOverrides()),
				tui.WithWorkspaceOpening(app.WorkspaceSvc),
				tui.WithAtlas(app.SpaceOverviewSvc),
				tui.WithAtlasTheme(app.atlasTheme()))
		},
	}
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
