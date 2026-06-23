---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: epic show/list count deprecated tasks in done/total (epic 18 reads 16/17); exclude deprecated from the denominator so rollups reflect real work
effort: Unknown
tier: 2
priority: low
autonomy_level: 3
tags: [cli, core]
created: "2026-06-21"
updated_at: "2026-06-23"
started_at: "2026-06-23"
completed_at: "2026-06-23"
---
## Objective

Exclude **deprecated** tasks from the epic rollup *denominator* so `epic
show`/`epic list` reflect real work, not withdrawn work. Today epic 18 reads
**16/17** purely because the deprecated `tui-sprint-2` (split into S2a/S2b) is
counted in `Total` — a deprecated task is neither done nor pending, so it
shouldn't drag the percentage or imply outstanding work.

## Where

`internal/core/service.go` — the per-epic `done/total` helper shared by
`ListEpics` and `Summary` (~L429–444): it does `es.Total++` for **every** task in
the epic, and `es.Done++` only for completed. The fix is to skip deprecated tasks
when accumulating `Total` (and they're already not counted in `Done`).

## Scope

1. Don't count `deprecated` tasks toward an epic's `Total` (nor `Done`). After
   this, epic 18 reads **16/16**.
2. One regression test on the rollup helper: an epic with a deprecated member
   counts N-1, and `Percent()` reflects it.

## Open question

- [ ] **Deferred tasks:** should `deferred` also leave the denominator? Unlike
      deprecated (withdrawn), deferred is "not now" — arguably still pending work
      that belongs in `Total`. Default recommendation: **keep `deferred` in**,
      exclude only `deprecated`. Confirm during implementation.

## Acceptance criteria

- [ ] The rollup helper excludes deprecated tasks from `Total`; `epic show`/`epic
      list`/`status` percentages reflect it (epic 18 → 16/16, 100%).
- [ ] Test covering the deprecated-exclusion case.
- [ ] `--json` rollup payloads carry the corrected counts; suite + lint green.

## Related

- Epic [[20-cli-ux-and-ergonomics]].
- Surfaced 2026-06-21 reviewing the board (epic 18's confusing 16/17).
- `internal/core/service.go` (the `EpicSummary` rollup helper);
  `internal/cli/render/render.go` (renders `done/total`).

## Review (2026-06-23)

Single-agent adversarial pass: core rule, edge cases (all-deprecated epic → Total=0 no div-by-zero; deferred stays in), JSON contract/schema, goldens (1.10→1.11 + deprecated field only), and docs all verified clean. One MAJOR caught: the **TUI epic-detail pane** (detail.go renderEpicMeta) recomputed progress from len(tasks), so it showed a different % than the epic list/rollup for any epic with a deprecated task (epic 18: list 100%, detail 94%). Fixed to mirror rollupEpics (deprecated out of done+total, surfaced as '(N deprecated)'); added a TUI regression test. Display-only fix — no schema/golden change.
