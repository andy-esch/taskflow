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
---
Audit 2026-06-27-consumer-data-flow-architecture H6. No core/store method takes a context.Context; a web adapter has no cancellation/deadline/tracing path. Add ctx as the first param to the Store ports + Service methods (CLI passes Background, FS checks ctx.Err() at ReadDir/ReadFile loop boundaries). Additive now, flag-day rewrite later.

**Deferred 2026-06-28.** Web-readiness only (audit H6 — see its Resolution note). The web companion (epic 19, `tskflwctl serve`) is not yet decided, and the ctx retrofit is additive + mechanical, so deferring is cheap and lets the real serve handler shape the seam instead of guessing now. Revisit when the web adapter is actually scoped.
