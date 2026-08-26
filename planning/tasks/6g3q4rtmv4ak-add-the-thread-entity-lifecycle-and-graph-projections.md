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
- Retire the unused Projects scaffold without deleting non-empty legacy content.

## Acceptance criteria

- [ ] Thread files own metadata and a sorted task-ID membership set only; task files remain the dependency source of truth.
- [ ] A task can belong to multiple Threads and an outside prerequisite appears as an external gate without entering progress totals.
- [ ] Start, complete, abandon, reopen, and membership mutation rules match ADR-0006.
- [ ] CLI and wire projections use the shared graph/eligibility analysis and expose stable member/external roles.
- [ ] Initialization creates `threads/`, stops creating `projects/`, and handles non-empty legacy Projects safely.

## Stress tests

- Shared members, external gates, empty Threads, deferred/deprecated members, completed-thread drift after upstream reopen, membership conflicts, ID drift, and watcher reload.
- Domain, descriptor, store, core fake, CLI, completion, schema-golden, and wire fan-out is explicitly covered.

## Sequencing

Requires the strict graph foundation, guarded writes, and shared eligibility semantics. Bulk linking waits for this persistence contract.
