package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

// deferStore drives DeferTask's single guarded lifecycle call in isolation. It
// records the plan and clock handed to the capability and can fail to prove the
// error propagates with its sentinel intact.
type deferStore struct {
	nopStore
	lifecycleCalls int
	lastPlan       TaskLifecyclePlan
	dryRun         bool
	lifecycleNow   time.Time
	lifecycleErr   error
}

type committedFailureStore struct {
	nopStore
	calls int
	cause error
}

func (s *committedFailureStore) ListEpics() ([]domain.Epic, []domain.FileProblem, error) {
	return []domain.Epic{{ID: "01-test", Status: "active"}}, nil, nil
}

func (s *committedFailureStore) MutateTaskLifecycle(_ time.Time, dryRun bool, planner TaskLifecyclePlanner) (TaskLifecycleMutationResult, error) {
	s.calls++
	existing := domain.Task{
		ID: "6g0000000001", Slug: "existing", Status: domain.StatusReadyToStart,
		Description: "existing", Tags: []string{"test"},
	}
	graph := NewTaskGraph([]domain.Task{existing}, nil)
	plan, err := planner(graph)
	if err != nil {
		return TaskLifecycleMutationResult{}, err
	}
	task := existing
	from := existing.Status
	before := graph.State(existing.ID)
	if plan.Create != nil {
		task = plan.Create.Task
		from = task.Status
		prospective := taskGraphWithTask(graph, task)
		before = prospective.State(task.ID)
	}
	task.Status = plan.To
	after := taskGraphWithTask(graph, task).State(task.ID)
	return TaskLifecycleMutationResult{
		Plan: plan, Task: task, From: from, Before: before, After: after,
		Changed: true, DryRun: dryRun, Committed: true,
	}, s.cause
}

func (s *deferStore) MutateTaskLifecycle(now time.Time, dryRun bool, planner TaskLifecyclePlanner) (TaskLifecycleMutationResult, error) {
	s.lifecycleCalls++
	s.dryRun = dryRun
	s.lifecycleNow = now
	if s.lifecycleErr != nil {
		return TaskLifecycleMutationResult{}, s.lifecycleErr
	}
	const taskID = "6g0000000001"
	graph := NewTaskGraph([]domain.Task{
		{ID: taskID, Slug: "alpha", Status: domain.StatusInProgress},
		{ID: "6g0000000002", Slug: "x", Status: domain.StatusInProgress},
	}, nil)
	plan, err := planner(graph)
	if err != nil {
		return TaskLifecycleMutationResult{}, err
	}
	s.lastPlan = plan
	task, _ := graph.Task(plan.TaskID)
	from := task.Status
	task.Status = plan.To
	task.RevisitAt = plan.RevisitAt
	return TaskLifecycleMutationResult{
		Plan: plan, Task: task, From: from, Before: graph.State(plan.TaskID),
		After:   taskGraphWithStatus(graph, plan.TaskID, plan.To).State(plan.TaskID),
		Changed: true, DryRun: dryRun,
	}, nil
}

// TestDeferTask_AtomicSingleWrite pins the audit-M4 fix: a `defer --until` is ONE
// store.Defer call carrying the date — not the old Move-then-SetFields two-write
// path that could leave a task deferred without its revisit_at.
func TestDeferTask_AtomicSingleWrite(t *testing.T) {
	st := &deferStore{}
	svc := NewService(st)

	got, err := svc.DeferTask("alpha", "2026-09-01", false)
	if err != nil {
		t.Fatalf("DeferTask: %v", err)
	}
	if st.lifecycleCalls != 1 {
		t.Errorf("want exactly one guarded lifecycle call, got %d", st.lifecycleCalls)
	}
	if st.lastPlan.TaskID != "6g0000000001" || st.lastPlan.RevisitAt != "2026-09-01" || st.dryRun {
		t.Errorf("Defer plan = (%q, %q, dryRun=%v), want (6g0000000001, 2026-09-01, false)", st.lastPlan.TaskID, st.lastPlan.RevisitAt, st.dryRun)
	}
	if got.Task.RevisitAt != "2026-09-01" || got.Task.Status != domain.StatusDeferred {
		t.Errorf("result = (status %q, revisit %q), want (deferred, 2026-09-01)", got.Task.Status, got.Task.RevisitAt)
	}
}

