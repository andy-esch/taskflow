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
}

// AuditStore is the audit-persistence port.
type AuditStore interface {
	ListAudits() ([]domain.Audit, []domain.FileProblem, error)
	GetAudit(slug string) (audit domain.Audit, body string, err error)
	MoveAudit(slug string, to domain.AuditBucket, dryRun bool) (domain.Audit, error)
	CreateAudit(a domain.Audit, body string, dryRun bool) (domain.Audit, error)
}

// Store is the full persistence port the Service depends on.
type Store interface {
	TaskStore
	EpicStore
	AuditStore

	// FixFrontmatter applies safe text-level frontmatter repairs across all
	// task and epic files (or previews them when dryRun is true).
	FixFrontmatter(dryRun bool) ([]domain.FixResult, error)

	// WatchPaths is the directory set a filesystem watcher should observe. The
	// store owns the on-disk layout; primary adapters (the TUI watcher) consume
	// this instead of rebuilding the path convention themselves.
	WatchPaths() []string
}
