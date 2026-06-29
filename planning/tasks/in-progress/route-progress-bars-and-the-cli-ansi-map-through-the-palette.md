---
schema: 1
status: in-progress
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
started_at: "2026-06-28"
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

**Implementation 2026-06-28 (branch feat/cli-bars-palette, off updated main w/ T2).** progressbar.Render/RenderSegments now take a design.Palette (gradient + empty track + segment bands from the semantic slots); render.Style carries a palette field (defaulted to design.Default().Dark in NewStyle — semantic ANSI slots are background-independent) and ansiCode is now a method returning sgr(palette.Of(c).ANSI); Green/Red/Warn route through it too, so the ansiRed..ansiGray literals are gone (only the structural Reset/Bold/Dim consts remain). TUI miniBar/segBar pass the package pal. Byte-stability: neon-night's ANSI slots map to the exact prior SGR codes (red=31…gray=90), so colored CLI output is unchanged and the ANSI-stripped goldens are untouched — verified by the full suite (render + fixture-driven CLI integration tests) plus gofmt/vet, all green. Gotcha: an in-package progressbar_test.go calls Render unprefixed (a plain-grep for 'progressbar.Render' missed it) — updated its 5 call sites to pass a test palette.
