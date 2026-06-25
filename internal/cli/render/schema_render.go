package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// --- schema (the tool's self-description for agents) ---

// SchemaStatus is one task status and whether it is part of the working set.
type SchemaStatus struct {
	Value  string `json:"value"`
	Active bool   `json:"active"`
}

// SchemaField is one known frontmatter field and its YAML storage type.
type SchemaField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// SchemaExitCode is one exit code and its stable machine name (also the `code`
// in the --json error envelope).
type SchemaExitCode struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}

// SchemaContract is the global machine contract (`tskflwctl schema`): everything
// an agent needs to drive the tool without parsing --help prose.
type SchemaContract struct {
	Statuses     []SchemaStatus   `json:"statuses"`
	EpicStatuses []string         `json:"epic_statuses"`
	AuditBuckets []string         `json:"audit_buckets"`
	TaskFields   []SchemaField    `json:"task_fields"`
	EpicFields   []string         `json:"epic_fields"`
	ExitCodes    []SchemaExitCode `json:"exit_codes"`
	Kinds        []string         `json:"kinds"`
}

// SchemaJSON writes the global contract envelope.
func SchemaJSON(w io.Writer, c SchemaContract) error {
	return encodeJSON(w, SchemaEnvelope{SchemaVersion: SchemaVersion, SchemaContract: c})
}

// SchemaHuman renders the global contract as readable sections.
func SchemaHuman(w io.Writer, st Style, c SchemaContract) error {
	fmt.Fprintf(w, "%s %s\n\n", st.Bold("tskflwctl schema"), st.Dim("v"+SchemaVersion))
	fmt.Fprintf(w, "%s:\n", st.Bold("Task statuses"))
	for _, s := range c.Statuses {
		active := ""
		if s.Active {
			active = st.Dim(" (active)")
		}
		fmt.Fprintf(w, "  %s%s\n", s.Value, active)
	}
	fmt.Fprintf(w, "\n%s: %s\n", st.Bold("Epic statuses"), strings.Join(c.EpicStatuses, ", "))
	fmt.Fprintf(w, "%s: %s\n", st.Bold("Audit buckets"), strings.Join(c.AuditBuckets, ", "))
	fmt.Fprintf(w, "%s:   %s\n", st.Bold("Doc kinds"), strings.Join(c.Kinds, ", "))
	fmt.Fprintf(w, "\n%s:\n", st.Bold("Task fields"))
	for _, f := range c.TaskFields {
		fmt.Fprintf(w, "  %-16s %s\n", f.Name, st.Dim(f.Type))
	}
	fmt.Fprintf(w, "\n%s: %s\n", st.Bold("Epic fields"), strings.Join(c.EpicFields, ", "))
	fmt.Fprintf(w, "\n%s:\n", st.Bold("Exit codes"))
	for _, e := range c.ExitCodes {
		fmt.Fprintf(w, "  %-3d %s\n", e.Code, st.Dim(e.Name))
	}
	fmt.Fprintf(w, "\n%s\n", st.Dim("`tskflwctl schema <task|epic|audit>` for per-kind authoring guidance."))
	return nil
}

// KindSchema is the per-kind authoring guidance (`tskflwctl schema <kind>`): how
// to compose a well-formed document of that kind.
type KindSchema struct {
	Kind         string            `json:"kind"`
	Sections     []string          `json:"sections"`
	BodyTemplate string            `json:"body_template"`
	Fields       []domain.FieldDoc `json:"fields"`
	Conventions  []string          `json:"conventions"`
	Templates    []TemplateInfo    `json:"templates"`
}

// SchemaKindJSON writes the per-kind authoring envelope.
func SchemaKindJSON(w io.Writer, ks KindSchema) error {
	return encodeJSON(w, SchemaKindEnvelope{SchemaVersion: SchemaVersion, KindSchema: ks})
}

// SchemaKindHuman renders the per-kind authoring guidance.
func SchemaKindHuman(w io.Writer, st Style, ks KindSchema) error {
	fmt.Fprintf(w, "%s %s\n\n", st.Bold("schema "+ks.Kind), st.Dim("— authoring guidance"))
	fmt.Fprintf(w, "%s: %s\n\n", st.Bold("Sections"), strings.Join(ks.Sections, " · "))
	fmt.Fprintf(w, "%s:\n", st.Bold("Frontmatter"))
	for _, f := range ks.Fields {
		req := ""
		if f.Required {
			req = st.Dim(" (required)")
		}
		fmt.Fprintf(w, "  %-15s %s%s — %s %s\n", f.Name, st.Dim(f.Type), req, f.Description, st.Dim("e.g. "+f.Example))
	}
	fmt.Fprintf(w, "\n%s:\n", st.Bold("Conventions"))
	for _, c := range ks.Conventions {
		fmt.Fprintf(w, "  %s %s\n", st.Dim("-"), c)
	}
	if len(ks.Templates) > 0 {
		fmt.Fprintf(w, "\n%s %s:\n", st.Bold("Templates"), st.Dim("(--template)"))
		for _, t := range ks.Templates {
			fmt.Fprintf(w, "  %-12s %s\n", t.Name, st.Dim(t.Description))
		}
	}
	fmt.Fprintf(w, "\n%s:\n%s\n", st.Bold("Body template"), ks.BodyTemplate)
	return nil
}
