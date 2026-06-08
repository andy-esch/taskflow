package core

import (
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// fakeStore is an in-memory Store for pure core unit tests.
type fakeStore struct {
	tasks []domain.Task
	epics []domain.Epic
}

func (f *fakeStore) ListTasks() ([]domain.Task, []domain.FileProblem, error) {
	return f.tasks, nil, nil
}
func (f *fakeStore) GetTask(slug string) (domain.Task, string, error) {
	for _, t := range f.tasks {
		if t.Slug == slug {
			return t, "", nil
		}
	}
	return domain.Task{}, "", domain.ErrNotFound
}
func (f *fakeStore) Move(string, domain.Status, time.Time) (domain.Task, error) {
	return domain.Task{}, nil
}
func (f *fakeStore) SetFields(string, map[string]any) (domain.Task, error) {
	return domain.Task{}, nil
}
func (f *fakeStore) ListEpics() ([]domain.Epic, []domain.FileProblem, error) {
	return f.epics, nil, nil
}
func (f *fakeStore) FixFrontmatter(bool) ([]domain.FixResult, error) { return nil, nil }
func (f *fakeStore) ListAudits() ([]domain.Audit, []domain.FileProblem, error) {
	return nil, nil, nil
}
func (f *fakeStore) GetAudit(string) (domain.Audit, string, error) {
	return domain.Audit{}, "", domain.ErrNotFound
}
func (f *fakeStore) MoveAudit(string, domain.AuditBucket) (domain.Audit, error) {
	return domain.Audit{}, nil
}
func (f *fakeStore) GetEpic(id string) (domain.Epic, string, error) {
	for _, e := range f.epics {
		if e.ID == id {
			return e, "epic body", nil
		}
	}
	return domain.Epic{}, "", domain.ErrNotFound
}

func TestService_ListEpics_Rollup(t *testing.T) {
	svc := NewService(&fakeStore{
		epics: []domain.Epic{{ID: "e1", Status: "in-progress"}, {ID: "e2"}},
		tasks: []domain.Task{
			{Slug: "a", Epic: "e1", Status: domain.StatusReadyToStart},
			{Slug: "b", Epic: "e1", Status: domain.StatusCompleted},
			{Slug: "c", Epic: "e1", Status: domain.StatusCompleted},
			{Slug: "d", Epic: "other", Status: domain.StatusInProgress}, // unknown epic ignored
		},
	})

	summaries, _, err := svc.ListEpics()
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 2 {
		t.Fatalf("want 2 epics, got %d", len(summaries))
	}
	e1 := summaries[0]
	if e1.Epic.ID != "e1" || e1.Total != 3 || e1.Done != 2 || e1.Percent() != 66 {
		t.Errorf("e1 rollup wrong: %+v pct=%d", e1, e1.Percent())
	}
	if summaries[1].Total != 0 || summaries[1].Percent() != 0 {
		t.Errorf("e2 should be empty: %+v", summaries[1])
	}
}

func TestService_ShowEpic(t *testing.T) {
	svc := NewService(&fakeStore{
		epics: []domain.Epic{{ID: "e1"}},
		tasks: []domain.Task{
			{Slug: "a", Epic: "e1"},
			{Slug: "b", Epic: "other"},
		},
	})
	epic, tasks, body, err := svc.ShowEpic("e1")
	if err != nil {
		t.Fatal(err)
	}
	if epic.ID != "e1" || len(tasks) != 1 || tasks[0].Slug != "a" || body != "epic body" {
		t.Errorf("ShowEpic wrong: %+v tasks=%v body=%q", epic, tasks, body)
	}
}
