package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/theme"
)

// loadedDashAt builds a model sized + initialized on the landing dashboard, over a
// custom repo root (loadedDash is the seedRepo variant).
func loadedDashAt(t *testing.T, root string, w, h int) Model {
	t.Helper()
	m := New(core.NewService(store.NewFS(root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	return tm.(Model)
}

// TestModel_DashboardIsDefaultLanding pins that the TUI opens onto the dashboard
// (not the tasks tab) and renders the v1 widgets.
func TestModel_DashboardIsDefaultLanding(t *testing.T) {
	m := loadedDash(t, 120, 40)
	if !m.onDash {
		t.Fatal("the TUI should land on the dashboard by default")
	}
	v := ansi.Strip(m.View().Content)
	for _, want := range []string{"overview", "in progress", "epics", "needs attention"} {
		if !strings.Contains(v, want) {
			t.Errorf("dashboard view should show %q:\n%s", want, v)
		}
	}
}

// TestModel_DashboardEnterJumpsToItem pins navigational behavior: enter on the
// first (in-progress) row leaves the dashboard for that task on the tasks tab.
func TestModel_DashboardEnterJumpsToItem(t *testing.T) {
	m := loadedDash(t, 120, 40) // seedRepo: alpha is the lone in-progress task
	tm, cmd := m.Update(press("enter"))
	m = drain(t, tm.(Model), cmd)
	if m.onDash {
		t.Fatal("enter should leave the dashboard")
	}
	if m.cur().kind != entityTasks || m.selectedID() != "alpha" {
		t.Errorf("enter on the in-progress row should select alpha on tasks, got tab=%q id=%q", m.cur().name, m.selectedID())
	}
}

// TestModel_DashboardTabNavWraps pins that ] leaves the dashboard for the first
// tab and [ from the first tab wraps back to the dashboard.
func TestModel_DashboardTabNavWraps(t *testing.T) {
	m := loadedDash(t, 120, 40)
	tm, cmd := m.Update(press("]"))
	m = drain(t, tm.(Model), cmd)
	if m.onDash || m.cur().name != "tasks" {
		t.Fatalf("] should leave the dashboard for tasks, got onDash=%v tab=%q", m.onDash, m.cur().name)
	}
	tm, cmd = m.Update(press("["))
	m = drain(t, tm.(Model), cmd)
	if !m.onDash {
		t.Error("[ from the first tab should wrap back to the dashboard")
	}
}

// TestModel_DashboardRevisitRowOpensView pins that the due-for-revisit row routes
// to the :revisit view (a view jump, not an item jump).
func TestModel_DashboardRevisitRowOpensView(t *testing.T) {
	m := loadedDashAt(t, revisitRepo(t), 120, 40) // alpha in-progress + one due-for-revisit task
	tm, _ := m.Update(press("j"))                 // in-progress row → due-for-revisit row
	m = tm.(Model)
	tm, cmd := m.Update(press("enter"))
	m = drain(t, tm.(Model), cmd)
	if m.onDash || m.cur().kind != entityTasks || m.cur().statusView != "revisit" {
		t.Errorf("the due-for-revisit row should open :revisit, got onDash=%v tab=%q view=%q",
			m.onDash, m.cur().name, m.cur().statusView)
	}
}

// TestModel_DashboardCommand pins the :dashboard / :d commands (the dashboard tab
// is otherwise `:`-unreachable) — both return to the landing screen from a tab.
func TestModel_DashboardCommand(t *testing.T) {
	for _, word := range []string{"overview", "o"} {
		m := loaded(t, 120, 40) // starts on the tasks tab
		if m.onDash {
			t.Fatal("setup: should be on a tab, not the dashboard")
		}
		m = cmdJump(t, m, word)
		if !m.onDash {
			t.Errorf(":%s should return to the dashboard, got tab=%q onDash=%v", word, m.cur().name, m.onDash)
		}
	}
}

// TestModel_CommandHintListsCommands pins the inline `:` discovery: an empty prompt
// lists the command vocabulary, and typing a prefix narrows it.
func TestModel_CommandHintListsCommands(t *testing.T) {
	m := loaded(t, 120, 40)
	tm, _ := m.Update(press(":"))
	m = tm.(Model)
	all := m.commandHint()
	for _, w := range []string{"overview", "tasks", "revisit"} {
		if !strings.Contains(all, w) {
			t.Errorf("empty `:` hint should list %q, got %q", w, all)
		}
	}
	m = typeRunes(t, m, "rev")
	if h := m.commandHint(); h != "revisit" {
		t.Errorf(":rev should narrow the hint to revisit, got %q", h)
	}
}

// dashOverflowRepo writes more in-progress tasks than dashListCap plus a few epics,
// so the composed dashboard overruns a short terminal and a widget shows "+N more".
func dashOverflowRepo(t *testing.T) string {
	t.Helper()
	r := testutil.NewRepo(t)
	r.Epic("01-epic.md", "---\nstatus: active\ndescription: an epic\n---\n# Epic\n")
	for i := 0; i < 8; i++ { // 8 > dashListCap (6) → one "+2 more" overflow row
		slug := fmt.Sprintf("wip-%02d", i)
		r.Task("in-progress", slug+".md",
			fmt.Sprintf("---\nstatus: in-progress\nepic: 01-epic\ndescription: task %d\n---\n# %s\n", i, slug))
	}
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("%02d-roll", i+2)
		r.Epic(id+".md", fmt.Sprintf("---\nstatus: active\ndescription: epic %d\n---\n# %s\n", i, id))
	}
	return r.Root
}

// dashEpicsRepo writes two epics whose "last updated" (derived from their tasks'
// updated_at) differ, so the dashboard's recency ordering + date are observable.
func dashEpicsRepo(t *testing.T) string {
	t.Helper()
	r := testutil.NewRepo(t)
	r.Epic("01-stale.md", "---\nstatus: active\ndescription: old epic\n---\n# Stale\n")
	r.Epic("02-fresh.md", "---\nstatus: active\ndescription: new epic\n---\n# Fresh\n")
	r.Task("completed", "old.md",
		"---\nstatus: completed\nepic: 01-stale\ndescription: x\nupdated_at: 2025-02-01\n---\n# old\n")
	r.Task("in-progress", "fresh-wip.md",
		"---\nstatus: in-progress\nepic: 02-fresh\ndescription: y\nupdated_at: 2026-06-25\n---\n# fresh-wip\n")
	return r.Root
}

// TestModel_DashboardEpicsSortedByRecentWithDate pins the epics widget ordering
// (most-recently-updated first, derived from tasks) and the relative date shown.
func TestModel_DashboardEpicsSortedByRecentWithDate(t *testing.T) {
	m := loadedDashAt(t, dashEpicsRepo(t), 120, 40)
	v := ansi.Strip(m.View().Content)
	fresh, stale := strings.Index(v, "02-fresh"), strings.Index(v, "01-stale")
	if fresh < 0 || stale < 0 {
		t.Fatalf("both epics should be listed:\n%s", v)
	}
	if fresh > stale {
		t.Errorf("the more-recently-updated epic (02-fresh) should sort above 01-stale:\n%s", v)
	}
	if want := theme.RelativeDate("2026-06-25"); want != "" && !strings.Contains(v, want) {
		t.Errorf("the epic row should show the derived last-updated date %q:\n%s", want, v)
	}
}

// TestModel_DashboardNeedsAttentionLabelsAudits pins that the renamed
// "needs attention" section names each row's category — in particular the
// open-audit queue (a bare count under a non-specific heading wouldn't say what it
// refers to).
func TestModel_DashboardNeedsAttentionLabelsAudits(t *testing.T) {
	m := loadedDash(t, 120, 40) // seedRepo has one open audit
	v := ansi.Strip(m.View().Content)
	if !strings.Contains(v, "needs attention") {
		t.Errorf("the section should be titled 'needs attention':\n%s", v)
	}
	if !strings.Contains(v, "open audit") {
		t.Errorf("the open-audit row should be labeled as audits, not a bare count:\n%s", v)
	}
}

// dashAlignRepo writes two epics whose rollup counts differ in width ("0/1" vs
// "0/12"), both dated, so the column-alignment of the epic ids is observable.
func dashAlignRepo(t *testing.T) string {
	t.Helper()
	r := testutil.NewRepo(t)
	r.Epic("01-small.md", "---\nstatus: active\ndescription: e\n---\n# s\n")
	r.Epic("02-large.md", "---\nstatus: active\ndescription: e\n---\n# l\n")
	r.Task("ready-to-start", "s-0.md",
		"---\nstatus: ready-to-start\nepic: 01-small\ndescription: x\nupdated_at: 2026-06-26\n---\n# s0\n")
	for i := 0; i < 12; i++ { // 12 tasks → counts "0/12" (4 wide vs 01-small's "0/1")
		slug := fmt.Sprintf("l-%02d", i)
		r.Task("ready-to-start", slug+".md",
			fmt.Sprintf("---\nstatus: ready-to-start\nepic: 02-large\ndescription: x\nupdated_at: 2026-06-25\n---\n# %s\n", slug))
	}
	return r.Root
}

// TestModel_DashboardEpicsColumnsAlign pins that the epic ids start at the same
// column even when the counts differ in width — i.e. the counts are padded to a
// shared width, not left jagged.
func TestModel_DashboardEpicsColumnsAlign(t *testing.T) {
	m := loadedDashAt(t, dashAlignRepo(t), 120, 40)
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	col := func(id string) int {
		for _, ln := range lines {
			if i := strings.Index(ln, id); i >= 0 {
				return visibleW(ln[:i]) // display column, not byte offset (the bar/cursor are multibyte)
			}
		}
		return -1
	}
	a, b := col("01-small"), col("02-large")
	if a < 0 || b < 0 {
		t.Fatalf("both epic ids should render (cols %d, %d):\n%s", a, b, strings.Join(lines, "\n"))
	}
	if a != b {
		t.Errorf("epic ids should start at the same column (counts padded), got %d vs %d:\n%s",
			a, b, strings.Join(lines, "\n"))
	}
}

// TestRollupCounts pins the shared done/total formatter: natural width at 0, and
// right-justified padding to a shared column width (so "3/8" lines up under "115/166").
func TestRollupCounts(t *testing.T) {
	if got := rollupCounts(3, 8, 0); got != "3/8" {
		t.Errorf("natural width: got %q, want %q", got, "3/8")
	}
	if got := rollupCounts(3, 8, 7); got != "    3/8" {
		t.Errorf("padded: got %q, want %q", got, "    3/8")
	}
	if got := rollupCounts(115, 166, 7); got != "115/166" {
		t.Errorf("at width: got %q, want %q", got, "115/166")
	}
}

// TestModel_EpicsTabCountsAlign pins that the epics *tab* (not just the dashboard)
// pads the done/total column so epic ids start at the same display column even when
// the counts differ in width. Single-pane width so the view is the list alone.
func TestModel_EpicsTabCountsAlign(t *testing.T) {
	m := toTasks(t, loadedDashAt(t, dashAlignRepo(t), 70, 24)) // 70 → single pane (no detail beside)
	m = cmdJump(t, m, "epics")
	if m.cur().kind != entityEpics {
		t.Fatalf("setup: should be on the epics tab, got %q", m.cur().name)
	}
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	col := func(id string) int {
		for _, ln := range lines {
			if i := strings.Index(ln, id); i >= 0 {
				return visibleW(ln[:i])
			}
		}
		return -1
	}
	a, b := col("01-small"), col("02-large")
	if a < 0 || b < 0 {
		t.Fatalf("both epics should render (cols %d, %d):\n%s", a, b, strings.Join(lines, "\n"))
	}
	if a != b {
		t.Errorf("epic ids should align on the tab (counts padded), got %d vs %d:\n%s",
			a, b, strings.Join(lines, "\n"))
	}
}

// TestModel_DashboardInProgressShowsDateAligned pins the in-progress staleness date
// (#5): each row shows its last-updated date, and the slugs line up because the date
// column is padded to a shared width.
func TestModel_DashboardInProgressShowsDateAligned(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Epic("01-e.md", "---\nstatus: active\ndescription: e\n---\n# e\n")
	r.Task("in-progress", "wip-short.md",
		"---\nstatus: in-progress\nepic: 01-e\ndescription: x\nupdated_at: 2026-06-26\n---\n# s\n")
	r.Task("in-progress", "wip-longer-name.md",
		"---\nstatus: in-progress\nepic: 01-e\ndescription: y\nupdated_at: 2026-06-20\n---\n# l\n")
	m := loadedDashAt(t, r.Root, 120, 40)
	v := ansi.Strip(m.View().Content)
	if want := theme.RelativeDate("2026-06-26"); want != "" && !strings.Contains(v, want) {
		t.Errorf("an in-progress row should show its last-updated date %q:\n%s", want, v)
	}
	lines := strings.Split(v, "\n")
	col := func(s string) int {
		for _, ln := range lines {
			if i := strings.Index(ln, s); i >= 0 {
				return visibleW(ln[:i])
			}
		}
		return -1
	}
	a, b := col("wip-short"), col("wip-longer-name")
	if a < 0 || b < 0 {
		t.Fatalf("both in-progress slugs should render (cols %d, %d):\n%s", a, b, strings.Join(lines, "\n"))
	}
	if a != b {
		t.Errorf("in-progress slugs should align (date column padded), got %d vs %d:\n%s",
			a, b, strings.Join(lines, "\n"))
	}
}

// auditFindingsRepo writes an open audit with actionable findings (open +
// in-progress) of mixed urgency/component, plus a "settled" open audit (every
// finding resolved) that should read as ready-to-close.
func auditFindingsRepo(t *testing.T) string {
	t.Helper()
	r := testutil.NewRepo(t)
	r.Audit("open", "2026-06-27-arch.md",
		"---\narea: arch\ndate: 2026-06-27\n---\n# Audit\n\n"+
			"#### H1. fence event_time on writes  · **Status:** open\n"+
			"**Component:** stravapipe / write paths · **Urgency:** acute\n\n"+
			"#### M2. reconcile the two stores  · **Status:** in-progress\n"+
			"**Component:** stravapipe / pubsub · **Urgency:** soon\n\n"+
			"#### L3. drive a decision off delivery_attempt  · **Status:** in-progress\n"+
			"**Component:** dispatcher · **Urgency:** eventually\n")
	r.Audit("open", "2026-06-20-web.md",
		"---\narea: web\ndate: 2026-06-20\n---\n# Audit\n\n"+
			"#### S1. tidy imports  · **Status:** fixed\n**Component:** web · **Urgency:** soon\n")
	return r.Root
}

// TestModel_DashboardAuditFindingsWidget pins the aggregated findings inbox: the
// open/in-progress tallies, the by-urgency and by-area breakdowns, the acute
// call-out, and the "ready to close" hygiene cue for the settled open audit.
func TestModel_DashboardAuditFindingsWidget(t *testing.T) {
	m := loadedDashAt(t, auditFindingsRepo(t), 120, 40)
	v := ansi.Strip(m.View().Content)
	for _, want := range []string{
		"audit findings (1 open · 2 in progress)",
		"by urgency:", "acute", "eventually",
		"by area:", "stravapipe 2",
		"H1 fence event_time on writes", // the acute finding, called out
		"ready to close",                // the settled open audit
	} {
		if !strings.Contains(v, want) {
			t.Errorf("dashboard should show %q:\n%s", want, v)
		}
	}
}

// TestModel_DashboardAcuteFindingJumpsToAudit pins that selecting the acute finding
// row navigates to its parent audit.
func TestModel_DashboardAcuteFindingJumpsToAudit(t *testing.T) {
	m := loadedDashAt(t, auditFindingsRepo(t), 120, 40)
	pos := -1
	for i, idx := range m.dash.nav {
		if tgt := m.dash.rows[idx].target; tgt != nil && tgt.kind == entityAudits && tgt.id == "2026-06-27-arch" {
			pos = i
		}
	}
	if pos < 0 {
		t.Fatalf("the acute finding row should navigate to its audit:\n%s", ansi.Strip(m.View().Content))
	}
	m.dash.cursor = pos
	tm, cmd := m.Update(press("enter"))
	m = drain(t, tm.(Model), cmd)
	if m.onDash || m.cur().kind != entityAudits || m.selectedID() != "2026-06-27-arch" {
		t.Errorf("enter on the acute finding should jump to its audit, got onDash=%v tab=%q id=%q",
			m.onDash, m.cur().name, m.selectedID())
	}
}

// TestModel_DashboardRefreshesOnMutation pins that a reload (the `r` / fsnotify
// path) refreshes the landing dashboard too — not just the entity tabs — so the
// at-a-glance summary can't silently go stale while you're looking at it.
func TestModel_DashboardRefreshesOnMutation(t *testing.T) {
	m := loadedDash(t, 120, 40)
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "in progress (1)") {
		t.Fatalf("setup: dashboard should show the seeded in-progress task:\n%s", v)
	}
	// Move alpha out of the working set behind the dashboard's back (as the CLI or
	// another process would), then reload.
	if _, err := m.svc.Move("alpha", domain.StatusCompleted, false); err != nil {
		t.Fatal(err)
	}
	m = drainBatch(t, m, m.reloadAll())
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "nothing in progress") {
		t.Errorf("reloadAll must refresh the dashboard (it goes stale otherwise):\n%s", v)
	}
}

