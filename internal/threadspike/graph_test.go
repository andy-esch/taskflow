package threadspike

import (
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func graphTask(seed string, status domain.Status, dependencies ...string) Task {
	taskID := testutil.TaskID(seed)
	return Task{
		Record:    domain.Task{ID: taskID, FilenameID: taskID, Slug: seed, Status: status},
		DependsOn: append([]string(nil), dependencies...),
	}
}

func scenarioGraph() (*Graph, map[string]string) {
	ids := map[string]string{}
	for _, seed := range []string{"external", "blocked-ready", "queued", "forced", "base-done", "downstream-done", "clear-ready"} {
		ids[seed] = testutil.TaskID(seed)
	}
	tasks := []Task{
		graphTask("external", domain.StatusReadyToStart),
		graphTask("blocked-ready", domain.StatusReadyToStart, ids["external"]),
		graphTask("queued", domain.StatusNextUp, ids["external"]),
		graphTask("forced", domain.StatusInProgress, ids["external"]),
		graphTask("base-done", domain.StatusCompleted),
		graphTask("downstream-done", domain.StatusCompleted, ids["base-done"]),
		graphTask("clear-ready", domain.StatusReadyToStart),
	}
	byID := make(map[string]Task, len(tasks))
	for _, task := range tasks {
		byID[task.Record.ID] = task
	}
	return NewGraph(byID), ids
}

func TestGraphProjectsLifecycleAndDependencyHealthSeparately(t *testing.T) {
	g, ids := scenarioGraph()
	if err := g.Validate(); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		seed         string
		role         LifecycleRole
		gate         GateState
		eligible     bool
		drained      bool
		inconsistent bool
	}{
		{"blocked-ready", RoleCandidate, GateBlocked, false, false, false},
		{"queued", RoleQueued, GateBlocked, false, false, false},
		{"forced", RoleInFlight, GateBlocked, false, false, true},
		{"base-done", RoleNominallyComplete, GateClear, false, true, false},
		{"downstream-done", RoleNominallyComplete, GateClear, false, true, false},
		{"clear-ready", RoleCandidate, GateClear, true, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.seed, func(t *testing.T) {
			view := g.ViewTask(ids[tc.seed], false)
			if view.Role != tc.role || view.Gate != tc.gate || view.Eligible != tc.eligible ||
				view.Drained != tc.drained || view.Inconsistent != tc.inconsistent {
				t.Fatalf("view = %+v", view)
			}
		})
	}
}

func TestThreadViewSupportsSharedMembersAndExternalGates(t *testing.T) {
	g, ids := scenarioGraph()
	thread := Thread{
		ID:     testutil.TaskID("thread-one"),
		Status: ThreadInProgress,
		Tasks: []string{
			ids["blocked-ready"], ids["queued"], ids["forced"],
			ids["base-done"], ids["downstream-done"], ids["clear-ready"],
		},
	}
	other := Thread{
		ID:     testutil.TaskID("thread-two"),
		Status: ThreadInProgress,
		Tasks:  []string{ids["base-done"], ids["clear-ready"]},
	}

	view := g.ViewThread(thread)
	if view.Done != 2 || view.Total != 6 || view.Withdrawn != 0 {
		t.Fatalf("rollup = %d/%d, withdrawn=%d", view.Done, view.Total, view.Withdrawn)
	}
	if !reflect.DeepEqual(view.Frontier, []string{ids["clear-ready"]}) {
		t.Fatalf("frontier = %v", view.Frontier)
	}
	if len(view.ExternalGates) != 1 || view.ExternalGates[0].ID != ids["external"] || !view.ExternalGates[0].External {
		t.Fatalf("external gates = %+v", view.ExternalGates)
	}
	if !view.Inconsistent || view.Broken {
		t.Fatalf("Thread health = inconsistent:%v broken:%v", view.Inconsistent, view.Broken)
	}
	otherView := g.ViewThread(other)
	if otherView.Total != 2 || len(otherView.Frontier) != 1 {
		t.Fatalf("shared task projection = %+v", otherView)
	}
}

