package tui

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// Inline field editing (the human face of `task set`): `e` opens a single form
// panel listing the typed, editable task fields with what each means; the active
// field's editor shows IN PLACE — an enum's options inline in its row, a long field
// (description) in a taller word-wrapped box below the list. Submit fires
// Service.SetFields (the SAME validated write `task set` uses), so the TUI is a
// third mutation face that adds no new validation path. Status is absent by design
// (status==directory; that's the `m` action menu).

// fieldKind selects the widget an editable field uses.
type fieldKind int

const (
	fieldText     fieldKind = iota // single-line text (tags as a comma-list, effort)
	fieldLongText                  // word-wrapped multi-row box (description)
	fieldEnum                      // a fixed option set shown inline (priority, tier)
)

// editField is one inline-editable task field: its frontmatter key (for
// SetFields), display label, what it means (desc, from the entity descriptor),
// widget kind, options (enum only), and current value (for prefill + display).
type editField struct {
	key     string
	label   string
	desc    string
	kind    fieldKind
	options []string
	current string
}

// editableFields are the typed task fields the TUI edits in place, in form order.
// Status is excluded by design. Tags ride as a comma-list and tier as a string —
// the SetFields coercion turns them into a YAML list / int, the same path `task
// set` uses, so the GUI and the agent face can't drift. Field meanings come from
// the entity descriptor (domain.AuthoringFields), the same source `schema task`
// shows, so they can't drift either.
func editableFields(t domain.Task) []editField {
	d := fieldDescs()
	return []editField{
		{key: "description", label: "description", desc: d["description"], kind: fieldLongText, current: t.Description},
		{key: "priority", label: "priority", desc: d["priority"], kind: fieldEnum, options: []string{"high", "medium", "low"}, current: t.Priority},
		{key: "tags", label: "tags", desc: d["tags"], kind: fieldText, current: strings.Join(t.Tags, ", ")},
		{key: "effort", label: "effort", desc: d["effort"], kind: fieldText, current: t.Effort},
		{key: "tier", label: "tier", desc: d["tier"], kind: fieldEnum, options: []string{"1", "2", "3", "4", "5"}, current: tierStr(t.Tier)},
	}
}

// fieldDescs maps a task field name to its one-line meaning from the descriptor.
func fieldDescs() map[string]string {
	out := map[string]string{}
	docs, _ := domain.AuthoringFields("task")
	for _, doc := range docs {
		out[doc.Name] = doc.Description
	}
	return out
}

func tierStr(n int) string {
	if n == 0 {
		return ""
	}
	return strconv.Itoa(n)
}

// editMenu is the inline field-edit form. cursor selects a field; editing flips
// that field into its widget (input/area for text, optCur for enums). Modal like
// actionMenu: the model routes every key here while active and floats it over the
// body (see overlay.go's editModal marker).
type editMenu struct {
	active  bool
	slug    string
	fields  []editField
	cursor  int             // selected field
	editing bool            // false: navigating fields; true: editing the selected one
	input   textinput.Model // single-line text fields
	area    textarea.Model  // the word-wrapped long-text field (description)
	optCur  int             // enum option cursor
	err     string          // last submit's validation error, shown until the next edit
}

// open shows the form for a task.
func (e *editMenu) open(t domain.Task) {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.SetWidth(36)

	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.CharLimit = domain.MaxDescriptionLen
	ta.SetWidth(48)
	ta.SetHeight(4)
	ta.KeyMap.InsertNewline.SetEnabled(false) // Enter submits; description stays one line

	*e = editMenu{active: true, slug: t.Slug, fields: editableFields(t), input: ti, area: ta}
}

func (e *editMenu) close() {
	e.active = false
	e.err = ""
	e.input.Blur()
	e.area.Blur()
}

