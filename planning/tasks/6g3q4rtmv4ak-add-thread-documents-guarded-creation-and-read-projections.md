---
schema: 1
id: 6g3q4rtmv4ak
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Introduce Thread documents, guarded unstarted creation, reusable materialization, and deterministic read projections.
effort: 5-8 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, domain, storage, cli]
created: "2026-08-25"
updated_at: "2026-08-29"
depends_on: [6g3q4rte8kc1]
---
# Add Thread documents, guarded creation, and read projections

## Objective

Introduce durable Thread documents and one deterministic read projection over the repository-global task DAG, with guarded creation and a materialization seam the later mutation and bulk slices can safely reuse.

## Scope

- Add Thread domain validation, the entity descriptor/template, flat ID-addressed Markdown storage, narrow read ports/use cases, CLI completion, wire/schema, init/layout, health, and watch coverage.
- Define the pure Thread projection over the shared task graph: many-valued membership, nominal and sound rollups, frontier, external gates, and completed-Thread inconsistency.
- Ship guarded creation in the explicit `unstarted` state. Creation validates membership IDs and cross-kind identity against one authoritative task/Thread snapshot and writes through a lock-free internal Thread materializer.
- Expose read-first CLI/wire surfaces for create, list, show, path, and frontier. Membership and lifecycle verbs belong to the immediately following task.
- Retire the unused Projects scaffold without deleting non-empty legacy content.

## Acceptance criteria

- [ ] Thread files own metadata and a sorted task-ID membership set only; task files remain the dependency source of truth.
- [ ] A task can belong to multiple Threads and an outside prerequisite appears as an external gate without entering progress totals while still preventing Thread completion until satisfied.
- [ ] Views expose nominal `done / total`, sound `drained / total`, graph health, and exact outstanding external gates without contradictory completion UX.
- [ ] CLI and wire create/read projections use the shared graph/eligibility analysis and expose stable member/external roles without reimplementing task graph rules.
- [ ] Initialization creates `threads/`, stops creating `projects/`, and handles non-empty legacy Projects safely.
- [ ] Task/Thread IDs are checked for cross-kind collisions, and an empty Projects scaffold is defined narrowly enough to permit only `.gitkeep` removal.
- [ ] Thread creation always persists `unstarted`, validates member existence and cross-kind ID collisions inside one canonical-root guard, and cannot accept an arbitrary lifecycle state through a generic create port.
- [ ] The store provides lock-free internal Thread materialization reusable by a
  later compound bulk capability without nesting guarded mutations.
- [ ] A real two-store race between Thread creation and task-graph/task-lifecycle mutation proves canonical-root serialization and fresh-snapshot validation; raw-file CAS coverage remains separate.

## Stress tests

- Shared members, external gates, empty Threads, deferred/deprecated members, completed-thread drift after upstream reopen, ID drift/collision, degraded diagnostic reads, guarded-creation races, and watcher reload.
- Domain, descriptor, store, core fake, CLI, completion, schema-golden, init/layout, health, and wire fan-out is explicitly covered.

## Sequencing

Requires completed dependency eligibility enforcement so guarded creation can reuse the settled committed-outcome and repository-guard rules. The guarded membership/lifecycle task follows this document/materialization contract; bulk linking waits for both.

## Readiness checkpoint

Completed 2026-08-29. The shipped lifecycle seam proved reusable, but the original task combined the full first-class-entity fan-out with a second guarded mutation family and task-lifecycle receipt augmentation. Those are independently reviewable boundaries and no longer fit one 5–8 day slice. This task now owns documents, guarded unstarted creation, materialization, and read projections; `ship-guarded-thread-membership-and-lifecycle-mutations` owns mutable membership/lifecycle and affected-Thread receipts.

## Guarded Thread mutation amendment (2026-08-27)

Diagnostic Thread projections may degrade explicitly, but guarded creation loads the required task graph and current Thread state inside one canonical-root repository guard. It validates task existence and cross-kind identity there and lands the Thread write before release. The following mutation task applies the same rule to membership and lifecycle.

Expose a use-case-specific guarded Thread-creation port backed by private store guard/materialization helpers. Keep a lock-free internal Thread document materializer so membership/lifecycle and bulk apply can compose the final Thread write under one outer guard. Calling `MutateTaskGraph` from a Thread planner, or nesting a guarded Thread method from another planner, is forbidden and will correctly fail callback exclusion.

Implement after the eligibility task establishes the first non-dependency guarded-write pattern. Guarded membership/lifecycle follows this Thread persistence/materialization contract; bulk linking waits for that second slice as well as dependency operations.
