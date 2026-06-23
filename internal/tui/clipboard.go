package tui

import (
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// copyToClipboard returns a command that puts text on the system clipboard.
//
// It prefers a native OS clipboard utility (pbcopy / wl-copy / xclip / …) because
// that's reliable everywhere a desktop session is reachable — including the many
// terminals that ignore OSC 52 by default (Terminal.app, VS Code, tmux without
// `set-clipboard on`). OSC 52 (tea.SetClipboard) is the fallback: it's the only
// thing that can reach the *local* clipboard from a remote/SSH session, but it
// depends on terminal support, so it's used only when no native tool is present —
// or when the native tool is present but fails (e.g. xclip with no $DISPLAY over
// SSH), in which case we still try OSC 52.
func copyToClipboard(text string) tea.Cmd {
	argv := clipboardArgv()
	if argv == nil {
		return tea.SetClipboard(text)
	}
	return func() tea.Msg {
		c := exec.Command(argv[0], argv[1:]...)
		c.Stdin = strings.NewReader(text)
		if err := c.Run(); err != nil {
			// Native tool present but unusable here — fall back to OSC 52 by
			// returning its message for the program loop to execute.
			return tea.SetClipboard(text)()
		}
		return nil
	}
}

// clipboardArgv returns the command+args for the first OS clipboard utility found
// on PATH, or nil if none is. Order is platform-canonical: pbcopy (macOS), then
// Wayland (wl-copy) before X11 (xclip/xsel), then clip.exe (Windows/WSL).
func clipboardArgv() []string {
	for _, argv := range [][]string{
		{"pbcopy"},
		{"wl-copy"},
		{"xclip", "-selection", "clipboard"},
		{"xsel", "--clipboard", "--input"},
		{"clip.exe"},
	} {
		if _, err := exec.LookPath(argv[0]); err == nil {
			return argv
		}
	}
	return nil
}
