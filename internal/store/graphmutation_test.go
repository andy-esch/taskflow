package store

import (
	"errors"
	"fmt"
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

var graphMutationNow = time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)

func writeGraphMutationTask(t *testing.T, root, seed string, status domain.Status, dependencies []string, extra string) string {
	t.Helper()
	taskID := testutil.TaskID(seed)
	dir := filepath.Join(root, domain.TasksDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	dependencyLine := ""
	if len(dependencies) > 0 {
		dependencyLine = "depends_on: [" + strings.Join(dependencies, ", ") + "]\n"
	}
	content := fmt.Sprintf("---\nschema: 1\nid: %s\nstatus: %s\nepic: 30\ndescription: %s\neffort: 1h\ntier: 1\npriority: high\nautonomy_level: 2\ntags: [graph]\ncreated: \"2026-08-27\"\n%s%s---\n# %s\n\nBody stays intact.\n",
		taskID, status, seed, dependencyLine, extra, seed)
	path := filepath.Join(dir, taskID+"-"+seed+".md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func addDependencyPlan(dependent, prerequisite string) core.TaskGraphPlanner {
	return func(graph *core.TaskGraph) (core.TaskGraphMutationPlan, error) {
		task, ok := graph.Task(dependent)
		if !ok {
			return core.TaskGraphMutationPlan{}, fmt.Errorf("missing dependent %s", dependent)
		}
		dependencies := append([]string(nil), task.DependsOn...)
		if !slices.Contains(dependencies, prerequisite) {
			dependencies = append(dependencies, prerequisite)
		}
		return core.TaskGraphMutationPlan{TaskWrites: []core.TaskDependencyWrite{{
			TaskID: dependent, DependsOn: dependencies,
		}}}, nil
	}
}

func TestMutateTaskGraphOwnsSemanticReadValidateWriteBoundary(t *testing.T) {
	root := t.TempDir()
	aID, bID, cID := testutil.TaskID("alpha"), testutil.TaskID("beta"), testutil.TaskID("charlie")
	writeGraphMutationTask(t, root, "alpha", domain.StatusCompleted, nil, "")
	bPath := writeGraphMutationTask(t, root, "beta", domain.StatusReadyToStart, nil, "custom_key: keep-me # preserve this comment\n")
	writeGraphMutationTask(t, root, "charlie", domain.StatusCompleted, nil, "")
	fs := NewFS(root)

	result, err := fs.MutateTaskGraph(graphMutationNow, false, func(graph *core.TaskGraph) (core.TaskGraphMutationPlan, error) {
		if graph.Health() != core.GraphHealthy || len(graph.TaskIDs()) != 3 {
			t.Fatalf("planner snapshot health=%s tasks=%v", graph.Health(), graph.TaskIDs())
		}
		if task, _ := graph.Task(bID); task.SourceVersion != "" {
			t.Fatal("planner-facing task exposed its persistence version")
		}
		return core.TaskGraphMutationPlan{TaskWrites: []core.TaskDependencyWrite{{
			TaskID: bID, DependsOn: []string{cID, aID},
		}}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(result.AppliedTaskIDs, []string{bID}) || result.DryRun {
		t.Fatalf("result = %+v", result)
	}
	if got := result.Plan.TaskWrites[0].DependsOn; !slices.IsSorted(got) {
		t.Fatalf("normalized dependency set is not sorted: %v", got)
	}
	b, err := os.ReadFile(bPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, preserved := range []string{"custom_key: keep-me # preserve this comment", "Body stays intact.", "updated_at: \"2026-08-27\""} {
		if !strings.Contains(string(b), preserved) {
			t.Errorf("materialized task lost %q:\n%s", preserved, b)
		}
	}
	graph, err := core.LoadTaskGraph(fs)
	if err != nil {
		t.Fatal(err)
	}
	updated, _ := graph.Task(bID)
	if graph.Health() != core.GraphHealthy || !slices.Equal(updated.DependsOn, result.Plan.TaskWrites[0].DependsOn) {
		t.Fatalf("reloaded graph health=%s task=%+v", graph.Health(), updated)
	}
}

func TestMutateTaskGraphDryRunValidatesWithoutWriting(t *testing.T) {
	root := t.TempDir()
	aID, bID := testutil.TaskID("alpha"), testutil.TaskID("beta")
	writeGraphMutationTask(t, root, "alpha", domain.StatusCompleted, nil, "")
	bPath := writeGraphMutationTask(t, root, "beta", domain.StatusReadyToStart, nil, "")
	before, _ := os.ReadFile(bPath)

	result, err := NewFS(root).MutateTaskGraph(graphMutationNow, true, addDependencyPlan(bID, aID))
	if err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(bPath)
	if !result.DryRun || len(result.AppliedTaskIDs) != 0 || len(result.Plan.TaskWrites) != 1 {
		t.Fatalf("dry-run result = %+v", result)
	}
	if !slices.Equal(before, after) {
		t.Fatal("dry-run changed the task file")
	}
}

func TestMaterializeTaskGraphPlanReportsTaskAbsentFromSnapshotAsNotFound(t *testing.T) {
	root := t.TempDir()
	writeGraphMutationTask(t, root, "alpha", domain.StatusReadyToStart, nil, "")
	graph, err := core.LoadTaskGraph(NewFS(root))
	if err != nil {
		t.Fatal(err)
	}

	// Make the planned task resolvable only after the snapshot. The materializer
	// must attribute the stale plan to the absent snapshot entity, not misreport
	// its empty snapshot path as a concurrent path change.
	newcomerID := testutil.TaskID("newcomer")
	writeGraphMutationTask(t, root, "newcomer", domain.StatusReadyToStart, nil, "")
	_, err = NewFS(root).materializeTaskGraphPlan(graph, core.TaskGraphMutationPlan{
		TaskWrites: []core.TaskDependencyWrite{{TaskID: newcomerID}},
	}, graphMutationNow)
	if !errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrConflict) {
		t.Fatalf("absent snapshot task error = %v, want ErrNotFound only", err)
	}
}

func TestMutateTaskGraphConcurrentOppositeEdgesCannotCommitCycle(t *testing.T) {
	root := t.TempDir()
	aID, bID := testutil.TaskID("alpha"), testutil.TaskID("beta")
	writeGraphMutationTask(t, root, "alpha", domain.StatusReadyToStart, nil, "")
	writeGraphMutationTask(t, root, "beta", domain.StatusReadyToStart, nil, "")

	planners := []core.TaskGraphPlanner{addDependencyPlan(aID, bID), addDependencyPlan(bID, aID)}
	start := make(chan struct{})
	errs := make(chan error, len(planners))
	var ready sync.WaitGroup
	ready.Add(len(planners))
	for _, planner := range planners {
		planner := planner
		go func() {
			ready.Done()
			<-start
			_, err := NewFS(root).MutateTaskGraph(graphMutationNow, false, planner)
			errs <- err
		}()
	}
	ready.Wait()
	close(start)

	succeeded, rejected := 0, 0
	for range planners {
		if err := <-errs; err == nil {
			succeeded++
		} else if errors.Is(err, domain.ErrValidation) && strings.Contains(err.Error(), "dependency cycle:") {
			rejected++
		} else {
			t.Fatalf("unexpected concurrent mutation result: %v", err)
		}
	}
	if succeeded != 1 || rejected != 1 {
		t.Fatalf("succeeded=%d rejected=%d", succeeded, rejected)
	}
	graph, err := core.LoadTaskGraph(NewFS(root))
	if err != nil {
		t.Fatal(err)
	}
	if graph.Health() != core.GraphHealthy {
		t.Fatalf("final graph health = %s, problems=%+v", graph.Health(), graph.Problems())
	}
	committed := 0
	for _, taskID := range []string{aID, bID} {
		task, _ := graph.Task(taskID)
		committed += len(task.DependsOn)
	}
	if committed != 1 {
		t.Fatalf("committed edge count = %d", committed)
	}
}

func TestMutateTaskGraphAllowsOneGuardedLegacyMigrationToHealthy(t *testing.T) {
	root := t.TempDir()
	aID, bID := testutil.TaskID("alpha"), testutil.TaskID("beta")
	writeGraphMutationTask(t, root, "alpha", domain.StatusCompleted, nil, "")
	writeGraphMutationTask(t, root, "beta", domain.StatusReadyToStart, nil, "blocked_by: [alpha]\n")
	fs := NewFS(root)

	_, err := fs.MutateTaskGraph(graphMutationNow, false, func(graph *core.TaskGraph) (core.TaskGraphMutationPlan, error) {
		if graph.Health() != core.GraphDegraded {
			t.Fatalf("migration planner health = %s", graph.Health())
		}
		return core.TaskGraphMutationPlan{TaskWrites: []core.TaskDependencyWrite{{
			TaskID: bID, DependsOn: []string{aID}, ClearLegacy: true,
		}}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	graph, err := core.LoadTaskGraph(fs)
	if err != nil {
		t.Fatal(err)
	}
	if graph.Health() != core.GraphHealthy || len(graph.LegacyDiagnostics()) != 0 {
		t.Fatalf("migrated health=%s legacy=%+v", graph.Health(), graph.LegacyDiagnostics())
	}
}

func TestMutateTaskGraphBrokenSnapshotFailsBeforePlanner(t *testing.T) {
	root := t.TempDir()
	aID, bID := testutil.TaskID("alpha"), testutil.TaskID("beta")
	writeGraphMutationTask(t, root, "alpha", domain.StatusReadyToStart, []string{bID}, "")
	writeGraphMutationTask(t, root, "beta", domain.StatusReadyToStart, []string{aID}, "")
	called := false
	_, err := NewFS(root).MutateTaskGraph(graphMutationNow, false, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
		called = true
		return core.TaskGraphMutationPlan{}, nil
	})
	if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "repository task graph is broken") || !strings.Contains(err.Error(), "dependency cycle:") {
		t.Fatalf("broken mutation error = %v", err)
	}
	if called {
		t.Fatal("planner ran against a broken authoritative snapshot")
	}
}

func TestMutateTaskGraphRejectsPlannerStoreCallsAndNestedMutationWithoutHanging(t *testing.T) {
	root := t.TempDir()
	aID := testutil.TaskID("alpha")
	writeGraphMutationTask(t, root, "alpha", domain.StatusReadyToStart, nil, "")
	fs := NewFS(root)
	done := make(chan error, 1)
	go func() {
		_, err := fs.MutateTaskGraph(graphMutationNow, false, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
			if _, _, err := fs.GetTask(aID); !errors.Is(err, domain.ErrConflict) {
				return core.TaskGraphMutationPlan{}, fmt.Errorf("nested read error = %v", err)
			}
			if _, err := fs.SetFields(aID, map[string]any{"priority": "low"}, false); !errors.Is(err, domain.ErrConflict) {
				return core.TaskGraphMutationPlan{}, fmt.Errorf("nested write error = %v", err)
			}
			if _, err := fs.MutateTaskGraph(graphMutationNow, false, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
				return core.TaskGraphMutationPlan{}, nil
			}); !errors.Is(err, domain.ErrConflict) {
				return core.TaskGraphMutationPlan{}, fmt.Errorf("nested graph mutation error = %v", err)
			}
			second := NewFS(root)
			if _, _, err := second.GetTask(aID); !errors.Is(err, domain.ErrConflict) {
				return core.TaskGraphMutationPlan{}, fmt.Errorf("second-store nested read error = %v", err)
			}
			if _, err := second.SetFields(aID, map[string]any{"priority": "low"}, false); !errors.Is(err, domain.ErrConflict) {
				return core.TaskGraphMutationPlan{}, fmt.Errorf("second-store nested write error = %v", err)
			}
			if _, err := second.DanglingLinks(); !errors.Is(err, domain.ErrConflict) {
				return core.TaskGraphMutationPlan{}, fmt.Errorf("nested dangling-links error = %v", err)
			}
			if _, err := second.FixFrontmatter(true); !errors.Is(err, domain.ErrConflict) {
				return core.TaskGraphMutationPlan{}, fmt.Errorf("nested frontmatter-fix error = %v", err)
			}
			return core.TaskGraphMutationPlan{}, nil
		})
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("nested Store call self-deadlocked")
	}
	if _, _, err := fs.GetTask(aID); err != nil {
		t.Fatalf("planner sentinel leaked after return: %v", err)
	}
}

func TestMutateTaskGraphRejectsConcurrentStoreAccessAtRepositoryScope(t *testing.T) {
	root := t.TempDir()
	aID := testutil.TaskID("alpha")
	writeGraphMutationTask(t, root, "alpha", domain.StatusReadyToStart, nil, "")
	fs := NewFS(root)
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := fs.MutateTaskGraph(graphMutationNow, false, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
			close(started)
			<-release
			return core.TaskGraphMutationPlan{}, nil
		})
		done <- err
	}()
	<-started
	_, _, err := NewFS(root).GetTask(aID)
	if !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "concurrent caller") {
		t.Fatalf("concurrent Store access error = %v", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestMutateTaskGraphPlannerPanicReleasesGuard(t *testing.T) {
	root := t.TempDir()
	aID := testutil.TaskID("alpha")
	writeGraphMutationTask(t, root, "alpha", domain.StatusReadyToStart, nil, "")
	fs := NewFS(root)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("planner panic did not propagate")
			}
		}()
		_, _ = fs.MutateTaskGraph(graphMutationNow, false, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
			panic("planner failed")
		})
	}()
	if _, err := fs.SetFields(aID, map[string]any{"priority": "low"}, false); err != nil {
		t.Fatalf("repository guard remained held after panic: %v", err)
	}
}

