package store

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func writeUnreadableGuardThread(t *testing.T, root, seed string) string {
	t.Helper()
	path := filepath.Join(root, domain.ThreadsDir, testutil.TaskID(seed)+"-"+seed+".md")
	testutil.Write(t, path, "---\nid: [unterminated\n---\n")
	return path
}

func TestGuardedMutationsRejectUnreadableThreadDocuments(t *testing.T) {
	t.Run("Thread creation", func(t *testing.T) {
		root := t.TempDir()
		writeUnreadableGuardThread(t, root, "unreadable-create-guard")
		threadID := testutil.TaskID("refused-create")
		svc := core.NewService(NewFS(root),
			core.WithIDGen(func() string { return threadID }),
			core.WithClock(func() time.Time { return threadCreationNow }),
			core.WithRetry(0, func(int) {}),
		)

		receipt, err := svc.NewThread(core.NewThreadParams{
			Title: "Refused create", Description: "Refuse incomplete authority", Goal: "Fail closed",
		})
		if !errors.Is(err, domain.ErrValidation) || receipt.Committed {
			t.Fatalf("receipt=%+v err=%v", receipt, err)
		}
		path := filepath.Join(root, domain.ThreadsDir, threadID+"-refused-create.md")
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("Thread creation wrote despite unreadable evidence: %v", statErr)
		}
	})

	t.Run("Thread membership", func(t *testing.T) {
		root := t.TempDir()
		taskID := testutil.TaskID("refused-membership-task")
		writeGraphMutationTask(t, root, "refused-membership-task", domain.StatusNextUp, nil, "")
		created, svc := createThreadForMutation(t, root, "refused-membership")
		before, err := os.ReadFile(created.Thread.Path)
		if err != nil {
			t.Fatal(err)
		}
		writeUnreadableGuardThread(t, root, "unreadable-membership-guard")

		receipt, err := svc.AddThreadMembers(created.Thread.ID, []string{taskID}, false)
		if !errors.Is(err, domain.ErrValidation) || receipt.Committed {
			t.Fatalf("receipt=%+v err=%v", receipt, err)
		}
		after, readErr := os.ReadFile(created.Thread.Path)
		if readErr != nil || !slices.Equal(before, after) {
			t.Fatalf("Thread membership wrote despite unreadable evidence: readErr=%v", readErr)
		}
	})

	t.Run("task lifecycle", func(t *testing.T) {
		root := t.TempDir()
		taskID := testutil.TaskID("refused-lifecycle-task")
		path := writeGraphMutationTask(t, root, "refused-lifecycle-task", domain.StatusNextUp, nil, "")
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		writeUnreadableGuardThread(t, root, "unreadable-lifecycle-guard")

		receipt, err := core.NewService(NewFS(root), core.WithRetry(0, func(int) {})).Move(
			taskID, domain.StatusReadyToStart, false, core.TaskLifecycleOverrideNone,
		)
		if !errors.Is(err, domain.ErrValidation) || receipt.Committed {
			t.Fatalf("receipt=%+v err=%v", receipt, err)
		}
		after, readErr := os.ReadFile(path)
		if readErr != nil || !slices.Equal(before, after) {
			t.Fatalf("task lifecycle wrote despite unreadable Thread evidence: readErr=%v", readErr)
		}
	})

	t.Run("Thread apply initial source", func(t *testing.T) {
		root := t.TempDir()
		repoID := "planning"
		taskID := testutil.TaskID("refused-apply-task")
		threadID := testutil.TaskID("refused-apply")
		writeGraphMutationTask(t, root, "refused-apply-task", domain.StatusNextUp, nil, "")
		writeUnreadableGuardThread(t, root, "unreadable-apply-guard")

		result, err := applyStoredThreadPlan(
			threadApplyStore(root, &repoID), storeThreadApplyPlan(threadID, []string{taskID}), false,
		)
		if !errors.Is(err, domain.ErrValidation) || result.Committed {
			t.Fatalf("result=%+v err=%v", result, err)
		}
		path := filepath.Join(root, domain.ThreadsDir, threadID+"-bulk-delivery.md")
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("Thread apply wrote despite unreadable evidence: %v", statErr)
		}
	})

	t.Run("Thread apply final snapshot verification", func(t *testing.T) {
		root := t.TempDir()
		repoID := "planning"
		taskID := testutil.TaskID("refused-final-apply-task")
		threadID := testutil.TaskID("refused-final-apply")
		writeGraphMutationTask(t, root, "refused-final-apply-task", domain.StatusNextUp, nil, "")

		original := testHookBeforeThreadApplyVerify
		t.Cleanup(func() { testHookBeforeThreadApplyVerify = original })
		testHookBeforeThreadApplyVerify = func() {
			testHookBeforeThreadApplyVerify = nil
			writeUnreadableGuardThread(t, root, "unreadable-final-apply-guard")
		}
		result, err := applyStoredThreadPlan(
			threadApplyStore(root, &repoID), storeThreadApplyPlan(threadID, []string{taskID}), false,
		)
		if !errors.Is(err, domain.ErrConflict) || result.Committed {
			t.Fatalf("result=%+v err=%v", result, err)
		}
		path := filepath.Join(root, domain.ThreadsDir, threadID+"-bulk-delivery.md")
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("Thread apply wrote after unreadable final evidence: %v", statErr)
		}
	})
}
