---
schema: 1
id: 6g4zvqddz3gz
bucket: closed
area: thread-documents-guarded-creation-and-read-projections-implementation-antigravity
date: "2026-08-29"
updated_at: "2026-08-29"
---

# Audit: Thread Documents, Guarded Creation, and Read Projections — 2026-08-29

> Edit findings in place and flip each `**Status:**` as you work it.

Adversarial implementation review of Thread documents, guarded unstarted creation, lock-free materialization, and deterministic read projections (task `6g3q4rtmv4ak`, epic `30-threads-and-task-dependency-graphs`, branch `feat/thread-documents-read-projections`, uncommitted tree over merge-base `93ba8d7`), evaluated against ADR-0006, `docs/ARCHITECTURE.md`, and the task specification.

**Executive Verdict: Ready.** The implementation establishes durable Thread documents as first-class, flat, ID-addressed entities while strictly keeping task documents as the sole dependency source of truth. The dedicated `ThreadCreationMutationStore` capability executes guarded creation inside the canonical-root write lock, loads an authoritative task-and-Thread snapshot, validates membership and bidirectional cross-kind ID collisions, enforces unstarted initial lifecycle status, materializes through a reusable lock-free internal seam, verifies whole-snapshot CAS tokens, and writes via exclusive creation. A single pure `ThreadView` projection drives `thread list`, `thread show`, and `thread frontier`, reporting stable member vs external-gate roles, nominal `done/total` vs sound `drained/total` rollups, graph health, and fail-closed frontier dispatch. Legacy Projects scaffolding is retired safely by removing only `.gitkeep`-only empty directories while preserving all user content.

---

## Findings

### Low

#### L1. Mutable membership operations and bulk linking sequenced into subsequent slice · **Status:** tracked by 6g4wm2yf6tyj

**File:** `internal/cli/thread.go:29-36`, `planning/tasks/6g3q4rtmv4ak-add-thread-documents-guarded-creation-and-read-projections.md:53-64` | **Component:** core / sequence boundary
**Effort:** XS · **Urgency:** eventually

**[monitor]**

**Description:**
In accordance with ADR-0006 and the task readiness checkpoint, mutable membership verbs (`thread add`, `thread remove`), Thread lifecycle transitions (`thread start`, `thread complete`, `thread abandon`), and bulk linking (`task link --thread`) are explicitly deferred to the immediately following planning task (`ship-guarded-thread-membership-and-lifecycle-mutations`). The current slice cleanly bounds itself to document structure, guarded unstarted creation, lock-free materialization, and deterministic read projections.

**Impact:**
Architecturally sound separation of concerns that prevents oversized pull requests while providing complete read-first support and a stable materialization seam.

---

**Resolution:** Mutable membership and lifecycle remain owned by 6g4wm2yf6tyj;
resumable bulk linking remains sequenced separately in 6g3q4rtv8d0a. The
creation/read slice does not widen.

#### L2. Unreadable or missing Thread member tasks included in progress denominator · **Status:** wontfix

**File:** `internal/core/thread_projection.go:101-112` | **Component:** core / projection rollup
**Effort:** XS · **Urgency:** eventually

**[monitor]**

**Description:**
When a Thread document declares a member task ID that does not exist in the repository task graph (e.g. from an unreadable task file or dangling reference), `ProjectThread` increments `view.Rollup.Total++` rather than skipping the entry. It marks `view.Health = GraphBroken` and attaches `ThreadProblemMissingMember`.

**Impact:**
Intentional fail-safe design. Counting declared missing members in `Total` prevents broken or unreadable tasks from manufacturing a false 100% or 0/0 completion rollup.

---

**Resolution:** Retained intentionally: a declared missing member stays in total
while projection health is broken, preventing malformed evidence from
manufacturing 0/0 or 100 percent completion.

## Traceability Table

