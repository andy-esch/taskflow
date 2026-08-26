package store

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/threadspike"
)

func spikeTaskFile(id, title string) string {
	return spikeTaskFileWith(id, title, domain.StatusReadyToStart, nil)
}

func spikeTaskFileWith(id, title string, status domain.Status, dependencies []string) string {
	dependsOn := ""
	if len(dependencies) > 0 {
		dependsOn = "depends_on:\n  - " + strings.Join(dependencies, "\n  - ") + "\n"
	}
	return fmt.Sprintf(`---
schema: 1
id: %s
status: %s
epic: 30-spike
description: Existing prerequisite
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags:
  - threads
created: "2026-08-24"
%s---
# %s
`, id, status, dependsOn, title)
}

func spikeThreadFile(id string, status threadspike.ThreadStatus, tasks []string) string {
	return fmt.Sprintf(`---
schema: 1
id: %s
status: %s
description: Filesystem-backed Thread projection
goal: Exercise the Thread view against real Markdown parsing.
created: "2026-08-24"
tags:
  - threads
tasks:
  - %s
---
# Thread fixture
`, id, status, strings.Join(tasks, "\n  - "))
}

func spikeRepo(t *testing.T, seeds ...string) (*testutil.Repo, map[string]string) {
	t.Helper()
	repo := testutil.NewRepo(t).Epic("30-spike.md", `---
status: active
priority: medium
description: Thread spike
created: "2026-08-24"
tags:
  - threads
---
# Thread spike
`)
	ids := map[string]string{}
	for _, seed := range seeds {
		id := testutil.TaskID(seed)
		ids[seed] = id
		repo.File("tasks/"+id+"-"+seed+".md", spikeTaskFile(id, seed))
	}
	return repo, ids
}

func TestThreadSpikeApplyIsDryRunnableAndConvergesAfterInterruption(t *testing.T) {
	repo, ids := spikeRepo(t, "existing")
	adapter := NewThreadSpikeRepository(repo.Root, "planning-space")
	snapshot, err := adapter.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Problems) != 0 {
		t.Fatalf("snapshot problems = %+v", snapshot.Problems)
	}
	threadID := testutil.TaskID("materialized-thread")
	firstID := testutil.TaskID("materialized-first")
	secondID := testutil.TaskID("materialized-second")
	minted := []string{threadID, firstID, secondID}
	manifest := threadspike.Manifest{
		Thread: threadspike.ThreadInput{
			Title:       "Persist the graph",
			Description: "Prove retryable multi-file graph authoring",
			Goal:        "Every interrupted prefix remains readable and retryable.",
			Tags:        []string{"threads", "spike"},
		},
		Nodes: []threadspike.NodeInput{
			{Key: "existing", TaskID: ids["existing"]},
			{Key: "first", NewTask: &threadspike.NewTaskInput{Title: "First planned task", Epic: "30-spike", Tags: []string{"threads"}}},
			{Key: "second", NewTask: &threadspike.NewTaskInput{Title: "Second planned task", Epic: "30-spike", Tags: []string{"threads"}}},
		},
		Dependencies: []threadspike.Edge{{From: "existing", To: "first"}, {From: "first", To: "second"}},
	}
	plan, err := threadspike.Compose(snapshot, manifest, func() string {
		id := minted[0]
		minted = minted[1:]
		return id
	}, time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	// The exact materialized plan—not the authoring manifest—is the retry token.
	encoded, err := yaml.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	var durablePlan threadspike.MaterializedPlan
	if err := yaml.Unmarshal(encoded, &durablePlan); err != nil {
		t.Fatal(err)
	}
	preview, err := adapter.Apply(durablePlan, threadspike.ApplyOptions{DryRun: true})
	if err != nil || !preview.Complete {
		t.Fatalf("dry-run receipt = %+v, error = %v", preview, err)
	}
	unchanged, err := adapter.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(unchanged.Tasks) != 1 || len(unchanged.Threads) != 0 {
		t.Fatalf("dry-run wrote state: %+v", unchanged)
	}

	stop := errors.New("simulated process interruption")
	writes := 0
	partial, err := adapter.Apply(durablePlan, threadspike.ApplyOptions{AfterWrite: func(threadspike.Operation) error {
		writes++
		if writes == 1 {
			return stop
		}
		return nil
	}})
	if !errors.Is(err, stop) || partial.Complete || writes != 1 {
		t.Fatalf("partial receipt = %+v, writes = %d, error = %v", partial, writes, err)
	}
	landed, err := adapter.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(landed.Problems) != 0 || len(landed.Tasks) != 2 || len(landed.Threads) != 0 {
		t.Fatalf("interrupted prefix is not healthy: %+v", landed)
	}
	if err := threadspike.NewGraph(landed.Tasks).Validate(); err != nil {
		t.Fatalf("interrupted prefix graph: %v", err)
	}

	completed, err := adapter.Apply(durablePlan, threadspike.ApplyOptions{})
	if err != nil || !completed.Complete {
		t.Fatalf("retry receipt = %+v, error = %v", completed, err)
	}
	final, err := adapter.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(final.Problems) != 0 || len(final.Tasks) != 3 || len(final.Threads) != 1 {
		t.Fatalf("final snapshot = %+v", final)
	}
	if !reflect.DeepEqual(final.Tasks[firstID].DependsOn, []string{ids["existing"]}) {
		t.Fatalf("first dependencies = %v", final.Tasks[firstID].DependsOn)
	}
	if !reflect.DeepEqual(final.Tasks[secondID].DependsOn, []string{firstID}) {
		t.Fatalf("second dependencies = %v", final.Tasks[secondID].DependsOn)
	}
	if got := final.Threads[threadID].Tasks; !reflect.DeepEqual(got, plan.Thread.Tasks) {
		t.Fatalf("Thread tasks = %v, want %v", got, plan.Thread.Tasks)
	}

	idempotent, err := adapter.Apply(durablePlan, threadspike.ApplyOptions{})
	if err != nil || !idempotent.Complete || len(idempotent.Entries) != 3 {
		t.Fatalf("idempotent receipt = %+v, error = %v", idempotent, err)
	}
	for _, operation := range idempotent.Entries {
		if operation.Action != "already-applied" {
			t.Fatalf("unexpected retry operation: %+v", operation)
		}
	}
}

