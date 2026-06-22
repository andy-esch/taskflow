package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// Inline field editing (the human face of `task set`): `e` opens a picker of the
// typed, editable task fields; choosing one opens a focused widget (text input or
// enum cursor); submit fires Service.SetFields — the SAME validated write `task
// set` uses, so the TUI is a third mutation face that adds no new validation path.
// Status is deliberately absent (status==directory; that's the `a` action menu).

// fieldKind selects the widget an editable field uses.
type fieldKind int

const (
	fieldText fieldKind = iota // free text (description, effort) or a comma-list (tags)
	fieldEnum                  // a fixed option set (priority, tier)
)

// editField is one inline-editable task field: its frontmatter key (for
// SetFields), display label, widget kind, options (enum only), and current value
// (for prefill + display).
type editField struct {
	key     string
	label   string
	kind    fieldKind
	options []string
	current string
}

// editableFields are the typed task fields the TUI edits in place, in picker order.
// Status is excluded by design. Tags are a comma-list the SetFields coercion turns
// into a YAML list; tier's "1".."5" coerces to an int — the same coercion `task
// set` relies on, so the GUI and the agent face can't drift.
func editableFields(t domain.Task) []editField {
	return []editField{
		{key: "description", label: "description", kind: fieldText, current: t.Description},
		{key: "priority", label: "priority", kind: fieldEnum, options: []string{"high", "medium", "low"}, current: t.Priority},
		{key: "tags", label: "tags", kind: fieldText, current: strings.Join(t.Tags, ", ")},
		{key: "effort", label: "effort", kind: fieldText, current: t.Effort},
		{key: "tier", label: "tier", kind: fieldEnum, options: []string{"1", "2", "3", "4", "5"}, current: tierStr(t.Tier)},
	}
}

func tierStr(n int) string {
	if n == 0 {
		return ""
	}
	return strconv.Itoa(n)
}

// editMenu is the inline field-edit modal. Phase 1 picks a field; phase 2 edits the
// chosen one. Modal like actionMenu: the model routes every key here while active
// and floats it over the body (see overlay.go's editModal marker).
type editMenu struct {
	active  bool
	slug    string
	fields  []editField
	cursor  int             // field-picker cursor (phase 1)
	editing bool            // false: picking a field; true: editing the selected one
	input   textinput.Model // phase 2, text fields
	optCur  int             // phase 2, enum option cursor
}

// open shows the field picker for a task.
func (e *editMenu) open(t domain.Task) {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Width = 40
	*e = editMenu{active: true, slug: t.Slug, fields: editableFields(t), input: ti}
}

func (e *editMenu) close() {
	e.active = false
	e.input.Blur()
}

func (e *editMenu) cur() editField { return e.fields[e.cursor] }

func (e *editMenu) move(d int) {
	if n := len(e.fields); n > 0 {
		e.cursor = ((e.cursor+d)%n + n) % n
	}
}

// beginEdit enters phase 2 for the selected field, seeding the widget with the
// current value (the enum cursor on the current option, the text input prefilled).
func (e *editMenu) beginEdit() tea.Cmd {
	e.editing = true
	f := e.cur()
	if f.kind == fieldEnum {
		e.optCur = indexOf(f.options, f.current)
		if e.optCur < 0 {
			e.optCur = 0
		}
		return nil
	}
	e.input.SetValue(f.current)
	e.input.CursorEnd()
	return e.input.Focus()
}

func (e *editMenu) optMove(d int) {
	if n := len(e.cur().options); n > 0 {
		e.optCur = ((e.optCur+d)%n + n) % n
	}
}

// value is the submitted value for the field being edited.
func (e *editMenu) value() string {
	if e.cur().kind == fieldEnum {
		return e.cur().options[e.optCur]
	}
	return strings.TrimSpace(e.input.Value())
}

func indexOf(opts []string, v string) int {
	for i, o := range opts {
		if o == v {
			return i
		}
	}
	return -1
}

