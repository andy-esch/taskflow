package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
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
		Use:         "new <title>",
		Short:       "Create a new epic (auto-numbered NN-slug)",
		Example:     "  tskflwctl epic new \"Billing overhaul\" --description \"Replace the legacy pipeline\"",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"safety": "mutating"},
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
			render.CreatedHuman(app.Out, app.Style, app.rel(e.Path), app.DryRun)
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
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	return cmd
}

func newEpicListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "list",
		Short:       "List epics with task rollup",
		Example:     "  tskflwctl epic list\n  tskflwctl epic list --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			epics, problems, err := app.Svc.ListEpics()
			if err != nil {
				return err
			}
			if app.JSON {
				if err := render.EpicsJSON(app.Out, epics, problems); err != nil {
					return err
				}
			} else {
				if err := render.EpicsHuman(app.Out, app.Style, epics); err != nil {
					return err
				}
				render.ProblemsHuman(app.ErrOut, app.Style, problems)
			}
			return problemsError(problems)
		},
	}
}

func newEpicShowCmd(app *App) *cobra.Command {
	return &cobra.Command{
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
			return render.EpicShowHuman(app.Out, app.Style, epic, tasks, body)
		},
	}
}
