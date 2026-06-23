---
status: reference
created: "2026-06-23"
tags: [tui, design, decision, lipgloss-v2]
---

# v2 TUI feature wins — adopt/decline decisions

Design output for [[design-the-v2-tui-feature-wins-overlays-clipboard-keyboard]].
The charm-v2 migration unlocked capabilities that the migration plan parked in a
"needs-scoping" bucket. This weighs each against a **real** TUI need and decides
adopt / decline / defer. Every capability below was verified present in our pinned
deps (bubbletea v2.0.7, bubbles v2.1.0, lipgloss v2.0.4). **No code here** — the
greenlit items are filed as their own build tasks.

## TL;DR

| Capability | Decision | Need it serves |
| :-- | :-- | :-- |
| Clipboard (OSC52 + native) | **DONE** | yank slug/path — shipped this cycle |
| Layers / overlays | **ADOPT** | retire hand-rolled `overlay()`; enable a command palette |
| Command palette | **ADOPT** (after the layers harvest) | fuzzy launcher over entities + verbs |
| Confirm-before-mutate modal | **DECLINE** | nothing destructive to guard |
| Mouse | **DECLINE (now)** | keyboard-first; breaks native text selection |
| Advanced cursor control | **DECLINE** | bubbles inputs already render cursors |
| Enhanced keyboard (release/disambig) | **DECLINE** | keymap is complete; Kitty not universal |
| Gradient progress bars | **ADOPT** | richer completion read on the rollup bars |
| Live light/dark re-theme | **ADOPT** | adapt to a terminal theme flip mid-session |
| Sync output / downsampling / wide-Unicode | **BANKED** | already gained by migrating |

Greenlit → four filed tasks: the **overlay-layers harvest** (foundation) then the
**command palette** (the flagship overlay feature), plus — bumped up on review (the
"DEFER" was too conservative) — **gradient progress bars** and **live re-theme**.
The latter two are independent of the layers work, so they can land in any order.

---

## Clipboard — DONE

`tea.SetClipboard` (OSC52) + a native-tool fallback (`pbcopy`/`wl-copy`/`xclip`/…).
Shipped as [[tui-clipboard-yank-for-slug-and-file-path-via-osc-52]] (`y` = slug,
`Y` = path). Lesson banked in code: OSC52 alone is unreliable (many terminals
ignore it), so native-first is the right default; OSC52 is the SSH fallback.
Possible future nicety (not filed): yank the rendered detail body, or a
ready-to-run `task start <slug>` line. Low priority.

## Layers / overlays — ADOPT

**Capability (verified):** `lipgloss.NewCanvas(w,h)` + `lipgloss.NewLayer(s).X().Y().Z()`
→ `canvas.Render() string`. True z-ordered compositing, replacing manual string
splicing.

**Need:** the TUI hand-rolls overlay compositing in `overlay()` (`help.go:109`),
called by the modal registry (`overlay.go`: help/action/follow/edit) via
`lipgloss.Place` + manual row splicing. It works but is bespoke and can't stack or
cast shadows.

**Decision: ADOPT**, in two slices:
1. **Overlay-layers harvest** (foundation) — replace `overlay()` with a
   `Canvas`/`Layer` composite. Net code deletion, correct z-order, sets up stacked
   floating UI. Mirrors the progress-bar harvest. **HIGH** value-to-risk, but needs
   a human visual pass (no PTY in CI) — the modals must render identically or better.
