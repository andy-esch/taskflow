package core

import (
	"errors"
	"fmt"

	"github.com/andy-esch/taskflow/internal/domain"
)

// ComposeThreadApply compiles a strict authoring manifest into a durable,
// planning-repository-bound plan without mutating repository state.
func (s *Service) ComposeThreadApply(planningRepoID string, manifest ThreadComposeManifest) (ThreadApplyPlan, error) {
	if s.store == nil || s.threads == nil {
		return ThreadApplyPlan{}, fmt.Errorf("thread apply composition is unavailable from this store")
	}
	graph, err := LoadTaskGraph(s.store)
	if err != nil {
		return ThreadApplyPlan{}, err
	}
	threads, problems, err := s.threads.ListThreads()
	if err != nil {
		return ThreadApplyPlan{}, err
	}
	if err := ValidateThreadCreationSource(graph, threads, problems); err != nil {
		return ThreadApplyPlan{}, err
	}
	template, err := s.templateBody("thread", "")
	if err != nil {
		return ThreadApplyPlan{}, err
	}
	body := renderTemplate(template, map[string]string{
		"title": manifest.Thread.Title,
		"goal":  manifest.Thread.Goal,
	})
	return ComposeThreadApplyPlan(ThreadApplySnapshot{
		PlanningRepoID: planningRepoID,
		Graph:          graph,
		Threads:        threads,
	}, manifest, body, s.newID, s.now())
}

// ApplyThreadPlan converges current repository state toward one materialized
// plan. A conflict is retried only before any durable write and only while work
// remains; committed prefixes and complete outcomes are always surfaced.
func (s *Service) ApplyThreadPlan(plan ThreadApplyPlan, dryRun bool) (ThreadApplyReceipt, error) {
	if s.threadApplies == nil {
		return ThreadApplyReceipt{}, fmt.Errorf("thread apply is unavailable from this store")
	}
	now := s.now()
	var receipt ThreadApplyReceipt
	for attempt := 0; ; attempt++ {
		result, err := s.threadApplies.MutateThreadApply(now, dryRun, func(snapshot ThreadApplySnapshot) (ThreadApplyPlan, error) {
			decision, prepareErr := PrepareThreadApply(snapshot, plan)
			if prepareErr != nil {
				return ThreadApplyPlan{}, prepareErr
			}
			return decision.Plan, nil
		})
		receipt = threadApplyReceipt(result)
		// Identity/source failures can occur before the guarded store accepts and
		// echoes the plan. Retain the caller's durable token in the typed recovery
		// receipt so machine errors still identify the attempted Thread.
		if receipt.Plan.Thread.ID == "" {
			receipt.Plan = cloneThreadApplyPlan(plan)
		}
		if dryRun || !errors.Is(err, domain.ErrConflict) || result.Committed || result.Complete || attempt >= s.maxRetries {
			if err != nil {
				return receipt, &ThreadApplyFailure{Cause: err, Receipt: receipt}
			}
			return receipt, nil
		}
		s.retrySleep(attempt + 1)
	}
}

func threadApplyReceipt(result ThreadApplyMutationResult) ThreadApplyReceipt {
	return ThreadApplyReceipt{
		Plan:       cloneThreadApplyPlan(result.Plan),
		Operations: append([]ThreadApplyOperation(nil), result.Operations...),
		Changed:    result.Changed, DryRun: result.DryRun,
		Complete: result.Complete, Committed: result.Committed,
	}
}
