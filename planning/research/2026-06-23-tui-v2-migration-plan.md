---
status: reference
created: "2026-06-23"
tags: [tui, ux, perf, bubbletea, lipgloss, migration, plan]
---

# TUI → charm v2 migration + perf wins — sprint plan

Greenlit from [2026-06-23-lipgloss-v2-charm-ecosystem](2026-06-23-lipgloss-v2-charm-ecosystem.md): move the TUI to charm
**v2** (bubbletea v2 + bubbles v2 + lipgloss v2) and harvest the perf + feature
wins. Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md). This is the
*scoping* — the win taxonomy (what's free vs simpler-code vs needs-design), each
mapped to a concrete tskflwctl need, plus the task breakdown. Governing rule
(epic 20) still holds: the agent/pipeline machine contract is sacrosanct — the
porcelain `-o table`/`csv`/`json` bytes must stay identical; everything here is
the human/TTY face.

## Today's surface (what we're moving)
- `internal/tui` (~7k LOC) — **all on v1**: 18 `bubbletea` sites, `lipgloss` v1,
  `bubbles/key`, `glamour` v1.
- `internal/cli/render` — `lipgloss` v1 (human output).
- `fang` already pulls **lipgloss/v2** alongside → we carry **two lipgloss
  majors**; this sprint retires the v1 one.

## Win taxonomy

### A. Free with the migration (no extra code — just land v2)
- **OSC-11 `init()` removal** — *the broad perf win.* v1's package-level init
  pings the terminal for bg color and can hang ~5s when unanswered; v2 deleted
  it. bubbletea is in **every binary's import graph**, so this can tax even
  `task list`. ⮕ **Step 1 of the migration: measure the v1 latency on a TTY** so
  we have a real before/after, not a claimed one.
- **Cursed renderer** (ncurses-style diffing) — faster redraws on big planning
  trees; ~orders-of-magnitude less bandwidth (matters if `wish` ever serves it).
- **Synchronized output (Mode 2026)** — atomic frames, no tearing/cursor flicker
  on the fsnotify live-reload.
- **Wide-Unicode (Mode 2027)** — emoji/wide chars stop breaking table layout (we
  already fuss over wide-char alignment).
- **Auto color downsampling** (colorprofile) — colors "just work" across terminals.

### B. No-brainer — *less* code (simpler, do right after the port)
- **Native progress bars (bubbles v2 `progress`)** — retire the hand-rolled
  `miniBar` (`internal/tui/style.go:62`, used in `detail.go` + `item.go`) and the
  CLI `Style.Bar` (`internal/cli/render/style.go:121`, the `status` dashboard).
  Net code deletion; `progress.ViewAs(pct)` renders a static styled bar, so it
  works in both the TUI and the non-interactive `render` path. Ties nicely to the
  epic-rollup percentages we just corrected.

### C. Needs scoping + design (connect each to a real need — its own design task)
- **Layers / overlays** (v2 `View.Layer`) ⮕ **command palette** (replaces/augments
  the `:`-jump), **confirm-before-mutate modals** (task start/complete/deprecate),
  **floating help/keymap** — instead of full-screen context swaps. Connects to the
  existing quit-key context layering + the mutation actions.
- **Native clipboard (OSC52, works over SSH)** ⮕ **yank a task slug / epic id** to
  paste into `task start …`. Needs a keybinding decision.
- **Advanced cursor control** (shape/blink/position) ⮕ sharpen the inline `e`
  field editor.
- **Enhanced keyboard** (shift+enter, ctrl+alt combos, key-release) ⮕ richer,
  unambiguous keybindings.

### D. Deferred / declined (recorded so they're not re-litigated)
- **`wish`** (SSH-served TUI) → epic [19-web-companion-apps-over-a-shared-core](../epics/19-web-companion-apps-over-a-shared-core.md).
- `table`/`list` reskins of CLI output, `charmbracelet/log` → declined in the
  ecosystem doc.

## Migration shape (phases) + watch-items
1. **Foundation port** — `internal/tui` v1→v2 (mechanical: `View()`→`tea.View`
   struct, `KeyMsg`→`KeyPressMsg` (`.Type`→`.Code`, `.Runes`→`.Text`), bubbles
   getter/setters, mouse interfaces) + the FREE perf wins land automatically.
   Charm's old→new lookup table makes it search-and-replace-able.
2. **Consolidate lipgloss** — move `internal/cli/render` to lipgloss/v2 too, so
   the v1 major is fully retired. ⚠️ **Contract:** the porcelain `-o table`/`csv`/
   `json` carry no lipgloss styling (already ANSI-free), so they stay
   byte-identical; only the *human* colored output changes (TTY-only, not
   golden-locked). Verify the goldens are untouched.
3. **Harvest** — (B) progress bars, then (C) the designed feature wins.

Watch-items: lipgloss v2 `Color` is now `image/color.Color` (func), rippling
through `theme`/`render`/`tui` color usage; **`AdaptiveColor` was removed** → manual
light/dark, but our `internal/theme` pkg already centralizes this so it's
contained; **glamour** v1.0.0 — confirm a v2-compatible release before the port.

## Proposed tasks (epic 18)
1. **Migrate the TUI to charm v2** (foundation; --next) — phases 1–2 above; step 1
   measures the OSC-11 latency; keep porcelain bytes identical.
2. **TUI native progress bars via bubbles v2** (no-brainer; after 1) — delete
   `miniBar` + `Style.Bar`.
3. **Design the v2 TUI feature wins** (after 1) — overlays/clipboard/cursor/
   keyboard: connect each to a need, propose build tasks (the bucket-C scoping).

## Cross-refs
- Source decision: [2026-06-23-lipgloss-v2-charm-ecosystem](2026-06-23-lipgloss-v2-charm-ecosystem.md) ·
  [2026-06-21-fang-evaluation-spike](2026-06-21-fang-evaluation-spike.md).
- Epics: [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md) ·
  [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md) · [19-web-companion-apps-over-a-shared-core](../epics/19-web-companion-apps-over-a-shared-core.md).
