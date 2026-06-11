package tui

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

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

	// One epic and one open audit so the epics/audits tabs have content.
	writeFile := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeFile("epics/01-test.md", "---\nstatus: planning\ndescription: a test epic\npriority: high\n---\n# Test epic\n")
	writeFile("audits/open/2026-06-01-thing.md", "---\narea: store\ndate: 2026-06-01\n---\n# Audit\n")
	return root
}

// drain runs a single (non-batch) command and applies its message — enough to
// resolve the async list/detail loads switchTab and refresh return in tests.
func drain(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		return m
	}
	msg := cmd()
	if msg == nil {
		return m
	}
	tm, _ := m.Update(msg)
	return tm.(Model)
}

// drainBatch runs a tea.Batch command, applying each sub-command's message — used
// for reloadAll, which returns one reload Cmd per loaded tab.
func drainBatch(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		return m
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected a tea.BatchMsg, got %T", cmd())
	}
	for _, c := range batch {
		m = drain(t, m, c)
	}
	return m
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
	if !m.cur().loaded {
		t.Fatal("tasks tab should be loaded")
	}
	// Working-set order: in-progress (alpha) before ready-to-start (beta).
	if m.selectedID() != "alpha" {
		t.Errorf("expected in-progress task first, got %q", m.selectedID())
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
	if m.selectedID() != "beta" {
		t.Fatalf("expected beta after j, got %q", m.selectedID())
	}
	if !m.detail.loading {
		t.Error("selection change should start a body load")
	}
	// A stale body (for alpha) must be dropped while beta is selected.
	tm, _ = m.Update(detailMsg{kind: entityTasks, id: "alpha", content: taskDetail{t: domain.Task{Slug: "alpha"}, body: "x"}})
	m = tm.(Model)
	if m.detail.hasContent {
		t.Error("stale body for a different selection must be ignored")
	}
	// The matching body sets the detail.
	tm, _ = m.Update(detailMsg{kind: entityTasks, id: "beta", content: taskDetail{t: domain.Task{Slug: "beta"}, body: "beta body"}})
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
	slug := m.selectedID()
	// An ambiguous-slug (duplicate across dirs) body error must not blank the UI.
	tm, _ := m.Update(detailErrMsg{kind: entityTasks, id: slug, err: domain.ErrAmbiguous})
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
	slug := m.selectedID()
	long := strings.Repeat("a line of text\n", 60)
	tm, _ := m.Update(detailMsg{kind: entityTasks, id: slug, content: taskDetail{t: domain.Task{Slug: slug}, body: long}})
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
	if m.cur().list.SettingFilter() {
		t.Fatal("should not be filtering initially")
	}
	tm, _ := m.Update(press("/"))
	m = tm.(Model)
	if !m.cur().list.SettingFilter() {
		t.Fatal("/ should open the list filter")
	}
	// Typing into the filter must not trigger global hotkeys (q/r).
	for _, k := range []string{"b", "q", "r"} {
		tm, _ = m.Update(press(k))
		m = tm.(Model)
	}
	if !m.cur().list.SettingFilter() {
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
	vis := fm.cur().list.VisibleItems()
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
		tm, _ := m.Update(detailMsg{kind: entityTasks, id: m.selectedID(), content: taskDetail{t: domain.Task{Slug: m.selectedID()}, body: "# body\n\nsome text here\n"}})
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

// --- S2a: multi-entity navigation ---

func TestModel_CycleTabsLoadsEntity(t *testing.T) {
	m := loaded(t, 120, 40)
	if m.cur().name != "tasks" {
		t.Fatalf("should start on tasks, got %q", m.cur().name)
	}
	// ] → epics; the returned cmd loads the epic list.
	tm, cmd := m.Update(press("]"))
	m = tm.(Model)
	if m.cur().name != "epics" {
		t.Fatalf("] should switch to epics, got %q", m.cur().name)
	}
	m = drain(t, m, cmd)
	if !m.cur().loaded {
		t.Fatal("epics tab should load on switch")
	}
	if _, ok := m.cur().list.SelectedItem().(epicItem); !ok {
		t.Errorf("epics tab should hold epicItems, got %T", m.cur().list.SelectedItem())
	}
	if m.selectedID() != "01-test" {
		t.Errorf("expected epic 01-test selected, got %q", m.selectedID())
	}
	// [ wraps back to tasks (cursor preserved — see the dedicated test).
	tm, _ = m.Update(press("["))
	m = tm.(Model)
	if m.cur().name != "tasks" {
		t.Errorf("[ should switch back to tasks, got %q", m.cur().name)
	}
}

func TestModel_CommandJumpSwitchesEntity(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press(":"))
	m = tm.(Model)
	if !m.cmd.active {
		t.Fatal(": should open the command bar")
	}
	for _, r := range "audits" {
		tm, _ = m.Update(press(string(r)))
		m = tm.(Model)
	}
	if m.cmd.value() != "audits" {
		t.Fatalf("command bar should hold %q, got %q", "audits", m.cmd.value())
	}
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	if m.cmd.active {
		t.Error("enter should close the command bar")
	}
	if m.cur().name != "audits" {
		t.Fatalf(":audits should switch to audits, got %q", m.cur().name)
	}
	m = drain(t, m, cmd)
	if !m.cur().loaded {
		t.Error("audits tab should load after the jump")
	}
	if m.selectedID() != "2026-06-01-thing" {
		t.Errorf("expected the seeded audit selected, got %q", m.selectedID())
	}
}

func TestModel_CommandBarCapturesKeysAndCompletes(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press(":"))
	m = tm.(Model)
	// A global hotkey (q) must not quit while the command bar is open — it's
	// captured as input instead (proven by it landing in the value).
	tm, _ = m.Update(press("q"))
	m = tm.(Model)
	if !m.cmd.active || m.cmd.value() != "q" {
		t.Errorf("q should be captured as input, value=%q active=%v", m.cmd.value(), m.cmd.active)
	}
	// Tab completes the unique prefix: "q" has no match, so reset and try "e".
	m.cmd.ti.SetValue("e")
	m.cmd.complete(m.entityNames())
	if m.cmd.value() != "epics" {
		t.Errorf("Tab should complete %q → %q, got %q", "e", "epics", m.cmd.value())
	}
}

