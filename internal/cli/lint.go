package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newLintCmd(app *App) *cobra.Command {
	var fix, dryRun bool
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Validate active task frontmatter (--fix to auto-repair)",
		Args:  cobra.NoArgs,
		// Read-only by default; --fix opts into mutation explicitly.
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			if fix {
				return runLintFix(app, dryRun)
			}
			return runLint(app)
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "auto-repair frontmatter (quote ':' values, normalize list fields)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "with --fix, show changes without writing")
	return cmd
}

func runLint(app *App) error {
	results, problems, err := app.Svc.Lint()
	if err != nil {
		return err
	}
	if app.JSON {
		if err := render.LintJSON(app.Out, results, problems); err != nil {
			return err
		}
	} else {
		render.ProblemsHuman(app.Out, problems)
		render.LintHuman(app.Out, results)
		if len(results) == 0 && len(problems) == 0 {
			fmt.Fprintln(app.Out, "all active tasks pass lint")
		}
	}
	if len(results)+len(problems) > 0 {
		return fmt.Errorf("%w: %d task(s) with issues, %d unreadable file(s)",
			domain.ErrValidation, len(results), len(problems))
	}
	return nil
}

func runLintFix(app *App, dryRun bool) error {
	results, err := app.Svc.LintFix(dryRun)
	if err != nil {
		return err
	}
	if app.JSON {
		return render.FixJSON(app.Out, results, dryRun)
	}
	render.FixHuman(app.Out, results, dryRun)
	return nil
}
