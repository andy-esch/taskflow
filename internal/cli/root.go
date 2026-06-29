// Package cli is the primary adapter: the cobra command tree over the core.
// The TUI (package tui, launched by `ui`) is a second primary adapter over the
// same core.
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
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

	JSON     bool
	DryRun   bool // preview mutations: full validation, no writes
	Chdir    string
	Color    string // auto | always | never
	NoColor  bool   // alias for --color=never
	NoInput  bool   // never prompt; missing required input is an error (also TSKFLW_NO_INPUT)
	NoPager  bool   // force paging off (--no-pager)
	Paginate bool   // force paging on, TTY gate permitting (--paginate)
	Theme    string // color theme name (--theme); overrides TSKFLW_THEME + [theme].name

	Style  render.Style
	Th     design.Theme    // the resolved active theme (flag > env > config > default)
	Gate   prompt.Gate     // may we prompt? (resolved once, like Style)
	Prompt prompt.Prompter // the human-recovery face (huh on a TTY)
	Cfg    *config.Config
	Svc    *core.Service
	// Fixer/Layout are the narrow fs/text ports that aren't core use cases:
	// `lint --fix` calls Fixer directly and the TUI watcher reads Layout, so
	// neither routes through the Service (see core.Fixer/core.Layout).
	Fixer  core.Fixer
	Layout core.Layout
}

// setStyle resolves the presentation "face" — output Style (color + width) and the
// input Gate/Prompter — from flags and environment. Called by every command's
// PreRun. The Gate is the single source of truth for "may I prompt?": stdin AND
// stderr must be TTYs, with --json and --no-input both off (the latter also via
// TSKFLW_NO_INPUT). Off a TTY the gate is closed, so the agent/pipeline path never
// blocks.
func (a *App) setStyle() {
	a.resolveTheme() // flag/env now; the [theme] config folds in once resolve() discovers it
	a.Style = render.NewStyle(wantColor(a.Color, a.NoColor, a.Out)).WithWidth(terminalWidth(a.Out)).WithPalette(a.Th.Dark)
	noInput := a.NoInput || envEnabled("TSKFLW_NO_INPUT")
	a.Gate = prompt.NewGate(gateOpen(a.JSON, noInput, isTerminalReader(a.In), isTerminal(a.ErrOut)))
	a.Prompt = prompt.NewTTY(a.In, a.ErrOut, a.Th)
}

// resolveTheme picks the active color theme by precedence: --theme flag >
// TSKFLW_THEME env > [theme].name in config > the built-in default. An unknown name
// degrades to the default (design.Lookup never errors), so a typo can't break a
// command. Cfg may be nil (pre-discovery): config is simply skipped. Called once in
// setStyle (flag/env) and again in resolve once Cfg is known.
func (a *App) resolveTheme() {
	cfgName := ""
	if a.Cfg != nil {
		cfgName = a.Cfg.Theme.Name
	}
	a.Th, _ = design.Lookup(themeName(a.Theme, os.Getenv("TSKFLW_THEME"), cfgName))
}

// themeName resolves the selected theme NAME by precedence — flag > env > config —
// trimming each, and "" when none is set (which design.Lookup maps to the default).
// Pure (no App/env access) so the precedence contract is unit-tested directly.
func themeName(flag, env, cfgName string) string {
	if s := strings.TrimSpace(flag); s != "" {
		return s
	}
	if s := strings.TrimSpace(env); s != "" {
		return s
	}
	return strings.TrimSpace(cfgName)
}

