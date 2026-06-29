---
schema: 1
status: in-progress
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
started_at: "2026-06-29"
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

**Slice A done 2026-06-29 (branch feat/theme-discovery).** Themes are now real + discoverable: registered a second theme **catppuccin** (Catppuccin Mocha dark / shared AA-Latte light — the dark variant + accent diverge from neon's synthwave), added `design.Names()` (sorted), and a `theme list` command (human + --json via a new wire ThemesEnvelope) that works anywhere and marks default+active. `--theme catppuccin` now renders visibly different (mauve #cba6f7 vs neon magenta). Tests: Names, Catppuccin Mocha slots, themeEntries, + regenerated docs/cli and schema_comments.json. gofmt/vet/full test green. REMAINING (slice B): `theme preview`, wire Theme.Markdown into glamour, and the deferred picker polish (slug accent / description dim).

**Slice B done 2026-06-29 (branch feat/theme-discovery).** (1) **Glamour Theme.Markdown wired** — the CLI `markdownStyle()` and the TUI both read the active theme's `palette.Markdown` (the `theme.MarkdownStyleFor` call sites are gone). catppuccin ships `tokyo-night`, neon `dracula`, so `--theme X task show` now renders the markdown body in the theme's own style (verified: distinct Dracula vs Tokyo-Night escape sets). (2) **`theme preview [name]`** — a swatch grid (truecolor block + hex per token + a sample gradient bar) for the background-appropriate variant; byte-stable plain (token->hex + bar glyphs) when piped/--color=never; `--json` emits a swatch envelope (token/hex/ansi); an unknown name errors with the available list (exit 10). Tests added (preview plain + JSON, catppuccin markdown pin); regenerated docs/cli + schema_comments.

DEFERRED: the picker two-tone polish (slug accent / description dim) is split to its own task [[two-tone-picker-rows-slug-in-accent-description-dim]] — it's fragile (huh re-styles each row over the label) and wants visual iteration. With Markdown + preview + second theme + `theme list` all done, T6's substance is complete; it closes on merge.

**Adversarial review addressed (2026-06-29, 3 lenses).** Verdict: integrates cleanly; byte-stability airtight; Mocha values canonical + AA-clean on the dark base; neon + the theme-independent glyphs untouched. Fixed: (1) **[blocker]** the new `theme list/preview --json` envelopes were missing from the `schema --json-schema` registry (jsonEnvelopes) — registered ThemesEnvelope + ThemePreviewEnvelope, added round-trip cases + ThemeSwatch field descriptions, regenerated the schema golden + schema_comments; (2) `theme preview --json` fired an OSC-11 terminal query making `variant` terminal-dependent on the machine path — gated the query on `!--json` (deterministic dark); (3) catppuccin gradient collapsed on 16-color (mauve+pink both -> magenta) — swapped the pink endpoint for teal #94e2d5 so it degrades to 3 distinct slots (13/4/6); (4) deleted the now-dead `theme.MarkdownStyleFor` + its test (both callers read palette.Markdown). Refuted/kept as defensible: shared latteAA light, truecolor-in-render (same on/off+strip contract as Bar), hardcoded canonical Mocha hexes. gofmt/vet/full test green.
