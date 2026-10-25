package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
)

// Task authoring is staged primary-adapter state, not an existing-record edit.
// The form captures its originating Service; commands copy inputs at submission.
// List generations may advance while typing without retargeting the selected epic.
// See docs/architecture/tui.md#staged-task-creation.
const (
	taskCreateTitle = iota
	taskCreateEpic
	taskCreateTags
	taskCreateDescription
	taskCreateMore
	taskCreatePriority
	taskCreateEffort
	taskCreateTier
	taskCreateAutonomy
	taskCreateCancel
	taskCreateSubmit
)

var taskCreateLabels = [...]string{
	"Title", "Epic", "Tags", "Description", "More options",
	"Priority", "Effort", "Tier", "Autonomy", "Cancel", "Create",
}

type taskCreateForm struct {
	creationState
	more          bool
	inputs        [taskCreateSubmit + 1]textinput.Model
	epic          entityRef
	initialEpic   entityRef
	epics         []core.EpicSummary
	catalogLoaded bool
	catalogGen    uint64
	catalogErr    string
	unreadable    int
	query         textinput.Model
	pickCursor    int
	receipt       core.TaskCreationReceipt
}

func (f *taskCreateForm) state() *creationState { return &f.creationState }
func (taskCreateForm) noun() string             { return "task" }

type taskCreateEpicsMsg struct {
	gen      uint64
	catalog  uint64
	epics    []core.EpicSummary
	problems []core.LoadProblem
	err      error
}

type taskCreatedMsg struct {
	gen, session uint64
	receipt      core.TaskCreationReceipt
	err          error
}

func taskCreateInput(row int) bool {
	return row == taskCreateTitle || row == taskCreateTags || row == taskCreateDescription || row == taskCreateEffort
}

func taskCreateEnumOptions(row int) []string {
	var field string
	switch row {
	case taskCreatePriority:
		field = "priority"
	case taskCreateTier:
		field = "tier"
	case taskCreateAutonomy:
		field = "autonomy_level"
	default:
		return nil
	}
	return append([]string{"default"}, editEnumOptions(field)...)
}

// newTaskCreateForm only builds widgets and initial draft state. Other entity
// kinds get their own field definitions and typed request conversion, not task
// fields disguised behind a universal create schema.
func newTaskCreateForm(origin *core.Service, gen, session uint64, destination string, epic entityRef) taskCreateForm {
	f := taskCreateForm{creationState: creationState{active: true, kind: entityTasks, gen: gen, session: session,
		origin: origin, destination: destination}, epic: epic, initialEpic: epic}
	for i := range f.inputs {
		f.inputs[i] = newCreationInput()
	}
	f.inputs[taskCreateTitle].Placeholder = "Task title"
	f.inputs[taskCreateTags].Placeholder = "comma-separated, at least one"
	f.inputs[taskCreateDescription].Placeholder = "Optional: what and why"
	f.inputs[taskCreatePriority].Placeholder = "medium (core default)"
	f.inputs[taskCreateEffort].Placeholder = "Unknown (core default)"
	f.inputs[taskCreateTier].Placeholder = "3 (core default)"
	f.inputs[taskCreateAutonomy].Placeholder = "3 (core default)"
	f.query = newCreationInput()
	f.query.Prompt = "Search epics: "
	return f
}

func (f *taskCreateForm) close() {
	f.active = false
	f.mode = creationFieldsMode
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
	f.query.Blur()
}

func (f taskCreateForm) dirty() bool {
	if f.epic != f.initialEpic {
		return true
	}
	for i, input := range f.inputs {
		if (taskCreateInput(i) || len(taskCreateEnumOptions(i)) > 0) && input.Value() != "" {
			return true
		}
	}
	return false
}

func (f taskCreateForm) rows() []int {
	rows := []int{taskCreateTitle, taskCreateEpic, taskCreateTags, taskCreateDescription, taskCreateMore}
	if f.more {
		rows = append(rows, taskCreatePriority, taskCreateEffort, taskCreateTier, taskCreateAutonomy)
	}
	return append(rows, taskCreateCancel, taskCreateSubmit)
}

func (f *taskCreateForm) controls() creationControls {
	return creationControls{state: &f.creationState, inputs: f.inputs[:], rows: f.rows(),
		isText: taskCreateInput, options: taskCreateEnumOptions, dirty: f.dirty(), close: f.close,
		enumValue: func(value string) string {
			if value == "default" {
				return "" // leave task defaulting to core, not the shared host
			}
			return value
		}}
}

