---
status: completed
epic: 17-pm-go-cli
description: 'Follow-ups deferred from the 2026-06-13 audit: epic numbering, scan dedup, Move CAS, epic JSON shape, TUI/completion layout dup, SetFields bool params'
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [audit, refactor, tech-debt]
created: "2026-06-13"
updated_at: "2026-06-14"
started_at: "2026-06-14"
completed_at: "2026-06-14"
---

# Address deferred code-audit findings (numbering, dedup, CAS, JSON layout)

## Objective

The 2026-06-13 codebase audit (`planning/audits/open/2026-06-13-codebase-quality-architecture.md`)
fixed the high/quick findings inline; this task collects the ones deliberately
deferred because they're larger refactors or design decisions rather than
one-line fixes. None is urgent — the code is correct today — but each removes a
real drift or robustness gap. Do them as independent commits.

## Acceptance criteria

- [x] **M3 — epic numbering.** Fixed the reproducible half: epics now sort by
      parsed `NN-` number (`epicNum` + `sort` in `store/ListEpics`), so `100` sorts
      after `99` regardless of zero-pad; `TestFS_ListEpics_NumericOrder`. The
      concurrent duplicate-number half is **documented as accepted** in
      `nextEpicNumber` — it needs OS-level locking for a race a single-user local
      CLI (no daemon) can't produce. (`store/create.go`, `epicstore.go`)
- [x] **L4 — `Move` compare-and-swap.** Added the re-resolve-before-write CAS
      guard (mirrors `SetFields`) + `testHookBeforeMoveWrite`; a concurrently-moved
      task now yields `ErrConflict` with nothing written.
      `TestFS_Move_ConflictsWhenMovedConcurrently`. (`store/fsstore.go`)
- [x] **L2 — scan dedup.** Extracted generic `scanDir` (the List loops) and
      `markdownCandidates` (the resolution loops) in `resolve.go`; routed all six
      sites through them. (`store/fsstore.go`, `epicstore.go`, `auditstore.go`,
      `resolve.go`)
- [x] **L6 — epic JSON shape.** `epicJSON` now embeds `epicMetaJSON` via a shared
      `toEpicMeta` helper, so `epic list`, `epic show`, and the dashboard share one
      meta shape (rollup additive). (`cli/render/render.go`)
- [x] **L8 — layout duplication.** Resolved as part of
      [[put-storage-layout-knowledge-back-behind-the-port]] (M15a): the store now
      owns `WatchPaths()` and the TUI consumes it. (Completion globbing is the
      documented, service-free exception and was left as-is.)
- [~] **M6 (optional) — `SetFields` bool params.** **WONTFIX** (judgment call):
      the Store port mutators are single-bool (idiomatic); only `Service.SetFields`
      has the `force, dryRun` adjacency, and no third flag is imminent. Revisit
      with a `MutateOpts` value only if one lands. (`core/service.go`)
- [x] `go build ./... && go test ./... && go vet ./...` green; `gofmt` clean.

## Out of scope

- The findings already fixed inline (H1, H2, M1, M2, M4, M5, L1, L3, L5, L7, L9) —
  see the audit's progress log.
- The TUI help-scroll bug (M5) — already fixed.

## Related

- Epic [[17-pm-go-cli]]
- Audit: `planning/audits/open/2026-06-13-codebase-quality-architecture.md`