// TestModel_DashboardCursorStaysVisibleOnShortTerminal pins the fix for a cursor
// that could walk off-screen: on a short terminal the composed widgets overrun the
// viewport, but the selected (last) row must stay rendered — never
// navigable-but-invisible.
func TestModel_DashboardCursorStaysVisibleOnShortTerminal(t *testing.T) {
	m := loadedDashAt(t, dashOverflowRepo(t), 80, 12) // short: rows overrun the body
	last := len(m.dash.nav) - 1
	if last <= 0 {
		t.Fatalf("setup: need several navigable rows, got %d", len(m.dash.nav))
	}
	m.dash.cursor = last // park on the row that a naive clip would drop off the bottom
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "›") {
		t.Errorf("the selected row must stay visible on a short terminal:\n%s", v)
	}
}

// TestModel_DashboardOverflowRowShown pins the "+N more" overflow row for a widget
// past dashListCap (untested before — every fixture stayed under the cap).
func TestModel_DashboardOverflowRowShown(t *testing.T) {
	m := loadedDashAt(t, dashOverflowRepo(t), 120, 40)
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "+2 more") {
		t.Errorf("an over-cap in-progress widget (8 tasks, cap 6) should show '+2 more':\n%s", v)
	}
}

// TestModel_DashboardLoadErrorIsDurable pins that a failed load is shown
// persistently (not a one-shot flash that leaves the screen stuck on "loading…"),
// and that a later successful load clears it.
func TestModel_DashboardLoadErrorIsDurable(t *testing.T) {
	m := newModel(t)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	// The initial dashboard load fails before any summary is shown.
	tm, _ = m.Update(dashLoadedMsg{err: domain.ErrNotFound})
	m = tm.(Model)
	if m.dash.loadErr == nil {
		t.Fatal("a failed dashboard load should be stored, not just flashed")
	}
	v := ansi.Strip(m.View().Content)
	if strings.Contains(v, "loading…") {
		t.Errorf("a failed load must not leave the dashboard stuck on loading:\n%s", v)
	}
	if !strings.Contains(v, "error:") {
		t.Errorf("the dashboard load error should be shown in the body:\n%s", v)
	}
	// Durable, not a flash: a keypress must not wipe it.
	tm, _ = m.Update(press("j"))
	m = tm.(Model)
	if !strings.Contains(ansi.Strip(m.View().Content), "error:") {
		t.Error("the dashboard error must survive a keypress (durable, unlike a flash)")
	}
	// Recovery: a successful load clears the error and renders the widgets.
	s, err := m.svc.Summary()
	if err != nil {
		t.Fatal(err)
	}
	tm, _ = m.Update(dashLoadedMsg{summary: s})
	m = tm.(Model)
	if m.dash.loadErr != nil || !m.dash.loaded {
		t.Errorf("a successful load should clear the error and load: err=%v loaded=%v", m.dash.loadErr, m.dash.loaded)
	}
	if strings.Contains(ansi.Strip(m.View().Content), "error:") {
		t.Error("the error pane should be gone after a successful load")
	}
}

