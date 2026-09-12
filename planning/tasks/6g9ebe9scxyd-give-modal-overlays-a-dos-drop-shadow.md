---
schema: 1
id: 6g9ebe9scxyd
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Add a one-cell offset shade under floating modals via the existing Z-layer compositor, and settle the size budget it costs.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, design, chrome, overlay]
created: "2026-09-12"
---
# Give modal overlays a DOS drop shadow

## Objective

The floating overlays (`?` help, action menu, follow picker, inline editor, command palette) composite
over the body with no separation beyond their border. The period-correct fix is the DOS dialog drop
shadow — a one-cell offset shade down and right — which also genuinely helps the box read as floating
rather than spliced into the rows behind it.

## Acceptance criteria

- [ ] Every modal in the registry renders with the shadow; none needs its own opt-in.
- [ ] The shadow never clips at the canvas edge and never pushes the box off-centre.
- [ ] The modal's usable inner size is unchanged, or the change is accounted for everywhere that sizes
  modal content.
- [ ] The effect degrades sanely with colour disabled.
- [ ] Whether the shadow is theme-owned or universal chrome is decided.

## Design notes

**The machinery is already there.** `overlay()` (`internal/tui/help.go:349`) composes
`lipgloss.NewLayer`s at explicit X/Y/Z through `lipgloss.NewCompositor`, so the shadow is one more
layer at X+1/Y+1 beneath the box's Z rather than any new rendering approach.

**The real constraint is the size budget, not the drawing.** `bodyView` hands each modal
`m.width-2, m.paneOuterH-2` and `overlay` centres it on a `width × height` canvas. A +1/+1 shadow needs
one more cell of room at the right and bottom or it is clipped by the canvas — so either the modal
budget shrinks by one in each axis, or the centring shifts to reserve it. Whichever is chosen has to
reach `helpMaxScroll`, which mirrors `helpBox`'s window math and would otherwise silently disagree
about how many lines fit.

**Open:** the shade itself — a `░` block in the idle border tone, versus re-rendering the covered body
cells dimmed (truer to the era's "darken what is behind" but more work and more fragile over already-
styled content).

## Out of scope

- Shadows on panes, list rows, or anything that is not a floating modal.
- Background/fill tokens for the panes themselves.

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Context: [mine the web mockup directions for the TUI](6g95m4eyvf4k-mine-the-web-mockup-directions-for-the-tui-data-organization-and-a-more-80s.md), axis B item 3.
