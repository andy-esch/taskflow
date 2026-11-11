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
updated_at: "2026-11-08"
audited: "2026-11-08"
audit_sources: [2026-11-08-weekly-task-sweep]
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

## Sweep verification (2026-09-06)

Automated weekly sweep. The task's premise is **fully intact** — the body palette
is still hardcoded to `a.Th.Dark` and is still the only surface that ignores the
terminal background. But `internal/cli/root.go` has moved since 2026-08-29 and
three of the line references in the Objective / "The tradeoff to settle first"
are now stale.

### Line references — corrected against HEAD `84b3798`

| Cited in body | Actual today | Note |
| --- | --- | --- |
| `root.go:101` | `root.go:101` | ✅ accurate — `setStyle`'s `WithPalette(a.Th.Dark)` |
| `root.go:352` | **`root.go:407`** | ⚠️ moved; line 352 is now unrelated space-resolution code |
| `root.go:409` (glamour markdown) | **`root.go:474`** | ⚠️ moved — `a.Th.For(lipgloss.HasDarkBackground(...)).Markdown` |
| `root.go:405` (the LAZY comment) | **`root.go:470–471`** | ⚠️ moved; comment text unchanged |
| `tui.go:36` | `internal/tui/tui.go:36` | ✅ accurate — `newStyles(th.For(dark))` |
| `prompt/tty.go:142` | `internal/cli/prompt/tty.go:142` | ✅ accurate — `p.theme.For(isDark).Accent` |

Both `a.Th.Dark` sites are unchanged in substance: `root.go:101` in `setStyle` and
`root.go:407` in the post-`config.Discover` re-skin. The lazy-markdown precedent the
task wants to imitate is intact at `root.go:470–474`, so the recommended shape
(option 1) is still the right one.

No acceptance criterion ticked — none is met.

> **Update 2026-11-08**: the `root.go` line numbers in this table have drifted again —
> the file is now 569 lines (was ~480). Current: the `setStyle` `WithPalette(a.Th.Dark)`
> is `root.go:107` (was 101), the post-`config.Discover` re-skin is **`root.go:477`**
> (was 407), the glamour markdown resolution is **`root.go:543`** (was 474), and the
> LAZY comment is **`root.go:539`** (was 470–471), its text unchanged. `tui.go:36` and
> `prompt/tty.go:142` are both still exact. Nothing moved in substance: both `a.Th.Dark`
> sites and the lazy-markdown precedent are intact, so option 1 remains the right shape.

## Sweep verification (2026-11-08)

Automated weekly sweep. The premise is **still fully intact** — the CLI body palette
is hardcoded to `a.Th.Dark` at both sites and remains the only surface that ignores
the terminal background. The line references drifted a second time and are corrected
in the blockquote above; `internal/cli/root.go` has absorbed ~90 lines since
2026-09-06 without touching either `Th.Dark` call.

The two sites, verified verbatim:

- `root.go:107` — `render.NewStyle(wantColor(…)).…WithPalette(a.Th.Dark)` in
  `setStyle`, immediately after `a.resolveTheme()`, whose own comment notes the repo
  `[theme]` folds in later.
- `root.go:477` — `a.Style = a.Style.WithPalette(a.Th.Dark)` in the
  post-`config.Discover` re-skin, where the comment explains the re-skin exists so a
  config-selected theme takes effect. Both are theme-aware and background-blind,
  exactly as the Objective describes.

The tradeoff the task says to settle first is also unchanged: `root.go:539–543` still
resolves the markdown style lazily through
`a.Th.For(lipgloss.HasDarkBackground(os.Stdin, os.Stdout)).Markdown`, with the LAZY
comment the task quotes still present word for word. So the precedent the task wants
to imitate is live, and AC 2's "no background query on the machine paths" is still
the binding constraint rather than a hypothetical one.

No acceptance criterion ticked — none is met.

## Progress Log

- 2026-09-06: automated weekly sweep — premise re-confirmed; corrected three stale `root.go` line refs (352→407, 409→474, 405→470) after the 2026-09-05 cli churn.
- 2026-11-08: automated weekly sweep — premise re-confirmed; corrected four `root.go`
  line refs for the second time (101→107, 407→477, 474→543, 470→539) after October's
  cli churn, with both `Th.Dark` sites unchanged in substance.
