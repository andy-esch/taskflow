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
	openFindingRe = regexp.MustCompile(`(?m)\*\*Status:\*\*\s*open(?:[^-\w]|$)`)
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
func (s *FS) MoveAudit(slug string, to domain.AuditBucket) (domain.Audit, error) {
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
	newDir := filepath.Join(s.auditsDir, to.Dir())
	newPath := filepath.Join(newDir, slug+".md")
	// Parse before the rename: a malformed file must fail with the audit still
	// in its original bucket, not move and then report failure.
	a, err := parseAudit(content, newPath, to)
	if err != nil {
		return domain.Audit{}, err
	}
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		return domain.Audit{}, fmt.Errorf("mkdir %s: %w", newDir, err)
	}
	if err := os.Rename(path, newPath); err != nil {
		return domain.Audit{}, fmt.Errorf("move audit: %w", err)
	}
	return a, nil
}

func (s *FS) resolveAudit(slug string) (path string, bucket domain.AuditBucket, err error) {
	var paths []string
	var buckets []domain.AuditBucket
	for _, b := range domain.AllAuditBuckets() {
		p := filepath.Join(s.auditsDir, b.Dir(), slug+".md")
		if info, statErr := os.Stat(p); statErr == nil && !info.IsDir() {
			paths = append(paths, p)
			buckets = append(buckets, b)
		}
	}
	switch len(paths) {
	case 0:
		return "", "", fmt.Errorf("audit %q: %w", slug, domain.ErrNotFound)
	case 1:
		return paths[0], buckets[0], nil
	default:
		return "", "", fmt.Errorf("%q matches %d audits: %w", slug, len(paths), domain.ErrAmbiguous)
	}
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
