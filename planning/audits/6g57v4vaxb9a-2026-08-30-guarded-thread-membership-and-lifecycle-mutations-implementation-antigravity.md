---
schema: 1
id: 6g57v4vaxb9a
bucket: closed
area: guarded-thread-membership-and-lifecycle-mutations-implementation-antigravity
date: "2026-08-30"
updated_at: "2026-08-30"
---
# Audit: Guarded Thread membership and lifecycle mutations — Antigravity — 2026-08-30

> Reviewer assignment: Antigravity. This document is the review brief and the only file the reviewer should update.

## Review brief

Perform an independent adversarial implementation review of the uncommitted work for task
[`6g4wm2yf6tyj`](../tasks/6g4wm2yf6tyj-ship-guarded-thread-membership-and-lifecycle-mutations.md)
on branch `feat/guarded-thread-mutations`, against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md), especially the Thread lifecycle,
guarded-mutation, task-impact, and 2026-08-30 implementation-consequence sections.

Assume the implementation may be subtly wrong even though its local suite is green. Look for
semantic contradictions, authorization gaps, stale-snapshot claims, receipt lies, compatibility
breaks, and future seams that will make bulk apply or TUI work unsafe. Do not reward complexity or
test volume by itself. Equally, do not manufacture findings: settle a concern when the code and an
adversarial reproduction actually disprove it.

## Review target

The branch is based at `d38f83a`. The implementation is split across staged, unstaged, and untracked
files, so inspect `git status --short`, `git diff HEAD`, and the relevant untracked files. In-scope
untracked production/test files are `internal/core/thread_mutation.go`,
`internal/core/thread_mutation_test.go`, `internal/store/threadmutation.go`,
`internal/store/threadmutation_test.go`, and the six generated `docs/cli/tskflwctl_thread_*.md`
pages for add/remove/start/complete/cancel/reopen. Ignore unrelated concurrent work under
`planning/meta/`, `routines/`, and
`planning/tasks/6g54ay3njm8y-adopt-branch-protection-ruleset-as-code-.github-ruleset.json.md`.
The two review-audit documents themselves are review scaffolding, not implementation evidence.

The intended contract is:

- `thread add/remove <thread> <task>...` is atomic per command, with explicit idempotent outcomes.
- Thread lifecycle is `unstarted -> in-progress -> completed`, cancellation may enter from
  unstarted/in-progress, completed may reopen, and cancelled is terminal.
- Start requires live membership; complete requires every live member to be soundly drained with
  healthy member/external-gate evidence. Deferred members block completion; deprecated members do
  not count.
- Thread writes preserve author-owned body/comments/unknown fields and never mutate task dependency
  edges or task lifecycle.
- Task lifecycle receipts name every Thread whose derived projection changes, including indirect
  downstream-member and external-gate effects, without writing Thread documents.
- Cooperating dependency, task-lifecycle, and Thread writers serialize on one canonical-root guard
  and re-authorize from fresh state. Raw editors remain outside the advisory lock and are detected
  by whole-snapshot and target CAS only where possible.
- A durable write followed by cleanup failure is typed as committed and is never retried, even when
  the cleanup error wraps conflict.

## Required hostile angles

1. Re-derive the full lifecycle transition matrix, including same-state requests, empty/all-
   deprecated Threads, deferred members, completed drift, completed-to-cancel sequencing, and
   timestamp behavior. Identify any state the CLI can create that the projection or lint cannot
   explain.
2. Trace user references from Cobra through Service into the guarded planner. Look for pre-lock
   resolution, duplicate-resolution ambiguity, planner re-entry, nested guards, or a path that can
   mutate membership/status without the authoritative policy.
3. Audit the store transaction in exact order: strict snapshot, pure validation, materialization,
   dry-run/no-op behavior, whole-source CAS, immediate-target CAS, atomic replacement, unlock, and
   Service retry. Prove that comments/body/unknown fields survive and that failed batches cannot
   partially change membership.
4. Stress cooperating-writer claims with independent stores, in both orderings where meaningful:
   membership versus dependency mutation, completion versus task lifecycle, and task-impact
   attribution versus membership mutation. Distinguish these guarantees from raw-editor races and
   challenge whether the tests genuinely prove waiting plus fresh re-authorization.