func TestModel_UnknownCommandReopensWithError(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press(":"))
	m = tm.(Model)
	for _, r := range "bogus" {
		tm, _ = m.Update(press(string(r)))
		m = tm.(Model)
	}
	tm, _ = m.Update(press("enter"))
	m = tm.(Model)
	if !m.cmd.active {
		t.Error("an unknown command should reopen the bar")
	}
	if m.cmd.err == "" {
		t.Error("an unknown command should show an inline error")
	}
	if m.cur().name != "tasks" {
		t.Errorf("an unknown command must not switch tabs, got %q", m.cur().name)
	}
}

func TestModel_PerTabCursorPreserved(t *testing.T) {
	m := loaded(t, 120, 40)
	// Move the tasks cursor to beta.
	tm, _ := m.Update(press("j"))
	m = tm.(Model)
	if m.selectedID() != "beta" {
		t.Fatalf("expected beta selected on tasks, got %q", m.selectedID())
	}
	// Cycle tasks → epics → audits → tasks, draining each load.
	for i := 0; i < 3; i++ {
		tm, cmd := m.Update(press("]"))
		m = drain(t, tm.(Model), cmd)
	}
	if m.cur().name != "tasks" {
		t.Fatalf("expected to land back on tasks, got %q", m.cur().name)
	}
	if m.selectedID() != "beta" {
		t.Errorf("tasks cursor should be preserved at beta, got %q", m.selectedID())
	}
}

