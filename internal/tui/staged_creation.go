package tui

import (
	"errors"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// Zero keeps Bubbles' default ~530ms virtual-cursor blink in production;
// package tests lower it so synchronous command settling and blink-continuation
// checks do not sleep 530ms on every focus or edit.
var creationBlinkSpeed time.Duration

func newCreationInput() textinput.Model {
	input := textinput.New()
	input.Prompt, input.CharLimit = "", 0
	if creationBlinkSpeed > 0 {
		styles := input.Styles()
		styles.Cursor.BlinkSpeed = creationBlinkSpeed
		input.SetStyles(styles)
	}
	return input
}

// Staged creation shares primary-adapter mechanics, not an entity schema or
// universal mutation port. Each form retains fixed-size widget structs,
// typed parameters/receipts and its own submission/recovery policy.
// See docs/architecture/tui.md#staged-task-creation.
type creationMode uint8

const (
	creationFieldsMode creationMode = iota
	creationPickerMode
	creationEnumMode
	creationDiscardMode
)

// Mode is exclusive interaction state. Saving/locked/displaced are orthogonal:
// an issued origin write can remain pending after workspace displacement.
type creationState struct {
	active, saving, locked, displaced bool
	kind                              entityKind
	mode                              creationMode
	cursor, enumCursor                int
	gen, session, inputGen            uint64
	origin                            *core.Service
	destination, err                  string
}

func (f *creationState) displace() {
	if !f.active {
		return
	}
	if !f.displaced {
		previous := f.err
		f.err = "Workspace changed. Draft retained for recovery; it cannot be submitted here."
		if previous != "" {
			f.err += " Original state: " + previous
		}
	}
	f.displaced, f.locked = true, true
	f.mode = creationFieldsMode
}

func (f creationState) editable() bool {
	return f.active && !f.saving && !f.locked && (f.mode == creationFieldsMode || f.mode == creationPickerMode)
}

func (f *creationState) focus(row int, inputs []textinput.Model, text bool) tea.Cmd {
	f.inputGen++
	f.mode, f.cursor = creationFieldsMode, row
	for i := range inputs {
		inputs[i].Blur()
	}
	if text && f.editable() {
		return f.widgetCmd(inputs[row].Focus())
	}
	return nil
}

type creationWidgetMsg struct {
	kind          entityKind
	gen, inputGen uint64
	msg           tea.Msg
}

// Focus-issued blink ticks are maintenance, not text input. All other delayed
// widget results (including opaque native paste/error messages) require the
// visible input owner. The textinput focus command produces cursor.BlinkMsg;
// canceled ticks need no continuation. Never identify private paste types by name.
func creationWidgetMaintenance(msg tea.Msg) bool {
	_, blink := msg.(cursor.BlinkMsg)
	return blink
}

// Copying a textinput.Model does not detach its internal rune buffer. Mid-string
// insert/delete can otherwise mutate the prior Model's draft. SetValue makes a
// fresh buffer through the public widget API, preserving cursor/focus/settings;
// preserve the previous diagnostic until Update applies the actual event.
// Pure blink ticks neither edit text nor need a proportional copy.
func updateCreationInput(input textinput.Model, msg tea.Msg) (textinput.Model, tea.Cmd) {
	if !creationWidgetMaintenance(msg) {
		err := input.Err
		input.SetValue(input.Value())
		input.Err = err
	}
	return input.Update(msg)
}

func (f creationState) widgetCmd(cmd tea.Cmd) tea.Cmd {
	return routeCreationWidget(f.kind, f.gen, f.inputGen, cmd)
}

// Capture routing tokens only: delayed clipboard/cursor commands must not retain
// the form or its Service. Kind, form and input generations all delimit delivery.
func routeCreationWidget(kind entityKind, gen, inputGen uint64, cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		if batch, ok := msg.(tea.BatchMsg); ok {
			wrapped := make([]tea.Cmd, len(batch))
			for i, child := range batch {
				wrapped[i] = routeCreationWidget(kind, gen, inputGen, child)
			}
			return tea.BatchMsg(wrapped)
		}
		return creationWidgetMsg{kind: kind, gen: gen, inputGen: inputGen, msg: msg}
	}
}

// A short-lived controls value borrows one form's fixed widget array. Text edits
// detach their internal buffer through updateCreationInput, not an array copy.
// Its callbacks are synchronous UI mechanics, never retained by tea.Cmd.
type creationControls struct {
	state   *creationState
	inputs  []textinput.Model
	rows    []int
	isText  func(int) bool
	options func(int) []string // nil for text-only forms
	// Optional per-kind presentation mapping, e.g. a task's displayed default
	// choice stages an empty value. The host does not interpret domain values.
	enumValue func(string) string
	dirty     bool
	close     func()
}

