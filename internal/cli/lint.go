package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newLintCmd(app *App) *cobra.Command {
	var fix bool
	cmd := &cobra.Command{
		Use:     "lint",
		Short:   "Validate active task frontmatter (--fix to auto-repair)",
		Example: "  tskflwctl lint\n  tskflwctl lint --fix --dry-run\n  tskflwctl lint --json",
		Args:    cobra.NoArgs,
		// Read-only by default; --fix opts into mutation explicitly.
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			if fix {
				return runLintFix(app, app.DryRun) // --dry-run is the persistent flag (root.go)
			}
			return runLint(app)
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "auto-repair frontmatter (quote ':' values, normalize list fields)")
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
		// Diagnostics go to stderr, matching the list commands — scripts that
		// capture stderr for problems must see them on one consistent stream.
		render.ProblemsHuman(app.ErrOut, app.Style, problems)
		render.LintHuman(app.Out, app.Style, results)
		if len(results) == 0 && len(problems) == 0 {
			fmt.Fprintf(app.Out, "%s all active tasks pass lint\n", app.Style.Green("✔"))
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
	// Dry-run only previews the repairs; nothing was written, so there's no
	// post-fix state to re-lint.
	if dryRun {
		if app.JSON {
			return render.FixJSON(app.Out, results, nil, dryRun)
		}
		render.FixHuman(app.Out, app.Style, results, dryRun)
		return nil
	}
	// The fixer only reports files it changed — a file it can't repair would
	// otherwise exit 0 in silence, leaving the tree broken while claiming
	// success. Re-lint and surface what's still wrong, with plain lint's exit.
	_, problems, err := app.Svc.Lint()
	if err != nil {
		return err
	}
	if app.JSON {
		// One envelope carrying both what was fixed and what couldn't be —
		// a --json consumer must never parse the prose error to learn that.
		if err := render.FixJSON(app.Out, results, problems, dryRun); err != nil {
			return err
		}
	} else {
		render.FixHuman(app.Out, app.Style, results, dryRun)
		render.ProblemsHuman(app.ErrOut, app.Style, problems)
	}
	if len(problems) > 0 {
		return fmt.Errorf("%w: %d file(s) could not be auto-repaired", domain.ErrValidation, len(problems))
	}
	return nil
}
