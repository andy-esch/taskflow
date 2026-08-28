package core

import (
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

type graphMutationFailure struct {
	err   error
	after int
}

// graphOperationStore is a small semantic implementation of the mutation port.
// The filesystem adapter has its own locking/CAS tests; these tests isolate the
// service planners, receipts, retry boundary, and query behavior.
type graphOperationStore struct {
	fakeStore
	calls    int
	failures []graphMutationFailure
}

func (s *graphOperationStore) MutateTaskGraph(_ time.Time, dryRun bool, planner TaskGraphPlanner) (TaskGraphMutationResult, error) {
	s.calls++
	result := TaskGraphMutationResult{DryRun: dryRun}
	graph := NewTaskGraph(s.tasks, s.problems)
	if err := ValidateTaskGraphMutationSource(graph); err != nil {
		return result, err
	}
	plan, err := planner(graph)
	if err != nil {
		return result, err
	}
	result.Plan, err = ValidateTaskGraphMutationPlan(graph, plan)
	if err != nil || dryRun {
		return result, err
	}
	var failure graphMutationFailure
	if len(s.failures) > 0 {
		failure, s.failures = s.failures[0], s.failures[1:]
		if failure.err != nil && failure.after == 0 {
			return result, failure.err
		}
	}
	for _, write := range result.Plan.TaskWrites {
		for i := range s.tasks {
			if s.tasks[i].ID != write.TaskID {
				continue
			}
			s.tasks[i].DependsOn = append([]string(nil), write.DependsOn...)
			if write.ClearLegacy {
				s.tasks[i].LegacyBlockedBy = nil
				s.tasks[i].LegacyDependencies = nil
				s.tasks[i].LegacyBlocks = nil
				s.tasks[i].LegacyDependencyFields = nil
			}
			break
		}
		result.AppliedTaskIDs = append(result.AppliedTaskIDs, write.TaskID)
		if failure.err != nil && len(result.AppliedTaskIDs) == failure.after {
			return result, failure.err
		}
	}
	return result, nil
}

func TestServiceDependencyAddRemoveDryRunAndIdempotence(t *testing.T) {
	alpha := graphRecord("alpha-prerequisite", domain.StatusCompleted)
	charlie := graphRecord("charlie-prerequisite", domain.StatusCompleted)
	dependent := graphRecord("dependent", domain.StatusReadyToStart, charlie.ID)
	store := &graphOperationStore{fakeStore: fakeStore{tasks: []domain.Task{dependent, charlie, alpha}}}
	svc := NewService(store)

	dry, err := svc.AddTaskDependencies("DEPEN", []string{"ALPHA", charlie.ID}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !dry.Changed || !dry.DryRun || len(dry.PlannedTaskIDs) != 1 || len(dry.AppliedTaskIDs) != 0 {
		t.Fatalf("dry-run receipt = %+v", dry)
	}
	if !slices.Equal(store.tasks[0].DependsOn, []string{charlie.ID}) {
		t.Fatalf("dry run changed store: %v", store.tasks[0].DependsOn)
	}

	added, err := svc.AddTaskDependencies(dependent.ID, []string{charlie.Slug, alpha.Slug}, false)
	if err != nil {
		t.Fatal(err)
	}
	wantDependencies := []string{alpha.ID, charlie.ID}
	slices.Sort(wantDependencies)
	if !added.Changed || !slices.Equal(added.AppliedTaskIDs, []string{dependent.ID}) ||
		!slices.Equal(store.tasks[0].DependsOn, wantDependencies) {
		t.Fatalf("add receipt=%+v dependencies=%v", added, store.tasks[0].DependsOn)
	}
	if len(added.Edges) != 2 || added.Edges[0].Outcome != "added" || added.Edges[1].Outcome != "skipped" {
		// Edge outcomes are canonical-ID sorted; determine the semantic counts below
		// rather than coupling this assertion to hash-derived ID order.
		outcomes := make(map[string]string)
		for _, edge := range added.Edges {
			outcomes[edge.PrerequisiteID] = edge.Outcome
		}
		if outcomes[alpha.ID] != "added" || outcomes[charlie.ID] != "skipped" {
			t.Fatalf("edge outcomes = %+v", added.Edges)
		}
	}

	noop, err := svc.AddTaskDependencies(dependent.Slug, []string{alpha.Slug}, false)
	if err != nil || noop.Changed || len(noop.AppliedTaskIDs) != 0 || noop.Edges[0].Outcome != "skipped" {
		t.Fatalf("idempotent add = %+v, %v", noop, err)
	}
	removed, err := svc.RemoveTaskDependencies(dependent.Slug, []string{alpha.Slug}, false)
	if err != nil || !removed.Changed || removed.Edges[0].Outcome != "removed" ||
		!slices.Equal(store.tasks[0].DependsOn, []string{charlie.ID}) {
		t.Fatalf("remove = %+v, %v; dependencies=%v", removed, err, store.tasks[0].DependsOn)
	}
}

func TestServiceDependencyMutationRejectsAmbiguousDuplicateSelfAndCycle(t *testing.T) {
	alpha := graphRecord("add-retry-alpha", domain.StatusReadyToStart)
	atom := graphRecord("add-retry-atom", domain.StatusReadyToStart)
	dependent := graphRecord("dependent", domain.StatusReadyToStart)
	store := &graphOperationStore{fakeStore: fakeStore{tasks: []domain.Task{dependent, atom, alpha}}}
	svc := NewService(store)

	if _, err := svc.AddTaskDependencies(dependent.Slug, []string{"add-retry"}, false); !errors.Is(err, domain.ErrAmbiguous) {
		t.Fatalf("ambiguous prerequisite = %v", err)
	}
	if _, err := svc.AddTaskDependencies(dependent.Slug, []string{alpha.Slug, alpha.ID}, false); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("duplicate canonical prerequisite = %v", err)
	}
	if _, err := svc.AddTaskDependencies(dependent.Slug, []string{dependent.ID}, false); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("self dependency = %v", err)
	}
	if _, err := svc.AddTaskDependencies(dependent.Slug, []string{alpha.Slug}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddTaskDependencies(alpha.Slug, []string{dependent.Slug}, false); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("cycle creation = %v", err)
	}
}

