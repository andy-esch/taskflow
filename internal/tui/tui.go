package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/andy-esch/taskflow/internal/core"
)

// Run launches the TUI program over the given service and planning root.
func Run(svc *core.Service, root string) error {
	_, err := tea.NewProgram(New(svc, root), tea.WithAltScreen()).Run()
	return err
}
