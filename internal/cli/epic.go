package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newEpicCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "epic", Short: "Work with epics"}
	cmd.AddCommand(newEpicNewCmd(app), newEpicListCmd(app), newEpicShowCmd(app), newEpicMoveCmd(app))
	return cmd
}

// newEpicMoveCmd is the epic analog of `task move`: it transitions an epic to a
// target status (active/retired/deprecated). Epic status is a frontmatter FIELD,
// not a directory, so the move rewrites the field in place — no file is relocated
// — but the verb name mirrors task/audit moves for UX parity. Runs through the
// shared runMoves engine + the same `moves` JSON envelope.
func newEpicMoveCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "move <epic>... <status>",
		Short:       "Transition epic(s) to <status> (active|retired|deprecated)",
		Example:     "  tskflwctl epic move 18-tui retired\n  tskflwctl epic move 18-tui 20-cli deprecated --dry-run",
		Args:        cobra.MinimumNArgs(2),
		Annotations: map[string]string{"safety": "mutating"},
		// Position-aware: the final arg is a status from a small closed set, so it
		// offers epic statuses there — never epic ids (which would actively mislead).
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			opts, directive := app.completeEpicIDs(cmd, args, toComplete)
			if len(args) >= 1 {
				opts = append(opts, domain.AllEpicStatuses()...)
				if cobra.GetActiveHelpConfig(cmd) != "off" {
					opts = cobra.AppendActiveHelp(opts, "the final argument is the target status")
				}
			}
			return opts, directive
		},
		RunE: func(_ *cobra.Command, args []string) error {
			status := args[len(args)-1]
			if err := domain.ValidateEpicStatus(status); err != nil {
				return err // wraps ErrValidation and lists the valid statuses
			}
			return runMoves(app, args[:len(args)-1], status,
				func(id string) (domain.Epic, error) { return app.Svc.MoveEpic(id, status, app.DryRun) },
				func(e domain.Epic) string { return e.ID })
		},
	}
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
			render.CreatedSlugNote(app.Out, app.Style, p.Title, e.ID)
			if !app.DryRun {
				fmt.Fprintf(app.Out, "%s\n", app.Style.Dim("→ next: tskflwctl task new \"Title\" --epic "+e.ID))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Description, "description", "", fmt.Sprintf("one-line description (required, <=%d chars)", domain.MaxDescriptionLen))
	cmd.Flags().StringVar(&p.Status, "status", "active", "epic status: active|retired|deprecated")
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
			"  tskflwctl epic list --status active\n" +
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
	cmd.Flags().StringVar(&statusFilter, "status", "", "filter by epic status (active|retired|deprecated)")
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
		Example:           "  tskflwctl epic show 01-api-gateway\n  tskflwctl epic show   # pick from a list",
		Args:              cobra.MaximumNArgs(1), // bare → picker on a TTY; non-interactive needs the id
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeEpicIDs,
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := app.resolveOne(args, "specify an epic to show", "no epics available", "Epic to show", app.epicOptions)
			if err != nil {
				return err
			}
			epic, tasks, body, err := app.Svc.ShowEpic(id)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.EpicShowJSON(app.Out, epic, tasks, body)
			}
			return app.paged(func(w io.Writer) error {
				return render.EpicShowHuman(w, app.Style, epic, tasks, render.RenderBody(app.Style, body, app.markdownStyle, raw))
			})
		},
	}
	cmd.Flags().BoolVar(&raw, "raw", false, "print the raw markdown body (skip rendering)")
	return cmd
}
