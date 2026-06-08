// Package domain holds the pure planning entities and their invariants.
// It imports no infrastructure (no fs, no cobra) and is unit-testable in
// isolation.
package domain

import "fmt"

// Status is a task lifecycle state. It is identical to the directory name a
// task file lives in; the "status == directory" invariant lives here.
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

// ParseStatus validates s and returns the typed Status.
func ParseStatus(s string) (Status, error) {
	for _, st := range allStatuses {
		if Status(s) == st {
			return st, nil
		}
	}
	return "", fmt.Errorf("invalid status %q", s)
}

// Valid reports whether s is a known status.
func (s Status) Valid() bool {
	_, err := ParseStatus(string(s))
	return err == nil
}

// Dir is the directory name for this status (identical to its string value).
func (s Status) Dir() string { return string(s) }
