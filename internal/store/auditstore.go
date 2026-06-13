package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
)

var (
	// findingHeaderRe matches a finding sub-header like "#### H1." / "### M2." / "#### S3.".
	findingHeaderRe = regexp.MustCompile(`(?m)^#{2,6}\s+[A-Z]+\d+\.`)
	// openFindingRe matches a "**Status:** open" line. The trailing `[^-\w]|$`
	// guard (RE2 has no lookahead) keeps "open-ish"/"openness" from matching
	// while still allowing "open" followed by punctuation/space.
	openFindingRe = regexp.MustCompile(`(?mi)\*\*Status:\*\*\s*open(?:[^-\w]|$)`)
	// fenceRe spans a ```-fenced code block, so example finding/status syntax in
	// docs isn't miscounted as a real finding.
	fenceRe = regexp.MustCompile("(?s)```.*?```")
)

// ListAudits scans every audit bucket. Unreadable audits are skipped and
// reported as FileProblems.
func (s *FS) ListAudits() ([]domain.Audit, []domain.FileProblem, error) {
	var audits []domain.Audit
	var problems []domain.FileProblem
	for _, bucket := range domain.AllAuditBuckets() {
		dir := filepath.Join(s.auditsDir, bucket.Dir())
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, fmt.Errorf("read audit bucket %s: %w", dir, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, nil, fmt.Errorf("read audit %s: %w", path, err)
			}
			a, err := parseAudit(content, path, bucket)
			if err != nil {
				problems = append(problems, domain.FileProblem{Path: path, Message: err.Error()})
				continue
			}
			audits = append(audits, a)
		}
	}
	return audits, problems, nil
}

// GetAudit returns one audit plus its markdown body.
func (s *FS) GetAudit(slug string) (domain.Audit, string, error) {
	path, bucket, err := s.resolveAudit(slug)
	if err != nil {
		return domain.Audit{}, "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Audit{}, "", fmt.Errorf("read audit %s: %w", path, err)
	}
	a, err := parseAudit(content, path, bucket)
	if err != nil {
		return domain.Audit{}, "", fmt.Errorf("%s: %w", path, err)
	}
	_, body := splitFrontmatter(content)
	return a, string(body), nil
}

// MoveAudit relocates an audit to another bucket (close/reopen/defer). Moving to
// the current bucket is an idempotent no-op. The bucket directory is the state,
// so only the file moves (no frontmatter rewrite).
func (s *FS) MoveAudit(slug string, to domain.AuditBucket, dryRun bool) (domain.Audit, error) {
	if !to.Valid() {
		return domain.Audit{}, fmt.Errorf("%q: %w", to, domain.ErrValidation)
	}
	path, from, err := s.resolveAudit(slug)
	if err != nil {
		return domain.Audit{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Audit{}, fmt.Errorf("read audit %s: %w", path, err)
	}
	if from == to {
		return parseAudit(content, path, to)
	}
	// Destination filename from the RESOLVED path, never the query (fuzzy
	// resolution must not rename the file to the abbreviation).
	canonical := strings.TrimSuffix(filepath.Base(path), ".md")
	newDir := filepath.Join(s.auditsDir, to.Dir())
	newPath := filepath.Join(newDir, canonical+".md")
	// Parse before the rename: a malformed file must fail with the audit still
	// in its original bucket, not move and then report failure.
	a, err := parseAudit(content, newPath, to)
	if err != nil {
		return domain.Audit{}, err
	}
	if dryRun {
		return a, nil // resolved + parsed; only the rename is skipped
	}
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		return domain.Audit{}, fmt.Errorf("mkdir %s: %w", newDir, err)
	}
	if err := os.Rename(path, newPath); err != nil {
		return domain.Audit{}, fmt.Errorf("move audit: %w", err)
	}
	return a, nil
}

// resolveAudit finds an audit by slug — exact first, then fuzzy, like resolve.
func (s *FS) resolveAudit(slug string) (path string, bucket domain.AuditBucket, err error) {
	var cands []candidate
	for _, b := range domain.AllAuditBuckets() {
		dir := filepath.Join(s.auditsDir, b.Dir())
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", "", fmt.Errorf("read audit dir %s: %w", dir, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			cands = append(cands, candidate{
				id:   strings.TrimSuffix(e.Name(), ".md"),
				path: filepath.Join(dir, e.Name()),
				dir:  b.Dir(),
			})
		}
	}
	c, err := resolveID("audit", slug, cands)
	if err != nil {
		return "", "", err
	}
	return c.path, domain.AuditBucket(c.dir), nil
}

func parseAudit(content []byte, path string, bucket domain.AuditBucket) (domain.Audit, error) {
	fm, body, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Audit{}, err
	}
	var a domain.Audit
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &a); err != nil {
			return domain.Audit{}, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
		}
	}
	a.Slug = strings.TrimSuffix(filepath.Base(path), ".md")
	a.Path = path
	a.Bucket = bucket
	// Count findings on the prose only — fenced code blocks may contain example
	// headers/status lines that aren't real findings.
	prose := fenceRe.ReplaceAll(body, nil)
	a.Findings = len(findingHeaderRe.FindAll(prose, -1))
	a.OpenFindings = len(openFindingRe.FindAll(prose, -1))
	return a, nil
}
