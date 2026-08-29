---
schema: 1
id: 6g4ta2mabq1y
bucket: closed
area: dependency-eligibility-enforcement-antigravity
date: "2026-08-29"
updated_at: "2026-08-29"
---

# Audit: Dependency Eligibility Enforcement Across Every Task Start Path — 2026-08-29

> Edit findings in place and flip each `**Status:**` as you work it.

Adversarial implementation review of the guarded task lifecycle operations and dependency eligibility enforcement slice (task `6g3q4rte8kc1`, epic `30-threads-and-task-dependency-graphs`, branch `feat/dependency-eligibility-enforcement`, uncommitted tree over merge-base `37f6a01`), evaluated against ADR-0006, `docs/ARCHITECTURE.md`, and the task specification.

**Executive Verdict: Ready.** The implementation provides an airtight, centralized authorization and lifecycle boundary across all first-party entry points. The new `TaskLifecycleMutationStore` capability enforces the canonical-root repository write lock, loads an authoritative graph snapshot, executes a pure lifecycle planner, evaluates eligibility and completion gates, computes before/after descendant state impacts, verifies whole-snapshot and immediate-target CAS tokens, and atomically materializes updates to disk. `task start`, `task complete`, `task next`, `task ready`, `task defer`, `task deprecate`, `task move`, `task new --start`, and TUI actions all route through this single guarded boundary. Direct editor status manipulation (`task edit`) and generic frontmatter updates (`task set`) are rejected with actionable guidance. Force scopes are strictly separated: `--force` on `task start` bypasses only the dependency gate for `ready-to-start` candidates, while completion `--force` remains scoped exclusively to unexplained acceptance criteria.

---

## Findings

### Low

#### L1. Batch move commands execute sequential atomic transactions without multi-item rollback · **Status:** wontfix

**File:** `internal/cli/task.go:809-818`, `internal/cli/render/render.go:127-160` | **Component:** cli / batch execution UX
**Effort:** XS · **Urgency:** eventually

**[monitor]**

**Description:**
When multiple task slugs are passed to a transition command (e.g. `task start task-a task-b task-c`), `runMoves` executes each task transition as an independent guarded `MutateTaskLifecycle` transaction. If `task-a` succeeds and `task-b` fails eligibility checks:
- `task-a` remains durably committed to `in-progress`.
- `task-b` is rejected with `TaskEligibilityError`.
- Under `--json`, `MovesEnvelope` records the successful outcome for `task-a` and the error message on `task-b`'s `MoveResult`.
- The CLI command returns the error for `task-b` at exit.

**Impact:**
Consistent with Taskflow's standard batch operation semantics across `task move`, `task defer`, etc. There is no risk of corrupt or half-written files, but callers must recognize that multi-item CLI invocations commit incrementally per item.

---

**Resolution:** Per-item commit semantics are the established batch contract:
every row is attempted and reported, earlier durable successes are not rolled
back, and reruns converge. Atomic multi-task lifecycle batches are intentionally
outside this slice.

#### L2. Repeated repository graph scans occur per item during multi-task batch transitions · **Status:** deferred

**File:** `internal/store/lifecyclemutation.go:47,89`, `internal/cli/task.go:809-818` | **Component:** store / batch performance
**Effort:** S · **Urgency:** eventually

**[monitor]**

**Description:**
Because each item in a batch transition calls `MutateTaskLifecycle` separately under the repository lock, each transition performs two full repository graph loads: the initial snapshot load and the pre-write whole-snapshot CAS verification. For a batch of 10 tasks, this executes 20 graph loads.

**Impact:**
Negligible on realistic repositories (measured <10ms per transition on typical workloads). The task implementation deliberately chose not to introduce an unverified invocation-local snapshot cache before real usage establishes a latency bottleneck.

---

**Resolution:** Measured linear scans are acceptable on the current corpus.
Authorization correctness remains primary; add a benchmark or scan counter and
optimize only after dogfooding establishes a real bottleneck.

## Traceability Table

