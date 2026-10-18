---
schema: 1
id: 6gmshzxksjcn
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Make small sequencing changes possible while inspecting a task or Thread, using explicit dependency direction and existing repository-global graph guards.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6gmshzw7we0m]
updated_at: "2026-10-17"
---
# Edit task dependencies from the TUI with guarded graph feedback

## Objective

Make small sequencing changes possible while inspecting a task or Thread, using explicit dependency direction and existing repository-global graph guards.

## Scope

Provide searchable add/remove dependency actions with clear prerequisite/dependent labels. Resolve selections by adapter-established identity and reuse core graph validation, impact and recovery projections.

## Acceptance criteria

- [ ] From a selected task, add or remove an explicit prerequisite with readable endpoint labels and identity disambiguation; direction is never inferred from visual screen position.
- [ ] Cycles, dangling/ambiguous endpoints and broken-graph constraints are enforced by core; dependency editing is not raw SetFields.
- [ ] Keep both selected endpoint intents bound across refresh, rename, deletion and workspace changes rather than starting a fresh human-reference lookup at submission.
- [ ] Show guard refusals, proposed effects where supported, applied/remaining progress and core remedies; do not replay partial work.
- [ ] Thread membership is unchanged unless a separate explicit action requests it, and focused graph/identity/UI tests plus normal gates and help/docs pass.

## Design attention

A task outside the displayed Thread may still be a valid prerequisite in the same planning space. Avoid hiding global effects behind a local graph view.

## Out of scope

Cross-planning-space edges, graph repair, bulk manifests, force overrides and drag-and-drop graph editing.

## Related

- [Authoring capability boundary and sequencing](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Owning dependency_operations.go](../../internal/core/dependency_operations.go).
- [Owning thread_projection.go](../../internal/tui/thread_projection.go).
- [Owning THREADS_COMPATIBILITY.md](../../docs/THREADS_COMPATIBILITY.md).
