package cli

import (
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
)

// newTemplateCmd is the body-template discovery surface: `template list` and
// `template show`. Resolution runs through core.Service (like every other
// read/create surface), so when epic 22 adds repo-local templates they layer on
// here with no CLI change. The built-in scaffolds still work repo-less — the
// PersistentPreRunE resolves a planning repo best-effort and falls back to a
// built-in-only service, so these run anywhere `schema` does.
func newTemplateCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "template",
		Short:       "List and inspect the body scaffolds `new --template` can use",
		Annotations: map[string]string{"safety": "read-only"},
		PersistentPreRunE: func(*cobra.Command, []string) error {
			app.setStyle()
			// Best-effort: a planning repo (when present) lets repo-local templates
			// layer on; when absent, fall back to the built-in source so the
			// built-in scaffolds still resolve anywhere.
			if err := app.resolve(); err != nil {
				app.Svc = core.NewBuiltinTemplateService()
			}
			return nil
		},
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
			infos, err := templateInfos(app, kind)
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
	var raw bool
	cmd := &cobra.Command{
		Use:               "show <kind> [name]",
		Short:             `Show a template's body (name defaults to "default"; --raw for the unrendered source)`,
		Example:           "  tskflwctl template show audit security\n  tskflwctl template show task --raw",
		Args:              cobra.RangeArgs(1, 2),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: completeTemplateShowArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			kind := args[0]
			name := ""
			if len(args) == 2 {
				name = args[1]
			}
			// One resolution through the service: ShowTemplate carries the metadata
			// and the raw body, so we render labels here rather than resolving twice.
			ti, raw0, err := app.Svc.ShowTemplate(kind, name) // validates kind+name → exit 11
			if err != nil {
				return err
			}
			body := raw0 // --raw: the unrendered {{placeholder}} source, for forking
			if !raw {
				body = core.RenderLabels(kind, raw0) // preview with <title>/<area> labels
			}
			info := render.TemplateInfo{Kind: ti.Kind, Name: ti.Name, Description: ti.Description}
			if app.JSON {
				return render.TemplateShowJSON(app.Out, info, body)
			}
			render.TemplateShowHuman(app.Out, app.Style, info, body)
			return nil
		},
	}
	cmd.Flags().BoolVar(&raw, "raw", false, "print the unrendered template source ({{placeholders}}) instead of the labelled preview")
	return cmd
}

// templateInfos gathers the listable templates for kind (or every kind, in schema
// order, when kind=="") through the service, then maps the core results into the
// render DTO. An unknown kind is ErrValidation (exit 11). The slice is always
// non-nil so `--json` emits [] rather than null.
func templateInfos(app *App, kind string) ([]render.TemplateInfo, error) {
	infos, err := app.Svc.ListTemplates(kind)
	if err != nil {
		return nil, err
	}
	out := make([]render.TemplateInfo, len(infos))
	for i, t := range infos {
		out[i] = render.TemplateInfo{Kind: t.Kind, Name: t.Name, Description: t.Description}
	}
	return out, nil
}
