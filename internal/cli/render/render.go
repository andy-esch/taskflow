// Package render turns typed core results into CLI output. It owns the human
// presentation (ANSI tables, lipgloss trees) and is a thin io.Writer wrapper over
// the machine wire contract in internal/wire: each *JSON emit func builds a wire
// envelope value (wire.ToXEnvelope) and encodes it, so the CLI and a future web
// adapter serialize identical JSON. The wire format itself — envelopes, DTOs,
// SchemaVersion, the JSON Schema — lives in wire, not here.
package render

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2/tree"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
	"github.com/andy-esch/taskflow/internal/wire"
)

// SchemaVersion re-exports the wire contract's version so existing CLI call sites
// (exit.go's error envelope, the projection writer in columns.go) keep one import.
// The canonical value + changelog live in internal/wire.
const SchemaVersion = wire.SchemaVersion

// TasksHuman writes a scannable table of tasks (empty input writes nothing).
func TasksHuman(w io.Writer, st Style, tasks []domain.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	rows := make([][]string, 0, len(tasks))
	misfiled := 0
	for _, t := range tasks {
		status := st.Status(t.Status)
		if t.Misfiled() {
			misfiled++
			status = st.Warn("⚠ ") + status
		}
		updated := t.Updated
		if updated == "" {
			updated = t.Created
		}
		// Description goes last so it's the column that truncates to terminal width.
		rows = append(rows, []string{status, st.Bold(t.Slug), st.Dim(theme.RelativeDate(updated)), t.Description})
	}
	writeTable(w, st.width, []string{st.Dim("STATUS"), st.Dim("TASK"), st.Dim("UPDATED"), st.Dim("DESCRIPTION")}, rows)
	fmt.Fprintf(w, "\n%s\n", st.Dim(plural(len(tasks), "task")))
	if misfiled > 0 {
		fmt.Fprintf(w, "%s\n", st.Warn(fmt.Sprintf("⚠ %d misfiled (status ≠ folder; run `lint --fix` to realign)", misfiled)))
	}
	return nil
}

// plural renders "N noun" / "N nouns".
func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// TasksJSON writes a stable, versioned JSON envelope of tasks, including any
// per-file load problems so a JSON consumer never silently loses unreadable
// files (mirrors LintJSON's `unreadable`).
func TasksJSON(w io.Writer, tasks []domain.Task, problems []domain.FileProblem) error {
	return wire.EncodeJSON(w, wire.ToTasksEnvelope(tasks, problems))
}

// TaskShowHuman prints a task's metadata followed by its body.
func TaskShowHuman(w io.Writer, st Style, t domain.Task, body string) error {
	field := func(label, value string) {
		lbl := fmt.Sprintf("%-12s", label+":")
		if st.width > 0 { // fit the value to the terminal (TTY only; piped stays full)
			value = truncate(value, st.width-visibleWidth(lbl)-1)
		}
		fmt.Fprintf(w, "%s %s\n", st.Dim(lbl), value)
	}
	field("slug", st.Bold(t.Slug))
	status := st.Status(t.Status)
	if t.Misfiled() {
		status += "  " + st.Warn(fmt.Sprintf("⚠ frontmatter says %q", t.Declared))
	}
	field("status", status)
	if t.Epic != "" {
		field("epic", t.Epic)
	}
	if t.Priority != "" {
		field("priority", st.Priority(t.Priority))
	}
	if t.Tier != 0 {
		field("tier", fmt.Sprintf("%d", t.Tier))
	}
	if len(t.Tags) > 0 {
		field("tags", strings.Join(t.Tags, ", "))
	}
	if t.Description != "" {
		field("description", t.Description)
	}
	if t.Created != "" {
		field("created", t.Created)
	}
	if t.Updated != "" {
		field("updated", fmt.Sprintf("%s %s", t.Updated, st.Dim("("+theme.RelativeDate(t.Updated)+")")))
	}
	fmt.Fprintf(w, "\n%s", body)
	return nil
}

// TaskShowJSON writes a task plus its body.
func TaskShowJSON(w io.Writer, t domain.Task, body string) error {
	return wire.EncodeJSON(w, wire.ToTaskShowEnvelope(t, body))
}

