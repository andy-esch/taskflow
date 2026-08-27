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
		Status: status, DependsOn: append([]string(nil), dependencies...),
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
	// Degraded means canonical queries remain explanatory, but dispatch does not
	// claim eligible work while the legacy constraint is still hidden from the DAG.
	if state := graph.State(dependent.ID); state.Gate != GateClear || state.Eligible {
		t.Fatalf("degraded candidate state = %+v", state)
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

func sortStrings(values []string) {
	slices.Sort(values)
}
