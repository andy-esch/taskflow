package core

import (
	"errors"
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
func (nopStore) EditBody(string, string, bool, time.Time, bool) (domain.Task, string, error) {
	return domain.Task{}, "", nil
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
func (nopStore) MoveEpic(string, string, bool) (domain.Epic, error) {
	return domain.Epic{}, nil
}
func (nopStore) SetEpicFields(string, map[string]any, bool) (domain.Epic, error) {
	return domain.Epic{}, nil
}
func (nopStore) EditEpic(string, func(string, error) (string, error)) (domain.Epic, bool, error) {
	return domain.Epic{}, false, nil
}
func (nopStore) ListAudits() ([]domain.Audit, []domain.FileProblem, error) { return nil, nil, nil }
func (nopStore) ListAuditsWithFindings() ([]AuditWithFindings, []domain.FileProblem, error) {
	return nil, nil, nil
}
func (nopStore) GetAudit(string) (domain.Audit, string, error) {
	return domain.Audit{}, "", domain.ErrNotFound
}
func (nopStore) GetAuditByPath(string) (domain.Audit, string, error) {
	return domain.Audit{}, "", domain.ErrNotFound
}
func (nopStore) MoveAudit(string, domain.AuditBucket, bool) (domain.Audit, error) {
	return domain.Audit{}, nil
}
func (nopStore) CreateAudit(domain.Audit, string, bool) (domain.Audit, error) {
	return domain.Audit{}, nil
}

// fakeStore is an in-memory Store for pure core unit tests; it overrides only
// the read/create methods its tests touch (the rest come from nopStore).
type fakeStore struct {
	nopStore
	tasks             []domain.Task
	epics             []domain.Epic
	audits            []domain.Audit
	problems          []domain.FileProblem // returned by ListTasks
	created           []domain.Task        // tasks passed to CreateTask
	createdBodies     []string             // bodies passed to CreateTask (parallel to created)
	createdAudits     []domain.Audit       // audits passed to CreateAudit
	auditCreateBodies []string             // bodies passed to CreateAudit (parallel to createdAudits)
	epicCreateBodies  []string             // bodies passed to CreateEpic
	auditBodies       map[string]string    // slug → body, for GetAudit (finding queries)
}

func (f *fakeStore) GetAudit(slug string) (domain.Audit, string, error) {
	for _, a := range f.audits {
		if a.Slug == slug {
			return a, f.auditBodies[slug], nil
		}
	}
	return domain.Audit{}, "", domain.ErrNotFound
}

// GetAuditByPath mirrors the real store's read-by-path: find the seeded audit
// whose .Path matches and return its body (auditBodies stays slug-keyed, so seed
// audits set .Path — typically to the slug — for the two keys to coincide).
func (f *fakeStore) GetAuditByPath(path string) (domain.Audit, string, error) {
	for _, a := range f.audits {
		if a.Path == path {
			return a, f.auditBodies[a.Slug], nil
		}
	}
	return domain.Audit{}, "", domain.ErrNotFound
}

func (f *fakeStore) ListTasks() ([]domain.Task, []domain.FileProblem, error) {
	return f.tasks, f.problems, nil
}
func (f *fakeStore) ListAudits() ([]domain.Audit, []domain.FileProblem, error) {
	return f.audits, nil, nil
}

// ListAuditsWithFindings mirrors the real store: one scan returning each seeded
// audit alongside the findings parsed from its (slug-keyed) body — the single read
// Summary's findings rollup now consumes instead of a GetAuditByPath re-read.
func (f *fakeStore) ListAuditsWithFindings() ([]AuditWithFindings, []domain.FileProblem, error) {
	out := make([]AuditWithFindings, 0, len(f.audits))
	for _, a := range f.audits {
		out = append(out, AuditWithFindings{Audit: a, Findings: domain.ParseFindings(f.auditBodies[a.Slug])})
	}
	return out, nil, nil
}
func (f *fakeStore) GetTask(slug string) (domain.Task, string, error) {
	for _, t := range f.tasks {
		if t.Slug == slug {
			return t, "", nil
		}
	}
	return domain.Task{}, "", domain.ErrNotFound
}
func (f *fakeStore) CreateTask(t domain.Task, body string, _ bool) (domain.Task, error) {
	f.created = append(f.created, t)
	f.createdBodies = append(f.createdBodies, body)
	return t, nil
}
func (f *fakeStore) CreateAudit(a domain.Audit, body string, _ bool) (domain.Audit, error) {
	f.createdAudits = append(f.createdAudits, a)
	f.auditCreateBodies = append(f.auditCreateBodies, body)
	return a, nil
}
func (f *fakeStore) ListEpics() ([]domain.Epic, []domain.FileProblem, error) {
	return f.epics, nil, nil
}
func (f *fakeStore) CreateEpic(slug string, e domain.Epic, body string, _ bool) (domain.Epic, error) {
	e.ID = slug
	f.epicCreateBodies = append(f.epicCreateBodies, body)
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

// MoveEpic mirrors the real store's field rewrite: resolve the seeded epic, set
// its status, and (unless dryRun) persist it back to the in-memory slice.
func (f *fakeStore) MoveEpic(id, status string, dryRun bool) (domain.Epic, error) {
	for i, e := range f.epics {
		if e.ID == id {
			e.Status = status
			if !dryRun {
				f.epics[i] = e
			}
			return e, nil
		}
	}
	return domain.Epic{}, domain.ErrNotFound
}

// SetEpicFields mirrors the real store: resolve the seeded epic and apply the
// already-validated/coerced updates the Service hands down (description/priority/
// tags), persisting unless dryRun. Unknown id is ErrNotFound, like the real store.
func (f *fakeStore) SetEpicFields(id string, updates map[string]any, dryRun bool) (domain.Epic, error) {
	for i, e := range f.epics {
		if e.ID != id {
			continue
		}
		if v, ok := updates["description"].(string); ok {
			e.Description = v
		}
		if v, ok := updates["priority"].(string); ok {
			e.Priority = v
		}
		if v, ok := updates["tags"].([]string); ok {
			e.Tags = v
		}
		if !dryRun {
			f.epics[i] = e
		}
		return e, nil
	}
	return domain.Epic{}, domain.ErrNotFound
}

func TestService_ListEpics_Rollup(t *testing.T) {
	svc := NewService(&fakeStore{
		epics: []domain.Epic{{ID: "e1", Status: "active"}, {ID: "e2"}},
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

// TestService_ListEpics_ExcludesDeprecated: deprecated (withdrawn) tasks leave
// total/done and are counted separately in Deprecated; deferred ("not now")
// stays in total as real pending work.
func TestService_ListEpics_ExcludesDeprecated(t *testing.T) {
	svc := NewService(&fakeStore{
		epics: []domain.Epic{{ID: "e1"}},
		tasks: []domain.Task{
			{Slug: "a", Epic: "e1", Status: domain.StatusCompleted},
			{Slug: "b", Epic: "e1", Status: domain.StatusCompleted},
			{Slug: "c", Epic: "e1", Status: domain.StatusDeferred},   // stays in total
			{Slug: "d", Epic: "e1", Status: domain.StatusDeprecated}, // excluded
			{Slug: "e", Epic: "e1", Status: domain.StatusDeprecated}, // excluded
		},
	})
	s, _, err := svc.ListEpics()
	if err != nil {
		t.Fatal(err)
	}
	e := s[0]
	if e.Total != 3 || e.Done != 2 || e.Deprecated != 2 || e.Percent() != 66 {
		t.Errorf("want total=3 done=2 deprecated=2 pct=66 (2 done + 1 deferred; 2 deprecated out), got %+v pct=%d", e, e.Percent())
	}

	// The epic-18 case: all real work done, one deprecated → 1/1 (100%), not 1/2.
	svc2 := NewService(&fakeStore{
		epics: []domain.Epic{{ID: "x"}},
		tasks: []domain.Task{
			{Slug: "p", Epic: "x", Status: domain.StatusCompleted},
			{Slug: "q", Epic: "x", Status: domain.StatusDeprecated},
		},
	})
	x, _, _ := svc2.ListEpics()
	if x[0].Total != 1 || x[0].Done != 1 || x[0].Percent() != 100 || x[0].Deprecated != 1 {
		t.Errorf("a fully-done epic with one deprecated task should read 1/1 (100%%) + 1 deprecated, got %+v pct=%d", x[0], x[0].Percent())
	}
}

// TestTaskRollup pins the shared counting rule the epic list/status, epic show,
// and TUI epic detail all derive from — including the empty-input zero case.
func TestTaskRollup(t *testing.T) {
	if done, total, dep := TaskRollup(nil); done != 0 || total != 0 || dep != 0 {
		t.Errorf("TaskRollup(nil) = %d,%d,%d, want 0,0,0", done, total, dep)
	}
	tasks := []domain.Task{
		{Status: domain.StatusCompleted},
		{Status: domain.StatusCompleted},
		{Status: domain.StatusInProgress},
		{Status: domain.StatusDeferred},   // stays in total, not done
		{Status: domain.StatusDeprecated}, // excluded from total, counted separately
	}
	if done, total, dep := TaskRollup(tasks); done != 2 || total != 4 || dep != 1 {
		t.Errorf("TaskRollup = done %d total %d deprecated %d, want 2 4 1", done, total, dep)
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

// TestService_NewTask_AppliesDefaults pins L23 (2026-06-22 audit): NewTask defaults
// zero-valued priority/tier/autonomy/effort so a caller that doesn't replicate the
// CLI flag defaults still produces a valid, lint-clean task.
func TestService_NewTask_AppliesDefaults(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1"}}}
	svc := NewService(fs)
	if _, err := svc.NewTask(NewTaskParams{Title: "Defaulted", Epic: "e1", Tags: []string{"x"}}); err != nil {
		t.Fatalf("NewTask with zero-valued fields should default and succeed, got %v", err)
	}
	if len(fs.created) != 1 {
		t.Fatalf("want 1 created task, got %d", len(fs.created))
	}
	got := fs.created[0]
	if got.Priority != "medium" || got.Tier != 3 || got.Autonomy != 3 || got.Effort != "Unknown" {
		t.Errorf("defaults not applied: priority=%q tier=%d autonomy=%d effort=%q",
			got.Priority, got.Tier, got.Autonomy, got.Effort)
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

// TestService_Summary_RevisitDue pins the snooze nudge: only deferred tasks whose
// revisit_at has arrived count toward RevisitDue. Uses a clearly-past date (always
// due) and a clearly-future one (never due) so it stays robust against the wall
// clock. A revisit_at on a non-deferred task is ignored by the count (Move clears
// it on leaving deferred, so such a value only arises via a manual `task set`).
func TestService_Summary_RevisitDue(t *testing.T) {
	svc := NewService(&fakeStore{
		tasks: []domain.Task{
			{Slug: "past-due", Status: domain.StatusDeferred, RevisitAt: "2020-01-01"},        // due
			{Slug: "future", Status: domain.StatusDeferred, RevisitAt: "2099-01-01"},          // not due
			{Slug: "no-date", Status: domain.StatusDeferred},                                  // indefinite, not due
			{Slug: "active-past", Status: domain.StatusReadyToStart, RevisitAt: "2020-01-01"}, // not deferred → ignored
		},
	})
	s, err := svc.Summary()
	if err != nil {
		t.Fatal(err)
	}
	if s.RevisitDue != 1 {
		t.Errorf("RevisitDue = %d, want 1 (only the past-due deferred task)", s.RevisitDue)
	}
}

// TestService_Summary_EpicLastUpdated pins the derived epic timestamp: epics have
// no updated_at of their own, so it's the max updated_at across their tasks
// (falling back to created), and "" for an epic with no tasks.
func TestService_Summary_EpicLastUpdated(t *testing.T) {
	svc := NewService(&fakeStore{
		epics: []domain.Epic{{ID: "e1"}, {ID: "e2"}, {ID: "e3"}},
		tasks: []domain.Task{
			{Slug: "a", Epic: "e1", Updated: "2026-01-10"},
			{Slug: "b", Epic: "e1", Updated: "2026-06-25"}, // newest in e1
			{Slug: "c", Epic: "e2", Created: "2025-03-01"}, // no updated_at → created fallback
		},
	})
	s, err := svc.Summary()
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]string{}
	for _, es := range s.Epics {
		by[es.Epic.ID] = es.LastUpdated
	}
	if by["e1"] != "2026-06-25" {
		t.Errorf("e1 LastUpdated = %q, want the max updated_at 2026-06-25", by["e1"])
	}
	if by["e2"] != "2025-03-01" {
		t.Errorf("e2 LastUpdated = %q, want the created fallback 2025-03-01", by["e2"])
	}
	if by["e3"] != "" {
		t.Errorf("e3 (no tasks) LastUpdated = %q, want empty", by["e3"])
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

// TestService_MoveEpic mirrors the MoveAudit tests: happy-path status change,
// an out-of-vocabulary status → ErrValidation, an unknown id → ErrNotFound.
func TestService_MoveEpic(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1", Status: "active"}}}
	svc := NewService(fs)

	e, err := svc.MoveEpic("e1", "retired", false)
	if err != nil {
		t.Fatalf("active→retired should succeed, got %v", err)
	}
	if e.Status != "retired" {
		t.Errorf("returned epic status = %q, want retired", e.Status)
	}
	if fs.epics[0].Status != "retired" {
		t.Errorf("store not updated: %q", fs.epics[0].Status)
	}
}

func TestService_MoveEpic_InvalidStatus(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1", Status: "active"}}}
	_, err := NewService(fs).MoveEpic("e1", "bogus", false)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("invalid status should be ErrValidation, got %v", err)
	}
	if fs.epics[0].Status != "active" {
		t.Errorf("nothing should change on a validation failure, got %q", fs.epics[0].Status)
	}
}

func TestService_MoveEpic_NotFound(t *testing.T) {
	_, err := NewService(&fakeStore{epics: []domain.Epic{{ID: "e1"}}}).MoveEpic("ghost", "retired", false)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown epic should be ErrNotFound, got %v", err)
	}
}

// TestService_SetEpicFields mirrors the SetFields tests: a valid field set lands,
// a bad priority is ErrValidation (nothing written), an unknown id is ErrNotFound.
func TestService_SetEpicFields(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1", Status: "active", Priority: "medium"}}}
	svc := NewService(fs)

	e, err := svc.SetEpicFields("e1", map[string]any{"priority": "high"}, false, false)
	if err != nil {
		t.Fatalf("setting a valid priority should succeed, got %v", err)
	}
	if e.Priority != "high" {
		t.Errorf("returned epic priority = %q, want high", e.Priority)
	}
	if fs.epics[0].Priority != "high" {
		t.Errorf("store not updated: %q", fs.epics[0].Priority)
	}
}

func TestService_SetEpicFields_BadPriority(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1", Status: "active", Priority: "medium"}}}
	_, err := NewService(fs).SetEpicFields("e1", map[string]any{"priority": "urgent"}, false, false)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("an invalid priority should be ErrValidation, got %v", err)
	}
	if fs.epics[0].Priority != "medium" {
		t.Errorf("nothing should change on a validation failure, got %q", fs.epics[0].Priority)
	}
}

func TestService_SetEpicFields_UnknownEpic(t *testing.T) {
	_, err := NewService(&fakeStore{epics: []domain.Epic{{ID: "e1"}}}).
		SetEpicFields("ghost", map[string]any{"priority": "high"}, false, false)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown epic should be ErrNotFound, got %v", err)
	}
}

// TestService_SetEpicFields_StatusRejected pins that status is NOT settable here —
// it moves via `epic move`. Both the typed-flag-less escape hatch and a bare status
// key are rejected before reaching the store.
func TestService_SetEpicFields_StatusRejected(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1", Status: "active"}}}
	_, err := NewService(fs).SetEpicFields("e1", map[string]any{"status": "retired"}, false, false)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("setting status via `set` should be ErrValidation (use `epic move`), got %v", err)
	}
	if fs.epics[0].Status != "active" {
		t.Errorf("status must not change via set, got %q", fs.epics[0].Status)
	}
}

// TestService_SetEpicFields_UnknownFieldNeedsForce mirrors the task contract: a key
// outside the epic registry is rejected without --force, accepted with it.
func TestService_SetEpicFields_UnknownFieldNeedsForce(t *testing.T) {
	svc := NewService(&fakeStore{epics: []domain.Epic{{ID: "e1", Status: "active"}}})
	if _, err := svc.SetEpicFields("e1", map[string]any{"owner": "me"}, false, false); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("an unknown field without --force should be ErrValidation, got %v", err)
	}
	if _, err := svc.SetEpicFields("e1", map[string]any{"owner": "me"}, true, false); err != nil {
		t.Errorf("--force should allow an unknown field, got %v", err)
	}
}

func TestService_SetEpicFields_NoFields(t *testing.T) {
	_, err := NewService(&fakeStore{epics: []domain.Epic{{ID: "e1"}}}).
		SetEpicFields("e1", map[string]any{}, false, false)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("no fields should be ErrValidation, got %v", err)
	}
}

// EditEpic delegates to the store's editor loop; the in-memory fakeStore inherits
// the nopStore EditEpic (returns a zero epic, unchanged), so this just pins the
// pass-through wiring exists and reports "no change" cleanly.
func TestService_EditEpic_PassThrough(t *testing.T) {
	svc := NewService(&fakeStore{epics: []domain.Epic{{ID: "e1"}}})
	_, changed, err := svc.EditEpic("e1", func(cur string, _ error) (string, error) { return cur, nil })
	if err != nil {
		t.Fatalf("EditEpic pass-through should not error, got %v", err)
	}
	if changed {
		t.Errorf("nopStore EditEpic reports no change; got changed=true")
	}
}
