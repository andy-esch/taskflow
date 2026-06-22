package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/theme"
)

// Run launches the TUI program over the given service. A filesystem watcher
// (live reload) is attached best-effort over the Layout's WatchPaths: if it can't
// start, the browser still runs and `r` refreshes manually — with a footer note
// so the degradation isn't silent. Layout is the narrow on-disk-layout port (the
// CLI injects the FS); reads still flow through the service as tea.Cmds.
func Run(svc *core.Service, layout core.Layout) error {
	m := New(svc)
	// Resolve the terminal background ONCE, here, before the program starts
	// reading input — querying it mid-program would race Bubble Tea's reader.
	m.detail.glamStyle = theme.MarkdownStyleFor(lipgloss.HasDarkBackground())
	if w, err := newWatcher(layout.WatchPaths()); err == nil {
		m.watch = w
		defer func() { _ = w.close() }()
	} else {
		m.watchOff = true
	}
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}
