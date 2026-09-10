---
schema: 1
id: 6g8ezj5e51hg
status: ready-to-start
epic: 30-threads-and-task-dependency-graphs
description: Bound oversized spatial-graph work before route layout and reuse one immutable layout per projection refresh across selection, navigation, and rendering.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [threads, tui, performance, hardening]
created: "2026-09-09"
depends_on: [6g86e10dztpf]
updated_at: "2026-09-09"
---
## Objective

Make the spatial Thread view pay for layout once per coherent projection refresh and reject oversized inputs before route materialization, without caching stale graph evidence across watcher reloads.

## Acceptance criteria

- [ ] Node and edge preflight limits run before ranking, gate fixed-point placement, lane assignment, and route materialization; the fallback remains explanatory and preserves access to the complete wave reader.
- [ ] One immutable spatial layout is reused across selection normalization, directional navigation, and rendering for a coherent Thread projection instead of being rebuilt multiple times per frame or keypress.
- [ ] Cache invalidation follows projection replacement and stable task identity, including add, rename, deletion, degraded refresh, and recovery paths.
- [ ] Benchmarks or bounded regression evidence cover near-limit nodes, edges, route length, canvas cells, and repeated render/navigation work.
- [ ] The optimization remains presentation-only and does not create a mutable graph model or redefine core waves, readiness, health, or dependencies.

## Context

The dense-route implementation reviews found that the canvas allocation guard is bounded but layout construction still happens before that guard and can be repeated by selection, navigation, and rendering. The route hardening task reduced per-cell ownership to an integer route index and lowered the canvas ceiling; this follow-up owns the remaining lifecycle, preflight, and reuse work.

## Related

- [Dense Thread route hardening](6g86e10dztpf-make-dense-thread-graph-routes-visually-trustworthy.md)
- [Claude implementation audit](../audits/6g8e0pyczpns-2026-09-09-dense-thread-graph-routes-implementation-claude.md)
- [Antigravity implementation audit](../audits/6g8e0pyn885r-2026-09-09-dense-thread-graph-routes-implementation-antigravity.md)
- [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
