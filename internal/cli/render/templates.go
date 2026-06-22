package render

import (
	"fmt"
	"io"
	"strings"
)

// TemplateInfo is one body template's listable metadata (kind/name/description),
// populated by the cli for `template list`/`show`. The rendered body is carried
// separately (TemplateShow*), since the list view never needs it.
type TemplateInfo struct {
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TemplatesJSON writes the `template list --json` envelope.
func TemplatesJSON(w io.Writer, ts []TemplateInfo) error {
	return encodeJSON(w, TemplatesEnvelope{SchemaVersion: SchemaVersion, Templates: ts})
}

// TemplatesHuman renders the template list as a kind / name / description table.
func TemplatesHuman(w io.Writer, st Style, ts []TemplateInfo) {
	if len(ts) == 0 {
		fmt.Fprintln(w, st.Dim("no templates"))
		return
	}
	for _, t := range ts {
		fmt.Fprintf(w, "  %-7s %-12s %s\n", t.Kind, t.Name, st.Dim(t.Description))
	}
}

// TemplateShowJSON writes the `template show --json` envelope.
func TemplateShowJSON(w io.Writer, info TemplateInfo, body string) error {
	return encodeJSON(w, TemplateShowEnvelope{SchemaVersion: SchemaVersion, Template: info, Body: body})
}

// TemplateShowHuman renders a single template: a header line then its body.
func TemplateShowHuman(w io.Writer, st Style, info TemplateInfo, body string) {
	fmt.Fprintf(w, "%s %s\n", st.Bold(info.Kind+" template: "+info.Name), st.Dim("— "+info.Description))
	fmt.Fprintf(w, "%s\n", strings.TrimRight(body, "\n"))
}
