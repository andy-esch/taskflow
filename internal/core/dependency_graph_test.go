package core

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func graphRecord(seed string, status domain.Status, dependencies ...string) domain.Task {
	taskID := testutil.TaskID(seed)
	return domain.Task{
		ID: taskID, FilenameID: taskID, Slug: seed, Path: "tasks/" + taskID + "-" + seed + ".md",
		Status: status, Description: seed, Tags: []string{"graph"},
		DependsOn: append([]string(nil), dependencies...),
	}
}

func TestTaskGraphHealthAndDeterministicStructuralProblems(t *testing.T) {
	a := graphRecord("a", domain.StatusCompleted)
	b := graphRecord("b", domain.StatusReadyToStart, a.ID, a.ID)
	c := graphRecord("c", domain.StatusReadyToStart)
	c.DependsOn = []string{c.ID}
	d := graphRecord("d", domain.StatusReadyToStart, testutil.TaskID("missing"))
	e := graphRecord("e", domain.Status("invented"))
	f := graphRecord("f", domain.StatusCompleted)
	f.ID = testutil.TaskID("drifted-frontmatter")

	tasks := []domain.Task{f, e, d, c, b, a}
	graph := NewTaskGraph(tasks, []domain.FileProblem{{Path: "tasks/broken.md", Message: "malformed frontmatter"}})
	if graph.Health() != GraphBroken || graph.MutationReady() {
		t.Fatalf("health = %s mutationReady=%v", graph.Health(), graph.MutationReady())
	}
	wantCodes := []GraphProblemCode{
		ProblemUnreadable, ProblemDuplicateDependency, ProblemSelfDependency,
		ProblemMissingDependency, ProblemInvalidStatus, ProblemTaskIDDrift,
	}
	gotCodes := make([]GraphProblemCode, 0)
	for _, problem := range graph.Problems() {
		gotCodes = append(gotCodes, problem.Code)
	}
	for _, code := range wantCodes {
		if !slices.Contains(gotCodes, code) {
			t.Errorf("problems %v do not contain %s", gotCodes, code)
		}
	}

	baseline := fmt.Sprintf("%+v", graph.Problems())
	for seed := int64(0); seed < 20; seed++ {
		shuffled := append([]domain.Task(nil), tasks...)
		rand.New(rand.NewSource(seed)).Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		for i := range shuffled {
			rand.New(rand.NewSource(seed+int64(i)+100)).Shuffle(len(shuffled[i].DependsOn), func(x, y int) {
				shuffled[i].DependsOn[x], shuffled[i].DependsOn[y] = shuffled[i].DependsOn[y], shuffled[i].DependsOn[x]
			})
		}
		if got := fmt.Sprintf("%+v", NewTaskGraph(shuffled, []domain.FileProblem{{Path: "tasks/broken.md", Message: "malformed frontmatter"}}).Problems()); got != baseline {
			t.Fatalf("seed %d changed diagnostics\nwant %s\n got %s", seed, baseline, got)
		}
	}
}

func TestTaskGraphDiagnosesEveryIdentityAndEdgeShape(t *testing.T) {
	missingFrontmatterID := graphRecord("missing-frontmatter-id", domain.StatusReadyToStart)
	missingFrontmatterID.ID = ""
	duplicateA := graphRecord("duplicate-a", domain.StatusReadyToStart)
	duplicateB := graphRecord("duplicate-b", domain.StatusReadyToStart)
	duplicateB.ID = duplicateA.ID
	duplicateB.FilenameID = duplicateA.ID
	invalidEdge := graphRecord("invalid-edge", domain.StatusReadyToStart, "not-a-stable-id")

	graph := NewTaskGraph([]domain.Task{invalidEdge, duplicateB, missingFrontmatterID, duplicateA}, nil)
	want := []GraphProblemCode{ProblemMissingTaskID, ProblemDuplicateTaskID, ProblemInvalidDependencyID}
	got := make([]GraphProblemCode, 0)
	for _, problem := range graph.Problems() {
		got = append(got, problem.Code)
	}
	for _, code := range want {
		if !slices.Contains(got, code) {
			t.Errorf("problems %v do not contain %s", got, code)
		}
	}
}