// seedManyTasks writes n ready-to-start tasks — enough to force bubbles/list to
// paginate at a small height (its `••` dots render one line beyond SetHeight).
func seedManyTasks(t *testing.T, n int) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "tasks", "ready-to-start")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		slug := fmt.Sprintf("task-%02d", i)
		body := fmt.Sprintf("---\nstatus: ready-to-start\ndescription: task %d\n---\n# %s\n", i, slug)
		if err := os.WriteFile(filepath.Join(dir, slug+".md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestModel_ChromeVisibleWhenListPaginates pins the footer/tab-strip-cropping
// regression: a paginated list must not push the chrome off-screen. (The list's
// pagination row renders beyond its SetHeight; the layout reserves a line for it
// and hard-clamps the body so the chrome always survives.)
func TestModel_ChromeVisibleWhenListPaginates(t *testing.T) {
	root := seedManyTasks(t, 20)
	m := New(core.NewService(store.NewFS(root)), root)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 14})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = tm.(Model)
	tm, _ = m.Update(press(":")) // open the command bar (bottom chrome)
	m = tm.(Model)

	lines := strings.Split(m.View(), "\n")
	if len(lines) != 14 {
		t.Fatalf("view must be exactly 14 lines, got %d", len(lines))
	}
	if top := ansi.Strip(lines[0]); !strings.Contains(top, "tasks") {
		t.Errorf("tab strip (top chrome) must stay visible, got %q", top)
	}
	if last := ansi.Strip(lines[len(lines)-1]); !strings.HasPrefix(last, ":") {
		t.Errorf("command bar (bottom chrome) must stay visible, got %q", last)
	}
}

func TestModel_CommandAliases(t *testing.T) {
	for _, tc := range []struct{ word, want string }{
		{"t", "tasks"}, {"e", "epics"}, {"a", "audits"},
		{"task", "tasks"}, {"epic", "epics"}, {"audits", "audits"},
	} {
		m := loaded(t, 120, 40)
		tm, _ := m.Update(press(":"))
		m = tm.(Model)
		for _, r := range tc.word {
			tm, _ = m.Update(press(string(r)))
			m = tm.(Model)
		}
		tm, _ = m.Update(press("enter"))
		m = tm.(Model)
		if m.cur().name != tc.want {
			t.Errorf(":%s → %q, want %q", tc.word, m.cur().name, tc.want)
		}
	}
}

func TestModel_EmptyTabShowsNothingSelected(t *testing.T) {
	// A repo with tasks but no audits dir → the audits tab loads empty.
	root := t.TempDir()
	dir := filepath.Join(root, "tasks", "ready-to-start")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "only.md"),
		[]byte("---\nstatus: ready-to-start\ndescription: x\n---\n# only\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := New(core.NewService(store.NewFS(root)), root)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = tm.(Model)
	// :audits → an empty list.
	tm, cmd := m.Update(press("]")) // → epics (also empty)
	m = drain(t, tm.(Model), cmd)
	tm, cmd = m.Update(press("]")) // → audits (empty)
	m = drain(t, tm.(Model), cmd)
	if m.cur().name != "audits" || len(m.cur().list.Items()) != 0 {
		t.Fatalf("expected an empty audits tab, got %q with %d items", m.cur().name, len(m.cur().list.Items()))
	}
	if m.detail.loading {
		t.Error("an empty tab must not sit on a perpetual loading state")
	}
	if v := ansi.Strip(m.detail.View()); !strings.Contains(v, "nothing selected") {
		t.Errorf("empty-tab detail should show the empty state, got %q", v)
	}
}

func TestEntityDetailRenderers(t *testing.T) {
	epic := epicDetail{
		e:     domain.Epic{ID: "17-x", Status: "in-progress", Priority: "high"},
		tasks: []domain.Task{{Slug: "a", Status: domain.StatusCompleted}, {Slug: "b", Status: domain.StatusReadyToStart}},
		body:  "# Epic body",
	}
	out := ansi.Strip(epic.Render(70))
	for _, want := range []string{"17-x", "1/2", "50%", "Epic body"} {
		if !strings.Contains(out, want) {
			t.Errorf("epic detail missing %q:\n%s", want, out)
		}
	}
	if epic.Title() != "17-x" {
		t.Errorf("epic title = %q", epic.Title())
	}

	audit := auditDetail{
		a:    domain.Audit{Slug: "2026-06-01-x", Bucket: domain.AuditOpen, Area: "store", Findings: 5, OpenFindings: 2},
		body: "# Audit body",
	}
	out = ansi.Strip(audit.Render(70))
	for _, want := range []string{"2026-06-01-x", "store", "2 open / 5 total", "Audit body"} {
		if !strings.Contains(out, want) {
			t.Errorf("audit detail missing %q:\n%s", want, out)
		}
	}
}

