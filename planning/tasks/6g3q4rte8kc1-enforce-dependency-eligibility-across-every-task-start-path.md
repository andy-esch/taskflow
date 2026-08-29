---
schema: 1
id: 6g3q4rte8kc1
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Centralize task eligibility and explanatory force behavior across every transition into in-progress.
effort: 5-8 days
tier: 1
priority: high
autonomy_level: 2
tags: [threads, graph, lifecycle, cli]
created: "2026-08-25"
updated_at: "2026-08-29"
depends_on: [6g3q4rt7mgjn]
started_at: "2026-08-28"
completed_at: "2026-08-29"
---
# Enforce dependency eligibility across every task start path

## Objective

Make dependency eligibility one authoritative core policy for all transitions into `in-progress`, with an explicit explanatory force escape hatch.

## Scope

- Consume the shared lifecycle-role, gate-state, sound-completion, eligibility, and inconsistency derivation from the strict graph foundation.
- Introduce a narrow guarded lifecycle-mutation capability: the store takes the canonical-root guard, loads one authoritative graph snapshot, invokes a pure core authorization/impact planner, and persists the transition before releasing the guard.
- Route `task start`, generic move, create-and-start, and reusable adapter/TUI entry points through that capability. `task new --start` must authorize and create the in-progress task as one guarded operation rather than create first and move afterward.
- Detect and reject accepted `task edit` status deltas before the generic editor writer can persist them, with an actionable pointer to the explicit lifecycle verbs. The editor itself always remains outside the repository guard.
- Replace the ambiguous internal force boolean with typed gate overrides while retaining contextual CLI `--force` spelling.
- Report descendant tasks whose derived gate state changes after a lifecycle transition. Design the receipt for later augmentation, but leave affected Thread discovery and Thread IDs to the Thread-entity task.
- Define one reusable before/after graph-state impact shape and add post-plan state for directly affected dependents to guarded dependency mutation receipts, so a legal edge to a withdrawn or unfinished prerequisite cannot look consequence-free.

## Acceptance criteria

- [x] Every first-party path capable of entering `in-progress` calls one policy and cannot bypass dependency checks accidentally; `task edit` is not such a path because it rejects every status delta.
- [x] `task new --start` performs graph authorization and creation under one repository guard; it cannot commit from a preflight snapshot or as a create-then-move sequence.
- [x] `task edit` rejects every status delta before writing and directs the user to the appropriate lifecycle verb; editing task content and transitioning lifecycle remain separate operations.
- [x] Ineligible starts fail by default with deterministic outstanding blockers.
- [x] `task start --force` and `task move ... in-progress --force` bypass only the dependency gate; completion force remains scoped to unexplained acceptance criteria.
- [x] Reopening an upstream task makes completed descendants unsound without rewriting their persisted statuses.
- [x] Deferred and deprecated prerequisites follow ADR-0006 semantics consistently.
- [x] Lifecycle receipts report descendant task IDs/counts whose gate state changed, their before/after derived states, the override used, and an explanatory remedy; the shape can later add affected Thread IDs without changing these meanings.
- [x] Dependency add/remove receipts report the resulting derived state of each directly affected dependent and human output calls out a newly blocked, broken, or inconsistent task.
- [x] Eligibility authorization and the persisted lifecycle transition occur
  under one repository guard over the same authoritative graph snapshot.
- [x] Lifecycle writes use a dedicated use-case-specific guarded capability
  sharing private store guard/materialization helpers; they neither nest
  MutateTaskGraph nor expose a generic filesystem callback.
- [x] Races with prerequisite lifecycle or dependency changes cannot commit a
  start authorized by stale graph state; default and forced paths have
  adversarial coverage.

## Stress tests

- Table-drive every task status against clear, blocked, and broken gates across every entry path.
- Cover direct/transitive blockers, forced completion, upstream reopen, missing/withdrawn prerequisites, create-and-start, rejected editor status deltas, and parity between human and machine receipts.
- Race a start against prerequisite lifecycle changes and dependency mutations. Verify that editor rejection remains deterministic across concurrent task-source changes without holding the repository guard while the editor is open.
- Exercise batch transitions and measure repeated repository scans before introducing any invocation-local snapshot cache.

## Sequencing