func TestTaskGraphGatePrecedenceReasonsAndShortestPaths(t *testing.T) {
	root := graphRecord("root", domain.StatusReadyToStart)
	inFlight := graphRecord("in-flight", domain.StatusInProgress)
	parked := graphRecord("parked", domain.StatusDeferred)
	withdrawn := graphRecord("withdrawn", domain.StatusDeprecated)
	invalid := graphRecord("invalid", domain.Status("bogus"))
	unsoundDone := graphRecord("unsound-done", domain.StatusCompleted, root.ID)
	missingID := testutil.TaskID("missing")
	target := graphRecord("target", domain.StatusReadyToStart,
		root.ID, inFlight.ID, parked.ID, withdrawn.ID, invalid.ID, unsoundDone.ID, missingID)

	graph := NewTaskGraph([]domain.Task{target, unsoundDone, invalid, withdrawn, parked, inFlight, root}, nil)
	if state := graph.State(target.ID); state.Gate != GateBroken || state.Eligible {
		t.Fatalf("target state = %+v; broken must outrank blocked", state)
	}
	reasons := map[string]BlockerReason{}
	for _, blocker := range graph.CausalBlockers(target.ID) {
		reasons[blocker.TaskID] = blocker.Reason
		if blocker.TaskID == root.ID && (!blocker.Direct || !reflect.DeepEqual(blocker.Path, []string{target.ID, root.ID})) {
			t.Errorf("direct root path = %+v", blocker)
		}
	}
	want := map[string]BlockerReason{
		root.ID:        BlockerNotStarted,
		inFlight.ID:    BlockerInFlight,
		parked.ID:      BlockerParked,
		withdrawn.ID:   BlockerWithdrawn,
		invalid.ID:     BlockerInvalidStatus,
		unsoundDone.ID: BlockerUnsoundCompleted,
		missingID:      BlockerMissing,
	}
	if !reflect.DeepEqual(reasons, want) {
		t.Fatalf("reasons = %v, want %v", reasons, want)
	}

	// Two equal-length paths to root choose the lexicographically first immediate
	// prerequisite, independent of insertion order.
	left := graphRecord("left", domain.StatusCompleted, root.ID)
	right := graphRecord("right", domain.StatusCompleted, root.ID)
	join := graphRecord("join", domain.StatusReadyToStart, right.ID, left.ID)
	shortest := NewTaskGraph([]domain.Task{join, right, left, root}, nil).CausalBlockers(join.ID)
	var rootPath []string
	for _, blocker := range shortest {
		if blocker.TaskID == root.ID {
			rootPath = blocker.Path
		}
	}
	first := left.ID
	if right.ID < left.ID {
		first = right.ID
	}
	if !reflect.DeepEqual(rootPath, []string{join.ID, first, root.ID}) {
		t.Fatalf("root shortest path = %v, want lexicographic tie via %s", rootPath, first)
	}
}

func TestTaskGraphPendingRolesShareGraphDrivenEligibility(t *testing.T) {
	completed := graphRecord("eligibility-completed", domain.StatusCompleted)
	unfinished := graphRecord("eligibility-unfinished", domain.StatusNextUp)
	for _, tc := range []struct {
		status domain.Status
		role   LifecycleRole
	}{
		{status: domain.StatusNextUp, role: RoleQueued},
		{status: domain.StatusReadyToStart, role: RoleCandidate},
	} {
		t.Run(string(tc.status), func(t *testing.T) {
			clear := graphRecord("clear-"+string(tc.status), tc.status, completed.ID)
			blocked := graphRecord("blocked-"+string(tc.status), tc.status, unfinished.ID)
			graph := NewTaskGraph([]domain.Task{clear, blocked, completed, unfinished}, nil)
			if state := graph.State(clear.ID); state.Role != tc.role || state.Gate != GateClear || !state.Eligible {
				t.Fatalf("clear state = %+v", state)
			}
			if state := graph.State(blocked.ID); state.Role != tc.role || state.Gate != GateBlocked || state.Eligible {
				t.Fatalf("blocked state = %+v", state)
			}
		})
	}
	for _, status := range []domain.Status{
		domain.StatusInProgress, domain.StatusDeferred, domain.StatusCompleted, domain.StatusDeprecated,
	} {
		task := graphRecord("not-pending-"+string(status), status)
		if state := NewTaskGraph([]domain.Task{task}, nil).State(task.ID); state.Eligible {
			t.Errorf("status %s unexpectedly eligible: %+v", status, state)
		}
	}
}

func TestTaskGraphLifecycleRoleCoversEveryPersistedStatus(t *testing.T) {
	for _, status := range domain.AllStatuses() {
		if role := roleForStatus(status); role == RoleUnknown {
			t.Errorf("persisted status %q maps to unknown lifecycle role", status)
		}
	}
}