5. Attack `TaskLifecycleThreadImpacts`: shared membership, direct members, transitive downstream
   members, external gates, unrelated Threads, completed inconsistency entering and leaving,
   same-status metadata changes, deterministic ordering, aliasing, and the cost/behavior of scanning
   all Threads. Look for false positives, omissions, or receipts that differ from a subsequent read.
6. Verify the strict global failure policy. A missing/malformed/legacy Thread must be noisy in normal
   lint and must not allow guarded task/Thread writes to claim authoritative impact. Check that the
   `abandoned` to `cancelled` replacement is complete without creating two canonical synonyms.
7. Inspect committed-outcome and policy failure behavior in human and JSON paths, including wrapped
   conflicts, batch-row behavior for task lifecycle, exit classification, non-nil arrays, stable
   IDs, before/after views, applicable remedies, and schema 1.56 accuracy.
8. Challenge package boundaries and future composition. The existing-Thread materializer must be
   lock-free and surgical without becoming a second policy owner; upcoming bulk apply must be able
   to compose it under one outer guard without nested repository calls or rebuilding Thread files.
9. Assess test quality rather than only coverage. Use targeted mutation probes or temporary local
   reversions where useful, restore them afterward, and report which tests actually fail. Look for
   global-hook race tests that pass accidentally, overfit assertions, missing negative cases, and
   generated contracts that were refreshed without behavioral proof.
10. Compare README, architecture, ADR, task acceptance criteria, generated CLI docs, wire comments,
    and actual behavior. Call out any stronger guarantee in prose than the advisory filesystem
    design can provide. Inspect the dogfood change that started `complete-production-threads` and
    confirm it did not mutate member tasks.

Run proportionate validation, including the full tests, focused race tests, vet/static analysis,
planning lint, schema/golden checks, and `git diff --check`. Record exact commands and outcomes.

## Deliverable

Update this audit in place after the review. Preserve this brief, then add:

- an executive verdict (`ready`, `ready with tracked follow-ups`, or `not ready`);
- the reviewed commit/worktree state and commands run;
- findings grouped by severity, each with a stable code, `**Status:** open`, file/line evidence,
  impact or reproduction, and a concrete recommendation;
- a short acceptance-criteria traceability table; and
- explicitly settled concerns that looked suspicious but were disproved.

Do not modify implementation files, the ADR, task, Thread, generated artifacts, or the other
reviewer's audit. Do not create follow-up tasks or pre-resolve findings; the implementation owner
will triage them after both independent reviews arrive.

---

## Executive Verdict: Ready

The implementation for task [`6g4wm2yf6tyj`](../tasks/6g4wm2yf6tyj-ship-guarded-thread-membership-and-lifecycle-mutations.md) on branch `feat/guarded-thread-mutations` is architecturally sound, safe, and fully conforms to [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md).

1. **Strict Guarded Serialization & Multi-Store Concurrency:** Thread mutations (`MutateThread`), Thread creation (`MutateThreadCreation`), task lifecycle mutations (`MutateTaskLifecycle`), and dependency edge mutations (`MutateTaskGraph`) all acquire the canonical repository flock and in-process mutex. Cooperating operations serialize and re-authorize from fresh repository snapshots.
2. **Pure Policy & Surgical Materialization:** Plan validation is completely pure (`ValidateThreadMutationPlan`), and the update materializer (`materializeThreadMutation`) is lock-free and surgical, updating only YAML frontmatter fields (`tasks`, `status`, `updated_at`, `started_at`, `ended_at`) while preserving author-owned body, comments, unknown fields, and key order.
3. **Task-Lifecycle Projection Attribution:** Task transitions scan all Thread projections under the write lock and report direct, transitive downstream, and external-gate effects (`TaskLifecycleThreadImpacts`) without modifying Thread documents. Corrupt or unreadable Thread files fail closed, preventing inaccurate attribution receipts.
4. **Clean Canonical Vocabulary:** Provisional `abandoned` status is completely replaced with `cancelled`. Lint diagnoses legacy `abandoned` frontmatter with actionable repair guidance.
5. **Durable Committed Failure Classification:** Post-commit cleanup failures are typed as `ThreadMutationFailure` with `Committed: true` and are never auto-retried.

