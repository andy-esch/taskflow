---
schema: 1
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: '[theme] config table (name=auto|preset|none, dark/light) mirroring [pager]; precedence --theme > env > config > auto; thread the chosen theme through App into render/tui/prompt.'
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, tui, design]
created: "2026-06-28"
blocked_by: [route-tui-chrome-through-the-palette, route-progress-bars-and-the-cli-ansi-map-through-the-palette, route-the-interactive-picker-theme-through-the-palette]
updated_at: "2026-06-28"
---
## Objective
Let users select a theme via config/env/flag, and feed it to every routed surface.

## Scope
- In: `ThemeConfig` mirroring `PagerConfig` — `[theme]` table (`name = "auto" | <preset> | "none"`, `dark`/`light`); precedence `--theme` > `TSKFLW_THEME` > `[theme].name` > `auto`, gated behind `wantColor()`. `App.Theme` set in `resolve()`; threaded into `render.Style` (`WithTheme`), `prompt.NewTTY`, `tui.Run` — no globals, background detection kept lazy (OSC-11 latency). Commented `[theme]` block in the default config + a config round-trip test. The hardcoded `Default()` from the routing tasks becomes config-selected.
- Out: discovery commands + polish (next task).

## Done when
A `[theme]` selection flows end-to-end to CLI + TUI + picker; precedence honored; `build/test/lint` green. If a `--theme` flag is added, regenerate docs (docs-check gate).

## Reference
Design doc §5. Depends on the three routing tasks (TUI chrome, bars/ANSI, picker).
