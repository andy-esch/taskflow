---
schema: 1
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Generalize core.Summary() into the canonical planning projection (structured data) and add a board command (human + JSON) that renders it; closes the projection-shape question. Per ADR-0003.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [core, cli]
created: "2026-07-01"
updated_at: "2026-07-03"
started_at: "2026-07-02"
completed_at: "2026-07-02"
id: 6fhnydm01vz7
---
# Core read-model projection and board command

## Objective

Lift the pattern behind `core.Summary()` into a first-class planning projection
and expose the active-work slice through a new `board` command. This closes epic
24's last open design piece — the projection shape — by shipping the read side
(the same projection the TUI renders today and a web endpoint will render later),
without yet touching the storage model (ADR-0003's stable-key layout stays in the
downstream epic-24 tasks).

## Acceptance criteria

- [x] `core.Board()` projection: active statuses (next-up → ready-to-start →
  in-progress) as ordered task columns, terminal/parked excluded, `FileProblem`s
  surfaced not swallowed (mirrors `Summary`).
- [x] `domain.ActiveStatuses()` owns the active-set definition (AllStatuses
  filtered by IsActive) — one order, one owner.
- [x] `board` command: human render (per-status sections, `(none)` for empty,
  problems footer) and `--json` (`BoardEnvelope`, schema_version 1.23), read-only
  annotation, non-zero exit on unreadable files.
- [x] TUI landing view renamed "dashboard" → "Overview" (display-only; `:o`
  command + shorthand behind an `overviewName` const).
- [x] Tests: core order/empty-column, CLI smoke + JSON + exit-code, wire envelope
  validation, golden `board_json`; `docs/cli` regenerated.

## Out of scope

- Storage-model change (status-in-frontmatter, flat id-led layout) — separate
  epic-24 tasks, per ADR-0003.
- A committed `BOARD.md` artifact — deferred to `serve` (ADR-0004); `board` is
  on-demand only.
- A TUI "board" tab — considered and declined; the active slice is the Overview
  plus the `task` views.

## Related

- Epic [24-data-model-evolution-stable-key-storage-read-model-content-occ](../epics/24-data-model-evolution-stable-key-storage-read-model-content-occ.md)
- ADR [0003-stable-key-id-addressed-storage](../adrs/0003-stable-key-id-addressed-storage.md) — projection-shape question