// TestModel_LongTitleKeepsDetailBorder pins the chrome-corruption fix: a detail
// title longer than the pane is truncated, not wrapped — so the pane keeps its
// bottom border (two `╯`, one per pane) at the narrowest two-pane width.
func TestModel_LongTitleKeepsDetailBorder(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "tasks", "in-progress")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	slug := "an-extremely-long-task-slug-well-past-the-detail-pane-inner-width"
	if err := os.WriteFile(filepath.Join(dir, slug+".md"),
		[]byte("---\nstatus: in-progress\ndescription: x\n---\n# body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := New(core.NewService(store.NewFS(root)), root)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 24}) // narrowest two-pane
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = tm.(Model)
	tm, _ = m.Update(detailMsg{kind: entityTasks, id: slug, content: taskDetail{t: domain.Task{Slug: slug}, body: "body"}})
	m = tm.(Model)

	lines := strings.Split(m.View(), "\n")
	if len(lines) != 24 {
		t.Fatalf("view must be 24 lines, got %d", len(lines))
	}
	// The pane bottom-border row is just above the footer; both panes must close.
	border := ansi.Strip(lines[len(lines)-2])
	if got := strings.Count(border, "╯"); got != 2 {
		t.Errorf("both panes must keep their bottom border (2 ╯), got %d: %q", got, border)
	}
}

// TestModel_RecoversFromFatalError pins that a fatal load error clears on the
// next successful load (otherwise a transient failure bricked the session).
func TestModel_RecoversFromFatalError(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(errMsg{err: domain.ErrNotFound})
	m = tm.(Model)
	if m.err == nil || !strings.Contains(m.View(), "error:") {
		t.Fatal("errMsg should show the error screen")
	}
	// A subsequent successful list load must clear it.
	tm, _ = m.Update(m.Init()())
	m = tm.(Model)
	if m.err != nil {
		t.Errorf("a successful load must clear the fatal error, got %v", m.err)
	}
}

// TestModel_RefreshFiresReloadMsg pins the reload path: `r` emits a reloadMsg
// (the seam fsnotify reuses in S3), which captures each loaded tab's cursor id.
func TestModel_RefreshFiresReloadMsg(t *testing.T) {
	m := loaded(t, 120, 40)
	_, cmd := m.Update(press("r"))
	if cmd == nil {
		t.Fatal("r must return a command")
	}
	if _, ok := cmd().(reloadMsg); !ok {
		t.Fatalf("r must fire reloadMsg, got %T", cmd())
	}
	tm, _ := m.Update(reloadMsg{})
	m = tm.(Model)
	if m.cur().restore != "alpha" {
		t.Errorf("reloadMsg should capture the active tab's cursor id, got %q", m.cur().restore)
	}
}

func TestModel_TabStripCollapsesWhenNarrow(t *testing.T) {
	wide := loaded(t, 120, 40)
	strip := wide.tabStrip()
	for _, name := range []string{"tasks", "epics", "audits"} {
		if !strings.Contains(strip, name) {
			t.Errorf("wide tab strip should list %q: %q", name, ansi.Strip(strip))
		}
	}
	narrow := loaded(t, 50, 20)
	if !strings.Contains(narrow.tabStrip(), "▾") {
		t.Errorf("narrow tab strip should collapse to a chip, got %q", ansi.Strip(narrow.tabStrip()))
	}
}

// --- S2b: search, status views, sort, help ---

func TestModel_SortReordersAndShowsChip(t *testing.T) {
	m := loaded(t, 120, 40)
	// Default (working-set): alpha (in-progress) before beta (ready-to-start).
	if got := m.cur().list.Items()[0].(taskItem).t.Slug; got != "alpha" {
		t.Fatalf("default order should lead with alpha, got %q", got)
	}
	// o×4 cycles default→priority→updated→tier→slug (sort applies synchronously).
	for i := 0; i < 4; i++ {
		tm, _ := m.Update(press("o"))
		m = tm.(Model)
	}
	if m.cur().sortKey != sortSlug {
		t.Fatalf("o×4 should land on sortSlug, got %v", m.cur().sortKey)
	}
	if !strings.Contains(m.cur().chip(), "sort:slug↓") {
		t.Errorf("chip should announce the sort, got %q", m.cur().chip())
	}
	// Reverse → beta sorts before alpha, and the arrow flips.
	tm, _ := m.Update(press("O"))
	m = tm.(Model)
	if got := m.cur().list.Items()[0].(taskItem).t.Slug; got != "beta" {
		t.Errorf("reversed slug sort should lead with beta, got %q", got)
	}
	if !strings.Contains(m.cur().chip(), "sort:slug↑") {
		t.Errorf("reversed chip should show ↑, got %q", m.cur().chip())
	}
}

func TestModel_StatusViewCyclesAndFilters(t *testing.T) {
	m := loaded(t, 120, 40)
	// s → the first non-default view is in-progress; the reload filters to alpha.
	tm, cmd := m.Update(press("s"))
	m = drain(t, tm.(Model), cmd)
	if m.cur().statusView != "in-progress" {
		t.Fatalf("s should select the in-progress view, got %q", m.cur().statusView)
	}
	if n := len(m.cur().list.Items()); n != 1 {
		t.Fatalf(":in-progress should show only alpha, got %d items", n)
	}
	if got := m.cur().list.Items()[0].(taskItem).t.Slug; got != "alpha" {
		t.Errorf("the in-progress view should hold alpha, got %q", got)
	}
	if !strings.Contains(m.cur().chip(), "view:in-progress") {
		t.Errorf("chip should announce the view, got %q", m.cur().chip())
	}
	// S steps backward: in-progress → "" (default), then wraps to the last entry.
	tm, _ = m.Update(press("S"))
	m = tm.(Model)
	tm, cmd = m.Update(press("S")) // "" → wrap to "all"
	m = drain(t, tm.(Model), cmd)
	if m.cur().statusView != "all" {
		t.Errorf("S past the default should wrap to all, got %q", m.cur().statusView)
	}
	if n := len(m.cur().list.Items()); n != 2 {
		t.Errorf(":all should show both tasks, got %d", n)
	}
}

func TestModel_StatusViewViaCommand(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press(":"))
	m = tm.(Model)
	for _, r := range "completed" {
		tm, _ = m.Update(press(string(r)))
		m = tm.(Model)
	}
	tm, cmd := m.Update(press("enter"))
	m = drain(t, tm.(Model), cmd)
	if m.cur().name != "tasks" || m.cur().statusView != "completed" {
		t.Fatalf(":completed should set the tasks completed view, got tab=%q view=%q", m.cur().name, m.cur().statusView)
	}
	if n := len(m.cur().list.Items()); n != 0 {
		t.Errorf("the seed has no completed tasks; expected an empty view, got %d", n)
	}
}

