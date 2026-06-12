package tui

import (
	"github.com/charmbracelet/bubbles/list"

	"github.com/andy-esch/taskflow/internal/domain"
)

// tasksLoadedMsg carries the result of an async task load.
type tasksLoadedMsg struct {
	items    []list.Item
	problems []domain.FileProblem
}

// taskBodyMsg carries a lazily-loaded task detail (frontmatter + body).
type taskBodyMsg struct {
	slug string
	task domain.Task
	body string
}

// bodyErrMsg carries a per-task body-load failure (e.g. an ambiguous duplicate
// slug). It's shown in the detail pane — it must not blank the whole browser.
type bodyErrMsg struct {
	slug string
	err  error
}

// reloadMsg requests a refresh (fired by `r` now; by fsnotify in a later sprint)
// — cursor is preserved by slug across the reload.
type reloadMsg struct{}

// errMsg carries a fatal async failure (e.g. the initial task-list load).
type errMsg struct{ err error }
