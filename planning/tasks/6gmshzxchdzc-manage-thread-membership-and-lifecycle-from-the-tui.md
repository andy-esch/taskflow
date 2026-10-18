---
schema: 1
id: 6gmshzxchdzc
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Allow routine Thread curation from the existing list/detail/graph presentations instead of requiring CLI round trips for every membership or lifecycle change.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-17"
depends_on: [6gmshzw7we0m]
updated_at: "2026-10-17"
---
# Manage Thread membership and lifecycle from the TUI

## Objective

Allow routine Thread curation from the existing list/detail/graph presentations instead of requiring CLI round trips for every membership or lifecycle change.

## Scope

Add/remove existing task members through searchable, identity-aware selection and expose the established start/complete/cancel/reopen lifecycle as applicable. Reuse core Thread projections and guarded mutation services.

## Acceptance criteria

- [ ] Thread detail exposes discoverable add/remove-member and applicable lifecycle actions with human labels and canonical identity disambiguation.
- [ ] Membership updates preserve repository-global dependency edges; external prerequisites remain visibly distinct from members.
- [ ] Lifecycle eligibility and completion/cancellation rules come from core, with refusals, impacts and partial/finalization receipts shown without a second UI policy.
- [ ] Selected Thread/task endpoints cannot be rebound by rename, deletion, impostor labels, refreshed selections or workspace switching; supply bounded selected-entry seams if necessary.
- [ ] Watcher refresh preserves the intended Thread/node selection, and tests/help/docs plus normal gates cover successful, refused, stale and partial outcomes.

## Design attention

No drag-to-link graph editing is required. Adding a task to a Thread is not a dependency declaration, and cancelling a Thread does not cancel its tasks.

## Out of scope

Creating/deleting tasks as a side effect, graph layout changes, bulk manifests, automatic edge cleanup, or new Thread semantics.

## Related

- [Authoring capability boundary and sequencing](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Owning service_thread.go](../../internal/core/service_thread.go).
- [Owning thread_mutation.go](../../internal/core/thread_mutation.go).
- [Owning thread_projection.go](../../internal/tui/thread_projection.go).
