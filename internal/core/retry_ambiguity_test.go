package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TestRetry_AmbiguityIsNotSleptThrough pins the cost side of the store guard's
// conflict/failure split. TestRetry_NonConflictErrorNotRetried already shows the
// loop ignores a non-conflict; this names the class that used to arrive here
// MISLABELLED as a conflict — a duplicate stable id — and counts the backoff
// sleeps, because the user-visible symptom was never the retry count but the
// four waits before advice that could not work.
//
// The store-side half is TestVerifyUnchanged_DuplicateIDSurfacesAsAmbiguous.
func TestRetry_AmbiguityIsNotSleptThrough(t *testing.T) {
	ambiguous := fmt.Errorf("id %q is claimed by 2 files: alpha (%s), beta (%s): %w",
		"6fjangd7kvc1", "6fjangd7kvc1", "6fjangd7kvc1", domain.ErrAmbiguous)
	cs := &conflictStore{failErr: ambiguous}
	sleeps := 0
	svc := NewService(cs, WithRetry(4, func(int) { sleeps++ }))

	_, err := svc.SetFields("alpha", map[string]any{"priority": "low"}, false, false)
	if !errors.Is(err, domain.ErrAmbiguous) {
		t.Fatalf("the ambiguity must reach the caller, got %v", err)
	}
	if errors.Is(err, domain.ErrConflict) {
		t.Errorf("an ambiguity must not also present as a conflict: %v", err)
	}
	if sleeps != 0 {
		t.Errorf("an ambiguity must not burn backoff sleeps, slept %d times", sleeps)
	}
	if cs.setCalls != 1 {
		t.Errorf("want a single attempt, got %d", cs.setCalls)
	}
	// The names are what make the state fixable, so they must survive the trip.
	for _, name := range []string{"alpha", "beta"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("the colliding file %q should still be named: %v", name, err)
		}
	}
}
