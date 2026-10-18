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
depends_on: [6g6dw5js81f3, 6g8vxcnezktm]
updated_at: "2026-10-18"
audit_sources: [2026-10-18-weekly-task-sweep]
audited: "2026-10-18"
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

## Sweep refresh (2026-10-18)

2026-10-18: automated weekly sweep — the Back action itself is still unbuilt, but
part of the infrastructure AC 2 asks for landed after this task was last updated,
and the Sequencing section is now out of date.

**Drift found — the restore machinery AC 1 describes already exists, under
`ctrl+o`.** When this task was written (last updated 2026-09-13) the follow stack
remembered only *where* a jump came from. `internal/tui/nav.go:32-38` now defines
`detailNavigationRestore{kind, key, view, selection, context}`, and
`internal/tui/model.go:103` holds `pendingDetailNavigation` to "restore
presentation/selection after an async `ctrl+o` return". That is precisely AC 1's
"returns ... to the same Thread presentation and selected canonical task
identity" — the capability exists; what is missing is the *discoverable, shared
Back action* that AC 2 and AC 3 are really about. The task is now less about
building history and more about naming, binding, and generalising history that
already works. Scope deliberately left unchanged; this is a framing note for
whoever picks it up.

**Verified accurate (2026-10-18).**

- `ctrl+o` is still the only jump-back binding (`internal/tui/keys.go:36,80`,
  `JumpBack`), still described as the vim jumplist, and the footer still
  advertises it as a breadcrumb rather than a Back affordance
  (`internal/tui/view.go:523-524`). AC 3's compatibility rule is still unwritten.
- `navStack` is still a shell-owned `[]navLoc` on the model
  (`internal/tui/model.go:102`) and `internal/tui/nav.go:211,253` still describe
  the shell as the owner of history — so AC 2's "not special-cased in the Thread
  renderer" premise holds today and is a property to preserve, not to create.

**Sequencing is satisfied but unrevised.** The Sequencing section says to follow
the spatial Thread graph prototype. That prototype (`6g6dw5js81f3`) is
`completed`, so the stated precondition is met and the pinned case it was meant
to supply is available. Left in place rather than rewritten, per sweep policy.

**Dependency state (2026-10-18).** `gate: blocked`, `eligible: false`,
`role: queued`. Of the two `depends_on` edges, `6g6dw5js81f3` is `completed` and
`6g8vxcnezktm` (`reassess-tui-navigation-and-information-architecture-at-current-scale`)
is `next-up` / `not-started` — that second edge is the sole remaining blocker.
The prose does not mention it: Sequencing discusses only the prototype. Not a
wrong edge (the reassessment's own AC 7 explicitly re-scopes "existing Back,
structured-target, and alternate-view tasks", so the edge is sound) — just a
Sequencing section that predates the edge and no longer tells the whole story.