func TestTaskGraphLegacyResolutionHealthAndDirection(t *testing.T) {
	prerequisite := graphRecord("prerequisite", domain.StatusCompleted)
	dependent := graphRecord("dependent", domain.StatusReadyToStart)
	dependent.LegacyBlockedBy = []string{prerequisite.Slug}
	dependent.LegacyDependencies = []string{prerequisite.ID}
	prerequisite.LegacyBlocks = []string{dependent.Slug}

	graph := NewTaskGraph([]domain.Task{dependent, prerequisite}, nil)
	if graph.Health() != GraphDegraded || graph.MutationReady() {
		t.Fatalf("resolved legacy health = %s mutationReady=%v", graph.Health(), graph.MutationReady())
	}
	if len(graph.Problems()) != 0 || len(graph.LegacyDiagnostics()) != 3 {
		t.Fatalf("problems=%+v legacy=%+v", graph.Problems(), graph.LegacyDiagnostics())
	}
	for _, diagnostic := range graph.LegacyDiagnostics() {
		for _, ref := range diagnostic.References {
			if ref.Resolution != LegacyResolved || ref.Edge != (DependencyEdge{From: prerequisite.ID, To: dependent.ID}) {
				t.Errorf("legacy %s resolution = %+v", diagnostic.Field, ref)
			}
		}
	}
	// Resolved legacy constraints are projected into explanatory reads and derived
	// gates even though dispatch still fails closed until migration makes them
	// canonical.
	if state := graph.State(dependent.ID); state.Gate != GateClear || state.Eligible {
		t.Fatalf("degraded candidate state = %+v", state)
	}
	prerequisite.Status = domain.StatusReadyToStart
	blocked := NewTaskGraph([]domain.Task{dependent, prerequisite}, nil)
	if state := blocked.State(dependent.ID); state.Gate != GateBlocked || state.Eligible {
		t.Fatalf("degraded blocked state = %+v", state)
	}
	if blockers := blocked.CausalBlockers(dependent.ID); len(blockers) != 1 || blockers[0].TaskID != prerequisite.ID {
		t.Fatalf("legacy blocker projection = %+v", blockers)
	}
	if impacts := blocked.DownstreamImpact(prerequisite.ID); len(impacts) != 1 || impacts[0].TaskID != dependent.ID {
		t.Fatalf("legacy downstream projection = %+v", impacts)
	}
}

func TestTaskGraphDiagnosesPresentEmptyLegacyFields(t *testing.T) {
	task := graphRecord("empty-legacy", domain.StatusReadyToStart)
	task.LegacyDependencyFields = []string{"blocked_by", "dependencies", "blocks"}
	graph := NewTaskGraph([]domain.Task{task}, nil)
	if graph.Health() != GraphDegraded || len(graph.LegacyDiagnostics()) != 3 {
		t.Fatalf("empty legacy health=%s diagnostics=%+v", graph.Health(), graph.LegacyDiagnostics())
	}
	for _, diagnostic := range graph.LegacyDiagnostics() {
		if len(diagnostic.References) != 0 {
			t.Fatalf("empty %s references = %+v", diagnostic.Field, diagnostic.References)
		}
	}
}

func TestTaskGraphLegacyMissingAndAmbiguousAreBroken(t *testing.T) {
	first := graphRecord("same-one", domain.StatusCompleted)
	second := graphRecord("same-two", domain.StatusCompleted)
	first.Slug, second.Slug = "same", "same"
	dependent := graphRecord("dependent", domain.StatusReadyToStart)
	dependent.LegacyBlockedBy = []string{"same", "gone"}
	graph := NewTaskGraph([]domain.Task{dependent, first, second}, nil)
	if graph.Health() != GraphBroken {
		t.Fatalf("health = %s", graph.Health())
	}
	codes := make([]GraphProblemCode, 0)
	for _, problem := range graph.Problems() {
		codes = append(codes, problem.Code)
	}
	if !slices.Contains(codes, ProblemLegacyAmbiguous) || !slices.Contains(codes, ProblemLegacyMissing) {
		t.Fatalf("legacy problem codes = %v", codes)
	}
	diagnostic := graph.LegacyDiagnostics()[0]
	if len(diagnostic.References) != 2 || diagnostic.References[0].Value != "gone" || diagnostic.References[1].Value != "same" {
		t.Fatalf("legacy references not stable: %+v", diagnostic.References)
	}
}

func TestTaskGraphTopologicalWavesAndDownstream(t *testing.T) {
	a := graphRecord("a", domain.StatusCompleted)
	b := graphRecord("b", domain.StatusCompleted, a.ID)
	c := graphRecord("c", domain.StatusCompleted, a.ID)
	d := graphRecord("d", domain.StatusReadyToStart, b.ID, c.ID)
	e := graphRecord("disconnected", domain.StatusReadyToStart)
	graph := NewTaskGraph([]domain.Task{d, b, e, c, a}, nil)
	waves, complete := graph.TopologicalWaves()
	first := []string{a.ID, e.ID}
	sortStrings(first)
	second := []string{b.ID, c.ID}
	sortStrings(second)
	if !complete || !reflect.DeepEqual(waves, [][]string{first, second, {d.ID}}) {
		t.Fatalf("waves = %v complete=%v", waves, complete)
	}
	downstream := []string{b.ID, c.ID, d.ID}
	sortStrings(downstream)
	if got := graph.Downstream(a.ID); !reflect.DeepEqual(got, downstream) {
		t.Fatalf("downstream = %v, want %v", got, downstream)
	}
	// Returned slices are copies; callers cannot mutate snapshot caches.
	got := graph.Downstream(a.ID)
	got[0] = "corrupt"
	if reflect.DeepEqual(got, graph.Downstream(a.ID)) {
		t.Fatal("downstream cache leaked a mutable slice")
	}
}

