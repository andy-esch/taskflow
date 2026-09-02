package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// resolveBody returns the body to use when creating or editing a document: --body
// verbatim, or the contents of --body-file (a path, or "-" for stdin). The two flags
// are mutually exclusive — every command that wires them up calls
// MarkFlagsMutuallyExclusive("body", "body-file") — so at most one is set here. That
// kills the heredoc-in-command-substitution quoting hazard for long bodies. (A new
// body-taking command MUST mark them, or this precedence silently prefers
// --body-file.)
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

// taskVerbHelp is the CLI-specific one-line help for each task lifecycle verb. The
// verb→destination mapping itself lives in the shared registry
// (domain.TaskTransitions()); this is the help text that registry deliberately
// does NOT carry (it's presentation, not vocabulary). Keyed by verb so the two
// can't fall out of step silently — newTaskCmd asserts every registry verb has an
// entry here.
var taskVerbHelp = map[string]string{
	"start":     "Move task(s) to in-progress",
	"next":      "Move task(s) to next-up",
	"ready":     "Move task(s) to ready-to-start",
	"complete":  "Move task(s) to completed",
	"defer":     "Move task(s) to deferred (optionally with a revisit date)",
	"deprecate": "Move task(s) to deprecated",
}

func newTaskCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "task", Short: "Work with tasks"}
	cmd.AddCommand(
		newTaskNewCmd(app),
		newTaskListCmd(app),
		newTaskShowCmd(app),
		newTaskInfoCmd(app),
		newTaskPathCmd(app),
		newTaskAcCmd(app),
		newTaskSetCmd(app),
		newTaskEditCmd(app),
		newTaskAppendCmd(app),
		newTaskRenameCmd(app),
		newTaskMoveCmd(app),
		newTaskDependCmd(app),
		newTaskBlockersCmd(app),
		newTaskUnblocksCmd(app),
	)
	// Explicit transition verbs over the internal move engine (no enum to
	// hallucinate; per-verb intent), built from the shared lifecycle registry so
	// the verb→destination mapping has ONE source the TUI also reads. Each verb
	// NAMES its destination status rather than implying a rank — `next`/`ready`
	// (not promote/demote), so a lateral status change never reads as a value
	// judgment and "leaving deferred" isn't a weird "demote". See the command spec.
	for _, tr := range domain.TaskTransitions() {
		short, ok := taskVerbHelp[tr.Verb]
		if !ok {
			// A new registry verb with no CLI help is a programming error caught the
			// first time the command tree is built (every test, every run).
			panic("cli: no help text for task transition verb " + tr.Verb)
		}
		if tr.Param == domain.ParamOptionalDate {
			// A verb that takes an optional date (defer) has its own builder: it mirrors
			// newTransitionCmd but adds the --until snooze flag, which the core records
			// atomically with the move (one store write — audit M4). The registry's
			// Param flag — not a hardcoded verb name — is what routes us here.
			cmd.AddCommand(newDeferCmd(app))
			continue
		}
		cmd.AddCommand(newTransitionCmd(app, tr.Verb, short, domain.Status(tr.To)))
	}
	// Hidden back-compat: the old hierarchy-flavored verbs still work (and warn)
	// so existing scripts/muscle memory don't hard-break, but they're off the
	// help surface so the dissonant names don't linger in the UI.
	cmd.AddCommand(
		deprecatedTransitionCmd(app, "promote", "next", domain.StatusNextUp),
		deprecatedTransitionCmd(app, "demote", "ready", domain.StatusReadyToStart),
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
					fmt.Sprintf("Description (one line, ≤%d chars)", domain.MaxDescriptionLen), "what & why")
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
				return render.CreatedJSON(app.Out, "task", t.ID, t.Slug, string(t.Status), app.rel(t.Path), app.DryRun, app.workspace())
			}
			render.CreatedHuman(app.Out, app.Style, app.linkPath(t.Path), app.DryRun)
			render.CreatedSlugNote(app.Out, app.Style, p.Title, t.Slug)
			if !app.DryRun {
				fmt.Fprintf(app.Out, "%s\n", app.Style.Dim("→ next: tskflwctl task start "+t.Slug))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Epic, "epic", "", "epic id (required)")
	cmd.Flags().StringVar(&p.Description, "description", "", fmt.Sprintf("one-line description (<=%d chars)", domain.MaxDescriptionLen))
	cmd.Flags().StringVar(&p.Effort, "effort", "Unknown", "effort estimate")
	cmd.Flags().StringVar(&p.Priority, "priority", "medium", "high|medium|low")
	cmd.Flags().IntVar(&p.Tier, "tier", 3, "tier 1-5")
	cmd.Flags().IntVar(&p.Autonomy, "autonomy", 3, "autonomy level 1-5")
	cmd.Flags().StringSliceVar(&p.Tags, "tags", nil, "comma-separated tags (at least one required)")
	cmd.Flags().BoolVar(&p.Next, "next", false, "create in next-up instead of ready-to-start")
	cmd.Flags().BoolVar(&p.Start, "start", false, "create directly in in-progress")
	cmd.Flags().StringVar(&p.Body, "body", "", "override the default body scaffold")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read the body from a file, or - for stdin (replaces --body)")
	cmd.Flags().StringVar(&p.Template, "template", "", `body scaffold to use (default "default"); completes the available names`)
	cmd.MarkFlagsMutuallyExclusive("next", "start")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file", "template")
	// NOTE: --epic is intentionally NOT MarkFlagRequired — newTaskNewCmd resolves
	// it via fillSelect (flag → prompt on a TTY → exit 11 otherwise), so a human
	// gets a picker while agents still get a validation error. cobra's required
	// check would short-circuit that (and exit 1, not our 11).
	_ = cmd.RegisterFlagCompletionFunc("epic", app.completeEpicIDs)
	_ = cmd.RegisterFlagCompletionFunc("template", completeTemplateNames("task"))
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
		Example:     "  tskflwctl task list\n  tskflwctl task list -q --tag tui | xargs tskflwctl task start\n  tskflwctl task list -o table -c slug,status,epic\n  tskflwctl task list --unblocked --json\n  tskflwctl task list --revisit-due -q | xargs tskflwctl task next   # resume snoozed tasks now due",
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
				"tasks", render.TaskColumns(), render.TasksJSON, render.TasksHuman); err != nil {
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
	cmd.Flags().BoolVar(&filter.RevisitDue, "revisit-due", false, "only deferred tasks whose revisit date has arrived (composes with --epic/--tag/-c)")
	cmd.Flags().BoolVar(&filter.Unblocked, "unblocked", false, "only next-up or ready-to-start tasks with a clear dependency gate")
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
	var (
		raw     bool
		section string
		fmOnly  bool
	)
	cmd := &cobra.Command{
		Use:               "show <task>",
		Short:             "Show a task's metadata and body",
		Example:           "  tskflwctl task show add-retry-backoff\n  tskflwctl task show add-retry-backoff --section acceptance\n  tskflwctl task show add-retry-backoff --frontmatter-only",
		Args:              cobra.MaximumNArgs(1), // bare → picker on a TTY; non-interactive needs the slug
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			slug, err := app.resolveOne(args, "specify a task to show", "no tasks available", "Task to show", app.taskOptions)
			if err != nil {
				return err
			}
			task, body, err := app.Svc.ShowTask(slug)
			if err != nil {
				return err
			}
			// --section / --frontmatter-only narrow the body an agent has to read:
			// one named section, or none at all. Both narrow the SAME body the full
			// view emits, so the task metadata (and the --json envelope shape) are
			// unchanged — only Body shrinks.
			body, err = narrowBody("task", slug, body, section, fmOnly)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.TaskShowJSON(app.Out, task, body)
			}
			return app.paged(func(w io.Writer) error {
				if fmOnly { // metadata block only — skip the (empty) body render entirely
					return render.TaskShowHuman(w, app.Style, task, "")
				}
				return render.TaskShowHuman(w, app.Style, task, render.RenderBody(app.Style, body, app.markdownStyle, raw))
			})
		},
	}
	cmd.Flags().BoolVar(&raw, "raw", false, "print the raw markdown body (skip rendering)")
	addBodyScopeFlags(cmd, &section, &fmOnly)
	return cmd
}

