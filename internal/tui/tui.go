package tui

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
)

// Run launches the TUI program over the given workspace. A filesystem watcher
// (live reload) is attached best-effort over the Layout's WatchPaths: if it can't
// start, the browser still runs and `r` refreshes manually — with a footer note
// so the degradation isn't silent. Layout is the narrow on-disk-layout port (the
// CLI injects the FS); reads still flow through the service as tea.Cmds.
func Run(workspace core.Workspace, th design.Theme, opts ...Option) error {
	m := New(workspace.Planning, opts...)
	m.workspace = workspace
	// Resolve the terminal background ONCE, here, before the program starts
	// reading input — querying it mid-program would race Bubble Tea's reader. The
	// same signal drives both the markdown style and the chrome palette: pick the
	// selected theme's background-appropriate palette and apply it before the first
	// render.
	dark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	m.configDark = dark
	// Repopulate the SHARED styles in place with the background-appropriate palette.
	// The list delegates hold this same pointer and deref it per render, so they pick
	// up the swap without being rebuilt — the crux of the per-Model theming. This runs
	// before the first render, so every surface is colored correctly at startup. A
	// future runtime retheme (on a BackgroundColorMsg) could repopulate *m.st the same
	// way, but would ALSO need to refresh the surfaces that render eagerly into cached
	// strings (the dashboard's setSummary rows, the detail pane's rendered body) —
	// those bake their output once and don't re-read *m.st until their next render.
	*m.st = newStyles(th.For(dark))
	atlasTheme := m.atlasTheme
	if atlasTheme.Name == "" {
		atlasTheme = th
	}
	*m.atlasSt = newStyles(atlasTheme.For(dark))
	// Layout is the only capability the browser can run without: no watcher means no live
	// reload, which the footer already reports honestly. Treat a missing one as exactly
	// that rather than panicking on a workspace some future adapter built incompletely.
	m.watchOff = true
	if workspace.Layout != nil {
		if w, err := newWatcher(workspace.Layout.WatchPaths()); err == nil {
			m.watch = w
			m.watchOff = false
		}
	}
	// Alt-screen is declarative in v2 (a View field, set in Model.View), not a
	// program option — so there's no tea.WithAltScreen here anymore.
	final, err := tea.NewProgram(m).Run()
	if fm, ok := final.(Model); ok {
		_ = fm.watch.close()
	} else {
		_ = m.watch.close()
	}
	return err
}
