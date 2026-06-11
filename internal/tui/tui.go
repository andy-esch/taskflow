package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/andy-esch/taskflow/internal/core"
)

// Run launches the TUI program over the given service and planning root. A
// filesystem watcher (live reload) is attached best-effort: if it can't start,
// the browser still runs and `r` refreshes manually.
func Run(svc *core.Service, root string) error {
	m := New(svc, root)
	if w, err := newWatcher(root); err == nil {
		m.watch = w
		defer func() { _ = w.close() }()
	}
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}
