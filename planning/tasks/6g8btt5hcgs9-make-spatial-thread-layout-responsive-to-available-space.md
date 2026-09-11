---
schema: 1
id: 6g8btt5hcgs9
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Make spatial Thread nodes, canvas, inspector, and surrounding pane budgets preserve identity and orientation across wide, ordinary, and narrow terminals.
effort: 2-3 days
tier: 2
priority: high
autonomy_level: 3
tags: [threads, tui, graph, ux, responsive]
created: "2026-09-09"
depends_on: [6g6dw5js81f3, 6g8by30btznq]
updated_at: "2026-09-11"
started_at: "2026-09-10"
completed_at: "2026-09-11"
---

# Make spatial Thread layout responsive to available space

## Objective

Make the spatial Thread presentation spend the terminal space it actually has on recognizable task
identity and complete visual units. Replace fixed node and pane assumptions with deterministic,
responsive budgets while preserving the full graph as a trustworthy flagship view and keeping all
layout decisions inside the TUI adapter over `ThreadGraphProjection`.

## Acceptance criteria

- [x] Node width and label treatment respond to the usable canvas width and visible layout layers:
  wider views expose materially more identity, while constrained views use an explicit code/legend
  or deterministic multi-line treatment instead of indistinguishable fixed-width ellipses.
- [x] Horizontal panning and clipping operate on complete node/layer units or render unmistakable
  boundary gutters; no partially visible node silently loses its status marker, identifying prefix,
  or border.
- [x] Vertical viewport budgeting accounts for complete node slots, headers, legends, and the focus
  inspector before choosing the visible window; the inspector cannot visually consume the bottom
  of a node row.
- [x] The list/detail split and full-screen presentation allocate width according to available space
  and content without starving Thread identity, current-view orientation, or useful graph detail.
- [x] Wide, ordinary, and narrow terminal regressions cover long and duplicate-prefix labels,
  scrolling in every direction, resize, reload, and selection restoration with deterministic output.
- [x] Layout remains a presentation-only transformation of the supplied projection and does not
  infer dependencies, readiness, waves, hidden repository nodes, or an automatic representation
  switch from an arbitrary graph-size threshold.

## Implementation checkpoint (2026-09-10)

Spatial layouts now choose an 18–42-cell node width from the usable viewport and a budget of up to
three concurrently visible layout layers. Wider terminals expose more of each task; constrained
nodes lead with their stable `[M#]` or `[G#]` code and preserve both ends of a long label. Geometry
is still cached per coherent projection: resize replaces the current immutable width bucket, while
navigation and same-size renders reuse it. If a preferred width would exceed the existing
layout-area guard, preparation selects the widest safe width instead of rejecting a graph that was
already supported.

The selection-centered window renders only complete node slots. Partial boundary nodes are omitted,
routes retain their clipped evidence, and gutter placement never overwrites a node or connector.
A dedicated viewport line now reports the number of complete nodes on screen plus visible layer and
row ranges with directional extent markers, making a three-node window into a much larger graph
explicit rather than relying on subtle edge glyphs. The minimum rendered height accounts for four
header/legend rows, one complete five-row node slot, and the five-row inspector. Below that bound,
an accented card reports the current and required dimensions and gives the fallback actions instead
of presenting a loose error paragraph. Thread summary and topology details also use a bounded
content-aware split: the list keeps an identity-safe floor, structured detail grows only to its
useful no-wrap width, and spatial mode continues to receive the full detail region.

Focused regressions cover long duplicate-prefix labels, every selected node across wide and narrow
viewports, horizontal and vertical gutters, safe gutter collisions, width-bucket cache reuse,
near-limit capacity negotiation, split/full-screen resize behavior, reload, and deterministic
navigation. Live dogfood passes at 80×24 and 140×36 used `complete-production-threads`; the full
race-enabled repository suite, lint, vet, and planning lint pass.

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
- Hands the larger navigation and representation question to
  [progressive-disclosure research](6g8vxbv3d4xn-explore-progressive-disclosure-and-navigation-for-large-thread-graphs.md)
- Coordinates shell-level discoverability with the
  [whole-TUI navigation and information-architecture review](6g8vxcnezktm-reassess-tui-navigation-and-information-architecture-at-current-scale.md)
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

## Adversarial review closeout (2026-09-11)

The independent [Claude](../audits/6g8w0pm8rztk-2026-09-10-responsive-spatial-thread-layout-implementation-claude.md) and [Antigravity](../audits/6g8w0pmhkd7n-2026-09-10-responsive-spatial-thread-layout-implementation-antigravity.md) reviews are closed with every demonstrated defect fixed. Their overlapping findings exposed four useful corrections: viewport placement now maximizes complete neighboring layers and rows instead of greedily excluding partial units; cold spatial entry preflights only width-independent ranking before materializing one responsive layout; empty and error detail states shed stale Thread pane preferences; and compact Unicode identities cannot overflow card borders.

Regression coverage now uses independent expectations for 18- and 42-cell geometry, the widest safe near-capacity width, literal 60x14 minimum boundaries, guarded gutter states, and brute-force-optimal layer/row visibility. The non-cache renderer shares the same capacity fallback as the production cache, and the responsive cached-render benchmark is valid again. The severity assigned to several performance and test-effectiveness findings was stronger than their user impact, but their underlying evidence was sound and required no architectural pivot.

Validation after remediation: the full race-enabled repository suite, go vet, golangci-lint, planning lint, documentation generation/check, spatial benchmarks, and git whitespace checks pass.
