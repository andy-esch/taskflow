---
schema: 1
id: 6g86c7y6hn41
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Zoom the spatial Thread graph into a selected task and its immediate prerequisites and dependents without losing whole-graph context or stable navigation.
effort: 2-3 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, tui, graph, ux, dogfood]
created: "2026-09-08"
depends_on: [6g6dw5js81f3]
updated_at: "2026-09-08"
---

# Add a one-hop focus subgraph mode to the TUI Thread graph

## Objective

Add a deliberate focus/zoom mode to the spatial Thread graph that shows the selected task and every
readable node exactly one supplied edge away: its immediate prerequisites and dependents. Make a
dense Thread locally comprehensible without turning the view into a new graph query engine or
losing the user's stable selection and route back to the whole graph.

## Acceptance criteria

- [ ] From a selected spatial node, one discoverable action enters a bounded view containing that
  node, its direct incoming neighbors, its direct outgoing neighbors, and only the supplied edges
  among those nodes.
- [ ] The focus view states how many nodes are shown versus hidden and preserves Thread graph and
  projection health; external gates and unreadable supplied nodes are not silently discarded.
- [ ] Entering, navigating, reloading, and leaving the focus view preserve canonical task identity;
  returning to the full graph restores the original selection and viewport context.
- [ ] `hjkl`, `Enter`, `y`, task following, and the eventual shared Back action have an explicit
  interaction contract rather than competing key or history behavior.
- [ ] Fan-in, fan-out, a leaf, a root, an isolated member, an external gate, incomplete topology,
  hostile labels, and narrow terminals have focused regression coverage.
- [ ] The implementation filters and lays out `ThreadGraphProjection` evidence in the TUI adapter;
  it neither mutates planning data nor derives readiness, transitive blockers, or new graph facts.

## Out of scope

- Arbitrary-depth expansion, recursive repository-wide blocker queries, persisted view state,
  critical-path analysis, or replacing the full Thread graph and linear wave reader.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Builds on the
  [two-dimensional Thread graph prototype](6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md)
- Complements the
  [reusable TUI Back action](6g86016qvk5d-add-a-reusable-back-action-for-tui-entity-navigation.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)

## Design attention

Evaluate using the otherwise redundant spatial-mode `z` binding as a local whole-graph/focus toggle,
but do not lock that key without checking the shared shell contract. The focused layout should make
incoming versus outgoing sides obvious and expose hidden-neighbor counts or boundary cues where
one-hop context ends. A focus subgraph is a presentation projection over the already bounded Thread
projection, not permission to walk deeper repository dependencies behind core's back.
