---
schema: 1
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: 'Visual+contrast pass on the light palette beyond the AA text fix already landed: find-highlight pair, gradient, gray/borders. Pairs with T5 (light path).'
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [tui, design]
created: "2026-06-29"
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
