---
schema: 1
id: 6g3q4rte8kc1
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Centralize task eligibility and explanatory force behavior across every transition into in-progress.
effort: 5-8 days
tier: 1
priority: high
autonomy_level: 2
tags: [threads, graph, lifecycle, cli]
created: "2026-08-25"
updated_at: "2026-08-28"
depends_on: [6g3q4rt7mgjn]
---
# Enforce dependency eligibility across every task start path

## Objective

Make dependency eligibility one authoritative core policy for all transitions into `in-progress`, with an explicit explanatory force escape hatch.

## Scope

- Consume the shared lifecycle-role, gate-state, sound-completion, eligibility, and inconsistency derivation from the strict graph foundation.
- Introduce a narrow guarded lifecycle-mutation capability: the store takes the canonical-root guard, loads one authoritative graph snapshot, invokes a pure core authorization/impact planner, and persists the transition before releasing the guard.
- Route `task start`, generic move, create-and-start, and reusable adapter/TUI entry points through that capability. `task new --start` must authorize and create the in-progress task as one guarded operation rather than create first and move afterward.
- Detect accepted `task edit` status deltas before the generic editor writer can persist them. The product decision below determines whether edit rejects those deltas or delegates them to the guarded lifecycle capability; the editor itself always remains outside the repository guard.
- Replace the ambiguous internal force boolean with typed gate overrides while retaining contextual CLI `--force` spelling.
- Report descendant tasks whose derived gate state changes after a lifecycle transition. Design the receipt for later augmentation, but leave affected Thread discovery and Thread IDs to the Thread-entity task.
- Define one reusable before/after graph-state impact shape and add post-plan state for directly affected dependents to guarded dependency mutation receipts, so a legal edge to a withdrawn or unfinished prerequisite cannot look consequence-free.

## Acceptance criteria

- [ ] Every first-party path that remains capable of entering `in-progress` calls one policy and cannot bypass dependency checks accidentally; rejecting edited status deltas removes `task edit` from that set rather than creating a second policy.
- [ ] `task new --start` performs graph authorization and creation under one repository guard; it cannot commit from a preflight snapshot or as a create-then-move sequence.
- [ ] `task edit` cannot persist a status delta through its ordinary editor write. It either rejects the delta before writing, or keeps the interactive editor outside the guard and detects concurrent source or graph changes before a delegated lifecycle commit.
- [ ] Ineligible starts fail by default with deterministic outstanding blockers.
- [ ] `task start --force` and `task move ... in-progress --force` bypass only the dependency gate; completion force remains scoped to unexplained acceptance criteria.
- [ ] Reopening an upstream task makes completed descendants unsound without rewriting their persisted statuses.
- [ ] Deferred and deprecated prerequisites follow ADR-0006 semantics consistently.
- [ ] Lifecycle receipts report descendant task IDs/counts whose gate state changed, their before/after derived states, the override used, and an explanatory remedy; the shape can later add affected Thread IDs without changing these meanings.
- [ ] Dependency add/remove receipts report the resulting derived state of each directly affected dependent and human output calls out a newly blocked, broken, or inconsistent task.
- [ ] Eligibility authorization and the persisted lifecycle transition occur
  under one repository guard over the same authoritative graph snapshot.
- [ ] Lifecycle writes use a dedicated use-case-specific guarded capability
  sharing private store guard/materialization helpers; they neither nest
  MutateTaskGraph nor expose a generic filesystem callback.
- [ ] Races with prerequisite lifecycle or dependency changes cannot commit a
  start authorized by stale graph state; default and forced paths have
  adversarial coverage.

## Stress tests

- Table-drive every task status against clear, blocked, and broken gates across every entry path.
- Cover direct/transitive blockers, forced completion, upstream reopen, missing/withdrawn prerequisites, create-and-start, rejected or delegated editor candidates, and parity between human and machine receipts.
- Race a start against prerequisite lifecycle changes and dependency mutations. Race an accepted editor candidate against both task-source and graph changes without holding the repository guard while the editor is open.
- Exercise batch transitions and measure repeated repository scans before introducing any invocation-local snapshot cache.

## Sequencing

Requires guarded dependency operations. This task settles the first non-dependency extension of the guarded-write pattern; Thread persistence follows it so the two slices do not invent incompatible guard and materialization seams.

## Product decision required before implementation

Decide whether `task edit` remains a lifecycle surface:

- **Reject status deltas (recommended):** preserve editing for task content and ordinary metadata, and direct every status change to explicit lifecycle verbs. This matches `task set`, keeps dates, gates, force semantics, and receipts on one obvious path, and substantially narrows the guarded editor contract.
- **Delegate status deltas:** accept a candidate edited outside the lock, then authorize the status transition and commit the complete candidate under the guarded lifecycle capability with exact-source CAS. This must enforce every applicable lifecycle rule—not only dependency eligibility—including completion criteria, lifecycle timestamps, reopen impact, typed overrides, and receipts.

Either choice satisfies the ADR rule that a first-party editor cannot bypass lifecycle policy. Directly saving an edited status through the generic editor writer is not an option.

## Atomic lifecycle boundary amendment (2026-08-27)

Eligibility is an authorization decision, so reading the graph and persisting the move into in-progress must be one guarded operation. A preflight graph query followed by an ordinary Move is insufficient: a prerequisite or dependency can change between those calls. Compute before/after derived state from snapshots owned by the same boundary so descendant-impact receipts describe the transition that actually committed.

Introduce a narrow lifecycle-mutation capability implemented over the store's private canonical-root guard and lock-free task materialization helpers. Do not broaden core into a filesystem callback and do not implement this by nesting MutateTaskGraph. This task is the first non-dependency extension of the guarded-write pattern and should settle that internal reuse seam before Thread persistence implements another entity kind.

Sequencing is therefore strict for implementation coordination: guarded dependency operations first, then this lifecycle boundary, then Thread persistence. The domain derivations remain independently testable; the ordering prevents two tasks from inventing incompatible guard extensions.
