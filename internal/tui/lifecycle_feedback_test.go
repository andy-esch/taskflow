package tui

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// Inject only the post-commit failure: projection, authorization and persistence
// still run through the real adapter, and Service supplies the recovery guidance.
type lifecycleCleanupStore struct {
	*store.FS
	cause error
	calls int
}

func (s *lifecycleCleanupStore) MutateTaskLifecycle(now time.Time, dryRun bool, planner core.TaskLifecyclePlanner) (core.TaskLifecycleMutationResult, error) {
	s.calls++
	result, err := s.FS.MutateTaskLifecycle(now, dryRun, planner)
	if err == nil && result.Committed {
		err = s.cause
	}
	return result, err
}

func TestTUITaskReopenSurfacesThreadOnlyRecovery(t *testing.T) {
	for _, cleanupFailure := range []bool{false, true} {
		t.Run(fmt.Sprintf("cleanup-failure=%t", cleanupFailure), func(t *testing.T) {
			r := testutil.NewRepo(t)
			taskID, threadID := testutil.TaskID("sole"), testutil.TaskID("sole-thread")
			r.Task("completed", "sole.md", fmt.Sprintf(
				"---\nid: %s\nstatus: completed\ndescription: sole member\ntags: [test]\n---\n# Sole\n", taskID))
			threadRel := filepath.Join(domain.ThreadsDir, threadID+"-delivery.md")
			r.File(threadRel, fmt.Sprintf(
				"---\nschema: 1\nid: %s\nstatus: completed\ndescription: delivery\n"+
					"goal: ship it\ncreated: \"2026-10-10\"\ntasks: [%s]\n---\n# Delivery\n", threadID, taskID))
			threadPath := filepath.Join(r.Root, threadRel)
			threadBefore, err := os.ReadFile(threadPath)
			if err != nil {
				t.Fatal(err)
			}
			adapter := &lifecycleCleanupStore{FS: testutil.Must(store.NewFS(r.Root, core.UnrestrictedMutations()))}
			if cleanupFailure {
				adapter.cause = fmt.Errorf("guard release failed: %w", domain.ErrConflict)
			}
			svc := core.MustNewService(adapter, core.WithRetry(3, func(int) {
				t.Fatal("a committed transition must not be retried")
			}))
			before, _, err := svc.ShowThread(threadID)
			if err != nil || before.Inconsistent || before.Rollup.Done != 1 {
				t.Fatalf("setup is not a sound completed Thread: %+v, %v", before, err)
			}
			m := New(svc)
			tm, _ := m.Update(tea.WindowSizeMsg{Width: 300, Height: 40})
			m = tm.(Model)
			m = drainNested(t, m, m.Init())
			m = toTasks(t, m)
			taskTab := m.cur()
			taskTab.statusView = "all" // the completed member must be selectable
			m = drainNested(t, m, taskTab.reload(svc, entityRef{}))
			m = openThreads(t, m) // a loaded projection must refresh after the move too
			m = drainNested(t, m, m.switchTab(indexOfKind(m.tabs, entityTasks)))
			if m.selectedRef().key != taskID {
				t.Fatalf("setup selected %q, want sole member %q", m.selectedRef().key, taskID)
			}
			threadGen, taskGen := threadTab(m).loadGen, m.tabs[indexOfKind(m.tabs, entityTasks)].loadGen

			tm, _ = m.Update(press("m"))
			m = cursorTo(t, tm.(Model), "ready")
			tm, move := m.Update(press("enter"))
			m = tm.(Model)
			if move == nil {
				t.Fatal("ready action did not launch its transition")
			}
			msg := move()
			moved, ok := mutationResult(msg).(movedMsg)
			if !ok {
				if failed, failedOK := mutationResult(msg).(actionErrMsg); failedOK {
					t.Fatalf("reopen failed: %v", failed.err)
				}
				t.Fatalf("reopen returned %T: %+v", msg, msg)
			}
			receipt := moved.lifecycle
			if receipt == nil || !receipt.Committed || !receipt.Changed || len(receipt.Impacts) != 0 ||
				len(receipt.ThreadImpacts) != 1 || !receipt.ThreadImpacts[0].NewlyInconsistent() || receipt.Remedy == "" {
				t.Fatalf("expected a committed Thread-only recovery receipt: %+v", receipt)
			}
			if (moved.warning != nil) != cleanupFailure || (cleanupFailure && !errors.Is(moved.warning, adapter.cause)) || adapter.calls != 1 {
				t.Fatalf("cleanup outcome lost or retried: warning=%v calls=%d", moved.warning, adapter.calls)
			}
			tm, reload := m.Update(msg)
			m = tm.(Model)
			if !strings.Contains(m.flash, receipt.Remedy) || strings.Contains(m.flash, "downstream task(s)") ||
				!strings.Contains(m.flash, "1 Thread projection(s) changed") || !strings.Contains(m.flash, "WARNING: 1 Thread(s) newly inconsistent") {
				t.Fatalf("Thread-only recovery was hidden: flash=%q remedy=%q", m.flash, receipt.Remedy)
			}
			if m.flashErr != cleanupFailure || (cleanupFailure && !strings.Contains(m.flash, moved.warning.Error())) {
				t.Fatalf("committed cleanup warning lost: flash=%q err=%v", m.flash, m.flashErr)
			}
			if reload == nil || threadTab(m).loadGen <= threadGen || m.tabs[indexOfKind(m.tabs, entityTasks)].loadGen <= taskGen {
				t.Fatal("committed transition did not schedule task and Thread reloads")
			}
			// The existing one-line footer clips at terminal width. A wide screen must
			// render the exact core guidance, not just keep it in the model.
			m.width = len(m.flash) + 10
			if !strings.Contains(ansi.Strip(m.footer()), receipt.Remedy) {
				t.Fatalf("footer discarded recovery guidance: %q", m.footer())
			}
			flash := m.flash
			m = drainNested(t, m, reload)
			reloadedThread, ok := threadTab(m).list.SelectedItem().(threadItem)
			if m.flash != flash || m.flashErr != cleanupFailure || !ok || !reloadedThread.view.Inconsistent {
				t.Fatalf("reload lost recovery or retained the old projection: flash=%q err=%v view=%+v", m.flash, m.flashErr, reloadedThread.view)
			}
			record, err := svc.ShowTask(taskID)
			if err != nil || record.Value.Task.Status != domain.StatusReadyToStart {
				t.Fatalf("reopen was not durable: task=%+v err=%v", record.Value.Task, err)
			}
			threadAfter, err := os.ReadFile(threadPath)
			if err != nil || !bytes.Equal(threadBefore, threadAfter) {
				t.Fatalf("task reopen unexpectedly rewrote the Thread: %v", err)
			}
		})
	}
}

