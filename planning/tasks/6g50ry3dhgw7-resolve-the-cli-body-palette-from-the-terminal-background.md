---
schema: 1
id: 6g50ry3dhgw7
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: internal/cli/root.go hardcodes Th.Dark, so on a light terminal the body renders dark hues while fang help renders light ones
effort: 4-8 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [design, palette]
created: "2026-08-29"
updated_at: "2026-08-29"
---
# Resolve the CLI body palette from the terminal background

## Objective

`internal/cli/root.go:101,352` build the render Style with `a.Th.Dark` unconditionally.
The CLI body therefore honors the *selected theme* but never the *terminal background*:
on a light terminal, `task list` renders dark-palette hues.

`design.Theme.For(darkBG)` exists and every other surface uses it — TUI chrome
(`tui.go:36`), the TUI atlas, glamour markdown (`root.go:409`), the huh picker
(`prompt/tty.go:142`), and `theme preview`. The body is the only holdout.

## Why now

This became sharper after
`route-fang-chrome-through-the-resolved-theme-and-fix-codeblock-contrast`. fang chrome
resolves light/dark correctly through its `LightDarkFunc`, so on a light terminal
`tskflwctl --help` now renders the light palette while `tskflwctl task list` renders the
dark one — two surfaces of the same command disagreeing in the same terminal.

A completed sibling, `validate-and-visually-tune-the-neon-day-light-palette`, tuned and
validated the light palette specifically so it could be used. The body never selects it.

## The tradeoff to settle first

This is not an oversight to delete; there is a real cost behind it.
`lipgloss.HasDarkBackground` fires an OSC-11 round-trip against the terminal, and
`root.go:405` deliberately resolves the *markdown* style lazily for exactly that reason:

> `render.RenderBody` as a LAZY provider (not called eagerly): HasDarkBackground fires an
> [OSC query]

`setStyle` runs on every command including `--json` and completion, so an eager query
there would put a terminal round-trip on the hot path of every agent and shell-completion
invocation. Whatever lands must not.

Plausible shapes, in rough order of preference:

1. Resolve lazily, the way the markdown style already does — only when a colored human
   surface is actually about to render.
2. Gate the query on `wantColor(...) && isTerminal(out)`, so `--json`, pipes, CI, and
   completion never pay it.
3. Cache one resolution per process.

## Acceptance criteria

- [ ] On a light terminal the CLI body renders the light palette; on a dark terminal, the dark one.
- [ ] `--json`, non-TTY, and shell-completion runs perform no background query and stay byte-identical.
- [ ] `--theme` / `[theme]` selection still composes with the detected background.
- [ ] Body and fang help agree on light vs dark in the same terminal.
- [ ] A test pins that the query is not issued on the machine paths.

## Out of scope

- fang's own background handling, which already works.

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Builds on the completed `validate-and-visually-tune-the-neon-day-light-palette`.
