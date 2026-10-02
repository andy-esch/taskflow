package core

import (
	"errors"
	"slices"
	"strconv"
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
	graph := localTaskGraph([]domain.Task{owner, prerequisite}, nil)

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

func TestTaskGraphRepairDoesNotSelectOpaqueLocations(t *testing.T) {
	task := graphRecord("repair-opaque-source", domain.StatusNextUp, "invalid-token")
	task.Path = ""
	for _, source := range []RecordSource{
		{ID: task.ID, Location: "db://tasks/opaque"},
		{ID: task.ID, Location: "tasks/path-shaped.md"},
		{ID: task.ID, Location: "tasks/path-shaped.md", LocationIsPath: true},
	} {
		t.Run(source.Location+"/path-hint="+strconv.FormatBool(source.LocationIsPath), func(t *testing.T) {
			graph := NewTaskGraphRead(TaskGraphRead{Records: []LoadedRecord[domain.Task]{{Value: task, Source: source}}})
			diagnosis, err := DiagnoseTaskGraphRepair(graph)
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnosis.Defects) != 1 || diagnosis.Defects[0].Repairable || diagnosis.Defects[0].Automatic ||
				diagnosis.Defects[0].Problem.Code != ProblemRepairUnavailable ||
				diagnosis.Defects[0].Problem.Location != source.Location {
				t.Fatalf("pathless repair diagnosis = %+v", diagnosis.Defects)
			}
			auto, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Auto: true})
			if err != nil || len(auto.Operations) != 0 {
				t.Fatalf("pathless auto plan = %+v, %v", auto, err)
			}
			for _, selector := range []TaskGraphSourceRef{{Location: source.Location}, {TaskID: task.ID}} {
				_, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{{
					Action: TaskGraphSourceDropDeclaration, Source: selector,
					Field: TaskDependencyDependsOn, Value: "invalid-token",
				}}})
				if !errors.Is(err, domain.ErrValidation) {
					t.Fatalf("opaque/pathless source %+v error = %v, want validation", selector, err)
				}
			}
		})
	}

	// Even when the readable location is opaque, an independently supplied local
	// path can select a repair. It is that path, not the URI, in the plan.
	task.Path = "/planning/tasks/local-copy.md"
	mixed := NewTaskGraphRead(TaskGraphRead{GuardedRecords: []VersionedRecord[domain.Task]{{
		Record:    LoadedRecord[domain.Task]{Value: task, Source: RecordSource{ID: task.ID, Location: "db://tasks/opaque"}},
		LocalPath: task.Path,
	}}})
	plan, err := PlanTaskGraphRepair(mixed, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: TaskGraphSourceRef{LocalPath: task.Path},
		Field: TaskDependencyDependsOn, Value: "invalid-token",
	}}})
	if err != nil || len(plan.Operations) != 1 || plan.Operations[0].Edit.Source.LocalPath != task.Path ||
		plan.Operations[0].Edit.Source.Location != "db://tasks/opaque" {
		t.Fatalf("mixed source plan = %+v, %v", plan, err)
	}
	_, analysis, err := ValidateTaskGraphRepairPlan(mixed, plan)
	if err != nil || analysis.After.Health != GraphHealthy ||
		analysis.Prospective.sourceRefs[0].Location != "db://tasks/opaque" ||
		analysis.Prospective.sourceRefs[0].LocalPath != task.Path {
		t.Fatalf("mixed source repair analysis = %+v, %v", analysis, err)
	}
}

func TestCompatibilityTaskGraphsDoNotPromoteSemanticPathToRepairAuthority(t *testing.T) {
	for _, path := range []string{"db://tasks/opaque", "tasks/path-shaped.md"} {
		t.Run(path, func(t *testing.T) {
			task := graphRecord("compatibility-opaque-repair", domain.StatusNextUp, "invalid-token")
			task.Path = path
			for name, graph := range map[string]*TaskGraph{
				"constructor":      NewTaskGraph([]domain.Task{task}, nil),
				"tasks projection": NewTaskGraphRead(TaskGraphRead{Tasks: []domain.Task{task}}),
			} {
				t.Run(name, func(t *testing.T) {
					diagnosis, err := DiagnoseTaskGraphRepair(graph)
					if err != nil {
						t.Fatal(err)
					}
					if len(diagnosis.Defects) != 1 || diagnosis.Defects[0].Repairable ||
						diagnosis.Defects[0].Problem.Code != ProblemRepairUnavailable ||
						diagnosis.Defects[0].Target.Source.LocalPath != "" {
						t.Fatalf("semantic Path gained repair authority: %+v", diagnosis.Defects)
					}
					auto, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Auto: true})
					if err != nil || len(auto.Operations) != 0 {
						t.Fatalf("compatibility auto plan = %+v, %v", auto, err)
					}
				})
			}
		})
	}
}

