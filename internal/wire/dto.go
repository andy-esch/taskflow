package wire

import (
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// This file holds the JSON DTOs and their mappers — the wire shape of each entity
// inside the --json envelopes (envelopes.go). Each ToXEnvelope constructor maps
// domain types into these and returns an envelope that embeds them; keeping the
// DTOs + mappers here lets both the CLI's emit funcs and a web adapter project the
// same shape.
//
// Field schema descriptions live in the `jsonschema:"description=…"` struct tags —
// the reflector's intended mechanism, and the one that yields a clean, precise
// machine-facing string. A field's Go doc comment is godoc-only: where a tag is
// present it wins, so a maintainer note in a comment never leaks into the wire
// contract. (Type-level descriptions, by contrast, can only come from doc comments,
// harvested into schema_comments.json.)

// TaskJSON is the wire shape of a task inside the --json envelopes.
type TaskJSON struct {
	Slug   string `json:"slug" jsonschema:"description=task identifier (filename without .md)"`
	Status string `json:"status" jsonschema:"description=lifecycle status — equals the task's directory under tasks/"`
	Epic   string `json:"epic,omitempty" jsonschema:"description=id of the epic this task belongs to"`
	// The "<=200" cap can't be computed (struct tags are static literals) — the only
	// hardcoded copy of domain.MaxDescriptionLen left. Kept honest by
	// TestTaskJSONDescriptionTagMatchesCap; update both if the cap changes.
	Description string   `json:"description,omitempty" jsonschema:"description=one-line summary (<=200 chars)"`
	Effort      string   `json:"effort,omitempty" jsonschema:"description=free-form effort estimate"`
	Tier        int      `json:"tier,omitempty" jsonschema:"description=importance 1 (highest) to 5 (lowest)"`
	Priority    string   `json:"priority,omitempty" jsonschema:"description=high | medium | low"`
	Autonomy    int      `json:"autonomy_level,omitempty" jsonschema:"description=how autonomously this can be done 1-5"`
	Created     string   `json:"created,omitempty" jsonschema:"description=creation date YYYY-MM-DD"`
	Updated     string   `json:"updated_at,omitempty" jsonschema:"description=last-modified date YYYY-MM-DD"`
	RevisitAt   string   `json:"revisit_at,omitempty" jsonschema:"description=snooze-until date YYYY-MM-DD for a deferred task (set by task defer)"`
	Tags        []string `json:"tags,omitempty" jsonschema:"description=topical tags"`
	// Misfiled/Declared surface status≠folder drift to JSON consumers (agents
	// are exactly who should detect it); declared_status only when misfiled.
	Misfiled bool   `json:"misfiled,omitempty"`
	Declared string `json:"declared_status,omitempty"`
}

// ToTaskJSON maps a domain task to its wire DTO.
func ToTaskJSON(t domain.Task) TaskJSON {
	j := TaskJSON{
		Slug: t.Slug, Status: string(t.Status), Epic: t.Epic,
		Description: t.Description, Effort: t.Effort, Tier: t.Tier,
		Priority: t.Priority, Autonomy: t.Autonomy,
		Created: t.Created, Updated: t.Updated, RevisitAt: t.RevisitAt, Tags: t.Tags,
	}
	if t.Misfiled() {
		j.Misfiled = true
		j.Declared = string(t.Declared)
	}
	return j
}

// StatusCountJSON is one status bucket and its task count.
type StatusCountJSON struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// EpicJSON is epic list output: the shared meta (embedded, so `epic list` and
// `epic show` can't drift) plus the task rollup.
type EpicJSON struct {
	EpicMetaJSON
	Total int `json:"total"`
	Done  int `json:"done"`
	// Open is the pending workload (total − done); 0 = nothing in flight.
	Open    int `json:"open"`
	Percent int `json:"percent"`
	// Deprecated is the withdrawn tasks, excluded from total/done.
	Deprecated int `json:"deprecated"`
	// Liveness is the derived activity band — computed from the rollup, not stored.
	Liveness string `json:"liveness" jsonschema:"description=derived activity: working | fresh | dormant"`
}

// ToEpicJSON maps a core epic summary to the epic list/rollup DTO.
func ToEpicJSON(e core.EpicSummary) EpicJSON {
	return EpicJSON{
		EpicMetaJSON: ToEpicMeta(e.Epic),
		Total:        e.Total, Done: e.Done, Open: e.Open(), Percent: e.Percent(),
		Deprecated: e.Deprecated, Liveness: string(e.Liveness()),
	}
}

// AuditJSON is the wire shape of an audit inside the --json envelopes.
type AuditJSON struct {
	Slug         string `json:"slug" jsonschema:"description=audit identifier (filename without .md)"`
	Bucket       string `json:"bucket" jsonschema:"description=open | closed | deferred — equals the audit's directory"`
	Area         string `json:"area,omitempty" jsonschema:"description=subsystem/topic audited"`
	Date         string `json:"date,omitempty" jsonschema:"description=audit date YYYY-MM-DD"`
	Findings     int    `json:"findings" jsonschema:"description=total findings parsed from the body"`
	OpenFindings int    `json:"open_findings" jsonschema:"description=findings whose status is open"`
	// The progress bar's disposition bands. open + in_progress + done + dropped ≤
	// findings (an unrecognized/missing status counts toward none).
	InProgressFindings int `json:"in_progress_findings" jsonschema:"description=findings whose status is in-progress"`
	DoneFindings       int `json:"done_findings" jsonschema:"description=findings whose status is fixed or landed (the bar's done band)"`
	DroppedFindings    int `json:"dropped_findings" jsonschema:"description=findings whose status is deferred, superseded, or wontfix"`
	// ReadyToClose is true for an OPEN audit whose findings are all resolved/dropped
	// (none open or in-progress) — a "ready to close" call-to-action.
	ReadyToClose bool `json:"ready_to_close,omitempty" jsonschema:"description=true when an open audit has no open/in-progress findings left (ready to close)"`
}

// ToAuditJSON maps a domain audit to its wire DTO.
func ToAuditJSON(a domain.Audit) AuditJSON {
	return AuditJSON{
		Slug: a.Slug, Bucket: string(a.Bucket), Area: a.Area, Date: a.Date,
		Findings: a.Findings, OpenFindings: a.OpenFindings,
		InProgressFindings: a.ActiveFindings, DoneFindings: a.DoneFindings, DroppedFindings: a.DroppedFindings,
		ReadyToClose: a.Bucket == domain.AuditOpen && a.Settled(),
	}
}

// FindingJSON is the wire shape of one audit finding.
type FindingJSON struct {
	Audit     string `json:"audit" jsonschema:"description=slug of the audit this finding belongs to"`
	Bucket    string `json:"bucket" jsonschema:"description=the audit's bucket — open | closed | deferred"`
	Code      string `json:"code" jsonschema:"description=finding code within the audit (H1/M2/S3…)"`
	Title     string `json:"title" jsonschema:"description=finding title"`
	Status    string `json:"status" jsonschema:"description=open | in-progress | fixed | landed | deferred | superseded | wontfix"`
	File      string `json:"file,omitempty" jsonschema:"description=file:line the finding refers to"`
	Component string `json:"component,omitempty" jsonschema:"description=component/subsystem"`
	Effort    string `json:"effort,omitempty" jsonschema:"description=XS | S | M | L"`
	Urgency   string `json:"urgency,omitempty" jsonschema:"description=acute | soon | eventually"`
}

// ToFindingJSON maps a core audit finding to its wire DTO.
func ToFindingJSON(f core.AuditFinding) FindingJSON {
	return FindingJSON{
		Audit: f.Audit, Bucket: f.Bucket, Code: f.Code, Title: f.Title, Status: f.Status,
		File: f.File, Component: f.Component, Effort: f.Effort, Urgency: f.Urgency,
	}
}

// CountByJSON is one bucket of a finding breakdown — an urgency value or a
// top-level component, and its count.
type CountByJSON struct {
	Key   string `json:"key" jsonschema:"description=the urgency value or top-level component"`
	Count int    `json:"count" jsonschema:"description=actionable findings in this bucket"`
}

// FindingsRollupJSON aggregates the actionable audit findings (status open or
// in-progress) across all audits — the `status` summary's "audit findings" view.
type FindingsRollupJSON struct {
	Open        int           `json:"open" jsonschema:"description=actionable findings with status open"`
	InProgress  int           `json:"in_progress" jsonschema:"description=actionable findings with status in-progress"`
	ByUrgency   []CountByJSON `json:"by_urgency,omitempty" jsonschema:"description=breakdown by urgency (acute, soon, eventually first)"`
	ByComponent []CountByJSON `json:"by_component,omitempty" jsonschema:"description=breakdown by top-level component, most findings first"`
	Acute       []FindingJSON `json:"acute,omitempty" jsonschema:"description=the acute findings, listed for a call-out"`
}

// ToFindingsRollup maps a core findings rollup to its wire DTO.
func ToFindingsRollup(r core.FindingsRollup) FindingsRollupJSON {
	out := FindingsRollupJSON{Open: r.Open, InProgress: r.InProgress}
	for _, c := range r.ByUrgency {
		out.ByUrgency = append(out.ByUrgency, CountByJSON{Key: c.Key, Count: c.Count})
	}
	for _, c := range r.ByComponent {
		out.ByComponent = append(out.ByComponent, CountByJSON{Key: c.Key, Count: c.Count})
	}
	for _, f := range r.Acute {
		out.Acute = append(out.Acute, ToFindingJSON(f))
	}
	return out
}

// LintTaskJSON is one entity's lint result (slug + field issues), the shape the
// `lint` / `audit lint` / `fix --remaining` envelopes carry per entity.
type LintTaskJSON struct {
	Slug   string         `json:"slug"`
	Issues []domain.Issue `json:"issues"`
}

// EpicMetaJSON is the shared epic meta fields, embedded by EpicJSON (`epic list`)
// and emitted directly by `epic show` / `epic set`.
type EpicMetaJSON struct {
	ID          string   `json:"id" jsonschema:"description=epic identifier (NN-slug)"`
	Status      string   `json:"status,omitempty" jsonschema:"description=active | retired | deprecated"`
	Description string   `json:"description,omitempty" jsonschema:"description=one-line epic goal"`
	Priority    string   `json:"priority,omitempty" jsonschema:"description=high | medium | low"`
	Created     string   `json:"created,omitempty" jsonschema:"description=creation date YYYY-MM-DD"`
	Tags        []string `json:"tags,omitempty" jsonschema:"description=topical tags"`
}

// ToEpicMeta is the one place epic meta fields are mapped to JSON, shared by
// `epic list` (embedded in EpicJSON) and `epic show`.
func ToEpicMeta(e domain.Epic) EpicMetaJSON {
	return EpicMetaJSON{
		ID: e.ID, Status: e.Status, Description: e.Description,
		Priority: e.Priority, Created: e.Created, Tags: e.Tags,
	}
}
