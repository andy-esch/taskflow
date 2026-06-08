package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// MaxDescriptionLen is the frontmatter description cap (matches the pm rule).
const MaxDescriptionLen = 150

var validPriorities = map[string]bool{"high": true, "medium": true, "low": true}

// ValidateField checks a constrained frontmatter field's string value. Returns
// an error wrapping ErrValidation when invalid. Unconstrained fields pass.
func ValidateField(field, value string) error {
	switch field {
	case "status":
		return fmt.Errorf("%w: set status with `task <verb>`/`task move`, not `set`", ErrValidation)
	case "priority":
		if !validPriorities[value] {
			return fmt.Errorf("%w: priority must be high|medium|low, got %q", ErrValidation, value)
		}
	case "tier", "autonomy_level":
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 || n > 5 {
			return fmt.Errorf("%w: %s must be an integer 1-5, got %q", ErrValidation, field, value)
		}
	case "description":
		if strings.ContainsAny(value, "\r\n") {
			return fmt.Errorf("%w: description must be a single line", ErrValidation)
		}
		if len(value) > MaxDescriptionLen {
			return fmt.Errorf("%w: description too long (%d > %d)", ErrValidation, len(value), MaxDescriptionLen)
		}
	}
	return nil
}
