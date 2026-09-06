package core

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestTaskGraphRepairAutoIsLimitedAndPreservesExplicitIntent(t *testing.T) {
	prerequisite := graphRecord("repair-auto-prerequisite", domain.StatusCompleted)
	dangling := testutil.TaskID("repair-auto-dangling")
	owner := graphRecord("repair-auto-owner", domain.StatusReadyToStart,
		prerequisite.ID, prerequisite.ID, "invalid-human-token", dangling)
	owner.DependsOn = append(owner.DependsOn, owner.ID)
	owner.LegacyDependencyFields = []string{"blocked_by"}
	graph := NewTaskGraph([]domain.Task{owner, prerequisite}, nil)

	diagnosis, err := DiagnoseTaskGraphRepair(graph)
	if err != nil {
		t.Fatal(err)
	}
	assertRepairReason(t, diagnosis.Defects, RepairDuplicate, true)
	assertRepairReason(t, diagnosis.Defects, RepairSelf, true)
	assertRepairReason(t, diagnosis.Defects, RepairLegacyEmpty, true)
	assertRepairReason(t, diagnosis.Defects, RepairInvalidID, false)
	assertRepairReason(t, diagnosis.Defects, RepairDangling, false)

	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Auto: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Operations) != 3 {
		t.Fatalf("auto operations = %+v", plan.Operations)
	}
	_, analysis, err := ValidateTaskGraphRepairPlan(graph, plan)
	if err != nil {
		t.Fatal(err)
	}
	if analysis.After.Health != GraphBroken {
		t.Fatalf("after health = %s, want broken residual explicit defects", analysis.After.Health)
	}
	values := repairRawValues(t, analysis.Prospective, owner.Path, TaskDependencyDependsOn)
	if !slices.Equal(values, []string{prerequisite.ID, "invalid-human-token", dangling}) {
		t.Fatalf("auto repair values = %v", values)
	}
	if len(analysis.Removed) != 2 {
		t.Fatalf("removed declarations = %+v", analysis.Removed)
	}
	record := sourceRecordFor(t, mustSourceRecords(t, analysis.Prospective), owner.Path)
	for _, field := range record.Fields {
		if field.Field == TaskDependencyBlockedBy {
			t.Fatalf("empty legacy field survived auto repair: %+v", record.Fields)
		}
	}
}

func TestTaskGraphRepairAutoHandlesRepeatedSelfDeclarationsWithoutOrderDependence(t *testing.T) {
	owner := graphRecord("repair-repeated-self", domain.StatusNextUp)
	owner.DependsOn = []string{owner.ID, owner.ID}
	graph := NewTaskGraph([]domain.Task{owner}, nil)
	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Auto: true})
	if err != nil {
		t.Fatal(err)
	}
	validated, analysis, err := ValidateTaskGraphRepairPlan(graph, plan)
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range validated.Operations {
		if !operation.Automatic {
			t.Fatalf("validated auto-selected operation lost provenance: %+v", operation)
		}
	}
	if analysis.After.Health != GraphHealthy || len(analysis.Removed) != 2 {
		t.Fatalf("repeated self repair = %+v", analysis)
	}
}

func TestTaskGraphRepairRejectsOverlappingDedupeAndExactDrop(t *testing.T) {
	owner := graphRecord("repair-overlap-owner", domain.StatusNextUp, "invalid-token", "invalid-token")
	graph := NewTaskGraph([]domain.Task{owner}, nil)
	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{
		Auto: true,
		Edits: []TaskGraphSourceEdit{{
			Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(owner),
			Field: TaskDependencyDependsOn, Value: "invalid-token",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := ValidateTaskGraphRepairPlan(graph, plan); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "overlap") {
		t.Fatalf("overlapping repair error = %v", err)
	}
}

func TestTaskGraphRepairValidationRejectsUnaccountedReceiptSelections(t *testing.T) {
	dangling := testutil.TaskID("repair-receipt-dangling")
	owner := graphRecord("repair-receipt-claim", domain.StatusNextUp, "invalid-token", dangling)
	graph := NewTaskGraph([]domain.Task{owner}, nil)
	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(owner),
		Field: TaskDependencyDependsOn, Value: "invalid-token",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	plan.Selections = []TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(owner),
		Field: TaskDependencyDependsOn, Value: dangling,
	}}
	if _, _, err := ValidateTaskGraphRepairPlan(graph, plan); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "does not exactly account") {
		t.Fatalf("unaccounted selected intent error = %v", err)
	}
}

