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
updated_at: "2026-06-29"
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

**Review notes for T6 (from the T5 adversarial review, 2026-06-29):**
- **Export a registry enumerator in `design`** for `theme list` / `theme preview` (no-arg): `registry` is unexported and only `Default()`/`Lookup(name)` exist — both resolve a single theme. Add e.g. `func Names() []string` (SORTED, for `--json` byte-stability) or `All() []Theme`. With it, T6 is a pure `cli` change; T6 edits `design` anyway (it registers the second theme), so land it there.
- **Wire `Theme.Markdown` (currently dead).** The field exists in `design.Palette` (set to `theme.MarkdownStyleDark/Light`) but NOTHING reads it: both the CLI (`root.markdownStyle`) and the TUI (`tui.Run`) still call `theme.MarkdownStyleFor` directly. "Palette-select the glamour markdown style" = read the resolved theme's `Markdown` instead, so a config-selected theme themes the body too.
- **`theme preview` must not stomp the global `pal`.** Rendering a sample dashboard/bar in a theme via the TUI path calls `applyTheme`, mutating the process-global `pal` (T2). Either render the sample without mutating `pal`, or save/restore it. `tui.Run(..., th)` is the threading seam — coordinate with [[tui-palette-as-model-scoped-styles-instead-of-package-globals]].
- **`name = "none"` (monochrome)** is still unimplemented; T5 added a stderr warning for it. If T6 wants it, register a `mono`/no-hue theme.
