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
		{name: "missing title", repoID: "planning", manifest: func() ThreadComposeManifest {
			value := base
			value.Thread.Title = " "
			return value
		}(), want: "Thread title is required"},
		{name: "empty slug", repoID: "planning", manifest: func() ThreadComposeManifest {
			value := base
			value.Thread.Title = "!!!"
			return value
		}(), want: "title produced an empty slug"},
		{name: "no nodes", repoID: "planning", manifest: ThreadComposeManifest{Thread: base.Thread}, want: "at least one node"},
		{name: "missing local key", repoID: "planning", manifest: ThreadComposeManifest{
			Thread: base.Thread, Nodes: []ThreadComposeNode{{Key: " ", TaskID: member.ID}},
		}, want: "requires a local key"},
		{name: "duplicate local key", repoID: "planning", manifest: ThreadComposeManifest{
			Thread: base.Thread, Nodes: []ThreadComposeNode{{Key: "member", TaskID: member.ID}, {Key: " member ", TaskID: nonmemberTask.ID}},
		}, want: `duplicate manifest node key "member"`},
		{name: "slug instead of stable id", repoID: "planning", manifest: ThreadComposeManifest{
			Thread: base.Thread, Nodes: []ThreadComposeNode{{Key: "member", TaskID: member.Slug}},
		}, want: "not an exact stable task id"},
		{name: "missing task", repoID: "planning", manifest: ThreadComposeManifest{
			Thread: base.Thread, Nodes: []ThreadComposeNode{{Key: "member", TaskID: testutil.TaskID("absent")}},
		}, want: "references missing task " + testutil.TaskID("absent")},
		{name: "no members", repoID: "planning", manifest: ThreadComposeManifest{
			Thread: base.Thread, Nodes: []ThreadComposeNode{{Key: "member", TaskID: member.ID, Member: &nonmember}},
		}, want: "at least one member node"},
		{name: "unknown prerequisite key", repoID: "planning", manifest: ThreadComposeManifest{
			Thread: base.Thread, Nodes: base.Nodes, Dependencies: []ThreadComposeDependency{{From: "missing", To: "member"}},
		}, want: `unknown local key "missing"`},
		{name: "unknown dependent key", repoID: "planning", manifest: ThreadComposeManifest{
			Thread: base.Thread, Nodes: base.Nodes, Dependencies: []ThreadComposeDependency{{From: "member", To: "missing"}},
		}, want: `unknown local key "missing"`},
		{name: "self dependency", repoID: "planning", manifest: ThreadComposeManifest{
			Thread: base.Thread, Nodes: base.Nodes, Dependencies: []ThreadComposeDependency{{From: "member", To: "member"}},
		}, want: "task " + member.ID + " cannot depend on itself"},
		{name: "duplicate dependency", repoID: "planning", manifest: ThreadComposeManifest{
			Thread: base.Thread, Nodes: []ThreadComposeNode{{Key: "member", TaskID: member.ID}, {Key: "gate", TaskID: nonmemberTask.ID}},
			Dependencies: []ThreadComposeDependency{{From: "gate", To: "member"}, {From: "gate", To: "member"}},
		}, want: "duplicate manifest dependency gate -> member"},
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
			if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestComposeThreadApplyPlanRequiresValidClockAndIDGenerator(t *testing.T) {
	member := graphRecord("composition-input-member", domain.StatusNextUp)
	snapshot := applySnapshot("planning", member)
	manifest := ThreadComposeManifest{
		Thread: ThreadComposeInput{Title: "Inputs", Description: "Validate composition inputs", Goal: "Fail before publication"},
		Nodes:  []ThreadComposeNode{{Key: "member", TaskID: member.ID}},
	}
	for _, tc := range []struct {
		name, candidate, diagnostic string
		now                         time.Time
		nilGenerator                bool
		want                        error
		calls                       int
	}{
		{name: "missing clock", candidate: testutil.TaskID("new-thread"), want: domain.ErrValidation, diagnostic: "compose time is required"},
		{name: "missing generator", now: threadApplyNow, nilGenerator: true, want: domain.ErrValidation, diagnostic: "id generator is required"},
		{name: "invalid generated id", now: threadApplyNow, candidate: "invalid", want: domain.ErrValidation, diagnostic: "is invalid", calls: 1},
		{name: "exhausted collisions", now: threadApplyNow, candidate: member.ID, want: domain.ErrConflict, diagnostic: "could not mint a unique Thread id", calls: 16},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			generator := func() string { calls++; return tc.candidate }
			if tc.nilGenerator {
				generator = nil
			}
			plan, err := ComposeThreadApplyPlan(snapshot, manifest, "body", generator, tc.now)
			if !errors.Is(err, tc.want) || !strings.Contains(err.Error(), tc.diagnostic) || calls != tc.calls || !reflect.DeepEqual(plan, ThreadApplyPlan{}) {
				t.Fatalf("plan=%+v err=%v calls=%d; want %v containing %q, calls=%d", plan, err, calls, tc.want, tc.diagnostic, tc.calls)
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
	gate := graphRecord("unsafe-plan-prerequisite", domain.StatusCompleted)
	plan := ThreadApplyPlan{
		Schema: ThreadApplyPlanSchema, PlanningRepoID: "planning", ComposedAt: "2026-08-30",
		Thread: ThreadApplyThread{
			ID: testutil.TaskID("unsafe-plan-thread"), Slug: "safe-thread", Status: domain.ThreadStatusUnstarted,
			Description: "Reject unsafe plan", Goal: "Keep creation in the guarded root", Created: "2026-08-30",
			Tasks: []string{task.ID}, Body: "body",
		},
		Dependencies: []ThreadApplyDependency{{From: gate.ID, To: task.ID}},
	}
	for _, tc := range []struct {
		name, diagnostic string
		edit             func(*ThreadApplySnapshot, *ThreadApplyPlan)
		conflict         bool
	}{
		{"unsafe slug", "canonical filename slug", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) { p.Thread.Slug = "../../outside" }, false},
		{"empty slug", "canonical filename slug", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) { p.Thread.Slug = "" }, false},
		{"created mismatch", "must equal composed_at", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) { p.Thread.Created = "2026-08-29" }, false},
		{"cross-kind collision", "already used by a task", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) { p.Thread.ID = task.ID }, true},
		{"memberless", "at least one member task", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) { p.Thread.Tasks = nil }, false},
		{"unsupported schema", "unsupported Thread apply-plan schema", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) { p.Schema++ }, false},
		{"missing planned identity", "apply plan has no planning_repo_id", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) { p.PlanningRepoID = "" }, false},
		{"missing current identity", "current planning repository has no durable id", func(s *ThreadApplySnapshot, _ *ThreadApplyPlan) { s.PlanningRepoID = "" }, false},
		{"invalid composition date", "apply plan composed_at", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) { p.ComposedAt = "invalid" }, false},
		{"advanced planned status", "planned Thread must be unstarted", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) { p.Thread.Status = domain.ThreadStatusInProgress }, false},
		{"prerequisite is a slug", "requires exact stable task ids", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) { p.Dependencies[0].From = gate.Slug }, false},
		{"dependent is a slug", "requires exact stable task ids", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) { p.Dependencies[0].To = task.Slug }, false},
		{"self dependency", "cannot depend on itself", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) { p.Dependencies[0].From = task.ID }, false},
		{"duplicate dependency", "duplicate planned dependency", func(_ *ThreadApplySnapshot, p *ThreadApplyPlan) {
			p.Dependencies = append(p.Dependencies, p.Dependencies[0])
		}, false},
		{"missing prerequisite", "planned prerequisite " + gate.ID + " does not exist", func(s *ThreadApplySnapshot, _ *ThreadApplyPlan) { s.Graph = NewTaskGraph([]domain.Task{task}, nil) }, false},
		{"missing dependent", "planned dependent " + task.ID + " does not exist", func(s *ThreadApplySnapshot, _ *ThreadApplyPlan) { s.Graph = NewTaskGraph([]domain.Task{gate}, nil) }, false},
		{"unavailable authoritative body", "authoritative body for existing planned Thread " + plan.Thread.ID + " is unavailable", func(s *ThreadApplySnapshot, p *ThreadApplyPlan) { s.Threads = []domain.Thread{p.Thread.domainThread()} }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := applySnapshot("planning", gate, task)
			edited := cloneThreadApplyPlan(plan)
			// Establish a valid baseline so unrelated fixture failures cannot pin a guard.
			if _, err := PrepareThreadApply(snapshot, edited); err != nil {
				t.Fatal(err)
			}
			tc.edit(&snapshot, &edited)
			before := cloneThreadApplyPlan(edited)
			decision, err := PrepareThreadApply(snapshot, edited)
			want := domain.ErrValidation
			if tc.conflict {
				want = domain.ErrConflict
			}
			if !errors.Is(err, want) || !strings.Contains(err.Error(), tc.diagnostic) || !reflect.DeepEqual(decision, ThreadApplyDecision{}) {
				t.Fatalf("decision=%+v err=%v; want %v containing %q", decision, err, want, tc.diagnostic)
			}
			if !reflect.DeepEqual(edited, before) {
				t.Fatal("refused preparation mutated caller's durable plan")
			}
		})
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
	existing.Description = "Different definition"
	snapshot.Threads = []domain.Thread{existing}
	if _, err := PrepareThreadApply(snapshot, plan); !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "different description") {
		t.Fatalf("different Thread error = %v", err)
	}
}