Requires guarded dependency operations. This task settles the first non-dependency extension of the guarded-write pattern; Thread persistence follows it so the two slices do not invent incompatible guard and materialization seams.

## Task-edit lifecycle decision (2026-08-28)

`task edit` does not own lifecycle transitions. It rejects every status delta and directs the user to explicit lifecycle verbs such as `task start`, `task complete`, `task defer`, or generic `task move`. This may require a second CLI call when a user edits task content and transitions it in the same workflow, but it keeps dates, eligibility, completion gates, typed overrides, and impact receipts on one explicit guarded path. `task new --start` remains the intentional one-command create-and-start operation.

## Atomic lifecycle boundary amendment (2026-08-27)

Eligibility is an authorization decision, so reading the graph and persisting the move into in-progress must be one guarded operation. A preflight graph query followed by an ordinary Move is insufficient: a prerequisite or dependency can change between those calls. Compute before/after derived state from snapshots owned by the same boundary so descendant-impact receipts describe the transition that actually committed.

Introduce a narrow lifecycle-mutation capability implemented over the store's private canonical-root guard and lock-free task materialization helpers. Do not broaden core into a filesystem callback and do not implement this by nesting MutateTaskGraph. This task is the first non-dependency extension of the guarded-write pattern and should settle that internal reuse seam before Thread persistence implements another entity kind.

Sequencing is therefore strict for implementation coordination: guarded dependency operations first, then this lifecycle boundary, then Thread persistence. The domain derivations remain independently testable; the ordering prevents two tasks from inventing incompatible guard extensions.

## Implementation progress (2026-08-29)

The implementation now has one typed `TaskLifecycleMutationStore` boundary beside guarded dependency mutation. It holds the canonical-root repository guard while loading a strict graph, invoking the pure lifecycle planner, validating eligibility or completion policy, deriving before/after impact, materializing status and timestamps, performing whole-snapshot and immediate-target CAS checks, and atomically writing. `task start`, every generic task lifecycle verb, defer, TUI actions, and the one-step `task new --start` path use it. The old store `Move`/`Defer` status-writing capability and ambiguous force boolean are gone.

`task edit`, `task set`, and the store field writer reject status changes and direct callers to lifecycle verbs. A dependency-gate override can enter `in-progress` only from the candidate role, cannot bypass a broken repository, and cannot create a lint-invalid active task; the acceptance-criteria override remains completion-only. Lifecycle JSON receipts now include exact source status, typed override, whether it was applied, outstanding blockers, descendant impact count/details, and remedy. Dependency add/remove receipts use the same impact shape for the directly affected dependent and human output makes newly unsafe state noisy.

Adversarial coverage includes blocked/parked/withdrawn prerequisites, every persisted source status entering `in-progress`, force-scope separation, broken repositories, unsound descendants after reopen, create-and-start, CLI and TUI parity, editor rejection, raw prerequisite/dependency changes before whole-graph CAS, and a target edit in the final CAS window. The race-enabled full suite passes. Batch transitions deliberately retain one guarded graph scan for a dry-run item and a verification re-scan for each committing item; no invocation-local snapshot cache was introduced before real usage establishes whether that repeated work matters.

Implementation acceptance was satisfied pending external review; the closeout below records the final dispositions.

## Adversarial review closeout (2026-08-29)

Two independent implementation audits were dispositioned and closed. The blocking committed-outcome defect is fixed with an explicit durable phase, typed recovery receipt, no post-commit retry, structured CLI recovery, and TUI warning plus reload. Ordinary creation now rejects lifecycle-owned or invalid statuses; same-status no-ops validate typed override scope first.

Machine refusals retain derived state, blocker paths, override eligibility, and remedy per batch row. TUI task moves retain descendant impacts. Deterministic two-store tests now exercise real dependency add/remove and prerequisite lifecycle changes under the shared repository guard, covering default and dependency-gate override authorization from fresh state.

The reviews batch-atomicity observation was rejected as outside the established per-item contract. Repeated graph scans remain deliberately deferred until dogfooding or a benchmark establishes a bottleneck. Validation passed: the full race-enabled suite, golangci-lint, go vet, module tidy diff, generated-schema drift checks, planning lint, and git diff hygiene.
