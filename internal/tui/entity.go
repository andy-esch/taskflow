package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// entityKind identifies a browsable entity. The registry (newEntityTabs) is the
// single place entities are declared, so adding Projects/ADRs/Research later is a
// new entry here — no new keybindings or layout.
type entityKind int

const (
	entityTasks entityKind = iota
	entityEpics
	entityAudits
)

// entityItem is a list row that knows its own stable id (slug / epic id), so the
// model can preserve the cursor and stale-guard detail loads generically.
type entityItem interface {
	list.Item
	id() string
}

// entityTab bundles one entity's static config (name, loaders, delegate via its
// list) with its runtime state (its own list.Model + cursor, loaded flag, load
// problems). Tabs are held by pointer so the value-typed root Model can mutate a
// tab's list in place.
type entityTab struct {
	kind     entityKind
	name     string   // the tab label and the canonical `:` command word
	aliases  []string // shorthands accepted by `:` (e.g. "t", "task")
	list     list.Model
	loadList func(*core.Service) tea.Cmd
	loadItem func(svc *core.Service, id string) tea.Cmd
	loaded   bool
	problems []domain.FileProblem
}

// matches reports whether a typed `:` word selects this tab (canonical name or a
// shorthand). Future entities just declare their own aliases (e.g. adr → "adr").
func (t *entityTab) matches(word string) bool {
	if word == t.name {
		return true
	}
	for _, a := range t.aliases {
		if word == a {
			return true
		}
	}
	return false
}

// newEntityTabs is the entity registry: the ordered set of browsable entities.
func newEntityTabs() []*entityTab {
	mk := func(title string, d list.ItemDelegate) list.Model {
		l := list.New(nil, d, 0, 0)
		l.Title = title
		l.Styles.Title = lipgloss.NewStyle().Bold(true)
		l.SetShowHelp(false)
		l.SetShowStatusBar(false)
		// Built-in fuzzy `/` filter over each item's FilterValue; S2b adds the
		// persistent chip + status views.
		return l
	}
	return []*entityTab{
		{
			kind: entityTasks, name: "tasks", aliases: []string{"t", "task"},
			list: mk("Tasks", taskDelegate{}), loadList: loadTaskList, loadItem: loadTaskDetail,
		},
		{
			kind: entityEpics, name: "epics", aliases: []string{"e", "epic"},
			list: mk("Epics", epicDelegate{}), loadList: loadEpicList, loadItem: loadEpicDetail,
		},
		{
			kind: entityAudits, name: "audits", aliases: []string{"a", "audit"},
			list: mk("Audits", auditDelegate{}), loadList: loadAuditList, loadItem: loadAuditDetail,
		},
	}
}

// indexOfKind returns the tab index for a kind (or -1).
func indexOfKind(tabs []*entityTab, k entityKind) int {
	for i, t := range tabs {
		if t.kind == k {
			return i
		}
	}
	return -1
}
