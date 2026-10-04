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
updated_at: "2026-10-04"
audited: "2026-10-04"
audit_sources: [2026-10-04-weekly-task-sweep]
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

## Sweep verification (2026-10-04)

Premise verified still accurate, with one refinement. The conflation is real and
still present: `internal/tui/atlas.go:879` renders the space row with
`st.miniBar(pct, 10)` while `internal/tui/dashboard.go:168` renders epic completion
with `st.miniBar(pct, 8)`, and `internal/tui/style.go:146` is a one-liner —
`func (s *styles) miniBar(pct, width int) string { return progressbar.Render(pct, width, s.pal) }`
— so both measures resolve to the identical gradient.

Sharper evidence than the task currently carries: `style.go:142`'s doc comment
describes `miniBar` as *"the epic rollup bar (epic-list rows, epic-detail line)"*.
The helper is documented as epic-specific and the atlas calls it anyway, so the
conflation is visible in the source's own vocabulary, not only on screen.

**A partial distinction already exists by accident.** The two call sites pass
different widths (10 on the atlas, 8 on the dashboard/epic rows). That is the
"Weight or width" option in *The actual question* — already half-applied, but
undeclared and therefore not a rule: `detail.go:874` also uses width 12 for an epic
bar, so width currently tracks available space, not measure kind. Whoever decides
this should know the width axis is already spent as a layout variable, which is an
argument against choosing it as the semantic one.

Out-of-scope claim re-confirmed: `segBar` is genuinely a different renderer
(`progressbar.RenderSegments`, `progressbar.go:85`), used at `detail.go:842`,
`detail.go:1703`, and `item.go:436`.

2026-10-04: automated weekly sweep — conflation confirmed at `atlas.go:879` vs
`dashboard.go:168` through one `progressbar.Render`; noted that width is already
used as a layout variable (8/10/12), weakening it as the semantic distinction.
