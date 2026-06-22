package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
)

// ListAudits scans every audit bucket. Unreadable audits are skipped and
// reported as FileProblems.
func (s *FS) ListAudits() ([]domain.Audit, []domain.FileProblem, error) {
	var audits []domain.Audit
	var problems []domain.FileProblem
	for _, bucket := range domain.AllAuditBuckets() {
		dir := filepath.Join(s.auditsDir, bucket.Dir())
		as, ps, err := scanDir(dir, func(path string, content []byte) (domain.Audit, error) {
			return parseAudit(content, path, bucket)
		})
		if err != nil {
			return nil, nil, err
		}
		audits = append(audits, as...)
		problems = append(problems, ps...)
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
	if testHookBeforeMoveAuditWrite != nil {
		testHookBeforeMoveAuditWrite()
	}
	// Re-resolve immediately before the rename (compare-and-swap), like Move/
	// SetFields: a concurrent relocation may have already moved this slug to another
	// bucket, so renaming from the now-stale path would fail with a generic error
	// (exit 1) instead of the exit-14 retry signal. Fail cleanly with nothing moved.
	if curPath, _, err := s.resolveAudit(slug); err != nil || curPath != path {
		return domain.Audit{}, fmt.Errorf("audit %q changed on disk during move; retry: %w", slug, domain.ErrConflict)
	}
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		return domain.Audit{}, fmt.Errorf("mkdir %s: %w", newDir, err)
	}
	if err := os.Rename(path, newPath); err != nil {
		return domain.Audit{}, fmt.Errorf("move audit: %w", err)
	}
	return a, nil
}

// testHookBeforeMoveAuditWrite runs between MoveAudit's validation and its
// compare-and-swap re-resolve — the seam tests use to interleave a concurrent
// relocation. Nil outside tests (mirrors testHookBeforeMoveWrite).
var testHookBeforeMoveAuditWrite func()

// auditCandidates lists every audit file across all buckets as a resolution
// candidate (the dir name IS the bucket). Shared by resolveAudit and the
// create-time cross-bucket collision check.
func (s *FS) auditCandidates() ([]candidate, error) {
	var cands []candidate
	for _, b := range domain.AllAuditBuckets() {
		cs, err := markdownCandidates(filepath.Join(s.auditsDir, b.Dir()), b.Dir())
		if err != nil {
			return nil, err
		}
		cands = append(cands, cs...)
	}
	return cands, nil
}

// resolveAudit finds an audit by slug — exact first, then fuzzy, like resolve.
func (s *FS) resolveAudit(slug string) (path string, bucket domain.AuditBucket, err error) {
	cands, err := s.auditCandidates()
	if err != nil {
		return "", "", err
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
	// The finding grammar (and "what counts as open") lives in the domain, so the
	// store just counts what ParseFindings reports.
	findings := domain.ParseFindings(string(body))
	a.Findings = len(findings)
	a.OpenFindings = domain.CountOpenFindings(findings)
	return a, nil
}
