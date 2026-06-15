// Package render turns typed core results into output. It is the only place
// that knows about presentation; the core stays presentation-agnostic. Human
// output may use ANSI; JSON output never does.
package render

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
const SchemaVersion = "1.3"

type taskJSON struct {
	Slug        string   `json:"slug"`
	Status      string   `json:"status"`
	Epic        string   `json:"epic,omitempty"`
	Description string   `json:"description,omitempty"`
	Effort      string   `json:"effort,omitempty"`
	Tier        int      `json:"tier,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	Autonomy    int      `json:"autonomy_level,omitempty"`
	Created     string   `json:"created,omitempty"`
	Updated     string   `json:"updated_at,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	// Misfiled/Declared surface status≠folder drift to JSON consumers (agents
	// are exactly who should detect it); declared_status only when misfiled.
	Misfiled bool   `json:"misfiled,omitempty"`
	Declared string `json:"declared_status,omitempty"`
}

func toJSON(t domain.Task) taskJSON {
	j := taskJSON{
		Slug: t.Slug, Status: string(t.Status), Epic: t.Epic,
		Description: t.Description, Effort: t.Effort, Tier: t.Tier,
		Priority: t.Priority, Autonomy: t.Autonomy,
		Created: t.Created, Updated: t.Updated, Tags: t.Tags,
	}
	if t.Misfiled() {
		j.Misfiled = true
		j.Declared = string(t.Declared)
	}
	return j
}

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
	payload := struct {
		SchemaVersion string               `json:"schema_version"`
		Tasks         []taskJSON           `json:"tasks"`
		Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
	}{SchemaVersion: SchemaVersion, Tasks: make([]taskJSON, 0, len(tasks)), Unreadable: problems}
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
	return encodeJSON(w, struct {
		SchemaVersion string   `json:"schema_version"`
		Task          taskJSON `json:"task"`
		Body          string   `json:"body"`
	}{SchemaVersion: SchemaVersion, Task: toJSON(t), Body: body})
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
func MovesHuman(w io.Writer, st Style, results []MoveResult, dryRun bool) {
	verb := "moved"
	if dryRun {
		verb = "would move"
	}
	for _, r := range results {
		if r.Error != "" {
			fmt.Fprintf(w, "%s %s: %s\n", st.Red("✘"), st.Bold(r.Slug), r.Error)
		} else {
			fmt.Fprintf(w, "%s %s %s -> %s\n", st.Green("✔"), verb, st.Bold(r.Slug), r.To)
		}
	}
}

// MovesJSON writes the structured per-task transition report; dry_run marks a
// preview (nothing was written).
func MovesJSON(w io.Writer, results []MoveResult, dryRun bool) error {
	return encodeJSON(w, struct {
		SchemaVersion string       `json:"schema_version"`
		DryRun        bool         `json:"dry_run"`
		Moves         []MoveResult `json:"moves"`
	}{SchemaVersion: SchemaVersion, DryRun: dryRun, Moves: results})
}

func encodeJSON(w io.Writer, payload any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
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

type statusCountJSON struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
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
			Total:        e.Total, Done: e.Done, Percent: e.Percent(),
		})
	}
	return encodeJSON(w, struct {
		SchemaVersion string               `json:"schema_version"`
		Counts        []statusCountJSON    `json:"counts"`
		InProgress    []taskJSON           `json:"in_progress"`
		Epics         []epicJSON           `json:"epics"`
		Misfiled      int                  `json:"misfiled"`
		Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
	}{
		SchemaVersion: SchemaVersion, Counts: counts, InProgress: inprog,
		Epics: epics, Misfiled: s.Misfiled, Unreadable: s.Problems,
	})
}

// VersionHuman prints the CLI version.
func VersionHuman(w io.Writer, st Style, version string) {
	fmt.Fprintf(w, "%s %s\n", st.Bold("tskflwctl"), version)
}