// TestModel_DashboardRefreshFailureKeepsRows pins the loaded-then-failed-refresh
// path: the last good rows survive and the footer flags the failure (mirroring the
// per-tab "reload failed" behaviour).
func TestModel_DashboardRefreshFailureKeepsRows(t *testing.T) {
	m := loadedDash(t, 120, 40)
	if !m.dash.loaded {
		t.Fatal("setup: dashboard should be loaded")
	}
	tm, _ := m.Update(dashLoadedMsg{err: domain.ErrNotFound})
	m = tm.(Model)
	if !m.dash.loaded {
		t.Error("a failed refresh must keep the last good load (loaded stays true)")
	}
	v := ansi.Strip(m.View().Content)
	if !strings.Contains(v, "in progress") {
		t.Errorf("the last good rows should survive a failed refresh:\n%s", v)
	}
	if !strings.Contains(v, "refresh failed") {
		t.Errorf("the footer should flag the failed dashboard refresh:\n%s", v)
	}
}

// TestDashboardMoveWrapsAndEmptyNavIsSafe pins cursor wrapping at the ends and the
// zero-navigable-rows edge case (an all-clear dashboard): move is a safe no-op and
// nothing is selectable.
func TestDashboardMoveWrapsAndEmptyNavIsSafe(t *testing.T) {
	// Empty nav: a repo with a single completed task — nothing in progress, no
	// epics, no audits, health all-clear → no navigable rows.
	r := testutil.NewRepo(t)
	r.Task("completed", "done.md", "---\nstatus: completed\ndescription: x\n---\n# done\n")
	empty := loadedDashAt(t, r.Root, 120, 40)
	if n := len(empty.dash.nav); n != 0 {
		t.Fatalf("setup: expected zero navigable rows, got %d", n)
	}
	empty.dash.move(1) // must not panic / corrupt state
	if _, ok := empty.dash.selectedTarget(); ok {
		t.Error("an empty dashboard must report no selection")
	}

	// Wrapping: k from the first row lands on the last; j past the last wraps to 0.
	m := loadedDash(t, 120, 40)
	n := len(m.dash.nav)
	if n < 2 {
		t.Fatalf("setup: need >=2 navigable rows, got %d", n)
	}
	m.dash.cursor = 0
	m.dash.move(-1)
	if m.dash.cursor != n-1 {
		t.Errorf("k at the first row should wrap to the last (%d), got %d", n-1, m.dash.cursor)
	}
	m.dash.move(1)
	if m.dash.cursor != 0 {
		t.Errorf("j past the last row should wrap to 0, got %d", m.dash.cursor)
	}
}

