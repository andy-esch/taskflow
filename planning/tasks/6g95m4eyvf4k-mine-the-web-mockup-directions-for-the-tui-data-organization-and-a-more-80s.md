---
schema: 1
id: 6g95m4eyvf4k
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Weigh the three web mockup board directions against the TUI, core.Board and core.Summary — data organization plus a more thoroughly 80s chrome — and document what transfers before deciding.
effort: M
tier: 2
priority: medium
autonomy_level: 2
tags: [tui, design, theme, ux, board]
created: "2026-09-11"
---
# Mine the web mockup directions for the TUI: data organization and a more 80s chrome

## Objective

Three web-UI board directions were drafted in a design canvas (`Taskflow Web.dc.html`: 1a
terminal-faithful, 1b miami vice, 1c vector arcade) against the real `board` projection and the
`neon-night` palette. They are worth mining on two independent axes — how the data is ORGANIZED, and
how the chrome LOOKS — but nothing here is decided yet. This task is the written record of what the
mockups contain, what each idea would cost in a terminal, and which of them the repo can already
express. Decisions and implementation spin out as their own tasks.

Framing that matters: the mockups are downstream of this repo, not upstream of it. Every hex in 1a is
verbatim `neonDark` from `internal/design/theme.go` and every glyph is verbatim `theme.Status`. There
is no palette or vocabulary to "port" — the transferable material is layout, density, and the
non-hue parts of the chrome.

## Acceptance criteria

- [ ] Axis A (data organization) is written up as a comparison of the mockup board against what
  `core.Board`, `core.Summary`, the CLI `board`/`status` renderers, and the TUI Overview each show
  today, naming which differences are deliberate and which are drift.
- [ ] Axis B (80s chrome) is written up as an inventory of terminal-available techniques beyond hue,
  each with its cost, its degradation behaviour, and whether the current `design.Palette` can express
  it at all.
- [ ] The width budget for any multi-column board layout is stated against the TUI's real breakpoints,
  with the narrow fallback named rather than assumed.
- [ ] Accessibility and degradation guardrails are stated for a new theme: AA coverage and its gaps,
  16-color fallback, `--color=never`/mono, and the dark/light pairing obligation.
- [ ] Each candidate idea is marked take / defer / reject with a one-line reason, and the takes are
  filed as their own tasks on the appropriate epic.
- [ ] No TUI, palette, or renderer code is changed by this task.

## Axis A — data organization

Weighed against the two projections that already exist.

**The board has no TUI surface at all.** `core.Board` has exactly one consumer,
`internal/cli/board.go:20`. The TUI Overview leads with `in progress` — work already started — and
never shows next-up or ready-to-start, so the screen does not answer "what do I do next". The mockups
answer it with three status columns plus a right rail of rollups. Both halves are computable today
from `core.Board` + `core.Summary` with no new core work. This is the one structural idea in the
mockups that is genuinely absent from the tool.

**The TUI Overview is a subset of CLI `status`.** Measured against the `demo-kitchen` fixture:

| | CLI `status` | TUI Overview | mockup |
| :-- | :-- | :-- | :-- |
| counts line | `active 4 next-up · 4 ready-to-start · 3 in-progress` / `archived 9 completed · 2 deprecated · 2 deferred` | absent | present |
| open audit | `2026-08-15-pantry-staples ███▓▓▒▒▒░░ 66% settled 4/6` | `◆ 1 open audit(s)` | stacked bar + legend |
| graph health | warning line when unhealthy | info row when unhealthy | persistent `graph healthy` chip |

The counts line is pure `Summary.Counts`, already loaded by the dashboard and already rendered by
`status`. `styles.segBar` already exists and is used at `item.go:437` and `detail.go:1444`, so the
dashboard degrading a settled bar to a bare count is a downgrade from both its own audits tab and the
CLI. Audit M2 recorded an intent that the two dashboards agree; these three rows are where they do
not.

**Blocked work never reaches the TUI.** `Board.Blocked` drives `sortEligibleFirst` and the CLI's
dimmed `⛔` row; `grep -n "locked" internal/tui/` finds only `GateBlocked` in the spatial graph.
`taskDelegate.Render` has no blocked marker. The mockups go one step further and NAME the blocker
(`⚠ blocked · needs 6g2pjz0h`), which the projection cannot do today — `Blocked` is a
`map[string]bool`, so naming the gating edge means widening the projection.

**Staleness is computed and then discarded.** `relDateCells` computes `theme.RelativeDate` and
unconditionally dims it. 1b renders `34d stale` in orange while sibling ages stay dim. A staleness
tier would follow the existing `theme.Percent()` tier precedent.

**Row shape is a real tradeoff, not a detail.** The mockup card is two lines (title, then
`id · epic · age`); `taskDelegate.Height()` is 1. Adopting the card halves rows per screen.

**Width.** The mockups are drawn at 1240px / ~12px IBM Plex Mono ≈ 170 columns: ~115 for three ~38-col
columns, ~46 for the rail. The TUI's own breakpoints are `twoPane` at ≥90 (`view.go:38`) and the
collapsed tab chip at <60. A board surface therefore needs a stated degradation path — drop the rail,
then stack the columns, which is what `tskflwctl board` already prints.

## Axis B — a more thoroughly 80s chrome

The palette is already synthwave (base16 Synth Midnight; `Default()` even calls itself "neon / 80s"),
so "more 80s" cannot mean more neon hue. Everything below is a non-hue lever.

