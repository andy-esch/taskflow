---
schema: 1
id: 6g86e10dztpf
status: in-progress
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
started_at: "2026-09-09"
---

# Make dense Thread graph routes visually trustworthy

## Objective

Turn the prototype's bounded edge router into a trustworthy dense-graph presentation. Preserve the
identity and direction of long, crossing, converging, and shared-lane dependencies so a route cannot
appear to terminate at the wrong task, pass through a node, or imply a junction absent from the
supplied `ThreadGraphProjection`.

## Acceptance criteria

- [x] Every rendered edge has a deterministic node-free route from its actual prerequisite to its
  actual dependent; intermediate arrowheads and boxes cannot visually terminate another edge.
- [x] Crossings, true junction-like shared geometry, parallel routes, fan-in, and fan-out remain
  distinguishable without inventing graph connectivity.
- [x] Selection highlighting identifies the complete incident route and endpoint under clipping and
  panning, while node boxes retain their semantic status colors.
- [x] External gates are positioned inside a valid dependency interval near the work they gate,
  without reversing supplied edges or misrepresenting member waves.
- [x] Dense, deep, wide, skipped-layer, same-row, reverse/cyclic-residue, cascaded external-gate,
  and narrow-terminal cases have deterministic layout and rendering regressions with bounded
  dimensions and runtime.
- [x] The work remains a TUI presentation adapter over `ThreadGraphProjection`; graph semantics,
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
- Implementation review briefs:
  [Claude](../audits/6g8e0pyczpns-2026-09-09-dense-thread-graph-routes-implementation-claude.md)
  and
  [Antigravity](../audits/6g8e0pyn885r-2026-09-09-dense-thread-graph-routes-implementation-antigravity.md)

## Sequencing

Follow the bounded correctness-hardening task, which repairs the prototype's missing deepest-row
track, and use that node-free route plus late external-gate placement as the baseline—not as proof
that dense routing is finished. This work may then proceed independently of the one-hop focus
subgraph and responsive layout: it owns whether edges remain identifiable and truthful at crossings
and viewport boundaries; responsive layout owns which complete node/layer units are visible. Prefer
testable routing invariants over committing in advance to any one researched technique such as
crossing gaps, waypoints, or outer-boundary skip routes.

## Implementation checkpoint (2026-09-09)

The spatial adapter now assigns deterministic per-edge boundary lanes and node-free inter-row
tracks before it allocates the canvas. Long and reverse routes retain their endpoints through
explicit waypoints and subdued corridors; unrelated perpendicular routes render as crossings rather
than false junctions; shared endpoint geometry and compact fan-in/fan-out counts remain distinct.
The selected task's incident routes stay magenta while semantic node colors and yellow direction
markers remain independent. When a selected incident edge is clipped, its visible boundary names
the offscreen endpoint alias instead of leaving an unattributed line. Panning also omits a route
whose two endpoint nodes are both outside the viewport, preventing unrelated offscreen work from
becoming a wall of unowned corridors while preserving every route incident to a visible node.

External-gate placement now settles a stable fixed point for cascaded gates, and column headings
identify presentation layers separately from core member waves. Regression fixtures cover dense
crossings and shared routes, equivalent-evidence permutations, skipped layers, the deepest route
track, same-layer cycles, a true reverse route, clipped endpoints, narrow terminals, capacity
fallbacks, supplied self-edges, a representative 20-node/76-edge layered graph, and routes
intersecting no node box.

Validation: `go test -race ./...`, repository-wide `go vet`, module-tidiness and generated-doc
checks, `golangci-lint run ./...` (with sandbox-local caches), `git diff --check`, and `tskflwctl
lint` all pass.

## Adversarial review closeout (2026-09-09)

The Claude and Antigravity implementation reviews correctly rejected several canvas shortcuts that
could still lie about connectivity. Route cells now carry a compact integer identity rather than
endpoint strings. Separate source and target lanes prevent unrelated turns from being flattened
into crossings; straight intersections stay non-junction crossings even when their routes share a
distant endpoint; a turn may use shared grammar only when it actually joins a common endpoint stub;
and unrelated collinear overlap has its own warning glyph. Production-path regressions pin every
arrow and corner in the minimal crossed-edge case and reject ambiguous collisions across the dense
20-node/76-edge fixture.

Counts are now separated by source/target role and entry side, with reserved cells that cannot erase
focus or self/reverse arrowheads. Clipped selected routes group all offscreen aliases and place the
annotation only in collision-free visible cells. Exact wave sets and the complete route glyph
grammar remain legible without claiming presentation layers are waves. The previously inert gate
fixture now requires a genuine second fixed-point pass, and dead route helpers are gone.

The canvas ceiling is reduced to 500,000 cells, with a regression keeping cell storage below 24
MiB. The remaining pre-layout work and repeated layout construction are deliberately tracked by
[cache and preflight spatial Thread layout work](6g8ezj5e51hg-cache-and-preflight-spatial-thread-layout-work.md),
sequenced immediately after this task in the production Thread. Both implementation audits are
closed with every finding either fixed here or linked to that follow-up.