// TestDashboardScrollTo unit-tests the cursor-following window: it always returns
// maxH lines and always contains the focused row.
func TestDashboardScrollTo(t *testing.T) {
	lines := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"} // 10 lines
	for _, tc := range []struct {
		focus, h            int
		wantFirst, wantLast string
	}{
		{focus: 0, h: 4, wantFirst: "0", wantLast: "3"},  // top
		{focus: 9, h: 4, wantFirst: "6", wantLast: "9"},  // bottom clamps
		{focus: 5, h: 4, wantFirst: "3", wantLast: "6"},  // centered-ish
		{focus: -1, h: 4, wantFirst: "0", wantLast: "3"}, // no cursor → top
	} {
		got := scrollTo(lines, tc.focus, tc.h)
		if len(got) != tc.h {
			t.Errorf("focus=%d h=%d: window len %d, want %d", tc.focus, tc.h, len(got), tc.h)
		}
		if got[0] != tc.wantFirst || got[len(got)-1] != tc.wantLast {
			t.Errorf("focus=%d h=%d: window [%s..%s], want [%s..%s]",
				tc.focus, tc.h, got[0], got[len(got)-1], tc.wantFirst, tc.wantLast)
		}
		if tc.focus >= 0 {
			var found bool
			for _, l := range got {
				if l == lines[tc.focus] {
					found = true
				}
			}
			if !found {
				t.Errorf("focus=%d h=%d: cursor line %q not in window %v", tc.focus, tc.h, lines[tc.focus], got)
			}
		}
	}
}