// newTaskInfoCmd is the token-cheap metadata read: where the file lives plus the
// triage fields and acceptance-criteria tally, WITHOUT the body `task show`
// carries. `--json` is the machine path (`{path,status,epic,ac:{checked,total}}`);
// the human face is a small aligned block.
func newTaskInfoCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "info <task>",
		Short:             "Show a task's metadata + file path + acceptance tally (no body)",
		Example:           "  tskflwctl task info add-retry-backoff\n  tskflwctl task info add-retry-backoff --json",
		Args:              cobra.MaximumNArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			slug, err := app.resolveOne(args, "specify a task", "no tasks available", "Task", app.taskOptions)
			if err != nil {
				return err
			}
			task, body, err := app.Svc.ShowTask(slug)
			if err != nil {
				return err
			}
			ac := domain.CountAcceptanceCriteria(body)
			path := absPath(task.Path)
			if app.JSON {
				return render.TaskInfoJSON(app.Out, task, ac, path)
			}
			render.TaskInfoHuman(app.Out, app.Style, task, ac, path)
			return nil
		},
	}
}

// newTaskPathCmd prints just the absolute path to a task's file — the minimal,
// pipe-friendly accessor (`$EDITOR "$(tskflwctl task path x)"`) that replaces
// globbing `find` on the id-led `<id>-<slug>.md` filename. It resolves the path
// WITHOUT parsing (Svc.TaskPath), so it still works on a file with broken
// frontmatter — exactly when you need the path to go fix it. `--json` wraps it so
// the "schema_version everywhere" contract holds; plain prints the bare path.
func newTaskPathCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "path <task>",
		Short:             "Print the absolute path to a task's file",
		Example:           "  tskflwctl task path add-retry-backoff\n  $EDITOR \"$(tskflwctl task path add-retry-backoff)\"",
		Args:              cobra.MaximumNArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			slug, err := app.resolveOne(args, "specify a task", "no tasks available", "Task", app.taskOptions)
			if err != nil {
				return err
			}
			p, err := app.Svc.TaskPath(slug)
			if err != nil {
				return err
			}
			return emitPath(app, absPath(p))
		},
	}
}

