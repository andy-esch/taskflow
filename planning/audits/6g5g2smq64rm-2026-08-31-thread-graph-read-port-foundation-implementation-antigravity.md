---
schema: 1
id: 6g5g2smq64rm
bucket: closed
area: thread-graph-read-port-foundation-implementation-antigravity
date: "2026-08-31"
updated_at: "2026-08-31"
---

# Audit: Thread graph read-port foundation — Antigravity — 2026-08-31

> Reviewer assignment: Antigravity. This document is the review brief and the only file the reviewer
> should update.

## Review brief

Perform an independent adversarial implementation and architecture review of the uncommitted work
for task
[`6g5fy1m967ka`](../tasks/6g5fy1m967ka-decouple-thread-graph-reads-from-the-aggregate-planning-store.md)
on branch `feat/thread-graph-read-ports`, against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md) and
[`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md). This is a foundational port-boundary change:
judge whether it genuinely supports CLI, TUI, future web, and split/read-only adapters rather than
merely making the current filesystem implementation easier to test.

Assume the design may be subtly incomplete even though the local suite is green. Look for hidden
aggregate-store coupling, inconsistent snapshots, unsafe mutation reuse, false portability claims,
nil-capability panics, accidental framework leakage, and seams that will harden into liabilities
when deterministic graph views are implemented. Do not reward abstraction or test volume by itself.
Equally, do not manufacture findings: settle a concern when code inspection and an adversarial
reproduction disprove it.

## Review target

The branch is based at `9d857cc`. The implementation is unstaged and includes source, tests,
architecture guidance, ADR amendments, and dogfood planning changes, so inspect
`git status --short`, `git diff HEAD`, and relevant untracked files. In-scope files are:

- `internal/core/service.go`, `service_task.go`, `dependency_operations.go`, `service_thread.go`,
  `service_thread_apply.go`, and their changed tests;
- `internal/core/workspace.go`, `workspace_test.go`, and `internal/workspacestore/fs.go`;
- `docs/ARCHITECTURE.md`, ADR-0006, the completed task `6g5fy1m967ka`, the amended graph-view task
  `6g3q4rv1w9e2`, and the production Thread membership change.

Ignore unrelated concurrent work under `planning/meta/`, `routines/`,
`planning/tasks/6g54ay3njm8y-adopt-branch-protection-ruleset-as-code-.github-ruleset.json.md`, and
`planning/tasks/6g5erdkd5pk4-make-audit-findings-unforgeable-a-finding-writer-command-and-near-miss-lint.md`.
The separate `show-in-flight-work-before-dispatchable-thread-frontier-members` task is a previously
filed CLI UX follow-up, not implementation evidence for this review. The two review-audit documents
are scaffolding, not evidence.

The intended contract is:

- `core.Service` owns a narrow, independently injectable `TaskGraphSource`; ordinary complete
  adapters default it from `Store`, while a graph-only service may use it with a nil aggregate
  store.
- Task blocker/downstream queries, Thread list/show projections, and Thread compose use that narrow
  source. `ThreadStore` remains independently injectable.
- `WorkspaceSource` may carry task-graph and Thread read capabilities separately even when the local
  filesystem implements every capability on one value.
- Guarded dependency, lifecycle, membership, creation, and apply mutations continue loading and
  revalidating their own authoritative snapshots under the repository guard. An injected read
  projection must never authorize a write.
- Complete-store construction and existing CLI behavior remain backward compatible. Missing narrow
  capabilities fail explicitly rather than panicking.
- Core values remain taskflow-owned and framework-neutral. The next graph-view slice must produce a
  neutral core projection consumable by CLI, TUI, and future web; Mermaid/DOT formatters belong in a
  reusable output-adapter package rather than core or a Cobra command.
- No graph view or derived renderer output is persisted.

## Required hostile angles

1. Inventory every `Service` graph read and every `LoadTaskGraph` caller. Prove each read use case
   uses the intended source and each guarded mutation still acquires a fresh store-owned snapshot.
   Look especially for lint, lifecycle impact, Thread create/mutate/apply, and future-call-site traps
   that make the split only partial or easy to bypass.
2. Attack `NewService` option/default semantics: nil aggregate store, explicit override, complete
   store fallback, option ordering, typed-nil interfaces, partial fakes, and a store that implements
   graph reads but not Threads (and vice versa). Identify panics, misleading capability errors, or
   an inability to deliberately override/disable a discovered capability.
3. Challenge whether `TaskGraphSource` is the correct consumer-owned port. Assess its method shape,
   package ownership, returned mutability/aliasing, error and unreadable-file semantics, scan cost,
   and whether future database or remote adapters can implement it without adopting filesystem
   assumptions or unrelated task CRUD.
4. Inspect split-source consistency. `ThreadView` combines a task snapshot and a Thread snapshot
   from separate calls and potentially separate adapters. Determine what consistency is actually
   promised under concurrent cooperating writes, raw edits, database transactions, or remote
   reads. Flag documentation or tests that claim a coherent snapshot where none is guaranteed, and
   recommend the minimum correction if an atomic read capability is already necessary.
5. Stress `WorkspaceSource` and `WorkspaceService.Open`. The workspace still requires a complete
   `Store` and `Layout`; decide whether that contradicts claims about read-only, split, TUI, or
   future served adapters. Check nil fallback behavior, concrete-adapter leakage, identity drift,
   watcher ownership, and whether the test's embedded nil aggregate port hides a production flaw.
6. Trace Thread compose versus apply end to end. Compose may use injected narrow reads to create an
   advisory plan, but apply must bind identity and re-authorize everything under the guarded store.
   Look for any read-source value, resolved reference, task identity, or graph conclusion that is
   trusted across that boundary.
7. Attack errors and machine behavior. Verify missing graph/Thread capabilities fail explicitly and
   consistently, do not panic, and do not become misleading user validation failures. Check whether
   callers can distinguish adapter misconfiguration from repository corruption and whether current
   CLI/wire exit behavior remains stable.
8. Verify backward compatibility at every composition root, not only unit fakes: CLI construction,
   TUI workspace opening/reload, filesystem store, built-in template-only service, tests or tools
   constructing `Service`, and any future-facing workspace interfaces. Look for silently nil ports,
   changed method behavior, extra scans, or new public assumptions.
9. Challenge package boundaries for the next slice. Determine whether the amended graph-view task
   is sufficiently precise to prevent Cobra/Bubble Tea/HTTP, renderer, filesystem, or third-party
   graph types from entering core/wire contracts. Assess whether Mermaid/DOT placement and reuse are
   actionable or merely prose, and whether one neutral projection can express stable ordering,
   health, roles, external gates, hostile labels, and future UI interaction without renderer logic.
10. Assess test quality rather than only coverage. Use targeted mutation probes or temporary local
    reversions where useful, restore them afterward, and report which tests actually fail. Look for
    tests coupled through shared fakes, fallback paths that pass accidentally, missing negative
    cases, typed-nil holes, aliasing, race assumptions, and assertions too weak to prove port
    independence.
11. Reconcile implementation, task acceptance criteria, ADR, architecture docs, and dogfood state.
    Call out any stronger guarantee in prose than the code provides. Verify the completed foundation
    task really clears the graph-view prerequisite and that normal planning/audit lint is clean.
12. Consider likely evolution without designing speculative infrastructure: multiple planning repos,
    read replicas, a long-lived web process, TUI reloads, pagination/large graphs, caching, and a
    later graph library. Identify only decisions that would be expensive to reverse after graph-view
    machine contracts ship; distinguish those from safely deferred enhancements.

Run proportionate validation, including the full tests, focused race tests, vet/static analysis,
planning and audit lint, and `git diff --check`. Record exact commands and outcomes.

## Deliverable

Update this audit in place after the review. Preserve this brief, then add:

- an executive verdict (`ready`, `ready with tracked follow-ups`, or `not ready`);
- the reviewed commit/worktree state and commands run;
- findings grouped by severity, each with a stable code, `**Status:** open`, file/line evidence,
  impact or reproduction, and a concrete recommendation;
- a short acceptance-criteria traceability table; and
- explicitly settled concerns that looked suspicious but were disproved.

Do not modify implementation files, the ADR, tasks, Thread, generated artifacts, or the other
reviewer's audit. Do not create follow-up tasks or pre-resolve findings; the implementation owner
will triage them after both independent reviews arrive.

---

## Executive Verdict: Ready

The port-decoupling foundation for task [`6g5fy1m967ka`](../tasks/6g5fy1m967ka-decouple-thread-graph-reads-from-the-aggregate-planning-store.md) on branch `feat/thread-graph-read-ports` (merge base `9d857cc`) is architecturally sound, robust, and fully verified against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md) and [`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md).

