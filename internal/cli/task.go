package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// resolveBody returns the body to use when creating a document: --body verbatim,
// or the contents of --body-file (a path, or "-" for stdin). The two flags are
// mutually exclusive (enforced by the command), so at most one is set — this
// kills the heredoc-in-command-substitution quoting hazard for long bodies.
func resolveBody(cmd *cobra.Command, body, bodyFile string) (string, error) {
	if bodyFile == "" {
		return body, nil
	}
	if bodyFile == "-" {
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return "", fmt.Errorf("read body from stdin: %w", err)
		}
		return string(data), nil
	}
	data, err := os.ReadFile(bodyFile)
	if err != nil {
		return "", fmt.Errorf("%w: read --body-file: %v", domain.ErrValidation, err)
	}
	return string(data), nil
}

func newTaskCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "task", Short: "Work with tasks"}
	cmd.AddCommand(
		newTaskNewCmd(app),
		newTaskListCmd(app),
		newTaskShowCmd(app),
		newTaskSetCmd(app),
		newTaskEditCmd(app),
		newTaskAppendCmd(app),
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
	var (
		p        core.NewTaskParams
		bodyFile string
	)
	cmd := &cobra.Command{
		Use:               "new <title>",
		Short:             "Create a new task (validated, handoff-ready scaffold)",
		Example:           "  tskflwctl task new \"Add retry backoff\" --epic 17-pm-go-cli --tags net\n  tskflwctl task new \"Triage flaky test\" --epic 17-pm-go-cli --next",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: activeHelpArg("provide a task title (quote it if it has spaces)"),
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Title = args[0]
			// epic: flag value, else prompt (interactive), else exit 11. The
			// required-input rule is unchanged for agents — this just adds a
			// human picker when a TTY is present.
			epic, err := app.fillSelect(p.Epic, "--epic is required",
				"no epics exist yet — create one with 'epic new' first", "Epic for this task", app.epicOptions)
			if err != nil {
				return err
			}
			p.Epic = epic
			// tags (≥1 required): flag values → free-form text prompt on a TTY →
			// exit 11 otherwise.
			tags, err := app.fillTags(p.Tags, app.tagHint)
			if err != nil {
				return err
			}
			p.Tags = tags
			// A next-up/in-progress task requires a description (the L4 rule); on a
			// TTY prompt for it, otherwise exit 11 — same flag-twin contract.
			if p.Next || p.Start {
				desc, err := app.fillText(p.Description,
					"--description is required for a --next/--start task",
					"Description (one line, ≤150 chars)", "what & why")
				if err != nil {
					return err
				}
				p.Description = desc
			}
			body, err := resolveBody(cmd, p.Body, bodyFile)
			if err != nil {
				return err
			}
			p.Body = body
			p.DryRun = app.DryRun
			t, err := app.Svc.NewTask(p)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.CreatedJSON(app.Out, "task", t.Slug, string(t.Status), app.rel(t.Path), app.DryRun)
			}
			render.CreatedHuman(app.Out, app.Style, app.linkPath(t.Path), app.DryRun)
			if !app.DryRun {
				fmt.Fprintf(app.Out, "%s\n", app.Style.Dim("→ next: tskflwctl task start "+t.Slug))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Epic, "epic", "", "epic id (required)")
	cmd.Flags().StringVar(&p.Description, "description", "", "one-line description (<=150 chars)")
	cmd.Flags().StringVar(&p.Effort, "effort", "Unknown", "effort estimate")
	cmd.Flags().StringVar(&p.Priority, "priority", "medium", "high|medium|low")
	cmd.Flags().IntVar(&p.Tier, "tier", 3, "tier 1-5")
	cmd.Flags().IntVar(&p.Autonomy, "autonomy", 3, "autonomy level 1-5")
	cmd.Flags().StringSliceVar(&p.Tags, "tags", nil, "comma-separated tags (at least one required)")
	cmd.Flags().BoolVar(&p.Next, "next", false, "create in next-up instead of ready-to-start")
	cmd.Flags().BoolVar(&p.Start, "start", false, "create directly in in-progress")
	cmd.Flags().StringVar(&p.Body, "body", "", "override the default body scaffold")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read the body from a file, or - for stdin (replaces --body)")
	cmd.MarkFlagsMutuallyExclusive("next", "start")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	// NOTE: --epic is intentionally NOT MarkFlagRequired — newTaskNewCmd resolves
	// it via fillSelect (flag → prompt on a TTY → exit 11 otherwise), so a human
	// gets a picker while agents still get a validation error. cobra's required
	// check would short-circuit that (and exit 1, not our 11).
	_ = cmd.RegisterFlagCompletionFunc("epic", app.completeEpicIDs)
	return cmd
}

