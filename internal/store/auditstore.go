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
	return scanDir(s.auditsDir, func(path string, content []byte) (domain.Audit, error) {
		return parseAudit(content, path)
	})
}

// ListAuditsWithFindings is ListAudits' scan that also keeps the findings parsed
// from each body (the same ParseFindings parseAudit already runs for the tally),
// so Summary reads each audit once for both the tally and the findings rollup
// instead of re-reading every body through GetAuditByPath.
func (s *FS) ListAuditsWithFindings() ([]core.AuditWithFindings, []domain.FileProblem, error) {
	return scanDir(s.auditsDir, func(path string, content []byte) (core.AuditWithFindings, error) {
		a, findings, err := parseAuditWithFindings(content, path)
		return core.AuditWithFindings{Audit: a, Findings: findings}, err
	})
}

// GetAudit returns one audit plus its markdown body.
func (s *FS) GetAudit(slug string) (domain.Audit, string, error) {
	path, err := s.resolveAudit(slug)
	if err != nil {
		return domain.Audit{}, "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Audit{}, "", fmt.Errorf("read audit %s: %w", path, err)
	}
	a, err := parseAudit(content, path)
	if err != nil {
		return domain.Audit{}, "", fmt.Errorf("%s: %w", path, err)
	}
	_, body := splitFrontmatter(content)
	return a, string(body), nil
}

// GetAuditByPath reads one audit directly by file path (bucket comes from
// frontmatter under the flat layout, ADR-0003 §4) instead of re-resolving the
// slug. The finding/lint sweeps use this to read each audit ListAudits already
// found exactly once, which also closes the concurrent-edit window a re-resolve opens.
func (s *FS) GetAuditByPath(path string) (domain.Audit, string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Audit{}, "", fmt.Errorf("read audit %s: %w", path, err)
	}
	a, err := parseAudit(content, path)
	if err != nil {
		return domain.Audit{}, "", fmt.Errorf("%s: %w", path, err)
	}
	_, body := splitFrontmatter(content)
	return a, string(body), nil
}

