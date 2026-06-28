---
schema: 1
status: in-progress
epic: 25-design-system-coherent-palette-and-selectable-themes
description: 'New internal/design package: Hue{True,ANSI}, Palette (semantic+chrome+gradient), Theme registry; define neon-night (base16) + neon-day (Catppuccin Latte). Additive, no consumers yet.'
effort: M
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, tui, design]
created: "2026-06-28"
updated_at: "2026-06-28"
started_at: "2026-06-28"
---
## Objective
Stand up `internal/design` as the single home for all concrete color decisions so every surface can later route through it. Additive only — nothing consumes it in this task. `internal/theme` stays untouched (domain-only semantic decisions).

## Scope
- In: `Hue{True color.Color, ANSI int}`; `Palette` — the 7 semantic slots keyed by `theme.Color` plus chrome tokens (accent, dim, selected, border-active, border-idle, danger, heading, match/match-current/match-fg, track) and the bar `Gradient`; `Theme{Name, Dark, Light}` with `Lookup`/`Default`/`For`. Define `neon-night` (base16 *Synth Midnight Terminal Dark*; danger-red → Outrun `#FF4242` for contrast) and `neon-day` (Catppuccin Latte via `github.com/catppuccin/go`, already in the graph). `design_test.go` pins each token's ANSI anchor + the gradient stops.
- Out: no surface wiring; no config; `internal/theme` enum/tokens unchanged.

## Done when
`internal/design` exists with palette/theme types + the two named themes + a pinning test; `build/test/lint` green; no other package imports it yet.

## Reference
Design doc §2,§3,§5 — `planning/research/2026-06-28-color-palette-and-theming-overhaul.md`. Umbrella: [[color-and-design-overhaul-one-coherent-palette-across-every-surface]].
