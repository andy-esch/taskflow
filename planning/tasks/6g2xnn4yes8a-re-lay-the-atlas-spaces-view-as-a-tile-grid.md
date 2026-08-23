---
schema: 1
id: 6g2xnn4yes8a
status: ready-to-start
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Replace the full-width card list with a responsive grid of mini-dashboard tiles, with per-space accents and a reserved slot for the deferred worktree marker.
effort: L
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, atlas, design, ux]
created: "2026-08-23"
updated_at: "2026-08-23"
---
# Re-lay the atlas spaces view as a tile grid

## Objective

The spaces view is a full-width vertical list. The design sketch drew a grid of bordered
mini-dashboard tiles, and that is the better instrument for the question this view answers
— *compare my projects at a glance* — because tiles put spaces side by side instead of
forcing a scroll to compare two of them.

This is the **delight** half of the atlas work, and is deliberately sequenced after the
[work view](6g2nnkffgyeg-render-the-cross-space-in-progress-rail-in-the-atlas.md) so the
expensive rendering rewrite is paid only once the useful half is already earning its keep.
Do not start it before living with the work view for a while.

See [The atlas as a dashboard of dashboards](../research/6g2qtp0022t7-the-atlas-as-a-dashboard-of-dashboards.md)
for the full reasoning, including the width arithmetic that dictates the entry-point
decision below.

## Acceptance criteria

- [ ] Spaces render as bordered tiles laid out in a responsive grid — multiple columns when
  the terminal affords it, degrading to one column when it does not, with no horizontal
  overflow at any width the list handles today.
- [ ] Each tile is a mini-dashboard: space name, a progress bar over member-task
  completion via the existing `s.miniBar`, the in-progress count, and epic / open-audit /
  finding counts. No new rendering primitives — compose `internal/progressbar` and the
  existing styles bundle.
- [ ] **Entry points: selected-only.** The tile shows the selected entry's role and path
  plus an `N entry points · h/l` affordance; `h`/`l` selection behaves exactly as it does
  now. This regresses a shipped acceptance criterion from
  [build-the-tui-atlas-and-navigate-into-spaces](6g2jhr3g20ss-build-the-tui-atlas-and-navigate-into-spaces.md)
  ("shows every entry-point directory in subdued text") — amend that criterion explicitly
  in this task's notes rather than dropping it silently. The full inventory remains in
  `space list` and `doctor`.
- [ ] An unhealthy space stays visible and still cannot be entered; its detail and remedy
  are reachable without blowing up the tile row's height.
- [ ] **A slot is reserved for the deferred worktree marker** — the tile's title row,
  right-aligned — and left empty. Retrofitting it into a full ~48-column tile later would
  force a re-layout; see the worktree task.
- [ ] Per-space accent derived deterministically from the durable planning id, over a
  curated subset of `design.Palette`, colouring the tile border. `core.SpaceEntryPoint.Accent`
  already exists and is unused; epic 29 currently lists accents as out of scope, so amend
  that too. Accents must remain legible in every registered theme and in both backgrounds.
- [ ] Grid scrolling works by tile-row, not by line; the pinned header and status row from
  the H2 fix survive, and the focused tile is never scrolled off.
- [ ] The three tests pinning current atlas rendering are reworked rather than deleted:
  header/status pinning, short-terminal focus retention, and grouped-card navigation.
- [ ] No new module dependency. `charm.land/lipgloss/v2` (`Border`, `Width`,
  `JoinHorizontal`/`JoinVertical`) plus existing internal widgets cover this; see the
  research doc's component survey for why `76creates/stickers` was rejected.

## Out of scope

- The work view and any aggregate banner.
- Worktree-aware entry selection or the marker's content — only its reserved slot.
- Making unhealthy-space remedies actionable from the atlas (`x forget` / `p repoint` in
  the sketch). That is mutation from a navigator and needs its own product decision.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Design: [The atlas as a dashboard of dashboards](../research/6g2qtp0022t7-the-atlas-as-a-dashboard-of-dashboards.md)
- Sequenced after: [Add a cross-space work view to the atlas](6g2nnkffgyeg-render-the-cross-space-in-progress-rail-in-the-atlas.md)

## Reopened, 2026-08-23 — decide this AFTER the table

The spaces view is being composed as an aligned table first
([compose-the-atlas-spaces-view-as-a-comparable-table](6g2zqyra2s6h-compose-the-atlas-spaces-view-as-a-comparable-table.md)).
Most of what this task was carrying — bars, richer cards, visual weight — lands there for a
fraction of the cost, with no grid reflow math and no 48-column squeeze, because a table
row can be as wide as the terminal.

So do not start this until the table exists. Then one of two things is true, and the answer
should be recorded here either way:

1. Tiles shrink to a pure layout change over already-rich rows — much smaller than the
   scope above, or
2. Tiles are no longer worth it, because the table delivered the delight without the
   constraint fight, and this task is closed `wontfix` with that reasoning.

The entry-point criterion above is also superseded: the table puts entry points in a pinned
band rather than choosing between hiding them and letting them blow up a row's height.
Note also that `v` now cycles a screen's views, so tile work must not reinvent view
switching.
