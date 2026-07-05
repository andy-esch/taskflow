---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: epic move <slug> <status> (mirrors task move + completion) and the TUI m action menu for epic active/retired/deprecated transitions. Status movement in both surfaces; split from epic set/edit.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, tui]
created: "2026-06-25"
updated_at: "2026-06-25"
started_at: "2026-06-25"
completed_at: "2026-06-25"
id: 6ffr4wc01g6s
---
# Epic status movement: `epic move` (CLI) + the `m` action menu (TUI) + completion

## Objective

Make an epic's status changeable in BOTH surfaces, the same way tasks/audits move
through their lifecycle — not via `epic set --status`, but a real `move` verb and
the TUI action menu. Targets the new `active / retired / deprecated` vocabulary.

## Context — mirror the existing task/audit patterns exactly

- **CLI**: `task move <task>... <status>` (`internal/cli/task.go`, `newTaskMoveCmd`)
  is the template: `MinimumNArgs(2)`, last arg is the target status, and a
  position-aware `ValidArgsFunction` that offers slugs for the entity args and
  `domain.AllStatuses()` for the final status arg. Both task and audit moves run
  through the generic `runMoves` helper (`internal/cli/moves.go`).
- **TUI**: `internal/tui/action.go` declares `taskTransitions` and
  `auditTransitions` (verb → target); `internal/tui/entity.go` wires per-entity
  transitions into the `m` action menu and a move `tea.Cmd` (`moveAudit` at ~221).
  Epics currently declare **none** — `entity.go` says so explicitly (~70, ~263:
  "epics have no in-TUI lifecycle move"). That's the gap.
- `domain.AllEpicStatuses()` = `{active, retired, deprecated}` (just shipped).
  `epicItem` already has `lifecycleState()` (its current status) for dropping the
  no-op transition.

Was split out of [epic-mutation-add-epic-set-and-epic-edit](6ffr4wc021q2-epic-mutation-add-epic-set-and-epic-edit.md) (which keeps the
NON-status field/body faces). Both surfaces in one pass so they stay consistent.

## Acceptance criteria

- [ ] **CLI** `epic move <epic>... <status>` — mirrors `task move`: validates the
      status against `AllEpicStatuses`, surgical status-field rewrite via a new
      `core.MoveEpic(id, status, dryRun)`, `--dry-run`, runs through `runMoves`
      (same `moves` JSON envelope), and position-aware completion (epic ids for the
      entity args, `AllEpicStatuses()` for the final arg, with the ActiveHelp hint).
- [ ] **TUI** pressing `m` on an epic opens the action menu with transitions to the
      other two statuses (drop the current via `lifecycleState()`), executes via
      `core.MoveEpic` (an epic `moveEpic` `tea.Cmd` mirroring `moveAudit`), and the
      list live-reloads. Add `epicTransitions` (verbs e.g. activate/retire/deprecate
      → target status) and wire epics in `entity.go`. Honor the `:`-verb path if
      task/audit do.
- [ ] go build/test/lint green; **schema_version bump** if the moves envelope newly
      carries epic moves and that changes the contract (likely additive/none — the
      moves envelope already exists; confirm); docs/cli regenerated; goldens updated.

## Implementation sketch

- `core.MoveEpic(id, status, dryRun) (domain.Epic, error)` — resolve, validate
  status (`ValidateEpicStatus`), surgical frontmatter write via the store's atomic
  helpers; mirror `MoveAudit`.
- CLI `newEpicMoveCmd` mirroring `newTaskMoveCmd` + `runMoves`.
- TUI: `epicTransitions` in action.go; wire into the epic entity config in
  entity.go; `moveEpic` cmd; ensure `epicItem.lifecycleState()` returns the status.

## Risks / gotchas

- Epic status is a frontmatter FIELD, not a directory (unlike tasks) — `MoveEpic`
  rewrites the field, it does NOT move the file. The verb name `move` is for UX
  parity; no file moves.
- Don't regress the rollup (task-status based, orthogonal to epic status).
- The `m` menu must drop the no-op (current-status) transition, like task/audit.
- Keep completion position-aware so the status arg never offers epic ids.

## Done when

`tskflwctl epic move 18-tui retired` works (with tab-completion offering
active/retired/deprecated on the status arg), and pressing `m` on an epic in the TUI
opens activate/retire/deprecate and live-reloads — both validated, both green.
