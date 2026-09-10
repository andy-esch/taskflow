---
schema: 1
id: 6g8k47wg8xks
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: Allocate and compose spatial Thread canvas cells from the visible viewport rather than the full bounded graph on every frame.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 3
tags: [threads, tui, performance, hardening]
created: "2026-09-10"
depends_on: [6g8ezj5e51hg]
updated_at: "2026-09-10"
---
# Render only the visible spatial Thread viewport

Make spatial Thread rendering allocate and compose terminal cells in proportion to the visible viewport rather than the full bounded graph canvas.

## Context

The cache/preflight implementation bounds and reuses projection-to-layout work, but adversarial profiling exposed the next bottleneck: `renderThreadSpatialCanvasWindow` still allocates the complete layout-sized cell matrix on every frame and clips it afterward. A supported near-canvas fixture retained one cached layout yet allocated roughly 27 MB and 20,000 objects per render. This is outside layout-cache ownership, but it makes the nominal 500,000-cell guard an expensive steady-state rendering allowance rather than only a safety ceiling.

The fix belongs in the TUI presentation adapter. It must not change `ThreadGraphProjection`, graph semantics, route ownership, navigation, or repository state.

## Acceptance criteria

- [ ] `renderThreadSpatialCanvasWindow` allocates cell storage from the requested visible width and height, not `layout.width × layout.height`, while retaining an explicit transform between layout and viewport coordinates.
- [ ] Nodes, focused expansion, routes, crossings/bundles, endpoint arrows and counts, offscreen endpoint annotations, and inspector behavior remain faithful when partially or wholly clipped on every boundary.
- [ ] Panning and resizing cannot index outside the viewport canvas or silently omit an incident route that the existing visibility contract requires.
- [ ] Empty, narrow, degraded, over-capacity, and partially projected Threads retain their current explanatory fallbacks.
- [ ] Bounded regression evidence covers viewports at graph origin and interior/extreme pans, partially visible nodes, long routes that enter/leave/re-enter the window, and several incident routes sharing a boundary.
- [ ] A benchmark over the existing near-canvas fixture demonstrates that per-frame cell memory scales with viewport dimensions; include a comparison at two viewport sizes and record what layout-wide work remains.
- [ ] The implementation remains deterministic and presentation-only; no global render cache or mutable semantic graph is introduced.

## Related evidence

- `cache-and-preflight-spatial-thread-layout-work` implementation audits measured the cached near-canvas render at roughly 27 MB and 20,000 allocations per frame while layout preparation itself was reused correctly.
- The dense-route trustworthiness work defines the connector, collision, clipping, and offscreen-boundary semantics this optimization must preserve.

## Out of scope

- Replacing the spatial layout algorithm or adopting a general graph library.
- Responsive node sizing, one-hop focus mode, or alternate-view discoverability.
- Persisting viewport position or rendered terminal cells across coherent projection refreshes.
