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
Audit 2026-06-27-consumer-data-flow-architecture H7. init/doctor/lint --fix + repo discovery bypass core.Service and live CLI-side (root.go resolve, config.* calls), so TUI/web cannot reuse them. Extract 'discover config -> build store -> build service' into a shared Resolve()->Workspace; promote Doctor()/FixFrontmatter() to core.Service; keep init as a cobra-free config function.