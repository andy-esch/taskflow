package store

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

var threadMutationStoreNow = time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC)

func newThreadMutationService(root, threadID string) *core.Service {
	return core.NewService(NewFS(root),
		core.WithIDGen(func() string { return threadID }),
		core.WithClock(func() time.Time { return threadMutationStoreNow }),
		core.WithRetry(0, func(int) {}),
	)
}

func createThreadForMutation(t *testing.T, root, seed string, taskRefs ...string) (core.ThreadCreationReceipt, *core.Service) {
	t.Helper()
	threadID := testutil.TaskID(seed)
	svc := newThreadMutationService(root, threadID)
	receipt, err := svc.NewThread(core.NewThreadParams{
		Title: seed, Description: "Exercise guarded Thread mutations", Goal: "Prove the existing-document writer", Tasks: taskRefs,
	})
	if err != nil {
		t.Fatal(err)
	}
	return receipt, svc
}

func TestThreadMembershipMutationIsSurgicalAtomicAndIdempotent(t *testing.T) {
	root := t.TempDir()
	aID := testutil.TaskID("membership-a")
	bID := testutil.TaskID("membership-b")
	cID := testutil.TaskID("membership-c")
	writeGraphMutationTask(t, root, "membership-a", domain.StatusNextUp, nil, "")
	writeGraphMutationTask(t, root, "membership-b", domain.StatusReadyToStart, nil, "")
	writeGraphMutationTask(t, root, "membership-c", domain.StatusCompleted, nil, "")
	created, svc := createThreadForMutation(t, root, "membership-thread", aID)

	content, err := os.ReadFile(created.Thread.Path)
	if err != nil {
		t.Fatal(err)
	}
	customized := strings.Replace(string(content), "status: unstarted", "status: unstarted # lifecycle comment\ncustom_field: keep-me", 1)
	if err := os.WriteFile(created.Thread.Path, []byte(customized), 0o644); err != nil {
		t.Fatal(err)
	}

	receipt, err := svc.AddThreadMembers(created.Thread.ID, []string{bID, aID}, false)
	if err != nil {
		t.Fatal(err)
	}
	wantMembers := []string{aID, bID}
	slices.Sort(wantMembers)
	if !receipt.Changed || !receipt.Committed || !slices.Equal(receipt.Thread.Tasks, wantMembers) {
		t.Fatalf("receipt = %+v", receipt)
	}
	outcomes := map[string]string{}
	for _, outcome := range receipt.MemberOutcomes {
		outcomes[outcome.TaskID] = outcome.Outcome
	}
	if len(receipt.MemberOutcomes) != 2 || outcomes[aID] != "skipped" || outcomes[bID] != "added" {
		t.Fatalf("outcomes = %+v", receipt.MemberOutcomes)
	}
	afterAdd, err := os.ReadFile(created.Thread.Path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"status: unstarted # lifecycle comment", "custom_field: keep-me", "updated_at: \"2026-08-30\"", "# Thread:"} {
		if !strings.Contains(string(afterAdd), want) {
			t.Errorf("surgical update lost %q:\n%s", want, afterAdd)
		}
	}

	noop, err := svc.RemoveThreadMembers(created.Thread.ID, []string{cID}, false)
	if err != nil || noop.Changed || noop.Committed || noop.MemberOutcomes[0].Outcome != "skipped" {
		t.Fatalf("no-op receipt=%+v err=%v", noop, err)
	}
	afterNoop, _ := os.ReadFile(created.Thread.Path)
	if !slices.Equal(afterAdd, afterNoop) {
		t.Fatal("idempotent remove rewrote the Thread")
	}

	beforeFailure := append([]byte(nil), afterNoop...)
	if _, err := svc.AddThreadMembers(created.Thread.ID, []string{cID, "missing-member"}, false); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("atomic failure = %v", err)
	}
	afterFailure, _ := os.ReadFile(created.Thread.Path)
	if !slices.Equal(beforeFailure, afterFailure) {
		t.Fatal("failed multi-member add partially changed the Thread")
	}
}

