package cli

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/domain"
)

// problemsError returns a validation error (non-zero exit) when any per-file
// load problems exist, else nil. It does not print: human commands render the
// problems to stderr themselves, and JSON commands carry them in the payload.
func problemsError(problems []domain.FileProblem) error {
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %d file(s) with unreadable frontmatter", domain.ErrValidation, len(problems))
}
