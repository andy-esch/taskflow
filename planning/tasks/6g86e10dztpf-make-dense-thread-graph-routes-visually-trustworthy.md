---
schema: 1
id: 6g86e10dztpf
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Make dense Thread graphs preserve edge identity across long routes, crossings, shared lanes, and viewport clipping without visually inventing dependencies.
effort: 2-3 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, tui, graph, ux, dogfood]
created: "2026-09-08"
depends_on: [6g6dw5js81f3, 6g8by30btznq]
updated_at: "2026-09-09"
---

# Make dense Thread graph routes visually trustworthy

## Objective

Turn the prototype's bounded edge router into a trustworthy dense-graph presentation. Preserve the
identity and direction of long, crossing, converging, and shared-lane dependencies so a route cannot
appear to terminate at the wrong task, pass through a node, or imply a junction absent from the
supplied `ThreadGraphProjection`.

## Acceptance criteria

- [ ] Every rendered edge has a deterministic node-free route from its actual prerequisite to its
  actual dependent; intermediate arrowheads and boxes cannot visually terminate another edge.
- [ ] Crossings, true junction-like shared geometry, parallel routes, fan-in, and fan-out remain
  distinguishable without inventing graph connectivity.
- [ ] Selection highlighting identifies the complete incident route and endpoint under clipping and
  panning, while node boxes retain their semantic status colors.
- [ ] External gates are positioned inside a valid dependency interval near the work they gate,
  without reversing supplied edges or misrepresenting member waves.
- [ ] Dense, deep, wide, skipped-layer, same-row, reverse/cyclic-residue, cascaded external-gate,
  and narrow-terminal cases have deterministic layout and rendering regressions with bounded
  dimensions and runtime.
- [ ] The work remains a TUI presentation adapter over `ThreadGraphProjection`; graph semantics,
  readiness, and persisted planning data stay unchanged.

## Out of scope

- Adding graph mutations, critical-path calculations, arbitrary repository traversal, or choosing a
  third-party layout engine without evidence that the bounded router cannot meet this contract.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Discovered by the false `G1 → M1` route during dogfooding of the
  [two-dimensional Thread graph prototype](6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
- Coordinates viewport boundaries with
  [responsive spatial layout](6g8btt5hcgs9-make-spatial-thread-layout-responsive-to-available-space.md)
- Tracked from the
  [spatial Thread graph experience design review](../audits/6g8bbjgcx9tj-2026-09-09-spatial-thread-graph-tui-experience-design-review.md)
- Also owns the chained external-gate placement finding from the
  [Antigravity implementation audit](../audits/6g8bb1mf9hw0-2026-09-09-spatial-thread-graph-prototype-implementation-antigravity.md)

## Sequencing

Follow the bounded correctness-hardening task, which repairs the prototype's missing deepest-row
track, and use that node-free route plus late external-gate placement as the baseline—not as proof
that dense routing is finished. This work may then proceed independently of the one-hop focus
subgraph and responsive layout: it owns whether edges remain identifiable and truthful at crossings
and viewport boundaries; responsive layout owns which complete node/layer units are visible. Prefer
testable routing invariants over committing in advance to any one researched technique such as
crossing gaps, waypoints, or outer-boundary skip routes.
