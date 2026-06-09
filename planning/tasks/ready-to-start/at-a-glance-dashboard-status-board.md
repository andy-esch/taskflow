---
status: ready-to-start
epic: 17-pm-go-cli
description: A no-arg or status command showing counts per status, in-progress items, and epic progress bars for a one-screen overview
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, cli, ergonomics]
created: "2026-06-09"
---

# At-a-glance dashboard (status board)

## Objective

The single most useful *human* command a planning tool can have: a one-screen
"where am I?" board. (pm had `index`.) Counts per status, what's in-progress,
epic progress bars, maybe stale items.

## Implementation sketch

- [ ] A `status` command (and make bare `tskflwctl` show it) — read-only, `--json`.
- [ ] Compose existing `core` reads (ListTasks/ListEpics) into a summary:
      counts per status, the in-progress list, per-epic progress bars (reuse
      `Style.Percent` + a bar), optional "stale" (oldest `updated_at`).
- [ ] New `core.Service.Summary()` returning a typed struct (keeps `cli` thin);
      `render.SummaryHuman`/`SummaryJSON` over it (reuse the color/table layer
      from [[cli-color-glyphs-table-headers-render-styling]]).
- [ ] Tests: summary counts against a fake store; CLI smoke + `--json` shape.

## Open questions

- "Stale" needs `updated_at`, which isn't on `domain.Task` yet — add the field,
  or derive recency another way. Decide when building.

## Out of scope

- Interactive navigation of the board — that's the planned Bubble Tea TUI, which
  reuses the same `core.Service.Summary()`.

## Related

- Epic [[17-pm-go-cli]]; builds on the render styling
  [[cli-color-glyphs-table-headers-render-styling]].