func TestThreadLifecycleMutationStampsAndClearsTerminalState(t *testing.T) {
	root := t.TempDir()
	memberID := testutil.TaskID("thread-lifecycle-member")
	writeGraphMutationTask(t, root, "thread-lifecycle-member", domain.StatusCompleted, nil, "")
	created, svc := createThreadForMutation(t, root, "lifecycle-thread", memberID)

	started, err := svc.StartThread(created.Thread.ID, false)
	if err != nil || started.Thread.Status != domain.ThreadStatusInProgress || started.Thread.StartedAt != "2026-08-30" {
		t.Fatalf("started=%+v err=%v", started, err)
	}
	startedBytes, _ := os.ReadFile(created.Thread.Path)
	startedNoop, err := svc.StartThread(created.Thread.ID, false)
	startedBytesAfterNoop, _ := os.ReadFile(created.Thread.Path)
	if err != nil || startedNoop.Changed || startedNoop.Committed || !slices.Equal(startedBytes, startedBytesAfterNoop) {
		t.Fatalf("same-state start receipt=%+v err=%v bytesChanged=%t", startedNoop, err, !slices.Equal(startedBytes, startedBytesAfterNoop))
	}
	completed, err := svc.CompleteThread(created.Thread.ID, false)
	if err != nil || completed.Thread.Status != domain.ThreadStatusCompleted || completed.Thread.EndedAt != "2026-08-30" {
		t.Fatalf("completed=%+v err=%v", completed, err)
	}
	completedBytes, _ := os.ReadFile(created.Thread.Path)
	completedNoop, err := svc.CompleteThread(created.Thread.ID, false)
	completedBytesAfterNoop, _ := os.ReadFile(created.Thread.Path)
	if err != nil || completedNoop.Changed || completedNoop.Committed || !slices.Equal(completedBytes, completedBytesAfterNoop) {
		t.Fatalf("same-state complete receipt=%+v err=%v bytesChanged=%t", completedNoop, err, !slices.Equal(completedBytes, completedBytesAfterNoop))
	}
	reopened, err := svc.ReopenThread(created.Thread.ID, false)
	if err != nil || reopened.Thread.Status != domain.ThreadStatusInProgress || reopened.Thread.EndedAt != "" {
		t.Fatalf("reopened=%+v err=%v", reopened, err)
	}
	cancelled, err := svc.CancelThread(created.Thread.ID, false)
	if err != nil || cancelled.Thread.Status != domain.ThreadStatusCancelled || cancelled.Thread.EndedAt != "2026-08-30" ||
		!slices.Equal(cancelled.Thread.Tasks, []string{memberID}) {
		t.Fatalf("cancelled=%+v err=%v", cancelled, err)
	}
	if !strings.Contains(cancelled.Remedy, "member tasks were not changed") {
		t.Fatalf("cancel remedy = %q", cancelled.Remedy)
	}
	if _, err := svc.AddThreadMembers(created.Thread.ID, []string{memberID}, false); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("cancelled membership error = %v", err)
	}
}

func TestThreadMutationDryRunReturnsProjectionWithoutWriting(t *testing.T) {
	root := t.TempDir()
	memberID := testutil.TaskID("dry-mutation-member")
	writeGraphMutationTask(t, root, "dry-mutation-member", domain.StatusNextUp, nil, "")
	created, svc := createThreadForMutation(t, root, "dry-mutation-thread")
	before, _ := os.ReadFile(created.Thread.Path)

	receipt, err := svc.AddThreadMembers(created.Thread.ID, []string{memberID}, true)
	if err != nil || !receipt.DryRun || !receipt.Changed || receipt.Committed || len(receipt.After.Members) != 1 {
		t.Fatalf("dry-run receipt=%+v err=%v", receipt, err)
	}
	after, _ := os.ReadFile(created.Thread.Path)
	if !slices.Equal(before, after) {
		t.Fatal("dry-run changed the Thread file")
	}
}

