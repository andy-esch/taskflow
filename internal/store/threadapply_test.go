package store

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

var threadApplyStoreNow = time.Date(2026, time.August, 30, 15, 0, 0, 0, time.UTC)

func threadApplyStore(root string, repoID *string) *FS {
	return testutil.Must(NewFS(root, core.UnrestrictedMutations(), WithPlanningIdentityReader(func() (string, string, error) {
		return root, *repoID, nil
	})))
}

func storeThreadApplyPlan(threadID string, members []string, dependencies ...core.ThreadApplyDependency) core.ThreadApplyPlan {
	members = append([]string(nil), members...)
	slices.Sort(members)
	return core.ThreadApplyPlan{
		Schema: core.ThreadApplyPlanSchema, PlanningRepoID: "planning", ComposedAt: "2026-08-30",
		Thread: core.ThreadApplyThread{
			ID: threadID, Slug: "bulk-delivery", Status: domain.ThreadStatusUnstarted,
			Description: "Link existing work safely", Goal: "Create a resumable Thread",
			Created: "2026-08-30", Tags: []string{"bulk", "threads"}, Tasks: members,
			Body: "# Thread: Bulk delivery\n\n## Goal\n\nCreate a resumable Thread\n",
		},
		Dependencies: dependencies,
	}
}

func applyStoredThreadPlan(fs *FS, plan core.ThreadApplyPlan, dryRun bool) (core.ThreadApplyMutationResult, error) {
	return fs.MutateThreadApply(threadApplyStoreNow, dryRun, func(core.ThreadApplySnapshot) (core.ThreadApplyPlan, error) {
		return plan, nil
	})
}

