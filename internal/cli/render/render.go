// Package render turns typed core results into output. It is the only place
// that knows about presentation; the core stays presentation-agnostic. Human
// output may use ANSI; JSON output never does.
package render

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2/tree"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// SchemaVersion is the semver of the --json payloads — ONE version for the
// whole CLI output schema, not per envelope (decided 2026-06-12). Adding a
// field bumps the minor; renaming/removing bumps the major. Key naming rule:
// JSON keys match the frontmatter keys exactly (`created`, `updated_at`).
// 1.1: every CLI-settable field round-trips (effort, autonomy_level), and the
// misfiled signal (previously human-output-only ⚠) is machine-readable.
// 1.2: mutation envelopes carry dry_run:true under --dry-run previews.
// 1.3: dry_run is always present on mutation envelopes (was omitted when false);
// the fix report carries `unreadable` (files it couldn't repair).
// 1.4: `schema` envelopes (the tool's self-description contract + per-kind
// authoring guidance) added.
// 1.5: the create envelope carries `status` (task status / epic status / audit
// bucket); its `path` is now relative to the planning root in both human and
// JSON modes (was absolute in JSON).
// 1.6: the `findings` envelope (audit finding-level query) added.
// 1.7: the `task_mutation` envelope (task set/append/set --body) added — it
// carries dry_run and the resulting body, which `task_show` (a read) does not.
// 1.8: the `init` envelope carries `mode` (scaffold|pointer) and `planning_repo`
// (set in pointer mode), for `init --planning-repo`.
// 1.9: the `doctor` envelope (planning_repo <-> tracked_repos linkback audit) added.
// 1.10: the `init` envelope carries `linked_back` (pointer-mode auto-link-back
// path) and `tracked` (scaffold-mode --track entries).
// 1.11: epic rollups exclude deprecated (withdrawn) tasks from total/done; the
// epic payload carries a separate `deprecated` count.
// 1.12: the `status` summary envelope carries `open_audits` — open-bucket audits
// (the actionable subset) with the same finding rollup `audit list` reports;
// omitted when there are none.
const SchemaVersion = "1.12"

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
	payload := TasksEnvelope{SchemaVersion: SchemaVersion, Tasks: make([]taskJSON, 0, len(tasks)), Unreadable: problems}
	for _, t := range tasks {
		payload.Tasks = append(payload.Tasks, toJSON(t))
	}
	return encodeJSON(w, payload)
}

// TaskShowHuman prints a task's metadata followed by its body.
func TaskShowHuman(w io.Writer, st Style, t domain.Task, body string) error {
	field := func(label, value string) {
		fmt.Fprintf(w, "%s %s\n", st.Dim(fmt.Sprintf("%-12s", label+":")), value)
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
	return encodeJSON(w, TaskShowEnvelope{SchemaVersion: SchemaVersion, Task: toJSON(t), Body: body})
}

// TaskMutationJSON writes the result of a task mutation (`task set`/`append`/`set
// --body`): the reloaded task, dry_run (always present — a preview must be
// distinguishable from a real write), and the resulting body for the body-editing
// commands (empty/omitted for field-only `set`). Distinct from TaskShowEnvelope so
// the mutation-only dry_run never lands on the `task show` read type.
func TaskMutationJSON(w io.Writer, t domain.Task, body string, dryRun bool) error {
	return encodeJSON(w, TaskMutationEnvelope{SchemaVersion: SchemaVersion, DryRun: dryRun, Task: toJSON(t), Body: body})
}

// MoveResult is the per-item outcome of a transition. `To` is the destination
// state — a task status or an audit bucket — so the JSON key is the neutral
// "to" rather than "status".
type MoveResult struct {
	Slug  string `json:"slug"`
	To    string `json:"to"`
	Error string `json:"error,omitempty"`
}

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
			fmt.Fprintf(out, "%s %s %s -> %s\n", st.Green("✔"), verb, st.Bold(r.Slug), r.To)
		}
	}
}

// MovesJSON writes the structured per-task transition report; dry_run marks a
// preview (nothing was written).
func MovesJSON(w io.Writer, results []MoveResult, dryRun bool) error {
	if results == nil {
		results = []MoveResult{} // empty, not null — schema is type: array (see FixJSON)
	}
	return encodeJSON(w, MovesEnvelope{SchemaVersion: SchemaVersion, DryRun: dryRun, Moves: results})
}

