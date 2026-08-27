---
schema: 1
id: 6g417v97bx8s
bucket: closed
area: canonical-task-dependency-read-foundation
date: "2026-08-26"
---

# Audit: Canonical Task-Dependency Read Foundation — 2026-08-26

Adversarial implementation review of the canonical task-dependency read foundation in Taskflow (task `6g3q4rst78qy`, epic `30-threads-and-task-dependency-graphs`, branch `feat/canonical-task-dependency-reads`), stress-tested against ADR-0006, the production domain/store/wire models, and real repository data.

**Executive Verdict: Safe with amendments.** The core architectural model holds: `depends_on` is cleanly modeled as a sorted, duplicate-free set of stable IDs; `TaskGraph` provides an immutable, thread-safe strict snapshot; reconvergent sound-completion memoization is proven O(V+E); and hard mutation guardrails on generic `task set` (including `--force`) and interactive `task edit` successfully prevent invalid graph states from entering the store. However, two defects must be resolved before this slice merges or immediately as part of slice 2:
1. Cycle detection in `deterministicCycles` (`internal/core/dependency_graph.go:795`) has an algorithmic blind spot on intersecting cycles (multi-branch feedback graphs), omitting cycles that traverse through completed DFS nodes (`state = 2`) and misclassifying participating tasks as unstarted or unsound-completed rather than `cycle` (H1).
2. The decision to make ordinary `tskflwctl lint` fail with exit code 11 on the six resolvable legacy `blocked_by` fields breaks repo CI while simultaneously prohibiting all CLI tools from repairing those fields until slice 2 ships the guarded migration (M1).

Findings are classified as **[merge-blocking]**, **[pre-mutation prerequisite]**, or **[tracked follow-up]**.

---

## Findings

### High

#### H1. Cycle detection in `deterministicCycles` misses cycles in multi-cycle or intersecting graphs, leaving cycle members misattributed · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:795-835 | **Component:** core / graph analysis
**Effort:** M · **Urgency:** acute

**[pre-mutation prerequisite / merge-blocking for cycle correctness]**

`deterministicCycles` implements a standard 3-color DFS (`state 0: unvisited, 1: visiting, 2: finished`) to extract attributable cycle paths. When a vertex finishes its post-order traversal, it is marked `state = 2`. The inner neighbor loop (`lines 807-820`) only handles `case 0` (unvisited) and `case 1` (back-edge to a node in the current recursion stack); it contains no handler for `case 2`.

When a directed graph contains intersecting cycles (for example, two parallel branches `A -> B -> C -> A` and `A -> D -> C -> A` sharing node `C` and `A`, or any figure-8 feedback topology), the first cycle path marks node `C` as `state = 2`. When DFS subsequently traverses branch `D`, neighbor `C` is in `state = 2` and is skipped without further traversal.

Consequently:
1. `structure.Cycles` records only `[A, B, C, A]` and misses `[A, D, C, A]`.
2. `g.cycleMembers[D]` remains `false`.
3. `g.Problems()` omits the cycle involving `D`.
4. When queried via `g.State(D)` or `g.Blockers(...)`, `g.blockerReason(D)` returns `BlockerNotStarted`, `BlockerInFlight`, or `BlockerUnsoundCompleted` instead of `BlockerCycle`.
5. In slice 2 (guarded dependency operations), an agent or developer diagnosing why task `D` is blocked will be told it is unstarted rather than cyclic. If the user breaks the edge `B -> C` to fix the first cycle, the second cycle `A -> D -> C -> A` will suddenly emerge on the next run, violating the guarantee of deterministic and complete cycle attribution.

**Why tests missed it:** Existing contract and stress tests (`TestOwnedDAGAnalyzerContract`, `TestTaskGraphCycleBlockerReason`) only tested single, isolated 2-node or 3-node cycles (`a <-> b` and `c -> a -> b -> c`), never multi-cycle or intersecting graphs sharing vertices.

**Recommendation:** Adopt Tarjan's Strongly Connected Components (SCC) algorithm (which is strictly O(V+E)) to identify all cyclic subgraphs. Every node in an SCC of size $> 1$ (or with a self-edge) is unconditionally marked in `g.cycleMembers`. Extract canonical cycle paths within each SCC.

---

### Medium

#### M1. Intentionally failing ordinary `lint` on resolvable legacy `blocked_by` fields breaks repository CI and leaves users without a CLI remediation path · **Status:** fixed 2026-08-27

**File:** internal/core/service.go:346-363, internal/cli/lint.go:62, planning/tasks/6g3q4rst78qy-establish-canonical-task-dependencies-and-strict-graph-reads.md:70-74 | **Component:** core / cli / lint
**Effort:** S · **Urgency:** acute

**[merge-blocking / operational]**