// VersionJSON writes the version in the standard envelope.
func VersionJSON(w io.Writer, version string) error {
	return encodeJSON(w, struct {
		SchemaVersion string `json:"schema_version"`
		Version       string `json:"version"`
	}{SchemaVersion: SchemaVersion, Version: version})
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
// marks a preview (nothing was written).
func CreatedJSON(w io.Writer, kind, id, path string, dryRun bool) error {
	return encodeJSON(w, struct {
		SchemaVersion string `json:"schema_version"`
		DryRun        bool   `json:"dry_run"`
		Created       struct {
			Kind string `json:"kind"`
			ID   string `json:"id"`
			Path string `json:"path"`
		} `json:"created"`
	}{SchemaVersion: SchemaVersion, DryRun: dryRun, Created: struct {
		Kind string `json:"kind"`
		ID   string `json:"id"`
		Path string `json:"path"`
	}{Kind: kind, ID: id, Path: path}})
}

// EpicsHuman writes a table of epics with task rollup.
func EpicsHuman(w io.Writer, st Style, epics []core.EpicSummary) error {
	if len(epics) == 0 {
		return nil
	}
	rows := make([][]string, 0, len(epics))
	for _, e := range epics {
		progress := fmt.Sprintf("%d/%d (%s)", e.Done, e.Total, st.Percent(e.Percent()))
		rows = append(rows, []string{st.Bold(e.Epic.ID), e.Epic.Status, progress, e.Epic.Description})
	}
	writeTable(w, st.width, []string{st.Dim("EPIC"), st.Dim("STATUS"), st.Dim("PROGRESS"), st.Dim("DESCRIPTION")}, rows)
	return nil
}

// epicJSON is epic list output: the shared meta (embedded, so `epic list` and
// `epic show` can't drift) plus the task rollup.
type epicJSON struct {
	epicMetaJSON
	Total   int `json:"total"`
	Done    int `json:"done"`
	Percent int `json:"percent"`
}

// EpicsJSON writes a versioned envelope of epics with rollup, including any
// per-file load problems (mirrors LintJSON's `unreadable`).
func EpicsJSON(w io.Writer, epics []core.EpicSummary, problems []domain.FileProblem) error {
	payload := struct {
		SchemaVersion string               `json:"schema_version"`
		Epics         []epicJSON           `json:"epics"`
		Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
	}{SchemaVersion: SchemaVersion, Epics: make([]epicJSON, 0, len(epics)), Unreadable: problems}
	for _, e := range epics {
		payload.Epics = append(payload.Epics, epicJSON{
			epicMetaJSON: toEpicMeta(e.Epic),
			Total:        e.Total, Done: e.Done, Percent: e.Percent(),
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
	fmt.Fprintf(w, "%s\n", st.Dim(fmt.Sprintf("tasks (%d):", len(tasks))))
	rows := make([][]string, 0, len(tasks))
	for _, t := range tasks {
		rows = append(rows, []string{"  " + st.Status(t.Status), st.Bold(t.Slug)})
	}
	writeTable(w, st.width, nil, rows)
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
		findings := fmt.Sprintf("%d/%d open", a.OpenFindings, a.Findings)
		rows = append(rows, []string{st.Bucket(string(a.Bucket)), st.Bold(a.Slug), findings, a.Area})
	}
	writeTable(w, st.width, []string{st.Dim("BUCKET"), st.Dim("AUDIT"), st.Dim("FINDINGS"), st.Dim("AREA")}, rows)
	return nil
}

type auditJSON struct {
	Slug         string `json:"slug"`
	Bucket       string `json:"bucket"`
	Area         string `json:"area,omitempty"`
	Date         string `json:"date,omitempty"`
	Findings     int    `json:"findings"`
	OpenFindings int    `json:"open_findings"`
}

func auditToJSON(a domain.Audit) auditJSON {
	return auditJSON{
		Slug: a.Slug, Bucket: string(a.Bucket), Area: a.Area, Date: a.Date,
		Findings: a.Findings, OpenFindings: a.OpenFindings,
	}
}

// AuditsJSON writes a versioned envelope of audits, including any per-file load
// problems (mirrors LintJSON's `unreadable`).
func AuditsJSON(w io.Writer, audits []domain.Audit, problems []domain.FileProblem) error {
	payload := struct {
		SchemaVersion string               `json:"schema_version"`
		Audits        []auditJSON          `json:"audits"`
		Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
	}{SchemaVersion: SchemaVersion, Audits: make([]auditJSON, 0, len(audits)), Unreadable: problems}
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
	fmt.Fprintf(w, "%s %d (%d open)\n\n%s", st.Dim("findings:"), a.Findings, a.OpenFindings, body)
	return nil
}

// AuditShowJSON writes an audit plus its body.
func AuditShowJSON(w io.Writer, a domain.Audit, body string) error {
	return encodeJSON(w, struct {
		SchemaVersion string    `json:"schema_version"`
		Audit         auditJSON `json:"audit"`
		Body          string    `json:"body"`
	}{SchemaVersion: SchemaVersion, Audit: auditToJSON(a), Body: body})
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
	if problems == nil {
		problems = []domain.FileProblem{}
	}
	return encodeJSON(w, struct {
		SchemaVersion string               `json:"schema_version"`
		DryRun        bool                 `json:"dry_run"`
		Fixed         []domain.FixResult   `json:"fixed"`
		Unreadable    []domain.FileProblem `json:"unreadable"`
	}{SchemaVersion: SchemaVersion, DryRun: dryRun, Fixed: results, Unreadable: problems})
}

// ProblemsHuman writes per-file load problems (unreadable frontmatter).
func ProblemsHuman(w io.Writer, st Style, problems []domain.FileProblem) {
	for _, p := range problems {
		fmt.Fprintf(w, "%s %s\n    %s\n", st.Red("!"), st.Bold(p.Path), p.Message)
	}
}

// LintHuman writes the per-task lint findings + a count footer.
func LintHuman(w io.Writer, st Style, results []core.LintResult) {
	for _, r := range results {
		fmt.Fprintf(w, "%s\n", st.Bold(r.Slug))
		for _, iss := range r.Issues {
			fmt.Fprintf(w, "  %s %s\n", st.Red(iss.Field+":"), iss.Message)
		}
	}
	if len(results) > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Dim(fmt.Sprintf("%d task(s) with issues", len(results))))
	}
}

type lintTaskJSON struct {
	Slug   string         `json:"slug"`
	Issues []domain.Issue `json:"issues"`
}

// LintJSON writes the structured lint report: unreadable files + field issues.
func LintJSON(w io.Writer, results []core.LintResult, problems []domain.FileProblem) error {
	payload := struct {
		SchemaVersion string               `json:"schema_version"`
		Unreadable    []domain.FileProblem `json:"unreadable"`
		Issues        []lintTaskJSON       `json:"issues"`
	}{SchemaVersion: SchemaVersion, Unreadable: problems, Issues: make([]lintTaskJSON, 0, len(results))}
	for _, r := range results {
		payload.Issues = append(payload.Issues, lintTaskJSON{Slug: r.Slug, Issues: r.Issues})
	}
	return encodeJSON(w, payload)
}

type epicMetaJSON struct {
	ID          string   `json:"id"`
	Status      string   `json:"status,omitempty"`
	Description string   `json:"description,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	Created     string   `json:"created,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// toEpicMeta is the one place epic meta fields are mapped to JSON, shared by
// `epic list` (embedded in epicJSON) and `epic show`.
func toEpicMeta(e domain.Epic) epicMetaJSON {
	return epicMetaJSON{
		ID: e.ID, Status: e.Status, Description: e.Description,
		Priority: e.Priority, Created: e.Created, Tags: e.Tags,
	}
}

// EpicShowJSON writes an epic, its tasks, and its body.
func EpicShowJSON(w io.Writer, epic domain.Epic, tasks []domain.Task, body string) error {
	jt := make([]taskJSON, 0, len(tasks))
	for _, t := range tasks {
		jt = append(jt, toJSON(t))
	}
	return encodeJSON(w, struct {
		SchemaVersion string       `json:"schema_version"`
		Epic          epicMetaJSON `json:"epic"`
		Tasks         []taskJSON   `json:"tasks"`
		Body          string       `json:"body"`
	}{
		SchemaVersion: SchemaVersion,
		Epic:          toEpicMeta(epic),
		Tasks:         jt,
		Body:          body,
	})
}

// InitJSON reports the scaffold result; created is empty (not null) when the
// tree already existed, so consumers can len() it without a nil check.
func InitJSON(w io.Writer, root string, created []string, dryRun bool) error {
	if created == nil {
		created = []string{}
	}
	return encodeJSON(w, struct {
		SchemaVersion string   `json:"schema_version"`
		DryRun        bool     `json:"dry_run"`
		Root          string   `json:"root"`
		Created       []string `json:"created"`
	}{SchemaVersion, dryRun, root, created})
}