func TestTaskGraphRepairValidationRejectsFalseAutomaticProvenance(t *testing.T) {
	owner := graphRecord("repair-false-auto", domain.StatusNextUp, "invalid-token")
	graph := NewTaskGraph([]domain.Task{owner}, nil)
	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(owner),
		Field: TaskDependencyDependsOn, Value: "invalid-token",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	plan.Operations[0].Automatic = true
	if _, _, err := ValidateTaskGraphRepairPlan(graph, plan); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "not safe for automatic selection") {
		t.Fatalf("false automatic provenance error = %v", err)
	}
}

func TestTaskGraphRepairRejectsSemanticallyInvalidEditShapes(t *testing.T) {
	owner := graphRecord("repair-edit-shapes", domain.StatusNextUp)
	graph := NewTaskGraph([]domain.Task{owner}, nil)
	for _, edit := range []TaskGraphSourceEdit{
		{Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(owner), Field: TaskDependencyDependsOn, Value: "missing", Occurrence: -1},
		{Action: TaskGraphSourceDedupe, Source: sourceRefForTask(owner), Field: TaskDependencyBlockedBy, Value: "missing"},
		{Action: TaskGraphSourceDedupe, Source: sourceRefForTask(owner), Field: TaskDependencyDependsOn, Value: "missing", Occurrence: 1},
		{Action: TaskGraphSourceDropEmptyField, Source: sourceRefForTask(owner), Field: TaskDependencyDependsOn},
		{Action: TaskGraphSourceDropEmptyField, Source: sourceRefForTask(owner), Field: TaskDependencyBlockedBy, Value: "missing"},
	} {
		if _, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{edit}}); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("edit %+v error = %v", edit, err)
		}
	}
}

func TestTaskGraphRepairExplicitDropsConvergeAndRejectValidConstraints(t *testing.T) {
	prerequisite := graphRecord("repair-explicit-prerequisite", domain.StatusCompleted)
	dangling := testutil.TaskID("repair-explicit-dangling")
	owner := graphRecord("repair-explicit-owner", domain.StatusReadyToStart,
		prerequisite.ID, "invalid-human-token", dangling)
	graph := NewTaskGraph([]domain.Task{owner, prerequisite}, nil)
	source := sourceRefForTask(owner)
	request := TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{
		{Action: TaskGraphSourceDropDeclaration, Source: TaskGraphSourceRef{TaskID: owner.ID}, Field: TaskDependencyDependsOn, Value: "invalid-human-token"},
		{Action: TaskGraphSourceDropDeclaration, Source: TaskGraphSourceRef{TaskSlug: owner.Slug}, Field: TaskDependencyDependsOn, Value: dangling},
	}}
	plan, err := PlanTaskGraphRepair(graph, request)
	if err != nil {
		t.Fatal(err)
	}
	_, analysis, err := ValidateTaskGraphRepairPlan(graph, plan)
	if err != nil {
		t.Fatal(err)
	}
	if analysis.After.Health != GraphHealthy {
		t.Fatalf("after health = %s; problems=%+v", analysis.After.Health, analysis.After.Problems)
	}
	if got := repairRawValues(t, analysis.Prospective, owner.Path, TaskDependencyDependsOn); !slices.Equal(got, []string{prerequisite.ID}) {
		t.Fatalf("remaining values = %v", got)
	}

	retry, err := PlanTaskGraphRepair(analysis.Prospective, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{
		{Action: TaskGraphSourceDropDeclaration, Source: source, Field: TaskDependencyDependsOn, Value: "invalid-human-token"},
	}})
	if err != nil || len(retry.Operations) != 0 {
		t.Fatalf("already-applied retry = %+v, %v", retry, err)
	}
	_, err = PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{
		{Action: TaskGraphSourceDropDeclaration, Source: source, Field: TaskDependencyDependsOn, Value: prerequisite.ID},
	}})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("valid constraint removal error = %v", err)
	}
}

