package tui

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

type selectedTaskAction struct {
	name   string
	status domain.Status
	cmd    func(*core.Service, entityRef) tea.Cmd
	want   string
}

func selectedTaskActions() []selectedTaskAction {
	return []selectedTaskAction{
		{"set", domain.StatusReadyToStart, func(s *core.Service, ref entityRef) tea.Cmd {
			return setFieldCmd(s, ref, "priority", "high")
		}, "priority: high"},
		{"unset", domain.StatusDeferred, func(s *core.Service, ref entityRef) tea.Cmd {
			return unsetFieldCmd(s, ref, "revisit_at")
		}, ""},
		{"move", domain.StatusReadyToStart, func(s *core.Service, ref entityRef) tea.Cmd {
			return moveTask(s, ref, transition{verb: "start", to: string(domain.StatusInProgress)})
		}, "status: in-progress"},
		{"defer", domain.StatusReadyToStart, func(s *core.Service, ref entityRef) tea.Cmd {
			return deferTaskCmd(s, ref, "2099-01-02")
		}, "revisit_at: \"2099-01-02\""},
	}
}

func selectedTaskActionSeed(slug string, status domain.Status) string {
	revisit := ""
	if status == domain.StatusDeferred {
		revisit = "revisit_at: \"2099-01-01\"\n"
	}
	return fmt.Sprintf("---\nid: %s\nstatus: %s\ndescription: selected action fixture\ntags: [identity]\npriority: medium\n%s---\n# %s\n", testutil.TaskID(slug), status, revisit, slug)
}

func selectedTaskActionModel(t *testing.T, status domain.Status) (Model, string, string) {
	t.Helper()
	r := testutil.NewRepo(t)
	for _, slug := range []string{"chosen", "sibling"} {
		r.File(filepath.Join(domain.TasksDir, testutil.TaskID(slug)+"-"+slug+".md"), selectedTaskActionSeed(slug, status))
	}
	m := loadedAt(t, r.Root, 120, 40)
	if !m.cur().selectByKey(testutil.TaskID("chosen")) {
		t.Fatal("selected task is missing")
	}
	return m, r.Root, filepath.Join(r.Root, domain.TasksDir, testutil.TaskID("chosen")+"-chosen.md")
}

// Construct the real tea.Cmd before an external edit, without a watcher refresh.
// The row key is already selected identity, never a new human-reference query.
func TestQueuedSelectedTaskActionsRetainIdentityBeforeExecution(t *testing.T) {
	for _, action := range selectedTaskActions() {
		for _, mode := range []string{"unchanged", "renamed", "missing", "impostor", "case-folded impostor", "duplicate"} {
			t.Run(action.name+"/"+mode, func(t *testing.T) {
				m, root, path := selectedTaskActionModel(t, action.status)
				ref := m.selectedRef()
				beforeConstruction := testutil.SnapshotTree(t, root)
				cmd := action.cmd(m.svc, ref)
				if !maps.Equal(beforeConstruction, testutil.SnapshotTree(t, root)) {
					t.Fatal("command construction changed persistent files")
				}
				switch mode {
				case "renamed":
					newPath := filepath.Join(filepath.Dir(path), ref.key+"-renamed.md")
					if err := os.Rename(path, newPath); err != nil {
						t.Fatal(err)
					}
					path = newPath
					testutil.Write(t, filepath.Join(filepath.Dir(path), testutil.TaskID("replacement")+"-chosen.md"), selectedTaskActionSeed("replacement", action.status))
				case "missing", "impostor", "case-folded impostor":
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					if mode != "missing" {
						label := ref.key
						if mode == "case-folded impostor" {
							label = strings.ToUpper(label)
						}
						testutil.Write(t, filepath.Join(filepath.Dir(path), testutil.TaskID("impostor")+"-"+label+".md"), selectedTaskActionSeed("impostor", action.status))
					}
				case "duplicate":
					testutil.Write(t, filepath.Join(filepath.Dir(path), ref.key+"-copy.md"), selectedTaskActionSeed("chosen", action.status))
				}
				before := testutil.SnapshotTree(t, root)
				result := cmd()
				after := testutil.SnapshotTree(t, root)
				if mode != "unchanged" && mode != "renamed" {
					failure, ok := result.(actionErrMsg)
					if !ok {
						t.Fatalf("selected action reported success for a vanished/colliding ID: %T %+v", result, result)
					}
					if mode != "duplicate" && !errors.Is(failure.err, domain.ErrNotFound) {
						t.Fatalf("missing selected ID should be not found: %v", failure.err)
					}
					if mode == "duplicate" && !errors.Is(failure.err, domain.ErrAmbiguous) && !errors.Is(failure.err, domain.ErrValidation) {
						t.Fatalf("collision was not attributed usefully: %v", failure.err)
					}
					if !maps.Equal(before, after) {
						t.Fatal("refused selected action changed another record")
					}
					return
				}
				switch result := result.(type) {
				case editedMsg:
					if result.ref != ref {
						t.Fatalf("edit result lost its selected identity: %+v", result)
					}
				case movedMsg:
					if result.ref != ref || result.lifecycle == nil || !result.lifecycle.Committed || result.lifecycle.Task.ID != ref.key {
						t.Fatalf("lifecycle receipt lost selected identity: %+v", result)
					}
				default:
					t.Fatalf("valid selected action failed: %T %+v", result, result)
				}
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if action.name == "unset" {
					if strings.Contains(string(content), "revisit_at:") {
						t.Fatal("selected revisit date was not removed")
					}
				} else if !strings.Contains(string(content), action.want) {
					t.Fatalf("selected action did not apply requested change:\n%s", content)
				}
				key := filepath.Join(domain.TasksDir, filepath.Base(path))
				if before[key] == after[key] {
					t.Fatal("successful action did not change the selected record")
				}
				delete(before, key)
				delete(after, key)
				if !maps.Equal(before, after) {
					t.Fatal("selected action changed a sibling")
				}
			})
		}
	}
}

