package core

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

var threadApplyNow = time.Date(2026, 8, 30, 14, 0, 0, 0, time.UTC)

func applySnapshot(repoID string, tasks ...domain.Task) ThreadApplySnapshot {
	return ThreadApplySnapshot{PlanningRepoID: repoID, Graph: NewTaskGraph(tasks, nil)}
}

func TestComposeThreadApplyPlanResolvesExistingTasksAndNonmemberGraphContext(t *testing.T) {
	context := graphRecord("bulk-context", domain.StatusNextUp)
	gate := graphRecord("bulk-boundary-gate", domain.StatusNextUp)
	first := graphRecord("bulk-first", domain.StatusNextUp)
	second := graphRecord("bulk-second", domain.StatusReadyToStart)
	threadID := testutil.TaskID("bulk-thread")
	external := false
	manifest := ThreadComposeManifest{
		Thread: ThreadComposeInput{
			Title: "Bulk delivery", Description: "Link existing tasks safely",
			Goal: "Create one resumable Thread plan", Tags: []string{"threads", "bulk", "threads"},
		},
		Nodes: []ThreadComposeNode{
			{Key: "context", TaskID: context.ID, Member: &external},
			{Key: "gate", TaskID: gate.ID, Member: &external},
			{Key: "first", TaskID: first.ID},
			{Key: "second", TaskID: second.ID},
		},
		Dependencies: []ThreadComposeDependency{
			{From: "first", To: "second"}, {From: "gate", To: "first"}, {From: "context", To: "gate"},
		},
	}
	snapshot := applySnapshot("planning-id", context, gate, first, second)
	plan, err := ComposeThreadApplyPlan(
		snapshot, manifest, "# Thread: Bulk delivery\n",
		func() string { return threadID }, threadApplyNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Schema != ThreadApplyPlanSchema || plan.PlanningRepoID != "planning-id" || plan.Thread.ID != threadID {
		t.Fatalf("plan identity = %+v", plan)
	}
	wantMembers := []string{first.ID, second.ID}
	sort.Strings(wantMembers)
	if !reflect.DeepEqual(plan.Thread.Tasks, wantMembers) || !reflect.DeepEqual(plan.Thread.Tags, []string{"bulk", "threads"}) {
		t.Fatalf("planned Thread = %+v", plan.Thread)
	}
	wantEdges := []ThreadApplyDependency{
		{From: context.ID, To: gate.ID}, {From: gate.ID, To: first.ID}, {From: first.ID, To: second.ID},
	}
	sortThreadApplyDependencies(wantEdges)
	if !reflect.DeepEqual(plan.Dependencies, wantEdges) {
		t.Fatalf("dependencies = %+v, want %+v", plan.Dependencies, wantEdges)
	}
	decision, err := PrepareThreadApply(snapshot, plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(decision.GraphPlan.TaskWrites) != 3 || decision.ThreadPlan == nil || len(decision.Operations) != 4 {
		t.Fatalf("decision = %+v", decision)
	}

	finalGraph := graphAfterTaskWrites(snapshot.Graph, decision.GraphPlan)
	view := ProjectThread(plan.Thread.domainThread(), finalGraph)
	if len(view.ExternalGates) != 1 || view.ExternalGates[0].Task.ID != gate.ID {
		t.Fatalf("external gates = %+v; only the direct membership boundary belongs in the projection", view.ExternalGates)
	}
	causal := finalGraph.CausalBlockers(first.ID)
	causalIDs := make(map[string]bool, len(causal))
	for _, blocker := range causal {
		causalIDs[blocker.TaskID] = true
	}
	if len(causal) != 2 || !causalIDs[context.ID] || !causalIDs[gate.ID] {
		t.Fatalf("causal blockers = %+v; transitive nonmember context must remain queryable", causal)
	}
}

func TestComposeThreadApplyPlanAcceptsExistingTransitiveNonmemberContext(t *testing.T) {
	context := graphRecord("existing-context", domain.StatusNextUp)
	gate := graphRecord("existing-boundary", domain.StatusNextUp, context.ID)
	member := graphRecord("existing-member", domain.StatusNextUp, gate.ID)
	nonmember := false
	manifest := ThreadComposeManifest{
		Thread: ThreadComposeInput{Title: "Existing context", Description: "Reuse the graph", Goal: "Keep roles precise"},
		Nodes: []ThreadComposeNode{
			{Key: "context", TaskID: context.ID, Member: &nonmember},
			{Key: "gate", TaskID: gate.ID, Member: &nonmember},
			{Key: "member", TaskID: member.ID},
		},
	}

	plan, err := ComposeThreadApplyPlan(
		applySnapshot("planning", context, gate, member), manifest, "body",
		func() string { return testutil.TaskID("existing-context-thread") }, threadApplyNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Dependencies) != 0 || !reflect.DeepEqual(plan.Thread.Tasks, []string{member.ID}) {
		t.Fatalf("plan = %+v", plan)
	}
}

func TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest(t *testing.T) {
	member := graphRecord("bulk-member", domain.StatusNextUp)
	nonmemberTask := graphRecord("bulk-unused-nonmember", domain.StatusCompleted)
	nonmember := false
	base := ThreadComposeManifest{
		Thread: ThreadComposeInput{Title: "Bulk", Description: "Bulk Thread", Goal: "Exercise validation"},
		Nodes:  []ThreadComposeNode{{Key: "member", TaskID: member.ID}},
	}
	tests := []struct {
		name     string
		repoID   string
		manifest ThreadComposeManifest
		want     string
	}{
		{name: "missing repository identity", manifest: base, want: "no durable id"},
		{name: "unsupported schema", repoID: "planning", manifest: func() ThreadComposeManifest {
			value := base
			value.Schema = 2
			return value
		}(), want: "unsupported"},
		{name: "duplicate task declaration", repoID: "planning", manifest: func() ThreadComposeManifest {
			value := base
			value.Nodes = append(value.Nodes, ThreadComposeNode{Key: "again", TaskID: member.ID})
			return value
		}(), want: "both declare"},
		{name: "nonmember is disconnected", repoID: "planning", manifest: func() ThreadComposeManifest {
			value := base
			value.Nodes = append(value.Nodes, ThreadComposeNode{Key: "context", TaskID: nonmemberTask.ID, Member: &nonmember})
			return value
		}(), want: "not upstream graph context"},
		{name: "nonmember is downstream", repoID: "planning", manifest: func() ThreadComposeManifest {
			value := base
			value.Nodes = append(value.Nodes, ThreadComposeNode{Key: "context", TaskID: nonmemberTask.ID, Member: &nonmember})
			value.Dependencies = []ThreadComposeDependency{{From: "member", To: "context"}}
			return value
		}(), want: "not upstream graph context"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ComposeThreadApplyPlan(
				applySnapshot(tc.repoID, member, nonmemberTask), tc.manifest, "body",
				func() string { return testutil.TaskID(tc.name) }, threadApplyNow,
			)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestPrepareThreadApplyIsAdditiveAndIdempotent(t *testing.T) {
	other := graphRecord("bulk-other", domain.StatusCompleted)
	gate := graphRecord("bulk-existing-edge", domain.StatusCompleted)
	member := graphRecord("bulk-additive-member", domain.StatusNextUp, other.ID, gate.ID)
	threadID := testutil.TaskID("bulk-additive-thread")
	plan := ThreadApplyPlan{
		Schema: ThreadApplyPlanSchema, PlanningRepoID: "planning", ComposedAt: "2026-08-30",
		Thread: ThreadApplyThread{
			ID: threadID, Slug: "bulk-additive", Status: domain.ThreadStatusUnstarted,
			Description: "Preserve unrelated edits", Goal: "Converge additive intent", Created: "2026-08-30",
			Tasks: []string{member.ID}, Body: "# Thread\n",
		},
		Dependencies: []ThreadApplyDependency{{From: gate.ID, To: member.ID}, {From: other.ID, To: member.ID}},
	}
	decision, err := PrepareThreadApply(applySnapshot("planning", other, gate, member), plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(decision.GraphPlan.TaskWrites) != 0 || decision.Operations[0].State != ThreadApplySkipped || decision.Operations[1].State != ThreadApplySkipped {
		t.Fatalf("decision = %+v", decision)
	}

	existing := plan.Thread.domainThread()
	existing.FilenameID, existing.Path = existing.ID, "/threads/"+existing.ID+".md"
	snapshot := applySnapshot("planning", other, gate, member)
	snapshot.Threads = []domain.Thread{existing}
	snapshot.ThreadBodies = map[string]string{threadID: plan.Thread.Body}
	idempotent, err := PrepareThreadApply(snapshot, plan)
	if err != nil {
		t.Fatal(err)
	}
	if idempotent.ThreadPlan != nil || idempotent.Operations[len(idempotent.Operations)-1].State != ThreadApplySkipped {
		t.Fatalf("idempotent decision = %+v", idempotent)
	}

	snapshot.ThreadBodies[threadID] = "different body"
	if _, err := PrepareThreadApply(snapshot, plan); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("same-ID different Thread error = %v", err)
	}
}

func TestPrepareThreadApplyDoesNotRewriteOwnerForExistingEdge(t *testing.T) {
	first := graphRecord("existing-order-first", domain.StatusCompleted)
	second := graphRecord("existing-order-second", domain.StatusCompleted)
	member := graphRecord("existing-order-member", domain.StatusNextUp, second.ID, first.ID)
	plan := ThreadApplyPlan{
		Schema: ThreadApplyPlanSchema, PlanningRepoID: "planning", ComposedAt: "2026-08-30",
		Thread: ThreadApplyThread{
			ID: testutil.TaskID("existing-order-thread"), Slug: "existing-order", Status: domain.ThreadStatusUnstarted,
			Description: "Preserve an existing owner", Goal: "Skip an existing edge", Created: "2026-08-30",
			Tasks: []string{member.ID}, Body: "body",
		},
		Dependencies: []ThreadApplyDependency{{From: first.ID, To: member.ID}},
	}
	decision, err := PrepareThreadApply(applySnapshot("planning", first, second, member), plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(decision.GraphPlan.TaskWrites) != 0 || decision.Operations[0].State != ThreadApplySkipped {
		t.Fatalf("decision = %+v", decision)
	}
}

func TestPrepareThreadApplyRejectsWrongIdentityAndCycle(t *testing.T) {
	a := graphRecord("bulk-cycle-a", domain.StatusNextUp)
	b := graphRecord("bulk-cycle-b", domain.StatusNextUp, a.ID)
	plan := ThreadApplyPlan{
		Schema: ThreadApplyPlanSchema, PlanningRepoID: "other", ComposedAt: "2026-08-30",
		Thread: ThreadApplyThread{
			ID: testutil.TaskID("bulk-cycle-thread"), Slug: "bulk-cycle", Status: domain.ThreadStatusUnstarted,
			Description: "Reject invalid apply", Goal: "Keep the graph sound", Created: "2026-08-30",
			Tasks: []string{a.ID, b.ID}, Body: "body",
		},
	}
	if _, err := PrepareThreadApply(applySnapshot("planning", a, b), plan); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("identity error = %v", err)
	}
	plan.PlanningRepoID = "planning"
	plan.Dependencies = []ThreadApplyDependency{{From: b.ID, To: a.ID}}
	if _, err := PrepareThreadApply(applySnapshot("planning", a, b), plan); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("cycle error = %v", err)
	}
}

func TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity(t *testing.T) {
	task := graphRecord("unsafe-plan-member", domain.StatusNextUp)
	plan := ThreadApplyPlan{
		Schema: ThreadApplyPlanSchema, PlanningRepoID: "planning", ComposedAt: "2026-08-30",
		Thread: ThreadApplyThread{
			ID: testutil.TaskID("unsafe-plan-thread"), Slug: "safe-thread", Status: domain.ThreadStatusUnstarted,
			Description: "Reject unsafe plan", Goal: "Keep creation in the guarded root", Created: "2026-08-30",
			Tasks: []string{task.ID}, Body: "body",
		},
	}
	snapshot := applySnapshot("planning", task)
	unsafe := plan
	unsafe.Thread.Slug = "../../outside"
	if _, err := PrepareThreadApply(snapshot, unsafe); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "canonical filename slug") {
		t.Fatalf("unsafe slug error = %v", err)
	}
	backdated := plan
	backdated.Thread.Created = "2026-08-29"
	if _, err := PrepareThreadApply(snapshot, backdated); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "composed_at") {
		t.Fatalf("created/composed mismatch = %v", err)
	}
	collision := plan
	collision.Thread.ID = task.ID
	if _, err := PrepareThreadApply(snapshot, collision); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("cross-kind collision = %v", err)
	}
	memberless := plan
	memberless.Thread.Tasks = nil
	if _, err := PrepareThreadApply(snapshot, memberless); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "at least one member") {
		t.Fatalf("memberless plan error = %v", err)
	}
}

