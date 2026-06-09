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
	Move(slug string, to domain.Status, now time.Time) (domain.Task, error)
	SetFields(slug string, updates map[string]any) (domain.Task, error)
	CreateTask(t domain.Task, body string) (domain.Task, error)
}

// EpicStore is the epic-persistence port.
type EpicStore interface {
	ListEpics() ([]domain.Epic, []domain.FileProblem, error)
	GetEpic(id string) (epic domain.Epic, body string, err error)
	CreateEpic(slug string, e domain.Epic, body string) (domain.Epic, error)
}

// AuditStore is the audit-persistence port.
type AuditStore interface {
	ListAudits() ([]domain.Audit, []domain.FileProblem, error)
	GetAudit(slug string) (audit domain.Audit, body string, err error)
	MoveAudit(slug string, to domain.AuditBucket) (domain.Audit, error)
}

// Store is the full persistence port the Service depends on.
type Store interface {
	TaskStore
	EpicStore
	AuditStore

	// FixFrontmatter applies safe text-level frontmatter repairs across all
	// task and epic files (or previews them when dryRun is true).
	FixFrontmatter(dryRun bool) ([]domain.FixResult, error)
}
