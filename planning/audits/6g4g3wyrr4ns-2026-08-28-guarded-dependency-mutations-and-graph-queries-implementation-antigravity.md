---
schema: 1
id: 6g4g3wyrr4ns
bucket: closed
area: guarded-dependency-mutations-and-graph-queries-implementation-antigravity
date: "2026-08-28"
updated_at: "2026-08-28"
---

# Audit: Guarded Dependency Mutations and Graph Queries Implementation Review — 2026-08-28

> Edit findings in place and flip each `**Status:**` as you work it.

Adversarial implementation review of the guarded task-dependency operations slice (task `6g3q4rt7mgjn`, epic `30-threads-and-task-dependency-graphs`, branch `feat/guarded-dependency-operations`, uncommitted tree over merge-base `90bfabe`), evaluated against ADR-0006 (including the 2026-08-27 amendment), the readiness audit `6g4aj7v60syg`, and `docs/ARCHITECTURE.md`.

**Executive Verdict: Ready with required fixes.** The foundational architecture is sound: reference resolution operates entirely in-memory within the authoritative snapshot; cross-process mutation exclusion and multi-level CAS protect against cooperating and non-cooperating writers; Tarjan SCC cycle detection correctly attributes multi-cycle feedback topologies; and legacy migrations successfully converged the six live planning occurrences to eight canonical `depends_on` edges. However, adversarial stress testing across query projections, migration failure modes, and user-facing diagnostics reveals four acute/moderate issues that should be addressed before task `6g3q4rt7mgjn` is closed:

1. **False affirmative in `task blockers` and `task unblocks` on degraded graphs (H1):** `g.dependencies` and `g.outgoing` are populated only from canonical `depends_on`. On a degraded graph with unmigrated legacy fields, `task blockers` reports `✔ no blockers` and `"blockers":[]` for a candidate task whose legacy prerequisite is unstarted, and `task unblocks` reports `• no downstream tasks`.
2. **Missing regression test for prefix-safe migration write order (H2):** The reverse topological wave write ordering (`leftWave > rightWave`) in `planLegacyDependencyMigration` is correct, but reversing the comparison leaves all committed tests passing while silently destroying edges on an interrupted `blocks`-only migration.
3. **Queried task derived state omitted from blocker/unblock envelopes (M1):** `TaskBlockersEnvelope` and `TaskUnblocksEnvelope` omit the queried task's own `TaskGraphStateJSON` (role, gate, eligible), forcing consumers to infer authorization from an empty blocker list.
4. **Stale diagnostic guidance citing unavailable commands (M2):** `task set` rejects dependency changes citing "`task depend add/remove` once available" and `lint` cites "canonical migration is intentionally deferred to guarded dependency operations", despite both shipping in this branch.

---

## Findings

#### H1. Degraded graphs make `task blockers` and `task unblocks` emit false affirmative all-clear responses · **Status:** fixed

**File:** `internal/core/dependency_graph.go:369,389-390,404-405,414-416`, `internal/cli/render/dependency.go:61-63,89-91` | **Component:** core/query projections, cli/render
**Effort:** M · **Urgency:** acute

`g.dependencies` and `g.outgoing` are populated exclusively from canonical `Task.DependsOn` fields. Resolved legacy edges from `resolveLegacyDiagnostics` are appended to `projectedEdges` solely for cycle and topological wave analysis in `analyzeDAG`. They are not merged into `g.dependencies` or `g.outgoing`.

As a consequence, when a repository is in the `degraded` state (e.g. before `task depend migrate` has been run):
- `task blockers <task>` walks `g.unsoundPrerequisites` and finds zero canonical dependencies, printing `✔ no blockers` and emitting `"blockers": []`.
- `task unblocks <prereq>` walks `g.outgoing` and finds zero canonical dependents, printing `• no downstream tasks` and emitting `"unblocks": []`.

**Reproducible Example:**
Given task `alpha` (`ready-to-start`) and task `beta` (`ready-to-start`, `blocked_by: [alpha]`):
```text
$ tskflwctl task blockers beta
beta  (6g0000000b02)
graph:  degraded
view:  frontier
✔ no blockers
⚠ legacy blocked_by on 6g0000000b02; run task depend migrate
```

