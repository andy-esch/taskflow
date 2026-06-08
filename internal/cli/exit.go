package cli

import (
	"errors"

	"github.com/andy-esch/taskflow/internal/domain"
)

// ExitCode maps an error to a semantic exit code, so agents can route on the
// code without parsing text. 0 also covers idempotent no-ops.
func ExitCode(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, domain.ErrNotFound):
		return 10
	case errors.Is(err, domain.ErrValidation):
		return 11
	case errors.Is(err, domain.ErrInvalidTransition):
		return 12
	case errors.Is(err, domain.ErrAmbiguous):
		return 13
	default:
		return 1
	}
}