func queueSelectedTaskActionThroughUI(t *testing.T, m Model, action selectedTaskAction) (Model, tea.Cmd) {
	t.Helper()
	if action.name == "set" || action.name == "unset" {
		tm, _ := m.Update(press("e"))
		m = tm.(Model)
		field := "priority"
		if action.name == "unset" {
			field = "revisit_at"
		}
		m = editCursorTo(t, m, field)
		tm, _ = m.Update(press("enter"))
		m = tm.(Model)
		if action.name == "set" {
			m = enumCursorTo(t, m, "high")
		} else {
			m.edit.input.SetValue("")
		}
		tm, cmd := m.Update(press("enter"))
		return tm.(Model), cmd
	}
	word := "start"
	if action.name == "defer" {
		word = "defer"
	}
	tm, _ := m.Update(press(":"))
	m = tm.(Model)
	for _, r := range word {
		tm, _ = m.Update(press(string(r)))
		m = tm.(Model)
	}
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	if action.name == "defer" {
		if !m.action.revisit {
			t.Fatal(":defer did not open the date prompt")
		}
		m.action.dateInput.SetValue("2099-01-02")
		tm, cmd = m.Update(press("enter"))
		m = tm.(Model)
	}
	return m, cmd
}

func TestSelectedTaskUIFormsRetainIdentityAndResultEnvelopes(t *testing.T) {
	for _, action := range selectedTaskActions() {
		t.Run(action.name, func(t *testing.T) {
			m, root, path := selectedTaskActionModel(t, action.status)
			m.sessionScope, m.sessionGen = true, 4
			ref := m.selectedRef()
			m, cmd := queueSelectedTaskActionThroughUI(t, m, action)
			if cmd == nil {
				t.Fatal("form did not submit a command")
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			testutil.Write(t, filepath.Join(filepath.Dir(path), testutil.TaskID("impostor")+"-"+ref.key+".md"), selectedTaskActionSeed("impostor", action.status))
			before := testutil.SnapshotTree(t, root)
			result := cmd()
			session, ok := result.(sessionMsg)
			if !ok || session.gen != 4 {
				t.Fatalf("form lost its workspace envelope: %T %+v", result, result)
			}
			scoped, ok := session.msg.(mutationResultMsg)
			if !ok || scoped.ref != ref || scoped.listGen != m.cur().loadGen {
				t.Fatalf("form lost its selected row/generation envelope: %T %+v", session.msg, session.msg)
			}
			failure, ok := scoped.result.(actionErrMsg)
			if !ok || !errors.Is(failure.err, domain.ErrNotFound) || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatalf("form action retargeted an impostor: %T %+v", scoped.result, scoped.result)
			}
			tm, _ := m.Update(result)
			m = tm.(Model)
			if !m.flashErr || !strings.Contains(m.flash, ref.key) {
				t.Fatalf("refusal was not surfaced for the selected identity: %q", m.flash)
			}
		})
	}
}

func TestQueuedSelectedTaskActionsRespectResultSelectionAndGeneration(t *testing.T) {
	for _, action := range selectedTaskActions() {
		for _, change := range []string{"selection", "reload"} {
			t.Run(action.name+"/"+change, func(t *testing.T) {
				m, _, _ := selectedTaskActionModel(t, action.status)
				ref := m.selectedRef()
				cmd := scopeMutation(entityTasks, ref, m.cur().loadGen, action.cmd(m.svc, ref))
				if change == "selection" {
					if !m.cur().selectByKey(testutil.TaskID("sibling")) {
						t.Fatal("sibling is missing")
					}
				} else {
					m.cur().loadGen++
					m.cur().coherentGen = m.cur().loadGen
				}
				selected := m.selectedRef()
				m.flash = "newer state"
				m.edit.open(selected, domain.Task{Status: action.status, Priority: "medium"})
				m.edit.err = "keep this editor"
				result := cmd()
				if _, ok := result.(mutationResultMsg); !ok {
					t.Fatalf("mutation lost its source-generation envelope: %T", result)
				}
				tm, reload := m.Update(result)
				m = tm.(Model)
				// reloadAll temporarily quarantines selection until its rows have
				// passed identity validation; its restore intent must stay current.
				if reload == nil || m.cur().restore != selected || m.flash != "newer state" || m.edit.err != "keep this editor" {
					t.Fatalf("stale committed result changed current UI state: restore=%+v flash=%q editor=%q reload=%v", m.cur().restore, m.flash, m.edit.err, reload != nil)
				}
			})
		}
	}
}

