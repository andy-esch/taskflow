package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// errBadFrontmatter marks a malformed-frontmatter parse failure (vs an I/O
// error), so listing can decide whether to skip or fail. It wraps
// domain.ErrValidation so a malformed file surfaces with the same exit code (11)
// on the single-item read/move paths (GetTask/GetEpic/GetAudit/Move) that the
// write paths (SetFields/EditBody) already produce — agents route on the code.
var errBadFrontmatter = fmt.Errorf("%w: malformed frontmatter", domain.ErrValidation)

// errNotEntity marks a file whose name is not id-led — a non-entity left in a flat
// entity directory (the carveout gate, ADR-0003 amendment 2026-07-04). It wraps
// ErrValidation so the listing surfaces it as a FileProblem, distinct from a broken
// entity (a real, id-led file whose frontmatter is malformed).
var errNotEntity = fmt.Errorf("%w: not an entity file", domain.ErrValidation)

// FS reads/writes a planning tree: tasks at <root>/tasks/<status>/<slug>.md
// and epics at <root>/epics/<id>.md.
type FS struct {
	root      string // the planning root; the write-lock (flock) is taken on this dir
	tasksDir  string
	epicsDir  string
	auditsDir string
}

// Compile-time assertions that FS satisfies the core ports. The use-case Store is
// the one the Service depends on; Fixer/Layout are the narrow fs/text ports the
// primary adapters (lint --fix, the TUI watcher) wire to the FS directly.
var (
	_ core.Store  = (*FS)(nil)
	_ core.Fixer  = (*FS)(nil)
	_ core.Layout = (*FS)(nil)
)

// NewFS returns a store rooted at a planning directory (the dir holding tasks/).
func NewFS(root string) *FS {
	return &FS{
		root:      root,
		tasksDir:  filepath.Join(root, domain.TasksDir),
		epicsDir:  filepath.Join(root, domain.EpicsDir),
		auditsDir: filepath.Join(root, domain.AuditsDir),
	}
}

// WatchPaths is the set of leaf directories a filesystem watcher must observe to
// catch every task/epic/audit change: the three entity parents plus each
// task-status and audit-bucket subdir. The store owns the on-disk layout, so
// this lives here rather than being reconstructed by the TUI watcher (which
// would otherwise duplicate the `tasks/<status>` / `audits/<bucket>` convention).
func (s *FS) WatchPaths() []string {
	// Tasks and audits are both flat now (ADR-0003 §4): each entity dir is the only
	// watch path for its kind — a status/bucket change is an in-place frontmatter write
	// that fires on the parent dir. Epics were always flat.
	return []string{s.epicsDir, s.tasksDir, s.auditsDir}
}

// ListTasks scans every status directory and parses each task's frontmatter.
// A file with unreadable frontmatter is skipped and reported as a FileProblem
// (so one bad file doesn't blind the whole listing); err is only for fatal I/O.
func (s *FS) ListTasks() ([]domain.Task, []domain.FileProblem, error) {
	return scanDir(s.tasksDir, func(path string, content []byte) (domain.Task, error) {
		return parseTask(content, path)
	})
}

// GetTask returns a single task plus its markdown body.
func (s *FS) GetTask(slug string) (domain.Task, string, error) {
	path, err := s.resolve(slug)
	if err != nil {
		return domain.Task{}, "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Task{}, "", fmt.Errorf("read task %s: %w", path, err)
	}
	t, err := parseTask(content, path)
	if err != nil {
		return domain.Task{}, "", fmt.Errorf("%s: %w", path, err)
	}
	_, body := splitFrontmatter(content)
	return t, string(body), nil
}

// Move transitions a task to status `to`: it updates frontmatter (status +
// dates) and relocates the file to the target status directory. Moving to the
// current status is an idempotent no-op.
func (s *FS) Move(slug string, to domain.Status, now time.Time, dryRun bool) (domain.Task, error) {
	return s.moveTask(slug, to, now, dryRun, nil)
}

// Defer moves a task to deferred and records `until` as revisit_at in the SAME
// atomic write (the audit-M4 fix), so the relocation and the snooze date can't
// land separately — a Move-then-SetFields could leave a task deferred without its
// date if the second write failed. An empty until is a plain move to deferred;
// re-deferring an already-deferred task rewrites revisit_at in place.
func (s *FS) Defer(slug, until string, now time.Time, dryRun bool) (domain.Task, error) {
	var extra map[string]any
	if until != "" {
		extra = map[string]any{"revisit_at": until}
	}
	return s.moveTask(slug, domain.StatusDeferred, now, dryRun, extra)
}