// TaskMutationJSON writes the result of a task mutation (`task set`/`append`/`set
// --body`): the reloaded task, dry_run (always present — a preview must be
// distinguishable from a real write), and the resulting body for the body-editing
// commands (empty/omitted for field-only `set`). Distinct from TaskShowEnvelope so
// the mutation-only dry_run never lands on the `task show` read type.
func TaskMutationJSON(w io.Writer, t domain.Task, body string, dryRun bool) error {
	return wire.EncodeJSON(w, wire.ToTaskMutationEnvelope(t, body, dryRun))
}

// MoveResult is the per-item outcome of a transition (the wire type), re-exported
// so the CLI's move loop (moves.go) keeps building it through the render package.
type MoveResult = wire.MoveResult

// MovesHuman prints one line per transition outcome ("would move" on a
// --dry-run preview).
func MovesHuman(out, errw io.Writer, st Style, results []MoveResult, dryRun bool) {
	verb := "moved"
	if dryRun {
		verb = "would move"
	}
	for _, r := range results {
		if r.Error != "" {
			// Failures are diagnostics → stderr, so a partial `… | xargs move`
			// doesn't interleave errors into the data stream.
			fmt.Fprintf(errw, "%s %s: %s\n", st.Red("✘"), st.Bold(r.Slug), r.Error)
		} else {
			// A recorded/would-be revisit date (defer --until) is shown inline so the
			// preview and the real run both confirm the snooze, not just the move.
			revisit := ""
			if r.RevisitAt != "" {
				revisit = st.Dim(fmt.Sprintf(" (revisit %s)", r.RevisitAt))
			}
			fmt.Fprintf(out, "%s %s %s -> %s%s\n", st.Green("✔"), verb, st.Bold(r.Slug), r.To, revisit)
		}
	}
}

// MovesJSON writes the structured per-task transition report; dry_run marks a
// preview (nothing was written).
func MovesJSON(w io.Writer, results []MoveResult, dryRun bool) error {
	return wire.EncodeJSON(w, wire.ToMovesEnvelope(results, dryRun))
}

// DoctorProblem is one linkback inconsistency (the wire type), re-exported so the
// CLI's doctor command keeps building it through the render package.
type DoctorProblem = wire.DoctorProblem

// DoctorJSON writes the linkback audit; problems is empty (not null) when the
// links are consistent, so a consumer can len() it without a nil check.
func DoctorJSON(w io.Writer, root string, problems []DoctorProblem) error {
	return wire.EncodeJSON(w, wire.ToDoctorEnvelope(root, problems))
}

// DoctorHuman writes the linkback audit: a ⚠ per problem (with a count footer),
// or a ✔ when the links are consistent.
func DoctorHuman(w io.Writer, st Style, problems []DoctorProblem) {
	if len(problems) == 0 {
		fmt.Fprintf(w, "%s linkback consistent\n", st.Green("✔"))
		return
	}
	for _, p := range problems {
		fmt.Fprintf(w, "%s %s\n", st.Warn("⚠"), p.Message)
	}
	fmt.Fprintf(w, "\n%s\n", st.Dim(plural(len(problems), "linkback problem")))
}

