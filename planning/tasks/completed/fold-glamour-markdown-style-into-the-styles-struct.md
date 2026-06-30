---
schema: 1
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: detail pane's glamStyle is threaded separately (field assignment in Run, a 2nd th.For call) — the one theme element bypassing the shared styles struct; fold it in so a retheme touches one place
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [design, tui, refactor]
created: "2026-06-29"
updated_at: "2026-06-30"
started_at: "2026-06-30"
completed_at: "2026-06-30"
---

# fold glamour markdown style into the styles struct

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [[25-design-system-coherent-palette-and-selectable-themes]]

**Done 2026-06-30.** Added a `markdown string` field to the `styles` struct, set in `newStyles` from `p.Markdown`. The detail pane now reads its glamour style from `d.st.markdown` (dropped its own `glamStyle` field + the `newDetailPane(st, glamStyle)` param). `tui.Run` no longer sets `m.detail.glamStyle` separately — the existing `*m.st = newStyles(th.For(dark))` now carries the markdown style too, so a theme/background swap touches one place. Glamour tests reworked: the style-rebuild test mutates a local `newStyles(...)` bundle (isolated from shared testStyles). gofmt/vet/full test green.
