package render

import (
	"github.com/andy-esch/taskflow/internal/domain"
)

// This file holds the unexported JSON DTOs and their mappers — the wire shape of
// each entity inside the --json envelopes (envelopes.go). Each *JSON render func
// maps domain types into these and marshals an envelope that embeds them; keeping
// the DTOs + mappers here keeps render.go to the generic + list/show renderers.

type taskJSON struct {
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

func toJSON(t domain.Task) taskJSON {
	j := taskJSON{
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

type statusCountJSON struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// epicJSON is epic list output: the shared meta (embedded, so `epic list` and
// `epic show` can't drift) plus the task rollup.
type epicJSON struct {
	epicMetaJSON
	Total      int `json:"total"`
	Done       int `json:"done"`
	Percent    int `json:"percent"`
	Deprecated int `json:"deprecated"` // withdrawn tasks, excluded from total/done
}

type auditJSON struct {
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
}

func auditToJSON(a domain.Audit) auditJSON {
	return auditJSON{
		Slug: a.Slug, Bucket: string(a.Bucket), Area: a.Area, Date: a.Date,
		Findings: a.Findings, OpenFindings: a.OpenFindings,
		InProgressFindings: a.ActiveFindings, DoneFindings: a.DoneFindings, DroppedFindings: a.DroppedFindings,
	}
}

type findingJSON struct {
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

type lintTaskJSON struct {
	Slug   string         `json:"slug"`
	Issues []domain.Issue `json:"issues"`
}

type epicMetaJSON struct {
	ID          string   `json:"id" jsonschema:"description=epic identifier (NN-slug)"`
	Status      string   `json:"status,omitempty" jsonschema:"description=active | retired | deprecated"`
	Description string   `json:"description,omitempty" jsonschema:"description=one-line epic goal"`
	Priority    string   `json:"priority,omitempty" jsonschema:"description=high | medium | low"`
	Created     string   `json:"created,omitempty" jsonschema:"description=creation date YYYY-MM-DD"`
	Tags        []string `json:"tags,omitempty" jsonschema:"description=topical tags"`
}

// toEpicMeta is the one place epic meta fields are mapped to JSON, shared by
// `epic list` (embedded in epicJSON) and `epic show`.
func toEpicMeta(e domain.Epic) epicMetaJSON {
	return epicMetaJSON{
		ID: e.ID, Status: e.Status, Description: e.Description,
		Priority: e.Priority, Created: e.Created, Tags: e.Tags,
	}
}