// SummaryHuman renders the at-a-glance dashboard.
func SummaryHuman(w io.Writer, st Style, s core.Summary) error {
	// Status counts — active line, then archived line, only non-zero buckets.
	active, archived := splitCounts(s.Counts)
	fmt.Fprintf(w, "%s\n", st.Bold("Tasks"))
	if line := countLine(st, active); line != "" {
		fmt.Fprintf(w, "  %s  %s\n", st.Dim("active  "), line)
	}
	if line := countLine(st, archived); line != "" {
		fmt.Fprintf(w, "  %s  %s\n", st.Dim("archived"), line)
	}

	if len(s.InProgress) > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Bold(fmt.Sprintf("In progress (%d)", len(s.InProgress))))
		rows := make([][]string, 0, len(s.InProgress))
		for _, t := range s.InProgress {
			rows = append(rows, []string{"  " + st.Bold(t.Slug), st.Dim(theme.RelativeDate(theme.TaskDate(t))), t.Description})
		}
		writeTable(w, st.width, nil, rows)
	}

	if len(s.Epics) > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Bold("Epics"))
		rows := make([][]string, 0, len(s.Epics))
		for _, e := range s.Epics {
			bar := fmt.Sprintf("%s %s", st.Bar(e.Percent(), 10), st.Percent(e.Percent()))
			rows = append(rows, []string{"  " + st.Bold(e.Epic.ID), bar, theme.Counts(e.Done, e.Total), e.Epic.Description})
		}
		writeTable(w, st.width, nil, rows)
	}

	// Only open audits, only when there are any — the actionable subset, rendered
	// with the same bar treatment as epics so the dashboard reads from one vocabulary.
	if len(s.OpenAudits) > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Bold(fmt.Sprintf("Open audits (%d)", len(s.OpenAudits))))
		rows := make([][]string, 0, len(s.OpenAudits))
		for _, a := range s.OpenAudits {
			bar := fmt.Sprintf("%s %s", st.SegmentBar(a.DoneFindings, a.ActiveFindings, a.DroppedFindings, a.Findings, 10), st.Percent(a.Percent()))
			rows = append(rows, []string{"  " + st.Bold(a.Slug), bar, theme.Counts(a.Resolved(), a.Findings), a.Area})
		}
		writeTable(w, st.width, nil, rows)
	}

	// Audit findings — the actionable cross-audit inbox, triaged. Same source as the
	// TUI dashboard's widget (core.Summary.Findings), so the two surfaces agree.
	if fr := s.Findings; fr.Open+fr.InProgress > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Bold(fmt.Sprintf("Audit findings (%d open · %d in progress)", fr.Open, fr.InProgress)))
		if line := countByLine(st, fr.ByUrgency); line != "" {
			fmt.Fprintf(w, "  %s  %s\n", st.Dim("by urgency"), line)
		}
		if line := countByLine(st, fr.ByComponent); line != "" {
			fmt.Fprintf(w, "  %s  %s\n", st.Dim("by area  "), line)
		}
	}
	if s.ReadyToClose > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Green(fmt.Sprintf("✓ %d audit(s) ready to close (all findings resolved; `audit close <slug>`)", s.ReadyToClose)))
	}

	if s.RevisitDue > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Warn(fmt.Sprintf("↻ %d deferred due to revisit (snooze date reached; `task ready`/`task next` to resume)", s.RevisitDue)))
	}
	if s.Misfiled > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Warn(fmt.Sprintf("⚠ %d misfiled (status ≠ folder; run `lint --fix`)", s.Misfiled)))
	}
	if s.BadEpicStatus > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Warn(fmt.Sprintf("⚠ %d epic(s) with unrecognized status (set active/retired/deprecated; run `lint`)", s.BadEpicStatus)))
	}
	if len(s.Problems) > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Red(fmt.Sprintf("! %d unreadable file(s) (run `lint`)", len(s.Problems))))
	}
	return nil
}

func splitCounts(counts []core.StatusCount) (active, archived []core.StatusCount) {
	for _, c := range counts {
		if c.Status.IsActive() {
			active = append(active, c)
		} else {
			archived = append(archived, c)
		}
	}
	return active, archived
}

// countByLine renders a finding breakdown ("1 acute · 12 soon · 23 eventually"),
// the dim-separated, uncolored counterpart of the dashboard's by-urgency / by-area
// lines. Shares the iterate/format/join STRUCTURE with them via theme.Breakdown;
// only this surface's plain segment format + dim separator differ (audit M10).
func countByLine(st Style, cs []core.CountBy) string {
	return theme.Breakdown(cs, st.Dim(" · "), 0,
		func(c core.CountBy) string { return fmt.Sprintf("%d %s", c.Count, c.Key) }, nil)
}

// countLine renders "3 next-up · 1 in-progress", skipping zero buckets.
func countLine(st Style, counts []core.StatusCount) string {
	var parts []string
	for _, c := range counts {
		if c.Count == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d %s", c.Count, st.Status(c.Status)))
	}
	return strings.Join(parts, st.Dim(" · "))
}

// SummaryJSON writes the dashboard as a versioned envelope.
func SummaryJSON(w io.Writer, s core.Summary) error {
	return wire.EncodeJSON(w, wire.ToSummaryEnvelope(s))
}

// VersionHuman prints the CLI version.
func VersionHuman(w io.Writer, st Style, version string) {
	fmt.Fprintf(w, "%s %s\n", st.Bold("tskflwctl"), version)
}

// VersionJSON writes the version in the standard envelope.
func VersionJSON(w io.Writer, version string) error {
	return wire.EncodeJSON(w, wire.ToVersionEnvelope(version))
}