// emitPath writes a resolved file path: the `path` --json envelope, or the bare
// path for piping. Shared by task/epic/audit path.
func emitPath(app *App, path string) error {
	if app.JSON {
		return render.PathJSON(app.Out, path)
	}
	fmt.Fprintln(app.Out, path)
	return nil
}

// joinIndices renders an index list for a status line: "3" for one, "1, 2 and 4" for
// several, so the message reads as a sentence rather than a slice literal.
func joinIndices(ns []int) string {
	parts := make([]string, len(ns))
	for i, n := range ns {
		parts[i] = strconv.Itoa(n)
	}
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	default:
		return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
	}
}

// newTaskAcCmd lists a task's acceptance criteria, or flips one by index — the CLI
// for close-out edits that otherwise force a hand-edit of `- [ ]` → `- [x]`.
// Index-based (`--list` to number them, then `--check 3`) is the robust form;
// substring matching is deliberately not offered. A flip goes through the atomic,
// frontmatter-preserving body-replace path and returns the task_mutation envelope.
func newTaskAcCmd(app *App) *cobra.Command {
	// --check/--uncheck take a LIST of indices: closing out several criteria at once is
	// the common shape at the end of a piece of work, and one flag call is one atomic
	// write instead of one per criterion. The four state flags below stay single-index —
	// each needs its own --reason, and one reason spread across several criteria would be
	// a different (and less honest) feature.
	var check, uncheck []int
	// One flag per state rather than a --state <word> pair, matching --check/--uncheck: the
	// index is the argument, the flag names the destination. Same shape as the lifecycle
	// verbs, where the verb names where the task is going.
	var defer_, wontfix, tracked, na int
	var add, replaceText string
	var remove, replace int
	var reason string
	var list bool
	cmd := &cobra.Command{
		Use:   "ac <task>",
		Short: "List a task's acceptance criteria, or check/uncheck one by index",
		Long: "List a task's acceptance criteria — the checkboxes under its " +
			"`## Acceptance criteria` section — or flip them by 1-based index. Run with no " +
			"flags (or --list) to number them, then --check <n> / --uncheck <n> to tick or " +
			"clear them. Both take a LIST: --check 1,2,4 (or --check 1 --check 2) flips all " +
			"of them in ONE atomic write, so closing out several criteria costs one file " +
			"rewrite and one updated_at bump rather than one per criterion. Indices are " +
			"deduplicated and order-independent. Matching is index-based, not substring, for " +
			"robustness. A flip rewrites only those checkboxes (the rest of the file is " +
			"preserved) and is idempotent — flipping to the current state writes nothing. " +
			"Checkboxes in fenced code blocks are ignored, and a missing section or an " +
			"out-of-range index is a validation error (exit 11) that rejects the whole " +
			"request before writing, so a bad index never leaves a half-applied body.\n\n" +
			"The criteria themselves can be edited too: --add <text> appends one, --remove <n> " +
			"deletes one, and --replace <n> --text <new> rewords one. A reworded criterion KEEPS " +
			"its checkbox and any state suffix — rewording is not a change of mind, and silently " +
			"dropping a `wontfix` and its reason would lose a decision. Added and reworded text " +
			"is wrapped to match the corpus. --add needs an existing `## Acceptance criteria` " +
			"section: creating one would mean guessing where it belongs in a body the tool did " +
			"not write.",
		Example:           "  tskflwctl task ac add-retry-backoff               # numbered list\n  tskflwctl task ac add-retry-backoff --check 3     # tick criterion 3\n  tskflwctl task ac add-retry-backoff --check 1,2,4 # tick three, one atomic write\n  tskflwctl task ac add-retry-backoff --uncheck 1,3\n  tskflwctl task ac add-retry-backoff --defer 2 --reason \"waiting on the schema ADR\"\n  tskflwctl task ac add-retry-backoff --add \"Retries stop at the configured ceiling\"\n  tskflwctl task ac add-retry-backoff --replace 3 --text \"Backoff is jittered\"\n  tskflwctl task ac add-retry-backoff --remove 4",
		Args:              cobra.MaximumNArgs(1),
		Annotations:       map[string]string{"safety": "mutating"}, // --check/--uncheck write; --list reads
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(c *cobra.Command, args []string) error {
			slug, err := app.resolveOne(args, "specify a task", "no tasks available", "Task", app.taskOptions)
			if err != nil {
				return err
			}
			// Exactly one destination flag may be given: they are alternatives, and silently
			// honouring the first would make a two-flag typo write something unasked-for.
			var chosen []string
			for _, f := range []string{"check", "uncheck", "defer", "wontfix", "tracked", "na", "add", "remove", "replace"} {
				if c.Flags().Changed(f) {
					chosen = append(chosen, "--"+f)
				}
			}
			if len(chosen) > 1 {
				return fmt.Errorf("%w: %s ask for different things; pass one", domain.ErrValidation, strings.Join(chosen, " and "))
			}
			// --text carries --replace's payload and means nothing alone. Falling through to
			// the list view would print criteria and report success while writing nothing —
			// the silent-no-op shape a mistyped flag name produces.
			if c.Flags().Changed("text") && !c.Flags().Changed("replace") {
				return fmt.Errorf("%w: --text is the new wording for --replace <n>; pass --replace too", domain.ErrValidation)
			}
			// The three that change WHICH criteria exist, rather than what one of them says.
			// They share the body-edit path; the list is reprinted afterwards because add
			// and remove renumber everything below them.
			if edit := criteriaEdit(chosen, add, remove, replace, replaceText); edit != nil {
				task, body, changed, err := app.Svc.EditCriteria(slug, app.DryRun, edit)
				if err != nil {
					return err
				}
				if !changed && !app.JSON {
					fmt.Fprintf(app.Out, "%s no change\n", app.Style.Dim("•"))
					return nil
				}
				return reportTaskMutation(app, task, body,
					strings.TrimPrefix(chosen[0], "--")+" acceptance criterion in",
					"would "+strings.TrimPrefix(chosen[0], "--")+" acceptance criterion in")
			}
			// No destination flag → the list view (the default; --list is explicit).
			if len(chosen) == 0 {
				canon, cs, err := app.Svc.AcceptanceCriteria(slug)
				if err != nil {
					return err
				}
				if app.JSON {
					return render.AcceptanceJSON(app.Out, canon, cs)
				}
				render.AcceptanceHuman(app.Out, app.Style, cs)
				return nil
			}
			idxs, state := check, domain.CriterionMet
			switch chosen[0] {
			case "--uncheck":
				idxs, state = uncheck, domain.CriterionUnmet
			case "--defer":
				idxs, state = []int{defer_}, domain.CriterionDeferred
			case "--wontfix":
				idxs, state = []int{wontfix}, domain.CriterionWontFix
			case "--tracked":
				idxs, state = []int{tracked}, domain.CriterionTracked
			case "--na":
				idxs, state = []int{na}, domain.CriterionNA
			}
			task, body, changed, err := app.Svc.SetCriteriaState(slug, idxs, state, reason, app.DryRun)
			if err != nil {
				return err
			}
			what := "criterion " + joinIndices(idxs)
			if len(idxs) > 1 {
				what = "criteria " + joinIndices(idxs)
			}
			if !changed && !app.JSON { // already in the target state — say so, no write
				verb := "is"
				if len(idxs) > 1 {
					verb = "are"
				}
				fmt.Fprintf(app.Out, "%s %s %s already %s\n", app.Style.Dim("•"), what, verb, state)
				return nil
			}
			return reportTaskMutation(app, task, body,
				"set "+what+" "+string(state),
				"would set "+what+" "+string(state))
		},
	}
	cmd.Flags().BoolVar(&list, "list", false, "list the acceptance criteria (the default)")
	cmd.Flags().IntSliceVar(&check, "check", nil, "check the criteria at these 1-based `indices` (comma-separated or repeatable)")
	cmd.Flags().IntSliceVar(&uncheck, "uncheck", nil, "uncheck the criteria at these 1-based `indices` (comma-separated or repeatable)")
	cmd.Flags().IntVar(&defer_, "defer", 0, "mark the criterion at this 1-based index deferred (needs --reason)")
	cmd.Flags().IntVar(&wontfix, "wontfix", 0, "mark the criterion at this 1-based index wontfix (needs --reason)")
	cmd.Flags().IntVar(&tracked, "tracked", 0, "mark the criterion at this 1-based index tracked — handed to another task (needs --reason naming it)")
	cmd.Flags().IntVar(&na, "na", 0, "mark the criterion at this 1-based index n/a — no longer applies (needs --reason)")
	cmd.Flags().StringVar(&add, "add", "", "append a new unchecked criterion with this `text`")
	cmd.Flags().IntVar(&remove, "remove", 0, "delete the criterion at this 1-based index")
	cmd.Flags().IntVar(&replace, "replace", 0, "reword the criterion at this 1-based index (needs --text)")
	cmd.Flags().StringVar(&replaceText, "text", "", "the new wording for --replace")
	cmd.Flags().StringVar(&reason, "reason", "", "why the criterion is deferred/wontfix/tracked/n-a — required for those, rejected otherwise")
	cmd.MarkFlagsMutuallyExclusive("check", "uncheck")
	cmd.MarkFlagsMutuallyExclusive("list", "check")
	cmd.MarkFlagsMutuallyExclusive("list", "uncheck")
	// --list is the read view; every writing flag contradicts it. Registered here as well
	// as caught by the `chosen` check so cobra reports the conflict before RunE.
	for _, f := range []string{"defer", "wontfix", "tracked", "na", "add", "remove", "replace"} {
		cmd.MarkFlagsMutuallyExclusive("list", f)
	}
	return cmd
}

