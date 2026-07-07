package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// firstH1Re matches the first ATX H1 line (`# title`) in a body.
var firstH1Re = regexp.MustCompile(`(?m)^# .*$`)

// RenameTask re-titles a task: it derives a new slug from newTitle, renames the file
// (id preserved — `<id>-<old>.md` → `<id>-<new>.md`), rewrites the body H1 to the new
// title, and CASCADES — every inbound relative-path markdown link across the planning
// tree that points at the old filename is repointed to the new one (and a link whose
// display text was the bare old slug is refreshed to the new slug). Returns the reloaded
// task and the count of inbound links repointed.
//
// It is a multi-file write serialized by the repo write-lock but NOT version-CAS-guarded:
// rename is a rare, deliberate, single-user operation (git is the undo). A dry run runs
// every check and returns the would-be result without touching disk.
func (s *FS) RenameTask(slug, newTitle string, dryRun bool) (domain.Task, int, error) {
	oldPath, err := s.resolve(slug)
	if err != nil {
		return domain.Task{}, 0, err
	}
	id, oldSlug, ok := splitFlatName(strings.TrimSuffix(filepath.Base(oldPath), ".md"))
	if !ok {
		return domain.Task{}, 0, fmt.Errorf("%w: %q is not an id-led task file", errNotEntity, filepath.Base(oldPath))
	}
	newSlug := domain.Slugify(newTitle)
	if newSlug == "" {
		return domain.Task{}, 0, fmt.Errorf("%w: title produced an empty slug: %q", domain.ErrValidation, newTitle)
	}
	oldName, newName := id+"-"+oldSlug+".md", id+"-"+newSlug+".md"
	newPath := filepath.Join(filepath.Dir(oldPath), newName)

	// Refuse to rename onto an existing file: the write loop below would silently
	// clobber it. newPath shares the id, so a collision means a duplicate-id sibling
	// already exists — fail loud on that corrupt state rather than destroy the file.
	if newPath != oldPath {
		if _, err := os.Stat(newPath); err == nil {
			return domain.Task{}, 0, fmt.Errorf("%w: target filename already exists: %s", domain.ErrConflict, newName)
		} else if !os.IsNotExist(err) {
			return domain.Task{}, 0, fmt.Errorf("stat target %s: %w", newPath, err)
		}
	}

	// Build every edit in one tree walk: the renamed file gets its H1 rewritten (and any
	// self-links repointed); every other file gets its inbound links repointed.
	type fileEdit struct {
		path    string
		content []byte
	}
	var edits []fileEdit
	cascade := 0
	var renamedContent []byte
	err = filepath.WalkDir(s.root, func(p string, d os.DirEntry, err error) error {
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
		isTarget := p == oldPath
		if isTarget {
			content = replaceFirstH1(content, newTitle)
		}
		rewritten, n := repointLinks(content, filepath.Dir(p), oldPath, oldName, newName, oldSlug, newSlug)
		switch {
		case isTarget:
			renamedContent = rewritten
			edits = append(edits, fileEdit{newPath, rewritten}) // the renamed file writes to the NEW path
			cascade += n
		case n > 0:
			edits = append(edits, fileEdit{p, rewritten})
			cascade += n
		}
		return nil
	})
	if err != nil {
		return domain.Task{}, 0, err
	}
	if renamedContent == nil {
		return domain.Task{}, 0, fmt.Errorf("rename %s: source file vanished under the walk", slug)
	}
	// Parse-before-commit: the renamed file must still read back as a task, or nothing changes.
	t, err := parseTask(renamedContent, newPath)
	if err != nil {
		return domain.Task{}, 0, err
	}
	if dryRun {
		return t, cascade, nil
	}
	unlock, err := s.writeLock()
	if err != nil {
		return domain.Task{}, 0, err
	}
	defer unlock()
	for _, e := range edits {
		if err := writeFileAtomic(e.path, e.content, 0o644); err != nil {
			return domain.Task{}, 0, err
		}
	}
	if newPath != oldPath {
		if err := os.Remove(oldPath); err != nil {
			return domain.Task{}, 0, fmt.Errorf("remove old %s: %w", oldPath, err)
		}
	}
	return t, cascade, nil
}

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