func newTaskListCmd(app *App) *cobra.Command {
	var (
		filter core.TaskFilter
		lm     listMode
	)
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List tasks (active by default)",
		Example:     "  tskflwctl task list\n  tskflwctl task list -q --tag tui | xargs tskflwctl task start\n  tskflwctl task list -o table -c slug,status,epic",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			mode, err := lm.resolve(cmd, app)
			if err != nil {
				return err
			}
			tasks, problems, err := app.Svc.ListTasks(filter)
			if err != nil {
				return err
			}
			if err := renderList(app, mode, lm.columns, tasks, problems,
				render.TaskColumns(), render.TasksJSON, render.TasksHuman); err != nil {
				return err
			}
			return problemsError(problems)
		},
	}
	lm.bind(cmd, render.Specs(render.TaskColumns()))
	cmd.Flags().StringVar(&filter.Status, "status", "", "filter by status")
	cmd.Flags().StringVar(&filter.Epic, "epic", "", "filter by epic")
	cmd.Flags().StringVar(&filter.Tag, "tag", "", "filter by tag")
	cmd.Flags().BoolVar(&filter.All, "all", false, "include completed/deprecated/deferred")
	_ = cmd.RegisterFlagCompletionFunc("status", completeStatusValues)
	_ = cmd.RegisterFlagCompletionFunc("epic", app.completeEpicIDs)
	return cmd
}

// completeStatusValues offers the closed status set for a --status flag.
func completeStatusValues(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	opts := make([]string, 0, len(domain.AllStatuses()))
	for _, st := range domain.AllStatuses() {
		opts = append(opts, string(st))
	}
	return opts, cobra.ShellCompDirectiveNoFileComp
}

func newTaskShowCmd(app *App) *cobra.Command {
	var raw bool
	cmd := &cobra.Command{
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
			return render.TaskShowHuman(app.Out, app.Style, task, render.RenderBody(app.Style, body, app.markdownStyle(), raw))
		},
	}
	cmd.Flags().BoolVar(&raw, "raw", false, "print the raw markdown body (skip rendering)")
	return cmd
}

func newTaskSetCmd(app *App) *cobra.Command {
	var (
		description, priority, epic, effort string
		tier, autonomy                      int
		tags, extra, unsets                 []string
		body, bodyFile                      string
		force                               bool
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
			for _, k := range unsets {
				if _, dup := updates[k]; dup {
					return fmt.Errorf("%w: %q is both set and unset", domain.ErrValidation, k)
				}
				updates[k] = domain.UnsetField{}
			}
			// --body/--body-file replace the markdown body (the agent face of
			// editing) — its own atomic write, not mixed with field surgery.
			if c.Flags().Changed("body") || c.Flags().Changed("body-file") {
				if len(updates) > 0 {
					return fmt.Errorf("%w: --body/--body-file can't be combined with field flags — set the body in its own call", domain.ErrValidation)
				}
				text, err := resolveBody(c, body, bodyFile)
				if err != nil {
					return err
				}
				if strings.TrimSpace(text) == "" {
					return fmt.Errorf("%w: --body is empty (nothing to write)", domain.ErrValidation)
				}
				task, newBody, err := app.Svc.ReplaceBody(args[0], text, app.DryRun)
				if err != nil {
					return err
				}
				return reportTaskMutation(app, task, newBody, "updated", "would update")
			}
			task, err := app.Svc.SetFields(args[0], updates, force, app.DryRun)
			if err != nil {
				return err
			}
			return reportTaskMutation(app, task, "", "updated", "would update")
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "one-line description (<=150 chars)")
	cmd.Flags().StringVar(&priority, "priority", "", "high|medium|low")
	cmd.Flags().StringVar(&epic, "epic", "", "epic id")
	cmd.Flags().StringVar(&effort, "effort", "", "effort estimate")
	cmd.Flags().IntVar(&tier, "tier", 0, "tier 1-5")
	cmd.Flags().IntVar(&autonomy, "autonomy", 0, "autonomy level 1-5")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "comma-separated tags")
	cmd.Flags().StringArrayVar(&extra, "set", nil,
		"key=value (repeatable); known fields are typed+validated, unknown keys need --force")
	cmd.Flags().StringArrayVar(&unsets, "unset", nil, "remove a frontmatter key (repeatable)")
	cmd.Flags().BoolVar(&force, "force", false, "allow --set of a field tskflwctl doesn't know")
	cmd.Flags().StringVar(&body, "body", "", "replace the markdown body (its own call — not combined with field flags)")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "replace the markdown body from a file (or - for stdin)")
	_ = cmd.RegisterFlagCompletionFunc("epic", app.completeEpicIDs)
	return cmd
}

