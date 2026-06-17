package core

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/domain"
)

// ScaffoldBody returns the default body template for a kind, rendered with
// placeholder values. It is the SAME template `task new`/`epic new`/`audit new`
// write, so `tskflwctl schema <kind>` can show the body skeleton without a
// second, drift-prone copy. Unknown kinds return ErrValidation.
func ScaffoldBody(kind string) (string, error) {
	switch kind {
	case "task":
		return fmt.Sprintf(taskBodyTemplate, "<title>", "<epic-id>"), nil
	case "epic":
		return fmt.Sprintf(epicBodyTemplate, "<title>", "<description>"), nil
	case "audit":
		return fmt.Sprintf(auditBodyTemplate, "<area>", "<date>"), nil
	}
	return "", fmt.Errorf("%w: unknown kind %q (task|epic|audit)", domain.ErrValidation, kind)
}
