package core

import (
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
	Audit  string // the audit's slug
	Bucket string // the audit's bucket (open|closed|deferred)
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

// QueryFindings parses findings across every audit (or just f.Audit) and returns
// those matching the filter, in (audit order, document order). Per-file load
// problems are returned separately so a single unreadable audit doesn't sink the
// whole query — the same resilient-read contract as ListTasks/ListAudits.
func (s *Service) QueryFindings(f FindingFilter) ([]AuditFinding, []domain.FileProblem, error) {
	var (
		out      []AuditFinding
		problems []domain.FileProblem
	)
	collect := func(slug, bucket, body string) {
		for _, fd := range domain.ParseFindings(body) {
			if findingMatches(fd, f) {
				out = append(out, AuditFinding{Finding: fd, Audit: slug, Bucket: bucket})
			}
		}
	}

	if f.Audit != "" {
		a, body, err := s.store.GetAudit(f.Audit) // resolves the slug; ErrNotFound/Ambiguous propagate
		if err != nil {
			return nil, nil, err
		}
		collect(a.Slug, string(a.Bucket), body)
		return out, nil, nil
	}

	audits, probs, err := s.store.ListAudits()
	if err != nil {
		return nil, nil, err
	}
	problems = probs
	for _, a := range audits {
		// Read by the path ListAudits already resolved, not GetAudit(a.Slug):
		// re-resolving every slug across all 3 bucket dirs per audit is the O(N^2)
		// sweep M16 flagged, and re-resolving also reopens a concurrent-edit window.
		_, body, err := s.store.GetAuditByPath(a.Path)
		if err != nil {
			problems = append(problems, domain.FileProblem{Path: a.Path, Message: err.Error()})
			continue
		}
		collect(a.Slug, string(a.Bucket), body)
	}
	return out, problems, nil
}

// LintAudits validates audit findings — status vocabulary, missing status, and the
// bucket↔state invariant — returning one LintResult per audit with issues (reusing
// the task-lint result + render shape). slug, if set, restricts to one audit.
func (s *Service) LintAudits(slug string) ([]LintResult, []domain.FileProblem, error) {
	var (
		results  []LintResult
		problems []domain.FileProblem
	)
	check := func(a domain.Audit, body string) {
		iss := domain.LintFindings(string(a.Bucket), domain.ParseFindings(body))
		iss = append(iss, domain.MissingIDIssue(a.ID)...)       // audits get a stable id too
		iss = append(iss, domain.FrontmatterBucketIssues(a)...) // and a missing/foreign bucket flag
		if len(iss) > 0 {
			results = append(results, LintResult{Slug: a.Slug, Issues: iss})
		}
	}
	if slug != "" {
		a, body, err := s.store.GetAudit(slug)
		if err != nil {
			return nil, nil, err
		}
		check(a, body)
		return results, nil, nil
	}
	audits, probs, err := s.store.ListAudits()
	if err != nil {
		return nil, nil, err
	}
	problems = probs
	for _, a := range audits {
		// By path, not GetAudit(a.Slug) — same O(N^2) re-resolve avoidance as
		// QueryFindings (the TUI runs LintAudits on every live reload).
		_, body, err := s.store.GetAuditByPath(a.Path)
		if err != nil {
			problems = append(problems, domain.FileProblem{Path: a.Path, Message: err.Error()})
			continue
		}
		check(a, body)
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
