---
schema: 1
id: 6g544nq49edg
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Settle whether [theme] stays a registered-name selector or gains user-supplied color, and what guards runtime palettes
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [design, palette]
created: "2026-08-30"
updated_at: "2026-08-30"
---
# Decide how far user-controlled color goes beyond named themes

## Objective

`[theme]` accepts a **registered name only** — `themeFileTOML` is a single `name` field,
and the generated config comment reads "a registered name, or `auto` for the default".
There is no way to change a color without editing `internal/design` and rebuilding. Two
themes ship: `neon` and `catppuccin`.

Settle whether that is the intended ceiling or a gap, and if it is a gap, pick the shape.
**The decision comes first — this is a scope question, not an implementation.**

## Why it is a real question rather than an obvious yes

Three deliberate decisions cut against free-form user color, and all three are recorded:

- Epic 25's goal is "established, **selectable** themes (base16, neon default)". Selectable,
  not authorable.
- The research settled on *"lean on the `base16` standard, not a custom system"* precisely
  so degradation is principled.
- Every `Hue` carries an explicit ANSI anchor *"so a neon hue degrades to a CHOSEN color
  rather than a runtime nearest-color guess"*. Hand-written hexes have no anchor unless
  the author supplies one.

So "fully configurable" may be a feature nobody decided to build rather than an adherence
gap. That distinction is the point of this task.

## What a complete palette actually costs

`design.Palette` is 6 semantic slots (red, green, yellow, blue, cyan, gray), 11 chrome
tokens (accent, border-active, border-idle, danger, heading, match, match-current,
match-fg, track, base, surface), 3 gradient stops, and a glamour style name. Both
variants of one theme is **40 hex + ANSI pairs plus 2 style names**. That is a lot of TOML
for a preference, and most of it is not what someone asking for this actually wants.

## The guards that exist only at test time

This matters more than the ergonomics. The built-in palettes are held to invariants that
live in `internal/design/design_test.go` and run at build time, not load time:

- `TestFindHighlightContrastAA` — `MatchFg` clears AA over both match backgrounds.
- `TestChromeSurfaceContrastAA` — every colored role fang draws in the help codeblock
  clears AA over `Surface`, and `Surface` stays on the base's side of the light/dark divide.
- `TestChromeSurfaceSlots`, `TestNeonDarkSemanticSlots`, `TestNeonAccent` — value pinning.

A runtime palette bypasses all of them. The unreadable-help defect fixed on 2026-08-29
(`route-fang-chrome-through-the-resolved-theme-and-fix-codeblock-contrast`) — a foreground
hue assigned to a background role, rendering `[command]` at 1.00:1 — becomes something a
user can reproduce in their own config with nothing to catch it. Any option below must say
what validates a palette at load time and what happens when it fails.

## Options

**A. Do nothing; document the ceiling.** Change the config comment from "a registered name"
to an explicit statement that palettes are built in, and close this. Cheapest, and honest
if the epic's goal is taken at face value. Everything below is strictly more work.

**B. Token overrides on a registered base** — `[theme] name = "neon"` plus a small
`[theme.overrides]` table setting individual tokens. Matches the actual request behind
this ("I like neon but its yellow is wrong"), keeps every unspecified token anchored to a
reviewed palette, and keeps the ANSI anchors for everything not overridden. Validation is
tractable because the result is a known palette with a few substitutions.

**C. Adopt base16 scheme files.** `neonDark` *is* a port of base16 Synth Midnight Terminal
Dark, and the mapping is consistent and already documented in the source: `base00`→Base,
`base01`→Surface, `base03`→BorderIdle/Track, `base04`→gray, `base0A`→yellow,
`base0B`→green, `base0C`→cyan, `base0D`→blue, `base0E`→accent/heading/border-active. Extract
that mapping into a function and any of the hundreds of published base16 schemes becomes a
palette mechanically — no per-token TOML, ANSI anchors preserved slot-for-slot, and fully
in the spirit of the standard the project already chose.

Not free: base16 covers 16 slots and our palette needs more. `Danger` is the known
exception (the scheme's `base08` failed contrast, hence the Outrun `#FF4242` swap), and
`Gradient` and `Markdown` have no base16 equivalent at all. Those need defaults or a small
per-theme supplement, and the AA invariants must run at load time rather than in a test.

**Recommendation:** decide between A and C. B is the smallest useful step but adds a
config surface that C would likely supersede. C is the only option that scales the theme
count without scaling the config surface, and it is the one the epic's own stated principle
points at.

## Acceptance criteria

- [ ] The decision is recorded — in this task, or as an ADR if it changes the epic's stated goal.
- [ ] If A: the `[theme]` config comment and `schema` output state that palettes are built in and not user-definable, and this task closes.
- [ ] If B or C: a user-supplied palette is validated at load time against the same contrast invariants the built-ins are held to in `design_test.go`, and the shared check is exercised from both places rather than duplicated.
- [ ] If B or C: an invalid palette degrades to a warning and the default (matching `warnUnknownTheme`'s existing behavior for an unrecognized name) rather than failing a command or rendering something unreadable.
- [ ] If B or C: `theme preview` renders the user palette so it can be reviewed before it is relied on.
- [ ] If B or C: `--json`, non-TTY, and completion paths are unaffected — no palette parsing on the machine path.

## Out of scope

- Per-command or per-entity color overrides.
- Runtime theme switching inside the TUI.

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Research [color-palette-and-theming-overhaul](../research/6fgq1n001pwm-color-palette-and-theming-overhaul.md) — the base16 decision this would extend or contradict.
- Raised while reviewing `route-fang-chrome-through-the-resolved-theme-and-fix-codeblock-contrast`, which supplied the load-time-validation argument.
