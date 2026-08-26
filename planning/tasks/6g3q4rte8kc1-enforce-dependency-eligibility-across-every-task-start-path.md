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

- Derive lifecycle role, clear/blocked/broken gate state, sound completion, eligibility, and inconsistency from persisted tasks.
- Route `task start`, generic move, create-and-start, and reusable adapter/TUI entry points through the same guard.
- Return forced-transition metadata without removing or waiving dependencies.

## Acceptance criteria

- [ ] Every current path into `in-progress` calls one policy and cannot bypass dependency checks accidentally.
- [ ] Ineligible starts fail by default with deterministic outstanding blockers.
- [ ] `--force` changes only lifecycle state and reports why the resulting task is inconsistent.
- [ ] Reopening an upstream task makes completed descendants unsound without rewriting their persisted statuses.
- [ ] Deferred and deprecated prerequisites follow ADR-0006 semantics consistently.

## Stress tests

- Table-drive every task status against clear, blocked, and broken gates across every entry path.
- Cover direct/transitive blockers, forced completion, upstream reopen, missing/withdrawn prerequisites, and parity between human and machine receipts.

## Sequencing

Requires guarded dependency operations. The Thread projection must consume this policy rather than recalculate readiness.
