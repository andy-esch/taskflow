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
updated_at: "2026-06-28"
---
Audit 2026-06-27-consumer-data-flow-architecture M2+M3+M9 (+L1). Push into core: epic recency-ordering (so CLI status and the TUI dashboard agree), EpicSummary.Percent() reuse on show/detail (ShowEpic returns a rollup), the settled/ready-to-close count (deduped from render.go + dashboard.go), and FindingsRollup as a composed Service view-model. Adapters should read aggregates, not re-derive off raw Summary lists.

**Progress 2026-06-28 (M2 done; L1 deferred).** M2 landed: epic recency-ordering moved into the aggregate — `core.Summary.Epics` is returned most-recently-updated first (`epicsByRecent` in service_epic.go, applied once in Summary()), so CLI `status` and the TUI dashboard share ONE order; the TUI`s local sort is deleted. Pinned by TestService_Summary_EpicsByRecent (CLI golden unchanged → non-breaking). L1 (FindingsRollup as a composed Service view-model) is DEFERRED pending the web effort (epic 19) — see the audit. Remaining here: M3 (ShowEpic returns a rollup so show/detail reuse EpicSummary.Percent()) and M9 (push the settled/ready-to-close count into Summary).
