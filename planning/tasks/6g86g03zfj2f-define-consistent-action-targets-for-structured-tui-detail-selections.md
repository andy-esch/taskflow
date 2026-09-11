---
schema: 1
id: 6g86g03zfj2f
status: next-up
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Give structured detail cursors one explicit shell-owned target contract for open, copy, edit, path, follow, and lifecycle actions.
effort: 2-3 days
tier: 2
priority: high
autonomy_level: 3
tags: [tui, navigation, ux, architecture, threads]
created: "2026-09-08"
depends_on: [6g6dw5js81f3]
updated_at: "2026-09-08"
---

# Define consistent action targets for structured TUI detail selections

## Objective

Define and implement one shell-owned rule for actions invoked while a structured detail cursor
highlights a child entity. Thread topology and spatial views made the ambiguity concrete: `Enter`
and `y` target the highlighted task, while path/editor and potential lifecycle actions still target
or derive from the parent Thread row. Make that distinction intentional, visible, portable, and
reusable by future structured entity presentations.

## Acceptance criteria

- [ ] A concise target matrix defines parent-versus-child behavior for `Enter`, `y`, `Y`, `E`, `e`,
  `m`, and `f` under list focus, ordinary detail focus, and structured child focus.
- [ ] Shared TUI infrastructure resolves the active action target by canonical identity; Thread
  renderers and future structured details do not special-case root-model actions individually.
- [ ] Footer/help text names the operative target where ambiguity is possible, and unsupported
  actions explain themselves instead of silently acting on the parent entity.
- [ ] Optional local-path capability is preserved: portable/pathless child targets degrade
  explicitly without manufacturing paths or preventing access to the parent document.
- [ ] Duplicate labels, unreadable children, reloads, cross-entity jumps, and switching focus back
  to the parent list have focused behavioral coverage.
- [ ] Existing list-row action behavior and stable navigation/history semantics remain compatible.

## Out of scope

- Inventing new entity mutations, embedding filesystem knowledge in portable projections, or
  replacing the separate reusable Back-action/history work.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Discovered while refining `y` in the
  [two-dimensional Thread graph prototype](6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md)
- Coordinate with the
  [reusable TUI Back action](6g86016qvk5d-add-a-reusable-back-action-for-tui-entity-navigation.md)
- Thread [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)
