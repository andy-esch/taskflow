// Package editor resolves and launches the user's text editor — the shared
// contract behind the CLI's `task edit` and the TUI's `E` (open in $EDITOR), so
// the two faces pick the same editor and split its command the same way.
package editor

import (
	"os"
	"os/exec"
	"strings"
)

// Resolve picks the editor the way every unix tool does: $VISUAL, then $EDITOR,
// then vi as the last-resort default. Always returns a non-empty program.
func Resolve() string {
	for _, env := range []string{"VISUAL", "EDITOR"} {
		if v := strings.TrimSpace(os.Getenv(env)); v != "" {
			return v
		}
	}
	return "vi"
}

// Command builds the *exec.Cmd that opens path in prog, splitting prog on spaces
// so a multi-word editor ("code -w", "emacsclient -nw") keeps its flags. The
// caller wires up stdio — the CLI sets it on the returned cmd; the TUI hands the
// cmd to tea.ExecProcess, which inherits the terminal. A blank prog falls back to
// Resolve so a caller can't build an empty command.
func Command(prog, path string) *exec.Cmd {
	fields := strings.Fields(prog)
	if len(fields) == 0 {
		fields = strings.Fields(Resolve())
	}
	return exec.Command(fields[0], append(fields[1:], path)...)
}
