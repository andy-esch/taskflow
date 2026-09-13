---
schema: 1
id: 6g86c7y6hn41
status: in-progress
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
started_at: "2026-09-12"
audited: "2026-09-12"
audit_sources: [planning/audits/6g9fg4a8xvyc-2026-09-12-tui-one-hop-thread-focus-implementation-claude.md, planning/audits/6g9fg4awm2cg-2026-09-12-tui-one-hop-thread-focus-implementation-antigravity.md]
---

# Add a one-hop focus subgraph mode to the TUI Thread graph

## Objective

Add a deliberate focus/zoom mode to the spatial Thread graph that shows the selected task and every
readable node exactly one supplied edge away: its immediate prerequisites and dependents. Make a
dense Thread locally comprehensible without turning the view into a new graph query engine or
losing the user's stable selection and route back to the whole graph.

## Acceptance criteria

- [x] From a selected spatial node, discoverable `z` replaces the main canvas with the shared
  one-hop selection containing that node, its direct incoming and outgoing neighbors, and only the
  supplied edges among those nodes; `z` retains pane-zoom behavior outside immersive spatial mode.
- [x] A small bordered scope card overlays an empty canvas corner, yields rather than obscuring
  topology, and states one-hop mode, shown/hidden counts, and return controls. The focus graph itself
  does not become a modal. Graph/projection health, external gates, and unreadable supplied nodes
  are not silently discarded.
- [x] First spatial entry centers an in-flight member, otherwise an eligible frontier member,
  otherwise the stable fallback. Entering focus, navigating, reloading, and leaving preserve
  canonical identity; returning to the full graph restores its exact selection and viewport.
- [x] `h/l` follows direct prerequisites/dependents: one target moves immediately, multiple targets
  open a compact directional neighbor chooser, and no target reports an explicit dead end. `j/k`,
  `f`, `Enter`, `y`, `Esc`, `ctrl+o`, and the eventual shared Back action retain the selected
  interaction contract without competing meanings.
- [x] Fan-in, fan-out, a leaf, a root, an isolated member, an external gate, incomplete topology,
  hostile labels, and narrow terminals have focused regression coverage.
- [x] The TUI consumes the adapter-neutral bounded selector established by the neighborhood-export
  task and lays it out locally; it neither mutates planning data nor derives readiness, transitive
  blockers, or new graph facts.

## Out of scope

- Arbitrary-depth expansion, recursive repository-wide blocker queries, persisted view state,
  critical-path analysis, or replacing the full Thread graph and linear dependency-rank reader.

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

## Implementation progress (2026-09-12)

The immersive spatial view now uses `z` to toggle the shared core one-hop selector around its
readable selected node. Full and focused projections retain separate lazy layout caches; leaving
focus restores the original stable selection and its derived viewport. A generic opaque detail
context carries focal/full-selection state through coherent reloads and shell-owned `ctrl+o`
navigation without teaching the shell about Threads.

Focused output keeps source graph/projection health and exact scope counts visible. A bordered
scope card searches empty canvas corners and yields instead of covering graph ink, while the fixed
header retains boundary and return guidance when no corner fits. Initial spatial entry prefers an
in-flight member, then the supplied Thread frontier, then a deterministic fallback.

Within focus, horizontal navigation follows only supplied incoming/outgoing edges. A single
readable neighbor moves immediately, fan-in/fan-out opens a compact shell-owned chooser, and a
missing readable target reports an explicit dead end. Existing vertical movement, picker, open,
yank, retreat, and follow/back behavior remains intact. Focused regression coverage exercises
branches, external gates, unreadable neighbors, isolated work, incomplete topology, narrow
terminals, reloads, and navigation round trips. The full repository suite, race-enabled TUI suite,
`go vet`, and `golangci-lint` pass.

Dogfooding the finished focus view exposed an overloaded presentation term: a core wave is a
topological generation, not a lifecycle phase or execution barrier. The stable projection retains
its `waves` contract, while the TUI now presents those values as **dependency ranks**, explains
their prerequisite-depth meaning, and continues to show lifecycle state independently.

The same pass made local zoom state visually persistent rather than relying on easily clipped
header prose or an opportunistic scope card. Every focused spatial render now leads with a
high-contrast `ZOOMED · ONE-HOP` badge; the collision-aware card and constrained/capacity fallbacks
repeat the textual signal, while the full graph never advertises the bounded mode.

A subsequent real-Thread capture still opened on the earliest historical external gate while the
sole in-flight member was off-screen, so the entry criterion was reopened rather than treated as
satisfied by unit-level preference alone. Deliberate entry now re-anchors to working context; reload
and follow/back restoration still preserve an explicit stable selection. Within each tier,
`updated_at` (falling back to `created`) chooses the freshest member: in-flight first, then supplied
frontier, then any readable member. External gates are the last resort when no member is readable.
After rebuilding, the maintainer confirmed that the production Thread centers the in-flight focus
task on entry, closing the criterion with real dogfood evidence.

## Adversarial review hardening (2026-09-12)

Independent [Claude](../audits/6g9fg4a8xvyc-2026-09-12-tui-one-hop-thread-focus-implementation-claude.md)
and [Antigravity](../audits/6g9fg4awm2cg-2026-09-12-tui-one-hop-thread-focus-implementation-antigravity.md)
reviews found real state and presentation gaps after the initial implementation passed its tests.
Directional branch choices are now re-derived when committed and dismissed on a coherent refresh;
the shell routes `z` to the one-hop lens only while that immersive graph is visibly active.
Opaque context restoration requires the exact readable focal identity, so deletion cannot fall
through the public fuzzy reference resolver.

Focused layouts retain the full graph's `[M#]` and `[G#]` aliases. The scope card can use blank
terminal space outside compact layout bounds without exceeding the canvas allocation ceiling, and
the fixed header keeps topology and both health verdicts ahead of optional viewport detail.
Constrained views name `z` as the route back. Entry selection validates portable activity dates,
examines all supplied members even when the full renderer hits its capacity guard, and prefers a
readable external prerequisite over unreadable evidence only after exhausting readable members.

Regression coverage now drives the real render, reload, history, view-cycle, shell-zoom, stale
chooser, alias, activity, capacity, and focal-deletion seams. The stale demo recording and older
record-limit mutation gaps remain meaningful but are explicitly sequenced in
[refresh-the-Threads-demo](6g9g37bvetjy-refresh-the-threads-demo-after-the-tui-navigation-redesign.md)
and [pin-Thread-graph-input-record-limits](6g9g37c50qp1-pin-thread-graph-input-record-limits-at-core-and-tui-boundaries.md).
