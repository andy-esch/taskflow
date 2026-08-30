package store

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

var threadCreationNow = time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)

func TestThreadCreationPersistsCanonicalDocumentAndReadsItBack(t *testing.T) {
	root := t.TempDir()
	aID := testutil.TaskID("thread-member-a")
	bID := testutil.TaskID("thread-member-b")
	writeGraphMutationTask(t, root, "thread-member-a", domain.StatusReadyToStart, nil, "")
	writeGraphMutationTask(t, root, "thread-member-b", domain.StatusNextUp, nil, "")
	threadID := testutil.TaskID("created-thread")
	svc := core.NewService(NewFS(root),
		core.WithIDGen(func() string { return threadID }),
		core.WithClock(func() time.Time { return threadCreationNow }),
	)
	receipt, err := svc.NewThread(core.NewThreadParams{
		Title: "Thread foundation", Description: "Ship the Thread foundation", Goal: "Dogfood Thread planning",
		Tasks: []string{bID, "thread-member-a"}, Tags: []string{"threads", "dogfood"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.Committed || !receipt.Changed || receipt.Thread.Status != domain.ThreadStatusUnstarted {
		t.Fatalf("receipt = %+v", receipt)
	}
	if !slices.Equal(receipt.Thread.Tasks, []string{aID, bID}) {
		t.Fatalf("members = %v", receipt.Thread.Tasks)
	}
	content, err := os.ReadFile(receipt.Thread.Path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, want := range []string{
		"schema: 1", "id: " + threadID, "status: unstarted", "goal: Dogfood Thread planning",
		"created: \"2026-08-29\"", "tasks:", "# Thread: Thread foundation",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("created Thread missing %q:\n%s", want, text)
		}
	}
	reloaded, body, err := NewFS(root).GetThread(threadID)
	if err != nil || !slices.Equal(reloaded.Tasks, []string{aID, bID}) || !strings.Contains(body, "Dogfood Thread planning") {
		t.Fatalf("reloaded=%+v body=%q err=%v", reloaded, body, err)
	}
	if got, err := NewFS(root).ResolveThreadPath("thread-foundation"); err != nil || got != receipt.Thread.Path {
		t.Fatalf("path=%q err=%v", got, err)
	}
}

func TestThreadCreationDryRunValidatesWithoutWriting(t *testing.T) {
	root := t.TempDir()
	threadID := testutil.TaskID("dry-thread")
	svc := core.NewService(NewFS(root), core.WithIDGen(func() string { return threadID }))
	receipt, err := svc.NewThread(core.NewThreadParams{
		Title: "Dry Thread", Description: "Preview a Thread", Goal: "Write nothing", DryRun: true,
	})
	if err != nil || !receipt.DryRun || !receipt.Changed || receipt.Committed {
		t.Fatalf("receipt=%+v err=%v", receipt, err)
	}
	if _, err := os.Stat(filepath.Join(root, domain.ThreadsDir)); !os.IsNotExist(err) {
		t.Fatalf("dry-run created threads dir: %v", err)
	}
}

func TestThreadCreationRejectsRawTaskRaceByWholeSnapshotCAS(t *testing.T) {
	root := t.TempDir()
	taskPath := writeGraphMutationTask(t, root, "member", domain.StatusReadyToStart, nil, "")
	memberID := testutil.TaskID("member")
	threadID := testutil.TaskID("raw-race-thread")
	thread := domain.Thread{
		ID: threadID, Slug: "raw-race-thread", Status: domain.ThreadStatusUnstarted,
		Description: "Race test", Goal: "Reject a stale snapshot", Created: "2026-08-29", Tasks: []string{memberID},
	}
	original := testHookBeforeThreadCreationVerify
	defer func() { testHookBeforeThreadCreationVerify = original }()
	testHookBeforeThreadCreationVerify = func() {
		testHookBeforeThreadCreationVerify = nil
		content, err := os.ReadFile(taskPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(taskPath, append(content, []byte("\n<!-- raw race -->\n")...), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	result, err := NewFS(root).MutateThreadCreation(threadCreationNow, false, func(core.ThreadCreationSnapshot) (core.ThreadCreationPlan, error) {
		return core.ThreadCreationPlan{Thread: thread, Body: "# Race\n"}, nil
	})
	if !errors.Is(err, domain.ErrConflict) || result.Committed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if _, statErr := os.Stat(filepath.Join(root, domain.ThreadsDir, threadID+"-raw-race-thread.md")); !os.IsNotExist(statErr) {
		t.Fatalf("stale create landed: %v", statErr)
	}
}

func TestThreadCreationRejectsRawThreadRaceByWholeSnapshotCAS(t *testing.T) {
	root := t.TempDir()
	targetID := testutil.TaskID("raw-thread-race-target")
	concurrentID := testutil.TaskID("raw-thread-race-peer")
	original := testHookBeforeThreadCreationVerify
	t.Cleanup(func() { testHookBeforeThreadCreationVerify = original })
	testHookBeforeThreadCreationVerify = func() {
		testHookBeforeThreadCreationVerify = nil
		testutil.Write(t, filepath.Join(root, domain.ThreadsDir, concurrentID+"-peer.md"), "---\n"+
			"schema: 1\nid: "+concurrentID+"\nstatus: unstarted\ndescription: Concurrent Thread\n"+
			"goal: Change the guarded snapshot\ncreated: \"2026-08-29\"\ntasks: []\n---\n# Peer\n")
	}
	result, err := NewFS(root).MutateThreadCreation(threadCreationNow, false, func(core.ThreadCreationSnapshot) (core.ThreadCreationPlan, error) {
		return core.ThreadCreationPlan{Thread: domain.Thread{
			ID: targetID, Slug: "target", Status: domain.ThreadStatusUnstarted,
			Description: "Race test", Goal: "Reject stale Thread state", Created: "2026-08-29",
		}, Body: "# Target\n"}, nil
	})
	if !errors.Is(err, domain.ErrConflict) || result.Committed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if _, statErr := os.Stat(filepath.Join(root, domain.ThreadsDir, targetID+"-target.md")); !os.IsNotExist(statErr) {
		t.Fatalf("stale create landed: %v", statErr)
	}
}

func TestThreadCreationPlannerCannotReenterStore(t *testing.T) {
	fs := NewFS(t.TempDir())
	_, err := fs.MutateThreadCreation(threadCreationNow, true, func(core.ThreadCreationSnapshot) (core.ThreadCreationPlan, error) {
		_, _, nestedErr := fs.ListThreads()
		return core.ThreadCreationPlan{}, nestedErr
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("planner re-entry error = %v", err)
	}
}

func TestTaskAndThreadCreationSerializeCrossKindIdentity(t *testing.T) {
	root := t.TempDir()
	fs := NewFS(root)
	epic, err := fs.CreateEpic("identity", domain.Epic{
		Status: domain.EpicStatusActive, Description: "Identity tests", Priority: "medium", Created: "2026-08-29",
	}, "# Identity\n", false)
	if err != nil {
		t.Fatal(err)
	}
	sharedID := testutil.TaskID("shared-cross-kind-id")
	newService := func() *core.Service {
		return core.NewService(NewFS(root),
			core.WithIDGen(func() string { return sharedID }),
			core.WithClock(func() time.Time { return threadCreationNow }),
			core.WithRetry(0, func(int) {}),
		)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	go func() {
		ready.Done()
		<-start
		_, err := newService().NewThread(core.NewThreadParams{
			Title: "Shared identity", Description: "Thread side", Goal: "Win or lose cleanly",
		})
		results <- err
	}()
	go func() {
		ready.Done()
		<-start
		_, err := newService().NewTask(core.NewTaskParams{
			Title: "Shared identity", Epic: epic.ID, Description: "Task side", Effort: "1h",
			Priority: "medium", Tier: 2, Autonomy: 3, Tags: []string{"identity"}, Start: true,
		})
		results <- err
	}()
	ready.Wait()
	close(start)
	errA, errB := <-results, <-results
	successes := 0
	conflicts := 0
	for _, result := range []error{errA, errB} {
		switch {
		case result == nil:
			successes++
		case errors.Is(result, domain.ErrConflict):
			conflicts++
		default:
			t.Fatalf("unexpected race result: %v", result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d errors=(%v, %v)", successes, conflicts, errA, errB)
	}
	tasks, _, taskErr := NewFS(root).ListTasks()
	threads, _, threadErr := NewFS(root).ListThreads()
	if taskErr != nil || threadErr != nil || len(tasks)+len(threads) != 1 {
		t.Fatalf("tasks=%d Threads=%d taskErr=%v threadErr=%v", len(tasks), len(threads), taskErr, threadErr)
	}
}

func TestThreadCreationWaitsForLifecycleMutationAndReadsFreshSnapshot(t *testing.T) {
	root := t.TempDir()
	memberID := testutil.TaskID("serialized-member")
	writeGraphMutationTask(t, root, "serialized-member", domain.StatusReadyToStart, nil, "")

	original := testHookBeforeLifecycleVerify
	t.Cleanup(func() { testHookBeforeLifecycleVerify = original })
	lifecycleHoldingGuard := make(chan struct{})
	releaseLifecycle := make(chan struct{})
	testHookBeforeLifecycleVerify = func() {
		testHookBeforeLifecycleVerify = nil
		close(lifecycleHoldingGuard)
		<-releaseLifecycle
	}

	type lifecycleOutcome struct {
		result core.TaskLifecycleMutationResult
		err    error
	}
	lifecycleDone := make(chan lifecycleOutcome, 1)
	go func() {
		result, err := NewFS(root).MutateTaskLifecycle(threadCreationNow, false, func(*core.TaskGraph) (core.TaskLifecyclePlan, error) {
			return core.TaskLifecyclePlan{TaskID: memberID, To: domain.StatusInProgress}, nil
		})
		lifecycleDone <- lifecycleOutcome{result: result, err: err}
	}()
	<-lifecycleHoldingGuard

	type threadOutcome struct {
		result core.ThreadCreationMutationResult
		err    error
	}
	threadDone := make(chan threadOutcome, 1)
	threadID := testutil.TaskID("serialized-thread")
	go func() {
		result, err := NewFS(root).MutateThreadCreation(threadCreationNow, false, func(snapshot core.ThreadCreationSnapshot) (core.ThreadCreationPlan, error) {
			member, ok := snapshot.Graph.Task(memberID)
			if !ok || member.Status != domain.StatusInProgress {
				return core.ThreadCreationPlan{}, errors.New("thread planner did not observe committed lifecycle state")
			}
			return core.ThreadCreationPlan{Thread: domain.Thread{
				ID: threadID, Slug: "serialized-thread", Status: domain.ThreadStatusUnstarted,
				Description: "Serialization test", Goal: "Read the fresh graph", Created: "2026-08-29",
				Tasks: []string{memberID},
			}, Body: "# Serialized Thread\n"}, nil
		})
		threadDone <- threadOutcome{result: result, err: err}
	}()

	select {
	case outcome := <-threadDone:
		close(releaseLifecycle)
		t.Fatalf("Thread creation bypassed lifecycle guard: result=%+v err=%v", outcome.result, outcome.err)
	case <-time.After(100 * time.Millisecond):
		// Expected: the independent FS instance shares the canonical-root guard.
	}
	close(releaseLifecycle)

	lifecycle := <-lifecycleDone
	if lifecycle.err != nil || !lifecycle.result.Committed {
		t.Fatalf("lifecycle result=%+v err=%v", lifecycle.result, lifecycle.err)
	}
	thread := <-threadDone
	if thread.err != nil || !thread.result.Committed {
		t.Fatalf("Thread result=%+v err=%v", thread.result, thread.err)
	}
}

func TestThreadCreationAttributesReleaseFailureAfterCommit(t *testing.T) {
	root := t.TempDir()
	threadID := testutil.TaskID("thread-release-failure")
	original := testHookRepositoryUnlockError
	t.Cleanup(func() { testHookRepositoryUnlockError = original })
	testHookRepositoryUnlockError = func() error { return errors.New("injected release failure") }

	result, err := NewFS(root).MutateThreadCreation(threadCreationNow, false, func(core.ThreadCreationSnapshot) (core.ThreadCreationPlan, error) {
		return core.ThreadCreationPlan{Thread: domain.Thread{
			ID: threadID, Slug: "release-failure", Status: domain.ThreadStatusUnstarted,
			Description: "Recovery test", Goal: "Retain committed truth", Created: "2026-08-29",
		}, Body: "# Recovery\n"}, nil
	})
	if err == nil || !strings.Contains(err.Error(), "release repository Thread creation guard") || !result.Committed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	testHookRepositoryUnlockError = nil
	thread, _, readErr := NewFS(root).GetThread(threadID)
	if readErr != nil || thread.ID != threadID {
		t.Fatalf("committed Thread was not readable: Thread=%+v err=%v", thread, readErr)
	}
}

func TestFixSweepIncludesThreadAtomicWriteOrphans(t *testing.T) {
	root := t.TempDir()
	orphan := filepath.Join(root, domain.ThreadsDir, ".tskflwctl-crashed.tmp")
	testutil.Write(t, orphan, "partial")
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(orphan, old, old); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Fatalf("aged Thread temp orphan survived sweep: %v", err)
	}
}

func TestFixRepairsOrdinaryThreadFrontmatterWithoutChangingMembership(t *testing.T) {
	root := t.TempDir()
	threadID := testutil.TaskID("fix-thread-frontmatter")
	memberID := testutil.TaskID("fix-thread-member")
	path := filepath.Join(root, domain.ThreadsDir, threadID+"-fixable.md")
	testutil.Write(t, path, "---\nstatus: unstarted\ndescription: Phase 1: ship it\n"+
		"goal: Repair ordinary fields\ncreated: \"2026-08-29\"\ntags: one,two\ntasks: ["+memberID+"]\n---\n# Fixable\n")

	results, err := NewFS(root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Skipped {
		t.Fatalf("results = %+v", results)
	}
	thread, _, err := NewFS(root).GetThread(threadID)
	if err != nil {
		t.Fatal(err)
	}
	if thread.ID != threadID || thread.Description != "Phase 1: ship it" || !slices.Equal(thread.Tasks, []string{memberID}) {
		t.Fatalf("repaired Thread = %+v", thread)
	}
}

func TestFixDoesNotNormalizeThreadMembershipSyntax(t *testing.T) {
	root := t.TempDir()
	threadID := testutil.TaskID("unsafe-membership-fix")
	memberID := testutil.TaskID("unsafe-membership-member")
	path := filepath.Join(root, domain.ThreadsDir, threadID+"-unsafe.md")
	original := "---\nid: " + threadID + "\nstatus: unstarted\ndescription: Unsafe membership\n" +
		"goal: Preserve guarded intent\ncreated: \"2026-08-29\"\ntasks: " + memberID + "\n---\n# Unsafe\n"
	testutil.Write(t, path, original)

	results, err := NewFS(root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("membership syntax was treated as fixable: %+v", results)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != original {
		t.Fatalf("membership changed: err=%v\n%s", err, raw)
	}
}

func TestFixRefusesCrossKindCollisionWhileCanonicalizingID(t *testing.T) {
	for _, tc := range []struct {
		name      string
		badDir    string
		otherDir  string
		otherKind string
	}{
		{name: "Thread into task", badDir: domain.ThreadsDir, otherDir: domain.TasksDir, otherKind: "task"},
		{name: "task into Thread", badDir: domain.TasksDir, otherDir: domain.ThreadsDir, otherKind: "Thread"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			badID, goodID := "6fbj87000lt6", "6fbj870001t6"
			testutil.Write(t, filepath.Join(root, tc.otherDir, goodID+"-owner.md"), "---\nid: "+goodID+"\n---\n")
			testutil.Write(t, filepath.Join(root, tc.badDir, badID+"-candidate.md"), "---\nid: "+badID+"\n---\n")

			results, err := NewFS(root).FixFrontmatter(false)
			if err != nil {
				t.Fatal(err)
			}
			var refusal *domain.FixResult
			for i := range results {
				if strings.Contains(results[i].Path, badID) {
					refusal = &results[i]
				}
			}
			if refusal == nil || !refusal.Skipped || !strings.Contains(refusal.Changes[0], tc.otherKind) {
				t.Fatalf("cross-kind refusal = %+v", results)
			}
			if _, err := os.Stat(filepath.Join(root, tc.badDir, badID+"-candidate.md")); err != nil {
				t.Fatalf("candidate moved despite collision: %v", err)
			}
		})
	}
}
