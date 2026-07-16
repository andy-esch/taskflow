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
	// ResolveTaskPath returns a task's file path from its slug/id WITHOUT parsing —
	// so `task path` works even on a file whose frontmatter won't parse (the case
	// where you most need the path, to open and repair it).
	ResolveTaskPath(slug string) (string, error)
	// Mutators take dryRun: true runs EVERY validation (resolve, parse-before-
	// commit, collision/CAS checks) and returns the would-be result, but stops
	// short of touching disk — so a dry-run that would fail fails identically.
	Move(slug string, to domain.Status, now time.Time, dryRun bool) (domain.Task, error)
	// Defer moves a task to deferred and, when until is non-empty, records it as
	// revisit_at ("snooze until") in the SAME atomic write — so a deferred task can
	// never be left without the snooze date it was deferred with (the lost-second-
	// write hazard a Move-then-SetFields had). An empty until is exactly
	// Move(StatusDeferred). The caller validates the date; the store writes it
	// verbatim. Re-deferring an already-deferred task rewrites revisit_at in place.
	Defer(slug, until string, now time.Time, dryRun bool) (domain.Task, error)
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

// Linter is the cross-link integrity port. Like Fixer it's an fs/text operation, not a
// core use case, so `lint --links` wires it directly to the FS rather than through the
// Service.
type Linter interface {
	// DanglingLinks reports every body markdown link whose target .md file is missing.
	DanglingLinks() ([]domain.FileProblem, error)
}

// Layout is the on-disk-layout port: the directory set a filesystem watcher must
// observe. The store owns the layout convention, so the TUI watcher consumes this
// instead of rebuilding the tasks/<status> + audits/<bucket> shape itself.
type Layout interface {
	WatchPaths() []string
}
