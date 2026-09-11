---
schema: 1
id: 6g8vxcnezktm
status: next-up
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Reassess the TUI's entity, view, focus, history, and mode hierarchy now that hidden states and tabs are straining navigation.
effort: 1-2 days
tier: 1
priority: high
autonomy_level: 2
tags: [tui, ux, navigation, information-architecture, research, design]
created: "2026-09-10"
---
# Reassess TUI navigation and information architecture at current scale

## Objective

Re-evaluate the TUI's navigation model now that the original entity tabs, alternate detail views,
structured child selections, immersive modes, history, and Atlas have accumulated into a real
application. Produce a concise interaction architecture that makes location, available
destinations, action targets, and escape routes understandable without requiring memorized keys.

## Acceptance criteria

- [ ] A current-state inventory maps entity switching, tabs, pane focus, detail views, zoom,
  overlays, filters, structured child selection, cross-entity jumps, and history, identifying key
  collisions, hidden modes, unclear state, and inconsistent retreat behavior.
- [ ] The review distinguishes primary entities, filtered collections, alternate presentations,
  local graph focus, and transient overlays, then proposes a visible hierarchy and vocabulary for
  each rather than presenting all of them as equivalent tabs or modes.
- [ ] At least two coherent navigation models are compared against novice discovery, expert speed,
  narrow terminals, keyboard accessibility, future entities, and eventual web-adapter concepts.
- [ ] A binding and affordance contract defines how users discover destinations, identify the
  current view and active action target, move locally, open an entity, go Back, retreat from a mode,
  and reach help without context-dependent surprises.
- [ ] Existing shell-owned extension seams are assessed before proposing a rewrite; renderer-local
  shortcuts, core leakage, and entity-specific duplication are called out explicitly.
- [ ] The recommendation is tested through compact wireframes or state-transition examples and a
  maintainer feedback checkpoint, including the current Threads and Atlas experiences.
- [ ] A migration map re-scopes and sequences existing Back, structured-target, and alternate-view
  tasks, filing only the additional bounded implementation slices needed.

## Design attention

The original TUI deliberately chose a thin entity tab strip plus command jumps, but the number of
entities and nested presentations has grown. Treat that decision as evidence, not an immutable
constraint. Prefer one teachable system over more footer text, and preserve fast keyboard paths
without making hidden bindings the only way to discover functionality.

## Out of scope

- Implementing the redesign, changing planning-domain semantics, adding mouse support, or treating
  a web interface as a substitute for a coherent shared interaction model.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Inform the
  [alternate-view and Thread identity discoverability task](6g87qn72901g-make-alternate-tui-views-and-thread-identity-discoverable.md)
- Coordinate with the reusable Back and structured-detail action-target tasks.
- Coordinate with the
  [large-Thread graph navigation investigation](6g8vxbv3d4xn-explore-progressive-disclosure-and-navigation-for-large-thread-graphs.md).
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
