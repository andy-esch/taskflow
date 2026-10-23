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
	auditCreateArea = iota
	auditCreateDate
	auditCreateCancel
	auditCreateSubmit
)

var auditCreateLabels = [...]string{"Area", "Date", "Cancel", "Create"}

// Audit area/date are creation intent, not a task title or lifecycle choice.
// Core owns date defaulting, open-bucket policy, identity and scaffolding.
// See docs/architecture/tui.md#audit-and-research-creation.
type auditCreateForm struct {
	creationState
	inputs  [auditCreateSubmit + 1]textinput.Model
	receipt core.AuditCreationReceipt
}

func (f *auditCreateForm) state() *creationState { return &f.creationState }
func (auditCreateForm) noun() string             { return "audit" }
func auditCreateInput(row int) bool              { return row == auditCreateArea || row == auditCreateDate }

func newAuditCreateForm(origin *core.Service, gen, session uint64, destination string) auditCreateForm {
	f := auditCreateForm{creationState: creationState{active: true, kind: entityAudits,
		origin: origin, gen: gen, session: session, destination: destination}}
	for i := range f.inputs {
		f.inputs[i] = textinput.New()
		f.inputs[i].Prompt, f.inputs[i].CharLimit = "", 0
	}
	f.inputs[auditCreateArea].Placeholder = "Required: what is being reviewed"
	f.inputs[auditCreateDate].Placeholder = "YYYY-MM-DD · blank uses today at creation"
	return f
}

func (f *auditCreateForm) close() {
	f.active, f.mode = false, creationFieldsMode
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
}

func (f auditCreateForm) dirty() bool {
	return f.inputs[auditCreateArea].Value() != "" || f.inputs[auditCreateDate].Value() != ""
}

func (f auditCreateForm) rows() []int {
	return []int{auditCreateArea, auditCreateDate, auditCreateCancel, auditCreateSubmit}
}

func (f *auditCreateForm) controls() creationControls {
	return creationControls{state: &f.creationState, inputs: f.inputs[:], rows: f.rows(),
		isText: auditCreateInput, dirty: f.dirty(), close: f.close}
}

func (f *auditCreateForm) updateInput(msg tea.Msg) tea.Cmd {
	if !f.editable() || !auditCreateInput(f.cursor) {
		return nil
	}
	var cmd tea.Cmd
	f.inputs[f.cursor], cmd = updateCreationInput(f.inputs[f.cursor], msg)
	return f.widgetCmd(cmd)
}

func (m *Model) openAuditCreate() tea.Cmd {
	if m.atlas.opening {
		m.flash, m.flashErr = "wait for the pending workspace open before creating an audit", true
		return nil
	}
	if m.activeCreation() != nil || m.svc == nil || m.onAtlas || m.onDash || m.cur().kind != entityAudits {
		return nil
	}
	return m.creation.openAudit(m.svc, m.sessionGen, m.creationDestination())
}

func (f *auditCreateForm) handleKey(m *Model, msg tea.KeyPressMsg) tea.Cmd {
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
			if !f.displaced && f.receipt.Committed && f.receipt.Audit.ID != "" && !f.receipt.DryRun {
				receipt := f.receipt
				f.close()
				return m.landCreatedAudit(receipt)
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
		case auditCreateCancel:
			f.controls().cancel()
		case auditCreateSubmit:
			return m.submitAuditCreate()
		}
		return nil
	}
	return f.updateInput(msg)
}

func (m *Model) submitAuditCreate() tea.Cmd {
	f := &m.creation.audit
	if !f.active || f.saving || f.locked {
		return nil
	}
	if f.origin != m.svc || f.session != m.sessionGen {
		f.displace()
		return nil
	}
	p := core.NewAuditParams{Area: f.inputs[auditCreateArea].Value(), Date: f.inputs[auditCreateDate].Value()}
	f.saving, f.err = true, ""
	svc, gen, session := f.origin, f.gen, f.session
	return func() tea.Msg {
		receipt, err := svc.NewAudit(p)
		return auditCreatedMsg{gen: gen, session: session, receipt: receipt, err: err}
	}
}

type auditCreatedMsg struct {
	gen, session uint64
	receipt      core.AuditCreationReceipt
	err          error
}

func (msg auditCreatedMsg) creationRoute() (entityKind, uint64, uint64) {
	return entityAudits, msg.gen, msg.session
}
func (msg auditCreatedMsg) applyCreation(m *Model) tea.Cmd { return m.handleAuditCreated(msg) }

// The host matched kind/attempt/origin. Only this controller interprets the audit receipt.
func (m *Model) handleAuditCreated(msg auditCreatedMsg) tea.Cmd {
	f := &m.creation.audit
	f.saving, f.receipt = false, msg.receipt
	if f.displaced || f.origin != m.svc || f.session != m.sessionGen {
		f.displace()
		if msg.err != nil {
			f.err += " Origin result: " + msg.err.Error()
		}
		return nil
	}
	if msg.err == nil && msg.receipt.Committed && msg.receipt.Audit.ID != "" && !msg.receipt.DryRun {
		f.close()
		return m.landCreatedAudit(msg.receipt)
	}
	var finalization *core.CreateFinalizationError
	uncertain := errors.As(msg.err, &finalization) || (msg.err != nil && !creationKnownRefusal(msg.err))
	if msg.receipt.Committed || msg.receipt.Local.CommittedPath != "" || uncertain || msg.err == nil {
		f.locked = true
		f.err = "Inspect the original audit attempt. Create is disabled; it will not be replayed."
		if msg.err != nil {
			f.err += " " + msg.err.Error()
		}
		return m.reloadAll()
	}
	f.err = domain.Reason(msg.err)
	return nil
}

func (m *Model) landCreatedAudit(receipt core.AuditCreationReceipt) tea.Cmd {
	return m.landCreatedEntity(entityAudits, entityRef{key: receipt.Audit.ID, label: receipt.Audit.Slug},
		"all", m.svc.HasLocalPath(core.EntityAudit))
}

func (f auditCreateForm) recoveryText() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Audit creation in %s\n", f.destination)
	for row, input := range f.inputs[:auditCreateCancel] {
		fmt.Fprintf(&b, "%s: %s\n", auditCreateLabels[row], input.Value())
	}
	fmt.Fprintf(&b, "Committed: %t\nAudit ID: %s\nDestination: %s\n%s", f.receipt.Committed,
		f.receipt.Audit.ID, f.receipt.Local.CommittedPath, f.err)
	return b.String()
}

func (f auditCreateForm) view(s *styles, maxW, maxH int) string {
	p := creationPresentation{noun: f.noun(), title: f.inputs[auditCreateArea].Value(),
		subtitle: "Nothing saved until Create · opens in the open bucket · blank date uses today",
		recordID: f.receipt.Audit.ID, path: f.receipt.Local.CommittedPath,
		inspect: f.receipt.Committed && f.receipt.Audit.ID != "" && !f.receipt.DryRun}
	frame := creationHeader(f.creationState, p, s, maxW, maxH)
	if frame.rendered != "" {
		return frame.rendered
	}
	frame.lines = creationErrorLines(frame.lines, f.err, frame.innerW, 3, s)
	return creationFieldsView(f.creationState, s, frame, creationFieldLayout{
		rows: f.rows(), labels: auditCreateLabels[:], cancel: auditCreateCancel, submit: auditCreateSubmit,
		cell: func(row, width int) string { return creationCell(f.creationState, s, f.inputs[row], nil, row, width) }})
}
