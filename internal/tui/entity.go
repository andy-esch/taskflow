package tui

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// entityKind identifies a browsable entity. The registry (newEntityTabs) is the
// single place entities are declared, so adding Projects/ADRs/Research later is a
// new entry here. Read/browse is keybinding-free; lifecycle (the `m` menu and `:`
// verbs) is declared per entity via its transition table + applyMove, so an entity
// that wants lifecycle wires it in the registry rather than editing the reducer.
type entityKind int

const (
	entityTasks entityKind = iota
	entityEpics
	entityThreads
	entityAudits
	entityResearch
)

// Non-tab screen sentinels let context surfaces such as `?` render screen-specific
// notes without pretending either composite surface is an entity tab.
const (
	entityDashboard entityKind = -1
	entityAtlas     entityKind = -2
)

// overviewName is the display name of the landing overview view — its window
// title, tab-strip label, `:` command word, and help heading — in one place so
// the label moves with a single edit.
const overviewName = "overview"

// atlasName is the canonical screen/command label for the cross-space navigator.
const atlasName = "atlas"

// entityRef keeps durable identity separate from human-facing text. key is the
// canonical reference accepted by the entity's store; label is the slug/id users
// see and copy. Keeping both in the registry contract prevents a future entity
// (notably Threads) from accidentally using a mutable or ambiguous label for
// selection, detail reads, navigation, or saved workspace state.
type entityRef struct {
	key   string
	label string
}

func (r entityRef) empty() bool { return r.key == "" }

// entityItem is a list row that knows its canonical reference, visible label,
// and sortable fields, so the model can preserve the cursor, stale-guard detail
// loads, and reorder lists generically across entities. displayLabel prefixes a
// short identity hint when duplicate labels coexist in the loaded result. Prefixing
// keeps the only distinguishing text visible when a long shared label is truncated.
type entityItem interface {
	list.Item
	ref() entityRef
	displayLabel() string
	hasIdentityHint() bool
	path() string // the entity's on-disk file path (for the clipboard yank)
	sortFields() sortFields
}

// validateEntityItems is the registry's last line of defence against an invalid
// portable adapter. A canonical key must be non-empty and unique within a loaded
// result; otherwise selection would mean "nothing" or silently choose the first
// matching row. Local stores enforce this more deeply, but future adapters must
// fail just as loudly instead of relying on filesystem invariants they do not have.
func validateEntityItems(items []list.Item) error {
	seen := make(map[string]string, len(items))
	for _, raw := range items {
		item, ok := raw.(entityItem)
		if !ok {
			return fmt.Errorf("entity registry received unsupported row %T", raw)
		}
		ref := item.ref()
		if ref.key == "" {
			return fmt.Errorf("entity %q has no canonical identity", ref.label)
		}
		if prior, ok := seen[ref.key]; ok {
			return fmt.Errorf("canonical identity %q is shared by %q and %q", ref.key, prior, ref.label)
		}
		seen[ref.key] = ref.label
	}
	return nil
}

// duplicateIdentityHints returns the shortest deterministic stable-ID prefix
// (six characters minimum) needed to distinguish rows that share a visible
// label. Unique labels stay unannotated; a full key is used only in the unlikely
// case that shorter prefixes still collide.
func duplicateIdentityHints(refs []entityRef) map[string]string {
	byLabel := make(map[string][]string)
	for _, ref := range refs {
		byLabel[ref.label] = append(byLabel[ref.label], ref.key)
	}
	hints := make(map[string]string)
	for _, keys := range byLabel {
		if len(keys) < 2 {
			continue
		}
		for keyIndex, key := range keys {
			found := false
			for n := min(6, len(key)); n <= len(key); n++ {
				prefix := key[:n]
				unique := true
				for otherIndex, other := range keys {
					if otherIndex != keyIndex && strings.HasPrefix(other, prefix) {
						unique = false
						break
					}
				}
				if unique {
					hints[key] = prefix
					found = true
					break
				}
			}
			// A non-filesystem adapter may supply unequal key lengths where one key is
			// a strict prefix of another. The shorter key has no unique prefix, but its
			// full value is still the most useful explicit discriminator. Equal or empty
			// keys remain an adapter-contract violation and are handled separately.
			if !found {
				hints[key] = key
			}
		}
	}
	return hints
}

func labelWithIdentityHint(label, hint string) string {
	if hint == "" {
		return label
	}
	return "[" + hint + "] " + label
}

