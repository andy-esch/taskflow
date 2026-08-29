// Package core holds the application use cases (the Service) and the ports the
// core needs. Interfaces are defined here, at the consumer, per the org's
// "keep interfaces close to where they're used" guidance.
package core

import (
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TaskStore is the task-persistence port. The list methods return per-file
// problems separately from a fatal error, so callers can show the good data
// and report unreadable files instead of dying on the first one.
type TaskStore interface {
	ListTasks() ([]domain.Task, []domain.FileProblem, error)
	// ListTasksWithBodies is ListTasks' scan with each task's markdown body kept
	// alongside, so a body-aware pass (lint's acceptance-criteria checks) reads every
	// file once instead of re-resolving each slug through GetTask. Same resilient-read
	// contract: an unreadable file is a FileProblem, not fatal.
	ListTasksWithBodies() ([]TaskWithBody, []domain.FileProblem, error)
	GetTask(slug string) (task domain.Task, body string, err error)
	// ResolveTaskPath returns a task's file path from its slug/id WITHOUT parsing —
	// so `task path` works even on a file whose frontmatter won't parse (the case
	// where you most need the path, to open and repair it).
	ResolveTaskPath(slug string) (string, error)
	// Ordinary task mutations take dryRun: true runs every validation and returns
	// the would-be result but stops short of disk. Lifecycle changes are deliberately
	// absent: TaskLifecycleMutationStore is the only status-write capability.
	SetFields(slug string, updates map[string]any, dryRun bool) (domain.Task, error)
	CreateTask(t domain.Task, body string, dryRun bool) (domain.Task, error)
	// EditTask hands the current file content to edit (which runs the caller's
	// editor) and accepts the result only if it still parses as a task —
	// parse-before-accept, looping on the editor for a broken edit. A changed save
	// is stamped with updated_at (now); reports whether the file changed.
	EditTask(slug string, now time.Time, edit func(current string, prevErr error) (string, error)) (domain.Task, bool, error)
	// EditBody replaces (appendMode=false) or appends to (true) a task's markdown
	// body in one atomic, validated write, preserving the frontmatter and stamping
	// updated_at. The agent face of body editing, beside EditTask's editor. Returns
	// the reloaded task and the resulting body (so a --json caller can echo it).
	EditBody(slug, text string, appendMode bool, now time.Time, dryRun bool) (domain.Task, string, error)
	// RenameTask re-titles a task: a new slug from newTitle, the file renamed (id kept),
	// the body H1 rewritten, and every inbound relative-path markdown link across the tree
	// repointed to the new filename. Returns the reloaded task and the count of links
	// repointed. Multi-file + write-locked but not version-CAS'd (a rare deliberate op).
	RenameTask(slug, newTitle string, dryRun bool) (domain.Task, int, error)
}

// TaskDependencyWrite is one semantic task-file change returned by a pure graph
// mutation planner. The store owns YAML surgery and atomic replacement; planners
// name only the task and its complete canonical dependency set. ClearLegacy is
// reserved for the guarded migration from blocked_by/dependencies/blocks.
type TaskDependencyWrite struct {
	TaskID      string
	DependsOn   []string
	ClearLegacy bool
}

// TaskGraphMutationPlan is the complete deterministic write set produced from one
// immutable repository snapshot. Writes are applied in the planner-provided order:
// ordering is semantic recovery data because every durable prefix must remain sound.
// Each file is replaced atomically; AppliedTaskIDs in the result is the durable prefix
// when a later write fails, which makes a multi-file operation diagnosable and resumable.
type TaskGraphMutationPlan struct {
	TaskWrites []TaskDependencyWrite
}

// TaskGraphMutationResult reports what the store planned and which task-file
// replacements actually landed. Dry runs return the normalized plan with no applied
// IDs. A non-nil error may accompany a non-empty AppliedTaskIDs prefix.
type TaskGraphMutationResult struct {
	Plan           TaskGraphMutationPlan
	AppliedTaskIDs []string
	DryRun         bool
}

// TaskGraphPlanner is deliberately control-inverted: the store calls a pure core
// planner while it owns the repository guard. The callback receives only the
// immutable taskflow graph and returns taskflow-owned semantic values; it must not
// call a Store method or begin another mutation.
type TaskGraphPlanner func(*TaskGraph) (TaskGraphMutationPlan, error)

// TaskGraphMutationStore owns the repository-wide graph read/validate/write
// critical section. It is a sibling capability rather than part of Store so read-
// only/test adapters do not acquire a mutation method they cannot implement.
type TaskGraphMutationStore interface {
	MutateTaskGraph(now time.Time, dryRun bool, planner TaskGraphPlanner) (TaskGraphMutationResult, error)
}

// TaskLifecyclePlanner resolves user intent and returns one semantic lifecycle
// plan from the immutable authoritative graph. It must not call a Store method or
// begin another guarded mutation.
type TaskLifecyclePlanner func(*TaskGraph) (TaskLifecyclePlan, error)

// TaskLifecycleMutationStore owns the guarded lifecycle read/authorize/write
// boundary. It is separate from TaskStore so read-only and lightweight test
// adapters do not claim atomic repository semantics they cannot provide.
type TaskLifecycleMutationStore interface {
	MutateTaskLifecycle(now time.Time, dryRun bool, planner TaskLifecyclePlanner) (TaskLifecycleMutationResult, error)
}

// EpicStore is the epic-persistence port.
type EpicStore interface {
	ListEpics() ([]domain.Epic, []domain.FileProblem, error)
	GetEpic(id string) (epic domain.Epic, body string, err error)
	// ResolveEpicPath returns an epic's file path from its id, parse-free (see
	// ResolveTaskPath).
	ResolveEpicPath(id string) (string, error)
	CreateEpic(slug string, e domain.Epic, body string, dryRun bool) (domain.Epic, error)
	// MoveEpic surgically rewrites an epic's `status` frontmatter field (epic
	// status is a field, not a directory, so the file stays put), stamping updated_at
	// on a real status change. dryRun runs every validation and returns the would-be
	// epic without touching disk.
	MoveEpic(id, status string, now time.Time, dryRun bool) (domain.Epic, error)
	// SetEpicFields surgically updates non-status frontmatter fields on an epic in
	// one atomic, validated write (status moves via MoveEpic). updated_at is injected
	// by the service. dryRun runs every validation and returns the would-be epic
	// without touching disk.
	SetEpicFields(id string, updates map[string]any, dryRun bool) (domain.Epic, error)
	// EditEpic hands the current file content to edit (which runs the caller's
	// editor) and accepts the result only if it still parses as an epic —
	// parse-before-accept, looping on the editor for a broken edit. A changed save is
	// stamped with updated_at (now); reports whether the file changed. The epic
	// counterpart to EditTask.
	EditEpic(id string, now time.Time, edit func(current string, prevErr error) (string, error)) (domain.Epic, bool, error)
}

// AuditWithFindings pairs an audit with the findings parsed from the SAME body
// read that produced its tally — so a sweep that needs both the audit-level
// counts and the per-finding rows reads each file once. ListAudits already parses
// the findings to compute the tally bands and then discards them; this surfaces
// them instead, in document order.
type AuditWithFindings struct {
	Audit    domain.Audit
	Findings []domain.Finding
}

// TaskWithBody is a task plus its markdown body, kept together by
// ListTasksWithBodies so lint's body-aware checks read each file once.
type TaskWithBody struct {
	Task domain.Task
	Body string
}

// AuditStore is the audit-persistence port.
type AuditStore interface {
	ListAudits() ([]domain.Audit, []domain.FileProblem, error)
	// ListAuditsWithFindings is ListAudits' scan with the parsed findings kept
	// alongside each audit, so Summary computes the audit tallies AND the findings
	// rollup from a single read of every body instead of re-reading each one through
	// GetAuditByPath. Same resilient-read contract: an unreadable file is a
	// FileProblem, not fatal.
	ListAuditsWithFindings() ([]AuditWithFindings, []domain.FileProblem, error)
	GetAudit(slug string) (audit domain.Audit, body string, err error)
	// ResolveAuditPath returns an audit's file path from its slug/id, parse-free
	// (see ResolveTaskPath).
	ResolveAuditPath(slug string) (string, error)
	// GetAuditByPath reads one audit directly by its file path (bucket read from
	// frontmatter, ADR-0003 §4) rather than re-resolving the slug. The finding/lint
	// sweeps use this to read each audit ListAudits already located exactly once,
	// instead of an O(N^2) re-resolve+re-read per audit.
	GetAuditByPath(path string) (audit domain.Audit, body string, err error)
	MoveAudit(slug string, to domain.AuditBucket, dryRun bool) (domain.Audit, error)
	CreateAudit(a domain.Audit, body string, dryRun bool) (domain.Audit, error)
	// EditAudit hands the current file content to edit (the caller's editor) and
	// accepts the result only if it still parses as an audit — parse-before-accept,
	// looping on a broken edit. A changed save is stamped with updated_at (now);
	// reports whether the file changed. The audit counterpart to EditTask;
	// finding-level lint is the caller's to surface.
	EditAudit(slug string, now time.Time, edit func(current string, prevErr error) (string, error)) (domain.Audit, bool, error)
	// AppendAuditBody appends markdown to an audit's body in one atomic, validated
	// write, stamping updated_at (the audit's `date` stays immutable — it's the slug).
	// The agent face of audit body editing, beside EditAudit's editor. Returns the
	// reloaded audit and the resulting body.
	AppendAuditBody(slug, text string, now time.Time, dryRun bool) (domain.Audit, string, error)
}

// ResearchStore is the research-persistence port. The narrowest entity port: research
// has no lifecycle (so no Move) and no cross-references (so nothing to resolve), which
// leaves scan, read, path, and create.
type ResearchStore interface {
	ListResearch() ([]domain.Research, []domain.FileProblem, error)
	GetResearch(slug string) (research domain.Research, body string, err error)
	// ResolveResearchPath returns a doc's file path from its slug/id, parse-free
	// (see ResolveTaskPath).
	ResolveResearchPath(slug string) (string, error)
	CreateResearch(r domain.Research, body string, dryRun bool) (domain.Research, error)
	// SetResearchFields surgically updates frontmatter fields in one atomic, validated
	// write. updated_at is injected by the service; `created` is rejected upstream (the
	// id encodes it). dryRun runs every validation without touching disk.
	SetResearchFields(slug string, updates map[string]any, dryRun bool) (domain.Research, error)
	// EditResearch hands the current file content to edit (which runs the caller's
	// editor) and accepts the result only if it still parses as a research doc —
	// parse-before-accept, looping on a broken edit. A changed save is stamped with
	// updated_at (now); reports whether the file changed.
	EditResearch(slug string, now time.Time, edit func(current string, prevErr error) (string, error)) (domain.Research, bool, error)
	// AppendResearchBody appends markdown to a doc's body in one atomic, validated
	// write, stamping updated_at (`created` stays immutable — the id is minted from it).
	// The agent face of body editing, beside EditResearch's editor.
	AppendResearchBody(slug, text string, now time.Time, dryRun bool) (domain.Research, string, error)
}

// Store is the use-case persistence port the Service depends on. It is
// deliberately narrow: only the task/epic/audit/research use cases live here. The three
// fs/text operations that aren't use cases (frontmatter repair, link checks,
// and watch-path layout) are split into Fixer/Linter/Layout below so a second
// Store implementation — and the test fakes — don't pay for methods the core
// never calls.
type Store interface {
	TaskStore
	EpicStore
	AuditStore
	ResearchStore
}

// SummaryStore is the narrow read port required by one dashboard scan. Store satisfies
// it, while cross-space status can ask its adapter for only these methods: the overview is
// read-only by construction, not merely by a primary-adapter annotation.
type SummaryStore interface {
	ListTasks() ([]domain.Task, []domain.FileProblem, error)
	ListEpics() ([]domain.Epic, []domain.FileProblem, error)
	ListAuditsWithFindings() ([]AuditWithFindings, []domain.FileProblem, error)
}

// Fixer is the frontmatter-repair port. It is an fs/text operation, not a core
// use case, so it sits beside Store rather than inside it; the CLI's `lint --fix`
// wires it directly to the FS instead of routing through the Service.
type Fixer interface {
	// FixFrontmatter applies safe text-level frontmatter repairs across all
	// task and epic files (or previews them when dryRun is true).
	FixFrontmatter(dryRun bool) ([]domain.FixResult, error)
}

// Linter is the cross-link integrity port. Like Fixer it's an fs/text operation, not a
// core use case, so `lint --links` wires it directly to the FS rather than through the
// Service.
type Linter interface {
	// DanglingLinks reports every body markdown link whose target .md file is missing.
	DanglingLinks() ([]domain.FileProblem, error)
}

// Layout is the on-disk-layout port: the directory set a filesystem watcher must
// observe. The store owns the layout convention, so the TUI watcher consumes this
// instead of rebuilding the flat entity-directory shape itself.
type Layout interface {
	WatchPaths() []string
}
