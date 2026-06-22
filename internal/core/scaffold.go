package core

import (
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// renderTemplate replaces {{key}} tokens in body with values[key]. Unknown tokens
// are left literal (visible) — so a typo'd placeholder surfaces instead of
// vanishing, and a body containing a literal '%' is harmless (no Printf arity to
// get wrong). This is the single fill primitive both the create paths (real values)
// and the preview renderers use.
func renderTemplate(body string, values map[string]string) string {
	for k, v := range values {
		body = strings.ReplaceAll(body, "{{"+k+"}}", v)
	}
	return body
}

// labelValues maps a kind's placeholder keys to their preview labels, registry-
// driven (no per-kind switch) — what previews show in lieu of real values.
func labelValues(kind string) map[string]string {
	ps := domain.Placeholders(kind)
	out := make(map[string]string, len(ps))
	for _, p := range ps {
		out[p.Key] = p.Label
	}
	return out
}

// RenderLabels renders a template body with the kind's placeholder preview labels
// (e.g. {{title}} -> <title>), for `template show` / `schema <kind>`. Unknown
// {{tokens}} are left literal. Exported so the cli can render a body it already
// looked up, without a second resolution.
func RenderLabels(kind, body string) string {
	return renderTemplate(body, labelValues(kind))
}

// TemplateBody looks up a kind's named template and renders it with placeholder
// labels — the preview `tskflwctl schema <kind>` and `template show` display, NOT
// the create path (which fills real values). Empty name selects the kind's default.
// Unknown kind or template name returns ErrValidation.
func TemplateBody(kind, name string) (string, error) {
	nt, err := domain.LookupTemplate(kind, name)
	if err != nil {
		return "", err
	}
	return RenderLabels(kind, nt.Body), nil
}

// ScaffoldBody renders a kind's DEFAULT body scaffold with placeholder labels, the
// preview `tskflwctl schema <kind>` shows — single-sourced with the create path's
// template so the two can't drift.
func ScaffoldBody(kind string) (string, error) { return TemplateBody(kind, "") }
