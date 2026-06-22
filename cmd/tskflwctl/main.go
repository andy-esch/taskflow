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
)

func main() {
	root := cli.NewRootCmd(os.Stdout, os.Stderr)

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

// repoColorScheme maps fang's help/error palette onto tskflwctl's own 16-color
// ANSI scheme. internal/cli/render uses the same SGR colors (red 31, green 32,
// yellow 33, blue 34, cyan 36, gray 90), so styled help/errors stay
// terminal-theme-adaptive and visually consistent with `list`/`status`/the TUI
// instead of fang's truecolor charmtone default. Palette indices: 1 red, 2 green,
// 3 yellow, 4 blue, 6 cyan, 8 bright-black, 15 bright-white.
//
// fang's LightDarkFunc parameter is intentionally ignored: ANSI 16-color indices
// are remapped by the terminal's own light/dark theme, so the scheme is already
// background-adaptive without per-slot light/dark selection. (Wire it only if a
// slot ever needs a truecolor value.)
func repoColorScheme(lipgloss.LightDarkFunc) fang.ColorScheme {
	var (
		def    = lipgloss.NoColor{} // terminal default foreground
		red    = lipgloss.Color("1")
		green  = lipgloss.Color("2")
		yellow = lipgloss.Color("3")
		blue   = lipgloss.Color("4")
		cyan   = lipgloss.Color("6")
		gray   = lipgloss.Color("8")
		bright = lipgloss.Color("15")
	)
	return fang.ColorScheme{
		Base:           def,
		Title:          cyan,   // section headers (USAGE / COMMANDS / FLAGS)
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
		ErrorHeader:    [2]color.Color{bright, red}, // bright-white on red badge
		ErrorDetails:   def,
	}
}
