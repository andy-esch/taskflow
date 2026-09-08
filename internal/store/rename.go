package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// firstH1Re matches the first ATX H1 line (`# title`) in a body.
var firstH1Re = regexp.MustCompile(`(?m)^# .*$`)

// RenameTask re-titles a task: it derives a new slug from newTitle, renames the file
// (id preserved — `<id>-<old>.md` → `<id>-<new>.md`), rewrites the body H1 to the new
// title, and CASCADES — every inbound relative-path markdown link across the planning
// tree that points at the old filename is repointed to the new one (and a link whose
// display text was the bare old slug is refreshed to the new slug). Returns the reloaded
// task and a durable-prefix receipt.
//
// Real writes capture the caller's source version before waiting, then acquire the
// canonical repository guard, reject a changed source, compile the cascade from the
// guarded current tree, and CAS every document immediately before replacing it. This
// prevents two concurrent renames from both committing and preserves cooperating writes
// to inbound-link documents. A dry run performs the same planning without a reservation.
func (s *FS) RenameTask(slug, newTitle string, dryRun bool) (result core.TaskRenameMutationResult, err error) {
	result.DryRun = dryRun
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return result, err
	}
	source, err := s.snapshotTaskRenameSource(slug)
	if err != nil {
		return result, err
	}
	newSlug := domain.Slugify(newTitle)
	if newSlug == "" {
		return result, fmt.Errorf("%w: title produced an empty slug: %q", domain.ErrValidation, newTitle)
	}

	if dryRun {
		plan, err := s.prepareTaskRename(source, newTitle, newSlug)
		if err != nil {
			return result, err
		}
		return taskRenameResult(plan, true), nil
	}
	if testHookBeforeTaskRenameLock != nil {
		testHookBeforeTaskRenameLock()
	}
	unlock, err := s.checkedWriteLock()
	if err != nil {
		return result, err
	}
	defer func() {
		if releaseErr := unlock(); releaseErr != nil {
			err = errors.Join(err, fmt.Errorf("release repository task rename guard: %w", releaseErr))
		}
	}()

	// Preserve the precise target-collision diagnosis even when the colliding file
	// already makes stable-id resolution ambiguous. This check is authoritative
	// because it now occurs after acquiring the repository guard.
	if err := s.ensureTaskRenameTargetAvailable(source, newSlug); err != nil {
		return result, err
	}
	// The pre-lock source version is the intent boundary: a competing rename or
	// source edit that completed while this caller waited makes this operation stale,
	// even if resolving by stable id could find the task at its new path.
	if err := verifyUnchanged(s.resolvePath, source.id, source.path, source.version, "task", "rename"); err != nil {
		return result, err
	}
	plan, err := s.prepareTaskRename(source, newTitle, newSlug)
	if err != nil {
		return result, err
	}
	result = taskRenameResult(plan, false)
	if !result.Changed {
		result.Complete = true
		return result, nil
	}

	// Apply same-path inbound-link rewrites first. Any durable prefix is convergent:
	// a retry scans the remaining old links and leaves already-repointed links alone.
	for _, edit := range plan.cascadeEdits {
		if testHookBeforeTaskRenameWrite != nil {
			testHookBeforeTaskRenameWrite(edit.path)
		}
		if err := verifyTaskRenamePath(edit.path, edit.ifVersion); err != nil {
			return result, err
		}
		if err := writeFileAtomic(edit.path, edit.content, 0o644); err != nil {
			return result, fmt.Errorf("write task rename cascade document %s: %w", edit.path, err)
		}
		result.AppliedDocuments++
		result.AppliedLinks += edit.links
		result.Committed = true
		if testHookAfterTaskRenameWrite != nil {
			if err := testHookAfterTaskRenameWrite(edit.path); err != nil {
				return result, fmt.Errorf("after task rename cascade document %s: %w", edit.path, err)
			}
		}
	}

	// Recheck the source immediately before materializing the destination. The
	// repository guard excludes cooperating writers; this CAS bounds the remaining
	// raw-editor window and prevents a stale H1/body from replacing newer bytes.
	if testHookBeforeTaskRenameWrite != nil {
		testHookBeforeTaskRenameWrite(source.path)
	}
	if err := verifyUnchanged(s.resolvePath, source.id, source.path, source.version, "task", "rename"); err != nil {
		return result, err
	}
	sourceInfo, err := os.Stat(source.path)
	if err != nil {
		return result, fmt.Errorf("stat task %s before rename destination create: %w", source.id, err)
	}
	if plan.newPath != source.path && testHookBeforeTaskRenameDestinationCreate != nil {
		testHookBeforeTaskRenameDestinationCreate(plan.newPath)
	}
	if plan.newPath == source.path {
		if err := writeFileAtomic(source.path, plan.renamedContent, 0o644); err != nil {
			return result, fmt.Errorf("write renamed task %s: %w", source.id, err)
		}
	} else if err := createFileAtomicExactMode(plan.newPath, plan.renamedContent, sourceInfo.Mode().Perm()); err != nil {
		if os.IsExist(err) {
			return result, fmt.Errorf("%w: target filename already exists: %s", domain.ErrConflict, filepath.Base(plan.newPath))
		}
		return result, fmt.Errorf("create renamed task %s: %w", source.id, err)
	}
	result.AppliedDocuments++
	result.AppliedLinks += plan.targetLinks
	result.Committed = true
	result.DestinationWritten = true
	if testHookAfterTaskRenameWrite != nil {
		if err := testHookAfterTaskRenameWrite(plan.newPath); err != nil {
			return result, fmt.Errorf("after writing renamed task %s: %w", source.id, err)
		}
	}

	if plan.newPath != source.path {
		if testHookBeforeTaskRenameSourceRemove != nil {
			if err := testHookBeforeTaskRenameSourceRemove(source.path, plan.newPath); err != nil {
				return result, fmt.Errorf("before removing old task %s: %w", source.id, err)
			}
		}
		// A raw editor does not honor the repository guard. Recheck the old source
		// bytes once more before deleting that path; after the destination exists,
		// a conflict intentionally leaves both copies for explicit inspection.
		if err := verifyTaskRenamePath(source.path, source.version); err != nil {
			return result, err
		}
		if err := os.Remove(source.path); err != nil {
			return result, fmt.Errorf("remove old %s: %w", source.path, err)
		}
		result.SourceRemoved = true
	}
	result.Complete = true
	return result, nil
}

