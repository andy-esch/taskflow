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
updated_at: "2026-09-12"
depends_on: [6g86c7y6hn41]
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

## Dogfood evidence: the spatial graph is a graph engine without a reading hierarchy

A 185-column capture of the 15-node navigation Thread makes the current scale problem concrete.
Before the first node, six dense rows mix Thread identity, topology completeness, selection layer,
dependency rank, boundary role, direction, health, viewport inventory, task status, node role, route
focus, and fan counts. The bottom inspector is semantically useful but can sit dozens of rows from
the selected node. Healthy evidence consumes the same prominence as a problem, while terms such as
`gate` and `fan` require graph-domain knowledge that the screen does not teach.

The `?` overlay already has section machinery and more than key bindings, but it is titled `Keys`,
fixed to a narrow key-reference layout, and receives only the active entity kind. It cannot know
whether Thread summary, dependency ranks, the full graph, or one-hop focus is active. Evaluate a
spatial-specific graph guide that leads with how to read the current view, defines Thread task and
external prerequisite concepts in plain language, explains multiple-route and crossing marks, then
lists navigation controls. Use that guide to remove permanent legend density rather than adding a
second wall of explanations.

The redesign should compare at least these concrete moves: a small mode/title strip; healthy state
collapsed while degraded evidence stays loud; selected-task context adjacent to the graph instead
of a distant bottom chasm; overflow cues shown only when relevant; plain-language node and route
vocabulary; and a shorter footer that makes `?` discoverable as the graph guide. Preserve exact
projection evidence and keyboard speed, but do not require users to decode renderer implementation
terms before they can answer what is active, what blocks it, and what it unlocks.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Inform the
  [alternate-view and Thread identity discoverability task](6g87qn72901g-make-alternate-tui-views-and-thread-identity-discoverable.md)
- Coordinate with the reusable Back and structured-detail action-target tasks.
- Coordinate with the
  [large-Thread graph navigation investigation](6g8vxbv3d4xn-explore-progressive-disclosure-and-navigation-for-large-thread-graphs.md).
- Thread [Refine Thread and TUI navigation](../threads/6g90h4pg0q7n-refine-thread-and-tui-navigation.md)
