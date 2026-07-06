---
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Esc in list focus quits the app (bubbles list default Quit binding); q from single-pane detail exits instead of popping back to the list
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [go, tui, bug]
created: "2026-06-12"
updated_at: "2026-06-12"
started_at: "2026-06-12"
completed_at: "2026-06-12"
id: 6fbj87001mq1
---
# TUI quit-key layering — Esc/q pop context before quitting

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([2026-06-12-critical-code-review-multi-lens](../research/2026-06-12-critical-code-review-multi-lens.md), findings H1/A3). H1 was
> hand-verified: no `DisableQuitKeybindings` call exists anywhere in
> `internal/tui`. Merge-blocker for `feat/tui-multi-entity-navigation`.

## Objective

1. **H1 — Esc in list focus quits the entire app.** `mk()`
   (`internal/tui/entity.go:121-132`) never calls
   `l.DisableQuitKeybindings()`, so with no filter applied Esc falls through
   the model's list branch (`model.go:247-255`) into `list.Update` →
   `handleBrowsing`, which returns `tea.Quit`. The UX spec
   ([2026-06-09-tui-ux-design-and-navigation-spec](../research/2026-06-09-tui-ux-design-and-navigation-spec.md)) says Esc returns focus
   to the left panel.
2. **A3 — `q` from single-pane detail quits the app** instead of popping back
   to the list: the global `keys.Quit` match (`model.go:219`) runs before
   focus routing, contradicting the locked "q context-quit" design. The
   detail footer even advertises "q quit".

## Acceptance criteria

- [x] `DisableQuitKeybindings()` (or equivalent interception) in `mk()`; Esc
      in list focus pops context / no-ops, never `tea.Quit`.
- [x] In single-pane drill with detail focus, `q` returns to the list; `q`
      from the list still quits.
- [x] Tests: Esc pressed in list focus (currently uncovered —
      `TestModel_FocusRouting` only presses Esc detail-focused), and `q` in
      single-pane detail focus.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Touches `internal/tui/entity.go`, `internal/tui/model.go`,
  `internal/tui/model_test.go`.