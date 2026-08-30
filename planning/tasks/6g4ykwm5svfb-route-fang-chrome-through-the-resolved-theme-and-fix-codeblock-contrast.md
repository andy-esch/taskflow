---
schema: 1
id: 6g4ykwm5svfb
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Give the palette a background token, map fang's Codeblock to it, and resolve fang chrome from the selected theme instead of design.Default()
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [design, palette]
created: "2026-08-29"
updated_at: "2026-08-29"
depends_on: [6g50akp94xwr]
started_at: "2026-08-29"
completed_at: "2026-08-29"
---
# Route fang chrome through the resolved theme and fix codeblock contrast

## Objective

`repoColorScheme` in [`cmd/tskflwctl/main.go`](../../cmd/tskflwctl/main.go) is the one place
in the project that bypasses the palette-resolution pipeline, and it carries two independent
defects. Both are fixed in the same ~50-line function, and fixing either alone leaves the
other visibly worse. Merge them.

## Defect 1 — a foreground hue is assigned to a background role

fang's `Codeblock` is a **background** color. `repoColorScheme` assigns it `theme.ColorGray`,
a *foreground* text hue (`#a3a5a6` on dark). fang paints the USAGE and Examples boxes in that
light grey, then draws every token on top of it.

Confirmed byte-level from the real binary under a pty
(`script -q /dev/null ./bin/tskflwctl task --help`):

```
[48;2;163;165;166m                      <- box background #a3a5a6
[38;2;163;165;166;48;2;163;165;166m [command]   <- fg == bg. 1.00:1. invisible.
[38;2;3;174;255;48;2;163;165;166mtskflwctl      <- #03aeff on #a3a5a6
```

| token in the box | role | ratio |
| :--- | :--- | ---: |
| `[command]`, `[--flags]` | DimmedArgument | **1.00:1** |
| `tskflwctl` | Program | **1.00:1** |
| comments | Comment | **1.00:1** |
| `task` | Command | **1.53:1** |
| `--flag` | Flag | **1.53:1** |
| body text | Base → terminal default fg | **1.71:1** |
| `"quoted"` | QuotedString | **1.99:1** |

Every one of those clears AA (5.4–10.8:1) over fang's own default backdrop `#2F2E36`, and over
both palette `Base` hues. The hues are correct; the **role assignment** is not. The same tokens
rendered outside the codeblock (COMMANDS, FLAGS) carry no background and are perfectly legible,
which is why only this box is affected.

Root cause: `design.Palette` has **no background/surface token**. Every Hue in it is a
foreground, so the fang mapping had nothing correct to reach for.

## Defect 2 — fang chrome ignores the selected theme

`repoColorScheme` resolves `design.Default()` (neon) unconditionally, so `[theme]` in
`.tskflwctl.toml` and `--theme` never reach styled help or styled errors. The existing code
comment calls this an accepted limitation and a follow-up; it is not, in practice, acceptable —
it is every `--help` and every styled error on the human path.

Verified with `[theme] name = "catppuccin"` set in the repository config:

- `task list` emits `38;2;137;220;235` — `#89dceb`, Mocha sky. Correct.
- `task --help` emits `234;92;226`, `201;211;100`, `3;174;255`, `6;234;97`, `163;165;166` —
  **every one a neon hue, zero Mocha**.

An audit of the other surfaces found this is the *only* theme-resolution bypass: CLI body, TUI
chrome, TUI atlas, glamour markdown, and the huh picker all resolve the selected theme
correctly, and there are no hardcoded color literals anywhere outside `internal/design` except
the documented error-badge foreground (`lipgloss.Color("15")`).

## Why they must land together

Defect 1's fix introduces a **per-palette** surface token. Fix contrast alone and a catppuccin
user gets a *neon-tinted* backdrop against a Mocha terminal — more conspicuously wrong than
today's uniform grey. Fix theming alone and the box stays unreadable in a second theme's colors.

## Hard constraint: fang has exactly one `Base`

`fang.ColorScheme.Base` is used for BOTH the top-level help `Text` and the codeblock's text
foreground (`theme.go:127,142` in v2.0.1; `:126,141` in v1.0.0), and fang exposes no
`WithStyles` escape hatch — only `WithColorSchemeFunc` and the deprecated `WithTheme`. So the
codeblock's text colour cannot be set independently of the body's.

Therefore `Base` must stay `lipgloss.NoColor{}` (terminal default), and the surface must be a
**subtle lift from the palette's existing `Base` hue** — exactly what fang's own `#2F2E36` is
on dark — so terminal-default foreground remains legible over it. An explicit `Base` would
recolor all help body text away from the terminal default and make things worse.

