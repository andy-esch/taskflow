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
	epicCreateTitle = iota
	epicCreateDescription
	epicCreateTags
	epicCreateStatus
	epicCreatePriority
	epicCreateCancel
	epicCreateSubmit
)

var epicCreateLabels = [...]string{"Title", "Description", "Tags", "Status", "Priority", "Cancel", "Create"}

// Epic input/default and receipt policy remain separate from task authoring.
// See docs/architecture/tui.md#epic-creation-and-shared-staging.
type epicCreateForm struct {
	creationState
	inputs  [epicCreateSubmit + 1]textinput.Model
	receipt core.EpicCreationReceipt
}

func (f *epicCreateForm) state() *creationState { return &f.creationState }

type epicCreatedMsg struct {
	gen, session uint64
	receipt      core.EpicCreationReceipt
	err          error
}

func epicCreateInput(row int) bool {
	return row == epicCreateTitle || row == epicCreateDescription || row == epicCreateTags
}

func epicCreateEnumOptions(row int) []string {
	switch row {
	case epicCreateStatus:
		return domain.AllEpicStatuses()
	case epicCreatePriority:
		return editEnumOptions("priority")
	}
	return nil
}

func newEpicCreateForm(origin *core.Service, gen, session uint64, destination string) epicCreateForm {
	f := epicCreateForm{creationState: creationState{active: true, kind: entityEpics,
		origin: origin, gen: gen, session: session, destination: destination}}
	for i := range f.inputs {
		f.inputs[i] = textinput.New()
		f.inputs[i].Prompt, f.inputs[i].CharLimit = "", 0
	}
	f.inputs[epicCreateTitle].Placeholder = "Epic title"
	f.inputs[epicCreateDescription].Placeholder = "Required: what this epic organizes"
	f.inputs[epicCreateTags].Placeholder = "Optional, comma-separated"
	// Unlike NewTask, NewEpic requires explicit status and priority. These are
	// visible initial form choices matching the CLI, not invented core defaults.
	f.inputs[epicCreateStatus].SetValue(domain.EpicStatusActive)
	f.inputs[epicCreatePriority].SetValue("medium")
	return f
}

func (f *epicCreateForm) close() {
	f.active, f.mode = false, creationFieldsMode
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
}

func (f epicCreateForm) dirty() bool {
	return f.inputs[epicCreateTitle].Value() != "" || f.inputs[epicCreateDescription].Value() != "" ||
		f.inputs[epicCreateTags].Value() != "" || f.inputs[epicCreateStatus].Value() != domain.EpicStatusActive ||
		f.inputs[epicCreatePriority].Value() != "medium"
}

func (f epicCreateForm) rows() []int {
	return []int{epicCreateTitle, epicCreateDescription, epicCreateTags, epicCreateStatus,
		epicCreatePriority, epicCreateCancel, epicCreateSubmit}
}

func (f *epicCreateForm) controls() creationControls {
	return creationControls{state: &f.creationState, inputs: f.inputs[:], rows: f.rows(),
		isText: epicCreateInput, options: epicCreateEnumOptions, dirty: f.dirty(), close: f.close}
}

func (f *epicCreateForm) updateInput(msg tea.Msg) tea.Cmd {
	if !f.editable() || !epicCreateInput(f.cursor) {
		return nil
	}
	var cmd tea.Cmd
	f.inputs[f.cursor], cmd = updateCreationInput(f.inputs[f.cursor], msg)
	return f.widgetCmd(cmd)
}

func (m *Model) openEpicCreate() tea.Cmd {
	if m.atlas.opening {
		m.flash, m.flashErr = "wait for the pending workspace open before creating an epic", true
		return nil
	}
	if m.activeCreation() != nil || m.svc == nil || m.onAtlas || m.onDash || m.cur().kind != entityEpics {
		return nil
	}
	m.creationGen++
	m.createEpic = newEpicCreateForm(m.svc, m.creationGen, m.sessionGen, m.creationDestination())
	return m.createEpic.controls().focus(epicCreateTitle)
}

