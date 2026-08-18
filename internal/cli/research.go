package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/editor"
)

// The research command surface is deliberately smaller than task/audit: there are no
// lifecycle verbs, because research has no status (epic 28). It does carry the two faces
// of mutation every other entity has — agent (`set`, `append`: field/body level,
// scriptable, atomic) and human (`edit`: $EDITOR on the whole file, re-validated on save).
func newResearchCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "research", Short: "Work with research docs"}
	cmd.AddCommand(
		newResearchNewCmd(app),
		newResearchListCmd(app),
		newResearchShowCmd(app),
		newResearchPathCmd(app),
		newResearchSetCmd(app),
		newResearchEditCmd(app),
		newResearchAppendCmd(app),
	)
	return cmd
}

func newResearchSetCmd(app *App) *cobra.Command {
	var (
		description         string
		tags, extra, unsets []string
		force               bool
	)
	cmd := &cobra.Command{
		Use:   "set <research>",
		Short: "Set one or more frontmatter fields (validated, single atomic write)",
		Long: "Update a research doc's frontmatter in one atomic, validated write. Unknown keys,\n" +
			"comments, and key order are preserved.\n\n" +
			"`created` cannot be set: the stable id is minted from it, so changing one would\n" +
			"desync the pair and break the id-order-is-date-order property. Re-dating a doc\n" +
			"means creating a new one.",
		Example:           "  tskflwctl research set theming-libs --description \"Weighed three TUI theming libs\"\n  tskflwctl research set theming-libs --tags tui,color",
		Args:              cobra.MaximumNArgs(1), // bare → picker on a TTY; non-interactive needs the slug
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.completeResearchSlugs,
		RunE: func(c *cobra.Command, args []string) error {
			slug, err := app.resolveOne(args, "specify a research doc to set", "no research docs available", "Research doc to set", app.researchOptions)
			if err != nil {
				return err
			}
			updates := map[string]any{}
			if c.Flags().Changed("description") {
				updates["description"] = description
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
			r, err := app.Svc.SetResearchFields(slug, updates, force, app.DryRun)
			if err != nil {
				return err
			}
			return reportResearchMutation(app, r, "", "updated", "would update")
		},
	}
	cmd.Flags().StringVar(&description, "description", "", fmt.Sprintf("one-line description (<=%d chars)", domain.MaxDescriptionLen))
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "comma-separated tags")
	cmd.Flags().StringArrayVar(&extra, "set", nil,
		"key=value (repeatable); known fields are typed+validated, unknown keys need --force")
	cmd.Flags().StringArrayVar(&unsets, "unset", nil, "remove a frontmatter key (repeatable)")
	cmd.Flags().BoolVar(&force, "force", false, "allow --set of a field tskflwctl doesn't know")
	return cmd
}

func newResearchEditCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "edit <research>",
		Short: "Open a research doc in your editor (whole file; re-validated on save)",
		Long: "Open the doc's markdown file in $VISUAL/$EDITOR (falling back to vi). On save the\n" +
			"file is re-parsed: a frontmatter break reopens the editor with the error rather than\n" +
			"landing on disk. The human counterpart to `research set` / `research append`.",
		Example:           "  tskflwctl research edit theming-libs\n  tskflwctl research edit   # pick from a list",
		Args:              cobra.MaximumNArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.completeResearchSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			// `edit` is interactive with no preview: reject --dry-run rather than open an
			// editor whose save is silently discarded.
			if app.DryRun {
				return fmt.Errorf("%w: `research edit` has no --dry-run preview (it's interactive) — use `research set --dry-run` or `research append --dry-run`", domain.ErrValidation)
			}
			value := ""
			if len(args) == 1 {
				value = args[0]
			}
			slug, err := app.fillSelect(value, "specify a research doc to edit",
				"no research docs available to edit", "Research doc to edit", app.researchOptions)
			if err != nil {
				return err
			}
			if !app.Gate.On() {
				return fmt.Errorf("%w: `research edit` needs an interactive terminal — use `research set`/`research append` non-interactively", domain.ErrValidation)
			}
			r, changed, err := app.Svc.EditResearch(slug, app.editViaEditor(editor.Resolve()))
			if err != nil {
				return err
			}
			if !changed {
				fmt.Fprintln(app.Out, app.Style.Dim("no changes to "+r.Slug))
				return nil
			}
			fmt.Fprintf(app.Out, "%s %s %s\n", app.Style.Green("✔"), "updated", app.Style.Bold(r.Slug))
			return nil
		},
	}
}