func TestLifecycleCommittedFailureIsNeverRetriedAndRetainsReceipt(t *testing.T) {
	for _, cause := range []error{
		errors.New("unlock failed"),
		fmt.Errorf("unlock failed: %w", domain.ErrConflict),
	} {
		for _, create := range []bool{false, true} {
			name := "move"
			if create {
				name = "new-start"
			}
			t.Run(fmt.Sprintf("%s/%v", name, domain.Classify(cause)), func(t *testing.T) {
				st := &committedFailureStore{cause: cause}
				svc := NewService(st, WithRetry(4, func(int) {}), WithIDGen(func() string { return "6g0000000002" }))

				var receipt TaskLifecycleReceipt
				var err error
				if create {
					var task domain.Task
					task, err = svc.NewTask(NewTaskParams{
						Title: "Created", Epic: "01-test", Description: "created",
						Tags: []string{"test"}, Start: true, Body: "# Created\n",
					})
					receipt.Task = task
				} else {
					receipt, err = svc.Move("existing", domain.StatusInProgress, false, TaskLifecycleOverrideNone)
				}
				var committed *TaskLifecycleMutationFailure
				if !errors.As(err, &committed) || !committed.Receipt.Committed ||
					committed.Receipt.Task.Status != domain.StatusInProgress {
					t.Fatalf("result receipt=%+v err=%v", receipt, err)
				}
				if st.calls != 1 {
					t.Fatalf("committed failure retried %d times", st.calls)
				}
				if !strings.Contains(err.Error(), "committed") || !strings.Contains(err.Error(), "inspect current task state") {
					t.Fatalf("committed failure is not actionable: %v", err)
				}
			})
		}
	}
}

// TestDeferTask_PropagatesStoreError pins that a NON-conflict store.Defer failure surfaces
// with its sentinel intact and is NOT retried — the write is atomic, so a failure means
// nothing changed, and the CLI still maps the sentinel to its exit code. (ErrConflict is the
// one error that IS retried; that path is covered by TestRetry_ExhaustionSurfacesConflict.)
func TestDeferTask_PropagatesStoreError(t *testing.T) {
	st := &deferStore{lifecycleErr: fmt.Errorf("%w: bad frontmatter", domain.ErrValidation)}
	svc := NewService(st)

	_, err := svc.DeferTask("alpha", "2026-09-01", false)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("error should keep its sentinel, got %v", err)
	}
	if st.lifecycleCalls != 1 {
		t.Errorf("a non-conflict error must not be retried; want exactly one lifecycle call, got %d", st.lifecycleCalls)
	}
}

// TestDeferTask_ValidatesDate pins that a malformed --until fails up front
// (ErrValidation) and never reaches the store — the guard the old SetFields path
// gave for free, kept now that the atomic write bypasses SetFields.
func TestDeferTask_ValidatesDate(t *testing.T) {
	st := &deferStore{}
	svc := NewService(st)

	_, err := svc.DeferTask("alpha", "next-week", false)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("a bad --until should be ErrValidation, got %v", err)
	}
	if st.lifecycleCalls != 0 {
		t.Errorf("a bad date must not reach the store, got %d lifecycle calls", st.lifecycleCalls)
	}
}

// TestDeferTask_DryRun pins that --dry-run reaches the store as a preview (no
// write) and still reflects the would-be revisit_at on the returned task.
func TestDeferTask_DryRun(t *testing.T) {
	st := &deferStore{}
	svc := NewService(st)

	got, err := svc.DeferTask("alpha", "2026-09-01", true)
	if err != nil {
		t.Fatalf("dry-run DeferTask: %v", err)
	}
	if st.lifecycleCalls != 1 || !st.dryRun {
		t.Errorf("dry-run should reach lifecycle mutation with dryRun=true, got calls=%d dryRun=%v", st.lifecycleCalls, st.dryRun)
	}
	if got.Task.RevisitAt != "2026-09-01" {
		t.Errorf("dry-run preview should carry the would-be revisit_at, got %q", got.Task.RevisitAt)
	}
}

// TestDeferTask_BareDefer pins that a defer with no date is a plain move to
// deferred — store.Defer with an empty until, no revisit_at.
func TestDeferTask_BareDefer(t *testing.T) {
	st := &deferStore{}
	svc := NewService(st)

	if _, err := svc.DeferTask("alpha", "", false); err != nil {
		t.Fatalf("bare DeferTask: %v", err)
	}
	if st.lifecycleCalls != 1 || st.lastPlan.RevisitAt != "" {
		t.Errorf("bare defer should use one lifecycle plan with empty revisit_at, got calls=%d until=%q", st.lifecycleCalls, st.lastPlan.RevisitAt)
	}
}

