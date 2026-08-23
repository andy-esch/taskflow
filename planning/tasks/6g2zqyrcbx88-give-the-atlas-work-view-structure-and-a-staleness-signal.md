---
schema: 1
id: 6g2zqyrcbx88
status: completed
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Surface how long each task has been in progress, add epic and priority columns, and give the work view its own orders including grouping by space.
effort: M
tier: 3
priority: high
autonomy_level: 3
tags: [tui, atlas, ux]
created: "2026-08-23"
updated_at: "2026-08-23"
started_at: "2026-08-23"
completed_at: "2026-08-23"
---
# Give the atlas work view structure and a staleness signal

## Objective

The work view is a flat list that repeats the space label on every row and shows
last-touched dates. It ignores `Epic`, `Priority`, `Effort`, and — most importantly —
`StartedAt`.

**The largest win is `started`, not `updated`.** For in-progress work the useful question
is how long something has been underway; "started 5 weeks ago" is what says a task is
stuck, and `StartedAt` is already stamped on every task that entered in-progress. Nothing
in the tool surfaces it today.

## Acceptance criteria

- [x] Each row shows how long the task has been in progress, from `StartedAt`, falling
  back to the existing date when a task carries none (legacy rows).
- [x] That age is coloured by staleness through a new `theme.Staleness`, built in the same
  shape as the existing `theme.Percent` so the palette owns the thresholds, not the view.
- [x] Rows carry epic and priority columns, priority via the existing `s.priorityText`,
  all aligned across the visible set.
- [x] `o`/`O` work in this view, cycling `started` (default) / `space` / `priority`, with
  the active order named in the header exactly as the spaces view names its own. The
  footer stops omitting them here.
- [x] The `space` order groups rows under a per-space heading and drops the now-redundant
  per-row space label; the other orders stay flat and keep it.
- [x] The cursor stays on the same task across an order change and across a refresh.
- [x] Description remains the last column and is the first thing sacrificed on a narrow
  terminal.

## Out of scope

- An aggregate banner. Still deliberately deferred — living with the view is what answers
  whether one earns its rows.
- A third "attention" view.
- Filtering (`/`) — tracked separately.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Design: [The atlas as a dashboard of dashboards](../research/6g2qtp0022t7-the-atlas-as-a-dashboard-of-dashboards.md)
- Builds on: [Add a cross-space work view to the atlas](6g2nnkffgyeg-render-the-cross-space-in-progress-rail-in-the-atlas.md)

## Implementation notes (2026-08-23)

The work view now carries epic and priority columns, orders on its own axis, and — the
point of the task — reports **how long each task has been underway** rather than when it
was last touched.

`theme.StartedDate` prefers `started_at` and falls back to the ordinary date for rows that
predate the stamp. `theme.Staleness(days)` colours it, built in the same shape as the
existing `theme.Percent` so the palette owns the thresholds rather than the view. Those
thresholds are deliberately generous — nothing is flagged before a month, because planning
is not a ticketing system and a fortnight on a hard task is ordinary. An unknown or future
date is neutral, never alarming; `theme.DaysSince` returns -1 for both so the two cases
cannot be confused with "started today".

`o`/`O` now work here, cycling `started` → `space` → `priority`. **Grouping is a property
of the order, not a separate toggle**: only the space order draws per-space headings,
because grouping by start date or priority would emit one heading per row. Grouped rows
drop the bracketed space label — the heading already carries it — and the freed width goes
to descriptions. The cursor holds its task across a reorder.

Reused rather than rebuilt: `priorityRank` already existed in `sort.go`, and a duplicate
was written and then removed once the collision surfaced. `s.priorityText` supplies the
colouring.

Two test-authoring mistakes worth recording, both caught before landing. An alignment
assertion measured `strings.Index` **byte** offsets, which reports phantom shear because
rows carry different counts of multi-byte `·` separators; it measures display columns now.
And an age assertion searched the whole rendered view for "yesterday" while the fixture's
own description contained the word — the test was matching itself.

Validation: `go test -race ./...`, `golangci-lint run ./...`, planning lint, generated
docs, `git diff --check`. Verified by recording: stalest-first by default, grouped headings
with per-space counts on `o`, priority coloured, cursor preserved.

Note for whoever runs the gate: `FORCE_COLOR` in the environment makes two unrelated
`internal/cli` colour tests fail (pre-existing, reproduces at `v0.17.0`). Use
`env -u FORCE_COLOR go test ./...`.

## Follow-up fixes from live use (2026-08-23)

Two problems the fixture hid, both found by running it against the real registry.

**The projection was never refreshed.** It loaded exactly twice — at `Init()` and on `r` —
so a task started while the TUI was open never appeared, and nothing on screen admitted the
list was a snapshot from program start. `reloadAll` now marks the atlas stale, and the next
entry re-reads; if the atlas is already on screen it re-reads immediately, because a
visibly wrong list is worse than one extra read. Deliberately lazy off-screen: `Overview()`
re-reads every registered planning tree, which is far too heavy for each fsnotify debounce.
This is why the atlas cannot be as live as an entity tab — the watcher covers only the
ACTIVE space, by design, so other spaces still need `r`.

**Right-edge truncation amputated the useful half of every row.** Against
`desirelines-planning` with 76-character slugs, a row wanted 138 columns on a 92-column
terminal, so priority, the staleness age, and the description never rendered — the view's
entire reason for existing was invisible. This was audit finding L1, which had been
deferred as "eventually" *on the strength of the demo fixture*, where spaces are called
`kitchen` and `bike-workshop`. Wrong call, corrected: columns are now fitted, dropped whole
in a declared sacrifice order, with slug and age always surviving. The slug is capped at a
share of the terminal first so one long name cannot starve everything beside it, then grows
back into any spare. The spaces table got the same treatment — it lands at exactly 92
columns with real names and shears at 80.

Lesson worth keeping: the demo fixture's short, tidy names make layout look healthier than
it is. Width work should be checked against the real registry, or the fixture should grow a
deliberately long name.
