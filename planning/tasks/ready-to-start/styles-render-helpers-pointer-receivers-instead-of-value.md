---
schema: 1
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: styles is a ~10KB struct (13 lipgloss.Style fields); render helpers are value receivers copying it per call in hot row paths — switch to pointer receivers to match the *styles storage elsewhere
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [design, tui, refactor]
created: "2026-06-29"
---

# styles render helpers: pointer receivers instead of value

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [[25-design-system-coherent-palette-and-selectable-themes]]
