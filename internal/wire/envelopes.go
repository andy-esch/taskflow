package wire

import (
	_ "embed"
	"encoding/json"

	"github.com/invopop/jsonschema"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// schemaComments is the Go-doc comment map for the envelope + domain types,
// generated at build time by internal/tools/schemacomments (the shipped binary has
// no source for invopop's AddGoComments to read). Regenerate with
// `go run ./internal/tools/schemacomments`; the drift test guards staleness.
//
//go:embed schema_comments.json
var schemaComments []byte

// This file is the named, reflectable form of the --json output contract. Every
// ToXEnvelope constructor returns one of these envelopes (and render's *JSON emit
// funcs marshal it), and JSONSchema() reflects them into a Draft 2020-12 schema
// (`tskflwctl schema --json-schema`) so an agent can validate the tool's output.
// Field names, json tags, and order mirror what the funcs emit exactly — output is
// byte-stable, and the existing output tests guard that the extraction changed
// nothing.

// TasksEnvelope is `task list --json`.
type TasksEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	Tasks         []TaskJSON           `json:"tasks"`
	Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
}

// ToTasksEnvelope builds the `task list --json` envelope value, including any
// per-file load problems so a JSON consumer never silently loses unreadable files.
func ToTasksEnvelope(tasks []domain.Task, problems []domain.FileProblem) TasksEnvelope {
	e := TasksEnvelope{SchemaVersion: SchemaVersion, Tasks: make([]TaskJSON, 0, len(tasks)), Unreadable: problems}
	for _, t := range tasks {
		e.Tasks = append(e.Tasks, ToTaskJSON(t))
	}
	return e
}

// TaskShowEnvelope is `task show --json`.
type TaskShowEnvelope struct {
	SchemaVersion string   `json:"schema_version"`
	Task          TaskJSON `json:"task"`
	Body          string   `json:"body"`
}

// ToTaskShowEnvelope builds the `task show --json` envelope value.
func ToTaskShowEnvelope(t domain.Task, body string) TaskShowEnvelope {
	return TaskShowEnvelope{SchemaVersion: SchemaVersion, Task: ToTaskJSON(t), Body: body}
}

// TaskMutationEnvelope is `task set` / `task append` / `task set --body` under
// --json: the reloaded task, dry_run, and (for the body commands) the resulting
// body. Separate from TaskShowEnvelope so the mutation-only dry_run stays off the
// read type.
type TaskMutationEnvelope struct {
	SchemaVersion string   `json:"schema_version"`
	DryRun        bool     `json:"dry_run"`
	Task          TaskJSON `json:"task"`
	Body          string   `json:"body,omitempty"`
}

// ToTaskMutationEnvelope builds the `task set`/`append`/`set --body` envelope
// value: the reloaded task, dry_run (always present — a preview must be
// distinguishable from a real write), and the resulting body for the body-editing
// commands (empty/omitted for field-only `set`).
func ToTaskMutationEnvelope(t domain.Task, body string, dryRun bool) TaskMutationEnvelope {
	return TaskMutationEnvelope{SchemaVersion: SchemaVersion, DryRun: dryRun, Task: ToTaskJSON(t), Body: body}
}

// EpicMutationEnvelope is `epic set --json`: the reloaded epic + dry_run. The
// epic counterpart to TaskMutationEnvelope; it carries the epic meta (not a task)
// and no body field (epic set is field-only — there's no `epic set --body`).
type EpicMutationEnvelope struct {
	SchemaVersion string       `json:"schema_version"`
	DryRun        bool         `json:"dry_run"`
	Epic          EpicMetaJSON `json:"epic"`
}

// ToEpicMutationEnvelope builds the `epic set --json` envelope value: the reloaded
// epic + dry_run (always present — a preview must be distinguishable from a real
// write). Field-only, so there's no body to echo.
func ToEpicMutationEnvelope(epic domain.Epic, dryRun bool) EpicMutationEnvelope {
	return EpicMutationEnvelope{SchemaVersion: SchemaVersion, DryRun: dryRun, Epic: ToEpicMeta(epic)}
}

// MoveResult is the per-item outcome of a transition. `To` is the destination
// state — a task status or an audit bucket — so the JSON key is the neutral
// "to" rather than "status".
type MoveResult struct {
	Slug      string `json:"slug"`
	To        string `json:"to"`
	RevisitAt string `json:"revisit_at,omitempty" jsonschema:"description=revisit (snooze-until) date recorded by task defer"`
	Error     string `json:"error,omitempty"`
}

