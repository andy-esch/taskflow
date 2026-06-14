package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/andy-esch/taskflow/internal/core"
)

// Run launches the TUI program over the given service. A filesystem watcher
// (live reload) is attached best-effort over the store's WatchPaths: if it can't
// start, the browser still runs and `r` refreshes manually — with a footer note
// so the degradation isn't silent.
func Run(svc *core.Service) error {
	m := New(svc)
	if w, err := newWatcher(svc.WatchPaths()); err == nil {
		m.watch = w
		defer func() { _ = w.close() }()
	} else {
		m.watchOff = true
	}
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}
