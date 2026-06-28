---
schema: 1
status: active
description: Unify all color/visual-design decisions behind one palette and established, selectable themes (base16, neon default) across CLI and TUI; home for cohesive design + theming work.
priority: medium
tags: [cli, tui, design]
created: "2026-06-28"
---

# Design system — coherent palette and selectable themes

**Goal.** Unify all color/visual-design decisions behind one palette and established, selectable themes (base16, neon default) across CLI and TUI; home for cohesive design + theming work.

## Why this is its own epic

The color decisions are spread across `theme`, `progressbar`, the picker, the TUI
lipgloss chrome, huh, and glamour with no single source. Unifying them is a
multi-step refactor (a new `internal/design` palette layer, per-surface routing,
config-driven selection, discovery commands) that's a coherent initiative in its
own right rather than a single CLI-ergonomics task. It's also the natural bucket
for future *design* concerns (glyph sets, borders, layout) beyond color.

## Decisions (from research)

- **Lean on the `base16` standard, not a custom system** — its 16 slots map
  slot==slot onto the CLI's 16 ANSI colors, so degradation is deliberate, not a
  runtime nearest-color guess.
- **Default `neon-night`** ported from base16 *Synth Midnight Terminal Dark*
  (danger-red -> Outrun `#FF4242` for contrast); light fallback `neon-day` ~=
  Catppuccin Latte (`github.com/catppuccin/go`, already in the graph).
- **New `internal/design` package** owns a truecolor `Palette` + `Theme` registry;
  `internal/theme` stays domain-only (the semantic decision layer). Each token
  carries an explicit ANSI-16 anchor.
- **`[theme]` config table** mirrors `[pager]`; precedence `--theme` > env >
  config > `auto`.

## Sequence

```
T1 foundation (done) --+-- T2 TUI chrome ---+
                       +-- T3 bars + ANSI --+-- T5 config/selection -- T6 discovery + polish --> umbrella
                       +-- T4 picker -------+
```

1. `design-package-foundation-palette-theme-registry-and-the-neon-default` (merged)
2. `route-tui-chrome-through-the-palette` · 3. `route-progress-bars-and-the-cli-ansi-map-through-the-palette` · 4. `route-the-interactive-picker-theme-through-the-palette` (T2-T4 parallel off T1)
5. `theme-config-table-and-selection-plumbing`
6. `theme-discovery-commands-glamour-polish-and-a-second-theme`

The broad `color-and-design-overhaul-one-coherent-palette-across-every-surface`
task is the north-star/definition-of-done, gated on the chain's tail.

## Out of scope (explicit non-goals)

- **Glyph set, border style (`RoundedBorder`), spacing** — stay fixed. The glyphs
  are the accessibility fallback (state survives `--color=never`/mono/colorblind);
  don't make them swappable. This overhaul is COLOR only.
- The agent/pipeline contract is untouched: color is TTY-gated; `--json`/piped
  output stays byte-stable plain.

## Reference

Full audit + design: `planning/research/2026-06-28-color-palette-and-theming-overhaul.md`.