// CreatedHuman prints the path of a newly created file (or, under --dry-run,
// the path that WOULD be created).
func CreatedHuman(w io.Writer, st Style, path string, dryRun bool) {
	verb := "created"
	if dryRun {
		verb = "would create"
	}
	fmt.Fprintf(w, "%s %s\n", st.Green(verb), st.Bold(path))
}

// CreatedSlugNote surfaces the derived slug on a dim line after the "created …"
// line, but only when Slugify did something beyond the obvious — a title with a
// colon, em-dash, arrow, or other dropped character shows where its filename came
// from, while the everyday "Add retry backoff" → add-retry-backoff stays silent
// (lowercasing + space→hyphen is no surprise). The JSON envelope already carries
// the id, so this is human-only. slug is the full filename id (NN-… for an epic,
// <date>-… for an audit); divergence is judged on the original title/area.
func CreatedSlugNote(w io.Writer, st Style, title, slug string) {
	if slug == "" || domain.Slugify(title) == naiveSlug(title) {
		return
	}
	fmt.Fprintf(w, "%s\n", st.Dim("→ slug: "+slug))
}

// naiveSlug is the "no surprise" slug: lowercase, apostrophes dropped (Slugify
// vanishes them silently, so "don't" → dont is no surprise), each run of whitespace
// collapsed to a single hyphen, and trailing '-'/'.' trimmed (Slugify trims them
// too, so a trailing dot/space is no surprise either). When Slugify's real output
// matches it, the only transforms were these obvious ones and the derivation needs
// no note; when it differs, a character was genuinely turned into a word-break (a
// colon, em-dash, arrow, …) and the note fires.
func naiveSlug(title string) string {
	lowered := strings.Map(func(r rune) rune {
		if r == '\'' || r == '’' || r == '‘' {
			return -1 // drop apostrophes, mirroring Slugify
		}
		return unicode.ToLower(r)
	}, title)
	return strings.Trim(strings.Join(strings.Fields(lowered), "-"), "-.")
}

// CreatedJSON writes a versioned envelope for a newly created item; dry_run
// marks a preview (nothing was written). status is the new item's status (task
// status / epic status / audit bucket); path is relative to the planning root,
// matching the human output.
func CreatedJSON(w io.Writer, kind, id, status, path string, dryRun bool) error {
	return wire.EncodeJSON(w, wire.ToCreatedEnvelope(kind, id, status, path, dryRun))
}

// EpicsHuman writes a table of epics with task rollup.
func EpicsHuman(w io.Writer, st Style, epics []core.EpicSummary) error {
	if len(epics) == 0 {
		return nil
	}
	rows := make([][]string, 0, len(epics))
	for _, e := range epics {
		pct := e.Percent()
		progress := fmt.Sprintf("%s %s %s", st.Bar(pct, 8), st.Percent(pct), theme.Counts(e.Done, e.Total))
		status := e.Epic.Status
		if !domain.IsKnownEpicStatus(e.Epic.Status) { // flag a fixable data problem inline
			disp := e.Epic.Status
			if disp == "" {
				disp = "—"
			}
			status = st.Warn("⚠ " + disp)
		}
		rows = append(rows, []string{st.Bold(e.Epic.ID), status, progress, e.Epic.Description})
	}
	writeTable(w, st.width, []string{st.Dim("EPIC"), st.Dim("STATUS"), st.Dim("PROGRESS"), st.Dim("DESCRIPTION")}, rows)
	return nil
}

// EpicsJSON writes a versioned envelope of epics with rollup, including any
// per-file load problems (mirrors LintJSON's `unreadable`).
func EpicsJSON(w io.Writer, epics []core.EpicSummary, problems []domain.FileProblem) error {
	return wire.EncodeJSON(w, wire.ToEpicsEnvelope(epics, problems))
}

