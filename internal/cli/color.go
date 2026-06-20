package cli

import (
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// wantColor decides whether to emit ANSI. Precedence (so agents have
// deterministic control): explicit off (--no-color / --color=never) → explicit
// on (--color=always) → FORCE_COLOR/CLICOLOR_FORCE → NO_COLOR → TTY autodetect.
func wantColor(mode string, noColor bool, out io.Writer) bool {
	if noColor || mode == "never" {
		return false
	}
	if mode == "always" {
		return true
	}
	// auto
	if forceColorEnv() {
		return true
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	return isTerminal(out)
}

// forceColorEnv reports an explicit "force color" request via the common env
// vars (set and not "0").
func forceColorEnv() bool {
	for _, k := range []string{"FORCE_COLOR", "CLICOLOR_FORCE"} {
		if v, ok := os.LookupEnv(k); ok && v != "" && v != "0" {
			return true
		}
	}
	return false
}

// isTerminal reports whether w is a real TTY. term.IsTerminal (the isatty
// ioctl) replaces the old ModeCharDevice stat: /dev/null is a char device but
// not a terminal, so it no longer receives ANSI under --color=auto.
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// isTerminalReader is the stdin counterpart of isTerminal: it reports whether r is
// a real TTY (false for a pipe or a test buffer), so the prompt gate only opens
// when input is genuinely interactive.
func isTerminalReader(r io.Reader) bool {
	f, ok := r.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// envEnabled reports whether an env var is set to a non-empty, non-"0" value — the
// "true-ish" convention used for TSKFLW_NO_INPUT (mirrors forceColorEnv).
func envEnabled(key string) bool {
	v, ok := os.LookupEnv(key)
	return ok && v != "" && v != "0"
}

// gateOpen resolves whether interactive prompting is permitted. Extracted as a
// pure function so the contract is unit-tested directly: prompt ONLY when neither
// --json nor --no-input is set AND both stdin and stderr are real TTYs (clig.dev:
// only prompt when stdin is interactive). Any false closes the gate, so no
// agent/pipeline path can ever block.
func gateOpen(jsonOut, noInput, stdinTTY, stderrTTY bool) bool {
	return !jsonOut && !noInput && stdinTTY && stderrTTY
}

// terminalWidth returns the column width of w's terminal, or 0 when w isn't a
// terminal (piped/redirected/tests) — which the renderer treats as "no limit",
// so scripts get full-width rows.
func terminalWidth(w io.Writer) int {
	// An explicit COLUMNS override wins (usually unset in non-interactive
	// subprocesses, so it doesn't disturb pipes or tests).
	if c := strings.TrimSpace(os.Getenv("COLUMNS")); c != "" {
		if n, err := strconv.Atoi(c); err == nil && n > 0 {
			return n
		}
	}
	f, ok := w.(*os.File)
	if !ok {
		return 0
	}
	width, _, err := term.GetSize(int(f.Fd()))
	if err != nil || width <= 0 {
		return 0
	}
	return width
}
