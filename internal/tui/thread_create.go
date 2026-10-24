package tui

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

const (
	threadCreateTitle = iota
	threadCreateDescription
	threadCreateGoal
	threadCreateTargetDate
	threadCreateTags
	threadCreateCancel
	threadCreateSubmit
)

var threadCreateLabels = [...]string{"Title", "Description", "Goal", "Target date", "Tags", "Cancel", "Create"}

// Thread creation deliberately has no task/member input or lifecycle choice.
// The guarded core use case creates an empty, unstarted Thread; graph validation,
// conflict retry and typed durability evidence remain there, not in the form.
// See docs/architecture/tui.md#thread-creation.
type threadCreateForm struct {
	creationState
	inputs  [threadCreateSubmit + 1]textinput.Model
	receipt core.ThreadCreationReceipt
}

func (f *threadCreateForm) state() *creationState { return &f.creationState }
func (threadCreateForm) noun() string             { return "Thread" }
func threadCreateInput(row int) bool              { return row >= threadCreateTitle && row <= threadCreateTags }

func newThreadCreateForm(origin *core.Service, gen, session uint64, destination string) threadCreateForm {
	f := threadCreateForm{creationState: creationState{active: true, kind: entityThreads,
		origin: origin, gen: gen, session: session, destination: destination}}
	for i := range f.inputs {
		f.inputs[i] = newCreationInput()
	}
	f.inputs[threadCreateTitle].Placeholder = "Thread title"
	f.inputs[threadCreateDescription].Placeholder = "Required: a short summary"
	f.inputs[threadCreateGoal].Placeholder = "Required: what success looks like"
	f.inputs[threadCreateTargetDate].Placeholder = "Optional, YYYY-MM-DD"
	f.inputs[threadCreateTags].Placeholder = "Optional, comma-separated"
	return f
}

func (f *threadCreateForm) close() {
	f.active, f.mode = false, creationFieldsMode
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
}

func (f threadCreateForm) dirty() bool {
	for _, input := range f.inputs[:threadCreateCancel] {
		if input.Value() != "" {
			return true
		}
	}
	return false
}

func (f threadCreateForm) rows() []int {
	return []int{threadCreateTitle, threadCreateDescription, threadCreateGoal,
		threadCreateTargetDate, threadCreateTags, threadCreateCancel, threadCreateSubmit}
}

func (f *threadCreateForm) controls() creationControls {
	return creationControls{state: &f.creationState, inputs: f.inputs[:], rows: f.rows(),
		isText: threadCreateInput, dirty: f.dirty(), close: f.close}
}

func (f *threadCreateForm) updateInput(msg tea.Msg) tea.Cmd {
	if !f.editable() || !threadCreateInput(f.cursor) {
		return nil
	}
	var cmd tea.Cmd
	f.inputs[f.cursor], cmd = updateCreationInput(f.inputs[f.cursor], msg)
	return f.widgetCmd(cmd)
}

func (m *Model) openThreadCreate() tea.Cmd {
	if m.atlas.opening {
		m.flash, m.flashErr = "wait for the pending workspace open before creating a Thread", true
		return nil
	}
	if m.activeCreation() != nil || m.svc == nil || m.onAtlas || m.onDash || m.cur().kind != entityThreads {
		return nil
	}
	return m.creation.openThread(m.svc, m.sessionGen, m.creationDestination())
}

func (f *threadCreateForm) handleKey(m *Model, msg tea.KeyPressMsg) tea.Cmd {
	if f.saving {
		if f.displaced && msg.String() == "y" {
			return copyCreationRecovery(f.recoveryText())
		}
		return nil
	}
	if f.locked {
		switch msg.String() {
		case "y":
			return copyCreationRecovery(f.recoveryText())
		case "enter":
			if !f.displaced && f.receipt.Committed && f.receipt.Thread.ID != "" && !f.receipt.DryRun {
				receipt := f.receipt
				f.close()
				return m.landCreatedThread(receipt)
			}
			f.close()
		case "esc":
			f.close()
		}
		return nil
	}
	if handled, cmd := f.controls().handleKey(msg); handled {
		return cmd
	}
	if msg.String() == "enter" {
		switch f.cursor {
		case threadCreateCancel:
			f.controls().cancel()
		case threadCreateSubmit:
			return m.submitThreadCreate()
		}
		return nil
	}
	return f.updateInput(msg)
}

