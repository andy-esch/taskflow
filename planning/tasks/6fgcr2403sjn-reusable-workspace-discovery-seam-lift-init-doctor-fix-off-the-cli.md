---
schema: 1
status: deferred
epic: 21-code-quality-architecture-hardening
description: ""
effort: L
tier: 3
priority: high
autonomy_level: 3
tags: [architecture, web]
created: "2026-06-27"
updated_at: "2026-08-15"
deferred_at: "2026-06-28"
id: 6fgcr2403sjn
---
Audit 2026-06-27-consumer-data-flow-architecture H7. init/doctor/lint --fix + repo discovery bypass core.Service and live CLI-side (root.go resolve, config.* calls), so TUI/web cannot reuse them. Extract 'discover config -> build store -> build service' into a shared Resolve()->Workspace; promote Doctor()/FixFrontmatter() to core.Service; keep init as a cobra-free config function.

**Deferred 2026-06-28.** Web-readiness only (audit H7 — see its Resolution note). The present-day win (dedup `doctor`s linkback check) is small; the reusable Resolve()->Workspace seam + core Doctor()/FixFrontmatter() only pay off once a second adapter (web) exists to reuse them. Deferred pending epic 19; revisit when `tskflwctl serve` is scoped.

### 2026-08-15 — a possible second trigger (still deferred)

This task's deferral rests on "the seam only pays off once a second adapter (web)
exists to reuse them." A **multi-space TUI** — one process holding N planning repos —
would be a second consumer of exactly this seam, and would want `Resolve() → Workspace`
per space rather than one CLI-side `resolve()`. See epic
[29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
and its sketch
[2026-08-15-multi-space-home-registry-and-the-atlas](../research/2026-08-15-multi-space-home-registry-and-the-atlas.md).

**Staying deferred** — epic 29 is an undecided direction, not a commitment, and its
sequencing deliberately puts the CLI-only slice first so the seam isn't needed until
the TUI board is actually chosen. Recorded here only so the trigger isn't lost.