func TestMutateTaskGraphAttributesReleaseFailureAfterUnlocking(t *testing.T) {
	root := t.TempDir()
	writeGraphMutationTask(t, root, "alpha", domain.StatusReadyToStart, nil, "")
	fs := NewFS(root)
	original := testHookRepositoryUnlockError
	defer func() { testHookRepositoryUnlockError = original }()
	testHookRepositoryUnlockError = func() error { return errors.New("injected unlock failure") }
	_, err := fs.MutateTaskGraph(graphMutationNow, false, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
		return core.TaskGraphMutationPlan{}, nil
	})
	if err == nil || !strings.Contains(err.Error(), "release repository graph mutation guard") || !strings.Contains(err.Error(), "injected unlock failure") {
		t.Fatalf("release error = %v", err)
	}
	testHookRepositoryUnlockError = nil
	if _, err := fs.MutateTaskGraph(graphMutationNow, false, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
		return core.TaskGraphMutationPlan{}, nil
	}); err != nil {
		t.Fatalf("guard stayed held after attributed release failure: %v", err)
	}
}

func TestMutateTaskGraphCASRejectsRawEditBeforeApply(t *testing.T) {
	root := t.TempDir()
	aID, bID := testutil.TaskID("alpha"), testutil.TaskID("beta")
	aPath := writeGraphMutationTask(t, root, "alpha", domain.StatusCompleted, nil, "")
	bPath := writeGraphMutationTask(t, root, "beta", domain.StatusReadyToStart, nil, "")
	fs := NewFS(root)
	original := testHookBeforeGraphVerify
	defer func() { testHookBeforeGraphVerify = original }()
	testHookBeforeGraphVerify = func() {
		// Edit a task the plan will NOT write. Per-file CAS on beta cannot see
		// this; the whole-snapshot CAS must still reject the stale plan.
		content, _ := os.ReadFile(aPath)
		_ = os.WriteFile(aPath, append(content, []byte("\nRaw edit survives.\n")...), 0o644)
		testHookBeforeGraphVerify = nil
	}
	_, err := fs.MutateTaskGraph(graphMutationNow, false, addDependencyPlan(bID, aID))
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("raw edit error = %v", err)
	}
	a, _ := os.ReadFile(aPath)
	b, _ := os.ReadFile(bPath)
	if !strings.Contains(string(a), "Raw edit survives.") || strings.Contains(string(b), "depends_on:") {
		t.Fatalf("stale graph mutation wrote after an unrelated raw edit:\nalpha:\n%s\nbeta:\n%s", a, b)
	}
}

