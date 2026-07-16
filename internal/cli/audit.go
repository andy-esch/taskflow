package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/editor"
)

// auditVerbHelp is the CLI-specific one-line help for each audit lifecycle verb.
// As with tasks, the verb→bucket mapping lives in the shared registry
// (domain.AuditTransitions()) while this presentation text stays CLI-local; keyed
// by verb so newAuditCmd can assert every registry verb has an entry.
var auditVerbHelp = map[string]string{
	"close":  "Move audit(s) to closed/",
	"reopen": "Move audit(s) back to open/",
	"defer":  "Move audit(s) to deferred/",
}

func newAuditCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "audit", Short: "Work with code audits"}
	cmd.AddCommand(
		newAuditNewCmd(app),
		newAuditListCmd(app),
		newAuditShowCmd(app),
		newAuditInfoCmd(app),
		newAuditPathCmd(app),
		newAuditEditCmd(app),
		newAuditAppendCmd(app),
		newAuditFindingsCmd(app),
		newAuditLintCmd(app),
	)
	// Bucket-move verbs from the shared lifecycle registry so the verb→destination
	// mapping has ONE source the TUI also reads (the CLI ignores the registry's
	// destructive flag — it stays non-interactive/scriptable, no confirm prompts).
	for _, tr := range domain.AuditTransitions() {
		short, ok := auditVerbHelp[tr.Verb]
		if !ok {
			panic("cli: no help text for audit transition verb " + tr.Verb)
		}
		cmd.AddCommand(newAuditMoveCmd(app, tr.Verb, short, domain.AuditBucket(tr.To)))
	}
	return cmd
}

func newAuditNewCmd(app *App) *cobra.Command {
	var (
		p        core.NewAuditParams
		bodyFile string
	)
	cmd := &cobra.Command{
		Use:               "new <area>",
		Short:             "Create a new audit (open bucket, scaffolded findings)",
		Example:           "  tskflwctl audit new dispatcher\n  tskflwctl audit new arch-data-flow --date 2026-06-16",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: activeHelpArg("provide an area to audit (e.g. dispatcher)"),
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Area = args[0]
			body, err := resolveBody(cmd, p.Body, bodyFile)
			if err != nil {
				return err
			}
			p.Body = body
			p.DryRun = app.DryRun
			a, err := app.Svc.NewAudit(p)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.CreatedJSON(app.Out, "audit", a.Slug, string(a.Bucket), app.rel(a.Path), app.DryRun)
			}
			render.CreatedHuman(app.Out, app.Style, app.linkPath(a.Path), app.DryRun)
			render.CreatedSlugNote(app.Out, app.Style, p.Area, a.Slug)
			if !app.DryRun {
				fmt.Fprintf(app.Out, "%s\n", app.Style.Dim("→ next: add findings, then tskflwctl audit close "+a.Slug))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Date, "date", "", "audit date YYYY-MM-DD (default today)")
	cmd.Flags().StringVar(&p.Body, "body", "", "override the default scaffold")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read the body from a file, or - for stdin (replaces --body)")
	cmd.Flags().StringVar(&p.Template, "template", "", `body scaffold to use (default "default"); e.g. "security". completes the available names`)
	cmd.MarkFlagsMutuallyExclusive("body", "body-file", "template")
	_ = cmd.RegisterFlagCompletionFunc("template", completeTemplateNames("audit"))
	return cmd
}

func newAuditListCmd(app *App) *cobra.Command {
	var (
		all, closed, deferred bool
		lm                    listMode
	)
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List audits (open by default)",
		Example:     "  tskflwctl audit list\n  tskflwctl audit list --all -o table -c slug,open\n  tskflwctl audit list --closed -o json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			mode, err := lm.resolve(cmd, app)
			if err != nil {
				return err
			}
			bucket := ""
			switch {
			case closed:
				bucket = string(domain.AuditClosed)
			case deferred:
				bucket = string(domain.AuditDeferred)
			}
			audits, problems, err := app.Svc.ListAudits(bucket, all)
			if err != nil {
				return err
			}
			if err := renderList(app, mode, lm.columns, audits, problems,
				"audits", render.AuditColumns(), render.AuditsJSON, render.AuditsHuman); err != nil {
				return err
			}
			return problemsError(problems)
		},
	}
	lm.bind(cmd, render.Specs(render.AuditColumns()))
	cmd.Flags().BoolVar(&all, "all", false, "all buckets")
	cmd.Flags().BoolVar(&closed, "closed", false, "closed audits only")
	cmd.Flags().BoolVar(&deferred, "deferred", false, "deferred audits only")
	cmd.MarkFlagsMutuallyExclusive("all", "closed", "deferred")
	return cmd
}

