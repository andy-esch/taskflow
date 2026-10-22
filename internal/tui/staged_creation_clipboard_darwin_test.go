package tui

import (
	"bytes"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/andy-esch/taskflow/internal/testutil"
)

const creationClipboardText = "CLIP-qhjkn-世界"

// This platform-specific integration test executes Bubbles' actual Ctrl+V
// command and private result, never the user's clipboard. On Darwin the pinned
// provider invokes pbpaste by name; an isolated PATH contains only this stub.
// Generic ownership, stale-input and buffer regressions run on every platform.
func stubCreationClipboard(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pbpaste"), []byte("#!/bin/sh\nprintf '%s' '"+creationClipboardText+"'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
}

func TestCreationNativeClipboardOwnership(t *testing.T) {
	stubCreationClipboard(t)
	for _, kind := range []string{"task", "epic"} {
		for _, blocker := range []string{"visible", "resize", "help", "config", "field changed"} {
			t.Run(kind+"/"+blocker, func(t *testing.T) {
				root := seedRepo(t)
				m := loadedAt(t, root, 120, 40)
				if kind == "task" {
					m = openCreate(t, m)
				} else {
					m = openEpicForm(t, m)
				}
				before := testutil.SnapshotTree(t, root)
				next, late := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
				m = next.(Model)
				if late == nil {
					t.Fatal("Ctrl+V did not issue the native command")
				}
				switch blocker {
				case "resize":
					next, _ = m.Update(tea.WindowSizeMsg{Width: 30, Height: 10})
					m = next.(Model)
					if !strings.Contains(ansi.Strip(m.View().Content), "Resize") {
						t.Fatal("fixture did not reach resize notice")
					}
				case "help":
					m.showHelp = true // defensive precedence case, not a creation hotkey
				case "config":
					m.configOpen = true
				case "field changed":
					m, _ = createKey(m, "tab")
				}
				result := late() // real native provider, not a fabricated tea.PasteMsg
				widget, ok := result.(creationWidgetMsg)
				if !ok {
					t.Fatalf("native command lost its scope: %T", result)
				}
				if _, terminal := widget.msg.(tea.PasteMsg); terminal {
					t.Fatal("fixture accidentally exercises terminal instead of native paste")
				}
				next, _ = m.Update(result)
				m = next.(Model)
				values := []string{m.createTask.inputs[taskCreateTitle].Value(), m.createTask.inputs[taskCreateTags].Value()}
				if kind == "epic" {
					values = []string{m.createEpic.inputs[epicCreateTitle].Value(), m.createEpic.inputs[epicCreateDescription].Value()}
				}
				want := ""
				if blocker == "visible" {
					want = creationClipboardText
				}
				if values[0] != want || values[1] != "" || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatalf("native ownership: values=%q; expected title=%q, no other edit/write", values, want)
				}
			})
		}
	}
}

// Hold result delivery, not command construction/execution: a running program
// must reject a native clipboard result while it renders the resize notice.
type creationClipboardRuntime struct {
	model                     Model
	ready, release, delivered chan struct{}
}

type creationClipboardDelivery struct{ msg tea.Msg }

func (r creationClipboardRuntime) Init() tea.Cmd  { return r.model.Init() }
func (r creationClipboardRuntime) View() tea.View { return r.model.View() }
func (r creationClipboardRuntime) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	delivery, delivered := msg.(creationClipboardDelivery)
	if delivered {
		msg = delivery.msg
	}
	next, cmd := r.model.Update(msg)
	r.model = next.(Model)
	if delivered {
		close(r.delivered)
	}
	if key, ok := msg.(tea.KeyPressMsg); ok && key.String() == "ctrl+v" && cmd != nil {
		original := cmd
		cmd = func() tea.Msg {
			result := original()
			close(r.ready)
			<-r.release
			return creationClipboardDelivery{msg: result}
		}
	}
	return r, cmd
}

func TestCreationNativeClipboardAfterRuntimeResize(t *testing.T) {
	stubCreationClipboard(t)
	r := creationClipboardRuntime{model: newModel(t), ready: make(chan struct{}), release: make(chan struct{}), delivered: make(chan struct{})}
	tm := teatest.NewTestModel(t, r, teatest.WithInitialTermSize(120, 40))
	// Release the held command even when an assertion fails, without leaking a
	// goroutine into the next test or reaching a real clipboard utility.
	released := false
	t.Cleanup(func() {
		if !released {
			close(r.release)
		}
	})
	wait := func(text string) {
		t.Helper()
		teatest.WaitFor(t, tm.Output(), func(b []byte) bool { return bytes.Contains(b, []byte(text)) }, teatest.WithDuration(5*time.Second))
	}
	tm.Send(press("]"))
	wait("beta")
	tm.Send(press("]"))
	wait("01-test")
	tm.Send(press("c"))
	wait("New epic")
	tm.Send(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
	select {
	case <-r.ready:
	case <-time.After(5 * time.Second):
		t.Fatal("native command did not execute")
	}
	tm.Send(tea.WindowSizeMsg{Width: 30, Height: 10})
	wait("Resize")
	close(r.release)
	released = true
	select {
	case <-r.delivered:
	case <-time.After(5 * time.Second):
		t.Fatal("native result was not delivered")
	}
	tm.Send(tea.WindowSizeMsg{Width: 46, Height: 18})
	wait("New epic")
	tm.Send(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(creationClipboardRuntime).model
	if final.width != 46 || final.height != 18 || final.createEpic.inputs[epicCreateTitle].Value() != "" ||
		!strings.Contains(ansi.Strip(final.View().Content), "New epic") {
		t.Fatal("running program accepted native paste behind its resize notice")
	}
}