2. **Command palette** (feature, after #1) — see below.

## Command palette — ADOPT (sequenced after the layers harvest)

**Need:** `:` already does entity/status/verb jump, but it's a single prompt line.
A floating, fuzzy-searchable palette over **all tasks/epics/audits + verbs** (think
Telescope / ⌘P) is the natural power-user upgrade and the flagship use of the new
layers. Reuses the existing `listfilter` matcher + bubbles/list (same components the
picker/TUI already use).

**Decision: ADOPT**, MEDIUM-large, **depends on the overlay-layers harvest** (it's a
floating layer). Scope: open with a key (e.g. `ctrl+p` / `P`), fuzzy over a flat
index of entities + verbs, enter to jump/run, esc to dismiss. Keep `:` as the terse
path.

## Confirm-before-mutate modal — DECLINE

**Considered for:** start/complete/defer/deprecate guarded by a confirm overlay.
**Why decline:** every TUI mutation is a **reversible file move** (deprecate →
`deprecated/`, undo by moving back); nothing is destructive. A confirm step would be
friction without a danger to justify it. Revisit only if a genuinely irreversible
op (e.g. delete) is ever added.

## Mouse — DECLINE (now)

**Capability (verified):** declarative via `View.MouseMode` (`MouseModeCellMotion`/
`AllMotion`) + `MouseClickMsg`/`MouseWheelMsg`/`MouseMotionMsg`/`MouseReleaseMsg`.

**Why decline:** this is a keyboard-first, vim-keyed tool whose users (and agents)
live on the keyboard. Critically, **enabling mouse mode at all disables the
terminal's native click-drag text selection** (you'd need ⌥/shift-click) — a real
regression for a tool people copy slugs out of. Net: low upside, concrete downside.
**Revisit** only as an explicit opt-in (`[tui] mouse = true` in `.tskflwctl.toml`,
default off) if a user asks. Not filed.

## Advanced cursor control — DECLINE

**Capability:** `View.Cursor *tea.Cursor` (position/shape/blink/visibility).
**Why decline:** the inline `e` editor, `/` find, and `:` inputs are bubbles
`textinput`/`textarea`, which **already render their own cursors**. A top-level
hardware cursor is redundant polish with no current need.

## Enhanced keyboard — DECLINE

**Capability:** `KeyReleaseMsg`, `KeyboardEnhancementsMsg` (Kitty disambiguation:
shift+enter, ctrl+alt+…, key-release events).
**Why decline:** the keymap (`keys.go`) is complete and conflict-free with plain
keys; there's no binding we can't express today. Kitty enhancements degrade on
terminals that lack them, so building on them invites inconsistency. No concrete
need → decline.

## Theming polish — ADOPT (reprioritized on review)

Originally parked as DEFER; bumped up on review. Two independent slices (neither
depends on the layers work):

- **Gradient progress bars** — `lipgloss.Blend1D` / bubbles `progress.WithColors`
  (multi-color blend) on the epic rollup bars. **Design constraint:** must not muddy
  today's completion semantics (the solid fill currently *means* the tier:
  gray < 34, yellow < 100, green at 100). A red→yellow→green blend across the fill
  is a richer read of the same signal; a gradient that obscures the tier is not.
  Small, low-risk, but needs a human visual pass.
- **Live light/dark re-theme** — handle `tea.BackgroundColorMsg` to re-resolve the
  glamour style + adaptive colors when the terminal flips theme mid-session, instead
  of today's one-shot `lipgloss.HasDarkBackground` at startup (`tui.go`). Medium;
  request the bg on init, update `m.detail.glamStyle` on the msg.

## Deferred / banked

- **Synchronized output** (Mode 2026, flicker-free), **auto color downsampling**,
  **wide-Unicode** (Mode 2027): **already banked** by the migration — the v2
  renderer does these for free. No action.

## Filed build tasks

- `tui-overlay-compositing-via-lipgloss-v2-layers-canvas` (foundation) — HIGH.
- `tui-command-palette-fuzzy-launcher-over-entities-and-verbs` (after the harvest) — MEDIUM.
- `tui-gradient-progress-bars-via-lipgloss-blend` (independent) — MEDIUM.
- `tui-live-light-dark-re-theme-via-backgroundcolormsg` (independent) — MEDIUM.

## Sources / related

- Verified against the module cache: `bubbletea/v2@v2.0.7` (`tea.go` View struct,
  `mouse.go`, `keyboard.go`, `clipboard.go`), `lipgloss/v2@v2.0.4`
  (`canvas.go`, `layer.go`).
- [[2026-06-23-tui-v2-migration-plan]] · [[2026-06-23-lipgloss-v2-charm-ecosystem]].
