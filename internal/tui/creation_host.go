package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
)

// Page creation has one metadata owner for palette and footer discovery. It is
// separate from contextual actions, so hints do not depend on their order/names.
func (m Model) pageCreationHint() string {
	if action, ok := m.pageCreationAction(); ok {
		return strings.ToLower(action.title)
	}
	return ""
}

// creationHost owns the active-kind/lifetime boundary, not entity policy. Concrete
// forms stay in value slots: storing a draft interface pointer here would alias
// mutable form state across copies of Model. The interface returned by draft is
// only a synchronous borrow from the current copy.
// See docs/architecture/tui.md#epic-creation-and-shared-staging.
type creationHost struct {
	kind     entityKind
	gen      uint64 // monotonic across kinds, closes and workspace activation
	task     taskCreateForm
	epic     epicCreateForm
	audit    auditCreateForm
	research researchCreateForm
	thread   threadCreateForm
}

func (m Model) pageCreationAction() (paletteItem, bool) {
	if m.onDash || m.onAtlas {
		return paletteItem{}, false
	}
	switch m.cur().kind {
	case entityTasks:
		return paletteItem{kind: palAction, word: "new-task", title: "New task", filter: "new create task"}, true
	case entityAudits:
		return paletteItem{kind: palAction, word: "new-audit", title: "New audit", filter: "new create audit"}, true
	case entityResearch:
		return paletteItem{kind: palAction, word: "new-research", title: "New research", filter: "new create research"}, true
	case entityThreads:
		return paletteItem{kind: palAction, word: "new-thread", title: "New Thread", filter: "new create thread"}, true
	case entityEpics:
		return paletteItem{kind: palAction, word: "new-epic", title: "New epic", filter: "new create epic"}, true
	}
	return paletteItem{}, false
}

func (m Model) creationActions() []paletteItem {
	action, ok := m.pageCreationAction()
	if !ok {
		return nil
	}
	if m.cur().kind != entityEpics {
		return []paletteItem{action}
	}
	ref := m.selectedRef()
	title := "New task in this epic — select an epic first"
	if _, ok := m.selectedEpic(); ok {
		title = "New task in " + ref.label
	} else {
		ref = entityRef{}
	}
	return []paletteItem{
		{kind: palAction, word: "new-task-in-epic", ref: ref, title: title, filter: "new create task in this epic " + ref.label},
		action,
	}
}

func (m *Model) openPageCreation() tea.Cmd {
	switch m.cur().kind {
	case entityTasks:
		return m.openTaskCreate(entityRef{})
	case entityEpics:
		return m.openEpicCreate()
	case entityAudits:
		return m.openAuditCreate()
	case entityResearch:
		return m.openResearchCreate()
	case entityThreads:
		return m.openThreadCreate()
	default:
		m.flash, m.flashErr = "creation is not available on this page yet", true
		return nil
	}
}

func (m *Model) runCreationAction(it paletteItem) tea.Cmd {
	switch it.word {
	case "new-epic":
		return m.openEpicCreate()
	case "new-task":
		return m.openTaskCreate(entityRef{})
	case "new-audit":
		return m.openAuditCreate()
	case "new-research":
		return m.openResearchCreate()
	case "new-thread":
		return m.openThreadCreate()
	case "new-task-in-epic":
		if it.ref.empty() || it.ref != m.selectedRef() {
			m.flash, m.flashErr = "select the epic again before creating a task", true
			return nil
		}
		return m.openTaskCreate(it.ref)
	}
	return nil
}

func (h *creationHost) draft() stagedCreation {
	var draft stagedCreation
	switch h.kind {
	case entityTasks:
		draft = &h.task
	case entityEpics:
		draft = &h.epic
	case entityAudits:
		draft = &h.audit
	case entityResearch:
		draft = &h.research
	case entityThreads:
		draft = &h.thread
	}
	if draft == nil || !draft.state().active {
		return nil
	}
	return draft
}

func (h *creationHost) active() *creationState {
	if draft := h.draft(); draft != nil {
		return draft.state()
	}
	return nil
}

func (h *creationHost) begin(kind entityKind) bool {
	if h.active() != nil {
		return false
	}
	switch kind {
	case entityTasks, entityEpics, entityAudits, entityResearch, entityThreads:
	default:
		return false
	}
	h.gen++
	h.kind = kind
	return true
}

func (h *creationHost) openTask(origin *core.Service, session uint64, destination string, epic entityRef) tea.Cmd {
	if !h.begin(entityTasks) {
		return nil
	}
	h.task = newTaskCreateForm(origin, h.gen, session, destination, epic)
	return tea.Batch(h.task.focus(taskCreateTitle), h.task.loadEpics())
}

func (h *creationHost) openEpic(origin *core.Service, session uint64, destination string) tea.Cmd {
	if !h.begin(entityEpics) {
		return nil
	}
	h.epic = newEpicCreateForm(origin, h.gen, session, destination)
	return h.epic.controls().focus(epicCreateTitle)
}

func (h *creationHost) displace() {
	if f := h.active(); f != nil {
		f.displace()
	}
}

func (h *creationHost) openAudit(origin *core.Service, session uint64, destination string) tea.Cmd {
	if !h.begin(entityAudits) {
		return nil
	}
	h.audit = newAuditCreateForm(origin, h.gen, session, destination)
	return h.audit.controls().focus(auditCreateArea)
}

func (h *creationHost) openResearch(origin *core.Service, session uint64, destination string) tea.Cmd {
	if !h.begin(entityResearch) {
		return nil
	}
	h.research = newResearchCreateForm(origin, h.gen, session, destination)
	return h.research.controls().focus(researchCreateTitle)
}

func (h *creationHost) openThread(origin *core.Service, session uint64, destination string) tea.Cmd {
	if !h.begin(entityThreads) {
		return nil
	}
	h.thread = newThreadCreateForm(origin, h.gen, session, destination)
	return h.thread.controls().focus(threadCreateTitle)
}

// Only typed create outcomes implement this private seam. It allows the shell's
// late-session exception without giving catalog, widget or ordinary read results
// permission to reach the incoming session. Each implementation retains its own
// typed receipt and recovery policy; this is not a core mutation interface.
type creationResult interface {
	creationRoute() (kind entityKind, gen, session uint64)
	applyCreation(*Model) tea.Cmd
}

func (h *creationHost) applyResult(m *Model, result creationResult) tea.Cmd {
	kind, gen, session := result.creationRoute()
	f := h.active()
	if f == nil || !f.saving || kind != f.kind || gen != f.gen || session != f.session {
		return nil
	}
	return result.applyCreation(m)
}
