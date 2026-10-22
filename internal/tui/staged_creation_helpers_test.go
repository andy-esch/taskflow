package tui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func createKey(m Model, key string) (Model, tea.Cmd) {
	msg := press(key)
	switch key {
	case "up":
		msg = tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		msg = tea.KeyPressMsg{Code: tea.KeyDown}
	case "shift+tab":
		msg = tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	case "ctrl+p":
		msg = tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}
	case "ctrl+r":
		msg = tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl}
	}
	next, cmd := m.Update(msg)
	return next.(Model), cmd
}

// Execute finite application/widget work through Model.Update. Only explicit
// cursor maintenance is left to the running-program tests; opaque clipboard
// results are delivered, not silently discarded with the rest of the widgets.
func settleCreate(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	next, err := settleCreationCommands(m, cmd, 128)
	if err != nil {
		t.Fatal(err)
	}
	return next
}

func settleCreationCommands(m Model, cmd tea.Cmd, budget int) (Model, error) {
	remaining := budget
	var run func(Model, tea.Cmd) (Model, error)
	run = func(m Model, cmd tea.Cmd) (Model, error) {
		if cmd == nil {
			return m, nil
		}
		if remaining <= 0 {
			return m, fmt.Errorf("creation command settling exceeded budget %d; possible maintenance loop", budget)
		}
		remaining--
		msg := cmd()
		if msg == nil {
			return m, nil
		}
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, child := range batch {
				var err error
				m, err = run(m, child)
				if err != nil {
					return m, err
				}
			}
			return m, nil
		}
		inner := msg
		if scoped, ok := inner.(sessionMsg); ok {
			inner = scoped.msg
		}
		if widget, ok := inner.(creationWidgetMsg); ok && creationWidgetMaintenance(widget.msg) {
			return m, nil
		}
		next, follow := m.Update(msg)
		return run(next.(Model), follow)
	}
	return run(m, cmd)
}
