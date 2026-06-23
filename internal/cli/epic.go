package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newEpicCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "epic", Short: "Work with epics"}
	cmd.AddCommand(newEpicNewCmd(app), newEpicListCmd(app), newEpicShowCmd(app))
	return cmd
}

func newEpicNewCmd(app *App) *cobra.Command {
	var (
		p        core.NewEpicParams
		bodyFile string
	)
	cmd := &cobra.Command{
		Use:               "new <title>",
		Short:             "Create a new epic (auto-numbered NN-slug)",
		Example:           "  tskflwctl epic new \"Billing overhaul\" --description \"Replace the legacy pipeline\"",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: activeHelpArg("provide an epic title (quote it if it has spaces)"),
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Title = args[0]
			body, err := resolveBody(cmd, p.Body, bodyFile)
			if err != nil {
				return err
			}
			p.Body = body
			p.DryRun = app.DryRun
			e, err := app.Svc.NewEpic(p)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.CreatedJSON(app.Out, "epic", e.ID, e.Status, app.rel(e.Path), app.DryRun)
			}
			render.CreatedHuman(app.Out, app.Style, app.linkPath(e.Path), app.DryRun)
			if !app.DryRun {
				fmt.Fprintf(app.Out, "%s\n", app.Style.Dim("→ next: tskflwctl task new \"Title\" --epic "+e.ID))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Description, "description", "", "one-line description (required, <=150 chars)")
	cmd.Flags().StringVar(&p.Status, "status", "planning", "epic status: planning|in-progress|completed|archived")
	cmd.Flags().StringVar(&p.Priority, "priority", "medium", "high|medium|low")
	cmd.Flags().StringSliceVar(&p.Tags, "tags", nil, "comma-separated tags")
	cmd.Flags().StringVar(&p.Body, "body", "", "override the default body scaffold")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read the body from a file, or - for stdin (replaces --body)")
	cmd.Flags().StringVar(&p.Template, "template", "", `body scaffold to use (default "default"); completes the available names`)
	cmd.MarkFlagsMutuallyExclusive("body", "body-file", "template")
	_ = cmd.RegisterFlagCompletionFunc("template", completeTemplateNames("epic"))
	return cmd
}

func newEpicListCmd(app *App) *cobra.Command {
	var (
		lm           listMode
		statusFilter string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List epics with task rollup",
		Example: "  tskflwctl epic list\n" +
			"  tskflwctl epic list --status in-progress\n" +
			"  tskflwctl epic list -o table -c id,status,percent,description",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			mode, err := lm.resolve(cmd, app)
			if err != nil {
				return err
			}
			// Validate the filter up front: epic status is a closed vocabulary, so a
			// typo is a loud error (exit 11), never a silently-empty list.
			if statusFilter != "" {
				if err := domain.ValidateEpicStatus(statusFilter); err != nil {
					return err
				}
			}
			epics, problems, err := app.Svc.ListEpics()
			if err != nil {
				return err
			}
			epics = filterEpicsByStatus(epics, statusFilter)
			if err := renderList(app, mode, lm.columns, epics, problems,
				"epics", render.EpicColumns(), render.EpicsJSON, render.EpicsHuman); err != nil {
				return err
			}
			return problemsError(problems)
		},
	}
	lm.bind(cmd, render.Specs(render.EpicColumns()))
	cmd.Flags().StringVar(&statusFilter, "status", "", "filter by epic status (planning|in-progress|completed|archived)")
	_ = cmd.RegisterFlagCompletionFunc("status",
		func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return domain.AllEpicStatuses(), cobra.ShellCompDirectiveNoFileComp
		})
	return cmd
}

// filterEpicsByStatus narrows the rollup list to a single epic status; an empty
// status keeps all. The cheap "don't pay for all epics" triage filter the other
// list commands already offer — done CLI-side since core.ListEpics has several
// callers and this is a small in-memory narrow, not a store query.
func filterEpicsByStatus(epics []core.EpicSummary, status string) []core.EpicSummary {
	if status == "" {
		return epics
	}
	var out []core.EpicSummary
	for _, e := range epics {
		if e.Epic.Status == status {
			out = append(out, e)
		}
	}
	return out
}

func newEpicShowCmd(app *App) *cobra.Command {
	var raw bool
	cmd := &cobra.Command{
		Use:               "show <epic>",
		Short:             "Show an epic and the tasks under it",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeEpicIDs,
		RunE: func(_ *cobra.Command, args []string) error {
			epic, tasks, body, err := app.Svc.ShowEpic(args[0])
			if err != nil {
				return err
			}
			if app.JSON {
				return render.EpicShowJSON(app.Out, epic, tasks, body)
			}
			return render.EpicShowHuman(app.Out, app.Style, epic, tasks, render.RenderBody(app.Style, body, app.markdownStyle(), raw))
		},
	}
	cmd.Flags().BoolVar(&raw, "raw", false, "print the raw markdown body (skip rendering)")
	return cmd
}