func TestTaskGraphRepairCycleRequiresExplicitEdgeAndCanLeaveOtherDefects(t *testing.T) {
	alpha := graphRecord("repair-cycle-alpha", domain.StatusNextUp)
	beta := graphRecord("repair-cycle-beta", domain.StatusNextUp)
	dangling := testutil.TaskID("repair-cycle-dangling")
	alpha.DependsOn = []string{beta.ID, dangling}
	beta.DependsOn = []string{alpha.ID}
	graph := NewTaskGraph([]domain.Task{beta, alpha}, nil)

	auto, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Auto: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(auto.Operations) != 0 {
		t.Fatalf("auto guessed cycle/dangling repair: %+v", auto.Operations)
	}
	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: TaskGraphSourceRef{TaskID: beta.ID},
		Field: TaskDependencyDependsOn, Value: alpha.ID,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	_, analysis, err := ValidateTaskGraphRepairPlan(graph, plan)
	if err != nil {
		t.Fatal(err)
	}
	if analysis.After.Health != GraphBroken {
		t.Fatalf("cycle repair should retain dangling defect; health=%s", analysis.After.Health)
	}
	if hasRepairReason(analysis.After.Defects, RepairCycle) {
		t.Fatalf("cycle remains after selected break: %+v", analysis.After.Defects)
	}
	if !hasRepairReason(analysis.After.Defects, RepairDangling) {
		t.Fatalf("unrelated dangling intent was lost: %+v", analysis.After.Defects)
	}
}

func TestTaskGraphRepairTargetsEveryLegacyFieldWithoutClearingItsNeighbors(t *testing.T) {
	for _, field := range []TaskDependencyField{TaskDependencyBlockedBy, TaskDependencyDependencies, TaskDependencyBlocks} {
		t.Run(string(field), func(t *testing.T) {
			owner := graphRecord("repair-legacy-"+string(field), domain.StatusNextUp)
			owner.LegacyDependencyFields = []string{string(field)}
			switch field {
			case TaskDependencyBlockedBy:
				owner.LegacyBlockedBy = []string{"missing-human-intent"}
			case TaskDependencyDependencies:
				owner.LegacyDependencies = []string{"missing-human-intent"}
			case TaskDependencyBlocks:
				owner.LegacyBlocks = []string{"missing-human-intent"}
			}
			graph := NewTaskGraph([]domain.Task{owner}, nil)
			plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Auto: true, Edits: []TaskGraphSourceEdit{
				{Action: TaskGraphSourceDropDeclaration, Source: TaskGraphSourceRef{TaskID: owner.ID}, Field: field, Value: "missing-human-intent"},
			}})
			if err != nil {
				t.Fatal(err)
			}
			_, analysis, err := ValidateTaskGraphRepairPlan(graph, plan)
			if err != nil {
				t.Fatal(err)
			}
			if analysis.After.Health != GraphHealthy {
				t.Fatalf("legacy repair health=%s defects=%+v", analysis.After.Health, analysis.After.Defects)
			}
			if len(sourceRecordFor(t, mustSourceRecords(t, analysis.Prospective), owner.Path).Fields) != 0 {
				t.Fatalf("legacy key survived combined explicit+auto repair")
			}
		})
	}
}

func TestTaskGraphRepairLegacyBlocksUsesDeclarationOwnerNotProjectedDependent(t *testing.T) {
	owner := graphRecord("repair-legacy-blocks-owner", domain.StatusNextUp)
	dependent := graphRecord("repair-legacy-blocks-dependent", domain.StatusNextUp)
	owner.DependsOn = []string{dependent.ID}
	owner.LegacyBlocks = []string{dependent.Slug}
	owner.LegacyDependencyFields = []string{"blocks"}
	graph := NewTaskGraph([]domain.Task{owner, dependent}, nil)
	diagnosis, err := DiagnoseTaskGraphRepair(graph)
	if err != nil {
		t.Fatal(err)
	}
	var legacy TaskGraphRepairDefect
	for _, defect := range diagnosis.Defects {
		if defect.Reason == RepairLegacyUnsafe {
			legacy = defect
			break
		}
	}
	if legacy.Target.Source.TaskID != owner.ID || legacy.ProjectedEdge != (DependencyEdge{From: owner.ID, To: dependent.ID}) {
		t.Fatalf("legacy blocks attribution = %+v", legacy)
	}
	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{legacy.Target}})
	if err != nil {
		t.Fatal(err)
	}
	_, analysis, err := ValidateTaskGraphRepairPlan(graph, plan)
	if err != nil || analysis.After.Health != GraphDegraded || hasRepairReason(analysis.After.Defects, RepairCycle) {
		t.Fatalf("legacy blocks cycle repair analysis=%+v err=%v", analysis, err)
	}
}

