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
depends_on: [6g6dw5js81f3, 6g8vxcnezktm]
updated_at: "2026-11-01"
audit_sources: [2026-11-01-weekly-task-sweep]
audited: "2026-11-01"
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

## Sweep annotation (2026-11-01)

The task's founding premise re-verified as **still accurate**: there is no shared
action-target resolver. Each binding resolves its own target independently in
`internal/tui/model.go` — `keys.Edit` via `m.selectedTask()` / `m.selectedEpic()`
/ `m.selectedRef()` (≈L928–951), `keys.Yank` via `m.selectedYankRef()` (L965),
`keys.YankPath` via `m.yankSelectedPath()` (L968) — which is exactly the
"root-model actions individually special-cased" condition AC2 asks to remove.

Partial, unticked progress on AC3's second half: the `keys.Edit` path on a Thread
with a structured child selection now explains itself rather than silently acting
on the parent, emitting "no inline edit here — press E to edit in $EDITOR" and
two sibling messages for the loading and no-local-path cases
(`internal/tui/model.go` ≈L940–948). AC3 also requires footer/help text naming
the operative target, which does not exist yet, so the criterion is left
unchecked deliberately rather than judged met.

Prerequisite note: `6g6dw5js81f3` (two-dimensional Thread graph prototype) is now
**completed**, so the "Discovered while refining `y`" citation is historical; the
remaining gate is `6g8vxcnezktm` (navigation reassessment, `next-up`).

## Progress Log

- 2026-11-01: automated weekly sweep — premise re-verified against current
  `model.go`; recorded that AC3's self-explaining half has partially landed while
  AC2's shared resolver has not, and that one of two prerequisites has completed.
