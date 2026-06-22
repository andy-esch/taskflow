package core

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TemplateBody renders a kind's named template scaffold with placeholder labels —
// the body `tskflwctl schema <kind>` and `template show` display, NOT the create
// path (which fills real title/epic/area values). An empty name selects the kind's
// default. Unknown kind or template name returns ErrValidation.
func TemplateBody(kind, name string) (string, error) {
	tmpl, err := domain.Template(kind, name)
	if err != nil {
		return "", err
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

// ScaffoldBody renders a kind's DEFAULT body scaffold with placeholder values, the
// same scaffold `task new`/`epic new`/`audit new` write — so `tskflwctl schema
// <kind>` shows the skeleton without a second, drift-prone copy.
func ScaffoldBody(kind string) (string, error) { return TemplateBody(kind, "") }
