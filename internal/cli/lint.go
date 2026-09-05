package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newLintCmd(app *App) *cobra.Command {
	var fix, links bool
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Validate entity frontmatter, audit findings, and task-dependency graph integrity",
		Long: "Validate task, epic, research, and Thread frontmatter plus audit findings, then validate the\n" +
			"repository-global task-dependency graph. Exactly resolved legacy dependency fields are visible\n" +
			"advisories; missing, ambiguous, or structurally unsafe references are errors.\n\n" +
			"--fix repairs ordinary frontmatter and missing ids. It never normalizes or changes\n" +
			"graph-owned task fields (depends_on, blocked_by, dependencies, or blocks); a\n" +
			"would-be graph repair is skipped and reported for deliberate remediation.",
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
	cmd.Flags().BoolVar(&fix, "fix", false, "auto-repair ordinary frontmatter and missing ids; graph-owned task fields are skipped")
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
			fmt.Fprintf(app.Out, "%s all planning entities and dependency links pass lint\n", app.Style.Green("✔"))
		}
	}
	blocking := core.BlockingLintResultCount(results)
	if blocking+len(problems) > 0 {
		return fmt.Errorf("%w: %d item(s) with issues, %d unreadable file(s)",
			domain.ErrValidation, blocking, len(problems))
	}
	return nil
}

func runLintFix(app *App, dryRun bool) error {
	results, err := app.Fixer.FixFrontmatter(dryRun)
	// Body repairs run after frontmatter: the frontmatter pass can rename a file to
	// heal a broken id, and the body pass resolves audits by slug.
	if err == nil {
		var headerFixes []domain.FixResult
		headerFixes, err = app.Svc.FixFindingHeaders(dryRun)
		results = append(results, headerFixes...)
	}
	if err != nil {
		// A mid-run write failure still repaired earlier files: report that partial
		// progress before surfacing the error, so the user can reconcile what landed.
		if len(results) > 0 {
			if app.JSON {
				_ = render.FixJSON(app.Out, results, nil, nil, dryRun, app.workspace())
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
			return render.FixJSON(app.Out, results, nil, nil, dryRun, app.workspace())
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
		if err := render.FixJSON(app.Out, results, problems, results2, dryRun, app.workspace()); err != nil {
			return err
		}
	} else {
		render.FixHuman(app.Out, app.Style, results, results2, dryRun)
		render.ProblemsHuman(app.ErrOut, app.Style, problems)
	}
	blocking := core.BlockingLintResultCount(results2)
	if blocking+len(problems) > 0 {
		return fmt.Errorf("%w: %d item(s) still with issues, %d unreadable file(s)",
			domain.ErrValidation, blocking, len(problems))
	}
	return nil
}
