---
schema: 1
id: 6gmshzwxxce3
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Close the routine task-field gaps in the existing inline editor, especially epic reassignment and autonomy, without introducing raw frontmatter editing.
effort: 1 day
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6gmshzw7we0m]
updated_at: "2026-10-17"
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
