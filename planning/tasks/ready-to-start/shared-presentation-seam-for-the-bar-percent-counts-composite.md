---
schema: 1
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: ""
effort: M
tier: 3
priority: high
autonomy_level: 3
tags: [architecture, render]
created: "2026-06-27"
---
Audit 2026-06-27-consumer-data-flow-architecture H5 (+M1,+M10). One formatter for the 'bar + N% + done/total' composite (and the CountBy breakdown lines) so CLI/TUI hand-assemble it once, not 7-10x with drifting %d%% vs %3d%% and bar widths 8/10/12. Also factor the dashboard setSummary column-building toward the shared table/alignment helper. Extends the bar-dedup from completed misc-v2-review-follow-ups.