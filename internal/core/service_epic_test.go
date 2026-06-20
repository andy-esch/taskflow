package core

import (
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// nopStore is a no-op core.Store: every method returns zero values. Tests embed
// it and override only the handful they exercise — so adding a Store method
// updates exactly one place, not every fake. The assertion below makes that a
// build error if nopStore ever falls out of sync with the port.
type nopStore struct{}

var _ Store = nopStore{}

func (nopStore) ListTasks() ([]domain.Task, []domain.FileProblem, error) { return nil, nil, nil }
func (nopStore) GetTask(string) (domain.Task, string, error) {
	return domain.Task{}, "", domain.ErrNotFound
}
func (nopStore) Move(string, domain.Status, time.Time, bool) (domain.Task, error) {
	return domain.Task{}, nil
}
func (nopStore) SetFields(string, map[string]any, bool) (domain.Task, error) {
	return domain.Task{}, nil
}
func (nopStore) CreateTask(domain.Task, string, bool) (domain.Task, error) { return domain.Task{}, nil }
func (nopStore) EditTask(string, func(string, error) (string, error)) (domain.Task, bool, error) {
	return domain.Task{}, false, nil
}
func (nopStore) EditBody(string, string, bool, time.Time, bool) (domain.Task, error) {
	return domain.Task{}, nil
}
func (nopStore) ListEpics() ([]domain.Epic, []domain.FileProblem, error) {
	return nil, nil, nil
}
func (nopStore) GetEpic(string) (domain.Epic, string, error) {
	return domain.Epic{}, "", domain.ErrNotFound
}
func (nopStore) CreateEpic(string, domain.Epic, string, bool) (domain.Epic, error) {
	return domain.Epic{}, nil
}
func (nopStore) ListAudits() ([]domain.Audit, []domain.FileProblem, error) { return nil, nil, nil }
func (nopStore) GetAudit(string) (domain.Audit, string, error) {
	return domain.Audit{}, "", domain.ErrNotFound
}
func (nopStore) MoveAudit(string, domain.AuditBucket, bool) (domain.Audit, error) {
	return domain.Audit{}, nil
}
func (nopStore) CreateAudit(domain.Audit, string, bool) (domain.Audit, error) {
	return domain.Audit{}, nil
}
func (nopStore) FixFrontmatter(bool) ([]domain.FixResult, error) { return nil, nil }
func (nopStore) WatchPaths() []string                            { return nil }

// fakeStore is an in-memory Store for pure core unit tests; it overrides only
// the read/create methods its tests touch (the rest come from nopStore).
type fakeStore struct {
	nopStore
	tasks         []domain.Task
	epics         []domain.Epic
	audits        []domain.Audit
	problems      []domain.FileProblem // returned by ListTasks
	created       []domain.Task        // tasks passed to CreateTask
	createdAudits []domain.Audit       // audits passed to CreateAudit
}

func (f *fakeStore) ListTasks() ([]domain.Task, []domain.FileProblem, error) {
	return f.tasks, f.problems, nil
}
func (f *fakeStore) ListAudits() ([]domain.Audit, []domain.FileProblem, error) {
	return f.audits, nil, nil
}
func (f *fakeStore) GetTask(slug string) (domain.Task, string, error) {
	for _, t := range f.tasks {
		if t.Slug == slug {
			return t, "", nil
		}
	}
	return domain.Task{}, "", domain.ErrNotFound
}
func (f *fakeStore) CreateTask(t domain.Task, _ string, _ bool) (domain.Task, error) {
	f.created = append(f.created, t)
	return t, nil
}
func (f *fakeStore) CreateAudit(a domain.Audit, _ string, _ bool) (domain.Audit, error) {
	f.createdAudits = append(f.createdAudits, a)
	return a, nil
}
func (f *fakeStore) ListEpics() ([]domain.Epic, []domain.FileProblem, error) {
	return f.epics, nil, nil
}
func (f *fakeStore) CreateEpic(slug string, e domain.Epic, _ string, _ bool) (domain.Epic, error) {
	e.ID = slug
	return e, nil
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

func TestService_NewTask_UnknownEpic(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1"}}}
	svc := NewService(fs)
	_, err := svc.NewTask(NewTaskParams{Title: "X", Epic: "nope", Tier: 3, Autonomy: 3, Priority: "medium"})
	if err == nil {
		t.Fatal("expected error for unknown epic")
	}
	if len(fs.created) != 0 {
		t.Errorf("nothing should be created on validation failure, got %d", len(fs.created))
	}
}

func TestService_NewTask_Valid(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1"}}}
	svc := NewService(fs)
	tk, err := svc.NewTask(NewTaskParams{Title: "My New Task", Epic: "e1", Tier: 3, Autonomy: 3, Priority: "medium", Effort: "Unknown", Tags: []string{"go"}})
	if err != nil {
		t.Fatal(err)
	}
	if tk.Slug != "my-new-task" || tk.Status != domain.StatusReadyToStart || tk.Created == "" {
		t.Errorf("unexpected created task: %+v", tk)
	}
	if len(fs.created) != 1 {
		t.Errorf("expected one CreateTask call, got %d", len(fs.created))
	}
}

func TestService_NewTask_Next(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1"}}}
	svc := NewService(fs)
	tk, err := svc.NewTask(NewTaskParams{Title: "T", Epic: "e1", Tier: 3, Autonomy: 3, Priority: "medium", Tags: []string{"go"}, Description: "do it", Next: true})
	if err != nil {
		t.Fatal(err)
	}
	if tk.Status != domain.StatusNextUp {
		t.Errorf("--next should yield next-up, got %s", tk.Status)
	}
}

func TestService_Summary(t *testing.T) {
	svc := NewService(&fakeStore{
		epics: []domain.Epic{{ID: "e1"}},
		tasks: []domain.Task{
			{Slug: "a", Status: domain.StatusInProgress, Epic: "e1"},
			{Slug: "b", Status: domain.StatusReadyToStart, Epic: "e1"},
			{Slug: "c", Status: domain.StatusCompleted, Epic: "e1"},
			{Slug: "d", Status: domain.StatusCompleted, Declared: domain.StatusReadyToStart}, // misfiled, no epic
		},
	})
	s, err := svc.Summary()
	if err != nil {
		t.Fatal(err)
	}
	counts := map[domain.Status]int{}
	for _, c := range s.Counts {
		counts[c.Status] = c.Count
	}
	if counts[domain.StatusInProgress] != 1 || counts[domain.StatusReadyToStart] != 1 || counts[domain.StatusCompleted] != 2 {
		t.Errorf("counts wrong: %+v", counts)
	}
	if len(s.InProgress) != 1 || s.InProgress[0].Slug != "a" {
		t.Errorf("in-progress wrong: %+v", s.InProgress)
	}
	if s.Misfiled != 1 {
		t.Errorf("misfiled = %d, want 1", s.Misfiled)
	}
	if len(s.Epics) != 1 || s.Epics[0].Total != 3 || s.Epics[0].Done != 1 {
		t.Errorf("epic rollup wrong: %+v", s.Epics)
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
