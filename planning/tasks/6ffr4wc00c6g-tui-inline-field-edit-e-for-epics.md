---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Extend the TUI inline 'e' field-editor to epics (currently task-only)
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui]
created: "2026-06-25"
updated_at: "2026-06-25"
started_at: "2026-06-25"
completed_at: "2026-06-25"
id: 6ffr4wc00c6g
---

# TUI inline field-edit (e) for epics

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)

## What's left

`epic set`/`epic edit` (CLI) + the entity-agnostic `E` ($EDITOR on the whole
file, TUI) now give epics field + body mutation. The one remaining parity gap is
the **TUI inline `e` field-editor**, which is still task-only.

It was NOT extended to epics in the `epic set`/`epic edit` work because the inline
form is deeply task-shaped:
- `internal/tui/edit.go` `editMenu` holds a task `slug`; `editableFields(domain.Task)`
  builds the field list from typed TASK fields (description/priority/tags/effort/tier
  — effort/tier don't exist on an epic).
- `setFieldCmd` calls `core.Service.SetFields` (task); entry is gated by
  `m.selectedTask()` in `model.go`.

The core + render plumbing it needs already exists: `core.SetEpicFields` and the
`epic_mutation` JSON envelope (schema_version 1.14). The follow-up is purely a TUI
generalization:
1. Parameterize `editMenu`/`editableFields` over the entity (an epic field set:
   description/priority/tags — no effort/tier), or add an epic-specific builder.
2. Route submit to `SetEpicFields` vs `SetFields` by selected entity.
3. Open `e` when `m.selectedEpic()` (or generalize `selectedTask`), updating the
   `model.go` Edit handler + the `TestModel_EditTasksOnly` pin.
4. Mirror the task inline-edit test for an epic.
