---
status: in-progress
epic: 18-tui-bubble-tea-interactive-planning-browser
description: FilterMatchesMsg misrouted to the active tab blanks filtered background tabs on reload; cursor restore lost; detail pane stale while filtering
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [go, tui, bug]
created: "2026-06-12"
updated_at: "2026-06-12"
started_at: "2026-06-12"
---
# TUI filter × reload correctness

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([[2026-06-12-critical-code-review-multi-lens]], findings H2 + A1 + the
> cursor-restore loss). Merge-blocker for `feat/tui-multi-entity-navigation`:
> all three are interactions between `/` filtering and the new
> watcher-triggered `reloadAll`.

## Objective

1. **H2 — Filtered background tab goes blank after a reload.** `reloadAll`
   calls `SetItems` on a background tab with a filter applied, which nils its
   `filteredItems` and returns a refilter cmd — but the resulting
   `FilterMatchesMsg` is forwarded by the default branch to the **active**
   tab's list (`internal/tui/model.go:139-143`, `167-168`). The background
   tab is left in `FilterApplied` state with nil matches (empty until the
   filter is cleared); if the active tab also has a filter, it receives the
   other entity's match set. Route filter cmds/msgs by tab (wrap the cmd),
   or clear/reapply filters explicitly on reload.
2. **Cursor restore is lost when a reload hits a filtered tab.**
   `selectByID` (`entity.go:63-70`) walks `VisibleItems()` while
   `filteredItems` is still nil (refilter is async), consuming `restore`
   with no effect — cursor lands at index 0 when matches arrive. Keep
   `restore` armed until a `selectByID` succeeds (or restore against
   `Items()` only when unfiltered).
3. **A1 — Detail pane stale while typing / after applying a `/` filter.**
   The `SettingFilter()` path (`model.go:201-206`) and the forwarded
   `FilterMatchesMsg` (`model.go:141-143`) both skip the prev/`selectedID()`
   diff that only `updateList` (`model.go:288-304`) performs. Funnel
   filter-affecting messages through the same tail: capture
   `prev := m.selectedID()` and `loadItem`/`showEmpty` on change.

## Acceptance criteria

- [ ] Reload while a background tab has an applied filter: switching back
      shows the (refiltered) rows, never a blank list or cross-entity rows.
- [ ] Cursor returns to the previously selected slug after a reload on a
      filtered tab when the item survives.
- [ ] Detail pane tracks the selection while typing a filter, after Enter,
      and shows empty-state on zero matches.
- [ ] teatest coverage for reload-while-filtered (the harness from
      `TestModel_FilterNarrows` is the right tool); a detail-content
      assertion in the filter test so A1 cannot silently regress.

## Related

- Epic [[18-tui-bubble-tea-interactive-planning-browser]]
- Touches `internal/tui/model.go`, `internal/tui/entity.go`,
  `internal/tui/model_test.go`.