// handleEditKey drives the modal: phase 1 picks a field, phase 2 edits it. Enter
// submits via SetFields, Esc backs out a level. Mutates the model copy (the modal
// loop passes &m); ForceQuit is the handleKey preamble's job.
func (m *Model) handleEditKey(msg tea.KeyMsg) tea.Cmd {
	if !m.edit.editing {
		switch msg.String() {
		case "j", "down":
			m.edit.move(1)
		case "k", "up":
			m.edit.move(-1)
		case "enter", "l":
			return m.edit.beginEdit()
		case "esc", "q":
			m.edit.close()
		}
		return nil
	}
	if m.edit.cur().kind == fieldEnum {
		switch msg.String() {
		case "j", "down", "l":
			m.edit.optMove(1)
		case "k", "up", "h":
			m.edit.optMove(-1)
		case "enter":
			return m.submitEdit()
		case "esc":
			m.edit.editing = false // back to the picker
		}
		return nil
	}
	// Text field: Enter submits, Esc backs to the picker, everything else types.
	switch msg.Type {
	case tea.KeyEnter:
		return m.submitEdit()
	case tea.KeyEsc:
		m.edit.editing = false
		m.edit.input.Blur()
		return nil
	}
	var cmd tea.Cmd
	m.edit.input, cmd = m.edit.input.Update(msg)
	return cmd
}

// submitEdit fires the SetFields write for the edited field and closes the modal.
// Core re-validates (enum/key-order/surgical frontmatter); an invalid value comes
// back as actionErrMsg (red flash, nothing written).
func (m *Model) submitEdit() tea.Cmd {
	f := m.edit.cur()
	slug, val := m.edit.slug, m.edit.value()
	m.edit.close()
	return setFieldCmd(m.svc, slug, f.key, val)
}

// setFieldCmd runs the field write off the event loop, reporting success
// (editedMsg → flash + reload) or the core validation error (actionErrMsg → flash,
// no reload). force=false, dryRun=false: a real, fully-validated set.
func setFieldCmd(svc *core.Service, slug, key, value string) tea.Cmd {
	return func() tea.Msg {
		if _, err := svc.SetFields(slug, map[string]any{key: value}, false, false); err != nil {
			return actionErrMsg{slug: slug, err: err}
		}
		return editedMsg{slug: slug, field: key}
	}
}

// view renders the picker (phase 1) or the field widget (phase 2) as a centered box
// + hint, ready to composite over the body with overlay(). Clamped to (maxW, maxH).
func (e editMenu) view(maxW, maxH int) string {
	var b strings.Builder
	if !e.editing {
		b.WriteString(actionHeading.Render("edit " + truncate(e.slug, max(maxW-8, 12))))
		b.WriteString("\n\n")
		for i, f := range e.fields {
			val := f.current
			if val == "" {
				val = dim("(empty)")
			}
			label := f.label + dim(": ") + truncate(val, max(maxW-22, 8))
			if i == e.cursor {
				b.WriteString(selectedStyle.Render("› ") + label + "\n")
			} else {
				b.WriteString("  " + label + "\n")
			}
		}
		box := actionBorder.Render(strings.TrimRight(b.String(), "\n"))
		hint := dim("↑↓/jk select · ⏎ edit · esc cancel")
		return clampBox(lipgloss.JoinVertical(lipgloss.Center, box, hint), maxW, maxH)
	}
	f := e.cur()
	b.WriteString(actionHeading.Render("edit " + f.label))
	b.WriteString("\n\n")
	if f.kind == fieldEnum {
		for i, o := range f.options {
			if i == e.optCur {
				b.WriteString(selectedStyle.Render("› ") + o + "\n")
			} else {
				b.WriteString("  " + o + "\n")
			}
		}
		box := actionBorder.Render(strings.TrimRight(b.String(), "\n"))
		hint := dim("↑↓/jk select · ⏎ apply · esc back")
		return clampBox(lipgloss.JoinVertical(lipgloss.Center, box, hint), maxW, maxH)
	}
	b.WriteString(e.input.View())
	box := actionBorder.Render(b.String())
	hint := dim("⏎ apply · esc back")
	return clampBox(lipgloss.JoinVertical(lipgloss.Center, box, hint), maxW, maxH)
}
