---
schema: 1
status: completed
epic: 25-design-system-coherent-palette-and-selectable-themes
description: '[theme] config table (name=auto|preset|none, dark/light) mirroring [pager]; precedence --theme > env > config > auto; thread the chosen theme through App into render/tui/prompt.'
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, tui, design]
created: "2026-06-28"
blocked_by: [route-tui-chrome-through-the-palette, route-progress-bars-and-the-cli-ansi-map-through-the-palette, route-the-interactive-picker-theme-through-the-palette]
updated_at: "2026-06-29"
started_at: "2026-06-29"
completed_at: "2026-06-29"
id: 6fgq1n0006y3
---
## Objective
Let users select a theme via config/env/flag, and feed it to every routed surface.

## Scope
- In: `ThemeConfig` mirroring `PagerConfig` — `[theme]` table (`name = "auto" | <preset> | "none"`, `dark`/`light`); precedence `--theme` > `TSKFLW_THEME` > `[theme].name` > `auto`, gated behind `wantColor()`. `App.Theme` set in `resolve()`; threaded into `render.Style` (`WithTheme`), `prompt.NewTTY`, `tui.Run` — no globals, background detection kept lazy (OSC-11 latency). Commented `[theme]` block in the default config + a config round-trip test. The hardcoded `Default()` from the routing tasks becomes config-selected.
- Out: discovery commands + polish (next task).

## Done when
A `[theme]` selection flows end-to-end to CLI + TUI + picker; precedence honored; `build/test/lint` green. If a `--theme` flag is added, regenerate docs (docs-check gate).

## Reference
Design doc §5. Depends on the three routing tasks (TUI chrome, bars/ANSI, picker).

**Implementation 2026-06-29 (worktree, branch feat/theme-config off main).** The `[theme]` config table + selection plumbing landed:
- **config:** `ThemeConfig{Name}` + `themeFileTOML` mirroring `PagerConfig`; rides on the discovered config (not resolved across a planning_repo pointer); commented `[theme]` block added to the default config. Round-trip test `TestDiscover_ParsesTheme`.
- **selection:** `--theme` persistent flag + `TSKFLW_THEME` env; precedence flag > env > `[theme].name` > default, extracted into a pure `themeName()` helper with a direct unit test (`TestThemeName_Precedence`). Unknown/empty names degrade to the default via `design.Lookup` (never errors).
- **threading (no globals):** `App.Th design.Theme` resolved in `setStyle` (flag/env) and refined in `resolve` once config is discovered; fed to `render.Style.WithTheme`, `prompt.NewTTY(..., th)`, and `tui.Run(..., th)`. The three hardcoded `design.Default()` literals are gone.
- **CLI uses the theme's Dark palette** (semantic ANSI slots are background-independent, so styled CLI text is correct on any terminal; the bar gradient staying dark on a light terminal is the deferred polish tracked by the neon-day validation task — no eager OSC-11 query added). The TUI still resolves `For(dark)` once at startup.
- **docs:** regenerated `docs/cli` via docgen for the new persistent `--theme` flag (docs-check gate).

DEFERRED: `name = "none"` (monochrome — structure without hues, distinct from `--color=never`) is not yet a registered theme; the plumbing supports it the moment a `none`/`mono` palette is added to the registry. A second *named* theme is T6.

Verified: gofmt, go build ./..., go vet, full go test ./... green; smoke-tested --theme + TSKFLW_THEME + unknown-name degrade.

**Review addressed (2026-06-29).** A 2-lens adversarial review found the wiring correct + contract-safe: a theme swap can't change a byte of `--json`/piped/`--color=never` output (palette only feeds the TTY-gated path), the setStyle→resolve double-resolve is idempotent, DI/no-globals + layering hold. Fixed in this branch:
- **Unknown/typo theme names now warn** instead of silently rendering neon — `design.Lookup`'s `ok` was discarded everywhere, so `[theme] name = "none"` or a typo fell back to neon with no feedback. `warnUnknownTheme` emits one ⚠ to stderr (stdout stays clean); `""`/`"auto"` are guarded as the intentional default. Test: `TestWarnUnknownTheme`.
- **`render.Style.WithTheme` → `WithPalette(p)`** — the builder took a whole `Theme` but silently used only `.Dark`; the caller now passes `a.Th.Dark` explicitly (honest API, future-proofs the deferred light-gradient work).
- **Stale `NewStyle` comment** (pointed at "a later task" that is this one) fixed.

Deferred to T6 (logged there): exported `design.Names()` for `theme list`, wiring the dead `Theme.Markdown`, and theme-preview-must-not-stomp-`pal`. gofmt/vet/full test green; `--theme none` smoke-tested (warns + still runs).
