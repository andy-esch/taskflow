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

// FS reads/writes a planning tree: tasks at <root>/tasks/<status>/<slug>.md
// and epics at <root>/epics/<id>.md.
type FS struct {
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
	dirs := []string{s.epicsDir, s.tasksDir, s.auditsDir}
	for _, st := range domain.AllStatuses() {
		dirs = append(dirs, filepath.Join(s.tasksDir, st.Dir()))
	}
	for _, b := range domain.AllAuditBuckets() {
		dirs = append(dirs, filepath.Join(s.auditsDir, b.Dir()))
	}
	return dirs
}

// ListTasks scans every status directory and parses each task's frontmatter.
// A file with unreadable frontmatter is skipped and reported as a FileProblem
// (so one bad file doesn't blind the whole listing); err is only for fatal I/O.
func (s *FS) ListTasks() ([]domain.Task, []domain.FileProblem, error) {
	var tasks []domain.Task
	var problems []domain.FileProblem
	for _, st := range domain.AllStatuses() {
		dir := filepath.Join(s.tasksDir, st.Dir())
		ts, ps, err := scanDir(dir, func(path string, content []byte) (domain.Task, error) {
			return parseTask(content, path, st)
		})
		if err != nil {
			return nil, nil, err
		}
		tasks = append(tasks, ts...)
		problems = append(problems, ps...)
	}
	return tasks, problems, nil
}

// GetTask returns a single task plus its markdown body.
func (s *FS) GetTask(slug string) (domain.Task, string, error) {
	path, st, err := s.resolve(slug)
	if err != nil {
		return domain.Task{}, "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Task{}, "", fmt.Errorf("read task %s: %w", path, err)
	}
	t, err := parseTask(content, path, st)
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
	if !to.Valid() {
		return domain.Task{}, fmt.Errorf("%q: %w", to, domain.ErrValidation)
	}
	path, from, err := s.resolve(slug)
	if err != nil {
		return domain.Task{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Task{}, fmt.Errorf("read task %s: %w", path, err)
	}
	if from == to { // idempotent no-op
		return parseTask(content, path, to)
	}

	date := now.Format("2006-01-02")
	updates := map[string]any{"status": string(to), "updated_at": date}
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

	newContent, err := updateFrontmatter(content, updates)
	if err != nil {
		return domain.Task{}, err
	}
	// Parse before committing: if the updated content wouldn't read back, fail
	// with nothing on disk changed. Parsing *after* the move reported failure on
	// a move that had already happened — a phantom failure a retrying caller
	// (or agent) would act on.
	//
	// The destination filename comes from the RESOLVED path, never the query:
	// with fuzzy resolution, `task complete retr` must not rename the file to
	// retr.md.
	canonical := strings.TrimSuffix(filepath.Base(path), ".md")
	newDir := filepath.Join(s.tasksDir, to.Dir())
	newPath := filepath.Join(newDir, canonical+".md")
	t, err := parseTask(newContent, newPath, to)
	if err != nil {
		return domain.Task{}, err
	}
	if dryRun {
		return t, nil // every check above ran; only the disk mutation is skipped
	}
	if testHookBeforeMoveWrite != nil {
		testHookBeforeMoveWrite()
	}
	// Re-resolve immediately before the write (compare-and-swap), like SetFields:
	// a concurrent Move may have already relocated this slug, so writing the new
	// file would leave a duplicate across two status dirs (a permanent
	// ErrAmbiguous). Fail cleanly with nothing written instead.
	if curPath, _, err := s.resolve(slug); err != nil || curPath != path {
		return domain.Task{}, fmt.Errorf("task %q changed on disk during move; retry: %w", slug, domain.ErrConflict)
	}
	// Write the updated content atomically into the *target* status dir, then
	// remove the old file last. A crash between the two leaves both files (a
	// recoverable duplicate), never one whose frontmatter status disagrees with
	// its directory — so the status==directory invariant is never broken.
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		return domain.Task{}, fmt.Errorf("mkdir %s: %w", newDir, err)
	}
	if err := writeFileAtomic(newPath, newContent, 0o644); err != nil {
		return domain.Task{}, err
	}
	if err := os.Remove(path); err != nil {
		return domain.Task{}, fmt.Errorf("remove old task file %s: %w", path, err)
	}
	return t, nil
}

// SetFields surgically updates frontmatter fields on a task (no status/dir
// change) and writes the file atomically in place.
func (s *FS) SetFields(slug string, updates map[string]any, dryRun bool) (domain.Task, error) {
	path, st, err := s.resolve(slug)
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
	t, err := parseTask(newContent, path, st)
	if err != nil {
		// Attribute the failure correctly: if the ORIGINAL file already fails to
		// parse the same way (e.g. pre-existing duplicate keys — a merge artifact
		// updateFrontmatter rewrites only the first of), blame the file, not the
		// user's update. Otherwise it's the update that wouldn't reload.
		if _, perr := parseTask(content, path, st); perr != nil {
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
	if testHookBeforeSetFieldsWrite != nil {
		testHookBeforeSetFieldsWrite()
	}
	// Re-resolve immediately before the write (compare-and-swap): a concurrent
	// Move may have relocated the file, and renaming onto the *original* path
	// would resurrect the slug in its old status directory — a permanent
	// ErrAmbiguous with no repair tooling. Atomicity alone only guards against
	// torn writes, not lost updates.
	if curPath, _, err := s.resolve(slug); err != nil || curPath != path {
		return domain.Task{}, fmt.Errorf("task %q changed on disk during update; retry: %w", slug, domain.ErrConflict)
	}
	if dryRun {
		return t, nil // validated end-to-end; only the write is skipped
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

// resolve finds the file and current status for a slug — exact first, then
// fuzzy (unique case-insensitive prefix, then substring) via resolveID, so a
// half-remembered name works anywhere a slug is accepted. Ambiguity (including
// the same slug in two status dirs) is an explicit ErrAmbiguous listing the
// candidates.
func (s *FS) resolve(slug string) (path string, status domain.Status, err error) {
	cands, err := s.taskCandidates()
	if err != nil {
		return "", "", err
	}
	c, err := resolveID("task", slug, cands)
	if err != nil {
		return "", "", err
	}
	return c.path, domain.Status(c.dir), nil
}

// taskCandidates lists every task file as a resolution candidate (the dir name
// IS the status, per the status==directory invariant).
func (s *FS) taskCandidates() ([]candidate, error) {
	var out []candidate
	for _, st := range domain.AllStatuses() {
		cs, err := markdownCandidates(filepath.Join(s.tasksDir, st.Dir()), st.Dir())
		if err != nil {
			return nil, err
		}
		out = append(out, cs...)
	}
	return out, nil
}

func parseTask(content []byte, path string, dirStatus domain.Status) (domain.Task, error) {
	fm, _, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Task{}, err
	}
	var t domain.Task
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &t); err != nil {
			return domain.Task{}, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
		}
	}
	// The directory is the source of truth for status — always. The frontmatter
	// value is kept as Declared so drift (a misfiled file) can be surfaced, but
	// it never overrides where the file physically lives.
	t.Declared = t.Status
	t.Status = dirStatus
	t.Slug = strings.TrimSuffix(filepath.Base(path), ".md")
	t.Path = path
	return t, nil
}