`dependencyLintIssues` (`service.go:346-363`) maps every `LegacyDependencyDiagnostic` into a `domain.Issue`. In `runLint` (`internal/cli/lint.go:62-65`), any issue causes `tskflwctl lint` to exit with status code 11 (`validation failed: 6 item(s) with issues`). Running `go run ./cmd/tskflwctl lint --json` against this repository confirms exit status 11 on the six existing `blocked_by` tasks.

At the same time, `task set`, `task set --force`, and `task edit` strictly reject removing or modifying `blocked_by` (with error `"legacy dependency fields are removed only by the guarded migration"`). `lint --fix` also does not migrate or remove them.

This creates an operational deadlock:
- Merging slice 1 to `main` immediately breaks `tskflwctl lint` in CI and developer pre-commit hooks.
- Developers cannot fix the legacy fields using any supported CLI command.
- The only options are disabling CI lint checks, making manual regex/editor filesystem edits (which the ADR and task documentation explicitly warn against), or waiting for slice 2 (`6g3q4rt7mgjn`) to land and execute the guarded migration.

**Why tests missed it:** `internal/cli/lint_test.go:88-90` asserts that `lint` exits with 11 on legacy fields, treating this as a test pass rather than evaluating its impact on live repository workflows.

**Recommendation:** Either:
1. Treat resolvable legacy references as non-fatal warnings in `lint` (reporting them on stderr/JSON while exiting 0, reserving exit 11 for unresolvable/ambiguous legacy references or broken graphs) until slice 2 ships the guarded migration; OR
2. Bundle the guarded migration into this slice or merge slice 1 and slice 2 in immediate sequence before tagging or enforcing CI lint gates.

#### M2. Unreadable task files lose identity association, reporting downstream blockers as `missing` rather than `unreadable` · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:269-274, 349-352, 703-705 | **Component:** core / graph snapshot
**Effort:** S · **Urgency:** soon

**[pre-mutation prerequisite]**

`NewTaskGraph` accepts `unreadable []domain.FileProblem`. Because an unreadable task failed YAML parsing, it is not present in `g.tasks`. `ProblemUnreadable` is recorded in `g.problems` with `Path` populated but `TaskID: ""` (`lines 270-274`).

When a valid downstream task declares `depends_on: [6fjangd7kvh0]`, and `tasks/6fjangd7kvh0-alpha.md` is the unreadable file, `!taskExists(g.tasks, prerequisite)` triggers `ProblemMissingDependency` (`"task beta depends on missing task 6fjangd7kvh0"`), and `g.Blockers(...)` reports `BlockerMissing` rather than an unreadable or corrupted blocker.

If a developer introduces a YAML indentation error in a prerequisite task file, downstream tasks will claim that the prerequisite was deleted or never existed (`missing`), misleading users who see the file on disk.

**Why tests missed it:** Test fixtures tested unreadable files and missing dependencies as disjoint scenarios, never asserting blocker reason fidelity for a valid task depending on a file that exists on disk but failed to parse.

**Recommendation:** Extract the filename ID from `FileProblem.Path` (via `id.Extract` or `splitFlatName`) and record unreadable task IDs in a `map[string]domain.FileProblem` within `TaskGraph`. When `prerequisite` is in this unreadable map, emit a specific `ProblemUnreadableDependency` and assign `BlockerInvalidStatus` (or a dedicated `BlockerUnreadable` token) with the path to the unreadable file.

---

### Low

#### L1. Minimal `dagcontract.Run` suite provides insufficient validation for alternative DAG analyzers · **Status:** fixed 2026-08-27

**File:** internal/core/testdata/dagcontract/contract.go:14-43 | **Component:** core / testdata / dagcontract
**Effort:** S · **Urgency:** eventually

**[tracked follow-up]**

`dagcontract.Run` contains only three micro-fixtures: a 4-node wave, a 3-node cycle, and a 1-node self cycle. It omits multi-cycle graphs, disconnected frontiers, diamond reconvergences, and deep dependency chains.

Any future spike or developer evaluating an external graph library (such as an updated `dominikbraun/graph` or `gonum`) against `dagcontract.Run` could observe 100% pass rates while the library actually fails on complex topological wave ties, deep recursion, or multi-cycle attribution.

**Why tests missed it:** The contract was created as a minimal structural interface check rather than an exhaustive property-based benchmark.

**Recommendation:** Expand `dagcontract.Run` to include the deep/wide and reconvergent stress tests currently present in `dependency_graph_test.go`.

#### L2. Self-dependencies generate redundant dual diagnostics in lint output · **Status:** fixed 2026-08-27

**File:** internal/core/dependency_graph.go:335-342, 374-376, internal/core/service.go:329-345 | **Component:** core / graph analysis / lint
**Effort:** XS · **Urgency:** eventually

**[tracked follow-up]**