// moveTask is the shared engine behind Move and Defer: it ensures the task ends up
// in the `to` status dir with the status/date stamps plus any `extra` frontmatter,
// in ONE atomic write. A real transition (from != to) relocates the file; an
// in-place rewrite (from == to, used by a re-defer that carries a new revisit_at)
// overwrites the existing file. When nothing would change it's an idempotent no-op.
func (s *FS) moveTask(slug string, to domain.Status, now time.Time, dryRun bool, extra map[string]any) (domain.Task, error) {
	if !to.Valid() {
		return domain.Task{}, fmt.Errorf("%q: %w", to, domain.ErrValidation)
	}
	path, err := s.resolve(slug)
	if err != nil {
		return domain.Task{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Task{}, fmt.Errorf("read task %s: %w", path, err)
	}
	// Under the flat layout status lives only in frontmatter (ADR-0003 §4) — there is
	// no directory to disagree with it, so a move is a pure in-place frontmatter edit.
	cur, err := parseTask(content, path)
	if err != nil {
		return domain.Task{}, err
	}
	from := cur.Status

	date := now.Format("2006-01-02")
	updates := map[string]any{}
	if from != to {
		// A real transition: stamp status, the activity date, and the destination's
		// entry date.
		updates["status"] = string(to)
		updates["updated_at"] = date
		switch to {
		case domain.StatusInProgress:
			updates["started_at"] = date
		case domain.StatusCompleted:
			updates["completed_at"] = date
		case domain.StatusDeprecated:
			updates["deprecated_at"] = date
		case domain.StatusDeferred:
			updates["deferred_at"] = date
		}
		// revisit_at is a live "snooze until" intent that only makes sense while a
		// task is parked in deferred. Leaving deferred (resume via next/ready, or
		// any other move) ends the snooze, so clear it — mirroring how entering a
		// state stamps its date. (deleteMapNode is a no-op when none is set.) A
		// re-defer (from==to==deferred) skips this branch, so an existing date is
		// kept unless `extra` overwrites it below.
		if from == domain.StatusDeferred {
			updates["revisit_at"] = domain.UnsetField{}
		}
	}
	// extra (Defer's revisit_at) rides the same write. When the status isn't
	// changing — a re-defer in place — it's the only field change, and we still
	// stamp updated_at so the activity date advances like the field-set path does.
	if len(extra) > 0 {
		if from == to {
			updates["updated_at"] = date
		}
		for k, v := range extra {
			updates[k] = v
		}
	}
	// No-op: a move to the current status with no extra field changes.
	if len(updates) == 0 {
		return cur, nil
	}
	newContent, err := updateFrontmatter(content, updates)
	if err != nil {
		return domain.Task{}, err
	}
	// Parse before committing: if the updated content wouldn't read back, fail with
	// nothing on disk changed. The file path never changes — a move is an in-place
	// frontmatter edit under the flat layout (no relocation, no dual-file window).
	t, err := parseTask(newContent, path)
	if err != nil {
		return domain.Task{}, err
	}
	if dryRun {
		return t, nil // every check above ran; only the disk mutation is skipped
	}
	if testHookBeforeMoveWrite != nil {
		testHookBeforeMoveWrite()
	}
	// Serialize the verify→write critical section (flock) so the version-CAS is atomic.
	unlock, err := s.writeLock()
	if err != nil {
		return domain.Task{}, err
	}
	defer unlock()
	// Version-CAS immediately before the write: re-hash the source so a concurrent
	// in-place edit is caught. The flat layout has no relocation, so content drift is
	// the only hazard; fail cleanly with nothing written.
	if err := verifyUnchanged(s.resolvePath, slug, path, hashContent(content), "task", "move"); err != nil {
		return domain.Task{}, err
	}
	if err := writeFileAtomic(path, newContent, 0o644); err != nil {
		return domain.Task{}, err
	}
	return t, nil
}

// SetFields surgically updates frontmatter fields on a task (no status/dir
// change) and writes the file atomically in place.
func (s *FS) SetFields(slug string, updates map[string]any, dryRun bool) (domain.Task, error) {
	// Defense-in-depth: a status change relocates the file, so it must go through Move,
	// never an in-place field write (the core SetFields already rejects it). A direct
	// store caller writing status here would desync the mirror dir from the frontmatter.
	if _, ok := updates["status"]; ok {
		return domain.Task{}, fmt.Errorf("%w: status is not a settable field — use Move", domain.ErrValidation)
	}
	path, err := s.resolve(slug)
	if err != nil {
		return domain.Task{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Task{}, fmt.Errorf("read task %s: %w", path, err)
	}
	newContent, err := updateFrontmatter(content, updates)
	if err != nil {
		return domain.Task{}, err
	}
	// Parse before committing: never leave an unreloadable file on disk. If the
	// updated frontmatter wouldn't read back (e.g. a value serialized to the wrong
	// YAML type), reject without writing rather than corrupt the source of truth.
	// The error is a *validation* failure (the update is bad, exit 11) — not a
	// file problem; the message must not blame a file that was never touched.
	t, err := parseTask(newContent, path)
	if err != nil {
		// Attribute the failure correctly: if the ORIGINAL file already fails to
		// parse the same way (e.g. pre-existing duplicate keys — a merge artifact
		// updateFrontmatter rewrites only the first of), blame the file, not the
		// user's update. Otherwise it's the update that wouldn't reload.
		if _, perr := parseTask(content, path); perr != nil {
			return domain.Task{}, fmt.Errorf("%w: %s already has malformed frontmatter (not caused by this update): %v", domain.ErrValidation, path, perr)
		}
		return domain.Task{}, fmt.Errorf("%w: update would not reload (%v); nothing was written", domain.ErrValidation, err)
	}
	// `set` must not be able to write a file the tool's own linter rejects: an
	// active task with emptied tags, or a next-up/in-progress task with its
	// description cleared. NewTask applies the identical domain rule at creation, so
	// the create and mutate paths can't diverge. Runs before the dry-run return so a
	// preview fails identically.
	if err := domain.ActiveTaskFieldErr(t); err != nil {
		return domain.Task{}, err
	}
	// A dry-run is a preview: it ran every validation above but writes nothing, so it takes
	// neither the write lock nor the version-CAS (both write-time concerns) — consistent
	// with the movers, which also return before locking/verifying on a dry-run.
	if dryRun {
		return t, nil
	}
	if testHookBeforeSetFieldsWrite != nil {
		testHookBeforeSetFieldsWrite()
	}
	// Serialize the verify→write critical section (flock) so the version-CAS is atomic.
	unlock, err := s.writeLock()
	if err != nil {
		return domain.Task{}, err
	}
	defer unlock()
	// Version-CAS immediately before the write: verifyUnchanged re-resolves (a concurrent
	// Move relocated the file → renaming onto the original path would resurrect the slug
	// in its old status dir, a permanent ErrAmbiguous) AND re-hashes the source (a
	// concurrent in-place edit the path-CAS alone missed). Atomicity guards only torn
	// writes, not lost updates. ifVersion is the hash of the bytes read above.
	if err := verifyUnchanged(s.resolvePath, slug, path, hashContent(content), "task", "update"); err != nil {
		return domain.Task{}, err
	}
	if err := writeFileAtomic(path, newContent, 0o644); err != nil {
		return domain.Task{}, err
	}
	return t, nil
}

// testHookBeforeSetFieldsWrite runs between SetFields' validation and its
// compare-and-swap re-resolve — the seam tests use to interleave a concurrent
// Move. Nil outside tests.
var testHookBeforeSetFieldsWrite func()

// testHookBeforeMoveWrite is Move's equivalent seam: it runs just before Move's
// compare-and-swap re-resolve, so a test can interleave a concurrent relocation.
// Nil outside tests.
var testHookBeforeMoveWrite func()

// resolve finds a task file by slug — exact first, then fuzzy (unique
// case-insensitive prefix, then substring) via resolveID, matching on the stable
// id or the human slug. Under the flat layout it returns just the path; status is
// read from the file's frontmatter, not its (now absent) directory.
func (s *FS) resolve(slug string) (string, error) {
	cands, err := s.taskCandidates()
	if err != nil {
		return "", err
	}
	c, err := resolveID("task", slug, cands)
	if err != nil {
		return "", err
	}
	return c.path, nil
}

// resolvePath is s.resolve reduced to (path, error) — the adapter the version-CAS guard
// (verifyUnchanged) takes, so the guard stays entity-agnostic across tasks and audits.
func (s *FS) resolvePath(slug string) (string, error) {
	return s.resolve(slug)
}

// taskCandidates lists every task file as a resolution candidate (the dir name
// IS the status, per the status==directory invariant).
func (s *FS) taskCandidates() ([]candidate, error) {
	return flatCandidates(s.tasksDir)
}

// parseTask reads a flat task file (`<id>-<slug>.md`) into a domain.Task. The slug
// comes from the id-led filename (splitFlatName); status is read purely from
// frontmatter — under the flat layout (ADR-0003 §4) there is no directory to fall
// back to, so a missing/unknown status is a hard read problem (flatten trap #3), and
// a non-id-led filename is a carveout stray (errNotEntity) — except a README, which
// is skipped silently.
func parseTask(content []byte, path string) (domain.Task, error) {
	base := filepath.Base(path)
	_, slug, ok := splitFlatName(strings.TrimSuffix(base, ".md"))
	if !ok {
		return domain.Task{}, fmt.Errorf("%w: %q has no leading id — move it to meta/ or delete it", errNotEntity, base)
	}
	fm, _, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Task{}, err
	}
	if fm == nil {
		return domain.Task{}, missingFrontmatterErr("task", "status, epic, tier, priority, effort, created, tags; see `tskflwctl schema task`")
	}
	var t domain.Task
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &t); err != nil {
			return domain.Task{}, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
		}
	}
	// Status is authoritative in frontmatter (ADR-0003 §4). There is no directory to
	// fall back to under the flat layout, but an id-led file with a missing/unrecognized
	// status is still a real task: it LISTS with its raw status and is FLAGGED by lint
	// (StatusFellBack), rather than dropped as a hard read problem — and a lifecycle verb
	// heals it (moveTask writes a valid status). A non-id-led stray is already rejected
	// above; a file with no frontmatter block at all remains a loud FileProblem.
	if !t.Status.Valid() {
		t.StatusFellBack = true
	}
	t.Slug = slug
	t.Path = path
	return t, nil
}
