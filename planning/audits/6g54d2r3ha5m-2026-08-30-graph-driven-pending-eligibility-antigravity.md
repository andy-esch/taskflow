---
schema: 1
id: 6g54d2r3ha5m
bucket: closed
area: graph-driven-pending-eligibility-antigravity
date: "2026-08-30"
updated_at: "2026-08-30"
---

# Audit: Graph-Driven Pending Eligibility for Queued and Ready Tasks — 2026-08-30

> Edit findings in place and flip each `**Status:**` as you work it.

Adversarial implementation review of the graph-driven pending eligibility slice (task `6g5075cga2nt`, epic `30-threads-and-task-dependency-graphs`, branch `feat/graph-driven-pending-eligibility`, uncommitted tree over merge-base `bb07221`), evaluated against ADR-0006, `docs/ARCHITECTURE.md`, and the task specification.

**Executive Verdict: Ready.** The implementation establishes graph-derived dependency eligibility across both pending work roles (`next-up` / `RoleQueued` and `ready-to-start` / `RoleCandidate`) without collapsing author-declared workflow metadata or introducing authorization bypasses. Centralized policy helper `isPendingWorkRole` unifies derived state computation in `internal/core/dependency_graph.go` (`Eligible = g.health == GraphHealthy && isPendingWorkRole(role) && gate == GateClear`), lifecycle transition authorization in `internal/core/task_lifecycle.go`, structured refusal formatting in `TaskEligibilityError`, and wire schema representations in `internal/wire/dependency.go`. The status/gate matrix strictly enforces force-scope separation: `--force` on `task start` bypasses only ordinary `GateBlocked` states for pending tasks, while locally broken evidence (`GateBroken`) and invalid lifecycle roles fail closed with non-overrideable diagnostic remedies. `thread frontier` and `task list --unblocked` inherit the identical pure derivation and fail closed on unsound repository graphs.

---

## Findings

### Low

#### L1. Guarded create-and-start validation retains candidate-only status precondition · **Status:** wontfix

**File:** `internal/core/task_lifecycle.go:259-261`, `internal/core/task_lifecycle.go:283` | **Component:** core / lifecycle creation
**Effort:** XS · **Urgency:** eventually

**[monitor]**

**Description:**
While existing-task transitions into `in-progress` (`validateExistingLifecyclePlan`) admit any pending work role (`isPendingWorkRole`), `validateCreateAndStartPlan` enforces that the prospective `TaskLifecycleCreation` document begins specifically as `task.Status == domain.StatusReadyToStart`. In CLI usage, `task new --start` constructs a candidate scaffold and transitions it directly to `in-progress` while enforcing `cmd.MarkFlagsMutuallyExclusive("next", "start")`.

**Impact:**
No user-facing CLI or TUI defect exists. At the internal API level, prospective single-step create-and-start operations require a candidate scaffold rather than accepting a queued scaffold.

---

**Resolution:** Retained intentionally: create-and-start creates a fresh task directly in
in-progress. Scaffolding it conceptually as a ready-to-start candidate before immediate
materialization is sound, and CLI flags enforce mutual exclusivity between --next and --start.

#### L2. In-flight tasks are excluded from eligibility and frontier read projections by design · **Status:** wontfix

**File:** `internal/core/dependency_graph.go:735-738`, `internal/core/service_task.go:80`, `internal/core/thread_projection.go:177-183` | **Component:** core / projection rollup
**Effort:** XS · **Urgency:** eventually

**[monitor]**

**Description:**
`TaskGraphState.Eligible` requires `isPendingWorkRole(role) && gate == GateClear`. A task currently in `in-progress` (`RoleInFlight`) evaluates to `Eligible: false`. Consequently, `thread frontier` and `task list --unblocked` emit only pending work waiting to be started, omitting active tasks already underway.