// MovesEnvelope is the transition report (`task start --json`, etc.).
type MovesEnvelope struct {
	SchemaVersion string       `json:"schema_version"`
	DryRun        bool         `json:"dry_run"`
	Moves         []MoveResult `json:"moves"`
}

// ToMovesEnvelope builds the per-task transition report; dry_run marks a preview
// (nothing was written). Nil results normalize to an empty (not null) array so the
// output validates against its own schema (type: array).
func ToMovesEnvelope(results []MoveResult, dryRun bool) MovesEnvelope {
	if results == nil {
		results = []MoveResult{}
	}
	return MovesEnvelope{SchemaVersion: SchemaVersion, DryRun: dryRun, Moves: results}
}

// SummaryEnvelope is `status --json`.
type SummaryEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	Counts        []StatusCountJSON    `json:"counts"`
	InProgress    []TaskJSON           `json:"in_progress"`
	Epics         []EpicJSON           `json:"epics"`
	OpenAudits    []AuditJSON          `json:"open_audits,omitempty"`
	Findings      *FindingsRollupJSON  `json:"findings,omitempty"`
	Misfiled      int                  `json:"misfiled"`
	RevisitDue    int                  `json:"revisit_due"`
	BadEpicStatus int                  `json:"bad_epic_status"`
	Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
}

// ToSummaryEnvelope builds the `status --json` dashboard envelope value.
func ToSummaryEnvelope(s core.Summary) SummaryEnvelope {
	counts := make([]StatusCountJSON, 0, len(s.Counts))
	for _, c := range s.Counts {
		counts = append(counts, StatusCountJSON{Status: string(c.Status), Count: c.Count})
	}
	inprog := make([]TaskJSON, 0, len(s.InProgress))
	for _, t := range s.InProgress {
		inprog = append(inprog, ToTaskJSON(t))
	}
	epics := make([]EpicJSON, 0, len(s.Epics))
	for _, e := range s.Epics {
		epics = append(epics, ToEpicJSON(e))
	}
	// open_audits is omitempty: absent unless there's actionable audit work, so a
	// repo with none sees no envelope change (the human dashboard self-hides too).
	audits := make([]AuditJSON, 0, len(s.OpenAudits))
	for _, a := range s.OpenAudits {
		audits = append(audits, ToAuditJSON(a))
	}
	// findings is omitted (nil) unless there's actionable audit work, paralleling
	// open_audits — a repo with none sees no envelope change.
	var findings *FindingsRollupJSON
	if fr := s.Findings; fr.Open+fr.InProgress > 0 {
		f := ToFindingsRollup(fr)
		findings = &f
	}
	return SummaryEnvelope{
		SchemaVersion: SchemaVersion, Counts: counts, InProgress: inprog,
		Epics: epics, OpenAudits: audits, Findings: findings,
		Misfiled: s.Misfiled, RevisitDue: s.RevisitDue, BadEpicStatus: s.BadEpicStatus,
		Unreadable: s.Problems,
	}
}

// VersionEnvelope is `version --json`.
type VersionEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	Version       string `json:"version"`
}

// ToVersionEnvelope builds the `version --json` envelope value.
func ToVersionEnvelope(version string) VersionEnvelope {
	return VersionEnvelope{SchemaVersion: SchemaVersion, Version: version}
}

// CreatedItem is the created document inside CreatedEnvelope.
type CreatedItem struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Status string `json:"status"`
	Path   string `json:"path"`
}

// CreatedEnvelope is `task/epic/audit new --json`.
type CreatedEnvelope struct {
	SchemaVersion string      `json:"schema_version"`
	DryRun        bool        `json:"dry_run"`
	Created       CreatedItem `json:"created"`
}

// ToCreatedEnvelope builds the `new --json` envelope value; dry_run marks a preview
// (nothing was written). status is the new item's status (task status / epic status
// / audit bucket); path is relative to the planning root.
func ToCreatedEnvelope(kind, id, status, path string, dryRun bool) CreatedEnvelope {
	return CreatedEnvelope{SchemaVersion: SchemaVersion, DryRun: dryRun, Created: CreatedItem{Kind: kind, ID: id, Status: status, Path: path}}
}

