package tui

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
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
	// reading input — querying it mid-program would race Bubble Tea's reader. The
	// same signal drives both the markdown style and the chrome palette: pick the
	// background-appropriate palette for the (default) theme and apply it before
	// the first render. Config-driven theme selection lands in a later task.
	dark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	applyTheme(design.Default().For(dark))
	m.detail.glamStyle = theme.MarkdownStyleFor(dark)
	if w, err := newWatcher(layout.WatchPaths()); err == nil {
		m.watch = w
		defer func() { _ = w.close() }()
	} else {
		m.watchOff = true
	}
	// Alt-screen is declarative in v2 (a View field, set in Model.View), not a
	// program option — so there's no tea.WithAltScreen here anymore.
	_, err := tea.NewProgram(m).Run()
	return err
}
