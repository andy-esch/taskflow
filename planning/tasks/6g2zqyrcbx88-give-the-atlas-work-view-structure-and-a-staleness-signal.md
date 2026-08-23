---
schema: 1
id: 6g2zqyrcbx88
status: ready-to-start
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Surface how long each task has been in progress, add epic and priority columns, and give the work view its own orders including grouping by space.
effort: M
tier: 3
priority: high
autonomy_level: 3
tags: [tui, atlas, ux]
created: "2026-08-23"
updated_at: "2026-08-23"
---
# Give the atlas work view structure and a staleness signal

## Objective

The work view is a flat list that repeats the space label on every row and shows
last-touched dates. It ignores `Epic`, `Priority`, `Effort`, and — most importantly —
`StartedAt`.

**The largest win is `started`, not `updated`.** For in-progress work the useful question
is how long something has been underway; "started 5 weeks ago" is what says a task is
stuck, and `StartedAt` is already stamped on every task that entered in-progress. Nothing
in the tool surfaces it today.

## Acceptance criteria

- [ ] Each row shows how long the task has been in progress, from `StartedAt`, falling
  back to the existing date when a task carries none (legacy rows).
- [ ] That age is coloured by staleness through a new `theme.Staleness`, built in the same
  shape as the existing `theme.Percent` so the palette owns the thresholds, not the view.
- [ ] Rows carry epic and priority columns, priority via the existing `s.priorityText`,
  all aligned across the visible set.
- [ ] `o`/`O` work in this view, cycling `started` (default) / `space` / `priority`, with
  the active order named in the header exactly as the spaces view names its own. The
  footer stops omitting them here.
- [ ] The `space` order groups rows under a per-space heading and drops the now-redundant
  per-row space label; the other orders stay flat and keep it.
- [ ] The cursor stays on the same task across an order change and across a refresh.
- [ ] Description remains the last column and is the first thing sacrificed on a narrow
  terminal.

## Out of scope

- An aggregate banner. Still deliberately deferred — living with the view is what answers
  whether one earns its rows.
- A third "attention" view.
- Filtering (`/`) — tracked separately.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Design: [The atlas as a dashboard of dashboards](../research/6g2qtp0022t7-the-atlas-as-a-dashboard-of-dashboards.md)
- Builds on: [Add a cross-space work view to the atlas](6g2nnkffgyeg-render-the-cross-space-in-progress-rail-in-the-atlas.md)
