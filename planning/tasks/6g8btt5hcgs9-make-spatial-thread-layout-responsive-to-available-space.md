---
schema: 1
id: 6g8btt5hcgs9
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Make spatial Thread nodes, canvas, inspector, and surrounding pane budgets preserve identity and orientation across wide, ordinary, and narrow terminals.
effort: 2-3 days
tier: 2
priority: high
autonomy_level: 3
tags: [threads, tui, graph, ux, responsive]
created: "2026-09-09"
depends_on: [6g6dw5js81f3, 6g8by30btznq]
updated_at: "2026-09-09"
---

# Make spatial Thread layout responsive to available space

## Objective

Make the spatial Thread presentation spend the terminal space it actually has on recognizable task
identity and complete visual units. Replace fixed node and pane assumptions with deterministic,
responsive budgets while preserving the full graph as a trustworthy flagship view and keeping all
layout decisions inside the TUI adapter over `ThreadGraphProjection`.

## Acceptance criteria

- [ ] Node width and label treatment respond to the usable canvas width and visible layout layers:
  wider views expose materially more identity, while constrained views use an explicit code/legend
  or deterministic multi-line treatment instead of indistinguishable fixed-width ellipses.
- [ ] Horizontal panning and clipping operate on complete node/layer units or render unmistakable
  boundary gutters; no partially visible node silently loses its status marker, identifying prefix,
  or border.
- [ ] Vertical viewport budgeting accounts for complete node slots, headers, legends, and the focus
  inspector before choosing the visible window; the inspector cannot visually consume the bottom
  of a node row.
- [ ] The list/detail split and full-screen presentation allocate width according to available space
  and content without starving Thread identity, current-view orientation, or useful graph detail.
- [ ] Wide, ordinary, and narrow terminal regressions cover long and duplicate-prefix labels,
  scrolling in every direction, resize, reload, and selection restoration with deterministic output.
- [ ] Layout remains a presentation-only transformation of the supplied projection and does not
  infer dependencies, readiness, waves, hidden repository nodes, or an automatic representation
  switch from an arbitrary graph-size threshold.

## Out of scope

- Replacing the edge router, solving crossing or shared-lane ambiguity, adding graph queries or
  mutations, adopting a matrix view, semantic zoom, animation, or changing one-hop focus behavior.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Builds on the
  [two-dimensional Thread graph prototype](6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md)
- Coordinates viewport boundaries with
  [dense-route trustworthiness](6g86e10dztpf-make-dense-thread-graph-routes-visually-trustworthy.md)
- Keeps the full graph complementary to the
  [one-hop focus subgraph](6g86c7y6hn41-add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph.md)
- Tracked from the
  [spatial Thread graph experience design review](../audits/6g8bbjgcx9tj-2026-09-09-spatial-thread-graph-tui-experience-design-review.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)

## Design attention

Do not treat node count alone as a representation cutoff. The dogfooded production Thread is a
sparse 40-node, 51-edge DAG, and its primary jobs are causal/path comprehension rather than generic
dense-network scanning. Test space allocation against real Thread shapes and viewport sizes. Reuse
the existing short node codes as an honest narrow-width identity channel, but do not make the focus
inspector the only place where a user can discover which task a node represents.

## Sequencing

This work proceeds after the bounded prototype-correctness hardening and may then run independently
of route hardening and one-hop focus. Keep responsibility crisp at their seam: this task owns which
complete node/layer units are visible and how much identity fits; route hardening owns whether edges
remain truthful at and across those viewport boundaries.
