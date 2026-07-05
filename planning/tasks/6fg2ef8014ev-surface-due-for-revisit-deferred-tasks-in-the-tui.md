---
schema: 1
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: 'Bring the CLI due-for-revisit surfacing to the TUI: an on-open banner, a non-emoji row marker, and due-deferred tasks sorted to the top. Several design forks left open (see task body).'
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [tui]
created: "2026-06-26"
updated_at: "2026-06-26"
started_at: "2026-06-26"
completed_at: "2026-06-26"
id: 6fg2ef8014ev
---

# Surface "due for revisit" deferred tasks in the TUI

## Objective

The CLI surfaces deferred tasks whose `revisit_at` has arrived (the `status` nudge
and `task list --revisit-due`), but the TUI — the human browse surface — has ZERO
awareness: no marker, no view, no use of `domain.IsRevisitDue`. A human in the
`:deferred` view can't tell which tasks have come due. Surface it in the TUI so due
tasks are easy to find and act on.

Constraint: the default TUI task view is the ACTIVE working set (in-progress →
next-up → ready-to-start); deferred tasks only appear via `:deferred` / `:all` /
the `s`/`S` cycle. That drives the one remaining open decision below.

Scope note: the on-open SUMMARY/banner idea graduated to its own concept — a TUI
landing Dashboard (see [add-a-tui-landing-dashboard-the-default-view](6fg2ef802k6s-add-a-tui-landing-dashboard-the-default-view.md)). THIS task
is just the per-row surfacing inside the views where deferred tasks live.

## Resolved decisions

- **Indicator glyph:** `◷` — a monochrome clock-face glyph (NO emoji), rendered in
  its OWN leading marker column (so existing columns don't shift), colored with the
  warn style.
- **No standalone banner here.** Surface via the marker + sort-to-top + a `:revisit`
  view (the TUI mirror of `task list --revisit-due`) + a jump key. The on-open
  summary/"pulse" belongs to the Dashboard task, not this one.
- **De-emoji the CLI `status` nudge:** YES. Replace the ⏰ (and sibling emoji such as
  ⚠ in that block) with non-emoji markers, per the standing no-emoji preference;
  regenerate any affected goldens.

## OPEN DECISION — resolve before / while building

1. **[BIG FORK] Where do the marker + sort-to-top apply?**
   - **(A)** Keep the default active view pure; a dedicated `:revisit` view (or the
     `:deferred` view) is where due tasks are marked and sorted to the top.
     Preserves "view = status". — **recommended.**
   - (B) Also pin due-deferred tasks to the top of the DEFAULT active list (zero
     keystrokes), but this injects a `deferred` row into the working set and muddies
     "what am I working on"; if chosen, style those rows as snoozed-but-due.

## Acceptance criteria

- [ ] A `:revisit` view lists deferred tasks whose revisit date has arrived
      (`domain.IsRevisitDue` + the injected core clock) — the TUI mirror of
      `task list --revisit-due`.
- [ ] A non-emoji `◷` marker flags due-for-revisit rows, in a leading column so the
      existing columns keep their positions.
- [ ] Due-for-revisit tasks sort to the top of the relevant view (per the open fork).
- [ ] The CLI `status` nudge is de-emojified (⏰ → non-emoji marker); goldens regenerated.
- [ ] All read-side: reuses `domain.IsRevisitDue` + the core clock + `Summary`; no new
      core mutations; the TUI reads through `core.Service` (no I/O in `Update`/`View`).
- [ ] Deterministic tests (inject the clock): marker on due rows, sort order, the
      `:revisit` view contents, and exclusion of not-due / no-date / non-deferred-with-stale-date.

## Out of scope

- The on-open summary/banner and any "pulse"/overview — that is the Dashboard task
  [add-a-tui-landing-dashboard-the-default-view](6fg2ef802k6s-add-a-tui-landing-dashboard-the-default-view.md).
- Auto-resuming due tasks — the snooze stays surface-only; resume is manual
  (`task next` / `task ready`).
- Changing how `revisit_at` is set or cleared (already shipped).

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- TUI mirror of [add-a-revisit-due-filter-to-task-list-for-deferred-task-triage](6fg2ef803rf8-add-a-revisit-due-filter-to-task-list-for-deferred-task-triage.md)
- Builds on [set-a-revisit-date-when-deferring-snooze-and-surface-what-is-due](6fg2ef801c1m-set-a-revisit-date-when-deferring-snooze-and-surface-what-is-due.md)
- Summary/banner concept moved to [add-a-tui-landing-dashboard-the-default-view](6fg2ef802k6s-add-a-tui-landing-dashboard-the-default-view.md)
