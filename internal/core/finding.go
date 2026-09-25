package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// CountBy is one bucket of a finding breakdown — a key (an urgency value or a
// top-level component) and how many actionable findings fell in it.
type CountBy struct {
	Key   string
	Count int
}

// FindingsRollup aggregates the ACTIONABLE audit findings (status open or
// in-progress) across all audits — the dashboard / `status` "audit findings" view.
// ByUrgency is in canonical order (acute, soon, eventually, then any others);
// ByComponent is most-findings-first on the top-level component. Acute lists the
// (rare, high-signal) acute findings for a call-out. Open+InProgress == total.
type FindingsRollup struct {
	Open        int
	InProgress  int
	ByUrgency   []CountBy
	ByComponent []CountBy
	Acute       []AuditFinding
}

// urgencyOrder is the canonical triage order; unknown/missing urgencies sort after.
var urgencyOrder = []string{"acute", "soon", "eventually"}

// actionableStatuses is the finding-status subset Summary's rollup surfaces — the
// work that is still outstanding. Kept beside the rollup so it stays the single
// definition of "actionable" the dashboard / `status` filter on (it must agree
// with what QueryFindings(Status:["open","in-progress"]) selects).
var actionableStatuses = []string{"open", "in-progress"}

// isActionableFinding reports whether a finding is open or in-progress — the
// case-insensitive match QueryFindings applies, so the single-scan rollup in
// Summary selects exactly the findings the old QueryFindings second pass did.
func isActionableFinding(fd domain.Finding) bool {
	return anyEqualFold(actionableStatuses, fd.Status)
}

// rollupFindings aggregates a set of actionable findings by urgency and top-level
// component, and collects the acute ones. Pure — the caller supplies the findings
// (Summary queries status open/in-progress).
func rollupFindings(fs []AuditFinding) FindingsRollup {
	var r FindingsRollup
	urg := map[string]int{}
	comp := map[string]int{}
	for _, f := range fs {
		switch strings.ToLower(strings.TrimSpace(f.Status)) {
		case "open":
			r.Open++
		case "in-progress":
			r.InProgress++
		}
		u := strings.ToLower(strings.TrimSpace(f.Urgency))
		if u == "" {
			u = "unspecified"
		}
		urg[u]++
		if u == "acute" {
			r.Acute = append(r.Acute, f)
		}
		if c := topComponent(f.Component); c != "" {
			comp[c]++
		}
	}
	r.ByUrgency = orderedCounts(urg, urgencyOrder)
	r.ByComponent = countsDesc(comp)
	return r
}

// topComponent is the first segment of a finding's component path
// ("stravapipe / write paths" → "stravapipe"), trimmed; "" when unset.
func topComponent(component string) string {
	c := strings.TrimSpace(component)
	if i := strings.IndexByte(c, '/'); i >= 0 {
		c = strings.TrimSpace(c[:i])
	}
	return c
}

// orderedCounts emits the known keys first (in `order`, skipping zeros), then any
// extras by count desc then key asc.
func orderedCounts(m map[string]int, order []string) []CountBy {
	var out []CountBy
	seen := map[string]bool{}
	for _, k := range order {
		if n := m[k]; n > 0 {
			out = append(out, CountBy{Key: k, Count: n})
			seen[k] = true
		}
	}
	var extra []CountBy
	for k, n := range m {
		if !seen[k] {
			extra = append(extra, CountBy{Key: k, Count: n})
		}
	}
	sortCountsDesc(extra)
	return append(out, extra...)
}

// countsDesc returns the map as counts sorted by count desc then key asc.
func countsDesc(m map[string]int) []CountBy {
	out := make([]CountBy, 0, len(m))
	for k, n := range m {
		out = append(out, CountBy{Key: k, Count: n})
	}
	sortCountsDesc(out)
	return out
}

func sortCountsDesc(cs []CountBy) {
	sort.SliceStable(cs, func(i, j int) bool {
		if cs[i].Count != cs[j].Count {
			return cs[i].Count > cs[j].Count
		}
		return cs[i].Key < cs[j].Key
	})
}

// AuditFinding is one parsed finding plus the audit it belongs to, so a
// cross-audit query result stays self-describing (which audit/bucket each hit
// came from).
type AuditFinding struct {
	domain.Finding
	Audit   string // the audit's slug
	AuditID string // canonical store identity; internal navigation must not resolve Audit
	Bucket  string // the audit's bucket (open|closed|deferred)
}

// FindingFilter narrows a finding query. Empty fields match everything. Audit, if
// set, restricts to a single audit (resolved like any slug). Status/Effort/Urgency
// are closed-vocabulary fields matched exactly (case-insensitive, any-of);
// Component is free-form and matched as a case-insensitive substring.
type FindingFilter struct {
	Audit     string
	Status    []string
	Effort    []string
	Urgency   []string
	Component string
}

