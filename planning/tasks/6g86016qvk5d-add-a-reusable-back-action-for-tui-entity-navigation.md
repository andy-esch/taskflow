---
schema: 1
id: 6g86016qvk5d
status: next-up
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Make cross-entity navigation reversibly discoverable through one shell-owned Back action, including task jumps from Thread graphs.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 3
tags: [tui, navigation, ux, threads]
created: "2026-09-08"
depends_on: [6g6dw5js81f3]
updated_at: "2026-09-08"
---

# Add a reusable Back action for TUI entity navigation

## Objective

Make following an entity in the TUI explicitly and consistently reversible. Introduce one
discoverable shell-owned Back action that works after opening a task from a Thread graph and across
the other cross-entity navigation surfaces, without making individual detail renderers own history.
Reconcile the new action with the existing `ctrl+o` jump history and contextual `Esc`/`h` behavior
instead of creating competing stacks.

## Acceptance criteria

- [ ] A documented Back binding returns from a Thread graph task jump to the same Thread
  presentation and selected canonical task identity.
- [ ] The action is owned by shared TUI navigation/history infrastructure and is reusable by every
  entity type rather than special-cased in the Thread renderer.
- [ ] Existing `ctrl+o`, contextual `Esc`, and `h` behavior have an explicit compatibility rule;
  the footer/help makes the preferred Back action discoverable where it is available.
- [ ] Cross-entity, reload, removed-target, empty-history, and repeated-back behavior is covered by
  focused tests.
- [ ] A bounded pass applies the shared action to existing cross-entity jumps and records any
  intentionally different local retreat behavior.

## Out of scope

- Browser-style forward history, persistence across process restarts, or changing entity-specific
  selection and scrolling semantics unrelated to returning from a jump.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Discovered while dogfooding the
  [two-dimensional Thread graph prototype](6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md)
- Thread [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)

## Sequencing

Follow the spatial Thread graph prototype so its real task-jump behavior supplies the first pinned
case, but keep this as general TUI navigation hardening. The prototype may continue using the
existing `ctrl+o` history while this task chooses and applies the reusable, discoverable contract.