func TestTaskGraphResolveTaskIDMatchesRepositoryReferenceTiers(t *testing.T) {
	polish := graphRecord("polish", domain.StatusReadyToStart)
	batch := graphRecord("polish-batch", domain.StatusReadyToStart)
	backoff := graphRecord("add-retry-backoff", domain.StatusReadyToStart)
	jitter := graphRecord("add-retry-jitter", domain.StatusReadyToStart)
	unreadableID := testutil.TaskID("unreadable")
	graph := NewTaskGraph(
		[]domain.Task{jitter, batch, backoff, polish},
		[]domain.FileProblem{{Path: "tasks/" + unreadableID + "-unreadable.md", Message: "bad YAML"}},
	)

	tests := []struct {
		query string
		want  string
	}{
		{query: "polish", want: polish.ID},        // exact beats a longer prefix
		{query: "POLISH-B", want: batch.ID},       // unique prefix, case-insensitive
		{query: "JITTER", want: jitter.ID},        // unique substring, case-insensitive
		{query: backoff.ID[:7], want: backoff.ID}, // stable-ID prefix
		{query: unreadableID, want: unreadableID}, // exact diagnostic addressability
		{query: "UNREAD", want: unreadableID},     // filename slug parity despite unreadable YAML
	}
	for _, tc := range tests {
		got, err := graph.ResolveTaskID(tc.query)
		if err != nil || got != tc.want {
			t.Errorf("ResolveTaskID(%q) = %q, %v; want %q", tc.query, got, err, tc.want)
		}
	}

	if _, err := graph.ResolveTaskID("add-retry"); !errors.Is(err, domain.ErrAmbiguous) ||
		!strings.Contains(err.Error(), backoff.ID) || !strings.Contains(err.Error(), jitter.ID) {
		t.Fatalf("ambiguous resolution should be classified and list canonical IDs: %v", err)
	}
	if _, err := graph.ResolveTaskID("does-not-exist"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing resolution = %v, want ErrNotFound", err)
	}
	for _, query := range []string{"", "../escape", `a\b`, "a/b", ".."} {
		if _, err := graph.ResolveTaskID(query); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("unsafe query %q = %v, want ErrValidation", query, err)
		}
	}

	duplicateA := graphRecord("duplicate-one", domain.StatusReadyToStart)
	duplicateB := graphRecord("duplicate-two", domain.StatusReadyToStart)
	duplicateB.ID, duplicateB.FilenameID = duplicateA.ID, duplicateA.ID
	duplicateGraph := NewTaskGraph([]domain.Task{duplicateA, duplicateB}, nil)
	if _, err := duplicateGraph.ResolveTaskID(duplicateA.ID); !errors.Is(err, domain.ErrAmbiguous) ||
		!strings.Contains(err.Error(), duplicateA.Slug) || !strings.Contains(err.Error(), duplicateB.Slug) {
		t.Fatalf("duplicate-id resolution = %v, want both source candidates and ErrAmbiguous", err)
	}
}

func TestTaskGraphDownstreamImpactUsesDeterministicShortestPaths(t *testing.T) {
	root := graphRecord("impact-root", domain.StatusCompleted)
	left := graphRecord("impact-left", domain.StatusCompleted, root.ID)
	right := graphRecord("impact-right", domain.StatusCompleted, root.ID)
	join := graphRecord("impact-join", domain.StatusReadyToStart, right.ID, left.ID)
	graph := NewTaskGraph([]domain.Task{join, right, left, root}, nil)

	got := graph.DownstreamImpact(root.ID)
	byID := make(map[string]DependentImpact, len(got))
	for _, impact := range got {
		byID[impact.TaskID] = impact
	}
	if !byID[left.ID].Direct || !byID[right.ID].Direct {
		t.Fatalf("immediate downstream tasks were not marked direct: %+v", got)
	}
	first := left.ID
	if right.ID < left.ID {
		first = right.ID
	}
	wantPath := []string{root.ID, first, join.ID}
	if impact := byID[join.ID]; impact.Direct || !reflect.DeepEqual(impact.Path, wantPath) {
		t.Fatalf("join impact = %+v, want shortest path %v", impact, wantPath)
	}

	// Cached paths remain immutable to callers.
	byID[join.ID].Path[0] = "corrupt"
	for _, impact := range graph.DownstreamImpact(root.ID) {
		if impact.TaskID == join.ID && impact.Path[0] != root.ID {
			t.Fatalf("downstream impact cache leaked mutable path: %+v", impact)
		}
	}
}