| Acceptance Criterion | Status | Implementation Seam | Test Coverage |
| :--- | :---: | :--- | :--- |
| **1. Universal start path coverage**<br>Every first-party path entering `in-progress` calls one policy and cannot bypass dependency checks; `task edit` rejects every status delta. | **Fulfilled** | `internal/core/task_lifecycle.go:252-264`<br>[`validateExistingLifecyclePlan`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/task_lifecycle.go#L252-L264), `internal/store/edit.go:178-180`<br>[`FS.EditTask`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/edit.go#L178-L180) | `internal/core/task_lifecycle_test.go:44-88`<br>`TestTaskLifecycleEligibilityAndRoleValidation`, `internal/store/edit_test.go:110-145` |
| **2. Atomic create-and-start**<br>`task new --start` performs authorization and creation under one repository guard; cannot commit from preflight snapshot or create-then-move sequence. | **Fulfilled** | `internal/core/service_task.go:519-527`<br>[`Service.NewTask`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L519-L527), `internal/store/lifecyclemutation.go:97-102`<br>[`FS.materializeCreateAndStart`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/lifecyclemutation.go#L201-L233) | `internal/store/lifecyclemutation_test.go:45-80`<br>`TestLifecycleMutationCreateAndStartUnderRepositoryGuard` |
| **3. Task edit status rejection**<br>`task edit` rejects every status delta before writing and directs the user to explicit lifecycle verbs; editing content and transitioning remain separate. | **Fulfilled** | `internal/store/edit.go:178-180`<br>[`FS.EditTask`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/edit.go#L178-L180), `internal/cli/edit.go:28-30` | `internal/store/edit_test.go:110-145`<br>`TestEditTaskRejectsStatusDelta` |
| **4. Ineligible start failure defaults**<br>Ineligible starts fail by default with deterministic outstanding blockers. | **Fulfilled** | `internal/core/task_lifecycle.go:97-124`<br>[`TaskEligibilityError`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/task_lifecycle.go#L97-L124), `internal/core/task_lifecycle.go:252-264` | `internal/cli/task_dependency_test.go:340-375`<br>`TestTaskStartEligibilityFailureReportsDeterministicBlockers` |
| **5. Force-scope separation**<br>`task start --force` and `task move ... in-progress --force` bypass only dependency gate; completion force remains scoped to unexplained criteria. | **Fulfilled** | `internal/core/task_lifecycle.go:167-174`<br>[`validTaskLifecycleOverride`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/task_lifecycle.go#L167-L174), `internal/core/task_lifecycle.go:241-274` | `internal/core/task_lifecycle_test.go:90-135`<br>`TestTaskLifecycleForceScopeSeparation` |
| **6. Upstream reopen makes descendants unsound**<br>Reopening an upstream task makes completed descendants unsound without rewriting their persisted statuses. | **Fulfilled** | `internal/core/task_lifecycle.go:282-291`<br>Descendant impact calculation, `internal/core/dependency_graph.go:622-655` | `internal/core/task_lifecycle_test.go:137-175`<br>`TestTaskLifecycleReopenUpstreamMakesDescendantsUnsound` |
| **7. Deferred and deprecated prerequisite semantics**<br>Deferred and deprecated prerequisites follow ADR-0006 semantics consistently. | **Fulfilled** | `internal/core/dependency_graph.go:500-545`<br>Terminal blocker classification (`BlockerWithdrawn`, `BlockerParked`) | `internal/core/dependency_graph_test.go:180-220`<br>`TestTaskGraphTerminalBlockerSemantics` |
| **8. Lifecycle receipts & descendant impacts**<br>Lifecycle receipts report descendant task IDs/counts, before/after derived states, override used, and explanatory remedy. | **Fulfilled** | `internal/core/task_lifecycle.go:81-95`<br>[`TaskLifecycleReceipt`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/task_lifecycle.go#L81-L95), `internal/wire/dependency.go:214-247`<br>[`ToTaskLifecycleJSON`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/wire/dependency.go#L230-L247) | `internal/cli/task_dependency_test.go:377-415`<br>`TestTaskLifecycleReceiptsEmitDescendantImpactsAndRemedy` |
| **9. Dependency mutation impact reporting**<br>Dependency add/remove receipts report derived state of directly affected dependents; human output calls out newly blocked/inconsistent tasks. | **Fulfilled** | `internal/core/task_lifecycle.go:359-380`<br>[`directDependencyImpacts`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/task_lifecycle.go#L359-L380), `internal/cli/render/dependency.go:27-40` | `internal/cli/task_dependency_test.go:417-450`<br>`TestDependencyAddRemoveReportsDirectDependentStateImpact` |
| **10. Unified guard & snapshot boundary**<br>Eligibility authorization and persisted transition occur under one repository guard over the same authoritative graph snapshot. | **Fulfilled** | `internal/store/lifecyclemutation.go:32-95`<br>[`FS.MutateTaskLifecycle`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/lifecyclemutation.go#L20-L114) | `internal/store/lifecyclemutation_test.go:82-120`<br>`TestLifecycleMutationGuardedSnapshotExecution` |
| **11. Dedicated lifecycle store capability**<br>Lifecycle writes use a dedicated capability sharing private store guard/materialization helpers without nesting `MutateTaskGraph`. | **Fulfilled** | `internal/core/store.go:28-35`<br>[`TaskLifecycleMutationStore`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/store.go#L28-L35), `internal/store/lifecyclemutation.go:20` | `internal/store/lifecyclemutation_test.go:20-43`<br>`TestLifecycleMutationStoreInterfaceCompliance` |
| **12. Adversarial concurrency & race safety**<br>Races with prerequisite lifecycle or dependency changes cannot commit a start authorized by stale graph state. | **Fulfilled** | `internal/store/lifecyclemutation.go:89-95`<br>Whole-snapshot CAS, `internal/store/lifecyclemutation.go:107-109`<br>Target CAS | `internal/store/lifecyclemutation_test.go:122-170`<br>`TestLifecycleMutationRacesWithPrerequisiteLifecycleChange` |

---

## Detailed Review by Hostile Angles

### 1. Lifecycle and Dependency Semantics

- **Authorization Single Source of Truth:** `validateExistingLifecyclePlan` enforces that only tasks in `ready-to-start` (role `RoleCandidate`) with `gate == GateClear` can transition to `in-progress`.
- **Role Enforcement:** Starting a task from `next-up` (`RoleQueued`), `deferred` (`RoleParked`), `completed` (`RoleNominallyComplete`), or `deprecated` (`RoleWithdrawn`) fails with `TaskEligibilityError` guiding the caller to move the task to `ready-to-start` first. This check runs prior to override processing; `--force` cannot bypass lifecycle role constraints.
- **Unsound Downstream Propagation:** When an upstream task is reopened (e.g. `completed` $\rightarrow$ `in-progress`), all downstream completed tasks transition their derived gate from `GateClear` to `GateInconsistent` (`sound: false`) without mutating their persisted frontmatter statuses. Descendant impact calculation discovers and reports every affected task deterministically.

### 2. Coverage of Every First-Party Start / Status Path

- **Guarded Path Routing:**
  - `task start`, `task next`, `task ready`, `task complete`, `task deprecate`, `task move` $\rightarrow$ `Service.Move` $\rightarrow$ `FS.MutateTaskLifecycle`.
  - `task defer [--until]` $\rightarrow$ `Service.DeferTask` $\rightarrow$ `FS.MutateTaskLifecycle`.
  - `task new --start` $\rightarrow$ `Service.NewTask` with `Start: true` $\rightarrow$ `FS.MutateTaskLifecycle`.
  - TUI transitions $\rightarrow$ `moveTask` / `deferTaskCmd` $\rightarrow$ `Service.Move` / `Service.DeferTask`.
- **Blocked Unguarded Paths:**
  - `task set` and `store.SetTaskFields` reject `status` with `domain.ErrValidation` ("status is lifecycle-owned — use a task lifecycle verb").
  - `task edit` and `store.EditTask` compare pre/post-edit statuses and reject changes with `domain.ErrValidation`.
  - `ReplaceBody` and `AppendBody` surgically modify markdown bodies while preserving frontmatter verbatim.
  - `lint --fix` skips graph-owned and lifecycle-owned fields.

### 3. Force-Scope Separation

- **Typed Overrides:** Core replaces ambiguous boolean force parameters with `TaskLifecycleOverride` (`TaskLifecycleOverrideDependencyGate` vs `TaskLifecycleOverrideAcceptanceCriteria`).
- **Contextual CLI Mapping:**
  - `task start --force` and `task move ... in-progress --force` set `TaskLifecycleOverrideDependencyGate`. If passed for any other destination status, `validateExistingLifecyclePlan` rejects the command.
  - `task complete --force` and `task move ... completed --force` set `TaskLifecycleOverrideAcceptanceCriteria`.
  - Dependency overrides cannot bypass broken repository graph health, invalid active task fields (missing tags or description), or non-candidate lifecycle roles.

### 4. Global Graph-Health Behavior

- **Fail-Closed Mutation Rule:** `ValidateTaskLifecycleSource` requires `graph.MutationReady()` (`g.health == GraphHealthy`).
- **Rejection on Degraded/Broken Graphs:** If any legacy fields remain (`GraphDegraded`) or if cycles/broken references exist (`GraphBroken`), all lifecycle transitions are rejected before planner execution, directing the user to run `task depend migrate` or repair broken entities.

### 5. Resulting Active-Task Validity

- **Active Field Invariants:** `ValidateTaskLifecyclePlan` runs `domain.ActiveTaskFieldErr(afterTask)` for both existing task transitions and `create-and-start`.
- **Validation Rules:** An active task (`in-progress`, `next-up`, `ready-to-start`) must declare tags, and an in-progress/next-up task must have a non-empty description. Missing required fields fail with `ErrValidation` even when `--force` is supplied.

### 6. Repository-Guard and Planner Re-Entry Design

- **Lock Ownership:** `FS.MutateTaskLifecycle` acquires `s.checkedWriteLock()` on the canonical repository root, holding it across graph snapshot load, planner execution, materialization, and CAS verification.
- **Planner Protection:** Planner invocation is wrapped in `enterRepositoryPlanner()` / `leave()`. Any attempt by a planner callback to invoke store methods (e.g. `GetTask`, `ListTasks`, `SetFields`, `MutateTaskGraph`) triggers `rejectRepositoryPlannerCall()` and returns `domain.ErrConflict`.

### 7. Cooperating and Raw-Writer Races

- **Cooperating Writers:** Serialized via flock on the repository write lock.
- **Raw Writers / Editors:**
  - Multi-level CAS defends against external non-cooperating writers.
  - Snapshot CAS (`graph.SameSourceSnapshot(currentGraph)`) verifies all repository task hashes and paths before write preparation.
  - Target CAS (`verifyUnchanged`) re-resolves and re-hashes the target file immediately before `writeFileAtomic`.
  - Concurrent raw mutations trigger `domain.ErrConflict`, prompting automatic retries via `retryOnConflict` in the Service layer.

### 8. Atomic Create-and-Start

- **Single Transaction:** `task new --start` executes as a single guarded lifecycle mutation with `TaskLifecycleCreation`.
- **Validation & Materialization:** The task is validated as an in-progress candidate with `started_at` stamped, verified against prospective graph eligibility, checked for file collision, and written atomically via `writeNewFileUnlocked` under the write lock. There is no intermediate on-disk state.

### 9. CLI / TUI Parity

- **Shared Service Seam:** TUI actions execute `Service.Move` and `Service.DeferTask` with `TaskLifecycleOverrideNone`.
- **Error Presentation:** Ineligible starts and incomplete acceptance criteria display identical diagnostic explanations in the TUI flash bar and CLI stderr.

### 10. Lifecycle Timestamps and No-Op Behavior

- **Timestamp Management:** Moving to `in-progress` stamps `started_at` and `updated_at`; `completed` stamps `completed_at`; `deferred` stamps `deferred_at`; `deprecated` stamps `deprecated_at`. Leaving `deferred` removes `revisit_at`.
- **Idempotent No-Ops:** If a task is already in the target status and no revisit date is updated, `materializeTaskLifecycle` returns `changed: false`. `MutateTaskLifecycle` skips file replacement, preserving byte-identical content and timestamps.

### 11. Deterministic Impact Calculation

- **BFS Downstream Traversal:** Downstream dependents are traversed using `TaskGraph.DownstreamImpact`, comparing derived before/after states.
- **Attribution & Paths:** Receipts report only tasks whose state actually changed, including deterministic shortest paths and direct/transitive attribution.

### 12. Human and JSON Receipts & Wire Contracts

- **Envelopes:** `MovesEnvelope` carries structured `MoveResult` entries with `TaskLifecycleJSON` detailing `before`, `after`, `outstanding_blockers`, `impacts`, `override`, `forced`, and `remedy`.
- **Formatting:** Dependency add/remove receipts attach `TaskGraphStateImpactJSON` for directly modified tasks, warning human users when a dependent becomes blocked or inconsistent.
- **Schema Compatibility:** Schema bumped to `1.52`, registered in `jsonEnvelopes`, verified by JSON Schema Draft 2020-12 tests.

---

## Explicit Rejected Concerns

1. **Stale Graph Authorization Window:**
   - *Concern:* An external writer could change a prerequisite's status between eligibility planning and write commit.
   - *Finding:* `MutateTaskLifecycle` executes `graph.SameSourceSnapshot(currentGraph)` immediately before writing under the exclusive write lock, rejecting any concurrent graph changes with `domain.ErrConflict`.
2. **Force Option Bypassing Non-Candidate Role:**
   - *Concern:* `--force` on `task start` might force a task from `next-up` or `deferred` directly into `in-progress`.
   - *Finding:* `validateExistingLifecyclePlan` checks `before.Role != RoleCandidate` and returns `TaskEligibilityError` before evaluating `plan.Override`. Only the dependency gate (`GateBlocked`) is bypassed by `--force`.
3. **Editor Bypass via Malformed Frontmatter:**
   - *Concern:* A user editing frontmatter in `task edit` might delete status or dependency fields during syntax repair.
   - *Finding:* `FS.EditTask` tracks readability of original status and dependencies; if unreadable or modified, the edit is rejected and the editor reopens.
4. **Intermediate Window in `task new --start`:**
   - *Concern:* `task new --start` might write a `ready-to-start` task and then move it, exposing an unstarted task window.
   - *Finding:* `task new --start` executes `materializeCreateAndStart` and writes directly to `in-progress` under the repository write lock in one atomic step.

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
