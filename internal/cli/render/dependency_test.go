package render

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestTaskGraphRepairHumanReportsRecoveryEvidence(t *testing.T) {
	edit := core.TaskGraphSourceEdit{
		Action: core.TaskGraphSourceDropDeclaration,
		Source: core.TaskGraphSourceRef{TaskID: "6g0000000001", Location: "planning/tasks/6g0000000001-owner.md", LocalPath: "planning/tasks/6g0000000001-owner.md"},
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

func TestGraphDiagnosticsHumanAttributesOpaqueOccurrences(t *testing.T) {
	problems := []core.GraphProblem{
		{Code: core.ProblemInvalidDependencyID, Message: "invalid dependency", Location: "db://tasks/a"},
		{Code: core.ProblemInvalidDependencyID, Message: "invalid dependency", Location: "db://tasks/b"},
	}
	var out bytes.Buffer
	graphDiagnosticsHuman(&out, NewStyle(false), problems, nil)
	for _, location := range []string{"db://tasks/a", "db://tasks/b"} {
		if !strings.Contains(out.String(), "(location: "+location+")") {
			t.Fatalf("missing distinct source %s:\n%s", location, out.String())
		}
	}
}

func TestTaskGraphRepairHumanAttributesNonRepairableDeclaration(t *testing.T) {
	defect := core.TaskGraphRepairDefect{
		Reason: core.RepairInvalidID,
		Target: core.TaskGraphSourceEdit{Action: core.TaskGraphSourceDropDeclaration,
			Source: core.TaskGraphSourceRef{TaskID: "6g0000000001", Location: "db://tasks/a"},
			Field:  core.TaskDependencyDependsOn, Value: "invalid-token", Occurrence: 1},
		Problem: core.GraphProblem{Code: core.ProblemRepairUnavailable, Message: "source has no explicit local repair path"},
	}
	var out bytes.Buffer
	if err := TaskGraphRepairHuman(&out, NewStyle(false), core.TaskGraphRepairReceipt{Residual: []core.TaskGraphRepairDefect{defect}}, wire.WorkspaceJSON{}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"declaration 6g0000000001:depends_on=\"invalid-token\"#1", "location db://tasks/a"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "tskflwctl task depend repair --drop") {
		t.Fatalf("non-repairable declaration gained a repair command:\n%s", out.String())
	}
}

func TestGraphDiagnosticsHumanOnlyOffersMigrationForSafeLegacyFields(t *testing.T) {
	legacy := []core.LegacyDependencyDiagnostic{
		{TaskID: "safe", Field: "blocked_by", References: []core.LegacyReference{{Resolution: core.LegacyResolved}}},
		{TaskID: "broken", Field: "dependencies", References: []core.LegacyReference{{Value: "gone", Resolution: core.LegacyMissing}}},
	}
	var out bytes.Buffer
	graphDiagnosticsHuman(&out, NewStyle(false), nil, legacy)
	for _, want := range []string{
		"legacy blocked_by on safe; run task depend migrate",
		"legacy dependencies on broken (\"gone\": missing); run task depend repair, then task depend migrate",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("diagnostics missing %q:\n%s", want, out.String())
		}
	}
}

func TestGraphLegacyDiagnosticsHumanKeepOpaqueContextOnOneSafeLine(t *testing.T) {
	location := "\x1b]0;fake title\adb://tasks/outside\nFORGED\x1b[31m"
	legacy := core.LegacyDependencyDiagnostic{TaskID: "source-id", Field: "blocks", Location: location}
	var out bytes.Buffer
	graphDiagnosticsHuman(&out, NewStyle(false), nil, []core.LegacyDependencyDiagnostic{legacy})
	want := "⚠ legacy blocks on source-id (empty field); run task depend migrate (location: db://tasks/outside FORGED)\n"
	if out.String() != want || legacy.Location != location {
		t.Fatalf("new legacy context was not safely presented without rewriting it: %q", out.String())
	}
}

func TestGraphDiagnosticsHumanKeepProducedStructuralContextOnOneSafeLine(t *testing.T) {
	for _, local := range []bool{false, true} {
		name := "opaque-location"
		if local {
			name = "local-path-in-message"
		}
		t.Run(name, func(t *testing.T) {
			location := "db://outside\nFORGED-READY\x1b[2J"
			member := domain.Task{ID: "6g0000000001", Slug: "member", Status: domain.StatusNextUp}
			owner := domain.Task{ID: "6g0000000002", Slug: "outside", Status: domain.StatusCompleted,
				LegacyBlockedBy: []string{"gone"}, DependsOn: []string{"invalid-token"}}
			source := core.RecordSource{ID: owner.ID, Location: location}
			localPath := ""
			if local {
				source.Location = "tasks/6g0000000002-outside\nFORGED-READY\x1b[2J.md"
				source.LocationIsPath = true
				localPath = source.Location
			}
			records := []core.VersionedRecord[domain.Task]{
				{Record: core.LoadedRecord[domain.Task]{Value: member, Source: core.RecordSource{ID: member.ID}}},
				{Record: core.LoadedRecord[domain.Task]{Value: owner, Source: source}, LocalPath: localPath},
			}
			if local {
				// Duplicate-source messages embed raw paths, separately from the optional
				// location suffix. Sanitizing only that suffix cannot protect this case.
				records = append(records, core.VersionedRecord[domain.Task]{LocalPath: "tasks/duplicate.md", Record: core.LoadedRecord[domain.Task]{
					Value:  domain.Task{ID: owner.ID, Slug: "duplicate", Status: domain.StatusCompleted},
					Source: core.RecordSource{ID: owner.ID, Location: "tasks/duplicate.md", LocationIsPath: true},
				}})
			}
			graph := core.NewTaskGraphRead(core.TaskGraphRead{GuardedRecords: records})
			view := core.ProjectThread(domain.Thread{ID: "6g0000000003", Slug: "delivery",
				Status: domain.ThreadStatusInProgress, Description: "Delivery", Goal: "Ship",
				Created: "2026-10-11", Tasks: []string{member.ID}}, graph)
			problems, legacy := graph.Problems(), graph.LegacyDiagnostics()
			if view.GraphHealth != core.GraphBroken || len(view.Frontier) != 0 || len(legacy) != 1 {
				t.Fatalf("hostile graph did not reach the intended diagnostic producer: %+v", view)
			}
			if local {
				found := false
				for _, problem := range problems {
					found = found || (problem.Code == core.ProblemDuplicateTaskID && strings.Contains(problem.Message, "\nFORGED-READY\x1b[2J"))
				}
				if !found {
					t.Fatal("fixture failed to produce raw local-path content in a structural message")
				}
			}
			for surface, render := range map[string]func(io.Writer) error{
				"frontier": func(w io.Writer) error { return ThreadFrontierHuman(w, NewStyle(false), view) },
				"blockers": func(w io.Writer) error {
					return TaskBlockersHuman(w, NewStyle(false), core.TaskBlockersResult{TaskID: member.ID, Task: member,
						State: graph.State(member.ID), Health: graph.Health(), Problems: problems, Legacy: legacy})
				},
				"unblocks": func(w io.Writer) error {
					return TaskUnblocksHuman(w, NewStyle(false), core.TaskUnblocksResult{TaskID: member.ID, Task: member,
						State: graph.State(member.ID), Health: graph.Health(), Problems: problems, Legacy: legacy})
				},
			} {
				t.Run(surface, func(t *testing.T) {
					var out bytes.Buffer
					if err := render(&out); err != nil {
						t.Fatal(err)
					}
					if strings.Contains(out.String(), "\nFORGED-READY") || strings.Contains(out.String(), "\x1b") {
						t.Fatalf("structural evidence forged a row or terminal control: %q", out.String())
					}
					if !strings.Contains(out.String(), "invalid-dependency-id:") || !strings.Contains(out.String(), "legacy blocked_by") {
						t.Fatalf("sanitization hid structural or legacy evidence: %s", out.String())
					}
				})
			}
			if !reflect.DeepEqual(graph.Problems(), problems) || !reflect.DeepEqual(graph.LegacyDiagnostics(), legacy) {
				t.Fatal("terminal rendering rewrote core evidence")
			}
			var out bytes.Buffer
			if err := ThreadFrontierJSON(&out, view); err != nil {
				t.Fatal(err)
			}
			var decoded wire.ThreadFrontierEnvelope
			if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded.View.GraphProblems, wire.ToGraphProblemsJSON(problems)) || decoded.View.GraphLegacy[0].Path != localPath || (!local && decoded.View.GraphLegacy[0].Location != location) {
				t.Fatal("machine output lost raw source context")
			}
		})
	}
}