// TestStatusViewsCoverAllStatuses guards the unified status-view table against
// drift: every domain status must be reachable as a `:` view, so a new status
// can't silently become unbrowsable in the TUI.
func TestStatusViewsCoverAllStatuses(t *testing.T) {
	for _, s := range domain.AllStatuses() {
		if v, ok := statusViewFor(string(s)); !ok || v != string(s) {
			t.Errorf("status %q is not reachable as a `:` view (ok=%v value=%q)", s, ok, v)
		}
	}
	// The default + "all" specials must resolve too.
	if v, ok := statusViewFor("active"); !ok || v != "" {
		t.Errorf("`:active` should map to the working-set view, got ok=%v value=%q", ok, v)
	}
	if v, ok := statusViewFor("all"); !ok || v != "all" {
		t.Errorf("`:all` should map to the all view, got ok=%v value=%q", ok, v)
	}
}

func TestTaskFilterValueIncludesTags(t *testing.T) {
	it := taskItem{domain.Task{Slug: "x", Description: "desc", Tags: []string{"go", "cli"}}}
	fv := it.FilterValue()
	for _, want := range []string{"x", "desc", "go", "cli"} {
		if !strings.Contains(fv, want) {
			t.Errorf("FilterValue %q should contain %q (tags are filterable)", fv, want)
		}
	}
}

