package store

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

var lifecycleMutationNow = time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)

func lifecyclePlan(ref string, to domain.Status, override core.TaskLifecycleOverride) core.TaskLifecyclePlanner {
	return func(graph *core.TaskGraph) (core.TaskLifecyclePlan, error) {
		taskID, err := graph.ResolveTaskID(ref)
		if err != nil {
			return core.TaskLifecyclePlan{}, err
		}
		return core.TaskLifecyclePlan{TaskID: taskID, To: to, Override: override}, nil
	}
}

func moveTaskForTest(fs *FS, ref string, to domain.Status, now time.Time, dryRun bool, override core.TaskLifecycleOverride) (domain.Task, error) {
	result, err := fs.MutateTaskLifecycle(now, dryRun, lifecyclePlan(ref, to, override))
	return result.Task, err
}

func deferTaskForTest(fs *FS, ref, until string, now time.Time, dryRun bool) (domain.Task, error) {
	result, err := fs.MutateTaskLifecycle(now, dryRun, func(graph *core.TaskGraph) (core.TaskLifecyclePlan, error) {
		taskID, err := graph.ResolveTaskID(ref)
		if err != nil {
			return core.TaskLifecyclePlan{}, err
		}
		return core.TaskLifecyclePlan{TaskID: taskID, To: domain.StatusDeferred, RevisitAt: until}, nil
	})
	return result.Task, err
}

func TestMutateTaskLifecycleEnforcesEligibilityAndPersistsForcedExplanation(t *testing.T) {
	root := t.TempDir()
	prerequisiteID := testutil.TaskID("prerequisite")
	targetID := testutil.TaskID("target")
	writeGraphMutationTask(t, root, "prerequisite", domain.StatusNextUp, nil, "")
	targetPath := writeGraphMutationTask(t, root, "target", domain.StatusReadyToStart, []string{prerequisiteID}, "")
	fs := NewFS(root)

	_, err := fs.MutateTaskLifecycle(lifecycleMutationNow, false,
		lifecyclePlan("target", domain.StatusInProgress, core.TaskLifecycleOverrideNone))
	var eligibility *core.TaskEligibilityError
	if !errors.As(err, &eligibility) || len(eligibility.Blockers) != 1 {
		t.Fatalf("default start refusal = %v", err)
	}
	if content, readErr := os.ReadFile(targetPath); readErr != nil || strings.Contains(string(content), "status: in-progress") {
		t.Fatalf("refused start changed target: readErr=%v\n%s", readErr, content)
	}

	result, err := fs.MutateTaskLifecycle(lifecycleMutationNow, false,
		lifecyclePlan("target", domain.StatusInProgress, core.TaskLifecycleOverrideDependencyGate))
	if err != nil {
		t.Fatal(err)
	}
	if result.From != domain.StatusReadyToStart || result.Task.ID != targetID ||
		result.Task.Status != domain.StatusInProgress || !result.OverrideApplied ||
		len(result.OutstandingBlockers) != 1 || !result.Changed {
		t.Fatalf("forced result = %+v", result)
	}
	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"status: in-progress", "started_at: \"2026-08-28\"", "updated_at: \"2026-08-28\""} {
		if !strings.Contains(string(content), want) {
			t.Errorf("persisted start missing %q:\n%s", want, content)
		}
	}
}

func TestMutateTaskLifecycleDryRunReturnsWouldBeStateWithoutWriting(t *testing.T) {
	root := t.TempDir()
	path := writeGraphMutationTask(t, root, "candidate", domain.StatusReadyToStart, nil, "")
	before, _ := os.ReadFile(path)

	result, err := NewFS(root).MutateTaskLifecycle(lifecycleMutationNow, true,
		lifecyclePlan("candidate", domain.StatusInProgress, core.TaskLifecycleOverrideNone))
	if err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if !result.DryRun || !result.Changed || result.Task.Status != domain.StatusInProgress ||
		!strings.EqualFold(string(before), string(after)) {
		t.Fatalf("dry-run result=%+v changedFile=%v", result, string(before) != string(after))
	}
}

