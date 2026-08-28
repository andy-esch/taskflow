---
schema: 1
id: 6g3q4rtmv4ak
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Introduce first-class Thread persistence, membership, lifecycle, rollup, frontier, external gates, CLI, and wire projections.
effort: 5-8 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, domain, storage, cli]
created: "2026-08-25"
updated_at: "2026-08-28"
depends_on: [6g3q4rte8kc1]
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
- [ ] Thread create, membership, start, complete, abandon, reopen, and
  cross-kind ID collision checks perform authoritative read, validation, and
  write under one repository guard.
- [ ] The store provides lock-free internal Thread materialization reusable by a
  later compound bulk capability without nesting guarded mutations.
- [ ] Concurrent Thread membership or lifecycle operations and task-graph
  changes either serialize to a valid state or return an attributable conflict.

## Stress tests

- Shared members, external gates, empty Threads, deferred/deprecated members, completed-thread drift after upstream reopen, membership conflicts, ID drift, and watcher reload.
- Domain, descriptor, store, core fake, CLI, completion, schema-golden, and wire fan-out is explicitly covered.

## Sequencing

Requires completed dependency eligibility enforcement so this task can extend the settled guarded lifecycle-write and materialization seam to a second entity kind. Bulk linking waits for this persistence contract and dependency operations.

## Readiness checkpoint

Review and re-estimate this task after eligibility ships against the actual reusable guard, planner, materialization, and receipt seams. Keep it as one vertical slice if Thread persistence, membership, lifecycle, and projections can land coherently within the estimate. Otherwise split persistence/read projections from guarded membership/lifecycle mutations before implementation; do not invent that split before the lifecycle seam exists.

## Guarded Thread mutation amendment (2026-08-27)

Diagnostic Thread projections may degrade explicitly, but every authoritative Thread mutation loads the required task graph and current Thread state inside one canonical-root repository guard. Creation and membership validate task existence and cross-kind identity there; start and complete validate membership, external gates, and sound completion there; the matching Thread write lands before release.

Expose use-case-specific Thread mutation ports backed by private store guard/materialization helpers. Keep a lock-free internal Thread document materializer so bulk apply can compose task dependency writes and the final Thread write under one outer guard. Calling MutateTaskGraph from a Thread planner, or calling a guarded Thread method from a graph planner, is forbidden and will correctly fail callback exclusion.

Implement after the eligibility task establishes the first non-dependency guarded-write pattern. Bulk linking waits for this Thread persistence/materialization contract as well as dependency operations.