type taskRenameSource struct {
	id      string
	oldSlug string
	path    string
	version string
}

func (s *FS) snapshotTaskRenameSource(ref string) (taskRenameSource, error) {
	path, err := s.resolve(ref)
	if err != nil {
		return taskRenameSource{}, err
	}
	id, oldSlug, ok := splitFlatName(strings.TrimSuffix(filepath.Base(path), ".md"))
	if !ok {
		return taskRenameSource{}, fmt.Errorf("%w: %q is not an id-led task file", errNotEntity, filepath.Base(path))
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return taskRenameSource{}, fmt.Errorf("task %q changed on disk during rename; retry: %w", ref, domain.ErrConflict)
		}
		return taskRenameSource{}, fmt.Errorf("read task %q for rename: %w", ref, err)
	}
	return taskRenameSource{id: id, oldSlug: oldSlug, path: path, version: hashContent(content)}, nil
}

type taskRenameEdit struct {
	path      string
	ifVersion string
	content   []byte
	links     int
}

type taskRenamePlan struct {
	task           domain.Task
	source         taskRenameSource
	newPath        string
	renamedContent []byte
	targetChanged  bool
	targetLinks    int
	cascadeEdits   []taskRenameEdit
	plannedLinks   int
}

func taskRenameResult(plan taskRenamePlan, dryRun bool) core.TaskRenameMutationResult {
	plannedDocuments := len(plan.cascadeEdits)
	if plan.targetChanged {
		plannedDocuments++
	}
	return core.TaskRenameMutationResult{
		Task: plan.task, FromSlug: plan.source.oldSlug,
		PlannedDocuments: plannedDocuments,
		PlannedLinks:     plan.plannedLinks,
		Changed:          plannedDocuments > 0,
		DryRun:           dryRun,
	}
}

func (s *FS) prepareTaskRename(source taskRenameSource, newTitle, newSlug string) (taskRenamePlan, error) {
	oldName := source.id + "-" + source.oldSlug + ".md"
	newName := source.id + "-" + newSlug + ".md"
	newPath := filepath.Join(filepath.Dir(source.path), newName)

	// This check now runs inside the repository guard for real writes. The final
	// create is also O_EXCL, preserving no-clobber protection from raw writers.
	if err := s.ensureTaskRenameTargetAvailable(source, newSlug); err != nil {
		return taskRenamePlan{}, err
	}

	// Build every edit in one tree walk: the renamed file gets its H1 rewritten (and any
	// self-links repointed); every other file gets its inbound links repointed.
	plan := taskRenamePlan{source: source, newPath: newPath}
	cascade := 0
	err := filepath.WalkDir(s.root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !markdownDoc(d) {
			return nil
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		isTarget := p == source.path
		if isTarget {
			content = replaceFirstH1(content, newTitle)
		}
		rewritten, n := repointLinks(content, filepath.Dir(p), source.path, oldName, newName, source.oldSlug, newSlug)
		switch {
		case isTarget:
			plan.renamedContent = rewritten
			plan.targetChanged = newPath != source.path || hashContent(rewritten) != source.version
			plan.targetLinks = n
			cascade += n
		case n > 0:
			plan.cascadeEdits = append(plan.cascadeEdits, taskRenameEdit{
				path: p, ifVersion: hashContent(content), content: rewritten, links: n,
			})
			cascade += n
		}
		return nil
	})
	if err != nil {
		return taskRenamePlan{}, err
	}
	if plan.renamedContent == nil {
		return taskRenamePlan{}, fmt.Errorf("task %s changed on disk during rename planning; retry: %w", source.id, domain.ErrConflict)
	}
	// Parse-before-commit: the renamed file must still read back as a task, or nothing changes.
	t, err := parseTask(plan.renamedContent, newPath)
	if err != nil {
		return taskRenamePlan{}, err
	}
	plan.task = t
	plan.plannedLinks = cascade
	return plan, nil
}