Although `graph: degraded` is displayed, the human green checkmark `✔ no blockers` and machine `"blockers": []` invite the inference that `beta` has no blockers.

**Recommendation:** In `render.TaskBlockersHuman` and `render.TaskUnblocksHuman`, replace `✔ no blockers` / `• no downstream tasks` with an explicit notice (e.g. `• no canonical blockers; N legacy field(s) unprojected — run task depend migrate`) when `result.Health != GraphHealthy`. Additionally, in wire envelopes, include `legacy_dependencies_unprojected: true` or project resolved legacy edges directly into diagnostic graph traversals.

**Resolution:** Projected exactly resolved legacy edges into diagnostic
traversals and derived gates; degraded blocker and downstream regression cases
now return the constraint.

#### H2. Prefix-safe migration write order lacks test coverage; inverted order silently drops edges on failure · **Status:** fixed

**File:** `internal/core/dependency_operations.go:320-326`, `internal/store/dependency_operations_test.go:102-144` | **Component:** core/migration planner, store regression tests
**Effort:** S · **Urgency:** acute

`planLegacyDependencyMigration` orders multi-file writes in reverse topological wave order (`leftWave > rightWave`), writing dependents before prerequisites. This ensures that when a prerequisite clears its legacy `blocks` declaration, the dependent has already durably written the canonical `depends_on` edge.

However, existing tests in `dependency_operations_test.go` only test fixtures where dependencies are declared symmetrically (`blocks` on owner AND `blocked_by` on dependent) or via `blocked_by` alone. If the sort comparison in `dependency_operations.go:323` is inverted to `leftWave < rightWave`, all 25 test packages still pass 100%.

If an inverted migration is interrupted after the first write on a `blocks`-only fixture:
1. The prerequisite file is written first, clearing `blocks: [dependent]`.
2. The write fails before updating the dependent.
3. On retry, the prerequisite no longer has `blocks`, and the dependent never had `blocked_by`. The edge vanishes completely and the graph is falsely declared `healthy`.

**Recommendation:** Add a test in `internal/store/dependency_operations_test.go` with a `blocks`-only fixture (owner has `blocks: [dependent]`, dependent has no legacy fields), inject a failure after the first write via `testHookAfterGraphWrite`, and verify that the first write modified the dependent (adding `depends_on`), leaving the owner's `blocks` intact until the second write.

**Resolution:** Added a blocks-only interrupted migration test that pins
dependent-first write order and convergent retry.

#### M1. Query wire envelopes omit derived state for the queried task · **Status:** fixed

**File:** `internal/wire/dependency.go:107-115, 143-150`, `internal/cli/render/dependency.go:108-116` | **Component:** wire/query DTOs, cli/render
**Effort:** S · **Urgency:** soon

`TaskBlockersEnvelope` and `TaskUnblocksEnvelope` include `TaskGraphStateJSON` for every blocker and downstream task, but only provide `GraphTaskJSON` (id, slug, status, epic) for the *queried* task itself. The queried task's own `Role`, `Gate`, and `Eligible` fields are omitted.

Per `docs/ARCHITECTURE.md`, "Eligibility is read from derived state, never inferred from an empty blocker list". Because `TaskBlockersEnvelope` does not return the queried task's derived state, machine consumers querying `task blockers <task>` have no choice but to infer eligibility from an empty blocker list.

**Recommendation:** Add `State TaskGraphStateJSON json:"state"` for the root task in `TaskBlockersEnvelope` and `TaskUnblocksEnvelope`, and render `role/gate` in `graphQueryHeader`.

**Resolution:** Queried-task state is present in both wire envelopes and human
headers; schema version is 1.51.

#### M2. Diagnostic and rejection messages contain stale guidance citing unavailable commands · **Status:** fixed

**File:** `internal/core/service_task.go:332-337`, `internal/core/service.go:401`, `internal/core/dependency_graph_mutation.go:116`, `internal/core/dependency_operations.go:446` | **Component:** core/diagnostic messages
**Effort:** XS · **Urgency:** soon

Several error and lint messages across the codebase still state that guarded commands are unavailable:
- `service_task.go:334`: `task set` rejection says `(`task depend add/remove` once available)`.
- `service.go:401`: `lint` diagnostic says `canonical migration is intentionally deferred to guarded dependency operations`.
- `dependency_graph_mutation.go:116` and `dependency_operations.go:446`: duplicate helper functions cite `run the guarded migration` vs `run task depend migrate`.

