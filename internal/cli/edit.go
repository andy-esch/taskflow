package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/domain"
)

// newTaskEditCmd is the human face of mutation: open the whole task file in the
// user's editor and re-validate on save. It complements the agent-facing,
// field-level `task set` — agents drive `set` (deterministic, scriptable); humans
// reach for `edit` to rewrite the body in their editor. The edit is accepted only
// if it still parses (parse-before-accept), so a fat-fingered frontmatter break
// never lands on disk.
func newTaskEditCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "edit <task>",
		Short: "Open a task in your editor (whole file; re-validated on save)",
		Long: "Open the task's markdown file in $VISUAL/$EDITOR (falling back to vi). On\n" +
			"save the file is re-parsed: a frontmatter break (or a value the loader can't\n" +
			"read) reopens the editor with the error rather than landing on disk — deeper\n" +
			"field checks remain `lint`'s job. The human counterpart to `task set`; agents\n" +
			"and scripts should drive `set` (deterministic) instead.",
		Example:           "  tskflwctl task edit add-retry-backoff\n  tskflwctl task edit   # pick from a list",
		Args:              cobra.MaximumNArgs(1), // bare → picker on a TTY; non-interactive needs the slug
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			// Bare `task edit` on a TTY → pick a task; a passed slug short-circuits
			// the picker; non-interactive with no slug → exit 11 (like the verbs).
			value := ""
			if len(args) == 1 {
				value = args[0]
			}
			slug, err := app.fillSelect(value, "specify a task to edit",
				"no tasks available to edit", "Task to edit", app.editOptions)
			if err != nil {
				return err
			}
			// Editing itself needs a terminal for the editor. The gate is the single
			// source of truth for "is a human here?" (TTY in/err, not --json, not
			// --no-input); closed → point agents at the scriptable path.
			if !app.Gate.On() {
				return fmt.Errorf("%w: `task edit` needs an interactive terminal — use `task set` to change fields non-interactively", domain.ErrValidation)
			}
			task, changed, err := app.Svc.EditTask(slug, app.editViaEditor(resolveEditor()))
			if err != nil {
				return err
			}
			if !changed {
				fmt.Fprintln(app.Out, app.Style.Dim("no changes to "+task.Slug))
				return nil
			}
			fmt.Fprintf(app.Out, "%s %s %s\n", app.Style.Green("✔"), "updated", app.Style.Bold(task.Slug))
			// status == directory: editing the frontmatter `status:` can't move a
			// task (and `lint --fix` would revert it). Flag the drift and point at the
			// verb that actually changes status, rather than letting it land silently.
			if task.Misfiled() {
				fmt.Fprintf(app.ErrOut, "%s frontmatter says status: %q but the file is in %q — use `task move`/`task <verb>` to change status\n",
					app.Style.Warn("⚠"), task.Declared, task.Status)
			}
			return nil
		},
	}
}

// resolveEditor picks the editor the way every unix tool does: $VISUAL, then
// $EDITOR, then vi as the last-resort default.
func resolveEditor() string {
	for _, env := range []string{"VISUAL", "EDITOR"} {
		if v := strings.TrimSpace(os.Getenv(env)); v != "" {
			return v
		}
	}
	return "vi"
}

// editViaEditor returns the edit callback the store drives: it writes the content
// to a temp file, runs the editor on it inheriting the terminal, and returns the
// saved bytes. On a prior parse error (a reopen) it prints the error first so the
// loop explains itself. The editor string is split on spaces so $EDITOR="code -w"
// works.
func (app *App) editViaEditor(editor string) func(string, error) (string, error) {
	return func(current string, prevErr error) (string, error) {
		if prevErr != nil {
			fmt.Fprintf(app.ErrOut, "%s %v\n%s\n",
				app.Style.Red("✘ invalid:"), prevErr,
				app.Style.Dim("reopening — fix and save, or save unchanged to cancel"))
		}
		f, err := os.CreateTemp("", "tskflwctl-edit-*.md")
		if err != nil {
			return "", err
		}
		name := f.Name()
		defer func() { _ = os.Remove(name) }()
		if _, err := f.WriteString(current); err != nil {
			_ = f.Close()
			return "", err
		}
		if err := f.Close(); err != nil {
			return "", err
		}
		fields := strings.Fields(editor)
		cmd := exec.Command(fields[0], append(fields[1:], name)...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = app.In, app.Out, app.ErrOut
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("%w: could not run editor %q: %v (set $EDITOR)", domain.ErrValidation, editor, err)
		}
		edited, err := os.ReadFile(name)
		return string(edited), err
	}
}
