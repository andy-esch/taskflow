---
schema: 1
id: 6g4ykwm5svfb
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Give the palette a background/surface token so fang's help codeblock stops painting light-grey-on-light-grey
effort: 2-4 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [design, palette]
created: "2026-08-29"
updated_at: "2026-08-29"
---
# fix contrast in entity help usage box

## Objective

`tskflwctl <entity> --help` paints its USAGE and Examples blocks in a light grey on which
nothing is readable. The cause is a **role mismatch**, not a hue that needs tuning:
fang's `Codeblock` is a *background* color, and
[`repoColorScheme`](../../cmd/tskflwctl/main.go) assigns it `theme.ColorGray` — a
*foreground* text hue. On a dark terminal that is neon's `#a3a5a6`, so fang paints the
whole box light grey and then draws every token on top of it. `[command]` and `[--flags]`
use `DimmedArgument`, which is mapped to the same `gray`, so they render at exactly
**1.00:1** — literally invisible.

The rest of the help output (short description, flag list, command list) is drawn outside
the codeblock on the terminal's own background, which is why only this box is unreadable.

The structural root cause: `design.Palette` has **no surface/background token**. Every Hue
in it is a foreground. The fang mapping had nothing correct to reach for.

## Evidence

fang (`fang@v1.0.0/theme.go:138-168`) builds the codeblock as
`Background(cs.Codeblock).Foreground(cs.<role>)` for `Base`, `Text`, `Comment`,
`Program.Name`, `Flag`, `Argument`, `DimmedArgument`, `Command`, and `QuotedString`.
Its own dark default is `#2F2E36`, a dark backdrop. We pass a light grey instead.

Measured WCAG 2.1 ratios over the resulting `#a3a5a6` box:

| token in the box | role | color | ratio | |
| :--- | :--- | :--- | ---: | :--- |
| `[command]`, `[--flags]` | DimmedArgument | `#a3a5a6` | **1.00:1** | invisible |
| `tskflwctl` | Program | `#03aeff` | **1.00:1** | invisible |
| comments | Comment | `#a3a5a6` | **1.00:1** | invisible |
| `task` | Command | `#c9d364` | **1.53:1** | fails |
| `--flag` | Flag | `#06ea61` | **1.53:1** | fails |
| body text | Base = `NoColor` → terminal fg | ~`#cdd6f4` | **1.71:1** | fails |
| `"quoted"` | QuotedString | `#42fff9` | **1.99:1** | fails |

Every one of those same tokens clears AA (5.4:1 – 10.8:1) over fang's own `#2F2E36`, and
they all clear AA over both palette base colors (`#050608` neon, `#1e1e2e` Mocha). The
hues are fine; the background assignment is wrong.

Reported on a Catppuccin Mocha dark terminal; the defect is background-independent because
`gray` is a mid-tone in both the light and dark palettes.

## Why the existing guard missed it

`internal/design/design_test.go` already owns `contrastRatio` and
`TestFindHighlightContrastAA`, which guards the *only other* foreground-over-colored-background
pair in the project (`MatchFg` over `Match`/`MatchCurrent`). The fang pairing escaped it for
two reasons: the pairing is declared in `cmd/tskflwctl/main.go`, outside `internal/design`,
and the palette exposes no background token the test could have checked against.

## Acceptance criteria

- [ ] `design.Palette` gains an explicit background/surface token (dark and light variants per palette), documented as a *background* role so it can't be confused with a foreground hue.
- [ ] `repoColorScheme` maps `fang.ColorScheme.Codeblock` to that token. No foreground hue is assigned to a background role anywhere in the mapping.
- [ ] Keep `fang.ColorScheme.Base` as `lipgloss.NoColor{}` (terminal default).
  fang has exactly ONE `Base`, shared by the top-level help `Text` and the
  codeblock text (`fang@v1.0.0/theme.go:126,141`), and exposes no `WithStyles`
  escape hatch — an explicit `Base` would recolor all help body text away from
  the terminal default. The surface must instead be a subtle lift from the
  palette's existing `Base` hue (as fang's own `#2F2E36` is on dark), so
  terminal-default foreground stays legible over it.
- [ ] A `design` test asserts (a) every *colored* fang role (`Comment`,
  `Program`, `Flag`, `DimmedArgument`, `Command`, `QuotedString`, `FlagDefault`,
  `Dash`, `Help`) clears WCAG AA 4.5:1 over the surface token, and (b) the
  surface stays within a small luminance delta of the palette `Base`, which is
  the property that keeps the uncheckable terminal-default `Base` legible over
  it. For every registered theme x both backgrounds, generalizing
  `TestFindHighlightContrastAA` rather than duplicating it.
- [ ] `tskflwctl task --help` (and epic/audit/research/thread) is legible end to end on a Catppuccin Mocha dark terminal and on a light terminal.

## Out of scope

- fang chrome always resolving `design.Default()` (neon) instead of the selected `--theme` / `[theme]`. `cmd/tskflwctl/main.go:86-89` calls this out as an accepted limitation. Note the interaction: once the codeblock backdrop becomes a real per-palette surface, a catppuccin user gets a *neon-tinted* box against a Mocha terminal, so this gets more visible, not less — but it is its own change.
- The CLI body hardcoding `a.Th.Dark` (`internal/cli/root.go:101,352`) while fang uses `lipgloss.LightDark`. On a light terminal `task list` renders dark-palette hues while `--help` renders light ones. Deliberate-looking (the comment at `root.go:405` notes `HasDarkBackground` fires an OSC query, hence the lazy markdown path), but it is an inconsistency in this epic's remit.
- The TUI's own overlays. They set only foregrounds over the terminal background and are unaffected.

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Builds on the completed `route-fangs-colorscheme-through-the-palette` and `validate-and-visually-tune-the-neon-day-light-palette`.
