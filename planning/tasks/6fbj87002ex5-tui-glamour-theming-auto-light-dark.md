---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Glamour body uses a fixed dark style; adapt to terminal background (auto light/dark) and consider a style aligned with the theme palette
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [tui, bubble-tea]
created: "2026-06-12"
started_at: "2026-06-14"
updated_at: "2026-06-14"
completed_at: "2026-06-14"
id: 6fbj87002ex5
---

# TUI glamour theming auto light dark

## Objective

The S5 glamour body uses a fixed **`"dark"`** style (`glamour.go: glamourStyle`).
On a light-background terminal it reads poorly. Adapt the render style to the
terminal background, and consider aligning it with the shared `theme` palette so
the body matches the rest of the TUI. Deferred from S5 (out-of-scope: theming).

## Scope

- [x] Pick the glamour style by terminal background. — `glamourStyleFor(darkBG)`
      maps light→"light", else→"dark"; `Run()` resolves `lipgloss.HasDarkBackground()`
      ONCE at startup (a mid-program OSC query would race Bubble Tea's reader) and
      threads the style through to the detail pane. Cached renderer now keyed by
      width **and** style.
- [~] (Optional) a glamour `StyleConfig` aligned with the `theme` colors. — **not
      done** (deferred): it's a taste/design call, not needed for the readability
      fix. Left for a follow-up if wanted.
- [x] Test: the chosen style follows the background signal. — `TestGlamourStyleFor`
      (mapping) + `TestDetailPane_GlamourRendererRebuildsOnStyle` (cache rekeys).

## Out of scope

- Async/off-loop glamour rendering — the per-width renderer cache (landed in the S5
  fast-follow) removed the recompile cost; only revisit if very long bodies lag.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Follows [tui-glamour-markdown-rendering-with-rawpretty-toggle](6fb7ym4017sj-tui-glamour-markdown-rendering-with-rawpretty-toggle.md)
