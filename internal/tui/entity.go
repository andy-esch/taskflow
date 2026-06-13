package tui

import (
	"strings"

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

// entityItem is a list row that knows its own stable id (slug / epic id) and the
// fields it can be sorted by, so the model can preserve the cursor, stale-guard
// detail loads, and reorder lists generically across entities.
type entityItem interface {
	list.Item
	id() string
	sortFields() sortFields
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
	loadList func(*entityTab, *core.Service) tea.Cmd // reads the tab's statusView
	loadItem func(svc *core.Service, id string) tea.Cmd
	loaded   bool
	loadGen  int   // bumped per reload; stale list results/errors are dropped by gen
	loadErr  error // this tab's last list-load failure (nil after a successful load)
	problems []domain.FileProblem

	// S2b list-scoped state (persists per tab across switches/reloads).
	statusView string    // tasks only: "" = working-set, "all", or a status string
	sortCols   []sortKey // the `o`-cycle columns this entity offers
	sortKey    sortKey   // interactive sort column ("o" cycles)
	sortRev    bool      // sort direction toggle ("O")

	restore string // id to re-select after this tab's next load (cursor preservation)
}

// reload re-fires the tab's list loader, passing the tab so the loader can read
// its current statusView (a value-typed Model still mutates via the pointer).
// Each reload bumps the load generation so an older in-flight load can't land
// over this one's result.
func (t *entityTab) reload(svc *core.Service) tea.Cmd {
	t.loadGen++
	return t.loadList(t, svc)
}

// selectByID moves the cursor to the row with the given id, reporting whether it
// was found. It ranges the *visible* items, since list.Select indexes the
// filtered/paginated view — immediately after SetItems on a filtered list the
// refilter is still in flight (VisibleItems is empty), so callers must keep the
// restore pending until this succeeds.
func (t *entityTab) selectByID(id string) bool {
	for i, it := range t.list.VisibleItems() {
		if ei, ok := it.(entityItem); ok && ei.id() == id {
			t.list.Select(i)
			return true
		}
	}
	return false
}

// markReload captures the current cursor id so the next load restores it.
func (t *entityTab) markReload() {
	if it, ok := t.list.SelectedItem().(entityItem); ok {
		t.restore = it.id()
	}
}

// chip is the per-tab state badge shown in the list's title slot: active status
// view, sort column/direction, and any applied `/` filter. Empty (the clean
// default) collapses the title row, giving the list one more visible row.
func (t *entityTab) chip() string {
	var parts []string
	if t.kind == entityTasks && t.statusView != "" {
		parts = append(parts, "view:"+t.statusView)
	}
	filtered := t.list.FilterState() == list.FilterApplied
	// Show the sort badge whenever the order is non-natural — a chosen column OR a
	// reversed default (so `O` on the working set is never a silent reorder). It's
	// suppressed while a filter is applied: the filtered view is ranked by match
	// relevance, so the sort has no visible effect and the badge would mislead.
	if (t.sortKey != sortDefault || t.sortRev) && !filtered {
		parts = append(parts, "sort:"+t.sortKey.label()+sortArrow(t.sortKey, t.sortRev))
	}
	if filtered {
		parts = append(parts, "filter:"+t.list.FilterValue())
	}
	return strings.Join(parts, "  ")
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
	mk := func(d list.ItemDelegate) list.Model {
		l := list.New(nil, d, 0, 0)
		// The title slot carries the S2b state chip (set each render from chip()),
		// not a static "Tasks" label — that duplicated the tab strip. An empty chip
		// collapses the row entirely. TitleBar padding is stripped so the chip (and
		// the `/` filter prompt, which bubbles draws in this same slot) sit on one
		// tight line.
		l.Styles.Title = lipgloss.NewStyle().Bold(true)
		l.Styles.TitleBar = lipgloss.NewStyle()
		l.SetShowHelp(false)
		l.SetShowStatusBar(false)
		// The embedded list must never quit the program itself: its default Quit
		// binding covers q AND esc, so an Esc in list focus (no filter applied)
		// would exit the whole app instead of being a context no-op. Quit layering
		// belongs to the root model (q global, ctrl+c force).
		l.DisableQuitKeybindings()
		return l
	}
	return []*entityTab{
		{
			kind: entityTasks, name: "tasks", aliases: []string{"t", "task"},
			list: mk(taskDelegate{}), loadList: loadTaskList, loadItem: loadTaskDetail,
			sortCols: taskSortCols,
		},
		{
			kind: entityEpics, name: "epics", aliases: []string{"e", "epic"},
			list: mk(epicDelegate{}), loadList: loadEpicList, loadItem: loadEpicDetail,
			sortCols: epicSortCols,
		},
		{
			kind: entityAudits, name: "audits", aliases: []string{"a", "audit"},
			list: mk(auditDelegate{}), loadList: loadAuditList, loadItem: loadAuditDetail,
			sortCols: auditSortCols,
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
