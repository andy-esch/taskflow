---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Epics are write-once via the CLI (only new/list/show); add epic set (fields incl. status) + epic edit ($EDITOR), mirroring task set/edit + the picker.
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli]
created: "2026-06-25"
updated_at: "2026-06-25"
started_at: "2026-06-25"
completed_at: "2026-06-25"
id: 6ffr4wc021q2
---
# Epic mutation: `epic set` + `epic edit` (parity with tasks)

## Objective

Epics are **write-once** via the CLI: `epic new` can set `--status`/`--priority`/
`--description`/`--tags` (you can even create a `completed` epic), but nothing can
*change* them afterward — there's no `epic set`, no `epic edit`, no lifecycle verb.
Advancing an epic `planning → in-progress → completed` means hand-editing the file.
Give epics the same two mutation faces tasks have.

## Context

`epic` today is only `{new, list, show}`. The pieces already exist:
- `domain.AllEpicStatuses()` + `domain.ValidateEpicStatus` (closed vocab:
  planning|in-progress|completed|archived) — validation is ready.
- `task set` (agent: field-level, atomic, `--dry-run`, `task_mutation` JSON) and
  `task edit` ($EDITOR, re-validated on save) are the templates.
- The picker (`resolveOne` + `epicOptions`) and the editor flow (`internal/editor`,
  edit.go) are in place from the 2026-06-25 pickers/width work — reuse both.

Relates to epic 20 (CLI UX). Sibling tasks: lint-covers-epics, audit-editing.

## Acceptance criteria

- [ ] `epic set <id> [--status|--priority|--description|--tags|--set k=v|--unset k]`
      — surgical, validated, single atomic write; `--dry-run` preview; an
      `epic_mutation` (or reused) JSON envelope; status checked via
      `ValidateEpicStatus`; bare invocation picks on a TTY (resolveOne + epicOptions).
- [ ] `epic edit [id]` — $EDITOR on the whole file, re-validated on save, picker on
      bare invocation, exit-11 non-interactively (mirror `task edit` exactly,
      including the `--dry-run` rejection).
- [ ] The epic ID (= filename stem) is NOT renamed by `set` (frontmatter only).
- [ ] go build/test/lint green; docs/cli regenerated; schema_version bumped if a new
      envelope/field is added.

## Implementation sketch

- `core.SetEpicFields(id, updates, …)` + (if needed) `ReplaceEpicBody`, mirroring
  `SetFields`/`ReplaceBody`; surgical frontmatter via the store's atomic helpers
  (preserve unknown fields, comments, key order).
- `render` envelope mirroring `task_mutation` (or generalize it).
- `epic.go`: `newEpicSetCmd`/`newEpicEditCmd` mirroring the task ones; reuse
  `resolveOne`, `epicOptions`, `fillSelect`, and the editor wiring.

## Risks / gotchas

- Decide whether lifecycle should ALSO get named verbs (`epic start/complete/…`) for
  full task-parity, or whether `epic set --status` is the intended single path
  (recommend the latter first; named verbs are a possible follow-up).
- `archived` is the epic "withdrawn" state — make sure rollups/`status`/`epic list`
  still treat it sensibly after a set.
- Surgical frontmatter must not reorder or drop keys (same contract as `task set`).

## Done when

`epic set 01-x --status in-progress` and `epic edit 01-x` both work — validated,
atomic, JSON, picker — so an epic is no longer write-once.

## Decision 2026-06-25 — depends on the new epic vocab

BLOCKED ON [[redefine-epic-status-model-to-active-retired-deprecated-and-migrate]]: target the new active/retired/deprecated vocabulary, not the old set. Command shape: 'epic set --status retired' (+ --priority/--description/--tags) is the path — named lifecycle verbs (epic retire/deprecate) are a possible later nicety, NOT in this task.

## Decision 2026-06-25 — status movement split out

Status movement (active/retired/deprecated) is now [[epic-status-movement-via-epic-move-cli-and-the-m-action-menu-tui]] (a real 'epic move' verb + the TUI 'm' action menu + completion, mirroring task/audit). So THIS task is **non-status fields + body only**: 'epic set <id> --priority|--description|--tags' (NO --status) + 'epic edit [id]' ($EDITOR). Plus TUI parity for field-edit ('e') and $EDITOR ('E') if tasks have them.