**Impact:**
Conforms precisely to ADR-0006 §3 ("In-flight, parked, nominally complete, and withdrawn tasks are not work waiting to start, so they stay outside the frontier"). In-progress work is surfaced via ordinary status filters and Thread member lists.

---

**Resolution:** Retained intentionally: eligibility and frontiers represent dispatchable
pending work waiting to start. Tasks already in progress are monitored through active
lists and Thread member rollups.

#### L3. Batch start commands commit incrementally per task without multi-item rollback · **Status:** wontfix

**File:** `internal/cli/task.go:809-818`, `internal/cli/transition_test.go:119-150` | **Component:** cli / batch execution UX
**Effort:** XS · **Urgency:** eventually

**[monitor]**

**Description:**
When multiple task slugs are passed to `task start task-a task-b task-c`, `runMoves` executes each task transition as an independent guarded `MutateTaskLifecycle` transaction under the repository write lock. If `task-a` is clear and `task-b` is blocked without `--force`, `task-a` commits to `in-progress`, `task-b` records a typed `TaskEligibilityFailureJSON` entry in `MovesEnvelope`, and the command exits with code 11.

**Impact:**
Consistent with Taskflow's standard batch operation semantics across all lifecycle commands. There is no risk of file corruption or half-written records, but multi-item CLI invocations commit per item.

---

**Resolution:** Per-item atomic commit semantics are the established Taskflow batch contract:
every row is attempted and reported, prior durable successes are not rolled back, and
reruns converge.

---

## Traceability Table

