package cli

import (
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
)

// newBoardCmd is the active-work board: tasks grouped by their active status (the
// next-up → ready-to-start → in-progress pipeline). The on-demand replacement for
// browsing tasks/<status>/; sibling to `status` (the aggregation dashboard).
func newBoardCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "board",
		Short:       "Active-work board: tasks by status (next-up → ready-to-start → in-progress)",
		Example:     "  tskflwctl board\n  tskflwctl board --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			b, err := app.Svc.Board()
			if err != nil {
				return err
			}
			if app.JSON {
				if err := render.BoardJSON(app.Out, b); err != nil {
					return err
				}
			} else if err := render.BoardHuman(app.Out, app.Style, b); err != nil {
				return err
			}
			// Render first, then exit non-zero if any file was unreadable — the same
			// contract as `status`/`list`, so an agent gating on `board` (incl. --json,
			// which carries the unreadable array) never gets exit 0 on a broken tree.
			return problemsError(b.Problems)
		},
	}
}
