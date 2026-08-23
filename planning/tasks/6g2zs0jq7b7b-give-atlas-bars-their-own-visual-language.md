---
schema: 1
id: 6g2zs0jq7b7b
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Distinguish the atlas's space-level progress bars from epic completion bars so the two are not read as the same measure.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, atlas, design-system]
created: "2026-08-23"
updated_at: "2026-08-23"
---
# Give atlas bars their own visual language

## Objective

The atlas spaces table renders each space's aggregate member-task completion with
`s.miniBar` — the same gradient `progressbar.Render` the dashboard and `status` use for
**epic** completion. Two different measures now share one visual language, and a reader
scanning the atlas can take a space bar for an epic bar.

The measures really are different in kind, which is the argument for separating them:

| | Epic bar | Atlas space bar |
| --- | --- | --- |
| Denominator | one epic's member tasks | every epic's member tasks in a space |
| Meaning | "this epic is 75% done" | "this project is 43% through its planned work" |
| Comparable across rows | yes, within a repo | yes, across repos |

## The actual question

Not "pick another color". The palette is a design system (epic 25), so the question is what
*principle* separates the two, applied once:

- **Hue** — a distinct palette ramp for aggregate/cross-space measures. Cheapest, but risks
  becoming "the atlas is the blue one" rather than a rule.
- **Glyph** — a different cell set (e.g. `▰▱` against the bars' `█░`) so the difference
  survives a monochrome terminal and colour-blind viewers, where hue alone does not.
- **Weight or width** — a shorter or thinner bar for aggregates, reading as "zoomed out".
- **Label** — leave the bar alone and let the adjacent `6/14` carry the distinction, on the
  argument that the confusion is speculative.

Worth deciding whether this generalises: if a third aggregate measure appears, does it join
the atlas language or get a third?

## Acceptance criteria

- [ ] The chosen distinction is written down as a rule in the design-system docs, not just
  applied — someone adding the next bar should know which language to use.
- [ ] Space-level and epic-level bars are distinguishable side by side, and the distinction
  survives **both** a monochrome terminal and every registered theme in both backgrounds.
- [ ] `internal/progressbar` grows the variant rather than the atlas hand-rolling one, so
  CLI surfaces can adopt it if `status --all` ever renders the same measure.
- [ ] Golden/rendering coverage for the new variant matching whatever the existing bars have.

## Out of scope

- The segmented finding bar (`s.segBar`), which is already visually distinct.
- Per-space accent colours — that is the tile task's question, and reopening it here would
  conflate "which measure is this" with "which space is this".

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Introduced by [Compose the atlas spaces view as a comparable table](6g2zqyra2s6h-compose-the-atlas-spaces-view-as-a-comparable-table.md)
- Design context: [The atlas as a dashboard of dashboards](../research/6g2qtp0022t7-the-atlas-as-a-dashboard-of-dashboards.md)
