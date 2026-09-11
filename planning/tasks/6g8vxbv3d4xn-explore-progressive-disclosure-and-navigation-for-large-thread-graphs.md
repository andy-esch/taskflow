---
schema: 1
id: 6g8vxbv3d4xn
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Research scalable overview, orientation, and navigation for large Thread DAGs, including the honest TUI-versus-web boundary.
effort: 2-3 days
tier: 1
priority: high
autonomy_level: 2
tags: [threads, tui, graph, ux, research, design]
created: "2026-09-10"
depends_on: [6g8btt5hcgs9]
updated_at: "2026-09-10"
---
# Explore progressive disclosure and navigation for large Thread graphs

## Objective

Determine how a large Thread DAG should support overview, orientation, causal inspection, and
movement without presenting a misleading handful of visible nodes as though they were the graph.
Use the production dogfood Thread as a real stress case, and decide which interactions belong in a
terminal UI versus a future web or extension surface before committing to more graph chrome.

## Acceptance criteria

- [ ] An evidence matrix evaluates representative sparse, deep, wide, fan-in/fan-out, and degraded
  Threads at narrow, ordinary, and wide terminal sizes against concrete user jobs: find current
  work, understand sequence, follow an edge, locate a known task, and regain whole-graph context.
- [ ] At least three coherent presentation strategies are compared, including an improved full
  graph, overview-plus-focus/progressive disclosure, and a richer web or extension surface; each
  records comprehension, navigation, implementation, accessibility, and portability tradeoffs.
- [ ] The proposed navigation model covers edge-following, spatial movement, search/picker jumps,
  off-screen neighbors, viewport extent, focus/return, and selection history without assigning the
  same key multiple incompatible meanings.
- [ ] Semantic zoom or progressive-disclosure levels state exactly which supplied nodes and edges
  they show, how hidden evidence is disclosed, and how the user returns to the complete projection.
- [ ] The recommendation preserves `ThreadGraphProjection` as the graph-fact boundary and explains
  which projection, layout, and navigation pieces a TUI and web adapter can share without making
  terminal geometry part of core.
- [ ] Text wireframes or a disposable prototype are exercised against the production Thread, then
  reviewed with the maintainer before a presentation strategy is selected.
- [ ] Resulting implementation slices are filed, dependency-sequenced, and added to the production
  Threads dogfood graph; this task does not silently turn the chosen direction into production UI.

## Design attention

Start from the observed failure mode: an initial viewport showing only a few named nodes gave no
prominent indication that dozens remained, while vertical movement selected a hidden node whose
edge led farther off-screen. Consider minimaps, semantic zoom, overview/focus pairing, neighbor
counts and bearings, breadcrumbs, jump/search palettes, edge-follow navigation, and an explicit
viewport inventory. Do not assume either that one giant node-link canvas must remain the default or
that a browser automatically solves graph comprehension.

## Out of scope

- Production implementation, changing dependency semantics, critical-path/scheduling analysis,
  replacing the graph core, or selecting a graph library without a demonstrated missing primitive.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Follow the
  [responsive spatial layout](6g8btt5hcgs9-make-spatial-thread-layout-responsive-to-available-space.md)
- Inform the
  [one-hop focus subgraph](6g86c7y6hn41-add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph.md)
- Coordinate with the
  [whole-TUI navigation and information-architecture review](6g8vxcnezktm-reassess-tui-navigation-and-information-architecture-at-current-scale.md).
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
