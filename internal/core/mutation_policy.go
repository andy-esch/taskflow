package core

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/domain"
)

// MutationPolicy is an explicit persistence-composition choice, not a security
// credential. Its zero value is invalid. Adapters validate it at construction
// and authorize every mutation, including previews, before any effects or callbacks.
// This value deliberately has no interface/typed-nil or implicit-allow mode.
type MutationPolicy struct {
	mode      mutationMode
	authorize func() error
}

type mutationMode uint8

const (
	mutationReadOnly mutationMode = iota + 1
	mutationGuarded
	mutationUnrestricted
)

var (
	ErrInvalidMutationPolicy = fmt.Errorf("%w: explicit mutation policy is required", domain.ErrValidation)
	ErrReadOnlyPersistence   = fmt.Errorf("%w: persistence is read-only", domain.ErrValidation)
)

// ReadOnlyMutations permits reads but denies all mutation APIs, even dry runs.
func ReadOnlyMutations() MutationPolicy { return MutationPolicy{mode: mutationReadOnly} }

// GuardedMutations invokes authorize at each mutation, never at construction.
// A nil function creates an invalid policy; constructors must refuse it.
func GuardedMutations(authorize func() error) MutationPolicy {
	return MutationPolicy{mode: mutationGuarded, authorize: authorize}
}

// UnrestrictedMutations deliberately opts trusted callers and writable fixtures
// out of invocation authorization. Production CLI/TUI composition uses guards.
// This does not bypass repository locks, CAS, or domain validation.
func UnrestrictedMutations() MutationPolicy { return MutationPolicy{mode: mutationUnrestricted} }

// Validate checks the composition choice without executing user code or I/O.
func (p MutationPolicy) Validate() error {
	switch p.mode {
	case mutationReadOnly, mutationUnrestricted:
		return nil
	case mutationGuarded:
		if p.authorize != nil {
			return nil
		}
	}
	return ErrInvalidMutationPolicy
}

// Authorize must be called before each adapter mutation. Decisions are not
// cached: invocation safety can be bound or changed after adapter construction.
// Guard errors are returned intact, preserving the caller's classification.
func (p MutationPolicy) Authorize() error {
	if err := p.Validate(); err != nil {
		return err
	}
	switch p.mode {
	case mutationReadOnly:
		return ErrReadOnlyPersistence
	case mutationGuarded:
		return p.authorize()
	default:
		return nil // explicitly unrestricted; Validate rejected every other mode
	}
}