func TestMutateTaskLifecycleCreateAndStartIsOneGuardedOperation(t *testing.T) {
	root := t.TempDir()
	writeGraphMutationTask(t, root, "existing", domain.StatusCompleted, nil, "")
	taskID := testutil.TaskID("created")
	task := domain.Task{
		ID: taskID, Slug: "created", Status: domain.StatusReadyToStart,
		Description: "created and started", Effort: "1h", Tier: 1, Priority: "high",
		Autonomy: 2, Tags: []string{"graph"}, Created: "2026-08-28",
	}
	result, err := NewFS(root).MutateTaskLifecycle(lifecycleMutationNow, false,
		func(*core.TaskGraph) (core.TaskLifecyclePlan, error) {
			return core.TaskLifecyclePlan{To: domain.StatusInProgress, Create: &core.TaskLifecycleCreation{
				Task: task, Body: "# Created\n",
			}}, nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if result.From != domain.StatusReadyToStart || result.Task.Status != domain.StatusInProgress || !result.Changed {
		t.Fatalf("create-and-start result = %+v", result)
	}
	reloaded, _, err := NewFS(root).GetTask(taskID)
	if err != nil || reloaded.Status != domain.StatusInProgress || reloaded.StartedAt != "2026-08-28" {
		t.Fatalf("reloaded create-and-start = %+v err=%v", reloaded, err)
	}
}

func TestMutateTaskLifecycleRejectsRawPrerequisiteRaceByWholeGraphCAS(t *testing.T) {
	root := t.TempDir()
	prerequisiteID := testutil.TaskID("prerequisite")
	prerequisitePath := writeGraphMutationTask(t, root, "prerequisite", domain.StatusCompleted, nil, "")
	targetPath := writeGraphMutationTask(t, root, "target", domain.StatusReadyToStart, []string{prerequisiteID}, "")

	testHookBeforeLifecycleVerify = func() {
		content, err := os.ReadFile(prerequisitePath)
		if err != nil {
			t.Fatal(err)
		}
		content = []byte(strings.Replace(string(content), "status: completed", "status: next-up", 1))
		if err := os.WriteFile(prerequisitePath, content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { testHookBeforeLifecycleVerify = nil })

	_, err := NewFS(root).MutateTaskLifecycle(lifecycleMutationNow, false,
		lifecyclePlan("target", domain.StatusInProgress, core.TaskLifecycleOverrideNone))
	if !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "graph changed") {
		t.Fatalf("raw prerequisite race = %v", err)
	}
	content, readErr := os.ReadFile(targetPath)
	if readErr != nil || !strings.Contains(string(content), "status: ready-to-start") {
		t.Fatalf("stale start committed: readErr=%v\n%s", readErr, content)
	}
}

func TestMutateTaskLifecycleRejectsRawDependencyRacesDefaultAndForced(t *testing.T) {
	tests := []struct {
		name       string
		initialDep bool
		override   core.TaskLifecycleOverride
	}{
		{name: "default gains blocker", initialDep: false, override: core.TaskLifecycleOverrideNone},
		{name: "forced loses blocker", initialDep: true, override: core.TaskLifecycleOverrideDependencyGate},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			prerequisiteID := testutil.TaskID("prerequisite")
			writeGraphMutationTask(t, root, "prerequisite", domain.StatusNextUp, nil, "")
			dependencies := []string(nil)
			if tc.initialDep {
				dependencies = []string{prerequisiteID}
			}
			targetPath := writeGraphMutationTask(t, root, "target", domain.StatusReadyToStart, dependencies, "")

			testHookBeforeLifecycleVerify = func() {
				content, err := os.ReadFile(targetPath)
				if err != nil {
					t.Fatal(err)
				}
				if tc.initialDep {
					content = []byte(strings.Replace(string(content), "depends_on: ["+prerequisiteID+"]\n", "", 1))
				} else {
					content = []byte(strings.Replace(string(content), "---\n# target", "depends_on: ["+prerequisiteID+"]\n---\n# target", 1))
				}
				if err := os.WriteFile(targetPath, content, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			t.Cleanup(func() { testHookBeforeLifecycleVerify = nil })

			_, err := NewFS(root).MutateTaskLifecycle(lifecycleMutationNow, false,
				lifecyclePlan("target", domain.StatusInProgress, tc.override))
			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("dependency race = %v", err)
			}
			content, _ := os.ReadFile(targetPath)
			if !strings.Contains(string(content), "status: ready-to-start") {
				t.Fatalf("stale start committed:\n%s", content)
			}
		})
	}
}

func TestMutateTaskLifecycleRejectsTargetRaceAtImmediateCAS(t *testing.T) {
	root := t.TempDir()
	targetPath := writeGraphMutationTask(t, root, "target", domain.StatusReadyToStart, nil, "")

	testHookBeforeLifecycleWrite = func(taskID string) {
		if taskID != testutil.TaskID("target") {
			t.Fatalf("write hook task = %q", taskID)
		}
		content, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatal(err)
		}
		content = []byte(strings.Replace(string(content), "description: target", "description: concurrent edit", 1))
		if err := os.WriteFile(targetPath, content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { testHookBeforeLifecycleWrite = nil })

	_, err := NewFS(root).MutateTaskLifecycle(lifecycleMutationNow, false,
		lifecyclePlan("target", domain.StatusInProgress, core.TaskLifecycleOverrideNone))
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("target race = %v", err)
	}
	content, readErr := os.ReadFile(targetPath)
	if readErr != nil || !strings.Contains(string(content), "description: concurrent edit") ||
		!strings.Contains(string(content), "status: ready-to-start") {
		t.Fatalf("concurrent target edit did not survive: readErr=%v\n%s", readErr, content)
	}
}

