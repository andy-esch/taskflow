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
depends_on: [6g6dw5js81f3]
updated_at: "2026-09-08"
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
- [ ] Dense, deep, wide, skipped-layer, same-row, reverse/cyclic-residue, and narrow-terminal cases
  have deterministic layout and rendering regressions with bounded dimensions and runtime.
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

## Sequencing

Follow the prototype and use its node-free track plus late external-gate placement as the baseline,
not as proof that dense routing is finished. It may proceed independently of the one-hop focus
subgraph: one hardens whole-graph truthfulness, while the other adds a deliberately bounded view.
