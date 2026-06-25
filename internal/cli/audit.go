package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newAuditCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "audit", Short: "Work with code audits"}
	cmd.AddCommand(
		newAuditNewCmd(app),
		newAuditListCmd(app),
		newAuditShowCmd(app),
		newAuditFindingsCmd(app),
		newAuditLintCmd(app),
		newAuditMoveCmd(app, "close", "Move audit(s) to closed/", domain.AuditClosed),
		newAuditMoveCmd(app, "reopen", "Move audit(s) back to open/", domain.AuditOpen),
		newAuditMoveCmd(app, "defer", "Move audit(s) to deferred/", domain.AuditDeferred),
	)
	return cmd
}

func newAuditNewCmd(app *App) *cobra.Command {
	var (
		p        core.NewAuditParams
		bodyFile string
	)
	cmd := &cobra.Command{
		Use:               "new <area>",
		Short:             "Create a new audit (open bucket, scaffolded findings)",
		Example:           "  tskflwctl audit new dispatcher\n  tskflwctl audit new arch-data-flow --date 2026-06-16",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: activeHelpArg("provide an area to audit (e.g. dispatcher)"),
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Area = args[0]
			body, err := resolveBody(cmd, p.Body, bodyFile)
			if err != nil {
				return err
			}
			p.Body = body
			p.DryRun = app.DryRun
			a, err := app.Svc.NewAudit(p)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.CreatedJSON(app.Out, "audit", a.Slug, string(a.Bucket), app.rel(a.Path), app.DryRun)
			}
			render.CreatedHuman(app.Out, app.Style, app.linkPath(a.Path), app.DryRun)
			if !app.DryRun {
				fmt.Fprintf(app.Out, "%s\n", app.Style.Dim("→ next: add findings, then tskflwctl audit close "+a.Slug))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Date, "date", "", "audit date YYYY-MM-DD (default today)")
	cmd.Flags().StringVar(&p.Body, "body", "", "override the default scaffold")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read the body from a file, or - for stdin (replaces --body)")
	cmd.Flags().StringVar(&p.Template, "template", "", `body scaffold to use (default "default"); e.g. "security". completes the available names`)
	cmd.MarkFlagsMutuallyExclusive("body", "body-file", "template")
	_ = cmd.RegisterFlagCompletionFunc("template", completeTemplateNames("audit"))
	return cmd
}

func newAuditListCmd(app *App) *cobra.Command {
	var (
		all, closed, deferred bool
		lm                    listMode
	)
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List audits (open by default)",
		Example:     "  tskflwctl audit list\n  tskflwctl audit list --all -o table -c slug,open\n  tskflwctl audit list --closed -o json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			mode, err := lm.resolve(cmd, app)
			if err != nil {
				return err
			}
			bucket := ""
			switch {
			case closed:
				bucket = string(domain.AuditClosed)
			case deferred:
				bucket = string(domain.AuditDeferred)
			}
			audits, problems, err := app.Svc.ListAudits(bucket, all)
			if err != nil {
				return err
			}
			if err := renderList(app, mode, lm.columns, audits, problems,
				"audits", render.AuditColumns(), render.AuditsJSON, render.AuditsHuman); err != nil {
				return err
			}
			return problemsError(problems)
		},
	}
	lm.bind(cmd, render.Specs(render.AuditColumns()))
	cmd.Flags().BoolVar(&all, "all", false, "all buckets")
	cmd.Flags().BoolVar(&closed, "closed", false, "closed audits only")
	cmd.Flags().BoolVar(&deferred, "deferred", false, "deferred audits only")
	cmd.MarkFlagsMutuallyExclusive("all", "closed", "deferred")
	return cmd
}

