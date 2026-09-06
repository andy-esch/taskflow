package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestTaskGraphRepairHumanReportsRecoveryEvidence(t *testing.T) {
	edit := core.TaskGraphSourceEdit{
		Action: core.TaskGraphSourceDropDeclaration,
		Source: core.TaskGraphSourceRef{TaskID: "6g0000000001", Location: "planning/tasks/6g0000000001-owner.md"},
		Field:  core.TaskDependencyDependsOn, Value: "raw invalid value", Occurrence: 0,
	}
	receipt := core.TaskGraphRepairReceipt{
		Changed: true, Committed: true, InitialHealth: core.GraphBroken, FinalHealth: core.GraphBroken,
		Selected:   []core.TaskGraphSourceEdit{edit},
		Operations: []core.TaskGraphRepairOperation{{Edit: edit, Reason: core.RepairInvalidID}},
		Residual: []core.TaskGraphRepairDefect{{Reason: core.RepairDirectEdit, Problem: core.GraphProblem{
			Code: core.ProblemUnreadable, Message: "task source is unreadable",
		}}},
		AppliedFiles: []string{"planning/tasks/6g0000000001-owner.md"},
		Impacts: []core.TaskGraphStateImpact{{TaskID: "6g0000000001",
			Before: core.TaskGraphState{Role: core.RoleQueued, Gate: core.GateBroken, Inconsistent: true},
			After:  core.TaskGraphState{Role: core.RoleQueued, Gate: core.GateClear, Eligible: true},
		}},
		ThreadImpacts: []core.ThreadProjectionImpact{{ThreadID: "6g0000000002", Slug: "repair-thread",
			ChangedTaskIDs: []string{"6g0000000001"},
			Before:         core.ThreadView{ProjectionHealth: core.GraphBroken},
			After:          core.ThreadView{ProjectionHealth: core.GraphHealthy},
		}},
		IncompleteThreads: []core.ThreadReadProblem{{Location: "planning/threads/broken.md", Message: "bad yaml"}},
	}
	var out bytes.Buffer
	if err := TaskGraphRepairHuman(&out, NewStyle(false), receipt, wire.WorkspaceJSON{PlanningRoot: "/repo/planning"}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"graph: broken -> broken · changed=true · committed=true",
		"workspace: /repo/planning",
		"explicit planning/tasks/6g0000000001-owner.md:depends_on=\"raw invalid value\"#0",
		"applied files: planning/tasks/6g0000000001-owner.md",
		"task 6g0000000001 state queued/broken sound=false eligible=false drained=false inconsistent=true -> queued/clear sound=false eligible=true drained=false inconsistent=false",
		"Thread repair-thread (6g0000000002) projection broken -> healthy",
		"needs-direct-edit",
		"planning/threads/broken.md: bad yaml",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("repair output missing %q:\n%s", want, out.String())
		}
	}
}

func TestGraphDiagnosticsHumanDeduplicatesRepeatedRepositoryProblems(t *testing.T) {
	problems := []core.GraphProblem{
		{Code: core.ProblemCycle, TaskID: "6g0000000001", Message: "dependency cycle: one -> two -> one"},
		{Code: core.ProblemCycle, TaskID: "6g0000000002", Message: "dependency cycle: one -> two -> one"},
		{Code: core.ProblemMissingDependency, TaskID: "6g0000000003", Message: "missing dependency"},
	}
	var out bytes.Buffer
	graphDiagnosticsHuman(&out, NewStyle(false), problems, nil)
	if got := strings.Count(out.String(), "dependency cycle:"); got != 1 {
		t.Fatalf("cycle rendered %d times:\n%s", got, out.String())
	}
	if !strings.Contains(out.String(), "missing dependency") {
		t.Fatalf("distinct repository problem was lost:\n%s", out.String())
	}
}

func TestGraphDiagnosticsHumanOnlyOffersMigrationForSafeLegacyFields(t *testing.T) {
	legacy := []core.LegacyDependencyDiagnostic{
		{TaskID: "safe", Field: "blocked_by", References: []core.LegacyReference{{Resolution: core.LegacyResolved}}},
		{TaskID: "broken", Field: "dependencies", References: []core.LegacyReference{{Resolution: core.LegacyMissing}}},
	}
	var out bytes.Buffer
	graphDiagnosticsHuman(&out, NewStyle(false), nil, legacy)
	for _, want := range []string{
		"legacy blocked_by on safe; run task depend migrate",
		"legacy dependencies on broken; run task depend repair, then task depend migrate",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("diagnostics missing %q:\n%s", want, out.String())
		}
	}
}
