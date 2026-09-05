package core

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// conflictStore returns domain.ErrConflict for the first `conflicts` calls to each mutating
// method, then succeeds — an in-memory stand-in for losing a version-CAS race N times before
// winning. It counts calls so a test can assert how many attempts the retry made.
type conflictStore struct {
	nopStore
	conflicts         int   // number of leading calls that return ErrConflict
	failErr           error // if non-nil, returned instead of ErrConflict (non-conflict passthrough)
	setCalls          int
	bodyCalls         int
	transformCalls    int
	auditMoveCalls    int
	auditBodyCalls    int
	researchSetCalls  int
	researchBodyCalls int
}

func (c *conflictStore) errFor(n int) error {
	if c.failErr != nil {
		return c.failErr
	}
	if n <= c.conflicts {
		return domain.ErrConflict
	}
	return nil
}

func (c *conflictStore) SetFields(slug string, _ map[string]any, _ bool) (domain.Task, error) {
	c.setCalls++
	return domain.Task{Slug: slug}, c.errFor(c.setCalls)
}
func (c *conflictStore) EditBody(slug, text string, _ bool, _ time.Time, _ bool) (domain.Task, string, error) {
	c.bodyCalls++
	return domain.Task{Slug: slug}, text, c.errFor(c.bodyCalls)
}
func (c *conflictStore) TransformTaskBody(slug string, _ time.Time, _ bool, transform func(string) (string, error)) (domain.Task, string, bool, error) {
	c.transformCalls++
	body, err := transform("# Task\n\n## Acceptance criteria\n\n- [ ] works\n")
	if err != nil {
		return domain.Task{}, "", false, err
	}
	if err := c.errFor(c.transformCalls); err != nil {
		return domain.Task{}, "", false, err
	}
	return domain.Task{Slug: slug}, body, true, nil
}
func (c *conflictStore) MoveAudit(slug string, _ domain.AuditBucket, _ bool) (domain.Audit, error) {
	c.auditMoveCalls++
	return domain.Audit{Slug: slug}, c.errFor(c.auditMoveCalls)
}
func (c *conflictStore) AppendAuditBody(slug, text string, _ time.Time, _ bool) (domain.Audit, string, error) {
	c.auditBodyCalls++
	return domain.Audit{Slug: slug}, text, c.errFor(c.auditBodyCalls)
}

func (c *conflictStore) SetResearchFields(slug string, _ map[string]any, _ bool) (domain.Research, error) {
	c.researchSetCalls++
	return domain.Research{Slug: slug}, c.errFor(c.researchSetCalls)
}
func (c *conflictStore) AppendResearchBody(slug, text string, _ time.Time, _ bool) (domain.Research, string, error) {
	c.researchBodyCalls++
	return domain.Research{Slug: slug}, text, c.errFor(c.researchBodyCalls)
}

// noopSleep is a WithRetry sleep that returns instantly, so retry tests don't wait.
func noopSleep(int) {}

func TestRetry_SucceedsAfterTransientConflicts(t *testing.T) {
	cs := &conflictStore{conflicts: 2}
	svc := NewService(cs, WithRetry(4, noopSleep))
	if _, err := svc.SetFields("t", map[string]any{"priority": "low"}, false, false); err != nil {
		t.Fatalf("2 transient conflicts (< 4 retries) should heal: %v", err)
	}
	if cs.setCalls != 3 {
		t.Errorf("want 3 store calls (1 + 2 retries), got %d", cs.setCalls)
	}
}

func TestRetry_ExhaustionSurfacesConflict(t *testing.T) {
	cs := &conflictStore{conflicts: 100}
	svc := NewService(cs, WithRetry(4, noopSleep))
	_, err := svc.SetFields("t", map[string]any{"priority": "low"}, false, false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("persistent contention must surface ErrConflict (exit 14), got %v", err)
	}
	if cs.setCalls != 5 {
		t.Errorf("want 5 store calls (1 + 4 retries) then give up, got %d", cs.setCalls)
	}
}

func TestRetry_DryRunNotRetried(t *testing.T) {
	cs := &conflictStore{conflicts: 100}
	svc := NewService(cs, WithRetry(4, noopSleep))
	_, err := svc.SetFields("t", map[string]any{"priority": "low"}, false, true) // dryRun
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a conflicting dry-run should return the conflict, got %v", err)
	}
	if cs.setCalls != 1 {
		t.Errorf("a dry-run is a preview and must not be retried; want 1 call, got %d", cs.setCalls)
	}
}

func TestRetry_NonConflictErrorNotRetried(t *testing.T) {
	cs := &conflictStore{failErr: domain.ErrValidation}
	svc := NewService(cs, WithRetry(4, noopSleep))
	_, err := svc.SetFields("t", map[string]any{"priority": "low"}, false, false)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("a non-conflict error must pass through unretried, got %v", err)
	}
	if cs.setCalls != 1 {
		t.Errorf("a non-conflict error must not be retried; want 1 call, got %d", cs.setCalls)
	}
}

// The body-append path retries too — and is safe to, because the store's version-CAS fails
// BEFORE the write, so a re-run re-appends onto fresh content exactly once (no double-apply).
func TestRetry_AppendBodyRetries(t *testing.T) {
	cs := &conflictStore{conflicts: 1}
	svc := NewService(cs, WithRetry(4, noopSleep))
	if _, _, err := svc.AppendBody("t", "more", false); err != nil {
		t.Fatalf("append should heal a transient conflict: %v", err)
	}
	if cs.bodyCalls != 2 {
		t.Errorf("want 2 store calls (1 + 1 retry), got %d", cs.bodyCalls)
	}
}