// The shell hosts concrete UI drafts through this mechanical seam. Domain
// requests, receipts and write policy remain in their per-kind controllers.
type stagedCreation interface {
	noun() string
	state() *creationState
	dirty() bool
	close()
	updateInput(tea.Msg) tea.Cmd
	handleKey(*Model, tea.KeyPressMsg) tea.Cmd
	view(*styles, int, int) string
}

func (c creationControls) enumOptions(row int) []string {
	if c.options == nil {
		return nil // text-only forms have no enum policy
	}
	return c.options(row)
}

func (c creationControls) focus(row int) tea.Cmd {
	return c.state.focus(row, c.inputs, c.isText(row))
}

func (c creationControls) advance(delta int) tea.Cmd {
	index := slices.Index(c.rows, c.state.cursor)
	return c.focus(c.rows[(index+delta+len(c.rows))%len(c.rows)])
}

func (c creationControls) cancel() {
	if c.dirty {
		c.state.mode = creationDiscardMode
	} else {
		c.close()
	}
}

// Fields, enum staging and dirty cancellation are shared. Per-kind controllers
// handle pending/recovery policy, custom pickers and explicit action rows.
func (c creationControls) handleKey(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	f := c.state
	if f.mode == creationDiscardMode {
		switch msg.String() {
		case "tab", "shift+tab", "left", "right":
		case "y", "Y":
			c.close()
		default:
			f.mode = creationFieldsMode
		}
		return true, nil
	}
	if f.mode == creationEnumMode {
		options := c.enumOptions(f.cursor)
		if len(options) == 0 {
			f.mode = creationFieldsMode
			return true, nil
		}
		switch msg.String() {
		case "j", "down", "l", "right":
			f.enumCursor = (f.enumCursor + 1) % len(options)
		case "k", "up", "h", "left":
			f.enumCursor = (f.enumCursor - 1 + len(options)) % len(options)
		case "enter", "tab", "shift+tab":
			value := options[f.enumCursor]
			if c.enumValue != nil {
				value = c.enumValue(value)
			}
			c.inputs[f.cursor].SetValue(value)
			if msg.String() == "shift+tab" {
				return true, c.advance(-1)
			}
			return true, c.advance(1)
		case "esc":
			f.mode = creationFieldsMode
		}
		return true, nil
	}
	switch msg.String() {
	case "q":
		if c.isText(f.cursor) {
			return false, nil
		}
		c.cancel()
		return true, nil
	case "esc":
		c.cancel()
		return true, nil
	case "tab":
		return true, c.advance(1)
	case "shift+tab":
		return true, c.advance(-1)
	case "enter":
		if options := c.enumOptions(f.cursor); len(options) > 0 {
			f.inputGen++
			f.mode = creationEnumMode
			f.enumCursor = max(slices.Index(options, c.inputs[f.cursor].Value()), 0)
			return true, nil
		}
		if c.isText(f.cursor) {
			return true, c.advance(1)
		}
	}
	return false, nil
}

func (m *Model) creationDraft() stagedCreation {
	return m.creation.draft()
}

func (m *Model) activeCreation() *creationState {
	return m.creation.active()
}

func (m *Model) creationDestination() string {
	destination := m.spaceName()
	if m.workspace.Checkout != "" {
		destination += " · " + m.workspace.Checkout
	}
	return destination
}

// Identity is supplied by the per-kind typed receipt, never its optional path.
func (m *Model) landCreatedEntity(kind entityKind, ref entityRef, view string, localEditor bool) tea.Cmd {
	i := indexOfKind(m.tabs, kind)
	m.exitDashboard(i)
	m.zoom, m.immersiveZoom = false, false
	m.setFocus(focusList)
	m.recomputeLayout()
	tab := m.tabs[i]
	tab.list.ResetFilter()
	tab.statusView = view
	tab.restore, tab.restoreWiden = ref, true
	m.flash, m.flashErr = "created "+ref.label, false
	if localEditor {
		m.flash += " — E opens its scaffold in your editor"
	}
	return m.reloadAll()
}

func creationTags(value string) []string {
	var tags []string
	for _, tag := range strings.Split(value, ",") {
		if tag = strings.TrimSpace(tag); tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// These sentinels establish an uncommitted refusal only when the typed receipt
// carries no progress and the error is not a finalization error. Each kind's
// result handler must make that decision before offering a deliberate retry.
func creationKnownRefusal(err error) bool {
	return errors.Is(err, domain.ErrValidation) || errors.Is(err, domain.ErrNotFound) ||
		errors.Is(err, domain.ErrAmbiguous) || errors.Is(err, domain.ErrConflict) || errors.Is(err, core.ErrReadOnlyPersistence)
}

func copyCreationRecovery(text string) tea.Cmd {
	return func() tea.Msg { return copyToClipboard(text)() }
}
