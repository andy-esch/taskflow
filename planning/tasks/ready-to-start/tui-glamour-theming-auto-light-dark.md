---
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Glamour body uses a fixed dark style; adapt to terminal background (auto light/dark) and consider a style aligned with the theme palette
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-12"
---

# TUI glamour theming auto light dark

## Objective

The S5 glamour body uses a fixed **`"dark"`** style (`glamour.go: glamourStyle`).
On a light-background terminal it reads poorly. Adapt the render style to the
terminal background, and consider aligning it with the shared `theme` palette so
the body matches the rest of the TUI. Deferred from S5 (out-of-scope: theming).

## Scope

- [ ] Pick the glamour style by terminal background — `WithAutoStyle()` or a
      `lipgloss.HasDarkBackground()` check — falling back to `dark`. Keep the cached
      renderer keyed by width (and now style).
- [ ] (Optional) a glamour `StyleConfig` aligned with the `theme` colors so headings
      / code / accents match the CLI/TUI palette.
- [ ] Test: the chosen style follows the background signal.

## Out of scope

- Async/off-loop glamour rendering — the per-width renderer cache (landed in the S5
  fast-follow) removed the recompile cost; only revisit if very long bodies lag.

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Follows [[tui-glamour-markdown-rendering-with-rawpretty-toggle]]
