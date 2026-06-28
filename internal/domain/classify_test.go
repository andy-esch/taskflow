package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Class
	}{
		{"nil", nil, ClassUnknown},
		{"not-found sentinel", ErrNotFound, ClassNotFound},
		{"validation sentinel", ErrValidation, ClassValidation},
		{"ambiguous sentinel", ErrAmbiguous, ClassAmbiguous},
		{"conflict sentinel", ErrConflict, ClassConflict},
		// The %w-wrapped form every service/store error actually returns must
		// classify identically to the bare sentinel.
		{"wrapped not-found", fmt.Errorf("task %q: %w", "x", ErrNotFound), ClassNotFound},
		{"wrapped validation", fmt.Errorf("%w: at least one tag is required", ErrValidation), ClassValidation},
		{"wrapped conflict", fmt.Errorf("%q: %w", "in-progress", ErrConflict), ClassConflict},
		{"unknown (no sentinel)", errors.New("disk on fire"), ClassUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Classify(tt.err); got != tt.want {
				t.Errorf("Classify(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestReason(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"nil", nil, ""},
		{"validation reason stripped", fmt.Errorf("%w: at least one tag is required", ErrValidation), "at least one tag is required"},
		// A reason that itself contains a colon must survive (the first-": " cut
		// would have mangled it).
		{"reason with colon", fmt.Errorf("%w: revisit offset %q: bad", ErrValidation, "2z"), `revisit offset "2z": bad`},
		{"not-found reason stripped", fmt.Errorf("%w: task %q", ErrNotFound, "x"), `task "x"`},
		// No sentinel prefix → returned verbatim.
		{"unknown verbatim", errors.New("disk on fire"), "disk on fire"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reason(tt.err); got != tt.want {
				t.Errorf("Reason(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}
