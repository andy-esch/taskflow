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
				if err := render.SummaryJSON(app.Out, s); err != nil {
					return err
				}
			} else if err := render.SummaryHuman(app.Out, app.Style, s); err != nil {
				return err
			}
			// Render the dashboard first, then exit non-zero if any file was
			// unreadable — matching the list/lint/audit contract so an agent gating
			// on `status` (incl. --json, which carries the unreadable array) doesn't
			// get exit 0 on a forked/broken tree.
			return problemsError(s.Problems)
		},
	}
}
