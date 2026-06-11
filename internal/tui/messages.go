package tui

import (
	"github.com/charmbracelet/bubbles/list"

	"github.com/andy-esch/taskflow/internal/domain"
)

// listLoadedMsg carries the result of an async entity-list load. kind tags which
// entity tab it belongs to, so a load that finishes after the user has switched
// tabs still lands in the right list.
type listLoadedMsg struct {
	kind     entityKind
	items    []list.Item
	problems []domain.FileProblem
}

// detailMsg carries a lazily-loaded item detail for the right pane. It's applied
// only when (kind, id) still match the active tab's selection — the stale guard.
type detailMsg struct {
	kind    entityKind
	id      string
	content detailContent
}

// detailErrMsg carries a per-item detail-load failure (e.g. an ambiguous
// duplicate slug). It's shown in the detail pane — it must not blank the browser.
type detailErrMsg struct {
	kind entityKind
	id   string
	err  error
}

// reloadMsg requests a refresh of the active tab (fired by `r` now; by fsnotify
// in a later sprint) — the cursor is preserved by id across the reload.
type reloadMsg struct{}

// errMsg carries a fatal async failure (e.g. the initial entity-list load).
type errMsg struct{ err error }
