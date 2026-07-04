package core

import (
	"errors"
	"math/rand"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// defaultMaxRetries bounds the OCC auto-retry for scriptable mutations. Small on purpose:
// a genuinely contended file should surface a loud ErrConflict (exit 14) quickly, not spin
// — matching how Firestore/S3 cap their retries and then fail. The concurrent writers here
// are a handful of cron agents, not a thundering herd.
const defaultMaxRetries = 4

// retryBackoff is the delay before retry `attempt` (1-based): capped exponential backoff
// with FULL jitter — a value uniformly in [0, min(base·2^(attempt-1), cap)). Pure (no
// sleep) so the bound + overflow guard are unit-testable without waiting. The raw shift
// wraps for an absurd attempt (only reachable via a huge injected maxRetries); the
// `d <= 0 || d > capDelay` guard catches any wrap, and `rand.Int63n`'s arg is thus always
// > 0 (never panics).
func retryBackoff(attempt int) time.Duration {
	const base = 2 * time.Millisecond
	const capDelay = 50 * time.Millisecond
	d := base << (attempt - 1)
	if d <= 0 || d > capDelay {
		d = capDelay
	}
	return time.Duration(rand.Int63n(int64(d)))
}

// defaultRetrySleep sleeps retryBackoff(attempt), scaled to the microsecond reality of
// local-FS contention (not the seconds of a network service). The point isn't to wait long
// — it's to DE-CORRELATE cron writers that woke on the same schedule and would otherwise
// collide every round (AWS's backoff-and-jitter result). Injected via WithRetry so tests
// run instantly and deterministically.
func defaultRetrySleep(attempt int) { time.Sleep(retryBackoff(attempt)) }

// WithRetry overrides the OCC auto-retry policy — injected so tests drive the loop
// deterministically (a no-op sleep) and pin the bound. A negative maxRetries or a nil
// sleep leaves the respective default untouched.
func WithRetry(maxRetries int, sleep func(attempt int)) Option {
	return func(s *Service) {
		if maxRetries >= 0 {
			s.maxRetries = maxRetries
		}
		if sleep != nil {
			s.retrySleep = sleep
		}
	}
}

// retryOnConflict runs fn and, on domain.ErrConflict, re-runs it up to s.maxRetries more
// times (backoff+jitter between attempts) — the bounded, plain-retry auto-heal that makes
// the scriptable mutations (set/append/move/defer) concurrency-safe WITHOUT each agent
// reimplementing a read→re-apply→rewrite loop. Each re-run goes through the store's
// self-contained read-modify-write, so it re-reads fresh and re-derives the change; and
// because the version-CAS fails BEFORE any write, a plain retry is safe even for a body
// append (it re-appends onto fresh content exactly once). A dry-run is a preview — never
// retried. Successes and non-conflict errors return immediately.
func retryOnConflict[T any](s *Service, dryRun bool, fn func() (T, error)) (T, error) {
	v, err := fn()
	if dryRun {
		return v, err
	}
	for attempt := 1; attempt <= s.maxRetries && errors.Is(err, domain.ErrConflict); attempt++ {
		s.retrySleep(attempt)
		v, err = fn()
	}
	return v, err
}
