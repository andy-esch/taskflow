package tui

import (
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fsnotify/fsnotify"
)

// fsDebounce is the quiet period a change must be followed by before the TUI
// reloads — long enough to coalesce an editor's write/rename/chmod storm into a
// single refresh, short enough to feel live.
const fsDebounce = 200 * time.Millisecond

// watcher wraps an fsnotify watcher over the planning tree. It's created once
// (outside the event loop) and lives for the program's duration; the model holds
// a pointer and drives it via Cmds. nil = live reload unavailable (the browser
// still works; `r` refreshes manually).
type watcher struct {
	fsw *fsnotify.Watcher
}

// newWatcher watches the given leaf dirs (fsnotify is non-recursive); the path
// set comes from the store via the core.Layout port (WatchPaths), so the TUI
// never reconstructs the planning-tree layout itself. Dirs that don't exist are
// skipped — `init` fixes the standard set, and watching missing optional buckets
// isn't worth failing over. (New status/bucket dirs at runtime are out of scope.)
//
// Individual Add failures are tolerated, but if NONE succeeded the watcher can
// never deliver an event, so we return an error (closing the fsnotify watcher)
// rather than a live-looking one — the caller then takes the watchOff branch and
// the footer honestly shows live-reload as unavailable instead of silently never
// firing (inotify limit, FUSE/overlay mount, all dirs absent).
func newWatcher(paths []string) (*watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	added := 0
	for _, d := range paths {
		if err := fsw.Add(d); err == nil { // best-effort: a missing optional dir mustn't kill live reload
			added++
		}
	}
	if added == 0 {
		_ = fsw.Close()
		return nil, errors.New("no watchable directories: live reload unavailable")
	}
	return &watcher{fsw}, nil
}

func (w *watcher) close() error {
	if w == nil || w.fsw == nil {
		return nil
	}
	return w.fsw.Close()
}

// waitForFS blocks until the next filesystem change, returning fsEventMsg. The
// model re-issues it after each event to keep listening. It returns nil (ending
// the listen loop) once the watcher is closed.
func waitForFS(w *watcher) tea.Cmd {
	if w == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case _, ok := <-w.fsw.Events:
			if !ok {
				return nil // watcher closed
			}
			return fsEventMsg{}
		case _, ok := <-w.fsw.Errors:
			if !ok {
				return nil
			}
			return fsEventMsg{} // treat a watch error as a nudge to reload too
		}
	}
}

// debounceTick fires a debounceMsg carrying gen after the quiet period. The model
// reloads only if gen is still current (no newer event arrived meanwhile).
func debounceTick(gen int) tea.Cmd {
	return tea.Tick(fsDebounce, func(time.Time) tea.Msg { return debounceMsg{gen: gen} })
}
