---
schema: 1
id: 6g87qn72901g
status: next-up
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Make alternate views visible and redesign Thread rows and focus as identifiable, status-aware cards instead of truncated symbol-heavy lines.
effort: 2-3 days
tier: 2
priority: high
autonomy_level: 3
tags: [tui, threads, ux, navigation, discoverability]
created: "2026-09-08"
depends_on: [6g6dw5js81f3]
updated_at: "2026-09-08"
---
# Make alternate TUI views and Thread identity discoverable

## Objective

Make Threads immediately identifiable in the TUI and make alternate presentations visible before
the user already knows the `v` binding. Redesign the Thread list/focus treatment around recognizable
content, then introduce one reusable visual affordance for discovering and identifying available
views in both Thread detail and Atlas.

## Acceptance criteria

- [ ] Thread list rows lead with a human-identifiable Thread title or name rather than status syntax
  and a readily truncated slug; lifecycle remains visible through the shared colored status marker.
- [ ] Health, progress, and other compact symbols are either self-explanatory or supported by a
  nearby legend/help treatment, without making color the only carrier of meaning.
- [ ] The focused Thread may expand into a bounded multi-line card showing its complete identity and
  a useful subset of description, goal, lifecycle, progress, and graph health without changing the
  height or information density of every row.
- [ ] Thread detail and Atlas visibly identify the current view and the available alternate views;
  discovering that `v` cycles presentations does not require opening global help or prior knowledge.
- [ ] View discovery is a reusable shell-level presentation contract rather than bespoke labels in
  the Thread graph and Atlas renderers; keyboard behavior and current-view restoration remain stable.
- [ ] Narrow terminals retain the most identifiable text first, expanded focus degrades cleanly,
  and selected/unselected, reload, resize, and all supported-view states have focused coverage.

## Out of scope

- Adding new Thread or Atlas views, changing Thread graph semantics, mouse interaction, or solving
  spatial edge routing and structured child-action targeting in this task.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Discovered while dogfooding the
  [two-dimensional Thread graph prototype](6g6dw5js81f3-prototype-a-two-dimensional-navigable-thread-graph-view.md)
- Coordinate with
  [structured-detail action targeting](6g86g03zfj2f-define-consistent-action-targets-for-structured-tui-detail-selections.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)

## Design attention

Treat the focused-card idea as a responsive list interaction, not permission to make every row
taller. Explore a compact view switcher such as labelled chips or a small tab strip that communicates
both current and adjacent presentations while preserving `v` as the fast cycle binding. The visual
affordance should work for Atlas modes and future multi-view entities without teaching the shell
Thread-specific view names.

## Sequencing

Follow the spatial prototype so all three Thread presentations are available as the first complete
test case. It may proceed independently of dense edge routing and one-hop focus; coordinate only at
the shared detail footer/view-discovery and action-target seams.
