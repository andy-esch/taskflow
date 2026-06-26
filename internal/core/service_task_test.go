package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// deferStore drives DeferTask's two-phase Move-then-SetFields path in isolation:
// Move always "succeeds" (records the call), and SetFields returns setErr — so a
// test can simulate the partial-failure window where the file has already moved
// into deferred/ but the revisit_at write fails. It records its calls so the test
// can assert the ordering/contract, not just the returned error.
type deferStore struct {
	nopStore
	moveCalls   int
	setCalls    int
	lastSlug    string
	lastUpdates map[string]any
	dryRun      bool
	setErr      error
}

func (s *deferStore) Move(slug string, to domain.Status, _ time.Time, dryRun bool) (domain.Task, error) {
	s.moveCalls++
	s.dryRun = dryRun
	return domain.Task{Slug: slug, Status: to}, nil
}

func (s *deferStore) SetFields(slug string, updates map[string]any, dryRun bool) (domain.Task, error) {
	s.setCalls++
	s.lastSlug = slug
	s.lastUpdates = updates
	if s.setErr != nil {
		return domain.Task{}, s.setErr
	}
	return domain.Task{Slug: slug, Status: domain.StatusDeferred, RevisitAt: fmt.Sprint(updates["revisit_at"])}, nil
}

// TestDeferTask_SetFieldsFailsAfterMove pins the non-atomic partial-failure
// contract the adversarial review flagged: Move persists (file already in
// deferred/) and then SetFields fails, so the task is deferred WITHOUT a
// revisit_at. DeferTask must surface the failure, keep the underlying sentinel
// (for the exit code), and name the partial state — not swallow it or report a
// plain success.
func TestDeferTask_SetFieldsFailsAfterMove(t *testing.T) {
	st := &deferStore{setErr: fmt.Errorf("%w: write clobbered", domain.ErrConflict)}
	svc := NewService(st)

	_, err := svc.DeferTask("alpha", "2026-09-01", false)
	if err == nil {
		t.Fatal("DeferTask must propagate a post-move SetFields failure, got nil")
	}
	// The sentinel survives the wrap, so the CLI still maps it to exit 14.
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("wrapped error should keep the ErrConflict sentinel, got %v", err)
	}
	// The message names the partial state (deferred, date not recorded) so the
	// report doesn't read as "nothing happened".
	if msg := err.Error(); !strings.Contains(msg, "deferred") || !strings.Contains(msg, "not recorded") {
		t.Errorf("error should describe the partial state, got %q", msg)
	}
	// Move ran (the file moved) and SetFields was attempted exactly once with the
	// revisit_at update — the two-phase ordering the contract depends on.
	if st.moveCalls != 1 || st.setCalls != 1 {
		t.Errorf("want exactly one Move + one SetFields, got move=%d set=%d", st.moveCalls, st.setCalls)
	}
	if st.lastUpdates["revisit_at"] != "2026-09-01" {
		t.Errorf("SetFields should carry the revisit_at update, got %v", st.lastUpdates)
	}
}

// TestDeferTask_DryRunSkipsSecondWrite pins that --dry-run never reaches the
// SetFields write yet still reflects the would-be revisit_at on the previewed
// task (the field the move report surfaces).
func TestDeferTask_DryRunSkipsSecondWrite(t *testing.T) {
	st := &deferStore{}
	svc := NewService(st)

	got, err := svc.DeferTask("alpha", "2026-09-01", true)
	if err != nil {
		t.Fatalf("dry-run DeferTask: %v", err)
	}
	if st.setCalls != 0 {
		t.Errorf("dry-run must not call SetFields, got %d calls", st.setCalls)
	}
	if !st.dryRun {
		t.Error("dry-run flag should reach the store's Move")
	}
	if got.RevisitAt != "2026-09-01" {
		t.Errorf("dry-run preview should carry the would-be revisit_at, got %q", got.RevisitAt)
	}
}

// TestDeferTask_BareDeferSkipsSetFields pins that a defer with no date is exactly
// Move(deferred) — no second write, no revisit_at.
func TestDeferTask_BareDeferSkipsSetFields(t *testing.T) {
	st := &deferStore{}
	svc := NewService(st)

	if _, err := svc.DeferTask("alpha", "", false); err != nil {
		t.Fatalf("bare DeferTask: %v", err)
	}
	if st.moveCalls != 1 || st.setCalls != 0 {
		t.Errorf("bare defer should Move once and never SetFields, got move=%d set=%d", st.moveCalls, st.setCalls)
	}
}
