---
schema: 1
id: 6g4aj7v60syg
bucket: closed
area: ship-guarded-dependency-mutations-and-graph-queries
date: "2026-08-27"
updated_at: "2026-08-27"
---

# Audit: Ship Guarded Dependency Mutations and Graph Queries — 2026-08-27

Adversarial pre-implementation readiness review of planned task `6g3q4rt7mgjn` (`ship-guarded-dependency-mutations-and-graph-queries`, epic `30-threads-and-task-dependency-graphs`), evaluated against ADR-0006, `docs/ARCHITECTURE.md`, and the merged canonical read (`6g3q4rst78qy`) and portable mutation guard (`6g3q4rt0wzkq`) foundations on `main`.

**Executive Verdict: Ready with amendments.** The merged foundations on `main` provide a complete and verified substrate: `core.LoadTaskGraph` unifies strict snapshot loading; `store.FS.MutateTaskGraph` provides atomic critical-section isolation and whole-snapshot CAS; `core.ValidateTaskGraphMutationPlan` preserves planner write order while validating all intermediate prefixes and final health; and Tarjan SCC cycle detection correctly attributes multi-cycle feedback topologies. However, before implementation begins on `6g3q4rt7mgjn`, one pre-implementation blocker and two specification gaps must be amended in the task definition:
1. **In-memory reference resolution seam (H1):** `TaskGraph` only looks up tasks by exact 12-character ID. Because planners are strictly forbidden from calling `Store` methods (which fail with `ErrConflict`), the core service cannot resolve user-supplied slugs or prefixes inside the planner callback unless `TaskGraph` provides pure in-memory resolution (`ResolveTask` / `ResolveID`).
2. **Unspecified legacy migration command shape (M1):** The acceptance criteria mandate migrating the 6 live `blocked_by` tasks, but the task specification does not define the CLI command signature (e.g. `tskflwctl task depend migrate [--dry-run]`).
3. **Query Wire DTOs and flag defaults (M2):** DTO schemas for `task blockers` and `task unblocks` and the default projection behavior (action frontier vs causal closure) require explicit definition to avoid wire drift.

Findings are classified as **[pre-implementation blocker]**, **[amendment required]**, **[follow-up task]**, or **[monitor]**.

---

## Findings

### High

#### H1. `TaskGraph` lacks in-memory slug and prefix resolution for mutation planners · **Status:** tracked by 6g3q4rt7mgjn

**File:** internal/core/dependency_graph.go:729-734, internal/core/store.go:87-91, planning/tasks/6g3q4rt7mgjn-ship-guarded-dependency-mutations-and-graph-queries.md:24
**Component:** core / graph snapshot & resolution
**Effort:** S · **Urgency:** acute

**[pre-implementation blocker]**

CLI commands like `task depend add <task> --on <prerequisite>` receive human-provided input such as full slugs (`6fjangd7kvh0-alpha`), titles/slugs (`alpha`), short prefixes (`6fjang`), or exact 12-character IDs (`6fjangd7kvh0`).

Currently, `TaskGraph` only exposes `Task(taskID string) (domain.Task, bool)`, which requires an exact 12-character task ID. It has no in-memory resolver for slugs or prefixes.

Under the control-inverted mutation guard:
1. Planners run inside `MutateTaskGraph` with `plannerActive` enabled. Calling `s.store.GetTask` or `s.store.ResolveTaskPath` inside the planner is rejected with `ErrConflict` ("graph mutation planner cannot call Store methods").
2. If the service resolves slugs through the Store *before* calling `MutateTaskGraph`, a TOCTOU race exists where a task renamed, moved, or deleted concurrently results in a stale or invalid mutation.
3. Therefore, all reference resolution must happen *inside* the planner callback against the immutable `*core.TaskGraph` snapshot.

**Why it matters:** Without an in-memory resolution method on `TaskGraph`, implementers will either hit re-entry deadlocks/conflicts by calling Store methods, or introduce TOCTOU race conditions by pre-resolving outside the lock.

**Recommendation:** Add a pure in-memory `ResolveTask(ref string) (domain.Task, error)` and `ResolveID(ref string) (string, error)` method to `*core.TaskGraph` that matches exact IDs, exact slugs, and unique unambiguous prefixes against `g.tasks`, returning typed errors (`domain.ErrNotFound`, `domain.ErrAmbiguous`).

---