func newAuditFindingsCmd(app *App) *cobra.Command {
	var (
		status, effort, urgency []string
		component               string
		lm                      listMode
	)
	cmd := &cobra.Command{
		Use:   "findings [audit]",
		Short: "Query findings across audits (or one) by status/effort/urgency/component",
		Long: "Search audit findings — the structured per-finding view, not the aggregate.\n" +
			"With no argument, searches every audit; with an audit slug, just that one.\n" +
			"status/effort/urgency match exactly (case-insensitive, comma = any-of);\n" +
			"--component is a case-insensitive substring. Each --json hit carries its\n" +
			"audit slug and bucket.",
		Example: "  tskflwctl audit findings --status open --effort XS,S --json\n" +
			"  tskflwctl audit findings 2026-06-14-simplify-apigateway --status in-progress\n" +
			"  tskflwctl audit findings --component stravapipe -o table",
		Args:              cobra.MaximumNArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeAuditSlugs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mode, err := lm.resolve(cmd, app)
			if err != nil {
				return err
			}
			f := core.FindingFilter{Status: status, Effort: effort, Urgency: urgency, Component: component}
			if len(args) == 1 {
				f.Audit = args[0]
			}
			findings, problems, err := app.Svc.QueryFindings(f)
			if err != nil {
				return err
			}
			if err := renderList(app, mode, lm.columns, findings, problems,
				"findings", render.FindingColumns(), render.FindingsJSON, render.FindingsHuman); err != nil {
				return err
			}
			return problemsError(problems)
		},
	}
	lm.bind(cmd, render.Specs(render.FindingColumns()))
	cmd.Flags().StringSliceVar(&status, "status", nil, "filter by finding status (comma-separated, any-of)")
	cmd.Flags().StringSliceVar(&effort, "effort", nil, "filter by effort XS,S,M,L (any-of)")
	cmd.Flags().StringSliceVar(&urgency, "urgency", nil, "filter by urgency acute,soon,eventually (any-of)")
	cmd.Flags().StringVar(&component, "component", "", "filter by component (case-insensitive substring)")
	return cmd
}

func newAuditLintCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "lint [audit]",
		Short: "Validate audit findings (status vocabulary, missing status, bucket↔state)",
		Long: "Lint audit findings — the audit analog of `lint` (which covers tasks/epics).\n" +
			"Checks every finding has a legal **Status:** (catching typos a free-text edit\n" +
			"allows) and that a non-open audit has no still-open findings. With no argument\n" +
			"it lints every audit; with a slug, just that one. Exit 11 when issues are found.",
		Example:           "  tskflwctl audit lint\n  tskflwctl audit lint 2026-06-14-gateway --json",
		Args:              cobra.MaximumNArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeAuditSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			slug := ""
			if len(args) == 1 {
				slug = args[0]
			}
			results, problems, err := app.Svc.LintAudits(slug)
			if err != nil {
				return err
			}
			if app.JSON {
				if err := render.LintJSON(app.Out, results, problems); err != nil {
					return err
				}
			} else {
				render.ProblemsHuman(app.ErrOut, app.Style, problems)
				render.LintHuman(app.Out, app.Style, results, "audit")
				if len(results) == 0 && len(problems) == 0 {
					fmt.Fprintf(app.Out, "%s all audit findings pass lint\n", app.Style.Green("✔"))
				}
			}
			if len(results)+len(problems) > 0 {
				return fmt.Errorf("%w: %d audit(s) with finding issues, %d unreadable file(s)",
					domain.ErrValidation, len(results), len(problems))
			}
			return nil
		},
	}
}

func newAuditShowCmd(app *App) *cobra.Command {
	var (
		raw     bool
		section string
		fmOnly  bool
	)
	cmd := &cobra.Command{
		Use:               "show <audit>",
		Short:             "Show an audit's metadata and body",
		Example:           "  tskflwctl audit show 2026-06-20-api-gateway\n  tskflwctl audit show 2026-06-20-api-gateway --section findings\n  tskflwctl audit show 2026-06-20-api-gateway --frontmatter-only",
		Args:              cobra.MaximumNArgs(1), // bare → picker on a TTY; non-interactive needs the slug
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeAuditSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			slug, err := app.resolveOne(args, "specify an audit to show", "no audits available", "Audit to show", app.auditOptions)
			if err != nil {
				return err
			}
			audit, body, err := app.Svc.ShowAudit(slug)
			if err != nil {
				return err
			}
			// --section / --frontmatter-only narrow the audit's markdown body only; the
			// metadata + finding tree always show. Parse findings from the FULL body so
			// the tree is unaffected by a narrowed view.
			findings := domain.ParseFindings(body)
			body, err = narrowBody("audit", slug, body, section, fmOnly)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.AuditShowJSON(app.Out, audit, body)
			}
			return app.paged(func(w io.Writer) error {
				rendered := ""
				if body != "" { // --frontmatter-only → no body render (and no trailing blank line)
					rendered = render.RenderBody(app.Style, body, app.markdownStyle, raw)
				}
				return render.AuditShowHuman(w, app.Style, audit, findings, rendered)
			})
		},
	}
	cmd.Flags().BoolVar(&raw, "raw", false, "print the raw markdown body (skip rendering)")
	addBodyScopeFlags(cmd, &section, &fmOnly)
	return cmd
}

