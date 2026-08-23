---
schema: 1
id: 6g2zqyra2s6h
status: completed
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Replace the flat card text with an aligned, bar-bearing table plus a folded attention marker and a pinned entry-point band.
effort: M
tier: 3
priority: high
autonomy_level: 3
tags: [tui, atlas, ux]
created: "2026-08-23"
updated_at: "2026-08-23"
started_at: "2026-08-23"
completed_at: "2026-08-23"
---
# Compose the atlas spaces view as a comparable table

## Objective

The spaces view renders each space as a flat run of prose — `2 in progress · 4 epic(s) ·
1 open audit(s) · 3 finding(s)` — with no bar, no alignment, and no way to compare two
spaces without reading both. Meanwhile the dashboard, two keystrokes away, composes
`s.miniBar`, `relDateCells`, `countsWidth`/`rollupCounts`, and `theme.Breakdown` into an
aligned epic row. The atlas hand-rolls text instead of using any of it.

This is composition, not invention: nearly every part already exists, is theme-aware, and
is tested. See the addendum in
[The atlas as a dashboard of dashboards](../research/6g2qtp0022t7-the-atlas-as-a-dashboard-of-dashboards.md).

Target shape:

```
  ● bike-workshop   ████░░░░░░  43%   6/14   ▸2   4 epics · 1 audit   ⚠3   1mo ago
› ● kitchen         ████░░░░░░  41%   9/22   ▸3   3 epics · 1 audit   ⚠2   yesterday

  ─ kitchen · 1 entry point ──────────────────────────────
  › ● kitchen   direct   /private/tmp/tskflw-atlas/kitchen
```

## Acceptance criteria

- [x] One aligned row per space, columns measured across the visible set so the bar,
  percentage, counts, and date columns line up — reusing `relDateCells`, `countsWidth`,
  `rollupCounts`, and `s.miniBar` rather than new measuring code.
- [x] Aggregate progress per space is member-task completion summed across its epics,
  rendered as a bar plus a `theme.Percent`-coloured percentage and a `done/total` rollup.
- [x] A single folded **attention** marker per space combines acute findings, audits ready
  to close, snoozed tasks now due, and unreadable files. It is absent, not zero, when there
  is nothing to attend to. Folding is deliberate — the atlas is an aggregate surface and
  the space's own overview is where you dig in.
- [x] A recency column shows the space's most recent task activity, derived from its epics'
  last-updated values.
- [x] **Entry points move to a pinned band below the table**, showing the focused space's
  entries with the selected one marked; `h`/`l` cycles within the band exactly as it does
  today. The band does not scroll with the table.
- [x] The pinned header and status row from finding H2 survive: everything the screen says
  about itself stays on screen, and the focused row is never scrolled off.
- [x] A space with no healthy entry point stays visible, keeps its error, and still cannot
  be entered.
- [x] Degrades on narrow terminals by dropping the least load-bearing columns rather than
  overflowing or wrapping.
- [x] Existing atlas tests are reworked rather than deleted where they assert the old flat
  layout.

## Out of scope

- The work view — its own task.
- Tiles and per-space accents. This task deliberately runs first so the tile decision can
  be re-taken against a table that may already deliver the wanted delight.
- Making attention actionable from the atlas (jumping to the acute finding). The fold is a
  signal here; digging in happens in the space.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Design: [The atlas as a dashboard of dashboards](../research/6g2qtp0022t7-the-atlas-as-a-dashboard-of-dashboards.md)

## Implementation notes (2026-08-23)

Each space is now one aligned row instead of a four-line prose card: health glyph, name,
aggregate progress bar, percentage, `done/total`, active count, a zero-dropping
`epics · audits · findings` breakdown, the folded `⚠` marker, and recency. Two spaces went
from ten lines to two, and they can now be compared by scanning a column.

**This removed more code than it added.** `renderAtlasCard` and `atlasSummaryLine` — 76
lines of hand-rolled text and width juggling — are gone, replaced by composition of
`s.miniBar`, `relDateCells`, `countsWidth`/`rollupCounts`, `theme.Percent`, and
`theme.Breakdown`, all of which the dashboard already used and the atlas had ignored.

**`⚠` means "wants a person", not "has findings".** It folds acute findings, audits ready
to close, snoozed tasks now due, and unreadable files; ordinary open findings are counted
in the breakdown beside it. The fixture demonstrates the distinction: bike-workshop shows
`3 findings` and no marker because all three are `soon`, while kitchen raises `⚠1` for a
single acute one. A pinned test asserts ordinary findings never raise it.

**Entry points moved to a pinned band** below the table, rendered for the focused space
with `h`/`l` cycling there unchanged. It sits outside the scrolled body deliberately: the
band answers "where would Enter take me", and a table tall enough to scroll is exactly when
that answer must not scroll away. This reuses the header/status split built for finding H2
rather than adding a layout mechanism. The band's height varies with the focused space's
entry count, so the table's scroll budget adapts — `scrollTo` keeps the focused row visible
either way.

A space whose summary could not be read keeps its row and shows the reason, rather than an
empty bar implying its tree was read and found empty.

Also fixed in passing: headers said `2 space(s)`. `countLabel` already knew how to say
`2 spaces`, so both view headers use it.

Verified by recording rather than by tests alone. Follow-on raised and tracked separately:
the space bar reuses the epic-completion gradient, so the two measures currently share one
visual language — [give-atlas-bars-their-own-visual-language](6g2zs0jq7b7b-give-atlas-bars-their-own-visual-language.md).

Validation: `go test -race ./...`, `golangci-lint run ./...`, planning lint, generated docs,
and `git diff --check` all clean.
