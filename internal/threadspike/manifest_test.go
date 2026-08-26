package threadspike

import (
	"reflect"
	"testing"
	"time"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestComposeMaterializesExistingAndNewNodesIntoDurablePlan(t *testing.T) {
	existingID := testutil.TaskID("existing-gate")
	threadID := testutil.TaskID("planned-thread")
	firstID := testutil.TaskID("planned-first")
	secondID := testutil.TaskID("planned-second")
	snapshot := Snapshot{
		RepoID: "planning-space",
		Tasks: map[string]Task{
			existingID: graphTask("existing-gate", domain.StatusReadyToStart),
		},
		Threads: map[string]Thread{},
		Epics:   map[string]bool{"30-spike": true},
	}
	external := false
	manifest := Manifest{
		Thread: ThreadInput{
			Title:       "Prove Thread DAGs",
			Description: "Exercise the full proposed vertical slice",
			Goal:        "Learn which assumptions survive real persistence.",
			Tags:        []string{"spike", "threads", "spike"},
		},
		Nodes: []NodeInput{
			{Key: "gate", TaskID: existingID, Member: &external},
			{Key: "first", NewTask: &NewTaskInput{Title: "Build first", Epic: "30-spike", Tags: []string{"threads"}}},
			{Key: "second", NewTask: &NewTaskInput{Title: "Build second", Status: domain.StatusNextUp, Epic: "30-spike", Description: "Depends on the first task", Tags: []string{"threads"}}},
		},
		Dependencies: []Edge{{From: "gate", To: "first"}, {From: "first", To: "second"}},
	}
	ids := []string{threadID, firstID, secondID}
	plan, err := Compose(snapshot, manifest, func() string {
		id := ids[0]
		ids = ids[1:]
		return id
	}, time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if plan.Thread.ID != threadID {
		t.Fatalf("Thread ID = %s", plan.Thread.ID)
	}
	if !reflect.DeepEqual(plan.Thread.Tasks, sortedUnique([]string{firstID, secondID})) {
		t.Fatalf("Thread members = %v", plan.Thread.Tasks)
	}
	if !reflect.DeepEqual(plan.Thread.Tags, []string{"spike", "threads"}) {
		t.Fatalf("Thread tags = %v", plan.Thread.Tags)
	}
	if len(plan.Tasks) != 2 || len(plan.Edges) != 2 {
		t.Fatalf("plan = %+v", plan)
	}

	encoded, err := yaml.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	var decoded MaterializedPlan
	if err := yaml.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan, decoded) {
		t.Fatalf("materialized plan did not round-trip\nwant: %#v\n got: %#v", plan, decoded)
	}
	decision, err := PrepareApply(snapshot, decoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(decision.CreateTasks) != 2 || decision.CreateThread == nil || len(decision.UpdateDependencies) != 0 {
		t.Fatalf("decision = %+v", decision)
	}
	if decision.CreateTasks[0].Task.ID != firstID || decision.CreateTasks[1].Task.ID != secondID {
		t.Fatalf("create order = %s, %s; want dependency order %s, %s",
			decision.CreateTasks[0].Task.ID, decision.CreateTasks[1].Task.ID, firstID, secondID)
	}

	stale := snapshot
	stale.Epics = map[string]bool{}
	if _, err := PrepareApply(stale, decoded); err == nil {
		t.Fatal("apply must revalidate a planned task's epic after compose")
	}
}
