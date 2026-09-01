---
schema: 1
id: 6g5rxq17px59
status: next-up
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Separate durable row/navigation identity from display slugs so duplicate-slug tasks, audits, research, and Threads cannot collide across reloads or workspace sessions.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, architecture, identity, correctness]
created: "2026-09-01"
---

# Make TUI entity navigation use stable identities

## Objective

Make the TUI's internal row, selection, reload, detail, back-stack, and workspace-session identity
match the stable IDs that persistence treats as canonical. Slugs remain the human-facing label and
filter vocabulary, but duplicate slugs must not make two distinct records select, restore, or open
as though they were one entity when Threads join the registry.

## Current constraint

`entityItem.id()` currently returns `Task.Slug`, `Audit.Slug`, and `Research.Slug`, even though those
documents have stable IDs and their stores deliberately permit duplicate slugs. The same value is
used for cursor restoration, selection checks, detail reads, navigation history, palette jumps, and
saved workspace sessions. Adding Thread rows with that convention would copy an existing ambiguous
identity bug into a new first-class entity.

## Scope

- Separate durable row/navigation identity from display/filter text in the entity-item contract;
  avoid overloading one `id()` method with both a canonical key and a human label.
- Use each entity's canonical stable reference for service calls, selection restore, stale-result
  guards, navigation/palette targets, the back stack, and per-workspace session state. Epics retain
  their existing canonical ID behavior.
- Keep slugs and descriptions in visible rows, detail titles, fuzzy filtering, and user-facing
  messages; include a short stable-ID disambiguator only where two visible rows would otherwise be
  indistinguishable.
- Define explicit behavior for malformed or drifting IDs that resilient list reads retain. Do not
  silently fall back to a colliding slug; preserve repair visibility and use the store's canonical
  resolvable identity when one is available.
- Make the entity registry contract safe for the later Thread item without adding the Thread tab in
  this task.

## Acceptance criteria

- [ ] Two tasks with the same slug but distinct stable IDs remain independently selectable,
  detail-loadable, filterable, and cursor-restorable across manual and watcher reloads.
- [ ] Equivalent duplicate-slug coverage exists for every stable-ID TUI entity whose store permits
  the state, and the future Thread item contract has one documented canonical-key path.
- [ ] Navigation history, command-palette jumps, stale async result guards, and saved workspace
  sessions carry canonical identity rather than display slug.
- [ ] Slug-based display and filtering remain readable; duplicate visible labels gain deterministic
  disambiguation without exposing IDs everywhere by default.
- [ ] Missing/drifting-ID fixtures fail visibly or use a documented canonical filename identity;
  they never collapse onto another row or open an arbitrary duplicate.
- [ ] Existing unique-slug navigation behavior and TUI golden output remain unchanged except where
  identity disambiguation is intentionally exercised.

## Stress tests

- Duplicate slugs in one list, duplicate slug hidden by a filter, reload with the second duplicate
  selected, detail loads completing out of order, back navigation across duplicates, workspace
  switch/restore, malformed frontmatter ID, and filename/frontmatter ID drift.

## Out of scope

- Persisting TUI sessions across process restarts, changing resolver ambiguity policy, renaming
  documents, adding Thread screens, or replacing user-facing slugs with IDs.

## Sequencing

Independent TUI foundation discovered while scoping Thread views. The Thread list/detail slice must
depend on it so the new entity never adopts slug-keyed state.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Downstream [Thread list and detail views](6g5rwjqr7rt8-add-thread-list-and-detail-views-to-the-tui.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
