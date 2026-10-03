package core

import (
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
}

// RepairPlanning performs ordinary frontmatter repairs, then finding-header
// repairs, then (for a real write only) validates the resulting corpus. Ordering
// matters: frontmatter repair may rename a source before body repair resolves it.
// The mutation adapter remains responsible for authorization and durable-prefix
// reporting; a service never retries this multi-file operation wholesale.
func (s *Service) RepairPlanning(dryRun bool) (PlanningRepairResult, error) {
	var result PlanningRepairResult
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
	var err error
	result.Fixes, err = s.frontmatterRepairs.FixFrontmatter(dryRun)
	if err != nil {
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
