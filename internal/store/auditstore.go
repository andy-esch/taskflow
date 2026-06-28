package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/core"
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

// ListAuditsWithFindings is ListAudits' scan that also keeps the findings parsed
// from each body (the same ParseFindings parseAudit already runs for the tally),
// so Summary reads each audit once for both the tally and the findings rollup
// instead of re-reading every body through GetAuditByPath.
func (s *FS) ListAuditsWithFindings() ([]core.AuditWithFindings, []domain.FileProblem, error) {
	var out []core.AuditWithFindings
	var problems []domain.FileProblem
	for _, bucket := range domain.AllAuditBuckets() {
		dir := filepath.Join(s.auditsDir, bucket.Dir())
		as, ps, err := scanDir(dir, func(path string, content []byte) (core.AuditWithFindings, error) {
			a, findings, err := parseAuditWithFindings(content, path, bucket)
			return core.AuditWithFindings{Audit: a, Findings: findings}, err
		})
		if err != nil {
			return nil, nil, err
		}
		out = append(out, as...)
		problems = append(problems, ps...)
	}
	return out, problems, nil
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

// GetAuditByPath reads one audit directly by file path, deriving the bucket from
// the parent directory (audits/<bucket>/<slug>.md — the bucket==directory
// invariant the store owns) instead of re-resolving the slug across every bucket.
// The finding/lint sweeps use this to read each audit ListAudits already found
// exactly once, which also closes the concurrent-edit window a re-resolve opens.
func (s *FS) GetAuditByPath(path string) (domain.Audit, string, error) {
	bucket, err := bucketFromPath(path)
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

// bucketFromPath maps an audit file path to its bucket via the parent directory
// name (audits/<bucket>/<slug>.md). A path whose parent isn't a known bucket is
// rejected rather than guessed, so a stray path can't be silently mis-bucketed.
func bucketFromPath(path string) (domain.AuditBucket, error) {
	bucket, err := domain.ParseAuditBucket(filepath.Base(filepath.Dir(path)))
	if err != nil {
		return "", fmt.Errorf("audit path %s: %w", path, err)
	}
	return bucket, nil
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
	// Bucket↔state invariant (the same rule `audit lint` enforces): a non-open
	// bucket must have no still-open findings. Refuse the move rather than write a
	// state the tool's own linter immediately rejects — resolve or defer the
	// findings first. Runs before the dry-run return so a preview fails identically.
	if to != domain.AuditOpen {
		_, body := splitFrontmatter(content)
		if open := domain.CountOpenFindings(domain.ParseFindings(string(body))); open > 0 {
			return domain.Audit{}, fmt.Errorf(
				"%w: audit %q has %d open finding(s); resolve or defer them before moving to %s",
				domain.ErrValidation, slug, open, to)
		}
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
	a, _, err := parseAuditWithFindings(content, path, bucket)
	return a, err
}

// parseAuditWithFindings parses an audit AND returns the findings it parsed to
// compute the tally — so a sweep that needs both (Summary's findings rollup)
// reuses the single ParseFindings call instead of re-reading the body. parseAudit
// is the body-only wrapper for callers that just want the audit + its tally.
func parseAuditWithFindings(content []byte, path string, bucket domain.AuditBucket) (domain.Audit, []domain.Finding, error) {
	fm, body, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Audit{}, nil, err
	}
	var a domain.Audit
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &a); err != nil {
			return domain.Audit{}, nil, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
		}
	}
	a.Slug = strings.TrimSuffix(filepath.Base(path), ".md")
	a.Path = path
	a.Bucket = bucket
	// The finding grammar (and "what each status means for progress") lives in the
	// domain, so the store just records the tally ParseFindings + TallyFindings report.
	findings := domain.ParseFindings(string(body))
	tally := domain.TallyFindings(findings)
	a.Findings = len(findings)
	a.OpenFindings = tally.Open
	a.ActiveFindings = tally.Active
	a.DoneFindings = tally.Done
	a.DroppedFindings = tally.Dropped
	return a, findings, nil
}
