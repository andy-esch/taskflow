---
schema: 1
id: 6g9fpr190sxd
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Render Summary.Counts on the Overview so the TUI dashboard shows the same active/archived shape CLI status already does.
effort: S
tier: 3
priority: medium
autonomy_level: 4
tags: [tui, ux, dashboard, parity]
created: "2026-09-12"
updated_at: "2026-09-12"
completed_at: "2026-09-12"
---
# Show the active/archived task counts on the TUI Overview

## Objective

CLI `status` opens with the shape of the whole task set — `active 4 next-up · 4
ready-to-start · 3 in-progress` over `archived 9 completed · 2 deprecated · 2 deferred`. The TUI
Overview never renders it, so the in-app dashboard leads with `in progress (3)` and shows no
next-up or ready-to-start count anywhere. Both surfaces already read the same `core.Summary`;
only one of them uses `Summary.Counts`. Close that gap.

## Acceptance criteria

- [x] The Overview renders the active and archived count lines from `Summary.Counts`, above the
  existing widgets.
- [x] Both surfaces agree on which statuses are active versus archived, and on the order within
  each line, without either re-deriving the split independently.
- [x] A bucket with no tasks is omitted rather than rendered as a zero, matching `status`.
- [x] Each count carries its status colour on the TUI while the CLI keeps its own treatment —
  shared structure, per-surface colour.
- [x] The lines are not selectable: they orient, they are not navigation targets.
- [x] Width-safe at narrow terminals like every other dashboard row.

## Design notes

`splitCounts` and `countLine` live in `internal/cli/render/status.go` — CLI-private, and the TUI
must not import another primary adapter. The established answer is already in this codebase:
`theme.Breakdown` is shared by the CLI's `countByLine` and the dashboard's `urgencyLine`, whose
comment records the pattern — same iterate/join structure, per-surface colouring (audit M10).
Follow that rather than inventing a second sharing mechanism or duplicating the split.

The active/archived split itself is domain knowledge (`domain.Status.IsActive()` already exists
and `core.Summary.Counts` is documented as "every status in display order"), so the honest home
for the split is below both adapters, not in either one.

Note this is the same class of drift as the open-audit row, which the dashboard renders as
`◆ 1 open audit(s)` where `status` shows a settled bar and percentage — `styles.segBar` already
exists and is used on the audits tab. Not in scope here, but worth filing if this lands cleanly.

## Out of scope

- The audit/graph-health parity gaps from the same review.
- Any board or column layout work.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Comes from [mine the web mockup directions for the TUI](6g95m4eyvf4k-mine-the-web-mockup-directions-for-the-tui-data-organization-and-a-more-80s.md), axis A.
