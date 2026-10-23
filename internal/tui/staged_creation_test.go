package tui

import (
	"errors"
	"maps"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestCreationSharedStatePreservesValueOwnershipAndRecovery(t *testing.T) {
	for name, kind := range creationContractKinds {
		t.Run(name, func(t *testing.T) {
			m := openCreationKind(t, loaded(t, 120, 40), kind)
			next, _ := m.Update(tea.PasteMsg{Content: "New input"})
			after := next.(Model)
			if creationTitle(m) != "" || creationTitle(after) != "New input" {
				t.Fatal("shared controls mutated the previous model's widgets")
			}
			f := after.activeCreation()
			f.locked, f.err = true, "Original cleanup failure"
			f.displace()
			first := f.err
			f.displace()
			if !f.displaced || !f.locked || !strings.Contains(first, "Original cleanup failure") || f.err != first {
				t.Fatal("displacement lost recovery diagnostics or duplicated them")
			}
		})
	}
}

func TestCreationTextEditsDetachPriorDraft(t *testing.T) {
	for _, surface := range []string{"task", "epic", "audit", "research", "epic picker"} {
		for _, edit := range []struct {
			name string
			msg  tea.Msg
			want string
		}{
			{"paste", tea.PasteMsg{Content: "X"}, "abcXd"},
			{"type", tea.KeyPressMsg{Code: 'X', Text: "X"}, "abcXd"},
			{"backspace", tea.KeyPressMsg{Code: tea.KeyBackspace}, "abd"},
			{"delete", tea.KeyPressMsg{Code: tea.KeyDelete}, "abc"},
		} {
			t.Run(surface+"/"+edit.name, func(t *testing.T) {
				m := loaded(t, 120, 40)
				if kind, ok := creationContractKinds[surface]; ok {
					m = openCreationKind(t, m, kind)
				} else {
					m = openCreate(t, m)
					if surface == "epic picker" {
						m.creation.task.focus(taskCreateEpic)
						m, _ = createKey(m, "enter")
					}
				}
				value := func(m Model) string {
					if surface == "epic picker" {
						return m.creation.task.query.Value()
					}
					return creationTitle(m)
				}
				m = pasteCreation(m, "abcd")
				next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
				prior := next.(Model)
				next, _ = prior.Update(edit.msg)
				if value(prior) != "abcd" || value(next.(Model)) != edit.want {
					t.Fatalf("retained prior=%q, next=%q; want abcd, %s", value(prior), value(next.(Model)), edit.want)
				}
			})
		}
	}
}

func TestCreationInputCopyPreservesWidgetState(t *testing.T) {
	input := textinput.New()
	input.Prompt, input.Placeholder = "field: ", "hint"
	input.CharLimit = 8
	input.SetWidth(3)
	input.SetValue("abcd")
	input.SetCursor(2)
	input.Focus()
	diagnostic := errors.New("clipboard failed")
	input.Err = diagnostic
	next, _ := updateCreationInput(input, struct{}{}) // non-editing event must retain diagnostics
	if next.Err != diagnostic || next.Position() != 2 || !next.Focused() || next.Width() != 3 ||
		next.Prompt != "field: " || next.Placeholder != "hint" || next.CharLimit != 8 {
		t.Fatal("buffer detachment changed widget settings, focus, cursor or error")
	}
	next, _ = updateCreationInput(next, tea.PasteMsg{Content: "世界"})
	if input.Value() != "abcd" || next.Value() != "ab世界cd" || next.Position() != 4 {
		t.Fatal("detached update changed the old buffer or lost its insertion position")
	}
}

func TestCreationHiddenWidgetKeepsBlinkContinuation(t *testing.T) {
	for name, kind := range creationContractKinds {
		t.Run(name, func(t *testing.T) {
			m := openCreationKind(t, loaded(t, 120, 40), kind)
			blink := creationContractControls(&m).focus(0)
			next, _ := m.Update(tea.WindowSizeMsg{Width: 30, Height: 10})
			m = next.(Model)
			msg := blink()
			widget, ok := msg.(creationWidgetMsg)
			if !ok || !creationWidgetMaintenance(widget.msg) {
				t.Fatalf("focus did not produce a wrapped maintenance tick: %T", msg)
			}
			_, follow := m.Update(msg)
			if follow == nil {
				t.Fatal("hidden form lost the cursor blink continuation")
			}
		})
	}
}

func TestCreationDelayedPasteRequiresVisibleOwner(t *testing.T) {
	for name, kind := range creationContractKinds {
		for _, blocker := range []string{"resize", "help", "config"} {
			t.Run(name+"/"+blocker, func(t *testing.T) {
				root := seedRepo(t)
				m := openCreationKind(t, loadedAt(t, root, 120, 40), kind)
				before := testutil.SnapshotTree(t, root)
				late := m.activeCreation().widgetCmd(func() tea.Msg { return tea.PasteMsg{Content: "hidden paste"} })
				switch blocker {
				case "resize":
					next, _ := m.Update(tea.WindowSizeMsg{Width: 30, Height: 10})
					m = next.(Model)
					if !strings.Contains(ansi.Strip(m.View().Content), "Resize") {
						t.Fatal("fixture did not reach resize notice")
					}
				case "help":
					m.showHelp = true // higher-priority modal; do not open via creation hotkeys
				case "config":
					m.configOpen = true
				}
				next, _ := m.Update(late())
				m = next.(Model)
				if creationTitle(m) != "" {
					t.Fatal("delayed paste edited a hidden form")
				}
				m.showHelp, m.configOpen = false, false
				next, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
				m = next.(Model)
				fresh := m.activeCreation().widgetCmd(func() tea.Msg { return tea.PasteMsg{Content: "visible paste"} })
				next, _ = m.Update(fresh())
				m = next.(Model)
				if creationTitle(m) != "visible paste" || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatal("restored form rejected current input or staging wrote files")
				}
			})
		}
	}
}
