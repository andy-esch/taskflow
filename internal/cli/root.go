// Package cli is the primary adapter: the cobra command tree over the core.
// A future TUI is a second primary adapter over the same core.
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/store"
)

// App is the dependency container. It is created empty by NewRootCmd and
// populated lazily in PersistentPreRunE — after flags are parsed, since deps
// (config, service) depend on flags like --chdir.
type App struct {
	Out    io.Writer
	ErrOut io.Writer

	JSON  bool
	Chdir string

	Cfg *config.Config
	Svc *core.Service
}

// NewRootCmd builds the command tree with explicit DI — no package globals.
// All output flows through the injected writers, which makes commands testable.
func NewRootCmd(out, errOut io.Writer) *cobra.Command {
	app := &App{Out: out, ErrOut: errOut}

	root := &cobra.Command{
		Use:           "tskflwctl",
		Short:         "Local-first planning CLI (tasks, epics, audits) over markdown",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Shell completion ('__complete') runs this hook too. Outside a
			// planning repo, resolve() errors — which would abort completion.
			// Stay silent there; completion funcs do their own forgiving
			// discovery (see completion.go).
			if isCompletionCommand(cmd) {
				_ = app.resolve()
				return nil
			}
			return app.resolve()
		},
	}
	root.PersistentFlags().BoolVar(&app.JSON, "json", false, "machine-readable JSON output")
	root.PersistentFlags().StringVarP(&app.Chdir, "chdir", "C", "", "anchor to the planning repo at this path")

	root.AddCommand(newInitCmd(app))
	root.AddCommand(newTaskCmd(app))
	root.AddCommand(newEpicCmd(app))
	root.AddCommand(newAuditCmd(app))
	root.AddCommand(newLintCmd(app))
	return root
}

// resolve discovers the planning repo and constructs the service. Runs once,
// after flag parsing, before any subcommand's RunE (the lazy App shell).
func (a *App) resolve() error {
	start := a.Chdir
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd: %w", err)
		}
		start = wd
	}
	cfg, err := config.Discover(start)
	if err != nil {
		return err
	}
	a.Cfg = cfg
	a.Svc = core.NewService(store.NewFS(cfg.Root))
	return nil
}

// rel renders path relative to the planning root for readable output, falling
// back to the original path.
func (a *App) rel(path string) string {
	if a.Cfg != nil {
		if r, err := filepath.Rel(a.Cfg.Root, path); err == nil {
			return r
		}
	}
	return path
}
