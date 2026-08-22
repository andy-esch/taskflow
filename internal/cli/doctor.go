package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

// newDoctorCmd preserves the historical top-level spelling without spending a visible
// root-help slot. config doctor is canonical; both route through the same command builder
// and configuration service, so output and exit behavior cannot drift.
func newDoctorCmd(app *App) *cobra.Command {
	cmd := newDoctorCommand(app)
	cmd.Hidden = true
	return cmd
}

func newConfigDoctorCmd(app *App) *cobra.Command {
	cmd := newDoctorCommand(app)
	cmd.Example = "  tskflwctl config doctor\n  tskflwctl config doctor --json"
	return cmd
}

func newDoctorCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Audit linkback integrity and the home space registry",
		Long: "Audit the cross-repo links: an impl repo's planning_repo pointer should be\n" +
			"matched by the planning repo tracking it back, and every tracked_repos entry\n" +
			"should exist and point its planning_repo back here. Also diagnose every home\n" +
			"space-registry entry. Reports each inconsistency and exits non-zero when any\n" +
			"is found — usable as a CI gate; nothing is repaired or forgotten.",
		Example:     "  tskflwctl doctor\n  tskflwctl doctor --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		// Own PreRunE: resolve the repo but SKIP the root's ambient ⚠ link warning —
		// doctor reports the same findings on stdout (with an exit code), so the
		// stderr warning would just duplicate them.
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			app.setStyle()
			err := app.resolve()
			// After resolve so a repo [theme] participates; the ambient link ⚠ stays
			// suppressed (this command reports those findings on stdout itself).
			app.warnPresentation(cmd)
			return err
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			start, err := app.startDir()
			if err != nil {
				return err
			}
			diagnosis, err := app.ConfigSvc.Diagnose(start)
			if err != nil {
				return err
			}
			problems := make([]render.DoctorProblem, len(diagnosis.Problems))
			for i, p := range diagnosis.Problems {
				problems[i] = render.DoctorProblem{Repo: p.Repo, Message: p.Message}
			}
			registry := wire.DoctorRegistry{Checked: diagnosis.Registry.Checked}
			for _, problem := range diagnosis.Registry.Problems {
				registry.Problems = append(registry.Problems, wire.DoctorSpaceProblem{
					ID: problem.ID, Path: problem.Path, Kind: problem.Kind,
					Message: problem.Message, Remedy: problem.Remedy,
				})
			}
			if app.JSON {
				if err := render.DoctorJSON(app.Out, diagnosis.Root, problems, registry); err != nil {
					return err
				}
			} else {
				render.DoctorHuman(app.Out, app.Style, problems, registry)
			}
			total := len(problems) + len(registry.Problems)
			if total > 0 {
				return fmt.Errorf(
					"%w: %d doctor problem(s) (%d linkback, %d registered space)",
					domain.ErrValidation, total, len(problems), len(registry.Problems),
				)
			}
			return nil
		},
	}
	return cmd
}
