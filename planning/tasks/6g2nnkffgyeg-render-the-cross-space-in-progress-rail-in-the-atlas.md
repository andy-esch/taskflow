---
schema: 1
id: 6g2nnkffgyeg
status: in-progress
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Cycle the atlas between a spaces view and a full-viewport cross-space work view, so the rail is a screen rather than a band competing with the cards.
effort: M
tier: 3
priority: high
autonomy_level: 3
tags: [tui, atlas, ux]
created: "2026-08-22"
updated_at: "2026-08-23"
started_at: "2026-08-23"
---
# Add a cross-space work view to the atlas

## Objective

The atlas answers "where can I go?" but not "what am I working on, anywhere?" —
the question the design sketch calls its actual payload. `SpaceOverview.InProgress` is
already computed, carries its space id, and is rendered by `status --all`; `atlas.view`
discards it and keeps only a count.

This task deliberately does **not** add it as a band beneath the cards, which is how it
was originally scoped. Two questions are being asked of one screen — *where can I go*
(spatial, comparative) and *what am I working on* (temporal, sorted) — and they want
opposite layouts. Making them two cycled views decouples this work from the tile rewrite:
a full-viewport work view ships against today's list layout and survives that rewrite
untouched, whereas a band would compete with tiles for vertical budget and have to be
built after them.

See [The atlas as a dashboard of dashboards](../research/6g2qtp0022t7-the-atlas-as-a-dashboard-of-dashboards.md).

## Acceptance criteria

- [ ] A key cycles the atlas between a **spaces** view (today's cards) and a **work** view,
  fitting the existing `o`/`O` and `s`/`S` cycling idiom rather than inventing a concept.
  The active view is named in the header, the footer, and `?` help.
- [ ] The work view renders `overview.InProgress` full-viewport: one row per in-progress
  task across every healthy space, each naming its space, slug, and age, in one
  byte-stable order.
- [ ] Rows are navigable, and entering one opens that task's space through the same
  `core.WorkspaceService` path `⏎` on a card already uses — landing on the task where
  possible rather than only on the space.
- [ ] Spaces whose summary failed to load contribute no rows; their card-local error stays
  visible in the spaces view rather than being silently dropped from both.
- [ ] The view survives the atlas lifecycle already established: refresh (`r`), ordering,
  registry-load failure, and the workspace-generation stamp that drops stale async results.
- [ ] Which view is active persists across an atlas round trip within a session, and resets
  to spaces on a fresh launch.
- [ ] **No banner.** Aggregate summary lines are deliberately excluded so that living with
  the work view answers whether one earns its rows — see the research doc's open forks.
- [ ] Tests cover the view cycle, the row → space navigation, an empty working set, and a
  partial-failure overview.

## Out of scope

- The tile grid, per-space accents, and any change to how a card renders — separate task.
- A banner or aggregate summary band (see above; deliberately deferred, not forgotten).
- A third "attention" view (acute findings / ready-to-close / revisits). Open fork 2.
- Worktree-aware entry selection.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Design: [The atlas as a dashboard of dashboards](../research/6g2qtp0022t7-the-atlas-as-a-dashboard-of-dashboards.md)
- Audit finding M3: [2026-08-22-multi-workspace-atlas](../audits/6g2k3qye4qma-2026-08-22-multi-workspace-atlas.md)
