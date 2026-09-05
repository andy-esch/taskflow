---
schema: 1
id: 6g71vzq8wdnj
bucket: closed
area: add-a-guarded-repair-path-for-broken-dependency-graphs-antigravity
date: "2026-09-05"
updated_at: "2026-09-05"
---
# Audit: add a guarded repair path for broken dependency graphs — antigravity — 2026-09-05

> Reviewer assignment: Antigravity.
> Design and architectural audit for task `6g4g8gatbnrs-add-a-guarded-repair-path-for-broken-dependency-graphs.md`.
>
> Finding grammar is exact: uses `#### H1. <title> · **Status:** open` (or M1/L1).
> Evaluated against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md), [ARCHITECTURE.md](../../docs/ARCHITECTURE.md), and production code in `internal/core` and `internal/store`.

## Context & Review Brief

Taskflow maintains a global, repository-wide Directed Acyclic Graph (DAG) of tasks within each planning space ([ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md)). Task dependencies are owned directly by the dependent task in its frontmatter (`depends_on: []`). Threads provide cross-cutting initiative projections over this global graph without owning dependency edges.

Under Taskflow's core integrity model:
1. Ordinary dependency mutations (`task depend add`, `task depend remove`, `task depend migrate`) and task lifecycle transitions (`task start`, `task complete`, generic `task move`) **fail closed** whenever the repository graph is in [`GraphBroken`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L25).
2. Generic field setters (`task set`, `task edit`) and automated text-fixers (`lint --fix`) explicitly refuse to mutate graph-owned frontmatter (`depends_on`, `blocked_by`, `dependencies`, `blocks`), even under `--force`.
3. Consequently, if a repository enters `GraphBroken`—via direct manual editing, conflicting Git merges, or corrupted files—**no product command currently exists to repair the graph**. Users are forced to manually hand-edit Markdown frontmatter.

Task [`6g4g8gatbnrs`](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6g4g8gatbnrs-add-a-guarded-repair-path-for-broken-dependency-graphs.md) proposes adding a guarded repair capability. This audit performs a rigorous, adversarial design pass on that capability, challenging its core assumptions, proving safety invariants, defining the core/store seams, and establishing the command and manifest contracts.

---

## Answers to the 12 Design Questions

### 1. What is the precise repair invariant?

A valid graph repair transformation $P: G_0 \to G_{\text{final}}$ must satisfy five strict invariants:

1. **Strictly Subtractive Edge Invariant:**
   $E(G_{\text{final}}) \subset E(G_0)$. A repair operation is strictly an edge-pruning or reference-purging action. It **never** introduces a new canonical dependency edge that did not already exist in $G_0$. (Adding dependencies remains reserved for [`AddTaskDependencies`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_operations.go#L91) on an already healthy graph).
2. **Non-Disruption of Unrelated Constraints:**
   For any task $T$ not targeted by $P$, $\text{DependsOn}(T, G_{\text{final}}) = \text{DependsOn}(T, G_0)$. For any task $T$ targeted by $P$, only the specific elements matching diagnosed problem tokens are removed. Valid, acyclic constraints are immutable during repair.
3. **Monotonic Problem Elimination:**
   The set of attributable problem instances in $G_{\text{final}}$ must be a strict subset of $G_0$:
   $$\text{Problems}(G_{\text{final}}) \subsetneq \text{Problems}(G_0) \quad \text{and} \quad \text{Problems}(G_{\text{final}}) \setminus \text{Problems}(G_0) = \emptyset$$
   No new problem codes, new problem tasks, or new problem edges may appear.
4. **Attributable Causality:**
   Every edge removal must be causally justified by an existing [`GraphProblem`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L49) in $G_0$. An edge can be broken if and only if it is an exact member of a diagnosed cyclic strongly connected component (SCC).
5. **Durable Prefix Commutativity:**
   Because all valid repair operations are set-subtractions on task-local `depends_on` lists, each task write is an independent edge deletion. For any two tasks $A$ and $B$, applying $A$ then $B$ yields the identical prospective graph state as applying $B$ then $A$. Any durable prefix leaves each modified file syntactically valid and strictly cleaner than before.

### 2. What deterministic problem measure proves monotonic progress without discarding unrelated constraints?

A scalar count like `len(graph.Problems())` is mathematically invalid because Tarjan’s SCC algorithm emits one [`ProblemCycle`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L42) per task in an SCC. Breaking an edge can split an SCC of 5 tasks into two smaller SCCs of 2 and 2 tasks, which changes component groupings without clean scalar monotonicity.

Instead, the deterministic measure is the **Attributable Problem Vector** $\vec{\mu}(G) \in \mathbb{N}^7$:
$$\vec{\mu}(G) = \langle P_{\text{unreadable}}, P_{\text{id\_drift}}, P_{\text{invalid\_id}}, P_{\text{missing\_dep}}, P_{\text{self}}, P_{\text{dup}}, |E_{\text{cyclic}}| \rangle$$
where $|E_{\text{cyclic}}|$ is the exact count of directed canonical edges $(u, v)$ such that $u$ and $v$ belong to the same cyclic strongly connected component.

**Proof of Monotonicity under Subtractive Pruning:**
Because repair only removes elements from `depends_on`:
- Pruning a dangling reference decrements $P_{\text{missing\_dep}}$ by 1, leaving other dimensions unchanged.
- Pruning an invalid ID decrements $P_{\text{invalid\_id}}$ by 1.
- Removing a self-edge decrements $P_{\text{self}}$ by 1.
- Deduplicating decrements $P_{\text{dup}}$ by 1.
- Removing a directed edge $e = (u, v)$ where $u, v \in \text{SCC}_k$ strictly removes $e$ from $E_{\text{cyclic}}$. Any cycle in $G' = G \setminus \{e\}$ is a subgraph of $G$; therefore, no new cycles can form, and no new nodes can enter any SCC. Thus, $|E_{\text{cyclic}}(G')| < |E_{\text{cyclic}}(G)|$.
- For every step $k$, $\vec{\mu}(G_k) <_{\text{lex}} \vec{\mu}(G_{k-1})$ and component-wise $\vec{\mu}(G_k) \le \vec{\mu}(G_{k-1})$.

### 3. Is “every durable prefix must improve” achievable for cycles and multi-file repairs?

**Yes, but only under the Vector Problem Measure, not under `prefixGraph.Health() != GraphBroken`.**
The existing task draft stated: *"rejects any plan whose prefix fails to improve or preserve safety while moving monotonically toward a healthy graph."*

If taken to mean "every prefix must be healthy", it is physically impossible when multiple defects exist across multiple files. For example, if Task A has a dangling ID and Task B participates in a cycle, writing Task A first leaves the repository graph in [`GraphBroken`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L25) due to Task B.

**The Correct Recovery Contract:**
1. **Preflight Plan Verification:** Core verifies that the *complete* plan moves the graph toward health (or full recovery), and that every planned task write $w_i \in P$ removes at least one diagnosed defect on that task.
2. **Deterministic Step Monotonicity:** For each durable write $w_i$, the intermediate graph $G_i$ satisfies $\text{Problems}(G_i) \subsetneq \text{Problems}(G_{i-1})$.
3. **Idempotent Resumability:** If execution aborts after durable prefix $k < n$, the receipt returns `AppliedTaskIDs: [T_1..T_k]` and `RemainingTaskIDs: [T_{k+1}..T_n]`. Re-running the repair on the resulting graph is a clean no-op for $T_1..T_k$ and converges $T_{k+1}..T_n$.

### 4. Which repairs can be safely inferred vs. require explicit user selection?

- **Safely Inferred (Deterministic, Unambiguous Auto-Resolution):**
  - Duplicate dependencies (`depends_on: [A, A]` $\to$ deduplicate to `[A]`).
  - Self-dependencies (`depends_on: [A]` on task A $\to$ purge `A`).
  - Dangling references (ID is valid 12-char Crockford Base32, but task does not exist $\to$ prune edge).
  - Malformed dependency tokens (syntactically invalid IDs, slugs, or empty strings $\to$ prune token).
- **Requires Explicit Edge Selection by User:**
  - **Cycles:** A cycle represents a circular causality contradiction (e.g., $A \to B \to C \to A$). The tool cannot know which dependency is the accidental or obsolete edge. Automatically guessing (e.g., dropping the newest edge) risks inverting critical work dependencies. The user must designate the edge to sever via `--break <dependent>:<prerequisite>`.
  - **Legacy Field Purging:** Clearing unparseable or ambiguous legacy fields (`blocked_by`, `dependencies`, `blocks`) must be explicitly requested via `--clear-legacy <task>` or confirmed in a manifest.

### 5. What should the core repair-plan and receipt types contain?

The core repair types must be adapter-neutral, stable-ID addressed, and carry full problem/impact attribution:
- [`GraphRepairPlan`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_operations.go): Holds planned repair items, targeted task IDs, expected problem delta, and ordered [`TaskDependencyWrite`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/store.go#L59) records.
- [`GraphRepairReceipt`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_operations.go): Carries `DryRun`, `Changed`, `InitialHealth`, `FinalHealth`, `ResolvedProblems`, `RemainingProblems`, `PlannedTaskIDs`, `AppliedTaskIDs`, and `RemainingTaskIDs`.

### 6. What belongs in core ports vs. filesystem adapter?

- **Core Ports:**
  - Graph diagnosis and problem extraction ([`TaskGraph.Problems`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L769)).
  - Auto-repair plan compilation (identifying dangling, duplicate, self, and invalid IDs).
  - Explicit cycle break validation (verifying candidate edge exists and is in a cyclic SCC).
  - Monotonicity proof validation over the problem vector.
  - Calculation of downstream unblocking/frontier state impacts.
- **Store Adapter (`internal/store`):**
  - Lock acquisition ([`checkedWriteLock`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/lock.go#L115)) and callback re-entry prevention.
  - Authoritative snapshot loading.
  - Whole-repository snapshot CAS comparison ([`SameSourceSnapshot`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L878)).
  - Surgical frontmatter YAML update preserving comments and bodies ([`updateFrontmatter`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/fsstore.go#L184)).
  - Immediate per-file pre-write CAS ([`verifyUnchanged`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/cas.go#L48)).
  - Atomic file replacement ([`writeFileAtomic`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/graphmutation.go#L100)).

### 7. How should stale snapshots, partial commits, retries, and already-applied operations behave?

- **Stale Snapshots / OCC Conflict:** If another writer modifies any task file between graph loading and write materialization, the whole-snapshot hash check fails with [`domain.ErrConflict`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/domain/errors.go) (exit 14).
- **Partial Commits:** If write $k$ of $n$ fails (I/O failure or permission error), already-written files $1..k-1$ remain durable. The error is wrapped in a typed failure preserving `AppliedTaskIDs: [1..k-1]` and `RemainingTaskIDs: [k..n]`.
- **Retries:** Retrying the exact same command or manifest is completely safe. Tasks $1..k-1$ are re-scanned, recognized as already repaired, and skipped. Only tasks $k..n$ are mutated.
- **Already-applied operations:** If an edge marked for removal is already gone, it results in an `Outcome: skipped` item, producing no disk write.

### 8. Which CLI shape is clearest?

A unified `task depend repair` command supporting:
1. Direct auto-repair flags: `tskflwctl task depend repair --prune-dangling --prune-invalid --dedup --remove-self`
2. Explicit cycle break flags: `tskflwctl task depend repair --break <dependent>:<prereq>...`
3. Manifest-based workflow for complex multi-cycle repairs:
   - `tskflwctl task depend repair plan --out repair.yaml`
   - `tskflwctl task depend repair apply repair.yaml [--dry-run]`

*Why not `lint --fix`?* [`lint --fix`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/cli/lint.go#L77) is an unassisted batch text fixer that bypasses [`core.Service`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service.go). Graph repair must be deliberate, guarded, and transactionally receipted.

### 9. What should human and JSON previews expose before mutation?

Before touching disk, `--dry-run` must expose:
- Initial health and problem count.
- Itemized removals: task ID, action (`prune-dangling`, `remove-self`, `break-cycle`, etc.), prerequisite ID/value, and attributable problem.
- Tasks whose files will be rewritten.
- Projected graph health (`healthy`, `degraded`, or remaining problems if partial).
- Downstream impact: tasks unblocked or transitioning to `eligible`.

### 10. How should lint/status guidance change once repair exists?

- Diagnostic messages in `status`, `board`, and `thread show` will change from:
  `"...repair the graph-owned frontmatter directly, then run 'tskflwctl lint'"`
  to:
  `"...run 'tskflwctl task depend repair'"` (or `--break <dep>:<prereq>` if cyclic).
- `lint --fix` will report skipped graph-owned fields with:
  `"graph-owned defect detected (%s); run 'tskflwctl task depend repair'"`.

### 11. Does the existing graph representation support this cleanly?

**Yes.** [`TaskGraph`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L274) already encapsulates Tarjan's SCC cycle detection, representative cycle paths, problem categorization, memoized sound completion, and topological ordering. No external graph package is required or justified.

### 12. What compatibility or machine-schema implications exist?

- Persisted Markdown schemas (`depends_on: []`) remain strictly unchanged.
- Wire envelopes add `wire.GraphRepairReceiptJSON`.
- CLI exit codes preserve standard conventions: `0` (success), `11` ([`domain.ErrValidation`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/domain/errors.go) for invalid plan/syntax), `14` ([`domain.ErrConflict`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/domain/errors.go) for concurrent OCC collision).

---

## Core Architecture, Ports & Store Boundary

```text
 ┌──────────────────────────────────────────────────────────────────┐
 │                        CLI / TUI Layer                           │
 │  tskflwctl task depend repair [--break|--prune-dangling|--plan]  │
 └──────────────────────────────┬───────────────────────────────────┘
                                │
                                ▼
 ┌──────────────────────────────────────────────────────────────────┐
 │                         core.Service                             │
 │  - PlanGraphRepair() [pure auto-diagnosis & plan compilation]    │
 │  - RepairTaskGraph() [coordinates store mutation & receipt]      │
 └──────────────┬───────────────────────────────────▲───────────────┘
                │ Invokes Planner                   │ Passes Snapshot
                ▼                                   │
 ┌──────────────────────────────────────────────────┴───────────────┐
 │               store.TaskGraphRepairStore (Port)                  │
 │  MutateTaskGraphRepair(now, dryRun, planner)                     │
 └──────────────────────────────┬───────────────────────────────────┘
                                │ Implemented by store.FS
                                ▼
 ┌──────────────────────────────────────────────────────────────────┐
 │                         store.FS                                 │
 │  1. checkedWriteLock() [process + OS platform file lock]         │
 │  2. LoadTaskGraph() [authoritative strict snapshot]              │
 │  3. Execute core planner callback -> GraphRepairPlan             │
 │  4. ValidateGraphRepairPlan() [pure problem vector check]        │
 │  5. materializeTaskGraphPlan() [YAML surgery + updated_at]       │
 │  6. SameSourceSnapshot() [whole-repo CAS preflight]              │
 │  7. Sequential per-file verifyUnchanged() CAS + writeFileAtomic()│
 │  8. Release lock & return TaskGraphMutationResult                │
 └──────────────────────────────────────────────────────────────────┘
```

### Proposed Core Types (`internal/core/dependency_operations.go`)

```go
// GraphRepairAction classifies the exact pruning operation on a task.
type GraphRepairAction string

const (
	RepairDeduplicate    GraphRepairAction = "deduplicate"
	RepairRemoveSelf     GraphRepairAction = "remove-self"
	RepairPruneDangling  GraphRepairAction = "prune-dangling"
	RepairPruneInvalidID GraphRepairAction = "prune-invalid-id"
	RepairBreakCycle     GraphRepairAction = "break-cycle"
	RepairClearLegacy    GraphRepairAction = "clear-legacy"
)

type GraphRepairItem struct {
	TaskID         string            `json:"task_id"`
	Action         GraphRepairAction `json:"action"`
	PrerequisiteID string            `json:"prerequisite_id"`
	Reason         string            `json:"reason"`
}

type GraphRepairPlan struct {
	Items      []GraphRepairItem
	TaskWrites []TaskDependencyWrite
}

type GraphRepairReceipt struct {
	DryRun            bool
	Changed           bool
	InitialHealth     GraphHealth
	FinalHealth       GraphHealth
	ResolvedProblems  []GraphProblem
	RemainingProblems []GraphProblem
	Items             []GraphRepairItem
	Impacts           []TaskGraphStateImpact
	PlannedTaskIDs    []string
	AppliedTaskIDs    []string
	RemainingTaskIDs  []string
}
```

### Proposed Store Port (`internal/core/store.go`)

```go
type TaskGraphRepairPlanner func(*TaskGraph) (GraphRepairPlan, error)

type TaskGraphRepairStore interface {
	MutateTaskGraphRepair(now time.Time, dryRun bool, planner TaskGraphRepairPlanner) (TaskGraphMutationResult, error)
}
```

---

## Validation & Store Materialization Implementation

### Core Validation Algorithm (`internal/core/dependency_graph_mutation.go`)

```go
func ValidateGraphRepairPlan(sourceGraph *TaskGraph, plan GraphRepairPlan) (GraphRepairPlan, error) {
	if sourceGraph == nil {
		return GraphRepairPlan{}, fmt.Errorf("%w: authoritative task graph is required", domain.ErrValidation)
	}
	if sourceGraph.Health() == GraphHealthy {
		return GraphRepairPlan{}, fmt.Errorf("%w: repository graph is already healthy", domain.ErrValidation)
	}

	taskIDs := sourceGraph.TaskIDs()
	prospective := make(map[string]domain.Task, len(taskIDs))
	for _, id := range taskIDs {
		t, _ := sourceGraph.Task(id)
		prospective[id] = t
	}

	removalsByTask := make(map[string][]GraphRepairItem)
	for _, item := range plan.Items {
		if !id.Valid(item.TaskID) {
			return GraphRepairPlan{}, fmt.Errorf("%w: invalid task id %q", domain.ErrValidation, item.TaskID)
		}
		if _, exists := prospective[item.TaskID]; !exists {
			return GraphRepairPlan{}, fmt.Errorf("%w: task %s does not exist", domain.ErrValidation, item.TaskID)
		}
		removalsByTask[item.TaskID] = append(removalsByTask[item.TaskID], item)
	}

	var writes []TaskDependencyWrite
	for taskID, items := range removalsByTask {
		task := prospective[taskID]
		newDependsOn := append([]string(nil), task.DependsOn...)
		clearLegacy := false

		for _, item := range items {
			switch item.Action {
			case RepairRemoveSelf:
				if !slices.Contains(newDependsOn, taskID) {
					return GraphRepairPlan{}, fmt.Errorf("%w: task %s has no self-dependency", domain.ErrValidation, taskID)
				}
				newDependsOn = slices.DeleteFunc(newDependsOn, func(v string) bool { return v == taskID })

			case RepairDeduplicate:
				seen := make(map[string]bool)
				deduped := make([]string, 0, len(newDependsOn))
				for _, dep := range newDependsOn {
					if !seen[dep] {
						seen[dep] = true
						deduped = append(deduped, dep)
					}
				}
				newDependsOn = deduped

			case RepairPruneDangling, RepairPruneInvalidID:
				newDependsOn = slices.DeleteFunc(newDependsOn, func(v string) bool { return v == item.PrerequisiteID })

			case RepairBreakCycle:
				if !slices.Contains(newDependsOn, item.PrerequisiteID) {
					return GraphRepairPlan{}, fmt.Errorf("%w: task %s does not depend on %s", domain.ErrValidation, taskID, item.PrerequisiteID)
				}
				if !sourceGraph.IsEdgeInCycle(item.PrerequisiteID, taskID) {
					return GraphRepairPlan{}, fmt.Errorf("%w: edge %s -> %s does not participate in a cycle",
						domain.ErrValidation, item.PrerequisiteID, taskID)
				}
				newDependsOn = slices.DeleteFunc(newDependsOn, func(v string) bool { return v == item.PrerequisiteID })

			case RepairClearLegacy:
				clearLegacy = true
			}
		}

		sort.Strings(newDependsOn)
		task.DependsOn = newDependsOn
		if clearLegacy {
			task.LegacyBlockedBy = nil
			task.LegacyDependencies = nil
			task.LegacyBlocks = nil
			task.LegacyDependencyFields = nil
		}
		prospective[taskID] = task

		writes = append(writes, TaskDependencyWrite{
			TaskID:      taskID,
			DependsOn:   newDependsOn,
			ClearLegacy: clearLegacy,
		})
	}

	sort.Slice(writes, func(i, j int) bool { return writes[i].TaskID < writes[j].TaskID })

	// Monotonicity verification
	sourceProblems := sourceGraph.Problems()
	currentTasks := make(map[string]domain.Task, len(taskIDs))
	for _, id := range taskIDs {
		t, _ := sourceGraph.Task(id)
		currentTasks[id] = t
	}

	for _, write := range writes {
		currentTasks[write.TaskID] = prospective[write.TaskID]
		prefixGraph := taskGraphFromMap(taskIDs, currentTasks)

		for _, p := range prefixGraph.Problems() {
			if !containsEquivalentProblem(sourceProblems, p) {
				return GraphRepairPlan{}, fmt.Errorf("%w: repair write on %s introduces new problem: %s",
					domain.ErrValidation, write.TaskID, p.Message)
			}
		}
	}

	plan.TaskWrites = writes
	return plan, nil
}
```

### Store Repair Method (`internal/store/graphmutation.go`)

```go
func (s *FS) MutateTaskGraphRepair(now time.Time, dryRun bool, planner core.TaskGraphRepairPlanner) (result core.TaskGraphMutationResult, err error) {
	result.DryRun = dryRun
	if planner == nil {
		return result, fmt.Errorf("%w: repair planner is required", domain.ErrValidation)
	}
	if now.IsZero() {
		return result, fmt.Errorf("%w: mutation time is required", domain.ErrValidation)
	}
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return result, err
	}

	unlock, err := s.checkedWriteLock()
	if err != nil {
		return result, err
	}
	defer func() {
		if releaseErr := unlock(); releaseErr != nil {
			err = errors.Join(err, fmt.Errorf("release graph repair lock: %w", releaseErr))
		}
	}()

	graph, err := core.LoadTaskGraph(s)
	if err != nil {
		return result, fmt.Errorf("load authoritative graph: %w", err)
	}

	plan, err := planner(graph)
	if err != nil {
		return result, err
	}

	validatedPlan, err := core.ValidateGraphRepairPlan(graph, plan)
	if err != nil {
		return result, err
	}
	result.Plan = core.TaskGraphMutationPlan{TaskWrites: validatedPlan.TaskWrites}

	writes, err := s.materializeTaskGraphPlan(graph, result.Plan, now)
	if err != nil {
		return result, err
	}
	if dryRun {
		return result, nil
	}

	currentGraph, err := core.LoadTaskGraph(s)
	if err != nil {
		return result, fmt.Errorf("re-read authoritative graph: %w", err)
	}
	if !graph.SameSourceSnapshot(currentGraph) {
		return result, fmt.Errorf("repository task graph changed while planning repair; retry: %w", domain.ErrConflict)
	}

	for _, write := range writes {
		if err := verifyUnchanged(s.resolvePath, write.taskID, write.path, write.ifVersion, "task", "repair dependency"); err != nil {
			return result, err
		}
		if err := writeFileAtomic(write.path, write.content, 0o644); err != nil {
			return result, fmt.Errorf("write repair for task %s: %w", write.taskID, err)
		}
		result.AppliedTaskIDs = append(result.AppliedTaskIDs, write.taskID)
	}

	return result, nil
}
```

---

## Command Surface & Manifest Contract

### CLI Invocations

```bash
# Preview unambiguous auto-repairs
tskflwctl task depend repair --dry-run

# Apply auto-repairs (duplicates, self-edges, dangling IDs, malformed tokens)
tskflwctl task depend repair --prune-dangling --prune-invalid --dedup --remove-self

# Break a specific cycle edge
tskflwctl task depend repair --break 6g4g8gatbnrs:6g3q4rt7mgjn

# Generate a repair manifest for complex or ambiguous states
tskflwctl task depend repair plan --out repair-plan.yaml

# Apply an edited repair manifest
tskflwctl task depend repair apply repair-plan.yaml [--dry-run]
```

### Repair Manifest Schema (`repair-plan.yaml`)

```yaml
schema: 1
space_id: spc_9f82k4ma81vd
generated_at: "2026-09-05T09:30:00Z"
target_health: healthy

diagnosed_problems:
  - code: cycle
    tasks: [6g4g8gatbnrs, 6g3q4rt7mgjn]
    path: 6g4g8gatbnrs -> 6g3q4rt7mgjn -> 6g4g8gatbnrs
  - code: missing-dependency
    task: 6g697mp8s4tx
    missing_id: 6g9999999999
  - code: duplicate-dependency
    task: 6g6scc9jgxae
    duplicate_id: 6g3q4rtv8d0a

repairs:
  - task_id: 6g4g8gatbnrs
    action: break-cycle
    remove_prerequisite: 6g3q4rt7mgjn

  - task_id: 6g697mp8s4tx
    action: prune-dangling
    remove_prerequisite: 6g9999999999

  - task_id: 6g6scc9jgxae
    action: deduplicate
```

---

## Failure & Recovery Matrix

| Failure Mode | Trigger Point | State on Disk | Recovery Path | Sentinels & Attributions |
| :--- | :--- | :--- | :--- | :--- |
| **Invalid Cycle Edge** | Core Plan Validation | Untouched | Fix flag: edge must belong to an actual cyclic SCC | Returns `domain.ErrValidation` (exit 11). Names edge and cycle members. |
| **Concurrent Out-of-Band Edit** | Whole-Snapshot Preflight CAS | Untouched | Re-run repair command on newly edited snapshot | Returns `domain.ErrConflict` (exit 14). Lock released cleanly. |
| **File CAS Mismatch** | Prior to Write of Task $k$ | Tasks $1..k-1$ applied; Task $k$ unwritten | Re-run repair command; already-repaired files are clean no-ops | `GraphRepairFailure`: preserves `AppliedTaskIDs` and `RemainingTaskIDs`. |
| **Disk/IO Failure on Task $k$** | Atomic rename of Task $k$ | Tasks $1..k-1$ applied; Task $k$ intact | Resolve OS disk/permission issue, then retry same command | Preserves exact receipt with remaining files. |
| **Re-entry Attempt** | Planner calls Store method | Untouched | Fix bug: planner must remain purely functional | Returns `repositoryPlannerReentryError()` (`domain.ErrConflict`). |
| **Partial Cycle Resolution** | Plan breaks 1 of 2 disjoint cycles | Cycle 1 resolved; Cycle 2 remains | Run repair with `--break` for remaining cycle | Exit 0 with `final_health: "broken"` and `RemainingProblems`. |

---

## Adversarial Test Strategy

1. **Deterministic Edge Pruning Unit Tests:**
   - Fixture with self-edge, duplicate edge, invalid string, and dangling 12-char ID. Proves all four prune deterministically and leave valid edges intact.
   - Fixture with figure-8 overlapping cycle: proves specifying one bridge edge breaks both cycles in a single write.
   - Rejection of acyclic edge removal: attempting to remove a valid prerequisite under `--break` must return `domain.ErrValidation`.
2. **Interrupted Multi-File Write & Resumption:**
   - Injected failure via `testHookAfterGraphWrite` after write 1 of 3.
   - Assert receipt contains `AppliedTaskIDs: [Task1]`, `RemainingTaskIDs: [Task2, Task3]`.
   - Assert Task 1 on disk is modified; Tasks 2 and 3 are untouched.
   - Execute second repair call without flags: proves Task 1 is recognized as no-op and Tasks 2 and 3 are successfully applied.
3. **Adversarial Concurrency (OCC & Races):**
   - Inject `testHookBeforeGraphVerify`: raw editor modifies unrelated frontmatter field of Task 1 before whole-snapshot verification. Asserts repair aborts with `domain.ErrConflict`.
   - Inject `testHookBeforeGraphWrite`: raw editor modifies Task 2 immediately before its per-file CAS. Asserts Task 1 committed, Task 2 aborts, receipt accurately records split.
   - Two concurrent processes executing `task depend repair` with different cycle-breaking choices: OS platform lock serializes them; the second process re-reads the snapshot, detects the cycle is already resolved, and exits with a zero-change receipt.
4. **Degraded vs. Broken Precedence:**
   - Repository has both a legacy `blocked_by` field and a broken canonical cycle.
   - Asserts initial health is `broken`.
   - Repairing the cycle transitions health to `degraded` (not `healthy`).
   - Asserts ordinary mutations remain blocked, directing user to `tskflwctl task depend migrate`.

---

## Findings Ranked by Severity

#### H1. Prefix Invariant Contradiction in Existing Planning · **Status:** tracked by 6g4g8gatbnrs
**File:** `planning/tasks/6g4g8gatbnrs-add-a-guarded-repair-path-for-broken-dependency-graphs.md`
**Component:** core/graph
**Effort:** M
**Urgency:** acute

Task `6g4g8gatbnrs` requires that every durable prefix must move monotonically toward a healthy graph. If implemented using the existing `ValidateTaskGraphMutationPlan`, every multi-file repair on an already-broken graph fails immediately at prefix 1 because intermediate graphs remain `GraphBroken`. The acceptance criteria must formally adopt the Attributable Problem Vector $\vec{\mu}(G)$ and require component-wise monotone decrease instead of prefix graph health.

**Resolution:** The task now separates structural prefix safety from final
strict improvement and permits explicit residual broken state.

#### H2. Missing Store Mutation Seam for Broken Graph Recovery · **Status:** tracked by 6g4g8gatbnrs
**File:** `internal/store/graphmutation.go`
**Component:** store/graphmutation
**Effort:** S
**Urgency:** acute

`TaskGraphMutationStore.MutateTaskGraph` hard-fails on line 57 via `ValidateTaskGraphMutationSource` if `graph.Health() == GraphBroken`. There is currently no store capability through which a repair planner can execute. `TaskGraphRepairStore.MutateTaskGraphRepair` must be introduced as a dedicated port to preserve fail-closed semantics for ordinary mutations.

**Resolution:** The task now requires a dedicated broken-source repair mutation
port rather than widening ordinary graph mutation.

#### M1. Ambiguity in Handling Cyclic Strongly Connected Components · **Status:** tracked by 6g4g8gatbnrs
**File:** `internal/core/dependency_graph.go`
**Component:** core/dependency_graph
**Effort:** S
**Urgency:** soon

`stronglyConnectedCycles` in `internal/core/dependency_graph.go` computes one representative cycle per SCC. In complex SCCs with multiple chord edges, severing the representative cycle might leave an alternate sub-cycle. The repair validation must evaluate the entire remaining prospective SCC to confirm whether the component is completely acyclic or requires additional edge cuts, reporting remaining cycles deterministically.

**Resolution:** Cycle edges require explicit operator selection; structural SCC
evidence validates non-worsening without heuristic cycle choice.

#### L1. Diagnostic Guidance Omission in Lint and Status · **Status:** tracked by 6g4g8gatbnrs
**File:** `internal/core/service.go`
**Component:** core/service
**Effort:** XS
**Urgency:** eventually

`dependencyLintIssues` and `taskGraphHealthDetail` instruct users to *"repair the graph-owned frontmatter directly, then run 'tskflwctl lint'"*. Once the repair path is implemented, these error strings and lint findings must be updated to reference `tskflwctl task depend repair`.

---

**Resolution:** The scoped CLI now includes defect-specific diagnosis, copyable
selectors, lint guidance, and residual-problem output.

## Maintainer Decisions Required Before Implementation

1. **CLI Top-Level Syntax:** Confirm adoption of `tskflwctl task depend repair` (with `--break`, `--prune-dangling`, `--dedup`) vs. top-level `tskflwctl repair graph`. *(Recommendation: keep under `task depend repair` to preserve command grouping).*
2. **Non-Interactive Default for Auto-Repairs:** Should `task depend repair` without flags run `--dry-run` by default (requiring `--write` / `-y` to commit), or run interactively when attached to a TTY? *(Recommendation: default to `--dry-run` unless `--write` or specific prune flags are passed).*
3. **Legacy Field Repair Scope:** Should `task depend repair` support clearing broken legacy fields (`--clear-legacy`), or should legacy remediation remain strictly manual until `task depend migrate` can run? *(Recommendation: support `--clear-legacy` on a broken task so users do not have to hand-edit YAML).*
