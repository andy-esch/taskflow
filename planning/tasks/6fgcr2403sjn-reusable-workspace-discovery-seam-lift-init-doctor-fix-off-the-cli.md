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
updated_at: "2026-06-28"
deferred_at: "2026-06-28"
id: 6fgcr2403sjn
---
Audit 2026-06-27-consumer-data-flow-architecture H7. init/doctor/lint --fix + repo discovery bypass core.Service and live CLI-side (root.go resolve, config.* calls), so TUI/web cannot reuse them. Extract 'discover config -> build store -> build service' into a shared Resolve()->Workspace; promote Doctor()/FixFrontmatter() to core.Service; keep init as a cobra-free config function.

**Deferred 2026-06-28.** Web-readiness only (audit H7 — see its Resolution note). The present-day win (dedup `doctor`s linkback check) is small; the reusable Resolve()->Workspace seam + core Doctor()/FixFrontmatter() only pay off once a second adapter (web) exists to reuse them. Deferred pending epic 19; revisit when `tskflwctl serve` is scoped.