// lifecycleItem is an entityItem with a mutable lifecycle state (a task's status,
// an audit's bucket, an epic's status). The action menu reads it to drop the no-op
// transition; every entity now implements it via lifecycleState(), so the reducer
// asks for the state generically instead of switching on concrete item types. Epics
// declare epicTransitions/moveEpic — they differ only in that the move rewrites a
// frontmatter FIELD (status:) rather than moving the file between directories.
type lifecycleItem interface {
	entityItem
	lifecycleState() string
}

// entityTab bundles one entity's static config (name, loaders, delegate via its
// list) with its runtime state (its own list.Model + cursor, loaded flag, load
// problems). Tabs are held by pointer so the value-typed root Model can mutate a
// tab's list in place.
type entityTab struct {
	kind    entityKind
	name    string   // the tab label and the canonical `:` command word
	aliases []string // shorthands accepted by `:` (e.g. "t", "task")

	// View axis (static config): the s/S cycle + leading `:` words for this
	// entity's status/bucket filter. nil for entities with no axis (epics).
	viewAxis    []statusView
	viewAliases []statusView // extra `:` words outside the s/S cycle

	list     list.Model
	loadList func(*entityTab, *core.Service) tea.Cmd // reads the tab's statusView
	loadItem func(svc *core.Service, id string) tea.Cmd
	loaded   bool
	// loadOrder is the items exactly as the loader produced them, before any interactive
	// sort. Every sort is computed FROM this rather than from the list's current slice:
	// sortItems reorders in place, so sorting the already-sorted slice made sorts
	// cumulative and left sortDefault — which by definition does not reorder — unable to
	// restore the loader's order. Cycling `o` all the way round then silently kept
	// whichever column ran last.
	loadOrder []list.Item
	loadGen   int   // bumped per reload; stale list results/errors are dropped by gen
	loadErr   error // this tab's last list-load failure (nil after a successful load)
	problems  []domain.FileProblem
	// Thread list reads carry repository-wide graph/read diagnostics that do not
	// belong to any one row. They live on the registry tab — not in a parallel
	// Thread state machine — and are nil for every other entity.
	threadDiagnostics *threadListDiagnostics

	// Lifecycle (registry-driven, S4/M10): the transitions this entity offers and
	// the move that applies one. Tasks declare status transitions (svc.Move); audits
	// declare bucket transitions (svc.MoveAudit); epics declare epicTransitions/moveEpic
	// (the move rewrites the status: frontmatter FIELD rather than relocating the file).
	// The `m` menu and `:` verbs read these off the active tab, so lifecycle is no
	// longer task-only plumbing in the reducer. nil transitions ⇒ no `m`/`:`-verb actions.
	transitions []transition
	applyMove   func(svc *core.Service, ref entityRef, tr transition) tea.Cmd

	// S2b list-scoped state (persists per tab across switches/reloads).
	statusView  string    // view axis: "" = default, "all", a task status, or an audit bucket
	sortCols    []sortKey // the `o`-cycle columns this entity offers
	sortKey     sortKey   // interactive sort column ("o" cycles)
	sortRev     bool      // sort direction toggle ("O")
	filterExact bool      // false = fuzzy (default), true = substring; toggled session-wide by "F"

	// restore is the canonical cursor key the next landing load (or its async refilter) should
	// select — a jumpTo target or a reload's cursor-preservation. restoreGen stamps
	// the loadGen it belongs to, so a newer reload supersedes a stale target and an
	// old filter-match callback (handleTabMsg) can't apply it. The id also rides on
	// each load's listLoadedMsg (gen-safe), so a dropped stale load never applies a
	// restore meant for another — the single-slot race M6 flagged. restoreWiden lets
	// an explicit navigation target retry once in :all when a tab's default view
	// hides it; ordinary cursor preservation never widens behind the user.
	restore      entityRef
	restoreGen   int
	restoreWiden bool
}

// reload re-fires the tab's list loader with the canonical cursor key to restore afterward,
// passing the tab so the loader can read its current statusView (a value-typed
// Model still mutates via the pointer). Each reload bumps the load generation so an
// older in-flight load can't land over this one's result, records restoreKey as the
// tab's pending intent (so a concurrent markReload carries it forward, not the
// stale cursor), and stamps it onto the load's message for the gen-safe consumer.
func (t *entityTab) reload(svc *core.Service, restore entityRef) tea.Cmd {
	// Keep the explicit-navigation fallback only while reloads carry the same pending target.
	// A watcher firing during the landing must not discard it, while an unrelated
	// navigation must not inherit it.
	widenOnMissing := t.restoreWiden && !restore.empty() && t.restore.key == restore.key
	t.loadGen++
	t.restore, t.restoreGen = restore, t.loadGen
	t.restoreWiden = widenOnMissing
	load := t.loadList(t, svc)
	stamped := func() tea.Msg {
		msg := load()
		if lm, ok := msg.(listLoadedMsg); ok {
			lm.restore = restore
			lm.widenOnMissing = widenOnMissing
			return lm
		}
		return msg // errMsg passes through unchanged
	}
	request := readRequest{surface: readEntityList, kind: t.kind, gen: t.loadGen}
	return withReadConflictRetry(request, stamped)
}

