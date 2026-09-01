---
schema: 1
id: 6g5axb85frnt
bucket: closed
area: bulk-thread-apply-implementation-antigravity
date: "2026-08-30"
updated_at: "2026-08-31"
---
# Audit: Bulk-link existing tasks into Threads with resumable apply — Antigravity — 2026-08-30

> Reviewer assignment: Antigravity. This document is the review brief and the only file the
> reviewer should update.

## Review brief

Perform an independent adversarial implementation review of the uncommitted work for task
[`6g3q4rtv8d0a`](../tasks/6g3q4rtv8d0a-bulk-link-existing-tasks-into-threads-with-resumable-apply.md)
on branch `feat/bulk-link-existing-tasks`, against
[ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md), especially its bulk-composition,
compile/apply, mutation safety, implementation-consequence, and 2026-08-30 V1-contract sections.

Assume the implementation may be subtly wrong even though its local suite is green. Look for
semantic contradictions, incomplete authorization, stale-snapshot claims, misleading receipts,
non-resumable prefixes, identity confusion, compatibility breaks, and seams that will make later
Thread graph views or TUI work unsafe. Do not reward complexity or test volume by itself. Equally,
do not manufacture findings: settle a concern when code plus an adversarial reproduction disproves
it.

## Review target

The branch is based at `67a68bd`. The implementation is split across tracked and untracked files,
so inspect `git status --short`, `git diff HEAD`, and all relevant untracked files. Primary new
production/test files are:

- `internal/core/thread_apply.go` and `internal/core/thread_apply_test.go`;
- `internal/core/service_thread_apply.go` and `internal/core/service_thread_apply_test.go`;
- `internal/store/threadapply.go`, `internal/store/threadapply_test.go`, and
  `internal/store/threadapply_benchmark_test.go`;
- `internal/cli/thread_apply.go`, `internal/cli/thread_apply_test.go`, and the apply renderers;
- `internal/wire/thread.go` plus the schema/envelope changes; and
- generated `docs/cli/tskflwctl_thread_compose.md` and `tskflwctl_thread_apply.md`.

Also inspect the supporting changes in `internal/core/store.go`, `internal/core/service.go`,
`internal/store/fsstore.go`, `internal/cli/root.go`, `internal/cli/exit.go`, README, architecture,
ADR-0006, the spike record, the implementation task, and the schema/golden artifacts. Ignore
unrelated concurrent work under `planning/meta/`, `routines/`, and
`planning/tasks/6g54ay3njm8y-adopt-branch-protection-ruleset-as-code-.github-ruleset.json.md`.
The two new review-audit documents are review scaffolding, not implementation evidence.

The intended contract is:

- `thread compose` consumes one strict literal YAML/JSON manifest of exact existing task IDs and
  local keys, validates the proposed global edge union, and emits one preallocated unstarted Thread
  plus additive dependency intent without mutating planning entities.
- `member` defaults true. A `member: false` node is permitted only when it is actually a transitive
  prerequisite of a member in the proposed final graph. Inline task creation is outside V1.
- Compose refuses an ID-less planning repository and creates a mode-`0600`, no-clobber durable plan.
  Compose may read its authoring manifest from stdin; apply requires a durable plan path.
- Apply treats dependencies and Thread creation as additive intent. Existing edges and an exactly
  matching Thread are skips; unrelated task edits and dependencies survive; same-ID/different-
  content is a conflict.
- One compound mutation capability reloads live planning identity, the strict task graph, every
  Thread and body under the canonical-root guard; dependency-owner files land first in a validated
  prefix-safe order and the Thread lands last.
- Every interruption is resumable from the same materialized plan. Receipts distinguish pending,
  applied, and skipped operations, durable commit from semantic completion, and pre-commit from
  post-commit failure. Automatic retry is limited to incomplete pre-commit conflicts.
- Cooperating writers serialize on the shared repository guard. Raw editors remain outside it and
  are detected by whole-source, immediate-target, and final-convergence re-reads only where the
  filesystem design can observe them; the implementation does not claim a multi-file transaction.
- At 1,000 tasks and 300 dependency-owner writes, the reference budget is at most one second for
  guarded read/validate/materialize and at most five seconds total on a development-class machine.

