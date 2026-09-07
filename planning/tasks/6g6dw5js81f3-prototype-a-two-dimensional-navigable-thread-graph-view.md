---
schema: 1
id: 6g6dw5js81f3
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Design and prove a flagship full-screen spatial Thread graph with deterministic layout and direct hjkl node navigation.
effort: 3-5 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, tui, graph, ux, dogfood]
created: "2026-09-03"
depends_on: [6g5rwjr0dz4p, 6g6scc9jgxae, 6g7fhfpmy032]
updated_at: "2026-09-07"
---
# Prototype a two-dimensional navigable Thread graph view

## Objective

Design and prototype a full-screen, genuinely spatial Thread graph in which nodes occupy two
dimensions and `h`/`j`/`k`/`l` navigate by graph/layout adjacency. This is the flagship feature
candidate for the first feature release after v0.20.x, not low-priority polish. It remains separate
from the compact linear wave reader so its richer interaction and layout model can be nailed without
stretching detail-pane text into a pseudo-graph. The removable presentation boundary is risk
containment, not a signal that the experience is disposable or unimportant.

## Dogfood evidence

The shipped wave reader makes execution order legible, and task selection plus `f`/`j`/`k`
navigation make its linear rows usable. It still presents a complex Thread as a sequence of task
bags: the earlier inline `<node> -> <node>` list was removed as verbose without conveying spatial
structure, while fan-out and fan-in remain difficult to perceive. That is sufficient evidence to
authorize a high-priority spatial design/prototype immediately after the v0.20 checkpoint. The
prototype is the gate to production-grade slices; it is deliberately not a reason to delay v0.20.

## Design questions

- Define what left/right/up/down mean when edges skip waves, fan out, fan in, or cross.
- Decide whether layout is deterministic and taskflow-owned or delegated to a terminal graph-layout
  library.
- Choose an experimental integration boundary—optional CLI renderer, gated TUI destination,
  separate module/binary, or another narrow adapter—and state its discovery/distribution tradeoff.
- Preserve stable task identity, member/external-gate roles, direction, health, and incomplete
  topology without recomputing core graph semantics.
- Specify selection, Enter-to-open, back-stack behavior, scrolling/panning, zoom, narrow-terminal
  fallback, reload stability, and accessibility without conflicting with existing TUI keys.
- Establish performance and visual-complexity limits for large/deep/wide graphs before choosing a
  renderer.

## Acceptance criteria

- [x] Dogfood evidence from the linear wave view identifies concrete questions that require spatial
  presentation.
- [ ] A short design note defines spatial layout, deterministic ordering, hjkl neighbor selection,
  focus, panning, reload, and narrow-terminal behavior.
- [ ] A bounded prototype renders the existing `ThreadGraphProjection` without parsing Mermaid/DOT
  or deriving task readiness/scheduling semantics.
- [ ] The dependency direction points from the extension toward the stable projection/wire
      contract: core code does not import the experiment, default CLI/TUI paths do not require it,
      and removing it requires no planning-data migration.
- [ ] Experimental launch/discovery is explicit, and renderer/layout failure cannot corrupt or
      authorize changes to the planning graph.
- [ ] Direct task navigation and ctrl+o return use canonical task/Thread identities.
- [ ] Fan-out, fan-in, edge crossing, skipped waves, disconnected members, external gates, hostile
  labels, incomplete topology, and large graphs are stress-tested.
- [ ] The prototype produces enough evidence to define and sequence a
  production-grade spatial graph slice as the flagship candidate for the first
  feature release after v0.20.x, or records a specific reason to revise or
  abandon that direction.

## Out of scope

Production graph mutation, critical-path/slack/forecasting analysis, web rendering, replacing the
linear wave view before the prototype is evaluated, or committing to a general plugin framework.

## Sequencing

Begin after the compatibility-hardened v0.20 checkpoint is complete. At that boundary this becomes
the highest-priority Thread presentation initiative: a deliberate design and prototype gate for a
production-quality flagship, not an opportunistic experiment. If the evidence holds, split and
sequence the production renderer, interaction hardening, accessibility, and release work rather
than hiding them inside this prototype. It remains independent of the correctness-only Thread
graduation decision; product importance does not make it a retroactive v0.20 or graduation gate.

## Related

- Predecessor [linear Thread topology view](6g5rwjr0dz4p-add-dogfooded-thread-graph-presentation-to-the-tui.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
