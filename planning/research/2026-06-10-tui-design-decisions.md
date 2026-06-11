---
status: reference
created: 2026-06-10
tags: [tui, bubble-tea, ux, reference]
---

# TUI design decisions & build reference

Distilled from two research agents (UX patterns: k9s/lazygit/gh-dash/gitui;
Bubble Tea architecture/testing) on 2026-06-10. The *decisions* live in epic
[[18-tui-bubble-tea-interactive-planning-browser]]; this is the **how-to-build**
reference for the sprint tasks. Supersedes the over-reaching parts of
`2026-06-09-tui-ux-design-and-navigation-spec.md` (Projects-tab, multi-select).

## Keybinding scheme (vim-first, locked)

Keys route through the **focused pane only** — never a global switch — so the
same key means different things per pane without collision.

**Global** (suppressed while `/` or `:` is capturing text):
`:` command/jump · `/` filter · `?` help overlay · `Esc` cancel/back (never
quits) · `Tab`/`Shift+Tab` cycle pane focus · `[`/`]` prev/next entity ·
`q` context-quit (close overlay/mode first) · `Ctrl+c` hard quit.

**List pane:** `j`/`k` move · `Ctrl+d`/`Ctrl+u` half-page · `g`/`G` top/bottom ·
`Enter`/`l` drill into detail · `h` back · `s` cycle status filter.

**Detail pane:** `j`/`k` + `Ctrl+d`/`u` + `g`/`G` scroll · `h`/`Esc` back to list.

**Do NOT** bind `1`–`6` to entity tabs (lazygit disabled this after collision
complaints — number keys stay free for vim count-prefixes).

## Entity switching: `:` jump, not tabs

k9s rejects persistent tabs for resource switching (its bottom bar is a nav
*stack*). A tab strip is fine at 3, cramped at 6, broken at <60 cols. So:
**`:` command-jump is primary** (`:tasks`/`:epics`/`:audits`, also `:in-progress`
filters) — O(1) horizontal cost at any entity count. A thin **tab strip + `[`/`]`**
is the discoverable affordance. Both read one **entity registry**, so
Projects/ADRs/Research later = a registration, not new keys/layout.

## Layout, focus, responsive

- **Focus = two signals:** accent border (lipgloss `BorderForeground`) **and** a
  pane-title marker (robust on mono/color-blind terminals). Selected row:
  reverse-video when focused, dim underline when not.
- **Responsive (recompute on every `WindowSizeMsg`):** ≥100 cols two-pane
  (≈40/60); 60–99 single-pane *drill* (list full-width → `Enter` replaces with
  full-width detail → `Esc` back); <60 collapse the tab strip to `[Entity ▾]`.
- **Truncate, never wrap** rows in bordered panes. Subtract the border+padding
  **frame** (`style.GetHorizontalFrameSize()`) from child sizes before sizing
  list/viewport — the #1 lipgloss layout bug. `JoinHorizontal` doesn't clip:
  make `listW + detailW + gutters == width` exactly.
- **Row design:** `status-glyph │ slug │ relative-date │ ⚠(if Misfiled)`.

## Bubble Tea architecture

**Package `internal/tui/`** (replace the stub): `tui.go` (`Run(svc, root)`),
`model.go` (root Model owns `svc`+size+focus+tab), `keys.go` (`key.Binding` +
`help.KeyMap`), `messages.go`, `commands.go` (Cmd factories calling `svc`),
`layout.go`, `panes/{list,detail,search}.go` (pure, never see `svc`), `item.go`
(delegates), `watch.go`. The `core.Service` lives on the **root model only**.

**Bubbles verdicts:** `list` ✓ (550 items fine — paginates/virtualizes; custom
`ItemDelegate` for the glyph row; `FilterValue()=slug+" "+description`);
`viewport` ✓ (detail; it does **not** wrap — pre-wrap content to inner width,
re-wrap on resize); `textinput` ✓ (`/`; **gate all key routing while
`Focused()`** except esc/enter); `help` ✓ (footer from the keymap); `spinner` ✓
(only while loading — it animates only as long as you keep returning its tick).