func TestQueuedSelectedTaskActionsRetainCapturedServiceAcrossSessionSwitch(t *testing.T) {
	for _, action := range selectedTaskActions() {
		t.Run(action.name, func(t *testing.T) {
			old, oldRoot, _ := selectedTaskActionModel(t, action.status)
			ref := old.selectedRef()
			cmd := scopeSession(1, scopeMutation(entityTasks, ref, old.cur().loadGen, action.cmd(old.svc, ref)))
			oldBefore := testutil.SnapshotTree(t, oldRoot)
			// Both spaces contain the same task IDs, so only the captured service
			// and session envelope distinguish them. This is not cancellation:
			// an already-issued action may finish in its original space.
			current, currentRoot, _ := selectedTaskActionModel(t, action.status)
			current.sessionScope, current.sessionGen = true, 2
			current.flash = "current workspace"
			before := testutil.SnapshotTree(t, currentRoot)
			result := cmd()
			tm, followup := current.Update(result)
			current = tm.(Model)
			if followup != nil || current.flash != "current workspace" || current.selectedRef() != ref || !maps.Equal(before, testutil.SnapshotTree(t, currentRoot)) {
				t.Fatal("old-session action or result retargeted the new workspace")
			}
			if maps.Equal(oldBefore, testutil.SnapshotTree(t, oldRoot)) {
				t.Fatal("queued command did not retain its originally selected service")
			}
		})
	}
}

func TestSelectedTaskActionsRespectReadOnlyPersistence(t *testing.T) {
	for _, action := range selectedTaskActions() {
		t.Run(action.name, func(t *testing.T) {
			m, root, _ := selectedTaskActionModel(t, action.status)
			svc := core.MustNewService(testutil.Must(store.NewFS(root, core.ReadOnlyMutations())))
			before := testutil.SnapshotTree(t, root)
			result := action.cmd(svc, m.selectedRef())()
			failed, ok := result.(actionErrMsg)
			if !ok || !errors.Is(failed.err, core.ErrReadOnlyPersistence) || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatalf("selected API bypassed authorization: result=%T %+v", result, result)
			}
		})
	}
}

type selectedCommittedLifecycleStore struct {
	*store.FS
	calls int
	cause error
}

func (s *selectedCommittedLifecycleStore) MutateTaskLifecycle(_ time.Time, _ bool, planner core.TaskLifecyclePlanner) (core.TaskLifecycleMutationResult, error) {
	s.calls++
	graph, err := core.LoadTaskGraph(s.FS)
	if err != nil {
		return core.TaskLifecycleMutationResult{}, err
	}
	plan, err := planner(graph)
	if err != nil {
		return core.TaskLifecycleMutationResult{}, err
	}
	task, _ := graph.Task(plan.TaskID)
	// This port double models a committed receipt plus cleanup conflict, not a
	// real filesystem write. The command must preserve it and never retry.
	return core.TaskLifecycleMutationResult{Plan: plan, Task: task, Committed: true, Changed: true}, s.cause
}

func TestSelectedTaskCommandsPreserveCommittedFailureReceipt(t *testing.T) {
	for _, action := range selectedTaskActions()[2:] {
		t.Run(action.name, func(t *testing.T) {
			m, root, _ := selectedTaskActionModel(t, action.status)
			st := &selectedCommittedLifecycleStore{FS: testutil.Must(store.NewFS(root, core.UnrestrictedMutations())), cause: fmt.Errorf("guard cleanup: %w", domain.ErrConflict)}
			result := action.cmd(core.MustNewService(st, core.WithRetry(2, func(int) { t.Error("committed action was retried") })), m.selectedRef())()
			moved, ok := result.(movedMsg)
			if !ok || st.calls != 1 || moved.lifecycle == nil || !moved.lifecycle.Committed || !errors.Is(moved.warning, st.cause) {
				t.Fatalf("committed failure was lost: result=%T %+v calls=%d", result, result, st.calls)
			}
			tm, reload := m.Update(moved)
			m = tm.(Model)
			if reload == nil || !m.flashErr || !strings.Contains(m.flash, "WARNING: committed") {
				t.Fatalf("committed warning did not reload: flash=%q reload=%v", m.flash, reload != nil)
			}
		})
	}
}