// absPath makes a store path absolute so `task path`/`task info` emit a path that
// resolves from anywhere, regardless of how the planning root was configured
// (relative config root, -C, etc.). A failure to absolutize (never expected for a
// real file) falls back to the store path rather than erroring a read.
func absPath(p string) string {
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
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
		Use:   "set <task>",
		Short: "Set one or more frontmatter fields (validated, single atomic write)",
		Long: "Set one or more task frontmatter fields in a single validated atomic write.\n" +
			"Graph-owned dependency fields (depends_on and the legacy blocked_by, dependencies,\n" +
			"and blocks fields) cannot be changed or removed here, including with --force.\n" +
			"They require the guarded dependency operations introduced by the dependency roadmap.",
		Example:           "  tskflwctl task set add-retry-backoff --priority high\n  tskflwctl task set --priority high   # pick the task from a list",
		Args:              cobra.MaximumNArgs(1), // bare → picker on a TTY; non-interactive needs the slug
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(c *cobra.Command, args []string) error {
			slug, err := app.resolveOne(args, "specify a task to set", "no tasks available", "Task to set", app.taskOptions)
			if err != nil {
				return err
			}
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
				task, newBody, err := app.Svc.ReplaceBody(slug, text, app.DryRun)
				if err != nil {
					return err
				}
				return reportTaskMutation(app, task, newBody, "updated", "would update")
			}
			task, err := app.Svc.SetFields(slug, updates, force, app.DryRun)
			if err != nil {
				return err
			}
			return reportTaskMutation(app, task, "", "updated", "would update")
		},
	}
	cmd.Flags().StringVar(&description, "description", "", fmt.Sprintf("one-line description (<=%d chars)", domain.MaxDescriptionLen))
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
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	_ = cmd.RegisterFlagCompletionFunc("epic", app.completeEpicIDs)
	return cmd
}

