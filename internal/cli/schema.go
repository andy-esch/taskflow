package cli

import (
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// sectionRe pulls the `## ` headings out of a scaffold body so the schema's
// section list is derived from the real template, never hand-maintained.
var sectionRe = regexp.MustCompile(`(?m)^##\s+(.+)$`)

func newSchemaCmd(app *App) *cobra.Command {
	var jsonSchema bool
	cmd := &cobra.Command{
		Use:   "schema [task|epic|audit]",
		Short: "Describe the tool's contract + per-kind authoring guidance (for agents)",
		Long: "For triage, lead with the terse path: `epic show <id>` for an epic's task\n" +
			"roster, and `task list -o table -c slug,status,description` for a compact,\n" +
			"byte-stable table. --json is compact and also takes -c to project just the\n" +
			"fields you need; reach for full --json (no -c) when you need every frontmatter\n" +
			"field. A --json -c projection is a string-valued column view — only full --json\n" +
			"validates against --json-schema.\n\n" +
			"With no argument, emit the machine contract — statuses, the epic/bucket\n" +
			"enums, the task field registry with types, and the exit/error codes — so an\n" +
			"agent can drive the tool without parsing --help prose. With a kind, emit how\n" +
			"to author that document: the body section template, per-field guidance, and\n" +
			"conventions. With --json-schema, emit a JSON Schema for the full --json output\n" +
			"envelopes so an agent can validate the tool's output. Everything is derived\n" +
			"from the tool's own types and data.",
		Example: "  # Triage first: cheap, scannable views\n" +
			"  tskflwctl epic show <id>\n" +
			"  tskflwctl task list -o table -c slug,status,description\n" +
			"  # Full frontmatter only when you need every field:\n" +
			"  tskflwctl schema --json\n" +
			"  tskflwctl schema task\n" +
			"  tskflwctl schema --json-schema",
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"safety": "read-only"},
		ValidArgs:   domain.SchemaKinds(),
		// Pure self-description — no planning repo needed. Overriding the root's
		// resolve() lets an agent run `schema` in any repo to learn the contract
		// (the strongest reason this command exists). Just set up styling.
		PersistentPreRunE: func(*cobra.Command, []string) error { app.setStyle(); return nil },
		RunE: func(_ *cobra.Command, args []string) error {
			if jsonSchema {
				return runJSONSchema(app)
			}
			if len(args) == 0 {
				return runSchemaContract(app)
			}
			return runSchemaKind(app, args[0])
		},
	}
	cmd.Flags().BoolVar(&jsonSchema, "json-schema", false,
		"emit a JSON Schema (Draft 2020-12) for the --json output envelopes")
	return cmd
}

// runJSONSchema emits the Draft 2020-12 JSON Schema for every --json envelope, so
// an agent can validate the tool's machine output against it.
func runJSONSchema(app *App) error {
	schema, err := render.JSONSchema()
	if err != nil {
		return err
	}
	_, err = app.Out.Write(append(schema, '\n'))
	return err
}

// runSchemaContract assembles the global contract from the domain enums/registry
// and the CLI's exit-code table — every value is read from its real source.
func runSchemaContract(app *App) error {
	statuses := make([]render.SchemaStatus, 0, len(domain.AllStatuses()))
	for _, s := range domain.AllStatuses() {
		statuses = append(statuses, render.SchemaStatus{Value: string(s), Active: s.IsActive()})
	}
	buckets := make([]string, 0, len(domain.AllAuditBuckets()))
	for _, b := range domain.AllAuditBuckets() {
		buckets = append(buckets, string(b))
	}
	fields := make([]render.SchemaField, 0)
	for _, name := range domain.KnownTaskFieldNames() {
		fields = append(fields, render.SchemaField{Name: name, Type: domain.FieldType(name)})
	}
	codes := make([]render.SchemaExitCode, 0, len(errCodes))
	for _, e := range errCodes {
		codes = append(codes, render.SchemaExitCode{Code: e.code, Name: e.name})
	}
	c := render.SchemaContract{
		Statuses:     statuses,
		EpicStatuses: domain.AllEpicStatuses(),
		AuditBuckets: buckets,
		TaskFields:   fields,
		ExitCodes:    codes,
		Kinds:        domain.SchemaKinds(),
	}
	if app.JSON {
		return render.SchemaJSON(app.Out, c)
	}
	return render.SchemaHuman(app.Out, app.Style, c)
}

// runSchemaKind assembles the per-kind authoring guidance: the body template (the
// same scaffold `<kind> new` writes), its section names, the authoring fields,
// and the conventions.
func runSchemaKind(app *App, kind string) error {
	fields, err := domain.AuthoringFields(kind)
	if err != nil {
		return err
	}
	body, err := core.ScaffoldBody(kind)
	if err != nil {
		return err
	}
	// Advertise the kind's templates so an agent reading `schema <kind>` discovers
	// `--template` options in-band, without a separate `template list` call.
	tmpls, err := domain.TemplatesFor(kind)
	if err != nil {
		return err
	}
	infos := make([]render.TemplateInfo, len(tmpls))
	for i, t := range tmpls {
		infos[i] = render.TemplateInfo{Kind: kind, Name: t.Name, Description: t.Description}
	}
	ks := render.KindSchema{
		Kind:         kind,
		Sections:     sections(body),
		BodyTemplate: body,
		Fields:       fields,
		Conventions:  domain.Conventions(kind),
		Templates:    infos,
	}
	if app.JSON {
		return render.SchemaKindJSON(app.Out, ks)
	}
	return render.SchemaKindHuman(app.Out, app.Style, ks)
}

func sections(body string) []string {
	matches := sectionRe.FindAllStringSubmatch(body, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, strings.TrimSpace(m[1]))
	}
	return out
}
