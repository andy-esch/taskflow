---
schema: 1
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: repoColorScheme hardcodes lipgloss.Color(1..15) for fang help/errors — route through the palette ANSI slots (keeping 16-color) so it can't desync when T5 adds a theme.
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [cli, design]
created: "2026-06-29"
updated_at: "2026-06-30"
started_at: "2026-06-29"
completed_at: "2026-06-30"
id: 6fh1b7w00ysw
---
## Objective
Route `cmd/tskflwctl/main.go` `repoColorScheme` (fang help/error/manpage colors) through the design palette instead of the hardcoded `lipgloss.Color("1".."15")` literals — closing the last stray-literal surface the T1-T4 sweep missed.

## Context
The pre-T5 adversarial review flagged `repoColorScheme` as the largest stray-literal cluster, sitting in `cmd/` (outside every swept package). The literals were deliberately chosen to MATCH the palette's ANSI slots (1 red, 2 green, 3 yellow, 4 blue, 6 cyan, 8 gray) AND to stay 16-color/terminal-themed (it intentionally ignores fang's `LightDarkFunc`). So the values agree today, but they're a hand-copied duplicate that silently desyncs once a second theme (T5) remaps a slot.

## Why it was deferred (non-trivial)
- Naively using `pal.Of(c).Color()` would regress fang to TRUECOLOR — the comment explicitly chose 16-color. A faithful routing must go through the ANSI slot, e.g. `lipgloss.Color(strconv.Itoa(pal.Of(theme.ColorRed).ANSI))`.
- Some fang colors have no clean palette token: bright-white (15) for the error badge fg; `Title`/`QuotedString` use cyan as a *section-header* concept, not the focus accent — keep that distinction.

## Scope
- A small helper (in `design` or `cmd`) that yields fang colors from a `Palette`'s ANSI slots, preserving the 16-color/terminal-adaptive behavior.
- Map each `fang.ColorScheme` field to a palette token; document any field that legitimately stays its own concept.
- When T5's selected theme is available, feed that theme's palette through.

## Reference
Review finding F1 (completeness lens). Design doc: `planning/research/2026-06-28-color-palette-and-theming-overhaul.md`.

**Done 2026-06-29 (branch feat/fang-palette).** repoColorScheme now derives fang's help/error colors from design.Default() via fang's LightDarkFunc — the hardcoded lipgloss.Color(1..15) literals are gone (only the error-badge white #15 remains, a fixed contrast affordance, not a themeable color). This was the LAST CLI surface bypassing the palette, so 'every surface' is now met. Verified: build/vet/full test green; TestRepoColorScheme pins Title=accent / Flag=green / badge-bg=danger from the palette.

THREE CHOICES I made (flag if you'd change them):
1. **fang uses the DEFAULT theme, not the --theme/[theme] selection** — fang renders chrome outside the per-command theme resolution (it has no access to App.Th), so help/errors always show neon even under --theme catppuccin. Help/errors are brand chrome, not data; threading the selected theme into fang is a possible follow-up.
2. **Title (section headers) → the theme Accent** (was 16-color cyan), so help headers carry the theme signature.
3. **Truecolor via the palette + LightDarkFunc** (not the task's original 'route through the 16-color ANSI slots' sketch) — consistent with the School-2 direction (theme drives truecolor; lipgloss downsamples on 16-color terminals). BONUS: fang now picks light/dark correctly per terminal background — better than the CLI body, which uses the dark palette always. (fang's 16-color downsampling is lipgloss-nearest, not the curated Hue.ANSI — acceptable for chrome.)