func TestCooperatingDependencyMutationsSerializeBeforeStartAuthorization(t *testing.T) {
	for _, tc := range []struct {
		name       string
		operation  core.DependencyOperation
		override   core.TaskLifecycleOverride
		wantStart  bool
		wantForced bool
	}{
		{name: "add/default refuses fresh blocker", operation: core.DependencyAdd},
		{name: "add/override starts with fresh blocker", operation: core.DependencyAdd, override: core.TaskLifecycleOverrideDependencyGate, wantStart: true, wantForced: true},
		{name: "remove/default starts after fresh clear", operation: core.DependencyRemove, wantStart: true},
		{name: "remove/override remains unnecessary", operation: core.DependencyRemove, override: core.TaskLifecycleOverrideDependencyGate, wantStart: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			prerequisiteID := testutil.TaskID("prerequisite")
			initial := []string(nil)
			if tc.operation == core.DependencyRemove {
				initial = []string{prerequisiteID}
			}
			writeGraphMutationTask(t, root, "prerequisite", domain.StatusNextUp, nil, "")
			writeGraphMutationTask(t, root, "target", domain.StatusReadyToStart, initial, "")

			entered, release := make(chan struct{}), make(chan struct{})
			original := testHookBeforeGraphVerify
			t.Cleanup(func() { testHookBeforeGraphVerify = original })
			testHookBeforeGraphVerify = func() {
				testHookBeforeGraphVerify = nil
				close(entered)
				<-release
			}

			dependencyDone := make(chan error, 1)
			dependencyService := core.NewService(NewFS(root), core.WithRetry(0, func(int) {}))
			go func() {
				var err error
				if tc.operation == core.DependencyAdd {
					_, err = dependencyService.AddTaskDependencies("target", []string{"prerequisite"}, false)
				} else {
					_, err = dependencyService.RemoveTaskDependencies("target", []string{"prerequisite"}, false)
				}
				dependencyDone <- err
			}()
			<-entered

			type lifecycleOutcome struct {
				receipt core.TaskLifecycleReceipt
				err     error
			}
			startDone := make(chan lifecycleOutcome, 1)
			startService := core.NewService(NewFS(root), core.WithRetry(0, func(int) {}))
			go func() {
				receipt, err := startService.Move("target", domain.StatusInProgress, false, tc.override)
				startDone <- lifecycleOutcome{receipt: receipt, err: err}
			}()
			select {
			case outcome := <-startDone:
				t.Fatalf("start escaped cooperating repository guard: %+v, %v", outcome.receipt, outcome.err)
			case <-time.After(75 * time.Millisecond):
			}

			close(release)
			if err := <-dependencyDone; err != nil {
				t.Fatalf("dependency mutation: %v", err)
			}
			outcome := <-startDone
			if tc.wantStart {
				if outcome.err != nil || outcome.receipt.Task.Status != domain.StatusInProgress || outcome.receipt.Forced != tc.wantForced {
					t.Fatalf("fresh start outcome = %+v, %v", outcome.receipt, outcome.err)
				}
			} else {
				var eligibility *core.TaskEligibilityError
				if !errors.As(outcome.err, &eligibility) || len(eligibility.Blockers) != 1 {
					t.Fatalf("fresh default refusal = %+v, %v", outcome.receipt, outcome.err)
				}
			}
		})
	}
}

