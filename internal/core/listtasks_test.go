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

// TestService_ListTasks_FilterByEpicNNKey pins the Scheme-2 filter join: `--epic
// 24-data-model` matches a task whose ref is the full stem, a bare NN, or a stale slug —
// all resolve on the NN key, mirroring validation and the rollup (was a raw string compare).
func TestService_ListTasks_FilterByEpicNNKey(t *testing.T) {
	svc := NewService(&fakeStore{
		epics: []domain.Epic{{ID: "24-data-model", Status: "active"}, {ID: "01-other", Status: "active"}},
		tasks: []domain.Task{
			{Slug: "a", Epic: "24-data-model", Status: domain.StatusReadyToStart},
			{Slug: "b", Epic: "24", Status: domain.StatusReadyToStart},          // bare NN
			{Slug: "c", Epic: "24-old-slug", Status: domain.StatusReadyToStart}, // drifted slug
			{Slug: "d", Epic: "01-other", Status: domain.StatusReadyToStart},    // different epic
		},
	})
	got, _, err := svc.ListTasks(TaskFilter{Epic: "24-data-model"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		var slugs []string
		for _, tk := range got {
			slugs = append(slugs, tk.Slug)
		}
		t.Errorf("--epic 24-data-model should match all 3 NN-24 refs (not 01-other), got %v", slugs)
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
