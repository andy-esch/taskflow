package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// errBadFrontmatter marks a malformed-frontmatter parse failure (vs an I/O
// error), so listing can decide whether to skip or fail. It wraps
// domain.ErrValidation so a malformed file surfaces with the same exit code (11)
// on the single-item read paths (GetTask/GetEpic/GetAudit) that the
// write paths (SetFields/EditBody) already produce — agents route on the code.
var errBadFrontmatter = fmt.Errorf("%w: malformed frontmatter", domain.ErrValidation)

// errNotEntity marks a file whose name is not id-led — a non-entity left in a flat
// entity directory (the carveout gate, ADR-0003 amendment 2026-07-04). It wraps
// ErrValidation so the listing surfaces it as a FileProblem, distinct from a broken
// entity (a real, id-led file whose frontmatter is malformed).
var errNotEntity = fmt.Errorf("%w: not an entity file", domain.ErrValidation)

// errBadEntityID marks a file that IS an entity — a 12-character id leads its name — whose
// id is misspelled. Kept distinct from errNotEntity because the two want opposite advice:
// a stray belongs in meta/, while this one belongs exactly where it is and needs one
// character corrected. Both wrap ErrValidation, so exit-code classification is unchanged.
var errBadEntityID = fmt.Errorf("%w: invalid entity id", domain.ErrValidation)

// FS reads and writes the flat, id-led entity directories under one planning
// root: tasks/, epics/, audits/, and research/.
type FS struct {
	root        string // the planning root; the write-lock (flock) is taken on this dir
	tasksDir    string
	epicsDir    string
	auditsDir   string
	researchDir string
}

// Compile-time assertions that FS satisfies the core ports. The use-case Store is
// the one the Service depends on; Fixer/Layout are the narrow fs/text ports the
// primary adapters (lint --fix, the TUI watcher) wire to the FS directly.
var (
	_ core.Store  = (*FS)(nil)
	_ core.Fixer  = (*FS)(nil)
	_ core.Linter = (*FS)(nil)
	_ core.Layout = (*FS)(nil)
)

// NewFS returns a store rooted at a planning directory (the dir holding tasks/).
func NewFS(root string) *FS {
	return &FS{
		root:        root,
		tasksDir:    filepath.Join(root, domain.TasksDir),
		epicsDir:    filepath.Join(root, domain.EpicsDir),
		auditsDir:   filepath.Join(root, domain.AuditsDir),
		researchDir: filepath.Join(root, domain.ResearchDir),
	}
}

// WatchPaths is the set of leaf directories a filesystem watcher must observe to
// catch every entity change. The store owns the on-disk layout, so this lives here
// rather than being reconstructed by the TUI watcher.
func (s *FS) WatchPaths() []string {
	// Every entity dir is flat (ADR-0003 §4): each is the only watch path for its kind —
	// a status/bucket change is an in-place frontmatter write that fires on the parent
	// dir. Epics were always flat; research has no lifecycle to change at all.
	return []string{s.epicsDir, s.tasksDir, s.auditsDir, s.researchDir}
}

// ListTasks scans the flat task directory and parses each task's frontmatter.
// A file with unreadable frontmatter is skipped and reported as a FileProblem
// (so one bad file doesn't blind the whole listing); err is only for fatal I/O.
func (s *FS) ListTasks() ([]domain.Task, []domain.FileProblem, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return nil, nil, err
	}
	return scanDir(s.tasksDir, func(path string, content []byte) (domain.Task, error) {
		return parseTask(content, path)
	})
}

// ListTasksWithBodies is ListTasks' scan with each task's body kept alongside (one
// pass), so lint's acceptance-criteria checks read every file once — the task twin of
// ListAuditsWithFindings.
func (s *FS) ListTasksWithBodies() ([]core.TaskWithBody, []domain.FileProblem, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return nil, nil, err
	}
	return scanDir(s.tasksDir, func(path string, content []byte) (core.TaskWithBody, error) {
		t, err := parseTask(content, path)
		if err != nil {
			return core.TaskWithBody{}, err
		}
		_, body := splitFrontmatter(content)
		return core.TaskWithBody{Task: t, Body: string(body)}, nil
	})
}

// GetTask returns a single task plus its markdown body.
func (s *FS) GetTask(slug string) (domain.Task, string, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return domain.Task{}, "", err
	}
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

// SetFields surgically updates frontmatter fields on a task (no status/dir
// change) and writes the file atomically in place.
func (s *FS) SetFields(slug string, updates map[string]any, dryRun bool) (domain.Task, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return domain.Task{}, err
	}
	// Defense-in-depth: status is lifecycle-owned and must go through the guarded
	// TaskLifecycleMutationStore capability, never generic frontmatter surgery.
	if _, ok := updates["status"]; ok {
		return domain.Task{}, fmt.Errorf("%w: status is lifecycle-owned — use a task lifecycle verb", domain.ErrValidation)
	}
	for field := range updates {
		if domain.IsGraphOwnedTaskField(field) {
			return domain.Task{}, fmt.Errorf("%w: %s is graph-owned — use guarded dependency operations", domain.ErrValidation, field)
		}
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
	// Re-resolve and re-hash the source immediately before replacement so a rename or
	// concurrent in-place edit cannot be overwritten. Atomicity guards torn writes,
	// not lost updates. ifVersion is the hash of the bytes read above.
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
func (s *FS) resolvePath(id string) (string, error) {
	cands, err := s.taskCandidates()
	if err != nil {
		return "", err
	}
	c, err := resolveExactID(cands, id)
	if err != nil {
		return "", err
	}
	return c.path, nil
}

// taskCandidates lists every flat task file as a resolution candidate — id-led, with
// status read from frontmatter (there is no status directory under the flat layout).
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
	fnID, slug, ok := splitFlatName(strings.TrimSuffix(base, ".md"))
	if !ok {
		reason, kind := entityNameProblem(base)
		return domain.Task{}, fmt.Errorf("%w: %q %s", kind, base, reason)
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
		var fields map[string]yaml.Node
		if err := yaml.Unmarshal(fm, &fields); err != nil {
			return domain.Task{}, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
		}
		for _, field := range []string{"blocked_by", "dependencies", "blocks"} {
			if _, present := fields[field]; present {
				t.LegacyDependencyFields = append(t.LegacyDependencyFields, field)
			}
		}
	}
	// Status is authoritative in frontmatter (ADR-0003 §4). There is no directory to
	// fall back to under the flat layout, but an id-led file with a missing/unrecognized
	// status is still a real task: it LISTS with its raw status and is FLAGGED by lint
	// (StatusFellBack), rather than dropped as a hard read problem. Guarded lifecycle
	// writes fail closed on this state, so repair the raw field and verify with lint. A
	// non-id-led stray is already rejected
	// above; a file with no frontmatter block at all remains a loud FileProblem.
	if !t.Status.Valid() {
		t.StatusFellBack = true
	}
	t.Slug = slug
	t.FilenameID = fnID
	t.Path = path
	t.SourceVersion = hashContent(content)
	return t, nil
}