func TestThreadMutationPlannerCannotReenterStore(t *testing.T) {
	root := t.TempDir()
	memberID := testutil.TaskID("thread-reentry-member")
	writeGraphMutationTask(t, root, "thread-reentry-member", domain.StatusNextUp, nil, "")
	created, _ := createThreadForMutation(t, root, "thread-reentry")
	fs := NewFS(root)
	_, err := fs.MutateThread(threadMutationStoreNow, true, func(core.ThreadMutationSnapshot) (core.ThreadMutationPlan, error) {
		_, nestedErr := fs.ReadThreads()
		return core.ThreadMutationPlan{ThreadID: created.Thread.ID, Operation: core.ThreadMutationAddMembers, TaskIDs: []string{memberID}}, nestedErr
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("planner re-entry error = %v", err)
	}
}

func TestTaskLifecycleReceiptAttributesThreadProjectionWithoutWritingThread(t *testing.T) {
	root := t.TempDir()
	upstreamID := testutil.TaskID("thread-impact-upstream")
	downstreamID := testutil.TaskID("thread-impact-downstream")
	writeGraphMutationTask(t, root, "thread-impact-upstream", domain.StatusCompleted, nil, "")
	writeGraphMutationTask(t, root, "thread-impact-downstream", domain.StatusCompleted, []string{upstreamID}, "")
	created, svc := createThreadForMutation(t, root, "impact-thread", downstreamID)
	if _, err := svc.StartThread(created.Thread.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CompleteThread(created.Thread.ID, false); err != nil {
		t.Fatal(err)
	}
	threadBefore, _ := os.ReadFile(created.Thread.Path)

	receipt, err := svc.Move(upstreamID, domain.StatusNextUp, false, core.TaskLifecycleOverrideNone)
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.ThreadImpacts) != 1 || receipt.ThreadImpacts[0].ThreadID != created.Thread.ID || !receipt.ThreadImpacts[0].After.Inconsistent {
		t.Fatalf("Thread impacts = %+v", receipt.ThreadImpacts)
	}
	if !strings.Contains(receipt.Remedy, "newly inconsistent Thread") {
		t.Fatalf("remedy = %q", receipt.Remedy)
	}
	threadAfter, _ := os.ReadFile(created.Thread.Path)
	if !slices.Equal(threadBefore, threadAfter) {
		t.Fatal("task lifecycle impact attribution wrote the Thread document")
	}
}

func TestTaskLifecycleRefusesInvalidThreadEvidenceBeforeWriting(t *testing.T) {
	root := t.TempDir()
	taskID := testutil.TaskID("invalid-thread-evidence-task")
	taskPath := writeGraphMutationTask(t, root, "invalid-thread-evidence-task", domain.StatusNextUp, nil, "")
	created, _ := createThreadForMutation(t, root, "invalid-thread-evidence")
	content, _ := os.ReadFile(created.Thread.Path)
	missingID := testutil.TaskID("invalid-thread-missing-member")
	content = []byte(strings.Replace(string(content), "tasks: []", "tasks: ["+missingID+"]", 1))
	if err := os.WriteFile(created.Thread.Path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := core.NewService(NewFS(root), core.WithRetry(0, func(int) {})).Move(
		taskID, domain.StatusReadyToStart, false, core.TaskLifecycleOverrideNone,
	)
	if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), missingID) {
		t.Fatalf("invalid Thread evidence error = %v", err)
	}
	taskContent, _ := os.ReadFile(taskPath)
	if !strings.Contains(string(taskContent), "status: next-up") {
		t.Fatalf("task lifecycle wrote despite invalid Thread evidence:\n%s", taskContent)
	}
}

func TestThreadMutationAttributesReleaseFailureAfterCommit(t *testing.T) {
	root := t.TempDir()
	memberID := testutil.TaskID("thread-mutation-release-member")
	writeGraphMutationTask(t, root, "thread-mutation-release-member", domain.StatusNextUp, nil, "")
	created, _ := createThreadForMutation(t, root, "thread-mutation-release")
	retries := 0
	svc := core.NewService(NewFS(root),
		core.WithClock(func() time.Time { return threadMutationStoreNow }),
		core.WithRetry(3, func(int) { retries++ }),
	)
	original := testHookRepositoryUnlockError
	t.Cleanup(func() { testHookRepositoryUnlockError = original })
	testHookRepositoryUnlockError = func() error {
		return fmt.Errorf("injected Thread release failure wrapping conflict: %w", domain.ErrConflict)
	}

	receipt, err := svc.AddThreadMembers(created.Thread.ID, []string{memberID}, false)
	var committed *core.ThreadMutationFailure
	if !errors.As(err, &committed) || !errors.Is(err, domain.ErrConflict) || !receipt.Committed ||
		!committed.Receipt.Committed || retries != 0 {
		t.Fatalf("receipt=%+v err=%v retries=%d", receipt, err, retries)
	}
	testHookRepositoryUnlockError = nil
	thread, _, readErr := NewFS(root).GetThread(created.Thread.ID)
	if readErr != nil || !slices.Equal(thread.Tasks, []string{memberID}) {
		t.Fatalf("durable Thread=%+v err=%v", thread, readErr)
	}
}