| Acceptance Criterion | Status | Implementation Seam | Test Coverage |
| :--- | :---: | :--- | :--- |
| **1. Shared eligibility derivation**<br>`Eligible` is true for `next-up` and `ready-to-start` only when the graph is healthy and dependency gate is clear; false for blocked/broken gates and all other roles. | **Fulfilled** | `internal/core/dependency_graph.go:107-113`<br>[`isPendingWorkRole`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L107-L113), `internal/core/dependency_graph.go:735`<br>[`deriveState`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph.go#L725-L739) | `internal/core/dependency_graph_test.go:150-185`<br>`TestTaskGraphLifecycleRoleAndGateMatrix` |
| **2. Universal start authorization & force scope**<br>`task start` and all paths into `in-progress` accept either pending role; `--force` bypasses only `blocked` gates and never bypasses broken evidence or invalid roles. | **Fulfilled** | `internal/core/task_lifecycle.go:311-322`<br>[`validateExistingLifecyclePlan`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/task_lifecycle.go#L286-L352), `internal/core/task_lifecycle.go:157-159`<br>[`DependencyGateOverrideAllowed`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/task_lifecycle.go#L157-L159) | `internal/core/task_lifecycle_test.go:23-95`<br>`TestValidateTaskLifecycleStartEligibilityAndTypedOverride`, `internal/cli/transition_test.go:85-117` |
| **3. Aligned Thread frontier & unblocked selector**<br>Thread frontier and `task list --unblocked` deterministically include clear-gated `next-up` and `ready-to-start` tasks, retain role labels, and fail closed on unsound graphs. | **Fulfilled** | `internal/core/thread_projection.go:177-183`<br>[`ProjectThread`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/thread_projection.go#L84-L218), `internal/core/service_task.go:48-54,80-82`<br>[`Service.ListTasks`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L36-L86) | `internal/core/thread_projection_test.go:43-61`<br>`TestProjectThreadFrontierUsesSharedEligibility`, `internal/cli/task_dependency_test.go:308-317,357-365` |
| **4. Readiness as workflow metadata**<br>`ready-to-start` remains an optional handoff signal; `task ready` is supported but no longer required before guarded execution. | **Fulfilled** | `internal/core/task_lifecycle.go:305-322`<br>Lifecycle transitions, `internal/cli/task.go:779-786` | `internal/cli/transition_test.go:85-117`<br>`TestTaskStartAcceptsQueuedWorkWithClearOrForcedBlockedGate` |
| **5. Wire contracts, schema & documentation**<br>Wire/schema comments, machine schema 1.55, CLI help text, generated reference docs, and golden tests describe corrected semantics without changing dependency ownership. | **Fulfilled** | `internal/wire/wire.go:182-185,238`<br>[`SchemaVersion 1.55`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/wire/wire.go#L238), `internal/wire/dependency.go:28,269`, `internal/wire/thread.go:82` | `internal/wire/envelopes_test.go:23-145`<br>`TestJSONSchema_ValidatesRealOutput`, `internal/cli/golden_test.go:20-65` |
| **6. Verification suite & diff hygiene**<br>Full race-enabled test suite, golangci-lint, go vet, CLI-doc sync, planning lint, audit lint, and diff hygiene pass cleanly. | **Fulfilled** | Repository test suites, linters, and generated documentation | Race suite (0 data races), golangci-lint (0 issues), git diff --check (clean) |

---

## Detailed Review by Hostile Angles

### 1. Domain Semantics

- **Pending Work Roles:** `isPendingWorkRole` recognizes both `RoleQueued` (`next-up`) and `RoleCandidate` (`ready-to-start`) as active pending work. Persisted lifecycle statuses remain distinct in frontmatter and query views, preserving author-declared readiness as descriptive workflow metadata.
- **Pure State Derivation:** `TaskGraphState.Eligible` is derived purely from graph structure and snapshot health:
  $$\text{Eligible} = (\text{Health} == \text{GraphHealthy}) \land \text{isPendingWorkRole}(\text{Role}) \land (\text{Gate} == \text{GateClear})$$
- **Independent Graph Health and Gate State:**
  - Graph health (`healthy`, `degraded`, `broken`) evaluates whole-snapshot integrity (unreadable files, ID drift, duplicate IDs, missing/cyclic dependencies, legacy fields).
  - Gate state (`clear`, `blocked`, `broken`) evaluates prerequisite sound completion along the task's upstream causal chain.
  - A task with a locally clear dependency set in an unsound repository graph evaluates to `Eligible: false`, preventing unsafe dispatch from partial or corrupted graph evidence.
- **Terminal and Non-Terminal Blocker Precedence:**
  - Deferred (`RoleParked`) prerequisites evaluate to `GateBlocked` with blocker reason `BlockerParked`. Because parked tasks represent active planning work that can be unparked, `--force` is permitted to override this gate.
  - Deprecated (`RoleWithdrawn`) prerequisites evaluate to `GateBroken` with blocker reason `BlockerWithdrawn`. Because withdrawn tasks will never complete, `--force` is strictly forbidden from bypassing this gate.
  - Missing, unreadable, invalid-status, and cyclic prerequisites evaluate to `GateBroken` and cannot be bypassed.
- **Reopen & Unsound Completion Propagation:** When an upstream task is reopened (e.g. `completed` $\rightarrow$ `in-progress`), all downstream completed tasks retain their persisted status but transition derived state to `Inconsistent: true`, `Gate: GateBlocked`, and `SoundlyCompleted: false`. Descendant tasks waiting on them transition to `GateBlocked` (`BlockerUnsoundCompleted`).

### 2. Mutation Authorization and Safety

- **Universal Start Path Routing:**
  - `task start <task> [--force]` $\rightarrow$ `Service.Move` $\rightarrow$ `FS.MutateTaskLifecycle`.
  - `task move <task> in-progress [--force]` $\rightarrow$ `Service.Move` $\rightarrow$ `FS.MutateTaskLifecycle`.
  - `task new <title> --start` $\rightarrow$ `Service.NewTask` $\rightarrow$ `FS.MutateTaskLifecycle` (`materializeCreateAndStart`).
  - TUI transitions $\rightarrow$ `moveTask` $\rightarrow$ `Service.Move` with `TaskLifecycleOverrideNone`.
  - Batch start invocations $\rightarrow$ `runMoves` executes each task through `Service.Move` under exclusive repository locking.
- **Guarded Path Integrity & Non-Bypassable Gates:**
  - `task edit` and `store.EditTask` compare pre/post-edit statuses and reject changes with `domain.ErrValidation`.
  - `task set` and `store.SetTaskFields` reject `status` and all graph-owned dependency fields with `domain.ErrValidation`.
  - `ReplaceBody` and `AppendBody` surgically modify markdown bodies while preserving frontmatter verbatim.
  - `task depend add/remove` modifies only dependency edges via `TaskGraphMutationStore`.
  - `lint --fix` skips lifecycle-owned and graph-owned fields.
- **Force-Scope Boundary:**
  - `--force` on `task start` and `task move ... in-progress` passes `TaskLifecycleOverrideDependencyGate`.
  - `validateExistingLifecyclePlan` checks:
    ```go
    if !isPendingWorkRole(before.Role) {
        return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, eligibilityError(graph, plan.TaskID, plan.Override)
    }
    if before.Gate != GateClear {
        analysis.OutstandingBlockers = graph.BlockingFrontier(plan.TaskID)
        if before.Gate != GateBlocked || plan.Override != TaskLifecycleOverrideDependencyGate {
            return TaskLifecyclePlan{}, TaskLifecycleAnalysis{}, eligibilityError(graph, plan.TaskID, plan.Override)
        }
        analysis.OverrideApplied = true
    }
    ```
  - This check guarantees that `--force` is accepted *only* when `before.Gate == GateBlocked` and `isPendingWorkRole(before.Role) == true`. If `before.Gate == GateBroken`, `--force` is rejected, and `DependencyGateOverrideAllowed()` reports `false`.
- **Double CAS and Concurrency Safety:**
  - `checkedWriteLock()` acquires root flock and mutex before reading graph state.
  - Pre-write whole-snapshot CAS (`graph.SameSourceSnapshot`) verifies that no prerequisite or concurrent task changed during planning.
  - Immediate target CAS (`verifyUnchanged`) re-resolves and hashes the target file immediately before atomic file replacement.

### 3. Read Projections

- **Projection Parity:**
  - `task list --unblocked` filters tasks with `graph.State(t.ID).Eligible == true`. Fails closed if `graph.Health() != GraphHealthy`.
  - `thread frontier <thread>` projects member tasks with `member.State.Eligible == true`. Fails closed (empty frontier `[]`) if `view.ProjectionHealth != GraphHealthy`.
  - `task blockers <task>` emits `State.Eligible`, `Gate`, and `Frontier` / `Causal` blocker lists with shortest paths.
  - `task unblocks <task>` emits `State` and downstream dependent impacts with before/after states.
- **Role Label Preservation:** Both `thread frontier` and `task list --unblocked` output the exact persisted role (`queued` for `next-up`, `candidate` for `ready-to-start`) rather than homogenizing the display.
- **Projection vs. Mutation Agreement:** No projection can report `Eligible: true` when `task start` would reject, or report `Eligible: false` when `task start` would accept (barring concurrent external modifications detected by CAS).

### 4. API and Compatibility Contracts

- **Wire Schema Bump:** Wire schema version bumped from `1.54` to `1.55` in `internal/wire/wire.go`.
- **JSON Schema Descriptions:** `TaskGraphStateJSON.Eligible` and `ThreadViewJSON.Frontier` carry explicit schema descriptions explaining that both `next-up` and `ready-to-start` are eligible when gates are clear in healthy graphs.
- **Refusal Payload Ownership:** `ToTaskEligibilityFailureJSON` delegates `OverrideAllowed` calculation directly to core's `failure.DependencyGateOverrideAllowed()`, preventing adapter-side drift.
- **Golden Fixtures & Generated Documentation:** All 26 JSON test fixtures, CLI markdown reference pages (`tskflwctl_task_start.md`, `tskflwctl_task_list.md`, `tskflwctl_thread_frontier.md`), and `README.md` are synchronized with the 1.55 contract.

### 5. Tests and Documentation

- **Matrix Coverage:** `TestValidateTaskLifecycleStartEligibilityAndTypedOverride` crosses `next-up` and `ready-to-start` sources against `clear`, `blocked` (not-started, in-flight, unsound-completed, parked), and `broken` (withdrawn, missing, cycle) gates.
- **Negative & Rejection Proofs:**
  - `TestValidateTaskLifecycleForceCannotBypassRoleOrBrokenRepository` proves that parked, completed, and withdrawn tasks cannot be started with `--force`.
  - `TestTaskListUnblockedFailsClosedOnBrokenGraph` proves that `--unblocked` exits 11 on corrupted graphs.
  - `TestTaskStartMixedBatchRetainsSuccessAndTypedFailure` proves that batch start reports successes and typed refusal rows without partial corruption.
- **TUI & CLI Integration:** `TestTUITaskStartUsesDependencyEligibilityPolicy` and `TestTaskStartAcceptsQueuedWorkWithClearOrForcedBlockedGate` prove end-to-end start execution for queued work across both interactive and CLI interfaces.

### 6. Future Resilience

- **Guarded Thread Mutations:** By centralizing eligibility on `isPendingWorkRole` and graph-derived gates, upcoming Thread mutation verbs (`thread start`, `thread complete`, `thread add/remove`) consume an invariant-backed foundation without needing custom eligibility branches.
- **Compound Bulk Operations:** The lock-free internal materialization seams (`materializeTaskLifecycle`, `materializeThreadCreation`) can be safely composed under compound guards during bulk apply operations.
- **Extensible Pending Roles:** If additional pending work statuses are introduced in future ADRs, adding them to `isPendingWorkRole` will immediately propagate consistent authorization, frontier filtering, and refusal explanations across all adapters.

---

## Explicit Rejected Concerns

1. **Weakening Start Safety by Allowing `next-up` to Start:**
   - *Concern:* Allowing queued (`next-up`) tasks to start directly might bypass intentional scoping checks.
   - *Finding:* In accordance with ADR-0006 §3, readiness is workflow metadata, not an authorization gate. External dependency clearance is the sole causal constraint. Starting a queued task requires all prerequisites to be soundly completed, preserving full causal safety.
2. **`--force` Bypassing Broken Dependencies:**
   - *Concern:* An operator using `task start --force` might bypass a broken or withdrawn prerequisite.
   - *Finding:* `validateExistingLifecyclePlan` enforces `before.Gate == GateBlocked`. If a prerequisite is deprecated, missing, or cyclic, the gate is `GateBroken`, and `--force` is rejected with `TaskEligibilityError` and `OverrideAllowed: false`.
3. **Discrepancy Between Thread Frontier and Task List `--unblocked`:**
   - *Concern:* Thread frontier might use a different filtering predicate than `task list --unblocked`.
   - *Finding:* Both projections evaluate the identical `TaskGraphState.Eligible` field derived by `internal/core/dependency_graph.go`.
4. **Stale Graph State in TUI Transitions:**
   - *Concern:* TUI task start might operate on stale cached graph data.
   - *Finding:* TUI delegates to `core.Service.Move`, which executes under `checkedWriteLock()` and verifies whole-snapshot and target CAS tokens before committing.

---

## Validation Commands and Results

```bash
go test ./...
# Result: ok across all 25 packages (0 failures)

go test -race ./...
# Result: ok across all packages (0 data races)

golangci-lint run ./...
# Result: 0 issues

git diff --check
# Result: clean diff hygiene

go run ./cmd/tskflwctl lint
# Result: ✔ all planning entities and dependency links pass lint

go run ./cmd/tskflwctl audit lint
# Result: ✔ all audit findings pass lint
```