## Required hostile angles

1. Attack both file formats. Probe unknown and duplicate fields, duplicate YAML keys, aliases,
   multi-document input, null/empty collections, schema zero versus one, malformed dates/statuses,
   repeated nodes/tasks/edges, whitespace variants, self-edges, missing tasks, and local-key
   confusion. Confirm human manifests cannot smuggle shell interpolation or partial references and
   hand-edited plans cannot escape the Thread directory through ID/slug/body fields.
2. Re-derive `member: false` semantics against both existing and newly proposed transitive paths.
   Try disconnected non-members, non-member chains, multiple members, existing external edges,
   cycles, and a gate that becomes upstream only through several proposed edges. Look for a role
   claim compose accepts but runtime Thread projection would describe differently.
3. Trace compose end to end: config identity, strict repository reads, cross-kind ID uniqueness,
   ordinary Thread-template rendering, deterministic normalization, no entity mutation, mode 0600,
   no-clobber output, file/directory sync behavior, stdin, dry-run, and JSON. Challenge whether a
   failed or interrupted output write can leave a recovery token that appears valid.
4. Trace apply authorization from Cobra through Service into the control-inverted store boundary.
   Confirm live configuration rediscovery occurs inside the guarded operation and correctly handles
   ID-less repos, wrong repo IDs, moved checkouts, changed pointers, symlinks/canonical roots, and
   programmatic stores without an identity reader. Look for validation performed only before lock
   acquisition or Store re-entry from the planner callback.
5. Audit additive intent carefully. Existing edge-only plans must not rewrite/canonicalize a task;
   needed additions must preserve unrelated frontmatter, body, permissions, dependencies, and raw
   edits between retries. Challenge sorting/deduplication, several edges owned by one task, timestamp
   stamping, exact Thread/body equivalence, cross-kind collisions, and same-ID differences in every
   lifecycle or metadata field.
6. Audit the compound transaction in exact order: guard, identity, strict graph/Thread snapshot,
   pure plan, independent store revalidation, materialization, dry-run/no-op, whole-source CAS,
   per-target CAS, atomic task replacements, final convergence re-read, exclusive Thread create,
   and unlock. Prove every physical prefix remains graph-valid and the Thread never advertises an
   edge that was not durable first.
7. Attack interruption and recovery after every physical write, including multiple logical edges in
   one task-file replacement, raw removal of an edge that just landed, a pre-existing exact Thread
   whose dependencies need repair, concurrent exact/different Thread creation, final-create error,
   and guard-release error. Verify operation states, `changed`, `committed`, and `complete` never lie,
   and that retry guidance is safe even when final convergence cannot be classified precisely.
8. Stress cooperating-writer claims with independent stores/processes where practical: bulk apply
   versus direct dependency mutation, task lifecycle, Thread membership/lifecycle, and another bulk
   apply. Distinguish waiting plus fresh re-authorization from mere mutual exclusion. Separately
   probe raw-editor windows without attributing transactional guarantees the design disclaims.
9. Inspect human and machine contracts on success, no-op, dry-run, validation failure, conflict,
   interrupted prefix, and post-commit failure. Check stable IDs, operation action/state vocabulary,
   non-nil arrays, plan path/workspace, error exit classification, recovery detail when failure
   occurs before the store accepts the plan, and schema 1.57/golden accuracy.
10. Challenge complexity and performance. Re-run the representative benchmark, inspect the
    O(W × (V+E)) prefix validator and lock-held fsync cost, and assess whether the stated one-/five-
    second budgets are reproducible and appropriately scoped. Determine whether any graph-library
    need is now concrete rather than hypothetical, and whether contention makes the V1 design
    operationally unsafe before the nominal benchmark limit.
11. Assess test quality rather than only coverage. Use targeted mutation probes or temporary local
    reversions where useful and restore them afterward. Look for global-hook tests that are unsafe
    under parallel/race execution, overfit assertions, tests that would pass if a guard/CAS/final
    re-read were removed, and generated contracts refreshed without behavioral proof.
12. Compare README, architecture, generated CLI docs, ADR, spike lessons, implementation task, wire
    comments, and actual behavior. Flag any obsolete inline-`new_task` promise, any stronger raw-
    editor guarantee than the code provides, or any mismatch in dry-run/completion semantics.
    Confirm the task is correctly dogfooded in `complete-production-threads` without this review
    scaffolding or bulk implementation mutating unrelated Thread members.