// applied returns the form to field navigation after a confirmed write, refreshing
// the just-set field's displayed value and clearing any prior error.
func (e *editMenu) applied(field, val string) {
	e.setCurrent(field, val)
	e.editing = false
	e.err = ""
	e.input.Blur()
	e.area.Blur()
}

// stopEditing drops back to the field picker (Esc from a widget), clearing the error.
func (e *editMenu) stopEditing() {
	e.editing = false
	e.err = ""
	e.input.Blur()
	e.area.Blur()
}

func (e *editMenu) cur() editField { return e.fields[e.cursor] }

func (e *editMenu) move(d int) {
	if n := len(e.fields); n > 0 {
		e.cursor = ((e.cursor+d)%n + n) % n
	}
}

// beginEdit enters editing for the selected field, seeding the widget with the
// current value (the enum cursor on the current option, the input/area prefilled).
func (e *editMenu) beginEdit() tea.Cmd {
	e.editing = true
	f := e.cur()
	switch f.kind {
	case fieldEnum:
		if e.optCur = indexOf(f.options, f.current); e.optCur < 0 {
			e.optCur = 0
		}
		return nil
	case fieldLongText:
		e.area.SetValue(f.current)
		e.area.CursorEnd()
		return e.area.Focus()
	default:
		e.input.SetValue(f.current)
		e.input.CursorEnd()
		return e.input.Focus()
	}
}

func (e *editMenu) optMove(d int) {
	if n := len(e.cur().options); n > 0 {
		e.optCur = ((e.optCur+d)%n + n) % n
	}
}

// value is the submitted value for the field being edited.
func (e editMenu) value() string {
	switch e.cur().kind {
	case fieldEnum:
		return e.cur().options[e.optCur]
	case fieldLongText:
		return strings.TrimSpace(e.area.Value())
	default:
		return strings.TrimSpace(e.input.Value())
	}
}

func indexOf(opts []string, v string) int {
	for i, o := range opts {
		if o == v {
			return i
		}
	}
	return -1
}

// handleEditKey drives the form: navigate fields, then edit the selected one. Enter
// submits via SetFields, Esc backs out a level. Mutates the model copy (the modal
// loop passes &m); ForceQuit is the handleKey preamble's job.
func (m *Model) handleEditKey(msg tea.KeyPressMsg) tea.Cmd {
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
	switch m.edit.cur().kind {
	case fieldEnum:
		switch msg.String() {
		case "j", "down", "l", "right":
			m.edit.optMove(1)
		case "k", "up", "h", "left":
			m.edit.optMove(-1)
		case "enter":
			return m.submitEdit()
		case "esc":
			m.edit.stopEditing() // back to field navigation
		}
		return nil
	case fieldLongText:
		switch msg.String() {
		case "enter":
			return m.submitEdit()
		case "esc":
			m.edit.stopEditing()
			return nil
		}
		m.edit.err = "" // any edit clears the stale validation error
		var cmd tea.Cmd
		m.edit.area, cmd = m.edit.area.Update(msg)
		return cmd
	default:
		switch msg.String() {
		case "enter":
			return m.submitEdit()
		case "esc":
			m.edit.stopEditing()
			return nil
		}
		m.edit.err = ""
		var cmd tea.Cmd
		m.edit.input, cmd = m.edit.input.Update(msg)
		return cmd
	}
}

// submitEdit fires the SetFields write for the edited field and waits on the result
// before leaving the field: on success (editedMsg) the form returns to the picker
// with the new value; on a core validation error (actionErrMsg) it stays on the
// field, keeping what was typed, and shows the error so the user can fix it in place
// (see the editedMsg/actionErrMsg handlers). The field stays focused meanwhile.
func (m *Model) submitEdit() tea.Cmd {
	m.edit.err = ""
	return setFieldCmd(m.svc, m.edit.slug, m.edit.cur().key, m.edit.value())
}

