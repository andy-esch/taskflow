package cli

import (
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/wire"
)

// newThemeCmd is the `theme` command group: inspect the color themes. Like
// `version`, it works ANYWHERE — its PreRun sets up styling and best-effort folds
// in the [theme] config when inside a planning repo (so the "active" marker is
// accurate there), without requiring one.
func newThemeCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "theme",
		Short:       "Inspect color themes",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		PersistentPreRunE: func(*cobra.Command, []string) error {
			app.setStyle()
			_ = app.resolve() // best-effort: pick up [theme] config in a repo; harmless outside one
			app.warnUnknownTheme()
			return nil
		},
	}
	cmd.AddCommand(newThemeListCmd(app))
	return cmd
}

// newThemeListCmd lists the registered themes, marking the default + the active one.
func newThemeListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "list",
		Short:       "List the available color themes",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			entries := themeEntries(app.Th.Name)
			if app.JSON {
				return render.ThemesJSON(app.Out, entries)
			}
			render.ThemesHuman(app.Out, app.Style, entries)
			return nil
		},
	}
}

// themeEntries builds the rows for `theme list`: every registered theme, flagged
// active (the one this invocation resolved to) and default.
func themeEntries(active string) []wire.ThemeEntry {
	def := design.Default().Name
	names := design.Names()
	out := make([]wire.ThemeEntry, 0, len(names))
	for _, name := range names {
		out = append(out, wire.ThemeEntry{Name: name, Active: name == active, Default: name == def})
	}
	return out
}