func TestThreadApplyMalformedAndStalePlansDoNotPersist(t *testing.T) {
	gateID := testutil.TaskID("refusal-gate")
	memberID := testutil.TaskID("refusal-member")
	threadID := testutil.TaskID("refusal-thread")
	for _, tc := range []struct {
		name, diagnostic string
		conflict         bool
	}{
		{"schema", "unsupported Thread apply-plan schema", false},
		{"status", "planned Thread must be unstarted", false},
		{"slug", "canonical filename slug", false},
		{"date", "must equal composed_at", false},
		{"edge slug", "requires exact stable task ids", false},
		{"self edge", "cannot depend on itself", false},
		{"duplicate edge", "duplicate planned dependency", false},
		{"missing prerequisite", "planned prerequisite " + gateID + " does not exist", false},
		{"missing dependent", "planned dependent " + memberID + " does not exist", false},
		{"changed identity", "apply plan belongs to planning repository", true},
		{"changed body", "different body", true},
	} {
		for _, dryRun := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/preview=%t", tc.name, dryRun), func(t *testing.T) {
				root := t.TempDir()
				repoID := "planning"
				gatePath := writeGraphMutationTask(t, root, "refusal-gate", domain.StatusCompleted, nil, "")
				memberPath := writeGraphMutationTask(t, root, "refusal-member", domain.StatusNextUp, nil, "")
				fs := threadApplyStore(root, &repoID)
				svc := testutil.Must(core.NewService(fs, core.WithClock(func() time.Time { return threadApplyStoreNow }),
					core.WithIDGen(func() string { return threadID }), core.WithRetry(0, func(int) {})))
				nonmember := false
				manifest := core.ThreadComposeManifest{
					Thread: core.ThreadComposeInput{Title: "Refusal delivery", Description: "Retain recoverable intent", Goal: "Refuse without effects"},
					Nodes: []core.ThreadComposeNode{
						{Key: "gate", TaskID: gateID, Member: &nonmember},
						{Key: "member", TaskID: memberID},
					},
					Dependencies: []core.ThreadComposeDependency{{From: "gate", To: "member"}},
				}
				before := testutil.SnapshotTree(t, root)
				plan, err := svc.ComposeThreadApply(repoID, manifest)
				if err != nil {
					t.Fatal(err)
				}
				baseline, err := svc.ApplyThreadPlan(plan, true)
				if err != nil || !baseline.DryRun || baseline.Committed || !baseline.Changed {
					t.Fatalf("valid preview=%+v err=%v", baseline, err)
				}
				if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatal("valid compose/preview changed repository entries, bytes, or modes")
				}

				// Alter one valid boundary at a time; capture intentional fixture edits
				// before measuring effects of the attempted apply.
				switch tc.name {
				case "schema":
					plan.Schema++
				case "status":
					plan.Thread.Status = domain.ThreadStatusInProgress
				case "slug":
					plan.Thread.Slug = "../../outside"
				case "date":
					plan.Thread.Created = "2026-08-29"
				case "edge slug":
					plan.Dependencies[0].From = "refusal-gate"
				case "self edge":
					plan.Dependencies[0].From = memberID
				case "duplicate edge":
					plan.Dependencies = append(plan.Dependencies, plan.Dependencies[0])
				case "missing prerequisite":
					if err := os.Remove(gatePath); err != nil {
						t.Fatal(err)
					}
				case "missing dependent":
					if err := os.Remove(memberPath); err != nil {
						t.Fatal(err)
					}
				case "changed identity":
					repoID = "replacement"
				case "changed body":
					if _, err := svc.ApplyThreadPlan(plan, false); err != nil {
						t.Fatal(err)
					}
					path, err := fs.ResolveThreadPath(threadID)
					if err != nil {
						t.Fatal(err)
					}
					content, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, append(content, []byte("\nHuman recovery note.\n")...), 0o644); err != nil {
						t.Fatal(err)
					}
				default:
					t.Fatal("missing fixture edit")
				}
				before = testutil.SnapshotTree(t, root)
				receipt, err := svc.ApplyThreadPlan(plan, dryRun)
				want := domain.ErrValidation
				if tc.conflict {
					want = domain.ErrConflict
				}
				var failure *core.ThreadApplyFailure
				if !errors.Is(err, want) || !strings.Contains(err.Error(), tc.diagnostic) || !errors.As(err, &failure) {
					t.Fatalf("receipt=%+v err=%v; want typed %v containing %q", receipt, err, want, tc.diagnostic)
				}
				if receipt.Committed || receipt.Complete || receipt.Changed || receipt.DryRun != dryRun ||
					!reflect.DeepEqual(receipt.Plan, plan) || !reflect.DeepEqual(failure.Receipt, receipt) {
					t.Fatalf("refusal lost the plan or reported effects: %+v", receipt)
				}
				if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatal("refused apply changed repository entries, bytes, or modes")
				}
			})
		}
	}
}

