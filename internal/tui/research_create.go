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
	researchCreateTitle = iota
	researchCreateDescription
	researchCreateTags
	researchCreateDate
	researchCreateCancel
	researchCreateSubmit
)

var researchCreateLabels = [...]string{"Title", "Description", "Tags", "Created date", "Cancel", "Create"}

// Research has no lifecycle or parent. Core alone mints from the chosen date
// and retries safe ID collisions; the UI submits one typed request per action.
// See docs/architecture/tui.md#audit-and-research-creation.
type researchCreateForm struct {
	creationState
	inputs  [researchCreateSubmit + 1]textinput.Model
	receipt core.ResearchCreationReceipt
}

func (f *researchCreateForm) state() *creationState { return &f.creationState }
func (researchCreateForm) noun() string             { return "research" }
func researchCreateInput(row int) bool {
	return row >= researchCreateTitle && row <= researchCreateDate
}

func newResearchCreateForm(origin *core.Service, gen, session uint64, destination string) researchCreateForm {
	f := researchCreateForm{creationState: creationState{active: true, kind: entityResearch,
		origin: origin, gen: gen, session: session, destination: destination}}
	for i := range f.inputs {
		f.inputs[i] = textinput.New()
		f.inputs[i].Prompt, f.inputs[i].CharLimit = "", 0
	}
	f.inputs[researchCreateTitle].Placeholder = "Research title"
	f.inputs[researchCreateDescription].Placeholder = "Optional: a short summary"
	f.inputs[researchCreateTags].Placeholder = "Optional, comma-separated"
	f.inputs[researchCreateDate].Placeholder = "YYYY-MM-DD · blank uses today at creation"
	return f
}

func (f *researchCreateForm) close() {
	f.active, f.mode = false, creationFieldsMode
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
}

func (f researchCreateForm) dirty() bool {
	for _, input := range f.inputs[:researchCreateCancel] {
		if input.Value() != "" {
			return true
		}
	}
	return false
}

func (f researchCreateForm) rows() []int {
	return []int{researchCreateTitle, researchCreateDescription, researchCreateTags,
		researchCreateDate, researchCreateCancel, researchCreateSubmit}
}

func (f *researchCreateForm) controls() creationControls {
	return creationControls{state: &f.creationState, inputs: f.inputs[:], rows: f.rows(),
		isText: researchCreateInput, dirty: f.dirty(), close: f.close}
}

func (f *researchCreateForm) updateInput(msg tea.Msg) tea.Cmd {
	if !f.editable() || !researchCreateInput(f.cursor) {
		return nil
	}
	var cmd tea.Cmd
	f.inputs[f.cursor], cmd = updateCreationInput(f.inputs[f.cursor], msg)
	return f.widgetCmd(cmd)
}

func (m *Model) openResearchCreate() tea.Cmd {
	if m.atlas.opening {
		m.flash, m.flashErr = "wait for the pending workspace open before creating research", true
		return nil
	}
	if m.activeCreation() != nil || m.svc == nil || m.onAtlas || m.onDash || m.cur().kind != entityResearch {
		return nil
	}
	return m.creation.openResearch(m.svc, m.sessionGen, m.creationDestination())
}

func (f *researchCreateForm) handleKey(m *Model, msg tea.KeyPressMsg) tea.Cmd {
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
			if !f.displaced && f.receipt.Committed && f.receipt.Research.ID != "" && !f.receipt.DryRun {
				receipt := f.receipt
				f.close()
				return m.landCreatedResearch(receipt)
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
		case researchCreateCancel:
			f.controls().cancel()
		case researchCreateSubmit:
			return m.submitResearchCreate()
		}
		return nil
	}
	return f.updateInput(msg)
}

