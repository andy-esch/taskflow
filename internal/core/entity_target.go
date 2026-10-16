package core

// These per-entity wrappers share the ordinary conflict budget between parse-free
// initial reference selection and exact selected writes. Failed selection binds
// nothing; once bound, even an empty error result cannot restart alias matching.
// This is not a transaction framework or authority to reuse a target with another
// Service. See docs/architecture/planning-data.md#identity-is-evidence-not-a-display-label.
func retryAuditMutation[T any](s *Service, ref string, dryRun bool, mutate func(AuditMutationTarget) (T, error)) (T, error) {
	var target AuditMutationTarget
	return retryOnConflict(s, dryRun, func() (T, error) {
		if target.ID == "" {
			source, err := s.store.ResolveAuditSource(ref)
			if err == nil {
				err = requireSourceID(EntityAudit, source)
			}
			if err != nil {
				var zero T
				return zero, err
			}
			target.ID = source.ID
		}
		return mutate(target)
	})
}

func retryResearchMutation[T any](s *Service, ref string, dryRun bool, mutate func(ResearchMutationTarget) (T, error)) (T, error) {
	var target ResearchMutationTarget
	return retryOnConflict(s, dryRun, func() (T, error) {
		if target.ID == "" {
			source, err := s.store.ResolveResearchSource(ref)
			if err == nil {
				err = requireSourceID(EntityResearch, source)
			}
			if err != nil {
				var zero T
				return zero, err
			}
			target.ID = source.ID
		}
		return mutate(target)
	})
}

func retryEpicMutation[T any](s *Service, ref string, dryRun bool, mutate func(EpicMutationTarget) (T, error)) (T, error) {
	var target EpicMutationTarget
	return retryOnConflict(s, dryRun, func() (T, error) {
		if target.ID == "" {
			source, err := s.store.ResolveEpicSource(ref)
			if err == nil {
				err = requireSourceID(EntityEpic, source)
			}
			if err != nil {
				var zero T
				return zero, err
			}
			target.ID = source.ID
		}
		return mutate(target)
	})
}