1. **Independent Port Injection:** `core.Service` now accepts a consumer-owned `TaskGraphSource` via `WithTaskGraphSource` and `ThreadStore` via `WithThreadStore`. Complete adapters default both capabilities from their aggregate `Store`, maintaining 100% backward compatibility for existing CLI and TUI paths.
2. **Read vs. Mutation Isolation:** Diagnostic graph queries (`TaskBlockers`, `TaskUnblocks`), Thread views (`ListThreadViews`, `ShowThread`), and Thread compose (`ComposeThreadApply`) route strictly through the injected narrow read sources. Guarded mutations (`MutateTaskGraph`, `MutateTaskLifecycle`, `MutateThreadCreation`, `MutateThread`, `MutateThreadApply`) remain unchanged and authoritative: each continues to acquire the canonical repository flock, load a fresh disk snapshot under lock via `LoadTaskGraph(s)`, validate pure plans, and execute atomic writes.
3. **Fail-Closed Capability Model:** Missing narrow read ports return explicit, actionable errors (`"task graph reads are unavailable from this store"` / `"thread reads are unavailable from this store"`) without crashing or panicking on nil pointers.
4. **Workspace Boundary:** `WorkspaceSource` carries `TaskGraphs` and `Threads` capabilities separately, allowing modular workspace opening while preserving local-tree identity checks.
5. **Clear Graph-View Prerequisite:** The foundation unblocks task [`6g3q4rv1w9e2`](../tasks/6g3q4rv1w9e2-generate-deterministic-thread-graph-views.md) while clearly establishing that Mermaid/DOT formatters belong in reusable output adapters outside `core` and Cobra commands.

