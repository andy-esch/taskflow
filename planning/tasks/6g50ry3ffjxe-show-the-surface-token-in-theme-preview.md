---
schema: 1
id: 6g50ry3ffjxe
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: The Surface background token has no review surface, unlike the find-highlight chip it is a sibling of
effort: 1-2 hours
tier: 4
priority: medium
autonomy_level: 3
tags: [design, palette]
created: "2026-08-29"
updated_at: "2026-08-29"
---
# Show the surface token in theme preview

## Objective

`theme preview` renders the semantic swatches, the find-highlight chip, the border rules,
and the bar — but not `Palette.Surface`, added by
`route-fang-chrome-through-the-resolved-theme-and-fix-codeblock-contrast`. It is the
palette's only background token and currently has no review surface, in the epic whose
purpose is making color decisions reviewable.

The find chip is the exact template: it exists because legibility of a background+foreground
pair can only be judged by seeing it, and `--variant light|dark` already lets a reviewer
inspect the palette they are not currently running.

## Why it matters beyond tidiness

The three surfaces are deliberately counter-intuitive and hard to sanity-check from the
source alone — neon lifts to base16 `base01`, Mocha *recesses* to `crust`, and Latte rises
to white. `TestChromeSurfaceContrastAA` and `TestChromeSurfaceSlots` pin the numbers, but
nothing lets a human confirm the box actually reads well before shipping a palette edit.

## Scope

- Add a surface row to `ThemePreviewHuman`, alongside the existing `find` and `border`
  rows, showing representative foreground roles drawn over the surface.
- Human output only. `theme preview --json` is the semantic-swatch machine contract; a
  chrome token does not belong in it, and keeping it out means no envelope change, no
  `schema_version` bump, and no golden churn.

## Acceptance criteria

- [ ] `theme preview` shows the surface with foreground text drawn on it, for both `--variant dark` and `--variant light`.
- [ ] `theme preview --json` is unchanged; no wire, schema, or golden edits.
- [ ] The row degrades to plain text with `--no-color`, like the existing chrome rows.

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
