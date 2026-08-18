package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// resolveResearchPathExact re-resolves a research doc by its EXACT stable id for the
// version-CAS guard (verifyUnchanged) — never the fuzzy id-OR-slug match resolveResearch
// uses, so a sibling doc whose slug equals this file's id can't lock it (see
// resolveExactID). Mirrors resolveAuditPath.
func (s *FS) resolveResearchPathExact(entityID string) (string, error) {
	cands, err := s.researchCandidates()
	if err != nil {
		return "", err
	}
	c, err := resolveExactID(cands, entityID)
	if err != nil {
		return "", err
	}
	return c.path, nil
}

// SetResearchFields surgically updates frontmatter fields on a research doc in one
// atomic, validated write. Unknown keys, comments, and key order survive (the vestigial
// `status: reference` on the legacy corpus rides along untouched). The service injects
// updated_at; protected fields are rejected before we get here.
func (s *FS) SetResearchFields(slug string, updates map[string]any, dryRun bool) (domain.Research, error) {
	path, err := s.resolveResearch(slug)
	if err != nil {
		return domain.Research{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Research{}, fmt.Errorf("read research %s: %w", path, err)
	}
	// Refuse a doc with no frontmatter block, exactly as GetResearch does. updateFrontmatter
	// would otherwise CREATE one, producing a doc whose only key is the field just set — no
	// id, no `created`. The parse-before-commit check below can't catch it, because that
	// fabricated content parses fine; the defect is that it shouldn't exist.
	if fm, _, ferr := splitFrontmatterStrict(content); ferr != nil {
		return domain.Research{}, ferr
	} else if fm == nil {
		return domain.Research{}, missingFrontmatterErr("research doc", "created; see `tskflwctl schema research`")
	}
	newContent, err := updateFrontmatter(content, updates)
	if err != nil {
		return domain.Research{}, err
	}
	// Parse before committing: never leave an unreloadable file on disk. Attribute a
	// parse failure correctly — if the ORIGINAL already fails the same way, blame the
	// file, not the user's update (mirrors the task/epic SetFields paths).
	r, err := parseResearch(newContent, path)
	if err != nil {
		if _, perr := parseResearch(content, path); perr != nil {
			return domain.Research{}, fmt.Errorf("%w: %s already has malformed frontmatter (not caused by this update): %v", domain.ErrValidation, path, perr)
		}
		return domain.Research{}, fmt.Errorf("%w: update would not reload (%v); nothing was written", domain.ErrValidation, err)
	}
	if dryRun {
		return r, nil // validated end-to-end; only the write is skipped
	}
	if testHookBeforeResearchWrite != nil {
		testHookBeforeResearchWrite()
	}
	unlock, err := s.writeLock()
	if err != nil {
		return domain.Research{}, err
	}
	defer unlock()
	// Version-CAS: catches a concurrent in-place edit during the read→write window.
	// Keyed on the FILENAME id, the canonical resolution key.
	if err := verifyUnchanged(s.resolveResearchPathExact, r.FilenameID, path, hashContent(content), "research doc", "update"); err != nil {
		return domain.Research{}, err
	}
	if err := writeFileAtomic(path, newContent, 0o644); err != nil {
		return domain.Research{}, err
	}
	return r, nil
}

// EditResearch is the research counterpart to EditAudit: resolve, read, and run the
// shared editor loop (parse-before-accept), accepting a save only if it still parses as
// a research doc. Research never moves (no lifecycle), so there is no relocation to
// guard, but the version-CAS recheck still catches a concurrent edit during the editor
// window. Returns the reloaded doc and whether it changed.
func (s *FS) EditResearch(slug string, now time.Time, edit func(current string, prevErr error) (string, error)) (domain.Research, bool, error) {
	path, err := s.resolveResearch(slug)
	if err != nil {
		return domain.Research{}, false, err
	}
	orig, err := os.ReadFile(path)
	if err != nil {
		return domain.Research{}, false, fmt.Errorf("read research %s: %w", path, err)
	}
	entityID, _, _ := splitFlatName(strings.TrimSuffix(filepath.Base(path), ".md"))
	ifVersion := hashContent(orig)
	return editFile("research doc", path, orig, now,
		func(content []byte) (domain.Research, error) { return parseResearch(content, path) },
		s.writeLock,
		func() error {
			return verifyUnchanged(s.resolveResearchPathExact, entityID, path, ifVersion, "research doc", "edit")
		},
		edit)
}

// AppendResearchBody appends markdown to a research doc's body in one atomic, validated
// write, stamping updated_at. `created` stays immutable — the id is minted from it. The
// agent face of body editing, beside EditResearch's editor.
func (s *FS) AppendResearchBody(slug, text string, now time.Time, dryRun bool) (domain.Research, string, error) {
	path, err := s.resolveResearch(slug)
	if err != nil {
		return domain.Research{}, "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Research{}, "", fmt.Errorf("read research %s: %w", path, err)
	}
	fm, body, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Research{}, "", err // can't body-edit a file whose frontmatter won't parse
	}
	// A write must not be MORE permissive than a read. splitFrontmatterStrict returns a nil
	// block (not an error) for a file with no `---` fence, and documentMapping would then
	// happily CREATE one — so without this guard an append to a frontmatter-less doc
	// succeeds and writes a block whose only key is `updated_at`, leaving the doc with no
	// id and no `created` (the anchor its id is minted from) and its prose `**Created**:`
	// stranded in the body. GetResearch rejects that same file, so the write path has to.
	if fm == nil {
		return domain.Research{}, "", missingFrontmatterErr("research doc", "created; see `tskflwctl schema research`")
	}
	entityID, _, _ := splitFlatName(strings.TrimSuffix(filepath.Base(path), ".md"))
	updatedAt := now.Format("2006-01-02")
	return writeBody(
		"research doc", path, content, appendSection(string(body), text),
		func(c []byte, nb string) ([]byte, error) { return replaceBodyStamped(c, nb, updatedAt) },
		func(c []byte) (domain.Research, error) { return parseResearch(c, path) },
		s.writeLock,
		func() error {
			return verifyUnchanged(s.resolveResearchPathExact, entityID, path, hashContent(content), "research doc", "edit")
		},
		dryRun,
	)
}

// testHookBeforeResearchWrite runs between SetResearchFields' validation and its
// compare-and-swap check, so a test can interleave a concurrent write into that exact
// window (the seam tasks/epics already have — testHookBeforeSetFieldsWrite,
// testHookBeforeEpicWrite). Without it the version-CAS is unreachable from a test, which
// is how research shipped with the CAS entirely uncovered. nil in production.
var testHookBeforeResearchWrite func()