// EpicShowHuman prints an epic, its tasks, and its body. The rollup arrives on the
// EpicSummary (computed once by ShowEpic) rather than re-derived here — same rule
// as epic list / status / the TUI detail (audit M3).
func EpicShowHuman(w io.Writer, st Style, es core.EpicSummary, tasks []domain.Task, body string) error {
	field := func(label, value string) {
		lbl := fmt.Sprintf("%-12s", label+":")
		if st.width > 0 { // fit the value to the terminal (TTY only; piped stays full)
			value = truncate(value, st.width-visibleWidth(lbl)-1)
		}
		fmt.Fprintf(w, "%s %s\n", st.Dim(lbl), value)
	}
	field("id", st.Bold(es.Epic.ID))
	field("status", es.Epic.Status)
	if es.Epic.Description != "" {
		field("description", es.Epic.Description)
	}
	// Deprecated tasks leave the denominator (counted separately in the tasks header).
	pct := es.Percent()
	field("progress", fmt.Sprintf("%s %s  %s", st.Bar(pct, 10), st.Percent(pct), theme.Counts(es.Done, es.Total)))
	header := fmt.Sprintf("tasks (%d):", len(tasks))
	if es.Deprecated > 0 {
		// Note the withdrawn count — those tasks are listed but excluded from the
		// done/total rollup shown by `epic list`/`status`.
		header = fmt.Sprintf("tasks (%d, %d deprecated — excluded from progress):", len(tasks), es.Deprecated)
	}
	fmt.Fprintf(w, "%s\n", st.Dim(header))
	if len(tasks) > 0 {
		// Render epic → tasks as a tree grouped by status (lipgloss/v2). Node text
		// is pre-styled by st (so --color is honored); the tree contributes only
		// its plain connectors. Rootless — the "tasks (…)" header is the label, so
		// the epic id isn't repeated. The --json envelope is unaffected.
		byStatus := make(map[domain.Status][]domain.Task, len(tasks))
		for _, task := range tasks {
			byStatus[task.Status] = append(byStatus[task.Status], task)
		}
		tr := tree.New()
		for _, s := range domain.AllStatuses() {
			grp := byStatus[s]
			if len(grp) == 0 {
				continue
			}
			// Fit nodes to the terminal: a header sits at one indent level (~4 cells),
			// a child at two (~8). Truncate ANSI-aware so a long slug can't overflow
			// (tree.Width pads but won't cut an unbreakable token). Width 0 (piped) = full.
			sub := tree.Root(fitNode(st, st.Status(s), 4))
			for _, task := range grp {
				sub.Child(fitNode(st, st.Bold(task.Slug), 8))
			}
			tr.Child(sub)
		}
		fmt.Fprintln(w, tr)
	}
	fmt.Fprintf(w, "\n%s", body)
	return nil
}

// AuditsHuman writes a table of audits with finding counts.
func AuditsHuman(w io.Writer, st Style, audits []domain.Audit) error {
	if len(audits) == 0 {
		return nil
	}
	rows := make([][]string, 0, len(audits))
	for _, a := range audits {
		bar := st.SegmentBar(a.DoneFindings, a.ActiveFindings, a.DroppedFindings, a.Findings, 8)
		progress := fmt.Sprintf("%s %s %s", bar, st.Percent(a.Percent()), theme.Counts(a.Resolved(), a.Findings))
		rows = append(rows, []string{st.Bucket(string(a.Bucket)), st.Bold(a.Slug), progress, a.Area})
	}
	writeTable(w, st.width, []string{st.Dim("BUCKET"), st.Dim("AUDIT"), st.Dim("PROGRESS"), st.Dim("AREA")}, rows)
	return nil
}

// AuditsJSON writes a versioned envelope of audits, including any per-file load
// problems (mirrors LintJSON's `unreadable`).
func AuditsJSON(w io.Writer, audits []domain.Audit, problems []domain.FileProblem) error {
	return wire.EncodeJSON(w, wire.ToAuditsEnvelope(audits, problems))
}

// findingStatusOrder renders the finding groups of `audit show` in lifecycle
// order (active work first, terminal states last). A status outside the
// vocabulary or missing entirely (audit lint flags those) sorts after these so
// the tree never drops a finding.
var findingStatusOrder = []string{"open", "in-progress", "fixed", "landed", "deferred", "superseded", "wontfix"}

