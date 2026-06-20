// Package prompt is the CLI adapter's human-recovery face: it acquires missing
// required input interactively, but ONLY on a TTY. Nothing here reaches the core —
// commands fill missing params via these helpers, then call core with complete
// params, so the agent/pipeline contract (today's exit codes) is untouched off a
// TTY. The UI libraries are isolated here: list pickers use bubbles/list
// (picker.go), text inputs use huh (tty.go); nothing leaks past this package, and
// tests use the Fake (fake.go).
package prompt

import "errors"

// ErrAborted is returned when the user cancels a prompt (ctrl-c / esc on an
// unfiltered list), so callers can exit cleanly rather than surfacing an error.
var ErrAborted = errors.New("prompt aborted")

// Option is one selectable choice: a human Label and the Value it yields.
type Option struct{ Label, Value string }

// Prompter acquires missing input from a human. Implementations: the huh-backed
// prompter (real TTY) and the Fake (tests). An aborted prompt (ctrl-c/esc)
// returns ErrAborted so callers can exit cleanly. The interface stays minimal —
// it grows a method only when a command needs one (SelectMany lands with the
// tags picker, post-D1).
type Prompter interface {
	SelectOne(title string, opts []Option) (string, error)
	Text(title, placeholder string) (string, error)
}

// Gate decides whether prompting is permitted. It is resolved ONCE (stdin and
// stderr are TTYs, --json off, --no-input unset) and injected, so the human/agent
// face is chosen consistently and the decision is unit-testable in isolation.
type Gate struct{ on bool }

// NewGate returns a gate; on enables prompting.
func NewGate(on bool) Gate { return Gate{on: on} }

// On reports whether prompting is permitted.
func (g Gate) On() bool { return g.on }
