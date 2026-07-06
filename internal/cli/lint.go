package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newLintCmd(app *App) *cobra.Command {
	var fix, links bool
	cmd := &cobra.Command{
		Use:     "lint",
		Short:   "Validate active task and epic frontmatter (--fix repairs tasks/audits and assigns missing ids)",
		Example: "  tskflwctl lint\n  tskflwctl lint --fix --dry-run\n  tskflwctl lint --links\n  tskflwctl lint --json",
		Args:    cobra.NoArgs,
		// Read-only by default; --fix opts into mutation explicitly.
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			if fix {
				return runLintFix(app, app.DryRun) // --dry-run is the persistent flag (root.go)
			}
			return runLint(app, links)
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "auto-repair frontmatter: quote ':' values, normalize lists, backfill missing task/audit ids; epics are text-only")
	cmd.Flags().BoolVar(&links, "links", false, "also check body cross-links: flag any [..](path.md) whose target file is missing (opt-in — a tree can carry pre-existing danglers)")
	return cmd
}

func runLint(app *App, links bool) error {
	results, problems, err := app.Svc.Lint()
	if err != nil {
		return err
	}
	// --links adds cross-reference integrity: a body link to a missing file surfaces as a
	// FileProblem, flowing through the same render + exit path. Opt-in, since a tree can
	// accumulate pre-existing danglers that would otherwise noise up the default gate.
	if links {
		danglers, err := app.Linter.DanglingLinks()
		if err != nil {
			return err
		}
		problems = append(problems, danglers...)
	}
	if app.JSON {
		if err := render.LintJSON(app.Out, results, problems); err != nil {
			return err
		}
	} else {
		// Diagnostics go to stderr, matching the list commands — scripts that
		// capture stderr for problems must see them on one consistent stream.
		render.ProblemsHuman(app.ErrOut, app.Style, problems)
		// Results mix tasks and epics now, so the footer noun is the neutral "item".
		render.LintHuman(app.Out, app.Style, results, "item")
		if len(results) == 0 && len(problems) == 0 {
			fmt.Fprintf(app.Out, "%s all active tasks and epics pass lint\n", app.Style.Green("✔"))
		}
	}
	if len(results)+len(problems) > 0 {
		return fmt.Errorf("%w: %d item(s) with issues, %d unreadable file(s)",
			domain.ErrValidation, len(results), len(problems))
	}
	return nil
}

func runLintFix(app *App, dryRun bool) error {
	results, err := app.Fixer.FixFrontmatter(dryRun)
	if err != nil {
		// A mid-run write failure still repaired earlier files: report that partial
		// progress before surfacing the error, so the user can reconcile what landed.
		if len(results) > 0 {
			if app.JSON {
				_ = render.FixJSON(app.Out, results, nil, nil, dryRun)
			} else {
				render.FixHuman(app.Out, app.Style, results, nil, dryRun)
			}
		}
		return err
	}
	// Dry-run only previews the repairs; nothing was written, so there's no
	// post-fix state to re-lint.
	if dryRun {
		if app.JSON {
			return render.FixJSON(app.Out, results, nil, nil, dryRun)
		}
		render.FixHuman(app.Out, app.Style, results, nil, dryRun)
		return nil
	}
	// The fixer only reports files it changed — issues it can't repair (epics are
	// report-only; some task issues aren't auto-fixable) and unreadable files would
	// otherwise exit 0 in silence, leaving the tree broken while claiming success.
	// Re-lint and surface BOTH the leftover results and problems, with plain lint's exit.
	results2, problems, err := app.Svc.Lint()
	if err != nil {
		return err
	}
	if app.JSON {
		// One envelope carrying what was fixed plus what couldn't be (leftover lint
		// findings + unreadable files) — a --json consumer must never parse the prose
		// error to learn that.
		if err := render.FixJSON(app.Out, results, problems, results2, dryRun); err != nil {
			return err
		}
	} else {
		render.FixHuman(app.Out, app.Style, results, results2, dryRun)
		render.ProblemsHuman(app.ErrOut, app.Style, problems)
	}
	if len(results2)+len(problems) > 0 {
		return fmt.Errorf("%w: %d item(s) still with issues, %d unreadable file(s)",
			domain.ErrValidation, len(results2), len(problems))
	}
	return nil
}