func newResearchAppendCmd(app *App) *cobra.Command {
	var body, bodyFile string
	cmd := &cobra.Command{
		Use:   "append <research>",
		Short: "Append a section to a research doc's body (atomic; agent-facing)",
		Long: "Append markdown to the end of a research doc's body in one atomic, validated write —\n" +
			"the scriptable counterpart to `research edit`. Content comes from --body, --body-file,\n" +
			"or stdin (--body-file -); a blank line separates it from the existing body.\n\n" +
			"Stamps updated_at. `created` stays immutable — the id is minted from it.",
		Example:           "  tskflwctl research append theming-libs --body \"## Addendum\\n\\nlipgloss v2 shipped.\"\n  cat notes.md | tskflwctl research append theming-libs --body-file -",
		Args:              cobra.MaximumNArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.completeResearchSlugs,
		RunE: func(c *cobra.Command, args []string) error {
			text, err := resolveBody(c, body, bodyFile)
			if err != nil {
				return err
			}
			if strings.TrimSpace(text) == "" {
				return fmt.Errorf("%w: nothing to append (provide --body, --body-file, or stdin via -)", domain.ErrValidation)
			}
			slug, err := app.resolveOne(args, "specify a research doc to append to", "no research docs available", "Research doc to append to", app.researchOptions)
			if err != nil {
				return err
			}
			r, newBody, err := app.Svc.AppendResearchBody(slug, text, app.DryRun)
			if err != nil {
				return err
			}
			return reportResearchMutation(app, r, newBody, "appended to", "would append to")
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "markdown to append")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read the markdown to append from a file (or - for stdin)")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	return cmd
}

// reportResearchMutation writes the standard research-mutation result: the
// research_mutation JSON envelope (carrying dry_run + the resulting body) under --json,
// else a styled one-line confirmation. body is "" for a field-only `set`.
func reportResearchMutation(app *App, r domain.Research, body, verb, dryVerb string) error {
	if app.JSON {
		return render.ResearchMutationJSON(app.Out, r, body, app.DryRun)
	}
	if app.DryRun {
		verb = dryVerb
	}
	fmt.Fprintf(app.Out, "%s %s %s\n", app.Style.Green("✔"), verb, app.Style.Bold(r.Slug))
	return nil
}

func newResearchNewCmd(app *App) *cobra.Command {
	var (
		p        core.NewResearchParams
		bodyFile string
	)
	cmd := &cobra.Command{
		Use:   "new <title>",
		Short: "Create a new research doc",
		Long: "Create a research doc: an exploration snapshot, true as of its date.\n\n" +
			"The id is minted from --created, so backdating a doc to when the work actually\n" +
			"happened sorts it into place chronologically. Research has no status and no\n" +
			"lifecycle verbs — a later doc supersedes an earlier one. A decision that needs a\n" +
			"lifecycle is an ADR, not research.",
		Example: "  tskflwctl research new \"Compare theming libraries\" --tags tui,color\n" +
			"  tskflwctl research new \"Storage model options\" --created 2026-06-24",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: activeHelpArg("provide a research title (e.g. \"Compare theming libraries\")"),
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Title = args[0]
			body, err := resolveBody(cmd, p.Body, bodyFile)
			if err != nil {
				return err
			}
			p.Body = body
			p.DryRun = app.DryRun
			r, err := app.Svc.NewResearch(p)
			if err != nil {
				return err
			}
			if app.JSON {
				// No status/bucket to report — research has none; the empty state field keeps
				// the shared created envelope's shape.
				return render.CreatedJSON(app.Out, "research", r.ID, r.Slug, "", app.rel(r.Path), app.DryRun, app.workspace())
			}
			render.CreatedHuman(app.Out, app.Style, app.linkPath(r.Path), app.DryRun)
			render.CreatedSlugNote(app.Out, app.Style, p.Title, r.Slug)
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Created, "created", "", "date the research was done, YYYY-MM-DD (default today); the id is minted from it")
	cmd.Flags().StringVar(&p.Description, "description", "", "one-line description (<=200 chars)")
	cmd.Flags().StringSliceVar(&p.Tags, "tags", nil, "comma-separated tags")
	cmd.Flags().StringVar(&p.Body, "body", "", "override the default scaffold")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read the body from a file, or - for stdin (replaces --body)")
	cmd.Flags().StringVar(&p.Template, "template", "", `body scaffold to use (default "default"); completes the available names`)
	cmd.MarkFlagsMutuallyExclusive("body", "body-file", "template")
	_ = cmd.RegisterFlagCompletionFunc("template", completeTemplateNames("research"))
	return cmd
}

func newResearchListCmd(app *App) *cobra.Command {
	var (
		tag string
		lm  listMode
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List research docs (newest first)",
		Long: "List the research corpus, newest first.\n\n" +
			"Unlike task/audit list there is no default-view filter — research has no status,\n" +
			"so the whole corpus is the listing.",
		Example:     "  tskflwctl research list\n  tskflwctl research list --tag tui\n  tskflwctl research list --json -c slug,created,description",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			mode, err := lm.resolve(cmd, app)
			if err != nil {
				return err
			}
			docs, problems, err := app.Svc.ListResearch(tag)
			if err != nil {
				return err
			}
			if err := renderList(app, mode, lm.columns, docs, problems,
				"research", render.ResearchColumns(), render.ResearchJSON, render.ResearchHuman); err != nil {
				return err
			}
			return problemsError(problems)
		},
	}
	lm.bind(cmd, render.Specs(render.ResearchColumns()))
	cmd.Flags().StringVar(&tag, "tag", "", "only docs carrying this tag")
	return cmd
}

