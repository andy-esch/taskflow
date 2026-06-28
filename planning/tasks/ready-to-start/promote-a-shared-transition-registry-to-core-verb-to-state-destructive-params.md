---
schema: 1
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: ""
effort: L
tier: 3
priority: high
autonomy_level: 3
tags: [architecture, core]
created: "2026-06-27"
updated_at: "2026-06-28"
---
Audit 2026-06-27-consumer-data-flow-architecture H3 (+M4,+M5). One verb->destination table in core/domain that CLI/TUI/web all consume (the missing peer of AllStatuses/AllAuditBuckets). Carry the destructive flag and an optional param spec so defer's revisit-date stops being special-cased in three layers and the destructive-confirm signal is shared. Builds on completed make-tui-lifecycle-action-machinery-registry-driven.

**Update 2026-06-28:** M5's clock-consistency half landed — the revisit-date relative-offset parse now uses svc.Now() (the injected clock) not time.Now(), at all 3 sites (cli/task.go, tui/edit.go, tui/model.go). Remaining in this task: the H3 shared transition registry + M4 defer-param + M5 destructive-confirm consolidation.
