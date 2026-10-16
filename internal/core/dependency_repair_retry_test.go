package core

import (
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

type repairRetryStep struct {
	graph     *TaskGraph // nil means contention before the planner runs
	err       error
	empty     bool // a post-planner conflict need not carry a result plan
	committed bool
	applied   bool
}

type repairRetryStore struct {
	nopStore
	steps  []repairRetryStep
	calls  int
	plans  []TaskGraphRepairPlan
	times  []time.Time
	before func()
}

func (s *repairRetryStore) MutateTaskGraphRepair(now time.Time, dryRun bool, planner TaskGraphRepairPlanner) (TaskGraphRepairMutationResult, error) {
	step := s.steps[s.calls]
	s.calls++
	s.times = append(s.times, now)
	if s.before != nil {
		s.before()
	}
	if step.graph == nil {
		return TaskGraphRepairMutationResult{DryRun: dryRun}, step.err
	}
	plan, err := planner(step.graph)
	if err != nil {
		return TaskGraphRepairMutationResult{DryRun: dryRun}, err
	}
	s.plans = append(s.plans, plan)
	if step.empty {
		return TaskGraphRepairMutationResult{DryRun: dryRun}, step.err
	}
	plan, analysis, err := ValidateTaskGraphRepairPlan(step.graph, plan)
	if err != nil {
		return TaskGraphRepairMutationResult{DryRun: dryRun}, err
	}
	result := TaskGraphRepairMutationResult{
		Plan: plan, Analysis: analysis, FinalGraph: analysis.Prospective,
		DryRun: dryRun, Committed: step.committed,
	}
	if step.applied {
		result.AppliedSources = []TaskGraphSourceRef{plan.Operations[0].Edit.Source}
	}
	return result, step.err
}

func repairDrop(source TaskGraphSourceRef, value string) TaskGraphSourceEdit {
	return TaskGraphSourceEdit{Action: TaskGraphSourceDropDeclaration, Source: source, Field: TaskDependencyDependsOn, Value: value}
}

func TestRepairRetriesRetainSourcesAcrossEmptyResultsAndReplanFreshContent(t *testing.T) {
	owner := graphRecord("repair-retry-owner", domain.StatusNextUp, "invalid-token")
	path := graphFixturePath(owner)
	first := localTaskGraph([]domain.Task{owner}, nil)
	renamed := owner
	renamed.Slug = "fresh-label" // a label is not physical occurrence identity
	renamed.DependsOn = []string{"concurrent-intent", "invalid-token"}
	current := localGuardedRecordAt(renamed, path)
	current.Record.Source.Location, current.Record.Source.LocationIsPath = "db://current/context", false
	impostor := graphRecord("repair-retry-impostor", domain.StatusNextUp, "invalid-token")
	impostor.Slug = owner.Slug
	fresh := NewTaskGraphRead(TaskGraphRead{GuardedRecords: []VersionedRecord[domain.Task]{current, localGuardedRecord(impostor)}})
	st := &repairRetryStore{steps: []repairRetryStep{
		{err: domain.ErrConflict},
		{graph: first, err: domain.ErrConflict, empty: true},
		{err: domain.ErrConflict},
		{graph: fresh},
	}}
	clockCalls := 0
	sleeps := 0
	svc := MustNewService(st, WithRetry(3, func(int) { sleeps++ }), WithClock(func() time.Time {
		clockCalls++
		return time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	}))
	receipt, err := svc.RepairTaskGraph(TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{
		repairDrop(TaskGraphSourceRef{TaskSlug: owner.Slug}, "invalid-token"),
	}}, false)
	if err != nil || st.calls != 4 || sleeps != 3 || clockCalls != 1 || len(st.plans) != 2 {
		t.Fatalf("calls=%d sleeps=%d clock=%d plans=%d err=%v", st.calls, sleeps, clockCalls, len(st.plans), err)
	}
	for _, now := range st.times {
		if now != st.times[0] {
			t.Fatal("retry changed its semantic timestamp")
		}
	}
	for _, plan := range st.plans {
		if len(plan.Operations) != 1 || plan.Operations[0].Edit.Source.TaskID != owner.ID || plan.Operations[0].Edit.Source.LocalPath != path {
			t.Fatalf("explicit selection drifted: %+v", plan)
		}
	}
	last := st.plans[1].Operations[0].Edit.Source
	if last.TaskSlug != renamed.Slug || last.Location != current.Record.Source.Location || len(receipt.Residual) != 2 {
		t.Fatalf("retry retained stale labels/diagnostics or lost concurrent defects: source=%+v residual=%+v", last, receipt.Residual)
	}
	_, analysis, err := ValidateTaskGraphRepairPlan(fresh, st.plans[1])
	if err != nil || !slices.Equal(repairRawValues(t, analysis.Prospective, current.Record.Source.Location, TaskDependencyDependsOn), []string{"concurrent-intent"}) {
		t.Fatalf("fresh source content was not preserved: %v", err)
	}
}

