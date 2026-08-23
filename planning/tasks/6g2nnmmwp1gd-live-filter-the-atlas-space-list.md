---
schema: 1
id: 6g2nnmmwp1gd
status: ready-to-start
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Wire the `/` filter key on the atlas screen so a large registry is navigable by label, id, or path instead of scrolling.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, atlas, ux]
created: "2026-08-22"
updated_at: "2026-08-23"
---
# Live-filter the atlas space list

## Objective

`/` opens live list filtering on every entity tab. `handleAtlasKey` leaves it unbound, so
a registry with many repos and worktrees is navigable only by `j`/`k` scrolling or by
cycling `o` order. The atlas is the one screen whose row count grows with the user's whole
machine rather than with one repo.

## Acceptance criteria

- [ ] `/` on the atlas opens a live filter over the cards, matching space id, entry-point
  label, and registry path.
- [ ] `esc` clears the filter and restores the previously selected logical space; `enter`
  on a filtered list opens the selection through the ordinary open path.
- [ ] The filter composes with `o`/`O` ordering without losing the selected space.
- [ ] The `?` help and the atlas footer name the key, consistent with the entity tabs.
- [ ] Tests cover filter → select → open and filter → clear → cursor restore.

## Out of scope

- Fuzzy/substring mode toggling (`F`) unless it falls out of reusing `internal/listfilter`.
- Persisting the filter across sessions.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Audit finding M4: [2026-08-22-multi-workspace-atlas](../audits/6g2k3qye4qma-2026-08-22-multi-workspace-atlas.md)

## Scope note (2026-08-23)

The atlas is becoming two cycled views. This task is the **spaces** view's filter — match
against space id, entry-point label, and registry path, as written above.

Whether the work view needs its own filter is a separate question and deliberately not
answered here: its rows are tasks, so a filter there would match task slugs and
descriptions and is closer to the entity tabs' existing `/` than to this one. Decide it
after living with the work view.

Order of work: this can land before or after the tile grid — it is cursor and matching
logic, not layout — but if it lands after, the filter's chrome has to fit whatever the
tile header becomes.
