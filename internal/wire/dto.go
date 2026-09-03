package wire

import (
	"sort"

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
	ID     string `json:"id,omitempty" jsonschema:"description=stable identifier — the immutable key that survives slug and status changes; absent on tasks created before id assignment"`
	Slug   string `json:"slug" jsonschema:"description=task slug (filename without .md) — the human handle"`
	Status string `json:"status" jsonschema:"description=lifecycle status — authoritative, read from frontmatter (ADR-0003 §4)"`
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
	DependsOn   []string `json:"depends_on,omitempty" jsonschema:"description=sorted stable task IDs that must be soundly completed before this task is ordinarily eligible to start"`
}

// ToTaskJSON maps a domain task to its wire DTO.
func ToTaskJSON(t domain.Task) TaskJSON {
	j := TaskJSON{
		ID: t.ID, Slug: t.Slug, Status: string(t.Status), Epic: t.Epic,
		Description: t.Description, Effort: t.Effort, Tier: t.Tier,
		Priority: t.Priority, Autonomy: t.Autonomy,
		Created: t.Created, Updated: t.Updated, RevisitAt: t.RevisitAt, Tags: t.Tags,
		DependsOn: sortedStrings(t.DependsOn),
	}
	return j
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

// ACJSON is a task's acceptance-criteria checkbox tally (the `ac` field of
// `task info`): how many criteria are checked out of the total.
type ACJSON struct {
	Checked int `json:"checked" jsonschema:"description=acceptance criteria currently checked (- [x])"`
	Total   int `json:"total" jsonschema:"description=total acceptance criteria in the task's acceptance-criteria section"`
	// Explained is how many UNMET criteria state why — any of the non-binary criterion
	// states (see domain.CriterionSuffixStates). Zero for
	// every task written before the criterion vocabulary existed. Schema 1.45.
	Explained int `json:"explained,omitempty"`
}

// TaskInfoJSON is the token-cheap metadata read for a task (`task info`): where
// the file lives plus the fields an agent triages on and the acceptance-criteria
// tally, WITHOUT the body — the machine counterpart to `task path` that avoids the
// full `task show` payload.
type TaskInfoJSON struct {
	ID     string `json:"id,omitempty" jsonschema:"description=stable identifier — absent on tasks created before id assignment"`
	Slug   string `json:"slug" jsonschema:"description=task slug (filename without .md)"`
	Status string `json:"status" jsonschema:"description=lifecycle status — authoritative from frontmatter (ADR-0003 §4)"`
	Epic   string `json:"epic,omitempty" jsonschema:"description=id of the epic this task belongs to"`
	Path   string `json:"path" jsonschema:"description=absolute path to the task's markdown file"`
	AC     ACJSON `json:"ac" jsonschema:"description=acceptance-criteria checkbox tally"`
}

// ToTaskInfoJSON maps a task + its acceptance tally + resolved path to the info DTO.
func ToTaskInfoJSON(t domain.Task, ac domain.ACCount, path string) TaskInfoJSON {
	return TaskInfoJSON{
		ID: t.ID, Slug: t.Slug, Status: string(t.Status), Epic: t.Epic,
		Path: path, AC: ACJSON{Checked: ac.Checked, Total: ac.Total, Explained: ac.Explained},
	}
}

// CriterionJSON is one acceptance-criteria checkbox for `task ac --list --json` —
// the list an agent then flips by index with `task ac --check/--uncheck`.
type CriterionJSON struct {
	Index   int    `json:"index" jsonschema:"description=1-based position of the criterion in the acceptance section"`
	Checked bool   `json:"checked" jsonschema:"description=whether the checkbox is checked (- [x])"`
	Text    string `json:"text" jsonschema:"description=the criterion text — the first line after the checkbox"`
	// State and Reason carry the disposition beyond the checkbox. Omitted when the
	// criterion is a plain met/not-met, so a body written before the vocabulary existed
	// serialises exactly as it did. Without these an agent reading `task ac --list --json`
	// could see an unchecked box but not whether it was still to do, deferred, abandoned,
	// or no longer applicable — the ambiguity the vocabulary exists to remove, reintroduced
	// at the machine boundary. Schema 1.46.
	State  string `json:"state,omitempty" jsonschema:"description=disposition beyond the checkbox — one of criterion_states in the schema contract; absent for a plain met/not-met criterion"`
	Reason string `json:"reason,omitempty" jsonschema:"description=why the criterion is deferred/wontfix/n-a — required for those states"`
	// TrackedBy is the destination task id of a `tracked` criterion. It is a field
	// rather than part of Reason because the point of `tracked` is that the work went
	// somewhere followable: held as prose, nothing resolves it and nothing notices when
	// the destination is renamed or completed without absorbing the criterion. Absent
	// for every other state, and for a tracked criterion written before the destination
	// was recorded.
	TrackedBy string `json:"tracked_by,omitempty" jsonschema:"description=stable id of the task a tracked criterion was handed to"`
}

// ToCriteriaJSON maps the domain criteria to their wire DTOs (never nil — an empty
// acceptance list marshals to []).
func ToCriteriaJSON(cs []domain.Criterion) []CriterionJSON {
	out := make([]CriterionJSON, len(cs))
	for i, c := range cs {
		j := CriterionJSON{Index: c.Index, Checked: c.Checked, Text: c.Text}
		if c.State.NeedsReason() {
			j.State, j.Reason, j.TrackedBy = string(c.State), c.Reason, c.TrackedBy
		}
		out[i] = j
	}
	return out
}

// FindingsTallyJSON is an audit's finding disposition tally (the `findings` field
// of `audit info`) — the audit analogue of a task's acceptance-criteria tally.
type FindingsTallyJSON struct {
	Total      int `json:"total" jsonschema:"description=total findings parsed from the audit body"`
	Open       int `json:"open" jsonschema:"description=findings whose status is open"`
	InProgress int `json:"in_progress" jsonschema:"description=findings whose status is in-progress"`
	Done       int `json:"done" jsonschema:"description=findings whose status is fixed or tracked"`
	Dropped    int `json:"dropped" jsonschema:"description=findings whose status is deferred, superseded, or wontfix"`
}

// AuditInfoJSON is the token-cheap metadata read for an audit (`audit info`): where
// the file lives, its bucket, and the finding tally — no body.
type AuditInfoJSON struct {
	ID       string            `json:"id,omitempty" jsonschema:"description=stable identifier — absent on audits created before id assignment"`
	Slug     string            `json:"slug" jsonschema:"description=audit slug (filename without .md)"`
	Bucket   string            `json:"bucket" jsonschema:"description=open | closed | deferred — authoritative from frontmatter (ADR-0003 §4)"`
	Path     string            `json:"path" jsonschema:"description=absolute path to the audit's markdown file"`
	Findings FindingsTallyJSON `json:"findings" jsonschema:"description=finding disposition tally"`
}

// ToAuditInfoJSON maps an audit (whose disposition tally is populated on load) + its
// resolved path to the info DTO.
func ToAuditInfoJSON(a domain.Audit, path string) AuditInfoJSON {
	return AuditInfoJSON{
		ID: a.ID, Slug: a.Slug, Bucket: string(a.Bucket), Path: path,
		Findings: FindingsTallyJSON{
			Total: a.Findings, Open: a.OpenFindings, InProgress: a.ActiveFindings,
			Done: a.DoneFindings, Dropped: a.DroppedFindings,
		},
	}
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
	ID           string `json:"id,omitempty" jsonschema:"description=stable identifier — the immutable key; absent on audits created before id assignment"`
	Slug         string `json:"slug" jsonschema:"description=audit slug (filename without .md) — the human handle"`
	Bucket       string `json:"bucket" jsonschema:"description=open | closed | deferred — authoritative, read from frontmatter (ADR-0003 §4)"`
	Area         string `json:"area,omitempty" jsonschema:"description=subsystem/topic audited"`
	Date         string `json:"date,omitempty" jsonschema:"description=audit date YYYY-MM-DD (immutable — part of the slug)"`
	Updated      string `json:"updated_at,omitempty" jsonschema:"description=audit's own last-edited date YYYY-MM-DD (edit/append); a bucket move does not change it"`
	Findings     int    `json:"findings" jsonschema:"description=total findings parsed from the body"`
	OpenFindings int    `json:"open_findings" jsonschema:"description=findings whose status is open"`
	// The progress bar's disposition bands. open + in_progress + done + dropped ≤
	// findings (an unrecognized/missing status counts toward none).
	InProgressFindings int `json:"in_progress_findings" jsonschema:"description=findings whose status is in-progress"`
	DoneFindings       int `json:"done_findings" jsonschema:"description=findings whose status is fixed or tracked (the bar's done band)"`
	DroppedFindings    int `json:"dropped_findings" jsonschema:"description=findings whose status is deferred, superseded, or wontfix"`
	// ReadyToClose is true for an OPEN audit whose findings are all resolved/dropped
	// (none open or in-progress) — a "ready to close" call-to-action.
	ReadyToClose bool `json:"ready_to_close,omitempty" jsonschema:"description=true when an open audit has no open/in-progress findings left (ready to close)"`
}

// ResearchJSON is the wire DTO for a research doc. Thin by design and the omissions
// are the contract: there is no `status`/`bucket` (research has no lifecycle) and no
// `epic`/`tasks` (provenance is a body concern, not frontmatter) — see domain.Research.
// An agent discovering this shape should read "snapshot, ordered by date", not
// "work item".
type ResearchJSON struct {
	ID          string   `json:"id,omitempty" jsonschema:"description=stable identifier — the immutable key, minted from created so lexical id order is authorship order"`
	Slug        string   `json:"slug" jsonschema:"description=research slug (filename without the leading id) — the human handle"`
	Created     string   `json:"created,omitempty" jsonschema:"description=date the research was done YYYY-MM-DD; the id is minted from it"`
	Description string   `json:"description,omitempty" jsonschema:"description=one-line summary of what was explored"`
	Tags        []string `json:"tags,omitempty" jsonschema:"description=topical tags"`
	Updated     string   `json:"updated_at,omitempty" jsonschema:"description=doc's own last-edited date YYYY-MM-DD; created stays immutable"`
}

// ToResearchJSON maps a domain research doc to its wire DTO.
func ToResearchJSON(r domain.Research) ResearchJSON {
	return ResearchJSON{
		ID: r.ID, Slug: r.Slug, Created: r.Created,
		Description: r.Description, Tags: r.Tags, Updated: r.Updated,
	}
}

// ToAuditJSON maps a domain audit to its wire DTO.
func ToAuditJSON(a domain.Audit) AuditJSON {
	return AuditJSON{
		ID: a.ID, Slug: a.Slug, Bucket: string(a.Bucket), Area: a.Area, Date: a.Date, Updated: a.Updated,
		Findings: a.Findings, OpenFindings: a.OpenFindings,
		InProgressFindings: a.ActiveFindings, DoneFindings: a.DoneFindings, DroppedFindings: a.DroppedFindings,
		ReadyToClose: a.ReadyToClose(),
	}
}

// FindingJSON is the wire shape of one audit finding.
type FindingJSON struct {
	Audit            string `json:"audit" jsonschema:"description=slug of the audit this finding belongs to"`
	Bucket           string `json:"bucket" jsonschema:"description=the audit's bucket — open | closed | deferred"`
	Code             string `json:"code" jsonschema:"description=finding code within the audit (H1/M2/S3…)"`
	Title            string `json:"title" jsonschema:"description=finding title"`
	Status           string `json:"status" jsonschema:"description=open | in-progress | fixed | tracked | deferred | superseded | wontfix"`
	File             string `json:"file,omitempty" jsonschema:"description=file:line the finding refers to"`
	Component        string `json:"component,omitempty" jsonschema:"description=component/subsystem"`
	Effort           string `json:"effort,omitempty" jsonschema:"description=XS | S | M | L"`
	Urgency          string `json:"urgency,omitempty" jsonschema:"description=acute | soon | eventually"`
	StatusDecoration string `json:"status_decoration,omitempty" jsonschema:"description=everything after the status word — the date, PR link, by <task-id> destination, or reason"`
	Note             string `json:"note,omitempty" jsonschema:"description=the finding's resolution paragraph — how it was resolved; absent when it carries none"`
}

// ToFindingJSON maps a core audit finding to its wire DTO.
func ToFindingJSON(f core.AuditFinding) FindingJSON {
	return FindingJSON{
		Audit: f.Audit, Bucket: f.Bucket, Code: f.Code, Title: f.Title, Status: f.Status,
		File: f.File, Component: f.Component, Effort: f.Effort, Urgency: f.Urgency,
		Note: f.Note, StatusDecoration: f.StatusDecoration,
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
	Updated     string   `json:"updated_at,omitempty" jsonschema:"description=epic's own last-edited date YYYY-MM-DD (set/edit/status-move); distinct from derived task activity"`
	Tags        []string `json:"tags,omitempty" jsonschema:"description=topical tags"`
}

// ToEpicMeta is the one place epic meta fields are mapped to JSON, shared by
// `epic list` (embedded in EpicJSON) and `epic show`.
func ToEpicMeta(e domain.Epic) EpicMetaJSON {
	return EpicMetaJSON{
		ID: e.ID, Status: e.Status, Description: e.Description,
		Priority: e.Priority, Created: e.Created, Updated: e.Updated, Tags: e.Tags,
	}
}
