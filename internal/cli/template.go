package cli

import (
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// newTemplateCmd is the body-template discovery surface: `template list` and
// `template show`. Templates are built-in (domain data), so — like `schema` —
// these need no planning repo and run anywhere an agent wants to learn which
// scaffolds `new --template` can use. (Repo-local templates, when epic 22 adds
// them, will resolve a repo on top of the built-ins.)
func newTemplateCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "template",
		Short:       "List and inspect the body scaffolds `new --template` can use",
		Annotations: map[string]string{"safety": "read-only"},
		// Pure self-description (like `schema`): no planning repo needed. Overriding
		// the root's resolve() lets this run in any directory.
		PersistentPreRunE: func(*cobra.Command, []string) error { app.setStyle(); return nil },
	}
	cmd.AddCommand(newTemplateListCmd(app), newTemplateShowCmd(app))
	return cmd
}

func newTemplateListCmd(app *App) *cobra.Command {
	var kind string
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List available body templates (kind, name, description)",
		Example:     "  tskflwctl template list\n  tskflwctl template list --kind audit --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			infos, err := templateInfos(kind)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.TemplatesJSON(app.Out, infos)
			}
			render.TemplatesHuman(app.Out, app.Style, infos)
			return nil
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "", "restrict to one kind (task|epic|audit)")
	_ = cmd.RegisterFlagCompletionFunc("kind", completeKinds)
	return cmd
}

func newTemplateShowCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "show <kind> [name]",
		Short:             `Show a template's rendered body (name defaults to "default")`,
		Example:           "  tskflwctl template show audit security\n  tskflwctl template show task --json",
		Args:              cobra.RangeArgs(1, 2),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: completeTemplateShowArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			kind := args[0]
			name := ""
			if len(args) == 2 {
				name = args[1]
			}
			nt, err := domain.LookupTemplate(kind, name) // validates kind+name → exit 11
			if err != nil {
				return err
			}
			body, err := core.TemplateBody(kind, name)
			if err != nil {
				return err
			}
			info := render.TemplateInfo{Kind: kind, Name: nt.Name, Description: nt.Description}
			if app.JSON {
				return render.TemplateShowJSON(app.Out, info, body)
			}
			render.TemplateShowHuman(app.Out, app.Style, info, body)
			return nil
		},
	}
}

// templateInfos gathers the listable templates for kind (or every kind, in schema
// order, when kind==""). An unknown kind is ErrValidation (exit 11). The slice is
// always non-nil so `--json` emits [] rather than null.
func templateInfos(kind string) ([]render.TemplateInfo, error) {
	kinds := domain.SchemaKinds()
	if kind != "" {
		if _, err := domain.TemplatesFor(kind); err != nil {
			return nil, err
		}
		kinds = []string{kind}
	}
	out := []render.TemplateInfo{}
	for _, k := range kinds {
		ts, _ := domain.TemplatesFor(k) // k is known here
		for _, t := range ts {
			out = append(out, render.TemplateInfo{Kind: k, Name: t.Name, Description: t.Description})
		}
	}
	return out, nil
}
