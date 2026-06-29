package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
)

// ListEpics parses every epics/*.md file. Unreadable epics are skipped and
// reported as FileProblems (resilient, like ListTasks).
func (s *FS) ListEpics() ([]domain.Epic, []domain.FileProblem, error) {
	epics, problems, err := scanDir(s.epicsDir, func(path string, content []byte) (domain.Epic, error) {
		return parseEpic(content, path)
	})
	if err != nil {
		return nil, nil, err
	}
	// Numeric order by the NN- prefix (10 after 9), not ReadDir's lexical order.
	sort.Slice(epics, func(i, j int) bool {
		if ni, nj := epicNum(epics[i].ID), epicNum(epics[j].ID); ni != nj {
			return ni < nj
		}
		return epics[i].ID < epics[j].ID
	})
	return epics, problems, nil
}

// GetEpic returns one epic plus its markdown body. The id resolves exact
// first, then fuzzy (unique prefix/substring), like task and audit slugs.
func (s *FS) GetEpic(id string) (domain.Epic, string, error) {
	cands, err := markdownCandidates(s.epicsDir, "") // epics have no status/bucket dir
	if err != nil {
		return domain.Epic{}, "", err
	}
	c, err := resolveID("epic", id, cands)
	if err != nil {
		return domain.Epic{}, "", err
	}
	content, err := os.ReadFile(c.path)
	if err != nil {
		return domain.Epic{}, "", fmt.Errorf("read epic %s: %w", c.path, err)
	}
	path := c.path
	ep, err := parseEpic(content, path)
	if err != nil {
		return domain.Epic{}, "", fmt.Errorf("%s: %w", path, err)
	}
	_, body := splitFrontmatter(content)
	return ep, string(body), nil
}

// MoveEpic surgically rewrites an epic's `status` frontmatter field. Unlike a
// task/audit move, epic status is a FIELD not a directory, so the file stays put
// — only the field is rewritten (unknown fields, comments, and key order survive).
// Moving to the current status is an idempotent no-op write. Mirrors SetFields'
// parse-before-commit guard: a status that wouldn't reload is rejected with the
// file untouched.
func (s *FS) MoveEpic(id, status string, now time.Time, dryRun bool) (domain.Epic, error) {
	if err := domain.ValidateEpicStatus(status); err != nil {
		return domain.Epic{}, err
	}
	cands, err := markdownCandidates(s.epicsDir, "") // epics have no status/bucket dir
	if err != nil {
		return domain.Epic{}, err
	}
	c, err := resolveID("epic", id, cands)
	if err != nil {
		return domain.Epic{}, err
	}
	path := c.path
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Epic{}, fmt.Errorf("read epic %s: %w", path, err)
	}
	updates := map[string]any{"status": status}
	// Stamp updated_at only on a real status change — a no-op move to the current
	// status isn't an edit. cur is best-effort: a currently-unparseable epic that the
	// move repairs counts as an edit and gets the stamp too.
	if cur, perr := parseEpic(content, path); perr != nil || cur.Status != status {
		updates["updated_at"] = now.Format("2006-01-02")
	}
	newContent, err := updateFrontmatter(content, updates)
	if err != nil {
		return domain.Epic{}, err
	}
	// Parse before committing: never leave an unreloadable file on disk.
	ep, err := parseEpic(newContent, path)
	if err != nil {
		return domain.Epic{}, fmt.Errorf("%w: update would not reload (%v); nothing was written", domain.ErrValidation, err)
	}
	if dryRun {
		return ep, nil // validated end-to-end; only the write is skipped
	}
	if err := writeFileAtomic(path, newContent, 0o644); err != nil {
		return domain.Epic{}, err
	}
	return ep, nil
}

// SetEpicFields surgically updates frontmatter fields on an epic and writes the
// file atomically in place. Unlike a task SetFields there is no status/directory
// concern (epics live flat, status is a field), so no compare-and-swap re-resolve
// is needed — the file never moves. Mirrors the task SetFields parse-before-commit
// guard: an update whose result wouldn't reload is rejected with the file
// untouched (ErrValidation, not a FileProblem — the user's update is bad, the file
// on disk was never the cause).
func (s *FS) SetEpicFields(id string, updates map[string]any, dryRun bool) (domain.Epic, error) {
	cands, err := markdownCandidates(s.epicsDir, "") // epics have no status/bucket dir
	if err != nil {
		return domain.Epic{}, err
	}
	c, err := resolveID("epic", id, cands)
	if err != nil {
		return domain.Epic{}, err
	}
	path := c.path
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Epic{}, fmt.Errorf("read epic %s: %w", path, err)
	}
	newContent, err := updateFrontmatter(content, updates)
	if err != nil {
		return domain.Epic{}, err
	}
	// Parse before committing: never leave an unreloadable file on disk. Attribute
	// a parse failure correctly — if the ORIGINAL already fails the same way, blame
	// the file, not the user's update (mirrors the task SetFields path).
	ep, err := parseEpic(newContent, path)
	if err != nil {
		if _, perr := parseEpic(content, path); perr != nil {
			return domain.Epic{}, fmt.Errorf("%w: %s already has malformed frontmatter (not caused by this update): %v", domain.ErrValidation, path, perr)
		}
		return domain.Epic{}, fmt.Errorf("%w: update would not reload (%v); nothing was written", domain.ErrValidation, err)
	}
	if dryRun {
		return ep, nil // validated end-to-end; only the write is skipped
	}
	if err := writeFileAtomic(path, newContent, 0o644); err != nil {
		return domain.Epic{}, err
	}
	return ep, nil
}

// EditEpic is the epic counterpart to EditTask: resolve id, read the file, and run
// the shared editor-loop (parse-before-accept), accepting a save only if it still
// parses as an epic. Epics never move directories, so there is no compare-and-swap
// concern — recheck is nil. Returns the reloaded epic and whether it changed.
func (s *FS) EditEpic(id string, now time.Time, edit func(current string, prevErr error) (string, error)) (domain.Epic, bool, error) {
	cands, err := markdownCandidates(s.epicsDir, "") // epics have no status/bucket dir
	if err != nil {
		return domain.Epic{}, false, err
	}
	c, err := resolveID("epic", id, cands)
	if err != nil {
		return domain.Epic{}, false, err
	}
	path := c.path
	orig, err := os.ReadFile(path)
	if err != nil {
		return domain.Epic{}, false, fmt.Errorf("read epic %s: %w", path, err)
	}
	return editFile("epic", path, orig, now,
		func(content []byte) (domain.Epic, error) { return parseEpic(content, path) },
		nil, // epics never move directories — no compare-and-swap needed
		edit)
}

func parseEpic(content []byte, path string) (domain.Epic, error) {
	fm, _, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Epic{}, err
	}
	var ep domain.Epic
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &ep); err != nil {
			return domain.Epic{}, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
		}
	}
	ep.ID = strings.TrimSuffix(filepath.Base(path), ".md")
	ep.Path = path
	return ep, nil
}