// newAuditInfoCmd is the token-cheap audit metadata read: file path, bucket, and
// the finding disposition tally (the audit analogue of `task info`'s acceptance
// tally), WITHOUT the body. `--json` is the machine path.
func newAuditInfoCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "info <audit>",
		Short:             "Show an audit's metadata + file path + finding tally (no body)",
		Example:           "  tskflwctl audit show 2026-06-20-api-gateway --frontmatter-only\n  tskflwctl audit info 2026-06-20-api-gateway --json",
		Args:              cobra.MaximumNArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeAuditSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			slug, err := app.resolveOne(args, "specify an audit", "no audits available", "Audit", app.auditOptions)
			if err != nil {
				return err
			}
			// ShowAudit populates the disposition tally on load (parseAudit), so no
			// re-parse is needed for the counts.
			audit, _, err := app.Svc.ShowAudit(slug)
			if err != nil {
				return err
			}
			path := absPath(audit.Path)
			if app.JSON {
				return render.AuditInfoJSON(app.Out, audit, path)
			}
			render.AuditInfoHuman(app.Out, app.Style, audit, path)
			return nil
		},
	}
}

// newAuditPathCmd prints just the absolute path to an audit's file — the audit
// counterpart to `task path`, parse-free so it works on a broken file too.
func newAuditPathCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "path <audit>",
		Short:             "Print the absolute path to an audit's file",
		Example:           "  tskflwctl audit path 2026-06-20-api-gateway\n  $EDITOR \"$(tskflwctl audit path 2026-06-20-api-gateway)\"",
		Args:              cobra.MaximumNArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeAuditSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			slug, err := app.resolveOne(args, "specify an audit", "no audits available", "Audit", app.auditOptions)
			if err != nil {
				return err
			}
			p, err := app.Svc.AuditPath(slug)
			if err != nil {
				return err
			}
			return emitPath(app, absPath(p))
		},
	}
}

func newAuditMoveCmd(app *App, use, short string, to domain.AuditBucket) *cobra.Command {
	return &cobra.Command{
		Use:               use + " <audit>...",
		Short:             short,
		Example:           "  tskflwctl audit " + use + " 2026-06-06-schemas-scripts\n  tskflwctl audit " + use + "   # pick from a list",
		Args:              cobra.ArbitraryArgs, // bare → picker on a TTY; non-interactive needs ≥1 arg
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.auditCompleter(to), // don't offer audits already at `to`
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Bare verb: pick an audit on a TTY; non-interactive → exit 11.
				slug, err := app.fillSelect("", "specify at least one audit to "+use,
					"no audits available to "+use, "Audit to "+use, app.auditMoveOptions(to))
				if err != nil {
					return err
				}
				args = []string{slug}
			}
			return runMoves(app, args, string(to),
				func(slug string) (domain.Audit, error) { return app.Svc.MoveAudit(slug, to, app.DryRun) },
				func(a domain.Audit) string { return a.Slug })
		},
	}
}