func TestThreadSpikeApplyIsBoundToPlanningSpaceAndFailsClosed(t *testing.T) {
	repo, _ := spikeRepo(t, "existing")
	adapter := NewThreadSpikeRepository(repo.Root, "planning-space")
	wrongSpace := threadspike.MaterializedPlan{Schema: threadspike.PlanSchema, RepoID: "another-space"}
	if _, err := adapter.Apply(wrongSpace, threadspike.ApplyOptions{DryRun: true}); err == nil || !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("planning-space mismatch error = %v", err)
	}

	badID := testutil.TaskID("broken-id")
	repo.File("tasks/"+badID+"-broken.md", strings.Replace(spikeTaskFile(badID, "broken"), "id: "+badID, "id: "+testutil.TaskID("different"), 1))
	snapshot, err := adapter.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Problems) == 0 {
		t.Fatal("expected ID drift to make the graph snapshot unhealthy")
	}
	plan := threadspike.MaterializedPlan{Schema: threadspike.PlanSchema, RepoID: "planning-space"}
	if _, err := adapter.Apply(plan, threadspike.ApplyOptions{DryRun: true}); err == nil || !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("unhealthy snapshot error = %v", err)
	}
}

func TestThreadSpikeFilesystemProjectionCoversLifecycleExternalGatesAndValidation(t *testing.T) {
	repo, _ := spikeRepo(t)
	seeds := []string{"external", "blocked-ready", "queued", "forced", "base-done", "downstream-done", "clear-ready"}
	ids := map[string]string{}
	for _, seed := range seeds {
		ids[seed] = testutil.TaskID(seed)
	}
	tasks := []struct {
		seed      string
		status    domain.Status
		dependsOn []string
	}{
		{"external", domain.StatusReadyToStart, nil},
		{"blocked-ready", domain.StatusReadyToStart, []string{ids["external"]}},
		{"queued", domain.StatusNextUp, []string{ids["external"]}},
		{"forced", domain.StatusInProgress, []string{ids["external"]}},
		{"base-done", domain.StatusCompleted, nil},
		{"downstream-done", domain.StatusCompleted, []string{ids["base-done"]}},
		{"clear-ready", domain.StatusReadyToStart, nil},
	}
	for _, task := range tasks {
		repo.File("tasks/"+ids[task.seed]+"-"+task.seed+".md", spikeTaskFileWith(ids[task.seed], task.seed, task.status, task.dependsOn))
	}
	threadOneID := testutil.TaskID("thread-one")
	threadTwoID := testutil.TaskID("thread-two")
	threadOneTasks := []string{ids["blocked-ready"], ids["queued"], ids["forced"], ids["base-done"], ids["downstream-done"], ids["clear-ready"]}
	threadTwoTasks := []string{ids["base-done"], ids["clear-ready"]}
	repo.File("threads/"+threadOneID+"-thread-one.md", spikeThreadFile(threadOneID, threadspike.ThreadInProgress, threadOneTasks))
	repo.File("threads/"+threadTwoID+"-thread-two.md", spikeThreadFile(threadTwoID, threadspike.ThreadInProgress, threadTwoTasks))

	snapshot, err := NewThreadSpikeRepository(repo.Root, "planning-space").Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Problems) != 0 {
		t.Fatalf("snapshot problems = %+v", snapshot.Problems)
	}
	graph := threadspike.NewGraph(snapshot.Tasks)
	if err := graph.Validate(); err != nil {
		t.Fatal(err)
	}
	view := graph.ViewThread(snapshot.Threads[threadOneID])
	if len(view.ExternalGates) != 1 || view.ExternalGates[0].ID != ids["external"] || !view.ExternalGates[0].External {
		t.Fatalf("external gates = %+v", view.ExternalGates)
	}
	if !reflect.DeepEqual(view.Frontier, []string{ids["clear-ready"]}) || view.Done != 2 || view.Total != len(threadOneTasks) {
		t.Fatalf("Thread projection = %+v", view)
	}
	if !view.Inconsistent || view.Broken {
		t.Fatalf("Thread graph health = inconsistent:%v broken:%v", view.Inconsistent, view.Broken)
	}
	if shared := graph.ViewThread(snapshot.Threads[threadTwoID]); shared.Total != 2 || len(shared.Frontier) != 1 {
		t.Fatalf("shared-task Thread projection = %+v", shared)
	}
	checks := map[string]threadspike.TaskView{
		"queued":        graph.ViewTask(ids["queued"], false),
		"blocked-ready": graph.ViewTask(ids["blocked-ready"], false),
		"forced":        graph.ViewTask(ids["forced"], false),
		"clear-ready":   graph.ViewTask(ids["clear-ready"], false),
		"sound-done":    graph.ViewTask(ids["downstream-done"], false),
	}
	if checks["queued"].Role != threadspike.RoleQueued || checks["queued"].Gate != threadspike.GateBlocked ||
		checks["blocked-ready"].Role != threadspike.RoleCandidate || checks["blocked-ready"].Gate != threadspike.GateBlocked ||
		!checks["forced"].Inconsistent || !checks["clear-ready"].Eligible || !checks["sound-done"].Drained {
		t.Fatalf("task projections = %+v", checks)
	}

	reopened := graph.Tasks()
	base := reopened[ids["base-done"]]
	base.Record.Status = domain.StatusReadyToStart
	reopened[ids["base-done"]] = base
	reopenedGraph := threadspike.NewGraph(reopened)
	if descendant := reopenedGraph.ViewTask(ids["downstream-done"], false); !descendant.Inconsistent || descendant.Drained {
		t.Fatalf("reopened descendant = %+v", descendant)
	}
	completedThread := snapshot.Threads[threadTwoID]
	completedThread.Status = threadspike.ThreadCompleted
	if completed := reopenedGraph.ViewThread(completedThread); !completed.Inconsistent {
		t.Fatalf("reopened completed Thread = %+v", completed)
	}

	invalid := []struct {
		name string
		edge threadspike.Edge
		want string
	}{
		{"self", threadspike.Edge{From: ids["clear-ready"], To: ids["clear-ready"]}, "cannot depend on itself"},
		{"missing", threadspike.Edge{From: testutil.TaskID("missing"), To: ids["clear-ready"]}, "does not exist"},
		{"duplicate", threadspike.Edge{From: ids["external"], To: ids["blocked-ready"]}, "already exists"},
		{"cycle", threadspike.Edge{From: ids["downstream-done"], To: ids["base-done"]}, "dependency cycle"},
	}
	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			if _, err := graph.WithEdges([]threadspike.Edge{test.edge}); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}