func TestTaskGraphRepairNeverGuessesAmbiguousLegacyIntent(t *testing.T) {
	first := graphRecord("repair-ambiguous-first", domain.StatusCompleted)
	second := graphRecord("repair-ambiguous-second", domain.StatusCompleted)
	first.Slug, second.Slug = "shared", "shared"
	owner := graphRecord("repair-ambiguous-owner", domain.StatusNextUp)
	owner.LegacyBlockedBy = []string{"shared"}
	owner.LegacyDependencyFields = []string{"blocked_by"}
	graph := NewTaskGraph([]domain.Task{owner, second, first}, nil)
	auto, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Auto: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(auto.Operations) != 0 {
		t.Fatalf("auto guessed ambiguous legacy intent: %+v", auto)
	}
	diagnosis, err := DiagnoseTaskGraphRepair(graph)
	if err != nil {
		t.Fatal(err)
	}
	for _, defect := range diagnosis.Defects {
		if defect.Reason == RepairLegacyAmbiguous {
			if defect.Automatic || len(defect.CandidateIDs) != 2 {
				t.Fatalf("ambiguous defect = %+v", defect)
			}
			return
		}
	}
	t.Fatalf("ambiguous legacy defect missing: %+v", diagnosis.Defects)
}

func TestTaskGraphRepairPreservesShadowAndUnreadableEvidence(t *testing.T) {
	prerequisite := graphRecord("repair-shadow-prerequisite", domain.StatusCompleted)
	representative := graphRecord("repair-shadow-a", domain.StatusNextUp, prerequisite.ID, prerequisite.ID)
	shadow := graphRecord("repair-shadow-b", domain.StatusNextUp, "shadow-invalid")
	shadow.ID = representative.ID
	shadow.FilenameID = representative.ID
	representative.Path = "tasks/a-representative.md"
	shadow.Path = "tasks/b-shadow.md"
	unreadable := TaskGraphLoadProblem{TaskID: testutil.TaskID("repair-shadow-unreadable"), Path: "tasks/unreadable.md", Message: "bad yaml", SourceVersion: "raw-v1"}
	graph := NewTaskGraphRead(TaskGraphRead{Tasks: []domain.Task{shadow, prerequisite, representative}, Problems: []TaskGraphLoadProblem{unreadable}})

	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Auto: true})
	if err != nil {
		t.Fatal(err)
	}
	_, analysis, err := ValidateTaskGraphRepairPlan(graph, plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Prospective.sourceTasks) != len(graph.sourceTasks) || !sameTaskGraphLoadProblem(analysis.Prospective.loadProblems[0], unreadable) {
		t.Fatalf("source evidence changed: tasks=%+v unreadable=%+v", analysis.Prospective.sourceTasks, analysis.Prospective.loadProblems)
	}
	if got := repairRawValues(t, analysis.Prospective, shadow.Path, TaskDependencyDependsOn); !slices.Equal(got, shadow.DependsOn) {
		t.Fatalf("shadow declaration changed: %v", got)
	}
}

func TestTaskGraphRepairAutoRemovesSelfDeclarationFromDuplicateIDShadow(t *testing.T) {
	representative := graphRecord("repair-shadow-self-primary", domain.StatusNextUp)
	shadow := graphRecord("repair-shadow-self-secondary", domain.StatusNextUp)
	shadow.ID = representative.ID
	shadow.FilenameID = representative.ID
	representative.Path = "tasks/a-primary.md"
	shadow.Path = "tasks/b-shadow.md"
	shadow.DependsOn = []string{shadow.ID}
	graph := NewTaskGraph([]domain.Task{shadow, representative}, nil)

	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Auto: true})
	if err != nil {
		t.Fatal(err)
	}
	_, analysis, err := ValidateTaskGraphRepairPlan(graph, plan)
	if err != nil {
		t.Fatal(err)
	}
	if got := repairRawValues(t, analysis.Prospective, shadow.Path, TaskDependencyDependsOn); len(got) != 0 {
		t.Fatalf("shadow self declaration survived: %v", got)
	}
	if analysis.After.Health != GraphBroken || !hasRepairProblem(analysis.After.Problems, ProblemDuplicateTaskID) {
		t.Fatalf("duplicate identity should remain residual: %+v", analysis.After)
	}
}