func TestPrepareThreadApplyExplainsExistingThreadDifference(t *testing.T) {
	task := graphRecord("advanced-plan-member", domain.StatusCompleted)
	plan := ThreadApplyPlan{
		Schema: ThreadApplyPlanSchema, PlanningRepoID: "planning", ComposedAt: "2026-08-30",
		Thread: ThreadApplyThread{
			ID: testutil.TaskID("advanced-plan-thread"), Slug: "advanced-plan", Status: domain.ThreadStatusUnstarted,
			Description: "Explain a collision", Goal: "Keep retry diagnosis actionable", Created: "2026-08-30",
			Tasks: []string{task.ID}, Body: "body",
		},
	}
	existing := plan.Thread.domainThread()
	existing.FilenameID, existing.Path = existing.ID, "/threads/"+existing.ID+".md"
	existing.Status = domain.ThreadStatusInProgress
	existing.Updated = "2026-08-31"
	existing.StartedAt = "2026-08-31"
	snapshot := applySnapshot("planning", task)
	snapshot.Threads = []domain.Thread{existing}
	snapshot.ThreadBodies = map[string]string{existing.ID: plan.Thread.Body}
	if _, err := PrepareThreadApply(snapshot, plan); !errors.Is(err, domain.ErrConflict) ||
		!strings.Contains(err.Error(), "has advanced since this plan was applied") || !strings.Contains(err.Error(), "status") {
		t.Fatalf("advanced Thread error = %v", err)
	}

	existing = plan.Thread.domainThread()
	existing.FilenameID, existing.Path = existing.ID, "/threads/"+existing.ID+".md"
	existing.Description = "Different definition"
	snapshot.Threads = []domain.Thread{existing}
	if _, err := PrepareThreadApply(snapshot, plan); !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "different description") {
		t.Fatalf("different Thread error = %v", err)
	}
}
