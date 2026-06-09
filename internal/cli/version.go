package cli

import (
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
)

// version is the build version, overridden via
// -ldflags "-X github.com/andy-esch/taskflow/internal/cli.version=...".
var version = "dev"

// versionString prefers the ldflags value, falling back to the module version
// recorded by `go install` (so an installed build still reports something).
func versionString() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return version
}

func newVersionCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "version",
		Short:       "Print the tskflwctl version",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		// Works anywhere — no planning repo needed; just set up styling.
		PersistentPreRunE: func(*cobra.Command, []string) error { app.setStyle(); return nil },
		RunE: func(_ *cobra.Command, _ []string) error {
			if app.JSON {
				return render.VersionJSON(app.Out, versionString())
			}
			render.VersionHuman(app.Out, app.Style, versionString())
			return nil
		},
	}
}
