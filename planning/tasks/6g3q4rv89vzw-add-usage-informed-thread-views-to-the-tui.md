---
schema: 1
id: 6g3q4rv89vzw
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Expose production-proven Thread lifecycle, rollup, frontier, gates, and graph information through read-oriented TUI views.
effort: 3-5 days
tier: 3
priority: medium
autonomy_level: 3
tags: [threads, tui, graph, ux]
created: "2026-08-25"
updated_at: "2026-08-27"
depends_on: [6g3q4rv1w9e2]
---
# Add usage-informed Thread views to the TUI

## Objective

Expose production-proven Thread lifecycle, rollup, frontier, member/external roles, and readable graph information in the TUI without creating a second graph engine.

## Scope

- Add Thread list/navigation and a detail view using existing core/wire projections.
- Begin read-oriented; direct graph editing is not required for the first TUI slice.
- Incorporate CLI dogfood feedback before freezing layout and interaction choices.

## Acceptance criteria

- [ ] TUI state matches CLI/core projections for lifecycle, progress, frontier, gates, and inconsistency.
- [ ] Watcher reload handles Thread and task dependency changes coherently.
- [ ] Narrow terminals degrade readably and navigation/completion remain consistent with other first-class entities.
- [ ] No TUI-local readiness, graph traversal, or direct filesystem mutation logic is introduced.
- [ ] Planner-phase ErrConflict from canonical-root callback exclusion is
  treated as transient watcher contention and retried or debounced without
  rendering an empty or permanently broken view.

## Stress tests

- Watcher reload during mutations, completed-thread drift, missing/broken members, large Thread lists, graph overflow, small terminals, and CLI/TUI parity fixtures.

## Sequencing

Last planned V1 slice, after real CLI usage of Thread projections and generated views.

## Mutation-contention amendment (2026-08-27)\n\nThe graph-mutation planner phase excludes every Store call at canonical-root scope. A watcher refresh that lands in that brief window can therefore receive ErrConflict even though no repository defect exists. Treat that result as transient contention: retain the last coherent model, debounce or retry after the mutation events settle, and never replace the view with empty data or a permanent broken-state banner.\n\nA multi-file mutation may also emit several watcher events for graph-valid durable prefixes. Coalesce them and reload the shared core projection after the burst rather than deriving intermediate graph state in the TUI.