**Resolution:** Accepted and recorded in ADR-0006 plus task 6g3q4rt7mgjn. The
guarded snapshot gains pure task-reference resolution, refined to preserve the
existing exact ID/slug, case-insensitive prefix, and substring policy through
shared logic or parity tests; the suggested duplicate ResolveTask and ResolveID
surface is not prescribed.

### Medium

#### M1. Unspecified CLI subcommand signature and receipt structure for legacy dependency migration · **Status:** tracked by 6g3q4rt7mgjn

**File:** planning/tasks/6g3q4rt7mgjn-ship-guarded-dependency-mutations-and-graph-queries.md:25, 36, planning/adrs/0006-adopt-threads-as-task-dags.md:428-432
**Component:** cli / legacy migration
**Effort:** S · **Urgency:** acute

**[amendment required]**

Acceptance criterion 36 requires migrating the six live `blocked_by` fields from resolvable slugs to stable `depends_on` IDs with frontmatter/body preservation. However, neither the task scope nor ADR-0006 defines the exact CLI command path, flags, or output format.

Without an explicit specification:
1. The implementer must guess whether the command is `tskflwctl task depend migrate`, `tskflwctl task migrate-dependencies`, or an option under `lint --fix`.
2. Diagnostic lint messages (which currently report `"run the guarded migration"`) cannot cite an exact command name.
3. The machine receipt and `--dry-run` behavior for the migration remain underspecified.

**Recommendation:** Explicitly specify the CLI command signature as `tskflwctl task depend migrate [--dry-run] [--json]` in task `6g3q4rt7mgjn`. Define its receipt to list migrated task IDs, resolved edge additions, cleared legacy fields, and dry-run indicators.

**Resolution:** Accepted and recorded in ADR-0006 plus task 6g3q4rt7mgjn as
repository-wide task depend migrate. Global dry-run and JSON modes are
inherited, V1 has no per-task selector, and failures report a sound durable
prefix for idempotent retry rather than claiming an all-files transaction.

#### M2. Query Wire DTOs and default projection flags need explicit specification · **Status:** tracked by 6g3q4rt7mgjn

**File:** planning/tasks/6g3q4rt7mgjn-ship-guarded-dependency-mutations-and-graph-queries.md:33-35, 38-40, internal/wire/dto.go
**Component:** wire / cli / query projections
**Effort:** S · **Urgency:** soon

**[amendment required]**

The task defines two distinct blocker projections: the action-oriented `BlockingFrontier` (which stops at terminal blockers and unstarted leaves) and the forensic `CausalBlockers` (which traverses the full unsound prerequisite closure).

However, the task does not define:
1. The CLI flag default for `tskflwctl task blockers <task-id>` (whether it defaults to `--frontier` with a `--causal` / `--all` flag, or vice versa).
2. The typed JSON wire structures (`TaskBlockersJSON`, `TaskUnblocksJSON`) in `internal/wire/dto.go`.

**Recommendation:**
1. Define that `tskflwctl task blockers <task-id>` defaults to the action frontier, with a `--causal` (or `--all`) flag to display full transitive causal closure.
2. Specify `TaskBlockersJSON` (with fields `task_id`, `projection`, `blockers: []BlockerJSON`, `health`) and `TaskUnblocksJSON` (with fields `task_id`, `unblocks: []DownstreamTaskJSON`) in `internal/wire/dto.go`.

---

**Resolution:** Accepted in substance and recorded in ADR-0006 plus task
6g3q4rt7mgjn. Blockers defaults to the action frontier with causal closure
selected explicitly; query envelopes require projection, health, structured
problems, and deterministic taskflow-owned data. The audit's tentative Go DTO
names and incomplete field lists are intentionally not frozen.

### Low

#### L1. `task list --unblocked` eligibility filter integration over graph snapshots · **Status:** tracked by 6g3q4rt7mgjn

**File:** planning/tasks/6g3q4rt7mgjn-ship-guarded-dependency-mutations-and-graph-queries.md:35, 40, planning/adrs/0006-adopt-threads-as-task-dags.md:443
**Component:** cli / task list / eligibility
**Effort:** S · **Urgency:** eventually

**[follow-up task / scope clarification]**

ADR-0006 §8 includes `tskflwctl task list [--unblocked]` as an actionable selector and specifies that it must filter by derived eligibility (`state.Eligible`, requiring `Role == RoleCandidate && Gate == GateClear`) rather than blocker list emptiness.

While criterion 35 mentions that "frontier/unblocked selectors return no eligible work on an unsound relevant graph", `task list --unblocked` is not explicitly listed in the scope list of task `6g3q4rt7mgjn`.