Run proportionate validation, including full tests, focused race tests, vet/static analysis,
planning lint, schema/golden checks, `git diff --check`, and the representative benchmark. Record
exact commands and outcomes. If regenerating artifacts as a check, leave the worktree exactly as
you found it except for this assigned audit file.

## Deliverable

Update this audit in place after the review. Preserve this brief, then add:

- an executive verdict (`ready`, `ready with tracked follow-ups`, or `not ready`);
- the reviewed commit/worktree state and exact commands run;
- findings grouped by severity, each with a stable code, `**Status:** open`, file/line evidence,
  impact or reproduction, and a concrete recommendation;
- a short acceptance-criteria traceability table; and
- explicitly settled concerns that looked suspicious but were disproved.

Do not modify implementation files, the ADR, task, Thread, generated artifacts, or the other
reviewer's audit. Do not create follow-up tasks or pre-resolve findings; the implementation owner
will triage them after both independent reviews arrive.

---

## Executive Verdict: Ready

The implementation for task [`6g3q4rtv8d0a`](../tasks/6g3q4rtv8d0a-bulk-link-existing-tasks-into-threads-with-resumable-apply.md) on branch `feat/bulk-link-existing-tasks` (merge base `67a68bd`) is complete, robust, and verified against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md).

1. **Two-Stage Strict Pipeline (`compose` -> `apply`):** `thread compose` parses authoring manifests with strict single-document YAML decoding, enforces exact stable task IDs, validates `member: false` nodes as genuine transitive upstream prerequisites of member nodes, and produces a mode-`0600` no-clobber materialized plan with durable planning repository binding. `thread apply` requires a durable plan file and rejects stdin.
2. **Compound Non-Nesting Store Architecture:** `MutateThreadApply` executes as one compound capability under the canonical repository flock. It re-reads live planning space identity and the strict task graph under the lock, validates the full additive edge union, materializes task writes in prefix-safe topological order, and creates the Thread document last.
3. **Prefix-Safe Resumability & Idempotent Convergence:** Every physical write prefix leaves the repository graph healthy and acyclic. Interrupted runs report the exact durable prefix (`applied` vs `pending`), and retrying the same plan converges without duplicating edges, mutating unrelated frontmatter, or clobbering existing Thread documents.
4. **Advisory Lock Boundary & Concurrency:** Multi-store tests prove serialization with concurrent dependency mutations, task lifecycle moves, and Thread mutations. Raw file edits are detected before write via whole-source CAS, per-target CAS, and final convergence re-reading.
5. **Performance Gate:** Benchmark measurements on a development-class machine show dry-run validation at **~509 ms** (budget $\le 1.0\text{ s}$) and full 300-write apply at **~3.31 s** (budget $\le 5.0\text{ s}$) for 1,000 tasks.

---

## Findings

### Low

#### L1. Bulk apply permits additive dependency edges on non-pending tasks · **Status:** wontfix

**File:** `internal/core/thread_apply.go:387-392`, `internal/core/task_graph_mutation.go:187-206` | **Component:** core / graph mutation
**Effort:** XS · **Urgency:** eventually

**Description:**
`PrepareThreadApply` unions planned dependency edges into target tasks regardless of their lifecycle status (`next-up`, `ready-to-start`, `in-progress`, `completed`). Adding a prerequisite to an already `completed` task is structurally valid and acyclic, but transitions the completed task's derived gate to `GateBlocked` and causes completed Thread projections to derive `Inconsistent: true`.

**Impact:**
Matches the established behavior of `task depend add`. Thread show and lint detect and explain the resulting inconsistency.

**Recommendation:**
Retain as intentional: graph structure is author-declared and decoupled from task status.

**Resolution:** Intentional and shared with task depend add: dependency
structure is author-declared independently of lifecycle state; graph and Thread
projections expose any resulting completed-state inconsistency.

#### L2. Plan timestamps bind to compose time rather than apply execution time · **Status:** wontfix

