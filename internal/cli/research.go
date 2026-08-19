package cli

import (
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// The research command surface is deliberately smaller than task/audit: new, list,
// show, path. There are no lifecycle verbs (research has no status — epic 28) and no
// `set`, because there are no cross-reference fields to maintain; a research doc is
// edited as prose, so `research path` + $EDITOR is the edit story for now.
func newResearchCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "research", Short: "Work with research docs"}
	cmd.AddCommand(
		newResearchNewCmd(app),
		newResearchListCmd(app),
		newResearchShowCmd(app),
		newResearchPathCmd(app),
	)
	return cmd
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
