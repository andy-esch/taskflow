// Package core holds the application use cases (the Service) and the ports the
// core needs. Interfaces are defined here, at the consumer, per the org's
// "keep interfaces close to where they're used" guidance.
// Semantic records stay separate from source evidence in [LoadedRecord] and
// [VersionedRecord]. [NewService] validates paired [SourceSetID] witnesses;
// optional local capabilities do not follow from semantic reads. Persistence
// adapters require an explicit [MutationPolicy]. The contributor checklist is
// docs/ARCHITECTURE.md#planning-data-change-checklist; detailed contracts remain beside their types.
package core

import (
	"fmt"
	"sort"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TaskStore is the task mutation and selected-read port. Repository-wide task
// reads use TaskGraphSource so list, board, graph, and Thread projections share
// one authoritative snapshot rather than rescanning through the aggregate store.
type TaskStore interface {
	ReadTasks() (TaskRead, error)
	ReadTask(ref string) (LoadedRecord[TaskWithBody], error)
	// Ordinary task mutations take dryRun: true runs every validation and returns
	// the would-be result but stops short of disk. Lifecycle changes are deliberately
	// absent: TaskLifecycleMutationStore is the only status-write capability.
	SetFields(slug string, updates map[string]any, dryRun bool) (domain.Task, error)
	CreateTask(t domain.Task, body string, dryRun bool) (TaskCreationReceipt, error)
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
	// TransformTaskBody is the read-modify-write counterpart to EditBody. The store
	// passes transform the body from the exact source snapshot its content CAS
	// protects, so a semantic edit can be safely re-applied after contention rather
	// than replacing a concurrent write with bytes derived from an earlier read.
	// transform is pure application logic and must not call back into the Store.
	TransformTaskBody(slug string, now time.Time, dryRun bool, transform func(current string) (string, error)) (domain.Task, string, bool, error)
	// RenameTask re-titles a task: a new slug from newTitle, the file renamed (id kept),
	// the body H1 rewritten, and every inbound relative-path markdown link across the tree
	// repointed to the new filename. Its result records a durable multi-document prefix
	// so post-commit failures are recoverable without guessing whether a retry is safe.
	RenameTask(slug, newTitle string, dryRun bool) (TaskRenameMutationResult, error)
}

// TaskPathSource is optional local, parse-free task navigation. Semantic task
// reads do not imply that their adapter has a filesystem path to resolve.
type TaskPathSource interface {
	ResolveTaskPath(ref string) (string, error)
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

// TaskGraphRepairPlanner is the sole planner admitted to an already-broken
// graph. It returns source-declaration removals, never replacement dependency
// sets, and must not call another store method while the repository guard is held.
type TaskGraphRepairPlanner func(*TaskGraph) (TaskGraphRepairPlan, error)

// TaskGraphRepairStore is deliberately separate from ordinary graph mutation.
// Its implementation owns the guarded task+Thread evidence snapshot, repair-only
// validation, surgical materialization, and durable-prefix receipt.
type TaskGraphRepairStore interface {
	MutateTaskGraphRepair(now time.Time, dryRun bool, planner TaskGraphRepairPlanner) (TaskGraphRepairMutationResult, error)
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

// ThreadReadProblem describes one Thread record that could not be decoded. Stable
// identity is carried explicitly when the adapter can recover it; Location is
// optional repair context such as a local path or remote URI and is never parsed
// by core for identity. SourceVersion is an opaque adapter-owned revision of the
// exact source that produced the problem. Core and guarded stores may compare it,
// but user-facing projections must not publish it.
type ThreadReadProblem struct {
	ThreadID       string
	ThreadSlug     string
	Location       string
	LocationIsPath bool
	Message        string
	SourceVersion  string `json:"-" yaml:"-"`
}

// ThreadRead is one adapter-owned Thread document snapshot. Readable records and
// unreadable problems both carry opaque source revisions so guarded stores can
// qualify the complete source set without exposing persistence-specific types.
type ThreadRead struct {
	Records  []VersionedRecord[domain.Thread]
	Problems []ThreadReadProblem
}

// LoadedThreads preserves readable source identity/location for projections
// without exposing guarded revisions or local mutation handles.
func (read ThreadRead) LoadedThreads() []LoadedRecord[domain.Thread] {
	threads := make([]LoadedRecord[domain.Thread], 0, len(read.Records))
	for _, record := range read.Records {
		threads = append(threads, LoadedRecord[domain.Thread]{Value: cloneThread(record.Record.Value), Source: record.Record.Source})
	}
	return threads
}

// SemanticThreads is the planner-facing view of one authoritative Thread read.
// It preserves source-record order and deliberately does not let a second,
// independently populated slice become mutation evidence. Revisions stay only
// in the guarded wrappers; callers must validate source identities before
// discarding the envelopes for ordinary mutation planning.
func (read ThreadRead) SemanticThreads() []domain.Thread {
	threads := make([]domain.Thread, 0, len(read.Records))
	for _, record := range read.Records {
		threads = append(threads, record.Record.Value)
	}
	return threads
}

// ValidateSources rejects a readable Thread set whose adapter identities cannot
// safely address one occurrence each. Ordinary guarded mutations must call this
// before constructing a semantic planner snapshot. Repair is intentionally not
// gated here: it may need to operate while Thread evidence is incomplete.
func (read ThreadRead) ValidateSources() error {
	seen := make(map[string]string, len(read.Records))
	ordered := append([]VersionedRecord[domain.Thread](nil), read.Records...)
	sort.Slice(ordered, func(i, j int) bool { return threadRecordLess(ordered[i].Record, ordered[j].Record) })
	for _, record := range ordered {
		if err := requireSourceID(EntityThread, record.Record.Source); err != nil {
			return err
		}
		id := record.Record.Source.ID
		name := threadRecordDiagnosticName(record.Record)
		if prior, duplicate := seen[id]; duplicate {
			return fmt.Errorf("%w: duplicate canonical Thread source ID %q across %s and %s", domain.ErrValidation, id, prior, name)
		}
		if declared := record.Record.Value.ID; declared != id {
			return fmt.Errorf("%w: Thread %s source ID %q disagrees with declared id %q", domain.ErrValidation, name, id, declared)
		}
		seen[id] = name
	}
	return nil
}

// ThreadStore is the narrow read capability for first-class Thread documents.
// It remains separate from Store so existing secondary adapters and focused test
// fakes do not claim Thread support accidentally. A Service joined projection
// reads the Thread record(s) first and TaskGraphSource second. Paired adapters must
// ensure that later task read is no older than the Thread read; a lagging or split
// backend must coordinate a compatible snapshot instead of presenting skew as a
// canonical repository view.
type ThreadStore interface {
	ReadThreads() (ThreadRead, error)
	ReadThread(ref string) (LoadedRecord[ThreadWithBody], error)
}

// ThreadPathSource is the optional local-navigation capability behind `thread
// path`. It is deliberately separate from semantic Thread reads: a database,
// API, or cache adapter can provide portable Thread records without inventing a
// filesystem path, while the local filesystem adapter can retain parse-free
// path recovery for malformed documents.
type ThreadPathSource interface {
	ResolveThreadPath(ref string) (string, error)
}

// EpicPathSource is optional local, parse-free epic navigation.
type EpicPathSource interface {
	ResolveEpicPath(ref string) (string, error)
}

// AuditPathSource is optional local, parse-free audit navigation.
type AuditPathSource interface {
	ResolveAuditPath(ref string) (string, error)
}

// ResearchPathSource is optional local, parse-free research navigation.
type ResearchPathSource interface {
	ResolveResearchPath(ref string) (string, error)
}

// ThreadCreationMutationStore owns guarded, unstarted Thread creation. The
// control-inverted planner receives only immutable semantic snapshot values.
type ThreadCreationMutationStore interface {
	MutateThreadCreation(now time.Time, dryRun bool, planner ThreadCreationPlanner) (ThreadCreationMutationResult, error)
}

// ThreadMutationStore owns guarded updates to an existing Thread document.
// Membership and lifecycle intents share the capability so neither can validate
// against a stale task/Thread snapshot or nest another repository guard.
type ThreadMutationStore interface {
	MutateThread(now time.Time, dryRun bool, planner ThreadMutationPlanner) (ThreadMutationResult, error)
}

// ThreadApplyMutationStore owns the one guarded compound operation used by a
// materialized bulk-link plan. Implementations must re-read planning identity,
// tasks, and Threads under one guard; apply dependency writes before the new
// Thread; and preserve the exact durable operation prefix on failure.
type ThreadApplyMutationStore interface {
	MutateThreadApply(now time.Time, dryRun bool, planner ThreadApplyPlanner) (ThreadApplyMutationResult, error)
}

// EpicStore is the entity-specific epic persistence port.
type EpicStore interface {
	ReadEpics() (EpicRead, error)
	ReadEpic(ref string) (LoadedRecord[EpicWithBody], error)
	CreateEpic(slug string, e domain.Epic, body string, dryRun bool) (EpicCreationReceipt, error)
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
	// NearMisses are headings that read as findings but do not parse as one, so
	// they contribute NOTHING to Findings above. Derived on the same scan for the
	// same reason Findings is: a dropped finding is invisible to any later pass
	// over the parsed set, and re-reading every audit to look for it would undo
	// the one-read contract this type exists to keep.
	NearMisses []domain.NearMissHeader
	// CandidateIssues are managed candidate-task v1 defects derived from the same
	// body read. Legacy unversioned candidate prose deliberately contributes none.
	CandidateIssues []domain.Issue
}

// TaskWithBody is a task plus its markdown body, kept together by
// ListTasksWithBodies so lint's body-aware checks read each file once.
type TaskWithBody struct {
	Task domain.Task
	Body string
}

// AuditStore is the audit-persistence port.
type AuditStore interface {
	ReadAudits() (AuditRead, error)
	ReadAudit(ref string) (LoadedRecord[AuditWithBody], error)
	MoveAudit(slug string, to domain.AuditBucket, dryRun bool) (domain.Audit, error)
	CreateAudit(a domain.Audit, body string, dryRun bool) (AuditCreationReceipt, error)
	// EditAudit hands the current file content to edit (the caller's editor) and
	// accepts the result only if it still parses as an audit — parse-before-accept,
	// looping on a broken edit. A changed save is stamped with updated_at (now);
	// reports whether the file changed. The audit counterpart to EditTask;
	// finding-level lint is the caller's to surface.
	EditAudit(slug string, now time.Time, edit func(current string, prevErr error) (string, error)) (domain.Audit, bool, error)
	// AppendAuditBody adds markdown to an audit's narrative in one atomic, validated
	// write, preserving a trailing managed Candidate tasks section as the final
	// projection and stamping updated_at (the audit's `date` stays immutable — it's
	// the slug). The agent face of audit body editing, beside EditAudit's editor.
	// Returns the reloaded audit and the resulting body.
	AppendAuditBody(slug, text string, now time.Time, dryRun bool) (domain.Audit, string, error)
	// TransformAuditBody is the read-modify-write counterpart to AppendAuditBody,
	// and the audit twin of TransformTaskBody. transform receives the body from the
	// exact snapshot the write's content CAS protects, together with that snapshot's
	// parsed audit metadata, so core can enforce bucket-sensitive invariants without a
	// racy read before the transform. On ErrConflict core can simply re-invoke the
	// operation and have the edit recomputed against fresh metadata and text.
	// Reports whether the normalized body actually changed; an unchanged body is a
	// no-op that stamps nothing.
	TransformAuditBody(slug string, now time.Time, dryRun bool, transform func(audit domain.Audit, current string) (string, error)) (domain.Audit, string, bool, error)
}

// ResearchStore is the research-persistence port. The narrowest entity port: research
// has no lifecycle (so no Move) and no cross-references (so nothing to resolve), which
// leaves scan, read, path, and create.
type ResearchStore interface {
	ReadResearch() (ResearchRead, error)
	ReadResearchDocument(ref string) (LoadedRecord[ResearchWithBody], error)
	CreateResearch(r domain.Research, body string, dryRun bool) (ResearchCreationReceipt, error)
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
// deliberately narrow: only the task/epic/audit/research persistence lives here.
// Optional repair, body-link, and completion capabilities are selected separately
// by Service; Layout remains a local watcher capability. A second Store need not
// implement optional workflows. NewService requires a non-zero SourceSetProvider witness at
// runtime; keeping that check at composition lets one complete adapter and
// independently supplied narrow ports share the same rule.
type Store interface {
	TaskStore
	EpicStore
	AuditStore
	ResearchStore
}

// SummaryStore is the non-task read port required by one dashboard scan. Tasks come
// from TaskGraphSource; audits reuse the same body-aware snapshot as finding queries.
type SummaryStore interface {
	ReadEpics() (EpicRead, error)
}

// Fixer is the optional frontmatter-repair persistence port. RepairPlanning owns
// orchestration; the adapter owns safe edits, authorization, and partial results.
type Fixer interface {
	// FixFrontmatter applies safe text-level frontmatter repairs across all
	// planning documents (or previews them when dryRun is true).
	// An error must retain the completed/proposed prefix; callers must not retry
	// the multi-document operation wholesale. Authorization belongs to the adapter.
	FixFrontmatter(dryRun bool) ([]domain.FixResult, error)
}

// Linter is the optional body cross-link integrity port used by LintWithLinks.
// Diagnostics stay adapter neutral; a non-filesystem adapter may check its own
// reference model without exposing local paths.
type Linter interface {
	// DanglingLinks reports unresolved body references. The filesystem adapter
	// checks local markdown links whose target .md file is missing.
	DanglingLinks() ([]LoadProblem, error)
}

// Layout is the on-disk-layout port: the desired directory set a filesystem
// watcher must observe, including configured entity directories that do not exist
// yet. The store owns the layout convention, so consumers recover desired paths
// instead of rebuilding the flat entity-directory shape themselves.
type Layout interface {
	WatchPaths() []string
}