**Recommendation:** Clarify in task `6g3q4rt7mgjn` whether `task list --unblocked` is included in this slice or tracked as a fast follow-up alongside eligibility enforcement (`Slice 3`). If included, assert that it filters on `TaskGraph.State(id).Eligible` over a healthy snapshot.

**Resolution:** Included in the current query slice. task list --unblocked
selects derived Eligible state and fails closed with an explicit graph-health
diagnosis on an unsound relevant graph; lifecycle transition enforcement remains
in its already-sequenced follow-up task.

#### L2. Mutation receipts should explicitly distinguish modified tasks from idempotent skips · **Status:** tracked by 6g3q4rt7mgjn

**File:** planning/tasks/6g3q4rt7mgjn-ship-guarded-dependency-mutations-and-graph-queries.md:31, 49-51, internal/store/graphmutation.go:26
**Component:** core / cli / receipts
**Effort:** XS · **Urgency:** eventually

**[monitor during implementation]**

When `task depend add A --on B` is executed where `B` is already a prerequisite of `A`, the mutation is an idempotent skip: no file is written and `updated_at` is not bumped. When `task depend add A --on B --on C` is executed where `B` is new and `C` is already present, `A` is modified, but only `B` represents an added edge.

Machine and human receipts should clearly separate applied changes from skipped/no-op targets so downstream automation and agents do not mistake a no-op for a state change.

**Recommendation:** Structure `DependencyMutationReceipt` with `TaskID`, `AddedDependencies: []string`, `RemovedDependencies: []string`, `SkippedDependencies: []string`, `Modified: bool`, and `DryRun: bool`.

---

**Resolution:** Accepted and strengthened in task 6g3q4rt7mgjn. Edge receipts
distinguish applied and idempotently skipped outcomes with canonical endpoints,
changed, dry-run, and workspace identity; migration and error output must also
preserve the durable applied prefix and remaining resumable work.

## Decisions Requiring Owner Input

1. **CLI Command Name for Legacy Migration:**
   - *Option A (Recommended):* `tskflwctl task depend migrate [--dry-run] [--json]` — groups the migration under the `task depend` namespace.
   - *Option B:* `tskflwctl task migrate-dependencies [--dry-run] [--json]` — top-level verb under `task`.
2. **Default Projection for `task blockers`:**
   - *Option A (Recommended):* Default to action frontier (`BlockingFrontier`), with `--causal` flag for full closure. This gives users immediate actionable items first.
   - *Option B:* Default to full causal closure (`CausalBlockers`), with `--frontier` flag for action frontier.
3. **Handling of `task edit` Dependency Deltas:**
   - *Option A (Recommended):* Strict rejection (currently implemented on `main`). If a user modifies `depends_on` or legacy fields in `task edit`, reject with: `"task edit cannot modify depends_on or legacy dependency fields; use task depend add/remove"`. This keeps the interactive editor session 100% outside the mutation guard.
   - *Option B:* Automatically pipe the parsed delta from `task edit` into `Service.AddDependency/RemoveDependency` after the editor exits. (Adds complexity and potential re-entry edge cases).

---

## Proposed Amendments to Task `6g3q4rt7mgjn`

1. **Update Scope Section:**
   - Add `task depend migrate [--dry-run]` to the explicit scope list.
   - Specify `task blockers [--causal]` and `task unblocks` command signatures.
   - Add pure in-memory slug/prefix resolution (`TaskGraph.ResolveTask`) to the core graph scope.
2. **Update Acceptance Criteria:**
   - Add: `TaskGraph exposes pure in-memory slug and prefix resolution so planners resolve user inputs without Store re-entry.`
   - Add: `task depend migrate converts all resolvable legacy blocked_by/dependencies/blocks references to depends_on, cleans legacy frontmatter, and is idempotent on clean repositories.`
   - Clarify: `task edit strictly rejects dependency field modifications and directs users to task depend add/remove.`
3. **Add Wire DTO Definitions:**
   - Specify `TaskBlockersJSON`, `TaskUnblocksJSON`, `DependencyMutationReceiptJSON`, and `DependencyMigrationReceiptJSON` in `internal/wire/dto.go`.

---

## Prioritized Test Matrix

