---
schema: 1
id: 6g4ym5y2dkp8
status: ready-to-start
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Create tasks through staged TUI forms and existing epic context, preserving input, origin-space binding, core defaults and truthful creation receipts.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 3
tags: [entity]
created: "2026-08-29"
updated_at: "2026-10-18"
audited: "2026-10-11"
audit_sources: [2026-10-11-weekly-task-sweep]
depends_on: [6gmshzw7we0m, 6gmy4c0x6e31]
---
# Create tasks from the TUI and epic context

## Objective

Create tasks without a CLI round trip, both from the tasks tab and from an existing
epic's detail context. This is the first staged creation slice, not all entity
creation in one change.

## Scope and implementation constraints

Collect fields before one explicit submission through the core task-creation use
case. The existing per-field editor mutates an existing record; reuse its widgets
where useful, not its write-on-every-field semantics. Use actual creation inputs
and defaults, not `domain.FieldDoc.Required` (persisted requiredness differs from
required user input).

Bind the form to its originating service/source set and preserve the selected
epic's intent. A reference-based API is not automatically safe for a displayed
selection; add a bounded selected-context seam if necessary. Preserve entered
values after validation failures and distinguish a committed create with cleanup
failure from a failed create. Never blindly remint/replay after progress.

Show the target space, use the default scaffold and offer $EDITOR for body prose.
After success, focus the created task without a watcher refresh stealing it.
Follow the shared authoring contract for action discovery, cancellation, dirty
input and narrow-screen behavior. Owner chose `c` for page-local entity creation:
tasks-page `c` creates a task; epic-page `c` creates an epic. Task creation from
epic detail is a separately named contextual action with that epic preselected.

The [adapter-safe gate](../research/6gmrkq00348a-tui-authoring-interaction-checkpoint.md#adapter-safe-implementation-gate)
is part of this slice: reuse per-kind application services, keep form/Bubble Tea
types out of core and prove any new selected-epic seam with portable and guarded
local adapters. No assumed filesystem path or universal CreateEntity API.

The [selected-epic creation foundation](6gmy4c0x6e31-bind-task-creation-to-the-selected-epic-identity.md)
owns that core/port/persistence seam and its portable/real-adapter evidence.
This UI slice consumes it and proves displayed selection, command closure,
origin-service binding and recovery presentation; do not implement a second
parent-selection policy inside the form.

## Scope transfer

The original audit/epic/research creation criteria are explicitly tracked to
[additional entity creation](6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md),
which also includes Threads. The TUI already has a Threads tab; the old exclusion
based on its absence is retired. This slice still creates tasks from existing
epics, with the epic preselected.

## Acceptance criteria

- [ ] Create a new task from the tasks tab.
- [ ] Create a new task from an epic's page, with that epic pre-filled and not re-asked.
- [ ] Create a new audit from the audits tab. · **tracked:** by 6gmshzwphadw — Additional entity creation is split from the first staged task-creation slice.
- [ ] Create a new epic from the epics tab. · **tracked:** by 6gmshzwphadw — Epic creation follows the proven staged form; this task retains task creation from an existing epic context.
- [ ] Create a new research doc from the research tab. · **tracked:** by 6gmshzwphadw — Research creation is delivered with the additional entity forms.
- [ ] Creation uses core application services and explicit origin-space binding,
  without direct persistence, duplicated validation, or I/O outside tea.Cmd; add
  bounded selected-context seams if required.
- [ ] Core validation refusals preserve all entered values; identify the
  offending field where core evidence supports it, otherwise show a form-level
  error without guessing policy from error strings.
- [ ] After a successful create the list reloads, the cursor lands on the new entity, and the fs-watcher reload does not move it again.
- [ ] Page-local c creates a task on Tasks without changing n/N search;
  epic-context task creation is separately named. Create, submit and cancel are
  discoverable through shared action/key/help routes.
- [ ] Workspace switches, duplicate submissions, selected epic drift and
  committed cleanup failures cannot retarget a create, duplicate it, or lose
  entered data/recovery evidence.
- [ ] Any new selected-epic creation seam stays adapter-neutral through
  consumer-owned ports and explicit source-set/mutation-policy binding;
  portable/pathless and real guarded-adapter tests cover stale selection,
  capability refusal and recovery without filesystem fallback. · **tracked:** by 6gmy4c0x6e31 — Selected-parent core/port/persistence implementation and portable/real-adapter evidence belong to the prerequisite; this task proves TUI selection and dispatch integration.

## Out of scope

Other entity creation, template-library selection, bulk creation, bootstrap and
implicit start/next transitions. Long body authoring remains editor-assisted.
Do not introduce a generic descriptor-driven form framework.

## Related

- [Authoring Thread and interaction contract](../threads/6gmshnmnew7p-enable-routine-planning-and-workspace-management-in-the-tui.md).
- [Task creation use case](../../internal/core/service_task.go).
- [Existing field editor](../../internal/tui/edit.go).
- [TUI command, refresh and local-capability contracts](../../docs/architecture/tui.md).
