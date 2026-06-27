package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/andy-esch/taskflow/internal/wire"
)

// --- schema (the tool's self-description for agents) ---
//
// The schema-contract DTOs are wire types (machine contract); they're re-exported
// here so the CLI's schema command keeps building them through the render package,
// while the human renderers below consume them.

// SchemaStatus is one task status and whether it is part of the working set.
type SchemaStatus = wire.SchemaStatus

// SchemaField is one known frontmatter field and its YAML storage type.
type SchemaField = wire.SchemaField

// SchemaExitCode is one exit code and its stable machine name (also the `code`
// in the --json error envelope).
type SchemaExitCode = wire.SchemaExitCode

// SchemaContract is the global machine contract (`tskflwctl schema`): everything
// an agent needs to drive the tool without parsing --help prose.
type SchemaContract = wire.SchemaContract

// KindSchema is the per-kind authoring guidance (`tskflwctl schema <kind>`): how
// to compose a well-formed document of that kind.
type KindSchema = wire.KindSchema

// SchemaJSON writes the global contract envelope.
func SchemaJSON(w io.Writer, c SchemaContract) error {
	return wire.EncodeJSON(w, wire.ToSchemaEnvelope(c))
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

// SchemaKindJSON writes the per-kind authoring envelope.
func SchemaKindJSON(w io.Writer, ks KindSchema) error {
	return wire.EncodeJSON(w, wire.ToSchemaKindEnvelope(ks))
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
