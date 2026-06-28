---
schema: 1
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: progressbar gradient/track/segments and render.Style.ansiCode read the palette; pin neon-night's ANSI slots to today's SGR so stripped goldens stay byte-stable.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, tui, design]
created: "2026-06-28"
blocked_by: [design-package-foundation-palette-theme-registry-and-the-neon-default]
updated_at: "2026-06-28"
---
## Objective
Route the rollup/segmented bars and the CLI's 16-color ANSI map through the palette, with no porcelain churn.

## Scope
- In: `progressbar.Render`/`RenderSegments` take/close over a `Palette` (gradient, empty track, segment bands from `Palette.Of(theme.Color)`); `render.Style.ansiCode` becomes `sgr(pal.Of(c).ANSI)`. **Pin neon-night's ANSI slots to today's exact SGR codes** (red=31 … gray=90) so ANSI-stripped goldens stay byte-identical.
- Out: TUI chrome, picker, config. The bar gradient stays the deliberate truecolor exception.

## Done when
Bars + `ansiCode` read the palette; goldens unchanged; `build/test/lint` green.

## Reference
Design doc §4,§5. Depends on the design foundation. Parallel with the TUI-chrome task.