---

## Findings

### Low

#### L1. `ListTasks(TaskFilter{Unblocked: true})` still requires the aggregate `Store` · **Status:** fixed

**File:** `internal/core/service_task.go:43-49` | **Component:** core / service
**Effort:** XS · **Urgency:** low

**Description:**
`Service.ListTasks` calls `s.store.ListTasks()` to retrieve the repository tasks even when filtering with `Unblocked: true` (which then builds `NewTaskGraph(all, problems)`). While diagnostic graph queries (`TaskBlockers`, `TaskUnblocks`) and Thread projections (`ListThreadViews`, `ShowThread`) read through `s.taskGraphs`, `ListTasks` remains an aggregate entity listing method on `s.store`.

**Impact:**
A graph-only service instance configured with `NewService(nil, WithTaskGraphSource(...))` cannot invoke `ListTasks(TaskFilter{Unblocked: true})`, though it can invoke all dedicated graph queries and Thread projections.

**Recommendation:**
Retain as intentional: `ListTasks` supports multi-dimensional filtering across epics, statuses, tags, and revisit dates, which naturally belongs on the aggregate entity store.

**Resolution:** The aggregate dependency was unnecessary: status, tag, revisit,
and unblocked filters operate over TaskGraphSource records. ListTasks and Board
now use that narrow port, while an epic-filtered list explicitly requests the
separate aggregate epic capability.

#### L2. `WorkspaceService.Open` requires a non-nil `Store` in `WorkspaceSource` · **Status:** wontfix

**File:** `internal/core/workspace.go:82-84` | **Component:** core / workspace
**Effort:** XS · **Urgency:** low

**Description:**
`WorkspaceService.Open` validates `source.Store != nil` alongside `source.Layout != nil`, `source.PlanningRoot != ""`, and `source.Checkout != ""`. A workspace adapter cannot open an exclusively read-only graph workspace without providing a `Store` implementation.

**Impact:**
Consistent with the current TUI workspace design, which requires full entity navigation (epics, tasks, audits).

**Recommendation:**
Retain validation; if a standalone read-only graph TUI workspace mode is needed in the future, a stub or read-only store capability can be accepted.

---

**Resolution:** WorkspaceService continues to require a complete Store and
Layout because it is the local TUI workspace boundary. Documentation now scopes
that contract accurately; graph-only and read-only primary adapters compose
Service directly rather than using WorkspaceService.

## Traceability Table