// newAuditEditCmd is the human face of audit mutation: open the audit file in the
// user's editor and re-validate on save — the audit twin of `task edit`, complementing
// the agent-facing `audit append`. The save is accepted only if it still parses
// (parse-before-accept); once it lands, the findings are lint-checked and any issues
// (a bad **Status:**, a bucket↔state drift a free-text edit can introduce) are surfaced
// as a WARNING, not a hard error — lint is advisory here, like `task edit`'s re-lint flag.
func newAuditEditCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "edit <audit>",
		Short: "Open an audit in your editor (whole file; re-validated on save)",
		Long: "Open the audit's markdown file in $VISUAL/$EDITOR (falling back to vi). On save\n" +
			"the file is re-parsed: a frontmatter break reopens the editor with the error rather\n" +
			"than landing on disk. The findings are then lint-checked and any issues (bad\n" +
			"**Status:**, bucket↔state drift) are surfaced as a warning. The human counterpart\n" +
			"to `audit append` (scriptable).",
		Example:           "  tskflwctl audit edit 2026-06-20-api-gateway\n  tskflwctl audit edit   # pick from a list",
		Args:              cobra.MaximumNArgs(1), // bare → picker on a TTY; non-interactive needs the slug
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.completeAuditSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			// `edit` is interactive ($EDITOR on the whole file) with no preview: reject
			// --dry-run rather than open an editor whose save is silently discarded.
			if app.DryRun {
				return fmt.Errorf("%w: `audit edit` has no --dry-run preview (it's interactive) — use `audit append --dry-run` for a non-interactive preview", domain.ErrValidation)
			}
			value := ""
			if len(args) == 1 {
				value = args[0]
			}
			slug, err := app.fillSelect(value, "specify an audit to edit",
				"no audits available to edit", "Audit to edit", app.auditOptions)
			if err != nil {
				return err
			}
			if !app.Gate.On() {
				return fmt.Errorf("%w: `audit edit` needs an interactive terminal — use `audit append` to add findings non-interactively", domain.ErrValidation)
			}
			audit, changed, err := app.Svc.EditAudit(slug, app.editViaEditor(editor.Resolve()))
			if err != nil {
				return err
			}
			if !changed {
				fmt.Fprintln(app.Out, app.Style.Dim("no changes to "+audit.Slug))
				return nil
			}
			fmt.Fprintf(app.Out, "%s %s %s\n", app.Style.Green("✔"), "updated", app.Style.Bold(audit.Slug))
			// Re-validate findings (parse-before-accept only guaranteed the file loads):
			// surface finding-level issues as a warning so a free-text slip doesn't land
			// silently, but don't fail — the edit already happened and lint is advisory.
			if results, _, lerr := app.Svc.LintAudits(slug); lerr == nil && len(results) > 0 {
				fmt.Fprintf(app.ErrOut, "%s findings need attention (see `audit lint %s`):\n", app.Style.Warn("⚠"), audit.Slug)
				render.LintHuman(app.ErrOut, app.Style, results, "audit")
			}
			return nil
		},
	}
}

// newAuditAppendCmd is the agent face of audit body editing: append a section
// (typically a finding) to the body in one atomic, validated write — the scriptable
// twin of `audit edit`, mirroring `task append`. Finding GRAMMAR correctness is left
// to `audit lint` (raw markdown is appended), so a malformed finding lands but is
// caught by lint rather than rejected inline.
func newAuditAppendCmd(app *App) *cobra.Command {
	var body, bodyFile string
	cmd := &cobra.Command{
		Use:   "append <audit>",
		Short: "Append a section to an audit's body (atomic; agent-facing)",
		Long: "Append markdown to the end of an audit's body in one atomic, validated write —\n" +
			"the scriptable counterpart to `audit edit`, e.g. to add a finding section. Content\n" +
			"comes from --body, --body-file, or stdin (--body-file -); a blank line separates it\n" +
			"from the existing body. Finding grammar is left to `audit lint`.",
		Example:           "  tskflwctl audit append my-audit --body '#### H1. Title  · **Status:** open'\n  printf '#### M3. ...\\n' | tskflwctl audit append my-audit --body-file -",
		Args:              cobra.MaximumNArgs(1), // bare → picker on a TTY; non-interactive needs the slug
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.completeAuditSlugs,
		RunE: func(c *cobra.Command, args []string) error {
			text, err := resolveBody(c, body, bodyFile)
			if err != nil {
				return err
			}
			if strings.TrimSpace(text) == "" {
				return fmt.Errorf("%w: nothing to append (provide --body, --body-file, or stdin via -)", domain.ErrValidation)
			}
			slug, err := app.resolveOne(args, "specify an audit to append to", "no audits available", "Audit to append to", app.auditOptions)
			if err != nil {
				return err
			}
			audit, newBody, err := app.Svc.AppendAuditBody(slug, text, app.DryRun)
			if err != nil {
				return err
			}
			return reportAuditMutation(app, audit, newBody, "appended to", "would append to")
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "markdown to append")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read the markdown to append from a file (or - for stdin)")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	return cmd
}

// reportAuditMutation renders the result of `audit append` — JSON envelope under
// --json, else a one-line confirmation ("would …" on a --dry-run preview). The audit
// counterpart to reportTaskMutation.
func reportAuditMutation(app *App, audit domain.Audit, body, verb, dryVerb string) error {
	if app.JSON {
		return render.AuditMutationJSON(app.Out, audit, body, app.DryRun)
	}
	if app.DryRun {
		verb = dryVerb
	}
	fmt.Fprintf(app.Out, "%s %s %s\n", app.Style.Green("✔"), verb, app.Style.Bold(audit.Slug))
	return nil
}
