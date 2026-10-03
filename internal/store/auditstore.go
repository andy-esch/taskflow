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
	records, problems, err := s.scanAudits()
	if err != nil {
		return nil, nil, err
	}
	audits := make([]domain.Audit, 0, len(records))
	for _, record := range records {
		audits = append(audits, record.Value)
	}
	return audits, problems, nil
}

func (s *FS) scanAudits() ([]core.LoadedRecord[domain.Audit], []domain.FileProblem, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return nil, nil, err
	}
	return scanDir(s.auditsDir, func(path string, content []byte) (core.LoadedRecord[domain.Audit], error) {
		audit, err := parseAudit(content, path)
		if err != nil {
			return core.LoadedRecord[domain.Audit]{}, err
		}
		return auditRecord(audit, path), nil
	})
}

// ListAuditsWithFindings is ListAudits' scan that also keeps the findings parsed
// from each body (the same ParseFindings parseAudit already runs for the tally),
// so consumers read each audit once for both metadata and body-derived views.
func (s *FS) ListAuditsWithFindings() ([]core.AuditWithFindings, []domain.FileProblem, error) {
	records, problems, err := s.scanAuditsWithFindings()
	if err != nil {
		return nil, nil, err
	}
	audits := make([]core.AuditWithFindings, 0, len(records))
	for _, record := range records {
		audits = append(audits, record.Value)
	}
	return audits, problems, nil
}

func (s *FS) scanAuditsWithFindings() ([]core.LoadedRecord[core.AuditWithFindings], []domain.FileProblem, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return nil, nil, err
	}
	return scanDirWithReader(s.auditsDir, s.auditReadFile, func(path string, content []byte) (core.LoadedRecord[core.AuditWithFindings], error) {
		a, findings, nearMisses, candidateIssues, err := parseAuditWithFindings(content, path)
		if err != nil {
			return core.LoadedRecord[core.AuditWithFindings]{}, err
		}
		return core.LoadedRecord[core.AuditWithFindings]{
			Value:  core.AuditWithFindings{Audit: a, Findings: findings, NearMisses: nearMisses, CandidateIssues: candidateIssues},
			Source: auditSource(path),
		}, nil
	})
}

// GetAudit returns one audit plus its markdown body.
func (s *FS) GetAudit(slug string) (domain.Audit, string, error) {
	record, err := s.readAudit(slug)
	if err != nil {
		return domain.Audit{}, "", err
	}
	return record.Value.Audit, record.Value.Body, nil
}

func (s *FS) readAudit(slug string) (core.LoadedRecord[core.AuditWithBody], error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return core.LoadedRecord[core.AuditWithBody]{}, err
	}
	path, err := s.resolveAudit(slug)
	if err != nil {
		return core.LoadedRecord[core.AuditWithBody]{}, err
	}
	content, err := s.auditReadFile(path)
	if err != nil {
		return core.LoadedRecord[core.AuditWithBody]{}, fmt.Errorf("read audit %s: %w", path, err)
	}
	a, err := parseAudit(content, path)
	if err != nil {
		return core.LoadedRecord[core.AuditWithBody]{}, fmt.Errorf("%s: %w", path, err)
	}
	_, body := splitFrontmatter(content)
	return core.LoadedRecord[core.AuditWithBody]{
		Value:  core.AuditWithBody{Audit: a, Body: string(body)},
		Source: auditSource(path),
	}, nil
}

// MoveAudit changes an audit's bucket (close/reopen/defer) by rewriting its authoritative
// `bucket:` frontmatter in place — under the flat layout (ADR-0003 §4) there is no bucket
// directory to move between. Moving to the bucket it already declares is an idempotent no-op.
func (s *FS) MoveAudit(slug string, to domain.AuditBucket, dryRun bool) (domain.Audit, error) {
	if err := s.authorizeMutation(); err != nil {
		return domain.Audit{}, err
	}
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return domain.Audit{}, err
	}
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
// relocation. Nil outside tests.
var testHookBeforeMoveAuditWrite func()

// resolveAuditPath re-resolves an audit by its EXACT stable id for the version-CAS
// guard (verifyUnchanged) — exact-id, never the fuzzy id-OR-slug match resolveAudit uses,
// so a sibling audit whose slug equals this file's id can't lock it (see resolveExactID).
func (s *FS) resolveAuditPath(id string) (string, error) {
	cands, err := s.auditCandidates()
	if err != nil {
		return "", err
	}
	c, err := resolveExactID(cands, id)
	if err != nil {
		return "", err
	}
	return c.path, nil
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
	a, _, _, _, err := parseAuditWithFindings(content, path)
	return a, err
}

// parseAuditWithFindings parses an audit AND returns the findings it parsed to
// compute the tally, plus the near-miss headings that parsed to NOTHING — so a
// sweep that needs them (Summary's rollup, lint) reuses this single pass instead
// of re-reading the body. parseAudit is the wrapper for callers that just want the
// audit + its tally.
func parseAuditWithFindings(content []byte, path string) (domain.Audit, []domain.Finding, []domain.NearMissHeader, []domain.Issue, error) {
	base := filepath.Base(path)
	_, slug, ok := splitFlatName(strings.TrimSuffix(base, ".md"))
	if !ok {
		reason, kind := entityNameProblem(base)
		return domain.Audit{}, nil, nil, nil, fmt.Errorf("%w: %q %s", kind, base, reason)
	}
	fm, body, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Audit{}, nil, nil, nil, err
	}
	if fm == nil {
		return domain.Audit{}, nil, nil, nil, missingFrontmatterErr("audit", "area, date; see `tskflwctl schema audit`")
	}
	var a domain.Audit
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &a); err != nil {
			return domain.Audit{}, nil, nil, nil, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
		}
	}
	a.Slug = slug
	// Bucket is authoritative in frontmatter (ADR-0003 §4). There is no directory to fall
	// back to under the flat layout, but an id-led file with a missing/unrecognized bucket
	// still LISTS (raw bucket) and is FLAGGED (BucketFellBack) — a lifecycle verb heals it.
	if !a.Bucket.Valid() {
		a.BucketFellBack = true
	}
	// The finding grammar (and "what each status means for progress") lives in the
	// domain, so the store just records the tally ParseFindings + TallyFindings report.
	bodyText := string(body)
	findings := domain.ParseFindings(bodyText)
	tally := domain.TallyFindings(findings)
	a.Findings = len(findings)
	a.OpenFindings = tally.Open
	a.ActiveFindings = tally.Active
	a.DoneFindings = tally.Done
	a.DroppedFindings = tally.Dropped
	return a, findings, domain.NearMissFindingHeaders(bodyText), domain.LintCandidateTasks(bodyText, findings), nil
}