func TestRepairRetriesRefuseMissingReplacedAndAmbiguousOccurrences(t *testing.T) {
	owner := graphRecord("repair-pinned-owner", domain.StatusNextUp, "invalid-token")
	path := graphFixturePath(owner)
	other := graphRecord("repair-pinned-other", domain.StatusNextUp, "invalid-token")
	duplicate := owner
	duplicate.Slug = "duplicate"
	renamed := owner
	renamed.Slug = "renamed"
	for _, tc := range []struct {
		name    string
		records []VersionedRecord[domain.Task]
		want    error
	}{
		{name: "deleted", want: domain.ErrNotFound},
		{name: "renamed", records: []VersionedRecord[domain.Task]{localGuardedRecord(renamed)}, want: domain.ErrNotFound},
		{name: "path reused with other ID", records: []VersionedRecord[domain.Task]{localGuardedRecordAt(other, path)}, want: domain.ErrNotFound},
		{name: "duplicate ID elsewhere cannot replace deleted occurrence", records: []VersionedRecord[domain.Task]{localGuardedRecord(duplicate)}, want: domain.ErrNotFound},
		{name: "indistinguishable occurrences", records: []VersionedRecord[domain.Task]{localGuardedRecord(owner), localGuardedRecord(owner)}, want: domain.ErrAmbiguous},
	} {
		for name, source := range map[string]TaskGraphSourceRef{
			"ID": {TaskID: owner.ID}, "slug": {TaskSlug: owner.Slug}, "path": {LocalPath: path},
		} {
			t.Run(tc.name+"/"+name, func(t *testing.T) {
				st := &repairRetryStore{steps: []repairRetryStep{
					{graph: localTaskGraph([]domain.Task{owner}, nil), err: domain.ErrConflict, empty: true},
					{graph: NewTaskGraphRead(TaskGraphRead{GuardedRecords: tc.records})},
				}}
				receipt, err := MustNewService(st, WithRetry(1, func(int) {})).RepairTaskGraph(TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{repairDrop(source, "invalid-token")}}, false)
				if !errors.Is(err, tc.want) || st.calls != 2 || len(st.plans) != 1 || receipt.Committed || len(receipt.AppliedFiles) != 0 {
					t.Fatalf("calls=%d plans=%d receipt=%+v err=%v, want %v", st.calls, len(st.plans), receipt, err, tc.want)
				}
			})
		}
	}
}

