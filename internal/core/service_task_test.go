package core

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

// deferStore drives DeferTask's single atomic store.Defer call in isolation: it
// records the call args (so a test can assert the contract) and can be made to
// fail (deferErr) to prove the error propagates with its sentinel intact.
type deferStore struct {
	nopStore
	deferCalls int
	lastSlug   string
	lastUntil  string
	dryRun     bool
	deferNow   time.Time // the clock value Defer was handed (to prove WithClock governs stamps)
	deferErr   error
}

func (s *deferStore) Defer(slug, until string, now time.Time, dryRun bool) (domain.Task, error) {
	s.deferCalls++
	s.lastSlug = slug
	s.lastUntil = until
	s.dryRun = dryRun
	s.deferNow = now
	if s.deferErr != nil {
		return domain.Task{}, s.deferErr
	}
	return domain.Task{Slug: slug, Status: domain.StatusDeferred, RevisitAt: until}, nil
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
	if st.deferCalls != 1 {
		t.Errorf("want exactly one atomic Defer call, got %d", st.deferCalls)
	}
	if st.lastSlug != "alpha" || st.lastUntil != "2026-09-01" || st.dryRun {
		t.Errorf("Defer args = (%q, %q, dryRun=%v), want (alpha, 2026-09-01, false)", st.lastSlug, st.lastUntil, st.dryRun)
	}
	if got.RevisitAt != "2026-09-01" || got.Status != domain.StatusDeferred {
		t.Errorf("result = (status %q, revisit %q), want (deferred, 2026-09-01)", got.Status, got.RevisitAt)
	}
}

// TestDeferTask_PropagatesStoreError pins that a store.Defer failure surfaces with
// its sentinel intact — the write is atomic, so a failure means nothing changed,
// and the CLI still maps the sentinel to its exit code.
func TestDeferTask_PropagatesStoreError(t *testing.T) {
	st := &deferStore{deferErr: fmt.Errorf("%w: changed on disk", domain.ErrConflict)}
	svc := NewService(st)

	_, err := svc.DeferTask("alpha", "2026-09-01", false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("error should keep the ErrConflict sentinel, got %v", err)
	}
	if st.deferCalls != 1 {
		t.Errorf("want exactly one Defer call, got %d", st.deferCalls)
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
	if st.deferCalls != 0 {
		t.Errorf("a bad date must not reach the store, got %d Defer calls", st.deferCalls)
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
	if st.deferCalls != 1 || !st.dryRun {
		t.Errorf("dry-run should reach Defer with dryRun=true, got calls=%d dryRun=%v", st.deferCalls, st.dryRun)
	}
	if got.RevisitAt != "2026-09-01" {
		t.Errorf("dry-run preview should carry the would-be revisit_at, got %q", got.RevisitAt)
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
	if st.deferCalls != 1 || st.lastUntil != "" {
		t.Errorf("bare defer should call Defer once with empty until, got calls=%d until=%q", st.deferCalls, st.lastUntil)
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
	if !st.deferNow.Equal(fixed) {
		t.Errorf("Defer should stamp via the injected clock; got %v, want %v", st.deferNow, fixed)
	}
	if !svc.Now().Equal(fixed) {
		t.Errorf("Service.Now() should expose the injected clock; got %v", svc.Now())
	}
}