func TestThreadApplyPersistsDependenciesThenThreadAndConverges(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	gateID := testutil.TaskID("apply-gate")
	firstID := testutil.TaskID("apply-first")
	secondID := testutil.TaskID("apply-second")
	threadID := testutil.TaskID("apply-thread")
	writeGraphMutationTask(t, root, "apply-gate", domain.StatusCompleted, nil, "")
	writeGraphMutationTask(t, root, "apply-first", domain.StatusNextUp, nil, "")
	writeGraphMutationTask(t, root, "apply-second", domain.StatusReadyToStart, nil, "")
	plan := storeThreadApplyPlan(threadID, []string{firstID, secondID},
		core.ThreadApplyDependency{From: gateID, To: firstID},
		core.ThreadApplyDependency{From: firstID, To: secondID},
	)

	result, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || !result.Complete || !result.Committed || result.DryRun {
		t.Fatalf("result = %+v", result)
	}
	for _, operation := range result.Operations {
		if operation.State != core.ThreadApplyApplied {
			t.Fatalf("operation was not applied: %+v", operation)
		}
	}
	graph, err := core.LoadTaskGraph(testutil.Must(NewFS(root, core.ReadOnlyMutations())))
	if err != nil || graph.Health() != core.GraphHealthy {
		t.Fatalf("graph health=%v err=%v", graph.Health(), err)
	}
	first, _ := graph.Task(firstID)
	second, _ := graph.Task(secondID)
	if !slices.Contains(first.DependsOn, gateID) || !slices.Contains(second.DependsOn, firstID) {
		t.Fatalf("dependencies not persisted: first=%v second=%v", first.DependsOn, second.DependsOn)
	}
	thread, body, err := testutil.Must(NewFS(root, core.ReadOnlyMutations())).GetThread(threadID)
	if err != nil || !slices.Equal(thread.Tasks, plan.Thread.Tasks) || body != plan.Thread.Body {
		t.Fatalf("thread=%+v body=%q err=%v", thread, body, err)
	}
	firstPath, _ := testutil.Must(NewFS(root, core.ReadOnlyMutations())).resolvePath(firstID)
	secondPath, _ := testutil.Must(NewFS(root, core.ReadOnlyMutations())).resolvePath(secondID)
	threadPath, err := testutil.Must(NewFS(root, core.ReadOnlyMutations())).ResolveThreadPath(threadID)
	if err != nil {
		t.Fatal(err)
	}
	beforeFirst, _ := os.ReadFile(firstPath)
	beforeSecond, _ := os.ReadFile(secondPath)
	beforeThread, _ := os.ReadFile(threadPath)

	converged, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if err != nil || converged.Changed || !converged.Complete || converged.Committed {
		t.Fatalf("idempotent result=%+v err=%v", converged, err)
	}
	for _, operation := range converged.Operations {
		if operation.State != core.ThreadApplySkipped {
			t.Fatalf("idempotent operation = %+v", operation)
		}
	}
	afterFirst, _ := os.ReadFile(firstPath)
	afterSecond, _ := os.ReadFile(secondPath)
	afterThread, _ := os.ReadFile(threadPath)
	if !slices.Equal(beforeFirst, afterFirst) || !slices.Equal(beforeSecond, afterSecond) || !slices.Equal(beforeThread, afterThread) {
		t.Fatal("idempotent retry rewrote an already-converged file")
	}
}

func TestThreadApplyInterruptedPrefixRetriesToCompletion(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	aID := testutil.TaskID("prefix-a")
	bID := testutil.TaskID("prefix-b")
	cID := testutil.TaskID("prefix-c")
	threadID := testutil.TaskID("prefix-thread")
	writeGraphMutationTask(t, root, "prefix-a", domain.StatusCompleted, nil, "")
	writeGraphMutationTask(t, root, "prefix-b", domain.StatusNextUp, nil, "")
	writeGraphMutationTask(t, root, "prefix-c", domain.StatusReadyToStart, nil, "")
	plan := storeThreadApplyPlan(threadID, []string{bID, cID},
		core.ThreadApplyDependency{From: aID, To: bID},
		core.ThreadApplyDependency{From: bID, To: cID},
	)

	original := testHookAfterThreadApplyWrite
	t.Cleanup(func() { testHookAfterThreadApplyWrite = original })
	testHookAfterThreadApplyWrite = func(kind, _ string) error {
		if kind == "dependency" {
			testHookAfterThreadApplyWrite = nil
			return errors.New("power loss")
		}
		return nil
	}
	interrupted, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if err == nil || !interrupted.Committed || interrupted.Complete {
		t.Fatalf("interrupted=%+v err=%v", interrupted, err)
	}
	applied, pending := 0, 0
	for _, operation := range interrupted.Operations {
		switch operation.State {
		case core.ThreadApplyApplied:
			applied++
		case core.ThreadApplyPending:
			pending++
		}
	}
	if applied != 1 || pending != 2 {
		t.Fatalf("durable prefix operations=%+v", interrupted.Operations)
	}
	if _, statErr := os.Stat(filepath.Join(root, domain.ThreadsDir, threadID+"-bulk-delivery.md")); !os.IsNotExist(statErr) {
		t.Fatalf("Thread landed before dependencies completed: %v", statErr)
	}
	graph, loadErr := core.LoadTaskGraph(testutil.Must(NewFS(root, core.ReadOnlyMutations())))
	if loadErr != nil || graph.Health() != core.GraphHealthy {
		t.Fatalf("interrupted graph health=%s err=%v", graph.Health(), loadErr)
	}

	resumed, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if err != nil || !resumed.Complete || !resumed.Committed {
		t.Fatalf("resumed=%+v err=%v", resumed, err)
	}
	if _, _, err := testutil.Must(NewFS(root, core.ReadOnlyMutations())).GetThread(threadID); err != nil {
		t.Fatalf("resumed Thread: %v", err)
	}
}

