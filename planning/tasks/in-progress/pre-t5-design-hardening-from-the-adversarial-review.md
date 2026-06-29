---
schema: 1
status: in-progress
epic: 25-design-system-coherent-palette-and-selectable-themes
description: 'Address the high-value findings from the T1-T4 adversarial review before T5: fang palette routing, ARCHITECTURE.md, panic-proof Of(), neon-day contrast, and contract/light tests.'
effort: M
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, tui, design]
created: "2026-06-29"
updated_at: "2026-06-29"
started_at: "2026-06-29"
---

# Pre-T5 design hardening from the adversarial review

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [[25-design-system-coherent-palette-and-selectable-themes]]

**Done 2026-06-29 (branch fix/design-pre-t5-hardening).** Landed: (1) Of() panic-proofed — Semantic is now a map[theme.Color]Hue, unknown slot degrades to no-color instead of an index-out-of-range panic; (2) neon-day light accents darkened to clear WCAG AA on the Latte bg (green #2e7d1f 4.6:1, amber #8a6000 5.0:1, teal #0e6e74 5.3:1, blue #2258cc 5.6:1) + the false 'glyphs not text' comment corrected; (3) dead Palette fields Fg/Dim/Selected dropped; (4) ARCHITECTURE.md gains an internal/design bullet + the TUI style.go line fixed; (5) tests added — sgr() 16-color contract (red=31..gray=90, NoANSI=''), neon-day semantic slots, Of() unknown-color degrade. Deferred to tasks: fang routing [[route-fang-s-colorscheme-through-the-palette]], light visual tuning [[validate-and-visually-tune-the-neon-day-light-palette]], and the applyTheme->newChrome + Hue.ANSI decision folded into [[tui-palette-as-model-scoped-styles-instead-of-package-globals]]. gofmt/vet/full go test green.
