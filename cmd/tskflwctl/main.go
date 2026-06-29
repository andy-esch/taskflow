// Command tskflwctl is the local-first planning CLI.
package main

import (
	"context"
	"image/color"
	"io"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang"
	"golang.org/x/term"

	"github.com/andy-esch/taskflow/internal/cli"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/theme"
)

func main() {
	root := cli.NewRootCmd(os.Stdin, os.Stdout, os.Stderr)

	// fang wraps the *human* face only. When stderr is not a TTY, or the run is
	// --json, fall through to the original machine path verbatim — so piped/agent
	// output is byte-identical with and without fang, by construction. The whole
	// machine contract (--json error envelope + semantic exit codes 10–14) lives
	// on that fall-through path; fang never touches it.
	if useFang(os.Args[1:], term.IsTerminal(int(os.Stderr.Fd()))) {
		err := fang.Execute(
			context.Background(),
			root,
			fang.WithoutVersion(), // keep our own version string + `version` subcommand
			fang.WithoutManpage(), // manpages come from ./internal/tools/mangen, not the runtime
			fang.WithColorSchemeFunc(repoColorScheme),
			fang.WithErrorHandler(fangErrorHandler),
		)
		if err != nil {
			os.Exit(cli.ExitCode(err)) // preserve semantic exit codes
		}
		return
	}

	// Machine / pipe path — identical to the pre-fang main().
	if err := root.Execute(); err != nil {
		// Under --json, errors are a machine-readable envelope on stderr
		// (stdout stays empty on failure); prose otherwise.
		asJSON, _ := root.PersistentFlags().GetBool("json")
		cli.WriteError(os.Stderr, err, asJSON)
		os.Exit(cli.ExitCode(err)) // semantic codes: 10 not-found … 14 conflict
	}
}

// fangErrorHandler renders errors on the human path. A prompt abort (exit 130)
// is the quiet path — never a styled "error" badge — so it reuses the existing
// writer; everything else gets fang's styled error. --json never reaches here
// (useFang is closed under --json), so the machine envelope is unaffected.
func fangErrorHandler(w io.Writer, styles fang.Styles, err error) {
	if cli.ExitCode(err) == 130 {
		cli.WriteError(os.Stderr, err, false)
		return
	}
	fang.DefaultErrorHandler(w, styles, err)
}

// useFang reports whether to take fang's styled human path. It is closed for any
// machine context: a non-TTY stderr (pipe / redirect / CI) or an explicit --json
// run. Pure (args + tty flag) so the contract gate is unit-testable; `--` ends
// flag scanning so a literal "--json" argument doesn't trip it.
func useFang(args []string, stderrIsTTY bool) bool {
	for _, a := range args {
		if a == "--" {
			break
		}
		if a == "--json" || a == "--json=true" {
			return false
		}
	}
	return stderrIsTTY
}

// repoColorScheme maps fang's help/error palette onto the project's DEFAULT theme,
// so styled help/errors carry the app's identity from the one palette — not hardcoded
// literals or fang's truecolor charmtone default. fang's LightDarkFunc picks each
// token's background-appropriate variant (so help/errors get light vs dark right —
// more than the CLI body, which renders the dark palette); on a non-truecolor
// terminal lipgloss downsamples to the nearest 16-color.
//
// fang renders chrome OUTSIDE the per-command theme resolution, so it always uses
// Default(), never the --theme / [theme] selection. Help/errors are brand chrome,
// not data, so that's an accepted limitation (threading the selected theme into fang
// would be a follow-up).
func repoColorScheme(ld lipgloss.LightDarkFunc) fang.ColorScheme {
	d := design.Default()
	// pick resolves token h to the detected terminal background via fang's func.
	pick := func(h func(design.Palette) design.Hue) color.Color {
		return ld(h(d.Light).Color(), h(d.Dark).Color())
	}
	sem := func(c theme.Color) func(design.Palette) design.Hue {
		return func(p design.Palette) design.Hue { return p.Of(c) }
	}
	var (
		def    = lipgloss.NoColor{} // terminal default foreground (body text)
		accent = pick(func(p design.Palette) design.Hue { return p.Accent })
		green  = pick(sem(theme.ColorGreen))
		yellow = pick(sem(theme.ColorYellow))
		blue   = pick(sem(theme.ColorBlue))
		cyan   = pick(sem(theme.ColorCyan))
		gray   = pick(sem(theme.ColorGray))
		danger = pick(func(p design.Palette) design.Hue { return p.Danger })
		// The error badge's foreground is a fixed high-contrast white — a universal
		// "error" affordance, not a themeable color (the palette has no badge-fg token).
		badgeFg = lipgloss.Color("15")
	)
	return fang.ColorScheme{
		Base:           def,
		Title:          accent, // section headers (USAGE / COMMANDS / FLAGS) carry the theme accent
		Command:        yellow, // subcommand names
		Flag:           green,  // flag names
		Program:        blue,   // the program name in the usage line
		Argument:       def,
		Description:    def,
		QuotedString:   cyan,
		DimmedArgument: gray,
		Comment:        gray,
		FlagDefault:    gray,
		Dash:           gray,
		Help:           gray,
		Codeblock:      gray,
		ErrorHeader:    [2]color.Color{badgeFg, danger}, // bright-white on a danger-red badge
		ErrorDetails:   def,
	}
}