func TestAnalyzeDAGDeepWideAndDisconnected(t *testing.T) {
	const depth = 2048
	nodes := make([]string, 0, depth+129)
	edges := make([]DependencyEdge, 0, depth-1)
	for i := 0; i < depth; i++ {
		node := fmt.Sprintf("%012d", i)
		nodes = append(nodes, node)
		if i > 0 {
			edges = append(edges, DependencyEdge{From: fmt.Sprintf("%012d", i-1), To: node})
		}
	}
	for i := 0; i < 128; i++ {
		nodes = append(nodes, fmt.Sprintf("w%011d", i))
	}
	nodes = append(nodes, "z-disconnected")

	analysis := analyzeDAG(dagInput{Nodes: nodes, Edges: edges})
	if !analysis.TopologicalComplete || len(analysis.CyclicComponents) != 0 || len(analysis.TopologicalWaves) != depth {
		t.Fatalf("analysis complete=%v cycles=%d waves=%d, want true/0/%d",
			analysis.TopologicalComplete, len(analysis.CyclicComponents), len(analysis.TopologicalWaves), depth)
	}
	if got := len(analysis.TopologicalWaves[0]); got != 130 {
		t.Fatalf("first wide/disconnected frontier has %d nodes, want 130", got)
	}
}

func TestTaskGraphSupportedDeepChainEnvelope(t *testing.T) {
	// 4,096 edges is deliberately far beyond a plausible markdown planning repo
	// while still cheap enough for every CI run. Unlike TestAnalyzeDAGDeep..., this
	// exercises the complete snapshot, recursive sound derivation, and explanatory
	// path materialization rather than only the structural analyzer.
	const edges = 4096
	tasks := make([]domain.Task, 0, edges+1)
	for i := 0; i <= edges; i++ {
		taskID := fmt.Sprintf("%012d", i)
		status := domain.StatusCompleted
		if i == 0 || i == edges {
			status = domain.StatusReadyToStart
		}
		task := domain.Task{
			ID: taskID, FilenameID: taskID, Slug: fmt.Sprintf("deep-%04d", i),
			Path: "tasks/" + taskID + "-deep.md", Status: status,
		}
		if i < edges {
			task.DependsOn = []string{fmt.Sprintf("%012d", i+1)}
		}
		tasks = append(tasks, task)
	}
	graph := NewTaskGraph(tasks, nil)
	if graph.Health() != GraphHealthy {
		t.Fatalf("deep graph health = %s", graph.Health())
	}
	state := graph.State("000000000000")
	if state.Gate != GateBlocked || state.Eligible {
		t.Fatalf("deep root state = %+v", state)
	}
	frontier := graph.BlockingFrontier("000000000000")
	if len(frontier) != 1 || frontier[0].TaskID != fmt.Sprintf("%012d", edges) || len(frontier[0].Path) != edges+1 {
		t.Fatalf("deep frontier count=%d blocker=%+v", len(frontier), frontier)
	}
}

func TestTaskGraphPathProjectionOutputEnvelope(t *testing.T) {
	// Full explanatory paths amplify a linear chain quadratically. Keep that cost
	// explicit at a large-but-CI-safe depth; the supported 4,096-edge structural and
	// frontier envelope remains covered separately above.
	const edges = 512
	tasks := make([]domain.Task, 0, edges+1)
	for i := 0; i <= edges; i++ {
		taskID := fmt.Sprintf("%012d", i)
		status := domain.StatusCompleted
		if i == edges {
			status = domain.StatusReadyToStart
		}
		task := domain.Task{ID: taskID, FilenameID: taskID, Slug: fmt.Sprintf("path-%04d", i), Status: status}
		if i < edges {
			task.DependsOn = []string{fmt.Sprintf("%012d", i+1)}
		}
		tasks = append(tasks, task)
	}
	graph := NewTaskGraph(tasks, nil)
	wantPathElements := edges * (edges + 3) / 2 // lengths 2..edges+1
	causal := graph.CausalBlockers("000000000000")
	if len(causal) != edges || totalBlockerPathElements(causal) != wantPathElements {
		t.Fatalf("causal count=%d path-elements=%d, want %d/%d", len(causal), totalBlockerPathElements(causal), edges, wantPathElements)
	}
	impacts := graph.DownstreamImpact(fmt.Sprintf("%012d", edges))
	if len(impacts) != edges || totalImpactPathElements(impacts) != wantPathElements {
		t.Fatalf("impact count=%d path-elements=%d, want %d/%d", len(impacts), totalImpactPathElements(impacts), edges, wantPathElements)
	}
}

func totalBlockerPathElements(blockers []Blocker) int {
	total := 0
	for _, blocker := range blockers {
		total += len(blocker.Path)
	}
	return total
}

func totalImpactPathElements(impacts []DependentImpact) int {
	total := 0
	for _, impact := range impacts {
		total += len(impact.Path)
	}
	return total
}