// reportTaskMutation writes the standard task-mutation result: the task_mutation
// JSON envelope (carrying dry_run + the resulting body) under --json, else a styled
// one-line confirmation. body is "" for field-only `set`. verb/dryVerb let the
// caller phrase the action ("updated"/"would update", "appended to"/…).
func reportTaskMutation(app *App, task domain.Task, body, verb, dryVerb string) error {
	if app.JSON {
		return render.TaskMutationJSON(app.Out, task, body, app.DryRun, app.workspace())
	}
	if app.DryRun {
		verb = dryVerb
	}
	fmt.Fprintf(app.Out, "%s %s %s\n", app.Style.Green("✔"), verb, app.Style.Bold(task.Slug))
	return nil
}

// newTaskRenameCmd is the Scheme-2 `rename` verb: re-title a task (a new slug from the
// title; the 12-char id — the stable key — is kept), rewrite the body H1, and cascade
// every inbound relative-path markdown link across the planning tree to the new filename.
func newTaskRenameCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "rename <task> <new-title>",
		Short:       "Re-title a task (new slug, id kept) and cascade its inbound body links",
		Args:        cobra.ExactArgs(2),
		Annotations: map[string]string{"safety": "mutating"},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return app.completeTaskSlugs(cmd, args, toComplete)
			}
			return nil, cobra.ShellCompDirectiveNoFileComp // the title is free text
		},
		RunE: func(_ *cobra.Command, args []string) error {
			task, cascade, err := app.Svc.RenameTask(args[0], args[1], app.DryRun)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.TaskMutationJSON(app.Out, task, "", app.DryRun, app.workspace())
			}
			verb := "renamed to"
			if app.DryRun {
				verb = "would rename to"
			}
			fmt.Fprintf(app.Out, "%s %s %s", app.Style.Green("✔"), verb, app.Style.Bold(task.Slug))
			if cascade > 0 {
				fmt.Fprintf(app.Out, " %s", app.Style.Dim(fmt.Sprintf("(%d inbound link(s) repointed)", cascade)))
			}
			fmt.Fprintln(app.Out)
			return nil
		},
	}
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
		// A heredoc leads, deliberately: planning bodies are full of percent signs
		// (coverage, gates, metrics) and printf reads them as format verbs, so the
		// obvious multi-line spelling silently truncates the write.
		Example:           "  tskflwctl task append my-task --body 'a one-line note'\n  tskflwctl task append my-task --body-file - <<'EOF'\n## Review\n- coverage held at 82% after the change\nEOF\n  cat notes.md | tskflwctl task append my-task --body-file -",
		Args:              cobra.MaximumNArgs(1), // bare → picker on a TTY; non-interactive needs the slug
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
			slug, err := app.resolveOne(args, "specify a task to append to", "no tasks available", "Task to append to", app.taskOptions)
			if err != nil {
				return err
			}
			task, newBody, err := app.Svc.AppendBody(slug, text, app.DryRun)
			if err != nil {
				return err
			}
			return reportTaskMutation(app, task, newBody, "appended to", "would append to")
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "markdown to append")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read the markdown to append from a file (or - for stdin)")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	return cmd
}

