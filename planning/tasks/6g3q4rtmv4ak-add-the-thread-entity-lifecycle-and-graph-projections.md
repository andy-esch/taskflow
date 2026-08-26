---
schema: 1
id: 6g3q4rtmv4ak
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Introduce first-class Thread persistence, membership, lifecycle, rollup, frontier, external gates, CLI, and wire projections.
effort: 4-7 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, domain, storage, cli]
created: "2026-08-25"
---
# Add the Thread entity, lifecycle, and graph projections

## Objective

Introduce Threads as first-class initiative documents whose membership and lifecycle project—but never own—the repository-global task DAG.

## Scope

- Add Thread domain validation, flat ID-addressed Markdown storage, narrow core ports/use cases, CLI, completion, wire/schema, init/layout, health, and watch coverage.
- Implement many-valued membership, lifecycle, rollup, frontier, external gates, and completed-Thread inconsistency.
- Report nominal completion separately from soundly drained members, and augment lifecycle-impact receipts with affected Thread IDs.
- Retire the unused Projects scaffold without deleting non-empty legacy content.

## Acceptance criteria

- [ ] Thread files own metadata and a sorted task-ID membership set only; task files remain the dependency source of truth.
- [ ] A task can belong to multiple Threads and an outside prerequisite appears as an external gate without entering progress totals while still preventing Thread completion until satisfied.
- [ ] Views expose nominal `done / total`, sound `drained / total`, graph health, and exact outstanding external gates without contradictory completion UX.
- [ ] Start and complete require at least one non-withdrawn member; complete, abandon, reopen, and membership mutation rules otherwise match ADR-0006.
- [ ] CLI and wire projections use the shared graph/eligibility analysis and expose stable member/external roles.
- [ ] Initialization creates `threads/`, stops creating `projects/`, and handles non-empty legacy Projects safely.
- [ ] Task/Thread IDs are checked for cross-kind collisions, and an empty Projects scaffold is defined narrowly enough to permit only `.gitkeep` removal.
- [ ] Task lifecycle receipts name affected Thread IDs once Thread documents can exist.

## Stress tests

- Shared members, external gates, empty Threads, deferred/deprecated members, completed-thread drift after upstream reopen, membership conflicts, ID drift, and watcher reload.
- Domain, descriptor, store, core fake, CLI, completion, schema-golden, and wire fan-out is explicitly covered.

## Sequencing

Requires the strict graph derivation and portable guarded writes, but not lifecycle enforcement. Bulk linking waits for this persistence contract and dependency operations.