// reportTaskMutation writes the standard task-mutation result: the task_mutation
// JSON envelope (carrying dry_run + the resulting body) under --json, else a styled
// one-line confirmation. body is "" for field-only `set`. verb/dryVerb let the
// caller phrase the action ("updated"/"would update", "appended to"/…).
func reportTaskMutation(app *App, task domain.Task, body, verb, dryVerb string) error {
	if app.JSON {
		return render.TaskMutationJSON(app.Out, task, body, app.DryRun)
	}
	if app.DryRun {
		verb = dryVerb
	}
	fmt.Fprintf(app.Out, "%s %s %s\n", app.Style.Green("✔"), verb, app.Style.Bold(task.Slug))
	return nil
}

// newTaskAppendCmd is the scriptable counterpart to `task edit`: append markdown
// to a task's body in one atomic, validated write, from --body/--body-file/stdin.
func newTaskAppendCmd(app *App) *cobra.Command {
	var body, bodyFile string
	cmd := &cobra.Command{
		Use:   "append <task>",
		Short: "Append a section to a task's body (atomic; agent-facing)",
		Long: "Append markdown to the end of a task's body in one atomic, validated write —\n" +
			"the scriptable counterpart to `task edit`. Content comes from --body, --body-file,\n" +
			"or stdin (--body-file -); a blank line separates it from the existing body.",
		// --body is one line as typed; multi-line content comes from a file or stdin
		// (a shell passes "\n" inside --body literally, it is not a newline).
		Example:           "  tskflwctl task append my-task --body 'a one-line note'\n  printf '## Review\\n- looks good\\n' | tskflwctl task append my-task --body-file -",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(c *cobra.Command, args []string) error {
			text, err := resolveBody(c, body, bodyFile)
			if err != nil {
				return err
			}
			if strings.TrimSpace(text) == "" {
				return fmt.Errorf("%w: nothing to append (provide --body, --body-file, or stdin via -)", domain.ErrValidation)
			}
			task, newBody, err := app.Svc.AppendBody(args[0], text, app.DryRun)
			if err != nil {
				return err
			}
			return reportTaskMutation(app, task, newBody, "appended to", "would append to")
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "markdown to append")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read the markdown to append from a file (or - for stdin)")
	return cmd
}

func newTaskMoveCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "move <task>... <status>",
		Short:       "Transition task(s) to <status> (generic escape hatch)",
		Args:        cobra.MinimumNArgs(2),
		Annotations: map[string]string{"safety": "mutating"},
		// Position-aware: the final arg is a status from a small closed set —
		// offering task slugs there (the old behavior) actively misled.
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			opts, directive := app.completeTaskSlugs(cmd, args, toComplete)
			if len(args) >= 1 {
				for _, st := range domain.AllStatuses() {
					opts = append(opts, string(st))
				}
				if cobra.GetActiveHelpConfig(cmd) != "off" {
					opts = cobra.AppendActiveHelp(opts, "the final argument is the target status")
				}
			}
			return opts, directive
		},
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
		Args:              cobra.ArbitraryArgs, // bare verb → picker on a TTY; non-interactive needs ≥1 arg
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.taskCompleter(to), // don't offer tasks already at `to`
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Bare verb: pick a task on a TTY; non-interactive → exit 11.
				slug, err := app.fillSelect("", "specify at least one task to "+use,
					"no tasks available to "+use, "Task to "+use, app.transitionOptions(to))
				if err != nil {
					return err
				}
				args = []string{slug}
			}
			return runTransition(app, to, args)
		},
	}
}

// runTransition moves each task to status `to`, via the shared runMoves report.
func runTransition(app *App, to domain.Status, slugs []string) error {
	return runMoves(app, slugs, string(to),
		func(slug string) (domain.Task, error) { return app.Svc.Move(slug, to, app.DryRun) },
		func(t domain.Task) string { return t.Slug })
}
