---
status: reference
created: "2026-06-28"
tags: [cli, tui, ux, theme, color, lipgloss, research, decision]
---

# Color/design overhaul — one palette, established themes, a neon default

Research + audit for
[[color-and-design-overhaul-one-coherent-palette-across-every-surface]] (epic
[[20-cli-ux-and-ergonomics]]). Builds on [[2026-06-23-lipgloss-v2-charm-ecosystem]],
[[2026-06-21-fang-evaluation-spike]], and [[2026-06-10-tui-design-decisions]].
Four parallel research lenses (internal audit, external theme libraries, Charm v2
theming mechanics, UX/accessibility) + source-verified against the module cache.
**No code here — the map, the options, and a recommended shape + sequence.**

Governing rule (epic 20): never compromise the agent/pipeline contract. Color is
TTY-gated; `--json`/piped output stays byte-stable plain. Every semantic color
keeps its distinct glyph so state survives `--color=never` / mono / colorblindness.

## TL;DR — recommendation

1. **Don't hand-roll a palette; lean on the `base16` standard.** Its 16 slots
   (`base00`–`base0F`) map *by construction* onto the 6 ANSI chromatic slots the
   CLI already emits — degradation is slot==slot, not a runtime nearest-color
   guess. This is the single best technical fit for a tool that is 16-color ANSI
   on the CLI and truecolor lipgloss on the TUI.
2. **Default neon theme: `neon-night`**, ported from the real, maintained base16
   scheme **Synth Midnight Terminal Dark** (Michaël Ball), with the danger-red
   swapped to Outrun's brighter `#FF4242` for legibility. Light fallback
   `neon-day` ≈ **Catppuccin Latte** (already an importable Go module —
   `github.com/catppuccin/go` is *already in our graph*).
3. **Architecture: a new `internal/design` package** owning the truecolor
   `Palette` + named-`Theme` registry. Keep `internal/theme` exactly as-is — it
   stays the domain-only *semantic* decision layer; `design` is the new tech-aware
   *color* layer both presenters consume. Each palette entry carries an explicit
   ANSI-16 anchor (base16 gives us these for free) so the CLI stays deterministic.
4. **Config: a `[theme]` table** mirroring `[pager]` exactly — `name = "auto"`
   resolving dark/light via the background detection we already do, with
   `--theme` flag > `TSKFLW_THEME` env > config > `auto` precedence (same ladder
   as `--color`). Add `tskflwctl theme list` / `theme preview`.