func TestCooperatingPrerequisiteReopenSerializesBeforeStartAuthorization(t *testing.T) {
	for _, override := range []core.TaskLifecycleOverride{
		core.TaskLifecycleOverrideNone,
		core.TaskLifecycleOverrideDependencyGate,
	} {
		t.Run(string(override), func(t *testing.T) {
			root := t.TempDir()
			prerequisiteID := testutil.TaskID("prerequisite")
			writeGraphMutationTask(t, root, "prerequisite", domain.StatusCompleted, nil, "")
			writeGraphMutationTask(t, root, "target", domain.StatusReadyToStart, []string{prerequisiteID}, "")

			entered, release := make(chan struct{}), make(chan struct{})
			original := testHookBeforeLifecycleVerify
			t.Cleanup(func() { testHookBeforeLifecycleVerify = original })
			testHookBeforeLifecycleVerify = func() {
				testHookBeforeLifecycleVerify = nil
				close(entered)
				<-release
			}

			reopenDone := make(chan error, 1)
			go func() {
				_, err := core.NewService(NewFS(root), core.WithRetry(0, func(int) {})).Move(
					"prerequisite", domain.StatusReadyToStart, false, core.TaskLifecycleOverrideNone)
				reopenDone <- err
			}()
			<-entered

			type lifecycleOutcome struct {
				receipt core.TaskLifecycleReceipt
				err     error
			}
			startDone := make(chan lifecycleOutcome, 1)
			go func() {
				receipt, err := core.NewService(NewFS(root), core.WithRetry(0, func(int) {})).Move(
					"target", domain.StatusInProgress, false, override)
				startDone <- lifecycleOutcome{receipt: receipt, err: err}
			}()
			select {
			case outcome := <-startDone:
				t.Fatalf("start escaped prerequisite lifecycle guard: %+v, %v", outcome.receipt, outcome.err)
			case <-time.After(75 * time.Millisecond):
			}

			close(release)
			if err := <-reopenDone; err != nil {
				t.Fatalf("prerequisite reopen: %v", err)
			}
			outcome := <-startDone
			if override == core.TaskLifecycleOverrideNone {
				var eligibility *core.TaskEligibilityError
				if !errors.As(outcome.err, &eligibility) || len(eligibility.Blockers) != 1 {
					t.Fatalf("fresh default refusal = %+v, %v", outcome.receipt, outcome.err)
				}
			} else if outcome.err != nil || !outcome.receipt.Forced || len(outcome.receipt.OutstandingBlockers) != 1 {
				t.Fatalf("fresh forced start = %+v, %v", outcome.receipt, outcome.err)
			}
		})
	}
}

func TestTaskLifecyclePlannerCannotReenterStore(t *testing.T) {
	root := t.TempDir()
	writeGraphMutationTask(t, root, "candidate", domain.StatusReadyToStart, nil, "")
	fs := NewFS(root)
	_, err := fs.MutateTaskLifecycle(lifecycleMutationNow, false, func(*core.TaskGraph) (core.TaskLifecyclePlan, error) {
		_, _, err := fs.GetTask("candidate")
		return core.TaskLifecyclePlan{}, err
	})
	if !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "mutation planner is active") {
		t.Fatalf("planner store reentry = %v", err)
	}

	// The sentinel must be released on the error path so a later guarded call works.
	if _, err := fs.MutateTaskLifecycle(lifecycleMutationNow, false,
		lifecyclePlan("candidate", domain.StatusInProgress, core.TaskLifecycleOverrideNone)); err != nil {
		t.Fatalf("planner sentinel leaked after reentry refusal: %v", err)
	}
}

func TestMutateTaskLifecycleAttributesReleaseFailureAfterCommit(t *testing.T) {
	root := t.TempDir()
	targetPath := writeGraphMutationTask(t, root, "target", domain.StatusReadyToStart, nil, "")
	fs := NewFS(root)
	original := testHookRepositoryUnlockError
	t.Cleanup(func() { testHookRepositoryUnlockError = original })
	testHookRepositoryUnlockError = func() error { return errors.New("injected unlock failure") }

	result, err := fs.MutateTaskLifecycle(lifecycleMutationNow, false,
		lifecyclePlan("target", domain.StatusInProgress, core.TaskLifecycleOverrideNone))
	if err == nil || !strings.Contains(err.Error(), "release repository task lifecycle guard") ||
		!strings.Contains(err.Error(), "injected unlock failure") || !result.Changed || !result.Committed {
		t.Fatalf("release result=%+v err=%v", result, err)
	}
	content, readErr := os.ReadFile(targetPath)
	if readErr != nil || !strings.Contains(string(content), "status: in-progress") {
		t.Fatalf("committed lifecycle write missing after release failure: readErr=%v\n%s", readErr, content)
	}

	testHookRepositoryUnlockError = nil
	if _, err := fs.MutateTaskLifecycle(lifecycleMutationNow, false,
		lifecyclePlan("target", domain.StatusReadyToStart, core.TaskLifecycleOverrideNone)); err != nil {
		t.Fatalf("guard stayed held after attributed release failure: %v", err)
	}
}

func TestMutateTaskLifecycleFailsClosedOnBrokenGraphEvenWhenForced(t *testing.T) {
	root := t.TempDir()
	writeGraphMutationTask(t, root, "target", domain.StatusReadyToStart, []string{testutil.TaskID("missing")}, "")
	_, err := NewFS(root).MutateTaskLifecycle(lifecycleMutationNow, false,
		lifecyclePlan("target", domain.StatusInProgress, core.TaskLifecycleOverrideDependencyGate))
	if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "repository task graph is broken") {
		t.Fatalf("forced start on broken graph = %v", err)
	}
}
