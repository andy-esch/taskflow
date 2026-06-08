// Package render turns typed core results into output. It is the only place
// that knows about presentation; the core stays presentation-agnostic. Human
// output may use ANSI; JSON output never does.
package render

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

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
	Tags        []string `json:"tags,omitempty"`
}

func toJSON(t domain.Task) taskJSON {
	return taskJSON{
		Slug: t.Slug, Status: string(t.Status), Epic: t.Epic,
		Description: t.Description, Tier: t.Tier, Priority: t.Priority, Tags: t.Tags,
	}
}

// TasksHuman writes a scannable table of tasks.
func TasksHuman(w io.Writer, tasks []domain.Task) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, t := range tasks {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", t.Status, t.Slug, t.Description)
	}
	return tw.Flush()
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
func TaskShowHuman(w io.Writer, t domain.Task, body string) error {
	fmt.Fprintf(w, "slug:        %s\n", t.Slug)
	fmt.Fprintf(w, "status:      %s\n", t.Status)
	if t.Epic != "" {
		fmt.Fprintf(w, "epic:        %s\n", t.Epic)
	}
	if t.Priority != "" {
		fmt.Fprintf(w, "priority:    %s\n", t.Priority)
	}
	if t.Tier != 0 {
		fmt.Fprintf(w, "tier:        %d\n", t.Tier)
	}
	if len(t.Tags) > 0 {
		fmt.Fprintf(w, "tags:        %s\n", strings.Join(t.Tags, ", "))
	}
	if t.Description != "" {
		fmt.Fprintf(w, "description: %s\n", t.Description)
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

// MoveResult is the per-task outcome of a transition.
type MoveResult struct {
	Slug   string `json:"slug"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// MovesHuman prints one line per transition outcome.
func MovesHuman(w io.Writer, results []MoveResult) {
	for _, r := range results {
		if r.Error != "" {
			fmt.Fprintf(w, "x %s: %s\n", r.Slug, r.Error)
		} else {
			fmt.Fprintf(w, "moved %s -> %s\n", r.Slug, r.Status)
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

// EpicsHuman writes a table of epics with task rollup.
func EpicsHuman(w io.Writer, epics []core.EpicSummary) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, e := range epics {
		fmt.Fprintf(tw, "%s\t%s\t%d/%d (%d%%)\t%s\n",
			e.Epic.ID, e.Epic.Status, e.Done, e.Total, e.Percent(), e.Epic.Description)
	}
	return tw.Flush()
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
func EpicShowHuman(w io.Writer, epic domain.Epic, tasks []domain.Task, body string) error {
	fmt.Fprintf(w, "id:          %s\n", epic.ID)
	fmt.Fprintf(w, "status:      %s\n", epic.Status)
	if epic.Description != "" {
		fmt.Fprintf(w, "description: %s\n", epic.Description)
	}
	fmt.Fprintf(w, "tasks (%d):\n", len(tasks))
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, t := range tasks {
		fmt.Fprintf(tw, "  %s\t%s\n", t.Status, t.Slug)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	fmt.Fprintf(w, "\n%s", body)
	return nil
}

// AuditsHuman writes a table of audits with finding counts.
func AuditsHuman(w io.Writer, audits []domain.Audit) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, a := range audits {
		fmt.Fprintf(tw, "%s\t%s\t%d/%d open\t%s\n", a.Bucket, a.Slug, a.OpenFindings, a.Findings, a.Area)
	}
	return tw.Flush()
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
func AuditShowHuman(w io.Writer, a domain.Audit, body string) error {
	fmt.Fprintf(w, "slug:     %s\n", a.Slug)
	fmt.Fprintf(w, "bucket:   %s\n", a.Bucket)
	if a.Area != "" {
		fmt.Fprintf(w, "area:     %s\n", a.Area)
	}
	if a.Date != "" {
		fmt.Fprintf(w, "date:     %s\n", a.Date)
	}
	fmt.Fprintf(w, "findings: %d (%d open)\n\n%s", a.Findings, a.OpenFindings, body)
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
func FixHuman(w io.Writer, results []domain.FixResult, dryRun bool) {
	verb := "fixed"
	if dryRun {
		verb = "would fix"
	}
	for _, r := range results {
		fmt.Fprintf(w, "%s %s\n", verb, r.Path)
		for _, c := range r.Changes {
			fmt.Fprintf(w, "  - %s\n", c)
		}
	}
	if len(results) == 0 {
		fmt.Fprintln(w, "nothing to fix")
	} else {
		fmt.Fprintf(w, "\n%d file(s) %s\n", len(results), verb)
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
func ProblemsHuman(w io.Writer, problems []domain.FileProblem) {
	for _, p := range problems {
		fmt.Fprintf(w, "! %s\n    %s\n", p.Path, p.Message)
	}
}

// LintHuman writes the per-task lint findings + a count footer.
func LintHuman(w io.Writer, results []core.LintResult) {
	for _, r := range results {
		fmt.Fprintf(w, "%s\n", r.Slug)
		for _, iss := range r.Issues {
			fmt.Fprintf(w, "  %s: %s\n", iss.Field, iss.Message)
		}
	}
	if len(results) > 0 {
		fmt.Fprintf(w, "\n%d task(s) with issues\n", len(results))
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