// selectByKey moves the cursor to the row with the given canonical key, reporting whether it
// was found. It ranges the *visible* items, since list.Select indexes the
// filtered/paginated view — immediately after SetItems on a filtered list the
// refilter is still in flight (VisibleItems is empty), so callers must keep the
// restore pending until this succeeds.
func (t *entityTab) selectByKey(key string) bool {
	for i, it := range t.list.VisibleItems() {
		if ei, ok := it.(entityItem); ok && ei.ref().key == key {
			t.list.Select(i)
			return true
		}
	}
	return false
}

// markReload returns the id a reload should re-select. A pending navigation target
// (a jumpTo whose load hasn't landed) outranks the current cursor, so a background
// reload firing mid-jump carries the jump target forward instead of yanking the
// cursor back to where it was — the M6 fix for the reload/jump race. With nothing
// pending it captures the current cursor (the ordinary reload-preserves-cursor case).
func (t *entityTab) markReload() entityRef {
	if !t.restore.empty() {
		return t.restore
	}
	if it, ok := t.list.SelectedItem().(entityItem); ok {
		return it.ref()
	}
	return entityRef{}
}

// viewFor maps a `:` word to a view value on this tab's axis.
func (t *entityTab) viewFor(word string) (string, bool) {
	return viewFor(t.viewAxis, t.viewAliases, word)
}

// viewWords is this tab's `:` view vocabulary (axis cycle + aliases).
func (t *entityTab) viewWords() []string {
	return viewWords(t.viewAxis, t.viewAliases)
}

// chip is the per-tab state badge shown in the list's title slot: active status/
// bucket view, sort column/direction, and any applied `/` filter. Empty (the clean
// default) collapses the title row, giving the list one more visible row.
func (t *entityTab) chip() string {
	var parts []string
	// Only axis-bearing tabs (tasks, audits, epics) ever set statusView non-empty, so
	// the non-default view shows here regardless of entity; the default ("") is silent.
	if t.statusView != "" {
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
		// Default fuzzy is silent here (like view:/sort:); only the non-default
		// substring mode annotates the badge, so an active exact filter is never a
		// silent surprise. The prompt (below) always names the mode while typing.
		label := "filter:"
		if t.filterExact {
			label = "filter(exact):"
		}
		parts = append(parts, label+t.list.FilterValue())
	}
	return strings.Join(parts, "  ")
}

