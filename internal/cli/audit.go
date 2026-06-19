package cli

import (
	"fmt"

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
		Use:         "new <area>",
		Short:       "Create a new audit (open bucket, scaffolded findings)",
		Example:     "  tskflwctl audit new dispatcher\n  tskflwctl audit new arch-data-flow --date 2026-06-16",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"safety": "mutating"},
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
			render.CreatedHuman(app.Out, app.Style, app.rel(a.Path), app.DryRun)
			if !app.DryRun {
				fmt.Fprintf(app.Out, "%s\n", app.Style.Dim("→ next: add findings, then tskflwctl audit close "+a.Slug))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Date, "date", "", "audit date YYYY-MM-DD (default today)")
	cmd.Flags().StringVar(&p.Body, "body", "", "override the default scaffold")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read the body from a file, or - for stdin (replaces --body)")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
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
		Example:     "  tskflwctl audit list\n  tskflwctl audit list --all --plain\n  tskflwctl audit list --closed --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			mode, err := lm.resolve(app)
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
			if err := renderList(app, mode, audits, problems,
				render.AuditsJSON, render.AuditsPlain, render.AuditsHuman,
				func(a domain.Audit) string { return a.Slug }); err != nil {
				return err
			}
			return problemsError(problems)
		},
	}
	lm.bind(cmd)
	cmd.Flags().BoolVar(&all, "all", false, "all buckets")
	cmd.Flags().BoolVar(&closed, "closed", false, "closed audits only")
	cmd.Flags().BoolVar(&deferred, "deferred", false, "deferred audits only")
	cmd.MarkFlagsMutuallyExclusive("all", "closed", "deferred")
	return cmd
}

func newAuditShowCmd(app *App) *cobra.Command {
	return &cobra.Command{
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
			return render.AuditShowHuman(app.Out, app.Style, audit, body)
		},
	}
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
