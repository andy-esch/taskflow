---
schema: 1
id: 6g8k47wg8xks
status: in-progress
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
started_at: "2026-09-10"
---
# Render only the visible spatial Thread viewport

Make spatial Thread rendering allocate and compose terminal cells in proportion to the visible viewport rather than the full bounded graph canvas.

## Context

The cache/preflight implementation bounds and reuses projection-to-layout work, but adversarial profiling exposed the next bottleneck: `renderThreadSpatialCanvasWindow` still allocates the complete layout-sized cell matrix on every frame and clips it afterward. A supported near-canvas fixture retained one cached layout yet allocated roughly 27 MB and 20,000 objects per render. This is outside layout-cache ownership, but it makes the nominal 500,000-cell guard an expensive steady-state rendering allowance rather than only a safety ceiling.

The fix belongs in the TUI presentation adapter. It must not change `ThreadGraphProjection`, graph semantics, route ownership, navigation, or repository state.

## Acceptance criteria

- [x] `renderThreadSpatialCanvasWindow` allocates cell storage from the requested visible width and height, not `layout.width × layout.height`, while retaining an explicit transform between layout and viewport coordinates.
- [x] Nodes, focused expansion, routes, crossings/bundles, endpoint arrows and counts, offscreen endpoint annotations, and inspector behavior remain faithful when partially or wholly clipped on every boundary.
- [x] Panning and resizing cannot index outside the viewport canvas or silently omit an incident route that the existing visibility contract requires.
- [x] Empty, narrow, degraded, over-capacity, and partially projected Threads retain their current explanatory fallbacks.
- [x] Bounded regression evidence covers viewports at graph origin and interior/extreme pans, partially visible nodes, long routes that enter/leave/re-enter the window, and several incident routes sharing a boundary.
- [x] A benchmark over the existing near-canvas fixture demonstrates that per-frame cell memory scales with viewport dimensions; include a comparison at two viewport sizes and record what layout-wide work remains.
- [x] The implementation remains deterministic and presentation-only; no global render cache or mutable semantic graph is introduced.

## Implementation checkpoint (2026-09-10)

The spatial renderer now allocates a canvas at the intersection of the requested terminal viewport and the preflighted layout bounds. Canvas primitives continue receiving global layout coordinates but translate through an explicit viewport origin and clip text, route segments, arrows, counts, nodes, and annotations before touching storage. Horizontal and vertical strokes clip their iteration ranges, while boundary extent calculation intersects route segments analytically instead of walking every cell of a long path. Oversized synthetic terminal dimensions cannot allocate more cells than the guarded layout.

Regression coverage compares viewport composition cell-for-cell with the former full-canvas-then-crop behavior at the graph origin, around an interior selection, through a partially clipped selected node, and at bottom/side extremes. Additional cases cover wide terminal glyphs, traversal direction, vertical boundary labels, and a selected route that enters, leaves, and re-enters the viewport. The existing narrow, degraded, capacity, dense-route, crossing, fan-count, and deterministic suites remain green.

On the near-500,000-cell fixture, cached 120×30 rendering fell from roughly 27 MB and 20,000 allocations per frame to about 220 KB and 2,700 allocations. Direct canvas benchmarks construct exactly 1,440 cells for 80×18 (69,120 bytes of cell storage; about 123 KB total per operation) and 5,760 cells for 160×36 (276,480 bytes of cell storage; about 352 KB total per operation). Remaining layout-wide work is bounded iteration over cached columns, nodes, and route metadata to select visible evidence; terminal-cell allocation and segment traversal are viewport-bounded.

Adversarial review caught two unintended presentation changes and three mutation-surviving coverage gaps. Routing conflicts are renderer-invariant failures, so the cached layout now indexes them once from route-turn candidates using the same connector grammar and the header stays stable across pans without recreating a full canvas. Route-count placement distinguishes clipping from genuine congestion, preventing an offscreen preferred count from appearing spuriously on a visible node border. Independent regressions now pin nonzero vertical origins, horizontal and vertical segment-work bounds, selected incident routes with both endpoints offscreen, wide-glyph accent clipping, and layout-wide conflict indexing; each targeted guard was mutation-checked. Both implementation audits are closed with every finding fixed.

Validation passed with the race-enabled repository suite, focused TUI suite, `golangci-lint`, `go vet`, tidy/docs checks, planning lint, and diff checks.

## Related evidence

- `cache-and-preflight-spatial-thread-layout-work` implementation audits measured the cached near-canvas render at roughly 27 MB and 20,000 allocations per frame while layout preparation itself was reused correctly.
- The dense-route trustworthiness work defines the connector, collision, clipping, and offscreen-boundary semantics this optimization must preserve.
- [Claude implementation review](../audits/6g8p95exeeje-2026-09-10-visible-spatial-thread-viewport-implementation-claude.md)
- [Antigravity implementation review](../audits/6g8p95f62d04-2026-09-10-visible-spatial-thread-viewport-implementation-antigravity.md)

## Out of scope

- Replacing the spatial layout algorithm or adopting a general graph library.
- Responsive node sizing, one-hop focus mode, or alternate-view discoverability.
- Persisting viewport position or rendered terminal cells across coherent projection refreshes.
