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
	"github.com/andy-esch/taskflow/internal/configstore"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/spacestore"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/userconfig"
	"github.com/andy-esch/taskflow/internal/workspacestore"
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
	Space    string // registered entry-point label (--space); explicit alternative to Chdir
	Color    string // auto | always | never
	NoColor  bool   // alias for --color=never
	NoInput  bool   // never prompt; missing required input is an error (also TSKFLW_NO_INPUT)
	NoPager  bool   // force paging off (--no-pager)
	Paginate bool   // force paging on, TTY gate permitting (--paginate)
	Theme    string // color theme name (--theme); overrides TSKFLW_THEME + [theme].name

	Style  render.Style
	Th     design.Theme    // the resolved active theme (flag > env > repo config > user config > default)
	Gate   prompt.Gate     // may we prompt? (resolved once, like Style)
	Prompt prompt.Prompter // the human-recovery face (huh on a TTY)
	Cfg    *config.Config
	// User is the home-scope config (theme/pager preferences that belong to a
	// person, not a repo). Always non-nil after setStyle; the zero value means
	// "nothing set here" and every field falls through to the tier below.
	User *userconfig.Config
	// userCfgErr is deferred, not printed at load time: the warning needs the Style
	// (built at the end of setStyle) AND must be suppressed on the completion path,
	// which only the command's own hook knows about. warnPresentation emits it.
	userCfgErr error
	// selectedSpace records the local registry label that selected this invocation's
	// entry point. Empty means ordinary -C/cwd discovery. It is carried onto workspace
	// receipts so an explicit cross-repo write cannot hide how its target was chosen.
	selectedSpace string
	Svc           *core.Service
	// SpaceSvc is the repo-independent application boundary for registry catalog,
	// selection, and mutation use cases.
	SpaceSvc *core.SpaceRegistryService
	// SpaceOverviewSvc is the repo-independent, read-only application service behind
	// `status --all`. It reads through a consumer-owned port so another primary adapter
	// can reuse the same grouping, selection, and failure-isolation rules.
	SpaceOverviewSvc *core.SpaceOverviewService
	// WorkspaceSvc opens arbitrary local planning entry points behind a neutral core
	// boundary. The atlas injects it into the TUI rather than teaching Bubble Tea about
	// config discovery or concrete Markdown storage.
	WorkspaceSvc *core.WorkspaceService
	// ConfigSvc is the framework-free configuration application core shared by
	// Cobra, both TUI contexts, and future adapters.
	ConfigSvc *core.ConfigurationService
	// Fixer/Layout/Linter are the narrow fs/text ports that aren't core use cases:
	// `lint --fix` calls Fixer, the TUI watcher reads Layout, and `lint --links` calls
	// Linter — none route through the Service (see core.Fixer/core.Layout/core.Linter).
	Fixer  core.Fixer
	Layout core.Layout
	Linter core.Linter
}

// setStyle resolves the presentation "face" — output Style (color + width) and the
// input Gate/Prompter — from flags and environment. Called by every command's
// PreRun. The Gate is the single source of truth for "may I prompt?": stdin AND
// stderr must be TTYs, with --json and --no-input both off (the latter also via
// TSKFLW_NO_INPUT). Off a TTY the gate is closed, so the agent/pipeline path never
// blocks.
func (a *App) setStyle() {
	// The home config loads HERE, not in resolve(): init/doctor/completion override
	// PersistentPreRunE and skip resolve() entirely, but they all run setStyle and
	// they all need the theme. Loading it downstream would mean `init` outside a
	// planning repo silently ignored your theme.
	userErr := a.loadUserConfig()
	a.resolveTheme() // flag/env/home now; the repo [theme] folds in once resolve() discovers it
	a.Style = render.NewStyle(wantColor(a.Color, a.NoColor, a.Out)).WithWidth(terminalWidth(a.Out)).WithTrueColor(trueColorCapable(a.Out)).WithPalette(a.Th.Dark)
	noInput := a.NoInput || envEnabled("TSKFLW_NO_INPUT")
	a.Gate = prompt.NewGate(gateOpen(a.JSON, noInput, isTerminalReader(a.In), isTerminal(a.ErrOut)))
	a.Prompt = prompt.NewTTY(a.In, a.ErrOut, a.Th)
	// Recorded, NOT printed here: see warnPresentation.
	a.userCfgErr = userErr
}

// loadUserConfig populates a.User, which is left as an empty (usable) config when
// none exists or it can't be read. The error is returned rather than printed so the
// caller can warn once the Style is built.
func (a *App) loadUserConfig() error {
	uc, err := userconfig.Load()
	a.User = uc
	return err
}