func TestPathlessRepairDefectsKeepDeclarationIdentity(t *testing.T) {
	owner := graphRecord("pathless-two-defects", domain.StatusNextUp, "invalid-one", "invalid-two")
	owner.Path = ""
	graph := NewTaskGraphRead(TaskGraphRead{Records: []LoadedRecord[domain.Task]{{
		Value: owner, Source: RecordSource{ID: owner.ID, Location: "db://tasks/owner"},
	}}})
	diagnosis, err := DiagnoseTaskGraphRepair(graph)
	if err != nil || len(diagnosis.Defects) != 2 {
		t.Fatalf("pathless diagnosis = %+v, %v", diagnosis, err)
	}
	for _, defect := range diagnosis.Defects {
		if defect.Repairable || defect.Target.Field != TaskDependencyDependsOn ||
			defect.Target.Source.Location != "db://tasks/owner" {
			t.Fatalf("unattributed pathless defect: %+v", defect)
		}
	}
	addressed := repairDefectDifference(diagnosis.Defects, diagnosis.Defects[1:])
	if len(addressed) != 1 || addressed[0].Target.Value != diagnosis.Defects[0].Target.Value {
		t.Fatalf("addressed defect was conflated with residual: %+v", addressed)
	}
	other := diagnosis.Defects[0]
	other.Target.Source.Location = "db://tasks/shadow"
	if repairDefectIdentity(other) == repairDefectIdentity(diagnosis.Defects[0]) {
		t.Fatal("distinct physical occurrences shared a defect identity")
	}
}

func TestMixedSourceLocationIsAStaleContextCheck(t *testing.T) {
	owner := graphRecord("mixed-stale-context", domain.StatusNextUp, "invalid-token")
	graph := NewTaskGraphRead(TaskGraphRead{GuardedRecords: []VersionedRecord[domain.Task]{{
		Record:    LoadedRecord[domain.Task]{Value: owner, Source: RecordSource{ID: owner.ID, Location: "db://tasks/current"}},
		LocalPath: owner.Path,
	}}})
	stale := TaskGraphSourceRef{TaskID: owner.ID, LocalPath: owner.Path, Location: "db://tasks/old"}
	if _, _, err := resolveSourceTask(graph.sourceRefs, stale); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("stale graph source context = %v, want conflict", err)
	}
	_, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: stale,
		Field: TaskDependencyDependsOn, Value: "invalid-token",
	}}})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("stale repair source context = %v, want conflict", err)
	}
}

func TestMixedSourceLocationRetainsLegacyRepairAttribution(t *testing.T) {
	owner := graphRecord("mixed-legacy-owner", domain.StatusNextUp)
	owner.LegacyBlockedBy = []string{"missing-human-intent"}
	owner.LegacyDependencyFields = []string{"blocked_by"}
	graph := NewTaskGraphRead(TaskGraphRead{GuardedRecords: []VersionedRecord[domain.Task]{{
		Record:    LoadedRecord[domain.Task]{Value: owner, Source: RecordSource{ID: owner.ID, Location: "db://tasks/legacy"}},
		LocalPath: owner.Path,
	}}})
	diagnosis, err := DiagnoseTaskGraphRepair(graph)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, defect := range diagnosis.Defects {
		if defect.Reason == RepairLegacyMissing && defect.Target.Source.Location == "db://tasks/legacy" &&
			defect.Target.Source.LocalPath == owner.Path && defect.Target.Value == "missing-human-intent" {
			found = true
		}
	}
	if !found {
		t.Fatalf("mixed-location legacy defect was lost: %+v", diagnosis.Defects)
	}
}

func TestURIValuedTransitionalTaskPathDoesNotOfferLocalRepair(t *testing.T) {
	owner := graphRecord("uri-path-owner", domain.StatusNextUp, "invalid-token")
	owner.Path = "db://tasks/owner"
	graph := NewTaskGraphRead(TaskGraphRead{Records: []LoadedRecord[domain.Task]{{
		Value: owner, Source: RecordSource{ID: owner.ID, Location: owner.Path},
	}}})
	diagnosis, err := DiagnoseTaskGraphRepair(graph)
	if err != nil || len(diagnosis.Defects) != 1 || diagnosis.Defects[0].Repairable {
		t.Fatalf("URI-valued domain path exposed local repair: %+v, %v", diagnosis.Defects, err)
	}
	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Auto: true})
	if err != nil || len(plan.Operations) != 0 {
		t.Fatalf("URI-valued domain path entered auto plan: %+v, %v", plan, err)
	}
}