**The rules:** no `svc.*` (I/O) in `Update`/`View` — every read is a `tea.Cmd`
→ custom `tea.Msg` (`Service` re-scans the fs per call, ~550 files). Lazy-load
the detail body via `ShowTask` on selection change; **guard stale results by
slug** (fast cursor moves can deliver an old body late). Model methods are
value-receivers: always reassign (`m.x, cmd = m.x.Update(msg)`). Don't render
before the first `WindowSizeMsg` (sizes are 0). Never mutate `m.*` from inside a
Cmd closure — build new slices, hand back via Msg.

## fsnotify live reload (S3)

~½–1 day. fsnotify is non-recursive → watch each of the ~9 status/bucket dirs.
Debounce editor save-storms: emit one `reloadMsg` per ~200ms quiet period
(coalesce via `tea.Tick`, never `time.Sleep` in a Cmd). On reload, re-fire
loaders + **re-arm** the watcher. **Preserve cursor by slug** (capture
`SelectedSlug()`, re-`Select()` after; clamp if it vanished) — never by index.
Plumb the `reloadMsg` path in S1 so S3 just wires the source.

## Testing

~80% **message-injection unit tests**: build the model, send msgs to `Update`,
assert on state / `View()` substrings; back it with a `fakeStore`-backed
`core.Service` (`core.Store` is already an interface). ~20% **`x/teatest`
golden** full-program tests for layout regressions (few — brittle across lipgloss
versions; always set an explicit term size; inject a no-op watcher so
`WaitFinished` doesn't hang).

## Shared theme (S0, do first)

Extract a dependency-free `theme` package (imports only `domain`) holding the
semantic map `Status/Bucket/Priority → {Glyph, Color hex}` — currently private in
`render/style.go` (`statusGlyph`). CLI `render` maps `Color`→nearest ANSI (keep
byte-stable output); TUI maps `Color`→`lipgloss.Color`. One place decides
"in-progress is yellow ●"; no import cycle (`render` and `tui` both import
`theme`).

## Top footguns to design against

I/O in `Update`/`View` · forgetting to reassign value-receiver updates ·
borders/padding not subtracted from child sizes · `textinput` swallowing nav
keys · rendering before first resize · spinner/ticker that never stops ·
out-of-order async results · un-debounced fsnotify storms · restoring cursor by
index not slug · `q` quitting unconditionally (layer `Esc` under it).

## Layout discipline (audited 2026-06-11, validated by research agent)

The S1 two-pane layout was audited after a clipped-top-border report and
hardened; these are the rules every future pane/tab must follow. The
**invariant test** (`TestModel_ViewFitsTerminal`) locks them: `View()` is
exactly `height` lines and **no line's display width exceeds `width`**.

- **Subtract the frame before sizing children.** Never hardcode `2`. Derive
  `paneHFrame`/`paneVFrame` from the pane style's
  `GetHorizontalFrameSize()`/`GetVerticalFrameSize()` (style.go) so a future
  border/padding change can't silently desync sizing.
- **Guard `View()` before the first `WindowSizeMsg`.** With unset (0) sizes,
  panes compute negative border dimensions → an oversized frame that corrupts
  the renderer's height tracking (the clipped-top-border bug). Return a plain
  `"loading…"` until `width>0 && height>0`.
- **Clamp every inner dimension to ≥1** (`max1`). A tiny terminal must degrade,
  never produce a negative-sized (panicking/broken) frame.
- **Truncate, never wrap, anything fed to a `Join`.** `JoinVertical` pads every
  row to the widest child — one overflowing line (e.g. the footer) widens the
  whole frame past the terminal. Use the ANSI/width-aware `truncate`
  (`ansi.Truncate`), not a rune slice.
- **Final clamp as last line of defense:** wrap the composed view in
  `MaxWidth(width).MaxHeight(height)` so a single missed truncation degrades
  gracefully instead of corrupting the screen.
- **No trailing newline from `View()`** in altscreen (it pushes the frame up a
  row → top border scrolls off).
- **Measure display cells, not bytes/runes,** anywhere width matters — ANSI
  escapes and wide runes both fool `len`/rune-count (`ansi.StringWidth`).

When the 3rd pane/tab lands, extract a small internal `pane`/layout helper
rather than copy-pasting this arithmetic.
