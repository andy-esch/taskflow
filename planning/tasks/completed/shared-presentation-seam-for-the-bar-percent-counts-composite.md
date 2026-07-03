---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: ""
effort: M
tier: 3
priority: high
autonomy_level: 3
tags: [architecture, render]
created: "2026-06-27"
updated_at: "2026-06-28"
completed_at: "2026-06-28"
id: 6fgcr24010wx
---
Audit 2026-06-27-consumer-data-flow-architecture H5 (+M1,+M10). One formatter for the 'bar + N% + done/total' composite (and the CountBy breakdown lines) so CLI/TUI hand-assemble it once, not 7-10x with drifting %d%% vs %3d%% and bar widths 8/10/12. Also factor the dashboard setSummary column-building toward the shared table/alignment helper. Extends the bar-dedup from completed misc-v2-review-follow-ups.

**Update 2026-06-28:** H5 landed — the progress-composite number formats (percent %d%%/%3d%%, done/total) are unified in theme.PercentLabel/PercentLabelPadded/Counts (the bar + percent color were already shared via progressbar.Render/theme.Percent); non-breaking, goldens unchanged. Remaining: M1 (the dashboard setSummary hand-rolled column structure — a separate TUI refactor toward writeTable-style alignment) and M10 (the CountBy breakdown line — low value: its only shared part is a trivial map-join while the per-surface styling/cap genuinely diverges).

**Update 2026-06-28:** M10 landed — the CountBy breakdown line`s iterate/format/join/cap STRUCTURE is now a single generic theme.Breakdown[T](items, sep, max, seg, more); countByLine (CLI) + urgencyLine/componentLine (TUI) supply only their per-segment styling, separator, and optional "+N more" cap, so the legitimately-presentational divergence stays per-surface while the loop is shared. The generic signature keeps theme free of a core import. Output unchanged (goldens green), cap behavior unit-tested (theme.TestBreakdown). Remaining in this task: M1 (the dashboard setSummary hand-rolled column structure → a shared writeTable-style alignment helper).

**Completed 2026-06-28 (M1 done; H5/M10 earlier).** The dashboard`s hand-rolled column logic is now two shared generic helpers in internal/tui/column.go — relDateCells (aligned dimmed relative-date column) and countsWidth (done/total column width). The in-progress + epics widgets (dashboard.go) and the epic/audit list loaders (commands.go) all use them instead of re-rolling dateW/countsW + %-*s, so the alignment is described once. Byte-identical render (dashboard tests unchanged), unit-tested (TestRelDateCells/TestCountsWidth). A full writeTable-style cell framework was deliberately not built — the rows are heterogeneous, so factoring the two duplicated measured columns is the right scope. All of this task`s scope is now done.