// AuditShowHuman prints an audit's metadata, a status-grouped finding tree, and
// its body. findings is parsed from the raw body by the caller; body is the
// already-rendered (glamour/raw) markdown.
func AuditShowHuman(w io.Writer, st Style, a domain.Audit, findings []domain.Finding, body string) error {
	field := func(label, value string) {
		lbl := fmt.Sprintf("%-9s", label+":")
		if st.width > 0 { // fit the value to the terminal (TTY only; piped stays full)
			value = truncate(value, st.width-visibleWidth(lbl)-1)
		}
		fmt.Fprintf(w, "%s %s\n", st.Dim(lbl), value)
	}
	field("slug", st.Bold(a.Slug))
	field("bucket", st.Bucket(string(a.Bucket)))
	if a.Area != "" {
		field("area", a.Area)
	}
	if a.Date != "" {
		field("date", a.Date)
	}
	bar := st.SegmentBar(a.DoneFindings, a.ActiveFindings, a.DroppedFindings, a.Findings, 10)
	progress := fmt.Sprintf("%s %s  %s", bar, st.Percent(a.Percent()), theme.Counts(a.Resolved(), a.Findings))
	if a.OpenFindings > 0 {
		progress += fmt.Sprintf("  (%d open)", a.OpenFindings)
	}
	field("findings", progress)
	// A status-grouped finding tree (glyph + code + title), mirroring epic show's
	// task tree — the CLI analog of the TUI audit detail's finding index. The
	// --json envelope (AuditShowJSON) is unaffected.
	if len(findings) > 0 {
		byStatus := map[string][]domain.Finding{}
		for _, f := range findings {
			key := strings.ToLower(strings.TrimSpace(f.Status))
			byStatus[key] = append(byStatus[key], f)
		}
		tr := tree.New()
		// Fit nodes to the terminal width (ANSI-aware), accounting for the connector
		// indent — header ~4 cells, child ~8 — so a long finding title can't overflow.
		// Width 0 (piped) leaves them full.
		addGroup := func(header string, grp []domain.Finding) {
			sub := tree.Root(fitNode(st, header, 4))
			for _, f := range grp {
				child := st.Bold(f.Code)
				if f.Title != "" {
					child += "  " + f.Title
				}
				sub.Child(fitNode(st, child, 8))
			}
			tr.Child(sub)
		}
		done := map[string]bool{}
		for _, s := range findingStatusOrder {
			if grp := byStatus[s]; len(grp) > 0 {
				addGroup(st.FindingStatus(s), grp)
				done[s] = true
			}
		}
		// Out-of-vocab / missing statuses, in first-appearance order — no finding dropped.
		for _, f := range findings {
			key := strings.ToLower(strings.TrimSpace(f.Status))
			if done[key] {
				continue
			}
			done[key] = true
			header := st.FindingStatus(f.Status)
			if key == "" {
				header = st.Dim("(no status)")
			}
			addGroup(header, byStatus[key])
		}
		fmt.Fprintln(w, tr)
	}
	fmt.Fprintf(w, "\n%s", body)
	return nil
}

// AuditShowJSON writes an audit plus its body.
func AuditShowJSON(w io.Writer, a domain.Audit, body string) error {
	return wire.EncodeJSON(w, wire.ToAuditShowEnvelope(a, body))
}

// AuditMutationJSON writes the result of `audit append`: the reloaded audit, dry_run
// (always present — a preview must be distinguishable from a real write), and the
// resulting body. The audit counterpart to TaskMutationJSON.
func AuditMutationJSON(w io.Writer, a domain.Audit, body string, dryRun bool) error {
	return wire.EncodeJSON(w, wire.ToAuditMutationEnvelope(a, body, dryRun))
}

// FindingsJSON writes the structured finding-query result: each parsed finding
// tagged with its audit slug and bucket, so a cross-audit query stays
// self-describing. Mirrors the list envelopes' `unreadable` for per-file problems.
func FindingsJSON(w io.Writer, fs []core.AuditFinding, problems []domain.FileProblem) error {
	return wire.EncodeJSON(w, wire.ToFindingsEnvelope(fs, problems))
}

// FindingsHuman writes a scannable table of findings (empty input writes nothing).
// Title goes last so it's the column that truncates to terminal width.
func FindingsHuman(w io.Writer, st Style, fs []core.AuditFinding) error {
	if len(fs) == 0 {
		return nil
	}
	rows := make([][]string, 0, len(fs))
	for _, f := range fs {
		rows = append(rows, []string{st.FindingStatus(f.Status), st.Bold(f.Code), f.Audit, f.Effort, f.Urgency, f.Component, f.Title})
	}
	writeTable(w, st.width, []string{
		st.Dim("STATUS"), st.Dim("CODE"), st.Dim("AUDIT"), st.Dim("EFFORT"),
		st.Dim("URGENCY"), st.Dim("COMPONENT"), st.Dim("TITLE"),
	}, rows)
	fmt.Fprintf(w, "\n%s\n", st.Dim(plural(len(fs), "finding")))
	return nil
}

