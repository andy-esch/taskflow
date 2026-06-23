package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newDoctorCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Audit planning_repo <-> tracked_repos linkback integrity",
		Long: "Audit the cross-repo links: an impl repo's planning_repo pointer should be\n" +
			"matched by the planning repo tracking it back, and every tracked_repos entry\n" +
			"should exist and point its planning_repo back here. Reports each inconsistency\n" +
			"and exits non-zero when any is found — usable as a CI gate.",
		Example:     "  tskflwctl doctor\n  tskflwctl doctor --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		// Own PreRunE: resolve the repo but SKIP the root's ambient ⚠ link warning —
		// doctor reports the same findings on stdout (with an exit code), so the
		// stderr warning would just duplicate them.
		PersistentPreRunE: func(*cobra.Command, []string) error { app.setStyle(); return app.resolve() },
		RunE: func(_ *cobra.Command, _ []string) error {
			links := config.CheckLinks(app.Cfg)
			problems := make([]render.DoctorProblem, len(links))
			for i, p := range links {
				problems[i] = render.DoctorProblem{Repo: p.Repo, Message: p.Message}
			}
			if app.JSON {
				if err := render.DoctorJSON(app.Out, app.Cfg.Root, problems); err != nil {
					return err
				}
			} else {
				render.DoctorHuman(app.Out, app.Style, problems)
			}
			if len(problems) > 0 {
				return fmt.Errorf("%w: %d linkback problem(s)", domain.ErrValidation, len(problems))
			}
			return nil
		},
	}
	return cmd
}
