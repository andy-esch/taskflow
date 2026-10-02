package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestTaskLocalActionDoesNotInferPathFromSemanticTask(t *testing.T) {
	repo := testutil.NewRepo(t)
	taskID := testutil.TaskID("pathless-tui-task")
	repo.Task("ready-to-start", "pathless-tui-task.md", "---\nstatus: ready-to-start\n---\n# Pathless\n")
	fs := store.NewFS(repo.Root)
	remote := &countingGraphSource{sourceSet: fs.SourceSetID(), tasks: []domain.Task{{
		ID: taskID, Slug: "pathless-tui-task", Status: domain.StatusReadyToStart,
		Path: "urn:task:must-not-open",
	}}}
	svc := core.MustNewService(fs, core.WithTaskGraphSource(remote))
	m := New(svc)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = tm.(Model)
	m = toTasks(t, m)
	if m.selectedRef().empty() || m.selectedPath() != "" {
		t.Fatalf("portable task selection gained local path: ref=%+v path=%q", m.selectedRef(), m.selectedPath())
	}
	for _, action := range []string{"E", "Y"} {
		tm, resolve := m.Update(press(action))
		if resolve == nil {
			t.Fatalf("%s did not request the optional path capability", action)
		}
		tm, effect := tm.(Model).Update(resolve())
		m = tm.(Model)
		if effect != nil || !m.flashErr || !strings.Contains(m.flash, "local path unavailable") {
			t.Fatalf("%s inferred local path: effect=%v flash=%q", action, effect != nil, m.flash)
		}
	}
}

func TestDelayedLocalPathActionCannotRetargetSelectionOrReload(t *testing.T) {
	m := loaded(t, 100, 30)
	first := m.selectedKey()
	tm, resolve := m.Update(press("E"))
	m = tm.(Model)
	if resolve == nil {
		t.Fatal("E did not request a path")
	}
	result := resolve()
	items := m.cur().list.Items()
	if len(items) < 2 || !m.cur().selectByKey(items[1].(entityItem).ref().key) || m.selectedKey() == first {
		t.Fatal("fixture did not move to another canonical row")
	}
	tm, effect := m.Update(result)
	if effect != nil || tm.(Model).flash != "" {
		t.Fatal("delayed editor path opened on another selection")
	}

	m = loaded(t, 100, 30)
	tm, resolve = m.Update(press("Y"))
	m = tm.(Model)
	result = resolve()
	m.cur().loadGen++
	m.cur().coherentGen = m.cur().loadGen // a newer coherent list can retain this selection
	tm, effect = m.Update(result)
	if effect != nil || tm.(Model).flash != "" {
		t.Fatal("delayed copy path survived a newer list generation")
	}
}