When a task declares `depends_on: [self_id]`, `newTaskGraph` emits `ProblemSelfDependency` (`"task X cannot depend on itself"`) and also appends the self-edge to `edges`. The analyzer then detects `[X, X]` as a cycle and emits `ProblemCycle` (`"dependency cycle: X -> X"`). In `dependencyLintIssues`, both issues are attached to task `X` under field `depends_on`.

A user who accidentally adds a self-dependency sees two separate lint lines describing the identical 1-node defect.

**Why tests missed it:** `TestLintReportsLegacyAndCanonicalDependencyDefects` explicitly asserted that both error messages were emitted simultaneously.

**Recommendation:** Suppress the generic `ProblemCycle` when `ProblemSelfDependency` is already recorded for that task ID, or filter 1-node self-cycles from the cycle problem list.

#### L3. Recursive DFS in `computeSound` creates unneeded stack frame allocation for deep graphs · **Status:** tracked by 6g3q4rt7mgjn

**File:** internal/core/dependency_graph.go:503-536 | **Component:** core / graph algorithms
**Effort:** S · **Urgency:** eventually

**[monitor / tracked follow-up]**

`OwnedDAGAnalyzer.Analyze` uses iterative Kahn's algorithm for topological sorting, but `computeSound` uses recursive DFS across prerequisites. While memoization prevents exponential re-traversal, stack depth equals graph diameter.

In an extreme artificial or generated graph with 20,000 linear chain tasks, `computeSound` will allocate 20,000 stack frames, placing unnecessary load on goroutine stack allocation.

**Why tests missed it:** Deep chain tests evaluated `Analyze` at 2048 nodes, but `TestTaskGraphSoundCompletionMemoizesReconvergentDiamonds` only tested diamond chains up to depth 120.

**Recommendation:** Replace recursive DFS in `computeSound` with an iterative post-order traversal or wave-based dynamic programming evaluation in a future optimization pass if planning graph diameters grow significantly.

---

## Assumptions That Survived Adversarial Review

These claims were independently checked, tested with counterexamples, and found sound:

1. **`depends_on` Set Semantics and Stable Serialization:**
   - Tasks store `DependsOn` as a stable task ID slice (`domain/task.go:41`).
   - `CreateTask` sorts IDs stably in YAML (`store/create.go:83-87`).
   - Domain `Task` reads raw values to preserve diagnostic fidelity for lint.
   - Wire contract (`TaskJSON.DependsOn`) sorts IDs and omits when empty (`wire/dto.go:41, 51`).
2. **Reconvergent Diamond Memoization & O(V+E) Complexity:**
   - Verified with visit-count assertions: `TestTaskGraphSoundCompletionMemoizesReconvergentDiamonds` proves that for a 120-layer reconvergent diamond (361 tasks, 480 edges), every task is computed exactly once in `computeSound` (`soundVisits == 1` for all nodes).
   - `Blockers` and `Downstream` are memoized behind mutexes and return deep copies (`dependency_graph.go:647, 730`).
3. **Hard Gating on Generic Mutators (`task set` and `task edit`):**
   - Verified: `task set` and `task set --force` strictly reject `depends_on`, `blocked_by`, `dependencies`, and `blocks` at both core (`Service.SetFields`) and store (`FS.SetFields`) layers.
   - Verified: `task edit` rejects any dependency delta (`depends_on` or legacy fields), while correctly permitting formatting / order-only adjustments and non-graph frontmatter edits (`store/edit.go:158-177`).
4. **Precedence Rules and Blocker Explanations:**
   - Broken precedence over blocked is strictly honored (`dependency_graph.go:565, 101`).
   - `domain.StatusDeferred` (parked) is correctly evaluated as `GateBlocked` / `BlockerParked` (not broken).
   - `domain.StatusDeprecated` (withdrawn) is correctly evaluated as `GateBroken` / `BlockerWithdrawn`.
   - `unsound-completed` accurately identifies completed tasks whose prerequisites are incomplete.
   - Shortest path BFS in `Blockers` produces deterministic, shortest explanatory paths with lexicographic tie-breaking (`dependency_graph.go:658-698`).
5. **Reopen Invalidation Semantics:**
   - Changing an upstream completed task to `ready-to-start` immediately causes downstream completed tasks to derive `SoundlyCompleted: false`, `Drained: false`, `Gate: GateBlocked`, `Inconsistent: true` without rewriting frontmatter on disk (`dependency_graph_test.go:291-304`).
6. **Legacy Field Direction and Slug Resolution:**
   - Inversion of `blocks` (`From: this_task, To: candidate`) vs `blocked_by` (`From: candidate, To: this_task`) is mathematically correct (`dependency_graph.go:484-488`).
   - Exact ID resolution takes precedence over slug resolution.
   - Ambiguous slugs (>1 task matching slug) and missing slugs/IDs are correctly flagged as `ProblemLegacyAmbiguous` / `ProblemLegacyMissing` and cause `GraphBroken` (`dependency_graph.go:477-494`).
