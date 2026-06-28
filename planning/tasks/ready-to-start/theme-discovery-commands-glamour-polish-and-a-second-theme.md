---
schema: 1
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Add theme list/preview; palette-select the glamour markdown style; land the deferred picker polish (slug accent/description dim); register a second named theme.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, tui, design]
created: "2026-06-28"
updated_at: "2026-06-28"
blocked_by: [theme-config-table-and-selection-plumbing]
---
## Objective
Make themes discoverable and prove the registry with a second theme + the deferred polish.

## Scope
- In: `tskflwctl theme list` (+ `--json`) and `theme preview [name]` (sample dashboard/bar/segmented-bar in the theme); palette-select the glamour markdown style name (`MarkdownDark/Light`); land the deferred picker polish (slug in `Accent`, description in `Dim`) on the real palette; register one more named theme (e.g. Catppuccin or Outrun) exercising the light-module path.
- Out: full glamour `ansi.StyleConfig` authoring (optional follow-up).

## Done when
`theme list`/`theme preview` work; a second theme is registered; the picker polish lands; `build/test/lint` green; CLI docs regenerated for the new commands (docs-check gate).

## Reference
Design doc §5, open questions. Depends on the config/selection task. Closes the picker stopgap in [[render-the-interactive-picker-inline-not-full-screen-alt-screen]].