| Priority | Test Category | Scenario & Invariants |
|---|---|---|
| **P0** | Mutation & Cycle Prevention | Add edge `A -> B -> A` fails with `ErrValidation` naming the cycle path; multi-node cycle `A -> B -> C -> A` fails closed; concurrent opposite edge additions across goroutines allow exactly one success and one cycle failure. |
| **P0** | Legacy Migration | Migrate the 6 live repository `blocked_by` tasks in a temporary fixture; verify all frontmatter comments, custom fields, and bodies are preserved; verify `blocked_by` is removed and `depends_on` contains resolved stable IDs; verify repository health transitions from `degraded` to `healthy`; verify subsequent run is a no-op. |
| **P0** | In-Memory Slug Resolution | Verify `TaskGraph.ResolveTask` resolves exact 12-char IDs, full slugs, unique slug prefixes, and returns `ErrNotFound` or `ErrAmbiguous` without invoking `Store` methods. |
| **P1** | Idempotency & No-Op | Adding already-present dependency produces 0 writes, does not bump `updated_at`, exits 0, and returns `skipped` receipt; removing absent dependency produces 0 writes, exits 0, and returns `skipped` receipt. |
| **P1** | Blocker Queries | Query `task blockers` on clear, blocked, unsound-completed, parked, withdrawn, and cyclic tasks; verify deterministic shortest paths, direct/transitive flags, and reason tokens; verify `--causal` vs `--frontier` outputs. |
| **P1** | Downstream Queries | Query `task unblocks` on root and intermediate tasks; verify transitive downstream dependent list and gate states. |
| **P2** | Dry-Run Fidelity | `--dry-run` on `add`, `remove`, and `migrate` runs identical snapshot loading, in-memory resolution, cycle checking, prefix validation, and receipt generation with 0 disk writes. |
| **P2** | `task edit` Guard | Verify `task edit` rejecting dependency field modifications with clear guidance to `task depend add/remove`. |
| **P3** | Wire DTO & JSON Parity | Verify JSON schema validation and stdout formatting across all mutation receipts and query DTOs. |

---

## Traceability Table

| Finding | Severity | Classification | Target Action |
|---|---|---|---|
| **H1** (In-memory slug resolution in `TaskGraph`) | High | Pre-implementation blocker | Amend `6g3q4rt7mgjn` to implement `TaskGraph.ResolveTask` before mutation service |
| **M1** (Unspecified legacy migration command shape) | Medium | Amendment required | Amend `6g3q4rt7mgjn` with `task depend migrate` command spec |
| **M2** (Query Wire DTOs and default projection flags) | Medium | Amendment required | Amend `6g3q4rt7mgjn` with DTO schemas and `--frontier` default |
| **L1** (`task list --unblocked` filter) | Low | Scope clarification / Follow-up | Clarify scope in `6g3q4rt7mgjn` or track in Slice 3 |
| **L2** (Receipt skip vs modified distinction) | Low | Monitor during implementation | Implement structured receipt in `Service` |

---

## Validation Commands and Results

All checks executed on `main` (commit `90bfabe`):

1. **Full Test Suite:**
   ```bash
   go test ./...
   ```
   *Result:* Passed (all 25 packages ok).

2. **Race Detector:**
   ```bash
   go test -race ./...
   ```
   *Result:* Passed (0 data races).

3. **Audit Lint Validation:**
   ```bash
   tskflwctl audit lint 2026-08-27-ship-guarded-dependency-mutations-and-graph-queries
   ```
   *Result:* Passed.

## Candidate tasks

- ⏳ `tskflwctl task new "Add pure in-memory reference resolution to TaskGraph" --epic 30-threads-and-task-dependency-graphs --tags core,graph` — Implement ResolveTask and ResolveID on TaskGraph for mutation planners (H1)
- ⏳ `tskflwctl task new "Add task depend migrate CLI command and DTOs" --epic 30-threads-and-task-dependency-graphs --tags cli,wire` — Define task depend migrate command and receipt DTOs (M1)
- ⏳ `tskflwctl task new "Define blocker and unblocks wire DTOs and CLI flags" --epic 30-threads-and-task-dependency-graphs --tags wire,cli` — Define TaskBlockersJSON and TaskUnblocksJSON schemas and CLI flag defaults (M2)

## Closeout disposition

All five findings are tracked by task 6g3q4rt7mgjn and the durable product contracts are recorded in ADR-0006. The three candidate tasks above were intentionally not created: resolver, migration, query, and receipt work are inseparable acceptance scope for this production slice. The illustrative DTO names remain non-binding. Closeout also adds the audit gaps around downstream-query semantics, the absence of a repository-global task plan command, and structured recovery data after a durable multi-file prefix.
