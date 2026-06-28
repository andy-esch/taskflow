---
schema: 1
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Convert the global lipgloss chrome vars (accent, pane/help/action/edit borders, dashHeading, find highlights) into palette-built styles; lipColor -> palette truecolor.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, design]
created: "2026-06-28"
blocked_by: [design-package-foundation-palette-theme-registry-and-the-neon-default]
updated_at: "2026-06-28"
---
## Objective
Make the TUI's structural chrome derive from the palette instead of scattered lipgloss literals.

## Scope
- In: convert the global style vars into a `styles` struct built from a `Palette` in `New` — `tui/style.go` (accent, paneActive/Inactive, selected/dim/activeTab), `dashboard.go` dashHeading, `help.go` border/heading, `action.go` action/danger border+heading, `find.go` match/current, `edit.go` edit-box border. Reimplement `lipColor` as `pal.Of(c).True`. Thread `design.Default().For(dark)` for now (config selection is a later task).
- Out: bars/CLI ANSI, picker, config wiring. Glyphs + border *style* unchanged.

## Done when
No stray chrome `lipgloss.Color(...)` literals in the TUI; styles built from the palette; `build/test/lint` green; TUI substring tests pass unchanged.

## Reference
Design doc §1,§5. Depends on the design foundation. Parallel with the bars task.