func TestModel_HelpOverlayTogglesAndFloats(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press("?"))
	m = tm.(Model)
	if !m.showHelp {
		t.Fatal("? should open the help overlay")
	}
	v := ansi.Strip(m.View())
	if !strings.Contains(v, "Keys") || !strings.Contains(v, "filter the list") {
		t.Errorf("help overlay should list keybindings:\n%s", v)
	}
	// It floats: the underlying list (alpha) stays partially visible around the box.
	if !strings.Contains(v, "alpha") {
		t.Errorf("help should float over the items, not blank them:\n%s", v)
	}
	// Any key dismisses it.
	tm, _ = m.Update(press("j"))
	m = tm.(Model)
	if m.showHelp {
		t.Error("a key press should dismiss the help overlay")
	}
}

func TestModel_DetailFindHighlightsAndNavigates(t *testing.T) {
	m := loaded(t, 120, 20)
	// Give the detail a body with two "find" matches on separate lines, plus
	// filler so the viewport must scroll.
	body := "alpha line\nbeta\nfind me here\nbeta again\nfind once more\n" + strings.Repeat("filler\n", 30)
	id := m.selectedID()
	tm, _ := m.Update(detailMsg{kind: entityTasks, id: id, content: taskDetail{t: domain.Task{Slug: id}, body: body}})
	m = tm.(Model)
	tm, _ = m.Update(press("l")) // focus detail
	m = tm.(Model)

	tm, _ = m.Update(press("/")) // open find
	m = tm.(Model)
	if !m.detail.finding() {
		t.Fatal("/ should open the find input in the detail pane")
	}
	for _, r := range "find" { // note: the 'n' is captured as input, not navigation
		tm, _ = m.Update(press(string(r)))
		m = tm.(Model)
	}
	tm, _ = m.Update(press("enter")) // apply
	m = tm.(Model)
	if m.detail.finding() {
		t.Error("enter should stop typing")
	}
	if !m.detail.findActive() {
		t.Fatal("enter should apply the query")
	}
	if n := len(m.detail.find.lines); n != 2 {
		t.Fatalf("expected 2 matching lines, got %d", n)
	}
	if m.detail.find.cur != 0 {
		t.Errorf("first match should be focused, got %d", m.detail.find.cur)
	}
	off1 := m.detail.vp.YOffset

	tm, _ = m.Update(press("n")) // next match
	m = tm.(Model)
	if m.detail.find.cur != 1 {
		t.Errorf("n should advance to the 2nd match, got %d", m.detail.find.cur)
	}
	if m.detail.vp.YOffset <= off1 {
		t.Errorf("n should scroll down to the lower match (%d → %d)", off1, m.detail.vp.YOffset)
	}
	if v := ansi.Strip(m.View()); !strings.Contains(v, "[2/2]") {
		t.Errorf("footer should show the match position [2/2]:\n%s", v)
	}

	tm, _ = m.Update(press("esc")) // first esc clears the find
	m = tm.(Model)
	if m.detail.findActive() {
		t.Error("esc should clear the active find")
	}
	if m.focus != focusDetail {
		t.Error("clearing the find should keep focus in the detail pane")
	}
	tm, _ = m.Update(press("esc")) // second esc leaves the detail
	m = tm.(Model)
	if m.focus != focusList {
		t.Error("a second esc should return to the list")
	}
}

func TestHighlightOccurrences(t *testing.T) {
	// lipgloss disables color off a TTY; force a profile so Render emits escapes.
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI)
	defer lipgloss.SetColorProfile(old)

	out := highlightOccurrences("the Find finds findings", "find", findMatch)
	// The plain text round-trips (highlight only adds escapes, never alters text).
	if ansi.Strip(out) != "the Find finds findings" {
		t.Errorf("highlight must preserve the plain text, got %q", ansi.Strip(out))
	}
	// Case-insensitive: "Find", "find", "find" each get styled → ≥3 escape runs.
	if got := strings.Count(out, "\x1b["); got < 3 {
		t.Errorf("expected each (case-insensitive) match styled, got %d escapes", got)
	}
}

