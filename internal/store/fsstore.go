package store

import (
	"errors"
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
// error), so listing can decide whether to skip or fail.
var errBadFrontmatter = errors.New("malformed frontmatter")

// FS reads/writes a planning tree: tasks at <root>/tasks/<status>/<slug>.md
// and epics at <root>/epics/<id>.md.
type FS struct {
	tasksDir  string
	epicsDir  string
	auditsDir string
}

// Compile-time assertion that FS satisfies the full core port.
var _ core.Store = (*FS)(nil)

// NewFS returns a store rooted at a planning directory (the dir holding tasks/).
func NewFS(root string) *FS {
	return &FS{
		tasksDir:  filepath.Join(root, "tasks"),
		epicsDir:  filepath.Join(root, "epics"),
		auditsDir: filepath.Join(root, "audits"),
	}
}

// ListTasks scans every status directory and parses each task's frontmatter.
// A file with unreadable frontmatter is skipped and reported as a FileProblem
// (so one bad file doesn't blind the whole listing); err is only for fatal I/O.
func (s *FS) ListTasks() ([]domain.Task, []domain.FileProblem, error) {
	var tasks []domain.Task
	var problems []domain.FileProblem
	for _, st := range domain.AllStatuses() {
		dir := filepath.Join(s.tasksDir, st.Dir())
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, fmt.Errorf("read status dir %s: %w", dir, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, nil, fmt.Errorf("read task %s: %w", path, err)
			}
			t, err := parseTask(content, path, st)
			if err != nil {
				problems = append(problems, domain.FileProblem{Path: path, Message: err.Error()})
				continue
			}
			tasks = append(tasks, t)
		}
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
func (s *FS) Move(slug string, to domain.Status, now time.Time) (domain.Task, error) {
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
	newDir := filepath.Join(s.tasksDir, to.Dir())
	newPath := filepath.Join(newDir, slug+".md")
	t, err := parseTask(newContent, newPath, to)
	if err != nil {
		return domain.Task{}, err
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
func (s *FS) SetFields(slug string, updates map[string]any) (domain.Task, error) {
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
		return domain.Task{}, fmt.Errorf("%w: update would not reload (%v); nothing was written", domain.ErrValidation, err)
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
	if err := writeFileAtomic(path, newContent, 0o644); err != nil {
		return domain.Task{}, err
	}
	return t, nil
}

// testHookBeforeSetFieldsWrite runs between SetFields' validation and its
// compare-and-swap re-resolve — the seam tests use to interleave a concurrent
// Move. Nil outside tests.
var testHookBeforeSetFieldsWrite func()

// resolve finds the file and current status for an exact slug match.
func (s *FS) resolve(slug string) (path string, status domain.Status, err error) {
	var paths []string
	var statuses []domain.Status
	for _, st := range domain.AllStatuses() {
		p := filepath.Join(s.tasksDir, st.Dir(), slug+".md")
		if info, statErr := os.Stat(p); statErr == nil && !info.IsDir() {
			paths = append(paths, p)
			statuses = append(statuses, st)
		}
	}
	switch len(paths) {
	case 0:
		return "", "", fmt.Errorf("task %q: %w", slug, domain.ErrNotFound)
	case 1:
		return paths[0], statuses[0], nil
	default:
		where := make([]string, len(statuses))
		for i, st := range statuses {
			where[i] = string(st)
		}
		return "", "", fmt.Errorf("%q matches %d tasks (in %s): %w",
			slug, len(paths), strings.Join(where, ", "), domain.ErrAmbiguous)
	}
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