7. **Thread Safety and Query Immutability:**
   - All slice and map returns from `TaskGraph` (`Blockers`, `Downstream`, `TaskIDs`, `TopologicalWaves`, `Problems`, `LegacyDiagnostics`) return cloned copies, preventing caller mutation from corrupting graph caches. Mutex locks protect memoized caches.
8. **Wire Version Bump and Backward Compatibility:**
   - `SchemaVersion` bumped to `"1.49"`.
   - `depends_on` field in JSON DTO is `omitempty` and sorted.
   - No breaking changes to existing fields or JSON envelopes.
9. **Multi-Repository Planning Space Compatibility:**
   - Tasks reference other tasks by their 12-character stable task ID within the planning space. The graph engine operates purely on task IDs and statuses within the planning repository, regardless of how many implementation checkouts are coordinated.

---

## Sequencing & Architectural Critique

1. **Store vs Core Separation:** The analysis interface correctly lives in `internal/core/dependency_graph.go` as pure in-memory algorithms over `domain.Task` records. It does not touch the filesystem or hold locks.
2. **Reentrancy Preparation for Slice 2:** `Service.ReadTaskGraph()` provides the strict snapshot factory. In slice 2 (`6g3q4rt0wzkq`), `WithGraphMutation` in store will take the lock, obtain `ReadTaskGraph()`, invoke the pure planner, apply writes, and release. This split cleanly prepares for slice 2 without coupling core to filesystem locking.
3. **Migration Boundary:** Slice 1 correctly scopes migration to *diagnosis only*. However, coupling this diagnosis to fatal exit 11 in `lint` (M1) before the mutation command exists causes an operational bind.

---

## Traceability Table

| Finding | Severity | Classification | Action / Target Destination |
|---|---|---|---|
| **H1** (Intersecting cycle detection) | High | Fixed | `6g3q4rst78qy`: Tarjan SCC membership and representative-cycle tests |
| **M1** (Fatal lint on degraded legacy fields) | Medium | Fixed | safe legacy references are advisories with exit zero; unsafe projections fail |
| **M2** (Unreadable task identity loss) | Medium | Fixed | recover filename ID and expose the `unreadable` blocker reason |
| **L1** (Minimal DAG contract suite) | Low | Fixed | remove the one-implementation seam and retain direct adversarial graph tests |
| **L2** (Duplicate self-dependency diagnostics) | Low | Fixed | self-edge emits one specific diagnostic, not a second generic cycle issue |
| **L3** (Recursive DFS in `computeSound`) | Low | Tracked | depth-envelope criterion on `6g3q4rt7mgjn` |

---

## Validation Commands and Results

All checks executed in worktree `/Users/andyeschbacher/git/andy-esch/taskflow-canonical-task-dependency-reads` on branch `feat/canonical-task-dependency-reads`:

1. **Full Unit Test Suite:**
   ```bash
   go test ./...
   ```
   *Result:* Passed (25 test packages ok, ~21s total execution time).

2. **Race Detection Test Suite:**
   ```bash
   go test -race ./...
   ```
   *Result:* Passed (all test packages passed with zero race conditions detected).

3. **CLI Lint Check against Live Planning Data:**
   ```bash
   go run ./cmd/tskflwctl lint --json
   ```
   *Result:* Exit code 11 (validation failed). Emitted 6 legacy `blocked_by` issues on live tasks:
   - `color-and-design-overhaul-one-coherent-palette-across-every-surface`
   - `theme-config-table-and-selection-plumbing`
   - `route-the-interactive-picker-theme-through-the-palette`
   - `route-tui-chrome-through-the-palette`
   - `theme-discovery-commands-glamour-polish-and-a-second-theme`
   - `route-progress-bars-and-the-cli-ansi-map-through-the-palette`

4. **Schema Wire Output:**
   ```bash
   go run ./cmd/tskflwctl schema --json
   ```
   *Result:* Passed. Correctly reports `schema_version: "1.49"` and includes `depends_on` in `task_fields`.

5. **Audit Lint Validation:**
   ```bash
   tskflwctl audit lint 2026-08-26-canonical-task-dependency-read-foundation
   ```
   *Result:* Passed.

## Remediation disposition (2026-08-27)

The current slice absorbed H1, M1, M2, L1, and L2. L3 remains a measured optimization trigger,
tracked by the explicit deep-chain envelope criterion on `6g3q4rt7mgjn`; it is not treated as a
present correctness failure because the audit's 100,000-node counterexample completed correctly.
Final post-remediation validation is recorded in task `6g3q4rst78qy`.