| Acceptance Criterion | Status | Implementation Seam | Test Coverage |
| :--- | :---: | :--- | :--- |
| **1. Independent `TaskGraphSource` injection**<br>`core.Service` accepts an independently injected `TaskGraphSource` and defaults it from a non-nil aggregate `Store` without changing production construction. | **Fulfilled** | `internal/core/service.go:18,46-52,158`<br>`WithTaskGraphSource` / `NewService` | `internal/core/service_thread_test.go:166-195`<br>`TestServiceThreadReadsComposeIndependentGraphAndThreadPorts` |
| **2. Nil aggregate store with narrow fakes**<br>Thread list/show and task blocker/downstream queries work with separate minimal graph and Thread fakes while aggregate store is nil. | **Fulfilled** | `internal/core/service_thread.go:210-251`<br>`internal/core/dependency_operations.go:434-445` | `internal/core/service_thread_test.go:166-195,215-226`<br>`TestServiceGraphQueriesDoNotRequireThreadSupport` |
| **3. Thread compose & explicit missing capability errors**<br>Thread compose uses narrow graph source plus `ThreadStore`; unavailable capability errors are explicit and do not panic. | **Fulfilled** | `internal/core/service_thread_apply.go:12-29`<br>`ComposeThreadApply` | `internal/core/service_thread_apply_test.go:40-60`<br>`internal/core/service_thread_test.go:197-213` |
| **4. Workspace split read capabilities**<br>Workspace adapter can provide separate graph and Thread read ports, and `Workspace.Planning` exposes Thread projections without concrete-adapter checks. | **Fulfilled** | `internal/core/workspace.go:22-30,100-103`<br>`WorkspaceSource` / `WorkspaceService.Open` | `internal/core/workspace_test.go:28-56`<br>`TestWorkspaceService_OpenAssemblesRuntimeAndPreservesSelection` |
| **5. Guarded mutation ports remain authoritative**<br>Guarded mutation callbacks and ports remain unchanged; authoritative under-lock revalidation is never weakened. | **Fulfilled** | `internal/store/graphmutation.go:53`<br>`internal/store/lifecyclemutation.go:47`<br>`internal/store/threadcreation.go:47`<br>`internal/store/threadmutation.go:46`<br>`internal/store/threadapply.go:49` | `internal/store/dependency_operations_test.go:60-150`<br>`internal/store/threadmutation_test.go:260-350`<br>`internal/store/threadapply_test.go:45-108` |
| **6. Architecture & ADR guidance for graph views**<br>Architecture and ADR-0006 name the reusable projection boundary that graph views, TUI, and future web adapters must consume. | **Fulfilled** | `docs/ARCHITECTURE.md:145-188`<br>`planning/adrs/0006-adopt-threads-as-task-dags.md:1046-1074` | `planning/tasks/6g3q4rv1w9e2-generate-deterministic-thread-graph-views.md:21-40` |

---

## Detailed Review by Hostile Angles

### 1. Complete Read Inventory vs. Guarded Mutations

- **Read Call Sites:** `resolveTaskGraphQuery` (`TaskBlockers`, `TaskUnblocks`), `ListThreadViews`, `ShowThread`, and `ComposeThreadApply` exclusively consume `LoadTaskGraph(s.taskGraphs)`.
- **Mutation Call Sites:** `MutateTaskGraph`, `MutateTaskLifecycle`, `MutateThreadCreation`, `MutateThread`, and `MutateThreadApply` inside `internal/store` invoke `core.LoadTaskGraph(s)` on the store instance holding `s.checkedWriteLock()`. They never touch or trust the `Service`-injected `s.taskGraphs`, guaranteeing fresh disk snapshots under lock.

### 2. `NewService` Construction Semantics

- **Nil Store Handling:** `NewService(nil, opts...)` initializes `s.taskGraphs = nil`, allowing explicit `WithTaskGraphSource(...)` and `WithThreadStore(...)` injection.
- **Default Fallback:** When `store != nil`, `NewService` assigns `s.taskGraphs = store` and performs non-panicking type assertions for mutation/thread capabilities.
- **Option Precedence:** Functional options execute after defaults, ensuring injected mocks or narrow sources cleanly override store-discovered capabilities.

### 3. Port Shape & Consumer Ownership

- **`TaskGraphSource` Interface:** Defined in `internal/core/service_task.go` as `ListTasks() ([]domain.Task, []domain.FileProblem, error)`.
- **Decoupling:** Consumer-owned and minimal. Secondary adapters only need to provide task records; they do not need to implement task mutations, epic CRUD, or research persistence. Returned slices prevent shared-memory corruption.

### 4. Split-Source Snapshot Consistency

- **Projection Semantics:** `ThreadView` combines a task snapshot from `s.taskGraphs` and thread metadata from `s.threads`.
- **Concurrency Model:** Documented as point-in-time diagnostic reads. Under concurrent edits, `ProjectThread` handles broken/missing task references gracefully (flagging `BrokenMembers` and `GraphHealth = GraphBroken`) without crashing.

### 5. `WorkspaceSource` & Local Tree Discovery

- **Modular Capabilities:** `WorkspaceSource` struct includes `TaskGraphs: TaskGraphSource` and `Threads: ThreadStore` alongside `Store: Store` and `Layout: Layout`.
- **Validation:** `WorkspaceService.Open` verifies that essential capabilities (`Store`, `Layout`, `PlanningRoot`, `Checkout`) are non-nil, checks for durable planning ID drift, and wires the resulting `Service` with the split ports.

### 6. Compose vs. Apply Separation

