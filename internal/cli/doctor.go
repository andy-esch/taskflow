package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/spacehealth"
	"github.com/andy-esch/taskflow/internal/wire"
)

func newDoctorCmd(app *App) *cobra.Command {
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
			links := config.CheckLinks(app.Cfg)
			problems := make([]render.DoctorProblem, len(links))
			for i, p := range links {
				problems[i] = render.DoctorProblem{Repo: p.Repo, Message: p.Message}
			}
			diagnoses, err := spacehealth.DiagnoseRegistry()
			if err != nil {
				return classifySpaceRegistryError(err)
			}
			registry := wire.DoctorRegistry{Checked: len(diagnoses)}
			for _, diagnosis := range diagnoses {
				if !diagnosis.Broken() {
					continue
				}
				registry.Problems = append(registry.Problems, wire.DoctorSpaceProblem{
					ID: diagnosis.Space.ID, Path: diagnosis.Space.Path, Kind: string(diagnosis.Kind),
					Message: diagnosis.Message, Remedy: diagnosis.Remedy,
				})
			}
			if app.JSON {
				if err := render.DoctorJSON(app.Out, app.Cfg.Root, problems, registry); err != nil {
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
