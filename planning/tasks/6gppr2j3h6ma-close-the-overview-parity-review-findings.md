---
schema: 1
id: 6gppr2j3h6ma
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Restore zero-bucket filtering in cross-space status, keep Overview counts reachable on short terminals, fix theme preview completion, and pin the staleness wiring.
effort: half day
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, cli, dashboard]
created: "2026-10-23"
audit_sources: [planning/audits/6g9g0jf3rsja-2026-09-12-tui-overview-parity-and-miami-vice-implementation.md]
updated_at: "2026-10-23"
---
# Close the Overview-parity review findings

## Objective

The 2026-09-12 adversarial review of the Overview parity and miami-vice changesets recorded four
bounded defects that were never triaged. Fix the two user-visible regressions and close the
completion and coverage gaps, re-verifying each against current code before changing it.

## Acceptance criteria

- [ ] `status --all` compact space summaries omit zero-valued status buckets (or render `none`), pinned by a human-output regression built from a Summary that carries every status.
- [ ] On a short terminal with no in-progress task, the Overview's task heading and active/archived counts are reachable while the current selection stays visible, covered by a short-height test.
- [ ] `theme preview <name> <TAB>` offers no second theme name while `--theme` completion still lists the registry; both positional cases are tested.
- [ ] A rendered-Overview test fails if the in-progress age column reverts from `staleDateCells` to the neutral `relDateCells`.

## Out of scope

- Redesigning the dashboard layout or the miami-vice palette.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Audit [2026-09-12 TUI Overview parity and miami-vice](../audits/6g9g0jf3rsja-2026-09-12-tui-overview-parity-and-miami-vice-implementation.md), findings M1, M2, L1, L2
