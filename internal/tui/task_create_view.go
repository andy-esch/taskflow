package tui

import "fmt"

func (f taskCreateForm) view(s *styles, maxW, maxH int) string {
	p := creationPresentation{noun: "task", title: f.inputs[taskCreateTitle].Value(),
		subtitle: "Nothing saved until Create · starts ready-to-start",
		recordID: f.receipt.Task.ID, path: f.receipt.Local.CommittedPath,
		inspect: f.receipt.Committed && f.receipt.Task.ID != "" && !f.receipt.DryRun}
	frame := creationHeader(f.creationState, p, s, maxW, maxH)
	if frame.rendered != "" {
		return frame.rendered
	}
	if f.mode == creationPickerMode {
		return f.pickerView(s, frame.lines, frame.innerW, maxW, maxH)
	}
	if status := f.catalogStatus(); status != "" {
		frame.lines = append(frame.lines, s.dim(truncate(status, frame.innerW)))
	}
	frame.lines = creationErrorLines(frame.lines, f.err, frame.innerW, 3, s)
	return creationFieldsView(f.creationState, s, frame, creationFieldLayout{
		rows: f.rows(), labels: taskCreateLabels[:], cancel: taskCreateCancel, submit: taskCreateSubmit,
		cell: func(row, width int) string {
			switch row {
			case taskCreateEpic:
				if !f.epic.empty() {
					return f.epic.label + " ▾"
				}
				return "Choose existing epic… ▾"
			case taskCreateMore:
				if f.more {
					return "▾ overrides (default keeps core values)"
				}
				return "▸ medium · Unknown · tier 3 · autonomy 3"
			default:
				return creationCell(f.creationState, s, f.inputs[row], taskCreateEnumOptions(row), row, width)
			}
		}})
}

func (f taskCreateForm) catalogStatus() string {
	switch {
	case !f.catalogLoaded:
		return "Loading existing epics…"
	case f.catalogErr != "":
		return "Epic catalog: " + f.catalogErr
	case len(f.epics) == 0:
		return "No readable epics. Create one on Epics (c), then reload this picker."
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
	return creationBox(s, lines, innerW, maxW, maxH)
}
