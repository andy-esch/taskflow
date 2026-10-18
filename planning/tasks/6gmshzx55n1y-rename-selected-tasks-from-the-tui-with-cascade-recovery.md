---
schema: 1
id: 6gmshzx55n1y
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Make task title changes a native, deliberate operation that preserves stable identity and the existing multi-document rename/cascade recovery contract.
effort: 1 day
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6gmshzw7we0m]
updated_at: "2026-10-17"
---
# Rename selected tasks from the TUI with cascade recovery

## Objective

Make task title changes a native, deliberate operation that preserves stable identity and the existing multi-document rename/cascade recovery contract.

## Scope

Offer a discoverable rename form showing current and proposed title in the active planning space. Invoke a selected-identity application path, with bounded core/store changes if the existing reference-based RenameTask cannot safely consume a prior TUI selection.

## Acceptance criteria

- [ ] A task can be renamed from the TUI without invoking the CLI or editing its filename manually; the same stable task remains selected afterward.
- [ ] Deleted/renamed sources, ID-shaped impostor slugs and workspace switching cannot redirect the selected action.
- [ ] Display affected-link/cascade context where available and preserve the exact core receipt on conflicts or committed-but-incomplete outcomes.
- [ ] Do not report partial completion as success or blindly retry/remint; render inspection/reconciliation guidance and reload coherent task/Thread views.
- [ ] Cancellation, duplicate submission, watcher refresh, successful/no-op rename and hostile identity/partial-failure paths have focused tests and updated help/docs.

## Design attention

Rename is not an ordinary title field setter: it may change a path and multiple inbound links. Keep semantic policy and recovery in core and persistence.

## Out of scope

Renaming other entity kinds, changing stable IDs, generic file moves, or automatic rollback/reconciliation.

## Related

- [Authoring capability boundary and sequencing](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Owning task_rename.go](../../internal/core/task_rename.go).
- [Owning service_task.go](../../internal/core/service_task.go).
- [Owning rename.go](../../internal/store/rename.go).
