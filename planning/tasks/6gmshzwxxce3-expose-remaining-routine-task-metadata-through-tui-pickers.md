---
schema: 1
id: 6gmshzwxxce3
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Close the routine task-field gaps in the inline editor — notably epic reassignment and autonomy — without raw frontmatter editing.
effort: 1 day
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6gmshzw7we0m]
updated_at: "2026-10-25"
audited: "2026-10-25"
---
# Expose remaining routine task metadata through TUI pickers

## Objective

Close the routine task-field gaps in the existing inline editor, especially epic reassignment and autonomy, without introducing raw frontmatter editing.

## Scope

Audit existing editableFields against routine core task fields, then add the justified missing fields with typed widgets and contextual help. Preserve existing description, priority, tags, effort, tier and revisit behavior.

## Acceptance criteria

- [ ] Record the concrete supported field set and deliberately excluded tool-owned/graph/lifecycle fields against current core contracts.
- [ ] An existing task can select a valid epic and change autonomy through discoverable native actions with current values and field meanings visible.
- [ ] Writes retain selected task and chosen epic intent across rename, deletion, duplicate labels and workspace switching; add bounded core support where a fresh-reference API cannot guarantee that.
- [ ] Save/cancel and validation behavior follows the reviewed action contract, preserves failed input, and retains partial/finalization feedback without blind replay.
- [ ] Focused interaction/identity/watch tests, normal code gates and owning help/architecture updates pass.

## Design attention

Keep lifecycle transitions, dependencies and title/cascade rename in their dedicated actions. Do not copy descriptor arrays or invent a second validation policy.

## Out of scope

Custom/unknown field editing, raw IDs/timestamps, schema controls, a full replacement for the existing editor, or cross-space batch edits.

## Related

- [Authoring capability boundary and sequencing](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Owning edit.go](../../internal/tui/edit.go).
- [Owning service_task.go](../../internal/core/service_task.go).

## Sweep refresh (2026-10-25)

2026-10-25: automated weekly sweep — the field-gap premise verified accurate
against the current editor; AC 1's inventory is partly answerable from code
today.

- Both cited paths exist: `internal/tui/edit.go` (487 lines) and
  `internal/core/service_task.go` (1045 lines).
- `editableFields` (`internal/tui/edit.go:55-77`) currently offers exactly
  `description`, `priority`, `tags`, `effort`, `tier`, plus `revisit_at` only
  while the task is `deferred`. That matches the Scope line's "preserve existing
  description, priority, tags, effort, tier and revisit behavior" with no drift,
  and confirms the two gaps the Objective names: **`epic` and `autonomy_level`
  are both absent.**
- `status` is excluded by an explicit design comment at `edit.go:52`, consistent
  with the Design attention section keeping lifecycle in its own action.
- Field meanings already come from `domain.AuthoringFields` via
  `fieldDescs("task")` (`edit.go:56`), so AC 2's "field meanings visible"
  has an existing source and needs no new descriptor array — which is also what
  Design attention asks for ("do not copy descriptor arrays").
- AC 3's selected-intent requirement has a precedent to copy rather than invent:
  the epic form already submits through `SetSelectedEpicFields`
  (`internal/core/service_epic.go:323`) with an adapter-established selection
  (`edit.go:82`), and the task side has `SetSelectedTaskFields`
  (`internal/core/service_task.go:772`). Both take a `*MutationTarget`, so the
  "add bounded core support where a fresh-reference API cannot guarantee that"
  escape hatch is likely unnecessary for this task's fields.

Nothing ticked: AC 1 asks for a *recorded* decision on the excluded set, which
is a human call, and the sweep only supplied the raw inventory above.

`description` tightened to 130 chars (was 152, over the 150-char sweep
threshold); meaning unchanged, and `## Objective` keeps the original sentence.