func (m *Model) submitThreadCreate() tea.Cmd {
	f := &m.creation.thread
	if !f.active || f.saving || f.locked {
		return nil
	}
	if f.origin != m.svc || f.session != m.sessionGen {
		f.displace()
		return nil
	}
	p := core.NewThreadParams{Title: f.inputs[threadCreateTitle].Value(),
		Description: f.inputs[threadCreateDescription].Value(), Goal: f.inputs[threadCreateGoal].Value(),
		TargetDate: f.inputs[threadCreateTargetDate].Value(), Tags: creationTags(f.inputs[threadCreateTags].Value())}
	// Tasks, Body, Template and DryRun are intentionally absent. The form cannot
	// create members/dependencies or replace core's ordinary scaffold/defaults.
	f.saving, f.err = true, ""
	svc, gen, session := f.origin, f.gen, f.session
	return func() tea.Msg {
		receipt, err := svc.NewThread(p)
		return threadCreatedMsg{gen: gen, session: session, receipt: receipt, err: err}
	}
}

type threadCreatedMsg struct {
	gen, session uint64
	receipt      core.ThreadCreationReceipt
	err          error
}

func (msg threadCreatedMsg) creationRoute() (entityKind, uint64, uint64) {
	return entityThreads, msg.gen, msg.session
}
func (msg threadCreatedMsg) applyCreation(m *Model) tea.Cmd { return m.handleThreadCreated(msg) }

// Routed only after the host matches kind/attempt/origin. Changed and a planned
// path describe a proposal, not a commit; committed/unknown outcomes cannot replay.
func (m *Model) handleThreadCreated(msg threadCreatedMsg) tea.Cmd {
	f := &m.creation.thread
	f.saving, f.receipt = false, msg.receipt
	if f.displaced || f.origin != m.svc || f.session != m.sessionGen {
		f.displace()
		if msg.err != nil {
			f.err += " Origin result: " + msg.err.Error()
		}
		return nil
	}
	if msg.err == nil && msg.receipt.Committed && msg.receipt.Thread.ID != "" && !msg.receipt.DryRun {
		f.close()
		return m.landCreatedThread(msg.receipt)
	}
	var failure *core.ThreadCreationMutationFailure
	uncertain := errors.As(msg.err, &failure) || (msg.err != nil && !creationKnownRefusal(msg.err))
	if msg.receipt.Committed || msg.receipt.Local.CommittedPath != "" || uncertain || msg.err == nil {
		f.locked = true
		f.err = "Inspect the original Thread attempt. Create is disabled; it will not be replayed."
		if msg.err != nil {
			f.err += " " + msg.err.Error()
		}
		return m.reloadAll()
	}
	f.err = domain.Reason(msg.err)
	return nil
}

func (m *Model) landCreatedThread(receipt core.ThreadCreationReceipt) tea.Cmd {
	return m.landCreatedEntity(entityThreads, entityRef{key: receipt.Thread.ID, label: receipt.Thread.Slug},
		"", m.svc.HasLocalPath(core.EntityThread))
}

func (f threadCreateForm) recoveryText() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Thread creation in %s\n", f.destination)
	for row, input := range f.inputs[:threadCreateCancel] {
		fmt.Fprintf(&b, "%s: %s\n", threadCreateLabels[row], input.Value())
	}
	fmt.Fprintf(&b, "Members: none requested\nChanged (proposal): %t\nCommitted: %t\nThread ID: %s\nDestination: %s\n%s",
		f.receipt.Changed, f.receipt.Committed, f.receipt.Thread.ID, f.receipt.Local.CommittedPath, f.err)
	return b.String()
}

func (f threadCreateForm) view(s *styles, maxW, maxH int) string {
	p := creationPresentation{noun: f.noun(), title: f.inputs[threadCreateTitle].Value(),
		subtitle: "Creates an empty, unstarted Thread · no task or dependency changes",
		recordID: f.receipt.Thread.ID, path: f.receipt.Local.CommittedPath,
		inspect: f.receipt.Committed && f.receipt.Thread.ID != "" && !f.receipt.DryRun}
	frame := creationHeader(f.creationState, p, s, maxW, maxH)
	if frame.rendered != "" {
		return frame.rendered
	}
	frame.lines = creationErrorLines(frame.lines, f.err, frame.innerW, 3, s)
	return creationFieldsView(f.creationState, s, frame, creationFieldLayout{
		rows: f.rows(), labels: threadCreateLabels[:], cancel: threadCreateCancel, submit: threadCreateSubmit,
		cell: func(row, width int) string { return creationCell(f.creationState, s, f.inputs[row], nil, row, width) }})
}
