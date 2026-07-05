---
schema: 1
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: Colors are scattered across theme, progressbar (neon hex), the picker (hardcoded purple), TUI lipgloss, huh, and glamour with no single palette. Define one palette and route every surface through it.
effort: L
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, tui]
created: "2026-06-25"
updated_at: "2026-06-29"
blocked_by: [theme-discovery-commands-glamour-polish-and-a-second-theme]
completed_at: "2026-06-29"
id: 6ffr4wc01gtv
---
# Color/design overhaul: one coherent palette across every surface

## Objective

The project's colors are spread across many uncoordinated sources — there is no
single palette. `internal/theme` is the *semantic* source of truth (status/bucket/
finding glyphs + a 16-color enum), but several surfaces bypass it with their own
literals: the progress bars use a neon hex gradient, the new interactive picker
hardcodes a neon purple, the TUI has ad-hoc lipgloss styles, prompts use huh's
Dracula, and markdown bodies use glamour's dracula/light. Result: "the colors are
all over the place." Establish ONE documented palette and route every surface
through it. (Raised during the 2026-06-25 picker work, where the neon-purple caret
was a deliberately-local stopgap pending this.)

## Context — the current color sources (the map)

- `internal/theme/theme.go` — semantic `Color` enum (Red/Green/Yellow/Blue/Cyan/
  Gray) + glyph tokens for task status / audit bucket / finding status / priority /
  percent. The intended single source of truth, but it only covers semantic
  glyph+color — not accent / dim / selection / border.
- `internal/progressbar/progressbar.go` — a neon hex gradient (`#b026ff` purple,
  `#00e5ff` cyan, `#ff2ec4` pink) for epic bars + segment colors (`"2"/"3"/"8"`)
  for the audit segmented bar — independent of theme.
- `internal/cli/prompt/tty.go` `pickerTheme` — hardcodes `#b026ff` for the picker
  caret + current row (a symptom of the gap; explicitly flagged as local-for-now).
- `internal/cli/render/style.go` — ANSI 16-color SGR codes mapped from `theme.Color`
  (the CLI's color tech).
- `internal/tui/style.go` — ad-hoc lipgloss styles (dim, selected, `fg(theme.Color)`)
  + miniBar / segBar.
- huh `ThemeDracula` (prompts) and glamour `dracula`/`light` (markdown bodies) —
  third-party palettes not coordinated with ours.

## Acceptance criteria

- [ ] A **documented palette**: a primary **accent** (the neon purple?), the
      **semantic** colors (status/bucket/priority/percent — already in theme), and
      consistent **dim / selected / highlight / border** tokens.
- [ ] `internal/theme` (or a small design package) is the single source for ALL
      color decisions — not just glyphs. The accent, the bar gradient/segments, and
      the picker/selection styling derive from it.
- [ ] The picker (`pickerTheme`), the progress bars (`progressbar`), the TUI lipgloss
      styles, and the CLI render styles all consume the shared palette — no stray
      hex/ANSI literals outside the palette definition.
- [ ] 16-color graceful degradation preserved (theme's 16-color ANSI on the CLI;
      bubbletea v2 colorprofile auto-downsampling on the TUI/prompts).
- [ ] The deferred picker polish lands here, done against the real palette: slug in
      the accent, description dim (the "slug one color / description another" idea;
      two-line optional).

## Implementation sketch

- Inventory every color literal: `grep` for hex `#`, `lipgloss.Color`, the `ansi*`
  SGR consts, and huh/glamour theme calls.
- Define palette tokens in `theme` (accent, dim, selected, border + the existing
  semantic set); decide whether the bar gradient is part of the palette or a
  deliberate exception.
- Refactor each surface to reference the tokens; align the huh prompt theme + the
  glamour style to the palette where feasible.

## Risks / gotchas

- Many colored outputs are ANSI-stripped in goldens (porcelain is plain), so golden
  churn may be small — but the theme decision-table tests + any colored snapshots
  move.
- CLI is 16-color ANSI; TUI/prompts are truecolor lipgloss — the palette must
  express both (a token → ANSI-16 on the CLI, truecolor on the TUI). The theme's
  existing `Color`→tech mapping already models this; EXTEND that pattern, don't fork.
- Don't regress the symbology work (status glyphs/colors, the segmented bar) — those
  are pinned by theme tests.

## Related

- The picker stopgap: [render-the-interactive-picker-inline-not-full-screen-alt-screen](6ffr4wc00dsr-render-the-interactive-picker-inline-not-full-screen-alt-screen.md)
- The symbology/segmented-bar work that established the current `theme` decisions.

## Done when

One documented palette; every surface (theme/render/tui/progressbar/prompt + huh/
glamour) is consistent with it; and no stray color literals live outside the palette
definition.

**Progress 2026-06-28.** A chunk landed via the audit`s theme/glyph work (epic 21): the SEMANTIC tokens (status/bucket/liveness/finding) AND the cross-surface markers (⚠/↻/✓/✔/!) are now centralized in internal/theme as glyph+color tokens (theme.Status/Bucket/Liveness/FindingStatus/Marker*), with a glyph() helper, so the CLI render layer and the TUI draw the same decisions. STILL OPEN (the actual "one palette" overhaul): the chrome/structural colors remain scattered + hardcoded — progressbar`s neon hex, the picker`s purple, huh/glamour themes, and UI-chrome lipgloss colors (dashHeading, accent, helpHeading, actionHeading all bypass theme). Define one palette and route those through it.

**Research 2026-06-28.** Full audit + unification design: `planning/research/2026-06-28-color-palette-and-theming-overhaul.md`. Recommendation: adopt the **base16** standard (degrades slot==slot to the CLI's 16 ANSI colors); default neon theme `neon-night` ported from base16 *Synth Midnight Terminal Dark* (danger-red → Outrun `#FF4242` for contrast), light fallback `neon-day` ≈ Catppuccin Latte (`github.com/catppuccin/go` already in the graph). New `internal/design` package owns a truecolor `Palette`+`Theme` registry (`theme` stays domain-only); each token carries an explicit ANSI-16 anchor. `[theme]` config table mirrors `[pager]`. Scope locked to COLOR (glyphs/borders fixed). To be split into a sequenced task chain — see research §7.
