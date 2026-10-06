package core

import (
	"errors"
	"fmt"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestMutationPolicyModesFailClosed(t *testing.T) {
	var nilGuard func() error
	for _, tc := range []struct {
		name   string
		policy MutationPolicy
		want   error
	}{
		{"zero", MutationPolicy{}, ErrInvalidMutationPolicy},
		{"nil guard", GuardedMutations(nilGuard), ErrInvalidMutationPolicy},
		{"invalid mode", MutationPolicy{mode: 255}, ErrInvalidMutationPolicy},
		{"read-only", ReadOnlyMutations(), ErrReadOnlyPersistence},
		{"unrestricted", UnrestrictedMutations(), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.policy.Authorize(); !errors.Is(err, tc.want) {
				t.Fatalf("Authorize = %v; want %v", err, tc.want)
			}
			if tc.want != nil && !errors.Is(tc.want, domain.ErrValidation) {
				t.Fatal("policy refusal lost domain classification")
			}
			wantValidation := tc.want
			if errors.Is(tc.want, ErrReadOnlyPersistence) {
				wantValidation = nil // a read-only policy is a valid composition choice
			}
			if err := tc.policy.Validate(); !errors.Is(err, wantValidation) {
				t.Fatalf("Validate = %v; want %v", err, wantValidation)
			}
		})
	}
}

func TestMutationPolicyValidatesWithoutCallingAndNeverCachesAuthorization(t *testing.T) {
	calls := 0
	cause := errors.New("invocation changed")
	var decision error
	policy := GuardedMutations(func() error { calls++; return decision })
	for range 2 {
		if err := policy.Validate(); err != nil || calls != 0 {
			t.Fatalf("construction authorization calls=%d err=%v", calls, err)
		}
	}
	if err := policy.Authorize(); err != nil || calls != 1 {
		t.Fatalf("first authorization calls=%d err=%v", calls, err)
	}
	decision = fmt.Errorf("caller context: %w", cause)
	if err := policy.Authorize(); err != decision || !errors.Is(err, cause) || calls != 2 {
		t.Fatalf("changed authorization calls=%d err=%v", calls, err)
	}
	decision = nil
	if err := policy.Authorize(); err != nil || calls != 3 {
		t.Fatalf("restored authorization calls=%d err=%v", calls, err)
	}
}
