package tui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/domain"
)

// movedMsg reports a successful lifecycle transition (S4). The model flashes a
// confirmation and fires a reload so the relocated entity shows in its new state.
// to is the destination as a string (a task status or an audit bucket), so one
// message serves every entity's lifecycle.
type movedMsg struct {
	slug string
	to   string
}

// actionErrMsg reports a failed mutation; the model flashes it (red) without
// reloading or corrupting state.
type actionErrMsg struct {
	slug string
	err  error
}

// editedMsg reports a successful inline field edit (SetFields). The model flashes
// it and reloads so the new value shows; unlike movedMsg the task doesn't change
// dirs, so it's just a refresh, cursor preserved by id. value is the written value,
// so the still-open editor can refresh the field it just set.
type editedMsg struct {
	slug  string
	field string
	value string
}

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
	// restore is the cursor id this specific load should re-select (a jump target
	// or a reload's cursor-preservation). Carrying it on the message — stamped with
	// the same gen — means a dropped stale load can't apply a restore meant for a
	// newer one, closing the M6 race where a reload and a jump shared one tab slot.
	restore string
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
