---
schema: 1
id: 6fq9zy102rea
status: completed
epic: 20-cli-ux-and-ergonomics
description: audit show/list % counts only fixed+landed while the bar bands 4 states, so a fully-superseded (triaged) audit reads 0%.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [audit, ux]
created: "2026-07-18"
updated_at: "2026-07-21"
started_at: "2026-07-21"
completed_at: "2026-07-21"
---
# Audit progress number contradicts its own bar

## Objective

`audit show` / `audit list` show a headline `N%` that counts only
fixed + landed (`FindingTally.Done`, internal/domain/finding.go:146), but the
segmented bar bands by four dispositions (open / in-progress / done / dropped).
So the number and the bar disagree in one widget: an audit whose findings are all
`superseded` (fully triaged into tasks, correctly closeable) shows `0%` —
indistinguishable *at the number* from an untouched audit. Confirmed: two
superseded findings render `▒▒▒▒▒▒▒▒▒▒ 0% 0/2`.

## Acceptance criteria

- [ ] The headline number and the bar tell one consistent story
- [ ] A fully-triaged audit (0 open, all superseded) is distinguishable from an untouched one
- [ ] The `audit list` `open` column reconciles with that same notion of progress

## Notes

- Options from the report: label it (`0% resolved`), show a stacked count
  (`2 superseded · 2 in-progress · 0 fixed`), or make the number track the bar's gradient.
- Loci: `TallyFindings`/`FindingTally` (internal/domain/finding.go:143), the
  headline `%` (`Percent`/`SegmentBar` in internal/cli/render/style.go).
- Source: https://github.com/andy-esch/taskflow/issues/105 (P1, High)

## Resolution (2026-07-21)

Kept the headline % meaning **fixed share** (never over-claims) and surfaced the
already-computed settled state as a green `✔ ready to close` marker — the
distinguisher between a triaged audit (full bar, 0% fixed) and an untouched one.

- `domain.Audit.ReadyToClose()` centralizes the open+Settled call-to-action (deduped from the JSON DTO); JSON `ready_to_close` unchanged.
- `theme.AuditPercentLabel`/`…Padded` label the audit number as `N% fixed` (epics keep the bare %).
- One `auditStateNote` helper drives `audit show`/`list`/dashboard; TUI list row + detail mirror it, so no surface drifts.
- `audit show`: `(N open)` while work remains, else `✔ ready to close`; list rows show only the ready marker to stay scannable.

All three acceptance criteria met: number+bar tell one story; triaged vs untouched distinguishable; `open` column reconciles (0 open ⟺ ready to close). Tests + lint green.
