---
schema: 1
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: 'Design: map v2 capabilities (overlays/clipboard/cursor/keyboard) to specific TUI needs and propose build tasks; no code. After the v2 migration.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, ux, design]
created: "2026-06-23"
updated_at: "2026-06-23"
started_at: "2026-06-23"
---
# Design the v2 TUI feature wins (overlays / clipboard / cursor / keyboard)

The "needs-scoping" bucket from the v2 plan: v2 unlocks capabilities that should
be connected to specific TUI needs before building. This is a DESIGN task —
weigh each capability against a real need and propose concrete build tasks; do
not build here.

## Capabilities → candidate needs (to weigh)
- **Layers / overlays** (`View.Layer`): a **command palette** (augment the `:`
  jump), **confirm-before-mutate modals** (task start/complete/deprecate), a
  **floating help/keymap** — instead of full-screen context swaps. Connects to
  the existing quit-key context layering + the mutation actions.
- **Clipboard (OSC52, works over SSH):** yank a task slug / epic id to paste into
  `task start …`. Decide the keybinding.
- **Advanced cursor control** (shape/blink/position): sharpen the inline `e`
  field editor.
- **Enhanced keyboard** (shift+enter, ctrl+alt combos, key-release): richer,
  unambiguous bindings.

## Acceptance criteria
- [ ] A decision per capability (adopt/decline) naming the specific need it serves.
- [ ] Build tasks proposed for whatever's greenlit (each its own slice).
- [ ] No implementation here.

## Depends on
- The v2 migration.

## Related
- Plan: `planning/research/2026-06-23-tui-v2-migration-plan.md`.
- [[18-tui-bubble-tea-interactive-planning-browser]].