func TestRepairRetryReauthorizesDefectsAndCallerLocation(t *testing.T) {
	prerequisite := graphRecord("repair-reauthorize-prerequisite", domain.StatusCompleted)
	owner := graphRecord("repair-reauthorize-owner", domain.StatusNextUp, prerequisite.ID)
	path := graphFixturePath(owner)
	for _, tc := range []struct {
		name   string
		fresh  *TaskGraph
		source TaskGraphSourceRef
		want   error
	}{
		{name: "former dangling reference now valid", fresh: localTaskGraph([]domain.Task{owner, prerequisite}, nil), source: TaskGraphSourceRef{TaskID: owner.ID}, want: domain.ErrValidation},
		{name: "already removed declaration", fresh: localTaskGraph([]domain.Task{graphRecord(owner.Slug, domain.StatusNextUp)}, nil), source: TaskGraphSourceRef{TaskID: owner.ID}},
		{name: "explicit stale context", fresh: NewTaskGraphRead(TaskGraphRead{GuardedRecords: []VersionedRecord[domain.Task]{{
			Record: LoadedRecord[domain.Task]{Value: owner, Source: RecordSource{ID: owner.ID, Location: "db://changed"}}, LocalPath: path,
		}}}), source: TaskGraphSourceRef{LocalPath: path, Location: path}, want: domain.ErrConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := &repairRetryStore{steps: []repairRetryStep{
				{graph: localTaskGraph([]domain.Task{owner}, nil), err: domain.ErrConflict}, {graph: tc.fresh},
			}}
			receipt, err := MustNewService(st, WithRetry(1, func(int) {})).RepairTaskGraph(TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{repairDrop(tc.source, prerequisite.ID)}}, false)
			if !errors.Is(err, tc.want) || st.calls != 2 || receipt.Committed || len(receipt.AppliedFiles) != 0 {
				t.Fatalf("calls=%d receipt=%+v err=%v, want %v", st.calls, receipt, err, tc.want)
			}
			if tc.want == nil && (receipt.Changed || len(receipt.Operations) != 0 || len(receipt.Selected) != 1) {
				t.Fatalf("already-satisfied intent was not a selected no-op: %+v", receipt)
			}
		})
	}
}

func TestRepairRetriesRediagnoseAutoWithoutExpandingExplicitBindings(t *testing.T) {
	owner := graphRecord("repair-auto-explicit", domain.StatusNextUp, "invalid-token")
	old := graphRecord("repair-auto-old", domain.StatusNextUp)
	old.DependsOn = []string{old.ID}
	newSource := graphRecord("repair-auto-new", domain.StatusNextUp)
	newSource.DependsOn = []string{newSource.ID}
	for _, explicit := range []bool{false, true} {
		t.Run(map[bool]string{false: "auto only", true: "mixed"}[explicit], func(t *testing.T) {
			request := TaskGraphRepairRequest{Auto: true}
			if explicit {
				request.Edits = []TaskGraphSourceEdit{repairDrop(TaskGraphSourceRef{TaskSlug: owner.Slug}, "invalid-token")}
			}
			st := &repairRetryStore{steps: []repairRetryStep{
				{graph: localTaskGraph([]domain.Task{owner, old}, nil), err: domain.ErrConflict, empty: true},
				{graph: localTaskGraph([]domain.Task{owner, newSource}, nil)},
			}}
			receipt, err := MustNewService(st, WithRetry(1, func(int) {})).RepairTaskGraph(request, false)
			if err != nil || st.calls != 2 {
				t.Fatalf("calls=%d err=%v", st.calls, err)
			}
			want := 1
			if explicit {
				want++
			}
			if len(receipt.Operations) != want {
				t.Fatalf("operations=%+v", receipt.Operations)
			}
			for _, operation := range receipt.Operations {
				if operation.Edit.Source.TaskID == old.ID || (operation.Automatic && operation.Edit.Source.TaskID != newSource.ID) {
					t.Fatalf("auto cached an old selection: %+v", operation)
				}
			}
		})
	}
}

func TestRepairRetryBoundariesAndDurableFailureReceipt(t *testing.T) {
	owner := graphRecord("repair-retry-boundary", domain.StatusNextUp, "invalid-token")
	for _, tc := range []struct {
		name   string
		step   repairRetryStep
		dryRun bool
	}{
		{name: "preview", step: repairRetryStep{err: domain.ErrConflict}, dryRun: true},
		{name: "non-conflict", step: repairRetryStep{err: domain.ErrValidation}},
		{name: "committed without applied sources", step: repairRetryStep{graph: localTaskGraph([]domain.Task{owner}, nil), err: domain.ErrConflict, committed: true}},
		{name: "applied prefix", step: repairRetryStep{graph: localTaskGraph([]domain.Task{owner}, nil), err: domain.ErrConflict, applied: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := &repairRetryStore{steps: []repairRetryStep{tc.step}}
			receipt, err := MustNewService(st, WithRetry(4, func(int) { t.Fatal("unexpected retry") })).RepairTaskGraph(TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{repairDrop(TaskGraphSourceRef{TaskID: owner.ID}, "invalid-token")}}, tc.dryRun)
			var failure *TaskGraphRepairFailure
			if !errors.Is(err, tc.step.err) || !errors.As(err, &failure) || st.calls != 1 || !reflect.DeepEqual(receipt, failure.Receipt) || receipt.Committed != tc.step.committed {
				t.Fatalf("calls=%d receipt=%+v err=%v", st.calls, receipt, err)
			}
		})
	}
}

