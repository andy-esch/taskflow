package cli

import (
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
)

func newStatusCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "status",
		Short:       "At-a-glance project dashboard (counts, in-progress, epic progress)",
		Example:     "  tskflwctl status\n  tskflwctl status --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			s, err := app.Svc.Summary()
			if err != nil {
				return err
			}
			if app.JSON {
				return render.SummaryJSON(app.Out, s)
			}
			return render.SummaryHuman(app.Out, app.Style, s)
		},
	}
}
