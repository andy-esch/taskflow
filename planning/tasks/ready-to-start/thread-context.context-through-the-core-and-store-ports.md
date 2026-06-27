---
schema: 1
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: ""
effort: L
tier: 3
priority: high
autonomy_level: 3
tags: [architecture, web]
created: "2026-06-27"
---
Audit 2026-06-27-consumer-data-flow-architecture H6. No core/store method takes a context.Context; a web adapter has no cancellation/deadline/tracing path. Add ctx as the first param to the Store ports + Service methods (CLI passes Background, FS checks ctx.Err() at ReadDir/ReadFile loop boundaries). Additive now, flag-day rewrite later.