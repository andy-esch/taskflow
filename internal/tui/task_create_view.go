package tui

import (
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/theme"
)

type taskCreateModal struct{}

func (taskCreateModal) active(m *Model) bool { return m.createTask.active }

func (taskCreateModal) inputVisible(m *Model) bool { return m.width >= 46 && m.paneOuterH >= 16 }

func (o taskCreateModal) handlePaste(m *Model, msg tea.PasteMsg) tea.Cmd {
	if !o.inputVisible(m) {
		return nil
	}
	return m.createTask.updateInput(msg)
}

func (o taskCreateModal) handleKey(m *Model, msg tea.KeyPressMsg) (bool, tea.Cmd) {
	// A resize notice is not an invisible authoring surface. Preserve the draft
	// and wait for enough space to show the target, input and explicit action.
	if !o.inputVisible(m) {
		if msg.String() == "esc" && !m.createTask.dirty() && !m.createTask.saving {
			m.createTask.close()
		}
		return true, nil
	}
	return true, m.handleTaskCreateKey(msg)
}

func (taskCreateModal) view(m *Model, w, h int) string {
	return m.createTask.view(m.st, w, h)
}

func (f taskCreateForm) view(s *styles, maxW, maxH int) string {
	w := min(84, maxW)
	innerW := max(w-s.actionBorder.GetHorizontalFrameSize(), 1)
	if maxW < 44 || maxH < 14 {
		return clampBox(s.actionBorder.Render("New task · draft retained\nResize to continue safely.\nSave to: "+truncateMiddle(f.destination, innerW)), maxW, maxH)
	}
	var lines []string
	lines = append(lines, s.actionHeading.Render("New task"), "Save to: "+truncateMiddle(f.destination, innerW-9))
	if f.locked {
		lines[0] = s.actionHeading.Render("Task creation · recovery")
		lines = append(lines, strings.Split(wrap(f.err, innerW), "\n")...)
		if f.receipt.Task.ID != "" {
			lines = append(lines, truncate("Task: "+f.receipt.Task.ID, innerW))
		}
		if f.receipt.Local.CommittedPath != "" {
			lines = append(lines, truncateMiddle("Destination: "+f.receipt.Local.CommittedPath, innerW))
		}
		lines = append(lines, "Draft: "+truncate(f.inputs[taskCreateTitle].Value(), innerW-7))
		hint := "y copy draft + receipt · esc close"
		if f.saving {
			lines = append(lines, "Origin submission still pending; it is not cancellable.")
			hint = "y copy draft · wait for origin result"
		} else if f.receipt.Committed && !f.displaced && f.receipt.Task.ID != "" {
			hint = "⏎ inspect created task · " + hint
		}
		// Keep recovery actions visible even with a verbose adapter error.
		lines = lines[:min(len(lines), maxH-4)]
		lines = append(lines, s.dim(truncate(hint, innerW)))
		return taskCreateBox(s, lines, innerW, maxW, maxH)
	}
	if f.saving {
		lines = append(lines, "Saving in the origin space…", s.dim("Already submitted · not cancellable · no repeat Create"))
		return taskCreateBox(s, lines, innerW, maxW, maxH)
	}
	if f.mode == taskCreateDiscard {
		lines = append(lines, "Discard unsaved input?", s.selected.Render("[Keep editing]"), "y discards · Enter / Esc keeps editing")
		return taskCreateBox(s, lines, innerW, maxW, maxH)
	}
	lines = append(lines, s.dim("Nothing saved until Create · starts ready-to-start"))
	if f.mode == taskCreateEpicPicker {
		return f.pickerView(s, lines, innerW, maxW, maxH)
	}
	if status := f.catalogStatus(); status != "" {
		lines = append(lines, s.dim(truncate(status, innerW)))
	}
	if f.err != "" {
		errLines := strings.Split(wrap(f.err, innerW), "\n")
		for _, line := range errLines[:min(len(errLines), 3)] {
			lines = append(lines, s.fg(theme.ColorRed, line))
		}
	}
	fields := f.rows()
	fields = fields[:len(fields)-2] // Cancel/Create stay pinned, independent of scrolling
	stacked := innerW < 54
	rowHeight := 1
	if stacked {
		rowHeight = 2
	}
	available := max(1, (maxH-2-len(lines)-3)/rowHeight)
	position := min(slices.Index(f.rows(), f.cursor), len(fields)-1)
	start := min(max(position-available+1, 0), max(len(fields)-available, 0))
	end := min(start+available, len(fields))
	for _, row := range fields[start:end] {
		label := fmt.Sprintf("%-13s", taskCreateLabels[row])
		cellW := max(innerW-15, 1)
		if stacked {
			label = taskCreateLabels[row]
			cellW = max(innerW-2, 1)
		}
		var value string
		switch row {
		case taskCreateEpic:
			value = "Choose existing epic…"
			if !f.epic.empty() {
				value = f.epic.label
			}
			value += " ▾"
		case taskCreateMore:
			value = "▸ medium · Unknown · tier 3 · autonomy 3"
			if f.more {
				value = "▾ overrides (default keeps core values)"
			}
		case taskCreatePriority, taskCreateTier, taskCreateAutonomy:
			if f.mode == taskCreateEnumPicker && row == f.cursor {
				value = enumInline(taskCreateEnumOptions(row), f.enumCursor, s)
			} else {
				value = f.inputs[row].Value()
				if value == "" {
					value = f.inputs[row].Placeholder
				}
				value += " ▾"
			}
		default:
			f.inputs[row].SetWidth(cellW)
			value = f.inputs[row].View()
		}
		prefix := "  "
		if row == f.cursor {
			prefix = s.accent("› ")
			label = s.selected.Render(label)
		}
		if stacked {
			lines = append(lines, prefix+label, "  "+truncate(value, cellW))
		} else {
			lines = append(lines, prefix+label+truncate(value, cellW))
		}
	}
	buttons := "[Cancel]  [Create]"
	switch f.cursor {
	case taskCreateCancel:
		buttons = s.accent("[Cancel]") + "  [Create]"
	case taskCreateSubmit:
		buttons = "[Cancel]  " + s.accent("[Create]")
	}
	if start > 0 || end < len(fields) {
		buttons += "  ↕ more fields"
	}
	hint := "Tab/Shift+Tab move · ⏎ advance/choose · Esc cancel"
	if f.mode == taskCreateEnumPicker {
		hint = "←→/jk choose · ⏎ stage value · Esc back"
	}
	lines = append(lines, "", buttons, s.dim(truncate(hint, innerW)))
	return taskCreateBox(s, lines, innerW, maxW, maxH)
}

