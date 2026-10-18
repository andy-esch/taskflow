package core

import (
	"errors"
	"fmt"

	"github.com/andy-esch/taskflow/internal/domain"
)

// PlanningRepairResult retains the completed/proposed prefix even if a later
// repair or validation phase fails. Fix paths are operation evidence, never
// semantic entity fields or authority for another write.
type PlanningRepairResult struct {
	Fixes       []domain.FixResult
	LintResults []LintResult
	Problems    []LoadProblem
	// Committed records at least one completed repair/cleanup or a written
	// relocation destination. Proposals and skipped rows never count. This is
	// historical operation evidence, not proof of current content or existence.
	Committed bool
	// Relocation reports progress inside a failed frontmatter move separately
	// from completed/proposed fixes. It is local adapter evidence, not authority.
	Relocation *FrontmatterRepairRelocation
}

// RecoveryError preserves cause classification and adds ordinary committed-prefix
// guidance, including failures while presenting a successful repair result.
// It neither duplicates an existing recovery error nor replaces relocation's
// more specific reconciliation advice. See docs/architecture/persistence.md.
func (r PlanningRepairResult) RecoveryError(cause error) error {
	if cause == nil || !r.Committed || r.Relocation != nil {
		return cause
	}
	var existing *PlanningRepairFailure
	if errors.As(cause, &existing) {
		return cause
	}
	return &PlanningRepairFailure{Cause: cause}
}

// RepairPlanning performs ordinary frontmatter repairs, then finding-header
// repairs, then (for a real write only) validates the resulting corpus. Ordering
// matters: frontmatter repair may rename a source before body repair resolves it.
// The mutation adapter remains responsible for authorization and durable-prefix
// reporting; a service never retries this multi-file operation wholesale.
func (s *Service) RepairPlanning(dryRun bool) (result PlanningRepairResult, err error) {
	defer func() {
		if !dryRun {
			result.Committed = result.Relocation != nil && result.Relocation.DestinationWritten
			for _, fixed := range result.Fixes {
				if !fixed.Skipped {
					result.Committed = true
					break
				}
			}
		}
		err = result.RecoveryError(err)
	}()
	if s.frontmatterRepairs == nil {
		return result, fmt.Errorf("%w: frontmatter repair is unavailable from this service", domain.ErrValidation)
	}
	// Reject a known incomplete workflow before the first potentially durable
	// edit. Dynamic read/write failures still return the completed prefix below.
	if s.auditReads == nil || s.store == nil {
		return result, fmt.Errorf("%w: finding-header repair capabilities are unavailable from this service", domain.ErrValidation)
	}
	if !dryRun && s.lintReads == nil {
		return result, fmt.Errorf("%w: post-repair lint is unavailable from this service", domain.ErrValidation)
	}
	result.Fixes, err = s.frontmatterRepairs.FixFrontmatter(dryRun)
	if err != nil {
		var failure *FrontmatterRepairFailure
		if errors.As(err, &failure) {
			outcome := failure.Relocation
			result.Relocation = &outcome
		}
		return result, err
	}
	headers, err := s.FixFindingHeaders(dryRun)
	result.Fixes = append(result.Fixes, headers...)
	if err != nil || dryRun {
		return result, err
	}
	result.LintResults, result.Problems, err = s.Lint()
	return result, err
}

// LintWithLinks adds opt-in body-link integrity to ordinary repository lint.
// Missing optional capability is explicit; callers never get a hidden local
// scan merely because another adapter cannot provide link diagnostics.
func (s *Service) LintWithLinks() ([]LintResult, []LoadProblem, error) {
	if s.bodyLinks == nil {
		return nil, nil, fmt.Errorf("%w: body link checks are unavailable from this service", domain.ErrValidation)
	}
	results, problems, err := s.Lint()
	if err != nil {
		return results, problems, err
	}
	links, err := s.bodyLinks.DanglingLinks()
	if err != nil {
		return results, problems, err
	}
	return results, append(problems, links...), nil
}
