package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/store"
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
	for _, want := range []string{"dashboard", "in progress", "epics", "health"} {
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
	for _, word := range []string{"dashboard", "d"} {
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
	for _, w := range []string{"dashboard", "tasks", "revisit"} {
		if !strings.Contains(all, w) {
			t.Errorf("empty `:` hint should list %q, got %q", w, all)
		}
	}
	m = typeRunes(t, m, "rev")
	if h := m.commandHint(); h != "revisit" {
		t.Errorf(":rev should narrow the hint to revisit, got %q", h)
	}
}