func newAuditFindingsCmd(app *App) *cobra.Command {
	var (
		status, effort, urgency []string
		component               string
		lm                      listMode
	)
	cmd := &cobra.Command{
		Use:   "findings [audit]",
		Short: "Query findings across audits (or one) by status/effort/urgency/component",
		Long: "Search audit findings — the structured per-finding view, not the aggregate.\n" +
			"With no argument, searches every audit; with an audit slug, just that one.\n" +
			"status/effort/urgency match exactly (case-insensitive, comma = any-of);\n" +
			"--component is a case-insensitive substring. Each --json hit carries its\n" +
			"audit slug and bucket.",
		Example: "  tskflwctl audit findings --status open --effort XS,S --json\n" +
			"  tskflwctl audit findings 2026-06-14-simplify-apigateway --status in-progress\n" +
			"  tskflwctl audit findings --component stravapipe -o table",
		Args:              cobra.MaximumNArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeAuditSlugs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mode, err := lm.resolve(cmd, app)
			if err != nil {
				return err
			}
			f := core.FindingFilter{Status: status, Effort: effort, Urgency: urgency, Component: component}
			if len(args) == 1 {
				f.Audit = args[0]
			}
			findings, problems, err := app.Svc.QueryFindings(f)
			if err != nil {
				return err
			}
			if err := renderList(app, mode, lm.columns, findings, problems,
				"findings", render.FindingColumns(), render.FindingsJSON, render.FindingsHuman); err != nil {
				return err
			}
			return problemsError(problems)
		},
	}
	lm.bind(cmd, render.Specs(render.FindingColumns()))
	cmd.Flags().StringSliceVar(&status, "status", nil, "filter by finding status (comma-separated, any-of)")
	cmd.Flags().StringSliceVar(&effort, "effort", nil, "filter by effort XS,S,M,L (any-of)")
	cmd.Flags().StringSliceVar(&urgency, "urgency", nil, "filter by urgency acute,soon,eventually (any-of)")
	cmd.Flags().StringVar(&component, "component", "", "filter by component (case-insensitive substring)")
	return cmd
}

func newAuditLintCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "lint [audit]",
		Short: "Validate audit findings (status vocabulary, missing status, bucket↔state)",
		Long: "Lint audit findings — the audit analog of `lint` (which covers tasks/epics).\n" +
			"Checks every finding has a legal **Status:** (catching typos a free-text edit\n" +
			"allows) and that a non-open audit has no still-open findings. With no argument\n" +
			"it lints every audit; with a slug, just that one. Exit 11 when issues are found.",
		Example:           "  tskflwctl audit lint\n  tskflwctl audit lint 2026-06-14-gateway --json",
		Args:              cobra.MaximumNArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeAuditSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			slug := ""
			if len(args) == 1 {
				slug = args[0]
			}
			results, problems, err := app.Svc.LintAudits(slug)
			if err != nil {
				return err
			}
			if app.JSON {
				if err := render.LintJSON(app.Out, results, problems); err != nil {
					return err
				}
			} else {
				render.ProblemsHuman(app.ErrOut, app.Style, problems)
				render.LintHuman(app.Out, app.Style, results, "audit")
				if len(results) == 0 && len(problems) == 0 {
					fmt.Fprintf(app.Out, "%s all audit findings pass lint\n", app.Style.Green("✔"))
				}
			}
			if len(results)+len(problems) > 0 {
				return fmt.Errorf("%w: %d audit(s) with finding issues, %d unreadable file(s)",
					domain.ErrValidation, len(results), len(problems))
			}
			return nil
		},
	}
}

func newAuditShowCmd(app *App) *cobra.Command {
	var raw bool
	cmd := &cobra.Command{
		Use:               "show <audit>",
		Short:             "Show an audit's metadata and body",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeAuditSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			audit, body, err := app.Svc.ShowAudit(args[0])
			if err != nil {
				return err
			}
			if app.JSON {
				return render.AuditShowJSON(app.Out, audit, body)
			}
			// Parse findings from the RAW body for the status-grouped tree; the body
			// passed to the renderer is the rendered (glamour/raw) markdown.
			findings := domain.ParseFindings(body)
			return app.paged(func(w io.Writer) error {
				return render.AuditShowHuman(w, app.Style, audit, findings, render.RenderBody(app.Style, body, app.markdownStyle, raw))
			})
		},
	}
	cmd.Flags().BoolVar(&raw, "raw", false, "print the raw markdown body (skip rendering)")
	return cmd
}

func newAuditMoveCmd(app *App, use, short string, to domain.AuditBucket) *cobra.Command {
	return &cobra.Command{
		Use:               use + " <audit>...",
		Short:             short,
		Example:           "  tskflwctl audit " + use + " 2026-06-06-schemas-scripts",
		Args:              cobra.MinimumNArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.auditCompleter(to), // don't offer audits already at `to`
		RunE: func(_ *cobra.Command, args []string) error {
			return runMoves(app, args, string(to),
				func(slug string) (domain.Audit, error) { return app.Svc.MoveAudit(slug, to, app.DryRun) },
				func(a domain.Audit) string { return a.Slug })
		},
	}
}
