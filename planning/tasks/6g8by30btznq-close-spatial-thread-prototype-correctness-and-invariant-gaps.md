---
schema: 1
id: 6g8by30btznq
status: in-progress
epic: 30-threads-and-task-dependency-graphs
description: Fix the concrete route and Atlas immersion bugs from implementation review, then pin ordering, capacity, and manual-zoom invariants before spatial feature work continues.
effort: 1-2 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, tui, graph, correctness, tests]
created: "2026-09-09"
updated_at: "2026-09-09"
depends_on: [6g6dw5js81f3]
started_at: "2026-09-09"
---

# Close spatial Thread prototype correctness and invariant gaps

## Objective

Close the concrete correctness and regression gaps found after the first spatial Thread prototype
merged, before building more interaction and layout behavior on top of them. Keep the pass narrowly
focused on preserving supplied edges, shell-owned zoom state, deterministic projection consumption,
and the resource guards already claimed by the prototype.

## Acceptance criteria

- [x] A skipped-layer edge whose shallower endpoint occupies the deepest layout row retains its
  complete horizontal track inside the canvas; its inspector relationship and rendered route cannot
  contradict each other.
- [x] Entering and exiting Atlas from an automatically immersive spatial graph preserves ownership
  of that zoom so retreating to waves restores the split, while an explicitly user-entered zoom is
  never consumed by spatial or Atlas transitions.
- [x] Spatial row placement, aliases, inspector connection order, and rendered output are invariant
  under equivalent permutations of projection node, edge, and wave slices, or the projection-order
  precondition is made explicit and enforced at one adapter boundary with equivalent regression
  strength.
- [x] Churning the node, edge, and canvas-cell capacity limits is killed by focused tests that reach
  each guard independently below the other limits; the narrow no-canvas path remains intentionally
  and visibly exempt.
- [x] Focused shell tests pin both manual-zoom preservation and `z` remaining inert while the spatial
  presentation owns immersion.
- [x] The fixes remain presentation/shell concerns over the supplied `ThreadGraphProjection`; no
  persisted data, graph semantics, readiness, or repository mutation behavior changes.

## Out of scope

- Dense crossing aesthetics, shared-lane routing, cascaded external-gate placement, responsive label
  and pane budgets, one-hop focus, alternate-view discoverability, or a general shell state rewrite.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Builds on the
  [two-dimensional Thread graph prototype](6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md)
- Reviewed by the
  [Claude implementation audit](../audits/6g8bb1m753yw-2026-09-09-spatial-thread-graph-prototype-implementation-claude.md)
  and
  [Antigravity implementation audit](../audits/6g8bb1mf9hw0-2026-09-09-spatial-thread-graph-prototype-implementation-antigravity.md)
- Unblocks
  [dense-route trustworthiness](6g86e10dztpf-make-dense-thread-graph-routes-visually-trustworthy.md),
  [responsive spatial layout](6g8btt5hcgs9-make-spatial-thread-layout-responsive-to-available-space.md),
  and
  [one-hop focus](6g86c7y6hn41-add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)

## Finding triage

Claude H1 and Antigravity H1 are reproduced user-visible defects. Claude M1/M2 are coverage gaps
around active production guards and shell ownership rules. Antigravity M1 is not user-reachable from
the current core producer because it sorts projection evidence, but normalizing or enforcing that
contract at the TUI boundary is cheap insurance for future portable producers and matches the
prototype's deterministic-equivalent-evidence promise. Antigravity L1 remains with dense-route
hardening because it concerns placement quality for chained external gates rather than this bounded
correctness closeout.

## Sequencing

Run this before dense-route hardening, responsive spatial layout, and one-hop focus so those tasks
start from a route-complete, deterministic, shell-state-safe baseline. General Thread/Atlas view
discoverability and reusable Back/action-target work may continue independently.

## Implementation closeout (2026-09-09)

The spatial renderer now reserves the deepest inter-row routing track, canonicalizes presentation
ordering independently of projection slice order, and exercises the node, edge, canvas-cell, and
narrow-terminal guard paths. Atlas snapshots and per-space sessions retain both the zoom value and
whether an immersive presentation owns it; manual zoom remains user-owned and `z` stays inert inside
the spatial graph.

Validation passed with `just build`, `just test` (including the race detector), `just lint` (0
issues), focused TUI tests, `git diff --check`, and `tskflwctl lint`. The change remains within the
TUI presentation and shell-session layers; core graph and repository semantics are unchanged.