// warnPresentation emits every ⚠ that belongs to the presentation layer, in one
// place. Two constraints force it to be its own step rather than living in setStyle:
// the warnings need the Style (built at the end of setStyle) and, for the theme, the
// repo config that only resolve() discovers; and they must NEVER fire on the shell
// completion path, which only the command's own hook can tell us about.
//
// Every PersistentPreRunE that calls setStyle must call this too. Commands that work
// outside a planning repo should use styleOnlyPreRun rather than hand-rolling the
// pair — `init`, `doctor` and `version` each silently dropped the theme warning by
// hand-rolling it (audit 2026-08-18-multi-space-config-foundation, M2 + L2).
func (a *App) warnPresentation(cmd *cobra.Command) {
	// Completion output is consumed by the shell; a stray ⚠ on stderr is noise at
	// best and corrupts the display at worst. Same rule warnUnknownTheme already had.
	if cmd != nil && isCompletionCommand(cmd) {
		return
	}
	if a.userCfgErr != nil {
		fmt.Fprintf(a.ErrOut, "%s ignoring user config: %v\n", a.Style.Warn("⚠"), a.userCfgErr)
	}
	a.warnUnknownTheme()
}

// styleOnlyPreRun is the PersistentPreRunE for commands that must run ANYWHERE, with
// no planning repo required (`version`, `init`, `schema`): resolve presentation, emit
// its warnings, skip discovery entirely.
func (a *App) styleOnlyPreRun(cmd *cobra.Command, _ []string) error {
	a.setStyle()
	a.warnPresentation(cmd)
	return nil
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
	userName := ""
	if a.User != nil {
		userName = a.User.Theme.Name
	}
	a.Th, _ = design.Lookup(themeName(a.Theme, os.Getenv("TSKFLW_THEME"), cfgName, userName))
}

// themeName resolves the selected theme NAME by precedence — flag > env > repo
// config > home config — trimming each, and "" when none is set (which design.Lookup
// maps to the default). The repo tier beats the home tier so a project can still pin
// a theme for everyone working in it, while the home tier is what a person sets once
// for their own terminal. Pure (no App/env access) so the precedence contract is
// unit-tested directly.
func themeName(flag, env, cfgName, userName string) string {
	if s := strings.TrimSpace(flag); s != "" {
		return s
	}
	if s := strings.TrimSpace(env); s != "" {
		return s
	}
	if s := strings.TrimSpace(cfgName); s != "" {
		return s
	}
	return strings.TrimSpace(userName)
}

// ChromeTheme resolves the active theme for OUT-OF-BAND chrome — fang's styled
// help and error boxes. Those render from cobra's help/error path, which returns
// before PersistentPreRunE ever runs, so App.Th is still the zero default when
// they draw. Without this, chrome silently ignored --theme and every [theme]
// table and always painted the built-in default.
//
// It reuses themeName, so chrome and body share ONE precedence contract and
// cannot drift. Every lookup is best-effort by design: `--help` must keep working
// outside a planning repo, with no home config, and with an unreadable one.
//
// Discovery starts at the process cwd. A run that retargets with -C/--space picks
// up that repo's theme for its BODY but not for chrome; chrome is brand framing,
// not data, and scanning those flags here would duplicate cobra's parser for a
// case that cannot change what any command does.
func ChromeTheme(args []string) design.Theme {
	cfgName := ""
	if start, err := os.Getwd(); err == nil {
		if cfg, cfgErr := config.Discover(start); cfgErr == nil && cfg != nil {
			cfgName = cfg.Theme.Name
		}
	}
	userName := ""
	if uc, ucErr := userconfig.Load(); ucErr == nil && uc != nil {
		userName = uc.Theme.Name
	}
	th, _ := design.Lookup(themeName(themeFlagFrom(args), os.Getenv("TSKFLW_THEME"), cfgName, userName))
	return th
}