func TestServiceDependencyMigrationConvergesLegacyVocabulary(t *testing.T) {
	prerequisite := graphRecord("legacy-prerequisite", domain.StatusCompleted)
	second := graphRecord("legacy-second", domain.StatusCompleted)
	dependent := graphRecord("legacy-dependent", domain.StatusReadyToStart)
	prerequisite.LegacyBlocks = []string{dependent.Slug}
	dependent.LegacyBlockedBy = []string{prerequisite.Slug}
	dependent.LegacyDependencies = []string{second.ID}
	store := &graphOperationStore{fakeStore: fakeStore{tasks: []domain.Task{prerequisite, second, dependent}}}
	svc := NewService(store)

	receipt, err := svc.MigrateTaskDependencies(false)
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.Changed || len(receipt.ClearedLegacyFields) != 3 || len(receipt.AppliedTaskIDs) != 2 {
		t.Fatalf("migration receipt = %+v", receipt)
	}
	for _, task := range store.tasks {
		if len(task.LegacyBlockedBy)+len(task.LegacyDependencies)+len(task.LegacyBlocks) != 0 {
			t.Fatalf("legacy fields remain on %+v", task)
		}
		if task.ID == dependent.ID {
			want := []string{prerequisite.ID, second.ID}
			slices.Sort(want)
			if !slices.Equal(task.DependsOn, want) {
				t.Fatalf("migrated dependencies = %v, want %v", task.DependsOn, want)
			}
		}
	}
	if graph := NewTaskGraph(store.tasks, nil); graph.Health() != GraphHealthy {
		t.Fatalf("migration did not converge to healthy graph: %s %+v", graph.Health(), graph.Problems())
	}
	noop, err := svc.MigrateTaskDependencies(false)
	if err != nil || noop.Changed || len(noop.AppliedTaskIDs) != 0 {
		t.Fatalf("migration rerun = %+v, %v", noop, err)
	}
}