// FixHuman writes the auto-repairs applied (or proposed under --dry-run), then the
// leftover lint findings the pass could NOT repair (report-only epic issues,
// unfixable task issues) — so a human sees the residual breakage `--fix` left
// behind, not just what it touched. remaining is empty on a dry-run.
func FixHuman(w io.Writer, st Style, results []domain.FixResult, remaining []core.LintResult, dryRun bool) {
	verb := "fixed"
	if dryRun {
		verb = "would fix"
	}
	for _, r := range results {
		fmt.Fprintf(w, "%s %s\n", st.Green(verb), st.Bold(r.Path))
		for _, c := range r.Changes {
			fmt.Fprintf(w, "  %s %s\n", st.Dim("-"), c)
		}
	}
	if len(results) == 0 {
		fmt.Fprintln(w, st.Dim("nothing to fix"))
	} else {
		fmt.Fprintf(w, "\n%s\n", st.Dim(fmt.Sprintf("%d file(s) %s", len(results), verb)))
	}
	// What's still wrong after the pass — same per-entity rendering plain `lint`
	// uses (epics are report-only; some task issues aren't auto-fixable).
	if len(remaining) > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Dim("could not auto-repair:"))
		LintHuman(w, st, remaining, "item")
	}
}

// FixJSON writes the structured fix report: what was repaired (`fixed`), any files
// that still can't be read after the pass (`unreadable`), and the per-entity lint
// findings the pass could NOT repair (`remaining` — report-only epics, unfixable
// task issues). All three are empty on a dry-run (which writes nothing) — so a
// --json consumer learns the residual breakage without parsing the prose error.
func FixJSON(w io.Writer, results []domain.FixResult, problems []domain.FileProblem, remaining []core.LintResult, dryRun bool) error {
	return wire.EncodeJSON(w, wire.ToFixEnvelope(results, problems, remaining, dryRun))
}

// ProblemsHuman writes per-file load problems (unreadable frontmatter).
func ProblemsHuman(w io.Writer, st Style, problems []domain.FileProblem) {
	for _, p := range problems {
		fmt.Fprintf(w, "%s %s\n    %s\n", st.Red("!"), st.Bold(p.Path), p.Message)
	}
}

// LintHuman writes the per-entity lint findings + a count footer. noun names the
// entity for the footer ("task", "audit") since the same result/render shape backs
// both `lint` and `audit lint`.
func LintHuman(w io.Writer, st Style, results []core.LintResult, noun string) {
	for _, r := range results {
		fmt.Fprintf(w, "%s\n", st.Bold(r.Slug))
		for _, iss := range r.Issues {
			fmt.Fprintf(w, "  %s %s\n", st.Red(iss.Field+":"), iss.Message)
		}
	}
	if len(results) > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Dim(fmt.Sprintf("%d %s(s) with issues", len(results), noun)))
	}
}

// LintJSON writes the structured lint report: unreadable files + field issues.
func LintJSON(w io.Writer, results []core.LintResult, problems []domain.FileProblem) error {
	return wire.EncodeJSON(w, wire.ToLintEnvelope(results, problems))
}

// EpicShowJSON writes an epic, its tasks, and its body.
func EpicShowJSON(w io.Writer, epic domain.Epic, tasks []domain.Task, body string) error {
	return wire.EncodeJSON(w, wire.ToEpicShowEnvelope(epic, tasks, body))
}

// EpicMutationJSON writes the result of an `epic set`: the reloaded epic + dry_run
// (always present — a preview must be distinguishable from a real write). The epic
// counterpart to TaskMutationJSON; field-only, so there's no body to echo.
func EpicMutationJSON(w io.Writer, epic domain.Epic, dryRun bool) error {
	return wire.EncodeJSON(w, wire.ToEpicMutationEnvelope(epic, dryRun))
}

// InitEnvelope is the `init --json` payload (the wire type), re-exported so the
// CLI's init command keeps building it through the render package.
type InitEnvelope = wire.InitEnvelope

// InitJSON reports the init result. The caller fills the envelope's named fields
// (mode/root/planning_repo/linked_back/tracked/created); InitJSON stamps the
// schema_version and normalizes created to an empty array (not null) so a
// consumer can len() it.
func InitJSON(w io.Writer, e InitEnvelope) error {
	return wire.EncodeJSON(w, wire.NormalizeInitEnvelope(e))
}
