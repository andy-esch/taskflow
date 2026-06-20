package core

import (
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

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
		_, body, err := s.store.GetAudit(a.Slug)
		if err != nil {
			problems = append(problems, domain.FileProblem{Path: a.Slug, Message: err.Error()})
			continue
		}
		collect(a.Slug, string(a.Bucket), body)
	}
	return out, problems, nil
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
		if strings.EqualFold(strings.TrimSpace(o), v) {
			return true
		}
	}
	return false
}