func newTaskMoveCmd(app *App) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
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
			return runTransition(app, to, args[:len(args)-1], force)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "override the destination's contextual gate (dependencies for in-progress; acceptance criteria for completed)")
	return cmd
}

// criteriaEdit maps the structural flags to their domain operation, or nil when the call is
// a state change or a plain list. Kept beside the command rather than inside it so the
// mutual-exclusion check above stays the single place that decides WHICH flag won.
func criteriaEdit(chosen []string, add string, remove, replace int, text string) func(string) (string, error) {
	if len(chosen) != 1 {
		return nil
	}
	switch chosen[0] {
	case "--add":
		return func(b string) (string, error) { return domain.AddCriterion(b, add) }
	case "--remove":
		return func(b string) (string, error) { return domain.RemoveCriterion(b, remove) }
	case "--replace":
		return func(b string) (string, error) {
			if strings.TrimSpace(text) == "" {
				return "", fmt.Errorf("%w: --replace needs the new wording — pass --text", domain.ErrValidation)
			}
			return domain.ReplaceCriterionText(b, replace, text)
		}
	}
	return nil
}

func newTransitionCmd(app *App, use, short string, to domain.Status) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
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
			return runTransition(app, to, args, force)
		},
	}
	switch to {
	case domain.StatusInProgress:
		cmd.Long = short + ".\n\n" +
			"Refuses a task unless it is next-up or ready-to-start with a clear dependency gate.\n" +
			"--force bypasses only a blocked dependency gate; it does not bypass another\n" +
			"lifecycle role or broken evidence, and the receipt names every outstanding blocker."
		cmd.Flags().BoolVar(&force, "force", false, "start despite outstanding dependency blockers")
	case domain.StatusCompleted:
		cmd.Long = short + ".\n\n" +
			"Refuses a task whose acceptance criteria are still unmet with no reason given —\n" +
			"the task counterpart of `audit close` refusing while findings are open. A criterion\n" +
			"carrying a state (`task ac --defer|--wontfix|--tracked|--na`) has been DECIDED and\n" +
			"does not block; only a silently unticked box does. --force completes anyway."
		cmd.Flags().BoolVar(&force, "force", false, "complete even with unmet, unexplained acceptance criteria")
	}
	return cmd
}