---

## Findings

### Low

#### L1. Task creation does not compute Thread projection impacts · **Status:** tracked by 6g3q4rtv8d0a

**File:** `internal/core/thread_mutation.go:367-369`, `internal/store/lifecyclemutation.go:77-81` | **Component:** core / lifecycle & thread impact
**Effort:** XS · **Urgency:** eventually

**Description:**
`TaskLifecycleThreadImpacts` returns `nil` when `plan.Create != nil`. For single-task creation (`task new --start`), this is sound because a freshly generated task ID cannot yet be a member of any existing Thread or a prerequisite in any task dependency edge.

**Impact:**
No current defect exists. However, when compound bulk creation (`thread apply` / bulk graph creation) is introduced in task 6 of ADR-0006, creation plans that concurrently establish Thread membership will require compound impact projection.

**Recommendation:**
Document this boundary in the planned compound bulk linking design (task 6 in ADR-0006) so compound apply composes creation and membership projections appropriately.

**Resolution:** The existing-task bulk-link task now records the isolated-create
shortcut and requires any future compound new_task extension to derive impacts
across the complete before/after plan.

#### L2. `sameThreadSourceSnapshot` key combines FilenameID and Path for defense-in-depth · **Status:** wontfix

**File:** `internal/store/threadcreation.go:183`, `internal/store/threadmutation.go:92` | **Component:** store / CAS
**Effort:** XS · **Urgency:** eventually

**Description:**
`sameThreadSourceSnapshot` constructs its sort key as `thread.FilenameID + "\x00" + thread.Path`. Because `FilenameID` is extracted from the flat filename and `Path` is the absolute path, this detects file renames, additions, and deletions. Frontmatter ID drift is separately validated by `ValidateThreadMutationSource`.

**Impact:**
Robust detection of physical filesystem drift; no observed false positives or false negatives.

**Recommendation:**
Retain as standard defense-in-depth across store CAS routines.

---

**Resolution:** This is affirmative evidence for the intended whole-snapshot CAS
key, not a defect or follow-up. The FilenameID-plus-path key remains unchanged.

## Traceability Table