// QueryFindings selects findings from one portable audit snapshot (or just the
// audit resolved by f.Audit), in (audit order, document order). The snapshot
// keeps metadata, parsed findings, and failed-record identity tied to the same
// adapter read; core never follows a persistence location to reread a body.
func (s *Service) QueryFindings(f FindingFilter) ([]AuditFinding, []LintLoadProblem, error) {
	if isNilCapability(s.auditReads) {
		return nil, nil, fmt.Errorf("audit snapshot reads are unavailable from this service")
	}
	snapshot, err := s.auditReads.ReadAuditSnapshot(f.Audit)
	if err != nil {
		return nil, nil, err
	}
	audits := snapshot.Audits

	var out []AuditFinding
	for _, record := range audits {
		for _, fd := range record.Findings {
			if findingMatches(fd, f) {
				out = append(out, AuditFinding{
					Finding: fd, Audit: record.Audit.Slug, AuditID: record.Audit.CanonicalID(),
					Bucket: string(record.Audit.Bucket),
				})
			}
		}
	}
	return out, snapshot.Problems, nil
}

// SetFindingStatus stamps one finding's status in place, through the audit body-replace
// path so the rest of the file is byte-identical. Returns the audit and whether anything
// changed — false means it already carried that exact value, so no write happened.
//
// This is the validated write path finding H1 of the 2026-08-17 audit asked for. Until now
// the only way to resolve a finding was a hand edit or a scripted search-and-replace, which
// is how a vocabulary drifts from its own documentation — and, in practice, how this repo's
// own audits were maintained.
func (s *Service) SetFindingStatus(slug, code, status string, dryRun bool) (domain.Audit, bool, error) {
	return s.EditFinding(slug, code, FindingEdit{Status: status}, dryRun)
}

// FindingEdit is what one `audit finding` call may change. A zero field is left alone; the
// two travel together so status and note land in ONE atomic write rather than two, which
// would leave a window where the finding claims `fixed` with last round's explanation.
type FindingEdit struct {
	Status    string  // "" leaves the status alone
	Note      *string // nil leaves the note alone; a pointer to "" REMOVES it
	Candidate *string // nil leaves the candidate alone; a pointer to "" REMOVES it
}

// apply runs the edit against a body, re-parsing between steps so the second edit sees the
// first one's offsets. Chaining rather than computing both spans up front is deliberate:
// stale offsets are exactly how finding H1 of the 2026-08-24 audit corrupted files.
func (e FindingEdit) apply(body, code string) (string, error) {
	out := body
	var err error
	if e.Status != "" {
		if out, err = domain.SetFindingStatus(out, code, e.Status); err != nil {
			return "", err
		}
	}
	if e.Note != nil {
		if out, err = domain.SetFindingNote(out, code, *e.Note); err != nil {
			return "", err
		}
	}
	if e.Candidate != nil {
		if out, err = domain.SetFindingCandidate(out, code, *e.Candidate); err != nil {
			return "", err
		}
	}
	return out, nil
}

// EditFinding stamps a finding's status and/or resolution note in place, through the audit
// body-replace path so the rest of the file is byte-identical. Returns the audit and
// whether anything changed — false means it already carried those exact values, so no
// write happened.
//
// This is the validated write path finding H1 of the 2026-08-17 audit asked for. Until now
// the only way to resolve a finding was a hand edit or a scripted search-and-replace, which
// is how a vocabulary drifts from its own documentation — and, in practice, how this repo's
// own audits were maintained.
func (s *Service) EditFinding(slug, code string, edit FindingEdit, dryRun bool) (domain.Audit, bool, error) {
	// The transform is recomputed from whatever body the store hands back, so a
	// conflict is retried rather than refused: the previous editor-callback path had
	// to escape its retry loop with an `attempted` flag because it replayed stale
	// precomputed text. A concurrent `audit append` now costs a retry, not the edit.
	now := s.now()
	type result struct {
		audit   domain.Audit
		changed bool
	}
	r, err := retryOnConflict(s, dryRun, func() (result, error) {
		audit, _, changed, err := s.store.TransformAuditBody(slug, now, dryRun,
			func(_ domain.Audit, current string) (string, error) { return edit.apply(current, code) })
		return result{audit: audit, changed: changed}, err
	})
	return r.audit, r.changed, err
}