func TestTaskGraphRepairPreservationRejectsUnauthorizedProspectiveRemoval(t *testing.T) {
	first := graphRecord("repair-proof-first", domain.StatusCompleted)
	second := graphRecord("repair-proof-second", domain.StatusCompleted)
	owner := graphRecord("repair-proof-owner", domain.StatusNextUp, first.ID, second.ID)
	owner.DependsOn = append(owner.DependsOn, owner.ID)
	before := NewTaskGraph([]domain.Task{owner, first, second}, nil)
	plan, err := PlanTaskGraphRepair(before, TaskGraphRepairRequest{Auto: true})
	if err != nil {
		t.Fatal(err)
	}
	forgedOwner := owner
	forgedOwner.DependsOn = []string{first.ID}
	forgedAfter := NewTaskGraph([]domain.Task{forgedOwner, first, second}, nil)
	removed, err := selectedRepairDeclarations(before, plan.Operations)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRepairPreservation(before, forgedAfter, plan.Operations, removed); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("unauthorized prospective removal error = %v", err)
	}
}

func TestExtendTaskGraphRepairAnalysisComposesSingleSourceProofs(t *testing.T) {
	first := graphRecord("repair-prefix-proof-first", domain.StatusNextUp, "first-invalid")
	second := graphRecord("repair-prefix-proof-second", domain.StatusNextUp, "second-invalid")
	graph := NewTaskGraph([]domain.Task{first, second}, nil)
	_, prefix, err := ValidateTaskGraphRepairPlan(graph, TaskGraphRepairPlan{})
	if err != nil {
		t.Fatal(err)
	}
	for _, edit := range []TaskGraphSourceEdit{
		{Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(first), Field: TaskDependencyDependsOn, Value: "first-invalid"},
		{Action: TaskGraphSourceDropDeclaration, Source: sourceRefForTask(second), Field: TaskDependencyDependsOn, Value: "second-invalid"},
	} {
		plan, planErr := PlanTaskGraphRepair(prefix.Prospective, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{edit}})
		if planErr != nil {
			t.Fatal(planErr)
		}
		_, step, validationErr := ValidateTaskGraphRepairPlan(prefix.Prospective, plan)
		if validationErr != nil {
			t.Fatal(validationErr)
		}
		prefix, err = ExtendTaskGraphRepairAnalysis(prefix, step)
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(prefix.SourceGroups) != 2 || len(prefix.Removed) != 2 || prefix.After.Health != GraphHealthy {
		t.Fatalf("composed repair prefix = %+v", prefix)
	}
}

func assertRepairReason(t *testing.T, defects []TaskGraphRepairDefect, reason TaskGraphRepairReason, automatic bool) {
	t.Helper()
	for _, defect := range defects {
		if defect.Reason == reason {
			if defect.Automatic != automatic {
				t.Fatalf("%s automatic=%v, want %v", reason, defect.Automatic, automatic)
			}
			return
		}
	}
	t.Fatalf("missing repair reason %s in %+v", reason, defects)
}

func hasRepairReason(defects []TaskGraphRepairDefect, reason TaskGraphRepairReason) bool {
	for _, defect := range defects {
		if defect.Reason == reason {
			return true
		}
	}
	return false
}

func hasRepairProblem(problems []GraphProblem, code GraphProblemCode) bool {
	for _, problem := range problems {
		if problem.Code == code {
			return true
		}
	}
	return false
}

func repairRawValues(t *testing.T, graph *TaskGraph, path string, field TaskDependencyField) []string {
	t.Helper()
	record := sourceRecordFor(t, mustSourceRecords(t, graph), path)
	for _, sourceField := range record.Fields {
		if sourceField.Field == field {
			return sourceField.Values
		}
	}
	return nil
}