func TestReopeningPrerequisiteInvalidatesSoundCompletion(t *testing.T) {
	g, ids := scenarioGraph()
	tasks := g.Tasks()
	reopened := tasks[ids["base-done"]]
	reopened.Record.Status = domain.StatusReadyToStart
	tasks[ids["base-done"]] = reopened
	g = NewGraph(tasks)

	downstream := g.ViewTask(ids["downstream-done"], false)
	if downstream.Gate != GateBlocked || downstream.Drained || !downstream.Inconsistent {
		t.Fatalf("reopened downstream view = %+v", downstream)
	}
	thread := Thread{ID: testutil.TaskID("thread"), Status: ThreadCompleted, Tasks: []string{ids["base-done"], ids["downstream-done"]}}
	if view := g.ViewThread(thread); !view.Inconsistent {
		t.Fatalf("completed Thread should be inconsistent after reopen: %+v", view)
	}
}

func TestGraphValidationDiagnostics(t *testing.T) {
	a := testutil.TaskID("a")
	b := testutil.TaskID("b")
	missing := testutil.TaskID("missing")
	cases := []struct {
		name string
		g    *Graph
		want string
	}{
		{"self", NewGraph(map[string]Task{a: graphTask("a", domain.StatusReadyToStart, a)}), "cannot depend on itself"},
		{"duplicate", NewGraph(map[string]Task{
			a: graphTask("a", domain.StatusReadyToStart),
			b: graphTask("b", domain.StatusReadyToStart, a, a),
		}), "repeats dependency"},
		{"missing", NewGraph(map[string]Task{
			a: graphTask("a", domain.StatusReadyToStart, missing),
		}), "depends on missing task"},
		{"cycle", NewGraph(map[string]Task{
			a: graphTask("a", domain.StatusReadyToStart, b),
			b: graphTask("b", domain.StatusReadyToStart, a),
		}), "dependency cycle"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.g.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want containing %q", err, tc.want)
			}
			if tc.name == "cycle" {
				cycle := tc.g.CyclePath()
				if len(cycle) != 3 || cycle[0] != cycle[2] {
					t.Fatalf("cycle path = %v", cycle)
				}
			}
		})
	}
}

func TestGraphQueriesAndPlanAreDeterministic(t *testing.T) {
	g, ids := scenarioGraph()
	if got := g.Blockers(ids["blocked-ready"]); !reflect.DeepEqual(got, []string{ids["external"]}) {
		t.Fatalf("blockers = %v", got)
	}
	if got := g.Blockers(ids["downstream-done"]); len(got) != 0 {
		t.Fatalf("soundly completed prerequisites are not blockers: %v", got)
	}
	wantUnblocks := []string{ids["blocked-ready"], ids["forced"], ids["queued"]}
	if got := g.Unblocks(ids["external"]); !reflect.DeepEqual(got, sortedUnique(wantUnblocks)) {
		t.Fatalf("unblocks = %v, want %v", got, sortedUnique(wantUnblocks))
	}
	members := []string{ids["downstream-done"], ids["base-done"], ids["clear-ready"]}
	waves, err := g.Plan(members)
	if err != nil {
		t.Fatal(err)
	}
	wantFirst := sortedUnique([]string{ids["base-done"], ids["clear-ready"]})
	if len(waves) != 2 || !reflect.DeepEqual(waves[0], wantFirst) || !reflect.DeepEqual(waves[1], []string{ids["downstream-done"]}) {
		t.Fatalf("waves = %v", waves)
	}
}

func TestBrokenReferenceProducesBrokenGate(t *testing.T) {
	task := graphTask("broken", domain.StatusReadyToStart, testutil.TaskID("missing"))
	g := NewGraph(map[string]Task{task.Record.ID: task})
	view := g.ViewTask(task.Record.ID, false)
	if view.Gate != GateBroken || view.Eligible {
		t.Fatalf("broken view = %+v", view)
	}
}
