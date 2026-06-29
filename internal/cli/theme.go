package cli

import (
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/domain"
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
	cmd.AddCommand(newThemePreviewCmd(app))
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

// newThemePreviewCmd renders a theme's palette — color swatches + a sample bar — for
// the background-appropriate variant. With no arg it previews the active theme.
func newThemePreviewCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "preview [name]",
		Short:       "Preview a theme's palette (color swatches + a sample bar)",
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, args []string) error {
			t := app.Th
			if len(args) == 1 {
				th, ok := design.Lookup(args[0])
				if !ok {
					return fmt.Errorf("%w: unknown theme %q (have: %s)",
						domain.ErrNotFound, args[0], strings.Join(design.Names(), ", "))
				}
				t = th
			}
			// Query the terminal background (an OSC-11 round-trip) only on the HUMAN
			// path with color on. --json stays deterministic (dark) — a machine
			// consumer must not depend on the reviewer's terminal background.
			dark := true
			if !app.JSON && wantColor(app.Color, app.NoColor, app.Out) {
				dark = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
			}
			variant := "dark"
			if !dark {
				variant = "light"
			}
			pal := t.For(dark)
			if app.JSON {
				return render.ThemePreviewJSON(app.Out, t.Name, variant, pal)
			}
			render.ThemePreviewHuman(app.Out, app.Style, t.Name, variant, pal)
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
