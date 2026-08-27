---
schema: 1
id: 6g3q4rte8kc1
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Centralize task eligibility and explanatory force behavior across every transition into in-progress.
effort: 3-5 days
tier: 1
priority: high
autonomy_level: 2
tags: [threads, graph, lifecycle, cli]
created: "2026-08-25"
updated_at: "2026-08-27"
---
# Enforce dependency eligibility across every task start path

## Objective

Make dependency eligibility one authoritative core policy for all transitions into `in-progress`, with an explicit explanatory force escape hatch.

## Scope

- Consume the shared lifecycle-role, gate-state, sound-completion, eligibility, and inconsistency derivation from the strict graph foundation.
- Route `task start`, generic move, create-and-start, accepted `task edit` status deltas, and reusable adapter/TUI entry points through the same guard.
- Replace the ambiguous internal force boolean with typed gate overrides while retaining contextual CLI `--force` spelling.
- Report descendant tasks whose derived gate state changes after a lifecycle transition; add affected Thread IDs when Thread persistence exists.

## Acceptance criteria

- [ ] Every first-party path into `in-progress`, including `task edit`, calls one policy and cannot bypass dependency checks accidentally.
- [ ] Ineligible starts fail by default with deterministic outstanding blockers.
- [ ] `task start --force` and `task move ... in-progress --force` bypass only the dependency gate; completion force remains scoped to unexplained acceptance criteria.
- [ ] Reopening an upstream task makes completed descendants unsound without rewriting their persisted statuses.
- [ ] Deferred and deprecated prerequisites follow ADR-0006 semantics consistently.
- [ ] Lifecycle receipts report descendant task IDs/counts whose gate state changed and, after Thread support lands, affected Thread IDs with an explanatory remedy.
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
- Cover direct/transitive blockers, forced completion, upstream reopen, missing/withdrawn prerequisites, and parity between human and machine receipts.
- Exercise batch transitions and measure repeated repository scans before introducing any invocation-local snapshot cache.

## Sequencing

Requires guarded dependency operations. Thread projections consume the derivation from the strict-read task; enforcement and Thread persistence may proceed independently once their own prerequisites land.

## Atomic lifecycle boundary amendment (2026-08-27)\n\nEligibility is an authorization decision, so reading the graph and persisting the move into in-progress must be one guarded operation. A preflight graph query followed by an ordinary Move is insufficient: a prerequisite or dependency can change between those calls. Compute before/after derived state from snapshots owned by the same boundary so descendant-impact receipts describe the transition that actually committed.\n\nIntroduce a narrow lifecycle-mutation capability implemented over the store's private canonical-root guard and lock-free task materialization helpers. Do not broaden core into a filesystem callback and do not implement this by nesting MutateTaskGraph. This task is the first non-dependency extension of the guarded-write pattern and should settle that internal reuse seam before Thread persistence implements another entity kind.\n\nSequencing is therefore strict for implementation coordination: guarded dependency operations first, then this lifecycle boundary, then Thread persistence. The domain derivations remain independently testable; the ordering prevents two tasks from inventing incompatible guard extensions.