// MoveAudit changes an audit's bucket (close/reopen/defer) by rewriting its authoritative
// `bucket:` frontmatter in place — under the flat layout (ADR-0003 §4) there is no bucket
// directory to move between. Moving to the bucket it already declares is an idempotent no-op.
func (s *FS) MoveAudit(slug string, to domain.AuditBucket, dryRun bool) (domain.Audit, error) {
	if !to.Valid() {
		return domain.Audit{}, fmt.Errorf("%q: %w", to, domain.ErrValidation)
	}
	path, err := s.resolveAudit(slug)
	if err != nil {
		return domain.Audit{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Audit{}, fmt.Errorf("read audit %s: %w", path, err)
	}
	// Under the flat layout bucket lives only in frontmatter (ADR-0003 §4) — a move is
	// a pure in-place frontmatter edit, no relocation.
	cur, err := parseAudit(content, path)
	if err != nil {
		return domain.Audit{}, err
	}
	from := cur.Bucket
	// Bucket↔state invariant (the same rule `audit lint` enforces): a non-open bucket
	// must have no still-open findings. Refuse rather than write a state the linter
	// immediately rejects. Runs before the dry-run return so a preview fails identically.
	if to != domain.AuditOpen {
		_, body := splitFrontmatter(content)
		if open := domain.CountOpenFindings(domain.ParseFindings(string(body))); open > 0 {
			return domain.Audit{}, fmt.Errorf(
				"%w: audit %q has %d open finding(s); resolve or defer them before moving to %s",
				domain.ErrValidation, slug, open, to)
		}
	}
	// No-op: already in the target bucket (there is no relocation to owe under flat).
	if from == to {
		return cur, nil
	}
	newContent, err := updateFrontmatter(content, map[string]any{"bucket": string(to)})
	if err != nil {
		return domain.Audit{}, err
	}
	// Parse before committing: a file that wouldn't read back fails with nothing written.
	// The file path never changes — the move is an in-place frontmatter edit.
	a, err := parseAudit(newContent, path)
	if err != nil {
		return domain.Audit{}, err
	}
	if dryRun {
		return a, nil // resolved + parsed; only the write is skipped
	}
	if testHookBeforeMoveAuditWrite != nil {
		testHookBeforeMoveAuditWrite()
	}
	// Serialize the verify→write critical section (flock) so the version-CAS is atomic.
	unlock, err := s.writeLock()
	if err != nil {
		return domain.Audit{}, err
	}
	defer unlock()
	// Version-CAS before the write: re-hash the source so a concurrent in-place edit is
	// caught (no relocation under the flat layout). Fail cleanly with nothing written.
	if err := verifyUnchanged(s.resolveAuditPath, slug, path, hashContent(content), "audit", "move"); err != nil {
		return domain.Audit{}, err
	}
	if err := writeFileAtomic(path, newContent, 0o644); err != nil {
		return domain.Audit{}, err
	}
	return a, nil
}

// testHookBeforeMoveAuditWrite runs between MoveAudit's validation and its
// compare-and-swap re-resolve — the seam tests use to interleave a concurrent
// relocation. Nil outside tests (mirrors testHookBeforeMoveWrite).
var testHookBeforeMoveAuditWrite func()

// resolveAuditPath is s.resolveAudit reduced to (path, error) — the adapter the
// version-CAS guard (verifyUnchanged) takes.
func (s *FS) resolveAuditPath(slug string) (string, error) {
	return s.resolveAudit(slug)
}

// auditCandidates lists every flat audit file (audits/<id>-<slug>.md) as a
// resolution candidate. Shared by resolveAudit and the create path.
func (s *FS) auditCandidates() ([]candidate, error) {
	return flatCandidates(s.auditsDir)
}

// resolveAudit finds an audit file by slug — exact first, then fuzzy, matching the
// stable id or the human slug. Under the flat layout it returns just the path;
// bucket is read from frontmatter, not the (now absent) directory.
func (s *FS) resolveAudit(slug string) (string, error) {
	cands, err := s.auditCandidates()
	if err != nil {
		return "", err
	}
	c, err := resolveID("audit", slug, cands)
	if err != nil {
		return "", err
	}
	return c.path, nil
}

func parseAudit(content []byte, path string) (domain.Audit, error) {
	a, _, err := parseAuditWithFindings(content, path)
	return a, err
}

// parseAuditWithFindings parses an audit AND returns the findings it parsed to
// compute the tally — so a sweep that needs both (Summary's findings rollup)
// reuses the single ParseFindings call instead of re-reading the body. parseAudit
// is the body-only wrapper for callers that just want the audit + its tally.
func parseAuditWithFindings(content []byte, path string) (domain.Audit, []domain.Finding, error) {
	base := filepath.Base(path)
	fnID, slug, ok := splitFlatName(strings.TrimSuffix(base, ".md"))
	if !ok {
		return domain.Audit{}, nil, fmt.Errorf("%w: %q has no leading id — move it to meta/ or delete it", errNotEntity, base)
	}
	fm, body, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Audit{}, nil, err
	}
	if fm == nil {
		return domain.Audit{}, nil, missingFrontmatterErr("audit", "area, date; see `tskflwctl schema audit`")
	}
	var a domain.Audit
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &a); err != nil {
			return domain.Audit{}, nil, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
		}
	}
	a.Slug = slug
	a.FilenameID = fnID
	a.Path = path
	// Bucket is authoritative in frontmatter (ADR-0003 §4). There is no directory to fall
	// back to under the flat layout, but an id-led file with a missing/unrecognized bucket
	// still LISTS (raw bucket) and is FLAGGED (BucketFellBack) — a lifecycle verb heals it.
	if !a.Bucket.Valid() {
		a.BucketFellBack = true
	}
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