// encodeJSON writes the payload as compact (un-indented) JSON with a single
// trailing newline. Machine output: pretty-printing is pure token cost for a
// consumer that parses it. Off-tree consumers pipe through `jq .` to read it.
func encodeJSON(w io.Writer, payload any) error {
	enc := json.NewEncoder(w)
	return enc.Encode(payload)
}

// DoctorJSON writes the linkback audit; problems is empty (not null) when the
// links are consistent, so a consumer can len() it without a nil check.
func DoctorJSON(w io.Writer, root string, problems []DoctorProblem) error {
	if problems == nil {
		problems = []DoctorProblem{}
	}
	return encodeJSON(w, DoctorEnvelope{SchemaVersion: SchemaVersion, Root: root, Problems: problems})
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
			rows = append(rows, []string{"  " + st.Bold(e.Epic.ID), bar, fmt.Sprintf("%d/%d", e.Done, e.Total), e.Epic.Description})
		}
		writeTable(w, st.width, nil, rows)
	}

	// Only open audits, only when there are any — the actionable subset, rendered
	// with the same bar treatment as epics so the dashboard reads from one vocabulary.
	if len(s.OpenAudits) > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Bold(fmt.Sprintf("Open audits (%d)", len(s.OpenAudits))))
		rows := make([][]string, 0, len(s.OpenAudits))
		for _, a := range s.OpenAudits {
			bar := fmt.Sprintf("%s %s", st.Bar(a.Percent(), 10), st.Percent(a.Percent()))
			rows = append(rows, []string{"  " + st.Bold(a.Slug), bar, fmt.Sprintf("%d/%d", a.Resolved(), a.Findings), a.Area})
		}
		writeTable(w, st.width, nil, rows)
	}

	if s.Misfiled > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Warn(fmt.Sprintf("⚠ %d misfiled (status ≠ folder; run `lint --fix`)", s.Misfiled)))
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
	counts := make([]statusCountJSON, 0, len(s.Counts))
	for _, c := range s.Counts {
		counts = append(counts, statusCountJSON{Status: string(c.Status), Count: c.Count})
	}
	inprog := make([]taskJSON, 0, len(s.InProgress))
	for _, t := range s.InProgress {
		inprog = append(inprog, toJSON(t))
	}
	epics := make([]epicJSON, 0, len(s.Epics))
	for _, e := range s.Epics {
		epics = append(epics, epicJSON{
			epicMetaJSON: toEpicMeta(e.Epic),
			Total:        e.Total, Done: e.Done, Percent: e.Percent(), Deprecated: e.Deprecated,
		})
	}
	// open_audits is omitempty: absent unless there's actionable audit work, so a
	// repo with none sees no envelope change (the human dashboard self-hides too).
	audits := make([]auditJSON, 0, len(s.OpenAudits))
	for _, a := range s.OpenAudits {
		audits = append(audits, auditToJSON(a))
	}
	return encodeJSON(w, SummaryEnvelope{
		SchemaVersion: SchemaVersion, Counts: counts, InProgress: inprog,
		Epics: epics, OpenAudits: audits, Misfiled: s.Misfiled, Unreadable: s.Problems,
	})
}

// VersionHuman prints the CLI version.
func VersionHuman(w io.Writer, st Style, version string) {
	fmt.Fprintf(w, "%s %s\n", st.Bold("tskflwctl"), version)
}