func newResearchShowCmd(app *App) *cobra.Command {
	var (
		raw     bool
		section string
		fmOnly  bool
	)
	cmd := &cobra.Command{
		Use:               "show <research>",
		Short:             "Show a research doc's metadata and body",
		Example:           "  tskflwctl research show lipgloss-v2-charm-ecosystem\n  tskflwctl research show tui-design-decisions --section findings",
		Args:              cobra.MaximumNArgs(1), // bare → picker on a TTY; non-interactive needs the slug
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeResearchSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			slug, err := app.resolveOne(args, "specify a research doc to show", "no research docs available", "Research doc to show", app.researchOptions)
			if err != nil {
				return err
			}
			r, body, err := app.Svc.ShowResearch(slug)
			if err != nil {
				return err
			}
			body, err = narrowBody("research", slug, body, section, fmOnly)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.ResearchShowJSON(app.Out, r, body)
			}
			return app.paged(func(w io.Writer) error {
				rendered := ""
				if body != "" { // --frontmatter-only → no body render
					rendered = render.RenderBody(app.Style, body, app.markdownStyle, raw)
				}
				return render.ResearchShowHuman(w, app.Style, r, rendered)
			})
		},
	}
	cmd.Flags().BoolVar(&raw, "raw", false, "print the body as raw markdown (no styling)")
	cmd.Flags().StringVar(&section, "section", "", "show only this body section (## heading, case-insensitive)")
	cmd.Flags().BoolVar(&fmOnly, "frontmatter-only", false, "show only the metadata, no body")
	cmd.MarkFlagsMutuallyExclusive("section", "frontmatter-only")
	return cmd
}

func newResearchPathCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "path <research>",
		Short:             "Print the absolute path to a research doc's file",
		Example:           "  tskflwctl research path tui-design-decisions\n  $EDITOR \"$(tskflwctl research path tui-design-decisions)\"",
		Args:              cobra.MaximumNArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeResearchSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			slug, err := app.resolveOne(args, "specify a research doc", "no research docs available", "Research doc", app.researchOptions)
			if err != nil {
				return err
			}
			p, err := app.Svc.ResearchPath(slug)
			if err != nil {
				return err
			}
			return emitPath(app, absPath(p))
		},
	}
}

// completeResearchSlugs completes research slugs from the flat, id-led filenames.
func (a *App) completeResearchSlugs(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	root, ok := a.planningRoot()
	if !ok {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	matches, _ := filepath.Glob(filepath.Join(root, domain.ResearchDir, "*.md"))
	return flatCompletions(matches, toComplete, nil, args), cobra.ShellCompDirectiveNoFileComp
}