func (s *FS) ensureTaskRenameTargetAvailable(source taskRenameSource, newSlug string) error {
	newName := source.id + "-" + newSlug + ".md"
	newPath := filepath.Join(filepath.Dir(source.path), newName)
	if newPath == source.path {
		return nil
	}
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("%w: target filename already exists: %s", domain.ErrConflict, newName)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat target %s: %w", newPath, err)
	}
	return nil
}

func verifyTaskRenamePath(path, ifVersion string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("task rename document %s changed on disk; retry: %w", path, domain.ErrConflict)
		}
		return fmt.Errorf("re-read task rename document %s: %w", path, err)
	}
	if hashContent(content) != ifVersion {
		return fmt.Errorf("task rename document %s changed on disk; retry: %w", path, domain.ErrConflict)
	}
	return nil
}

var testHookBeforeTaskRenameLock func()
var testHookBeforeTaskRenameWrite func(path string)
var testHookAfterTaskRenameWrite func(path string) error
var testHookBeforeTaskRenameDestinationCreate func(path string)
var testHookBeforeTaskRenameSourceRemove func(oldPath, newPath string) error

// repointLinks rewrites every markdown link — inline or reference-style — that RESOLVES to
// oldPath (relative to sourceDir — the file the link lives in) so its filename becomes
// newName, freshening an inline link whose display text was the old slug or full stem.
// Matching by resolved path, not bare basename, means a same-named file in a different
// directory is left untouched; links inside fenced code blocks (examples) are skipped. A
// trailing #fragment or ?query is split off before resolving and re-appended. Returns the
// new content and the count of links repointed; a no-op (0) when the name is unchanged.
func repointLinks(content []byte, sourceDir, oldPath, oldName, newName, oldSlug, newSlug string) ([]byte, int) {
	if oldName == newName {
		return content, 0
	}
	oldStem, newStem := strings.TrimSuffix(oldName, ".md"), strings.TrimSuffix(newName, ".md")
	// Collect non-overlapping (target, and optionally display) edits, then splice once.
	type edit struct {
		start, end int
		repl       string
	}
	var edits []edit
	n := 0
	for _, r := range scanLinks(content) {
		linkPath, suffix := r.target, ""
		if i := strings.IndexAny(linkPath, "#?"); i >= 0 {
			linkPath, suffix = linkPath[:i], linkPath[i:]
		}
		if linkPath == "" || filepath.Clean(filepath.Join(sourceDir, filepath.FromSlash(linkPath))) != oldPath {
			continue
		}
		n++
		edits = append(edits, edit{r.tStart, r.tEnd, linkPath[:len(linkPath)-len(oldName)] + newName + suffix})
		if r.inline { // a reference-style [label]: key is not display text — never rewrite it
			switch r.display {
			case oldSlug:
				edits = append(edits, edit{r.dStart, r.dEnd, newSlug})
			case oldStem:
				edits = append(edits, edit{r.dStart, r.dEnd, newStem})
			}
		}
	}
	if n == 0 {
		return content, 0
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	out := make([]byte, 0, len(content))
	last := 0
	for _, e := range edits {
		out = append(out, content[last:e.start]...)
		out = append(out, e.repl...)
		last = e.end
	}
	return append(out, content[last:]...), n
}

// replaceFirstH1 rewrites the FIRST `# …` line of content to `# newTitle` (the re-title).
// A body with no H1 is returned unchanged.
func replaceFirstH1(content []byte, newTitle string) []byte {
	loc := firstH1Re.FindIndex(content)
	if loc == nil {
		return content
	}
	out := make([]byte, 0, len(content)+len(newTitle))
	out = append(out, content[:loc[0]]...)
	out = append(out, "# "+newTitle...)
	out = append(out, content[loc[1]:]...)
	return out
}
