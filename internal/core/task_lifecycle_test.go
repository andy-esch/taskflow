package core

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestValidateTaskLifecycleStartEligibilityAndTypedOverride(t *testing.T) {
	tests := []struct {
		name       string
		prereq     domain.Status
		gate       GateState
		wantReason BlockerReason
	}{
		{name: "clear", prereq: domain.StatusCompleted, gate: GateClear},
		{name: "unfinished", prereq: domain.StatusNextUp, gate: GateBlocked, wantReason: BlockerNotStarted},
		{name: "parked", prereq: domain.StatusDeferred, gate: GateBlocked, wantReason: BlockerParked},
		{name: "withdrawn", prereq: domain.StatusDeprecated, gate: GateBroken, wantReason: BlockerWithdrawn},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prerequisite := graphRecord("prerequisite-"+tc.name, tc.prereq)
			target := graphRecord("target-"+tc.name, domain.StatusReadyToStart, prerequisite.ID)
			graph := NewTaskGraph([]domain.Task{target, prerequisite}, nil)
			plan := TaskLifecyclePlan{TaskID: target.ID, To: domain.StatusInProgress}

			validated, analysis, err := ValidateTaskLifecyclePlan(graph, plan, "")
			if tc.gate == GateClear {
				if err != nil || validated != plan || analysis.OverrideApplied {
					t.Fatalf("clear start = plan %+v analysis %+v err %v", validated, analysis, err)
				}
				if analysis.From != domain.StatusReadyToStart || analysis.After.Role != RoleInFlight {
					t.Fatalf("clear start analysis = %+v", analysis)
				}
				return
			}

			var eligibility *TaskEligibilityError
			if !errors.As(err, &eligibility) || eligibility.State.Gate != tc.gate {
				t.Fatalf("default refusal = %#v, want eligibility gate %s", err, tc.gate)
			}
			if len(eligibility.Blockers) != 1 || eligibility.Blockers[0].Reason != tc.wantReason {
				t.Fatalf("blockers = %+v, want reason %s", eligibility.Blockers, tc.wantReason)
			}

			plan.Override = TaskLifecycleOverrideDependencyGate
			_, forced, err := ValidateTaskLifecyclePlan(graph, plan, "")
			if err != nil || !forced.OverrideApplied || forced.After.Role != RoleInFlight {
				t.Fatalf("forced start analysis=%+v err=%v", forced, err)
			}
			if !reflect.DeepEqual(forced.OutstandingBlockers, eligibility.Blockers) {
				t.Fatalf("forced blockers = %+v, want refusal blockers %+v", forced.OutstandingBlockers, eligibility.Blockers)
			}
		})
	}
}

func TestValidateTaskLifecycleForceCannotBypassRoleOrBrokenRepository(t *testing.T) {
	queued := graphRecord("queued", domain.StatusNextUp)
	graph := NewTaskGraph([]domain.Task{queued}, nil)
	_, _, err := ValidateTaskLifecyclePlan(graph, TaskLifecyclePlan{
		TaskID: queued.ID, To: domain.StatusInProgress, Override: TaskLifecycleOverrideDependencyGate,
	}, "")
	var eligibility *TaskEligibilityError
	if !errors.As(err, &eligibility) || eligibility.State.Role != RoleQueued {
		t.Fatalf("force should not bypass queued role: %v", err)
	}

	missing := testutil.TaskID("missing")
	candidate := graphRecord("candidate", domain.StatusReadyToStart, missing)
	broken := NewTaskGraph([]domain.Task{candidate}, nil)
	_, _, err = ValidateTaskLifecyclePlan(broken, TaskLifecyclePlan{
		TaskID: candidate.ID, To: domain.StatusInProgress, Override: TaskLifecycleOverrideDependencyGate,
	}, "")
	if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "repository task graph is broken") {
		t.Fatalf("force should fail closed on a broken repository: %v", err)
	}
}

func TestValidateTaskLifecycleEveryPersistedStatusEnteringInProgress(t *testing.T) {
	for _, status := range domain.AllStatuses() {
		t.Run(string(status), func(t *testing.T) {
			task := graphRecord("status-"+string(status), status)
			plan := TaskLifecyclePlan{TaskID: task.ID, To: domain.StatusInProgress}
			_, analysis, err := ValidateTaskLifecyclePlan(NewTaskGraph([]domain.Task{task}, nil), plan, "")
			switch status {
			case domain.StatusReadyToStart:
				if err != nil || analysis.After.Role != RoleInFlight {
					t.Fatalf("candidate start analysis=%+v err=%v", analysis, err)
				}
			case domain.StatusInProgress:
				if err != nil || analysis.From != domain.StatusInProgress || analysis.Before != analysis.After {
					t.Fatalf("idempotent start analysis=%+v err=%v", analysis, err)
				}
			default:
				var eligibility *TaskEligibilityError
				if !errors.As(err, &eligibility) {
					t.Fatalf("status %s should require an explicit move to ready-to-start first: %v", status, err)
				}
			}
		})
	}
}

