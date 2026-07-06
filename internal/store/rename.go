package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// mdLinkRe matches a markdown link `[display](target)` — the Scheme-2 body-link form.
var mdLinkRe = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)

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
		rewritten, n := repointLinks(content, oldName, newName, oldSlug, newSlug)
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

// repointLinks rewrites every markdown link whose target basename is oldName to newName,
// freshening a link whose display text was exactly oldSlug to newSlug. Returns the new
// content and the count of links repointed. A no-op (0) when the name is unchanged.
func repointLinks(content []byte, oldName, newName, oldSlug, newSlug string) ([]byte, int) {
	if oldName == newName {
		return content, 0
	}
	n := 0
	out := mdLinkRe.ReplaceAllFunc(content, func(m []byte) []byte {
		sub := mdLinkRe.FindSubmatch(m)
		display, target := string(sub[1]), string(sub[2])
		linkPath, anchor := target, ""
		if i := strings.IndexByte(linkPath, '#'); i >= 0 {
			linkPath, anchor = linkPath[:i], linkPath[i:]
		}
		if filepath.Base(linkPath) != oldName {
			return m
		}
		n++
		if display == oldSlug {
			display = newSlug
		}
		return []byte("[" + display + "](" + linkPath[:len(linkPath)-len(oldName)] + newName + anchor + ")")
	})
	return out, n
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
