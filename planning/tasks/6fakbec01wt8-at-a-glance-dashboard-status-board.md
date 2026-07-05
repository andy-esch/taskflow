---
status: completed
epic: 17-pm-go-cli
description: A no-arg or status command showing counts per status, in-progress items, and epic progress bars for a one-screen overview
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, cli, ergonomics]
created: "2026-06-09"
updated_at: "2026-06-09"
started_at: "2026-06-09"
completed_at: "2026-06-09"
id: 6fakbec01wt8
---

# At-a-glance dashboard (status board)

## Objective

The single most useful *human* command a planning tool can have: a one-screen
"where am I?" board. (pm had `index`.) Counts per status, what's in-progress,
epic progress bars, maybe stale items.

## Done

- **`tskflwctl status`** (read-only, `--json`) — counts per status (active vs
  archived, glyphs, zero buckets hidden), the in-progress working set (with
  relative dates), epic **progress bars** (`Style.Bar` + colored percent +
  done/total), and a `⚠ misfiled` / `! unreadable` line for data hygiene.
- **`core.Service.Summary()`** returns a typed `Summary{Counts, InProgress,
  Epics, Misfiled, Problems}` from a single tasks+epics scan; extracted a shared
  `rollupEpics` (now used by `ListEpics` too). `cli` stays thin.
- **`render.SummaryHuman`/`SummaryJSON`** reuse the color/table/width layer.
- Tests: `Service.Summary` against the fake store (counts/in-progress/rollup/
  misfiled); `Style.Bar`; CLI `status` smoke + `--json` shape.

Decided against hijacking bare `tskflwctl` → keep `--help` for discovery; the
board is the explicit `status` command (trivial to add bare-root later).
"Stale" deferred — `updated_at` now exists on `Task`, so it's an easy follow-on.

## Acceptance

- [x] `status` shows counts + in-progress + epic progress bars; `--json` carries
      a versioned summary; reuses the render styling; tests + lint green; demoed
      on real planning (17-pm-go-cli at 77%).

## Open questions

- "Stale" (oldest `updated_at` among active) — a clean follow-on now that
  `Task.Updated` exists.

## Out of scope

- Interactive navigation of the board — that's the planned Bubble Tea TUI, which
  reuses the same `core.Service.Summary()`.

## Related

- Epic [[17-pm-go-cli]]; builds on the render styling
  [[cli-color-glyphs-table-headers-render-styling]].
