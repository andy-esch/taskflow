package tui

import (
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"

	"github.com/andy-esch/taskflow/internal/domain"
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

// newWatcher watches the planning tree's leaf dirs (fsnotify is non-recursive).
// Dirs that don't exist are skipped — `init` fixes the standard set, and watching
// missing optional buckets isn't worth failing over.
func newWatcher(root string) (*watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	for _, d := range watchDirs(root) {
		_ = fsw.Add(d) // best-effort: a missing optional dir mustn't kill live reload
	}
	return &watcher{fsw}, nil
}

// watchDirs is the set of directories to watch: the entity parents plus every
// task-status and audit-bucket subdir, so file create/write/move/delete are all
// seen. (New status/bucket dirs at runtime are out of scope — they're fixed by
// `init`.)
func watchDirs(root string) []string {
	tasks := filepath.Join(root, "tasks")
	audits := filepath.Join(root, "audits")
	dirs := []string{filepath.Join(root, "epics"), tasks, audits}
	for _, st := range domain.AllStatuses() {
		dirs = append(dirs, filepath.Join(tasks, st.Dir()))
	}
	for _, b := range domain.AllAuditBuckets() {
		dirs = append(dirs, filepath.Join(audits, b.Dir()))
	}
	return dirs
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
