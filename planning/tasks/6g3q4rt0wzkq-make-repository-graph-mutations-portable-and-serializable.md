---
schema: 1
id: 6g3q4rt0wzkq
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Provide a cross-platform store-owned guard for authoritative repository graph read, validation, and write operations.
effort: 2-4 days
tier: 1
priority: high
autonomy_level: 2
tags: [threads, graph, storage, concurrency]
created: "2026-08-25"
---
# Make repository graph mutations portable and serializable

## Objective

Provide one store-owned repository mutation boundary that makes final graph read, validation, and write indivisible for cooperating writers on every supported platform.

## Scope

- Replace or explicitly close the non-Unix no-op locking gap.
- Expose a narrow control-inverted mutation port: the store locks, loads the strict snapshot, invokes a core-supplied pure planner, applies returned writes through package-internal lock-free helpers, and unlocks.
- Forbid Store calls from inside the planner callback and make invalid nested mutation fail explicitly rather than self-deadlock.
- Preserve atomic replacement and content-version conflict behavior for ordinary entity writes.

## Acceptance criteria

- [ ] Every supported platform has an explicit, tested cooperating-writer serialization contract; unsupported guarantees are rejected or documented rather than silently no-op.
- [ ] A graph mutation performs its authoritative scan, pure planning/validation, and writes within one repository guard without exposing filesystem locking or teaching the store graph semantics.
- [ ] The callback contract accepts and returns taskflow-owned snapshot/planned-write values, permits no nested Store calls, and detects invalid nesting without hanging.
- [ ] Lock acquisition/release errors are attributable and process termination does not leave unrecoverable stale state.
- [ ] Existing optimistic concurrency and ordinary write behavior remain compatible.

## Stress tests

- Concurrent opposite edge intents cannot both commit a cycle.
- Concurrent same-file and different-file writers cannot lose updates silently.
- Error, panic-equivalent, and process-exit paths release or recover the guard as designed.
- A faithful nested-acquisition regression test cannot reproduce the current self-deadlock.

## Sequencing

The control-inversion contract is fixed by ADR-0006's 2026-08-26 amendment, so implementation can proceed alongside the strict read foundation. Guarded dependency writes require both tasks to land.