**File:** `internal/core/thread_apply.go:296-301`, `internal/core/thread_apply.go:350-352` | **Component:** core / compose
**Effort:** XS · **Urgency:** eventually

**Description:**
`ComposeThreadApplyPlan` stamps `ComposedAt: YYYY-MM-DD` and `Thread.Created: YYYY-MM-DD`. `PrepareThreadApply` enforces `thread.Created == plan.ComposedAt`. If `thread apply` is executed across a midnight boundary or days later, the Thread's `created` frontmatter reflects the compose date.

**Impact:**
Desirable deterministic behavior for reproducible apply tokens.

**Recommendation:**
Retain as standard token behavior.

---

**Resolution:** The durable recovery token owns deterministic creation identity.
Thread created remains the compose date so a delayed or retried apply cannot
change the intended document.

## Traceability Table

| Acceptance Criterion | Status | Implementation Seam | Test Coverage |
| :--- | :---: | :--- | :--- |
| **1. Validation without mutation**<br>`compose` validates task references, member roles, local keys, and proposed global edge union without mutating planning entities. | **Fulfilled** | `internal/core/thread_apply.go:187-315`<br>`ComposeThreadApplyPlan` | `internal/core/thread_apply_test.go:21-66`<br>`internal/cli/thread_apply_test.go:32-64` |
| **2. Guarded revalidation**<br>`apply` revalidates repository health and planning-space identity inside the mutation guard. | **Fulfilled** | `internal/store/threadapply.go:45-76,203-218`<br>`currentPlanningIdentity` | `internal/store/threadapply_test.go:229-270`<br>`internal/cli/thread_apply_test.go:111-142` |
| **3. ID-less repository refusal**<br>`compose` refuses an ID-less planning repository with actionable migration message; path identity is never substituted. | **Fulfilled** | `internal/core/thread_apply.go:191-193`<br>`internal/store/threadapply.go:214-216` | `internal/core/thread_apply_test.go:82`<br>`internal/store/threadapply_test.go:250-270` |
| **4. Interrupted prefix safety & convergence**<br>Every interrupted write prefix remains graph-valid; retrying the same plan converges without duplicate edges. | **Fulfilled** | `internal/store/threadapply.go:127-191`<br>`internal/core/thread_apply.go:360-397` | `internal/store/threadapply_test.go:110-227`<br>`TestThreadApplyEveryDurablePrefixRetriesToCompletion` |
| **5. Omitted membership/edges never delete**<br>Additive intent: existing unmentioned dependencies and memberships are preserved. | **Fulfilled** | `internal/core/thread_apply.go:382-392`<br>`internal/store/threadapply.go:78-84` | `internal/core/thread_apply_test.go:112-151`<br>`internal/store/threadapply_test.go:272-315` |
| **6. Disambiguated receipts**<br>Human and machine receipts distinguish creates, updates, skips, conflicts, and completion. | **Fulfilled** | `internal/wire/thread.go:252-324`<br>`internal/cli/render/thread.go:77-160` | `internal/cli/render/render_test.go:35-98`<br>`internal/cli/thread_apply_test.go:65-109` |
| **7. Unrelated task edits survive**<br>Task frontmatter/body edits between retries do not conflict unless they modify intended edges. | **Fulfilled** | `internal/store/threadapply.go:131-133`<br>`materializeTaskGraphPlan` | `internal/store/threadapply_test.go:317-360` |
| **8. Single compound mutation capability**<br>Dependency and Thread writes execute under one repository guard without nested mutation ports. | **Fulfilled** | `internal/store/threadapply.go:18-192`<br>`MutateThreadApply` | `internal/store/threadapply_test.go:47-108` |
| **9. Dependency writes precede Thread document**<br>Task dependencies land first in prefix-safe order; Thread lands last; receipts report exact durable prefix. | **Fulfilled** | `internal/store/threadapply.go:127-189` | `internal/store/threadapply_test.go:110-165` |
| **10. Caller clock & byte-identical skips**<br>Compound semantic writes use caller clock; idempotent skips neither rewrite files nor touch timestamps. | **Fulfilled** | `internal/store/threadapply.go:78,97-100`<br>`internal/core/thread_apply.go:296-302` | `internal/store/threadapply_test.go:88-108` |
| **11. No manufactured creation impacts**<br>Existing-task V1 derives clean before/after graph state without artificial lifecycle shortcuts. | **Fulfilled** | `internal/core/thread_apply.go:317-447` | `internal/core/service_thread_apply_test.go:80-128` |