func TestTaskGraphCycleBlockerReason(t *testing.T) {
	a := graphRecord("cycle-a", domain.StatusCompleted)
	b := graphRecord("cycle-b", domain.StatusCompleted, a.ID)
	a.DependsOn = []string{b.ID}
	target := graphRecord("cycle-target", domain.StatusReadyToStart, a.ID)
	graph := NewTaskGraph([]domain.Task{target, b, a}, nil)

	if state := graph.State(target.ID); state.Gate != GateBroken || state.Eligible {
		t.Fatalf("target state = %+v", state)
	}
	blockers := graph.CausalBlockers(target.ID)
	if len(blockers) != 2 {
		t.Fatalf("cycle blockers = %+v, want both cycle members", blockers)
	}
	for _, blocker := range blockers {
		if blocker.Reason != BlockerCycle {
			t.Fatalf("cycle blocker = %+v", blocker)
		}
	}
	for _, impact := range graph.DownstreamImpact(a.ID) {
		if impact.TaskID == a.ID {
			t.Fatalf("cyclic downstream query reported its source as its own impact: %+v", impact)
		}
	}
}

func TestTaskGraphSCCMarksEveryMemberAndEmitsRepresentativePath(t *testing.T) {
	a := graphRecord("scc-a", domain.StatusCompleted)
	b := graphRecord("scc-b", domain.StatusCompleted)
	c := graphRecord("scc-c", domain.StatusCompleted)
	a.DependsOn = []string{b.ID}
	b.DependsOn = []string{a.ID, c.ID}
	c.DependsOn = []string{a.ID}
	graph := NewTaskGraph([]domain.Task{c, a, b}, nil)

	cycleProblems := make(map[string]GraphProblem)
	for _, problem := range graph.Problems() {
		if problem.Code == ProblemCycle {
			cycleProblems[problem.TaskID] = problem
		}
	}
	if len(cycleProblems) != 3 {
		t.Fatalf("cycle attribution = %+v, want one problem for every SCC member", cycleProblems)
	}
	for _, task := range []domain.Task{a, b, c} {
		problem := cycleProblems[task.ID]
		if len(problem.Cycle) < 3 || problem.Cycle[0] != problem.Cycle[len(problem.Cycle)-1] {
			t.Fatalf("representative cycle for %s = %v", task.ID, problem.Cycle)
		}
		if graph.blockerReason(task.ID) != BlockerCycle {
			t.Fatalf("member %s reason = %s", task.ID, graph.blockerReason(task.ID))
		}
		byID := map[string]domain.Task{a.ID: a, b.ID: b, c.ID: c}
		for i := 0; i < len(problem.Cycle)-1; i++ {
			prerequisite, dependent := problem.Cycle[i], problem.Cycle[i+1]
			if !slices.Contains(byID[dependent].DependsOn, prerequisite) {
				t.Fatalf("representative path contains non-edge %s -> %s: %v", prerequisite, dependent, problem.Cycle)
			}
		}
	}
}

func TestTaskGraphSelfDependencyDoesNotDuplicateCycleDiagnostic(t *testing.T) {
	task := graphRecord("self-only", domain.StatusReadyToStart)
	task.DependsOn = []string{task.ID}
	graph := NewTaskGraph([]domain.Task{task}, nil)
	var self, cycle int
	for _, problem := range graph.Problems() {
		switch problem.Code {
		case ProblemSelfDependency:
			self++
		case ProblemCycle:
			cycle++
		}
	}
	if self != 1 || cycle != 0 {
		t.Fatalf("self=%d cycle=%d problems=%+v", self, cycle, graph.Problems())
	}
}

func TestTaskGraphLegacyProjectedCycleIsUnsafeAndBroken(t *testing.T) {
	a := graphRecord("legacy-cycle-a", domain.StatusReadyToStart)
	b := graphRecord("legacy-cycle-b", domain.StatusReadyToStart)
	a.LegacyBlockedBy = []string{b.ID}
	b.LegacyBlockedBy = []string{a.ID}
	graph := NewTaskGraph([]domain.Task{b, a}, nil)
	if graph.Health() != GraphBroken || graph.MutationReady() {
		t.Fatalf("health=%s mutationReady=%v", graph.Health(), graph.MutationReady())
	}
	for _, diagnostic := range graph.LegacyDiagnostics() {
		if len(diagnostic.References) != 1 || diagnostic.References[0].Resolution != LegacyUnsafe {
			t.Fatalf("legacy diagnostic = %+v, want unsafe", diagnostic)
		}
	}
	if _, complete := graph.TopologicalWaves(); complete {
		t.Fatal("a legacy-projected cycle cannot yield a complete topology")
	}
}

