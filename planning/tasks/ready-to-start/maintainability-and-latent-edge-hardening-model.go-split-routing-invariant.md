---
schema: 1
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: ""
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, tui]
created: "2026-06-27"
---
Audit 2026-06-27-consumer-data-flow-architecture M6+M8+L3+L4+L5. Split the 1537-line model.go by concern (view.go, command_dispatch.go); make the Update fall-through routing invariant executable instead of comment-only; bounds-guard actionMenu/editMenu selection; guard setMapNode comment-carry; fix the scrollToCurrent >0 boundary. Follows completed split-render.go-and-service.go and harden-tui-dispatch.