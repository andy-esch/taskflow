// Package domain holds the pure planning entities and their invariants.
// It imports no infrastructure (no fs, no cobra) and is unit-testable in
// isolation.
package domain

import (
	"fmt"
	"strings"
)

// Status is a task lifecycle state; its string value is also the directory name a
// task file mirrors into. The authoritative value now lives in frontmatter (ADR-0003
// Phase A) — the directory is a lock-step mirror, not the source of truth.
type Status string

// The lifecycle states, in display order.
const (
	StatusNextUp       Status = "next-up"
	StatusReadyToStart Status = "ready-to-start"
	StatusInProgress   Status = "in-progress"
	StatusCompleted    Status = "completed"
	StatusDeprecated   Status = "deprecated"
	StatusDeferred     Status = "deferred"
)

var allStatuses = []Status{
	StatusNextUp, StatusReadyToStart, StatusInProgress,
	StatusCompleted, StatusDeprecated, StatusDeferred,
}

// AllStatuses returns every lifecycle status, in display order.
func AllStatuses() []Status { return allStatuses }

// ActiveStatuses returns the active-pipeline statuses (those Status.IsActive
// reports) in display order — the working set a pipeline/board view iterates. It's
// AllStatuses filtered by IsActive, so "the active set" has one definition and one
// order, owned here in the domain rather than re-derived per use case.
func ActiveStatuses() []Status {
	out := make([]Status, 0, len(allStatuses))
	for _, st := range allStatuses {
		if st.IsActive() {
			out = append(out, st)
		}
	}
	return out
}

// ParseStatus validates s and returns the typed Status. The failure wraps
// ErrValidation (exit 11 at the CLI) and enumerates the valid statuses — a typo
// must be a loud error, not a silently empty result.
func ParseStatus(s string) (Status, error) {
	for _, st := range allStatuses {
		if Status(s) == st {
			return st, nil
		}
	}
	return "", fmt.Errorf("%w: invalid status %q (valid: %s)", ErrValidation, s, statusList())
}

// statusList renders the valid statuses for error messages.
func statusList() string {
	names := make([]string, len(allStatuses))
	for i, st := range allStatuses {
		names[i] = string(st)
	}
	return strings.Join(names, ", ")
}

// Valid reports whether s is a known status.
func (s Status) Valid() bool {
	_, err := ParseStatus(string(s))
	return err == nil
}

// Dir is the directory name for this status (identical to its string value).
func (s Status) Dir() string { return string(s) }

// IsActive reports whether a task in this status is part of the working set
// (not completed/deprecated/deferred). The lifecycle invariant lives on the
// domain type, not in a use-case.
func (s Status) IsActive() bool {
	switch s {
	case StatusNextUp, StatusReadyToStart, StatusInProgress:
		return true
	default:
		return false
	}
}
