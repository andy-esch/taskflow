package cli

import (
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newAuditCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "audit", Short: "Work with code audits"}
	cmd.AddCommand(
		newAuditListCmd(app),
		newAuditShowCmd(app),
		newAuditMoveCmd(app, "close", "Move audit(s) to closed/", domain.AuditClosed),
		newAuditMoveCmd(app, "reopen", "Move audit(s) back to open/", domain.AuditOpen),
		newAuditMoveCmd(app, "defer", "Move audit(s) to deferred/", domain.AuditDeferred),
	)
	return cmd
}

func newAuditListCmd(app *App) *cobra.Command {
	var all, closed, deferred bool
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List audits (open by default)",
		Example:     "  tskflwctl audit list\n  tskflwctl audit list --all\n  tskflwctl audit list --closed --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
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
			if app.JSON {
				if err := render.AuditsJSON(app.Out, audits, problems); err != nil {
					return err
				}
			} else {
				if err := render.AuditsHuman(app.Out, app.Style, audits); err != nil {
					return err
				}
				render.ProblemsHuman(app.ErrOut, app.Style, problems)
			}
			return problemsError(problems)
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "all buckets")
	cmd.Flags().BoolVar(&closed, "closed", false, "closed audits only")
	cmd.Flags().BoolVar(&deferred, "deferred", false, "deferred audits only")
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
				func(slug string) (domain.Audit, error) { return app.Svc.MoveAudit(slug, to) },
				func(a domain.Audit) string { return a.Slug })
		},
	}
}