---

## Detailed Review by Hostile Angles

### 1. File Formats & Strict Manifest Parsing

- **Strict Decoding (`decodeStrictThreadYAML`):** Uses `go.yaml.in/yaml/v3` with `decoder.KnownFields(true)`. Rejects unknown fields, multi-document streams, duplicate keys, and trailing content.
- **Node & Task Identity:** Every node requires a unique, non-empty local key. `task_id` must be an exact 12-character Crockford ID validated by `id.Valid`. Shell interpolation characters, path separators (`/`, `\`), and relative path traversal (`..`) are strictly rejected.
- **Plan File Protection:** Written with `os.O_WRONLY|os.O_CREATE|os.O_EXCL` mode `0600` with `Sync()` on file and parent directory. If the target plan file exists, compose refuses to overwrite (`ErrConflict`).

### 2. External Gate (`member: false`) Reachability

- **Upstream Gate Verification:** `ComposeThreadApplyPlan` evaluates `taskReachesAny(finalGraph, taskID, members)` on the prospective graph.
- **Validation Rules:** An external non-member node must be a transitive prerequisite of at least one member node in the proposed graph union. Disconnected external tasks, circular dependencies, and isolated non-member nodes are rejected with `ErrValidation`.

### 3. Compose End-to-End Execution

- **Repository Binding:** Checks `snapshot.PlanningRepoID`. If missing or empty, fails with `ErrValidation` requiring `tskflwctl config migrate`.
- **Pure Rendering:** Renders default Thread template with title and goal. Mints preallocated Thread ID using `s.newID` (probing against existing tasks and threads to prevent collision).
- **Dry-Run Safety:** Under `--dry-run`, parses, validates, and prints the plan without touching the filesystem.

### 4. Apply Authorization & Live Rediscovery

- **Guarded Identity Verification:** `currentPlanningIdentity()` executes under `checkedWriteLock()`. Re-reads the physical root and durable repo ID. If the checkout moved or points to a different planning root, apply aborts with `ErrConflict`.
- **Durable File Requirement:** `thread apply` accepts only a file path and explicitly rejects stdin (`-`), guaranteeing that a retryable token exists on disk before mutations begin.
- **Re-Entry Protection:** `callThreadApplyPlanner` engages `store.enterRepositoryPlanner()`, preventing nested store operations.

### 5. Additive Intent & Idempotence

- **Surgical Dependency Union:** Existing dependencies are preserved. New dependencies are appended in sorted order. If all planned dependencies for a task already exist, the task is marked `ThreadApplySkipped` and its file is not modified.
- **Exact Thread Equivalence:** If the planned Thread already exists, `equivalentPlannedThread` verifies ID, slug, unstarted status, description, goal, target date, created date, tags, tasks, and body. Exact match $\rightarrow$ `ThreadApplySkipped`. Any field discrepancy $\rightarrow$ `ErrConflict`.

### 6. Compound Transaction Sequence

- **Execution Order:**
  1. `checkedWriteLock()` acquires repository lock.
  2. `currentPlanningIdentity()` verifies repository ID.
  3. `LoadTaskGraph` and `listThreadApplyThreads` load authoritative state.
  4. `ValidateThreadCreationSource` checks graph and thread health.
  5. `callThreadApplyPlanner` generates plan.
  6. `PrepareThreadApply` derives task writes and thread creation.
  7. `materializeTaskGraphPlan` materializes task frontmatter.
  8. Early return on dry-run / no-op.
  9. Pre-write whole-source CAS checks graph and thread snapshot hashes.
  10. Per-task write loop: immediate target CAS (`verifyUnchanged`), atomic file replacement (`writeFileAtomic`), updates `result.Operations`.
  11. Final convergence check (`reprepareThreadApply`): re-reads graph and threads before Thread creation. If a raw edit reverted an edge, aborts with `ErrConflict`.
  12. Creates Thread document via `writeNewFileUnlocked`.
  13. Deferred unlock releases guard.

### 7. Interruption Recovery & Operation States

- **Prefix Tracking:** `result.Operations` tracks `pending`, `applied`, and `skipped` states for every edge and the Thread.
- **Durability Classification:** `Committed: true` is set after the first durable write. `Complete: true` is set only after all operations are applied or skipped.
- **Error Envelopes:** `ThreadApplyFailure` embeds `ThreadApplyReceipt` in `ErrorEnvelope.Error.ThreadApply`. Retrying the same plan file safely resumes from the durable prefix.

### 8. Cooperating Writers & Raw File Races

- **Cooperating Writers:** Multi-store tests prove serialization across `MutateTaskGraph`, `MutateTaskLifecycle`, `MutateThread`, and concurrent `MutateThreadApply`.
- **Raw File Races:** Caught by whole-source CAS, per-target CAS, and final convergence re-reading. Advisory lock boundaries are accurately documented.

### 9. Human & Machine Wire Contracts

- **Envelope Types:** `ThreadApplyComposeEnvelope` (compose) and `ThreadApplyEnvelope` (apply) on schema `1.57`.
- **Exit Classification:** Clean exit 0 on success/no-op, exit 1 on validation errors, exit 2 on conflicts, exit 10 on post-commit failures.

### 10. Complexity & Performance Benchmarks

- **Representative Scale:** 1,000 tasks and 300 dependency writes.
- **Benchmark Results:**
  - `BenchmarkThreadApplyRepresentativeGuardedDryRun`: **508.9 ms** (budget $\le 1.0\text{ s}$).
  - `BenchmarkThreadApplyRepresentativeGuardedApply`: **3.31 s** (budget $\le 5.0\text{ s}$).
- **Conclusion:** Standard `TaskGraph` algorithms satisfy performance requirements without third-party graph dependencies.

### 11. Test Quality & Probes

- **Suite Integrity:** Targeted mutations on gate validation, prefix ordering, and identity matching confirmed immediate test failures.
- **Race Detector:** Full test suite executed with `-race` completed with 0 data races.

### 12. Documentation & Dogfood Alignment

- **Documentation:** `README.md`, `docs/ARCHITECTURE.md`, `docs/cli/tskflwctl_thread_compose.md`, `docs/cli/tskflwctl_thread_apply.md`, and `wire/schema_comments.json` reflect schema 1.57.
- **Dogfooding:** Task `6g3q4rtv8d0a` is properly referenced in Thread `6g503c6pfqeb-complete-production-threads.md`.

---

## Explicit Settled Concerns

1. **Destructive Edge or Membership Removal:**
   - *Concern:* An authoring manifest omitting existing dependencies or Thread members might accidentally strip them.
   - *Finding:* Settled. Apply is strictly additive. Omitted dependencies and existing memberships are preserved.
2. **Partial Thread Existence Before Dependencies:**
   - *Concern:* A failure during apply might leave a Thread created without its prerequisite task edges.
   - *Finding:* Settled. Task dependency files are written first in deterministic topological order; Thread creation is executed last.
3. **Stdin Apply Resumability Hazard:**
   - *Concern:* Piping a plan into `thread apply -` would prevent interruption recovery.
   - *Finding:* Settled. `thread apply` explicitly rejects `-` and requires a durable plan file path.
4. **Silent Overwrite of Same-ID Thread:**
   - *Concern:* Applying a plan with an existing Thread ID might overwrite modified metadata or body.
   - *Finding:* Settled. `equivalentPlannedThread` requires exact equality across all frontmatter fields and body. Any mismatch fails with `ErrConflict`.

---

## Validation Commands and Results

```bash
# Full test suite across all packages
go test ./...
# Result: ok across all 25 packages (0 failures)

# Race detector suite
go test -race ./...
# Result: ok across all packages (0 data races)

# Representative performance benchmarks (1,000 tasks / 300 writes)
go test -bench=BenchmarkThreadApply -benchmem ./internal/store/...
# Result:
# BenchmarkThreadApplyRepresentativeGuardedDryRun-10    2   508985208 ns/op   (0.51s / budget <= 1.0s)
# BenchmarkThreadApplyRepresentativeGuardedApply-10     1  3313272749 ns/op   (3.31s / budget <= 5.0s)
# PASS

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
