package tui

import (
	"fmt"
	"slices"
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

// One interaction owns the keys at a time. Saving/recovery and displacement are
// separate write/authority state: for example, an issued write can still be
// pending after its origin workspace has been displaced.
type taskCreateMode uint8

const (
	taskCreateFields taskCreateMode = iota
	taskCreateEpicPicker
	taskCreateEnumPicker
	taskCreateDiscard
)

type taskCreateForm struct {
	active        bool
	more          bool
	mode          taskCreateMode
	enumCursor    int
	saving        bool
	locked        bool
	displaced     bool
	gen, session  uint64
	inputGen      uint64 // delayed paste/cursor results belong to one focused widget
	origin        *core.Service
	destination   string
	inputs        [taskCreateSubmit + 1]textinput.Model
	cursor        int
	epic          entityRef
	initialEpic   entityRef
	epics         []core.EpicSummary
	catalogLoaded bool
	catalogGen    uint64
	catalogErr    string
	unreadable    int
	query         textinput.Model
	pickCursor    int
	err           string
	receipt       core.TaskCreationReceipt
}

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

type taskCreateWidgetMsg struct {
	gen, inputGen uint64
	msg           tea.Msg
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
	f := taskCreateForm{active: true, gen: gen, session: session,
		origin: origin, destination: destination, epic: epic, initialEpic: epic}
	for i := range f.inputs {
		f.inputs[i] = textinput.New()
		f.inputs[i].Prompt = ""
		f.inputs[i].CharLimit = 0 // core owns validation; resizing never changes input
	}
	f.inputs[taskCreateTitle].Placeholder = "Task title"
	f.inputs[taskCreateTags].Placeholder = "comma-separated, at least one"
	f.inputs[taskCreateDescription].Placeholder = "Optional: what and why"
	f.inputs[taskCreatePriority].Placeholder = "medium (core default)"
	f.inputs[taskCreateEffort].Placeholder = "Unknown (core default)"
	f.inputs[taskCreateTier].Placeholder = "3 (core default)"
	f.inputs[taskCreateAutonomy].Placeholder = "3 (core default)"
	f.query = textinput.New()
	f.query.Prompt = "Search epics: "
	f.query.CharLimit = 0
	return f
}

func (f *taskCreateForm) close() {
	f.active = false
	f.mode = taskCreateFields
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
	f.query.Blur()
}

func (f *taskCreateForm) displace() {
	if !f.active {
		return
	}
	f.displaced, f.locked = true, true
	f.mode = taskCreateFields
	f.err = "Workspace changed. Draft retained for recovery; it cannot be submitted here."
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

func (f *taskCreateForm) focus(row int) tea.Cmd {
	f.inputGen++
	f.mode = taskCreateFields
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
	f.cursor = row
	if taskCreateInput(row) && !f.locked && !f.saving {
		return f.widgetCmd(f.inputs[row].Focus())
	}
	return nil
}

func (f *taskCreateForm) advance(delta int) tea.Cmd {
	rows := f.rows()
	index := slices.Index(rows, f.cursor)
	return f.focus(rows[(index+delta+len(rows))%len(rows)])
}

func (f taskCreateForm) widgetCmd(cmd tea.Cmd) tea.Cmd {
	return routeTaskCreateWidget(f.gen, f.inputGen, cmd)
}

// Capture only routing tokens, not a complete draft and its retained Service.
// Cursor/paste timers can outlive the field or form that created them.
func routeTaskCreateWidget(gen, inputGen uint64, cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		if batch, ok := msg.(tea.BatchMsg); ok {
			wrapped := make([]tea.Cmd, len(batch))
			for i, child := range batch {
				wrapped[i] = routeTaskCreateWidget(gen, inputGen, child)
			}
			return tea.BatchMsg(wrapped)
		}
		return taskCreateWidgetMsg{gen: gen, inputGen: inputGen, msg: msg}
	}
}

func (f *taskCreateForm) updateInput(msg tea.Msg) tea.Cmd {
	if f.saving || f.locked || (f.mode != taskCreateFields && f.mode != taskCreateEpicPicker) {
		return nil
	}
	var cmd tea.Cmd
	if f.mode == taskCreateEpicPicker {
		if _, ok := msg.(tea.PasteMsg); ok {
			f.pickCursor = 0
		}
		f.query, cmd = f.query.Update(msg)
	} else if taskCreateInput(f.cursor) {
		f.inputs[f.cursor], cmd = f.inputs[f.cursor].Update(msg)
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

func (m *Model) handleTaskCreateKey(msg tea.KeyPressMsg) tea.Cmd {
	f := &m.createTask
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
			if !f.displaced && f.receipt.Committed && f.receipt.Task.ID != "" {
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
	if f.mode == taskCreateDiscard {
		switch msg.String() {
		case "tab", "shift+tab", "left", "right":
			// Keep editing is the default. Discard needs an explicit y, not Enter.
		case "y", "Y":
			f.close()
		default:
			f.mode = taskCreateFields
		}
		return nil
	}
	if f.mode == taskCreateEpicPicker {
		switch msg.String() {
		case "ctrl+r":
			f.catalogLoaded = false
			return f.loadEpics()
		case "esc":
			f.inputGen++
			f.mode = taskCreateFields
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
				f.mode, f.err = taskCreateFields, ""
				f.query.Blur()
				return f.advance(1)
			}
		default:
			f.pickCursor = 0
			return f.updateInput(msg)
		}
		return nil
	}
	if f.mode == taskCreateEnumPicker {
		options := taskCreateEnumOptions(f.cursor)
		switch msg.String() {
		case "j", "down", "l", "right":
			f.enumCursor = (f.enumCursor + 1) % len(options)
		case "k", "up", "h", "left":
			f.enumCursor = (f.enumCursor - 1 + len(options)) % len(options)
		case "enter", "tab", "shift+tab":
			value := options[f.enumCursor]
			if value == "default" {
				value = "" // preserve core defaults rather than pinning them in UI state
			}
			f.inputs[f.cursor].SetValue(value)
			f.mode = taskCreateFields
			if msg.String() == "shift+tab" {
				return f.advance(-1)
			}
			return f.advance(1)
		case "esc":
			f.mode = taskCreateFields // discard only the unaccepted choice
		}
		return nil
	}
	switch msg.String() {
	case "q":
		if taskCreateInput(f.cursor) {
			return f.updateInput(msg)
		}
		if f.dirty() {
			f.mode = taskCreateDiscard
		} else {
			f.close()
		}
		return nil
	case "esc":
		if f.dirty() {
			f.mode = taskCreateDiscard
		} else {
			f.close()
		}
		return nil
	case "tab":
		return f.advance(1)
	case "shift+tab":
		return f.advance(-1)
	case "enter":
		if options := taskCreateEnumOptions(f.cursor); len(options) > 0 {
			f.inputGen++
			f.mode = taskCreateEnumPicker
			f.enumCursor = max(slices.Index(options, f.inputs[f.cursor].Value()), 0)
			return nil
		}
		switch f.cursor {
		case taskCreateEpic:
			f.inputGen++
			f.mode = taskCreateEpicPicker
			f.query.SetValue("")
			f.pickCursor = 0
			return f.widgetCmd(f.query.Focus())
		case taskCreateMore:
			f.more = !f.more
		case taskCreateCancel:
			if f.dirty() {
				f.mode = taskCreateDiscard
			} else {
				f.close()
			}
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
	text := f.recoveryText()
	return func() tea.Msg {
		// Even native clipboard-tool discovery belongs off the reducer.
		return copyToClipboard(text)()
	}
}
