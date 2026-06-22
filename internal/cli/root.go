// Package cli is the primary adapter: the cobra command tree over the core.
// The TUI (package tui, launched by `ui`) is a second primary adapter over the
// same core.
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/theme"
)

// App is the dependency container. It is created empty by NewRootCmd and
// populated lazily in PersistentPreRunE — after flags are parsed, since deps
// (config, service) depend on flags like --chdir.
type App struct {
	Out    io.Writer
	ErrOut io.Writer
	In     io.Reader // stdin (for interactive prompts; non-TTY in tests/pipes)

	JSON    bool
	DryRun  bool // preview mutations: full validation, no writes
	Chdir   string
	Color   string // auto | always | never
	NoColor bool   // alias for --color=never
	NoInput bool   // never prompt; missing required input is an error (also TSKFLW_NO_INPUT)

	Style  render.Style
	Gate   prompt.Gate     // may we prompt? (resolved once, like Style)
	Prompt prompt.Prompter // the human-recovery face (huh on a TTY)
	Cfg    *config.Config
	Svc    *core.Service
}

// setStyle resolves the presentation "face" — output Style (color + width) and the
// input Gate/Prompter — from flags and environment. Called by every command's
// PreRun. The Gate is the single source of truth for "may I prompt?": stdin AND
// stderr must be TTYs, with --json and --no-input both off (the latter also via
// TSKFLW_NO_INPUT). Off a TTY the gate is closed, so the agent/pipeline path never
// blocks.
func (a *App) setStyle() {
	a.Style = render.NewStyle(wantColor(a.Color, a.NoColor, a.Out)).WithWidth(terminalWidth(a.Out))
	noInput := a.NoInput || envEnabled("TSKFLW_NO_INPUT")
	a.Gate = prompt.NewGate(gateOpen(a.JSON, noInput, isTerminalReader(a.In), isTerminal(a.ErrOut)))
	a.Prompt = prompt.NewTTY(a.In, a.ErrOut)
}

// NewRootCmd builds the command tree with explicit DI — no package globals.
// All output flows through the injected writers, which makes commands testable.
func NewRootCmd(out, errOut io.Writer) *cobra.Command {
	app := &App{Out: out, ErrOut: errOut, In: os.Stdin}

	root := &cobra.Command{
		Use:           "tskflwctl",
		Short:         "Local-first planning CLI (tasks, epics, audits) over markdown",
		Version:       versionString(),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			app.setStyle()
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
	// Cobra's own output (help, usage errors, completion scripts) must follow
	// the injected writers too, or it leaks to os.Stdout/os.Stderr and escapes
	// both tests and callers that capture output.
	root.SetOut(out)
	root.SetErr(errOut)
	root.PersistentFlags().BoolVar(&app.JSON, "json", false, "machine-readable JSON output")
	root.PersistentFlags().BoolVar(&app.DryRun, "dry-run", false, "preview the mutation without writing (validation still runs)")
	root.PersistentFlags().StringVarP(&app.Chdir, "chdir", "C", "", "anchor to the planning repo at this path")
	root.PersistentFlags().StringVar(&app.Color, "color", "auto", "colorize output: auto|always|never")
	root.PersistentFlags().BoolVar(&app.NoColor, "no-color", false, "disable colored output (alias for --color=never)")
	root.PersistentFlags().BoolVar(&app.NoInput, "no-input", false, "never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)")

	root.AddCommand(newInitCmd(app))
	root.AddCommand(newVersionCmd(app))
	root.AddCommand(newStatusCmd(app))
	root.AddCommand(newUICmd(app))
	root.AddCommand(newTaskCmd(app))
	root.AddCommand(newEpicCmd(app))
	root.AddCommand(newAuditCmd(app))
	root.AddCommand(newLintCmd(app))
	root.AddCommand(newSchemaCmd(app))
	return root
}

// startDir is the single source of the discovery start directory: --chdir if
// given, else the cwd. resolve() (fatal) and completion's planningRoot() (forgiving)
// share it so the "where do we start discovery" contract can't drift between them.
func (a *App) startDir() (string, error) {
	if a.Chdir != "" {
		return a.Chdir, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	return wd, nil
}

// resolve discovers the planning repo and constructs the service. Runs once,
// after flag parsing, before any subcommand's RunE (the lazy App shell).
func (a *App) resolve() error {
	start, err := a.startDir()
	if err != nil {
		return err
	}
	cfg, err := config.Discover(start)
	if err != nil {
		return err
	}
	a.Cfg = cfg
	a.Svc = core.NewService(store.NewFS(cfg.Root))
	return nil
}

// markdownStyle resolves the glamour style for `show` body rendering from the
// terminal background (dracula on dark, light on light). Background detection is
// a terminal concern, so it lives here rather than in the render layer; it's
// called only on the `show` path, where color is on.
func (a *App) markdownStyle() string {
	return theme.MarkdownStyleFor(lipgloss.HasDarkBackground())
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

// linkPath renders an absolute path as a relative display string that is, on a
// TTY, a click-to-open OSC 8 `file://` hyperlink (relative for readability, the
// absolute path in the URL so the terminal can resolve it). Off a TTY / under
// --json it's just the plain relative path, so machine output is unchanged.
func (a *App) linkPath(abs string) string {
	return a.Style.Link(a.rel(abs), "file://"+abs)
}
