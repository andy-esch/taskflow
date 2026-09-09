---
schema: 1
id: 6g6dw5js81f3
status: in-progress
epic: 30-threads-and-task-dependency-graphs
description: Design and prove a flagship full-screen spatial Thread graph with deterministic layout and direct hjkl node navigation.
effort: 3-5 days
tier: 1
priority: high
autonomy_level: 3
tags: [threads, tui, graph, ux, dogfood]
created: "2026-09-03"
depends_on: [6g5rwjr0dz4p, 6g6scc9jgxae, 6g7fhfpmy032]
updated_at: "2026-09-08"
started_at: "2026-09-08"
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
- [x] A short design note defines spatial layout, deterministic ordering, hjkl neighbor selection,
  focus, panning, reload, and narrow-terminal behavior.
- [x] A bounded prototype renders the existing `ThreadGraphProjection` without parsing Mermaid/DOT
  or deriving task readiness/scheduling semantics.
- [x] The dependency direction points from the extension toward the stable projection/wire
      contract: core code does not import the experiment, default CLI/TUI paths do not require it,
      and removing it requires no planning-data migration.
- [x] Experimental launch/discovery is explicit, and renderer/layout failure cannot corrupt or
      authorize changes to the planning graph.
- [x] Direct task navigation and ctrl+o return use canonical task/Thread identities.
- [x] Fan-out, fan-in, edge crossing, skipped waves, disconnected members, external gates, hostile
  labels, incomplete topology, and large graphs are stress-tested.
- [x] The prototype produces enough evidence to define and sequence a
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

## Prototype design decision (2026-09-08)

The first usable prototype is a third Thread detail presentation, cycled as
`summary → topology → spatial`. Entering it gives the graph the full TUI content region; `Esc`
returns to the topology/wave reader, while `v` continues the presentation cycle.

- Dependencies flow left to right in deterministic presentation layers computed only from the
  projection's supplied nodes and edges. Column labels retain the core member-wave number while
  allowing external gates to appear between member waves; cyclic or unranked residue remains
  clearly qualified as partial topology. The renderer never parses Mermaid/DOT or recomputes
  readiness, membership, health, or scheduling semantics.
- Node and edge order is deterministic: projection wave order first, then canonical task ID for
  presentation-only ties. `h`/`l` first follow an actual prerequisite/dependent edge in that
  direction, even when it skips a populated visual column. Multiple qualifying edges prefer the
  nearest connected column, then row distance; only a node with no edge in that direction falls
  back to the nearest populated column. `j`/`k` select the next/previous node in the current column.
  Remaining geometric ties use projection order, then canonical ID.
- Compact nodes show the shared status glyph/color and a bounded label. A visible legend explains
  status and role markers. A fixed inspector shows the selected node's full name, canonical ID,
  description, role, graph state, and direct connections.
- Selection is canonical-ID based and survives reload when that ID remains readable; otherwise it
  falls back to the first deterministic node. Navigation redraws a bounded viewport around the
  selection, keeping it visible without storing graph coordinates in planning data. Narrow
  terminals show an explicit size hint and retain the wave reader one `Esc` away. There is no zoom
  model in this prototype.
- `Enter` opens the selected task through the existing stable-identity navigation seam and
  `ctrl+o` returns to the Thread. Renderer/layout failure is display-only and cannot authorize or
  mutate graph state.

The prototype keeps a small taskflow-owned layered presentation adapter over
`core.ThreadGraphProjection`. General graph containers do not help because topology and waves are
already supplied by core; D2 and Graphviz wrappers target richer image/DOT rendering and add a much
larger dependency/runtime surface than a terminal cell layout warrants. The adapter remains
replaceable if stress testing later justifies a dedicated layout engine.

## First usable prototype checkpoint (2026-09-08)

The TUI now cycles `summary → topology → spatial`. Spatial mode takes the full detail region,
renders status-colored compact nodes and supplied dependency edges in deterministic left-to-right
layers, keeps a legend and selected-node inspector fixed around the panned graph canvas, and uses
`hjkl` for spatial navigation. `Enter` opens a readable selected task by canonical ID;
`ctrl+o` restores the same Thread, spatial presentation, and node; `Esc` returns to waves.

The first PTY dogfood pass against `complete-production-threads` exposed that an external gate may
sit between member waves rather than before the entire Thread. The layout was corrected to rank all
supplied projection nodes and edges while retaining core wave numbers in column labels; a focused
regression pins member → external gate → member ordering. No third-party graph library was adopted:
the core already owns graph semantics/waves, while D2 or embedded Graphviz would add a large image
rendering/runtime surface for a terminal-cell placement problem. The adapter remains isolated in
the TUI and replaceable.