// VersionJSON writes the version in the standard envelope.
func VersionJSON(w io.Writer, version string) error {
	return encodeJSON(w, VersionEnvelope{SchemaVersion: SchemaVersion, Version: version})
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

// CreatedJSON writes a versioned envelope for a newly created item; dry_run
// marks a preview (nothing was written). status is the new item's status (task
// status / epic status / audit bucket); path is relative to the planning root,
// matching the human output.
func CreatedJSON(w io.Writer, kind, id, status, path string, dryRun bool) error {
	return encodeJSON(w, CreatedEnvelope{SchemaVersion: SchemaVersion, DryRun: dryRun, Created: CreatedItem{Kind: kind, ID: id, Status: status, Path: path}})
}

// EpicsHuman writes a table of epics with task rollup.
func EpicsHuman(w io.Writer, st Style, epics []core.EpicSummary) error {
	if len(epics) == 0 {
		return nil
	}
	rows := make([][]string, 0, len(epics))
	for _, e := range epics {
		pct := e.Percent()
		progress := fmt.Sprintf("%s %s %d/%d", st.Bar(pct, 8), st.Percent(pct), e.Done, e.Total)
		rows = append(rows, []string{st.Bold(e.Epic.ID), e.Epic.Status, progress, e.Epic.Description})
	}
	writeTable(w, st.width, []string{st.Dim("EPIC"), st.Dim("STATUS"), st.Dim("PROGRESS"), st.Dim("DESCRIPTION")}, rows)
	return nil
}

// EpicsJSON writes a versioned envelope of epics with rollup, including any
// per-file load problems (mirrors LintJSON's `unreadable`).
func EpicsJSON(w io.Writer, epics []core.EpicSummary, problems []domain.FileProblem) error {
	payload := EpicsEnvelope{SchemaVersion: SchemaVersion, Epics: make([]epicJSON, 0, len(epics)), Unreadable: problems}
	for _, e := range epics {
		payload.Epics = append(payload.Epics, epicJSON{
			epicMetaJSON: toEpicMeta(e.Epic),
			Total:        e.Total, Done: e.Done, Percent: e.Percent(), Deprecated: e.Deprecated,
		})
	}
	return encodeJSON(w, payload)
}

// EpicShowHuman prints an epic, its tasks, and its body.
func EpicShowHuman(w io.Writer, st Style, epic domain.Epic, tasks []domain.Task, body string) error {
	field := func(label, value string) {
		fmt.Fprintf(w, "%s %s\n", st.Dim(fmt.Sprintf("%-12s", label+":")), value)
	}
	field("id", st.Bold(epic.ID))
	field("status", epic.Status)
	if epic.Description != "" {
		field("description", epic.Description)
	}
	// One pass for the rollup: deprecated tasks leave the denominator (counted
	// separately in the tasks header), matching epic list / status / the TUI detail.
	done, total, deprecated := 0, 0, 0
	for _, t := range tasks {
		if t.Status == domain.StatusDeprecated {
			deprecated++
			continue
		}
		total++
		if t.Status == domain.StatusCompleted {
			done++
		}
	}
	pct := 0
	if total > 0 {
		pct = done * 100 / total
	}
	field("progress", fmt.Sprintf("%s %s  %d/%d", st.Bar(pct, 10), st.Percent(pct), done, total))
	header := fmt.Sprintf("tasks (%d):", len(tasks))
	if deprecated > 0 {
		// Note the withdrawn count — those tasks are listed but excluded from the
		// done/total rollup shown by `epic list`/`status`.
		header = fmt.Sprintf("tasks (%d, %d deprecated — excluded from progress):", len(tasks), deprecated)
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
			sub := tree.Root(st.Status(s))
			for _, task := range grp {
				sub.Child(st.Bold(task.Slug))
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
		pct := a.Percent()
		progress := fmt.Sprintf("%s %s %d/%d", st.Bar(pct, 8), st.Percent(pct), a.Resolved(), a.Findings)
		rows = append(rows, []string{st.Bucket(string(a.Bucket)), st.Bold(a.Slug), progress, a.Area})
	}
	writeTable(w, st.width, []string{st.Dim("BUCKET"), st.Dim("AUDIT"), st.Dim("PROGRESS"), st.Dim("AREA")}, rows)
	return nil
}

// AuditsJSON writes a versioned envelope of audits, including any per-file load
// problems (mirrors LintJSON's `unreadable`).
func AuditsJSON(w io.Writer, audits []domain.Audit, problems []domain.FileProblem) error {
	payload := AuditsEnvelope{SchemaVersion: SchemaVersion, Audits: make([]auditJSON, 0, len(audits)), Unreadable: problems}
	for _, a := range audits {
		payload.Audits = append(payload.Audits, auditToJSON(a))
	}
	return encodeJSON(w, payload)
}

// AuditShowHuman prints an audit's metadata and body.
func AuditShowHuman(w io.Writer, st Style, a domain.Audit, body string) error {
	field := func(label, value string) {
		fmt.Fprintf(w, "%s %s\n", st.Dim(fmt.Sprintf("%-9s", label+":")), value)
	}
	field("slug", st.Bold(a.Slug))
	field("bucket", st.Bucket(string(a.Bucket)))
	if a.Area != "" {
		field("area", a.Area)
	}
	if a.Date != "" {
		field("date", a.Date)
	}
	pct := a.Percent()
	progress := fmt.Sprintf("%s %s  %d/%d", st.Bar(pct, 10), st.Percent(pct), a.Resolved(), a.Findings)
	if a.OpenFindings > 0 {
		progress += fmt.Sprintf("  (%d open)", a.OpenFindings)
	}
	field("findings", progress)
	fmt.Fprintf(w, "\n%s", body)
	return nil
}

// AuditShowJSON writes an audit plus its body.
func AuditShowJSON(w io.Writer, a domain.Audit, body string) error {
	return encodeJSON(w, AuditShowEnvelope{SchemaVersion: SchemaVersion, Audit: auditToJSON(a), Body: body})
}

// FindingsJSON writes the structured finding-query result: each parsed finding
// tagged with its audit slug and bucket, so a cross-audit query stays
// self-describing. Mirrors the list envelopes' `unreadable` for per-file problems.
func FindingsJSON(w io.Writer, fs []core.AuditFinding, problems []domain.FileProblem) error {
	payload := FindingsEnvelope{SchemaVersion: SchemaVersion, Findings: make([]findingJSON, 0, len(fs)), Unreadable: problems}
	for _, f := range fs {
		payload.Findings = append(payload.Findings, findingJSON{
			Audit: f.Audit, Bucket: f.Bucket, Code: f.Code, Title: f.Title, Status: f.Status,
			File: f.File, Component: f.Component, Effort: f.Effort, Urgency: f.Urgency,
		})
	}
	return encodeJSON(w, payload)
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

// FixHuman writes the auto-repairs applied (or proposed under --dry-run).
func FixHuman(w io.Writer, st Style, results []domain.FixResult, dryRun bool) {
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
}

// FixJSON writes the structured fix report: what was repaired, plus any files
// that still can't be read after the pass (empty on a dry-run, which writes
// nothing) — so a --json consumer learns the residual breakage without parsing
// the prose error.
func FixJSON(w io.Writer, results []domain.FixResult, problems []domain.FileProblem, dryRun bool) error {
	// Empty (not null) for the array fields, so a consumer can len() without a nil
	// check — and so the output validates against its own schema (type: array).
	if problems == nil {
		problems = []domain.FileProblem{}
	}
	if results == nil {
		results = []domain.FixResult{}
	}
	return encodeJSON(w, FixEnvelope{SchemaVersion: SchemaVersion, DryRun: dryRun, Fixed: results, Unreadable: problems})
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
	if problems == nil {
		problems = []domain.FileProblem{} // empty, not null (see FixJSON) — schema is type: array
	}
	payload := LintEnvelope{SchemaVersion: SchemaVersion, Unreadable: problems, Issues: make([]lintTaskJSON, 0, len(results))}
	for _, r := range results {
		issues := r.Issues
		if issues == nil {
			issues = []domain.Issue{} // empty, not null — the per-row issues are type: array too
		}
		payload.Issues = append(payload.Issues, lintTaskJSON{Slug: r.Slug, Issues: issues})
	}
	return encodeJSON(w, payload)
}

// EpicShowJSON writes an epic, its tasks, and its body.
func EpicShowJSON(w io.Writer, epic domain.Epic, tasks []domain.Task, body string) error {
	jt := make([]taskJSON, 0, len(tasks))
	for _, t := range tasks {
		jt = append(jt, toJSON(t))
	}
	return encodeJSON(w, EpicShowEnvelope{
		SchemaVersion: SchemaVersion,
		Epic:          toEpicMeta(epic),
		Tasks:         jt,
		Body:          body,
	})
}

// InitJSON reports the init result. The caller fills the envelope's named fields
// (mode/root/planning_repo/linked_back/tracked/created); InitJSON stamps the
// schema_version and normalizes created to an empty array (not null) so a
// consumer can len() it.
func InitJSON(w io.Writer, e InitEnvelope) error {
	e.SchemaVersion = SchemaVersion
	if e.Created == nil {
		e.Created = []string{}
	}
	return encodeJSON(w, e)
}
