package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newTaskCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "task", Short: "Work with tasks"}
	cmd.AddCommand(
		newTaskNewCmd(app),
		newTaskListCmd(app),
		newTaskShowCmd(app),
		newTaskSetCmd(app),
		newTaskMoveCmd(app),
		// Explicit transition verbs over the internal move engine (no enum to
		// hallucinate; per-verb intent). See the command spec.
		newTransitionCmd(app, "start", "Move task(s) to in-progress", domain.StatusInProgress),
		newTransitionCmd(app, "promote", "Move task(s) to next-up", domain.StatusNextUp),
		newTransitionCmd(app, "demote", "Move task(s) to ready-to-start", domain.StatusReadyToStart),
		newTransitionCmd(app, "complete", "Move task(s) to completed", domain.StatusCompleted),
		newTransitionCmd(app, "defer", "Move task(s) to deferred", domain.StatusDeferred),
		newTransitionCmd(app, "deprecate", "Move task(s) to deprecated", domain.StatusDeprecated),
	)
	return cmd
}

func newTaskNewCmd(app *App) *cobra.Command {
	var p core.NewTaskParams
	cmd := &cobra.Command{
		Use:         "new <title>",
		Short:       "Create a new task (validated, handoff-ready scaffold)",
		Example:     "  tskflwctl task new \"Add retry backoff\" --epic 17-pm-go-cli --tags net\n  tskflwctl task new \"Triage flaky test\" --epic 17-pm-go-cli --next",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"safety": "mutating"},
		RunE: func(_ *cobra.Command, args []string) error {
			p.Title = args[0]
			t, err := app.Svc.NewTask(p)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.CreatedJSON(app.Out, "task", t.Slug, t.Path)
			}
			render.CreatedHuman(app.Out, app.Style, app.rel(t.Path))
			fmt.Fprintf(app.Out, "%s\n", app.Style.Dim("→ next: tskflwctl task start "+t.Slug))
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Epic, "epic", "", "epic id (required)")
	cmd.Flags().StringVar(&p.Description, "description", "", "one-line description (<=150 chars)")
	cmd.Flags().StringVar(&p.Effort, "effort", "Unknown", "effort estimate")
	cmd.Flags().StringVar(&p.Priority, "priority", "medium", "high|medium|low")
	cmd.Flags().IntVar(&p.Tier, "tier", 3, "tier 1-5")
	cmd.Flags().IntVar(&p.Autonomy, "autonomy", 3, "autonomy level 1-5")
	cmd.Flags().StringSliceVar(&p.Tags, "tags", nil, "comma-separated tags")
	cmd.Flags().BoolVar(&p.Next, "next", false, "create in next-up instead of ready-to-start")
	cmd.Flags().StringVar(&p.Body, "body", "", "override the default body scaffold")
	_ = cmd.MarkFlagRequired("epic")
	_ = cmd.RegisterFlagCompletionFunc("epic", app.completeEpicIDs)
	return cmd
}

func newTaskListCmd(app *App) *cobra.Command {
	var filter core.TaskFilter
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List tasks (active by default)",
		Example:     "  tskflwctl task list\n  tskflwctl task list --all --epic 17-pm-go-cli\n  tskflwctl task list --status in-progress --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			tasks, problems, err := app.Svc.ListTasks(filter)
			if err != nil {
				return err
			}
			if app.JSON {
				if err := render.TasksJSON(app.Out, tasks, problems); err != nil {
					return err
				}
			} else {
				if err := render.TasksHuman(app.Out, app.Style, tasks); err != nil {
					return err
				}
				render.ProblemsHuman(app.ErrOut, app.Style, problems)
			}
			return problemsError(problems)
		},
	}
	cmd.Flags().StringVar(&filter.Status, "status", "", "filter by status")
	cmd.Flags().StringVar(&filter.Epic, "epic", "", "filter by epic")
	cmd.Flags().StringVar(&filter.Tag, "tag", "", "filter by tag")
	cmd.Flags().BoolVar(&filter.All, "all", false, "include completed/deprecated/deferred")
	return cmd
}

func newTaskShowCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "show <task>",
		Short:             "Show a task's metadata and body",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			task, body, err := app.Svc.ShowTask(args[0])
			if err != nil {
				return err
			}
			if app.JSON {
				return render.TaskShowJSON(app.Out, task, body)
			}
			return render.TaskShowHuman(app.Out, app.Style, task, body)
		},
	}
}

func newTaskSetCmd(app *App) *cobra.Command {
	var (
		description, priority, epic, effort string
		tier, autonomy                      int
		tags, extra                         []string
	)
	cmd := &cobra.Command{
		Use:               "set <task>",
		Short:             "Set one or more frontmatter fields (validated, single atomic write)",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(c *cobra.Command, args []string) error {
			updates := map[string]any{}
			if c.Flags().Changed("description") {
				updates["description"] = description
			}
			if c.Flags().Changed("priority") {
				updates["priority"] = priority
			}
			if c.Flags().Changed("epic") {
				updates["epic"] = epic
			}
			if c.Flags().Changed("effort") {
				updates["effort"] = effort
			}
			if c.Flags().Changed("tier") {
				updates["tier"] = tier
			}
			if c.Flags().Changed("autonomy") {
				updates["autonomy_level"] = autonomy
			}
			if c.Flags().Changed("tags") {
				updates["tags"] = tags
			}
			for _, kv := range extra {
				k, v, ok := strings.Cut(kv, "=")
				if !ok || k == "" {
					return fmt.Errorf("%w: --set expects key=value, got %q", domain.ErrValidation, kv)
				}
				updates[k] = v
			}
			task, err := app.Svc.SetFields(args[0], updates)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.TaskShowJSON(app.Out, task, "")
			}
			fmt.Fprintf(app.Out, "%s updated %s\n", app.Style.Green("✔"), app.Style.Bold(task.Slug))
			return nil
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "one-line description (<=150 chars)")
	cmd.Flags().StringVar(&priority, "priority", "", "high|medium|low")
	cmd.Flags().StringVar(&epic, "epic", "", "epic id")
	cmd.Flags().StringVar(&effort, "effort", "", "effort estimate")
	cmd.Flags().IntVar(&tier, "tier", 0, "tier 1-5")
	cmd.Flags().IntVar(&autonomy, "autonomy", 0, "autonomy level 1-5")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "comma-separated tags")
	cmd.Flags().StringArrayVar(&extra, "set", nil, "arbitrary key=value (repeatable)")
	return cmd
}

func newTaskMoveCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "move <task>... <status>",
		Short:             "Transition task(s) to <status> (generic escape hatch)",
		Args:              cobra.MinimumNArgs(2),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			to, err := domain.ParseStatus(args[len(args)-1])
			if err != nil {
				return err // already wraps ErrValidation and lists valid statuses
			}
			return runTransition(app, to, args[:len(args)-1])
		},
	}
}

func newTransitionCmd(app *App, use, short string, to domain.Status) *cobra.Command {
	return &cobra.Command{
		Use:               use + " <task>...",
		Short:             short,
		Example:           "  tskflwctl task " + use + " my-task\n  tskflwctl task " + use + " task-a task-b",
		Args:              cobra.MinimumNArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.taskCompleter(to), // don't offer tasks already at `to`
		RunE: func(_ *cobra.Command, args []string) error {
			return runTransition(app, to, args)
		},
	}
}

// runTransition moves each task to status `to`, via the shared runMoves report.
func runTransition(app *App, to domain.Status, slugs []string) error {
	return runMoves(app, slugs, string(to),
		func(slug string) (domain.Task, error) { return app.Svc.Move(slug, to) },
		func(t domain.Task) string { return t.Slug })
}