## fang v2 sequencing (verified, not assumed)

Renovate PR #150 proposes `github.com/charmbracelet/fang v1.0.0` → `charm.land/fang/v2 v2.0.1`,
tracked by [upgrade-fang-to-v2](6g50akp94xwr-upgrade-fang-to-v2.md). Checked against the real
v2.0.1 source:

- `ColorScheme` is **byte-identical** between v1.0.0 and v2.0.1.
- `Codeblock` is still applied as `Background(...)`; `Base` is still the shared foreground.
- The option set is unchanged; there is still no `WithStyles`.
- `theme.go` differs only in the first-word capitalization helper (the multiline-error fix).

So this work is **version-agnostic** — no technical blocker either way. The ordering is
practical: both changes edit `cmd/tskflwctl/main.go`, and this task rewrites both
`repoColorScheme` and `TestRepoColorScheme`, which the upgrade task claims to verify. Doing the
mechanical import-path upgrade first means writing this once, against the final module path,
without conflicting with an open Renovate PR. Hence the `depends_on` edge — an ordering
constraint, not a technical one; drop it if the upgrade stalls.

Note for whoever picks up the upgrade task: its acceptance criterion "styled help … renders
correctly **with the active theme colors**" is false today and stays false after the upgrade.
Only this task makes it true.

## Acceptance criteria

- [x] `design.Palette` gains an explicit background/surface token (per palette, dark and light), documented as a *background* role so it can't be confused with a foreground hue. Values sourced from each theme's own scheme rather than invented — Catppuccin names `surface0` for Mocha and Latte; neon should take the corresponding base16 Synth Midnight Terminal Dark slot.
- [x] `repoColorScheme` maps `fang.ColorScheme.Codeblock` to that token. No foreground hue is assigned to a background role anywhere in the mapping.
- [x] Keep `fang.ColorScheme.Base` as `lipgloss.NoColor{}`. fang shares one `Base` between the help body text and the codeblock text and offers no `WithStyles` hatch, so the surface must instead stay close enough to the palette `Base` that terminal-default foreground remains legible over it.
- [x] `repoColorScheme` resolves the **selected** theme, not `design.Default()`. `[theme]` in config and `--theme` change styled help and styled error colors.
- [x] `TestRepoColorScheme` (`cmd/tskflwctl/main_test.go:195`) is updated — it currently asserts `design.Default().Dark`, which encodes defect 2 as expected behaviour.
- [x] A `design` test asserts (a) every *colored* fang role (`Comment`, `Program`, `Flag`, `DimmedArgument`, `Command`, `QuotedString`, `FlagDefault`, `Dash`, `Help`) clears WCAG AA 4.5:1 over the surface token, and (b) the surface stays within a small luminance delta of the palette `Base`, which is the property that keeps the uncheckable terminal-default `Base` legible over it. For every registered theme × both backgrounds — generalizing `TestFindHighlightContrastAA` rather than duplicating it.
- [x] Verified under a pty (`script -q /dev/null ./bin/tskflwctl task --help`): with `[theme] name = "catppuccin"` the emitted foregrounds are Mocha hues, and no token in the codeblock shares its background's color.
- [x] Eyeballed on a Catppuccin Mocha dark terminal and on a light terminal.

## Out of scope

- **The CLI body's hardcoded `a.Th.Dark`** (`internal/cli/root.go:101,352`). It honors `[theme]` but never resolves light vs dark, so on a light terminal `task list` renders dark-palette hues while `--help` renders light ones. A different axis from either defect here, and it has a real tradeoff behind it — `HasDarkBackground` fires an OSC-11 round-trip, which is why `root.go:405` resolves the markdown style lazily. Its own task.
- **User-defined palettes in `.tskflwctl.toml`.** `[theme]` accepts a *registered name* only (`config.go:465`). Epic 25's goal is "established, selectable themes", so this is an unbuilt feature rather than an adherence gap — worth deciding on separately.
- Adding a surface swatch row to `theme preview` human output. Cheap and useful (the existing `find` chip is the template, and `--variant` already lets you review the light palette from a dark terminal), but not required. Note it needs no wire work: `theme preview --json` emits only semantic swatches, so a new chrome field touches no envelope, no `schema_version`, and no golden.

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Sequenced after [upgrade-fang-to-v2](6g50akp94xwr-upgrade-fang-to-v2.md) (Renovate PR #150).
- Follows the completed `route-fangs-colorscheme-through-the-palette` and `validate-and-visually-tune-the-neon-day-light-palette`.