func (f *taskCreateForm) focus(row int) tea.Cmd {
	return f.controls().focus(row)
}

func (f *taskCreateForm) advance(delta int) tea.Cmd {
	return f.controls().advance(delta)
}

func (f *taskCreateForm) updateInput(msg tea.Msg) tea.Cmd {
	if !f.editable() {
		return nil
	}
	var cmd tea.Cmd
	if f.mode == creationPickerMode {
		if _, ok := msg.(tea.PasteMsg); ok {
			f.pickCursor = 0
		}
		f.query, cmd = updateCreationInput(f.query, msg)
	} else if taskCreateInput(f.cursor) {
		f.inputs[f.cursor], cmd = updateCreationInput(f.inputs[f.cursor], msg)
	}
	return f.widgetCmd(cmd)
}

func (f taskCreateForm) matches() []core.EpicSummary {
	if !f.catalogLoaded || f.catalogErr != "" {
		return nil
	}
	query := strings.ToLower(strings.TrimSpace(f.query.Value()))
	var out []core.EpicSummary
	for _, epic := range f.epics {
		if strings.Contains(strings.ToLower(epic.Epic.ID+" "+epic.Epic.Description+" "+epic.Source.ID), query) {
			out = append(out, epic)
		}
	}
	return out
}

func (f *taskCreateForm) handleKey(m *Model, msg tea.KeyPressMsg) tea.Cmd {
	if f.saving {
		if f.displaced && msg.String() == "y" {
			return f.copyRecovery()
		}
		return nil // an issued write is not cancellable; Ctrl+C remains the shell escape
	}
	if f.locked {
		switch msg.String() {
		case "y":
			return f.copyRecovery()
		case "enter":
			if !f.displaced && f.receipt.Committed && f.receipt.Task.ID != "" && !f.receipt.DryRun {
				receipt := f.receipt
				f.close()
				return m.landCreatedTask(receipt)
			}
			f.close()
		case "esc":
			f.close()
		}
		return nil
	}
	if f.mode == creationPickerMode {
		switch msg.String() {
		case "ctrl+r":
			f.catalogLoaded = false
			return f.loadEpics()
		case "esc":
			f.inputGen++
			f.mode = creationFieldsMode
			f.query.Blur()
		case "up":
			f.pickCursor = max(f.pickCursor-1, 0)
		case "down":
			f.pickCursor = min(f.pickCursor+1, max(len(f.matches())-1, 0))
		case "enter":
			matches := f.matches()
			if len(matches) > 0 {
				ep := matches[min(f.pickCursor, len(matches)-1)]
				f.epic = entityRef{key: ep.Source.ID, label: ep.Epic.ID}
				f.mode, f.err = creationFieldsMode, ""
				f.query.Blur()
				return f.advance(1)
			}
		default:
			f.pickCursor = 0
			return f.updateInput(msg)
		}
		return nil
	}
	if handled, cmd := f.controls().handleKey(msg); handled {
		return cmd
	}
	switch msg.String() {
	case "enter":
		switch f.cursor {
		case taskCreateEpic:
			f.inputGen++
			f.mode = creationPickerMode
			f.query.SetValue("")
			f.pickCursor = 0
			return f.widgetCmd(f.query.Focus())
		case taskCreateMore:
			f.more = !f.more
		case taskCreateCancel:
			f.controls().cancel()
		case taskCreateSubmit:
			return m.submitTaskCreate()
		default:
			return f.advance(1)
		}
		return nil
	}
	return f.updateInput(msg) // q/h/j/k are ordinary text here
}

func (f taskCreateForm) recoveryText() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Task creation in %s\nEpic: %s (%s)\n", f.destination, f.epic.label, f.epic.key)
	for i, input := range f.inputs {
		if taskCreateInput(i) || len(taskCreateEnumOptions(i)) > 0 {
			fmt.Fprintf(&b, "%s: %s\n", taskCreateLabels[i], input.Value())
		}
	}
	fmt.Fprintf(&b, "Committed: %t\nTask ID: %s\nDestination: %s\n%s", f.receipt.Committed, f.receipt.Task.ID, f.receipt.Local.CommittedPath, f.err)
	return b.String()
}

func (f taskCreateForm) copyRecovery() tea.Cmd {
	return copyCreationRecovery(f.recoveryText())
}