// themeFlagFrom scans raw args for --theme, the one flag chrome must honor before
// cobra parses anything. Pure (args in, name out) so the scan is unit-tested
// directly, mirroring useFang's --json scan. `--` ends flag scanning, so a literal
// "--theme" argument after it is not read as the flag.
func themeFlagFrom(args []string) string {
	for i, a := range args {
		if a == "--" {
			return ""
		}
		if name, ok := strings.CutPrefix(a, "--theme="); ok {
			return name
		}
		if a == "--theme" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// NewRootCmd builds the command tree with explicit DI — no package globals.
// All I/O flows through the injected streams, which makes commands testable.
// in is the single stdin owner: it feeds App.In (the prompt gate, prompter, and
// editor) AND the cobra root (cmd.InOrStdin, which resolveBody reads for
// `--body-file -`), so a caller/test injects one reader and every input path
// agrees — production passes os.Stdin.
func NewRootCmd(in io.Reader, out, errOut io.Writer) *cobra.Command {
	spaceAdapter := spacestore.New()
	spaceSvc := core.NewSpaceRegistryService(spaceAdapter)
	app := &App{
		Out: out, ErrOut: errOut, In: in, Th: design.Default(),
		ConfigSvc: core.NewConfigurationService(configstore.New(),
			core.WithConfigurationThemes(design.Names()), core.WithSpaceRegistry(spaceSvc)),
		SpaceSvc:         spaceSvc,
		SpaceOverviewSvc: core.NewSpaceOverviewService(spaceSvc, spaceAdapter),
		WorkspaceSvc:     core.NewWorkspaceService(workspacestore.New()),
	}

	root := &cobra.Command{
		Use:               "tskflwctl",
		Short:             "Local-first planning CLI (tasks, Threads, epics, audits, research) over markdown",
		Version:           versionString(),
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: app.repoPreRun,
	}
	// Cobra's own output (help, usage errors, completion scripts) must follow
	// the injected writers too, or it leaks to os.Stdout/os.Stderr and escapes
	// both tests and callers that capture output.
	root.SetIn(in)
	root.SetOut(out)
	root.SetErr(errOut)
	root.PersistentFlags().BoolVar(&app.JSON, "json", false, "machine-readable JSON output")
	root.PersistentFlags().BoolVar(&app.DryRun, "dry-run", false, "preview the mutation without writing (validation still runs)")
	root.PersistentFlags().StringVarP(&app.Chdir, "chdir", "C", "", "anchor to the planning repo at this path (conflicts with --space)")
	root.PersistentFlags().StringVar(&app.Space, "space", "", "select a registered entry point by label (also TSKFLW_SPACE; conflicts with -C)")
	_ = root.RegisterFlagCompletionFunc("space", completeSpaceIDs(spaceSvc))
	root.PersistentFlags().StringVar(&app.Color, "color", "auto", "colorize output: auto|always|never")
	root.PersistentFlags().BoolVar(&app.NoColor, "no-color", false, "disable colored output (alias for --color=never)")
	root.PersistentFlags().BoolVar(&app.NoInput, "no-input", false, "never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)")
	root.PersistentFlags().BoolVar(&app.NoPager, "no-pager", false, "do not pipe long human output through a pager")
	root.PersistentFlags().BoolVar(&app.Paginate, "paginate", false, "page long human output through $PAGER (on a TTY), even if disabled in config")
	root.PersistentFlags().StringVar(&app.Theme, "theme", "", "color theme name (overrides TSKFLW_THEME and [theme].name in config)")

	root.AddCommand(newInitCmd(app))
	root.AddCommand(newVersionCmd(app))
	root.AddCommand(newStatusCmd(app))
	root.AddCommand(newBoardCmd(app))
	root.AddCommand(newUICmd(app))
	root.AddCommand(newTaskCmd(app))
	root.AddCommand(newThreadCmd(app))
	root.AddCommand(newEpicCmd(app))
	root.AddCommand(newAuditCmd(app))
	root.AddCommand(newResearchCmd(app))
	root.AddCommand(newLintCmd(app))
	root.AddCommand(newConfigCmd(app))
	root.AddCommand(newDoctorCmd(app)) // hidden compatibility forwarding command
	root.AddCommand(newWorkspaceCmd(app))
	root.AddCommand(newSpaceCmd(app))
	root.AddCommand(newSchemaCmd(app))
	root.AddCommand(newTemplateCmd(app))
	root.AddCommand(newThemeCmd(app))
	return root
}

// repoPreRun is the default command hook: resolve presentation and the current
// planning repo. Commands that can work without a current repo override it with
// styleOnlyPreRun or their own conditional hook.
func (a *App) repoPreRun(cmd *cobra.Command, _ []string) error {
	a.setStyle()
	// Shell completion ('__complete') runs this hook too. Outside a planning repo,
	// resolve() errors — which would abort completion. Stay silent there; completion
	// funcs do their own forgiving discovery (see completion.go).
	if isCompletionCommand(cmd) {
		_ = a.resolve()
		return nil
	}
	if err := a.resolve(); err != nil {
		return err
	}
	a.warnLinks()
	a.warnPresentation(cmd)
	return nil
}

// startDir is the single source of the discovery start directory. An explicit --space
// selects the exact registered entry point; -C selects a path; otherwise TSKFLW_SPACE
// may select an entry point before falling back to cwd. resolve() (fatal), config/TUI
// adapters, and completion's planningRoot() (forgiving) share it so the "where do we
// start discovery" contract can't drift between consumers.
func (a *App) startDir() (string, error) {
	spaceFlag := strings.TrimSpace(a.Space)
	if a.Chdir != "" && spaceFlag != "" {
		return "", fmt.Errorf("%w: --space and -C are two answers to one question; pass one", domain.ErrValidation)
	}
	if a.Chdir != "" {
		// An explicit path flag overrides an ambient TSKFLW_SPACE. This follows the
		// ordinary flag-over-environment rule and gives scripts a way to pin a path
		// without having to sanitize the parent environment first.
		a.selectedSpace = ""
		return a.Chdir, nil
	}
	spaceID := spaceFlag
	if spaceID == "" {
		spaceID = strings.TrimSpace(os.Getenv("TSKFLW_SPACE"))
	}
	if spaceID != "" {
		start, err := a.registeredSpaceStart(spaceID)
		if err != nil {
			return "", err
		}
		a.selectedSpace = spaceID
		return start, nil
	}
	a.selectedSpace = ""
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	return wd, nil
}

// wantsSpace distinguishes an explicit registry selection from ordinary best-effort cwd
// discovery. Commands such as template/theme may ignore an ordinary discovery miss, but
// must never swallow an unknown or broken --space and silently fall back to local data.
func (a *App) wantsSpace() bool {
	if strings.TrimSpace(a.Space) != "" {
		return true
	}
	return a.Chdir == "" && strings.TrimSpace(os.Getenv("TSKFLW_SPACE")) != ""
}

// registeredSpaceStart resolves one local registry label into the exact recorded entry
// point. It diagnoses that entry before returning it: explicit selection is also the
// wrong-repo guard, so a missing, unreadable, or identity-mismatched target must fail
// loudly and can never fall back to cwd discovery.
func (a *App) registeredSpaceStart(id string) (string, error) {
	entry, err := a.SpaceSvc.Resolve(id)
	if err != nil {
		return "", err
	}
	return entry.Checkout, nil
}

// resolve discovers the planning repo and constructs the service. Runs once,
// after flag parsing, before any subcommand's RunE (the lazy App shell).
func (a *App) resolve() error {
	start, err := a.startDir()
	if err != nil {
		return err
	}
	return a.resolveFrom(start)
}

// resolveFrom constructs current-repo dependencies from an explicit start path. Splitting
// it from startDir lets `status --all` implement its empty-registry compatibility fallback
// without letting ambient TSKFLW_SPACE turn an all-spaces request into a named selection.
func (a *App) resolveFrom(start string) error {
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
	// One *FS satisfies all the core ports; the Service gets the use-case Store,
	// the adapters get the narrow Fixer/Layout/Linter (see the App field comment).
	fs := store.NewFS(cfg.Root)
	a.Svc = core.NewService(fs)
	a.Fixer = fs
	a.Layout = fs
	a.Linter = fs
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
// env / repo config / home config) didn't match a registered theme — so a typo, or a
// not-yet-supported name like "none", isn't a silent fall-back to the default. Empty
// and "auto" mean "the default" and are intentional, so they don't warn. stderr-only
// (so --json stdout stays clean), and not called on the completion path.
func (a *App) warnUnknownTheme() {
	cfgName := ""
	if a.Cfg != nil {
		cfgName = a.Cfg.Theme.Name
	}
	userName := ""
	if a.User != nil {
		userName = a.User.Theme.Name
	}
	name := themeName(a.Theme, os.Getenv("TSKFLW_THEME"), cfgName, userName)
	if name == "" || strings.EqualFold(name, "auto") {
		return
	}
	if _, ok := design.Lookup(name); !ok {
		fmt.Fprintf(a.ErrOut, "%s unknown theme %q; using %q\n", a.Style.Warn("⚠"), name, a.Th.Name)
	}
}

// markdownStyle resolves the glamour style for `show` body rendering from the ACTIVE
// theme's palette for the terminal background — the theme OWNS its markdown style, so
// a theme can ship its own (e.g. catppuccin uses tokyo-night). Background detection is
// a terminal concern, so it lives here rather than in the render layer. Passed to
// render.RenderBody as a LAZY provider (not called eagerly): HasDarkBackground fires an
// OSC-11 query that can stall, so it runs only when styled markdown is actually
// rendered — never on --raw / --color=never / piped / empty-body.
func (a *App) markdownStyle() string {
	return a.Th.For(lipgloss.HasDarkBackground(os.Stdin, os.Stdout)).Markdown
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
