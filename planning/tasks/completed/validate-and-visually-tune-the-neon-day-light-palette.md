---
schema: 1
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: 'Visual+contrast pass on the light palette beyond the AA text fix already landed: find-highlight pair, gradient, gray/borders. Pairs with T5 (light path).'
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [tui, design]
created: "2026-06-29"
updated_at: "2026-06-30"
started_at: "2026-06-30"
completed_at: "2026-06-30"
---
## Objective
A proper visual + contrast pass on the neon-day (light) palette, beyond the AA text-color fix already landed in [[pre-t5-design-hardening-from-the-adversarial-review]].

## Context
The light palette (Catppuccin-Latte-inspired) was never visually validated. The hardening pass darkened the four semantic TEXT accents (green/yellow/teal/blue) to clear WCAG AA (>=4.5:1) on the Latte bg — but other light tokens still need review:
- **Find-highlight pair:** neonLight `Match` is a yellow bg (`#df8e1d`) with `MatchFg` = near-white (`#eff1f5`) — light-on-yellow, low contrast. neonDark uses dark-on-yellow. The light pair likely needs dark text on the highlight (or a darker highlight bg).
- **Light gradient** (mauve -> sapphire -> pink) on a light terminal — visual check.
- **Gray** (`#6c6f85`), **BorderIdle**, **Track** legibility on a light bg.
- A real eyeball pass on a light terminal — the contrast math is necessary but not sufficient.

## Scope
- Validate/tune `Match`/`MatchFg`, the gradient, borders, and gray for the light bg.
- Extend `design_test` to pin any chrome value that gets tuned.
- Pairs naturally with T5 (which wires background -> palette selection and exercises the light path).

## Reference
Review findings F1-F2 (accessibility lens) + the hardening pass that fixed the text-contrast subset.

**Progress 2026-06-30 (autonomous slice).**
- **Bug fixed:** light find-highlight was white `#eff1f5` on amber `#df8e1d` = ~2.3:1 (unreadable). Flipped the light pair to dark-text-on-light-bg (highlighter-pen look) so one shared MatchFg clears AA on both backgrounds: MatchFg → `#1e1e2e`, MatchCurrent → `#c9a6f8` (lightened mauve; was the saturated `#8839ef` accent that needed white text). Match (amber) kept. Now ~6.3:1 / ~8:1.
- **Test added:** `TestFindHighlightContrastAA` computes real WCAG ratios and asserts ≥4.5:1 for match+current on BOTH variants (would have caught the bug; guards regressions).
- **Tooling:** `theme preview` gained `--variant auto|dark|light` so the light palette is reviewable from any terminal; the HUMAN preview now also renders find-highlight chips + border samples (JSON unchanged). docs/cli regenerated.
- **Still needs eyeballs (visual sign-off):** the tuned find values, the light gradient (mauve→sapphire→pink), BorderIdle/Track/gray on a Latte bg. Math is necessary, not sufficient.

- **Preview canvas:** the human `theme preview` now paints swatch tiles + border rules on the variant's intended bg (Latte `#eff1f5` for light, dark base for dark), so the light palette is judged on its own background — no need to switch the terminal to light. find chips carry their own bg; the bar is self-contained.

**Sign-off 2026-06-30.** User reviewed `theme preview neon --variant light` (on the new light canvas) and approved all four: find chips, gradient, borders, gray. Tuned values pinned by `TestNeonLightHighlight` (exact hexes) + `TestFindHighlightContrastAA` (WCAG ≥4.5:1). Gradient/borders/gray were validated as-is (unchanged). Done pending merge.