func (m *Model) submitResearchCreate() tea.Cmd {
	f := &m.creation.research
	if !f.active || f.saving || f.locked {
		return nil
	}
	if f.origin != m.svc || f.session != m.sessionGen {
		f.displace()
		return nil
	}
	p := core.NewResearchParams{Title: f.inputs[researchCreateTitle].Value(),
		Description: f.inputs[researchCreateDescription].Value(), Tags: creationTags(f.inputs[researchCreateTags].Value()),
		Created: f.inputs[researchCreateDate].Value()}
	f.saving, f.err = true, ""
	svc, gen, session := f.origin, f.gen, f.session
	return func() tea.Msg {
		receipt, err := svc.NewResearch(p)
		return researchCreatedMsg{gen: gen, session: session, receipt: receipt, err: err}
	}
}

type researchCreatedMsg struct {
	gen, session uint64
	receipt      core.ResearchCreationReceipt
	err          error
}

func (msg researchCreatedMsg) creationRoute() (entityKind, uint64, uint64) {
	return entityResearch, msg.gen, msg.session
}
func (msg researchCreatedMsg) applyCreation(m *Model) tea.Cmd { return m.handleResearchCreated(msg) }

// The host matched kind/attempt/origin. Only this controller interprets the research receipt.
func (m *Model) handleResearchCreated(msg researchCreatedMsg) tea.Cmd {
	f := &m.creation.research
	f.saving, f.receipt = false, msg.receipt
	if f.displaced || f.origin != m.svc || f.session != m.sessionGen {
		f.displace()
		if msg.err != nil {
			f.err += " Origin result: " + msg.err.Error()
		}
		return nil
	}
	if msg.err == nil && msg.receipt.Committed && msg.receipt.Research.ID != "" && !msg.receipt.DryRun {
		f.close()
		return m.landCreatedResearch(msg.receipt)
	}
	var finalization *core.CreateFinalizationError
	uncertain := errors.As(msg.err, &finalization) || (msg.err != nil && !creationKnownRefusal(msg.err))
	if msg.receipt.Committed || msg.receipt.Local.CommittedPath != "" || uncertain || msg.err == nil {
		f.locked = true
		f.err = "Inspect the original research attempt. Create is disabled; it will not be replayed."
		if msg.err != nil {
			f.err += " " + msg.err.Error()
		}
		return m.reloadAll()
	}
	f.err = domain.Reason(msg.err)
	return nil
}

func (m *Model) landCreatedResearch(receipt core.ResearchCreationReceipt) tea.Cmd {
	return m.landCreatedEntity(entityResearch, entityRef{key: receipt.Research.ID, label: receipt.Research.Slug},
		"", m.svc.HasLocalPath(core.EntityResearch))
}

func (f researchCreateForm) recoveryText() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Research creation in %s\n", f.destination)
	for row, input := range f.inputs[:researchCreateCancel] {
		fmt.Fprintf(&b, "%s: %s\n", researchCreateLabels[row], input.Value())
	}
	fmt.Fprintf(&b, "Committed: %t\nResearch ID: %s\nDestination: %s\n%s", f.receipt.Committed,
		f.receipt.Research.ID, f.receipt.Local.CommittedPath, f.err)
	return b.String()
}

func (f researchCreateForm) view(s *styles, maxW, maxH int) string {
	p := creationPresentation{noun: f.noun(), title: f.inputs[researchCreateTitle].Value(),
		subtitle: "Nothing saved until Create · no lifecycle · created date determines identity date",
		recordID: f.receipt.Research.ID, path: f.receipt.Local.CommittedPath,
		inspect: f.receipt.Committed && f.receipt.Research.ID != "" && !f.receipt.DryRun}
	frame := creationHeader(f.creationState, p, s, maxW, maxH)
	if frame.rendered != "" {
		return frame.rendered
	}
	frame.lines = creationErrorLines(frame.lines, f.err, frame.innerW, 3, s)
	return creationFieldsView(f.creationState, s, frame, creationFieldLayout{
		rows: f.rows(), labels: researchCreateLabels[:], cancel: researchCreateCancel, submit: researchCreateSubmit,
		cell: func(row, width int) string { return creationCell(f.creationState, s, f.inputs[row], nil, row, width) }})
}
