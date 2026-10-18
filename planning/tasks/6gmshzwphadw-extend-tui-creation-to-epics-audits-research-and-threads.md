---
schema: 1
id: 6gmshzwphadw
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Reuse the proven staged task-creation interaction to create the remaining four planning entity kinds without pretending their required inputs or mutation policies are identical.
effort: 2-3 days
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6g4ym5y2dkp8]
updated_at: "2026-10-17"
---
# Extend TUI creation to epics audits research and Threads

## Objective

Reuse the proven staged task-creation interaction to create the remaining four planning entity kinds without pretending their required inputs or mutation policies are identical.

## Scope

Add entity-specific form definitions and contextual create actions for epics, audits, research and Threads. Use existing New* use cases, their defaults and default body scaffolds; long bodies remain editor-assisted. Threads have a real tab and are explicitly included.

## Acceptance criteria

- [ ] Create an epic, audit, research document and unstarted Thread from their respective TUI contexts through core services.
- [ ] Each kind collects its actual creation inputs and defaults without treating persisted-field Required metadata as a generic creation form specification.
- [ ] Thread initial-member selection, if offered, uses guarded creation and exact intended task identities; creating a Thread does not silently create dependencies or start it.
- [ ] One explicit submission creates one record; duplicate submission, validation refusal, committed cleanup failure and watcher reload cannot create duplicates or steal selection.
- [ ] Forms preserve input and origin-space binding, open/select the created record by stable identity, and pass focused tests, normal code gates and updated help/architecture guidance.

## Design attention

Reuse widgets and staging mechanics, not a universal CreateEntity port. Add a bounded core seam if a real interface contract is missing rather than promising this is adapter-only.

## Out of scope

Bulk import, template-library design, long Markdown editing, implicit cross-entity creation chains and bootstrap.

## Related

- [Authoring capability boundary and sequencing](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Owning service_epic.go](../../internal/core/service_epic.go).
- [Owning service_audit.go](../../internal/core/service_audit.go).
- [Owning service_research.go](../../internal/core/service_research.go).
- [Owning service_thread.go](../../internal/core/service_thread.go).
