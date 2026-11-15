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
updated_at: "2026-11-15"
audited: "2026-11-15"
audit_sources: [2026-11-15-weekly-task-sweep]
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

> **Update 2026-11-15** (automated weekly sweep): `internal/cli/root.go` churned again on
> 2026-10-25, so the "Actual today" column above is itself stale — `101`→`107`,
> `407`→`477`, `470–471`→`539–540`, `474`→`543`. The substance of this block is unchanged;
> see the 2026-11-15 table below for the current map.

No acceptance criterion ticked — none is met.

## Sweep verification 2026-11-15

Automated weekly sweep, re-checked against `main` (`5513c3d`).

**The premise is still fully intact.** Both body-palette sites still pass `a.Th.Dark`
unconditionally, and the CLI body is still the only surface that ignores the terminal
background. Nothing is ticked; no acceptance criterion is met.

**`internal/cli/root.go` has moved again.** The file churned on 2026-10-25, after the
2026-09-06 audit, so every `root.go` line reference in the body — including the ones
the 2026-09-06 table corrected — is stale again. Corrected against `5513c3d`:

| Reference | 2026-09-06 said | Actual today | Note |
| --- | --- | --- | --- |
| `setStyle`'s `WithPalette(a.Th.Dark)` | `root.go:101` | **`root.go:107`** | ⚠️ moved |
| post-`config.Discover` re-skin | `root.go:407` | **`root.go:477`** | ⚠️ moved |
| the LAZY / OSC-11 comment | `root.go:470–471` | **`root.go:539–540`** | ⚠️ moved; text unchanged |
| glamour markdown via `Th.For(HasDarkBackground(...))` | `root.go:474` | **`root.go:543`** | ⚠️ moved |
| `newStyles(th.For(dark))` | `internal/tui/tui.go:36` | `internal/tui/tui.go:36` | ✅ accurate |
| `p.theme.For(isDark).Accent` | `internal/cli/prompt/tty.go:142` | `internal/cli/prompt/tty.go:142` | ✅ accurate |

Both `a.Th.Dark` sites are unchanged in substance, and the lazy-markdown precedent the
task wants to imitate is intact at `root.go:539–543` — so recommended shape 1 (resolve
lazily, the way the markdown style already does) is still the right one, and the
`--json`/completion hot-path concern the task raises is unchanged.

**Note for whoever implements this:** the body's own line citations in the Objective and
"The tradeoff to settle first" sections still read `root.go:101,352` and `root.go:405/409`
as originally filed. They were left as filed rather than rewritten; this table and the
2026-09-06 one are the current map.

## Progress Log

- 2026-09-06: automated weekly sweep — premise re-confirmed; corrected three stale `root.go` line refs (352→407, 409→474, 405→470) after the 2026-09-05 cli churn.
- 2026-11-15: automated weekly sweep — premise re-confirmed; after the 2026-10-25 cli churn all four `root.go` refs moved again (101→107, 407→477, 470–471→539–540, 474→543); `tui.go:36` and `prompt/tty.go:142` still exact.