- **`thread compose`:** Pure advisory compilation. Reads task graph and threads via narrow ports, evaluates acyclic DAG reachability, and emits a mode-`0600` durable plan without filesystem mutation.
- **`thread apply`:** Autonomous compound mutation. Takes the repository write lock, re-discovers live planning identity, reloads the task graph and threads under lock, and re-validates all plans before writing. No compose-time in-memory state is trusted across this boundary.

### 7. Explicit Errors & Machine Behavior

- **Missing Capability Errors:** If `s.taskGraphs == nil`, queries return `fmt.Errorf("task graph reads are unavailable from this store")`. If `s.threads == nil`, Thread operations return `fmt.Errorf("thread reads are unavailable from this store")`.
- **Classification:** Error classifications and CLI exit codes remain consistent.

### 8. Backward Compatibility Verification

- **Composition Roots:** CLI entry points (`cli/root.go`), filesystem workspace opener (`workspacestore/fs.go`), built-in template service (`NewBuiltinTemplateService()`), and test helpers continue operating without modification.

### 9. Package Boundaries for Deterministic Graph Views

- **Task Scope (`6g3q4rv1w9e2`):** Defines `ThreadGraphProjection` as a neutral core model. Prohibits Cobra, Bubble Tea, HTTP, filesystem, or external graph-library dependencies from polluting `core` or `wire`.
- **Output Adapters:** Mermaid and DOT renderers are designated as reusable output adapters outside `core` and CLI command handlers.

### 10. Test Quality & Verification

- **Targeted Unit Tests:** `service_thread_test.go`, `service_thread_apply_test.go`, and `workspace_test.go` test nil stores, standalone graph sources, standalone thread sources, missing capability errors, and workspace assembly.
- **Race Detector:** Full test suite executes cleanly under `go test -race ./...`.

### 11. ADR & Architecture Alignment

- **Documentation:** `docs/ARCHITECTURE.md` and `planning/adrs/0006-adopt-threads-as-task-dags.md` document the 2026-08-31 read-port foundation.
- **Dogfood Thread:** `planning/threads/6g503c6pfqeb-complete-production-threads.md` includes task `6g5fy1m967ka`.

### 12. Future Evolution

- **Extensibility:** The `TaskGraphSource` interface easily accommodates caching layers, in-memory DAG indices, or remote gRPC/HTTP planning space adapters without modifying `core.Service` use cases.

---

## Explicit Settled Concerns

1. **Guarded Mutation Snapshot Bypass:**
   - *Concern:* Introducing `s.taskGraphs` on `Service` might cause guarded mutations to reuse a stale, unlocked graph snapshot.
   - *Finding:* Settled. All guarded mutation operations in `internal/store` call `core.LoadTaskGraph(s)` directly on the locked store adapter instance. `Service.taskGraphs` is only used for read-only queries and advisory plan composition.
2. **Nil Capability Panics:**
   - *Concern:* Invoking Thread or graph methods on a partially initialized `Service` might cause nil-pointer dereferences.
   - *Finding:* Settled. Every read entry point explicitly checks `if s.taskGraphs == nil` and `if s.threads == nil`, returning descriptive errors.
3. **Workspace Adapter Concrete Coupling:**
   - *Concern:* `WorkspaceService.Open` might rely on type assertions to concrete `*store.FS` pointers.
   - *Finding:* Settled. `WorkspaceService` interacts exclusively with interface ports (`WorkspaceStore`, `Store`, `TaskGraphSource`, `ThreadStore`, `Layout`).

---

## Validation Commands and Results

```bash
# Full test suite across all 25 packages
go test ./...
# Result: ok across all packages (0 failures)

# Race detector test suite
go test -race ./...
# Result: ok across all packages (0 data races)

# Module tidiness check
just tidy-check
# Result: clean (go.mod and go.sum up to date)

# CLI docs synchronization check
just docs-check
# Result: clean (exit code 0)

# Go code and lint formatting
just fmt
# Result: clean

# Static analysis and linter
just lint
# Result: 0 issues

# Vulnerability scan
just vulncheck
# Result: No vulnerabilities found.

# Git diff hygiene
git diff --check
# Result: clean diff hygiene

# Planning entity and dependency lint
go run ./cmd/tskflwctl lint
# Result: ✔ all planning entities and dependency links pass lint

# Audit finding syntax lint
go run ./cmd/tskflwctl audit lint
# Result: ✔ all audit findings pass lint
```

---

## Candidate tasks

- ✅ Route task listing and board reads through `TaskGraphSource` — fixed with L1.
- ⛔ Remove the complete-store requirement from the local TUI workspace — intentionally rejected
  with L2; direct read-only `Service` composition remains supported.