func TestRepairOwnsCallerEditSlice(t *testing.T) {
	owner := graphRecord("repair-owned-request", domain.StatusNextUp, "invalid-token")
	impostor := graphRecord("repair-owned-impostor", domain.StatusNextUp, "invalid-token")
	request := TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{repairDrop(TaskGraphSourceRef{TaskID: owner.ID}, "invalid-token")}}
	st := &repairRetryStore{steps: []repairRetryStep{{graph: localTaskGraph([]domain.Task{owner, impostor}, nil)}}}
	st.before = func() { request.Edits[0].Source.TaskID = impostor.ID }
	receipt, err := MustNewService(st).RepairTaskGraph(request, false)
	if err != nil || len(receipt.Selected) != 1 || receipt.Selected[0].Source.TaskID != owner.ID {
		t.Fatalf("caller changed an in-flight selection: receipt=%+v err=%v", receipt, err)
	}
}

func TestRepairRetryReauthorizesDedupeAndEmptyFieldIntent(t *testing.T) {
	prerequisite := graphRecord("repair-dedupe-prerequisite", domain.StatusCompleted)
	owner := graphRecord("repair-dedupe-owner", domain.StatusNextUp, prerequisite.ID, prerequisite.ID)
	freshOwner := owner
	freshOwner.DependsOn = []string{prerequisite.ID, "concurrent-intent", prerequisite.ID, prerequisite.ID}
	empty := graphRecord("repair-empty-owner", domain.StatusNextUp)
	empty.LegacyDependencyFields = []string{"blocked_by"}
	nonempty := empty
	nonempty.LegacyBlockedBy = []string{"unresolved-intent"}
	for _, tc := range []struct {
		name   string
		before domain.Task
		after  domain.Task
		edit   TaskGraphSourceEdit
		want   error
	}{
		{name: "dedupe current occurrences", before: owner, after: freshOwner, edit: TaskGraphSourceEdit{
			Action: TaskGraphSourceDedupe, Source: TaskGraphSourceRef{TaskSlug: owner.Slug}, Field: TaskDependencyDependsOn, Value: prerequisite.ID,
		}},
		{name: "empty field now has intent", before: empty, after: nonempty, edit: TaskGraphSourceEdit{
			Action: TaskGraphSourceDropEmptyField, Source: TaskGraphSourceRef{TaskSlug: empty.Slug}, Field: TaskDependencyBlockedBy,
		}, want: domain.ErrValidation},
		{name: "empty field already removed", before: empty, after: graphRecord(empty.Slug, domain.StatusNextUp), edit: TaskGraphSourceEdit{
			Action: TaskGraphSourceDropEmptyField, Source: TaskGraphSourceRef{LocalPath: graphFixturePath(empty)}, Field: TaskDependencyBlockedBy,
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := &repairRetryStore{steps: []repairRetryStep{
				{graph: localTaskGraph([]domain.Task{tc.before, prerequisite}, nil), err: domain.ErrConflict, empty: true},
				{graph: localTaskGraph([]domain.Task{tc.after, prerequisite}, nil)},
			}}
			receipt, err := MustNewService(st, WithRetry(1, func(int) {})).RepairTaskGraph(TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{tc.edit}}, false)
			if !errors.Is(err, tc.want) || st.calls != 2 {
				t.Fatalf("calls=%d err=%v, want %v", st.calls, err, tc.want)
			}
			if tc.edit.Action == TaskGraphSourceDedupe {
				if len(receipt.Removed) != 2 || len(receipt.Residual) != 1 || receipt.Residual[0].Target.Value != "concurrent-intent" {
					t.Fatalf("dedupe reused old occurrences or lost concurrent intent: %+v", receipt)
				}
			} else if len(receipt.Operations) != 0 {
				t.Fatalf("empty-field retry removed new intent or invented work: %+v", receipt)
			}
		})
	}
}