5. **Scope: lock to COLOR.** Glyph set, border style (`RoundedBorder`), and
   spacing stay fixed and out of scope (the glyphs are accessibility-load-bearing
   — don't make them swappable). Document this as an explicit non-goal so the
   "design overhaul" title doesn't scope-creep.
6. **Sequence chrome-first, config-last** — the TUI chrome literals have *no*
   test coverage and zero golden footprint, so they're the safe place to start;
   wire config once at the end against already-routed surfaces.

---

## 1. The audit — every color source, mapped

The project's color lives in three tiers: a clean semantic core, two
tech-adapters that consume it correctly, and a scatter of chrome literals that
bypass it. The overhaul is about that third tier.

### Tier 1 — the semantic core (keep, untouched)

- `internal/theme/theme.go` — **single source of truth for *semantic*
  presentation**. A 7-value `Color` enum (`ColorNone..ColorGray`) + glyph+color
  `Token`s for `Status`/`Bucket`/`FindingStatus`/`Liveness`/`Priority`/`Percent`/
  `Marker*`. **Hard invariant (load-bearing):** imports only `internal/domain` —
  no ANSI, no lipgloss, no `image/color`. That's what lets the CLI render it
  16-color and the TUI render it truecolor from the *same* decision.
  `theme_test.go` pins the glyph+enum decisions (it asserts no hex/ANSI — so it
  won't move).
- Markdown style names already centralized here (`MarkdownStyleDark="dracula"` /
  `MarkdownStyleLight="light"` + `MarkdownStyleFor(darkBG)`), resolved per-surface
  from background detection (`root.go:182`, `tui/tui.go:22`). **This is the model
  to imitate** — a named decision in one place, terminal detection at the adapter.

### Tier 2 — the two tech-adapters (keep the pattern, repoint the values)

- `internal/cli/render/style.go` — `ansiCode(theme.Color)` maps the enum to
  hand-written 16-color SGR consts (`ansiRed="\x1b[31m"`…`ansiGray="\x1b[90m"`).
  There is **no colorprofile/downsampler in the CLI path** — it literally needs a
  16-color slot to print.
- `internal/tui/style.go` — `lipColor(theme.Color)` maps the enum to lipgloss
  16-color indices (`"1".."8"`); `fg()`/`glyph()` render it.

### Tier 3 — the bypass set (the actual work: route these through one palette)

| file:line | current literal | role |
|---|---|---|
| `progressbar.go:16-19` | `#b026ff`→`#00e5ff`→`#ff2ec4` neon gradient | epic rollup bar fill |
| `progressbar.go:36` | `EmptyColor = Color("8")` | bar empty track |
| `progressbar.go:60-63` | seg bands `"2"/"3"/"8"` | audit segmented bar |
| `cli/prompt/tty.go:91-98` | `huh.ThemeDracula` + `#b026ff` caret | the interactive picker |
| `tui/style.go:18` | `accent = Color("6")` | focus/selection accent |
| `tui/style.go:26-27` | pane borders `accent` / `"8"` | active/idle pane chrome |
| `tui/dashboard.go:26` | `dashHeading` `Color("6")` | dashboard heading |
| `tui/help.go:152-153` | `helpBorder`/`helpHeading` `"6"` | help overlay |
| `tui/action.go:164-166` | `actionBorder`/`dangerBorder`/`actionHeading` `"6"/"1"` | action overlays |
| `tui/find.go:42-43` | `findMatch` bg`"3"` / `findCurrent` bg`"11"`, fg`"0"` | find highlight |
| `tui/edit.go:382-383` | edit-box border `Color("8")` | edit overlay |

Third-party palettes not coordinated with ours: huh `ThemeDracula` (prompts),
glamour `dracula`/`light` (markdown bodies).

**The asymmetry the task asks about:** `progressbar` importing lipgloss is *fine*
— it's already a presenter (emits ANSI, the CLI strips it). `theme` importing
lipgloss is *not* — it's the pre-presentation decision layer both adapters
consume. So the new truecolor palette must NOT live in `theme`.

---

## 2. Established theme libraries — the landscape

The owner's constraint: lean on an existing **library** of named palettes, not a
bespoke color system. The findings, ranked by fit:

### base16 / tinted-theming — the recommended foundation

- A 16-slot palette **standard** (`base00`–`base0F`) with defined slot semantics
  and **326 base16 + 190 base24 community schemes** in one MIT repo
  (`github.com/tinted-theming/schemes`, actively maintained into 2026).
- Slot semantics: `base00–07` a dark→light mono ramp (bg/fg/UI); `base08–0F` the
  accent ramp — **08 red, 09 orange, 0A yellow, 0B green, 0C cyan, 0D blue, 0E
  magenta/purple, 0F brown**.
- **Why it's the right fit (the crux):** base16's accent ramp was *designed* to
  correspond to the ANSI chromatic slots. tinted-shell's own template maps
  `base08→red, base0B→green, base0A→yellow, base0D→blue, base0E→magenta,
  base0C→cyan` directly onto ANSI 1–6/9–14. So choosing colors in base16 form
  means **we've already decided the 16-color rendering** — the CLI degradation is
  an identity map, not termenv's HSLuv nearest-neighbor guess (which silently
  collapses both `#b026ff` purple and `#ff2ec4` pink onto the *same* magenta
  slot — see §4).
- **Go story:** no off-the-shelf importable scheme-loader library; the realistic
  path is to **vendor the handful of hexes we need as Go constants** (we only
  ship a few curated themes, not all 326). The `.yaml` schemes are MIT.
- Honest limit: base16 has 8 accents but ANSI has only 6 chromatic slots —
  **orange (`base09`) and brown (`base0F`) have no ANSI home.** Our semantic
  vocabulary (red/green/yellow/blue/cyan/gray) doesn't use them, so this costs us
  nothing; just don't build the neon identity on orange.

### Catppuccin — the one with an official Go module (and it's already in our graph)

- `github.com/catppuccin/go` v0.3.0, MIT, official, method-based API
  (`catppuccin.Mocha.Mauve().Hex` → `"#cba6f7"`). 4 flavors (Latte/Frappé/
  Macchiato/Mocha), 26 named roles.
- **Already an indirect dependency** (pulled by `huh.ThemeCatppuccin`) — verified
  in `go.mod`. So Catppuccin Latte as the **light** fallback costs zero new deps
  and is a real library, not hardcoded hexes.
- Take: Mocha is *pastel/cozy*, not 80s — wrong for the neon default, right for
  the legible light variant (Latte).

### The rest — editor themes, no official Go bindings

Dracula, Rosé Pine, Tokyo Night, Gruvbox, Nord, Everforest, Kanagawa: all are
canonical editor/terminal schemes but ship **no official Go module** (only
hardcoded hexes or unofficial ports). Several are reachable *through Charm*
already, though (next section).

### What Charm ships natively (free, already in our stack)

- **glamour v2 styles:** `dark, light, dracula, tokyo-night, pink, notty, ascii`.
  `tokyo-night` is built-in — a neon-adjacent markdown style for free.
- **huh v2 themes:** `ThemeBase`, `ThemeCharm`, `ThemeDracula`, **`ThemeBase16`**,
  **`ThemeCatppuccin`** (all `isDark`-parameterized). `ThemeBase16` is notable —
  huh already speaks base16.
- **lipgloss v2:** no named themes, just primitives (`Color`, `LightDark`,
  `Blend1D/2D`, `Lighten/Darken/Complementary`).
- **fang v1** (direct dep): `fang.ColorScheme` is itself a "struct of named
  colors" fed via `WithColorSchemeFunc` — prior art for the pattern, and the
  [[2026-06-21-fang-evaluation-spike]] already flagged feeding our theme into it.

---

## 3. The neon/80s question — concrete candidates

All in base16 form (hexes source-verified from `tinted-theming/schemes`):

| candidate | bg | the neon read | availability |
|---|---|---|---|
| **Synth Midnight Terminal Dark** | `#050608` | near-black void; `#42fff9` cyan + `#06ea61` green pop like a CRT — "cyberpunk terminal" | real base16 file ✔ |
| **Outrun Dark** | `#00002A` | inky navy; screaming `#0EF0F0` cyan + `#F10596` magenta — pure OutRun arcade sunset | real base16 file ✔ |
| **Synthwave '84** (Robb Owen) | `#262335` | purple-haze bg, hot-pink `#ff7edb` + cyan `#36f9f6` — the *definitive* look, max "rad" | VS Code theme; **no** base16 port — must hand-author |
| Tokyo Night (neon config) | `#1a1b26` | city-glow, cooler/bluer — "neon-tasteful" not arcade | glamour built-in; hexes for the rest |

Catppuccin Mocha, Tokyo Night (default), twilight, tomorrow-night,
material-darker, espresso — all too pastel/muted to read as true neon. (The
oft-cited `cyberpunk` base16 scheme does not exist in the repo — 404.)

**Recommended default — `neon-night`:** base it on **Synth Midnight Terminal
Dark** (it's real, maintained, near-black so accents glow, and its accent ramp is
already base16-clean), with **one substitution**: swap the danger-red from Synth's
`#b53b50` (fails contrast, see §4) to **Outrun's `#FF4242`**. This gives a
maintained-scheme spine with our one deliberate legibility fix. If we want maximum
"rad" later, a hand-authored Synthwave '84 port can be added as a *second* named
theme — but the default should be the legible, real-scheme one.

`neon-night` role mapping (dark):

| role | hex | base16 | ANSI anchor |
|---|---|---|---|
| background | `#050608` | base00 | (terminal default) |
| foreground | `#c1c3c4` | base05 | 37 |
| dim/gray | `#a3a5a6` / track `#474849` | base04 / base03 | 90 |
| red (danger) | `#FF4242` *(Outrun)* | base08 | 31 |
| green (success) | `#06ea61` | base0B | 32 |
| yellow (warn) | `#c9d364` | base0A | 33 |
| blue (next-up) | `#03aeff` | base0D | 34 |
| cyan (info/ready) | `#42fff9` | base0C | 36 |
| **accent/selection** | `#ea5ce2` magenta | base0E | **95** (bright magenta) |
| bar gradient | `#b026ff → #00e5ff → #ff2ec4` | (deliberate truecolor exception) | — |

`neon-day` (light fallback) = **Catppuccin Latte** via the already-present Go
module: text `#4c4f69`, accent mauve `#8839ef`, red `#d20f39`, blue `#1e66f5`,
etc. (legible accents for text; the sub-4.5:1 ones lean on the fixed glyphs).

---

## 4. The two-tech reality — degradation & accessibility (hard constraints)

### Truecolor → 16-color is unpredictable; base16 makes it deliberate

- termenv/lipgloss downsample truecolor via **HSLuv nearest-neighbor** over the 16
  slots. Verified collapses: `#b026ff` neon purple → **bright magenta (13)**;
  `#00e5ff` → bright cyan (14); `#ff2ec4` pink → **also magenta** — purple and
  pink become *indistinguishable* in 16-color.
- lipgloss v2 does **not** auto-downsample inside `Render` — it happens at the
  output writer (`colorprofile.Writer`, which honors `NO_COLOR`/`CLICOLOR_FORCE`).
  Bubble Tea v2 wraps output in one automatically; the CLI does not.
- **Therefore: each palette entry carries an explicit ANSI-16 anchor.** base16
  hands us these for free (slot==slot). The CLI's `ansiCode` becomes
  `sgr(palette.Of(c).ANSI)` — deterministic and testable — instead of a
  hardcoded switch. The primary magenta accent anchors to **bright magenta (95)**
  so the synthwave identity survives a 16-color tmux/CI terminal. The TUI uses the
  truecolor value and lets the profile writer degrade it.

### Accessibility — confirmed constraints the overhaul must not regress

- **Contrast (computed vs `#050608`):** cyan/green/yellow/blue all pass AA with
  huge headroom (8–16:1) → safe for colored *text*. **`#b026ff` purple is 4.4:1
  and Synth red `#b53b50` is 3.6:1 — both FAIL AA for small text.** Rule: reserve
  sub-4.5:1 neons (purple, saturated red) for **glyphs, bars, borders,
  selection** only; use cyan/green/yellow/bright-pink for any colored text. (This
  is why the default danger-red moves to `#FF4242`, 5.9:1.)
- **Light mode:** neon assumes dark bg; the same hues are unreadable on light
  (Latte: pink 2.3:1). **Never auto-invert** — ship a separate hand-tuned light
  variant (the bat `theme-light` / our existing glamour dracula-vs-light model).
- **Glyph pairing is load-bearing and must stay fixed:** status `●○✔✘◌`, liveness
  `●✦○`, segmented bar `█▓▒░`. These carry state through `--color=never`/mono/
  colorblindness (WCAG 1.4.1). A theme may remap *hues* but must never collapse
  two states onto one glyph, and the destyled (`!s.on`) path stays byte-identical.

---

## 5. Architecture — where the palette lives and how it flows

### A new `internal/design` package (not an extension of `theme`)

Three options were weighed; **(b) a new package** wins:

- **(a) extend `theme`** — rejected: a real neon palette needs truecolor, which
  forces `image/color`/lipgloss into a package whose entire reason to exist is
  being free of a color tech. Breaks the invariant both adapters depend on.
- **(c) resolve the existing enum per-theme** — rejected: chrome (`accent`,
  `border`, `selected-bg`, `match-highlight`, the **gradient stops**) is *not*
  semantic and has no home in a Red/Green/Yellow vocabulary. Forcing
  `accent→ColorCyan→"6"` is exactly the flattening we're undoing.
- **(b) `internal/design`** — owns the truecolor `Palette` + named-`Theme`
  registry. Dependency arrows stay acyclic: `theme` keeps importing only
  `domain`; `design` imports `image/color` (+ may import `theme` to key its
  semantic slots by `theme.Color`); `render`/`tui`/`progressbar`/`prompt` import
  `design`. **One-way: `theme` never imports `design`.**

Type sketch:

```go
// internal/design — Hue carries BOTH techs from one definition.
type Hue struct {
    True color.Color // truecolor (the TUI/huh/glamour use this; lipgloss degrades it)
    ANSI int         // 0..15 SGR slot (the CLI's hand-written path needs this)
}

type Palette struct {
    Semantic [7]Hue // indexed by theme.Color — backs Status/Bucket/Priority/...
    // chrome (the gap the enum can't express):
    Accent, Dim, SelectedFg, BorderActive, BorderIdle, Danger,
    Heading, MatchBg, MatchCurBg, MatchFg, Track Hue
    Gradient []color.Color // the deliberate truecolor bar exception
    MarkdownDark, MarkdownLight string // glamour names move here from theme
}
func (p Palette) Of(c theme.Color) Hue { return p.Semantic[c] }

type Theme struct { Name string; Dark, Light Palette }
func Lookup(name string) (Theme, bool)
func Default() Theme // neon-night
func (t Theme) For(darkBG bool) Palette
```

**On "use a library, not custom":** define `Palette` as the internal *contract*
(it has to express "this token is `#ea5ce2` truecolor AND ANSI slot 95" — no raw
library does that), but **populate the built-in themes from established schemes**:
base16 hexes for `neon-night`, the `catppuccin/go` module (already present) for
the Latte light variant. Library for the *values*, a thin `design.Palette` for
the *routing*.

### Driving all four Charm surfaces from one `Palette`

Verified mechanics (against the `charm.land` fork source — flag: the **bubbles
`progress` fork API differs from upstream**; it uses `WithColors(...color.Color)`
+ exported `EmptyColor`, which our code already uses correctly — do not port from
upstream progress docs):

- **lipgloss/TUI** — `Foreground(p.Accent.True)` etc. (already the `tui/style.go`
  pattern). Light/dark via `tea.BackgroundColorMsg` → `lipgloss.LightDark(isDark)`
  (this also enables the deferred live re-theme task,
  [[tui-live-light-dark-re-theme-via-backgroundcolormsg]]).
- **huh/picker** — build a full `*huh.Styles` from the palette inside
  `huh.ThemeFunc(func(isDark bool)...)` (the `isDark` flag is exactly where to
  pick `For(isDark)`); set `Focused.SelectSelector`/`SelectedOption` from
  `p.Accent`. Can start from `ThemeBase16`/`ThemeDracula` and migrate fields
  incrementally.
- **glamour/markdown** — `glamour.WithStyles(ansi.StyleConfig)` pours a palette
  into the markdown style (`*string` hex fields for Document/Heading/Link/Code…).
  Caveat: glamour does **not** self-downsample (v2 dropped `WithColorProfile`) —
  route its output through the same `colorprofile.Writer`, or author 16-safe
  hexes. Simplest first step: keep `WithStandardStyle` but source the *name* from
  `palette.MarkdownDark/Light` so it's at least palette-selected.
- **bubbles progress** — `progress.WithColors(p.Gradient...)` +
  `p.EmptyColor = p.Track.True`; build gradients from palette anchors with
  `lipgloss.Blend1D`.

### Config plumbing — a `[theme]` table mirroring `[pager]`

`internal/config/config.go` already has the exact template: `PagerConfig` is an
optional local-terminal table read straight off the discovered config, **not**
resolved across a `planning_repo` pointer. A theme is the same kind of concern.

```toml
[theme]
name  = "auto"        # "auto" | <preset> | "none"   (default: auto)
dark  = "neon-night"
light = "neon-day"
```

- `name = "auto"` → resolve dark/light by the background detection we already do.
  A concrete preset pins it; `"none"` = monochrome (glyphs/bold/dim, no hue) —
  distinct from `--color=never` (which kills all ANSI).
- **Precedence (same ladder as `--color`):** `--theme` flag > `TSKFLW_THEME` env
  > `[theme].name` > `auto`. Gated behind the existing `wantColor()` result.
- Config stays `domain`-only: it carries the theme *name string*, never a
  `design.Palette` — resolving name→Theme happens in the adapter, exactly as
  `Root` (a string) resolves to a store in `root.go`.
- **The flow:** add `App.Theme design.Theme`, set it in `resolve()` from
  `cfg.Theme.Name`; thread it into `render.Style` (new `WithTheme`/`WithPalette`
  builder), `prompt.NewTTY(in, out, th)`, and `tui.Run(svc, layout, th)`. No
  globals — rides on `*App` like `Cfg`/`Svc`. **Watch the OSC-11 latency:** keep
  background detection lazy (store the `Theme` with both palettes; resolve
  `For(dark)` only where color actually renders, like `markdownStyle()` does).
- **Discovery UX:** add `tskflwctl theme list` (+ `--json`) and `theme preview
  [name]` (render a sample dashboard/bar in the theme) — a neon feature is
  unsellable without a preview. Name presets after their base16/scheme origin for
  terminal-wide interop; do **not** read tinty's `current_scheme` file.

---

## 6. Scope boundary (explicit non-goals)

Terminal "themes" in the wild span palette + border style + glyph set + bar chars
+ spacing. **Lock this overhaul to COLOR** (palette + the semantic→color maps +
bar gradient stops + accent). Keep **glyphs, `RoundedBorder`, and spacing fixed
and out of scope**:

- The glyphs are the accessibility fallback (§4) — user-swappable glyphs invite
  breaking the mono/colorblind guarantee.
- A color-only theme slots into the existing `theme.Color`→tech seam with near-zero
  structural churn; glyph/border theming would touch layout/sizing
  (`paneHFrame`, `apportion`, truncation) and the `?` legend.

Themeable now: bg, fg/dim, the 6 semantic slots, the accent/selection, the bar
gradient stops. Not themeable: glyph set, border style, spacing, destyled output.

---

## 7. Migration sequence (each step keeps `just build/test/lint` green)

Test landscape (verified): `theme_test.go` is a glyph+enum decision table (no
hex/ANSI — won't move). CLI goldens are ANSI-stripped (no ESC in
`internal/cli/testdata` — porcelain is plain), so color routing is ~zero golden
churn. TUI tests are `View()` substring assertions — chrome moving from a global
to a model field is invisible to them.

1. **Introduce `internal/design`** (additive; nothing consumes it). `Hue`,
   `Palette`, `Theme`, the `neon-night`/`neon-day` defaults, `Lookup`/`Default`/
   `For` + a `design_test.go` pinning each token's ANSI slot + the gradient.
2. **Route TUI chrome** — convert the global style vars (`tui/style.go`,
   `dashboard.go`, `help.go`, `action.go`, `find.go`, `edit.go`) into a `styles`
   struct built from a `Palette` in `New`; thread `Default().For(true)` for now.
   (No test coverage to break — safest first move.)
3. **Route progressbar + render's ANSI** — gradient/track/seg-bands from the
   palette; `ansiCode` → `sgr(pal.Of(c).ANSI)`, **pinning neon-night's ANSI slots
   to today's exact SGR values** so plain-stripped goldens stay byte-identical.
4. **Route the picker** — `pickerTheme` reads `Palette.Accent` instead of
   `#b026ff`.
5. **Wire config + selection** — `[theme]` table (mirror Pager), `App.Theme` in
   `resolve()`, thread through the three seams. The hardcoded `Default()` from
   steps 2-4 becomes config-selected.
6. **Add named themes + land the deferred picker polish** (slug in `Accent`,
   description in `Dim` — the [[render-the-interactive-picker-inline-not-full-screen-alt-screen]]
   stopgap's debt) against the real palette. Optionally finish glamour/huh onto
   the palette.

Invariant never touched throughout: `theme`'s enum + tokens (and `theme_test.go`)
— the symbology/segmented-bar work the task warns against regressing stays pinned.

---

## Open questions for the implementer

- **glamour depth:** is palette-selecting the *style name* (`dracula`/
  `tokyo-night`) enough, or do we author a full `ansi.StyleConfig` from the
  palette? (Recommend: name-only first; full StyleConfig is a follow-up.)
- **Second theme:** ship just `neon-night`+`neon-day` at first, or also a
  hand-authored Synthwave '84 and a base16-classic? (Recommend: ship the default
  pair + prove the registry with one more, e.g. Catppuccin, to exercise the
  light-module path.)
- **`theme preview`** richness — a static swatch grid vs a live mini-dashboard.

## Sources

base16 spec/schemes: github.com/chriskempson/base16/blob/main/styling.md ·
github.com/tinted-theming/schemes · base16-shell template (ANSI map). Synthwave:
github.com/robb0wen/synthwave-vscode · synth-midnight-dark.yaml · outrun-dark.yaml.
Charm v2: pkg.go.dev/charm.land/{lipgloss,huh,glamour}/v2 ·
github.com/charmbracelet/{fang,glamour,huh}; verified against the local module
cache (`charm.land/*`). Catppuccin: github.com/catppuccin/go (already an indirect
dep). Degradation: github.com/muesli/termenv (color.go/profile.go) ·
lipgloss CompleteColor. Accessibility: WebAIM contrast · Catppuccin palette ·
octalwave terminal-scheme generator. Config prior art: bat (`--theme`/BAT_THEME/
`--list-themes`), delta, k9s skins, starship `[palettes]`, helix `base16_*`, tinty.
