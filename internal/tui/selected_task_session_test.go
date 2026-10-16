package tui

import (
	"maps"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/testutil"
)

// An Atlas open may finish after Esc has returned to the old workspace and a
// task form has opened there. Matching row IDs and tab generations in two
// checkouts must not make that form valid against the newly activated Service.
func TestSelectedTaskFormsCloseWhenPendingAtlasOpenActivates(t *testing.T) {
	for _, form := range []string{"edit", "menu", "confirm", "defer"} {
		t.Run(form, func(t *testing.T) {
			// A delayed y becomes ordinary yank after the form closes. Keep that
			// command on the OSC52 path; no terminal runtime executes it here.
			t.Setenv("PATH", t.TempDir())
			m, _, alpha, beta := atlasTestModel(t)
			key := func(k string) tea.Cmd {
				tm, cmd := m.Update(press(k))
				m = tm.(Model)
				return cmd
			}

			// Visit beta's task tab first so activation will restore a same-ID
			// selected row, rather than landing on a dashboard that cannot write.
			key("j")
			cmd := key("enter")
			m = applyAtlasOpen(t, m, cmd)
			m = toTasks(t, m)
			ref := testTaskRef("beta") // ready-to-start in both checkouts
			if !m.cur().selectByKey(ref.key) || m.selectedRef() != ref {
				t.Fatal("beta checkout is missing the selected task")
			}
			incomingGen := m.cur().loadGen + 1 // activation refreshes the cached tab
			key("a")
			key("k") // beta card -> alpha card
			cmd = key("enter")
			m = applyAtlasOpen(t, m, cmd)
			m = toTasks(t, m)
			// Align through ordinary refreshes, not assigned counters. Generations
			// are per tab, not a workspace identity; coincidence is legitimate.
			for i := 0; i < 3 && m.cur().loadGen < incomingGen; i++ {
				cmd = key("r")
				m = settleAtlasCmd(t, m, cmd)
			}
			if m.workspace.Checkout != alpha || m.cur().loadGen != incomingGen || !m.cur().selectByKey(ref.key) {
				t.Fatal("alpha checkout is not ready for the matching-generation form")
			}

			key("a")
			key("j") // alpha card -> beta card
			pendingOpen := key("enter")
			if pendingOpen == nil || !m.atlas.opening {
				t.Fatal("beta workspace open was not queued")
			}
			key("esc") // does not cancel the already-issued workspace open
			if m.onAtlas || m.workspace.Checkout != alpha || m.selectedRef() != ref {
				t.Fatal("Esc did not return to alpha's selected task")
			}

			switch form {
			case "edit":
				key("e")
				m = editCursorTo(t, m, "priority")
				key("enter")
				m = enumCursorTo(t, m, "high")
				if !m.edit.active || !m.edit.editing || m.edit.ref != ref || m.edit.sourceGen != incomingGen {
					t.Fatal("task edit form was not opened in alpha")
				}
			case "menu", "confirm", "defer":
				key("m")
				verb := "start"
				switch form {
				case "confirm":
					verb = "deprecate"
				case "defer":
					verb = "defer"
				}
				m = cursorTo(t, m, verb)
				if form != "menu" {
					key("enter")
				}
				if form == "defer" {
					m.action.dateInput.SetValue("2099-01-02")
				}
				if !m.action.active || m.action.ref != ref || m.action.sourceGen != incomingGen ||
					m.action.confirm != (form == "confirm") || m.action.revisit != (form == "defer") {
					t.Fatal("task lifecycle form was not opened in alpha")
				}
			}
			beforeAlpha := testutil.SnapshotTree(t, alpha)
			beforeBeta := testutil.SnapshotTree(t, beta)
			oldService, oldSession := m.svc, m.sessionGen

			// Deliver the real command's current-generation result through Update,
			// including activation and the restored tab's ordinary data refresh.
			m = settleAtlasCmd(t, m, pendingOpen)
			if m.workspace.Checkout != beta || m.svc == oldService || m.sessionGen != oldSession+1 ||
				m.onDash || m.selectedRef() != ref || m.cur().loadGen != incomingGen || m.cur().coherentGen != incomingGen {
				t.Fatal("pending open did not restore beta's same-ID task and generation")
			}
			if !maps.Equal(beforeAlpha, testutil.SnapshotTree(t, alpha)) || !maps.Equal(beforeBeta, testutil.SnapshotTree(t, beta)) {
				t.Fatal("activation changed planning records")
			}
			if m.edit.active || m.action.active {
				t.Fatal("task form survived activation and could submit through the new Service")
			}

			// Delayed confirmation keys cannot submit the departed form. Ordinary
			// Enter may open a detail view, so settle its harmless read commands.
			for _, k := range []string{"y", "enter"} {
				cmd = key(k)
				m = settleAtlasCmd(t, m, cmd)
			}
			if !maps.Equal(beforeAlpha, testutil.SnapshotTree(t, alpha)) || !maps.Equal(beforeBeta, testutil.SnapshotTree(t, beta)) {
				t.Fatal("departed task form wrote to a checkout after activation")
			}
		})
	}
}