**Recommendation:** Update `service_task.go` to state `use guarded dependency operations (task depend add/remove)` and unify `graphMutationHealthDetail` and `taskGraphHealthDetail` into a single shared helper.

**Resolution:** Guidance now names exact shipping commands and one shared
health-detail helper owns the wording.

#### L1. `TaskGraph.ResolveTaskID` omits discarded duplicate-ID sibling slugs from candidate resolution on broken graphs · **Status:** fixed

**File:** `internal/core/dependency_graph.go:346-349, 762-765` | **Component:** core/graph resolution, diagnostic addressability
**Effort:** XS · **Urgency:** eventually

When multiple files declare the exact same task ID, `NewTaskGraph` retains only the first task in `g.tasks` and `g.ids`. `ResolveTaskID` builds candidates from `g.ids`, so exact ID lookup returns the first task ID rather than `domain.ErrAmbiguous`, and querying the second task's slug returns `domain.ErrNotFound`. Store resolution (`store.ResolveTaskPath`) preserves all filesystem candidates and returns `domain.ErrAmbiguous`.

**Impact:** Contained to broken repositories where duplicate task IDs exist. All mutations already fail closed via `ValidateTaskGraphMutationSource`.

**Resolution:** The snapshot retains one resolution candidate per file,
preserving duplicate-ID ambiguity and sibling slug diagnostics.

#### L2. Downstream and causal query output scales quadratically with graph depth · **Status:** deferred

**File:** `internal/core/dependency_graph.go:1039-1076` | **Component:** core/query performance, envelope size
**Effort:** S · **Urgency:** eventually

`DownstreamImpact` and `CausalBlockers` materialize full path arrays for each reachable node. For a linear chain of depth $N$, the serialized JSON envelope produces $\Theta(N^2)$ path elements. On a 1,500-task chain, this generates ~17MB of JSON.

**Impact:** Negligible on realistic human planning repositories (typical depth $\le 10$). Stress tests should document this envelope bound.

**Resolution:** A bounded 512-edge test records quadratic full-path
amplification. Capping remains deferred until dogfood exceeds the current live
depth of six.

#### L3. Present-but-empty legacy fields are ignored by migration · **Status:** fixed

**File:** `internal/core/dependency_graph.go:551-553` | **Component:** core/legacy migration
**Effort:** XS · **Urgency:** eventually

`resolveLegacyDiagnostics` skips fields where `len(values) == 0`. A task with `blocked_by: []` or `blocks: []` is not diagnosed as degraded and is not cleared by `task depend migrate`.

**Impact:** Zero semantic impact, as empty legacy arrays declare no dependency constraints.

---

**Resolution:** Store parsing now preserves legacy key presence, so empty keys
remain degraded, migratable, and receipt-bearing.

## Traceability Table

