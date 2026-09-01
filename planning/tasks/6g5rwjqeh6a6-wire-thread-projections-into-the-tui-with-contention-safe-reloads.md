---
schema: 1
id: 6g5rwjqeh6a6
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Load shared Thread projections through TUI adapters and retain the last coherent model across watcher bursts and transient planner conflicts.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, tui, architecture, concurrency]
created: "2026-09-01"
depends_on: [6g3q4rv1w9e2, 6g5m69wpydzw, 6g5rxq1g5mp1, 6g5rxq1ravd3]
updated_at: "2026-09-01"
---

# Wire Thread projections into the TUI with contention-safe reloads

## Objective

Give the TUI one faithful, reload-safe path to the production Thread projections before deciding
how those values should look on screen. The adapter must consume the same core services as the CLI,
keep split task-graph and Thread stores independently injectable, and retain the last coherent model
when a watcher refresh overlaps a guarded repository mutation.

## Scope

- Add TUI loaders and typed messages for `ListThreadViews`, `ShowThread`, and, where needed by the
  later topology view, `ShowThreadGraph`; carry taskflow-owned projection values rather than
  flattening them into a second TUI semantic model.
- Extend the entity/reload plumbing only as far as needed to load, generation-stamp, stale-drop, and
  cursor-restore Thread data. Visible list/detail design belongs to the next task.
- Make task and Thread filesystem events invalidate Thread projections through the existing shared
  reload/debounce path.
- Treat planner-window `domain.ErrConflict` as bounded transient contention: keep the last coherent
  list/detail, coalesce the event burst, and retry after it settles. A first load may remain visibly
  loading/retrying, but must not become a false empty state.
- Put transient read handling in the shared asynchronous list/detail/dashboard machinery rather than
  adding a Thread-only error branch; existing entities can hit the same canonical-root exclusion
  window and should receive the same bounded behavior.
- Keep durable validation, I/O, and projection errors visible and non-spinning; only the documented
  conflict class receives transient treatment.
- Exercise workspace construction with independently supplied `TaskGraphSource` and `ThreadStore`
  fakes so later TUI work cannot accidentally collapse the reviewed port boundary.

## Acceptance criteria

- [ ] Thread list, detail, and graph loaders call the existing core service methods and preserve
  their lifecycle, rollup, frontier, gate, health, topology, and diagnostic values without local
  traversal or readiness calculations.
- [ ] A workspace test proves the TUI can read Threads when the aggregate store, task graph source,
  and Thread store are distinct compatible capabilities.
- [ ] Task dependency changes and Thread document changes both trigger a coherent Thread reload;
  generation stamps prevent an older result from overwriting a newer one and cursor/selection intent
  survives the refresh.
- [ ] A refresh-time `domain.ErrConflict` retains the last coherent model and schedules a bounded,
  coalesced retry; a deterministic test covers both initial-load and already-loaded behavior.
- [ ] The retry/retain policy is entity-agnostic and covers list, detail, and dashboard reads without
  letting one failing background tab blank or repeatedly reload unrelated surfaces.
- [ ] Non-conflict failures remain visible, do not spin, and recover on a later successful manual or
  watcher reload.

## Stress tests

- Mutation events emitted in several durable prefixes, conflict followed by success, repeated
  conflict at the retry bound, stale success arriving after a newer load, first-load contention,
  task-only edits that change Thread eligibility, and split-adapter causal compatibility.

## Out of scope

- Thread row/detail layout, topology visualization, lifecycle mutations, direct filesystem reads,
  and any redesign of the core projection or workspace ports without a demonstrated contract gap.

## Sequencing

First replacement slice for superseded task `6g3q4rv89vzw`. It follows the v0.18.0 CLI preview and
unblocks the read-only list/detail task; topology presentation remains behind real TUI dogfooding.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
- Supersedes part of [the original combined TUI slice](6g3q4rv89vzw-add-usage-informed-thread-views-to-the-tui.md)