// AuditLintIssues is the single audit check-set, shared by `audit lint` and the
// top-level `lint` roster so the two cannot drift. Findings and near-misses arrive
// already derived from the same AuditSnapshot source read for both repository
// sweeps and single-audit queries.
//
// nearMisses is a separate input rather than something recomputed from findings
// because a dropped finding is by construction ABSENT from findings — the parsed
// set can never reveal what failed to parse into it.
func AuditLintIssues(a domain.Audit, findings []domain.Finding, nearMisses []domain.NearMissHeader, candidateIssues []domain.Issue) []domain.Issue {
	iss := domain.NearMissFindingIssues(nearMisses)
	iss = append(iss, domain.LintFindings(string(a.Bucket), findings)...)
	iss = append(iss, candidateIssues...)
	iss = append(iss, domain.MissingIDIssue(a.ID)...)             // audits get a stable id too
	iss = append(iss, domain.IDDriftIssue(a.ID, a.FilenameID)...) // …that must match the filename
	iss = append(iss, domain.FrontmatterBucketIssues(a)...)       // and a missing/foreign bucket flag
	return iss
}

// FixFindingHeaders canonicalizes every near-miss finding header in the repository,
// one guarded body write per audit. This is the repair half of the near-miss
// contract: detection alone still ends with a human or an agent re-editing the
// document by hand, which is the loop the tool exists to remove.
//
// It writes through TransformAuditBody rather than the frontmatter fixer because
// the change is a BODY edit and must carry the audit's content CAS — FixFrontmatter
// deliberately rewrites frontmatter only and passes the body through untouched.
// Audits with nothing to repair are not written at all, so a clean corpus is a no-op.
func (s *Service) FixFindingHeaders(dryRun bool) ([]domain.FixResult, error) {
	audits, _, err := s.store.ListAuditsWithFindings()
	if err != nil {
		return nil, err
	}
	now := s.now()
	var out []domain.FixResult
	for _, a := range audits {
		if len(a.NearMisses) == 0 {
			continue
		}
		slug := a.Audit.Slug
		changes, err := retryOnConflict(s, dryRun, func() ([]string, error) {
			var applied []string
			_, _, changed, err := s.store.TransformAuditBody(slug, now, dryRun, func(_ domain.Audit, current string) (string, error) {
				fixed, hits := domain.CanonicalizeFindingHeaders(current)
				applied = applied[:0]
				for _, h := range hits {
					applied = append(applied, fmt.Sprintf("line %d: %q → %q", h.Line, h.Text, h.Canonical))
				}
				return fixed, nil
			})
			if err != nil || !changed {
				return nil, err
			}
			return applied, nil
		})
		if err != nil {
			// Report the prefix that already landed, then surface the failure — the
			// same partial-progress contract the frontmatter fixer keeps.
			return out, fmt.Errorf("canonicalize finding headers in %s: %w", slug, err)
		}
		if len(changes) > 0 {
			out = append(out, domain.FixResult{Path: a.Audit.Path, Changes: changes})
		}
	}
	return out, nil
}

// LintAudits validates findings, managed candidate projections, and the bucket↔state
// invariant, returning one LintResult per audit with issues. slug restricts it to one audit.
func (s *Service) LintAudits(slug string) ([]LintResult, []LintLoadProblem, error) {
	var (
		results  []LintResult
		problems []LintLoadProblem
	)
	if isNilCapability(s.auditReads) {
		return nil, nil, fmt.Errorf("audit snapshot reads are unavailable from this service")
	}
	snapshot, err := s.auditReads.ReadAuditSnapshot(slug)
	if err != nil {
		return nil, nil, err
	}
	problems = snapshot.Problems
	for _, record := range snapshot.Audits {
		iss := AuditLintIssues(record.Audit, record.Findings, record.NearMisses, record.CandidateIssues)
		if len(iss) > 0 {
			results = append(results, LintResult{Slug: record.Audit.Slug, Issues: iss})
		}
	}
	return results, problems, nil
}

func findingMatches(fd domain.Finding, f FindingFilter) bool {
	if len(f.Status) > 0 && !anyEqualFold(f.Status, fd.Status) {
		return false
	}
	if len(f.Effort) > 0 && !anyEqualFold(f.Effort, fd.Effort) {
		return false
	}
	if len(f.Urgency) > 0 && !anyEqualFold(f.Urgency, fd.Urgency) {
		return false
	}
	if f.Component != "" && !strings.Contains(strings.ToLower(fd.Component), strings.ToLower(strings.TrimSpace(f.Component))) {
		return false
	}
	return true
}

func anyEqualFold(opts []string, v string) bool {
	for _, o := range opts {
		// Skip empty/whitespace tokens — a stray comma (`--status "open,"` →
		// ["open",""]) must NOT match findings with a missing field (where v == "").
		if o = strings.TrimSpace(o); o != "" && strings.EqualFold(o, v) {
			return true
		}
	}
	return false
}
