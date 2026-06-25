package tui

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// seedRepo writes a tiny planning tree: alpha (in-progress), beta (ready-to-start),
// one epic, and one open audit so every tab has content.
func seedRepo(t *testing.T) string {
	t.Helper()
	r := testutil.NewRepo(t)
	task := func(status, slug, desc string) {
		body := fmt.Sprintf("---\nstatus: %s\nepic: 01-test\ndescription: %s\n---\n# %s\n", status, desc, slug)
		r.Task(status, slug+".md", body)
	}
	task("in-progress", "alpha", "the alpha task")
	task("ready-to-start", "beta", "the beta task")
	r.Epic("01-test.md", "---\nstatus: active\ndescription: a test epic\npriority: high\n---\n# Test epic\n")
	r.Audit("open", "2026-06-01-thing.md", "---\narea: store\ndate: 2026-06-01\n---\n# Audit\n")
	return r.Root
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
	return New(core.NewService(store.NewFS(root)))
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

func press(s string) tea.KeyPressMsg {
	// v2 keys are a single struct: special keys carry only a Code (so String()
	// renders the name, e.g. "enter"), ctrl combos a Code+Mod, and printable runes
	// a Code+Text (Text is what String() returns and what inputs insert).
	switch s {
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	case "ctrl+o":
		return tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl}
	default:
		return tea.KeyPressMsg{Code: []rune(s)[0], Text: s}
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
	if !strings.Contains(m.View().Content, "alpha") || !strings.Contains(m.View().Content, "beta") {
		t.Errorf("view should list both tasks:\n%s", m.View().Content)
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
	tm, _ = m.Update(detailMsg{kind: entityTasks, id: "alpha", gen: m.detailGen, content: taskDetail{t: domain.Task{Slug: "alpha"}, body: "x"}})
	m = tm.(Model)
	if m.detail.hasContent {
		t.Error("stale body for a different selection must be ignored")
	}
	// An outdated request generation for the RIGHT id must be dropped too (two
	// loads for the same id aren't ordered by (kind, id) alone).
	tm, _ = m.Update(detailMsg{kind: entityTasks, id: "beta", gen: m.detailGen - 1, content: taskDetail{t: domain.Task{Slug: "beta"}, body: "old"}})
	m = tm.(Model)
	if m.detail.hasContent {
		t.Error("an older detail load for the same id must be ignored")
	}
	// The matching body sets the detail.
	tm, _ = m.Update(detailMsg{kind: entityTasks, id: "beta", gen: m.detailGen, content: taskDetail{t: domain.Task{Slug: "beta"}, body: "beta body"}})
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
	_ = m.View().Content
}

func TestModel_BodyErrorDoesNotBrick(t *testing.T) {
	m := loaded(t, 120, 40)
	slug := m.selectedID()
	// An ambiguous-slug (duplicate across dirs) body error must not blank the UI.
	tm, _ := m.Update(detailErrMsg{kind: entityTasks, id: slug, gen: m.detailGen, err: domain.ErrAmbiguous})
	m = tm.(Model)
	if m.cur().loadErr != nil {
		t.Error("a per-task body error must not set the tab's list-load error")
	}
	if !m.detail.hasContent {
		t.Error("the error should be shown in the detail pane")
	}
	if !strings.Contains(m.View().Content, "alpha") {
		t.Errorf("the list must still render (not an error screen):\n%s", m.View().Content)
	}
}

func TestModel_DetailScrollKeys(t *testing.T) {
	m := loaded(t, 120, 12) // short, so the body overflows
	slug := m.selectedID()
	long := strings.Repeat("a line of text\n", 60)
	tm, _ := m.Update(detailMsg{kind: entityTasks, id: slug, gen: m.detailGen, content: taskDetail{t: domain.Task{Slug: slug}, body: long}})
	m = tm.(Model)
	tm, _ = m.Update(press("l")) // focus detail
	m = tm.(Model)

	tm, _ = m.Update(press("G")) // bottom
	m = tm.(Model)
	if m.detail.vp.YOffset() == 0 {
		t.Error("G should scroll the detail pane down")
	}
	tm, _ = m.Update(press("g")) // top
	m = tm.(Model)
	if m.detail.vp.YOffset() != 0 {
		t.Errorf("g should scroll to top, YOffset=%d", m.detail.vp.YOffset())
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
		tm.Send(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter}) // apply the filter
	tm.Send(press("q"))                          // quit (no longer filtering)

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
		tm, _ := m.Update(detailMsg{kind: entityTasks, id: m.selectedID(), gen: m.detailGen, content: taskDetail{t: domain.Task{Slug: m.selectedID()}, body: "# body\n\nsome text here\n"}})
		m = tm.(Model)
		lines := strings.Split(m.View().Content, "\n")
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
	v := m.View().Content
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
	m := New(core.NewService(store.NewFS(root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 14})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = tm.(Model)
	tm, _ = m.Update(press(":")) // open the command bar (bottom chrome)
	m = tm.(Model)

	lines := strings.Split(m.View().Content, "\n")
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
	m := New(core.NewService(store.NewFS(root)))
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
		e:     domain.Epic{ID: "17-x", Status: "active", Priority: "high"},
		tasks: []domain.Task{{Slug: "a", Status: domain.StatusCompleted}, {Slug: "b", Status: domain.StatusReadyToStart}},
		body:  "# Epic body",
	}
	out := ansi.Strip(epic.meta(70) + "\n" + epic.rawBody())
	for _, want := range []string{"17-x", "1/2", "50%", "Epic body"} {
		if !strings.Contains(out, want) {
			t.Errorf("epic detail missing %q:\n%s", want, out)
		}
	}
	if epic.Title() != "17-x" {
		t.Errorf("epic title = %q", epic.Title())
	}

	// Regression: the detail progress must exclude deprecated tasks from the
	// denominator (mirroring the rollup / epic list), not divide by len(tasks).
	dep := epicDetail{
		e: domain.Epic{ID: "18-x"},
		tasks: []domain.Task{
			{Slug: "a", Status: domain.StatusCompleted},
			{Slug: "b", Status: domain.StatusCompleted},
			{Slug: "c", Status: domain.StatusDeprecated},
		},
	}
	depOut := ansi.Strip(dep.meta(70))
	for _, want := range []string{"2/2", "100%", "1 deprecated"} {
		if !strings.Contains(depOut, want) {
			t.Errorf("epic detail should exclude deprecated (want %q):\n%s", want, depOut)
		}
	}
	if strings.Contains(depOut, "2/3") {
		t.Errorf("a deprecated task must not be in the detail denominator:\n%s", depOut)
	}

	audit := auditDetail{
		a:    domain.Audit{Slug: "2026-06-01-x", Bucket: domain.AuditOpen, Area: "store", Findings: 5, OpenFindings: 2, DoneFindings: 3},
		body: "# Audit body",
	}
	out = ansi.Strip(audit.meta(70) + "\n" + audit.rawBody())
	for _, want := range []string{"2026-06-01-x", "store", "3/5", "60%", "2 open", "Audit body"} {
		if !strings.Contains(out, want) {
			t.Errorf("audit detail missing %q:\n%s", want, out)
		}
	}
}

// TestAuditDetailFindingIndex pins the glyph-coded finding index the audit detail
// renders above the prose body: one scannable line per finding (status glyph +
// code + title), the audit analog of the epic detail's task list.
func TestAuditDetailFindingIndex(t *testing.T) {
	d := auditDetail{
		a: domain.Audit{Slug: "2026-06-01-x", Bucket: domain.AuditOpen, Area: "store", Findings: 2, OpenFindings: 1},
		body: "# Audit\n" +
			"#### H1. retry waste  · **Status:** open\n\n" +
			"#### M2. dead method  · **Status:** fixed\n",
	}
	meta := ansi.Strip(d.meta(70))
	// each finding's code + title, glyph-coded by status (○ open, ✔ fixed).
	for _, want := range []string{"○", "H1", "retry waste", "✔", "M2", "dead method"} {
		if !strings.Contains(meta, want) {
			t.Errorf("finding index missing %q:\n%s", want, meta)
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
		[]byte("---\nstatus: active\ndescription: x\n---\n# body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := New(core.NewService(store.NewFS(root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 24}) // narrowest two-pane
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = tm.(Model)
	tm, _ = m.Update(detailMsg{kind: entityTasks, id: slug, gen: m.detailGen, content: taskDetail{t: domain.Task{Slug: slug}, body: "body"}})
	m = tm.(Model)

	lines := strings.Split(m.View().Content, "\n")
	if len(lines) != 24 {
		t.Fatalf("view must be 24 lines, got %d", len(lines))
	}
	// The pane bottom-border row is just above the footer; both panes must close.
	border := ansi.Strip(lines[len(lines)-2])
	if got := strings.Count(border, "╯"); got != 2 {
		t.Errorf("both panes must keep their bottom border (2 ╯), got %d: %q", got, border)
	}
}

// TestModel_RecoversFromFailedInitialLoad pins the `r` escape hatch: a failed
// FIRST load (nothing `loaded` yet) shows the error pane, and the r key — the
// path a user actually has — must reload the active tab and recover. (The old
// version of this test recovered via m.Init()(), a path no key takes, which
// masked the dead-end.)
func TestModel_RecoversFromFailedInitialLoad(t *testing.T) {
	m := newModel(t)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	// The initial load fails before anything is loaded.
	tm, _ = m.Update(errMsg{kind: entityTasks, gen: m.cur().loadGen, err: domain.ErrNotFound})
	m = tm.(Model)
	if m.cur().loadErr == nil || !strings.Contains(m.View().Content, "error:") {
		t.Fatal("a failed initial load should show the error pane")
	}
	// r → reloadMsg → reloadAll must reload the active tab even though !loaded.
	tm, cmd := m.Update(press("r"))
	m = tm.(Model)
	tm, cmd = m.Update(cmd()) // reloadMsg
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("r must produce a reload for the active tab after a failed initial load")
	}
	m = pump(t, m, cmd, 8)
	if !m.cur().loaded || m.cur().loadErr != nil {
		t.Errorf("r should recover the session: loaded=%v err=%v", m.cur().loaded, m.cur().loadErr)
	}
	if !strings.Contains(m.View().Content, "alpha") {
		t.Error("the recovered list should render its rows")
	}
}

// TestModel_LoadErrorIsPerTab pins that one tab's loader failing neither blanks
// the other tabs nor loses the failing tab's last good rows (the failure is
// flagged in the footer instead).
func TestModel_LoadErrorIsPerTab(t *testing.T) {
	m := loaded(t, 120, 40)
	// Visit epics so it's loaded, then fail a tasks reload in the background.
	tm, cmd := m.Update(press("]"))
	m = drain(t, tm.(Model), cmd)
	tasks := m.tabs[indexOfKind(m.tabs, entityTasks)]
	tm, _ = m.Update(errMsg{kind: entityTasks, gen: tasks.loadGen, err: domain.ErrNotFound})
	m = tm.(Model)
	// The active epics tab must be untouched.
	if !strings.Contains(m.View().Content, "01-test") {
		t.Errorf("a background tab's failure must not blank the active tab:\n%s", m.View().Content)
	}
	// Back on tasks: the stale rows survive, and the footer flags the failure.
	tm, _ = m.Update(press("["))
	m = tm.(Model)
	v := ansi.Strip(m.View().Content)
	if !strings.Contains(v, "alpha") {
		t.Error("a failed reload should keep the last good rows visible")
	}
	if !strings.Contains(v, "reload failed") {
		t.Errorf("the footer should flag the failed reload:\n%s", v)
	}
	// A stale failure (older generation) must be dropped entirely.
	tm, _ = m.Update(errMsg{kind: entityEpics, gen: m.tabs[indexOfKind(m.tabs, entityEpics)].loadGen - 1, err: domain.ErrNotFound})
	m = tm.(Model)
	if m.tabs[indexOfKind(m.tabs, entityEpics)].loadErr != nil {
		t.Error("an error from a superseded load generation must be ignored")
	}
}

// TestModel_StaleListLoadDropped pins the load-generation guard: an older list
// load finishing after a newer one must not clobber its result.
func TestModel_StaleListLoadDropped(t *testing.T) {
	m := loaded(t, 120, 40)
	tab := m.cur()
	if n := len(tab.list.Items()); n != 2 {
		t.Fatalf("setup: want 2 items, got %d", n)
	}
	stale := listLoadedMsg{kind: entityTasks, gen: tab.loadGen - 1,
		items: []list.Item{taskItem{domain.Task{Slug: "ghost", Status: domain.StatusInProgress}}}}
	tm, _ := m.Update(stale)
	m = tm.(Model)
	if n := len(m.cur().list.Items()); n != 2 {
		t.Errorf("a stale list load must be dropped, got %d items", n)
	}
}

// TestModel_DetailScrollSurvivesReload pins the live-reload affordance: a
// refresh of the item already on screen keeps the scroll position; a different
// item still snaps to the top.
func TestModel_DetailScrollSurvivesReload(t *testing.T) {
	m := loaded(t, 120, 12)
	slug := m.selectedID()
	long := strings.Repeat("a line of text\n", 60)
	feed := func(id, body string) {
		tm, _ := m.Update(detailMsg{kind: entityTasks, id: id, gen: m.detailGen, content: taskDetail{t: domain.Task{Slug: id}, body: body}})
		m = tm.(Model)
	}
	feed(slug, long)
	tm, _ := m.Update(press("l"))
	m = tm.(Model)
	tm, _ = m.Update(press("G")) // scroll to bottom
	m = tm.(Model)
	off := m.detail.vp.YOffset()
	if off == 0 {
		t.Fatal("setup: the body should have scrolled")
	}
	feed(slug, long) // same item reloaded (e.g. an external write)
	if m.detail.vp.YOffset() != off {
		t.Errorf("a same-item refresh must keep the scroll: %d → %d", off, m.detail.vp.YOffset())
	}
	// A different item snaps back to the top.
	tm, _ = m.Update(press("h"))
	m = tm.(Model)
	tm, _ = m.Update(press("j"))
	m = tm.(Model)
	feed(m.selectedID(), long)
	if m.detail.vp.YOffset() != 0 {
		t.Errorf("a different item should start at the top, got offset %d", m.detail.vp.YOffset())
	}
}

// TestModel_EscInListFocusDoesNotQuit guards the embedded list's default quit
// binding (q AND esc): Esc in list focus must be a context no-op, not exit the
// program.
func TestModel_EscInListFocusDoesNotQuit(t *testing.T) {
	m := loaded(t, 120, 40)
	if m.focus != focusList {
		t.Fatal("setup: should be list-focused")
	}
	tm, cmd := m.Update(press("esc"))
	m = tm.(Model)
	if cmd != nil {
		if _, quits := cmd().(tea.QuitMsg); quits {
			t.Fatal("esc in list focus must not quit the app")
		}
	}
	if !strings.Contains(m.View().Content, "alpha") {
		t.Error("the browser should still be rendering after esc")
	}
}

// TestModel_QuitPopsSinglePaneDetail pins q's context-quit layering: in
// single-pane drill, q returns to the list (like Esc/h); from the list q quits.
func TestModel_QuitPopsSinglePaneDetail(t *testing.T) {
	m := loaded(t, 70, 24) // single-pane
	if m.twoPane {
		t.Fatal("setup: 70 cols should be single-pane")
	}
	tm, _ := m.Update(press("l")) // drill into detail
	m = tm.(Model)
	if m.focus != focusDetail {
		t.Fatal("setup: should be detail-focused")
	}
	tm, cmd := m.Update(press("q"))
	m = tm.(Model)
	if m.focus != focusList {
		t.Error("q from single-pane detail should pop back to the list")
	}
	if cmd != nil {
		if _, quits := cmd().(tea.QuitMsg); quits {
			t.Fatal("q from single-pane detail must not quit the app")
		}
	}
	// From the list, q still quits.
	_, cmd = m.Update(press("q"))
	if cmd == nil {
		t.Fatal("q from the list should quit")
	}
	if _, quits := cmd().(tea.QuitMsg); !quits {
		t.Errorf("q from the list should quit, got %T", cmd())
	}
	// In two-pane, q quits from detail focus too (it isn't a drill layer there).
	w := loaded(t, 120, 40)
	tm, _ = w.Update(press("l"))
	w = tm.(Model)
	_, cmd = w.Update(press("q"))
	if cmd == nil {
		t.Fatal("two-pane q should quit")
	}
	if _, quits := cmd().(tea.QuitMsg); !quits {
		t.Errorf("two-pane q from detail should quit, got %T", cmd())
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
	// The arrow shows the column's ACTUAL direction (sortArrow): slug defaults
	// to ascending → ↑; reversing flips both the order and the glyph.
	if !strings.Contains(m.cur().chip(), "sort:slug↑") {
		t.Errorf("chip should announce ascending slug sort, got %q", m.cur().chip())
	}
	// Reverse → beta sorts before alpha, and the arrow flips to descending.
	tm, _ := m.Update(press("O"))
	m = tm.(Model)
	if got := m.cur().list.Items()[0].(taskItem).t.Slug; got != "beta" {
		t.Errorf("reversed slug sort should lead with beta, got %q", got)
	}
	if !strings.Contains(m.cur().chip(), "sort:slug↓") {
		t.Errorf("reversed chip should show ↓, got %q", m.cur().chip())
	}
}

// TestSortArrow pins the per-column direction semantics: updated is the one
// column that defaults to descending (newest first).
func TestSortArrow(t *testing.T) {
	for _, tc := range []struct {
		k    sortKey
		rev  bool
		want string
	}{
		{sortSlug, false, "↑"}, {sortSlug, true, "↓"},
		{sortPriority, false, "↑"},
		{sortUpdated, false, "↓"}, {sortUpdated, true, "↑"},
	} {
		if got := sortArrow(tc.k, tc.rev); got != tc.want {
			t.Errorf("sortArrow(%v, %v) = %q, want %q", tc.k, tc.rev, got, tc.want)
		}
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

// auditModel returns a loaded model whose repo has one audit in each bucket, so
// the bucket-view axis has something distinct to show per view.
func auditModel(t *testing.T) Model {
	t.Helper()
	r := testutil.NewRepo(t)
	r.Epic("01-test.md", "---\nstatus: active\ndescription: a test epic\n---\n# Test epic\n")
	r.Audit("open", "2026-06-01-open-a.md", "---\narea: store\ndate: 2026-06-01\n---\n# Open A\n")
	r.Audit("closed", "2026-05-01-closed-a.md", "---\narea: cli\ndate: 2026-05-01\n---\n# Closed A\n")
	r.Audit("deferred", "2026-04-01-deferred-a.md", "---\narea: tui\ndate: 2026-04-01\n---\n# Deferred A\n")
	m := New(core.NewService(store.NewFS(r.Root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	return tm.(Model)
}

// auditSlugs is the slugs of the active list's audit rows, in order.
func auditSlugs(m Model) []string {
	items := m.cur().list.Items()
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.(auditItem).id())
	}
	return out
}

// cmdJump types `:word<enter>` and drains the resulting load — the command-bar
// equivalent of a tab/view jump.
func cmdJump(t *testing.T, m Model, word string) Model {
	t.Helper()
	tm, _ := m.Update(press(":"))
	m = tm.(Model)
	for _, r := range word {
		tm, _ = m.Update(press(string(r)))
		m = tm.(Model)
	}
	tm, cmd := m.Update(press("enter"))
	return drain(t, tm.(Model), cmd)
}

func TestModel_AuditBucketCyclesAndFilters(t *testing.T) {
	m := cmdJump(t, auditModel(t), "audits")
	if m.cur().kind != entityAudits {
		t.Fatalf(":audits should select the audits tab, got %q", m.cur().name)
	}
	// Default view is the open bucket: open audit only, and a silent chip.
	if got := auditSlugs(m); len(got) != 1 || got[0] != "2026-06-01-open-a" {
		t.Fatalf("default audits view should be open-only, got %v", got)
	}
	if c := m.cur().chip(); c != "" {
		t.Errorf("the open (default) bucket should render no chip, got %q", c)
	}
	// s → next bucket (closed): closed audit only, chip announces it.
	tm, cmd := m.Update(press("s"))
	m = drain(t, tm.(Model), cmd)
	if m.cur().statusView != "closed" {
		t.Fatalf("s should advance to the closed bucket, got %q", m.cur().statusView)
	}
	if got := auditSlugs(m); len(got) != 1 || got[0] != "2026-05-01-closed-a" {
		t.Errorf("closed bucket should list the closed audit, got %v", got)
	}
	if !strings.Contains(m.cur().chip(), "view:closed") {
		t.Errorf("chip should announce the bucket, got %q", m.cur().chip())
	}
	// S steps backward: closed → "" (open default), then wraps to all.
	tm, _ = m.Update(press("S"))
	m = tm.(Model)
	tm, cmd = m.Update(press("S")) // "" → wrap to all
	m = drain(t, tm.(Model), cmd)
	if m.cur().statusView != "all" {
		t.Fatalf("S past the default should wrap to all, got %q", m.cur().statusView)
	}
	if n := len(auditSlugs(m)); n != 3 {
		t.Errorf(":all should span every bucket (3 audits), got %d", n)
	}
}

// TestModel_AuditViewViaCommand covers the `:` words, including the shared
// "all"/"deferred" words resolving against the active tab (audits here) rather
// than always falling through to tasks.
func TestModel_AuditViewViaCommand(t *testing.T) {
	// :closed is audits-only — reachable straight from the tasks tab.
	m := cmdJump(t, auditModel(t), "closed")
	if m.cur().kind != entityAudits || m.cur().statusView != "closed" {
		t.Fatalf(":closed should land on the audits closed bucket, got tab=%q view=%q", m.cur().name, m.cur().statusView)
	}
	if got := auditSlugs(m); len(got) != 1 || got[0] != "2026-05-01-closed-a" {
		t.Errorf("closed bucket should list the closed audit, got %v", got)
	}
	// :all is shared with tasks; on the audits tab it acts in context (audits all).
	m = cmdJump(t, m, "all")
	if m.cur().kind != entityAudits || m.cur().statusView != "all" {
		t.Fatalf(":all on the audits tab should stay on audits, got tab=%q view=%q", m.cur().name, m.cur().statusView)
	}
	if n := len(auditSlugs(m)); n != 3 {
		t.Errorf("audits :all should span every bucket, got %d", n)
	}
	// :deferred (also shared) selects the audits deferred bucket from audits.
	m = cmdJump(t, m, "deferred")
	if m.cur().kind != entityAudits || m.cur().statusView != "deferred" {
		t.Fatalf(":deferred on audits should select the deferred bucket, got tab=%q view=%q", m.cur().name, m.cur().statusView)
	}
	// Back-compat: :all from the tasks tab still targets tasks.
	m = cmdJump(t, m, "tasks")
	m = cmdJump(t, m, "all")
	if m.cur().kind != entityTasks || m.cur().statusView != "all" {
		t.Fatalf(":all from the tasks tab should target tasks, got tab=%q view=%q", m.cur().name, m.cur().statusView)
	}
}

// TestStatusViewsCoverAllStatuses guards the unified status-view table against
// drift: every domain status must be reachable as a `:` view, so a new status
// can't silently become unbrowsable in the TUI.
func TestStatusViewsCoverAllStatuses(t *testing.T) {
	for _, s := range domain.AllStatuses() {
		if v, ok := viewFor(statusViews, statusViewAliases, string(s)); !ok || v != string(s) {
			t.Errorf("status %q is not reachable as a `:` view (ok=%v value=%q)", s, ok, v)
		}
	}
	// The default + "all" specials must resolve too.
	if v, ok := viewFor(statusViews, statusViewAliases, "active"); !ok || v != "" {
		t.Errorf("`:active` should map to the working-set view, got ok=%v value=%q", ok, v)
	}
	if v, ok := viewFor(statusViews, statusViewAliases, "all"); !ok || v != "all" {
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
	// Short enough that the focus-filtered help still overflows its box, so j/k
	// scrolling is exercised (context filtering hides the inactive pane's keys).
	m := loaded(t, 120, 24)
	tm, _ := m.Update(press("?"))
	m = tm.(Model)
	if !m.showHelp {
		t.Fatal("? should open the help overlay")
	}
	v := ansi.Strip(m.View().Content)
	if !strings.Contains(v, "Keys") || !strings.Contains(v, "filter the list") {
		t.Errorf("help overlay should list keybindings:\n%s", v)
	}
	// It floats: the underlying list (alpha) stays partially visible around the box.
	if !strings.Contains(v, "alpha") {
		t.Errorf("help should float over the items, not blank them:\n%s", v)
	}
	// j/k scroll the overlay (it can outgrow a short terminal) without closing.
	tm, _ = m.Update(press("j"))
	m = tm.(Model)
	if !m.showHelp || m.helpScroll != 1 {
		t.Errorf("j should scroll the help, not dismiss it (open=%v scroll=%d)", m.showHelp, m.helpScroll)
	}
	tm, _ = m.Update(press("k"))
	m = tm.(Model)
	if m.helpScroll != 0 {
		t.Errorf("k should scroll back up, got %d", m.helpScroll)
	}
	// Any other key dismisses it (and resets the scroll).
	tm, _ = m.Update(press("x"))
	m = tm.(Model)
	if m.showHelp {
		t.Error("a key press should dismiss the help overlay")
	}
}

// TestModel_HelpScrollRevealsTail pins that a short terminal can still reach
// the bottom help entries by scrolling: the audits scope note is the last line.
func TestModel_HelpScrollRevealsTail(t *testing.T) {
	m := loaded(t, 100, 14) // too short for the full help content
	tm, _ := m.Update(press("?"))
	m = tm.(Model)
	// "force-quit" is the last row of the now-last Global section (Global sorts
	// after the active pane's keys and Notes), so it's the tail to scroll to.
	if v := ansi.Strip(m.View().Content); strings.Contains(v, "force-quit") {
		t.Skip("terminal tall enough to show the tail without scrolling")
	}
	for i := 0; i < len(helpLines(m.focus)); i++ { // scroll past the clamp
		tm, _ = m.Update(press("j"))
		m = tm.(Model)
	}
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "force-quit") {
		t.Errorf("scrolling should reveal the last help entries:\n%s", v)
	}
	// The layout invariant holds while scrolled.
	lines := strings.Split(m.View().Content, "\n")
	if len(lines) != 14 {
		t.Errorf("scrolled help: view has %d lines, want 14", len(lines))
	}
}

func TestModel_DetailFindHighlightsAndNavigates(t *testing.T) {
	m := loaded(t, 120, 20)
	// Give the detail a body with two "find" matches on separate lines, plus
	// filler so the viewport must scroll.
	body := "alpha line\nbeta\nfind me here\nbeta again\nfind once more\n" + strings.Repeat("filler\n", 30)
	id := m.selectedID()
	tm, _ := m.Update(detailMsg{kind: entityTasks, id: id, gen: m.detailGen, content: taskDetail{t: domain.Task{Slug: id}, body: body}})
	m = tm.(Model)
	tm, _ = m.Update(press("l")) // focus detail
	m = tm.(Model)
	tm, _ = m.Update(press("R")) // raw mode: deterministic line structure (glamour
	m = tm.(Model)               // would paragraph-collapse this body)

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
	if n := len(m.detail.find.matches); n != 2 {
		t.Fatalf("expected 2 occurrences, got %d", n)
	}
	if m.detail.find.cur != 0 {
		t.Errorf("first match should be focused, got %d", m.detail.find.cur)
	}
	off1 := m.detail.vp.YOffset()

	tm, _ = m.Update(press("n")) // next match
	m = tm.(Model)
	if m.detail.find.cur != 1 {
		t.Errorf("n should advance to the 2nd match, got %d", m.detail.find.cur)
	}
	if m.detail.vp.YOffset() <= off1 {
		t.Errorf("n should scroll down to the lower match (%d → %d)", off1, m.detail.vp.YOffset())
	}
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "[2/2]") {
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

func TestFoldMatches(t *testing.T) {
	// Case-insensitive, non-overlapping, left-to-right; ranges index the original.
	got := foldMatches("the Find finds findings", "find")
	if len(got) != 3 {
		t.Fatalf("want 3 occurrences (Find/find/find), got %d: %v", len(got), got)
	}
	for _, r := range got {
		if strings.ToLower("the Find finds findings"[r[0]:r[1]]) != "find" {
			t.Errorf("range %v does not bound a 'find': %q", r, "the Find finds findings"[r[0]:r[1]])
		}
	}
	if got := foldMatches("abc", "x"); got != nil {
		t.Errorf("no match should be nil, got %v", got)
	}
}

// TestFoldMatchesUnicode is the regression for the length-changing case fold (the
// U+0130 class): folding must not misalign byte offsets or index past the end.
func TestFoldMatchesUnicode(t *testing.T) {
	// "İ" (U+0130, 2 bytes) lowercases to "i̇" (2 runes / 3 bytes) under ToLower —
	// the exact mismatch the old strings.ToLower-then-slice approach tripped on.
	s := "the İstanbul plan" // contains the dotted capital I
	for _, q := range []string{"plan", "the", "İ"} {
		for _, r := range foldMatches(s, q) {
			if r[0] < 0 || r[1] > len(s) || r[0] > r[1] {
				t.Fatalf("foldMatches(%q,%q) produced an out-of-range span %v (len=%d)", s, q, r, len(s))
			}
		}
	}
	// A plain ASCII query after the multibyte rune still resolves to the right text.
	got := foldMatches(s, "plan")
	if len(got) != 1 || s[got[0][0]:got[0][1]] != "plan" {
		t.Errorf("expected one 'plan' match with correct offsets, got %v", got)
	}
}

func TestHighlightLine(t *testing.T) {
	// v2 lipgloss has no global color profile: Render always emits ANSI escapes
	// (downsampling happens at the output writer), so no profile setup is needed —
	// the styled prefix below carries its escape unconditionally.

	// A styled line: a green prefix glyph then plain text. Highlighting "find"
	// must keep the prefix's color and only restyle the match.
	plain := "● find me"
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("●") + " find me"
	occ := foldMatches(plain, "find")
	out := highlightLine(styled, plain, occ, occ[0][0]) // the lone match is current
	if ansi.Strip(out) != plain {
		t.Errorf("highlight must preserve the plain text, got %q", ansi.Strip(out))
	}
	// The green prefix escape survives (field colors preserved around the match).
	if !strings.Contains(out, "\x1b[") {
		t.Error("the line should still carry styling")
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
		lines := strings.Split(m.View().Content, "\n")
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

// pump applies a Cmd and recursively feeds every resulting message back through
// Update (depth-limited) — enough to settle the reload → SetItems → routed
// refilter → cursor-restore chain that drain (one level) can't reach.
func pump(t *testing.T, m Model, cmd tea.Cmd, depth int) Model {
	t.Helper()
	if cmd == nil || depth <= 0 {
		return m
	}
	msg := cmd()
	if msg == nil {
		return m
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			m = pump(t, m, c, depth-1)
		}
		return m
	}
	tm, next := m.Update(msg)
	return pump(t, tm.(Model), next, depth-1)
}

// TestModel_ReloadWithBackgroundFilterApplied pins the filter×reload misroute:
// a reload while a BACKGROUND tab has an applied filter must re-deliver that
// tab's async filter matches to that tab (not the active one). Before the fix,
// the background tab came back blank (FilterApplied with nil matches) and the
// active tab could receive the other entity's match set.
func TestModel_ReloadWithBackgroundFilterApplied(t *testing.T) {
	m := loaded(t, 120, 40)
	// Apply a filter on tasks (synchronously, via the list's programmatic API)
	// and put the cursor on its only match.
	tasks := m.tabs[indexOfKind(m.tabs, entityTasks)]
	tasks.list.SetFilterText("beta")
	if n := len(tasks.list.VisibleItems()); n != 1 {
		t.Fatalf("setup: filter should leave 1 visible task, got %d", n)
	}
	// Switch to epics (tasks is now a filtered background tab), then reload all.
	tm, cmd := m.Update(press("]"))
	m = drain(t, tm.(Model), cmd)
	if m.cur().name != "epics" {
		t.Fatalf("setup: expected epics active, got %q", m.cur().name)
	}
	tm, cmd = m.Update(reloadMsg{})
	m = pump(t, tm.(Model), cmd, 8)

	// The background tasks tab keeps its filter AND its matches.
	tasks = m.tabs[indexOfKind(m.tabs, entityTasks)]
	if tasks.list.FilterState() != list.FilterApplied {
		t.Fatalf("tasks filter should survive the reload, state=%v", tasks.list.FilterState())
	}
	vis := tasks.list.VisibleItems()
	if len(vis) != 1 {
		t.Fatalf("filtered background tab must not go blank after a reload: %d visible", len(vis))
	}
	if it, _ := vis[0].(taskItem); it.t.Slug != "beta" {
		t.Errorf("expected beta visible after reload, got %q", it.t.Slug)
	}
	// The cursor restore (pending until the refilter landed) points at beta.
	if id := tasks.list.SelectedItem().(entityItem).id(); id != "beta" {
		t.Errorf("cursor should be restored to beta after the refilter, got %q", id)
	}
	// The active epics tab was not polluted by the tasks tab's matches.
	epics := m.tabs[indexOfKind(m.tabs, entityEpics)]
	if n := len(epics.list.VisibleItems()); n != 1 {
		t.Errorf("epics must keep its own (1) row, got %d", n)
	}
	if _, ok := epics.list.VisibleItems()[0].(epicItem); !ok {
		t.Errorf("epics rows must stay epicItems, got %T", epics.list.VisibleItems()[0])
	}
}

// TestModel_DetailFollowsFilter runs the real program (the list's filter is
// async) and pins A1: while a `/` filter is typed, the detail pane follows the
// moving selection instead of showing the pre-filter item until j/k is pressed.
func TestModel_DetailFollowsFilter(t *testing.T) {
	tm := teatest.NewTestModel(t, newModel(t), teatest.WithInitialTermSize(120, 40))
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("beta"))
	}, teatest.WithDuration(3*time.Second))
	tm.Send(press("/"))
	for _, r := range "beta" {
		tm.Send(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	// While still typing, the filter narrows to beta and the detail must follow:
	// beta's status (ready-to-start) appears only via its detail pane — the rows
	// render the status as a glyph, the body as text.
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("ready-to-start"))
	}, teatest.WithDuration(3*time.Second))
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter}) // apply
	tm.Send(press("q"))                          // quit
	fm := tm.FinalModel(t, teatest.WithFinalTimeout(3*time.Second)).(Model)
	if fm.detail.title != "beta" {
		t.Errorf("detail should track the filtered selection, showing %q", fm.detail.title)
	}
}
