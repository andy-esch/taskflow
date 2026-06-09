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
)

// SchemaVersion is the semver of the --json payloads. Adding a field bumps the
// minor; renaming/removing bumps the major.
const SchemaVersion = "1.0"

type taskJSON struct {
	Slug        string   `json:"slug"`
	Status      string   `json:"status"`
	Epic        string   `json:"epic,omitempty"`
	Description string   `json:"description,omitempty"`
	Tier        int      `json:"tier,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	Created     string   `json:"created,omitempty"`
	Updated     string   `json:"updated_at,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func toJSON(t domain.Task) taskJSON {
	return taskJSON{
		Slug: t.Slug, Status: string(t.Status), Epic: t.Epic,
		Description: t.Description, Tier: t.Tier, Priority: t.Priority,
		Created: t.Created, Updated: t.Updated, Tags: t.Tags,
	}
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
		rows = append(rows, []string{status, st.Bold(t.Slug), st.Dim(RelativeDate(updated)), t.Description})
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
		field("updated", fmt.Sprintf("%s %s", t.Updated, st.Dim("("+RelativeDate(t.Updated)+")")))
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

// MovesHuman prints one line per transition outcome.
func MovesHuman(w io.Writer, st Style, results []MoveResult) {
	for _, r := range results {
		if r.Error != "" {
			fmt.Fprintf(w, "%s %s: %s\n", st.Red("✘"), st.Bold(r.Slug), r.Error)
		} else {
			fmt.Fprintf(w, "%s moved %s -> %s\n", st.Green("✔"), st.Bold(r.Slug), r.To)
		}
	}
}

// MovesJSON writes the structured per-task transition report.
func MovesJSON(w io.Writer, results []MoveResult) error {
	return encodeJSON(w, struct {
		SchemaVersion string       `json:"schema_version"`
		Moves         []MoveResult `json:"moves"`
	}{SchemaVersion: SchemaVersion, Moves: results})
}

func encodeJSON(w io.Writer, payload any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
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

// CreatedHuman prints the path of a newly created file.
func CreatedHuman(w io.Writer, st Style, path string) {
	fmt.Fprintf(w, "%s %s\n", st.Green("created"), st.Bold(path))
}

// CreatedJSON writes a versioned envelope for a newly created item.
func CreatedJSON(w io.Writer, kind, id, path string) error {
	return encodeJSON(w, struct {
		SchemaVersion string `json:"schema_version"`
		Created       struct {
			Kind string `json:"kind"`
			ID   string `json:"id"`
			Path string `json:"path"`
		} `json:"created"`
	}{SchemaVersion: SchemaVersion, Created: struct {
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

type epicJSON struct {
	ID          string `json:"id"`
	Status      string `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
	Total       int    `json:"total"`
	Done        int    `json:"done"`
	Percent     int    `json:"percent"`
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
			ID: e.Epic.ID, Status: e.Epic.Status, Description: e.Epic.Description,
			Total: e.Total, Done: e.Done, Percent: e.Percent(),
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

// FixJSON writes the structured fix report.
func FixJSON(w io.Writer, results []domain.FixResult, dryRun bool) error {
	return encodeJSON(w, struct {
		SchemaVersion string             `json:"schema_version"`
		DryRun        bool               `json:"dry_run"`
		Fixed         []domain.FixResult `json:"fixed"`
	}{SchemaVersion: SchemaVersion, DryRun: dryRun, Fixed: results})
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
	ID          string `json:"id"`
	Status      string `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
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
		Epic:          epicMetaJSON{ID: epic.ID, Status: epic.Status, Description: epic.Description},
		Tasks:         jt,
		Body:          body,
	})
}
