package core

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/domain"
)

// SelectedEpicTaskCreationStore creates a non-started task under an already
// selected epic. It is optional: semantic reads, ordinary creation and local
// paths do not imply this capability. Supply it through composition with the
// same [SourceSetID] as the originating service.
//
// Implementations require exact, unique ownership of target.ID and a readable
// epic, without human-reference matching or declaration/location fallback.
// Revalidate that parent in the same guarded snapshot/transaction as the task
// create, including previews; require task.Epic == target.ID. Reauthorize after
// waiting for a commit guard. Keep authorization,
// no-clobber/ID collision checks and committed/finalization receipts identical
// to ordinary creation. Preview reserves nothing. This is selection safety,
// not an enduring foreign-key constraint or same-ID delete/recreate detection.
// See docs/architecture/planning-data.md#identity-is-evidence-not-a-display-label.
type SelectedEpicTaskCreationStore interface {
	CreateTaskForSelectedEpic(target EpicMutationTarget, task domain.Task, body string, dryRun bool) (TaskCreationReceipt, error)
}

// WithSelectedEpicTaskCreationStore supplies selected-parent creation separately
// from the aggregate Store. No path or aggregate read capability is inferred.
func WithSelectedEpicTaskCreationStore(store SelectedEpicTaskCreationStore) Option {
	return func(s *Service) {
		if !isNilCapability(store) {
			s.selectedTaskCreations = store
		}
	}
}

// NewTaskForSelectedEpic shares [Service.NewTask]'s defaults, validation and
// scaffolding while retaining an epic selected through RecordSource.ID. Epic
// must be blank or exactly target.ID; it is never a second human reference.
// The primary adapter must retain the originating Service/source-set binding.
//
// Start is intentionally unsupported: selected-parent creation cannot borrow
// the separate lifecycle capability's graph authorization. Next remains a
// non-start state. This method makes one create attempt, preserving its exact
// receipt and error; even conflicts/finalization errors never remint or replay.
func (s *Service) NewTaskForSelectedEpic(target EpicMutationTarget, p NewTaskParams) (TaskCreationReceipt, error) {
	if err := requireSourceID(EntityEpic, RecordSource{ID: target.ID}); err != nil {
		return TaskCreationReceipt{}, err
	}
	if p.Epic != "" && p.Epic != target.ID {
		return TaskCreationReceipt{}, fmt.Errorf("%w: epic input must match the selected epic identity", domain.ErrValidation)
	}
	if p.Start {
		return TaskCreationReceipt{}, fmt.Errorf("%w: selected-epic creation does not support create-and-start", domain.ErrValidation)
	}
	if err := validateTaskCreationMode(p); err != nil {
		return TaskCreationReceipt{}, err
	}
	if s.selectedTaskCreations == nil {
		return TaskCreationReceipt{}, fmt.Errorf("%w: selected-epic task creation is unavailable from this service", domain.ErrValidation)
	}
	p.Epic = target.ID
	prepared, err := s.prepareTaskCreation(p)
	if err != nil {
		return TaskCreationReceipt{}, err
	}
	return s.selectedTaskCreations.CreateTaskForSelectedEpic(target, prepared.Task, prepared.Body, p.DryRun)
}