func TestTaskGraphBrokenOrDegradedTopologyIsNeverComplete(t *testing.T) {
	missing := graphRecord("topology-missing", domain.StatusReadyToStart, testutil.TaskID("absent"))
	broken := NewTaskGraph([]domain.Task{missing}, nil)
	if waves, complete := broken.TopologicalWaves(); complete || len(waves) == 0 {
		t.Fatalf("broken waves=%v complete=%v; partial waves are diagnostic only", waves, complete)
	}

	prerequisite := graphRecord("topology-legacy-prereq", domain.StatusCompleted)
	dependent := graphRecord("topology-legacy-dependent", domain.StatusReadyToStart)
	dependent.LegacyBlockedBy = []string{prerequisite.ID}
	degraded := NewTaskGraph([]domain.Task{dependent, prerequisite}, nil)
	if waves, complete := degraded.TopologicalWaves(); complete || len(waves) != 2 {
		t.Fatalf("degraded projected waves=%v complete=%v", waves, complete)
	}
}

func TestTaskGraphSeparatesCausalBlockersFromActionFrontier(t *testing.T) {
	root := graphRecord("frontier-root", domain.StatusReadyToStart)
	middle := graphRecord("frontier-middle", domain.StatusCompleted, root.ID)
	target := graphRecord("frontier-target", domain.StatusReadyToStart, middle.ID)
	graph := NewTaskGraph([]domain.Task{target, middle, root}, nil)

	causal := graph.CausalBlockers(target.ID)
	if len(causal) != 2 {
		t.Fatalf("causal blockers = %+v, want middle and root", causal)
	}
	frontier := graph.BlockingFrontier(target.ID)
	if len(frontier) != 1 || frontier[0].TaskID != root.ID || frontier[0].Reason != BlockerNotStarted {
		t.Fatalf("frontier = %+v, want root only", frontier)
	}

	broken := graphRecord("locally-broken", domain.Status("invalid"))
	explanation := NewTaskGraph([]domain.Task{broken, root}, nil).ExplainGate(broken.ID)
	if explanation.State.Eligible || explanation.State.Gate != GateBroken || len(explanation.LocalProblems) == 0 || len(explanation.Frontier) != 0 {
		t.Fatalf("broken explanation = %+v", explanation)
	}
}

func TestTaskGraphFrontierStopsAtWithdrawnWhileCausalProjectionContinues(t *testing.T) {
	root := graphRecord("withdrawn-root", domain.StatusReadyToStart)
	withdrawn := graphRecord("withdrawn-middle", domain.StatusDeprecated, root.ID)
	target := graphRecord("withdrawn-target", domain.StatusReadyToStart, withdrawn.ID)
	graph := NewTaskGraph([]domain.Task{target, withdrawn, root}, nil)
	if causal := graph.CausalBlockers(target.ID); len(causal) != 2 {
		t.Fatalf("causal blockers = %+v, want withdrawn and its upstream", causal)
	}
	frontier := graph.BlockingFrontier(target.ID)
	if len(frontier) != 1 || frontier[0].TaskID != withdrawn.ID || frontier[0].Reason != BlockerWithdrawn {
		t.Fatalf("frontier = %+v, want terminal withdrawn task", frontier)
	}
}

func TestTaskGraphDistinguishesUnreadableAndInvalidReferences(t *testing.T) {
	unreadableID := testutil.TaskID("unreadable")
	target := graphRecord("unreadable-target", domain.StatusReadyToStart, unreadableID, "not-an-id")
	graph := NewTaskGraph([]domain.Task{target}, []domain.FileProblem{{
		Path: "tasks/" + unreadableID + "-broken.md", Message: "malformed frontmatter",
	}})
	reasons := make(map[string]BlockerReason)
	for _, blocker := range graph.CausalBlockers(target.ID) {
		reasons[blocker.TaskID] = blocker.Reason
	}
	if reasons[unreadableID] != BlockerUnreadable || reasons["not-an-id"] != BlockerInvalidReference {
		t.Fatalf("blocker reasons = %+v", reasons)
	}
}

func TestTaskGraphDuplicateIDsRetainPathFaithfulDiagnostics(t *testing.T) {
	first := graphRecord("duplicate-path-first", domain.StatusReadyToStart)
	second := graphRecord("duplicate-path-second", domain.StatusReadyToStart, "bad-reference")
	second.ID, second.FilenameID = first.ID, first.ID
	graph := NewTaskGraph([]domain.Task{second, first}, nil)
	duplicates := make(map[string]bool)
	invalidOnSecond := false
	for _, problem := range graph.Problems() {
		if problem.Code == ProblemDuplicateTaskID {
			duplicates[problem.Path] = true
		}
		if problem.Code == ProblemInvalidDependencyID && problem.Path == second.Path {
			invalidOnSecond = true
		}
	}
	if !duplicates[first.Path] || !duplicates[second.Path] || !invalidOnSecond {
		t.Fatalf("path-faithful problems = %+v", graph.Problems())
	}
	lintByPath := dependencyLintIssues(graph)
	if len(lintByPath[first.Path]) == 0 || len(lintByPath[second.Path]) < 2 {
		t.Fatalf("path-faithful lint issues = %+v", lintByPath)
	}
	for _, issue := range lintByPath[first.Path] {
		if strings.Contains(issue.Message, "bad-reference") {
			t.Fatalf("second record defect leaked onto first path: %+v", lintByPath)
		}
	}
}