// NewRootCmd builds the command tree with explicit DI — no package globals.
// All I/O flows through the injected streams, which makes commands testable.
// in is the single stdin owner: it feeds App.In (the prompt gate, prompter, and
// editor) AND the cobra root (cmd.InOrStdin, which resolveBody reads for
// `--body-file -`), so a caller/test injects one reader and every input path
// agrees — production passes os.Stdin.
func NewRootCmd(in io.Reader, out, errOut io.Writer) *cobra.Command {
	app := &App{Out: out, ErrOut: errOut, In: in, Th: design.Default()}

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
			if err := app.resolve(); err != nil {
				return err
			}
			app.warnLinks()
			app.warnUnknownTheme()
			return nil
		},
	}
	// Cobra's own output (help, usage errors, completion scripts) must follow
	// the injected writers too, or it leaks to os.Stdout/os.Stderr and escapes
	// both tests and callers that capture output.
	root.SetIn(in)
	root.SetOut(out)
	root.SetErr(errOut)
	root.PersistentFlags().BoolVar(&app.JSON, "json", false, "machine-readable JSON output")
	root.PersistentFlags().BoolVar(&app.DryRun, "dry-run", false, "preview the mutation without writing (validation still runs)")
	root.PersistentFlags().StringVarP(&app.Chdir, "chdir", "C", "", "anchor to the planning repo at this path")
	root.PersistentFlags().StringVar(&app.Color, "color", "auto", "colorize output: auto|always|never")
	root.PersistentFlags().BoolVar(&app.NoColor, "no-color", false, "disable colored output (alias for --color=never)")
	root.PersistentFlags().BoolVar(&app.NoInput, "no-input", false, "never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)")
	root.PersistentFlags().BoolVar(&app.NoPager, "no-pager", false, "do not pipe long human output through a pager")
	root.PersistentFlags().BoolVar(&app.Paginate, "paginate", false, "page long human output through $PAGER (on a TTY), even if disabled in config")
	root.PersistentFlags().StringVar(&app.Theme, "theme", "", "color theme name (overrides TSKFLW_THEME and [theme].name in config)")

	root.AddCommand(newInitCmd(app))
	root.AddCommand(newVersionCmd(app))
	root.AddCommand(newStatusCmd(app))
	root.AddCommand(newUICmd(app))
	root.AddCommand(newTaskCmd(app))
	root.AddCommand(newEpicCmd(app))
	root.AddCommand(newAuditCmd(app))
	root.AddCommand(newLintCmd(app))
	root.AddCommand(newDoctorCmd(app))
	root.AddCommand(newSchemaCmd(app))
	root.AddCommand(newTemplateCmd(app))
	root.AddCommand(newThemeCmd(app))
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
	// The [theme].name can now participate in selection (lowest precedence, so this
	// only changes anything when neither --theme nor TSKFLW_THEME pinned it). Re-skin
	// the output Style + prompter so a config-selected theme takes effect.
	a.resolveTheme()
	a.Style = a.Style.WithPalette(a.Th.Dark)
	a.Prompt = prompt.NewTTY(a.In, a.ErrOut, a.Th)
	// One *FS satisfies all three core ports; the Service gets the use-case Store,
	// the adapters get the narrow Fixer/Layout (see the App field comment).
	fs := store.NewFS(cfg.Root)
	a.Svc = core.NewService(fs)
	a.Fixer = fs
	a.Layout = fs
	return nil
}

// warnLinks emits the ambient linkback-integrity warnings — one ⚠ per finding to
// stderr, so --json stdout stays clean and a pipe consuming data is unaffected.
// Silent when the links are consistent (or absent); suppressed entirely by
// TSKFLW_NO_LINK_WARN. The `doctor` command reports the same findings explicitly,
// so its own PreRunE overrides the root hook that calls this.
func (a *App) warnLinks() {
	if envEnabled("TSKFLW_NO_LINK_WARN") {
		return
	}
	for _, p := range config.CheckLinks(a.Cfg) {
		fmt.Fprintf(a.ErrOut, "%s %s\n", a.Style.Warn("⚠"), p.Message)
	}
}

// warnUnknownTheme emits one ⚠ to stderr when an explicitly-set theme name (flag /
// env / config) didn't match a registered theme — so a typo, or a not-yet-supported
// name like "none", isn't a silent fall-back to the default. Empty and "auto" mean
// "the default" and are intentional, so they don't warn. stderr-only (so --json
// stdout stays clean), and not called on the completion path.
func (a *App) warnUnknownTheme() {
	cfgName := ""
	if a.Cfg != nil {
		cfgName = a.Cfg.Theme.Name
	}
	name := themeName(a.Theme, os.Getenv("TSKFLW_THEME"), cfgName)
	if name == "" || strings.EqualFold(name, "auto") {
		return
	}
	if _, ok := design.Lookup(name); !ok {
		fmt.Fprintf(a.ErrOut, "%s unknown theme %q; using %q\n", a.Style.Warn("⚠"), name, a.Th.Name)
	}
}

// markdownStyle resolves the glamour style for `show` body rendering from the
// terminal background (dracula on dark, light on light). Background detection is a
// terminal concern, so it lives here rather than in the render layer. It is passed
// to render.RenderBody as a LAZY provider (not called eagerly): HasDarkBackground
// fires an OSC-11 terminal query that can stall on terminals that don't answer, so
// it must run only when styled markdown is actually rendered — never on
// --raw / --color=never / piped / empty-body, where the result would be discarded.
func (a *App) markdownStyle() string {
	return theme.MarkdownStyleFor(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
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
