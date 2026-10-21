---
schema: 1
id: 6g4ym5y2dkp8
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Create tasks through staged TUI forms and existing epic context, preserving input, origin-space binding, core defaults and truthful creation receipts.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 3
tags: [entity]
created: "2026-08-29"
updated_at: "2026-10-21"
audited: "2026-10-11"
audit_sources: [2026-10-11-weekly-task-sweep]
depends_on: [6gmshzw7we0m, 6gmy4c0x6e31]
started_at: "2026-10-20"
completed_at: "2026-10-21"
---
# Create tasks from the TUI and epic context

## Objective

Create tasks without a CLI round trip, both from the tasks tab and from an existing
epic's detail context. This is the first staged creation slice, not all entity
creation in one change.

## Scope and implementation constraints

Collect fields before one explicit submission through the selected-epic core
task-creation use case. The existing per-field editor mutates an existing record; reuse its widgets
where useful, not its write-on-every-field semantics. Use actual creation inputs
and defaults, not `domain.FieldDoc.Required` (persisted requiredness differs from
required user input).

Bind the form to its originating service/source set and preserve the selected
epic's intent using the completed selected-parent prerequisite, not fresh reference
matching or a second TUI parent-selection policy. Preserve entered
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

- [x] Create a new task from the tasks tab.
- [x] Create a new task from an epic's page, with that epic pre-filled and not re-asked.
- [ ] Create a new audit from the audits tab. · **tracked:** by 6gmshzwphadw — Additional entity creation is split from the first staged task-creation slice.
- [ ] Create a new epic from the epics tab. · **tracked:** by 6gmshzwphadw — Epic creation follows the proven staged form; this task retains task creation from an existing epic context.
- [ ] Create a new research doc from the research tab. · **tracked:** by 6gmshzwphadw — Research creation is delivered with the additional entity forms.
- [x] Creation uses core application services and explicit origin-space binding,
  without direct persistence, duplicated validation, or I/O outside tea.Cmd; add
  bounded selected-context seams if required.
- [x] Core validation refusals preserve all entered values; identify the
  offending field where core evidence supports it, otherwise show a form-level
  error without guessing policy from error strings.
- [x] After a successful create the list reloads, the cursor lands on the new entity, and the fs-watcher reload does not move it again.
- [x] Page-local c creates a task on Tasks without changing n/N search;
  epic-context task creation is separately named. Create, submit and cancel are
  discoverable through shared action/key/help routes.
- [x] Workspace switches, duplicate submissions, selected epic drift and
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

## Implementation checkpoint (2026-10-20)

- Tasks-page `c` and the named palette action open the same staged form; the
  epic-context palette action preselects its adapter-established source ID.
  Epic-page `c` does not create a task. Other entity forms remain separate work.
- Title, searchable epic, tags and description precede collapsed metadata options.
  Owner feedback aligned priority/tier/autonomy with the existing edit selectors:
  Enter opens bounded choices, arrows/hjkl select, Enter stages, and Esc backs out
  without applying the candidate. An explicit default choice preserves core defaults.
  Tab/Shift+Tab move focus; Enter in text advances and only the explicit Create
  action submits. Dirty cancellation defaults to Keep editing. Narrow forms stack
  labels and scroll fields while retaining the destination and action; resize
  notices never become invisible mutation surfaces.
- Commands consume `NewTaskForSelectedEpic`, retaining their originating Service,
  selected epic and copied input. Known refusals keep the draft; committed cleanup,
  uncertain outcomes and incomplete receipts disable replay and expose recovery
  evidence. Workspace activation keeps a non-submittable recovery copy; an issued
  origin result updates that copy without navigating the incoming space.
- Selection is carried through pending list generations, status/filter widening
  and watcher refresh. Successful creation offers the existing selected-ID `E`
  editor action for body prose, without automatically launching an editor.
- [Tests](../../internal/tui/task_create_test.go) cover both entry points, defaults
  and overrides, no partial writes, cancellation, stale parent identity, duplicate
  submission, portable/absent/read-only capabilities, catalog ordering and identity
  quarantine, recovery receipts, session switches, concurrent watcher landing,
  terminal fit and a complete running-program keyboard flow.
- Validation: fresh `just build`, full `just test` race suite, `just lint`,
  planning/link lint, audit lint and `git diff --check` pass.

Owner interaction feedback and the completed Codex review are addressed. No new
core API, persisted schema, machine-output revision or CLI command metadata is
introduced by this UI slice.

### Owner feedback and quality lookback

The owner exercised creation and approved its overall interaction, requesting the
bounded selectors now shared with `e`. The quality pass separates pure widget/state
construction from [task-specific commands](../../internal/tui/task_create_commands.go),
replaces mutually exclusive interaction booleans with `taskCreateMode`, and captures
only form/field tokens in delayed widget routing rather than retaining a whole draft.
Pathless success no longer advertises an unavailable editor. Existing task enum
choices/rendering are reused without moving UI types or another validation policy
into core. The [next entity-creation task](6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md)
now explicitly owns common-host extraction against genuinely different forms, not
copying task-parent or lifecycle assumptions.

Implementation reviews:

- [Codex: architecture, extension seams and independent test oracles](../audits/6gnqmm1qp9mz-2026-10-20-tui-staged-task-creation-implementation-codex.md): processed; M1 fixed and audit closed.
- Antigravity: async ordering, replay/recovery and rendered interaction; report
  pending in the owner's local audit. On 2026-10-21 the owner accepted proceeding
  without waiting. That unfinished audit is intentionally excluded from this PR;
  later findings will be triaged separately, not represented as a completed review.

Codex reproduced a dropped terminal-paste event across the form. Live paste now
follows the first active modal's optional input handler and only edits visible,
editable fields/search, never blocked modes or underlying widgets. A running-program
regression pastes title, epic search, tags, description and effort through actual
events and creation; reducer checks pin modal precedence, blocked states and no
partial writes. Delayed widget generation guards remain separate. Its extension
assessment fits the already-scoped additional-creation task; no duplicate follow-up
is needed. All criteria in this task are met or explicitly transferred to their
owning tasks; the remaining review is not a completion gate.
Repeated runtime tests also exposed a test-only early-quit race: the success flash
precedes asynchronous selection landing. Both keyboard and paste program tests
now wait for loaded created-task detail before quitting, without timing sleeps.

### Interactive check

Run `just run ui` in a disposable planning space (or deliberately create real work).
On Tasks, press `c`; Tab through title/epic/tags/description, optionally expand More
options, then focus Create and press Enter. Check cancellation first if no record
should be written. From an epic detail, use Ctrl+P and select **New task in …**;
the epic should already be selected. After a successful create, the task remains
selected through refresh; `E` opens its scaffold via the normal optional editor route.
