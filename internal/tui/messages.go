package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/andy-esch/taskflow/internal/domain"
)

// listLoadedMsg carries the result of an async entity-list load. kind tags which
// entity tab it belongs to, so a load that finishes after the user has switched
// tabs still lands in the right list. gen is the tab's load generation at fire
// time: list loads run concurrently, so an older load finishing *after* a newer
// one must be dropped, not applied over it.
type listLoadedMsg struct {
	kind     entityKind
	gen      int
	items    []list.Item
	problems []domain.FileProblem
}

// detailMsg carries a lazily-loaded item detail for the right pane. It's applied
// only when (kind, id) still match the active tab's selection — the stale guard —
// and when gen is the latest detail request (two loads for the *same* id aren't
// ordered otherwise).
type detailMsg struct {
	kind    entityKind
	id      string
	gen     int
	content detailContent
}

// detailErrMsg carries a per-item detail-load failure (e.g. an ambiguous
// duplicate slug). It's shown in the detail pane — it must not blank the browser.
type detailErrMsg struct {
	kind entityKind
	id   string
	gen  int
	err  error
}

// tabMsg wraps a list-internal message (notably the async FilterMatchesMsg) with
// the tab it belongs to. Without the tag, a background tab's refilter after a
// reload would be applied to whichever tab happens to be active — blanking the
// filtered background tab and polluting the active one with foreign matches.
type tabMsg struct {
	kind entityKind
	msg  tea.Msg
}

// reloadMsg requests a refresh of every loaded tab (fired by `r` and by the
// fsnotify debounce) — each tab's cursor is preserved by id across the reload.
type reloadMsg struct{}

// fsEventMsg is a raw filesystem change from the watcher. It (re)arms the debounce
// rather than reloading directly, so an editor's save-storm coalesces.
type fsEventMsg struct{}

// debounceMsg fires fsDebounce after an fs event; the model reloads only if gen
// still matches m.dirtyGen (i.e. no newer event re-armed the window).
type debounceMsg struct{ gen int }

// errMsg carries a tab's list-load failure. It's stored per tab (not globally):
// one failing loader must not blank tabs that loaded fine, and concurrent
// reloads must not race on a shared error slot. gen orders it against newer
// loads, like listLoadedMsg.
type errMsg struct {
	kind entityKind
	gen  int
	err  error
}
