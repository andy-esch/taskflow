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
		// init creates a NEW planning repo, so it must NOT require an existing
		// one. A subcommand's own PersistentPreRunE overrides the root's
		// resolve(), so this no-op skips the discovery step. (Non-interactive
		// by design → no TTY-hang risk for headless agents.)
		PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
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
				fmt.Fprintf(app.Out, "already initialized: %s\n", abs)
				return nil
			}
			fmt.Fprintf(app.Out, "initialized %s\n", abs)
			for _, c := range created {
				fmt.Fprintf(app.Out, "  + %s\n", c)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", ".", "directory to initialize")
	return cmd
}
