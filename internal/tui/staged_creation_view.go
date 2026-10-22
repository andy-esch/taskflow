package tui

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/theme"
)

type creationModal struct{}

func (creationModal) active(m *Model) bool { return m.activeCreation() != nil }
func (creationModal) inputVisible(m *Model) bool {
	return creationInputFits(m.modalSize())
}

// Dimensions are the modal's outer bounds after the shell gutter, before the
// action border/padding. The same rule selects resize rendering and input rights.
func creationInputFits(width, height int) bool {
	return width >= 44 && height >= 14
}

// inputOwner applies to terminal paste and opaque native clipboard results alike.
// Do not identify native input by its library-private message type.
func (o creationModal) inputOwner(m *Model) bool {
	if !o.inputVisible(m) {
		return false
	}
	// Delayed clipboard deliveries must still belong to the visible modal, not
	// merely to a live form generation underneath help or configuration.
	for _, candidate := range m.modals {
		if candidate.active(m) {
			_, owns := candidate.(creationModal)
			return owns
		}
	}
	return false
}

func (o creationModal) handlePaste(m *Model, msg tea.PasteMsg) tea.Cmd {
	if !o.inputOwner(m) {
		return nil
	}
	return m.creationDraft().updateInput(msg)
}

func (o creationModal) handleKey(m *Model, msg tea.KeyPressMsg) (bool, tea.Cmd) {
	draft := m.creationDraft()
	if !o.inputVisible(m) {
		// A resize notice must never be an invisible authoring surface.
		if msg.String() == "esc" && !draft.dirty() && !draft.state().saving {
			draft.close()
		}
		return true, nil
	}
	return true, draft.handleKey(m, msg)
}

func (creationModal) view(m *Model, w, h int) string {
	return m.creationDraft().view(m.st, w, h)
}

// Evidence is display-only and supplied by a typed per-kind controller. It does
// not interpret receipts, grant inspection authority or become a core DTO.
type creationPresentation struct {
	noun, title, subtitle, recordID, path string
	inspect                               bool
}

// A completed header mode owns the whole modal; ordinary fields receive its
// unrendered lines and dimensions. Do not overload the first line as a result.
type creationFrame struct {
	lines              []string
	innerW, maxW, maxH int
	rendered           string
}

func creationHeader(f creationState, p creationPresentation, s *styles, maxW, maxH int) creationFrame {
	innerW := max(min(84, maxW)-s.actionBorder.GetHorizontalFrameSize(), 1)
	frame := creationFrame{innerW: innerW, maxW: maxW, maxH: maxH}
	if !creationInputFits(maxW, maxH) {
		frame.rendered = clampBox(s.actionBorder.Render("New "+p.noun+" · draft retained\nResize to continue safely.\nSave to: "+truncateMiddle(f.destination, innerW)), maxW, maxH)
		return frame
	}
	lines := []string{s.actionHeading.Render("New " + p.noun), "Save to: " + truncateMiddle(f.destination, innerW-9)}
	if f.locked {
		lines[0] = s.actionHeading.Render(p.noun + " creation · recovery")
		lines = append(lines, "Create disabled · do not retry")
		if p.recordID != "" {
			lines = append(lines, "Record: "+truncateMiddle(p.recordID, innerW-8))
		}
		if p.path != "" {
			lines = append(lines, "Destination: "+truncateMiddle(p.path, innerW-13))
		}
		lines = append(lines, "Draft: "+truncate(p.title, innerW-7))
		hints := []string{"y copy full evidence · esc close"}
		if f.saving {
			hints = []string{"Origin pending · not cancellable", "y copy draft · wait for origin result"}
		} else if p.inspect && !f.displaced {
			hints = append([]string{"⏎ inspect created record"}, hints...)
		}
		// Identity and actions have priority over verbose adapter diagnostics. The
		// unabridged evidence remains available via recoveryText / y.
		lines = creationErrorLines(lines, f.err, innerW, min(3, maxH-2-len(lines)-len(hints)), s)
		for _, hint := range hints {
			lines = append(lines, s.dim(truncate(hint, innerW)))
		}
		frame.rendered = creationBox(s, lines, innerW, maxW, maxH)
		return frame
	}
	if f.saving {
		lines = append(lines, "Saving in the origin space…", s.dim("Already submitted · not cancellable · no repeat Create"))
		frame.rendered = creationBox(s, lines, innerW, maxW, maxH)
		return frame
	}
	if f.mode == creationDiscardMode {
		lines = append(lines, "Discard unsaved input?", s.selected.Render("[Keep editing]"), "y discards · Enter / Esc keeps editing")
		frame.rendered = creationBox(s, lines, innerW, maxW, maxH)
		return frame
	}
	frame.lines = append(lines, s.dim(p.subtitle))
	return frame
}