// EpicsEnvelope is `epic list --json`.
type EpicsEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	Epics         []EpicJSON           `json:"epics"`
	Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
}

// ToEpicsEnvelope builds the `epic list --json` envelope value with rollup,
// including any per-file load problems.
func ToEpicsEnvelope(epics []core.EpicSummary, problems []domain.FileProblem) EpicsEnvelope {
	e := EpicsEnvelope{SchemaVersion: SchemaVersion, Epics: make([]EpicJSON, 0, len(epics)), Unreadable: problems}
	for _, es := range epics {
		e.Epics = append(e.Epics, ToEpicJSON(es))
	}
	return e
}

// EpicShowEnvelope is `epic show --json`.
type EpicShowEnvelope struct {
	SchemaVersion string       `json:"schema_version"`
	Epic          EpicMetaJSON `json:"epic"`
	Tasks         []TaskJSON   `json:"tasks"`
	Body          string       `json:"body"`
}

// ToEpicShowEnvelope builds the `epic show --json` envelope value (epic + tasks + body).
func ToEpicShowEnvelope(epic domain.Epic, tasks []domain.Task, body string) EpicShowEnvelope {
	jt := make([]TaskJSON, 0, len(tasks))
	for _, t := range tasks {
		jt = append(jt, ToTaskJSON(t))
	}
	return EpicShowEnvelope{SchemaVersion: SchemaVersion, Epic: ToEpicMeta(epic), Tasks: jt, Body: body}
}

// AuditsEnvelope is `audit list --json`.
type AuditsEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	Audits        []AuditJSON          `json:"audits"`
	Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
}

// ToAuditsEnvelope builds the `audit list --json` envelope value, including any
// per-file load problems.
func ToAuditsEnvelope(audits []domain.Audit, problems []domain.FileProblem) AuditsEnvelope {
	e := AuditsEnvelope{SchemaVersion: SchemaVersion, Audits: make([]AuditJSON, 0, len(audits)), Unreadable: problems}
	for _, a := range audits {
		e.Audits = append(e.Audits, ToAuditJSON(a))
	}
	return e
}

// AuditShowEnvelope is `audit show --json`.
type AuditShowEnvelope struct {
	SchemaVersion string    `json:"schema_version"`
	Audit         AuditJSON `json:"audit"`
	Body          string    `json:"body"`
}

// ToAuditShowEnvelope builds the `audit show --json` envelope value (audit + body).
func ToAuditShowEnvelope(a domain.Audit, body string) AuditShowEnvelope {
	return AuditShowEnvelope{SchemaVersion: SchemaVersion, Audit: ToAuditJSON(a), Body: body}
}

// FindingsEnvelope is `audit findings --json` (the finding-level query).
type FindingsEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	Findings      []FindingJSON        `json:"findings"`
	Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
}

// ToFindingsEnvelope builds the `audit findings --json` envelope value: each parsed
// finding tagged with its audit slug and bucket, plus any per-file load problems.
func ToFindingsEnvelope(fs []core.AuditFinding, problems []domain.FileProblem) FindingsEnvelope {
	e := FindingsEnvelope{SchemaVersion: SchemaVersion, Findings: make([]FindingJSON, 0, len(fs)), Unreadable: problems}
	for _, f := range fs {
		e.Findings = append(e.Findings, ToFindingJSON(f))
	}
	return e
}

// FixEnvelope is `lint --fix --json`. `remaining` carries the per-entity lint
// findings the fix pass could NOT repair (report-only epic issues, unfixable task
// issues) — the same slug+issues shape `lint --json` emits, so a --json consumer
// learns what's still broken without re-running plain lint.
type FixEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	DryRun        bool                 `json:"dry_run"`
	Fixed         []domain.FixResult   `json:"fixed"`
	Unreadable    []domain.FileProblem `json:"unreadable"`
	Remaining     []LintTaskJSON       `json:"remaining"`
}

