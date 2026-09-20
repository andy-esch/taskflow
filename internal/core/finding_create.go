package core

import (
	"fmt"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// NewFindingParams are semantic inputs for one tool-authored audit finding. Code,
// status, Markdown structure, placement, and optional candidate linkage are owned by
// the domain transform rather than accepted from the adapter.
type NewFindingParams struct {
	Band           string
	Title          string
	File           string
	Component      string
	Effort         string
	Urgency        string
	Body           string
	Recommendation string
	Candidate      *string
	DryRun         bool
}

// FindingCreationReceipt identifies the exact finding planned or persisted. The compact
// receipt deliberately omits the complete audit body; callers that need prose can use the
// ordinary audit read surface after storing this stable audit-local code.
type FindingCreationReceipt struct {
	Audit   domain.Audit
	Finding domain.Finding
	DryRun  bool
}

// NewFinding allocates and writes one canonical finding through the audit body CAS. The
// allocation happens inside the transform closure, so every conflict retry reads current
// codes and cannot replay an identity chosen from a stale snapshot.
func (s *Service) NewFinding(slug string, params NewFindingParams) (FindingCreationReceipt, error) {
	if params.Candidate != nil && strings.TrimSpace(*params.Candidate) == "" {
		return FindingCreationReceipt{}, fmt.Errorf("%w: --candidate for a new finding cannot be empty; omit the flag when no candidate is needed", domain.ErrValidation)
	}
	now := s.now()
	return retryOnConflict(s, params.DryRun, func() (FindingCreationReceipt, error) {
		var created domain.Finding
		audit, _, changed, err := s.store.TransformAuditBody(slug, now, params.DryRun,
			func(audit domain.Audit, current string) (string, error) {
				if audit.Bucket != domain.AuditOpen {
					return "", fmt.Errorf("%w: cannot create an open finding in a %s audit; reopen the audit first", domain.ErrValidation, audit.Bucket)
				}
				next, finding, err := domain.CreateFinding(current, domain.FindingDraft{
					Band: params.Band, Title: params.Title, File: params.File,
					Component: params.Component, Effort: params.Effort, Urgency: params.Urgency,
					Body: params.Body, Recommendation: params.Recommendation,
				})
				if err != nil {
					return "", err
				}
				if params.Candidate != nil {
					next, err = domain.SetFindingCandidate(next, finding.Code, *params.Candidate)
					if err != nil {
						return "", err
					}
				}
				created = finding
				return next, nil
			})
		if err != nil {
			return FindingCreationReceipt{}, err
		}
		if !changed || created.Code == "" {
			return FindingCreationReceipt{}, fmt.Errorf("%w: finding creation produced no change", domain.ErrConflict)
		}
		return FindingCreationReceipt{Audit: audit, Finding: created, DryRun: params.DryRun}, nil
	})
}