func creationErrorLines(lines []string, err string, innerW, budget int, s *styles) []string {
	if err != "" && budget > 0 {
		// Only format a bounded prefix for this viewport, not an entire external
		// diagnostic that we would immediately discard. Iterate without splitting
		// a UTF-8 rune or allocating a rune slice proportional to the full error.
		remaining := max(innerW, 1) * budget
		for i := range err {
			if remaining == 0 {
				err = err[:i]
				break
			}
			remaining--
		}
		parts := strings.Split(wrap(err, innerW), "\n")
		for _, part := range parts[:min(len(parts), budget)] {
			lines = append(lines, s.fg(theme.ColorRed, part))
		}
	}
	return lines
}

type creationFieldLayout struct {
	rows           []int
	labels         []string
	cancel, submit int
	cell           func(row, width int) string
}

func creationFieldsView(f creationState, s *styles, frame creationFrame, layout creationFieldLayout) string {
	lines, innerW, maxW, maxH := frame.lines, frame.innerW, frame.maxW, frame.maxH
	fields := layout.rows[:len(layout.rows)-2] // explicit actions stay pinned while fields scroll
	stacked := innerW < 54
	rowHeight := 1
	if stacked {
		rowHeight = 2
	}
	available := max(1, (maxH-2-len(lines)-3)/rowHeight)
	position := min(slices.Index(layout.rows, f.cursor), len(fields)-1)
	start := min(max(position-available+1, 0), max(len(fields)-available, 0))
	end := min(start+available, len(fields))
	for _, row := range fields[start:end] {
		label, cellW := fmt.Sprintf("%-13s", layout.labels[row]), max(innerW-15, 1)
		if stacked {
			label, cellW = layout.labels[row], max(innerW-2, 1)
		}
		prefix := "  "
		if row == f.cursor {
			prefix, label = s.accent("› "), s.selected.Render(label)
		}
		if stacked {
			lines = append(lines, prefix+label, "  "+truncate(layout.cell(row, cellW), cellW))
		} else {
			lines = append(lines, prefix+label+truncate(layout.cell(row, cellW), cellW))
		}
	}
	buttons := "[Cancel]  [Create]"
	switch f.cursor {
	case layout.cancel:
		buttons = s.accent("[Cancel]") + "  [Create]"
	case layout.submit:
		buttons = "[Cancel]  " + s.accent("[Create]")
	}
	if start > 0 || end < len(fields) {
		buttons += "  ↕ more fields"
	}
	hint := "Tab/Shift+Tab move · ⏎ advance/choose · Esc cancel"
	if f.mode == creationEnumMode {
		hint = "←→/jk choose · ⏎ stage value · Esc back"
	}
	lines = append(lines, "", buttons, s.dim(truncate(hint, innerW)))
	return creationBox(s, lines, innerW, maxW, maxH)
}

func creationCell(f creationState, s *styles, input textinput.Model, options []string, row, width int) string {
	if len(options) > 0 {
		if f.mode == creationEnumMode && row == f.cursor {
			return enumInline(options, f.enumCursor, s)
		}
		value := input.Value()
		if value == "" {
			value = input.Placeholder
		}
		return value + " ▾"
	}
	input.SetWidth(width)
	return input.View()
}

func creationBox(s *styles, lines []string, innerW, maxW, maxH int) string {
	for i, line := range lines {
		lines[i] = truncate(line, innerW)
	}
	return clampBox(s.actionBorder.Width(innerW+s.actionBorder.GetHorizontalFrameSize()).Render(strings.Join(lines, "\n")), maxW, maxH)
}
