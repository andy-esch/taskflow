---
schema: 1
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Derive a working/fresh/dormant liveness signal from the epic task rollup (no new stored status) so the TUI dashboard shows live epics bright, dims dormant ones, and hides retired/deprecated.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, epics, dashboard]
created: "2026-06-28"
updated_at: "2026-06-28"
completed_at: "2026-06-28"
id: 6fgq1n002kg5
---
# Epic liveness — derive working/dormant so the dashboard shows only live epics

The TUI dashboard blindly lists every epic, so finished/quiet domain buckets
crowd out the live ones. The fix is **not** a new stored status — it's reading
the task rollup we already compute. Decided with the user 2026-06-28.

## The model (decided)

Two axes, kept separate:

- **Status (intent, stored):** `active` / `retired` / `deprecated` — unchanged.
  This is the closed vocabulary in `domain/epic.go` (decided 2026-06-25). No new
  value, no migration, no lint/schema change. Status changes rarely and
  deliberately (a *domain* closes), not as work ebbs and flows.
- **Liveness (activity, derived):** `working` / `fresh` / `dormant` — computed
  from the epic's task rollup, never stored, only meaningful for `active` epics.

Epics deliberately get **no "completed"/"done" status**: they are perpetual
domain buckets whose arc is liveness (working ↔ dormant, reversible), not
completion. The start→ship→end arc belongs to **projects** (not yet formalized);
keeping epics off a completion lifecycle reserves that semantics for projects and
keeps `retired` meaning "this domain is closed," not "this feature shipped."

## Liveness classifier

A pure function on `EpicSummary` (the rollup already carries `Total`, `Done`,
`Deprecated`, `LastUpdated` — see `core/service_epic.go`). `Total` already
excludes withdrawn/deprecated tasks via `TaskRollup`.

```
open := Total - Done
working : open > 0              // pending/in-progress work exists   -> bright
fresh   : Total == 0           // declared bucket, no tasks yet       -> bright
dormant : Total > 0 && open == 0  // had work, all done/withdrawn    -> dimmed
```

The `Total > 0` guard keeps a just-created empty epic out of the dormant band —
*new* and *drained* are different and both stay visible. A dormant epic
auto-wakes to `working` the moment a task is filed against it; zero bookkeeping.

## Dashboard behavior (decided: dim, keep in list)

- Default epics view = **`active` only**.
- Sort **working/fresh first (bright), dormant below (dimmed)**, tiebroken by
  `LastUpdated` descending.
- `retired` / `deprecated` excluded from the default, revealed by a toggle key.
- Nothing vanishes — one glance still shows the whole live domain map, with quiet
  buckets visibly receded.

## Scope / where it lands

- **core** — add the derived bit beside `TaskRollup` / `EpicSummary` (e.g.
  `EpicSummary.Open()` + a `Liveness()` returning working/fresh/dormant), so the
  rule lives in ONE place like the rollup does. Unit-test the classifier
  (boundaries: `Total==0`, `open==0 & Total>0`, `open>0`).
- **core `--json`** — expose `liveness` and `open` on `epic list --json` so no
  surface re-derives the rule.
- **tui** — `epicDelegate` dims dormant rows; epic list default-filters to
  `active` and sorts working-first; a key toggles retired/deprecated. No logic in
  `View`/`Update` — read liveness off the summary loaded via `core.Service`.
- **domain/epic.go** — sharpen the `active` doc comment to name the
  working/dormant derived split (documentation only; the vocabulary is unchanged).
- **cli (optional)** — mirror the dim/sort in `epic list -o table`.

## Out of scope

- Any change to the stored epic-status vocabulary (no `dormant` status).
- Projects / the project lifecycle — referenced only to explain why epics have no
  completion state.
- Staleness-by-date as a classifier input — `open`-count is the crisp signal;
  `LastUpdated` is used only for sort tiebreaking here.

**Verified shipped & completed 2026-06-28.** Landed in commit fb2eb80 (PR #65, feat: epic liveness): core EpicSummary.Open()/Liveness()/Live() (service_epic.go) pinned by TestEpicSummary_Liveness; open + liveness on `epic list --json`; TUI active-default view + working-first sort + dormant dimming (statusview.go epicViews, commands.go filterEpicsByView/sortEpicsForView, item.go epicDelegate) + the ? legend. The task was just left at next-up after the work merged; closing it.