func TestThreadApplyEveryDurablePrefixRetriesToCompletion(t *testing.T) {
	const dependencyWrites = 3
	for failAfter := range dependencyWrites + 1 { // three task replacements, then the Thread create
		t.Run(fmt.Sprintf("after operation %d", failAfter+1), func(t *testing.T) {
			root := t.TempDir()
			repoID := "planning"
			ids := []string{
				testutil.TaskID(fmt.Sprintf("all-prefixes-%d-a", failAfter)),
				testutil.TaskID(fmt.Sprintf("all-prefixes-%d-b", failAfter)),
				testutil.TaskID(fmt.Sprintf("all-prefixes-%d-c", failAfter)),
				testutil.TaskID(fmt.Sprintf("all-prefixes-%d-d", failAfter)),
			}
			for index, suffix := range []string{"a", "b", "c", "d"} {
				status := domain.StatusNextUp
				if index == 0 {
					status = domain.StatusCompleted
				}
				writeGraphMutationTask(t, root, fmt.Sprintf("all-prefixes-%d-%s", failAfter, suffix), status, nil, "")
			}
			threadID := testutil.TaskID(fmt.Sprintf("all-prefixes-%d-thread", failAfter))
			plan := storeThreadApplyPlan(threadID, ids[1:],
				core.ThreadApplyDependency{From: ids[0], To: ids[1]},
				core.ThreadApplyDependency{From: ids[1], To: ids[2]},
				core.ThreadApplyDependency{From: ids[2], To: ids[3]},
			)

			original := testHookAfterThreadApplyWrite
			t.Cleanup(func() { testHookAfterThreadApplyWrite = original })
			operation := 0
			testHookAfterThreadApplyWrite = func(_, _ string) error {
				current := operation
				operation++
				if current == failAfter {
					return errors.New("injected interruption")
				}
				return nil
			}
			interrupted, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
			if err == nil || !interrupted.Committed {
				t.Fatalf("interrupted=%+v err=%v", interrupted, err)
			}
			wantComplete := failAfter == dependencyWrites
			if interrupted.Complete != wantComplete {
				t.Fatalf("complete=%v, want %v; operations=%+v", interrupted.Complete, wantComplete, interrupted.Operations)
			}
			graph, loadErr := core.LoadTaskGraph(testutil.Must(NewFS(root, core.ReadOnlyMutations())))
			if loadErr != nil || graph.Health() != core.GraphHealthy {
				t.Fatalf("prefix graph health=%s err=%v", graph.Health(), loadErr)
			}

			testHookAfterThreadApplyWrite = nil
			resumed, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
			if err != nil || !resumed.Complete {
				t.Fatalf("resumed=%+v err=%v", resumed, err)
			}
			if _, _, err := testutil.Must(NewFS(root, core.ReadOnlyMutations())).GetThread(threadID); err != nil {
				t.Fatalf("resumed Thread: %v", err)
			}
		})
	}
}