func TestTaskGraphRepairAutoHandlesRepeatedSelfDeclarationsWithoutOrderDependence(t *testing.T) {
	owner := graphRecord("repair-repeated-self", domain.StatusNextUp)
	owner.DependsOn = []string{owner.ID, owner.ID}
	graph := localTaskGraph([]domain.Task{owner}, nil)
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
	graph := localTaskGraph([]domain.Task{owner}, nil)
	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{
		Auto: true,
		Edits: []TaskGraphSourceEdit{{
			Action: TaskGraphSourceDropDeclaration, Source: localSourceRefForTask(owner),
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
	graph := localTaskGraph([]domain.Task{owner}, nil)
	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: localSourceRefForTask(owner),
		Field: TaskDependencyDependsOn, Value: "invalid-token",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	plan.Selections = []TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: localSourceRefForTask(owner),
		Field: TaskDependencyDependsOn, Value: dangling,
	}}
	if _, _, err := ValidateTaskGraphRepairPlan(graph, plan); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "does not exactly account") {
		t.Fatalf("unaccounted selected intent error = %v", err)
	}
}

func TestTaskGraphRepairValidationRejectsFalseAutomaticProvenance(t *testing.T) {
	owner := graphRecord("repair-false-auto", domain.StatusNextUp, "invalid-token")
	graph := localTaskGraph([]domain.Task{owner}, nil)
	plan, err := PlanTaskGraphRepair(graph, TaskGraphRepairRequest{Edits: []TaskGraphSourceEdit{{
		Action: TaskGraphSourceDropDeclaration, Source: localSourceRefForTask(owner),
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
	graph := localTaskGraph([]domain.Task{owner}, nil)
	for _, edit := range []TaskGraphSourceEdit{
		{Action: TaskGraphSourceDropDeclaration, Source: localSourceRefForTask(owner), Field: TaskDependencyDependsOn, Value: "missing", Occurrence: -1},
		{Action: TaskGraphSourceDedupe, Source: localSourceRefForTask(owner), Field: TaskDependencyBlockedBy, Value: "missing"},
		{Action: TaskGraphSourceDedupe, Source: localSourceRefForTask(owner), Field: TaskDependencyDependsOn, Value: "missing", Occurrence: 1},
		{Action: TaskGraphSourceDropEmptyField, Source: localSourceRefForTask(owner), Field: TaskDependencyDependsOn},
		{Action: TaskGraphSourceDropEmptyField, Source: localSourceRefForTask(owner), Field: TaskDependencyBlockedBy, Value: "missing"},
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
	graph := localTaskGraph([]domain.Task{owner, prerequisite}, nil)
	source := localSourceRefForTask(owner)
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
	graph := localTaskGraph([]domain.Task{beta, alpha}, nil)

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
			graph := localTaskGraph([]domain.Task{owner}, nil)
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
	graph := localTaskGraph([]domain.Task{owner, dependent}, nil)
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
	graph := localTaskGraph([]domain.Task{owner, second, first}, nil)
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
	graph := localTaskGraph([]domain.Task{shadow, representative}, nil)

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
	before := localTaskGraph([]domain.Task{owner, first, second}, nil)
	plan, err := PlanTaskGraphRepair(before, TaskGraphRepairRequest{Auto: true})
	if err != nil {
		t.Fatal(err)
	}
	forgedOwner := owner
	forgedOwner.DependsOn = []string{first.ID}
	forgedAfter := localTaskGraph([]domain.Task{forgedOwner, first, second}, nil)
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
	graph := localTaskGraph([]domain.Task{first, second}, nil)
	_, prefix, err := ValidateTaskGraphRepairPlan(graph, TaskGraphRepairPlan{})
	if err != nil {
		t.Fatal(err)
	}
	for _, edit := range []TaskGraphSourceEdit{
		{Action: TaskGraphSourceDropDeclaration, Source: localSourceRefForTask(first), Field: TaskDependencyDependsOn, Value: "first-invalid"},
		{Action: TaskGraphSourceDropDeclaration, Source: localSourceRefForTask(second), Field: TaskDependencyDependsOn, Value: "second-invalid"},
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