// TestModel_HelpOverlayFitsTerminal extends the layout invariant to the `?`
// overlay: floating the help box must not change the view's height or overflow
// the width at any size.
func TestModel_HelpOverlayFitsTerminal(t *testing.T) {
	for _, d := range []struct{ w, h int }{
		{120, 40}, {100, 24}, {80, 20}, {40, 12}, {24, 8},
	} {
		m := loaded(t, d.w, d.h)
		tm, _ := m.Update(press("?"))
		m = tm.(Model)
		lines := strings.Split(m.View(), "\n")
		if len(lines) != d.h {
			t.Errorf("%dx%d with help: %d lines, want %d", d.w, d.h, len(lines), d.h)
		}
		for i, ln := range lines {
			if w := visibleW(ln); w > d.w {
				t.Errorf("%dx%d with help: line %d is %d wide > %d: %q", d.w, d.h, i, w, d.w, ansi.Strip(ln))
			}
		}
	}
}

// --- S3: fsnotify live reload ---

// TestModel_ReloadAllTabsPreservesCursors pins that a reload refreshes every
// loaded tab and keeps each cursor on its slug (not the active tab only).
func TestModel_ReloadAllTabsPreservesCursors(t *testing.T) {
	m := loaded(t, 120, 40)
	// tasks: move the cursor to beta.
	tm, _ := m.Update(press("j"))
	m = tm.(Model)
	if m.selectedID() != "beta" {
		t.Fatalf("setup: want beta on tasks, got %q", m.selectedID())
	}
	// Visit epics so that tab is loaded too, and note its selection.
	tm, cmd := m.Update(press("]"))
	m = drain(t, tm.(Model), cmd)
	if m.cur().name != "epics" {
		t.Fatalf("setup: expected epics, got %q", m.cur().name)
	}
	epicID := m.selectedID()
	// Back to tasks, then reload everything.
	tm, _ = m.Update(press("["))
	m = tm.(Model)
	tm, cmd = m.Update(reloadMsg{})
	m = drainBatch(t, tm.(Model), cmd)

	if m.selectedID() != "beta" {
		t.Errorf("tasks cursor should survive reload at beta, got %q", m.selectedID())
	}
	tm, _ = m.Update(press("]")) // epics is already loaded; no reload needed
	m = tm.(Model)
	if m.selectedID() != epicID {
		t.Errorf("epics cursor should survive reload at %q, got %q", epicID, m.selectedID())
	}
}

// TestModel_FsEventDebounces pins the coalescing: rapid fs events bump the
// generation, and only a debounce tick whose gen is still current reloads.
func TestModel_FsEventDebounces(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(fsEventMsg{})
	m = tm.(Model)
	tm, _ = m.Update(fsEventMsg{}) // a second event during the window
	m = tm.(Model)
	if m.dirtyGen != 2 {
		t.Fatalf("two events should advance dirtyGen to 2, got %d", m.dirtyGen)
	}
	// The first event's debounce is now stale → must not reload.
	if _, cmd := m.Update(debounceMsg{gen: 1}); cmd != nil {
		t.Error("a superseded debounce tick must not reload")
	}
	// The latest event's debounce fires the reload.
	_, cmd := m.Update(debounceMsg{gen: 2})
	if cmd == nil {
		t.Fatal("the current debounce tick should reload")
	}
	if _, ok := cmd().(reloadMsg); !ok {
		t.Fatalf("debounce should fire reloadMsg, got %T", cmd())
	}
}

func TestWatchDirs(t *testing.T) {
	root := filepath.Join("x", "plan")
	got := watchDirs(root)
	for _, want := range []string{
		filepath.Join(root, "epics"),
		filepath.Join(root, "tasks"),
		filepath.Join(root, "audits"),
		filepath.Join(root, "tasks", "in-progress"),
		filepath.Join(root, "tasks", "ready-to-start"),
		filepath.Join(root, "audits", "open"),
	} {
		found := false
		for _, d := range got {
			if d == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("watchDirs(%q) missing %q; got %v", root, want, got)
		}
	}
}
