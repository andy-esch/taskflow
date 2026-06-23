package cli

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// paged runs render with the writer human output should go to. On the gated human
// path that writer is the stdin of a spawned pager (its stdout is a.Out, the TTY);
// otherwise — and on any spawn failure — it is a.Out unchanged, so the machine path
// is byte-identical to no pager at all. The pager is closed and waited on after
// render returns; quitting the pager early (a broken pipe) is not an error.
func (a *App) paged(render func(io.Writer) error) error {
	if !a.pagerActive() {
		return render(a.Out)
	}
	prog := strings.TrimSpace(a.pagerProgram())
	// An empty pager or "cat" means "don't page" (matching git's PAGER="" / cat).
	if prog == "" || prog == "cat" {
		return render(a.Out)
	}
	return a.pipeToPager(prog, render)
}

// pipeToPager spawns prog (via sh -c, like git, so it can carry its own
// flags/quoting), streams render's output into its stdin, and waits for it. The
// pager's stdout is a.Out (the TTY in production). Any spawn failure degrades to
// writing a.Out directly so output is never silenced; quitting the pager early
// (a broken pipe) is expected, not an error.
func (a *App) pipeToPager(prog string, render func(io.Writer) error) error {
	cmd := exec.Command("sh", "-c", prog)
	cmd.Stdout = a.Out // the TTY (the gate guarantees a.Out is a real *os.File)
	cmd.Stderr = a.ErrOut
	cmd.Env = pagerEnv()
	w, err := cmd.StdinPipe()
	if err != nil {
		return render(a.Out)
	}
	if err := cmd.Start(); err != nil {
		// No pager on PATH (or unrunnable): degrade to direct output, never silence.
		return render(a.Out)
	}

	renderErr := render(w)
	_ = w.Close()  // let the pager drain and exit
	_ = cmd.Wait() // the pager's own exit status isn't ours to surface
	if renderErr != nil && !isBrokenPipe(renderErr) {
		return renderErr
	}
	return nil
}

// pagerActive resolves whether to page human output. The hard gate comes first —
// the machine contract: only a real stdout TTY, never under --json or
// --no-input/TSKFLW_NO_INPUT — so off a TTY or for agents this is always false and
// output is emitted raw. pagerWanted then applies the on/off precedence.
func (a *App) pagerActive() bool {
	if a.JSON || a.NoInput || envEnabled("TSKFLW_NO_INPUT") || !isTerminal(a.Out) {
		return false
	}
	return a.pagerWanted()
}

// pagerWanted is the on/off decision without the TTY/machine gate (so it is unit-
// testable on its own). Precedence mirrors git: --no-pager (off) > --paginate (on)
// > [pager].enabled > default on.
func (a *App) pagerWanted() bool {
	switch {
	case a.NoPager:
		return false
	case a.Paginate:
		return true
	default:
		return a.Cfg == nil || a.Cfg.Pager.Enabled == nil || *a.Cfg.Pager.Enabled
	}
}

// pagerProgram resolves the pager command string, mirroring git's
// GIT_PAGER > core.pager > PAGER > less: TSKFLW_PAGER > [pager].command > $PAGER >
// "less -FRX" (F = skip the pager if it fits one screen, R = keep ANSI colors,
// X = leave the scrollback intact).
func (a *App) pagerProgram() string {
	if v := strings.TrimSpace(os.Getenv("TSKFLW_PAGER")); v != "" {
		return v
	}
	if a.Cfg != nil {
		if v := strings.TrimSpace(a.Cfg.Pager.Command); v != "" {
			return v
		}
	}
	if v := strings.TrimSpace(os.Getenv("PAGER")); v != "" {
		return v
	}
	return "less -FRX"
}

// pagerEnv passes the parent environment to the pager, defaulting LESS/LV the way
// git does so a bare `less`/`lv` (e.g. PAGER=less) still keeps colors and quits on
// a single screen without clobbering scrollback.
func pagerEnv() []string {
	env := os.Environ()
	if _, ok := os.LookupEnv("LESS"); !ok {
		env = append(env, "LESS=FRX")
	}
	if _, ok := os.LookupEnv("LV"); !ok {
		env = append(env, "LV=-c")
	}
	return env
}

// isBrokenPipe reports whether err is the pager closing its read end (the user quit
// before all output was written) — expected, not a failure.
func isBrokenPipe(err error) bool {
	return errors.Is(err, syscall.EPIPE) || errors.Is(err, io.ErrClosedPipe)
}