func TestThreadApplyIdentityAndDryRunFailClosed(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	taskID := testutil.TaskID("identity-member")
	threadID := testutil.TaskID("identity-thread")
	path := writeGraphMutationTask(t, root, "identity-member", domain.StatusNextUp, nil, "")
	before, _ := os.ReadFile(path)
	plan := storeThreadApplyPlan(threadID, []string{taskID})

	preview, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, true)
	if err != nil || !preview.DryRun || !preview.Changed || preview.Complete || preview.Committed {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
	if _, statErr := os.Stat(filepath.Join(root, domain.ThreadsDir)); !os.IsNotExist(statErr) {
		t.Fatalf("dry-run created Thread directory: %v", statErr)
	}
	after, _ := os.ReadFile(path)
	if !slices.Equal(before, after) {
		t.Fatal("dry-run changed a task")
	}

	repoID = "other"
	wrong, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if !errors.Is(err, domain.ErrConflict) || wrong.Committed {
		t.Fatalf("wrong identity result=%+v err=%v", wrong, err)
	}
	repoID = ""
	if _, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "config migrate") {
		t.Fatalf("missing identity error=%v", err)
	}
	repoID = "planning"
	wrongRoot := testutil.Must(NewFS(root, core.UnrestrictedMutations(), WithPlanningIdentityReader(func() (string, string, error) {
		return t.TempDir(), repoID, nil
	})))
	if _, err := applyStoredThreadPlan(wrongRoot, plan, false); !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "guarded root") {
		t.Fatalf("wrong root error=%v", err)
	}
}

