package tui

import (
	"errors"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// Task-specific controller/command boundary: field and widget mechanics stay in
// task_create.go; core invocation, selected-parent intent and typed receipts stay
// here. Future entity forms must supply their own request/receipt policy rather
// than borrowing task fields, status assumptions or parent validation.

func (m Model) taskCreationActions() []paletteItem {
	if m.onDash || m.onAtlas {
		return nil
	}
	switch m.cur().kind {
	case entityTasks:
		return []paletteItem{{kind: palAction, word: "new-task", title: "New task", filter: "new create task"}}
	case entityEpics:
		ref := m.selectedRef()
		title := "New task in this epic — select an epic first"
		if _, ok := m.selectedEpic(); ok {
			title = "New task in " + ref.label
		} else {
			ref = entityRef{}
		}
		return []paletteItem{{kind: palAction, word: "new-task-in-epic", ref: ref, title: title, filter: "new create task in this epic " + ref.label}}
	}
	return nil
}

func (m *Model) openTaskCreate(epic entityRef) tea.Cmd {
	if m.atlas.opening {
		m.flash, m.flashErr = "wait for the pending workspace open before creating a task", true
		return nil
	}
	if m.createTask.active || m.svc == nil || m.onAtlas || m.onDash {
		return nil
	}
	if epic.empty() && m.cur().kind != entityTasks {
		return nil
	}
	if !epic.empty() {
		if _, ok := m.selectedEpic(); !ok || m.selectedRef() != epic {
			m.flash, m.flashErr = "select the epic again before creating a task", true
			return nil
		}
	}
	m.createTaskGen++
	destination := m.spaceName()
	if m.workspace.Checkout != "" {
		destination += " · " + m.workspace.Checkout
	}
	m.createTask = newTaskCreateForm(m.svc, m.createTaskGen, m.sessionGen, destination, epic)
	return tea.Batch(m.createTask.focus(taskCreateTitle), m.createTask.loadEpics())
}

func (f *taskCreateForm) loadEpics() tea.Cmd {
	f.catalogGen++
	f.catalogLoaded = false
	svc, gen, catalog := f.origin, f.gen, f.catalogGen
	return func() tea.Msg {
		epics, problems, err := svc.ListEpics()
		return taskCreateEpicsMsg{gen: gen, catalog: catalog, epics: epics, problems: problems, err: err}
	}
}

func (m *Model) applyTaskCreateEpics(msg taskCreateEpicsMsg) {
	f := &m.createTask
	if !f.active || f.displaced || msg.gen != f.gen || msg.catalog != f.catalogGen {
		return
	}
	f.catalogLoaded = true
	if msg.err == nil {
		refs := make([]entityRef, 0, len(msg.epics))
		for _, epic := range msg.epics {
			refs = append(refs, entityRef{key: epic.Source.ID, label: epic.Epic.ID})
		}
		msg.err = validateEntityRefs(refs)
	}
	f.catalogErr = ""
	f.epics = nil
	f.unreadable = len(msg.problems)
	if msg.err != nil {
		f.catalogErr = msg.err.Error()
		return
	}
	f.epics = msg.epics
}

func (m *Model) submitTaskCreate() tea.Cmd {
	f := &m.createTask
	if !f.active || f.saving || f.locked {
		return nil
	}
	if f.origin != m.svc || f.session != m.sessionGen {
		f.displace()
		return nil
	}
	if f.epic.empty() {
		f.err = "Choose an existing epic before creating a task."
		return f.focus(taskCreateEpic)
	}
	p := core.NewTaskParams{Title: f.inputs[taskCreateTitle].Value(), Description: f.inputs[taskCreateDescription].Value(),
		Priority: f.inputs[taskCreatePriority].Value(), Effort: f.inputs[taskCreateEffort].Value()}
	for _, tag := range strings.Split(f.inputs[taskCreateTags].Value(), ",") {
		if tag = strings.TrimSpace(tag); tag != "" {
			p.Tags = append(p.Tags, tag)
		}
	}
	for _, field := range []struct {
		row int
		out *int
	}{{taskCreateTier, &p.Tier}, {taskCreateAutonomy, &p.Autonomy}} {
		value := strings.TrimSpace(f.inputs[field.row].Value())
		if value != "" {
			n, err := strconv.Atoi(value)
			if err != nil {
				f.err = taskCreateLabels[field.row] + " must be a whole number."
				f.more = true
				return f.focus(field.row)
			}
			*field.out = n
		}
	}
	f.saving, f.err = true, ""
	svc, target, gen, session := f.origin, core.EpicMutationTarget{ID: f.epic.key}, f.gen, f.session
	return func() tea.Msg {
		receipt, err := svc.NewTaskForSelectedEpic(target, p)
		return taskCreatedMsg{gen: gen, session: session, receipt: receipt, err: err}
	}
}

func (m Model) handleTaskCreated(msg taskCreatedMsg) (tea.Model, tea.Cmd) {
	f := &m.createTask
	if !f.active || !f.saving || msg.gen != f.gen || msg.session != f.session {
		return m, nil
	}
	f.saving, f.receipt = false, msg.receipt
	if f.displaced || f.origin != m.svc || f.session != m.sessionGen {
		f.displace()
		if msg.err != nil {
			f.err += " Origin result: " + msg.err.Error()
		}
		return m, nil
	}
	var finalization *core.CreateFinalizationError
	uncertain := errors.As(msg.err, &finalization) || (msg.err != nil && !taskCreateKnownRefusal(msg.err))
	if msg.receipt.Committed || msg.receipt.Local.CommittedPath != "" || uncertain {
		if msg.err == nil && msg.receipt.Committed && msg.receipt.Task.ID != "" && !msg.receipt.DryRun {
			receipt := msg.receipt
			f.close()
			return m, m.landCreatedTask(receipt)
		}
		f.locked = true
		f.err = "Inspect the original attempt before doing anything else. Create is disabled; it will not be replayed."
		if msg.err != nil {
			f.err += " " + msg.err.Error()
		}
		return m, m.reloadAll()
	}
	if msg.err == nil {
		// A non-preview success without commit evidence is not permission to mint
		// again, even if a faulty/remote adapter returned an incomplete receipt.
		f.locked = true
		f.err = "Create returned no durable confirmation. Inspect the origin; do not create again."
		return m, nil
	}
	f.err = domain.Reason(msg.err) // no field attribution guessed from error strings
	return m, nil
}

func taskCreateKnownRefusal(err error) bool {
	return errors.Is(err, domain.ErrValidation) || errors.Is(err, domain.ErrNotFound) ||
		errors.Is(err, domain.ErrAmbiguous) || errors.Is(err, domain.ErrConflict) || errors.Is(err, core.ErrReadOnlyPersistence)
}

func (m *Model) landCreatedTask(receipt core.TaskCreationReceipt) tea.Cmd {
	i := indexOfKind(m.tabs, entityTasks)
	m.exitDashboard(i)
	m.zoom, m.immersiveZoom = false, false
	m.setFocus(focusList)
	m.recomputeLayout()
	tab := m.tabs[i]
	tab.list.ResetFilter()
	if tab.statusView != "all" && tab.statusView != string(domain.StatusReadyToStart) {
		tab.statusView = ""
	}
	tab.restore = entityRef{key: receipt.Task.ID, label: receipt.Task.Slug}
	tab.restoreWiden = true
	m.flash, m.flashErr = "created "+receipt.Task.Slug, false
	if m.svc.HasLocalPath(core.EntityTask) {
		m.flash += " — E opens its scaffold in your editor"
	}
	return m.reloadAll()
}
