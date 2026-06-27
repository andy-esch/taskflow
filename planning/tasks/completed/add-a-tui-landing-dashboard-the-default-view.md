---
schema: 1
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: A TUI landing Dashboard — the TUI counterpart of tskflwctl status — as the default view; mostly a render of core.Summary, with a larger-goals widget gated on the future Projects entity.
effort: L
tier: 3
priority: medium
autonomy_level: 3
tags: [tui]
created: "2026-06-26"
updated_at: "2026-06-27"
started_at: "2026-06-27"
completed_at: "2026-06-27"
---

# Add a TUI landing Dashboard (the default view)

## Objective

Give the TUI a landing **Dashboard** — the TUI counterpart of `tskflwctl status` —
that opens as the default view instead of dropping straight into the tasks tab. An
orienting glance over project state (what's in progress, what moved recently, what
needs attention), with each widget **navigational**: select an item to jump into
its tab/detail. "`status` the command ↔ Dashboard the view," over the same core.

## Key insight — mostly NOT data-blocked

A v1 Dashboard is ~90% a NEW RENDER of data `core.Summary` already computes in one
call: status counts, the in-progress list, per-epic rollups (done/total), open
audits, misfiled count, and `revisit_due`. No new core plumbing for those. Only two
pieces need more:
- **Recent pulse** (recently completed / updated) — a small new read off
  `updated_at` / `completed_at`.
- **Larger goals / cross-cutting clusters** — needs the **Projects** entity
  (groupings that span epics; see
  `planning/research/2026-06-20-adrs-and-projects-format-design.md`). GATED — ship
  everything else before Projects lands.

## Proposed widgets

- **In progress** — the active work (`Summary.InProgress`).
- **Pulse** — recently completed / updated (last N).
- **Status counts** — per-status tallies.
- **Epic rollups** — done/total per epic (`Summary.Epics`).
- **Due for revisit** — deferred tasks whose revisit date has arrived
  (`Summary.RevisitDue` + the list). This absorbs the "banner" idea from
  [[surface-due-for-revisit-deferred-tasks-in-the-tui]].
- **Health** — misfiled + open audits + lint state.
- **Larger goals (GATED on Projects)** — cross-cutting clusters of completed +
  in-progress toward a goal.

## Open questions

> Decide before / while building.

1. **Epic placement:** keep as a task under epic 18, or promote to its own epic
   (multiple widgets + a UX shift + Projects overlap)? Leaning own-epic if it grows.
2. **Default-landing + fast path:** Dashboard as the leftmost/default tab, but power
   users need a fast path — remember the last-used view across sessions and/or a
   one-key jump to tasks — so the extra glance isn't a tax.
3. **v1 widget set:** which ship first? (likely in-progress + due-for-revisit + epic
   rollups + health; pulse next; goals when Projects lands.)
4. **Layout:** single scrollable column vs panes; how it composites with the tab strip.

## Acceptance criteria (v1 — refine per the open questions)

- [ ] A Dashboard view renders `core.Summary` (in-progress, epic rollups,
      due-for-revisit, health) READ-ONLY; the TUI reads through `core.Service` (no
      I/O in `Update`/`View`).
- [ ] Opens as the default landing, with a fast path to the tasks tab (decision #2).
- [ ] Widgets are navigational: selecting an item jumps to its tab/detail
      (in-progress task → tasks; epic → epics; due-for-revisit → the `:revisit` view).
- [ ] Deterministic tests (inject the core clock for any date-relative widget).

## Out of scope

- Mutations from the Dashboard (v1 is navigational / read-only).
- The Projects entity itself (its own design/epic) — only the goals widget depends on it.
- The per-row revisit marker + sort + `:revisit` view — those live in
  [[surface-due-for-revisit-deferred-tasks-in-the-tui]].

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- TUI counterpart of the `status` command (`core.Summary` / `SummaryHuman`)
- Absorbs the summary/banner from [[surface-due-for-revisit-deferred-tasks-in-the-tui]]
