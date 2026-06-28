---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: ""
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, core]
created: "2026-06-27"
updated_at: "2026-06-28"
completed_at: "2026-06-28"
---
Audit 2026-06-27-consumer-data-flow-architecture M2+M3+M9 (+L1). Push into core: epic recency-ordering (so CLI status and the TUI dashboard agree), EpicSummary.Percent() reuse on show/detail (ShowEpic returns a rollup), the settled/ready-to-close count (deduped from render.go + dashboard.go), and FindingsRollup as a composed Service view-model. Adapters should read aggregates, not re-derive off raw Summary lists.

**Progress 2026-06-28 (M2 done; L1 deferred).** M2 landed: epic recency-ordering moved into the aggregate — `core.Summary.Epics` is returned most-recently-updated first (`epicsByRecent` in service_epic.go, applied once in Summary()), so CLI `status` and the TUI dashboard share ONE order; the TUI`s local sort is deleted. Pinned by TestService_Summary_EpicsByRecent (CLI golden unchanged → non-breaking). L1 (FindingsRollup as a composed Service view-model) is DEFERRED pending the web effort (epic 19) — see the audit. Remaining here: M3 (ShowEpic returns a rollup so show/detail reuse EpicSummary.Percent()) and M9 (push the settled/ready-to-close count into Summary).

**Completed 2026-06-28 (M3 + M9 landed; M2 earlier today; L1 deferred).** M3: ShowEpic now returns an EpicSummary built by a shared rollupEpic(epic, tasks) helper (used by both rollupEpics and ShowEpic), so the show/detail paths consume es.Percent()/Done/Total instead of re-deriving the rule; EpicShowHuman + the TUI renderEpicMeta updated. M9: the settled/ready-to-close count is now the Summary.ReadyToClose aggregate, computed once from the audit sweep and read by both dashboards; the verbatim settledCount/countSettled helpers are deleted. Pinned by TestService_ShowEpic + TestService_Summary_ReadyToClose; goldens unchanged. L1 (FindingsRollup composed view-model) stays DEFERRED pending the web effort (epic 19) — tracked in the audit, out of this task`s present-day scope. All actionable items done → complete.
