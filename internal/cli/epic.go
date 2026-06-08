package cli

import (
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
)

func newEpicCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "epic", Short: "Work with epics"}
	cmd.AddCommand(newEpicListCmd(app), newEpicShowCmd(app))
	return cmd
}

func newEpicListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "list",
		Short:       "List epics with task rollup",
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
				if err := render.EpicsHuman(app.Out, epics); err != nil {
					return err
				}
				render.ProblemsHuman(app.ErrOut, problems)
			}
			return problemsError(problems)
		},
	}
}

func newEpicShowCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "show <epic>",
		Short:       "Show an epic and the tasks under it",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, args []string) error {
			epic, tasks, body, err := app.Svc.ShowEpic(args[0])
			if err != nil {
				return err
			}
			if app.JSON {
				return render.EpicShowJSON(app.Out, epic, tasks, body)
			}
			return render.EpicShowHuman(app.Out, epic, tasks, body)
		},
	}
}