func TestMutateTaskGraphReturnsDurablePrefixAndRerunConverges(t *testing.T) {
	root := t.TempDir()
	aID, bID, cID := testutil.TaskID("alpha"), testutil.TaskID("beta"), testutil.TaskID("charlie")
	writeGraphMutationTask(t, root, "alpha", domain.StatusCompleted, nil, "")
	writeGraphMutationTask(t, root, "beta", domain.StatusReadyToStart, nil, "")
	writeGraphMutationTask(t, root, "charlie", domain.StatusReadyToStart, nil, "")
	fs := NewFS(root)
	planner := func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
		return core.TaskGraphMutationPlan{TaskWrites: []core.TaskDependencyWrite{
			{TaskID: bID, DependsOn: []string{aID}},
			{TaskID: cID, DependsOn: []string{aID}},
		}}, nil
	}
	original := testHookAfterGraphWrite
	defer func() { testHookAfterGraphWrite = original }()
	testHookAfterGraphWrite = func(string) error {
		testHookAfterGraphWrite = nil
		return errors.New("injected interruption")
	}
	partial, err := fs.MutateTaskGraph(graphMutationNow, false, planner)
	if err == nil || len(partial.AppliedTaskIDs) != 1 {
		t.Fatalf("partial result=%+v err=%v", partial, err)
	}
	graph, loadErr := core.LoadTaskGraph(fs)
	if loadErr != nil || graph.Health() != core.GraphHealthy {
		t.Fatalf("prefix left invalid graph: health=%s err=%v", graph.Health(), loadErr)
	}
	completed, err := fs.MutateTaskGraph(graphMutationNow, false, planner)
	if err != nil {
		t.Fatal(err)
	}
	if len(completed.AppliedTaskIDs) != 1 {
		t.Fatalf("rerun should skip durable prefix and apply one remainder: %+v", completed)
	}
	graph, _ = core.LoadTaskGraph(fs)
	for _, taskID := range []string{bID, cID} {
		task, _ := graph.Task(taskID)
		if !slices.Equal(task.DependsOn, []string{aID}) {
			t.Fatalf("task %s dependencies = %v", taskID, task.DependsOn)
		}
	}
}

