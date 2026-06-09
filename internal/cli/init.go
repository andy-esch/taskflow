package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/config"
)

func newInitCmd(app *App) *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:         "init",
		Short:       "Scaffold a planning tree (tasks/ epics/ projects/ audits/) + config",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "mutating"},
		Example:     "  tskflwctl init\n  tskflwctl init --path ./planning",
		// init creates a NEW planning repo, so it must NOT require an existing
		// one. A subcommand's own PersistentPreRunE overrides the root's
		// resolve(), so this skips discovery (but still sets up styling).
		// Non-interactive by design → no TTY-hang risk for headless agents.
		PersistentPreRunE: func(*cobra.Command, []string) error { app.setStyle(); return nil },
		RunE: func(_ *cobra.Command, _ []string) error {
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			created, err := config.Init(abs)
			if err != nil {
				return err
			}
			if len(created) == 0 {
				fmt.Fprintf(app.Out, "%s already initialized: %s\n", app.Style.Dim("·"), abs)
				return nil
			}
			fmt.Fprintf(app.Out, "%s initialized %s\n", app.Style.Green("✔"), app.Style.Bold(abs))
			for _, c := range created {
				fmt.Fprintf(app.Out, "  %s %s\n", app.Style.Dim("+"), c)
			}
			fmt.Fprintf(app.Out, "\n%s\n", app.Style.Dim(`→ next: tskflwctl epic new "Title" --description "..."`))
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", ".", "directory to initialize")
	return cmd
}
