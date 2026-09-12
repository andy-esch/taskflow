---
schema: 1
id: 6g9ebe9xzspb
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Wire the already-built theme.Staleness helper into relDateCells so dashboard dates carry the signal instead of a flat dim.
effort: S
tier: 3
priority: medium
autonomy_level: 4
tags: [tui, design, ux]
created: "2026-09-12"
updated_at: "2026-09-12"
completed_at: "2026-09-12"
---
# Colour the relative-date columns by staleness

## Objective

`theme.Staleness(days)` and `theme.DaysSince(date)` already exist, with thresholds chosen and
documented (gray under 30 days, yellow at 30, red at 90 — deliberately generous, because planning is
not a ticketing system). They have exactly one consumer, the atlas work list. Every other relative-date
column renders through `relDateCells`, which dims unconditionally and throws the signal away. Wire the
existing helper into those columns: an in-progress task untouched for two months should say so.

## Acceptance criteria

- [x] The dashboard's in-progress dates carry the staleness colour instead of a flat dim.
- [x] Column alignment is unchanged, with padding measured in display cells rather than bytes.
- [x] An undated or unparseable date stays neutral and never renders as alarming.
- [x] Whether the epics widget and the entity-list date columns also adopt it is decided rather than
  left inconsistent.
- [x] No new threshold is introduced — `theme.Staleness` stays the single place the boundaries live.

## Design notes

`relDateCells` (`internal/tui/column.go`) pre-measures the column, then wraps each cell in `st.dim(...)`
to pad it. Colouring means the pad and the colour interact, so the width measurement has to keep using
`ansi.StringWidth` over the styled string — the same discipline `padRight` already documents. The
existing comment in `relDateCells` about a blank cell still padding stays true and stays load-bearing.

Worth noting this is the one item from the mockup review that needs no new design decision at all: the
semantics, the thresholds, and the rationale were already settled in `theme/theme.go`, and only the
wiring is missing.

**Open:** whether epics inherit it. An epic's `LastUpdated` is a rollup over member tasks, so "stale"
may not mean the same thing there, and colouring it could imply neglect where there is none.

## Out of scope

- New staleness thresholds, or a staleness filter/sort.
- The CLI's date columns.

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Context: [mine the web mockup directions for the TUI](6g95m4eyvf4k-mine-the-web-mockup-directions-for-the-tui-data-organization-and-a-more-80s.md), axis A.