func TestMutateTaskGraphPerFileCASPreservesRawEditAfterDurablePrefix(t *testing.T) {
	root := t.TempDir()
	aID, bID, cID := testutil.TaskID("alpha"), testutil.TaskID("beta"), testutil.TaskID("charlie")
	writeGraphMutationTask(t, root, "alpha", domain.StatusCompleted, nil, "")
	writeGraphMutationTask(t, root, "beta", domain.StatusReadyToStart, nil, "")
	cPath := writeGraphMutationTask(t, root, "charlie", domain.StatusReadyToStart, nil, "")
	fs := NewFS(root)
	original := testHookBeforeGraphWrite
	defer func() { testHookBeforeGraphWrite = original }()
	testHookBeforeGraphWrite = func(taskID string) {
		if taskID != cID {
			return
		}
		content, _ := os.ReadFile(cPath)
		_ = os.WriteFile(cPath, append(content, []byte("\nRaw edit survives.\n")...), 0o644)
		testHookBeforeGraphWrite = nil
	}
	result, err := fs.MutateTaskGraph(graphMutationNow, false, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
		return core.TaskGraphMutationPlan{TaskWrites: []core.TaskDependencyWrite{
			{TaskID: bID, DependsOn: []string{aID}},
			{TaskID: cID, DependsOn: []string{aID}},
		}}, nil
	})
	if !errors.Is(err, domain.ErrConflict) || !slices.Equal(result.AppliedTaskIDs, []string{bID}) {
		t.Fatalf("partial CAS result=%+v err=%v", result, err)
	}
	content, _ := os.ReadFile(cPath)
	if !strings.Contains(string(content), "Raw edit survives.") || strings.Contains(string(content), "depends_on:") {
		t.Fatalf("later graph write clobbered raw edit:\n%s", content)
	}
}