func TestThreadApplyDetectsIdentityChangeBeforeFirstWrite(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	taskID := testutil.TaskID("identity-race-member")
	threadID := testutil.TaskID("identity-race-thread")
	writeGraphMutationTask(t, root, "identity-race-member", domain.StatusNextUp, nil, "")
	plan := storeThreadApplyPlan(threadID, []string{taskID})
	original := testHookBeforeThreadApplyVerify
	t.Cleanup(func() { testHookBeforeThreadApplyVerify = original })
	testHookBeforeThreadApplyVerify = func() {
		testHookBeforeThreadApplyVerify = nil
		repoID = "changed"
	}
	result, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if !errors.Is(err, domain.ErrConflict) || result.Committed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestThreadApplyWholeSourceCASRejectsRawTaskEditBeforeFirstWrite(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	gateID := testutil.TaskID("whole-cas-gate")
	memberID := testutil.TaskID("whole-cas-member")
	unrelatedPath := writeGraphMutationTask(t, root, "whole-cas-unrelated", domain.StatusNextUp, nil, "")
	writeGraphMutationTask(t, root, "whole-cas-gate", domain.StatusCompleted, nil, "")
	writeGraphMutationTask(t, root, "whole-cas-member", domain.StatusNextUp, nil, "")
	plan := storeThreadApplyPlan(testutil.TaskID("whole-cas-thread"), []string{memberID},
		core.ThreadApplyDependency{From: gateID, To: memberID})
	original := testHookBeforeThreadApplyVerify
	t.Cleanup(func() { testHookBeforeThreadApplyVerify = original })
	testHookBeforeThreadApplyVerify = func() {
		testHookBeforeThreadApplyVerify = nil
		content, err := os.ReadFile(unrelatedPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(unrelatedPath, append(content, []byte("\n<!-- raw whole-source edit -->\n")...), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	result, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if !errors.Is(err, domain.ErrConflict) || result.Committed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	graph, loadErr := core.LoadTaskGraph(testutil.Must(NewFS(root, core.ReadOnlyMutations())))
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	memberTask, _ := graph.Task(memberID)
	if slices.Contains(memberTask.DependsOn, gateID) {
		t.Fatal("dependency landed despite whole-source task CAS conflict")
	}
}

func TestThreadApplyWholeSourceCASRejectsRawThreadEditBeforeFirstWrite(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	memberID := testutil.TaskID("whole-thread-cas-member")
	peerID := testutil.TaskID("whole-thread-cas-peer")
	writeGraphMutationTask(t, root, "whole-thread-cas-member", domain.StatusNextUp, nil, "")
	plan := storeThreadApplyPlan(testutil.TaskID("whole-thread-cas-target"), []string{memberID})
	original := testHookBeforeThreadApplyVerify
	t.Cleanup(func() { testHookBeforeThreadApplyVerify = original })
	testHookBeforeThreadApplyVerify = func() {
		testHookBeforeThreadApplyVerify = nil
		testutil.Write(t, filepath.Join(root, domain.ThreadsDir, peerID+"-peer.md"), "---\n"+
			"schema: 1\nid: "+peerID+"\nstatus: unstarted\ndescription: Concurrent Thread\n"+
			"goal: Change the guarded snapshot\ncreated: \"2026-08-30\"\ntasks: []\n---\n# Peer\n")
	}
	result, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if !errors.Is(err, domain.ErrConflict) || result.Committed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if _, _, err := testutil.Must(NewFS(root, core.ReadOnlyMutations())).GetThread(plan.Thread.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("planned Thread landed despite whole-source Thread CAS conflict: %v", err)
	}
}

func TestThreadApplyFailureAfterFinalCreateReportsCompleteCommit(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	taskID := testutil.TaskID("final-member")
	threadID := testutil.TaskID("final-thread")
	writeGraphMutationTask(t, root, "final-member", domain.StatusNextUp, nil, "")
	plan := storeThreadApplyPlan(threadID, []string{taskID})
	original := testHookAfterThreadApplyWrite
	t.Cleanup(func() { testHookAfterThreadApplyWrite = original })
	testHookAfterThreadApplyWrite = func(kind, _ string) error {
		if kind == "thread" {
			return errors.New("cleanup failed")
		}
		return nil
	}
	result, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if err == nil || !result.Complete || !result.Committed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if _, _, readErr := testutil.Must(NewFS(root, core.ReadOnlyMutations())).GetThread(threadID); readErr != nil {
		t.Fatalf("final Thread was not durable: %v", readErr)
	}
}

func TestThreadApplyExactConcurrentThreadCreateReportsUnchangedSkip(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	memberID := testutil.TaskID("exact-race-member")
	writeGraphMutationTask(t, root, "exact-race-member", domain.StatusNextUp, nil, "")
	plan := storeThreadApplyPlan(testutil.TaskID("exact-race-thread"), []string{memberID})
	fs := threadApplyStore(root, &repoID)
	graph, err := core.LoadTaskGraph(fs)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := core.PrepareThreadApply(core.ThreadApplySnapshot{
		PlanningRepoID: repoID, Graph: graph,
	}, plan)
	if err != nil {
		t.Fatal(err)
	}
	materialized, err := fs.materializeThreadCreation(*decision.ThreadPlan)
	if err != nil {
		t.Fatal(err)
	}
	original := testHookBeforeThreadApplyWrite
	t.Cleanup(func() { testHookBeforeThreadApplyWrite = original })
	testHookBeforeThreadApplyWrite = func(kind, _ string) {
		if kind != "thread" {
			return
		}
		testHookBeforeThreadApplyWrite = nil
		if err := os.MkdirAll(filepath.Dir(materialized.path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(materialized.path, materialized.content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	result, err := applyStoredThreadPlan(fs, plan, false)
	if err != nil || result.Changed || result.Committed || !result.Complete {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if len(result.Operations) != 1 || result.Operations[0].State != core.ThreadApplySkipped {
		t.Fatalf("operations=%+v", result.Operations)
	}
}

func TestThreadApplyImmediateTargetCASRejectsRawEditAndRetryPreservesIt(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	gateID := testutil.TaskID("raw-cas-gate")
	memberID := testutil.TaskID("raw-cas-member")
	threadID := testutil.TaskID("raw-cas-thread")
	writeGraphMutationTask(t, root, "raw-cas-gate", domain.StatusCompleted, nil, "")
	memberPath := writeGraphMutationTask(t, root, "raw-cas-member", domain.StatusNextUp, nil, "")
	plan := storeThreadApplyPlan(threadID, []string{memberID}, core.ThreadApplyDependency{From: gateID, To: memberID})
	original := testHookBeforeThreadApplyWrite
	t.Cleanup(func() { testHookBeforeThreadApplyWrite = original })
	testHookBeforeThreadApplyWrite = func(kind, id string) {
		if kind != "dependency" || id != memberID {
			return
		}
		testHookBeforeThreadApplyWrite = nil
		content, err := os.ReadFile(memberPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(memberPath, append(content, []byte("\n<!-- raw edit -->\n")...), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	result, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if !errors.Is(err, domain.ErrConflict) || result.Committed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	resumed, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if err != nil || !resumed.Complete {
		t.Fatalf("resumed=%+v err=%v", resumed, err)
	}
	content, _ := os.ReadFile(memberPath)
	if !strings.Contains(string(content), "<!-- raw edit -->") {
		t.Fatal("retry erased unrelated raw task content")
	}
}

func TestThreadApplySerializesWithDirectDependencyMutation(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	aID := testutil.TaskID("serialize-a")
	bID := testutil.TaskID("serialize-b")
	cID := testutil.TaskID("serialize-c")
	threadID := testutil.TaskID("serialize-thread")
	writeGraphMutationTask(t, root, "serialize-a", domain.StatusCompleted, nil, "")
	writeGraphMutationTask(t, root, "serialize-b", domain.StatusNextUp, nil, "")
	writeGraphMutationTask(t, root, "serialize-c", domain.StatusNextUp, nil, "")
	plan := storeThreadApplyPlan(threadID, []string{bID}, core.ThreadApplyDependency{From: aID, To: bID})

	original := testHookBeforeThreadApplyVerify
	t.Cleanup(func() { testHookBeforeThreadApplyVerify = original })
	applyHoldingGuard := make(chan struct{})
	releaseApply := make(chan struct{})
	testHookBeforeThreadApplyVerify = func() {
		testHookBeforeThreadApplyVerify = nil
		close(applyHoldingGuard)
		<-releaseApply
	}
	type applyOutcome struct {
		result core.ThreadApplyMutationResult
		err    error
	}
	applyDone := make(chan applyOutcome, 1)
	go func() {
		result, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
		applyDone <- applyOutcome{result: result, err: err}
	}()
	<-applyHoldingGuard

	directDone := make(chan error, 1)
	go func() {
		_, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).MutateTaskGraph(threadApplyStoreNow, false, addDependencyPlan(cID, bID))
		directDone <- err
	}()
	select {
	case err := <-directDone:
		t.Fatalf("direct mutation bypassed compound guard: %v", err)
	case <-time.After(75 * time.Millisecond):
	}
	close(releaseApply)
	bulk := <-applyDone
	if bulk.err != nil || !bulk.result.Complete {
		t.Fatalf("bulk=%+v err=%v", bulk.result, bulk.err)
	}
	if err := <-directDone; err != nil {
		t.Fatal(err)
	}
	graph, err := core.LoadTaskGraph(testutil.Must(NewFS(root, core.UnrestrictedMutations())))
	if err != nil || graph.Health() != core.GraphHealthy {
		t.Fatalf("graph health=%s err=%v", graph.Health(), err)
	}
	b, _ := graph.Task(bID)
	c, _ := graph.Task(cID)
	if !slices.Contains(b.DependsOn, aID) || !slices.Contains(c.DependsOn, bID) {
		t.Fatalf("serialized graph b=%v c=%v", b.DependsOn, c.DependsOn)
	}
}

func TestThreadApplyRawEditAfterDurablePrefixIsReportedAndResumable(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	aID := testutil.TaskID("raw-prefix-a")
	bID := testutil.TaskID("raw-prefix-b")
	cID := testutil.TaskID("raw-prefix-c")
	threadID := testutil.TaskID("raw-prefix-thread")
	writeGraphMutationTask(t, root, "raw-prefix-a", domain.StatusCompleted, nil, "")
	writeGraphMutationTask(t, root, "raw-prefix-b", domain.StatusNextUp, nil, "")
	writeGraphMutationTask(t, root, "raw-prefix-c", domain.StatusNextUp, nil, "")
	plan := storeThreadApplyPlan(threadID, []string{bID, cID},
		core.ThreadApplyDependency{From: aID, To: bID},
		core.ThreadApplyDependency{From: bID, To: cID},
	)
	original := testHookBeforeThreadApplyWrite
	t.Cleanup(func() { testHookBeforeThreadApplyWrite = original })
	dependencyWrites := 0
	testHookBeforeThreadApplyWrite = func(kind, id string) {
		if kind != "dependency" {
			return
		}
		dependencyWrites++
		if dependencyWrites != 2 {
			return
		}
		testHookBeforeThreadApplyWrite = nil
		path, err := testutil.Must(NewFS(root, core.ReadOnlyMutations())).resolvePath(id)
		if err != nil {
			t.Fatal(err)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(content, []byte("\n<!-- raw after prefix -->\n")...), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	interrupted, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if !errors.Is(err, domain.ErrConflict) || !interrupted.Committed || interrupted.Complete {
		t.Fatalf("interrupted=%+v err=%v", interrupted, err)
	}
	resumed, err := applyStoredThreadPlan(threadApplyStore(root, &repoID), plan, false)
	if err != nil || !resumed.Complete {
		t.Fatalf("resumed=%+v err=%v", resumed, err)
	}
}

func TestThreadApplyReverifiesRepairedEdgesWhenThreadAlreadyExists(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	gateID := testutil.TaskID("existing-thread-gate")
	memberID := testutil.TaskID("existing-thread-member")
	threadID := testutil.TaskID("existing-thread-thread")
	writeGraphMutationTask(t, root, "existing-thread-gate", domain.StatusCompleted, nil, "")
	memberPath := writeGraphMutationTask(t, root, "existing-thread-member", domain.StatusNextUp, nil, "")
	withoutDependency, err := os.ReadFile(memberPath)
	if err != nil {
		t.Fatal(err)
	}
	plan := storeThreadApplyPlan(threadID, []string{memberID}, core.ThreadApplyDependency{From: gateID, To: memberID})
	fs := threadApplyStore(root, &repoID)
	if initial, applyErr := applyStoredThreadPlan(fs, plan, false); applyErr != nil || !initial.Complete {
		t.Fatalf("initial=%+v err=%v", initial, applyErr)
	}

	// Simulate an out-of-band edit that removed an intended edge after the first
	// successful apply, then race the repair by removing it again immediately
	// before final convergence verification.
	if err := os.WriteFile(memberPath, withoutDependency, 0o644); err != nil {
		t.Fatal(err)
	}
	original := testHookBeforeThreadApplyWrite
	t.Cleanup(func() { testHookBeforeThreadApplyWrite = original })
	testHookBeforeThreadApplyWrite = func(kind, _ string) {
		if kind != "thread" {
			return
		}
		testHookBeforeThreadApplyWrite = nil
		if err := os.WriteFile(memberPath, withoutDependency, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	interrupted, err := applyStoredThreadPlan(fs, plan, false)
	if !errors.Is(err, domain.ErrConflict) || !interrupted.Committed || interrupted.Complete {
		t.Fatalf("interrupted=%+v err=%v", interrupted, err)
	}
	if interrupted.Operations[0].State != core.ThreadApplyPending || interrupted.Operations[1].State != core.ThreadApplySkipped {
		t.Fatalf("recovery operations=%+v", interrupted.Operations)
	}

	resumed, err := applyStoredThreadPlan(fs, plan, false)
	if err != nil || !resumed.Complete || !resumed.Committed {
		t.Fatalf("resumed=%+v err=%v", resumed, err)
	}
}

func TestThreadApplyPlannerCannotReenterStore(t *testing.T) {
	root := t.TempDir()
	repoID := "planning"
	fs := threadApplyStore(root, &repoID)
	_, err := fs.MutateThreadApply(threadApplyStoreNow, true, func(core.ThreadApplySnapshot) (core.ThreadApplyPlan, error) {
		_, nestedErr := fs.ReadThreads()
		return core.ThreadApplyPlan{}, nestedErr
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("planner re-entry error = %v", err)
	}
}