// deprecatedTransitionCmd builds a hidden back-compat alias for a renamed verb:
// it behaves exactly like newTransitionCmd(oldVerb) but is hidden from help and
// prints cobra's deprecation notice pointing at the new name. Lets `task promote`
// keep working (with a nudge) after the rename to `task next`/`task ready`.
func deprecatedTransitionCmd(app *App, oldVerb, newVerb string, to domain.Status) *cobra.Command {
	cmd := newTransitionCmd(app, oldVerb, "Move task(s) to "+string(to), to)
	cmd.Hidden = true
	cmd.Deprecated = "use `task " + newVerb + "` (lifecycle verbs now name the destination status)"
	return cmd
}

// runTransition moves each task to status `to`, via the shared runMoves report.
func runTransition(app *App, to domain.Status, slugs []string, force bool) error {
	return runMoves(app, slugs, string(to),
		func(slug string) (core.TaskLifecycleReceipt, error) {
			return app.Svc.Move(slug, to, app.DryRun, lifecycleOverride(to, force))
		},
		func(receipt core.TaskLifecycleReceipt) string { return receipt.Task.Slug },
		func(receipt core.TaskLifecycleReceipt, result *render.MoveResult) {
			*result = render.TaskMoveResult(receipt)
		})
}

func lifecycleOverride(to domain.Status, force bool) core.TaskLifecycleOverride {
	if !force {
		return core.TaskLifecycleOverrideNone
	}
	if to == domain.StatusCompleted {
		return core.TaskLifecycleOverrideAcceptanceCriteria
	}
	return core.TaskLifecycleOverrideDependencyGate
}