func TestValidateTaskLifecycleRejectsInvalidResultingActiveState(t *testing.T) {
	task := graphRecord("missing-description", domain.StatusReadyToStart)
	task.Description = ""
	_, _, err := ValidateTaskLifecyclePlan(NewTaskGraph([]domain.Task{task}, nil), TaskLifecyclePlan{
		TaskID: task.ID, To: domain.StatusInProgress,
	}, "")
	if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "description is required") {
		t.Fatalf("start should reject a lint-invalid resulting active task: %v", err)
	}
}

func TestValidateTaskLifecycleCompletionOverrideIsSeparate(t *testing.T) {
	task := graphRecord("complete-me", domain.StatusInProgress)
	graph := NewTaskGraph([]domain.Task{task}, nil)
	body := "# Complete me\n\n## Acceptance criteria\n\n- [ ] unexplained\n"

	_, _, err := ValidateTaskLifecyclePlan(graph, TaskLifecyclePlan{TaskID: task.ID, To: domain.StatusCompleted}, body)
	if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "acceptance criterion") {
		t.Fatalf("unforced completion = %v", err)
	}
	_, _, err = ValidateTaskLifecyclePlan(graph, TaskLifecyclePlan{
		TaskID: task.ID, To: domain.StatusCompleted, Override: TaskLifecycleOverrideDependencyGate,
	}, body)
	if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "applies only when entering in-progress") {
		t.Fatalf("dependency override should not complete: %v", err)
	}
	_, analysis, err := ValidateTaskLifecyclePlan(graph, TaskLifecyclePlan{
		TaskID: task.ID, To: domain.StatusCompleted, Override: TaskLifecycleOverrideAcceptanceCriteria,
	}, body)
	if err != nil || !analysis.OverrideApplied || len(analysis.OutstandingBlockers) != 0 {
		t.Fatalf("completion override analysis=%+v err=%v", analysis, err)
	}
}

func TestValidateTaskLifecycleNoOpStillValidatesOverrideScope(t *testing.T) {
	task := graphRecord("candidate", domain.StatusReadyToStart)
	graph := NewTaskGraph([]domain.Task{task}, nil)
	for _, override := range []TaskLifecycleOverride{
		TaskLifecycleOverrideDependencyGate,
		TaskLifecycleOverrideAcceptanceCriteria,
	} {
		_, _, err := ValidateTaskLifecyclePlan(graph, TaskLifecyclePlan{
			TaskID: task.ID, To: domain.StatusReadyToStart, Override: override,
		}, "")
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("same-status override %q error = %v, want validation", override, err)
		}
	}

	_, analysis, err := ValidateTaskLifecyclePlan(graph, TaskLifecyclePlan{
		TaskID: task.ID, To: domain.StatusReadyToStart,
	}, "")
	if err != nil || analysis.From != domain.StatusReadyToStart || analysis.Before != analysis.After {
		t.Fatalf("ordinary same-status no-op = %+v, %v", analysis, err)
	}
}

func TestValidateTaskLifecycleReopenReportsUnsoundDescendants(t *testing.T) {
	upstream := graphRecord("upstream", domain.StatusCompleted)
	direct := graphRecord("direct", domain.StatusCompleted, upstream.ID)
	transitive := graphRecord("transitive", domain.StatusCompleted, direct.ID)
	graph := NewTaskGraph([]domain.Task{transitive, direct, upstream}, nil)

	_, analysis, err := ValidateTaskLifecyclePlan(graph, TaskLifecyclePlan{
		TaskID: upstream.ID, To: domain.StatusReadyToStart,
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	if analysis.From != domain.StatusCompleted || len(analysis.Impacts) != 2 {
		t.Fatalf("reopen analysis = %+v", analysis)
	}
	for _, impact := range analysis.Impacts {
		if impact.Before.Gate != GateClear || impact.After.Gate != GateBlocked ||
			impact.Before.Inconsistent || !impact.After.Inconsistent || impact.After.SoundlyCompleted {
			t.Errorf("unexpected descendant impact: %+v", impact)
		}
	}
	directCount := 0
	for _, impact := range analysis.Impacts {
		if impact.Direct {
			directCount++
		}
	}
	if directCount != 1 {
		t.Fatalf("direct/transitive impacts = %+v", analysis.Impacts)
	}
}

func TestValidateCreateAndStartRequiresHealthyCandidateAndProducesExactFrom(t *testing.T) {
	graph := NewTaskGraph(nil, nil)
	task := graphRecord("new", domain.StatusReadyToStart)
	task.Path = ""
	task.Description = "new task"
	task.Tags = []string{"test"}
	plan := TaskLifecyclePlan{To: domain.StatusInProgress, Create: &TaskLifecycleCreation{Task: task, Body: "# New\n"}}
	_, analysis, err := ValidateTaskLifecyclePlan(graph, plan, "")
	if err != nil {
		t.Fatal(err)
	}
	if analysis.From != domain.StatusReadyToStart || analysis.Before.Role != RoleCandidate || analysis.After.Role != RoleInFlight {
		t.Fatalf("create-and-start analysis = %+v", analysis)
	}

	plan.Override = TaskLifecycleOverrideDependencyGate
	if _, _, err := ValidateTaskLifecyclePlan(graph, plan, ""); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("create-and-start must reject force: %v", err)
	}
}