// setCurrent updates the form's displayed value for a field after a confirmed
// write, so the just-edited field isn't stale while the form stays open.
func (e *editMenu) setCurrent(key, val string) {
	for i := range e.fields {
		if e.fields[i].key == key {
			e.fields[i].current = val
		}
	}
}

// setFieldCmd runs the field write off the event loop, reporting success
// (editedMsg → flash + reload) or the core validation error (actionErrMsg → flash,
// no reload). force=false, dryRun=false: a real, fully-validated set.
func setFieldCmd(svc *core.Service, slug, key, value string) tea.Cmd {
	return func() tea.Msg {
		if _, err := svc.SetFields(slug, map[string]any{key: value}, false, false); err != nil {
			return actionErrMsg{slug: slug, err: err}
		}
		return editedMsg{slug: slug, field: key, value: value}
	}
}

// --- view ---

var editAreaBox = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("8")).
	Padding(0, 1)

const editLabelW = 12 // field-label column

// view renders the whole form (field list + the active field's inline editor) as a
// centered box + hint, ready to composite over the body. Clamped to (maxW, maxH).
func (e editMenu) view(maxW, maxH int) string {
	innerW := max(maxW-8, 28) // inside the box border + padding
	var rows []string
	for i, f := range e.fields {
		marker := "  "
		if i == e.cursor {
			marker = selectedStyle.Render("› ")
		}
		rows = append(rows, marker+padField(f.label, editLabelW)+"  "+e.cell(i, f, innerW))
	}
	body := actionHeading.Render("edit "+truncate(e.slug, max(innerW-6, 8))) + "\n\n" + strings.Join(rows, "\n")
	// The long-text field gets a roomy word-wrapped box below the list.
	if e.editing && e.cur().kind == fieldLongText {
		body += "\n\n" + editAreaBox.Render(e.area.View())
	}
	box := actionBorder.Render(body)
	lines := []string{box}
	if e.err != "" {
		lines = append(lines, fg(theme.ColorRed, "✘ "+truncate(e.err, max(maxW-2, 8))))
	}
	lines = append(lines, e.hint())
	return clampBox(lipgloss.JoinVertical(lipgloss.Left, lines...), maxW, maxH)
}

// cell renders field i's value column: the inline editor when it's the one being
// edited, else the current value followed by its dim description.
func (e editMenu) cell(i int, f editField, innerW int) string {
	editing := e.editing && i == e.cursor
	switch {
	case editing && f.kind == fieldEnum:
		return enumInline(f.options, e.optCur)
	case editing && f.kind == fieldText:
		return e.input.View()
	case editing && f.kind == fieldLongText:
		return dim("editing ↓")
	}
	const valueW = 18
	val := f.current
	if val == "" {
		val = "—"
	}
	cell := padField(val, valueW)
	if descW := innerW - editLabelW - valueW - 6; descW >= 8 && f.desc != "" {
		cell += "  " + dim(truncate(f.desc, descW))
	}
	return cell
}

// enumInline renders an enum's options on one line, the selected one bracketed +
// accented — so choosing a value happens right in the field's row, not a new pane.
func enumInline(opts []string, cur int) string {
	parts := make([]string, len(opts))
	for i, o := range opts {
		if i == cur {
			parts[i] = selectedStyle.Render("‹" + o + "›")
		} else {
			parts[i] = dim(o)
		}
	}
	return strings.Join(parts, " ")
}

func (e editMenu) hint() string {
	if !e.editing {
		return dim("↑↓ field · ⏎ edit · esc cancel")
	}
	if e.cur().kind == fieldEnum {
		return dim("←→/jk choose · ⏎ apply · esc back")
	}
	return dim("⏎ apply · esc back")
}

// padField truncates s to width w (ANSI-aware) and right-pads with spaces so the
// columns line up across rows.
func padField(s string, w int) string {
	s = truncate(s, w)
	if n := w - ansi.StringWidth(s); n > 0 {
		s += strings.Repeat(" ", n)
	}
	return s
}