| Acceptance Criterion | Status | Implementation Seam | Test Coverage |
| :--- | :---: | :--- | :--- |
| **1. Metadata & sorted membership ownership**<br>Thread files own metadata and sorted task-ID membership only; task files remain the dependency source of truth. | **Fulfilled** | `internal/domain/thread.go:48-66`<br>[`Thread struct`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/domain/thread.go#L48-L66), `internal/domain/thread.go:105-117`<br>[`ValidateThreadDocument`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/domain/thread.go#L71-L119) | `internal/domain/thread_test.go:20-65`<br>`TestValidateThreadDocumentInvariants` |
| **2. Shared membership & external gates**<br>Tasks can belong to multiple Threads; outside prerequisites appear as external gates without entering progress totals while preventing sound completion. | **Fulfilled** | `internal/core/thread_projection.go:90-148`<br>[`ProjectThread`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/thread_projection.go#L70-L162) | `internal/core/thread_projection_test.go:45-90`<br>`TestProjectThreadSharedMembersAndExternalGates` |
| **3. Nominal vs sound rollups**<br>Views expose nominal `done / total`, sound `drained / total`, graph health, and exact outstanding external gates without contradictory completion UX. | **Fulfilled** | `internal/core/thread_projection.go:115-126`<br>Rollup computation, `internal/cli/render/thread.go:30-80` | `internal/core/thread_projection_test.go:92-140`<br>`TestProjectThreadNominalVsSoundRollups` |
| **4. Shared graph & eligibility analysis reuse**<br>CLI and wire create/read projections use shared graph/eligibility analysis and expose stable member/external roles without reimplementing task graph rules. | **Fulfilled** | `internal/core/thread_projection.go:36-46`<br>[`ThreadTaskView`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/thread_projection.go#L36-L46), `internal/wire/thread.go:42-84`<br>[`ThreadViewJSON`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/wire/thread.go#L74-L85) | `internal/wire/envelopes_test.go:180-220`<br>`TestThreadViewJSONSchemaValidation` |
| **5. Safe legacy Projects retirement**<br>Initialization creates `threads/`, stops creating `projects/`, and handles non-empty legacy Projects safely. | **Fulfilled** | `internal/config/config.go:618-663`<br>[`retireProjectsScaffold`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/config/config.go#L618-L663), `internal/config/config.go:580-605` | `internal/config/init_test.go:440-525`<br>`TestInitRetiresProjectsScaffoldSafely` |
| **6. Cross-kind ID collision prevention**<br>Task/Thread IDs are checked for cross-kind collisions, and an empty Projects scaffold permits only `.gitkeep` removal. | **Fulfilled** | `internal/store/create.go:191-202`<br>[`ensureTaskIDNotThread`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/create.go#L191-L202), `internal/core/thread_creation.go:88-90,120-122` | `internal/store/threadcreation_test.go:90-130`<br>`TestCrossKindTaskAndThreadIdentityCollisions` |
| **7. Guarded unstarted Thread creation**<br>Creation always persists `unstarted`, validates member existence and cross-kind ID collisions inside one canonical-root guard, and rejects arbitrary initial lifecycle states. | **Fulfilled** | `internal/core/thread_creation.go:100-134`<br>[`ValidateThreadCreationPlan`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/thread_creation.go#L100-L134), `internal/store/threadcreation.go:18-96` | `internal/store/threadcreation_test.go:30-85`<br>`TestThreadCreationGuardedExecution` |
| **8. Lock-free internal materialization**<br>The store provides lock-free internal Thread materialization reusable by compound bulk capabilities without nesting guarded mutations. | **Fulfilled** | `internal/store/threadcreation.go:115-139`<br>[`materializeThreadCreation`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/threadcreation.go#L115-L139) | `internal/store/threadcreation_test.go:132-160`<br>`TestLockFreeThreadMaterialization` |
| **9. Real two-store concurrency serialization**<br>A real two-store race between Thread creation and task lifecycle mutation proves canonical-root serialization and fresh-snapshot validation. | **Fulfilled** | `internal/store/threadcreation.go:88-90`<br>Whole-snapshot CAS, `internal/store/threadcreation_test.go:162-210` | `internal/store/threadcreation_test.go:162-210`<br>`TestThreadCreationRacesWithTaskLifecycleMutation` |

---

## Detailed Review by Hostile Angles

### 1. Thread Domain Invariants, Document Parsing, Stable Identity, and Malformed Hand-Edited State

- **Domain Model:** `domain.Thread` enforces 12-character Crockford base32 stable IDs, closed status vocabulary (`unstarted`, `in-progress`, `completed`, `abandoned`), required description ($\le 120$ chars), required single-line goal, valid `YYYY-MM-DD` dates, and a duplicate-free, sorted list of member task IDs.
- **Parsing:** `store.parseThread` splits flat filenames (`<id>-<slug>.md`), parses YAML frontmatter strictly via `splitFrontmatterStrict`, and records `SourceVersion` hash.
- **Diagnostic Tolerance:** Malformed hand-edited Thread documents are retained by `ListThreads` as `FileProblem` records so `thread list`, `thread show`, and `tskflwctl lint` diagnose field defects with actionable remediation guidance instead of failing silently.

### 2. Projection Semantics

- **Shared Membership:** Tasks can belong to multiple Threads simultaneously. `ProjectThread` reads task state from the shared immutable `TaskGraph` without duplicating graph traversal rules.
- **External Gates:** Prerequisites of member tasks outside the Thread's membership are classified as `ThreadTaskExternalGate`. They never contribute to `Total`, `Done`, or `Drained`, but `gate.Outstanding = !state.SoundlyCompleted` ensures that outstanding external gates prevent sound Thread completion.
- **Nominal vs Sound Rollups:**
  - `Rollup.Total`: excludes deprecated members, includes active/deferred/unreadable members.
  - `Rollup.Done`: counts nominally completed tasks (`task.Status == completed`).
  - `Rollup.Drained`: counts soundly completed tasks (`state.Drained == true`).
  - `Rollup.Deprecated`: tracks deprecated members separately.
- **Completed-Thread Inconsistency:** When a Thread is marked `completed`, `view.Inconsistent` becomes `true` if `view.Health != GraphHealthy`, `view.Rollup.Total == 0`, `view.Rollup.Drained != view.Rollup.Total`, or any external gate is outstanding. Reopening an upstream task immediately flips completed downstream Threads to inconsistent.
- **Fail-Closed Frontier:** `view.Frontier` evaluates only member tasks with `member.State.Eligible == true`. If `view.Health != GraphHealthy` (broken/degraded graph), `view.Frontier` returns empty (`[]`), failing closed against automated dispatch.

### 3. Guarded Creation

- **Canonical-Root Guard:** `FS.MutateThreadCreation` acquires `s.checkedWriteLock()` before loading snapshots or executing planners.
- **Authoritative Snapshot Verification:** Validates `core.ValidateThreadCreationSource(graph, threads, problems)` to ensure graph is `MutationReady()`, all existing Thread documents are valid, and no ID drift exists.
- **Bidirectional Cross-Kind Identity Collision Prevention:**
  - `ValidateThreadCreationPlan` checks that the new Thread ID is not already used by any task or existing Thread.
  - `store.CreateTask` and `store.MutateTaskLifecycle(Create)` invoke `s.ensureTaskIDNotThread(taskID)` under the write lock, ensuring task creation cannot collide with a Thread ID.
- **Re-Entry Protection:** Planner callbacks run under `enterRepositoryPlanner()`. Any attempt to call store methods during planning is rejected with `domain.ErrConflict`.
- **Multi-Level CAS:** Re-reads `currentGraph` and `currentThreads` before write, comparing `graph.SameSourceSnapshot` and `sameThreadSourceSnapshot`. Writes via `createFileAtomic` with `O_CREAT | O_EXCL`.
- **Committed Post-Write Failure Recovery:** If writing the Thread file succeeds but guard release or cleanup fails, the error is wrapped in `ThreadCreationMutationFailure{Committed: true}` so CLI and JSON consumers recognize that the file exists on disk and avoid blind retries.

### 4. Safe Legacy Projects Retirement

- **Surgical Retirement:** `retireProjectsScaffold` uses `os.Lstat` and `os.ReadDir` to inspect `projects/`.
- **Narrow Policy:** Only an empty directory or one containing exactly a single regular `.gitkeep` file is removed.
- **User Content Preservation:** If `projects/` is a symlink, contains other files/directories, or contains Markdown notes, it is left untouched and reported for manual operator review.
- **Dry-Run Equivalence:** `tskflwctl init --dry-run` reports would-be retirements without modifying the filesystem.

### 5. Lint and Diagnostic Behavior

- **Thread Document Linting:** `domain.LintThread` validates all frontmatter fields, ID-filename agreement, date formatting, and member task existence.
- **Duplicate & Cross-Kind ID Detection:** `core.Service.Lint` detects duplicate Thread IDs and cross-kind collisions with tasks across the entire planning tree.
- **Unsafe Automatic Repair Prevention:** `lint --fix` sweeps stale temp orphans and repairs misspelled entity IDs only when a deterministic canonical Crockford decode exists and no external references are broken; it never rewrites Thread membership sets automatically.

### 6. CLI and Wire Contracts

- **Commands:** Exposes `thread new <title>`, `thread list`, `thread show <thread>`, `thread path <thread>`, and `thread frontier <thread>`.
- **Deterministic Ordering:** Tasks, members, external gates, and frontiers are sorted deterministically by stable ID.
- **Schema & Envelopes:** Wire schemas updated to version `1.54` (`ThreadsEnvelope`, `ThreadShowEnvelope`, `ThreadFrontierEnvelope`, `ThreadMutationEnvelope`). Empty arrays serialize as `[]` rather than `null`.
- **Human & JSON Diagnostics:** `thread frontier` output retains diagnostic problems and graph health even when returning zero dispatchable tasks.

### 7. Architectural Fan-Out and Materialization Seams

- **Ports & Interfaces:** `ThreadStore` and `ThreadCreationMutationStore` defined in `internal/core/store.go`.
- **Lock-Free Materializer:** `store.FS.materializeThreadCreation` is decoupled from locking, allowing future compound bulk operations to compose Thread materialization under one outer repository guard.
- **Subsystem Integration:** Space health checks `threads/`, watcher registers `threadsDir`, shell completion supports Thread slugs, and CLI documentation is generated.

---

## Explicit Rejected Concerns

1. **Premature Mutable Membership Verbs:**
   - *Concern:* The slice does not include `thread add` or `thread remove`.
   - *Finding:* Sequenced intentionally per ADR-0006 and the readiness checkpoint into `ship-guarded-thread-membership-and-lifecycle-mutations`. The document schema, lock-free materializer, and read projections are fully established.
2. **Excluding Missing Members from Denominators:**
   - *Concern:* An unreadable member task could be excluded from the `Total` count.
   - *Finding:* `ProjectThread` includes missing members in `Rollup.Total`, marks the Thread `GraphBroken`, and surfaces `ThreadProblemMissingMember`, preventing false 100% completion rollups.
3. **Recursive Deletion of Legacy Projects:**
   - *Concern:* `init` could delete legacy notes stored in `projects/`.
   - *Finding:* `retireProjectsScaffold` removes only `.gitkeep`-only empty directories, preserving all other files.
4. **Task/Thread ID Collision Vulnerability:**
   - *Concern:* A task and Thread might receive identical 12-char Crockford base32 IDs.
   - *Finding:* Bidirectional checks in `ValidateThreadCreationPlan` and `ensureTaskIDNotThread` prevent collisions on all creation paths under the repository write lock.

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
# Result: ✔ all planning entities and dependency links pass lint

go run ./cmd/tskflwctl audit lint
# Result: ✔ all audit findings pass lint
```
