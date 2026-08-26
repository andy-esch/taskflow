---
schema: 1
id: 6g3q4rte8kc1
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Centralize task eligibility and explanatory force behavior across every transition into in-progress.
effort: 2-4 days
tier: 1
priority: high
autonomy_level: 2
tags: [threads, graph, lifecycle, cli]
created: "2026-08-25"
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

## Stress tests

- Table-drive every task status against clear, blocked, and broken gates across every entry path.
- Cover direct/transitive blockers, forced completion, upstream reopen, missing/withdrawn prerequisites, and parity between human and machine receipts.
- Exercise batch transitions and measure repeated repository scans before introducing any invocation-local snapshot cache.

## Sequencing

Requires guarded dependency operations. Thread projections consume the derivation from the strict-read task; enforcement and Thread persistence may proceed independently once their own prerequisites land.