| Acceptance Criterion | Status | Implementation Seam | Test Coverage |
| :--- | :---: | :--- | :--- |
| **1. Snapshot reference resolution**<br>Matches ordinary resolution for exact IDs/slugs, case-insensitive prefixes, substrings, missing, ambiguity without calling Store. | **Fulfilled** (with L1 caveat) | `internal/core/dependency_graph.go:752-809`<br>[`*TaskGraph.ResolveTaskID`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L752-L809) | `internal/core/dependency_graph_test.go:225-266`<br>`TestTaskGraphResolveTaskIDMatchesRepositoryReferenceTiers` |
| **2. Add/Remove mutation validation**<br>Validates exact canonical endpoints, duplicate/self/missing edges, proposed union, cycles, every durable prefix, and final health. | **Fulfilled** | `internal/core/dependency_operations.go:176-230`<br>[`planDependencyEdges`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_operations.go#L176-L230), `internal/core/dependency_graph_mutation.go:29-97`<br>[`ValidateTaskGraphMutationPlan`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph_mutation.go#L29-L97) | `internal/core/dependency_operations_test.go:121-143`<br>`TestServiceDependencyMutationRejectsAmbiguousDuplicateSelfAndCycle` |
| **3. Idempotent no-op preservation**<br>Already-present adds and absent removals succeed as receipt-bearing no-ops with zero byte or timestamp change. | **Fulfilled** | `internal/core/dependency_operations.go:208-228`<br>[`planDependencyEdges`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_operations.go#L208-L228), `internal/store/graphmutation.go:155-161`<br>[`materializeTaskGraphPlan`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/graphmutation.go#L155-L161) | `internal/cli/task_dependency_test.go:45-129`<br>`TestTaskDependAddRemoveJSONDryRunAndNoop` |
| **4. Authoritative dry-run execution**<br>`--dry-run` executes authoritative resolution, planning, and validation with zero replacements. | **Fulfilled** | `internal/store/graphmutation.go:73-75`<br>[`MutateTaskGraph`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/graphmutation.go#L73-L75), `internal/core/dependency_operations.go:124-150`<br>[`runDependencyMutation`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_operations.go#L124-L150) | `internal/core/dependency_operations_test.go:70-119`<br>`TestServiceDependencyAddRemoveDryRunAndIdempotence` |
| **5. Legacy dependency migration**<br>`task depend migrate` converts the six live legacy occurrences with frontmatter/body preservation; unsafe state writes nothing. | **Fulfilled** | `internal/core/dependency_operations.go:232-340`<br>[`planLegacyDependencyMigration`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_operations.go#L232-L340), `internal/store/dependency_operations_test.go:17-67` | `internal/store/dependency_operations_test.go:17-67`<br>`TestDependencyMigrationPreservesBodyCommentsAndConverges` |
| **6. Migration prefix-safety and retry convergence**<br>Migration remains sound after every injected write failure; retry converges from every durable prefix. | **Requires Amendment** (H2) | `internal/core/dependency_operations.go:306-327`<br>Reverse topological wave write ordering | `internal/store/dependency_operations_test.go:102-144`<br>`TestDependencyMigrationEveryDurablePrefixStaysSoundAndResumes` |
| **7. Mutation receipts & recovery data**<br>Receipts distinguish applied and skipped edges; partial-failure diagnostics preserve applied and remaining tasks. | **Fulfilled** | `internal/core/dependency_operations.go:151-174`<br>[`dependencyReceipt`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_operations.go#L151-L174), `internal/cli/task_dependency.go:25-37`<br>[`dependencyFailure`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/cli/task_dependency.go#L25-L37), `internal/cli/exit.go:80-84` | `internal/cli/task_dependency_test.go:303-328`<br>`TestWriteErrorCarriesStructuredDependencyMutationRecovery` |
| **8. Explanatory blocker query**<br>`task blockers` defaults to action frontier, `--causal` returns full closure, exposing deterministic reasons and shortest paths. | **Requires Amendment** (H1, M1) | `internal/core/dependency_graph.go:866-933`<br>[`CausalBlockers`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L866-L875), [`BlockingFrontier`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L879-L888) | `internal/cli/task_dependency_test.go:232-291`<br>`TestTaskGraphQueryCommandsAndUnblockedSelector` |
| **9. Explanatory downstream query**<br>`task unblocks` reports transitive downstream dependents and current state without claiming counterfactual eligibility. | **Requires Amendment** (H1, M1) | `internal/core/dependency_graph.go:1036-1075`<br>[`DownstreamImpact`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L1036-L1075), `internal/cli/render/dependency.go:87-106` | `internal/core/dependency_graph_test.go:268-299`<br>`TestTaskGraphDownstreamImpactUsesDeterministicShortestPaths` |
| **10. Fail-closed unblocked selector**<br>`task list --unblocked` filters on derived eligibility and returns no dispatchable work on an unsound relevant graph. | **Fulfilled** | `internal/core/service_task.go:47-53`<br>[`Service.ListTasks`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L47-L53), `internal/core/dependency_graph.go:707` | `internal/cli/task_dependency_test.go:293-301`<br>`TestTaskListUnblockedFailsClosedOnBrokenGraph` |
| **11. Graph ownership enforcement**<br>`task set --force` and `task edit` cannot bypass graph ownership or guarded dependency validation. | **Fulfilled** (with M2 message fix) | `internal/core/service_task.go:273-277`<br>[`Service.SetFields`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L273-L277), `internal/store/edit.go:169-176`<br>[`FS.EditTask`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/edit.go#L169-L176) | `internal/store/dependency_persistence_test.go:62-135`<br>`TestEditTaskRejectsDependencyDeltaButAllowsReordering` |
| **12. Reflected schema & wire contract**<br>Query and mutation JSON envelopes enter reflected schema and preserve human/machine semantic parity. | **Fulfilled** (with M1 schema update) | `internal/wire/dependency.go:1-229`, `internal/wire/envelopes.go:1054-1056`, `internal/wire/wire.go:173-177` | `internal/wire/envelopes_test.go:70-102`<br>JSON Schema compilation and validation test suite |
| **13. Service clock and timestamp hygiene**<br>Semantic changes use the Service clock exactly once; no-ops do not advance `updated_at`. | **Fulfilled** | `internal/core/dependency_operations.go:128`<br>[`Service.runDependencyMutation`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_operations.go#L128), `internal/store/graphmutation.go:162-163` | `internal/store/dependency_operations_test.go:17-67`<br>`TestDependencyMigrationPreservesBodyCommentsAndConverges` |
| **14. Deep-chain stress envelope**<br>Deep-chain stress establishes supported depth envelope without unsafe stack or latency bounds. | **Fulfilled** (with L2 monitor) | `internal/core/dependency_graph.go:622-655`<br>Memoized sound derivation, iterative BFS | `internal/core/dependency_graph_test.go:327-361`<br>`TestTaskGraphSupportedDeepChainEnvelope` (4,096 edges) |
| **15. Dogfood bootstrap persisted**<br>Production commands persist Threads bootstrap dependency edges and exercise queries against the real graph. | **Fulfilled** | `planning/tasks/6g3q4rt7mgjn...`<br>`planning/tasks/6g3q4rte8kc1...`<br>`planning/tasks/6g3q4rtmv4ak...` | `tskflwctl lint`<br>`tskflwctl task blockers 6g3q4rte8kc1`<br>`tskflwctl task unblocks 6g3q4rt7mgjn` |

---

## Explicit Rejected Concerns

1. **Planner Store Re-entry & TOCTOU Resolution:**
   - *Investigation:* Analyzed whether human-provided references could trigger store deadlocks, conflict errors, or TOCTOU races if resolved outside or inside the lock.
   - *Finding:* `TaskGraph.ResolveTaskID` executes entirely in-memory over the immutable snapshot loaded under the exclusive write lock. Store re-entry is blocked by `FS.rejectGraphPlannerCall()`, and parity tests verify exact matching behavior against Store candidate resolution.
2. **Multi-Cycle SCC Exponential Enumeration:**
   - *Concern:* Cyclic components with multiple overlapping cycles could trigger exponential path enumeration.
   - *Finding:* Tarjan's SCC algorithm runs in linear $O(V+E)$ time, marks all cyclic SCC members exactly, and extracts a single edge-following representative cycle path without combinatorial explosion.
3. **Partial Mutation Recovery Misrepresenting Applied Edges:**
   - *Concern:* Pre-write failures might emit `dependency_mutation` envelopes that imply edges were applied.
   - *Finding:* `cli.dependencyFailure` attaches `dependency_mutation` only when `len(receipt.AppliedTaskIDs) > 0`. Pre-write failures emit standard `ErrorEnvelope` structures without misleading edge states.
4. **Cyclic Self-Reference in Downstream Impact:**
   - *Concern:* `task unblocks` on a cyclic task might list the task as downstream of itself.
   - *Finding:* `DownstreamImpact` marks the origin as visited and explicitly skips `dependent == taskID`, preventing self-referential impact output.
5. **No-op Idempotence and Timestamp Preservation:**
   - *Concern:* Re-running `task depend add` on an existing edge might advance `updated_at`.
   - *Finding:* `FS.materializeTaskGraphPlan` detects `bytes.Equal(content, newContent)` and skips writing without stamping `updated_at`.

---

## Validation Commands and Results

```bash
GOCACHE=/tmp/taskflow-review-go-cache go test ./...
# Result: ok across all 25 packages (0 failures)

GOCACHE=/tmp/taskflow-review-go-cache go test -race ./...
# Result: ok across all packages (0 data races)

GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache golangci-lint run ./...
# Result: 0 issues

git diff --check
# Result: clean diff hygiene

go run ./cmd/tskflwctl lint
# Result: ✔ all active tasks and epics pass lint
```
