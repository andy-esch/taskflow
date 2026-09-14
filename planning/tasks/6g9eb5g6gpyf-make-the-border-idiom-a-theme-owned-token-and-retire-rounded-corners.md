---
schema: 1
id: 6g9eb5g6gpyf
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Move the hardcoded RoundedBorder out of newStyles onto the theme so chrome structure is themeable, and switch neon to the double-line DOS idiom.
effort: S
tier: 2
priority: high
autonomy_level: 3
tags: [tui, design, theme, chrome]
created: "2026-09-12"
audited: "2026-09-13"
updated_at: "2026-09-13"
---
# Make the border idiom a theme-owned token and retire rounded corners

## Objective

`RoundedBorder()` is hardcoded in six places in `newStyles` (`internal/tui/style.go:69,70,78,86,87`
plus the `NormalBorder` at `:93`), so no theme can change the chrome's structure — only its colour.
Rounded corners are also the single most anti-80s element in the current look. Make the border idiom a
property of the theme, and switch `neon` to `DoubleBorder` (╔═╗, the IBM PC / ANSI-art dialog idiom).
This is the enabler the other chrome work sits on.

## Acceptance criteria

- [ ] The border idiom is declared once per theme and reaches every chrome box without a per-call
  literal.
- [ ] `neon` renders its panes, help box, action menu and danger box in the double-line idiom.
- [ ] Swapping a theme swaps the idiom with it, and no pane, overlay, or list is mis-sized by the swap
  at any terminal width.
- [ ] `catppuccin`'s idiom is a deliberate choice recorded in the palette, not an accident of the
  default.
- [ ] Whether `theme preview` and its `--json` envelope expose the idiom is decided, and the envelope's
  `schema_version` is bumped if it does.
- [ ] The frame-size comment at `style.go:55-57` no longer claims border styles are theme-independent.

## Design notes

**It must live on `design.Theme`, not `design.Palette`.** `latteAA` is shared by BOTH `neon` and
`catppuccin` as their light variant, so a `Palette` field could not carry two different idioms. Put it
on `Theme` and have `Theme.For()` — which already returns a `Palette` by value — stamp it into the
returned copy. That keeps `newStyles(design.Palette)`'s signature and all of its call sites
(`tui.go:36,41`, `model.go:174,176`) untouched.

**Prefer a named idiom type in `design` over a raw `lipgloss.Border`,** so it is printable and
serialisable for `theme preview`/`--json` and stays declarative. `design/palette.go` already imports
lipgloss for `lipgloss.Color`, so this adds no dependency and does not disturb the
design→theme direction.

**Sizing is safe.** All nine lipgloss borders (`Rounded`, `Normal`, `Thick`, `Double`, `Block`,
`OuterHalfBlock`, `InnerHalfBlock`, `ASCII`, `Markdown`) measure `hframe=2, vframe=2`. `paneHFrame`,
`paneVFrame` and `helpHFrame` are already DERIVED from their styles rather than hardcoded, so they stay
correct across a swap and `recomputeLayout`/`helpMaxScroll` need no change. Verify rather than assume
if a non-box idiom (`HiddenBorder`) is ever offered.

**Open:** whether `editAreaBox` (`:93`, today `NormalBorder` — deliberately lighter than the panes)
follows the theme or stays fixed as a contrast device.

## Out of scope

- Background/fill tokens, selection restyling, and any new palette colour.
- Letting a theme change padding or pane geometry — this is the border glyph set only.

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Enabler for [the miami-vice theme](6g95m4eyvf4k-mine-the-web-mockup-directions-for-the-tui-data-organization-and-a-more-80s.md)'s chrome half.
- Context: [mine the web mockup directions for the TUI](6g95m4eyvf4k-mine-the-web-mockup-directions-for-the-tui-data-organization-and-a-more-80s.md), axis B item 2.

## Sweep audit 2026-09-13

Automated weekly sweep. Every code citation re-checked against `main` (`1ca31b9`).

**Verified accurate (2026-09-13):** `RoundedBorder()` at `internal/tui/style.go:69,70,78,86,87`
and `NormalBorder()` at `:93` — all six still literal, still unthemeable. The
"border STYLES are theme-independent" claim is still at `style.go:57`. `newStyles`
call sites in `tui.go:36,41` are exact. `Theme.For()` still returns a `Palette` by
value (`design/palette.go:100`), and `design/palette.go` still imports lipgloss (`:19`),
so the design note's chosen approach still holds.

**Drifted — `latteAA` is now shared by THREE themes, not two.** `miami-vice` landed
2026-09-12 (`c037150`) and also pairs with `latteAA` (`design/theme.go:209`, alongside
`neon` at `:201` and `catppuccin` at `:204`). This *strengthens* the design note's
argument for putting the idiom on `design.Theme` rather than `design.Palette` — the
shared-light-palette collision is now three-way — but the note's "BOTH `neon` and
`catppuccin`" phrasing is now undercounting.

**Drifted — minor.** The `newStyles` call sites cited as `model.go:174,176` are now at
`model.go:175,177`.

**Left for a human:** acceptance criterion 4 names only `catppuccin`'s idiom as a
deliberate choice. `miami-vice` now needs the same decision, and it is the theme whose
whole point is the 80s idiom this task introduces. Widening a criterion is a scope
change, so this sweep flagged it rather than editing it.

## Progress log

- 2026-09-13: automated weekly sweep — all six `RoundedBorder`/`NormalBorder` sites and the `style.go:57` comment re-verified; `latteAA` is now shared by three themes (`miami-vice` landed 2026-09-12), so the "put it on `Theme`, not `Palette`" argument strengthens and AC 4 undercounts.