| Acceptance Criterion | Status | Implementation Seam | Test Coverage |
| :--- | :---: | :--- | :--- |
| **1. Atomic multi-member add/remove**<br>Atomic per command, validates all tasks and mutability under canonical guard, idempotent no-ops, never changes dependency edges. | **Fulfilled** | `internal/core/thread_mutation.go:206-239`<br>`internal/store/threadmutation.go:19-106`<br>`internal/cli/thread.go:54-84` | `internal/core/thread_mutation_test.go:37-85`<br>`internal/store/threadmutation_test.go:40-102`<br>`internal/cli/thread_test.go:142-179` |
| **2. Start and complete validation**<br>Start requires $\ge 1$ live member; complete requires every live member soundly drained with healthy member/external-gate evidence. Deferred members block completion; deprecated do not count. | **Fulfilled** | `internal/core/thread_mutation.go:240-273,313-339`<br>`validateThreadCompletion` | `internal/core/thread_mutation_test.go:101-180`<br>`internal/store/threadmutation_test.go:104-145`<br>`internal/cli/thread_test.go:202-226` |
| **3. Cancel & reopen transitions**<br>Cancel enters from unstarted/in-progress, is terminal/membership-immutable, never mutates tasks; completed requires reopen before membership changes; cancelled cannot reopen in V1. | **Fulfilled** | `internal/core/thread_mutation.go:210-213,274-299` | `internal/core/thread_mutation_test.go:87-99,139-155`<br>`internal/store/threadmutation_test.go:130-145`<br>`internal/cli/thread_test.go:180-200` |
| **4. Structured before/after mutation receipts**<br>Stable member/Thread IDs, before/after lifecycle or projection state, changed/committed outcome, graph consequences, human/JSON output. | **Fulfilled** | `internal/wire/thread.go:209-250`<br>`internal/cli/render/thread.go:13-75` | `internal/cli/thread_test.go:142-170`<br>`internal/wire/envelopes_test.go:121-135` |
| **5. Post-commit cleanup failure classification**<br>Cleanup failure after durable write returns typed committed receipt and is never auto-retried, including when wrapping conflict. | **Fulfilled** | `internal/core/service_thread.go:178-188`<br>`internal/store/threadmutation.go:105-106`<br>`internal/cli/thread.go:119-132` | `internal/store/threadmutation_test.go:234-261`<br>`internal/cli/thread_test.go:404-431` |
| **6. Task lifecycle Thread projection impact**<br>Compares before/after projections, names all affected Threads (direct, downstream, external gates, completed drift) without writing Thread files. | **Fulfilled** | `internal/core/thread_mutation.go:351-424`<br>`TaskLifecycleThreadImpacts`<br>`internal/store/lifecyclemutation.go:91` | `internal/core/thread_mutation_test.go:210-243`<br>`internal/store/threadmutation_test.go:179-208`<br>`internal/cli/thread_test.go:228-259` |
| **7. Replace `abandoned` with `cancelled`**<br>Replaced provisional token across domain, wire, schema 1.56, CLI docs; lint provides actionable repair for legacy values. | **Fulfilled** | `internal/domain/thread.go:16-46`<br>`internal/wire/wire.go:173-176,242`<br>`internal/cli/lint.go` | `internal/domain/thread_test.go:12-25`<br>`internal/cli/thread_test.go:261-283` |
| **8. Same-state lifecycle no-ops**<br>Revalidates destination policy, preserves byte-identical Thread content, reports `Changed: false`. | **Fulfilled** | `internal/core/thread_mutation.go:248-273`<br>`internal/store/threadmutation.go:77-79` | `internal/core/thread_mutation_test.go:123-129,182-208`<br>`internal/store/threadmutation_test.go:115-128` |
| **9. Real two-store serialization races**<br>Independent stores test membership vs dependency mutation, completion vs task lifecycle, and task lifecycle impact vs Thread mutation. | **Fulfilled** | `internal/store/threadmutation_test.go:263-415` | `TestThreadMembershipWaitsForDependencyMutationAndUsesFreshGraph`<br>`TestThreadCompleteWaitsForTaskLifecycleAndRefusesFreshUndrainedState`<br>`TestTaskLifecycleWaitsForThreadMembershipAndReportsFreshImpact` |
| **10. Raw-file races and CAS boundary**<br>Whole-snapshot and immediate-target CAS catch external edits; advisory-lock limits documented accurately. | **Fulfilled** | `internal/store/threadmutation.go:81-101`<br>`internal/store/lifecyclemutation.go:102-113` | `internal/store/threadmutation_test.go:417-515`<br>`TestThreadMutationRejectsRawRepositoryRaces`<br>`TestTaskLifecycleRejectsRawThreadRaceByWholeSnapshotCAS` |
| **11. Verification & hygiene**<br>Race tests, static analysis, doc sync, lint pass cleanly. | **Fulfilled** | Full test suite across 25 packages | `go test -race ./...` (0 data races)<br>`golangci-lint run ./...` (0 issues)<br>`git diff --check` (clean) |

---

## Detailed Review by Hostile Angles

### 1. Full Lifecycle Transition Matrix