func TestTUILifecycleFeedbackUsesIndependentReceiptFields(t *testing.T) {
	newlyInconsistent := core.ThreadProjectionImpact{After: core.ThreadView{Inconsistent: true}}
	alreadyInconsistent := core.ThreadProjectionImpact{
		Before: core.ThreadView{Inconsistent: true}, After: core.ThreadView{Inconsistent: true},
	}
	restored := core.ThreadProjectionImpact{Before: core.ThreadView{Inconsistent: true}}
	for _, tc := range []struct {
		name    string
		receipt *core.TaskLifecycleReceipt
		suffix  string
	}{
		{name: "non-task move"},
		{name: "empty receipt", receipt: &core.TaskLifecycleReceipt{}},
		{
			name: "standalone remedy", receipt: &core.TaskLifecycleReceipt{Remedy: "core-only recovery instruction"},
			suffix: "; core-only recovery instruction",
		},
		{
			name: "task-only", receipt: &core.TaskLifecycleReceipt{
				Impacts: []core.TaskGraphStateImpact{{}}, Remedy: "exact task recovery",
			},
			suffix: "; 1 downstream task(s) changed derived state; exact task recovery",
		},
		{
			name: "healthy Thread-only", receipt: &core.TaskLifecycleReceipt{ThreadImpacts: []core.ThreadProjectionImpact{{}}},
			suffix: "; 1 Thread projection(s) changed",
		},
		{
			name: "newly inconsistent Thread-only", receipt: &core.TaskLifecycleReceipt{
				ThreadImpacts: []core.ThreadProjectionImpact{newlyInconsistent}, Remedy: "exact Thread recovery",
			},
			suffix: "; 1 Thread projection(s) changed; WARNING: 1 Thread(s) newly inconsistent; exact Thread recovery",
		},
		{
			name: "already inconsistent and restored", receipt: &core.TaskLifecycleReceipt{
				ThreadImpacts: []core.ThreadProjectionImpact{alreadyInconsistent, restored},
			},
			suffix: "; 2 Thread projection(s) changed",
		},
		{
			name: "combined", receipt: &core.TaskLifecycleReceipt{
				Impacts:       []core.TaskGraphStateImpact{{}, {}},
				ThreadImpacts: []core.ThreadProjectionImpact{newlyInconsistent, alreadyInconsistent, newlyInconsistent},
				Remedy:        "exact combined recovery",
			},
			suffix: "; 2 downstream task(s) changed derived state; 3 Thread projection(s) changed; WARNING: 2 Thread(s) newly inconsistent; exact combined recovery",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := loaded(t, 120, 40)
			m.flashErr = true // successful feedback must clear a previous error
			gen := m.cur().loadGen
			tm, reload := m.Update(movedMsg{
				ref: testEntityRef("target"), to: string(domain.StatusDeferred), revisit: "2026-11-01", lifecycle: tc.receipt,
			})
			m = tm.(Model)
			want := "moved target → deferred (revisit 2026-11-01)" + tc.suffix
			if m.flash != want || m.flashErr || reload == nil || m.cur().loadGen <= gen {
				t.Fatalf("feedback=%q err=%v reload=%v; want %q", m.flash, m.flashErr, reload != nil, want)
			}
		})
	}
}
