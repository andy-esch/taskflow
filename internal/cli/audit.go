package cli

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/editor"
)

// auditVerbShort is the one-line help for an audit lifecycle verb, DERIVED from the
// shared registry's destination rather than hand-keyed beside it. The hand-written
// version outlived the flat-layout migration and still promised to move audits into
// `closed/`, `open/`, and `deferred/` directories that no longer exist — bucket is a
// frontmatter value now. Deriving it removes that drift class instead of correcting
// three strings: the text cannot disagree with the transition it describes.
func auditVerbShort(to string) string { return "Move audit(s) to the " + to + " bucket" }

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
		newAuditFindingCmd(app),
		newAuditLintCmd(app),
	)
	// Bucket-move verbs from the shared lifecycle registry so the verb→destination
	// mapping has ONE source the TUI also reads (the CLI ignores the registry's
	// destructive flag — it stays non-interactive/scriptable, no confirm prompts).
	for _, tr := range domain.AuditTransitions() {
		cmd.AddCommand(newAuditMoveCmd(app, tr.Verb, auditVerbShort(tr.To), domain.AuditBucket(tr.To)))
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
				return render.CreatedJSON(app.Out, "audit", a.ID, a.Slug, string(a.Bucket), app.rel(a.Path), app.DryRun, app.workspace())
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
		Use:   "list",
		Short: "List audits (open by default)",
		Long: "List audits with a segmented progress bar per row.\n\n" +
			"The headline number is the SETTLED share — findings that have reached a terminal\n" +
			"disposition, however they got there — so 100% is exactly the point an open audit\n" +
			"becomes `✔ ready to close`. The bar says how it settled, grouping the seven statuses\n" +
			"into four bands so the shape reads at a glance:\n\n" +
			"  █ green   settled here         fixed · tracked\n" +
			"  ▓ yellow  still being worked   in-progress\n" +
			"  ▒ gray    settled by dropping  deferred · superseded · wontfix\n" +
			"  ░ dim     still open           open\n\n" +
			"The glyphs differ as well as the colors, so the bands survive --color=never.",
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

// newAuditFindingCmd is the validated write path for a finding's resolution — singular
// `finding` beside the plural `findings` query, the same read/write pairing `task set` has
// beside `task list`.
//
// Until this existed the only way to resolve a finding was to hand-edit the markdown or run
// a search-and-replace over it, which is how the vocabulary drifted from its own
// documentation (finding H1 of 2026-08-17-finding-status-surface). A status a tool cannot
// write is a status nobody can be held to — and the same is true of the paragraph that says
// how it was resolved, which is why --note is here rather than left to the editor.
func newAuditFindingCmd(app *App) *cobra.Command {
	var status, note string
	var pr int
	cmd := &cobra.Command{
		Use:   "finding <audit> <code>",
		Short: "Set one finding's status and resolution note in place (validated, atomic)",
		Long: "Stamp a finding's **Status:** and **Resolution:** without touching the rest of the audit.\n\n" +
			"The status is validated against the finding vocabulary, and only the leading token\n" +
			"is normalised — decoration the line formats carry (`fixed 2026-08-24 (PR #12)`,\n" +
			"`deferred (see ADR-0003)`, `superseded by <link>`) is written verbatim, because it\n" +
			"holds dates, links, and document names whose spelling is not the tool's to flatten.\n" +
			"`tracked` additionally REQUIRES a destination (`tracked by <task-id>`), so a finding\n" +
			"handed to a task always says where it went.\n\n" +
			"--note writes the `**Resolution:**` paragraph as the finding's last block: one\n" +
			"paragraph, no newlines, placed inside the right finding by construction rather than\n" +
			"by careful typing. Passing an empty --note removes it. Both flags REPLACE what was\n" +
			"there, and given together they land in a single atomic write.\n\n" +
			"--pr N is sugar for the canonical `(PR #N)` decoration, so the reference is spelled\n" +
			"one way across the corpus and stays greppable.",
		Example: "  tskflwctl audit finding 2026-06-14-gateway H1 --status fixed\n" +
			"  tskflwctl audit finding 2026-06-14-gateway M2 --status \"deferred (see ADR-0003)\"\n" +
			"  tskflwctl audit finding 2026-06-14-gateway H1 --status \"tracked by 6g392b0rps7w\"\n" +
			"  tskflwctl audit finding 2026-06-14-gateway H1 --status fixed --note \"Widened the regex; regression test added.\"",
		Args:              cobra.ExactArgs(2),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.completeAuditSlugs,
		RunE: func(c *cobra.Command, args []string) error {
			if !c.Flags().Changed("status") && !c.Flags().Changed("note") && !c.Flags().Changed("pr") {
				return fmt.Errorf("%w: pass --status (one of: %s) and/or --note",
					domain.ErrValidation, strings.Join(domain.FindingStatuses(), ", "))
			}
			if c.Flags().Changed("status") && strings.TrimSpace(status) == "" {
				return fmt.Errorf("%w: --status was given an empty value — pass one of: %s",
					domain.ErrValidation, strings.Join(domain.FindingStatuses(), ", "))
			}
			if c.Flags().Changed("pr") {
				var err error
				if status, err = decorateWithPR(status, pr, c.Flags().Changed("status")); err != nil {
					return err
				}
			}
			edit := core.FindingEdit{Status: status}
			if c.Flags().Changed("note") {
				edit.Note = &note
			}
			a, changed, err := app.Svc.EditFinding(args[0], args[1], edit, app.DryRun)
			if err != nil {
				return err
			}
			what := describeFindingEdit(args[1], status, c.Flags().Changed("note"), note)
			if !changed && !app.JSON { // already exactly these values — say so, no write
				// Naming the value is the useful half of this message, so the status-only
				// case (the overwhelmingly common one) keeps saying it.
				if !c.Flags().Changed("note") {
					fmt.Fprintf(app.Out, "%s %s in %s is already %s\n", app.Style.Dim("•"),
						app.Style.Bold(args[1]), app.Style.Bold(a.Slug), app.Style.Bold(status))
					return nil
				}
				fmt.Fprintf(app.Out, "%s %s in %s already reads that way\n",
					app.Style.Dim("•"), app.Style.Bold(args[1]), app.Style.Bold(a.Slug))
				return nil
			}
			_, body, err := app.Svc.ShowAudit(a.Slug)
			if err != nil {
				return err
			}
			return reportAuditMutation(app, a, body,
				"set "+what+" in", "would set "+what+" in")
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "the finding's new status — one of: "+strings.Join(domain.FindingStatuses(), " | ")+" (decoration after the token is kept verbatim)")
	cmd.Flags().IntVar(&pr, "pr", 0, "append `(PR #N)` to the status — the one canonical spelling, so the reference stays greppable")
	cmd.Flags().StringVar(&note, "note", "", "the finding's `**Resolution:**` paragraph — how it was resolved; empty removes it")
	return cmd
}

// existingPRRe matches the PR references a status might already carry — `(PR #9)`, `PR#9`,
// `pr 9`. It is deliberately looser than what --pr writes: the point is to catch the
// hand-spelled forms this flag exists to displace, not only the one it produces.
var existingPRRe = regexp.MustCompile(`(?i)\bPR\s*#?\s*\d`)

// decorateWithPR appends the canonical PR reference to a status. It is sugar over free-text
// decoration for a reason: `(PR #12)`, `PR 12`, and `pull/12` all read the same to a human
// and differently to grep, and the corpus already spells its dates two ways. One flag, one
// spelling.
func decorateWithPR(status string, pr int, hasStatus bool) (string, error) {
	if !hasStatus {
		return "", fmt.Errorf("%w: --pr decorates a status — pass --status too", domain.ErrValidation)
	}
	if pr <= 0 {
		return "", fmt.Errorf("%w: --pr takes a positive pull-request number, got %d", domain.ErrValidation, pr)
	}
	if existingPRRe.MatchString(status) {
		return "", fmt.Errorf("%w: --status already carries a PR reference — pass one or the other", domain.ErrValidation)
	}
	return strings.TrimSpace(status) + " (PR #" + strconv.Itoa(pr) + ")", nil
}

// describeFindingEdit names what the receipt should say happened. Either flag is optional
// and either may be a removal, so the object of "set …" is built rather than interpolated.
// The status-only case is spelled exactly as it was before --note existed, because that is
// the overwhelmingly common receipt and there is no reason to churn it.
func describeFindingEdit(code, status string, noteSet bool, note string) string {
	switch {
	case status == "" && note == "":
		return code + "'s resolution note removed"
	case status == "":
		return code + "'s resolution note"
	case noteSet && note == "":
		return code + " " + status + " and removed its resolution note"
	case noteSet:
		return code + " " + status + " and its resolution note"
	}
	return code + " " + status
}

func newAuditLintCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "lint [audit]",
		Short: "Validate audit findings (status vocabulary, missing status, bucket↔state)",
		Long: "Lint audit findings — the audit analog of `lint` (which covers tasks, epics, and research).\n" +
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
		// A heredoc leads: findings quote percentages and code, and printf reads a
		// bare % as a format verb, truncating the write at the first one.
		Example:           "  tskflwctl audit append my-audit --body '#### H1. Title  · **Status:** open'\n  tskflwctl audit append my-audit --body-file - <<'EOF'\n#### M3. Cache hit rate fell to 40% · **Status:** open\nEOF\n  cat findings.md | tskflwctl audit append my-audit --body-file -",
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
		return render.AuditMutationJSON(app.Out, audit, body, app.DryRun, app.workspace())
	}
	if app.DryRun {
		verb = dryVerb
	}
	fmt.Fprintf(app.Out, "%s %s %s\n", app.Style.Green("✔"), verb, app.Style.Bold(audit.Slug))
	return nil
}
