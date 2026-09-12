---
schema: 1
id: 6g9eba1j545t
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: 'Register miami-vice: hot pink and neon cyan on indigo, retinting only the freely-assignable blue and cyan slots, plus the find-highlight AA coverage gap a third theme exposes.'
effort: M
tier: 2
priority: high
autonomy_level: 3
tags: [design, theme, palette, a11y]
created: "2026-09-12"
updated_at: "2026-09-12"
started_at: "2026-09-12"
completed_at: "2026-09-12"
---
# Add the miami-vice theme

## Objective

Add `miami-vice` as a third registered theme — hot pink and neon cyan on indigo, taken from the 1b
mockup direction. The palette below is already curated and contrast-checked; this task is the
implementation, the tests, and the two coverage gaps a third theme exposes.

## Acceptance criteria

- [x] `miami-vice` is registered and selectable through every existing path: `--theme`, `TSKFLW_THEME`,
  the repo and user config tables, `theme list`, `theme preview`, and the TUI.
- [x] Its dark palette is the curated set below, with each semantic slot carrying an explicit ANSI slot
  for the 16-colour path.
- [x] Its semantic slots are pinned by a test in the same shape as `TestNeonDarkSemanticSlots`, so a
  later edit is a reviewed decision.
- [x] `TestFindHighlightContrastAA` covers every registered theme rather than only `Default()`.
- [x] The light variant is a deliberate, documented choice.
- [x] `theme preview miami-vice --variant dark` renders swatches and a bar that read correctly on a
  truecolour terminal.

## The curated palette (verified)

Ratios are against `Base #160b2e`; the AA-gated roles are re-checked against `Surface` in the table
below that.

```
Semantic                                        on Base
  ColorNone   {"",        NoANSI}
  ColorRed    {"#FF4242", 1}    5.45:1   the Outrun red, the same legible swap neon makes
  ColorGreen  {"#06ea61", 2}   11.55:1
  ColorYellow {"#c9d364", 3}   11.58:1   UNCHANGED — see the slot note below
  ColorBlue   {"#00e5ff", 4}   12.17:1   next-up; the mockup's neon cyan
  ColorCyan   {"#c88cff", 6}    7.71:1   ready-to-start; the mockup's own label colour
  ColorGray   {"#8b7bb5", 8}    4.99:1

Chrome
  Accent       {"#ff2ec4", 13}  5.74:1   hot pink — see why this matters below
  BorderActive {"#ff2ec4", 13}
  BorderIdle   {"#4a3a6b", 8}   1.88:1   chrome only (cf. neon's own idle border at 2.21:1)
  Danger       {"#FF4242", 1}
  Heading      {"#ff2ec4", 13}
  Match        {"#ffd166", 3}   with MatchFg 12.99:1
  MatchCurrent {"#ff2ec4", 13}  with MatchFg  5.74:1
  MatchFg      {"#160b2e", 0}
  Track        {"#3a2b57", 8}
  Base         {"#160b2e", 0}   the mockup's body fill
  Surface      {"#0a041a", 0}   RECESSED — the mockup's own sidebar fill

Gradient  #b026ff -> #00e5ff -> #ff2ec4    (the mockup's own bar gradient)
Markdown  dracula
```

`TestChromeSurfaceContrastAA` gates five text roles over `Surface` plus a lift and a light/dark-side
check. Measured for `Surface #0a041a`: blue 13.06:1 · yellow 12.42:1 · green 12.39:1 · cyan 8.28:1 ·
gray 5.36:1, lift 1.07 (needs ≥1.03), same side of the divide. It passes with headroom, and it is
recessed for the same reason Mocha's is — layering away from the base widens every pair instead of
narrowing it.

## Design notes

**Why yellow is untouched.** `ColorYellow` carries in-progress AND `⚠` warn AND `↻` revisit AND the
audit-open `◆` AND priority-medium AND the mid `Percent` band AND `BandActive`. Retinting it hot pink
to match the mockup's in-progress column would turn every warning pink and weaken warn-vs-error against
the red. `ColorCyan` by contrast is used in exactly ONE place (`theme.Status` ready-to-start) and
`ColorBlue` in three, so both retint freely. The pipeline therefore reads cyan → purple → yellow:
recognisably Miami at the front, unchanged at the back. Decided; not an oversight.

**Hot pink lands on `Accent`, and that is the point.** Because yellow keeps the warning vocabulary,
`#ff2ec4` stays scarce — borders, headings, selection — which is exactly where the mockup's pink
carries its charge. Retinting yellow would have made pink omnipresent and killed it as an accent.

**`#b026ff` is deliberately not used for text.** The mockup's purple dot/border is 4.07:1 on its own
background and fails AA for small text. `#c88cff` — the mockup's own READY TO START *label* colour at
7.71:1 — is the semantic slot; `#b026ff` survives as the gradient's first stop, where it is decoration
rather than text.

**Keep the conventional ANSI slots (blue 4, cyan 6) even though the hues are cyan and purple.**
`theme --help` documents that on a 16-colour terminal the semantic colours fall back to the terminal's
own palette and "look the same across themes"; honouring the slot convention is what preserves that.

**Light variant:** reuse `latteAA`, as both existing themes do. An 80s palette has no honest light
form, and inventing a pastel Miami would be worse than sharing the AA-tuned Latte. Say so in a comment
rather than leaving it to look accidental.

**Coverage gap this exposes:** `TestFindHighlightContrastAA` loops only `Default()`, unlike
`TestChromeSurfaceContrastAA` which loops `Names()`. A third theme's `Match`/`MatchCurrent`/`MatchFg`
triple would be entirely unchecked. Broaden it.

## Out of scope

- Border idiom and any other chrome STRUCTURE (see the border-token task).
- Background/fill tokens beyond the existing single `Surface`.
- Changing `theme.Status` glyph or slot assignments.

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Chrome half depends on [the theme-owned border idiom](6g9eb5g6gpyf-make-the-border-idiom-a-theme-owned-token-and-retire-rounded-corners.md).
- Context: [mine the web mockup directions for the TUI](6g95m4eyvf4k-mine-the-web-mockup-directions-for-the-tui-data-organization-and-a-more-80s.md), direction 1b.
