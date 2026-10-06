package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func completeLocalPathAction(t *testing.T, m Model, initial tea.Cmd) (Model, tea.Cmd) {
	t.Helper()
	if initial == nil {
		t.Fatal("local path action did not request an initial lookup")
	}
	tm, confirm := m.Update(initial())
	if confirm == nil {
		t.Fatal("local path action did not confirm the selected stable ID")
	}
	tm, effect := tm.(Model).Update(confirm())
	return tm.(Model), effect
}

func TestTaskLocalActionDoesNotInferPathFromSourceLocation(t *testing.T) {
	for _, source := range []core.RecordSource{
		{Location: "urn:task:must-not-open"},
		{Location: "/planning/tasks/must-not-open.md"},
		{Location: "/planning/tasks/must-not-open.md", LocationIsPath: true},
	} {
		t.Run(fmt.Sprintf("%s/path-hint=%t", source.Location, source.LocationIsPath), func(t *testing.T) {
			repo := testutil.NewRepo(t)
			taskID := testutil.TaskID("pathless-tui-task")
			repo.Task("ready-to-start", "pathless-tui-task.md", "---\nstatus: ready-to-start\n---\n# Pathless\n")
			fs := testutil.Must(store.NewFS(repo.Root, core.UnrestrictedMutations()))
			source.ID = taskID
			remote := &countingGraphSource{sourceSet: fs.SourceSetID(), records: []core.LoadedRecord[domain.Task]{{
				Value:  domain.Task{ID: taskID, Slug: "pathless-tui-task", Status: domain.StatusReadyToStart},
				Source: source,
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
				m, effect := completeLocalPathAction(t, tm.(Model), resolve)
				if effect != nil || !m.flashErr || !strings.Contains(m.flash, "local path unavailable") {
					t.Fatalf("%s inferred local path: effect=%v flash=%q", action, effect != nil, m.flash)
				}
			}
		})
	}
}

func TestLocalPathActionFollowsRenameByStableID(t *testing.T) {
	for _, action := range []string{"E", "Y"} {
		t.Run(action, func(t *testing.T) {
			m := loaded(t, 100, 30)
			id := m.selectedKey()
			oldPath, err := m.svc.TaskPath(id)
			if err != nil {
				t.Fatal(err)
			}
			tm, lookup := m.Update(press(action))
			m = tm.(Model)
			if lookup == nil {
				t.Fatal("action did not request a local path")
			}
			first := lookup() // the initial lookup now carries the old pathname
			newPath := filepath.Join(filepath.Dir(oldPath), id+"-renamed.md")
			if err := os.Rename(oldPath, newPath); err != nil {
				t.Fatal(err)
			}
			tm, confirm := m.Update(first)
			m = tm.(Model)
			if confirm == nil {
				t.Fatal("accepted initial lookup did not re-resolve the stable ID")
			}
			second, ok := confirm().(localPathResultMsg)
			if !ok || !second.rechecked || second.path != newPath || second.err != nil {
				t.Fatalf("confirmation did not follow rename: %+v", second)
			}
			tm, effect := m.Update(second)
			m = tm.(Model)
			if effect == nil || m.flashErr {
				t.Fatalf("renamed action effect=%v flash=%q", effect != nil, m.flash)
			}
			if action == "Y" && m.flash != "copied path: "+newPath {
				t.Fatalf("copied stale path: %q", m.flash)
			}
		})
	}
}

func TestLocalPathActionReportsMissingStableIDAfterRename(t *testing.T) {
	m := loaded(t, 100, 30)
	id := m.selectedKey()
	oldPath, err := m.svc.TaskPath(id)
	if err != nil {
		t.Fatal(err)
	}
	tm, lookup := m.Update(press("Y"))
	m = tm.(Model)
	if lookup == nil {
		t.Fatal("Y did not request a local path")
	}
	first := lookup()
	otherPath := filepath.Join(filepath.Dir(oldPath), testutil.TaskID("different-task")+"-renamed.md")
	if err := os.Rename(oldPath, otherPath); err != nil {
		t.Fatal(err)
	}
	tm, confirm := m.Update(first)
	m = tm.(Model)
	if confirm == nil {
		t.Fatal("accepted initial lookup did not re-resolve the stable ID")
	}
	tm, effect := m.Update(confirm())
	m = tm.(Model)
	if effect != nil || !m.flashErr || !strings.Contains(m.flash, "local path unavailable") {
		t.Fatalf("missing ID action effect=%v flash=%q", effect != nil, m.flash)
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
