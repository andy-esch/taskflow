---
schema: 1
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: ""
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, core]
created: "2026-06-27"
---
Audit 2026-06-27-consumer-data-flow-architecture M2+M3+M9 (+L1). Push into core: epic recency-ordering (so CLI status and the TUI dashboard agree), EpicSummary.Percent() reuse on show/detail (ShowEpic returns a rollup), the settled/ready-to-close count (deduped from render.go + dashboard.go), and FindingsRollup as a composed Service view-model. Adapters should read aggregates, not re-derive off raw Summary lists.