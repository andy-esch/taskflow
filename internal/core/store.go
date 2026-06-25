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
	GetTask(slug string) (task domain.Task, body string, err error)
	// Mutators take dryRun: true runs EVERY validation (resolve, parse-before-
	// commit, collision/CAS checks) and returns the would-be result, but stops
	// short of touching disk — so a dry-run that would fail fails identically.
	Move(slug string, to domain.Status, now time.Time, dryRun bool) (domain.Task, error)
	SetFields(slug string, updates map[string]any, dryRun bool) (domain.Task, error)
	CreateTask(t domain.Task, body string, dryRun bool) (domain.Task, error)
	// EditTask hands the current file content to edit (which runs the caller's
	// editor) and accepts the result only if it still parses as a task —
	// parse-before-accept, looping on the editor for a broken edit. Reports
	// whether the file changed.
	EditTask(slug string, edit func(current string, prevErr error) (string, error)) (domain.Task, bool, error)
	// EditBody replaces (appendMode=false) or appends to (true) a task's markdown
	// body in one atomic, validated write, preserving the frontmatter and stamping
	// updated_at. The agent face of body editing, beside EditTask's editor. Returns
	// the reloaded task and the resulting body (so a --json caller can echo it).
	EditBody(slug, text string, appendMode bool, now time.Time, dryRun bool) (domain.Task, string, error)
}

// EpicStore is the epic-persistence port.
type EpicStore interface {
	ListEpics() ([]domain.Epic, []domain.FileProblem, error)
	GetEpic(id string) (epic domain.Epic, body string, err error)
	CreateEpic(slug string, e domain.Epic, body string, dryRun bool) (domain.Epic, error)
	// MoveEpic surgically rewrites an epic's `status` frontmatter field (epic
	// status is a field, not a directory, so the file stays put). dryRun runs every
	// validation and returns the would-be epic without touching disk.
	MoveEpic(id, status string, dryRun bool) (domain.Epic, error)
}

// AuditStore is the audit-persistence port.
type AuditStore interface {
	ListAudits() ([]domain.Audit, []domain.FileProblem, error)
	GetAudit(slug string) (audit domain.Audit, body string, err error)
	// GetAuditByPath reads one audit directly by its file path, deriving the
	// bucket from the parent directory (the bucket==directory invariant) rather
	// than re-resolving the slug across every bucket dir. The finding/lint sweeps
	// use this to read each audit ListAudits already located exactly once, instead
	// of an O(N^2) re-resolve+re-read per audit.
	GetAuditByPath(path string) (audit domain.Audit, body string, err error)
	MoveAudit(slug string, to domain.AuditBucket, dryRun bool) (domain.Audit, error)
	CreateAudit(a domain.Audit, body string, dryRun bool) (domain.Audit, error)
}

// Store is the use-case persistence port the Service depends on. It is
// deliberately narrow: only the task/epic/audit use cases live here. The two
// fs/text operations that aren't use cases (frontmatter repair, watch-path
// layout) are split into Fixer/Layout below so a second Store implementation —
// and the test fakes — don't pay for methods the core never calls.
type Store interface {
	TaskStore
	EpicStore
	AuditStore
}

// Fixer is the frontmatter-repair port. It is an fs/text operation, not a core
// use case, so it sits beside Store rather than inside it; the CLI's `lint --fix`
// wires it directly to the FS instead of routing through the Service.
type Fixer interface {
	// FixFrontmatter applies safe text-level frontmatter repairs across all
	// task and epic files (or previews them when dryRun is true).
	FixFrontmatter(dryRun bool) ([]domain.FixResult, error)
}

// Layout is the on-disk-layout port: the directory set a filesystem watcher must
// observe. The store owns the layout convention, so the TUI watcher consumes this
// instead of rebuilding the tasks/<status> + audits/<bucket> shape itself.
type Layout interface {
	WatchPaths() []string
}
