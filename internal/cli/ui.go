package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/tui"
)

func newUICmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Launch the interactive TUI (Bubble Tea)",
		Long: "Browse and update planning entities in the full-screen TUI. Open the shared\n" +
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
			return tui.Run(app.Svc, app.Layout, app.Th,
				tui.WithConfiguration(app.ConfigSvc, start, app.configurationOverrides()))
		},
	}
}
