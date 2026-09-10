---
schema: 1
id: 6g8ezj5e51hg
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Bound oversized spatial-graph work before route layout and reuse one shared read-only layout per projection refresh across selection, navigation, and rendering.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [threads, tui, performance, hardening]
created: "2026-09-09"
depends_on: [6g86e10dztpf]
updated_at: "2026-09-10"
started_at: "2026-09-09"
completed_at: "2026-09-10"
---
## Objective

Make the spatial Thread view pay for layout once per coherent projection refresh and reject oversized inputs before route materialization, without caching stale graph evidence across watcher reloads.

## Acceptance criteria

- [x] Node and edge preflight limits run before ranking, gate fixed-point placement, lane assignment, and route materialization; the fallback remains explanatory and preserves access to the complete wave reader.
- [x] One shared read-only spatial layout is reused across selection normalization, directional navigation, and rendering for a coherent Thread projection instead of being rebuilt multiple times per frame or keypress.
- [x] Cache invalidation follows projection replacement and stable task identity, including add, rename, deletion, degraded refresh, and recovery paths.
- [x] Benchmarks or bounded regression evidence cover near-limit nodes, edges, route length, canvas cells, and repeated render/navigation work.
- [x] The optimization remains presentation-only and does not create a mutable graph model or redefine core waves, readiness, health, or dependencies.

## Context

The dense-route implementation reviews found that the canvas allocation guard is bounded but layout construction still happens before that guard and can be repeated by selection, navigation, and rendering. The route hardening task reduced per-cell ownership to an integer route index and lowered the canvas ceiling; this follow-up owns the remaining lifecycle, preflight, and reuse work.

## Implementation checkpoint (2026-09-10)

Spatial preparation now rejects excessive node, edge, and malformed raw wave-record inputs before graph ranking or geometry, then checks planned canvas dimensions before materializing routes. A lazy cache owned by each coherent Thread detail read supplies one shared read-only prepared layout to selection normalization, directional navigation, and rendering; same-Thread refreshes preserve stable task identity while replacing the cache on add, rename, deletion, coherent degradation, and recovery. A failed transient refresh continues presenting the last coherent layout.

Adversarial review tightened the avoided-work evidence. An injected planner proves rejected inputs never enter layout planning, while an injected sentinel prepared result makes render or navigation cache bypasses observable. Capacity fallback selection is computed once with bounded input work and then reused without rescanning the projection; selection normalization has one owner per render or movement. Over-canvas fallback retains a standalone placement layout rather than the rejected route seeds and lane maps, and rendering/navigation are checked not to mutate their shared result.

Focused regressions exercise limits and their first rejected values for nodes, edges, raw wave records, and canvas cells, plus wave-only nodes, long routes, lazy preparation, stable selection, refresh invalidation, and runtime cache consumption. Benchmarks now separate near-canvas preparation, near-canvas cached rendering, and active cached navigation at the node limit. They also exposed a distinct full-canvas-per-frame allocation problem, tracked by `render-only-the-visible-spatial-thread-viewport` rather than folded into layout-cache ownership. All behavior remains inside the TUI presentation adapter over `core.ThreadGraphProjection`.

## Related

- [Dense Thread route hardening](6g86e10dztpf-make-dense-thread-graph-routes-visually-trustworthy.md)
- [Claude implementation audit](../audits/6g8e0pyczpns-2026-09-09-dense-thread-graph-routes-implementation-claude.md)
- [Antigravity implementation audit](../audits/6g8e0pyn885r-2026-09-09-dense-thread-graph-routes-implementation-antigravity.md)
- [Cache/preflight implementation audit — Claude](../audits/6g8hpbg6sttx-2026-09-09-cache-and-preflight-spatial-thread-layout-work-implementation-claude.md)
- [Cache/preflight implementation audit — Antigravity](../audits/6g8hpc8wreyz-2026-09-09-cache-and-preflight-spatial-thread-layout-work-implementation-antigravity.md)
- [Visible-viewport rendering follow-up](6g8k47wg8xks-render-only-the-visible-spatial-thread-viewport.md)
- [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