// newDeferCmd mirrors newTransitionCmd (bare verb → picker, ArbitraryArgs, the
// move) but adds the optional --until snooze date: when set, each task is moved
// to deferred AND has revisit_at recorded, so `status` can nudge you when the date
// arrives. On a TTY without --until it brings up a separate prompt to choose a
// revisit date (an absolute date or a relative offset like 2w/10d), so a human
// deferring interactively is offered a snooze without remembering the flag; blank
// skips it (park indefinitely), and off a TTY there's no prompt — exactly the old
// agent `task defer`. The date is validated up front (a bad --until errors before
// anything moves) and recorded in the SAME atomic write as the move (audit M4).
func newDeferCmd(app *App) *cobra.Command {
	var until string
	to := domain.StatusDeferred
	cmd := &cobra.Command{
		Use:   "defer <task>...",
		Short: "Move task(s) to deferred (optionally with a revisit date)",
		Example: "  tskflwctl task defer my-task                      # on a TTY, prompts for a revisit date\n" +
			"  tskflwctl task defer my-task --until 2026-09-01   # snooze until a date\n" +
			"  tskflwctl task defer task-a task-b",
		Args:              cobra.ArbitraryArgs, // bare verb → picker on a TTY; non-interactive needs ≥1 arg
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.taskCompleter(to), // don't offer tasks already deferred
		RunE: func(c *cobra.Command, args []string) error {
			changed := c.Flags().Changed("until")
			// Validate an explicit --until BEFORE moving anything — a bad date must
			// fail fast (exit 11) and leave every task where it is.
			if changed {
				if err := domain.ValidateDate(until); err != nil {
					return err
				}
			}
			if len(args) == 0 {
				slug, err := app.fillSelect("", "specify at least one task to defer",
					"no tasks available to defer", "Task to defer", app.transitionOptions(to))
				if err != nil {
					return err
				}
				args = []string{slug}
			}
			// After the task is chosen, offer a revisit date on a TTY (no-op off a
			// TTY or when --until was given), so the same value drives every slug.
			revisit, err := app.fillRevisitDate(changed, until, app.Svc.Now())
			if err != nil {
				return err
			}
			return runMoves(app, args, string(to),
				func(slug string) (core.TaskLifecycleReceipt, error) {
					return app.Svc.DeferTask(slug, revisit, app.DryRun)
				},
				func(receipt core.TaskLifecycleReceipt) string { return receipt.Task.Slug },
				func(receipt core.TaskLifecycleReceipt, result *render.MoveResult) {
					*result = render.TaskMoveResult(receipt)
				})
		},
	}
	cmd.Flags().StringVar(&until, "until", "", "revisit date YYYY-MM-DD (snooze until); records revisit_at on each task")
	return cmd
}
