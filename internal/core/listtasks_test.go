package core

import (
	"errors"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// failingEpicStore proves ListTasks doesn't touch ListEpics unless the epic
// filter is set (the validation must not tax the common path).
type failingEpicStore struct{ fakeStore }

func (f *failingEpicStore) ListEpics() ([]domain.Epic, []domain.FileProblem, error) {
	return nil, nil, errors.New("ListEpics must not be called without an epic filter")
}

func TestService_ListTasks_RejectsInvalidFilters(t *testing.T) {
	svc := NewService(&fakeStore{
		epics: []domain.Epic{{ID: "e1"}},
		tasks: []domain.Task{{Slug: "a", Epic: "e1", Status: domain.StatusInProgress}},
	})

	// A typo'd status is ErrValidation, not a silently empty list.
	if _, _, err := svc.ListTasks(TaskFilter{Status: "bogus"}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("invalid status filter should be ErrValidation, got %v", err)
	}
	// An unknown epic likewise.
	if _, _, err := svc.ListTasks(TaskFilter{Epic: "nope"}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("unknown epic filter should be ErrValidation, got %v", err)
	}
	// Valid filters still work.
	tasks, _, err := svc.ListTasks(TaskFilter{Status: "in-progress", Epic: "e1"})
	if err != nil || len(tasks) != 1 {
		t.Errorf("valid filters should pass: %v (%d tasks)", err, len(tasks))
	}
}

func TestService_ListTasks_EmptyEpicFilterSkipsEpicScan(t *testing.T) {
	svc := NewService(&failingEpicStore{fakeStore{
		tasks: []domain.Task{{Slug: "a", Status: domain.StatusInProgress}},
	}})
	if _, _, err := svc.ListTasks(TaskFilter{}); err != nil {
		t.Errorf("no epic filter must not consult ListEpics: %v", err)
	}
}