func TestThreadMembershipWaitsForDependencyMutationAndUsesFreshGraph(t *testing.T) {
	root := t.TempDir()
	prerequisiteID := testutil.TaskID("serialized-thread-prerequisite")
	memberID := testutil.TaskID("serialized-thread-member")
	writeGraphMutationTask(t, root, "serialized-thread-prerequisite", domain.StatusNextUp, nil, "")
	writeGraphMutationTask(t, root, "serialized-thread-member", domain.StatusReadyToStart, nil, "")
	created, _ := createThreadForMutation(t, root, "serialized-membership-thread")

	entered, release := make(chan struct{}), make(chan struct{})
	original := testHookBeforeGraphVerify
	t.Cleanup(func() { testHookBeforeGraphVerify = original })
	testHookBeforeGraphVerify = func() {
		testHookBeforeGraphVerify = nil
		close(entered)
		<-release
	}

	dependencyDone := make(chan error, 1)
	go func() {
		_, err := core.NewService(NewFS(root), core.WithRetry(0, func(int) {})).AddTaskDependencies(
			memberID, []string{prerequisiteID}, false,
		)
		dependencyDone <- err
	}()
	<-entered

	type mutationOutcome struct {
		receipt core.ThreadMutationReceipt
		err     error
	}
	threadDone := make(chan mutationOutcome, 1)
	go func() {
		receipt, err := newThreadMutationService(root, "unused").AddThreadMembers(created.Thread.ID, []string{memberID}, false)
		threadDone <- mutationOutcome{receipt: receipt, err: err}
	}()
	select {
	case outcome := <-threadDone:
		close(release)
		t.Fatalf("Thread membership escaped dependency guard: receipt=%+v err=%v", outcome.receipt, outcome.err)
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	if err := <-dependencyDone; err != nil {
		t.Fatalf("dependency mutation: %v", err)
	}
	outcome := <-threadDone
	if outcome.err != nil || !outcome.receipt.Committed || len(outcome.receipt.After.ExternalGates) != 1 ||
		outcome.receipt.After.ExternalGates[0].State.TaskID != prerequisiteID || !outcome.receipt.After.ExternalGates[0].Outstanding {
		t.Fatalf("fresh Thread mutation = %+v err=%v", outcome.receipt, outcome.err)
	}
}

func TestThreadCompleteWaitsForTaskLifecycleAndRefusesFreshUndrainedState(t *testing.T) {
	root := t.TempDir()
	memberID := testutil.TaskID("serialized-completion-member")
	writeGraphMutationTask(t, root, "serialized-completion-member", domain.StatusCompleted, nil, "")
	created, svc := createThreadForMutation(t, root, "serialized-completion-thread", memberID)
	if _, err := svc.StartThread(created.Thread.ID, false); err != nil {
		t.Fatal(err)
	}

	entered, release := make(chan struct{}), make(chan struct{})
	original := testHookBeforeLifecycleVerify
	t.Cleanup(func() { testHookBeforeLifecycleVerify = original })
	testHookBeforeLifecycleVerify = func() {
		testHookBeforeLifecycleVerify = nil
		close(entered)
		<-release
	}

	lifecycleDone := make(chan error, 1)
	go func() {
		_, err := core.NewService(NewFS(root), core.WithRetry(0, func(int) {})).Move(
			memberID, domain.StatusNextUp, false, core.TaskLifecycleOverrideNone,
		)
		lifecycleDone <- err
	}()
	<-entered

	completeDone := make(chan error, 1)
	go func() {
		_, err := newThreadMutationService(root, "unused").CompleteThread(created.Thread.ID, false)
		completeDone <- err
	}()
	select {
	case err := <-completeDone:
		close(release)
		t.Fatalf("Thread completion escaped lifecycle guard: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	if err := <-lifecycleDone; err != nil {
		t.Fatalf("member lifecycle mutation: %v", err)
	}
	var policy *core.ThreadMutationPolicyError
	if err := <-completeDone; !errors.As(err, &policy) || !strings.Contains(policy.Reason, "soundly completed") {
		t.Fatalf("fresh completion refusal = %v", err)
	}
	thread, _, err := NewFS(root).GetThread(created.Thread.ID)
	if err != nil || thread.Status != domain.ThreadStatusInProgress {
		t.Fatalf("Thread after refusal = %+v err=%v", thread, err)
	}
}

func TestTaskLifecycleWaitsForThreadMembershipAndReportsFreshImpact(t *testing.T) {
	root := t.TempDir()
	memberID := testutil.TaskID("serialized-impact-member")
	writeGraphMutationTask(t, root, "serialized-impact-member", domain.StatusCompleted, nil, "")
	created, _ := createThreadForMutation(t, root, "serialized-impact-thread")

	entered, release := make(chan struct{}), make(chan struct{})
	original := testHookBeforeThreadMutationVerify
	t.Cleanup(func() { testHookBeforeThreadMutationVerify = original })
	testHookBeforeThreadMutationVerify = func() {
		testHookBeforeThreadMutationVerify = nil
		close(entered)
		<-release
	}

	threadDone := make(chan error, 1)
	go func() {
		_, err := newThreadMutationService(root, "unused").AddThreadMembers(created.Thread.ID, []string{memberID}, false)
		threadDone <- err
	}()
	<-entered

	type lifecycleOutcome struct {
		receipt core.TaskLifecycleReceipt
		err     error
	}
	lifecycleDone := make(chan lifecycleOutcome, 1)
	go func() {
		receipt, err := core.NewService(NewFS(root), core.WithRetry(0, func(int) {})).Move(
			memberID, domain.StatusNextUp, false, core.TaskLifecycleOverrideNone,
		)
		lifecycleDone <- lifecycleOutcome{receipt: receipt, err: err}
	}()
	select {
	case outcome := <-lifecycleDone:
		close(release)
		t.Fatalf("task lifecycle escaped Thread guard: receipt=%+v err=%v", outcome.receipt, outcome.err)
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	if err := <-threadDone; err != nil {
		t.Fatalf("Thread membership mutation: %v", err)
	}
	outcome := <-lifecycleDone
	if outcome.err != nil || !outcome.receipt.Committed || len(outcome.receipt.ThreadImpacts) != 1 ||
		outcome.receipt.ThreadImpacts[0].ThreadID != created.Thread.ID || !outcome.receipt.ThreadImpacts[0].Direct {
		t.Fatalf("fresh lifecycle impact = %+v err=%v", outcome.receipt, outcome.err)
	}
}

func TestThreadMutationRejectsRawRepositoryRaces(t *testing.T) {
	t.Run("whole graph snapshot", func(t *testing.T) {
		root := t.TempDir()
		memberID := testutil.TaskID("raw-graph-member")
		taskPath := writeGraphMutationTask(t, root, "raw-graph-member", domain.StatusNextUp, nil, "")
		created, svc := createThreadForMutation(t, root, "raw-graph-thread")
		original := testHookBeforeThreadMutationVerify
		t.Cleanup(func() { testHookBeforeThreadMutationVerify = original })
		testHookBeforeThreadMutationVerify = func() {
			testHookBeforeThreadMutationVerify = nil
			content, _ := os.ReadFile(taskPath)
			if err := os.WriteFile(taskPath, append(content, []byte("\n<!-- raw graph edit -->\n")...), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := svc.AddThreadMembers(created.Thread.ID, []string{memberID}, false); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("raw graph race = %v", err)
		}
		thread, _, _ := NewFS(root).GetThread(created.Thread.ID)
		if len(thread.Tasks) != 0 {
			t.Fatalf("stale membership committed: %+v", thread)
		}
	})

	t.Run("whole Thread snapshot", func(t *testing.T) {
		root := t.TempDir()
		memberID := testutil.TaskID("raw-thread-member")
		writeGraphMutationTask(t, root, "raw-thread-member", domain.StatusNextUp, nil, "")
		created, svc := createThreadForMutation(t, root, "raw-thread-snapshot")
		original := testHookBeforeThreadMutationVerify
		t.Cleanup(func() { testHookBeforeThreadMutationVerify = original })
		testHookBeforeThreadMutationVerify = func() {
			testHookBeforeThreadMutationVerify = nil
			content, _ := os.ReadFile(created.Thread.Path)
			if err := os.WriteFile(created.Thread.Path, append(content, []byte("\n<!-- raw Thread edit -->\n")...), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := svc.AddThreadMembers(created.Thread.ID, []string{memberID}, false); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("raw Thread snapshot race = %v", err)
		}
		content, _ := os.ReadFile(created.Thread.Path)
		if !strings.Contains(string(content), "raw Thread edit") || strings.Contains(string(content), "tasks: ["+memberID+"]") {
			t.Fatalf("raw edit was lost or stale membership committed:\n%s", content)
		}
	})

	t.Run("immediate target CAS", func(t *testing.T) {
		root := t.TempDir()
		memberID := testutil.TaskID("raw-target-member")
		writeGraphMutationTask(t, root, "raw-target-member", domain.StatusNextUp, nil, "")
		created, svc := createThreadForMutation(t, root, "raw-target-thread")
		original := testHookBeforeThreadMutationWrite
		t.Cleanup(func() { testHookBeforeThreadMutationWrite = original })
		testHookBeforeThreadMutationWrite = func(string) {
			testHookBeforeThreadMutationWrite = nil
			content, _ := os.ReadFile(created.Thread.Path)
			content = []byte(strings.Replace(string(content), "description: Exercise guarded Thread mutations", "description: Concurrent raw edit", 1))
			if err := os.WriteFile(created.Thread.Path, content, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := svc.AddThreadMembers(created.Thread.ID, []string{memberID}, false); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("raw target race = %v", err)
		}
		content, _ := os.ReadFile(created.Thread.Path)
		if !strings.Contains(string(content), "description: Concurrent raw edit") || strings.Contains(string(content), "tasks: ["+memberID+"]") {
			t.Fatalf("target edit was lost or stale membership committed:\n%s", content)
		}
	})
}

func TestTaskLifecycleRejectsRawThreadRaceByWholeSnapshotCAS(t *testing.T) {
	root := t.TempDir()
	memberID := testutil.TaskID("raw-lifecycle-thread-member")
	taskPath := writeGraphMutationTask(t, root, "raw-lifecycle-thread-member", domain.StatusNextUp, nil, "")
	created, _ := createThreadForMutation(t, root, "raw-lifecycle-thread", memberID)
	original := testHookBeforeLifecycleVerify
	t.Cleanup(func() { testHookBeforeLifecycleVerify = original })
	testHookBeforeLifecycleVerify = func() {
		testHookBeforeLifecycleVerify = nil
		content, _ := os.ReadFile(created.Thread.Path)
		if err := os.WriteFile(created.Thread.Path, append(content, []byte("\n<!-- raw Thread lifecycle race -->\n")...), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	_, err := core.NewService(NewFS(root), core.WithRetry(0, func(int) {})).Move(
		memberID, domain.StatusReadyToStart, false, core.TaskLifecycleOverrideNone,
	)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("raw Thread lifecycle race = %v", err)
	}
	taskContent, _ := os.ReadFile(taskPath)
	threadContent, _ := os.ReadFile(created.Thread.Path)
	if !strings.Contains(string(taskContent), "status: next-up") || !strings.Contains(string(threadContent), "raw Thread lifecycle race") {
		t.Fatalf("stale lifecycle landed or raw Thread edit was lost:\ntask:\n%s\nThread:\n%s", taskContent, threadContent)
	}
}
