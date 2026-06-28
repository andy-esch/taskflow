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
updated_at: "2026-06-28"
---
Audit 2026-06-27-consumer-data-flow-architecture H5 (+M1,+M10). One formatter for the 'bar + N% + done/total' composite (and the CountBy breakdown lines) so CLI/TUI hand-assemble it once, not 7-10x with drifting %d%% vs %3d%% and bar widths 8/10/12. Also factor the dashboard setSummary column-building toward the shared table/alignment helper. Extends the bar-dedup from completed misc-v2-review-follow-ups.

**Update 2026-06-28:** H5 landed — the progress-composite number formats (percent %d%%/%3d%%, done/total) are unified in theme.PercentLabel/PercentLabelPadded/Counts (the bar + percent color were already shared via progressbar.Render/theme.Percent); non-breaking, goldens unchanged. Remaining: M1 (the dashboard setSummary hand-rolled column structure — a separate TUI refactor toward writeTable-style alignment) and M10 (the CountBy breakdown line — low value: its only shared part is a trivial map-join while the per-surface styling/cap genuinely diverges).