// ToFixEnvelope builds the structured fix report value: what was repaired
// (`fixed`), files still unreadable after the pass (`unreadable`), and the
// per-entity lint findings the pass could NOT repair (`remaining`). The array
// fields normalize to empty (not null) so a consumer can len() them and the output
// validates against its own schema (type: array).
func ToFixEnvelope(results []domain.FixResult, problems []domain.FileProblem, remaining []core.LintResult, dryRun bool) FixEnvelope {
	if problems == nil {
		problems = []domain.FileProblem{}
	}
	if results == nil {
		results = []domain.FixResult{}
	}
	rem := make([]LintTaskJSON, 0, len(remaining))
	for _, r := range remaining {
		issues := r.Issues
		if issues == nil {
			issues = []domain.Issue{} // empty, not null — the per-row issues are type: array too
		}
		rem = append(rem, LintTaskJSON{Slug: r.Slug, Issues: issues})
	}
	return FixEnvelope{SchemaVersion: SchemaVersion, DryRun: dryRun, Fixed: results, Unreadable: problems, Remaining: rem}
}

// LintEnvelope is `lint --json` and `audit lint --json` (the same per-entity
// slug+issues shape backs both).
type LintEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	Unreadable    []domain.FileProblem `json:"unreadable"`
	Issues        []LintTaskJSON       `json:"issues"`
}

// ToLintEnvelope builds the structured lint report value: unreadable files + field
// issues. The array fields normalize to empty (not null) so the output validates
// against its own schema (type: array).
func ToLintEnvelope(results []core.LintResult, problems []domain.FileProblem) LintEnvelope {
	if problems == nil {
		problems = []domain.FileProblem{}
	}
	e := LintEnvelope{SchemaVersion: SchemaVersion, Unreadable: problems, Issues: make([]LintTaskJSON, 0, len(results))}
	for _, r := range results {
		issues := r.Issues
		if issues == nil {
			issues = []domain.Issue{} // empty, not null — the per-row issues are type: array too
		}
		e.Issues = append(e.Issues, LintTaskJSON{Slug: r.Slug, Issues: issues})
	}
	return e
}

// InitEnvelope is `init --json`. Mode is "scaffold" (a planning tree written
// under Root) or "pointer" (a planning_repo pointing at an external repo, no
// tree); PlanningRepo is set only in pointer mode. LinkedBack is the
// planning→impl path recorded in the planning repo's tracked_repos by pointer-
// mode auto-link-back (empty when none was written). Tracked is the entries
// `--track` added to this planning repo's tracked_repos (scaffold mode).
type InitEnvelope struct {
	SchemaVersion string   `json:"schema_version"`
	DryRun        bool     `json:"dry_run"`
	Mode          string   `json:"mode"`
	Root          string   `json:"root"`
	PlanningRepo  string   `json:"planning_repo,omitempty"`
	LinkedBack    string   `json:"linked_back,omitempty"`
	Tracked       []string `json:"tracked,omitempty"`
	Created       []string `json:"created"`
}

// NormalizeInitEnvelope stamps the schema_version and normalizes created to an
// empty array (not null) so a consumer can len() it. The caller fills the named
// fields (mode/root/planning_repo/linked_back/tracked/created); the wire package
// owns the version + the array-emptiness invariant.
func NormalizeInitEnvelope(e InitEnvelope) InitEnvelope {
	e.SchemaVersion = SchemaVersion
	if e.Created == nil {
		e.Created = []string{}
	}
	return e
}

// SchemaEnvelope is `schema --json` (the global contract).
type SchemaEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	SchemaContract
}

// ToSchemaEnvelope builds the `schema --json` global-contract envelope value.
func ToSchemaEnvelope(c SchemaContract) SchemaEnvelope {
	return SchemaEnvelope{SchemaVersion: SchemaVersion, SchemaContract: c}
}

// SchemaKindEnvelope is `schema <kind> --json` (per-kind authoring guidance).
type SchemaKindEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	KindSchema
}

// ToSchemaKindEnvelope builds the `schema <kind> --json` per-kind authoring envelope value.
func ToSchemaKindEnvelope(ks KindSchema) SchemaKindEnvelope {
	return SchemaKindEnvelope{SchemaVersion: SchemaVersion, KindSchema: ks}
}

// TemplatesEnvelope is `template list --json`.
type TemplatesEnvelope struct {
	SchemaVersion string         `json:"schema_version"`
	Templates     []TemplateInfo `json:"templates"`
}

// ToTemplatesEnvelope builds the `template list --json` envelope value.
func ToTemplatesEnvelope(ts []TemplateInfo) TemplatesEnvelope {
	return TemplatesEnvelope{SchemaVersion: SchemaVersion, Templates: ts}
}

// TemplateShowEnvelope is `template show --json` (a template's metadata + body).
type TemplateShowEnvelope struct {
	SchemaVersion string       `json:"schema_version"`
	Template      TemplateInfo `json:"template"`
	Body          string       `json:"body"`
}