func TestRetry_SetAcceptanceCriterionReappliesAfterConflict(t *testing.T) {
	cs := &conflictStore{conflicts: 2}
	svc := NewService(cs, WithRetry(4, noopSleep))
	if _, body, changed, err := svc.SetAcceptanceCriterion("t", 1, true, false); err != nil {
		t.Fatalf("criterion edit should heal transient conflicts: %v", err)
	} else if !changed || !strings.Contains(body, "- [x] works") {
		t.Fatalf("criterion edit result = changed %v body %q", changed, body)
	}
	if cs.transformCalls != 3 {
		t.Errorf("want 3 transformed snapshots (1 + 2 retries), got %d", cs.transformCalls)
	}
}

func TestRetry_SetAcceptanceCriterionExhaustionSurfacesConflict(t *testing.T) {
	cs := &conflictStore{conflicts: 100}
	svc := NewService(cs, WithRetry(2, noopSleep))
	_, _, _, err := svc.SetAcceptanceCriterion("t", 1, true, false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("persistent criterion contention must surface ErrConflict, got %v", err)
	}
	if cs.transformCalls != 3 {
		t.Errorf("want 3 transformed snapshots (1 + 2 retries), got %d", cs.transformCalls)
	}
}

func TestRetry_SetAcceptanceCriterionDryRunIsNotRetried(t *testing.T) {
	cs := &conflictStore{conflicts: 100}
	svc := NewService(cs, WithRetry(4, noopSleep))
	_, _, _, err := svc.SetAcceptanceCriterion("t", 1, true, true)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a conflicting criterion dry-run should return the conflict, got %v", err)
	}
	if cs.transformCalls != 1 {
		t.Errorf("a criterion dry-run must be attempted once, got %d calls", cs.transformCalls)
	}
}

// Cross-entity coverage: the retry helper is entity-agnostic, so an audit mutation heals a
// transient conflict exactly like a task one (guards against a per-method wrap regression —
// wrong `now` capture or return-packing — in the audit/epic wraps the task tests don't touch).
func TestRetry_MoveAuditRetries(t *testing.T) {
	cs := &conflictStore{conflicts: 2}
	svc := NewService(cs, WithRetry(4, noopSleep))
	if _, err := svc.MoveAudit("a", domain.AuditClosed, false); err != nil {
		t.Fatalf("MoveAudit should heal transient conflicts: %v", err)
	}
	if cs.auditMoveCalls != 3 {
		t.Errorf("want 3 store calls (1 + 2 retries), got %d", cs.auditMoveCalls)
	}
}

func TestRetry_AppendAuditBodyRetries(t *testing.T) {
	cs := &conflictStore{conflicts: 1}
	svc := NewService(cs, WithRetry(4, noopSleep))
	if _, _, err := svc.AppendAuditBody("a", "x", false); err != nil {
		t.Fatalf("AppendAuditBody should heal a transient conflict: %v", err)
	}
	if cs.auditBodyCalls != 2 {
		t.Errorf("want 2 store calls (1 + 1 retry), got %d", cs.auditBodyCalls)
	}
}

// TestRetryBackoff_BoundedAndPanicFree exercises the REAL backoff math (defaultRetrySleep's
// pure core) past the point the exponential saturates AND past where the raw shift overflows
// int64 — the delay stays within [0, cap) and rand.Int63n never gets a <= 0 arg (no panic).
func TestRetryBackoff_BoundedAndPanicFree(t *testing.T) {
	const capDelay = 50 * time.Millisecond
	for attempt := 1; attempt <= 70; attempt++ {
		if d := retryBackoff(attempt); d < 0 || d >= capDelay {
			t.Fatalf("retryBackoff(%d) = %v, want within [0, %v)", attempt, d, capDelay)
		}
	}
}

// Research rides the same entity-agnostic retry helper, so a transient version-CAS loss must
// heal for `research set` / `research append` exactly as it does for tasks and audits. Added
// because the research write paths shipped with NO retry coverage at all: unwrapping both
// retryOnConflict calls in service_research.go left the whole suite green.
func TestRetry_SetResearchFieldsRetries(t *testing.T) {
	cs := &conflictStore{conflicts: 2}
	svc := NewService(cs, WithRetry(4, noopSleep))
	if _, err := svc.SetResearchFields("r", map[string]any{"description": "d"}, false, false); err != nil {
		t.Fatalf("SetResearchFields should heal transient conflicts: %v", err)
	}
	if cs.researchSetCalls != 3 {
		t.Errorf("want 3 store calls (1 + 2 retries), got %d", cs.researchSetCalls)
	}
}

func TestRetry_AppendResearchBodyRetries(t *testing.T) {
	cs := &conflictStore{conflicts: 1}
	svc := NewService(cs, WithRetry(4, noopSleep))
	if _, _, err := svc.AppendResearchBody("r", "more", false); err != nil {
		t.Fatalf("AppendResearchBody should heal a transient conflict: %v", err)
	}
	if cs.researchBodyCalls != 2 {
		t.Errorf("want 2 store calls (1 + 1 retry), got %d", cs.researchBodyCalls)
	}
}

// A dry run must not retry — it never writes, so a conflict from it is not transient.
func TestRetry_ResearchDryRunNotRetried(t *testing.T) {
	cs := &conflictStore{conflicts: 3}
	svc := NewService(cs, WithRetry(4, noopSleep))
	if _, err := svc.SetResearchFields("r", map[string]any{"description": "d"}, false, true); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want the conflict surfaced, got %v", err)
	}
	if cs.researchSetCalls != 1 {
		t.Errorf("a dry run must be attempted once, got %d calls", cs.researchSetCalls)
	}
}