**1. Fill / surface.** The TUI paints foreground only. Much of the mockups' period feel comes from
filled panels — 1b's indigo body and gradient header band, 1a's `#0b0c10` header strip and `#111216`
card fills over the `#050608` body. `lipgloss.Style.Background()` is available; so are zebra row
tints and reverse-video selection (the DOS idiom) in place of the `› ` marker. This is the biggest
lever and the one with real architectural weight: `design.Palette` has exactly ONE background token
(`Surface`) and its doc comment explicitly declares that policy, so growing it is a decision to
revisit deliberately, not an oversight to patch.

**2. Border vocabulary.** `RoundedBorder()` is hardcoded in six places in `newStyles`
(`style.go:69,70,78,86,87,93`) and cannot be themed. Rounded corners are the single most anti-80s
element in the current chrome. lipgloss ships `DoubleBorder()` (╔═╗, the DOS/ANSI-art idiom),
`ThickBorder()`, `BlockBorder()` and `OuterHalfBlockBorder()` (chunky bevelled slab). Making border
style a palette token is cheap because `paneHFrame`/`paneVFrame` are already DERIVED from the style
rather than hardcoded — though the comment at `style.go:57` asserting borders are theme-independent
would become stale.

**3. Modal drop shadow.** The classic DOS dialog shadow. `overlay()` (`help.go:349`) already composes
Z-ordered layers via `lipgloss.NewCompositor`, so a shadow is one more layer offset +1/+1 at lower Z.

**4. Display type / wordmark.** 1b uses Archivo Black, 1c Michroma. The terminal analog is block-glyph
lettering built from ▀▄█ for a splash, atlas header, or `theme preview`. `progressbar.Render` already
interpolates a multi-stop gradient per cell via `progress.WithColors`, so a chrome-type colour sweep
across a wordmark reuses machinery that exists.

**5. Half-block gradient bands.** 1b's header is a 4-stop vertical gradient. ▀/▄ give two vertical
colour rows per cell row, so a 3-row band yields ~6 steps. `Palette.Gradient` carries 3 stops today
(purple→cyan→pink); a sunset band would want its own stop set.

**6. Texture that encodes state.** 1c hatches the blocked card with diagonal stripes. A ░▒▓ row fill
is the terminal cousin, and it is the one texture idea that earns its cost because it carries state
rather than decoration — consistent with the principle stated in `theme.Bucket`'s comment that a
state must survive a mono terminal.

**7. Authentically 80s, probably still bad.** `lipgloss.Style.Blink()` exists and is period-correct,
hostile, and inconsistently supported. Per-pixel scanlines are impossible; at one-row granularity the
effect is just zebra striping (see 1). Recorded here so neither is re-litigated.

**Framing.** The cleanest home for 1b and 1c is as THEMES, not redesigns — siblings of `neon` and
`catppuccin` in the `design` registry, plugging into the existing `theme list` / `theme preview` /
picker / `--theme` / `TSKFLW_THEME` / config selection story. 1a is approximately the current theme
already.

## Guardrails measured, not assumed

- `TestChromeSurfaceContrastAA` loops `Names()`, so ANY new theme is automatically held to AA for the
  five text roles over `Surface`. Measured against their own backgrounds: 1b miami is nearly AA-clean
  as drawn — only `#b026ff` (the ready-to-start accent) misses at 4.07:1. 1c vector arcade fails
  exactly where its look lives: tinting each column's metadata to its column hue gives `#6a4a86`
  2.94:1 (fail), `#3d6a70` 3.49:1 and `#8a4a72` 3.30:1 (large-text only).
- That test covers TEXT over `Surface`, not borders or backgrounds. The existing idle border is 2.21:1
  on base and is fine because it is a border — so new border/fill tokens are a judgment call the suite
  will not catch.
- Every `Hue` carries an explicit ANSI slot for the 16-color path. A fill-heavy theme degrades far
  worse there than a foreground-heavy one, since backgrounds collapse to 8 clashing colors.
- `--color=never` and mono terminals: state encoded in a background violates the repo's stated
  shape-survives-color principle unless it is paired with a glyph.
- Every theme is a Dark/Light pair and both existing themes share `latteAA`. An 80s theme has no
  honest light variant; reusing `latteAA` is likely right but should be said out loud rather than a
  pastel Miami invented.

## Design attention

Deliberately undecided:

- Whether the board becomes the Overview, a new tab beside it, or a mode of the tasks tab — and
  whether the rail is part of it or stays the Overview's job.
- Whether the dashboard/`status` parity gaps are one task or three, and whether parity is even the
  goal (the TUI may be right to be terser).
- Whether `Board.Blocked` widens to carry the gating edge, and whether that is a projection change or
  a renderer lookup against the existing graph.
- How far `design.Palette` grows for fill and border tokens, and whether a theme may change chrome
  STRUCTURE (border glyphs, selection mechanism) or only colour.
- Whether 1b/1c land as new registry themes, as variants of `neon`, or not at all.
- Two-line cards versus the current one-line rows.

## Out of scope

- Any web UI, `tskflwctl serve`, or the marketing/landing strip in the canvas (1d) beyond noting that
  its board blurb — "Blocked work sinks to the bottom of its column so the top of the list is always
  startable" — is a crisp statement of the invariant `sortEligibleFirst` already implements and is
  good help/docs text.
- Changing `theme.Status`/`theme.Bucket` glyph or colour semantics.
- Thread graph and spatial-view presentation.

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Data-organization takes likely spin out onto
  [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md).
- Source canvas: `~/Downloads/Taskflow Web Interface Mockup/Taskflow Web.dc.html` (turn 1: options
  1a/1b/1c/1d), synced 2026-09-03 per its `github.md`.
