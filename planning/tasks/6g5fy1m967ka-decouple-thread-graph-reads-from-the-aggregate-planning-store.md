---
schema: 1
id: 6g5fy1m967ka
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Give Thread graph reads an independently injectable task-graph source so CLI, TUI, web, and split adapters share one neutral core projection boundary.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 3
tags: [threads, architecture, ports, graph]
created: "2026-08-31"
updated_at: "2026-08-31"
started_at: "2026-08-31"
completed_at: "2026-08-31"
---

# Decouple Thread graph reads from the aggregate planning store

## Objective

Remove the aggregate planning `Store` as an accidental prerequisite of read-only graph use cases.
Give `core.Service` an independently injectable `TaskGraphSource`, default it from complete adapters,
and prove Thread projections can be assembled from separate task and Thread ports. This preserves
one framework-neutral foundation for CLI, TUI, future web, and split persistence adapters before
generated Thread graph views deepen the read surface.

## Current constraint

`ThreadStore` is already independently injectable, but `ShowThread`, `ListThreadViews`, graph
queries, and Thread compose still call `LoadTaskGraph(s.store)`. Since `store` is the aggregate task,
epic, audit, and research port, a read-only or split adapter must implement unrelated entity methods
merely to supply the task DAG. `NewService(nil, WithThreadStore(...))` therefore cannot actually
perform a Thread read without an aggregate store.

## Scope

- Add one consumer-owned `TaskGraphSource` capability to `Service`, with an explicit option and
  backward-compatible discovery from the ordinary aggregate `Store`.
- Route diagnostic graph queries, Thread list/show projections, and Thread compose through that
  capability. Guarded mutation ports continue to own their authoritative snapshots.
- Fail explicitly when the graph source is unavailable instead of reaching a nil aggregate store.
- Let `WorkspaceSource` carry separate task-graph and Thread read capabilities so the local TUI
  workspace path does not depend on both happening to be methods on one concrete filesystem value.
- Keep all projected values taskflow-owned and framework-neutral; no CLI, TUI, HTTP, filesystem, or
  third-party graph type may cross the core boundary.

## Acceptance criteria

- [x] `core.Service` accepts an independently injected `TaskGraphSource` and defaults it from a
  non-nil aggregate `Store` without changing current production construction.
- [x] Thread list/show and task blocker/downstream queries work with separate minimal graph and
  Thread fakes while the aggregate store is nil.
- [x] Thread compose uses the narrow graph source plus `ThreadStore`; unavailable capability errors
  are explicit and do not panic.
- [x] A workspace adapter can provide separate graph and Thread read ports, and the resulting
  `Workspace.Planning` service exposes the same Thread projection without concrete-adapter checks.
- [x] Guarded mutation callbacks and ports remain unchanged; this task does not weaken authoritative
  under-lock revalidation.
- [x] Architecture and ADR guidance name the reusable projection boundary that graph views, TUI,
  and future web adapters must consume.

## Stress tests

- Nil aggregate store with both narrow read ports; graph-only without Thread support; Thread-only
  without graph support; separate fake implementations; complete filesystem adapter fallback; and
  workspace opening with split capabilities.

## Out of scope

- Implementing `ThreadGraphProjection`, Mermaid/DOT, CLI commands, TUI screens, HTTP transport,
  database storage, or redesigning mutation capabilities.

## Sequencing

This is a direct prerequisite of
[generate-deterministic-thread-graph-views](6g3q4rv1w9e2-generate-deterministic-thread-graph-views.md).
The graph-view task is intentionally derived as in-flight/blocked while this foundation is active;
completion restores its clear gate without changing either task's lifecycle status.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)

## Implementation outcome

`core.Service` now owns an independently injectable `TaskGraphSource`, with the complete `Store`
retained as the backward-compatible default. Thread list/show, graph queries, and Thread compose read
through the narrow capability; workspace composition carries task-graph and Thread reads separately,
while guarded mutation callbacks remain unchanged and authoritative.

Tests exercise nil aggregate-store composition, graph-only and Thread-only capability failures,
distinct graph/Thread implementations, workspace assembly, compose, and the existing complete-store
fallback. Architecture guidance and ADR-0006 now require an adapter-neutral core graph projection
with reusable Mermaid/DOT formatters outside CLI/framework packages. Full tests, focused race tests,
vet, golangci-lint, and diff checks pass.

## Adversarial-review hardening

Two independent reviews confirmed that guarded mutation remains authoritative but exposed several
read-boundary gaps. Task listing and board now read through `TaskGraphSource`; typed-nil aggregate
and narrow capabilities become explicit unavailable errors; and Thread-backed projections read
Threads before tasks under a documented causally-compatible source-pair contract. Workspace prose
now distinguishes its intentionally complete local-TUI service from direct read-only `Service`
composition, and deterministic graph formatting has a named dependency-free package and escaping
owner.

The filesystem-shaped unreadable-record attribution was fail-closed but not portable. Follow-up
[make-task-graph-load-diagnostics-adapter-neutral](6g5gbk5a5bt0-make-task-graph-load-diagnostics-adapter-neutral.md)
replaced it with a neutral record-identity channel as a hard prerequisite of graph-view machine
contracts rather than hiding the larger change inside this port-wiring slice.
