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

## Stress tests

- Watcher reload during mutations, completed-thread drift, missing/broken members, large Thread lists, graph overflow, small terminals, and CLI/TUI parity fixtures.

## Sequencing

Last planned V1 slice, after real CLI usage of Thread projections and generated views.
