package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

// threadCreationCommandFailure enriches a post-commit cleanup failure with the
// workspace/path facts only the primary adapter knows. Error classification is
// preserved through Unwrap; JSON errors can carry a complete inspection receipt.
type threadCreationCommandFailure struct {
	cause     error
	receipt   core.ThreadCreationReceipt
	path      string
	workspace wire.WorkspaceJSON
}

func (e *threadCreationCommandFailure) Error() string { return e.cause.Error() }
func (e *threadCreationCommandFailure) Unwrap() error { return e.cause }

func newThreadCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "thread", Short: "Work with initiative Threads over the task DAG"}
	cmd.AddCommand(
		newThreadNewCmd(app), newThreadListCmd(app), newThreadShowCmd(app),
		newThreadPathCmd(app), newThreadFrontierCmd(app),
	)
	return cmd
}

func newThreadNewCmd(app *App) *cobra.Command {
	var (
		params   core.NewThreadParams
		bodyFile string
	)
	cmd := &cobra.Command{
		Use:   "new <title>",
		Short: "Create an unstarted Thread with optional initial task members",
		Long:  "Create one unstarted Thread. Repeat --task to add initial members; every reference is resolved and validated with current tasks and Threads under the repository guard.",
		Example: "  tskflwctl thread new \"Thread delivery\" --description \"Ship production Threads\" --goal \"Dogfood the remaining implementation\"\n" +
			"  tskflwctl thread new \"Thread delivery\" --description \"Ship production Threads\" --goal \"Dogfood it\" --task documents --task lifecycle",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: activeHelpArg("provide a Thread title (quote it if it has spaces)"),
		RunE: func(cmd *cobra.Command, args []string) error {
			params.Title = args[0]
			body, err := resolveBody(cmd, params.Body, bodyFile)
			if err != nil {
				return err
			}
			params.Body, params.DryRun = body, app.DryRun
			receipt, err := app.Svc.NewThread(params)
			if err != nil {
				var committed *core.ThreadCreationMutationFailure
				if errors.As(err, &committed) {
					return &threadCreationCommandFailure{
						cause: err, receipt: committed.Receipt, path: app.rel(committed.Receipt.Thread.Path), workspace: app.workspace(),
					}
				}
				return err
			}
			path := app.rel(receipt.Thread.Path)
			if app.JSON {
				return render.ThreadMutationJSON(app.Out, receipt, path, app.workspace())
			}
			render.ThreadCreatedHuman(app.Out, app.Style, receipt, app.linkPath(receipt.Thread.Path))
			return nil
		},
	}
	cmd.Flags().StringVar(&params.Description, "description", "", fmt.Sprintf("one-line description (<=%d chars; required)", domain.MaxDescriptionLen))
	cmd.Flags().StringVar(&params.Goal, "goal", "", "one-line observable finish line (required)")
	cmd.Flags().StringVar(&params.TargetDate, "target-date", "", "optional human planning target (YYYY-MM-DD)")
	cmd.Flags().StringSliceVar(&params.Tags, "tags", nil, "comma-separated tags")
	cmd.Flags().StringArrayVar(&params.Tasks, "task", nil, "initial task reference (repeatable)")
	cmd.Flags().StringVar(&params.Body, "body", "", "replace the default body scaffold")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read body from a file, or - for stdin")
	cmd.Flags().StringVar(&params.Template, "template", "", "body template name (default: default)")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	_ = cmd.RegisterFlagCompletionFunc("task", app.completeTaskSlugs)
	_ = cmd.RegisterFlagCompletionFunc("template", completeTemplateNames("thread"))
	return cmd
}

func newThreadListCmd(app *App) *cobra.Command {
	var status string
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List Threads with nominal and sound progress",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			if status != "" {
				if err := domain.ValidateThreadStatus(domain.ThreadStatus(status)); err != nil {
					return err
				}
			}
			list, problems, err := app.Svc.ListThreadViews()
			if err != nil {
				return err
			}
			if status != "" {
				filtered := list.Threads[:0]
				for _, view := range list.Threads {
					if string(view.Thread.Status) == status {
						filtered = append(filtered, view)
					}
				}
				list.Threads = filtered
			}
			if app.JSON {
				if err := render.ThreadsJSON(app.Out, list, problems); err != nil {
					return err
				}
			} else {
				if err := render.ThreadsHuman(app.Out, app.Style, list); err != nil {
					return err
				}
				render.ProblemsHuman(app.ErrOut, app.Style, problems)
			}
			return problemsError(problems)
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "filter by Thread status")
	_ = cmd.RegisterFlagCompletionFunc("status", completeThreadStatusValues)
	return cmd
}

func completeThreadStatusValues(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	values := domain.AllThreadStatuses()
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = string(value)
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func newThreadShowCmd(app *App) *cobra.Command {
	var raw bool
	cmd := &cobra.Command{
		Use:               "show <thread>",
		Short:             "Show Thread progress, members, gates, frontier, and body",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeThreadSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			view, body, err := app.Svc.ShowThread(args[0])
			if err != nil {
				return err
			}
			if app.JSON {
				return render.ThreadShowJSON(app.Out, view, body)
			}
			return app.paged(func(w io.Writer) error {
				return render.ThreadShowHuman(w, app.Style, view, render.RenderBody(app.Style, body, app.markdownStyle, raw))
			})
		},
	}
	cmd.Flags().BoolVar(&raw, "raw", false, "print the raw Markdown body")
	return cmd
}

func newThreadPathCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "path <thread>",
		Short:             "Print the absolute path to a Thread file",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeThreadSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			path, err := app.Svc.ThreadPath(args[0])
			if err != nil {
				return err
			}
			return emitPath(app, absPath(path))
		},
	}
}

func newThreadFrontierCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "frontier <thread>",
		Short:             "Show currently eligible member tasks",
		Long:              "Show dispatchable members from the shared task-graph projection. An unhealthy projection returns no tasks and retains its diagnosis.",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeThreadSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			view, _, err := app.Svc.ShowThread(args[0])
			if err != nil {
				return err
			}
			if app.JSON {
				return render.ThreadFrontierJSON(app.Out, view)
			}
			return render.ThreadFrontierHuman(app.Out, app.Style, view)
		},
	}
}