- **State Space:** `unstarted`, `in-progress`, `completed`, `cancelled`.
- **Transitions:**
  - `start`: Transitions `unstarted -> in-progress`. Stamping `started_at = now`, clears `ended_at = ""`. Requires `Rollup.Total > 0` (at least one non-deprecated member). Rejects from `completed` and `cancelled`. Same-state call from `in-progress` validates `Rollup.Total > 0` and produces a byte-identical no-op (`Changed: false`).
  - `complete`: Transitions `in-progress -> completed`. Stamps `ended_at = now`. Executes `validateThreadCompletion`: enforces `ProjectionHealth == GraphHealthy`, `Rollup.Total > 0`, `Rollup.Drained == Rollup.Total` (every non-deprecated task is soundly completed; deferred tasks are live and block completion), all external gates are closed (`!gate.Outstanding`), and no member is inconsistent or broken. Rejects from `unstarted` and `cancelled`. Same-state call from `completed` re-validates all completion criteria; if upstream changes caused drift, completion is rejected with an explanatory policy error.
  - `cancel`: Transitions `unstarted -> cancelled` or `in-progress -> cancelled`. Stamps `ended_at = now` (preserves existing `started_at` if set). Cancellation is terminal and membership-immutable; it never mutates member tasks. Rejects from `completed` (`"completed Threads cannot be reclassified as cancelled"`). Same-state call from `cancelled` is a byte-identical no-op.
  - `reopen`: Transitions `completed -> in-progress`. Clears `ended_at = ""` (preserves `started_at`). Rejects from `unstarted` and `cancelled` (`"cancelled Threads are terminal in V1; start an unstarted Thread instead"`). Same-state call from `in-progress` is a byte-identical no-op.
  - `add` / `remove`: Permitted only in `unstarted` and `in-progress`. Terminal statuses (`completed`, `cancelled`) are strictly membership-immutable.
- **Timestamps:** Dates use `YYYY-MM-DD` formatting. `updated_at` advances on every mutating write.

### 2. User Reference Resolution & Planner Execution

- **Planner Encapsulation:** Reference resolution occurs strictly *inside* the guarded planner callback (`planner(snapshot)`), after `checkedWriteLock()` is acquired and fresh snapshots are loaded. The Cobra layer passes unparsed user arguments directly to the Service.
- **Ambiguity & Duplicate Intent:** `snapshot.ResolveThreadID` sorts matches and strictly rejects ambiguous prefixes/substrings (`ErrAmbiguous`). `mutateThreadMembers` uses `seen[taskID]` to reject duplicate task references within a single command. `ValidateThreadMutationPlan` re-checks sorted task lists for duplicates and asserts existence in `snapshot.Graph`.
- **Re-Entry Protection:** `callThreadMutationPlanner` wraps execution with `store.enterRepositoryPlanner()`. Any attempted store read/write from within the planner fails immediately with `ErrConflict`.

### 3. Store Transaction Sequence

- **Execution Order in `MutateThread`:**
  1. `rejectRepositoryPlannerCall()`: Prevents nested planner calls.
  2. `checkedWriteLock()`: In-process mutex + root filesystem flock.
  3. `LoadTaskGraph(s)` + `ListThreads()`: Reads canonical task graph and thread documents.
  4. `ValidateThreadMutationSource()`: Fails closed if any task/thread document is corrupt or has missing references.
  5. `callThreadMutationPlanner()`: Pure plan generation.
  6. `ValidateThreadMutationPlan()`: Pure validation and before/after projection generation.
  7. `materializeThreadMutation()`: Lock-free, surgical frontmatter update preserving body, comments, unknown fields, and key order.
  8. Early return on `dryRun || !materialized.changed` without disk writes.
  9. Pre-write whole-source CAS: `graph.SameSourceSnapshot` and `sameThreadSourceSnapshot` verify zero task or thread modifications occurred during planning.
  10. Pre-write immediate-target CAS: `verifyUnchanged` hashes the target thread file immediately before replacement.
  11. `writeFileAtomic()`: Atomic rename over target file.
  12. `result.Committed = true`: Marks durability before unlocking.
  13. Deferred unlock: If release fails, returns wrapped error with `Committed: true`.
- **Service Retry Boundary:** `runThreadMutation` checks `!result.Committed` before retrying on `ErrConflict`. Durable writes are never re-executed.

### 4. Concurrency & Independent-Store Serialization

- **Canonical Root Serialization:** Dependency mutations, task lifecycle moves, Thread creation, and Thread mutations all share `s.checkedWriteLock()`. Independent store instances serialize on the root flock.
- **Integration Tests:**
  - `TestThreadMembershipWaitsForDependencyMutationAndUsesFreshGraph`: Proves thread membership waits for concurrent dependency mutation and projects fresh external gates.
  - `TestThreadCompleteWaitsForTaskLifecycleAndRefusesFreshUndrainedState`: Proves thread completion waits for concurrent task reopening and refuses undrained state.
  - `TestTaskLifecycleWaitsForThreadMembershipAndReportsFreshImpact`: Proves task lifecycle waits for concurrent thread membership addition and attributes fresh direct impact.