// filterPrompt is the bubbles/list filter-input prefix, so the active mode is
// visible while typing a filter ("filter (fuzzy): " / "filter (exact): ").
func filterPrompt(exact bool) string {
	if exact {
		return "filter (exact): "
	}
	return "filter (fuzzy): "
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

// moveTask applies a task status transition off the event loop, reporting success
// (movedMsg → flash + reload) or failure (actionErrMsg → flash, no reload).
func moveTask(svc *core.Service, ref entityRef, tr transition) tea.Cmd {
	return func() tea.Msg {
		// force=false: a TUI completion is held to the same acceptance-criteria gate as
		// the CLI's, and the refusal surfaces as the action's error flash.
		receipt, err := svc.Move(ref.key, domain.Status(tr.to), false, core.TaskLifecycleOverrideNone)
		if err != nil {
			var committed *core.TaskLifecycleMutationFailure
			if errors.As(err, &committed) {
				return movedMsg{ref: ref, to: tr.to, lifecycle: &committed.Receipt, warning: err}
			}
			return actionErrMsg{ref: ref, err: err}
		}
		return movedMsg{ref: ref, to: tr.to, lifecycle: &receipt}
	}
}

// deferTaskCmd defers a task with an optional revisit ("snooze until") date — the
// TUI face of `task defer [--until]`. An empty date parks the task indefinitely
// (a plain Move); a non-empty one also records revisit_at. Mirrors moveTask's
// success/failure reporting (movedMsg → flash + reload, actionErrMsg → flash).
func deferTaskCmd(svc *core.Service, ref entityRef, revisit string) tea.Cmd {
	return func() tea.Msg {
		receipt, err := svc.DeferTask(ref.key, revisit, false)
		if err != nil {
			var committed *core.TaskLifecycleMutationFailure
			if errors.As(err, &committed) {
				return movedMsg{ref: ref, to: string(domain.StatusDeferred), revisit: revisit, lifecycle: &committed.Receipt, warning: err}
			}
			return actionErrMsg{ref: ref, err: err}
		}
		return movedMsg{ref: ref, to: string(domain.StatusDeferred), revisit: revisit, lifecycle: &receipt}
	}
}

// moveAudit applies an audit bucket transition (close/reopen/defer). The store
// refuses closing/deferring an audit with still-open findings (M4); that surfaces
// as an actionErrMsg (red flash, no move), matching the CLI.
func moveAudit(svc *core.Service, ref entityRef, tr transition) tea.Cmd {
	return func() tea.Msg {
		if _, err := svc.MoveAudit(ref.key, domain.AuditBucket(tr.to), false); err != nil {
			return actionErrMsg{ref: ref, err: err}
		}
		return movedMsg{ref: ref, to: tr.to}
	}
}

// moveEpic applies an epic status transition (activate/retire/deprecate). Epic
// status is a frontmatter field, not a directory, so MoveEpic rewrites it in place
// — the file never moves. Success → movedMsg (flash + reload); failure →
// actionErrMsg (red flash, no reload), matching the CLI.
func moveEpic(svc *core.Service, ref entityRef, tr transition) tea.Cmd {
	return func() tea.Msg {
		if _, err := svc.MoveEpic(ref.key, tr.to, false); err != nil {
			return actionErrMsg{ref: ref, err: err}
		}
		return movedMsg{ref: ref, to: tr.to}
	}
}

// newEntityTabs is the entity registry: the ordered set of browsable entities. st
// is the Model's shared *styles, threaded into each list delegate so the rows
// render through the same per-Model palette (and see a Run-time repopulation).
func newEntityTabs(st *styles) []*entityTab {
	mk := func(d list.ItemDelegate) list.Model {
		l := list.New(nil, d, 0, 0)
		// The title slot carries the S2b state chip (set each render from chip()),
		// not a static "Tasks" label — that duplicated the tab strip. An empty chip
		// collapses the row entirely. TitleBar padding is stripped so the chip (and
		// the `/` filter prompt, which bubbles draws in this same slot) sit on one
		// tight line.
		l.Styles.Title = lipgloss.NewStyle().Bold(true)
		l.Styles.TitleBar = lipgloss.NewStyle()
		// Default filter is fuzzy (list.DefaultFilter); the prompt advertises the
		// mode, which `F` toggles to substring at runtime.
		l.FilterInput.Prompt = filterPrompt(false)
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
			viewAxis: statusViews, viewAliases: statusViewAliases,
			list: mk(taskDelegate{st: st}), loadList: loadTaskList, loadItem: loadTaskDetail,
			sortCols: taskSortCols, transitions: taskTransitions, applyMove: moveTask,
		},
		{
			kind: entityEpics, name: "epics", aliases: []string{"e", "epic"},
			viewAxis: epicViews, viewAliases: epicViewAliases,
			list: mk(epicDelegate{st: st}), loadList: loadEpicList, loadItem: loadEpicDetail,
			// Epic status is a frontmatter field, not a directory: the `m` menu / `:`
			// verbs flip it via svc.MoveEpic (the file stays put), mirroring task/audit.
			sortCols: epicSortCols, transitions: epicTransitions, applyMove: moveEpic,
		},
		{
			// Threads are deliberately read-only in this first visible slice. Lifecycle,
			// membership, and graph semantics all come from core projections; the registry
			// owns only list/detail loading, selection, sorting, and presentation.
			kind: entityThreads, name: "threads", aliases: []string{"th", "thread"},
			list: mk(threadDelegate{st: st}), loadList: loadThreadList, loadItem: loadThreadDetail,
			sortCols: threadSortCols,
		},
		{
			kind: entityAudits, name: "audits", aliases: []string{"a", "audit"},
			viewAxis: auditViews,
			list:     mk(auditDelegate{st: st}), loadList: loadAuditList, loadItem: loadAuditDetail,
			sortCols: auditSortCols, transitions: auditTransitions, applyMove: moveAudit,
		},
		{
			// Research is the only AXIS-LESS, LIFECYCLE-LESS tab: no viewAxis (nothing to
			// filter by — research has no status), and no transitions/applyMove (no verbs to
			// move it between). Both are already guarded — cycleView returns early on an
			// empty axis and the `m` menu checks len(transitions) — so the reducer needs no
			// research-specific branch.
			kind: entityResearch, name: "research", aliases: []string{"r"},
			list: mk(researchDelegate{st: st}), loadList: loadResearchList, loadItem: loadResearchDetail,
			sortCols: researchSortCols,
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