// TestListTasks_RevisitDue pins the `task list --revisit-due` predicate: only
// deferred tasks whose revisit_at is on or before the (injected) clock day are
// returned — today counts as due, future/no-date don't, and a revisit_at on a
// non-deferred task is ignored. It composes with the other filters.
func TestListTasks_RevisitDue(t *testing.T) {
	now := func() time.Time { return time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC) }
	svc := NewService(&fakeStore{
		tasks: []domain.Task{
			{Slug: "due-past", Status: domain.StatusDeferred, RevisitAt: "2020-01-01", Tags: []string{"net"}},
			{Slug: "due-today", Status: domain.StatusDeferred, RevisitAt: "2026-06-26", Tags: []string{"ui"}},
			{Slug: "future", Status: domain.StatusDeferred, RevisitAt: "2099-01-01"},
			{Slug: "no-date", Status: domain.StatusDeferred},
			{Slug: "active", Status: domain.StatusReadyToStart, RevisitAt: "2020-01-01"}, // not deferred → excluded
		},
	}, WithClock(now))

	// --revisit-due alone: the two deferred tasks whose snooze date has arrived
	// (today is due), nothing else — and it bypasses the active-only default.
	got, _, err := svc.ListTasks(TaskFilter{RevisitDue: true})
	if err != nil {
		t.Fatal(err)
	}
	if set := slugSet(got); len(set) != 2 || !set["due-past"] || !set["due-today"] {
		t.Errorf("--revisit-due = %v, want {due-past, due-today}", set)
	}

	// Composes with --tag.
	got, _, err = svc.ListTasks(TaskFilter{RevisitDue: true, Tag: "ui"})
	if err != nil {
		t.Fatal(err)
	}
	if set := slugSet(got); len(set) != 1 || !set["due-today"] {
		t.Errorf("--revisit-due --tag ui = %v, want {due-today}", set)
	}
}

func slugSet(tasks []domain.Task) map[string]bool {
	m := map[string]bool{}
	for _, t := range tasks {
		m[t.Slug] = true
	}
	return m
}

func TestNewTask_MintsValidID(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1", Status: "active"}}}
	svc := NewService(fs)
	got, err := svc.NewTask(NewTaskParams{Title: "Add retry", Epic: "e1", Description: "d", Tags: []string{"net"}, Body: "# x\n"})
	if err != nil {
		t.Fatal(err)
	}
	if !id.Valid(got.ID) {
		t.Errorf("NewTask minted an invalid id: %q", got.ID)
	}
	// The id must reach CreateTask (be persisted), not just the returned value.
	if len(fs.created) != 1 || fs.created[0].ID != got.ID {
		t.Errorf("id not passed to CreateTask: created=%+v", fs.created)
	}
}

func TestNewAudit_MintsValidID(t *testing.T) {
	fs := &fakeStore{}
	svc := NewService(fs)
	got, err := svc.NewAudit(NewAuditParams{Area: "storage", Date: "2026-07-02", Body: "# x\n"})
	if err != nil {
		t.Fatal(err)
	}
	if !id.Valid(got.ID) {
		t.Errorf("NewAudit minted an invalid id: %q", got.ID)
	}
	if len(fs.createdAudits) != 1 || fs.createdAudits[0].ID != got.ID {
		t.Errorf("id not passed to CreateAudit: created=%+v", fs.createdAudits)
	}
}

func TestNewTask_UsesInjectedIDGen(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1"}}}
	svc := NewService(fs, WithIDGen(func() string { return "0000000000zz" }))
	got, err := svc.NewTask(NewTaskParams{Title: "x", Epic: "e1", Tags: []string{"a"}, Body: "b"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "0000000000zz" {
		t.Errorf("NewTask ignored the injected id gen: got %q", got.ID)
	}
}

// TestWithClock_GovernsWriteStamps pins the clock unification: an injected clock
// drives the time handed to write paths (here store.Defer's stamp time), not just
// the revisit read paths — so WithClock makes date stamping deterministic too.
func TestWithClock_GovernsWriteStamps(t *testing.T) {
	fixed := time.Date(2031, 7, 8, 9, 0, 0, 0, time.UTC)
	st := &deferStore{}
	svc := NewService(st, WithClock(func() time.Time { return fixed }))

	if _, err := svc.DeferTask("x", "2031-09-01", false); err != nil {
		t.Fatalf("DeferTask: %v", err)
	}
	if !st.lifecycleNow.Equal(fixed) {
		t.Errorf("lifecycle mutation should stamp via the injected clock; got %v, want %v", st.lifecycleNow, fixed)
	}
	if !svc.Now().Equal(fixed) {
		t.Errorf("Service.Now() should expose the injected clock; got %v", svc.Now())
	}
}