Validation at this checkpoint: full `go test -race ./...`, full `golangci-lint run ./...`, planning
lint, and a live 140×36 PTY pass over the 35-node/46-edge dogfood Thread. The broad stress/evaluation
criterion remains open pending hands-on visual feedback; this is intentionally the first usable
prototype, not the final renderer decision.

## Hands-on feedback refinement (2026-09-08)

The first interactive passes found useful product evidence rather than a reason to discard the
spatial direction. Focus was too cyan/subtle, route bends collapsed into ambiguous plus signs,
horizontal movement lacked a clear contract between graph and geometric adjacency, and the fixed
inspector was informative but visually flat. The refined prototype therefore:

- keeps every node box in its semantic status color across focus changes, uses the active palette's
  magenta/mauve accent for touching connector strokes, and uses the established yellow active-work
  treatment for focus pointers and dependency arrowheads;
- composes connector cells from directional geometry, producing real elbows and tees while
  reserving `┼` for a genuine four-way crossing;
- makes `h`/`l` graph-first: direct prerequisite/dependent edges outrank unrelated boxes in the
  adjacent visual column, while the closest connected column and geometric row break fan-in/fan-out
  ties; spatial column movement remains the fallback when no edge exists that way;
- places an external boundary gate in the latest safe layer before its first dependent rather than
  grouping every source gate beside unrelated wave-one roots, and routes remaining multi-column
  edges along node-free inter-row tracks;
- reserves a stable five-row slot per node so the selected compact box can expand in place with its
  alias, status, and graph role/gate; and
- turns the bottom inspector into a palette-accented focus card with clearer status, state,
  relationship, identity, and description hierarchy.

The same pass exposed a separate navigation-system gap: following a graph node should offer an
obvious Back action, not rely only on knowledge of `ctrl+o`. That work is intentionally tracked in
[a reusable TUI Back-action task](6g86016qvk5d-add-a-reusable-back-action-for-tui-entity-navigation.md),
sequenced after this prototype and dogfooded in the same Thread rather than being hidden inside the
renderer.

Structured-detail selection now also owns `y`: with focus inside topology or spatial presentation,
the highlighted task slug is copied rather than the parent Thread slug; list focus continues to
copy the Thread, and the detail footer advertises the target. The broader parent-versus-child
contract for path, editor, lifecycle, and follow actions is tracked in
[structured-detail action targeting](6g86g03zfj2f-define-consistent-action-targets-for-structured-tui-detail-selections.md)
rather than being guessed inside this renderer. Dogfooding also identified a useful next level of detail—a bounded view of the
selected node and everything one edge away—which is tracked as the
[one-hop focus subgraph task](6g86c7y6hn41-add-a-one-hop-focus-subgraph-mode-to-the-tui-thread-graph.md)
rather than expanding this first renderer checkpoint without an interaction contract.

The same walkthrough also showed that reaching the spatial view still depends on already knowing
the unobtrusive `v` binding, while the Thread list leads with compact status/health syntax and can
truncate the identifying slug. Atlas has the same hidden alternate-view problem. That shared
discoverability and focused-row design work is tracked in
[alternate TUI views and Thread identity](6g87qn72901g-make-alternate-tui-views-and-thread-identity-discoverable.md),
including a bounded expanded card for the selected Thread rather than making every list row taller.

The routing rule was prompted by a concrete dogfood failure: `G1` genuinely unlocks `M25`, while
`M1` is unlocked by `G2`; the original long `G1 → M25` stroke crossed `G2 → M1` at M1's arrow and
visually invented `G1 → M1`. Navigation correctly followed the supplied edge to `M25`, exposing
that the picture—not canonical identity or graph traversal—was wrong. Focused regressions now pin
both late-gate placement and a long edge staying out of an intermediate node's incoming route.
Production-strength crossing, shared-lane, clipping, and route-identity guarantees continue in
[dense spatial routing hardening](6g86e10dztpf-make-dense-thread-graph-routes-visually-trustworthy.md)
rather than treating this first correction as proof that the home-grown router is finished.

Prototype resource use is now explicitly bounded before allocating the dense terminal canvas: more
than 512 nodes, 2,048 edges, or 750,000 layout cells produces a deterministic capacity explanation,
retains the selected-node inspector, and points to the complete wave reader/task picker instead of
rendering a misleading partial graph or risking an unresponsive TUI. Reload regressions also pin the
manual watcher dogfood result: adding a member and renaming the selected task update the open spatial
projection while preserving its canonical selection and immersive presentation.
