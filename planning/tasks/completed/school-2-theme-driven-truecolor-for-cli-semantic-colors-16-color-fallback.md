---
schema: 1
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: CLI semantic glyph/status/percent colors emit the theme's truecolor hues on truecolor terminals (curated ANSI slot as 16-color fallback), so the theme shows on every CLI surface, not just bars/TUI.
effort: M
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, design]
created: "2026-06-29"
updated_at: "2026-06-30"
started_at: "2026-06-29"
completed_at: "2026-06-30"
---
## Objective
Make the CLI's **semantic** colors (status/finding glyphs, checkmarks, percent, priority, Green/Red/Warn) emit the active theme's **truecolor** hues on truecolor terminals, with the curated 16-color ANSI slot as the fallback. Today they always emit the 16-color slot, so the theme is invisible on the CLI glyphs (only the bars + TUI + picker show it). This is the "School 2 / holistic palette" decision — the theme should show on every surface.

## Why
The `Hue{Hex, ANSI}` model was built for exactly this: truecolor on capable terminals, a *chosen* ANSI slot on 16-color ones. The CLI was wired to always use the ANSI side; the TUI already renders truecolor. This completes the "one coherent palette across every surface" vision and resolves the current split-brain (CLI bars are theme-driven truecolor, CLI glyphs are terminal-native 16-color, on the same screen).

## Approach (contained — one method + a detector)
- `render.Style` gains `trueColor bool` + `WithTrueColor(bool)`.
- `ansiCode(c)` becomes: NoANSI → ""; else `trueColor` → `\x1b[38;2;r;g;bm` from `Hue.Hex` (a small `truecolorSeq`/hex parser); else `sgr(Hue.ANSI)` (unchanged 16-color fallback).
- Detect capability in `setStyle` via charm `colorprofile.Detect(out, env) == TrueColor` (a capability probe; on/off stays gated by `wantColor`).

## Degradation model
- Truecolor terminal → theme's exact hues (catppuccin pastel, neon glow).
- 16-color terminal → curated ANSI slot (today's exact behavior survives).
- 256-only terminal → the 16-color slot (Hue has no 256 value; acceptable, rare).
- Piped / `--json` / `--color=never` → plain, byte-stable (unchanged).

## Preserved / de-risked
- Byte-stability holds (still gated on `on`; truecolor strips like 16-color).
- Goldens are ANSI-stripped → zero churn. `TestSGR` stays (now the fallback path). Colored-output tests check for `\x1b[` presence → still pass.
- The palette + the TUI are untouched.

## Known caveat (folds into the light-mode work)
The CLI Style uses the theme's **Dark** palette (no eager OSC-11 background query, per the latency discipline). So on a *light* terminal with truecolor, semantic colors render the dark-tuned hexes (a contrast concern) — same deferral as the bar gradient. Ties into [[validate-and-visually-tune-the-neon-day-light-palette]] and a possible future lazy CLI background detection. Also: the bars still emit truecolor with no 16-color fallback (pre-existing; a separate polish).

## Reference
Supersedes the terminal-native assumption baked into [[route-progress-bars-and-the-cli-ansi-map-through-the-palette]] (T3). Design doc: `planning/research/2026-06-28-color-palette-and-theming-overhaul.md`.

**Done 2026-06-29 (branch feat/cli-truecolor-semantics).** CLI semantic colors now emit the active theme's truecolor hue on a truecolor terminal, falling back to the curated 16-color slot otherwise. Change is contained: `render.Style` gains `trueColor bool` + `WithTrueColor()`; `ansiCode` branches (NoANSI -> ""; trueColor -> `\x1b[38;2;r;g;bm` from Hue.Hex via `truecolorSeq`; else `sgr(Hue.ANSI)`); `trueColorCapable()` probes via charm `colorprofile.Detect` (now a direct dep); wired in setStyle. Verified: on a truecolor terminal the glyphs differ by theme (neon yellow #c9d364 vs catppuccin #f9e2af); on a non-truecolor terminal they fall back to the same curated slots as before; --json/piped stays byte-identical. Unit tests TestStyle_TrueColor + TestTruecolorSeq; goldens (ANSI-stripped) unchanged; gofmt/vet/full test green. CAVEAT (deferred, light-mode): CLI Style uses the theme's Dark palette (no eager OSC-11 query), so light-terminal truecolor semantics use dark-tuned hexes — folds into the light-palette work. Bars still emit truecolor with no 16-color fallback (pre-existing).

**Adversarial review 2026-06-29 (2 lenses) — delivers the vision, mergeable, no blockers.** Verified: every semantic CLI surface (status/bucket/finding/priority/percent + Green/Red/Warn + warnLinks ⚠, across status/doctor/lint/audit/epic) routes through the palette; truecolor CANNOT reach a pipe (Detect(non-TTY)=NoTTY) so --json/piped/--color=never stay byte-stable; truecolorSeq hex-parse + NoANSI ordering + 256→16 + colorprofile direct-dep all correct; goldens unchanged; TestSGR intact. Applied from the review: renamed ansiCode→colorSeq (it emits truecolor now), and documented the forced-color-to-pipe→16-color edge. DEFERRED (flagged, fold into the light/degradation work): (1) the CLI Style uses a.Th.Dark unconditionally, so on a LIGHT terminal School 2 now renders semantic glyph TEXT in dark-tuned hexes (contrast) — fix is to pass For(isDark) (needs CLI background detection); ties into [[validate-and-visually-tune-the-neon-day-light-palette]]. (2) bars still emit truecolor with no 16-color fallback — the lone surface that doesn't honor the degradation tier (pre-existing, low impact; modern emulators approximate).
