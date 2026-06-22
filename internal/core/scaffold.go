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
	tmpl := domain.BodyTemplate(kind)
	if tmpl == "" {
		return "", fmt.Errorf("%w: unknown kind %q (task|epic|audit)", domain.ErrValidation, kind)
	}
	// The template content is single-sourced on the descriptor; only the per-kind
	// placeholder values stay here (each scaffold is a 2-arg Printf format).
	switch kind {
	case "task":
		return fmt.Sprintf(tmpl, "<title>", "<epic-id>"), nil
	case "epic":
		return fmt.Sprintf(tmpl, "<title>", "<description>"), nil
	case "audit":
		return fmt.Sprintf(tmpl, "<area>", "<date>"), nil
	}
	return "", fmt.Errorf("%w: unknown kind %q (task|epic|audit)", domain.ErrValidation, kind)
}