func TestServiceDependencyMutationRetriesOnlyBeforeDurablePrefix(t *testing.T) {
	prerequisite := graphRecord("retry-prerequisite", domain.StatusCompleted)
	dependent := graphRecord("retry-dependent", domain.StatusReadyToStart)
	store := &graphOperationStore{
		fakeStore: fakeStore{tasks: []domain.Task{dependent, prerequisite}},
		failures:  []graphMutationFailure{{err: domain.ErrConflict}},
	}
	svc := NewService(store, WithRetry(2, func(int) {}))
	receipt, err := svc.AddTaskDependencies(dependent.Slug, []string{prerequisite.Slug}, false)
	if err != nil || store.calls != 2 || !slices.Equal(receipt.AppliedTaskIDs, []string{dependent.ID}) {
		t.Fatalf("pre-write retry receipt=%+v calls=%d err=%v", receipt, store.calls, err)
	}

	legacyOwner := graphRecord("prefix-owner", domain.StatusCompleted)
	legacyDependent := graphRecord("prefix-dependent", domain.StatusReadyToStart)
	legacyOwner.LegacyBlocks = []string{legacyDependent.Slug}
	legacyDependent.LegacyBlockedBy = []string{legacyOwner.Slug}
	partialStore := &graphOperationStore{
		fakeStore: fakeStore{tasks: []domain.Task{legacyOwner, legacyDependent}},
		failures:  []graphMutationFailure{{err: domain.ErrConflict, after: 1}},
	}
	partialSvc := NewService(partialStore, WithRetry(4, func(int) {}))
	partial, err := partialSvc.MigrateTaskDependencies(false)
	var failure *DependencyMutationFailure
	if !errors.Is(err, domain.ErrConflict) || !errors.As(err, &failure) || partialStore.calls != 1 {
		t.Fatalf("partial mutation calls=%d receipt=%+v err=%v", partialStore.calls, partial, err)
	}
	if len(partial.AppliedTaskIDs) != 1 || len(partial.RemainingTaskIDs) != 1 ||
		!reflect.DeepEqual(partial, failure.Receipt) {
		t.Fatalf("typed durable-prefix receipt=%+v failure=%+v", partial, failure.Receipt)
	}
}

func TestServiceTaskGraphQueriesExplainCurrentSnapshot(t *testing.T) {
	root := graphRecord("query-root", domain.StatusReadyToStart)
	middle := graphRecord("query-middle", domain.StatusCompleted, root.ID)
	target := graphRecord("query-target", domain.StatusReadyToStart, middle.ID)
	store := &graphOperationStore{fakeStore: fakeStore{tasks: []domain.Task{target, middle, root}}}
	svc := NewService(store)

	frontier, err := svc.TaskBlockers(target.Slug, false)
	if err != nil || frontier.Projection != "frontier" || len(frontier.Blockers) != 1 ||
		frontier.Blockers[0].Blocker.TaskID != root.ID || frontier.Blockers[0].Blocker.Direct ||
		frontier.State.Gate != GateBlocked || frontier.State.Eligible {
		t.Fatalf("frontier = %+v, %v", frontier, err)
	}
	causal, err := svc.TaskBlockers(target.Slug, true)
	if err != nil || causal.Projection != "causal" || len(causal.Blockers) != 2 {
		t.Fatalf("causal = %+v, %v", causal, err)
	}
	unblocks, err := svc.TaskUnblocks(root.Slug)
	if err != nil || len(unblocks.Unblocks) != 2 || unblocks.State.Role != RoleCandidate {
		t.Fatalf("unblocks = %+v, %v", unblocks, err)
	}
	for _, detail := range unblocks.Unblocks {
		if detail.Impact.TaskID == target.ID && !slices.Equal(detail.Impact.Path, []string{root.ID, middle.ID, target.ID}) {
			t.Fatalf("target downstream path = %v", detail.Impact.Path)
		}
	}
}