func TestTaskGraphSoundCompletionMemoizesReconvergentDiamonds(t *testing.T) {
	tasks := []domain.Task{graphRecord("root", domain.StatusCompleted)}
	previous := tasks[0].ID
	for i := 0; i < 120; i++ {
		left := graphRecord(fmt.Sprintf("left-%03d", i), domain.StatusCompleted, previous)
		right := graphRecord(fmt.Sprintf("right-%03d", i), domain.StatusCompleted, previous)
		join := graphRecord(fmt.Sprintf("join-%03d", i), domain.StatusCompleted, left.ID, right.ID)
		tasks = append(tasks, left, right, join)
		previous = join.ID
	}
	graph := NewTaskGraph(tasks, nil)
	sound, broken := graph.SoundlyCompleted(previous)
	if !sound || broken {
		t.Fatalf("last join sound=%v broken=%v", sound, broken)
	}
	for taskID, visits := range graph.soundVisits {
		if visits != 1 {
			t.Fatalf("task %s computed %d times; want exactly once", taskID, visits)
		}
	}
	if len(graph.soundVisits) != len(tasks) {
		t.Fatalf("visited %d tasks, want %d", len(graph.soundVisits), len(tasks))
	}
}

func TestTaskGraphReopenInvalidatesCompletedDescendantsWithoutRewriting(t *testing.T) {
	base := graphRecord("base", domain.StatusCompleted)
	downstream := graphRecord("downstream", domain.StatusCompleted, base.ID)
	initial := NewTaskGraph([]domain.Task{downstream, base}, nil)
	if state := initial.State(downstream.ID); !state.Drained || state.Inconsistent {
		t.Fatalf("initial downstream = %+v", state)
	}
	base.Status = domain.StatusReadyToStart
	reopened := NewTaskGraph([]domain.Task{downstream, base}, nil)
	state := reopened.State(downstream.ID)
	if state.Drained || !state.Inconsistent || state.Gate != GateBlocked || downstream.Status != domain.StatusCompleted {
		t.Fatalf("reopened downstream = %+v persisted=%s", state, downstream.Status)
	}
}

func TestTaskGraphSameSourceSnapshotComparesUnreadableIdentity(t *testing.T) {
	left := NewTaskGraph(nil, []domain.FileProblem{{Path: "/planning/tasks/aaaaaaaaaaaa-left.md", Message: "bad yaml"}})
	right := NewTaskGraph(nil, []domain.FileProblem{{Path: "/planning/tasks/bbbbbbbbbbbb-right.md", Message: "bad yaml"}})
	if left.SameSourceSnapshot(right) {
		t.Fatal("different unreadable task sets compared as the same source snapshot")
	}
}

func TestValidateTaskGraphMutationPlanPreservesSemanticWriteOrder(t *testing.T) {
	alpha := graphRecord("alpha", domain.StatusReadyToStart)
	beta := graphRecord("beta", domain.StatusReadyToStart, alpha.ID)
	graph := NewTaskGraph([]domain.Task{alpha, beta}, nil)
	plan := TaskGraphMutationPlan{TaskWrites: []TaskDependencyWrite{
		{TaskID: beta.ID},
		{TaskID: alpha.ID, DependsOn: []string{beta.ID}},
	}}
	validated, err := ValidateTaskGraphMutationPlan(graph, plan)
	if err != nil {
		t.Fatal(err)
	}
	if validated.TaskWrites[0].TaskID != beta.ID || validated.TaskWrites[1].TaskID != alpha.ID {
		t.Fatalf("validator reordered semantic durable prefixes: %+v", validated.TaskWrites)
	}

	unsafe := TaskGraphMutationPlan{TaskWrites: []TaskDependencyWrite{
		{TaskID: alpha.ID, DependsOn: []string{beta.ID}},
		{TaskID: beta.ID},
	}}
	if _, err := ValidateTaskGraphMutationPlan(graph, unsafe); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("unsafe planner order error = %v", err)
	}
}

func TestValidateTaskGraphMutationSourceNamesManualRepairPath(t *testing.T) {
	task := graphRecord("manual-repair", domain.StatusReadyToStart, "not-a-stable-id")
	err := ValidateTaskGraphMutationSource(NewTaskGraph([]domain.Task{task}, nil))
	if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), task.Path) ||
		!strings.Contains(err.Error(), "field depends_on") ||
		!strings.Contains(err.Error(), "repair the graph-owned frontmatter directly") ||
		!strings.Contains(err.Error(), "tskflwctl lint") {
		t.Fatalf("broken graph recovery guidance = %v", err)
	}
}

func sortStrings(values []string) {
	slices.Sort(values)
}
