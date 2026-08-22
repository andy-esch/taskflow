---
schema: 1
id: 6g2nnkffgyeg
status: ready-to-start
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Show overview.InProgress under the atlas cards so the atlas answers what am I working on everywhere, the payload status --all already renders.
effort: M
tier: 3
priority: high
autonomy_level: 3
tags: [tui, atlas, ux]
created: "2026-08-22"
updated_at: "2026-08-22"
---
# Render the cross-space in-progress rail in the atlas

## Objective

`core.SpaceOverview` already carries `InProgress` (one row per in-progress task across
every healthy space, with its space id), and `status --all` renders it. `atlas.view`
discards it and shows only aggregated counts per card ("1 in progress · 12 epics").

The research sketch names this rail the atlas's actual payload — cards orient, the rail
answers "what am I in the middle of, anywhere?" — so the atlas currently ships the
orienting half without the answering half.

## Acceptance criteria

- [ ] The atlas renders `overview.InProgress` as a labelled section beneath the cards,
  each row naming its space, the task slug, and its age, in one byte-stable order.
- [ ] The rail is bounded on a short terminal: cards and rail share the body budget
  without the card list becoming unnavigable, and the selected card stays visible.
- [ ] Rows are navigable or explicitly not; if selecting a row enters that task's space,
  it opens through `core.WorkspaceService` on the same path `Enter` on a card uses.
- [ ] Spaces whose summary failed to load contribute no rows and still show their
  card-local error.
- [ ] Rendering tests cover a multi-space rail, an empty rail, and the short-terminal
  budget split.

## Out of scope

- Per-space accent colors (the other unbuilt idea from the sketch).
- Persisted or cached cross-space stats.
- Mutating tasks from the atlas.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Design sketch: [Multi-space: home registry and the atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)
- Audit finding M3: [2026-08-22-multi-workspace-atlas](../audits/6g2k3qye4qma-2026-08-22-multi-workspace-atlas.md)
