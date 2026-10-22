package tui

import (
	"errors"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// Task-specific controller/command boundary: task fields stay in task_create.go
// over shared staged_creation mechanics; core invocation, selected-parent intent
// and typed receipts stay here. Other forms supply their own request/receipt policy rather
// than borrowing task fields, status assumptions or parent validation.

func (m *Model) openTaskCreate(epic entityRef) tea.Cmd {
	if m.atlas.opening {
		m.flash, m.flashErr = "wait for the pending workspace open before creating a task", true
		return nil
	}
	if m.activeCreation() != nil || m.svc == nil || m.onAtlas || m.onDash {
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
	m.creationGen++
	m.createTask = newTaskCreateForm(m.svc, m.creationGen, m.sessionGen, m.creationDestination(), epic)
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
	p.Tags = creationTags(f.inputs[taskCreateTags].Value())
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
	uncertain := errors.As(msg.err, &finalization) || (msg.err != nil && !creationKnownRefusal(msg.err))
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

func (m *Model) landCreatedTask(receipt core.TaskCreationReceipt) tea.Cmd {
	i := indexOfKind(m.tabs, entityTasks)
	view := m.tabs[i].statusView
	if view != "all" && view != string(domain.StatusReadyToStart) {
		view = ""
	}
	return m.landCreatedEntity(entityTasks, entityRef{key: receipt.Task.ID, label: receipt.Task.Slug},
		view, m.svc.HasLocalPath(core.EntityTask))
}
