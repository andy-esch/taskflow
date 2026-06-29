package cli

import (
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/tui"
)

func newUICmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "ui",
		Short:       "Launch the interactive TUI (Bubble Tea)",
		Example:     "  tskflwctl ui",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			return tui.Run(app.Svc, app.Layout, app.Th)
		},
	}
}
