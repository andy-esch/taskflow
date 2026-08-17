package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Research persistence. The thinnest entity store in the package, and deliberately
// so: research has no status/bucket, so there is no Move, and no cross-references, so
// there is no reference resolution. What's left is scan, read, and create.

// ListResearch scans the research dir. An unreadable doc is skipped and reported as
// a FileProblem (one bad file doesn't blind the listing); err is only for fatal I/O.
func (s *FS) ListResearch() ([]domain.Research, []domain.FileProblem, error) {
	return scanDir(s.researchDir, func(path string, content []byte) (domain.Research, error) {
		return parseResearch(content, path)
	})
}

// GetResearch returns one research doc plus its markdown body.
func (s *FS) GetResearch(slug string) (domain.Research, string, error) {
	path, err := s.resolveResearch(slug)
	if err != nil {
		return domain.Research{}, "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Research{}, "", fmt.Errorf("read research %s: %w", path, err)
	}
	r, err := parseResearch(content, path)
	if err != nil {
		return domain.Research{}, "", fmt.Errorf("%s: %w", path, err)
	}
	_, body := splitFrontmatter(content)
	return r, string(body), nil
}

// researchCandidates lists every flat research file (research/<id>-<slug>.md) as a
// resolution candidate. Shared by resolveResearch and the create path.
func (s *FS) researchCandidates() ([]candidate, error) {
	return flatCandidates(s.researchDir)
}

// resolveResearch finds a research file by slug — exact first, then fuzzy, matching
// the stable id or the human slug (the same contract as tasks/audits).
func (s *FS) resolveResearch(slug string) (string, error) {
	cands, err := s.researchCandidates()
	if err != nil {
		return "", err
	}
	c, err := resolveID("research", slug, cands)
	if err != nil {
		return "", err
	}
	return c.path, nil
}

// parseResearch reads a flat research file (`<id>-<slug>.md`) into a domain.Research.
// The slug comes from the id-led filename; there is no status or bucket to fall back
// on, so the parse is purely "is this an id-led file with readable frontmatter".
func parseResearch(content []byte, path string) (domain.Research, error) {
	base := filepath.Base(path)
	fnID, slug, ok := splitFlatName(strings.TrimSuffix(base, ".md"))
	if !ok {
		return domain.Research{}, fmt.Errorf("%w: %q has no leading id — move it to meta/ or delete it", errNotEntity, base)
	}
	fm, _, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Research{}, err
	}
	if fm == nil {
		return domain.Research{}, missingFrontmatterErr("research doc", "created; see `tskflwctl schema research`")
	}
	var r domain.Research
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &r); err != nil {
			return domain.Research{}, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
		}
	}
	r.Slug = slug
	r.FilenameID = fnID
	r.Path = path
	return r, nil
}