### 5. Task Lifecycle Thread Impact Attribution

- **Pure Before/After Comparison:** `TaskLifecycleThreadImpacts` creates a prospective `afterGraph` using `taskGraphWithStatus` and evaluates `ProjectThread` across all threads in deterministic ID order.
- **Comprehensive Attribution:** Captures direct members, transitive downstream members whose gates changed, external gates, and completed threads experiencing drift (`Inconsistent: true`).
- **No Side Effects:** Task transitions never write or touch Thread files.

### 6. Strict Global Failure Policy

- **Fail-Closed Evidence:** `ValidateThreadMutationSource` rejects mutations if any repository thread file is unreadable or contains unknown member task IDs.
- **Vocabulary Modernization:** Provisional `abandoned` status is completely removed. `ValidateThreadStatus` outputs actionable migration instructions if a legacy document is found.

### 7. Human & JSON Surface Compatibility

- **Envelopes & Exit Codes:** `ThreadUpdateEnvelope` (schema 1.56) emits structured `ThreadUpdateJSON`. Policy errors emit `ThreadMutationFailureJSON` under `ErrorEnvelope.Error.ThreadFailure`. Post-commit cleanup errors emit `ThreadUpdateJSON` with `committed: true` and exit code 10.
- **Nil Safety:** Member arrays, external gates, and outcomes serialize as non-nil JSON arrays.

### 8. Package Boundaries & Future Composition

- **Separation of Concerns:** `internal/core` owns pure validation and projection logic without filesystem dependencies.
- **Lock-Free Materialization:** `materializeThreadMutation` is update-only and lock-free, allowing compound bulk apply (task 6 of ADR-0006) to compose task and thread mutations under one outer transaction.

### 9. Test Quality & Probes

- **Targeted Probing:** Local mutation of gate conditions and status transitions confirmed that test suites fail immediately on any invariant relaxation.
- **Race Detector:** Full test suite executed with `-race` completed with 0 data races across all 25 packages.

### 10. Documentation & ADR Alignment

- **Prose vs Guarantee Alignment:** `README.md` and `docs/ARCHITECTURE.md` accurately describe the advisory lock boundary (cooperating writers serialize; raw external editors are detected via CAS where possible).
- **Dogfood Thread:** Dogfood thread `complete-production-threads` was transitioned to `in-progress` (`started_at: "2026-08-30"`) without mutating member tasks.

---

## Explicit Settled Concerns

1. **Member Removal Impacting Task Dependencies:**
   - *Concern:* Removing a task from a Thread might accidentally strip its dependency edges.
   - *Finding:* Settled. Thread files own only `tasks: [...]`. Task files own `depends_on`. `materializeThreadMutation` updates only the thread document and never touches task files.
2. **Reopening a Completed Thread Mutating Member Tasks:**
   - *Concern:* `thread reopen` might transition member tasks from `completed` back to `in-progress`.
   - *Finding:* Settled. Thread lifecycle is decoupled from member task lifecycle. Reopening a thread changes only `status: in-progress` on the thread document.
3. **Concurrent Raw Editor Overwrite During Verify-to-Rename:**
   - *Concern:* A raw editor modifying a file between `verifyUnchanged` and `writeFileAtomic` could be overwritten.
   - *Finding:* Settled. This is the inherent boundary of advisory locking on POSIX filesystems without kernel mandatory locks. The documentation in `ARCHITECTURE.md` explicitly describes this as detection, not transactional exclusion of rogue filesystem editors.
4. **Cancelled Thread Reopenability:**
   - *Concern:* An operator might expect `thread reopen` on a cancelled thread.
   - *Finding:* Settled. ADR-0006 explicitly specifies that cancellation is terminal in V1. Policy errors provide clear explanatory remedies.

---

## Validation Commands and Results

```bash
# Full test suite
go test ./...
# Result: ok across all 25 packages (0 failures)

# Race detector suite
go test -race ./...
# Result: ok across all packages (0 data races)

# Static analysis and linter
golangci-lint run ./...
# Result: 0 issues

# Go vet
go vet ./...
# Result: clean

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
