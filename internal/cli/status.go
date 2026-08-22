package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newStatusCmd(app *App) *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "At-a-glance project dashboard (counts, in-progress, epic progress)",
		Long: "Show the current planning repo's dashboard. With --all, summarize every\n" +
			"logical planning space in the home registry and combine their in-progress work.\n" +
			"Multiple registered entry points sharing one planning id are read only once. The\n" +
			"command works from any directory; -C is used only when the registry is empty.\n\n" +
			"Broken registry entries remain inline and informational. Unreadable planning files\n" +
			"or a selected tree that fails to load still render every available result, then make\n" +
			"the command exit non-zero so automation can detect the partial result.",
		Example:     "  tskflwctl status\n  tskflwctl status --json\n  tskflwctl status --all\n  tskflwctl status --all --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !all {
				return app.repoPreRun(cmd, args)
			}
			// Cross-space status works anywhere. Presentation warnings are deferred to
			// RunE because an empty registry falls back to current-repo resolution first.
			app.setStyle()
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !all {
				return runCurrentStatus(app)
			}
			if strings.TrimSpace(app.Space) != "" {
				return fmt.Errorf("%w: --all and --space select different scopes; pass one", domain.ErrValidation)
			}
			overview, err := app.SpaceOverviewSvc.Overview()
			if err != nil {
				return err
			}
			if len(overview.Spaces) == 0 {
				start := app.Chdir
				if start == "" {
					start, err = os.Getwd()
					if err != nil {
						return fmt.Errorf("getwd: %w", err)
					}
				}
				if err := app.resolveFrom(start); err != nil {
					return err
				}
				app.warnLinks()
				app.warnPresentation(cmd)
				return runCurrentStatus(app)
			}
			app.warnPresentation(cmd)
			var renderErr error
			if app.JSON {
				renderErr = render.StatusAllJSON(app.Out, overview)
			} else {
				renderErr = render.StatusAllHuman(app.Out, app.Style, overview)
			}
			if renderErr != nil {
				return renderErr
			}
			return statusAllProblemsError(overview)
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "summarize every logical planning space in the registry")
	return cmd
}

// statusAllProblemsError applies the same unreadable-entity exit contract as ordinary
// status after every available space has rendered. Registry health remains informational:
// a group with no healthy entry point is already diagnosed inline and does not fail the
// command. A selected tree that subsequently cannot be read is a partial planning-data
// failure and must not look successful to automation.
func statusAllProblemsError(overview core.SpaceOverview) error {
	unreadableFiles := 0
	failedSpaces := 0
	for _, space := range overview.Spaces {
		if space.Summary != nil {
			unreadableFiles += len(space.Summary.Problems)
			continue
		}
		if space.Selected != nil && space.LoadError != "" {
			failedSpaces++
		}
	}
	if unreadableFiles == 0 && failedSpaces == 0 {
		return nil
	}
	parts := make([]string, 0, 2)
	if unreadableFiles > 0 {
		parts = append(parts, fmt.Sprintf("%d file(s) with unreadable frontmatter", unreadableFiles))
	}
	if failedSpaces > 0 {
		parts = append(parts, fmt.Sprintf("%d planning space(s) failed to load", failedSpaces))
	}
	return fmt.Errorf("%w: cross-space status incomplete: %s", domain.ErrValidation, strings.Join(parts, "; "))
}

func runCurrentStatus(app *App) error {
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
	// Render the dashboard first, then exit non-zero if any file was unreadable —
	// matching the list/lint/audit contract so an agent gating on `status` (incl.
	// --json, which carries the unreadable array) doesn't get exit 0 on a broken tree.
	return problemsError(s.Problems)
}
