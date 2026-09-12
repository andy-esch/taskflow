---
schema: 1
id: 6g86c7y6hn41
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Zoom the spatial Thread graph into a selected task and its immediate prerequisites and dependents without losing whole-graph context or stable navigation.
effort: 3-5 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, tui, graph, ux, dogfood]
created: "2026-09-08"
depends_on: [6g6dw5js81f3, 6g8by30btznq, 6g8vxbv3d4xn, 6g9150nrt4p9]
updated_at: "2026-09-12"
---

# Add a one-hop focus subgraph mode to the TUI Thread graph

## Objective

Add a deliberate focus/zoom mode to the spatial Thread graph that shows the selected task and every
readable node exactly one supplied edge away: its immediate prerequisites and dependents. Make a
dense Thread locally comprehensible without turning the view into a new graph query engine or
losing the user's stable selection and route back to the whole graph.

## Acceptance criteria

- [ ] From a selected spatial node, discoverable `z` replaces the main canvas with the shared
  one-hop selection containing that node, its direct incoming and outgoing neighbors, and only the
  supplied edges among those nodes; `z` retains pane-zoom behavior outside immersive spatial mode.
- [ ] A small bordered scope card overlays an empty canvas corner, yields rather than obscuring
  topology, and states one-hop mode, shown/hidden counts, and return controls. The focus graph itself
  does not become a modal. Graph/projection health, external gates, and unreadable supplied nodes
  are not silently discarded.
- [ ] First spatial entry centers an in-flight member, otherwise an eligible frontier member,
  otherwise the stable fallback. Entering focus, navigating, reloading, and leaving preserve
  canonical identity; returning to the full graph restores its exact selection and viewport.
- [ ] `h/l` follows direct prerequisites/dependents: one target moves immediately, multiple targets
  open a compact directional neighbor chooser, and no target reports an explicit dead end. `j/k`,
  `f`, `Enter`, `y`, `Esc`, `ctrl+o`, and the eventual shared Back action retain the selected
  interaction contract without competing meanings.
- [ ] Fan-in, fan-out, a leaf, a root, an isolated member, an external gate, incomplete topology,
  hostile labels, and narrow terminals have focused regression coverage.
- [ ] The TUI consumes the adapter-neutral bounded selector established by the neighborhood-export
  task and lays it out locally; it neither mutates planning data nor derives readiness, transitive
  blockers, or new graph facts.

## Out of scope

- Arbitrary-depth expansion, recursive repository-wide blocker queries, persisted view state,
  critical-path analysis, or replacing the full Thread graph and linear wave reader.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Builds on the
  [two-dimensional Thread graph prototype](6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md)
- Complements the
  [reusable TUI Back action](6g86016qvk5d-add-a-reusable-back-action-for-tui-entity-navigation.md)
- Builds on [bounded Thread neighborhood exports](6g9150nrt4p9-export-bounded-thread-neighborhoods-around-a-task.md).
- Thread [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)
- Tracked from the
  [spatial Thread graph experience design review](../audits/6g8bbjgcx9tj-2026-09-09-spatial-thread-graph-tui-experience-design-review.md)

## Design attention

The maintainer selected spatial-mode `z` as the local whole-graph/focus toggle after checking the
shared shell contract. Use the main canvas for the focused graph and a small collision-aware overlay
only for scope and return guidance; a modal graph cannot honestly contain large one-hop fans.
Incoming versus outgoing sides stay obvious, and boundary evidence discloses where one-hop context
ends. Preserve semantic `h/l` walks with a directional chooser at branches and explicit dead-end
feedback. Focus remains a complementary causal-inspection lens: neither node count alone nor the
cited design research justifies replacing or automatically suppressing the full sparse-DAG view.