func (f *epicCreateForm) handleKey(m *Model, msg tea.KeyPressMsg) tea.Cmd {
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
			if !f.displaced && f.receipt.Committed && f.receipt.Epic.ID != "" && !f.receipt.DryRun {
				receipt := f.receipt
				f.close()
				return m.landCreatedEpic(receipt)
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
		case epicCreateCancel:
			f.controls().cancel()
		case epicCreateSubmit:
			return m.submitEpicCreate()
		}
		return nil
	}
	return f.updateInput(msg)
}

func (m *Model) submitEpicCreate() tea.Cmd {
	f := &m.createEpic
	if !f.active || f.saving || f.locked {
		return nil
	}
	if f.origin != m.svc || f.session != m.sessionGen {
		f.displace()
		return nil
	}
	p := core.NewEpicParams{Title: f.inputs[epicCreateTitle].Value(), Description: f.inputs[epicCreateDescription].Value(),
		Status: f.inputs[epicCreateStatus].Value(), Priority: f.inputs[epicCreatePriority].Value(),
		Tags: creationTags(f.inputs[epicCreateTags].Value())}
	f.saving, f.err = true, ""
	svc, gen, session := f.origin, f.gen, f.session
	return func() tea.Msg {
		receipt, err := svc.NewEpic(p)
		return epicCreatedMsg{gen: gen, session: session, receipt: receipt, err: err}
	}
}

func (m Model) handleEpicCreated(msg epicCreatedMsg) (tea.Model, tea.Cmd) {
	f := &m.createEpic
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
	if msg.err == nil && msg.receipt.Committed && msg.receipt.Epic.ID != "" && !msg.receipt.DryRun {
		f.close()
		return m, m.landCreatedEpic(msg.receipt)
	}
	var finalization *core.CreateFinalizationError
	uncertain := errors.As(msg.err, &finalization) || (msg.err != nil && !creationKnownRefusal(msg.err))
	if msg.receipt.Committed || msg.receipt.Local.CommittedPath != "" || uncertain || msg.err == nil {
		f.locked = true
		f.err = "Inspect the original epic attempt. Create is disabled; it will not be replayed."
		if msg.err != nil {
			f.err += " " + msg.err.Error()
		}
		return m, m.reloadAll()
	}
	f.err = domain.Reason(msg.err)
	return m, nil
}

func (m *Model) landCreatedEpic(receipt core.EpicCreationReceipt) tea.Cmd {
	return m.landCreatedEntity(entityEpics, entityRef{key: receipt.Epic.ID, label: receipt.Epic.ID},
		"all", m.svc.HasLocalPath(core.EntityEpic))
}

func (f epicCreateForm) recoveryText() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Epic creation in %s\n", f.destination)
	for row, input := range f.inputs[:epicCreateCancel] {
		fmt.Fprintf(&b, "%s: %s\n", epicCreateLabels[row], input.Value())
	}
	fmt.Fprintf(&b, "Committed: %t\nEpic ID: %s\nDestination: %s\n%s", f.receipt.Committed,
		f.receipt.Epic.ID, f.receipt.Local.CommittedPath, f.err)
	return b.String()
}

func (f epicCreateForm) view(s *styles, maxW, maxH int) string {
	p := creationPresentation{noun: "epic", title: f.inputs[epicCreateTitle].Value(),
		subtitle: "Nothing saved until Create · status and priority are explicit choices",
		recordID: f.receipt.Epic.ID, path: f.receipt.Local.CommittedPath,
		inspect: f.receipt.Committed && f.receipt.Epic.ID != "" && !f.receipt.DryRun}
	frame := creationHeader(f.creationState, p, s, maxW, maxH)
	if frame.rendered != "" {
		return frame.rendered
	}
	frame.lines = creationErrorLines(frame.lines, f.err, frame.innerW, 3, s)
	return creationFieldsView(f.creationState, s, frame, creationFieldLayout{
		rows: f.rows(), labels: epicCreateLabels[:], cancel: epicCreateCancel, submit: epicCreateSubmit,
		cell: func(row, width int) string {
			return creationCell(f.creationState, s, f.inputs[row], epicCreateEnumOptions(row), row, width)
		}})
}
