---
schema: 1
id: 6g6dw5js81f3
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Explore a full-screen spatial Thread graph with direct hjkl node navigation after the linear wave view has accumulated enough dogfood evidence.
effort: 3-5 days
tier: 4
priority: low
autonomy_level: 3
tags: [threads, tui, graph, ux, dogfood]
created: "2026-09-03"
depends_on: [6g5rwjr0dz4p]
updated_at: "2026-09-03"
---
# Prototype a two-dimensional navigable Thread graph view

## Objective

Explore and prototype a full-screen, genuinely spatial Thread graph in which nodes occupy two
dimensions and `h`/`j`/`k`/`l` navigate by graph/layout adjacency. This is deliberately separate
from the compact linear wave reader: it should earn a richer interaction and layout model from
dogfood evidence rather than stretching detail-pane text into a pseudo-graph.

## Design questions

- Define what left/right/up/down mean when edges skip waves, fan out, fan in, or cross.
- Decide whether layout is deterministic and taskflow-owned or delegated to a terminal graph-layout
  library.
- Preserve stable task identity, member/external-gate roles, direction, health, and incomplete
  topology without recomputing core graph semantics.
- Specify selection, Enter-to-open, back-stack behavior, scrolling/panning, zoom, narrow-terminal
  fallback, reload stability, and accessibility without conflicting with existing TUI keys.
- Establish performance and visual-complexity limits for large/deep/wide graphs before choosing a
  renderer.

## Acceptance criteria

- [ ] Dogfood evidence from the linear wave view identifies concrete questions that require spatial
  presentation.
- [ ] A short design note defines spatial layout, deterministic ordering, hjkl neighbor selection,
  focus, panning, reload, and narrow-terminal behavior.
- [ ] A bounded prototype renders the existing `ThreadGraphProjection` without parsing Mermaid/DOT
  or deriving task readiness/scheduling semantics.
- [ ] Direct task navigation and ctrl+o return use canonical task/Thread identities.
- [ ] Fan-out, fan-in, edge crossing, skipped waves, disconnected members, external gates, hostile
  labels, incomplete topology, and large graphs are stress-tested.
- [ ] The prototype produces enough evidence to accept a production slice, revise the
  projection/layout boundary, adopt a suitable library, or abandon the spatial view.

## Out of scope

Production graph mutation, critical-path/slack/forecasting analysis, web rendering, or replacing
the linear wave view before the prototype is evaluated.

## Sequencing

Follow the linear TUI topology task. Keep this task queued until dogfooding shows that compact waves
plus inline prerequisites cannot answer an important planning question.

## Related

- Predecessor [linear Thread topology view](6g5rwjr0dz4p-add-dogfooded-thread-graph-presentation-to-the-tui.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
