package tui

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
)

// seedRepo writes a tiny planning tree: alpha (in-progress), beta (ready-to-start).
func seedRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(status, slug, desc string) {
		dir := filepath.Join(root, "tasks", status)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		body := fmt.Sprintf("---\nstatus: %s\ndescription: %s\n---\n# %s\n", status, desc, slug)
		if err := os.WriteFile(filepath.Join(dir, slug+".md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("in-progress", "alpha", "the alpha task")
	write("ready-to-start", "beta", "the beta task")
	return root
}

func newModel(t *testing.T) Model {
	t.Helper()
	root := seedRepo(t)
	return New(core.NewService(store.NewFS(root)), root)
}

// loaded returns a model that's been sized and had its task load applied.
func loaded(t *testing.T, w, h int) Model {
	t.Helper()
	m := newModel(t)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()()) // run loadTasks synchronously, feed result
	return tm.(Model)
}

func press(s string) tea.KeyMsg {
	switch s {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestModel_LoadsWorkingSetOrder(t *testing.T) {
	m := loaded(t, 120, 40)
	if m.loading {
		t.Fatal("still loading")
	}
	// Working-set order: in-progress (alpha) before ready-to-start (beta).
	if m.selectedSlug() != "alpha" {
		t.Errorf("expected in-progress task first, got %q", m.selectedSlug())
	}
	if !strings.Contains(m.View(), "alpha") || !strings.Contains(m.View(), "beta") {
		t.Errorf("view should list both tasks:\n%s", m.View())
	}
}

func TestModel_SelectionLoadsBodyWithStaleGuard(t *testing.T) {
	m := loaded(t, 120, 40)
	// Move down → select beta, which triggers a body load.
	tm, _ := m.Update(press("j"))
	m = tm.(Model)
	if m.selectedSlug() != "beta" {
		t.Fatalf("expected beta after j, got %q", m.selectedSlug())
	}
	if !m.detail.loading {
		t.Error("selection change should start a body load")
	}
	// A stale body (for alpha) must be dropped while beta is selected.
	tm, _ = m.Update(taskBodyMsg{slug: "alpha", task: domain.Task{Slug: "alpha"}, body: "x"})
	m = tm.(Model)
	if m.detail.hasContent {
		t.Error("stale body for a different selection must be ignored")
	}
	// The matching body sets the detail.
	tm, _ = m.Update(taskBodyMsg{slug: "beta", task: domain.Task{Slug: "beta"}, body: "beta body"})
	m = tm.(Model)
	if !m.detail.hasContent || m.detail.title != "beta" {
		t.Errorf("detail should show beta, got title %q hasContent=%v", m.detail.title, m.detail.hasContent)
	}
}

func TestModel_FocusRouting(t *testing.T) {
	m := loaded(t, 120, 40)
	if m.focus != focusList {
		t.Fatal("should start focused on the list")
	}
	for _, step := range []struct {
		k    string
		want focus
	}{
		{"l", focusDetail},   // drill into detail
		{"h", focusList},     // back
		{"tab", focusDetail}, // toggle
		{"esc", focusList},   // esc returns from detail
	} {
		tm, _ := m.Update(press(step.k))
		m = tm.(Model)
		if m.focus != step.want {
			t.Errorf("after %q: focus = %d, want %d", step.k, m.focus, step.want)
		}
	}
}

func TestModel_Responsive(t *testing.T) {
	m := loaded(t, 120, 40)
	if !m.twoPane {
		t.Error("120 cols should be two-pane")
	}
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 70, Height: 40})
	m = tm.(Model)
	if m.twoPane {
		t.Error("70 cols should be single-pane")
	}
	// Tiny terminal must not panic.
	tm, _ = m.Update(tea.WindowSizeMsg{Width: 20, Height: 6})
	m = tm.(Model)
	_ = m.View()
}

func TestModel_BodyErrorDoesNotBrick(t *testing.T) {
	m := loaded(t, 120, 40)
	slug := m.selectedSlug()
	// An ambiguous-slug (duplicate across dirs) body error must not blank the UI.
	tm, _ := m.Update(bodyErrMsg{slug: slug, err: domain.ErrAmbiguous})
	m = tm.(Model)
	if m.err != nil {
		t.Error("a per-task body error must not set the fatal error")
	}
	if !m.detail.hasContent {
		t.Error("the error should be shown in the detail pane")
	}
	if !strings.Contains(m.View(), "alpha") {
		t.Errorf("the list must still render (not an error screen):\n%s", m.View())
	}
}

func TestModel_DetailScrollKeys(t *testing.T) {
	m := loaded(t, 120, 12) // short, so the body overflows
	slug := m.selectedSlug()
	long := strings.Repeat("a line of text\n", 60)
	tm, _ := m.Update(taskBodyMsg{slug: slug, task: domain.Task{Slug: slug}, body: long})
	m = tm.(Model)
	tm, _ = m.Update(press("l")) // focus detail
	m = tm.(Model)

	tm, _ = m.Update(press("G")) // bottom
	m = tm.(Model)
	if m.detail.vp.YOffset == 0 {
		t.Error("G should scroll the detail pane down")
	}
	tm, _ = m.Update(press("g")) // top
	m = tm.(Model)
	if m.detail.vp.YOffset != 0 {
		t.Errorf("g should scroll to top, YOffset=%d", m.detail.vp.YOffset)
	}
}

func TestModel_SlashFiltersAndCapturesKeys(t *testing.T) {
	m := loaded(t, 120, 40)
	if m.list.SettingFilter() {
		t.Fatal("should not be filtering initially")
	}
	tm, _ := m.Update(press("/"))
	m = tm.(Model)
	if !m.list.SettingFilter() {
		t.Fatal("/ should open the list filter")
	}
	// Typing into the filter must not trigger global hotkeys (q/r).
	for _, k := range []string{"b", "q", "r"} {
		tm, _ = m.Update(press(k))
		m = tm.(Model)
	}
	if !m.list.SettingFilter() {
		t.Error("global hotkeys must not leak out of the filter input")
	}
}

func TestModel_FilterNarrows(t *testing.T) {
	// Run the real program (teatest runs cmds in the runtime, so the list's async
	// FilterMatchesMsg is actually applied) and inspect the final model.
	tm := teatest.NewTestModel(t, newModel(t), teatest.WithInitialTermSize(120, 40))
	// Wait for the async task load to render before filtering.
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("beta"))
	}, teatest.WithDuration(3*time.Second))
	tm.Send(press("/"))
	for _, r := range "beta" {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // apply the filter
	tm.Send(press("q"))                     // quit (no longer filtering)

	fm := tm.FinalModel(t, teatest.WithFinalTimeout(3*time.Second)).(Model)
	vis := fm.list.VisibleItems()
	if len(vis) != 1 {
		t.Fatalf("applied filter %q should leave 1 item, got %d", "beta", len(vis))
	}
	if it, _ := vis[0].(taskItem); it.t.Slug != "beta" {
		t.Errorf("expected beta visible, got %q", it.t.Slug)
	}
}