// ToTemplateShowEnvelope builds the `template show --json` envelope value.
func ToTemplateShowEnvelope(info TemplateInfo, body string) TemplateShowEnvelope {
	return TemplateShowEnvelope{SchemaVersion: SchemaVersion, Template: info, Body: body}
}

// ErrorItem is the error body inside ErrorEnvelope.
type ErrorItem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorEnvelope is the failure payload emitted under --json (see cli.WriteError).
type ErrorEnvelope struct {
	SchemaVersion string    `json:"schema_version"`
	Error         ErrorItem `json:"error"`
}

// DoctorEnvelope is `doctor --json`: the linkback audit result. Problems is
// empty (not null) when the planning_repo <-> tracked_repos links are consistent.
type DoctorEnvelope struct {
	SchemaVersion string          `json:"schema_version"`
	Root          string          `json:"root"`
	Problems      []DoctorProblem `json:"problems"`
}

// ToDoctorEnvelope builds the `doctor --json` envelope value; problems normalizes
// to empty (not null) when the links are consistent, so a consumer can len() it
// without a nil check.
func ToDoctorEnvelope(root string, problems []DoctorProblem) DoctorEnvelope {
	if problems == nil {
		problems = []DoctorProblem{}
	}
	return DoctorEnvelope{SchemaVersion: SchemaVersion, Root: root, Problems: problems}
}

// DoctorProblem is one linkback inconsistency: the offending repo + a message.
type DoctorProblem struct {
	Repo    string `json:"repo"`
	Message string `json:"message"`
}

// jsonEnvelopes registers every envelope so a single Reflect pulls them all (and
// their shared types) into one schema document's $defs.
type jsonEnvelopes struct {
	Tasks        TasksEnvelope        `json:"tasks"`
	TaskShow     TaskShowEnvelope     `json:"task_show"`
	TaskMutation TaskMutationEnvelope `json:"task_mutation"`
	EpicMutation EpicMutationEnvelope `json:"epic_mutation"`
	Moves        MovesEnvelope        `json:"moves"`
	Summary      SummaryEnvelope      `json:"summary"`
	Version      VersionEnvelope      `json:"version"`
	Created      CreatedEnvelope      `json:"created"`
	Epics        EpicsEnvelope        `json:"epics"`
	EpicShow     EpicShowEnvelope     `json:"epic_show"`
	Audits       AuditsEnvelope       `json:"audits"`
	AuditShow    AuditShowEnvelope    `json:"audit_show"`
	Findings     FindingsEnvelope     `json:"findings"`
	Fix          FixEnvelope          `json:"fix"`
	Lint         LintEnvelope         `json:"lint"`
	Init         InitEnvelope         `json:"init"`
	Doctor       DoctorEnvelope       `json:"doctor"`
	Schema       SchemaEnvelope       `json:"schema"`
	SchemaKind   SchemaKindEnvelope   `json:"schema_kind"`
	Templates    TemplatesEnvelope    `json:"templates"`
	TemplateShow TemplateShowEnvelope `json:"template_show"`
	Error        ErrorEnvelope        `json:"error"`
}

// Envelopes returns the reflect type of the registry so a coverage test can
// enumerate every registered envelope without re-declaring the list.
func Envelopes() jsonEnvelopes { return jsonEnvelopes{} }

// JSONSchema returns the Draft 2020-12 JSON Schema for every --json envelope, as
// one document. Each property of the root object names an envelope and $refs its
// definition; validate a command's --json output against the matching entry in
// $defs (e.g. `task list --json` against $defs/TasksEnvelope).
func JSONSchema() ([]byte, error) {
	r := &jsonschema.Reflector{}
	// Field/type descriptions come from the build-time-generated comment map, so a
	// shipped binary emits them without needing the source tree at runtime.
	var comments map[string]string
	if err := json.Unmarshal(schemaComments, &comments); err != nil {
		return nil, err
	}
	r.CommentMap = comments
	s := r.Reflect(&jsonEnvelopes{})
	s.Title = "tskflwctl --json output (schema_version " + SchemaVersion + ")"
	s.Description = "Each property of the root names a --json envelope and references its definition in $defs; " +
		"validate a command's --json output against the matching definition."
	return json.MarshalIndent(s, "", "  ")
}
