package render

import (
	_ "embed"
	"encoding/json"

	"github.com/invopop/jsonschema"

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
// *JSON render func marshals one of these envelopes, and JSONSchema() reflects
// them into a Draft 2020-12 schema (`tskflwctl schema --json-schema`) so an agent
// can validate the tool's output. Field names, json tags, and order mirror what
// the funcs emit exactly — output is byte-stable, and the existing output tests
// guard that the extraction changed nothing.

// TasksEnvelope is `task list --json`.
type TasksEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	Tasks         []taskJSON           `json:"tasks"`
	Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
}

// TaskShowEnvelope is `task show --json`.
type TaskShowEnvelope struct {
	SchemaVersion string   `json:"schema_version"`
	Task          taskJSON `json:"task"`
	Body          string   `json:"body"`
}

// TaskMutationEnvelope is `task set` / `task append` / `task set --body` under
// --json: the reloaded task, dry_run, and (for the body commands) the resulting
// body. Separate from TaskShowEnvelope so the mutation-only dry_run stays off the
// read type.
type TaskMutationEnvelope struct {
	SchemaVersion string   `json:"schema_version"`
	DryRun        bool     `json:"dry_run"`
	Task          taskJSON `json:"task"`
	Body          string   `json:"body,omitempty"`
}

// EpicMutationEnvelope is `epic set --json`: the reloaded epic + dry_run. The
// epic counterpart to TaskMutationEnvelope; it carries the epic meta (not a task)
// and no body field (epic set is field-only — there's no `epic set --body`).
type EpicMutationEnvelope struct {
	SchemaVersion string       `json:"schema_version"`
	DryRun        bool         `json:"dry_run"`
	Epic          epicMetaJSON `json:"epic"`
}

// MovesEnvelope is the transition report (`task start --json`, etc.).
type MovesEnvelope struct {
	SchemaVersion string       `json:"schema_version"`
	DryRun        bool         `json:"dry_run"`
	Moves         []MoveResult `json:"moves"`
}

// SummaryEnvelope is `status --json`.
type SummaryEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	Counts        []statusCountJSON    `json:"counts"`
	InProgress    []taskJSON           `json:"in_progress"`
	Epics         []epicJSON           `json:"epics"`
	OpenAudits    []auditJSON          `json:"open_audits,omitempty"`
	Misfiled      int                  `json:"misfiled"`
	Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
}

// VersionEnvelope is `version --json`.
type VersionEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	Version       string `json:"version"`
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

// EpicsEnvelope is `epic list --json`.
type EpicsEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	Epics         []epicJSON           `json:"epics"`
	Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
}

// EpicShowEnvelope is `epic show --json`.
type EpicShowEnvelope struct {
	SchemaVersion string       `json:"schema_version"`
	Epic          epicMetaJSON `json:"epic"`
	Tasks         []taskJSON   `json:"tasks"`
	Body          string       `json:"body"`
}

// AuditsEnvelope is `audit list --json`.
type AuditsEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	Audits        []auditJSON          `json:"audits"`
	Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
}

// AuditShowEnvelope is `audit show --json`.
type AuditShowEnvelope struct {
	SchemaVersion string    `json:"schema_version"`
	Audit         auditJSON `json:"audit"`
	Body          string    `json:"body"`
}

// FindingsEnvelope is `audit findings --json` (the finding-level query).
type FindingsEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	Findings      []findingJSON        `json:"findings"`
	Unreadable    []domain.FileProblem `json:"unreadable,omitempty"`
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
	Remaining     []lintTaskJSON       `json:"remaining"`
}

// LintEnvelope is `lint --json` and `audit lint --json` (the same per-entity
// slug+issues shape backs both).
type LintEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	Unreadable    []domain.FileProblem `json:"unreadable"`
	Issues        []lintTaskJSON       `json:"issues"`
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

// SchemaEnvelope is `schema --json` (the global contract).
type SchemaEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	SchemaContract
}

// SchemaKindEnvelope is `schema <kind> --json` (per-kind authoring guidance).
type SchemaKindEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	KindSchema
}

// TemplatesEnvelope is `template list --json`.
type TemplatesEnvelope struct {
	SchemaVersion string         `json:"schema_version"`
	Templates     []TemplateInfo `json:"templates"`
}

// TemplateShowEnvelope is `template show --json` (a template's metadata + body).
type TemplateShowEnvelope struct {
	SchemaVersion string       `json:"schema_version"`
	Template      TemplateInfo `json:"template"`
	Body          string       `json:"body"`
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