func TestMutateTaskGraphRejectsBrokenIntermediatePrefixBeforeWriting(t *testing.T) {
	root := t.TempDir()
	aID, bID := testutil.TaskID("alpha"), testutil.TaskID("beta")
	firstID, secondID := aID, bID
	firstSeed, secondSeed := "alpha", "beta"
	if secondID < firstID {
		firstID, secondID = secondID, firstID
		firstSeed, secondSeed = secondSeed, firstSeed
	}
	firstPath := writeGraphMutationTask(t, root, firstSeed, domain.StatusReadyToStart, nil, "")
	secondPath := writeGraphMutationTask(t, root, secondSeed, domain.StatusReadyToStart, []string{firstID}, "")
	firstBefore, _ := os.ReadFile(firstPath)
	secondBefore, _ := os.ReadFile(secondPath)

	_, err := NewFS(root).MutateTaskGraph(graphMutationNow, false, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
		// Final state reverses the edge and is acyclic. Task-ID ordering would add
		// the reverse edge before removing the old one, however, so a crash after
		// the first replacement would leave a cycle.
		return core.TaskGraphMutationPlan{TaskWrites: []core.TaskDependencyWrite{
			{TaskID: firstID, DependsOn: []string{secondID}},
			{TaskID: secondID, DependsOn: nil},
		}}, nil
	})
	if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "write prefix") || !strings.Contains(err.Error(), "dependency cycle:") {
		t.Fatalf("unsafe prefix error = %v", err)
	}
	firstAfter, _ := os.ReadFile(firstPath)
	secondAfter, _ := os.ReadFile(secondPath)
	if !slices.Equal(firstBefore, firstAfter) || !slices.Equal(secondBefore, secondAfter) {
		t.Fatal("unsafe multi-file plan wrote before prefix validation completed")
	}
}