// visibleW is the display-cell width (ANSI- and wide-rune-aware), so the
// invariant can't be fooled by a wide rune that passes a naive rune count.
func visibleW(s string) int { return ansi.StringWidth(s) }

// TestModel_ViewFitsTerminal is the layout invariant guard: the rendered View
// must be exactly the terminal height and never wider than the terminal — so the
// top border is never clipped and rows never wrap. Locks this for future sprints.
func TestModel_ViewFitsTerminal(t *testing.T) {
	for _, d := range []struct{ w, h int }{
		{120, 40}, {100, 24}, {90, 30}, {80, 24}, {70, 20}, {40, 12}, {24, 8},
	} {
		m := loaded(t, d.w, d.h)
		tm, _ := m.Update(taskBodyMsg{slug: m.selectedSlug(), task: domain.Task{Slug: m.selectedSlug()}, body: "# body\n\nsome text here\n"})
		m = tm.(Model)
		lines := strings.Split(m.View(), "\n")
		if len(lines) != d.h {
			t.Errorf("%dx%d: View has %d lines, want exactly %d", d.w, d.h, len(lines), d.h)
		}
		for i, ln := range lines {
			if w := visibleW(ln); w > d.w {
				t.Errorf("%dx%d: line %d is %d wide > terminal %d: %q", d.w, d.h, i, w, d.w, ansi.Strip(ln))
			}
		}
	}
}

// TestModel_NoFrameBeforeSize guards the clipped-top-border regression: if data
// loads before the first WindowSizeMsg, View must not draw a (broken) full frame.
func TestModel_NoFrameBeforeSize(t *testing.T) {
	m := newModel(t)
	tm, _ := m.Update(m.Init()()) // tasksLoadedMsg, but no WindowSizeMsg yet
	m = tm.(Model)
	v := m.View()
	if strings.ContainsAny(v, "╭╮│") || strings.Count(v, "\n") > 1 {
		t.Errorf("must not render a frame before the first WindowSizeMsg:\n%q", v)
	}
}

func TestModel_QuitsCleanly(t *testing.T) {
	tm := teatest.NewTestModel(t, newModel(t), teatest.WithInitialTermSize(120, 40))
	tm.Send(press("q"))
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}
