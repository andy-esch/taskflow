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
---
Audit 2026-06-27-consumer-data-flow-architecture H3 (+M4,+M5). One verb->destination table in core/domain that CLI/TUI/web all consume (the missing peer of AllStatuses/AllAuditBuckets). Carry the destructive flag and an optional param spec so defer's revisit-date stops being special-cased in three layers and the destructive-confirm signal is shared. Builds on completed make-tui-lifecycle-action-machinery-registry-driven.