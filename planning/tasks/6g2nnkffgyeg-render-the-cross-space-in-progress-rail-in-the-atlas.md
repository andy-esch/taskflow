---
schema: 1
id: 6g2nnkffgyeg
status: completed
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
completed_at: "2026-08-23"
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

## Implementation notes (2026-08-23)

The atlas now cycles between a **spaces** view (the existing cards) and a **work** view:
one row per in-progress task across every healthy space, most-recently-touched first, with
space, slug, age, and description in data-derived columns capped so one long slug cannot
push every description off screen.

**The key is `[` / `]`, not a new binding.** Those already mean "move between sibling
screens" everywhere else in this TUI and were dead keys on the atlas. Atlas views are
siblings, so they keep that meaning — which also closes a gap noted during the v0.17.0
audit. `h`/`l` and `o`/`O` stay spaces-only and are no longer promised by the footer while
the work list shows, since they would do nothing there.

**Landing on the task, not the space, needed an intent that outlives the switch.** The
tabs to land on do not exist until the open succeeds, so `Model.pendingJump` carries the
target across and `activateWorkspace` consumes it *instead of* the ordinary dashboard load
— `jumpTo` threads the id through the tab's own loader, so it stays one load rather than a
jump stacked on a dashboard fetch. A failed open drops the intent so it cannot fire against
whatever workspace is opened next.

The space owning a row is resolved by durable planning identity, falling back to the
registry label only for id-less legacy trees, so a relabelled entry still matches.

Verified by recording the real thing, not just tests: `]` reaches the work view showing
5 tasks across 2 spaces, `j`+`⏎` enters `kitchen` and lands on the tasks tab with the
cursor already on the chosen task and its detail open.

Two pre-existing assertions were updated rather than deleted: the spaces header gained its
view name, and a help assertion was matching a phrase that now word-wraps — it was
narrowed to wrap-safe fragments and extended to cover the work view.

**No banner**, per the design: living with a full-viewport work view is what should answer
whether an aggregate band earns its rows.

Validation: `go test -race ./...`, `golangci-lint run ./...`, planning lint, generated
docs, and `git diff --check` all clean.

## Amendment (2026-08-23) — the view-cycle key

The first cut used `[`/`]` to cycle the atlas's two views, on the reasoning that those keys
already mean "move between sibling screens". Review rejected it, correctly: the page strip
is visibly `atlas · overview · tasks · epics · audits · research`, so keys that appear to
walk that strip must actually walk it. Scoping them to atlas sub-views made the atlas a
dead end you could only leave with `a`.

Two changes:

- **`v` cycles the current screen's views.** It is a different axis from `s` (which *rows*)
  and `[`/`]` (which *screen*), so it gets its own key. Deliberately general rather than
  atlas-specific, since other pages may grow alternate views. It is advertised only in the
  atlas help, not the global section — promising it on screens where it is inert would be
  a lie.
- **The atlas joined the `[`/`]` page ring, in strip order.** `]` from the atlas reaches the
  overview; `[` wraps back to the last tab; `[` from the overview reaches the atlas; `]`
  from the last tab wraps forward to it. The ring skips the atlas entirely for consumers
  that mount no overview service, so an embedded single-space TUI is unchanged.

The atlas is now a first-class page rather than a modal side-trip.

Tests: `TestPageRingIncludesTheAtlasInStripOrder` walks every hop, and
`TestPageRingSkipsTheAtlasWhenItIsNotMounted` pins the unchanged fallback. Verified by
recording: `v` → work, `]` → overview, `]` → tasks, `[` `[` → back to the atlas with the
work view still active.