func TestMutateTaskGraphPreservesPrefixSafePlannerOrder(t *testing.T) {
	root := t.TempDir()
	aID, bID := testutil.TaskID("alpha"), testutil.TaskID("beta")
	aPath := writeGraphMutationTask(t, root, "alpha", domain.StatusReadyToStart, nil, "")
	bPath := writeGraphMutationTask(t, root, "beta", domain.StatusReadyToStart, []string{aID}, "")

	result, err := NewFS(root).MutateTaskGraph(graphMutationNow, false, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
		// Remove the old edge before adding its reverse. Stable-ID sorting is unsafe
		// here; the planner-provided sequence is part of the recovery contract.
		return core.TaskGraphMutationPlan{TaskWrites: []core.TaskDependencyWrite{
			{TaskID: bID, DependsOn: nil},
			{TaskID: aID, DependsOn: []string{bID}},
		}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(result.AppliedTaskIDs, []string{bID, aID}) {
		t.Fatalf("applied order = %v", result.AppliedTaskIDs)
	}
	a, _ := os.ReadFile(aPath)
	b, _ := os.ReadFile(bPath)
	if !strings.Contains(string(a), "depends_on: ["+bID+"]") || strings.Contains(string(b), "depends_on:") {
		t.Fatalf("edge reversal did not materialize safely:\nalpha:\n%s\nbeta:\n%s", a, b)
	}
}

func TestMutateTaskGraphStampsUpdatedAtOnlyForSemanticChanges(t *testing.T) {
	root := t.TempDir()
	aID, bID := testutil.TaskID("alpha"), testutil.TaskID("beta")
	writeGraphMutationTask(t, root, "alpha", domain.StatusCompleted, nil, "")
	bPath := writeGraphMutationTask(t, root, "beta", domain.StatusReadyToStart, nil, "updated_at: \"2026-01-01\"\n")
	fs := NewFS(root)

	first, err := fs.MutateTaskGraph(graphMutationNow, false, addDependencyPlan(bID, aID))
	if err != nil || !slices.Equal(first.AppliedTaskIDs, []string{bID}) {
		t.Fatalf("first mutation=%+v err=%v", first, err)
	}
	content, _ := os.ReadFile(bPath)
	if !strings.Contains(string(content), "updated_at: \"2026-08-27\"") {
		t.Fatalf("dependency mutation did not stamp updated_at:\n%s", content)
	}
	second, err := fs.MutateTaskGraph(graphMutationNow.AddDate(0, 0, 1), false, addDependencyPlan(bID, aID))
	if err != nil || len(second.AppliedTaskIDs) != 0 {
		t.Fatalf("idempotent rerun=%+v err=%v", second, err)
	}
	content, _ = os.ReadFile(bPath)
	if strings.Contains(string(content), "2026-08-28") {
		t.Fatalf("no-op dependency mutation changed updated_at:\n%s", content)
	}
}