func taskCreateBox(s *styles, lines []string, innerW, maxW, maxH int) string {
	for i, line := range lines {
		lines[i] = truncate(line, innerW)
	}
	// Lip Gloss Width includes the frame. Passing only the content width makes
	// fields wrap and consumes the rows reserved for the explicit submit button.
	return clampBox(s.actionBorder.Width(innerW+s.actionBorder.GetHorizontalFrameSize()).Render(strings.Join(lines, "\n")), maxW, maxH)
}

func (f taskCreateForm) catalogStatus() string {
	switch {
	case !f.catalogLoaded:
		return "Loading existing epics…"
	case f.catalogErr != "":
		return "Epic catalog: " + f.catalogErr
	case len(f.epics) == 0:
		return "No readable epics. Create one with the CLI first."
	case f.unreadable > 0:
		return fmt.Sprintf("%d planning diagnostics; readable epics remain selectable", f.unreadable)
	default:
		return ""
	}
}

func (f taskCreateForm) pickerView(s *styles, lines []string, innerW, maxW, maxH int) string {
	f.query.SetWidth(max(innerW-15, 1))
	lines = append(lines, f.query.View())
	matches := f.matches()
	if status := f.catalogStatus(); status != "" {
		lines = append(lines, truncate(status, innerW))
	}
	if len(matches) == 0 {
		lines = append(lines, s.dim("No matching readable epic."))
	}
	refs := make([]entityRef, 0, len(f.epics))
	for _, epic := range f.epics {
		refs = append(refs, entityRef{key: epic.Source.ID, label: epic.Epic.ID})
	}
	hints := duplicateIdentityHints(refs)
	budget := max(1, maxH-len(lines)-3)
	cursor := min(f.pickCursor, max(len(matches)-1, 0))
	start := max(cursor-budget+1, 0)
	for i := start; i < min(start+budget, len(matches)); i++ {
		ep := matches[i]
		prefix := "  "
		if i == cursor {
			prefix = s.accent("› ")
		}
		label := labelWithIdentityHint(ep.Epic.ID, hints[ep.Source.ID])
		lines = append(lines, prefix+truncate(label+" · "+ep.Epic.Description, innerW-2))
	}
	lines = append(lines, s.dim(truncate("Type to search · ↑↓ choose · ⏎ accept · Esc back · Ctrl+R reload", innerW)))
	return taskCreateBox(s, lines, innerW, maxW, maxH)
}

// Keep modal rendering dependent on styles, not persistence or filesystem paths.
var _ modal = taskCreateModal{}
